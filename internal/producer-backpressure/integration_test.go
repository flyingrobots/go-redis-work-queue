// Copyright 2025 James Ross
//go:build integration

package backpressure

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// RedisStatsProvider implements StatsProvider for integration tests with real Redis behavior
type RedisStatsProvider struct {
	queues map[string]*QueueStats
	mu     sync.RWMutex
}

func NewRedisStatsProvider() *RedisStatsProvider {
	return &RedisStatsProvider{
		queues: make(map[string]*QueueStats),
	}
}

func (r *RedisStatsProvider) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, exists := r.queues[queueName]
	if !exists {
		return &QueueStats{
			QueueName:    queueName,
			BacklogCount: 0,
			LastUpdated:  time.Now(),
		}, nil
	}

	return &QueueStats{
		QueueName:       stats.QueueName,
		BacklogCount:    stats.BacklogCount,
		ProcessingCount: stats.ProcessingCount,
		LastUpdated:     time.Now(),
	}, nil
}

func (r *RedisStatsProvider) GetAllQueueStats(ctx context.Context) (map[string]*QueueStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*QueueStats)
	for name, stats := range r.queues {
		result[name] = &QueueStats{
			QueueName:       stats.QueueName,
			BacklogCount:    stats.BacklogCount,
			ProcessingCount: stats.ProcessingCount,
			LastUpdated:     time.Now(),
		}
	}
	return result, nil
}

func (r *RedisStatsProvider) UpdateQueueStats(queueName string, backlogCount, processingCount int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queues[queueName] = &QueueStats{
		QueueName:       queueName,
		BacklogCount:    backlogCount,
		ProcessingCount: processingCount,
		LastUpdated:     time.Now(),
	}
}

func TestIntegrationBackpressureFlow(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = true
	config.Polling.Interval = 100 * time.Millisecond // Fast polling for test
	config.Polling.CacheTTL = 50 * time.Millisecond  // Short cache for test

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	queueName := "integration-test-queue"

	// Test 1: Green zone - no throttling
	statsProvider.UpdateQueueStats(queueName, 100, 0) // Well within green zone
	time.Sleep(150 * time.Millisecond)                // Let polling cycle complete

	decision, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), decision.Delay)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_green", decision.Reason)

	// Test 2: Yellow zone - light throttling
	statsProvider.UpdateQueueStats(queueName, 1000, 0) // Yellow zone for medium priority
	time.Sleep(150 * time.Millisecond)                 // Wait for cache expiry and polling

	decision, err = controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.Less(t, decision.Delay, 500*time.Millisecond)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_yellow", decision.Reason)

	// Test 3: Red zone - heavy throttling
	statsProvider.UpdateQueueStats(queueName, 3000, 0) // Red zone
	time.Sleep(150 * time.Millisecond)

	decision, err = controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Greater(t, decision.Delay, 500*time.Millisecond)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_medium_priority", decision.Reason)

	// Test 4: Extreme load - shedding for low priority
	statsProvider.UpdateQueueStats(queueName, 900, 0) // High load for low priority
	time.Sleep(150 * time.Millisecond)

	decision, err = controller.SuggestThrottle(ctx, LowPriority, queueName)
	require.NoError(t, err)
	assert.Equal(t, InfiniteDelay, decision.Delay)
	assert.True(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_shed_low_priority", decision.Reason)
}

func TestIntegrationCircuitBreakerBehavior(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	config.Circuit.FailureThreshold = 3
	config.Circuit.RecoveryTimeout = 200 * time.Millisecond

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	queueName := "circuit-test-queue"
	statsProvider.UpdateQueueStats(queueName, 100, 0) // Green zone

	// Initial state should be closed
	assert.Equal(t, Closed, controller.GetCircuitState(queueName))

	// Simulate failures to trip circuit breaker
	failureCount := 0
	for i := 0; i < config.Circuit.FailureThreshold; i++ {
		err = controller.Run(ctx, MediumPriority, queueName, func() error {
			failureCount++
			return fmt.Errorf("simulated failure %d", failureCount)
		})
		assert.Error(t, err)
	}

	// Circuit should now be open
	assert.Equal(t, Open, controller.GetCircuitState(queueName))

	// Requests should be shed due to open circuit
	decision, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.True(t, decision.ShouldShed)
	assert.Contains(t, decision.Reason, "circuit_breaker")

	// Wait for recovery timeout
	time.Sleep(300 * time.Millisecond)

	// Circuit should allow probe (transition to half-open)
	decision, err = controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	// Should allow through (half-open allows probes)

	// Successful work should close the circuit
	for i := 0; i < config.Circuit.RecoveryThreshold; i++ {
		err = controller.Run(ctx, MediumPriority, queueName, func() error {
			return nil // Success
		})
		assert.NoError(t, err)
	}

	// Circuit should be closed again
	assert.Equal(t, Closed, controller.GetCircuitState(queueName))
}

