// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestManager_ProcessWithIdempotency(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Idempotency.Storage.Memory.MaxKeys = 100
	cfg.Metrics.Enabled = false // Disable metrics to avoid registration conflicts

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key-1",
		QueueName: "test-queue",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	processCount := 0
	processor := func() (interface{}, error) {
		processCount++
		return map[string]interface{}{
			"result": "success",
			"processed_at": time.Now(),
			"count": processCount,
		}, nil
	}

	// First call should execute the processor
	result1, err := manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)
	assert.Equal(t, 1, processCount)

	resultMap1, ok := result1.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "success", resultMap1["result"])

	// Second call with same key should return cached result without executing processor
	result2, err := manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)
	assert.Equal(t, 1, processCount) // Should still be 1, not incremented

	resultMap2, ok := result2.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "success", resultMap2["result"])

	// Results should be equivalent
	assert.Equal(t, resultMap1["count"], resultMap2["count"])
}

func TestManager_ProcessWithIdempotency_DisabledIdempotency(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Enabled = false
	cfg.Metrics.Enabled = false // Disable metrics to avoid registration conflicts

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	processCount := 0
	processor := func() (interface{}, error) {
		processCount++
		return "result", nil
	}

	// First call
	_, err := manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)
	assert.Equal(t, 1, processCount)

	// Second call should execute processor again since idempotency is disabled
	_, err = manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)
	assert.Equal(t, 2, processCount)
}

func TestManager_GenerateIdempotencyKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Metrics.Enabled = false
	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	t.Run("basic key generation", func(t *testing.T) {
		key := manager.GenerateIdempotencyKey("test-queue", "tenant-1")

		assert.NotEmpty(t, key.ID)
		assert.Equal(t, "test-queue", key.QueueName)
		assert.Equal(t, "tenant-1", key.TenantID)
		assert.Equal(t, cfg.Idempotency.DefaultTTL, key.TTL)
		assert.False(t, key.CreatedAt.IsZero())
	})

	t.Run("key generation with custom suffix", func(t *testing.T) {
		key := manager.GenerateIdempotencyKey("test-queue", "", "custom", "suffix")

		assert.Contains(t, key.ID, "custom-suffix")
		assert.Equal(t, "test-queue", key.QueueName)
		assert.Empty(t, key.TenantID)
	})

	t.Run("unique keys", func(t *testing.T) {
		key1 := manager.GenerateIdempotencyKey("test-queue", "")
		key2 := manager.GenerateIdempotencyKey("test-queue", "")

		assert.NotEqual(t, key1.ID, key2.ID)
	})
}

func TestManager_Hooks(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	var hookCalls []string
	hook := &testHook{calls: &hookCalls}
	manager.RegisterHook(hook)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "hook-test",
		QueueName: "test-queue",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	processor := func() (interface{}, error) {
		return "success", nil
	}

	// First call should trigger before and after hooks
	_, err := manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)

	expectedCalls := []string{"before:hook-test", "after:hook-test"}
	assert.Equal(t, expectedCalls, hookCalls)

	// Reset calls
	hookCalls = nil

	// Second call should trigger before and duplicate hooks
	_, err = manager.ProcessWithIdempotency(ctx, key, processor)
	require.NoError(t, err)

	expectedCalls = []string{"before:hook-test", "duplicate:hook-test"}
	assert.Equal(t, expectedCalls, hookCalls)
}

func TestManager_StoreInOutbox(t *testing.T) {
	t.Run("outbox disabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Outbox.Enabled = false

		log := zaptest.NewLogger(t)
		manager := NewManager(cfg, nil, log)

		event := OutboxEvent{
			AggregateID: "test-123",
			EventType:   "test.event",
			Payload:     json.RawMessage(`{"test": true}`),
		}

		err := manager.StoreInOutbox(context.Background(), event)
		assert.Equal(t, ErrOutboxDisabled, err)
	})

	t.Run("outbox enabled but no redis client", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Outbox.Enabled = true
		cfg.Metrics.Enabled = false

		log := zaptest.NewLogger(t)
		// Pass nil Redis client
		manager := NewManager(cfg, nil, log)

		event := OutboxEvent{
			AggregateID: "test-123",
			EventType:   "test.event",
			Payload:     json.RawMessage(`{"test": true}`),
		}

		// This should fail because Redis client is nil but outbox tries to use it
		err := manager.StoreInOutbox(context.Background(), event)
		assert.Error(t, err)
		// The exact error may vary, just ensure it fails
	})
}

func TestManager_PublishOutboxEvents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Outbox.Enabled = false
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	err := manager.PublishOutboxEvents(context.Background())
	assert.Equal(t, ErrOutboxDisabled, err)
}

func TestManager_GetDedupStats(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	ctx := context.Background()

	// Add some test data
	key := IdempotencyKey{
		ID:        "stats-test",
		QueueName: "test-queue",
		TenantID:  "tenant-1",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	err := manager.storage.Set(ctx, key, "test-result")
	require.NoError(t, err)

	// Get stats
	stats, err := manager.GetDedupStats(ctx, "test-queue", "tenant-1")
	require.NoError(t, err)

	assert.Equal(t, "test-queue", stats.QueueName)
	assert.Equal(t, "tenant-1", stats.TenantID)
	assert.GreaterOrEqual(t, stats.TotalKeys, int64(1))
	assert.False(t, stats.LastUpdated.IsZero())
}

func TestManager_CleanupExpiredKeys(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	err := manager.CleanupExpiredKeys(context.Background())
	assert.NoError(t, err)
}

func TestManager_CleanupOutboxEvents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Outbox.Enabled = false
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, log)

	err := manager.CleanupOutboxEvents(context.Background())
	assert.Equal(t, ErrOutboxDisabled, err)
}

// testHook is a test implementation of ProcessingHook
type testHook struct {
	calls *[]string
}

func (h *testHook) BeforeProcessing(ctx context.Context, jobID string, idempotencyKey IdempotencyKey) error {
	*h.calls = append(*h.calls, "before:"+jobID)
	return nil
}

func (h *testHook) AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error {
	*h.calls = append(*h.calls, "after:"+jobID)
	return nil
}

func (h *testHook) OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error {
	*h.calls = append(*h.calls, "duplicate:"+jobID)
	return nil
}

// Benchmark tests
func BenchmarkManager_ProcessWithIdempotency(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Idempotency.Storage.Memory.MaxKeys = 10000
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(b)
	manager := NewManager(cfg, nil, log)

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := IdempotencyKey{
				ID:        fmt.Sprintf("bench-key-%d", i),
				QueueName: "benchmark",
				CreatedAt: time.Now(),
				TTL:       time.Hour,
			}

			_, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
				return "result", nil
			})
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkManager_ProcessWithIdempotency_Duplicate(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Type = "memory"
	cfg.Metrics.Enabled = false

	log := zaptest.NewLogger(b)
	manager := NewManager(cfg, nil, log)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "bench-duplicate-key",
		QueueName: "benchmark",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	// Pre-populate the key
	_, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
		return "result", nil
	})
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
				return "result", nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}