// Copyright 2025 James Ross
package exactly_once

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQueue implements Queue interface for testing
type MockQueue struct {
	mu          sync.Mutex
	enqueuedJobs []struct {
		QueueName      string
		Payload        []byte
		IdempotencyKey string
	}
	shouldFail bool
}

func (m *MockQueue) Enqueue(ctx context.Context, queueName string, payload []byte, idempotencyKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock queue error")
	}

	m.enqueuedJobs = append(m.enqueuedJobs, struct {
		QueueName      string
		Payload        []byte
		IdempotencyKey string
	}{
		QueueName:      queueName,
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
	})
	return nil
}

func (m *MockQueue) GetEnqueuedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.enqueuedJobs)
}

// MockIdempotencyManager for testing
type MockIdempotencyManager struct {
	keys map[string]bool
	mu   sync.Mutex
}

func NewMockIdempotencyManager() *MockIdempotencyManager {
	return &MockIdempotencyManager{
		keys: make(map[string]bool),
	}
}

func (m *MockIdempotencyManager) CheckAndReserve(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.keys[key] {
		return true, nil // duplicate
	}
	m.keys[key] = true
	return false, nil
}

func (m *MockIdempotencyManager) Release(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, key)
	return nil
}

func (m *MockIdempotencyManager) Confirm(ctx context.Context, key string) error {
	return nil
}

func (m *MockIdempotencyManager) Stats(ctx context.Context) (*DedupStats, error) {
	return &DedupStats{}, nil
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create outbox table
	_, err = db.Exec(`
		CREATE TABLE outbox_events (
			id VARCHAR(255) PRIMARY KEY,
			queue_name VARCHAR(255) NOT NULL,
			payload TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			processed_at TIMESTAMP,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			attempts INT NOT NULL DEFAULT 0,
			last_error TEXT
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestSQLOutboxManager_ExecuteWithOutbox(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	queue := &MockQueue{}
	dedupManager := NewMockIdempotencyManager()
	manager := NewSQLOutboxManager(db, queue, dedupManager)

	ctx := context.Background()

	t.Run("successful transaction with outbox event", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)

		businessLogicExecuted := false
		businessLogic := func(tx *sql.Tx) error {
			businessLogicExecuted = true
			// Simulate business logic
			return nil
		}

		event := OutboxEvent{
			ID:        "test-event-1",
			QueueName: "test-queue",
			Payload:   json.RawMessage(`{"test": "data"}`),
		}

		err = manager.ExecuteWithOutbox(ctx, tx, businessLogic, event)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		assert.True(t, businessLogicExecuted, "business logic should have been executed")

		// Verify event was inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = ?", event.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "event should be in database")
	})

	t.Run("rollback on business logic failure", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)

		businessLogic := func(tx *sql.Tx) error {
			return fmt.Errorf("business logic error")
		}

		event := OutboxEvent{
			ID:        "test-event-2",
			QueueName: "test-queue",
			Payload:   json.RawMessage(`{"test": "data"}`),
		}

		err = manager.ExecuteWithOutbox(ctx, tx, businessLogic, event)
		assert.Error(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		// Verify event was NOT inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = ?", event.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "event should not be in database after rollback")
	})

	t.Run("multiple events in single transaction", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)

		businessLogic := func(tx *sql.Tx) error {
			return nil
		}

		events := []OutboxEvent{
			{
				ID:        "multi-event-1",
				QueueName: "queue-1",
				Payload:   json.RawMessage(`{"id": 1}`),
			},
			{
				ID:        "multi-event-2",
				QueueName: "queue-2",
				Payload:   json.RawMessage(`{"id": 2}`),
			},
		}

		err = manager.ExecuteWithOutbox(ctx, tx, businessLogic, events...)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify both events were inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id IN (?, ?)",
			"multi-event-1", "multi-event-2").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "both events should be in database")
	})
}

func TestSQLOutboxManager_ProcessPending(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	queue := &MockQueue{}
	dedupManager := NewMockIdempotencyManager()
	manager := NewSQLOutboxManager(db, queue, dedupManager)

	ctx := context.Background()

	t.Run("processes pending events", func(t *testing.T) {
		// Insert pending events
		for i := 0; i < 3; i++ {
			_, err := db.Exec(`
				INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
				VALUES (?, ?, ?, ?, 'pending', 0)
			`, fmt.Sprintf("pending-%d", i), "test-queue", `{"test": "data"}`, time.Now())
			require.NoError(t, err)
		}

		// Process pending events
		err := manager.ProcessPending(ctx)
		require.NoError(t, err)

		// Verify events were enqueued
		assert.Equal(t, 3, queue.GetEnqueuedCount(), "all pending events should be enqueued")

		// Verify events marked as processed
		var processedCount int
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE status = 'processed'").Scan(&processedCount)
		require.NoError(t, err)
		assert.Equal(t, 3, processedCount, "all events should be marked as processed")
	})

	t.Run("handles failed enqueue", func(t *testing.T) {
		// Clear queue
		queue = &MockQueue{shouldFail: true}
		manager.queue = queue

		// Insert a pending event
		_, err := db.Exec(`
			INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
			VALUES ('fail-test', 'test-queue', '{"test": "data"}', ?, 'pending', 0)
		`, time.Now())
		require.NoError(t, err)

		// Process pending events
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)

		// Verify event has increased attempts
		var attempts int
		var lastError sql.NullString
		err = db.QueryRow("SELECT attempts, last_error FROM outbox_events WHERE id = 'fail-test'").
			Scan(&attempts, &lastError)
		require.NoError(t, err)
		assert.Equal(t, 1, attempts, "attempts should be incremented")
		assert.True(t, lastError.Valid, "last_error should be set")
	})

	t.Run("respects max retries", func(t *testing.T) {
		// Insert event with max attempts
		_, err := db.Exec(`
			INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
			VALUES ('max-retry', 'test-queue', '{"test": "data"}', ?, 'pending', ?)
		`, time.Now(), manager.maxRetries)
		require.NoError(t, err)

		// Process pending events
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)

		// Verify event was not processed (exceeded max retries)
		var status string
		err = db.QueryRow("SELECT status FROM outbox_events WHERE id = 'max-retry'").Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "pending", status, "event should remain pending when max retries exceeded")
	})

	t.Run("prevents duplicate processing", func(t *testing.T) {
		// Clear and reset queue
		queue = &MockQueue{}
		manager.queue = queue

		// Insert an event
		_, err := db.Exec(`
			INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
			VALUES ('dedup-test', 'test-queue', '{"test": "data"}', ?, 'pending', 0)
		`, time.Now())
		require.NoError(t, err)

		// Process once
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)
		firstCount := queue.GetEnqueuedCount()

		// Reset status to pending (simulate retry)
		_, err = db.Exec("UPDATE outbox_events SET status = 'pending' WHERE id = 'dedup-test'")
		require.NoError(t, err)

		// Process again
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)
		secondCount := queue.GetEnqueuedCount()

		// Should not enqueue again due to idempotency
		assert.Equal(t, firstCount, secondCount, "should not enqueue duplicate")
	})
}

func TestSQLOutboxManager_Status(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	queue := &MockQueue{}
	dedupManager := NewMockIdempotencyManager()
	manager := NewSQLOutboxManager(db, queue, dedupManager)

	ctx := context.Background()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
		VALUES
			('pending-1', 'queue', '{}', ?, 'pending', 0),
			('pending-2', 'queue', '{}', ?, 'pending', 0),
			('processed-1', 'queue', '{}', ?, 'processed', 1)
	`, time.Now(), time.Now(), time.Now())
	require.NoError(t, err)

	status, err := manager.Status(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, status.PendingCount, "should have 2 pending events")
	assert.False(t, status.IsRunning, "should not be running initially")
}

