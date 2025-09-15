// Copyright 2025 James Ross
package backpressure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockStatsProvider implements StatsProvider for testing
type MockStatsProvider struct {
	queues       map[string]*QueueStats
	shouldFail   bool
	failureError error
	callCount    int
}

func NewMockStatsProvider() *MockStatsProvider {
	return &MockStatsProvider{
		queues: make(map[string]*QueueStats),
	}
}

func (m *MockStatsProvider) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	m.callCount++
	if m.shouldFail {
		return nil, m.failureError
	}

	stats, exists := m.queues[queueName]
	if !exists {
		return &QueueStats{
			QueueName:    queueName,
			BacklogCount: 0,
			LastUpdated:  time.Now(),
		}, nil
	}

	return stats, nil
}

func (m *MockStatsProvider) GetAllQueueStats(ctx context.Context) (map[string]*QueueStats, error) {
	m.callCount++
	if m.shouldFail {
		return nil, m.failureError
	}

	result := make(map[string]*QueueStats)
	for name, stats := range m.queues {
		result[name] = stats
	}
	return result, nil
}

func (m *MockStatsProvider) SetQueueStats(queueName string, backlogCount int) {
	m.queues[queueName] = &QueueStats{
		QueueName:    queueName,
		BacklogCount: backlogCount,
		LastUpdated:  time.Now(),
	}
}

func (m *MockStatsProvider) SetFailure(shouldFail bool, err error) {
	m.shouldFail = shouldFail
	m.failureError = err
}

func TestNewController(t *testing.T) {
	config := DefaultConfig()
	statsProvider := NewMockStatsProvider()
	logger := zap.NewNop()

	controller, err := NewController(config, statsProvider, logger)
	require.NoError(t, err)
	assert.NotNil(t, controller)

	// Test with nil stats provider
	_, err = NewController(config, nil, logger)
	assert.Error(t, err)

	// Test with invalid config
	invalidConfig := config
	invalidConfig.Circuit.FailureThreshold = -1
	_, err = NewController(invalidConfig, statsProvider, logger)
	assert.Error(t, err)
}

func TestControllerStartStop(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false // Disable polling for simpler test
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	// Test starting
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Test starting again (should fail)
	err = controller.Start(ctx)
	assert.Error(t, err)

	// Test stopping
	err = controller.Stop()
	require.NoError(t, err)

	// Test starting after stop (should fail)
	err = controller.Start(ctx)
	assert.Error(t, err)
}

func TestSuggestThrottleGreenZone(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set low backlog (green zone)
	statsProvider.SetQueueStats("test-queue", 100)

	decision, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)

	assert.Equal(t, MediumPriority, decision.Priority)
	assert.Equal(t, "test-queue", decision.QueueName)
	assert.Equal(t, time.Duration(0), decision.Delay)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_green", decision.Reason)
	assert.Equal(t, 100, decision.BacklogSize)
}

func TestSuggestThrottleYellowZone(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set medium backlog (yellow zone for medium priority)
	statsProvider.SetQueueStats("test-queue", 1000)

	decision, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)

	assert.Equal(t, MediumPriority, decision.Priority)
	assert.Equal(t, "test-queue", decision.QueueName)
	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.Less(t, decision.Delay, 500*time.Millisecond)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_yellow", decision.Reason)
}

func TestSuggestThrottleRedZone(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set high backlog (red zone)
	statsProvider.SetQueueStats("test-queue", 3000)

	// Test high priority (reduced throttling)
	decision, err := controller.SuggestThrottle(ctx, HighPriority, "test-queue")
	require.NoError(t, err)
	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_high_priority", decision.Reason)

	// Test medium priority (full throttling)
	decision, err = controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)
	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_medium_priority", decision.Reason)

	// Test low priority (heavy throttling)
	decision, err = controller.SuggestThrottle(ctx, LowPriority, "test-queue")
	require.NoError(t, err)
	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_low_priority", decision.Reason)
}

func TestSuggestThrottleShedding(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set extremely high backlog that should cause shedding for low priority
	statsProvider.SetQueueStats("test-queue", 900) // Close to red threshold for low priority

	decision, err := controller.SuggestThrottle(ctx, LowPriority, "test-queue")
	require.NoError(t, err)
	assert.Equal(t, InfiniteDelay, decision.Delay)
	assert.True(t, decision.ShouldShed)
	assert.Equal(t, "backlog_red_shed_low_priority", decision.Reason)
}

func TestRunWithThrottling(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set medium backlog for throttling
	statsProvider.SetQueueStats("test-queue", 1000)

	executed := false
	startTime := time.Now()

	err = controller.Run(ctx, MediumPriority, "test-queue", func() error {
		executed = true
		return nil
	})

	duration := time.Since(startTime)
	require.NoError(t, err)
	assert.True(t, executed)
	assert.Greater(t, duration, 10*time.Millisecond) // Should have been throttled
}

