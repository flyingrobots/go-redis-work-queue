//go:build advanced_rate_limiting_tests && integration
// +build advanced_rate_limiting_tests,integration

// Copyright 2025 James Ross

package ratelimiting

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestIntegrationMultiTenantIsolation tests rate limiting across multiple tenants
func TestIntegrationMultiTenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use real Redis instance
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use separate DB for testing
	})
	defer client.Close()

	// Check Redis connectivity
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	logger := zap.NewNop()
	config := &Config{
		GlobalRatePerSecond:  1000,
		GlobalBurstSize:      2000,
		DefaultRatePerSecond: 50,
		DefaultBurstSize:     100,
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
	ctx := context.Background()

	// Clean up before test
	client.FlushDB(ctx)

	t.Run("tenants_isolated_from_each_other", func(t *testing.T) {
		var wg sync.WaitGroup
		tenantResults := make(map[string]*atomic.Int32)

		// Create 5 tenants
		for i := 0; i < 5; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i)
			tenantResults[tenantID] = &atomic.Int32{}

			wg.Add(1)
			go func(tid string, counter *atomic.Int32) {
				defer wg.Done()

				// Each tenant tries to consume 200 tokens (double their burst)
				for j := 0; j < 200; j++ {
					result, err := rl.Consume(ctx, tid, 1, "normal")
					if err == nil && result.Allowed {
						counter.Add(1)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}(tenantID, tenantResults[tenantID])
		}

		wg.Wait()

		// Each tenant should get roughly their burst size
		for tenantID, counter := range tenantResults {
			allowed := counter.Load()
			t.Logf("Tenant %s: %d requests allowed", tenantID, allowed)
			// Should get at least burst size, but not much more due to refill
			assert.GreaterOrEqual(t, allowed, int32(90)) // At least 90% of burst
			assert.LessOrEqual(t, allowed, int32(150))   // Not more than 150% due to refill
		}
	})

	t.Run("global_limit_enforced", func(t *testing.T) {
		client.FlushDB(ctx)

		var totalAllowed atomic.Int32
		var wg sync.WaitGroup

		// 20 tenants trying to consume simultaneously
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				tenantID := fmt.Sprintf("burst-tenant-%d", id)

				// Each tries to consume 200 tokens
				result, err := rl.Consume(ctx, tenantID, 200, "normal")
				if err == nil && result.Allowed {
					totalAllowed.Add(200)
				}
			}(i)
		}

		wg.Wait()

		// Total should not exceed global burst
		allowed := totalAllowed.Load()
		t.Logf("Total allowed across all tenants: %d", allowed)
		assert.LessOrEqual(t, allowed, int32(config.GlobalBurstSize))
	})
}

// TestIntegrationPriorityUnderLoad tests priority fairness under sustained load
func TestIntegrationPriorityUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2,
	})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available")
	}

	logger := zap.NewNop()
	config := &Config{
		GlobalRatePerSecond:  100,
		GlobalBurstSize:      100,
		DefaultRatePerSecond: 100,
		DefaultBurstSize:     100,
		PriorityWeights: map[string]float64{
			"high":   3.0,
			"normal": 1.5,
			"low":    0.5,
		},
		RefillInterval: 100 * time.Millisecond,
		KeyTTL:         1 * time.Hour,
		DryRun:         false,
	}

	rl := NewRateLimiter(client, logger, config)
	ctx := context.Background()
	client.FlushDB(ctx)

	// Track success by priority
	successByPriority := map[string]*atomic.Int32{
		"high":   &atomic.Int32{},
		"normal": &atomic.Int32{},
		"low":    &atomic.Int32{},
	}

	var wg sync.WaitGroup
	done := make(chan bool)

	// Generate load for each priority
	for _, priority := range []string{"high", "normal", "low"} {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					result, err := rl.Consume(ctx, "load-test", 1, p)
					if err == nil && result.Allowed {
						successByPriority[p].Add(1)
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(priority)
	}

	// Run for 3 seconds
	time.Sleep(3 * time.Second)
	close(done)
	wg.Wait()

	// Check that higher priorities got more throughput
	highCount := successByPriority["high"].Load()
	normalCount := successByPriority["normal"].Load()
	lowCount := successByPriority["low"].Load()

	t.Logf("Priority throughput - High: %d, Normal: %d, Low: %d", highCount, normalCount, lowCount)

	// High should get more than normal
	assert.Greater(t, highCount, normalCount)
	// Normal should get more than low
	assert.Greater(t, normalCount, lowCount)
	// Ratios should roughly match weights
	highToNormalRatio := float64(highCount) / float64(normalCount)
	assert.InDelta(t, 2.0, highToNormalRatio, 1.0) // 3.0/1.5 = 2.0, allow variance
}