func TestSQLOutboxManager_StartStop(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	queue := &MockQueue{}
	dedupManager := NewMockIdempotencyManager()
	manager := NewSQLOutboxManager(db, queue, dedupManager)
	manager.interval = 100 * time.Millisecond // Fast interval for testing

	ctx := context.Background()

	t.Run("start and stop processing", func(t *testing.T) {
		err := manager.Start(ctx)
		require.NoError(t, err)

		status, err := manager.Status(ctx)
		require.NoError(t, err)
		assert.True(t, status.IsRunning, "should be running after start")

		// Try to start again (should fail)
		err = manager.Start(ctx)
		assert.Error(t, err, "should error when starting twice")

		// Stop processing
		err = manager.Stop()
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond) // Wait a bit

		status, err = manager.Status(ctx)
		require.NoError(t, err)
		assert.False(t, status.IsRunning, "should not be running after stop")
	})

	t.Run("background processing", func(t *testing.T) {
		// Insert a pending event
		_, err := db.Exec(`
			INSERT INTO outbox_events (id, queue_name, payload, created_at, status, attempts)
			VALUES ('bg-test', 'test-queue', '{"test": "data"}', ?, 'pending', 0)
		`, time.Now())
		require.NoError(t, err)

		// Start background processing
		err = manager.Start(ctx)
		require.NoError(t, err)

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Stop processing
		err = manager.Stop()
		require.NoError(t, err)

		// Verify event was processed
		assert.Greater(t, queue.GetEnqueuedCount(), 0, "event should have been processed in background")
	})
}

func TestOutboxIdempotency(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	queue := &MockQueue{}
	dedupManager := NewMockIdempotencyManager()
	manager := NewSQLOutboxManager(db, queue, dedupManager)

	ctx := context.Background()

	// Simulate a complete flow with idempotency
	t.Run("end-to-end idempotent processing", func(t *testing.T) {
		// Business transaction with outbox
		tx, err := db.Begin()
		require.NoError(t, err)

		businessDataInserted := false
		businessLogic := func(tx *sql.Tx) error {
			// Simulate inserting business data
			businessDataInserted = true
			return nil
		}

		event := OutboxEvent{
			ID:        "idempotent-test",
			QueueName: "payment-queue",
			Payload:   json.RawMessage(`{"amount": 100, "user": "123"}`),
		}

		err = manager.ExecuteWithOutbox(ctx, tx, businessLogic, event)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		assert.True(t, businessDataInserted, "business data should be inserted")

		// Process the outbox event
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)

		initialCount := queue.GetEnqueuedCount()
		assert.Equal(t, 1, initialCount, "should enqueue once")

		// Try to process again (should be idempotent)
		err = manager.ProcessPending(ctx)
		require.NoError(t, err)

		finalCount := queue.GetEnqueuedCount()
		assert.Equal(t, initialCount, finalCount, "should not enqueue again due to idempotency")
	})
}