func TestRunWithShedding(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set extremely high backlog for shedding
	statsProvider.SetQueueStats("test-queue", 900) // Should shed low priority

	executed := false
	err = controller.Run(ctx, LowPriority, "test-queue", func() error {
		executed = true
		return nil
	})

	assert.Error(t, err)
	assert.True(t, IsShedError(err))
	assert.False(t, executed)
}

func TestRunWithWorkFailure(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	statsProvider.SetQueueStats("test-queue", 100) // Green zone

	workError := errors.New("work failed")
	err = controller.Run(ctx, MediumPriority, "test-queue", func() error {
		return workError
	})

	assert.Error(t, err)
	assert.Equal(t, workError, err)

	// Circuit breaker should have recorded the failure
	state := controller.GetCircuitState("test-queue")
	assert.Equal(t, Closed, state) // Still closed after one failure
}

func TestProcessBatch(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	statsProvider.SetQueueStats("test-queue", 100) // Green zone

	executed := 0
	jobs := []BatchJob{
		{
			Priority:  HighPriority,
			QueueName: "test-queue",
			Work: func() error {
				executed++
				return nil
			},
		},
		{
			Priority:  MediumPriority,
			QueueName: "test-queue",
			Work: func() error {
				executed++
				return nil
			},
		},
	}

	err = controller.ProcessBatch(ctx, jobs)
	require.NoError(t, err)
	assert.Equal(t, 2, executed)
}

func TestManualOverride(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Set high backlog that would normally cause throttling
	statsProvider.SetQueueStats("test-queue", 3000)

	// Enable manual override
	controller.SetManualOverride(true)

	decision, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)

	assert.Equal(t, time.Duration(0), decision.Delay)
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "manual_override_enabled", decision.Reason)
}

func TestCacheHitBehavior(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	config.Polling.CacheTTL = 100 * time.Millisecond
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	statsProvider.SetQueueStats("test-queue", 1000)

	// First call should hit the stats provider
	initialCallCount := statsProvider.callCount
	decision1, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)
	assert.Greater(t, statsProvider.callCount, initialCallCount)

	// Second call should hit cache
	callCountAfterFirst := statsProvider.callCount
	decision2, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)
	assert.Equal(t, callCountAfterFirst, statsProvider.callCount) // No additional calls

	// Decisions should be identical
	assert.Equal(t, decision1.Delay, decision2.Delay)
	assert.Equal(t, decision1.Reason, decision2.Reason)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third call should hit stats provider again
	_, err = controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)
	assert.Greater(t, statsProvider.callCount, callCountAfterFirst)
}

func TestStatsUnavailableFallback(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	config.Recovery.FallbackMode = true
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Make stats provider fail
	statsProvider.SetFailure(true, errors.New("stats unavailable"))

	decision, err := controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)

	assert.Greater(t, decision.Delay, time.Duration(0))
	assert.False(t, decision.ShouldShed)
	assert.Equal(t, "fallback_conservative", decision.Reason)
}

func TestInvalidInputs(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Test invalid priority
	_, err = controller.SuggestThrottle(ctx, Priority(-1), "test-queue")
	assert.Error(t, err)
	assert.True(t, IsBackpressureError(err))

	// Test invalid queue name
	_, err = controller.SuggestThrottle(ctx, MediumPriority, "")
	assert.Error(t, err)
	assert.True(t, IsBackpressureError(err))

	// Test controller not started
	controller2, _ := NewController(config, statsProvider, zap.NewNop())
	_, err = controller2.SuggestThrottle(ctx, MediumPriority, "test-queue")
	assert.Error(t, err)
	assert.True(t, IsBackpressureError(err))
}

func TestHealthStatus(t *testing.T) {
	config := DefaultConfig()
	config.Polling.Enabled = false
	statsProvider := NewMockStatsProvider()

	controller, err := NewController(config, statsProvider, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)
	defer controller.Stop()

	// Trigger some throttling to populate health data
	statsProvider.SetQueueStats("test-queue", 1000)
	_, err = controller.SuggestThrottle(ctx, MediumPriority, "test-queue")
	require.NoError(t, err)

	health := controller.Health()
	assert.True(t, health["started"].(bool))
	assert.False(t, health["stopped"].(bool))
	assert.False(t, health["manual_override"].(bool))
	assert.False(t, health["emergency_mode"].(bool))
	assert.Contains(t, health, "cache_hit_rate")
	assert.Contains(t, health, "cache_size")
	assert.Contains(t, health, "circuit_states")
}