// TestIntegrationFairnessWithStarvation tests starvation prevention
func TestIntegrationFairnessWithStarvation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   3,
	})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available")
	}

	logger := zap.NewNop()
	fairnessConfig := &FairnessConfig{
		Weights: map[string]float64{
			"critical": 10.0,
			"low":      0.1,
		},
		MinGuaranteedShare: 0.1, // 10% minimum
		MaxWaitTime:        1 * time.Second,
		EnableAdaptive:     true,
		AdaptiveWindow:     5 * time.Second,
		BurstMultiplier:    1.5,
	}

	pf := NewPriorityFairness(client, logger, fairnessConfig)
	ctx := context.Background()
	client.FlushDB(ctx)

	// Simulate heavy critical load
	var criticalSuccess atomic.Int32
	var lowSuccess atomic.Int32
	var wg sync.WaitGroup
	done := make(chan bool)

	// Critical priority hammering the system
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					decision, err := pf.CheckFairness(ctx, "critical", 10)
					if err == nil && decision.Allowed {
						criticalSuccess.Add(1)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}

	// Low priority trying to get through
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				decision, err := pf.CheckFairness(ctx, "low", 1)
				if err == nil && decision.Allowed {
					lowSuccess.Add(1)
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	// Run for 3 seconds
	time.Sleep(3 * time.Second)
	close(done)
	wg.Wait()

	critical := criticalSuccess.Load()
	low := lowSuccess.Load()

	t.Logf("Starvation test - Critical: %d, Low: %d", critical, low)

	// Low priority should get some requests through (starvation prevention)
	assert.Greater(t, low, int32(0), "Low priority was completely starved")
	// But critical should still get significantly more
	assert.Greater(t, critical, low*5, "Critical didn't get enough priority")
}

// TestIntegrationRateLimiterRecovery tests recovery after overload
func TestIntegrationRateLimiterRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   4,
	})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available")
	}

	logger := zap.NewNop()
	config := &Config{
		GlobalRatePerSecond:  10,
		GlobalBurstSize:      20,
		DefaultRatePerSecond: 10,
		DefaultBurstSize:     20,
		PriorityWeights: map[string]float64{
			"normal": 1.0,
		},
		RefillInterval: 100 * time.Millisecond,
		KeyTTL:         1 * time.Hour,
		DryRun:         false,
	}

	rl := NewRateLimiter(client, logger, config)
	ctx := context.Background()
	client.FlushDB(ctx)

	// Phase 1: Exhaust the bucket
	var deniedCount int
	for i := 0; i < 30; i++ {
		result, err := rl.Consume(ctx, "recovery-test", 1, "normal")
		require.NoError(t, err)
		if !result.Allowed {
			deniedCount++
		}
	}

	assert.Greater(t, deniedCount, 0, "Should have some denied requests after exhausting bucket")

	// Phase 2: Wait for recovery
	time.Sleep(2 * time.Second) // Allow refill at 10/sec

	// Phase 3: Should be able to consume again
	var allowedAfterRecovery int
	for i := 0; i < 20; i++ {
		result, err := rl.Consume(ctx, "recovery-test", 1, "normal")
		require.NoError(t, err)
		if result.Allowed {
			allowedAfterRecovery++
		}
	}

	t.Logf("After recovery: %d requests allowed", allowedAfterRecovery)
	assert.GreaterOrEqual(t, allowedAfterRecovery, 15, "Should allow most requests after recovery")
}

// TestIntegrationDryRunMode tests dry-run functionality
func TestIntegrationDryRunMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   5,
	})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available")
	}

	logger := zap.NewNop()
	config := &Config{
		GlobalRatePerSecond:  1,
		GlobalBurstSize:      1,
		DefaultRatePerSecond: 1,
		DefaultBurstSize:     1,
		PriorityWeights: map[string]float64{
			"normal": 1.0,
		},
		RefillInterval: 1 * time.Second,
		KeyTTL:         1 * time.Hour,
		DryRun:         true, // Enable dry-run
	}

	rl := NewRateLimiter(client, logger, config)
	ctx := context.Background()
	client.FlushDB(ctx)

	// In dry-run, all requests should be allowed
	for i := 0; i < 100; i++ {
		result, err := rl.Consume(ctx, "dry-run-test", 1, "normal")
		require.NoError(t, err)
		assert.True(t, result.Allowed, "Dry-run should allow all requests")

		if i > 0 {
			// After first request, would normally be denied
			assert.False(t, result.DryRunWouldAllow, "Should indicate would be denied in production")
		}
	}

	// Verify no state was actually consumed
	status, err := rl.GetStatus(ctx, "dry-run-test")
	require.NoError(t, err)
	assert.Equal(t, config.DefaultBurstSize, status.Available, "Dry-run should not consume tokens")
}
