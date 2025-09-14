// Copyright 2025 James Ross
package exactly_once

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRedisIdempotencyManager_CheckAndReserve(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	t.Run("first reservation succeeds", func(t *testing.T) {
		isDuplicate, err := manager.CheckAndReserve(ctx, "key1", time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate, "first reservation should not be a duplicate")
	})

	t.Run("duplicate reservation is detected", func(t *testing.T) {
		// First reservation
		isDuplicate, err := manager.CheckAndReserve(ctx, "key2", time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		// Second reservation (duplicate)
		isDuplicate, err = manager.CheckAndReserve(ctx, "key2", time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate, "second reservation should be detected as duplicate")
	})

	t.Run("different keys don't conflict", func(t *testing.T) {
		isDuplicate1, err := manager.CheckAndReserve(ctx, "unique1", time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate1)

		isDuplicate2, err := manager.CheckAndReserve(ctx, "unique2", time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate2)
	})
}

func TestRedisIdempotencyManager_TTLExpiry(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	// Reserve with short TTL
	isDuplicate, err := manager.CheckAndReserve(ctx, "ttl_test", 100*time.Millisecond)
	require.NoError(t, err)
	assert.False(t, isDuplicate)

	// Check it's reserved
	isDuplicate, err = manager.CheckAndReserve(ctx, "ttl_test", time.Hour)
	require.NoError(t, err)
	assert.True(t, isDuplicate)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be able to reserve again
	isDuplicate, err = manager.CheckAndReserve(ctx, "ttl_test", time.Hour)
	require.NoError(t, err)
	assert.False(t, isDuplicate, "key should be available after TTL expiry")
}

func TestRedisIdempotencyManager_Release(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	// Reserve a key
	isDuplicate, err := manager.CheckAndReserve(ctx, "release_test", time.Hour)
	require.NoError(t, err)
	assert.False(t, isDuplicate)

	// Release the key
	err = manager.Release(ctx, "release_test")
	require.NoError(t, err)

	// Should be able to reserve again
	isDuplicate, err = manager.CheckAndReserve(ctx, "release_test", time.Hour)
	require.NoError(t, err)
	assert.False(t, isDuplicate, "key should be available after release")
}

func TestRedisIdempotencyManager_Confirm(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	// Reserve a key
	isDuplicate, err := manager.CheckAndReserve(ctx, "confirm_test", 100*time.Millisecond)
	require.NoError(t, err)
	assert.False(t, isDuplicate)

	// Confirm extends TTL
	err = manager.Confirm(ctx, "confirm_test")
	require.NoError(t, err)

	// Wait original TTL
	time.Sleep(150 * time.Millisecond)

	// Should still be reserved due to confirm extending TTL
	isDuplicate, err = manager.CheckAndReserve(ctx, "confirm_test", time.Hour)
	require.NoError(t, err)
	assert.True(t, isDuplicate, "key should still be reserved after confirm")
}

func TestRedisIdempotencyManager_Stats(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	// Create some test data
	for i := 0; i < 5; i++ {
		manager.CheckAndReserve(ctx, fmt.Sprintf("stats_key_%d", i), time.Hour)
	}

	// Create duplicates
	for i := 0; i < 3; i++ {
		manager.CheckAndReserve(ctx, fmt.Sprintf("stats_key_%d", i), time.Hour)
	}

	// Get stats
	stats, err := manager.Stats(ctx)
	require.NoError(t, err)

	assert.Equal(t, int64(5), stats.Processed, "should have 5 processed")
	assert.Equal(t, int64(3), stats.Duplicates, "should have 3 duplicates")
	assert.Equal(t, int64(5), stats.ActiveKeys, "should have 5 active keys")
	assert.Greater(t, stats.HitRate, 0.0, "hit rate should be greater than 0")
}

func TestRedisIdempotencyManager_ConcurrentAccess(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	const numGoroutines = 100
	const key = "concurrent_test"

	var wg sync.WaitGroup
	successCount := int32(0)
	duplicateCount := int32(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			isDuplicate, err := manager.CheckAndReserve(ctx, key, time.Hour)
			if err != nil {
				return
			}
			if isDuplicate {
				duplicateCount++
			} else {
				successCount++
			}
		}()
	}

	wg.Wait()

	// Only one should succeed, rest should be duplicates
	assert.Equal(t, int32(1), successCount, "exactly one reservation should succeed")
	assert.Equal(t, int32(numGoroutines-1), duplicateCount, "all others should be duplicates")
}

