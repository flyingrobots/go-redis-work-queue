// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Manager coordinates exactly-once processing patterns
type Manager struct {
	cfg     *Config
	rdb     *redis.Client
	log     *zap.Logger
	storage IdempotencyStorage
	outbox  OutboxStorage
	metrics *MetricsCollector
	hooks   []ProcessingHook
	mu      sync.RWMutex
}

// NewManager creates a new exactly-once patterns manager
func NewManager(cfg *Config, rdb *redis.Client, log *zap.Logger) *Manager {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	m := &Manager{
		cfg:   cfg,
		rdb:   rdb,
		log:   log,
		hooks: make([]ProcessingHook, 0),
	}

	// Initialize storage backend
	if cfg.Idempotency.Enabled {
		m.storage = m.createIdempotencyStorage()
	}

	// Initialize outbox storage if enabled
	if cfg.Outbox.Enabled {
		m.outbox = m.createOutboxStorage()
	}

	// Initialize metrics collector
	if cfg.Metrics.Enabled {
		m.metrics = NewMetricsCollector(cfg.Metrics)
	}

	return m
}

// RegisterHook adds a processing hook
func (m *Manager) RegisterHook(hook ProcessingHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// ProcessWithIdempotency wraps job processing with idempotency checks
func (m *Manager) ProcessWithIdempotency(ctx context.Context, key IdempotencyKey, processor func() (interface{}, error)) (interface{}, error) {
	if !m.cfg.Idempotency.Enabled {
		return processor()
	}

	startTime := time.Now()
	defer func() {
		if m.metrics != nil {
			m.metrics.RecordProcessingLatency(time.Since(startTime), key.QueueName)
		}
	}()

	// Call before processing hooks
	for _, hook := range m.hooks {
		if err := hook.BeforeProcessing(ctx, key.ID, key); err != nil {
			m.log.Warn("Before processing hook failed", zap.Error(err), zap.String("key", key.ID))
		}
	}

	// Check if already processed
	result, err := m.storage.Check(ctx, key)
	if err != nil {
		if m.metrics != nil {
			m.metrics.IncrementStorageErrors(key.QueueName)
		}
		return nil, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	if !result.IsFirstTime {
		// Already processed - call duplicate hooks and return existing result
		for _, hook := range m.hooks {
			if err := hook.OnDuplicate(ctx, key.ID, result.ExistingValue); err != nil {
				m.log.Warn("Duplicate processing hook failed", zap.Error(err), zap.String("key", key.ID))
			}
		}

		if m.metrics != nil {
			m.metrics.IncrementDuplicatesAvoided(key.QueueName)
		}

		m.log.Debug("Duplicate processing detected, returning cached result",
			zap.String("key", key.ID),
			zap.String("queue", key.QueueName))

		return result.ExistingValue, nil
	}

	// First time processing - execute the processor
	processResult, processErr := processor()

	// Store the result for future idempotency checks
	if processErr == nil {
		if err := m.storage.Set(ctx, key, processResult); err != nil {
			// Log the error but don't fail the processing since it succeeded
			m.log.Error("Failed to store idempotency result",
				zap.Error(err),
				zap.String("key", key.ID))

			if m.metrics != nil {
				m.metrics.IncrementStorageErrors(key.QueueName)
			}
		} else {
			if m.metrics != nil {
				m.metrics.IncrementSuccessfulProcessing(key.QueueName)
			}
		}
	}

	// Call after processing hooks
	for _, hook := range m.hooks {
		if err := hook.AfterProcessing(ctx, key.ID, processResult, processErr); err != nil {
			m.log.Warn("After processing hook failed", zap.Error(err), zap.String("key", key.ID))
		}
	}

	return processResult, processErr
}

// GenerateIdempotencyKey creates a unique idempotency key
func (m *Manager) GenerateIdempotencyKey(queueName, tenantID string, customSuffix ...string) IdempotencyKey {
	id := generateRandomID()
	if len(customSuffix) > 0 {
		id = fmt.Sprintf("%s-%s", id, strings.Join(customSuffix, "-"))
	}

	return IdempotencyKey{
		ID:        id,
		QueueName: queueName,
		TenantID:  tenantID,
		CreatedAt: time.Now().UTC(),
		TTL:       m.cfg.Idempotency.DefaultTTL,
	}
}

// StoreInOutbox stores an event in the transactional outbox
func (m *Manager) StoreInOutbox(ctx context.Context, tx interface{}, event OutboxEvent) error {
	if !m.cfg.Outbox.Enabled {
		return ErrOutboxDisabled
	}

	if m.outbox == nil {
		return fmt.Errorf("outbox storage not initialized")
	}

	// Ensure event has required fields
	if event.ID == "" {
		event.ID = generateRandomID()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.MaxRetries == 0 {
		event.MaxRetries = m.cfg.Outbox.MaxRetries
	}

	return m.outbox.Store(ctx, tx, event)
}

// PublishOutboxEvents processes unpublished events from the outbox
func (m *Manager) PublishOutboxEvents(ctx context.Context) error {
	if !m.cfg.Outbox.Enabled {
		return ErrOutboxDisabled
	}

	if m.outbox == nil {
		return fmt.Errorf("outbox storage not initialized")
	}

	events, err := m.outbox.GetUnpublished(ctx, m.cfg.Outbox.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	m.log.Debug("Processing outbox events", zap.Int("count", len(events)))

	published := make([]string, 0, len(events))
	failed := make([]string, 0)

	for _, event := range events {
		if err := m.publishEvent(ctx, event); err != nil {
			m.log.Error("Failed to publish outbox event",
				zap.Error(err),
				zap.String("event_id", event.ID),
				zap.String("event_type", event.EventType))

			failed = append(failed, event.ID)
			continue
		}

		published = append(published, event.ID)
	}

	// Mark successfully published events
	if len(published) > 0 {
		if err := m.outbox.MarkPublished(ctx, published); err != nil {
			m.log.Error("Failed to mark events as published", zap.Error(err))
			return err
		}

		if m.metrics != nil {
			m.metrics.IncrementOutboxPublished(len(published))
		}
	}

	// Mark failed events with retry schedule
	if len(failed) > 0 {
		nextRetryAt := m.calculateNextRetry(1) // TODO: track actual retry count
		if err := m.outbox.MarkFailed(ctx, failed, nextRetryAt); err != nil {
			m.log.Error("Failed to mark events as failed", zap.Error(err))
		}

		if m.metrics != nil {
			m.metrics.IncrementOutboxFailed(len(failed))
		}
	}

	return nil
}

// GetDedupStats returns deduplication statistics for a queue/tenant
func (m *Manager) GetDedupStats(ctx context.Context, queueName, tenantID string) (*DedupStats, error) {
	if !m.cfg.Idempotency.Enabled || m.storage == nil {
		return nil, ErrIdempotencyDisabled
	}

	return m.storage.Stats(ctx, queueName, tenantID)
}

// CleanupExpiredKeys removes expired idempotency keys
func (m *Manager) CleanupExpiredKeys(ctx context.Context) error {
	if !m.cfg.Idempotency.Enabled {
		return ErrIdempotencyDisabled
	}

	// This is a simplified cleanup - in a real implementation,
	// you'd scan for expired keys and remove them
	m.log.Debug("Cleaning up expired idempotency keys")

	// The actual cleanup depends on the storage backend
	// For Redis, expired keys are automatically removed by TTL
	// For database storage, you'd run a DELETE query for expired entries

	return nil
}

// CleanupOutboxEvents removes old processed events from the outbox
func (m *Manager) CleanupOutboxEvents(ctx context.Context) error {
	if !m.cfg.Outbox.Enabled || m.outbox == nil {
		return ErrOutboxDisabled
	}

	cutoffTime := time.Now().UTC().Add(-m.cfg.Outbox.CleanupAfter)
	return m.outbox.Cleanup(ctx, cutoffTime)
}

// Close shuts down the manager and releases resources
func (m *Manager) Close() error {
	// TODO: Close connections, stop background workers, etc.
	return nil
}

// createIdempotencyStorage creates the appropriate idempotency storage backend
func (m *Manager) createIdempotencyStorage() IdempotencyStorage {
	switch m.cfg.Idempotency.Storage.Type {
	case "redis":
		return NewRedisIdempotencyStorage(m.rdb, m.cfg, m.log)
	case "memory":
		return NewMemoryIdempotencyStorage(m.cfg, m.log)
	case "database":
		// TODO: Implement database storage
		m.log.Warn("Database idempotency storage not implemented, falling back to Redis")
		return NewRedisIdempotencyStorage(m.rdb, m.cfg, m.log)
	default:
		m.log.Warn("Unknown storage type, falling back to Redis", zap.String("type", m.cfg.Idempotency.Storage.Type))
		return NewRedisIdempotencyStorage(m.rdb, m.cfg, m.log)
	}
}

// createOutboxStorage creates the appropriate outbox storage backend
func (m *Manager) createOutboxStorage() OutboxStorage {
	// TODO: Implement outbox storage backends
	m.log.Warn("Outbox storage not yet implemented")
	return nil
}

// publishEvent publishes a single outbox event to its configured destinations
func (m *Manager) publishEvent(ctx context.Context, event OutboxEvent) error {
	// TODO: Implement actual publishing based on configured publishers
	// For now, just log the event
	m.log.Debug("Publishing outbox event",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.EventType),
		zap.String("aggregate_id", event.AggregateID))

	return nil
}

// calculateNextRetry calculates the next retry time using exponential backoff
func (m *Manager) calculateNextRetry(retryCount int) time.Time {
	backoff := m.cfg.Outbox.RetryBackoff
	delay := backoff.InitialDelay

	for i := 1; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * backoff.Multiplier)
		if delay > backoff.MaxDelay {
			delay = backoff.MaxDelay
			break
		}
	}

	// Add jitter if configured
	if backoff.Jitter && delay > 0 {
		// Add up to 25% jitter
		jitterAmount := time.Duration(float64(delay) * 0.25 * randomFloat())
		delay += jitterAmount
	}

	return time.Now().UTC().Add(delay)
}

// Helper functions

func generateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func randomFloat() float64 {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	// Convert to float64 between 0 and 1
	var result uint64
	for i := 0; i < 8; i++ {
		result = (result << 8) | uint64(bytes[i])
	}

	return float64(result) / float64(^uint64(0))
}