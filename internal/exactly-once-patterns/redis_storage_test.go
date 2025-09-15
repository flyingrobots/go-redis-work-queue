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

func setupRedisStorage(t *testing.T) (*RedisIdempotencyStorage, *redis.Client, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cfg := DefaultConfig()
	logger := zaptest.NewLogger(t)
	storage := NewRedisIdempotencyStorage(rdb, cfg, logger)

	cleanup := func() {
		rdb.Close()
		s.Close()
	}

	return storage, rdb, cleanup
}

func TestRedisIdempotencyStorage_Check(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	t.Run("first time check", func(t *testing.T) {
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.True(t, result.IsFirstTime)
		assert.Nil(t, result.ExistingValue)
		assert.Equal(t, key.ID, result.Key)
	})

	// Set a value first
	testValue := map[string]interface{}{"status": "processed"}
	err := storage.Set(ctx, key, testValue)
	require.NoError(t, err)

	t.Run("duplicate check", func(t *testing.T) {
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.False(t, result.IsFirstTime)
		assert.NotNil(t, result.ExistingValue)
		assert.Equal(t, key.ID, result.Key)
	})
}

func TestRedisIdempotencyStorage_Set(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	testValue := map[string]interface{}{
		"status": "processed",
		"result": "success",
	}

	t.Run("set value successfully", func(t *testing.T) {
		err := storage.Set(ctx, key, testValue)
		assert.NoError(t, err)

		// Verify it was set
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.False(t, result.IsFirstTime)
		assert.NotNil(t, result.ExistingValue)
	})

	t.Run("set with nil value", func(t *testing.T) {
		key2 := key
		key2.ID = "nil-test"
		err := storage.Set(ctx, key2, nil)
		assert.NoError(t, err)
	})
}

func TestRedisIdempotencyStorage_Delete(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	// Set a value first
	testValue := "test-value"
	err := storage.Set(ctx, key, testValue)
	require.NoError(t, err)

	t.Run("delete existing key", func(t *testing.T) {
		err := storage.Delete(ctx, key)
		assert.NoError(t, err)

		// Verify it was deleted
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.True(t, result.IsFirstTime)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		key2 := key
		key2.ID = "non-existent"
		err := storage.Delete(ctx, key2)
		assert.NoError(t, err) // Should not error even if key doesn't exist
	})
}

func TestRedisIdempotencyStorage_Stats(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("stats for empty queue", func(t *testing.T) {
		stats, err := storage.Stats(ctx, "empty-queue", "")
		assert.NoError(t, err)
		assert.Equal(t, "empty-queue", stats.QueueName)
		assert.Equal(t, int64(0), stats.TotalKeys)
	})

	// Add some keys
	for i := 0; i < 5; i++ {
		key := IdempotencyKey{
			ID:        fmt.Sprintf("key-%d", i),
			QueueName: "test-queue",
			TenantID:  "tenant-1",
			CreatedAt: time.Now(),
			TTL:       time.Hour,
		}
		err := storage.Set(ctx, key, fmt.Sprintf("value-%d", i))
		require.NoError(t, err)
	}

	t.Run("stats with data", func(t *testing.T) {
		stats, err := storage.Stats(ctx, "test-queue", "tenant-1")
		assert.NoError(t, err)
		assert.Equal(t, "test-queue", stats.QueueName)
		assert.Equal(t, "tenant-1", stats.TenantID)
		assert.True(t, stats.TotalKeys > 0)
		assert.True(t, stats.LastUpdated.After(time.Time{}))
	})
}

func TestRedisIdempotencyStorage_HashMode(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	cfg := DefaultConfig()
	cfg.Idempotency.Storage.Redis.UseHashes = true
	logger := zaptest.NewLogger(t)
	storage := NewRedisIdempotencyStorage(rdb, cfg, logger)

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "hash-test",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	testValue := "hash-value"

	t.Run("hash mode operations", func(t *testing.T) {
		// Set
		err := storage.Set(ctx, key, testValue)
		assert.NoError(t, err)

		// Check
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.False(t, result.IsFirstTime)
		assert.NotNil(t, result.ExistingValue)

		// Delete
		err = storage.Delete(ctx, key)
		assert.NoError(t, err)

		// Verify deleted
		result, err = storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.True(t, result.IsFirstTime)
	})
}

func TestRedisIdempotencyStorage_BuildKeys(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	key := IdempotencyKey{
		ID:        "test-key",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
	}

	t.Run("build redis key", func(t *testing.T) {
		redisKey := storage.buildRedisKey(key)
		expected := "test-queue:idempotency:test-tenant:test-key"
		assert.Equal(t, expected, redisKey)
	})

	t.Run("build hash key", func(t *testing.T) {
		hashKey := storage.buildHashKey(key)
		expected := "test-queue:idempotency:test-tenant"
		assert.Equal(t, expected, hashKey)
	})

	t.Run("build key pattern", func(t *testing.T) {
		pattern := storage.buildKeyPattern("test-queue", "test-tenant")
		expected := "test-queue:idempotency:test-tenant:*"
		assert.Equal(t, expected, pattern)
	})

	t.Run("build hash key pattern", func(t *testing.T) {
		pattern := storage.buildHashKeyPattern("test-queue", "test-tenant")
		expected := "test-queue:idempotency:test-tenant"
		assert.Equal(t, expected, pattern)
	})
}

func TestRedisIdempotencyStorage_WithoutTenant(t *testing.T) {
	storage, _, cleanup := setupRedisStorage(t)
	defer cleanup()

	ctx := context.Background()
	key := IdempotencyKey{
		ID:        "no-tenant",
		QueueName: "test-queue",
		TenantID:  "", // No tenant
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	testValue := "no-tenant-value"

	t.Run("operations without tenant", func(t *testing.T) {
		// Set
		err := storage.Set(ctx, key, testValue)
		assert.NoError(t, err)

		// Check
		result, err := storage.Check(ctx, key)
		assert.NoError(t, err)
		assert.False(t, result.IsFirstTime)

		// Stats
		stats, err := storage.Stats(ctx, "test-queue", "")
		assert.NoError(t, err)
		assert.Equal(t, "test-queue", stats.QueueName)
		assert.Equal(t, "", stats.TenantID)
	})
}