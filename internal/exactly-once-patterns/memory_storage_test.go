// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestMemoryIdempotencyStorage_Check(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Memory.MaxKeys = 10

	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	// Check non-existent key
	result, err := storage.Check(ctx, key)
	require.NoError(t, err)
	assert.True(t, result.IsFirstTime)
	assert.Nil(t, result.ExistingValue)

	// Set the key
	testValue := map[string]interface{}{"result": "success"}
	err = storage.Set(ctx, key, testValue)
	require.NoError(t, err)

	// Check existing key
	result, err = storage.Check(ctx, key)
	require.NoError(t, err)
	assert.False(t, result.IsFirstTime)
	assert.Equal(t, testValue, result.ExistingValue)
}

func TestMemoryIdempotencyStorage_TTL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.CleanupInterval = 10 * time.Millisecond

	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "ttl-test",
		QueueName: "test-queue",
		CreatedAt: time.Now(),
		TTL:       50 * time.Millisecond, // Very short TTL
	}

	// Set the key
	err := storage.Set(ctx, key, "test-value")
	require.NoError(t, err)

	// Should exist initially
	result, err := storage.Check(ctx, key)
	require.NoError(t, err)
	assert.False(t, result.IsFirstTime)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	result, err = storage.Check(ctx, key)
	require.NoError(t, err)
	assert.True(t, result.IsFirstTime)
}

func TestMemoryIdempotencyStorage_Delete(t *testing.T) {
	cfg := DefaultConfig()
	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "delete-test",
		QueueName: "test-queue",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	// Set the key
	err := storage.Set(ctx, key, "test-value")
	require.NoError(t, err)

	// Verify it exists
	result, err := storage.Check(ctx, key)
	require.NoError(t, err)
	assert.False(t, result.IsFirstTime)

	// Delete the key
	err = storage.Delete(ctx, key)
	require.NoError(t, err)

	// Should not exist now
	result, err = storage.Check(ctx, key)
	require.NoError(t, err)
	assert.True(t, result.IsFirstTime)
}

func TestMemoryIdempotencyStorage_Stats(t *testing.T) {
	cfg := DefaultConfig()
	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()
	queueName := "stats-queue"
	tenantID := "stats-tenant"

	// Add some test data
	for i := 0; i < 3; i++ {
		key := IdempotencyKey{
			ID:        fmt.Sprintf("stats-key-%d", i),
			QueueName: queueName,
			TenantID:  tenantID,
			CreatedAt: time.Now(),
			TTL:       time.Hour,
		}

		err := storage.Set(ctx, key, fmt.Sprintf("value-%d", i))
		require.NoError(t, err)

		// Check to generate stats
		_, err = storage.Check(ctx, key)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := storage.Stats(ctx, queueName, tenantID)
	require.NoError(t, err)

	assert.Equal(t, queueName, stats.QueueName)
	assert.Equal(t, tenantID, stats.TenantID)
	assert.Equal(t, int64(3), stats.TotalKeys)
	assert.GreaterOrEqual(t, stats.TotalRequests, int64(3))
	assert.False(t, stats.LastUpdated.IsZero())
}

func TestMemoryIdempotencyStorage_Eviction(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Memory.MaxKeys = 2
	cfg.Idempotency.Storage.Memory.EvictionPolicy = "fifo"

	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()

	// Add keys up to the limit
	key1 := IdempotencyKey{
		ID:        "evict-key-1",
		QueueName: "test",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}
	key2 := IdempotencyKey{
		ID:        "evict-key-2",
		QueueName: "test",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}
	key3 := IdempotencyKey{
		ID:        "evict-key-3",
		QueueName: "test",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	err := storage.Set(ctx, key1, "value1")
	require.NoError(t, err)

	err = storage.Set(ctx, key2, "value2")
	require.NoError(t, err)

	// Adding third key should trigger eviction of first key
	err = storage.Set(ctx, key3, "value3")
	require.NoError(t, err)

	// key1 should be evicted
	result, err := storage.Check(ctx, key1)
	require.NoError(t, err)
	assert.True(t, result.IsFirstTime)

	// key2 should still exist
	result, err = storage.Check(ctx, key2)
	require.NoError(t, err)
	assert.False(t, result.IsFirstTime)

	// key3 should exist
	result, err = storage.Check(ctx, key3)
	require.NoError(t, err)
	assert.False(t, result.IsFirstTime)
}

func TestMemoryIdempotencyStorage_Concurrent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Memory.MaxKeys = 100

	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()
	numGoroutines := 10
	keysPerGoroutine := 10

	// Concurrent writes
	errCh := make(chan error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < keysPerGoroutine; j++ {
				key := IdempotencyKey{
					ID:        fmt.Sprintf("concurrent-%d-%d", goroutineID, j),
					QueueName: "test",
					CreatedAt: time.Now(),
					TTL:       time.Hour,
				}

				if err := storage.Set(ctx, key, fmt.Sprintf("value-%d-%d", goroutineID, j)); err != nil {
					errCh <- err
					return
				}
			}
			errCh <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-errCh
		require.NoError(t, err)
	}

	// Verify all keys exist
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < keysPerGoroutine; j++ {
			key := IdempotencyKey{
				ID:        fmt.Sprintf("concurrent-%d-%d", i, j),
				QueueName: "test",
				CreatedAt: time.Now(),
				TTL:       time.Hour,
			}

			result, err := storage.Check(ctx, key)
			require.NoError(t, err)
			assert.False(t, result.IsFirstTime)
			assert.Equal(t, fmt.Sprintf("value-%d-%d", i, j), result.ExistingValue)
		}
	}
}

func TestMemoryIdempotencyStorage_Close(t *testing.T) {
	cfg := DefaultConfig()
	log := zaptest.NewLogger(t)
	storage := NewMemoryIdempotencyStorage(cfg, log)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "close-test",
		QueueName: "test",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	// Set a key
	err := storage.Set(ctx, key, "test")
	require.NoError(t, err)

	// Close the storage
	err = storage.Close()
	require.NoError(t, err)

	// Operations after close should fail
	err = storage.Set(ctx, key, "test2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage is stopped")

	_, err = storage.Check(ctx, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage is stopped")
}

// Benchmark tests for memory storage
func BenchmarkMemoryStorage_Set(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Memory.MaxKeys = 100000

	log := zaptest.NewLogger(b)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

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

			if err := storage.Set(ctx, key, "benchmark-value"); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkMemoryStorage_Check(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Memory.MaxKeys = 1000

	log := zaptest.NewLogger(b)
	storage := NewMemoryIdempotencyStorage(cfg, log)
	defer storage.Close()

	ctx := context.Background()

	// Pre-populate with some keys
	for i := 0; i < 100; i++ {
		key := IdempotencyKey{
			ID:        fmt.Sprintf("pre-key-%d", i),
			QueueName: "benchmark",
			CreatedAt: time.Now(),
			TTL:       time.Hour,
		}
		storage.Set(ctx, key, "pre-value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := IdempotencyKey{
				ID:        fmt.Sprintf("pre-key-%d", i%100),
				QueueName: "benchmark",
				CreatedAt: time.Now(),
				TTL:       time.Hour,
			}

			if _, err := storage.Check(ctx, key); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}