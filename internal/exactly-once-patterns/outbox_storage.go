// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// OutboxStorage defines the interface for outbox event storage
type OutboxStorage interface {
	// Store saves an outbox event
	Store(ctx context.Context, event OutboxEvent) error

	// GetPending retrieves pending events up to the specified limit
	GetPending(ctx context.Context, limit int) ([]OutboxEvent, error)

	// MarkPublished marks an event as published
	MarkPublished(ctx context.Context, eventID string) error

	// MarkFailed marks an event as failed with retry information
	MarkFailed(ctx context.Context, eventID string, err error) error

	// Cleanup removes old processed events
	Cleanup(ctx context.Context, before time.Time) error
}

// RedisOutboxStorage implements OutboxStorage using Redis
type RedisOutboxStorage struct {
	client *redis.Client
	cfg    *Config
	log    *zap.Logger
	prefix string
}

// NewRedisOutboxStorage creates a new Redis-based outbox storage
func NewRedisOutboxStorage(client *redis.Client, cfg *Config, log *zap.Logger) *RedisOutboxStorage {
	prefix := "outbox"
	if cfg.Outbox.KeyPrefix != "" {
		prefix = cfg.Outbox.KeyPrefix
	}

	return &RedisOutboxStorage{
		client: client,
		cfg:    cfg,
		log:    log,
		prefix: prefix,
	}
}

func (s *RedisOutboxStorage) eventKey(eventID string) string {
	return fmt.Sprintf("%s:event:%s", s.prefix, eventID)
}

func (s *RedisOutboxStorage) pendingKey() string {
	return fmt.Sprintf("%s:pending", s.prefix)
}

func (s *RedisOutboxStorage) failedKey() string {
	return fmt.Sprintf("%s:failed", s.prefix)
}

// Store saves an outbox event
func (s *RedisOutboxStorage) Store(ctx context.Context, event OutboxEvent) error {
	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	pipe := s.client.TxPipeline()

	// Store event data
	pipe.Set(ctx, s.eventKey(event.ID), data, s.cfg.Outbox.RetentionPeriod)

	// Add to pending queue with timestamp score
	score := float64(event.CreatedAt.Unix())
	pipe.ZAdd(ctx, s.pendingKey(), redis.Z{
		Score:  score,
		Member: event.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store outbox event: %w", err)
	}

	s.log.Debug("Stored outbox event",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.EventType))

	return nil
}

// GetPending retrieves pending events up to the specified limit
func (s *RedisOutboxStorage) GetPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	// Get oldest pending event IDs
	eventIDs, err := s.client.ZRange(ctx, s.pendingKey(), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(eventIDs) == 0 {
		return []OutboxEvent{}, nil
	}

	// Fetch event data
	events := make([]OutboxEvent, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		data, err := s.client.Get(ctx, s.eventKey(eventID)).Result()
		if err != nil {
			if err == redis.Nil {
				// Event data missing, remove from pending
				s.client.ZRem(ctx, s.pendingKey(), eventID)
				continue
			}
			return nil, fmt.Errorf("failed to get event data: %w", err)
		}

		var event OutboxEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			s.log.Error("Failed to unmarshal event",
				zap.String("event_id", eventID),
				zap.Error(err))
			continue
		}

		events = append(events, event)
	}

	return events, nil
}

