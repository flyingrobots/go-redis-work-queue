// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupManagerWithRedis(t *testing.T) (*Manager, *redis.Client, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cfg := DefaultConfig()
	cfg.Outbox.Enabled = true
	cfg.Metrics.Enabled = true
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, rdb, logger)

	cleanup := func() {
		manager.Close()
		rdb.Close()
		s.Close()
	}

	return manager, rdb, cleanup
}

func TestManager_StoreInOutboxDetailed(t *testing.T) {
	manager, _, cleanup := setupManagerWithRedis(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("store basic event", func(t *testing.T) {
		event := OutboxEvent{
			AggregateID: "user-123",
			EventType:   "user.created",
			Payload:     json.RawMessage(`{"user_id": "123", "name": "Test User"}`),
		}

		err := manager.StoreInOutbox(ctx, event)
		assert.NoError(t, err)
	})

	t.Run("store event with all fields", func(t *testing.T) {
		event := OutboxEvent{
			ID:            "event-456",
			AggregateID:   "order-456",
			AggregateType: "order",
			EventType:     "order.completed",
			Payload:       json.RawMessage(`{"order_id": "456", "total": 99.99}`),
			Headers:       map[string]string{"source": "api"},
			Metadata:      map[string]interface{}{"version": 1},
			MaxRetries:    3,
			Status:        "pending",
		}

		err := manager.StoreInOutbox(ctx, event)
		assert.NoError(t, err)
	})
}

func TestManager_PublishOutboxEventsDetailed(t *testing.T) {
	manager, _, cleanup := setupManagerWithRedis(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("publish events", func(t *testing.T) {
		// Store some events first
		events := []OutboxEvent{
			{
				AggregateID: "user-1",
				EventType:   "user.created",
				Payload:     json.RawMessage(`{"user_id": "1"}`),
			},
			{
				AggregateID: "user-2",
				EventType:   "user.created",
				Payload:     json.RawMessage(`{"user_id": "2"}`),
			},
		}

		for _, event := range events {
			err := manager.StoreInOutbox(ctx, event)
			require.NoError(t, err)
		}

		// Publish them
		err := manager.PublishOutboxEvents(ctx)
		assert.NoError(t, err)
	})
}

func TestManager_CleanupOutboxEventsDetailed(t *testing.T) {
	manager, _, cleanup := setupManagerWithRedis(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("cleanup events", func(t *testing.T) {
		err := manager.CleanupOutboxEvents(ctx)
		assert.NoError(t, err)
	})
}

func TestManager_CreateStorageBackends(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("create redis idempotency storage", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Idempotency.Storage.Type = "redis"
		manager := NewManager(cfg, nil, logger)
		defer manager.Close()

		assert.NotNil(t, manager.storage)
	})

	t.Run("create memory idempotency storage", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Idempotency.Storage.Type = "memory"
		manager := NewManager(cfg, nil, logger)
		defer manager.Close()

		assert.NotNil(t, manager.storage)
	})

	t.Run("create database idempotency storage (fallback to redis)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Idempotency.Storage.Type = "database"
		manager := NewManager(cfg, nil, logger)
		defer manager.Close()

		assert.NotNil(t, manager.storage)
	})

	t.Run("create unknown idempotency storage (fallback to redis)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Idempotency.Storage.Type = "unknown"
		manager := NewManager(cfg, nil, logger)
		defer manager.Close()

		assert.NotNil(t, manager.storage)
	})

	t.Run("create redis outbox storage", func(t *testing.T) {
		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		defer rdb.Close()

		cfg := DefaultConfig()
		cfg.Outbox.Enabled = true
		cfg.Outbox.StorageType = "redis"
		manager := NewManager(cfg, rdb, logger)
		defer manager.Close()

		assert.NotNil(t, manager.outbox)
	})

	t.Run("create database outbox storage (fallback to redis)", func(t *testing.T) {
		s, err := miniredis.Run()
		require.NoError(t, err)
		defer s.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		defer rdb.Close()

		cfg := DefaultConfig()
		cfg.Outbox.Enabled = true
		cfg.Outbox.StorageType = "database"
		manager := NewManager(cfg, rdb, logger)
		defer manager.Close()

		assert.NotNil(t, manager.outbox)
	})
}

