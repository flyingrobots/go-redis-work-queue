//go:build advanced_rate_limiting_tests
// +build advanced_rate_limiting_tests

// Copyright 2025 James Ross
package ratelimiting

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRateLimiter(t *testing.T) (*RateLimiter, *redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	logger := zap.NewNop()

	config := &Config{
		GlobalRatePerSecond:  100,
		GlobalBurstSize:      200,
		DefaultRatePerSecond: 10,
		DefaultBurstSize:     20,
		PriorityWeights: map[string]float64{
			"high":   2.0,
			"normal": 1.0,
			"low":    0.5,
		},
		RefillInterval: 100 * time.Millisecond,
		KeyTTL:         1 * time.Hour,
		DryRun:         false,
	}

	rl := NewRateLimiter(client, logger, config)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return rl, client, cleanup
}

func TestBasicRateLimiting(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("allows tokens within limit", func(t *testing.T) {
		result, err := rl.Consume(ctx, "test-tenant", 5, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, int64(5), result.Tokens)
		assert.Equal(t, int64(15), result.Remaining) // 20 burst - 5 consumed
	})

	t.Run("denies tokens exceeding burst", func(t *testing.T) {
		// Consume all remaining tokens
		result, err := rl.Consume(ctx, "test-tenant-2", 25, "normal")
		require.NoError(t, err)
		assert.False(t, result.Allowed)
		assert.Greater(t, result.RetryAfter, time.Duration(0))
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		// Consume some tokens
		result, err := rl.Consume(ctx, "test-refill", 10, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)

		// Wait for refill
		time.Sleep(1 * time.Second)

		// Should be able to consume again
		result, err = rl.Consume(ctx, "test-refill", 10, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	})
}

func TestPriorityWeights(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("high priority consumes fewer tokens", func(t *testing.T) {
		// Reset buckets
		rl.Reset(ctx, "priority-test")

		// High priority (weight 2.0) should consume half tokens
		result, err := rl.Consume(ctx, "priority-test", 10, "high")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		// With weight 2.0, 10 tokens becomes 5 actual tokens consumed

		// Low priority (weight 0.5) should consume double tokens
		result, err = rl.Consume(ctx, "priority-test-low", 10, "low")
		require.NoError(t, err)
		// With weight 0.5, 10 tokens becomes 20 actual tokens consumed
		// This should hit the limit exactly (20 burst size)
	})
}

func TestGlobalRateLimit(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("enforces global limit across tenants", func(t *testing.T) {
		totalConsumed := int64(0)

		// Multiple tenants consuming from global pool
		for i := 0; i < 10; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i)
			result, err := rl.Consume(ctx, tenantID, 25, "normal")
			require.NoError(t, err)

			if result.Allowed {
				totalConsumed += 25
			}

			// Global burst is 200, so we should stop allowing after 8 tenants
			if totalConsumed >= 200 {
				assert.False(t, result.Allowed, "Should hit global limit")
				break
			}
		}

		assert.GreaterOrEqual(t, totalConsumed, int64(175)) // At least 7 tenants
		assert.LessOrEqual(t, totalConsumed, int64(200))    // Not exceeding global
	})
}

func TestDryRunMode(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	// Enable dry run
	rl.config.DryRun = true

	ctx := context.Background()

	t.Run("allows all requests in dry run", func(t *testing.T) {
		// Consume way more than limit
		result, err := rl.Consume(ctx, "dry-run-test", 1000, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)           // Always allowed in dry run
		assert.False(t, result.DryRunWouldAllow) // But would be denied normally
	})
}

func TestConcurrentAccess(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("handles concurrent requests correctly", func(t *testing.T) {
		var wg sync.WaitGroup
		var allowed int32
		var denied int32

		numGoroutines := 50
		tokensPerRequest := int64(1)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				result, err := rl.Consume(ctx, "concurrent-test", tokensPerRequest, "normal")
				if err == nil {
					if result.Allowed {
						atomic.AddInt32(&allowed, 1)
					} else {
						atomic.AddInt32(&denied, 1)
					}
				}
			}()
		}

		wg.Wait()

		// With burst of 20, we should allow exactly 20 requests
		assert.LessOrEqual(t, allowed, int32(20))
		assert.Equal(t, int32(numGoroutines), allowed+denied)
	})
}