func TestUUIDKeyGenerator(t *testing.T) {
	gen := NewUUIDKeyGenerator("test", "prefix")

	t.Run("generates unique keys", func(t *testing.T) {
		key1 := gen.Generate("payload1")
		key2 := gen.Generate("payload1") // Same payload
		assert.NotEqual(t, key1, key2, "should generate different keys for same payload")
	})

	t.Run("includes namespace and prefix", func(t *testing.T) {
		key := gen.Generate("payload")
		assert.Contains(t, key, "test")
		assert.Contains(t, key, "prefix")
		assert.Contains(t, key, "uuid")
	})

	t.Run("validates keys", func(t *testing.T) {
		err := gen.Validate("valid_key_123")
		assert.NoError(t, err)

		err = gen.Validate("short")
		assert.Error(t, err)
	})
}

func TestContentHashGenerator(t *testing.T) {
	gen := NewContentHashGenerator("test", nil)

	t.Run("generates same key for same content", func(t *testing.T) {
		payload := map[string]string{"user": "123", "action": "buy"}
		key1 := gen.Generate(payload)
		key2 := gen.Generate(payload)
		assert.Equal(t, key1, key2, "should generate same key for identical payload")
	})

	t.Run("generates different keys for different content", func(t *testing.T) {
		key1 := gen.Generate("payload1")
		key2 := gen.Generate("payload2")
		assert.NotEqual(t, key1, key2, "should generate different keys for different payloads")
	})

	t.Run("includes namespace", func(t *testing.T) {
		key := gen.Generate("payload")
		assert.Contains(t, key, "test")
		assert.Contains(t, key, "hash")
	})
}

func TestHybridKeyGenerator(t *testing.T) {
	gen := NewHybridKeyGenerator("test")

	t.Run("generates unique keys for same content", func(t *testing.T) {
		payload := "same_payload"
		key1 := gen.Generate(payload)
		key2 := gen.Generate(payload)

		// Keys should be different due to UUID suffix
		assert.NotEqual(t, key1, key2, "should generate different keys even for same payload")

		// But should share same prefix (content hash)
		assert.Equal(t, key1[:20], key2[:20], "should have same content hash prefix")
	})

	t.Run("validates keys", func(t *testing.T) {
		key := gen.Generate("payload")
		err := gen.Validate(key)
		assert.NoError(t, err)

		err = gen.Validate("short")
		assert.Error(t, err)
	})
}

func TestDedupGuardEdgeCases(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	t.Run("empty key handling", func(t *testing.T) {
		isDuplicate, err := manager.CheckAndReserve(ctx, "", time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		// Even empty keys should be deduplicated
		isDuplicate, err = manager.CheckAndReserve(ctx, "", time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("very long key handling", func(t *testing.T) {
		longKey := ""
		for i := 0; i < 1000; i++ {
			longKey += "a"
		}

		isDuplicate, err := manager.CheckAndReserve(ctx, longKey, time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		isDuplicate, err = manager.CheckAndReserve(ctx, longKey, time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("special characters in key", func(t *testing.T) {
		specialKey := "key:with:colons|pipes|and-dashes_underscores"

		isDuplicate, err := manager.CheckAndReserve(ctx, specialKey, time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		isDuplicate, err = manager.CheckAndReserve(ctx, specialKey, time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("zero TTL uses default", func(t *testing.T) {
		isDuplicate, err := manager.CheckAndReserve(ctx, "zero_ttl", 0)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		// Should still be reserved (using default TTL)
		isDuplicate, err = manager.CheckAndReserve(ctx, "zero_ttl", time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})
}

func TestStatsAccuracy(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewRedisIdempotencyManager(client, "test", time.Hour)

	// Process 10 unique keys
	for i := 0; i < 10; i++ {
		isDuplicate, err := manager.CheckAndReserve(ctx, fmt.Sprintf("key_%d", i), time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	}

	// Create 5 duplicates
	for i := 0; i < 5; i++ {
		isDuplicate, err := manager.CheckAndReserve(ctx, fmt.Sprintf("key_%d", i), time.Hour)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	}

	stats, err := manager.Stats(ctx)
	require.NoError(t, err)

	assert.Equal(t, int64(10), stats.Processed)
	assert.Equal(t, int64(5), stats.Duplicates)

	// Hit rate should be 5/(10+5) = 33.33%
	expectedHitRate := (5.0 / 15.0) * 100
	assert.InDelta(t, expectedHitRate, stats.HitRate, 0.1)
}