func TestManager_CalculateNextRetry(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Outbox.RetryBackoff = BackoffConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	defer manager.Close()

	t.Run("first retry", func(t *testing.T) {
		nextRetry := manager.calculateNextRetry(1)
		assert.True(t, nextRetry.After(time.Now()))
	})

	t.Run("exponential backoff", func(t *testing.T) {
		retry1 := manager.calculateNextRetry(1)
		retry2 := manager.calculateNextRetry(2)
		retry3 := manager.calculateNextRetry(3)

		// Each retry should be later than the previous (within some tolerance)
		assert.True(t, retry2.After(retry1.Add(-100*time.Millisecond)))
		assert.True(t, retry3.After(retry2.Add(-100*time.Millisecond)))
	})

	t.Run("max delay cap", func(t *testing.T) {
		// High retry count should hit max delay
		nextRetry := manager.calculateNextRetry(10)
		maxTime := time.Now().Add(cfg.Outbox.RetryBackoff.MaxDelay + 1*time.Second)
		assert.True(t, nextRetry.Before(maxTime))
	})

	t.Run("with jitter", func(t *testing.T) {
		cfg.Outbox.RetryBackoff.Jitter = true
		manager2 := NewManager(cfg, nil, logger)
		defer manager2.Close()

		retry1 := manager2.calculateNextRetry(1)
		retry2 := manager2.calculateNextRetry(1)

		// With jitter, two calls with same retry count might differ
		// (though they might be the same due to randomness)
		assert.True(t, retry1.After(time.Now()))
		assert.True(t, retry2.After(time.Now()))
	})
}

func TestManager_Close(t *testing.T) {
	t.Run("close with metrics", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Metrics.Enabled = true
		logger := zaptest.NewLogger(t)
		manager := NewManager(cfg, nil, logger)

		err := manager.Close()
		assert.NoError(t, err)
	})

	t.Run("close without metrics", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Metrics.Enabled = false
		logger := zaptest.NewLogger(t)
		manager := NewManager(cfg, nil, logger)

		err := manager.Close()
		assert.NoError(t, err)
	})

	t.Run("close with hooks", func(t *testing.T) {
		cfg := DefaultConfig()
		logger := zaptest.NewLogger(t)
		manager := NewManager(cfg, nil, logger)

		// Add a hook with cleanup
		hook := &TestCleanupHook{}
		manager.RegisterHook(hook)

		err := manager.Close()
		assert.NoError(t, err)
		assert.True(t, hook.CleanupCalled)
	})
}

// TestCleanupHook is a test hook that implements cleanup
type TestCleanupHook struct {
	CleanupCalled bool
}

func (h *TestCleanupHook) BeforeProcessing(ctx context.Context, jobID string, key IdempotencyKey) error {
	return nil
}

func (h *TestCleanupHook) AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error {
	return nil
}

func (h *TestCleanupHook) OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error {
	return nil
}

func (h *TestCleanupHook) Cleanup() error {
	h.CleanupCalled = true
	return nil
}

func TestManager_PublishEvent(t *testing.T) {
	manager, _, cleanup := setupManagerWithRedis(t)
	defer cleanup()

	ctx := context.Background()
	event := OutboxEvent{
		ID:          "test-event",
		AggregateID: "test-123",
		EventType:   "test.event",
		Payload:     json.RawMessage(`{"test": true}`),
	}

	// publishEvent is a private method, but we can test it indirectly
	// through PublishOutboxEvents after storing an event
	err := manager.StoreInOutbox(ctx, event)
	require.NoError(t, err)

	err = manager.PublishOutboxEvents(ctx)
	assert.NoError(t, err)
}

func TestHelperFunctions(t *testing.T) {
	t.Run("generateRandomID", func(t *testing.T) {
		id1 := generateRandomID()
		id2 := generateRandomID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2) // Should be unique
		assert.True(t, len(id1) > 10) // Should be reasonable length
	})

	t.Run("randomFloat", func(t *testing.T) {
		f1 := randomFloat()
		f2 := randomFloat()

		assert.True(t, f1 >= 0.0 && f1 <= 1.0)
		assert.True(t, f2 >= 0.0 && f2 <= 1.0)
		// Usually different, but might be same due to randomness
	})
}