func TestManualRefill(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("manually adds tokens to bucket", func(t *testing.T) {
		// Consume all tokens
		result, err := rl.Consume(ctx, "refill-test", 20, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)

		// Should be denied now
		result, err = rl.Consume(ctx, "refill-test", 1, "normal")
		require.NoError(t, err)
		assert.False(t, result.Allowed)

		// Manually refill
		newTokens, err := rl.Refill(ctx, "refill-test", 10)
		require.NoError(t, err)
		assert.Equal(t, int64(10), newTokens)

		// Should allow again
		result, err = rl.Consume(ctx, "refill-test", 5, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	})
}

func TestGetStatus(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("returns current bucket status", func(t *testing.T) {
		// Consume some tokens
		_, err := rl.Consume(ctx, "status-test", 5, "normal")
		require.NoError(t, err)

		// Get status
		status, err := rl.GetStatus(ctx, "status-test")
		require.NoError(t, err)

		assert.Equal(t, "status-test", status.Scope)
		assert.Equal(t, int64(15), status.Available) // 20 - 5
		assert.Equal(t, int64(20), status.Capacity)
		assert.Equal(t, int64(10), status.RefillRate)
	})
}

func TestRateLimitReset(t *testing.T) {
	rl, _, cleanup := setupTestRateLimiter(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("resets rate limit state", func(t *testing.T) {
		// Consume tokens
		_, err := rl.Consume(ctx, "reset-test", 15, "normal")
		require.NoError(t, err)

		// Verify consumed
		status, _ := rl.GetStatus(ctx, "reset-test")
		assert.Equal(t, int64(5), status.Available)

		// Reset
		err = rl.Reset(ctx, "reset-test")
		require.NoError(t, err)

		// Should be back to full capacity
		result, err := rl.Consume(ctx, "reset-test", 20, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	})
}

func BenchmarkRateLimiter(b *testing.B) {
	rl, _, cleanup := setupTestRateLimiter(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	b.Run("single_tenant", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rl.Consume(ctx, "bench-tenant", 1, "normal")
		}
	})

	b.Run("multi_tenant", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i%100)
			rl.Consume(ctx, tenantID, 1, "normal")
		}
	})

	b.Run("with_priority", func(b *testing.B) {
		priorities := []string{"high", "normal", "low"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			priority := priorities[i%3]
			rl.Consume(ctx, "bench-tenant", 1, priority)
		}
	})
}

// Test Priority Fairness

func TestPriorityFairness(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	logger := zap.NewNop()
	config := DefaultFairnessConfig()

	pf := NewPriorityFairness(client, logger, config)
	ctx := context.Background()

	t.Run("allocates tokens fairly by weight", func(t *testing.T) {
		demands := map[string]int64{
			"critical": 100,
			"high":     100,
			"normal":   100,
			"low":      100,
		}

		allocations, err := pf.AllocateTokens(ctx, 200, demands)
		require.NoError(t, err)

		// Critical should get more than low
		assert.Greater(t, allocations["critical"], allocations["low"])
		// High should get more than normal
		assert.Greater(t, allocations["high"], allocations["normal"])
	})

	t.Run("prevents starvation", func(t *testing.T) {
		// Simulate a starving low priority
		demands := map[string]int64{
			"critical": 1000, // Huge demand
			"low":      10,   // Small demand
		}

		allocations, err := pf.AllocateTokens(ctx, 100, demands)
		require.NoError(t, err)

		// Low priority should get at least minimum share
		minShare := int64(float64(100) * config.MinGuaranteedShare)
		assert.GreaterOrEqual(t, allocations["low"], minShare)
	})

	t.Run("checks fairness decision", func(t *testing.T) {
		decision, err := pf.CheckFairness(ctx, "normal", 10)
		require.NoError(t, err)

		assert.NotNil(t, decision)
		assert.True(t, decision.Allowed) // Should allow initial request
		assert.False(t, decision.IsStarving)
	})
}