// MarkPublished marks an event as published
func (s *RedisOutboxStorage) MarkPublished(ctx context.Context, eventID string) error {
	pipe := s.client.TxPipeline()

	// Remove from pending queue
	pipe.ZRem(ctx, s.pendingKey(), eventID)

	// Update event with published timestamp
	data, err := s.client.Get(ctx, s.eventKey(eventID)).Result()
	if err == nil {
		var event OutboxEvent
		if json.Unmarshal([]byte(data), &event) == nil {
			now := time.Now().UTC()
			event.PublishedAt = &now
			event.Status = "published"

			if updatedData, err := json.Marshal(event); err == nil {
				pipe.Set(ctx, s.eventKey(eventID), updatedData, s.cfg.Outbox.RetentionPeriod)
			}
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	s.log.Debug("Marked event as published", zap.String("event_id", eventID))
	return nil
}

// MarkFailed marks an event as failed with retry information
func (s *RedisOutboxStorage) MarkFailed(ctx context.Context, eventID string, failureErr error) error {
	// Get current event data
	data, err := s.client.Get(ctx, s.eventKey(eventID)).Result()
	if err != nil {
		return fmt.Errorf("failed to get event data: %w", err)
	}

	var event OutboxEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Update retry information
	event.RetryCount++
	now := time.Now().UTC()
	event.LastError = failureErr.Error()
	event.LastAttemptAt = &now

	// Calculate next retry time with exponential backoff
	backoffSeconds := 1 << event.RetryCount // 2^retryCount seconds
	if backoffSeconds > 3600 {
		backoffSeconds = 3600 // Cap at 1 hour
	}
	nextRetry := now.Add(time.Duration(backoffSeconds) * time.Second)
	event.NextRetryAt = &nextRetry

	// Check if max retries exceeded
	if event.RetryCount >= s.cfg.Outbox.MaxRetries {
		event.Status = "failed"
		// Move to failed queue
		s.client.ZAdd(ctx, s.failedKey(), redis.Z{
			Score:  float64(now.Unix()),
			Member: event.ID,
		})
		// Remove from pending
		s.client.ZRem(ctx, s.pendingKey(), event.ID)
	} else {
		// Update position in pending queue with next retry time
		s.client.ZAdd(ctx, s.pendingKey(), redis.Z{
			Score:  float64(nextRetry.Unix()),
			Member: event.ID,
		})
	}

	// Update event data
	updatedData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal updated event: %w", err)
	}

	if err := s.client.Set(ctx, s.eventKey(event.ID), updatedData, s.cfg.Outbox.RetentionPeriod).Err(); err != nil {
		return fmt.Errorf("failed to update event data: %w", err)
	}

	s.log.Debug("Marked event as failed",
		zap.String("event_id", eventID),
		zap.Int("retry_count", event.RetryCount),
		zap.Time("next_retry", nextRetry))

	return nil
}

// Cleanup removes old processed events
func (s *RedisOutboxStorage) Cleanup(ctx context.Context, before time.Time) error {
	cutoff := float64(before.Unix())

	// Get old event IDs from pending queue
	oldEventIDs, err := s.client.ZRangeByScore(ctx, s.pendingKey(), &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", cutoff),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get old events: %w", err)
	}

	if len(oldEventIDs) == 0 {
		return nil
	}

	// Delete old events
	pipe := s.client.TxPipeline()
	for _, eventID := range oldEventIDs {
		pipe.Del(ctx, s.eventKey(eventID))
		pipe.ZRem(ctx, s.pendingKey(), eventID)
		pipe.ZRem(ctx, s.failedKey(), eventID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	s.log.Info("Cleaned up old outbox events",
		zap.Int("count", len(oldEventIDs)),
		zap.Time("before", before))

	return nil
}

// SQLOutboxStorage implements OutboxStorage using SQL database
type SQLOutboxStorage struct {
	db     *sql.DB
	cfg    *Config
	log    *zap.Logger
	table  string
}

// NewSQLOutboxStorage creates a new SQL-based outbox storage
func NewSQLOutboxStorage(db *sql.DB, cfg *Config, log *zap.Logger) *SQLOutboxStorage {
	table := "outbox_events"
	if cfg.Outbox.TableName != "" {
		table = cfg.Outbox.TableName
	}

	return &SQLOutboxStorage{
		db:    db,
		cfg:   cfg,
		log:   log,
		table: table,
	}
}

// Store saves an outbox event
func (s *SQLOutboxStorage) Store(ctx context.Context, event OutboxEvent) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, event_type, aggregate_id, aggregate_type,
			payload, metadata, created_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, s.table)

	metadataJSON, _ := json.Marshal(event.Metadata)

	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.EventType, event.AggregateID, event.AggregateType,
		event.Payload, metadataJSON, event.CreatedAt, event.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to store outbox event: %w", err)
	}

	return nil
}

// GetPending retrieves pending events up to the specified limit
func (s *SQLOutboxStorage) GetPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := fmt.Sprintf(`
		SELECT id, event_type, aggregate_id, aggregate_type,
		       payload, metadata, created_at, retry_count,
		       last_error, last_attempt_at, next_retry_at
		FROM %s
		WHERE status = 'pending'
		  AND (next_retry_at IS NULL OR next_retry_at <= ?)
		ORDER BY created_at ASC
		LIMIT ?
	`, s.table)

	rows, err := s.db.QueryContext(ctx, query, time.Now().UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var event OutboxEvent
		var metadataJSON []byte
		var lastError sql.NullString
		var lastAttemptAt, nextRetryAt sql.NullTime

		err := rows.Scan(
			&event.ID, &event.EventType, &event.AggregateID, &event.AggregateType,
			&event.Payload, &metadataJSON, &event.CreatedAt, &event.RetryCount,
			&lastError, &lastAttemptAt, &nextRetryAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &event.Metadata)
		}

		if lastError.Valid {
			event.LastError = lastError.String
		}
		if lastAttemptAt.Valid {
			event.LastAttemptAt = &lastAttemptAt.Time
		}
		if nextRetryAt.Valid {
			event.NextRetryAt = &nextRetryAt.Time
		}

		event.Status = "pending"
		events = append(events, event)
	}

	return events, nil
}

// MarkPublished marks an event as published
func (s *SQLOutboxStorage) MarkPublished(ctx context.Context, eventID string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = 'published', published_at = ?
		WHERE id = ?
	`, s.table)

	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	return nil
}

// MarkFailed marks an event as failed with retry information
func (s *SQLOutboxStorage) MarkFailed(ctx context.Context, eventID string, failureErr error) error {
	// First, get current retry count
	var retryCount int
	err := s.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT retry_count FROM %s WHERE id = ?", s.table),
		eventID,
	).Scan(&retryCount)

	if err != nil {
		return fmt.Errorf("failed to get retry count: %w", err)
	}

	retryCount++
	now := time.Now().UTC()

	// Calculate next retry time with exponential backoff
	backoffSeconds := 1 << retryCount
	if backoffSeconds > 3600 {
		backoffSeconds = 3600
	}
	nextRetry := now.Add(time.Duration(backoffSeconds) * time.Second)

	status := "pending"
	if retryCount >= s.cfg.Outbox.MaxRetries {
		status = "failed"
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET status = ?, retry_count = ?, last_error = ?,
		    last_attempt_at = ?, next_retry_at = ?
		WHERE id = ?
	`, s.table)

	_, err = s.db.ExecContext(ctx, query,
		status, retryCount, failureErr.Error(),
		now, nextRetry, eventID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	return nil
}

// Cleanup removes old processed events
func (s *SQLOutboxStorage) Cleanup(ctx context.Context, before time.Time) error {
	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE (status = 'published' OR status = 'failed')
		  AND created_at < ?
	`, s.table)

	result, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		s.log.Info("Cleaned up old outbox events",
			zap.Int64("count", rowsAffected),
			zap.Time("before", before))
	}

	return nil
}