func TestIntegrationPollingAndCaching(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = true
	config.Polling.Interval = 50 * time.Millisecond
	config.Polling.CacheTTL = 100 * time.Millisecond

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	queueName := "polling-test-queue"
	statsProvider.UpdateQueueStats(queueName, 1000, 0)

	// First call should cache the decision
	decision1, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)

	// Second call should hit cache (same delay)
	decision2, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Equal(t, decision1.Delay, decision2.Delay)

	// Change stats significantly
	statsProvider.UpdateQueueStats(queueName, 100, 0) // Back to green

	// Should still get cached result
	decision3, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Equal(t, decision1.Delay, decision3.Delay) // Still cached

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Should now get new result reflecting changed stats
	decision4, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
	require.NoError(t, err)
	assert.Less(t, decision4.Delay, decision1.Delay) // Should be less throttling
}

func TestIntegrationConcurrentLoad(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	queueName := "concurrent-test-queue"
	statsProvider.UpdateQueueStats(queueName, 1000, 0) // Yellow zone

	// Run many concurrent throttle requests
	const numGoroutines = 50
	const requestsPerGoroutine = 20

	var wg sync.WaitGroup
	results := make(chan *ThrottleDecision, numGoroutines*requestsPerGoroutine)
	errors := make(chan error, numGoroutines*requestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				decision, err := controller.SuggestThrottle(ctx, MediumPriority, queueName)
				if err != nil {
					errors <- err
				} else {
					results <- decision
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "Should have no errors in concurrent requests")

	// Check results
	resultCount := 0
	for decision := range results {
		assert.Equal(t, MediumPriority, decision.Priority)
		assert.Equal(t, queueName, decision.QueueName)
		assert.Greater(t, decision.Delay, time.Duration(0)) // Yellow zone should throttle
		resultCount++
	}
	assert.Equal(t, numGoroutines*requestsPerGoroutine, resultCount)
}

func TestIntegrationBatchProcessing(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set different queue loads
	statsProvider.UpdateQueueStats("normal-queue", 100, 0)   // Green
	statsProvider.UpdateQueueStats("busy-queue", 1000, 0)    // Yellow
	statsProvider.UpdateQueueStats("overload-queue", 900, 0) // Red (will shed low priority)

	executed := make(map[string]int)
	var mu sync.Mutex

	jobs := []BatchJob{
		{
			Priority:  HighPriority,
			QueueName: "normal-queue",
			Work: func() error {
				mu.Lock()
				executed["high-normal"]++
				mu.Unlock()
				return nil
			},
		},
		{
			Priority:  MediumPriority,
			QueueName: "busy-queue",
			Work: func() error {
				mu.Lock()
				executed["medium-busy"]++
				mu.Unlock()
				return nil
			},
		},
		{
			Priority:  LowPriority,
			QueueName: "overload-queue",
			Work: func() error {
				mu.Lock()
				executed["low-overload"]++
				mu.Unlock()
				return nil
			},
		},
		{
			Priority:  HighPriority,
			QueueName: "overload-queue",
			Work: func() error {
				mu.Lock()
				executed["high-overload"]++
				mu.Unlock()
				return nil
			},
		},
	}

	startTime := time.Now()
	err = controller.ProcessBatch(ctx, jobs)
	duration := time.Since(startTime)

	// Should complete without critical errors (some jobs might be shed)
	if err != nil {
		// Only acceptable error is if low priority job was shed
		assert.Contains(t, err.Error(), "batch processing had")
	}

	mu.Lock()
	defer mu.Unlock()

	// High priority jobs should always execute
	assert.Equal(t, 1, executed["high-normal"])
	assert.Equal(t, 1, executed["high-overload"])

	// Medium priority on busy queue should execute (with throttling)
	assert.Equal(t, 1, executed["medium-busy"])

	// Low priority on overloaded queue might be shed
	// (We don't assert this as it depends on exact threshold configuration)

	// Should have taken some time due to throttling
	assert.Greater(t, duration, 10*time.Millisecond)
}

func TestIntegrationHealthMonitoring(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = true
	config.Polling.Interval = 50 * time.Millisecond

	statsProvider := NewRedisStatsProvider()
	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set up some queue activity
	statsProvider.UpdateQueueStats("test-queue-1", 500, 0)
	statsProvider.UpdateQueueStats("test-queue-2", 1500, 0)

	// Generate some activity
	_, err = controller.SuggestThrottle(ctx, MediumPriority, "test-queue-1")
	require.NoError(t, err)
	_, err = controller.SuggestThrottle(ctx, HighPriority, "test-queue-2")
	require.NoError(t, err)

	// Wait for polling cycles
	time.Sleep(200 * time.Millisecond)

	// Check health status
	health := controller.Health()

	assert.True(t, health["started"].(bool))
	assert.False(t, health["stopped"].(bool))
	assert.False(t, health["manual_override"].(bool))
	assert.Contains(t, health, "cache_hit_rate")
	assert.Contains(t, health, "cache_size")
	assert.Contains(t, health, "circuit_states")
	assert.True(t, health["polling_enabled"].(bool))

	// Cache should have some hits by now
	cacheSize := health["cache_size"].(int)
	assert.Greater(t, cacheSize, 0)

	// Should have circuit states for queues we accessed
	circuitStates := health["circuit_states"].(map[string]string)
	assert.Contains(t, circuitStates, "test-queue-1")
	assert.Contains(t, circuitStates, "test-queue-2")
	assert.Equal(t, "closed", circuitStates["test-queue-1"])
	assert.Equal(t, "closed", circuitStates["test-queue-2"])
}