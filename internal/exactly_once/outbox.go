// Copyright 2025 James Ross
package exactly_once

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// OutboxEvent represents an event to be published
type OutboxEvent struct {
	ID         string          `json:"id"`
	QueueName  string          `json:"queue_name"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
	ProcessedAt *time.Time     `json:"processed_at"`
	Status     string          `json:"status"` // pending, processed, failed
	Attempts   int             `json:"attempts"`
	LastError  *string         `json:"last_error"`
}

// OutboxStatus represents the current state of the outbox processor
type OutboxStatus struct {
	PendingCount   int       `json:"pending_count"`
	ProcessedCount int       `json:"processed_count"`
	FailedCount    int       `json:"failed_count"`
	LastProcessed  time.Time `json:"last_processed"`
	IsRunning      bool      `json:"is_running"`
}

// OutboxManager handles transactional outbox pattern
type OutboxManager interface {
	// ExecuteWithOutbox runs business logic with outbox event in transaction
	ExecuteWithOutbox(ctx context.Context, tx *sql.Tx, logic func(*sql.Tx) error, events ...OutboxEvent) error

	// ProcessPending processes pending outbox events
	ProcessPending(ctx context.Context) error

	// Status returns outbox processing status
	Status(ctx context.Context) (*OutboxStatus, error)

	// Start begins background processing
	Start(ctx context.Context) error

	// Stop halts background processing
	Stop() error
}

// Queue interface for job enqueueing
type Queue interface {
	Enqueue(ctx context.Context, queueName string, payload []byte, idempotencyKey string) error
}

// SQLOutboxManager implements OutboxManager with SQL database
type SQLOutboxManager struct {
	db           *sql.DB
	queue        Queue
	dedupManager IdempotencyManager
	interval     time.Duration
	maxRetries   int
	batchSize    int

	mu        sync.RWMutex
	isRunning bool
	stopCh    chan struct{}
	status    OutboxStatus
}

// NewSQLOutboxManager creates a new SQL-backed outbox manager
func NewSQLOutboxManager(db *sql.DB, queue Queue, dedupManager IdempotencyManager) *SQLOutboxManager {
	return &SQLOutboxManager{
		db:           db,
		queue:        queue,
		dedupManager: dedupManager,
		interval:     5 * time.Second,
		maxRetries:   3,
		batchSize:    100,
		stopCh:       make(chan struct{}),
		status:       OutboxStatus{},
	}
}

// ExecuteWithOutbox runs business logic with outbox event in transaction
func (o *SQLOutboxManager) ExecuteWithOutbox(ctx context.Context, tx *sql.Tx, logic func(*sql.Tx) error, events ...OutboxEvent) error {
	// Execute business logic first
	if err := logic(tx); err != nil {
		return fmt.Errorf("business logic failed: %w", err)
	}

	// Insert outbox events in same transaction
	for _, event := range events {
		if event.ID == "" {
			event.ID = fmt.Sprintf("outbox_%d", time.Now().UnixNano())
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = time.Now()
		}
		if event.Status == "" {
			event.Status = "pending"
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, event.ID, event.QueueName, event.Payload, event.CreatedAt, event.Status, 0)

		if err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	return nil
}

// ProcessPending processes pending outbox events
func (o *SQLOutboxManager) ProcessPending(ctx context.Context) error {
	// Query pending events
	rows, err := o.db.QueryContext(ctx, `
		SELECT id, queue_name, payload, attempts
		FROM outbox_events
		WHERE status = 'pending' AND attempts < $1
		ORDER BY created_at ASC
		LIMIT $2
	`, o.maxRetries, o.batchSize)

	if err != nil {
		return fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	processedCount := 0
	failedCount := 0

	for rows.Next() {
		var event OutboxEvent
		var payload []byte
		err := rows.Scan(&event.ID, &event.QueueName, &payload, &event.Attempts)
		if err != nil {
			continue
		}
		event.Payload = json.RawMessage(payload)

		// Use idempotency key to prevent duplicate processing
		idempotencyKey := fmt.Sprintf("outbox_%s", event.ID)
		isDuplicate, err := o.dedupManager.CheckAndReserve(ctx, idempotencyKey, 24*time.Hour)
		if err != nil || isDuplicate {
			// Already processed or error checking, mark as processed
			o.markEventProcessed(ctx, event.ID)
			processedCount++
			continue
		}

		// Enqueue job
		err = o.queue.Enqueue(ctx, event.QueueName, event.Payload, idempotencyKey)
		if err != nil {
			// Update attempts and error
			o.markEventFailed(ctx, event.ID, err.Error())
			failedCount++
		} else {
			// Mark as processed
			o.markEventProcessed(ctx, event.ID)
			processedCount++
		}
	}

	// Update status
	o.mu.Lock()
	o.status.ProcessedCount += processedCount
	o.status.FailedCount += failedCount
	if processedCount > 0 {
		o.status.LastProcessed = time.Now()
	}
	o.mu.Unlock()

	return nil
}

func (o *SQLOutboxManager) markEventProcessed(ctx context.Context, eventID string) {
	now := time.Now()
	_, _ = o.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = 'processed', processed_at = $1
		WHERE id = $2
	`, now, eventID)
}

func (o *SQLOutboxManager) markEventFailed(ctx context.Context, eventID string, errorMsg string) {
	_, _ = o.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET attempts = attempts + 1, last_error = $1,
		    status = CASE
		        WHEN attempts + 1 >= $2 THEN 'failed'
		        ELSE status
		    END
		WHERE id = $3
	`, errorMsg, o.maxRetries, eventID)
}

// Status returns outbox processing status
func (o *SQLOutboxManager) Status(ctx context.Context) (*OutboxStatus, error) {
	o.mu.RLock()
	status := o.status
	o.mu.RUnlock()

	// Query current counts
	var pendingCount int
	err := o.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM outbox_events WHERE status = 'pending'
	`).Scan(&pendingCount)
	if err != nil {
		return nil, err
	}

	status.PendingCount = pendingCount
	status.IsRunning = o.isRunning

	return &status, nil
}

// Start begins background processing
func (o *SQLOutboxManager) Start(ctx context.Context) error {
	o.mu.Lock()
	if o.isRunning {
		o.mu.Unlock()
		return fmt.Errorf("already running")
	}
	o.isRunning = true
	o.mu.Unlock()

	go func() {
		ticker := time.NewTicker(o.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-o.stopCh:
				return
			case <-ticker.C:
				_ = o.ProcessPending(ctx)
			}
		}
	}()

	return nil
}

// Stop halts background processing
func (o *SQLOutboxManager) Stop() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.isRunning {
		return fmt.Errorf("not running")
	}

	close(o.stopCh)
	o.isRunning = false
	return nil
}

// CreateOutboxTable creates the outbox events table
func CreateOutboxTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS outbox_events (
			id VARCHAR(255) PRIMARY KEY,
			queue_name VARCHAR(255) NOT NULL,
			payload TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			processed_at TIMESTAMP,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			attempts INT NOT NULL DEFAULT 0,
			last_error TEXT,
			INDEX idx_status_created (status, created_at)
		)
	`)
	return err
}