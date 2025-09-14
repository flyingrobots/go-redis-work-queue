// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create outbox events table
	_, err = db.Exec(`
		CREATE TABLE outbox_events (
			id VARCHAR(255) PRIMARY KEY,
			event_type VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			aggregate_type VARCHAR(255),
			payload TEXT NOT NULL,
			metadata TEXT,
			created_at TIMESTAMP NOT NULL,
			published_at TIMESTAMP,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			retry_count INT NOT NULL DEFAULT 0,
			last_error TEXT,
			last_attempt_at TIMESTAMP,
			next_retry_at TIMESTAMP
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestRedisOutboxStorage_Store(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewRedisOutboxStorage(client, cfg, logger)

	ctx := context.Background()

	t.Run("stores event successfully", func(t *testing.T) {
		event := OutboxEvent{
			ID:            "test-event-1",
			EventType:     "order.created",
			AggregateID:   "order-123",
			AggregateType: "order",
			Payload:       json.RawMessage(`{"amount": 100}`),
			Metadata:      map[string]interface{}{"user": "test"},
			CreatedAt:     time.Now().UTC(),
			Status:        "pending",
		}

		err := storage.Store(ctx, event)
		require.NoError(t, err)

		// Verify event is in pending queue
		pending, err := storage.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 1)
		assert.Equal(t, event.ID, pending[0].ID)
	})

	t.Run("stores multiple events", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			event := OutboxEvent{
				ID:        fmt.Sprintf("event-%d", i),
				EventType: "test.event",
				Payload:   json.RawMessage(`{}`),
				CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
				Status:    "pending",
			}
			err := storage.Store(ctx, event)
			require.NoError(t, err)
		}

		pending, err := storage.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pending), 5)
	})
}

func TestRedisOutboxStorage_GetPending(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewRedisOutboxStorage(client, cfg, logger)

	ctx := context.Background()

	// Store events with different timestamps
	baseTime := time.Now().UTC()
	for i := 0; i < 5; i++ {
		event := OutboxEvent{
			ID:        fmt.Sprintf("pending-%d", i),
			EventType: "test.pending",
			Payload:   json.RawMessage(fmt.Sprintf(`{"index": %d}`, i)),
			CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
			Status:    "pending",
		}
		err := storage.Store(ctx, event)
		require.NoError(t, err)
	}

	t.Run("respects limit", func(t *testing.T) {
		pending, err := storage.GetPending(ctx, 3)
		require.NoError(t, err)
		assert.Len(t, pending, 3)
		// Should get oldest first
		assert.Equal(t, "pending-0", pending[0].ID)
		assert.Equal(t, "pending-1", pending[1].ID)
		assert.Equal(t, "pending-2", pending[2].ID)
	})

	t.Run("returns empty list when no pending", func(t *testing.T) {
		// Mark all as published
		for i := 0; i < 5; i++ {
			err := storage.MarkPublished(ctx, fmt.Sprintf("pending-%d", i))
			require.NoError(t, err)
		}

		pending, err := storage.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.Empty(t, pending)
	})
}

func TestRedisOutboxStorage_MarkPublished(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewRedisOutboxStorage(client, cfg, logger)

	ctx := context.Background()

	event := OutboxEvent{
		ID:        "publish-test",
		EventType: "test.publish",
		Payload:   json.RawMessage(`{}`),
		CreatedAt: time.Now().UTC(),
		Status:    "pending",
	}

	err := storage.Store(ctx, event)
	require.NoError(t, err)

	// Mark as published
	err = storage.MarkPublished(ctx, event.ID)
	require.NoError(t, err)

	// Should not appear in pending
	pending, err := storage.GetPending(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, pending)
}

func TestRedisOutboxStorage_MarkFailed(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.Outbox.MaxRetries = 3
	logger := zap.NewNop()
	storage := NewRedisOutboxStorage(client, cfg, logger)

	ctx := context.Background()

	event := OutboxEvent{
		ID:        "fail-test",
		EventType: "test.fail",
		Payload:   json.RawMessage(`{}`),
		CreatedAt: time.Now().UTC(),
		Status:    "pending",
	}

	err := storage.Store(ctx, event)
	require.NoError(t, err)

	t.Run("increments retry count", func(t *testing.T) {
		err := storage.MarkFailed(ctx, event.ID, fmt.Errorf("test error"))
		require.NoError(t, err)

		// Event should still be pending but with retry info
		pending, err := storage.GetPending(ctx, 10)
		require.NoError(t, err)
		require.Len(t, pending, 1)
		assert.Equal(t, 1, pending[0].RetryCount)
		assert.Equal(t, "test error", pending[0].LastError)
		assert.NotNil(t, pending[0].LastAttemptAt)
		assert.NotNil(t, pending[0].NextRetryAt)
	})

	t.Run("marks as failed after max retries", func(t *testing.T) {
		// Fail multiple times to exceed max retries
		for i := 0; i < cfg.Outbox.MaxRetries; i++ {
			err := storage.MarkFailed(ctx, event.ID, fmt.Errorf("error %d", i))
			require.NoError(t, err)
		}

		// Should no longer be in pending
		pending, err := storage.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.Empty(t, pending)
	})
}

func TestRedisOutboxStorage_Cleanup(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewRedisOutboxStorage(client, cfg, logger)

	ctx := context.Background()

	// Store old events
	oldTime := time.Now().UTC().Add(-48 * time.Hour)
	for i := 0; i < 3; i++ {
		event := OutboxEvent{
			ID:        fmt.Sprintf("old-%d", i),
			EventType: "test.old",
			Payload:   json.RawMessage(`{}`),
			CreatedAt: oldTime,
			Status:    "pending",
		}
		err := storage.Store(ctx, event)
		require.NoError(t, err)
	}

	// Store recent events
	recentTime := time.Now().UTC()
	for i := 0; i < 2; i++ {
		event := OutboxEvent{
			ID:        fmt.Sprintf("recent-%d", i),
			EventType: "test.recent",
			Payload:   json.RawMessage(`{}`),
			CreatedAt: recentTime,
			Status:    "pending",
		}
		err := storage.Store(ctx, event)
		require.NoError(t, err)
	}

	// Cleanup old events
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	err := storage.Cleanup(ctx, cutoff)
	require.NoError(t, err)

	// Only recent events should remain
	pending, err := storage.GetPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 2)
	for _, event := range pending {
		assert.Contains(t, event.ID, "recent")
	}
}

func TestSQLOutboxStorage_Store(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewSQLOutboxStorage(db, cfg, logger)

	ctx := context.Background()

	event := OutboxEvent{
		ID:            "sql-test-1",
		EventType:     "order.created",
		AggregateID:   "order-456",
		AggregateType: "order",
		Payload:       json.RawMessage(`{"amount": 200}`),
		Metadata:      map[string]interface{}{"source": "api"},
		CreatedAt:     time.Now().UTC(),
		Status:        "pending",
	}

	err := storage.Store(ctx, event)
	require.NoError(t, err)

	// Verify event was stored
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = ?", event.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLOutboxStorage_GetPending(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewSQLOutboxStorage(db, cfg, logger)

	ctx := context.Background()

	// Store test events
	baseTime := time.Now().UTC()
	for i := 0; i < 5; i++ {
		event := OutboxEvent{
			ID:        fmt.Sprintf("sql-pending-%d", i),
			EventType: "test.event",
			Payload:   json.RawMessage(fmt.Sprintf(`{"index": %d}`, i)),
			CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
			Status:    "pending",
		}
		err := storage.Store(ctx, event)
		require.NoError(t, err)
	}

	pending, err := storage.GetPending(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, pending, 3)
	// Should be ordered by created_at
	assert.Equal(t, "sql-pending-0", pending[0].ID)
}

func TestSQLOutboxStorage_MarkPublished(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewSQLOutboxStorage(db, cfg, logger)

	ctx := context.Background()

	event := OutboxEvent{
		ID:        "sql-publish",
		EventType: "test.publish",
		Payload:   json.RawMessage(`{}`),
		CreatedAt: time.Now().UTC(),
		Status:    "pending",
	}

	err := storage.Store(ctx, event)
	require.NoError(t, err)

	err = storage.MarkPublished(ctx, event.ID)
	require.NoError(t, err)

	// Verify status changed
	var status string
	err = db.QueryRow("SELECT status FROM outbox_events WHERE id = ?", event.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "published", status)
}

func TestSQLOutboxStorage_MarkFailed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.Outbox.MaxRetries = 2
	logger := zap.NewNop()
	storage := NewSQLOutboxStorage(db, cfg, logger)

	ctx := context.Background()

	event := OutboxEvent{
		ID:        "sql-fail",
		EventType: "test.fail",
		Payload:   json.RawMessage(`{}`),
		CreatedAt: time.Now().UTC(),
		Status:    "pending",
	}

	err := storage.Store(ctx, event)
	require.NoError(t, err)

	// First failure
	err = storage.MarkFailed(ctx, event.ID, fmt.Errorf("first error"))
	require.NoError(t, err)

	var retryCount int
	var status string
	err = db.QueryRow("SELECT retry_count, status FROM outbox_events WHERE id = ?", event.ID).Scan(&retryCount, &status)
	require.NoError(t, err)
	assert.Equal(t, 1, retryCount)
	assert.Equal(t, "pending", status)

	// Second failure (should exceed max retries)
	err = storage.MarkFailed(ctx, event.ID, fmt.Errorf("second error"))
	require.NoError(t, err)

	err = db.QueryRow("SELECT retry_count, status FROM outbox_events WHERE id = ?", event.ID).Scan(&retryCount, &status)
	require.NoError(t, err)
	assert.Equal(t, 2, retryCount)
	assert.Equal(t, "failed", status)
}

func TestSQLOutboxStorage_Cleanup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := DefaultConfig()
	logger := zap.NewNop()
	storage := NewSQLOutboxStorage(db, cfg, logger)

	ctx := context.Background()

	// Insert old published event
	oldTime := time.Now().UTC().Add(-48 * time.Hour)
	_, err := db.Exec(`
		INSERT INTO outbox_events (id, event_type, aggregate_id, payload, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "old-published", "test", "agg-1", "{}", oldTime, "published")
	require.NoError(t, err)

	// Insert recent event
	recentTime := time.Now().UTC()
	_, err = db.Exec(`
		INSERT INTO outbox_events (id, event_type, aggregate_id, payload, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "recent-pending", "test", "agg-2", "{}", recentTime, "pending")
	require.NoError(t, err)

	// Cleanup old events
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	err = storage.Cleanup(ctx, cutoff)
	require.NoError(t, err)

	// Verify old published event was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = 'old-published'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify recent event remains
	err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = 'recent-pending'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}