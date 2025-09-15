// Copyright 2025 James Ross
package patternedloadgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*LoadGenerator, *redis.Client, func()) {
	// Create miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create test config
	config := &GeneratorConfig{
		DefaultGuardrails: Guardrails{
			MaxRate:         100,
			MaxTotal:        1000,
			MaxDuration:     10 * time.Second,
			MaxQueueDepth:   500,
			RateLimitWindow: 100 * time.Millisecond,
		},
		MetricsInterval:  100 * time.Millisecond,
		ProfilesPath:     t.TempDir(),
		EnableCharts:     true,
		ChartUpdateRate:  100 * time.Millisecond,
		MaxHistoryPoints: 100,
	}

	logger := zap.NewNop()
	generator := NewLoadGenerator(config, client, logger)

	cleanup := func() {
		generator.Shutdown()
		client.Close()
		mr.Close()
	}

	return generator, client, cleanup
}

func TestLoadPatterns(t *testing.T) {
	generator, _, cleanup := setupTest(t)
	defer cleanup()

	t.Run("sine pattern calculation", func(t *testing.T) {
		pattern := LoadPattern{
			Type:     PatternSine,
			Duration: 10 * time.Second,
			Parameters: map[string]interface{}{
				"amplitude": 50.0,
				"baseline":  100.0,
				"period":    2.0, // Will be converted to Duration
				"phase":     0.0,
			},
		}

		// Convert period to Duration for the actual params
		params := SineParameters{
			Amplitude: 50,
			Baseline:  100,
			Period:    2 * time.Second,
			Phase:     0,
		}
		pattern.Parameters = map[string]interface{}{
			"amplitude": params.Amplitude,
			"baseline":  params.Baseline,
			"period":    params.Period,
			"phase":     params.Phase,
		}

		// Test at different points in the wave
		rate0 := generator.calculateSineRate(pattern, 0)
		assert.InDelta(t, 100.0, rate0, 1.0) // At t=0, sin(0) = 0

		rate500ms := generator.calculateSineRate(pattern, 500*time.Millisecond)
		assert.InDelta(t, 150.0, rate500ms, 1.0) // At t=0.5s (1/4 period), sin(π/2) = 1

		rate1s := generator.calculateSineRate(pattern, 1*time.Second)
		assert.InDelta(t, 100.0, rate1s, 1.0) // At t=1s (1/2 period), sin(π) = 0

		rate1500ms := generator.calculateSineRate(pattern, 1500*time.Millisecond)
		assert.InDelta(t, 50.0, rate1500ms, 1.0) // At t=1.5s (3/4 period), sin(3π/2) = -1
	})

	t.Run("burst pattern calculation", func(t *testing.T) {
		pattern := LoadPattern{
			Type:     PatternBurst,
			Duration: 10 * time.Second,
			Parameters: map[string]interface{}{
				"burst_rate":     200.0,
				"burst_duration": 2 * time.Second,
				"idle_duration":  3 * time.Second,
				"burst_count":    0,
			},
		}

		// During burst
		rateBurst := generator.calculateBurstRate(pattern, 1*time.Second)
		assert.Equal(t, 200.0, rateBurst)

		// During idle
		rateIdle := generator.calculateBurstRate(pattern, 3*time.Second)
		assert.Equal(t, 0.0, rateIdle)

		// Next burst cycle
		rateNextBurst := generator.calculateBurstRate(pattern, 6*time.Second)
		assert.Equal(t, 200.0, rateNextBurst)
	})

	t.Run("ramp pattern calculation", func(t *testing.T) {
		pattern := LoadPattern{
			Type:     PatternRamp,
			Duration: 10 * time.Second,
			Parameters: map[string]interface{}{
				"start_rate":    10.0,
				"end_rate":      100.0,
				"ramp_duration": 3 * time.Second,
				"hold_duration": 2 * time.Second,
				"ramp_down":     true,
			},
		}

		// Start of ramp
		rateStart := generator.calculateRampRate(pattern, 0)
		assert.Equal(t, 10.0, rateStart)

		// Middle of ramp up
		rateMid := generator.calculateRampRate(pattern, 1500*time.Millisecond)
		assert.InDelta(t, 55.0, rateMid, 1.0) // Halfway between 10 and 100

		// During hold
		rateHold := generator.calculateRampRate(pattern, 4*time.Second)
		assert.Equal(t, 100.0, rateHold)

		// During ramp down
		rateDown := generator.calculateRampRate(pattern, 6500*time.Millisecond)
		assert.InDelta(t, 55.0, rateDown, 1.0) // Halfway down
	})

	t.Run("step pattern calculation", func(t *testing.T) {
		pattern := LoadPattern{
			Type:     PatternStep,
			Duration: 10 * time.Second,
			Parameters: map[string]interface{}{
				"steps": []StepLevel{
					{Rate: 10, Duration: 2 * time.Second},
					{Rate: 50, Duration: 2 * time.Second},
					{Rate: 100, Duration: 2 * time.Second},
				},
				"step_duration": 2 * time.Second,
				"repeat":        false,
			},
		}

		// First step
		rate1 := generator.calculateStepRate(pattern, 1*time.Second)
		assert.Equal(t, 10.0, rate1)

		// Second step
		rate2 := generator.calculateStepRate(pattern, 3*time.Second)
		assert.Equal(t, 50.0, rate2)

		// Third step
		rate3 := generator.calculateStepRate(pattern, 5*time.Second)
		assert.Equal(t, 100.0, rate3)

		// After all steps (no repeat)
		rate4 := generator.calculateStepRate(pattern, 7*time.Second)
		assert.Equal(t, 0.0, rate4)
	})

	t.Run("custom pattern calculation", func(t *testing.T) {
		pattern := LoadPattern{
			Type:     PatternCustom,
			Duration: 10 * time.Second,
			Parameters: map[string]interface{}{
				"points": []DataPoint{
					{Time: 0, Rate: 10},
					{Time: 2 * time.Second, Rate: 50},
					{Time: 4 * time.Second, Rate: 30},
					{Time: 6 * time.Second, Rate: 80},
				},
				"loop": false,
			},
		}

		// At defined points
		rate0 := generator.calculateCustomRate(pattern, 0)
		assert.Equal(t, 10.0, rate0)

		rate2s := generator.calculateCustomRate(pattern, 2*time.Second)
		assert.Equal(t, 50.0, rate2s)

		// Between points (interpolated)
		rate1s := generator.calculateCustomRate(pattern, 1*time.Second)
		assert.InDelta(t, 30.0, rate1s, 1.0) // Halfway between 10 and 50

		rate3s := generator.calculateCustomRate(pattern, 3*time.Second)
		assert.InDelta(t, 40.0, rate3s, 1.0) // Halfway between 50 and 30
	})
}

func TestLoadGeneration(t *testing.T) {
	generator, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("start and stop generation", func(t *testing.T) {
		pattern := &LoadPattern{
			Type:     PatternConstant,
			Duration: 2 * time.Second,
			Parameters: map[string]interface{}{
				"rate": 10.0,
			},
		}

		err := generator.StartPattern(pattern, nil)
		require.NoError(t, err)

		// Check status
		status := generator.GetStatus()
		assert.True(t, status.Running)
		assert.Equal(t, PatternConstant, status.Pattern)

		// Let it run briefly
		time.Sleep(500 * time.Millisecond)

		// Stop
		err = generator.Stop()
		require.NoError(t, err)

		// Check jobs were generated
		status = generator.GetStatus()
		assert.False(t, status.Running)
		assert.Greater(t, status.JobsGenerated, int64(0))

		// Check queue
		length, _ := client.LLen(ctx, "queue:default").Result()
		assert.Greater(t, length, int64(0))
	})

	t.Run("pause and resume", func(t *testing.T) {
		// Clear queue
		client.Del(ctx, "queue:default")

		pattern := &LoadPattern{
			Type:     PatternConstant,
			Duration: 5 * time.Second,
			Parameters: map[string]interface{}{
				"rate": 20.0,
			},
		}

		err := generator.StartPattern(pattern, nil)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		// Pause
		err = generator.Pause()
		require.NoError(t, err)

		// Record jobs at pause
		statusPaused := generator.GetStatus()
		jobsAtPause := statusPaused.JobsGenerated

		// Wait while paused
		time.Sleep(500 * time.Millisecond)

		// Should not generate more jobs
		statusStillPaused := generator.GetStatus()
		assert.Equal(t, jobsAtPause, statusStillPaused.JobsGenerated)

		// Resume
		err = generator.Resume()
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		// Should generate more jobs
		statusResumed := generator.GetStatus()
		assert.Greater(t, statusResumed.JobsGenerated, jobsAtPause)

		generator.Stop()
	})

	t.Run("guardrails enforcement", func(t *testing.T) {
		// Clear previous state
		client.Del(ctx, "queue:test")

		guardrails := &Guardrails{
			MaxTotal:        50,
			MaxRate:         100,
			MaxDuration:     1 * time.Second,
			RateLimitWindow: 100 * time.Millisecond,
		}

		pattern := &LoadPattern{
			Type:     PatternConstant,
			Duration: 10 * time.Second, // Longer than guardrail allows
			Parameters: map[string]interface{}{
				"rate": 100.0,
			},
		}

		err := generator.StartPattern(pattern, guardrails)
		require.NoError(t, err)

		// Wait for guardrail to trigger
		time.Sleep(1500 * time.Millisecond)

		// Should have stopped due to guardrails
		status := generator.GetStatus()
		assert.False(t, status.Running)
		assert.LessOrEqual(t, status.JobsGenerated, int64(50))
	})
}

func TestProfiles(t *testing.T) {
	generator, _, cleanup := setupTest(t)
	defer cleanup()

	t.Run("save and load profile", func(t *testing.T) {
		profile := &LoadProfile{
			Name:        "Test Profile",
			Description: "Test profile for unit tests",
			Patterns: []LoadPattern{
				{
					Type:     PatternSine,
					Duration: 30 * time.Second,
					Parameters: map[string]interface{}{
						"amplitude": 25.0,
						"baseline":  50.0,
						"period":    10 * time.Second,
					},
				},
				{
					Type:     PatternBurst,
					Duration: 20 * time.Second,
					Parameters: map[string]interface{}{
						"burst_rate":     100.0,
						"burst_duration": 5 * time.Second,
						"idle_duration":  5 * time.Second,
					},
				},
			},
			Guardrails: Guardrails{
				MaxRate:  200,
				MaxTotal: 10000,
			},
			QueueName: "test-queue",
			Tags:      []string{"test", "performance"},
		}

		// Save profile
		err := generator.SaveProfile(profile)
		require.NoError(t, err)
		assert.NotEmpty(t, profile.ID)

		// Load profile
		loaded, err := generator.LoadProfile(profile.ID)
		require.NoError(t, err)
		assert.Equal(t, profile.Name, loaded.Name)
		assert.Equal(t, len(profile.Patterns), len(loaded.Patterns))
		assert.Equal(t, profile.QueueName, loaded.QueueName)
	})

	t.Run("list profiles", func(t *testing.T) {
		// Create multiple profiles
		for i := 0; i < 3; i++ {
			profile := &LoadProfile{
				Name: fmt.Sprintf("Profile %d", i),
				Patterns: []LoadPattern{
					{
						Type:     PatternConstant,
						Duration: 10 * time.Second,
						Parameters: map[string]interface{}{
							"rate": float64(10 * (i + 1)),
						},
					},
				},
			}
			generator.SaveProfile(profile)
		}

		// List profiles
		profiles, err := generator.ListProfiles()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(profiles), 3)
	})

	t.Run("delete profile", func(t *testing.T) {
		profile := &LoadProfile{
			Name: "To Delete",
		}
		generator.SaveProfile(profile)

		// Delete
		err := generator.DeleteProfile(profile.ID)
		require.NoError(t, err)

		// Should not be able to load
		_, err = generator.LoadProfile(profile.ID)
		assert.Error(t, err)
	})
}

func TestMetricsCollection(t *testing.T) {
	generator, _, cleanup := setupTest(t)
	defer cleanup()

	t.Run("collect metrics during generation", func(t *testing.T) {
		pattern := &LoadPattern{
			Type:     PatternConstant,
			Duration: 1 * time.Second,
			Parameters: map[string]interface{}{
				"rate": 50.0,
			},
		}

		err := generator.StartPattern(pattern, nil)
		require.NoError(t, err)

		// Let metrics collect
		time.Sleep(500 * time.Millisecond)

		// Get metrics
		metrics := generator.GetMetrics(1 * time.Minute)
		assert.Greater(t, len(metrics), 0)

		// Check metrics values
		for _, m := range metrics {
			assert.InDelta(t, 50.0, m.TargetRate, 5.0)
			assert.GreaterOrEqual(t, m.JobsGenerated, int64(0))
		}

		generator.Stop()
	})

	t.Run("chart data generation", func(t *testing.T) {
		pattern := &LoadPattern{
			Type:     PatternSine,
			Duration: 2 * time.Second,
			Parameters: map[string]interface{}{
				"amplitude": 20.0,
				"baseline":  30.0,
				"period":    1 * time.Second,
			},
		}

		err := generator.StartPattern(pattern, nil)
		require.NoError(t, err)

		// Let it run and collect data
		time.Sleep(1 * time.Second)

		// Get chart data
		chartData := generator.GetChartData()
		assert.NotNil(t, chartData)
		assert.Greater(t, len(chartData.TimePoints), 0)
		assert.Equal(t, len(chartData.TimePoints), len(chartData.TargetRates))
		assert.Equal(t, len(chartData.TimePoints), len(chartData.ActualRates))

		generator.Stop()
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("token bucket rate limiting", func(t *testing.T) {
		limiter := NewRateLimiter(10, 100*time.Millisecond) // 10 per 100ms = 100/s

		// Should allow initial burst
		allowed := 0
		for i := 0; i < 20; i++ {
			if limiter.Allow() {
				allowed++
			}
		}
		assert.LessOrEqual(t, allowed, 10) // Should not exceed capacity

		// Wait for refill
		time.Sleep(100 * time.Millisecond)

		// Should allow more
		allowed = 0
		for i := 0; i < 20; i++ {
			if limiter.Allow() {
				allowed++
			}
		}
		assert.Greater(t, allowed, 0)
	})

	t.Run("rate adjustment", func(t *testing.T) {
		limiter := NewRateLimiter(10, 100*time.Millisecond)

		// Use up tokens
		for i := 0; i < 10; i++ {
			limiter.Allow()
		}

		// Should be exhausted
		assert.False(t, limiter.Allow())

		// Increase rate
		limiter.SetRate(100)

		// Wait for refill at new rate
		time.Sleep(50 * time.Millisecond)

		// Should have more tokens available
		allowed := 0
		for i := 0; i < 10; i++ {
			if limiter.Allow() {
				allowed++
			}
		}
		assert.Greater(t, allowed, 0)
	})
}

func TestEventSystem(t *testing.T) {
	generator, _, cleanup := setupTest(t)
	defer cleanup()

	t.Run("receive events during generation", func(t *testing.T) {
		eventCh := generator.GetEventChannel()
		events := make([]GeneratorEvent, 0)

		// Collect events in background
		done := make(chan bool)
		go func() {
			for {
				select {
				case event := <-eventCh:
					events = append(events, event)
				case <-done:
					return
				}
			}
		}()

		// Generate load
		pattern := &LoadPattern{
			Type:     PatternConstant,
			Duration: 500 * time.Millisecond,
			Parameters: map[string]interface{}{
				"rate": 10.0,
			},
		}

		generator.StartPattern(pattern, nil)
		time.Sleep(600 * time.Millisecond)

		done <- true

		// Should have received events
		assert.Greater(t, len(events), 0)

		// Check for expected events
		hasStarted := false
		hasStopped := false
		hasMetrics := false

		for _, event := range events {
			switch event.Type {
			case EventStarted:
				hasStarted = true
			case EventStopped:
				hasStopped = true
			case EventMetricsUpdate:
				hasMetrics = true
			}
		}

		assert.True(t, hasStarted, "Should have started event")
		assert.True(t, hasStopped, "Should have stopped event")
		assert.True(t, hasMetrics, "Should have metrics events")
	})
}

func TestJobGeneration(t *testing.T) {
	t.Run("simple job generator", func(t *testing.T) {
		generator := &SimpleJobGenerator{
			Template: map[string]interface{}{
				"type":     "test",
				"priority": 5,
			},
		}

		// Generate jobs
		job1, err := generator.GenerateJob()
		require.NoError(t, err)

		job2, err := generator.GenerateJob()
		require.NoError(t, err)

		// Check job structure
		j1 := job1.(map[string]interface{})
		assert.Equal(t, "test", j1["type"])
		assert.Equal(t, 5, j1["priority"])
		assert.Equal(t, "job-1", j1["id"])
		assert.Equal(t, int64(1), j1["sequence"])

		j2 := job2.(map[string]interface{})
		assert.Equal(t, "job-2", j2["id"])
		assert.Equal(t, int64(2), j2["sequence"])
	})
}

func TestIntegration(t *testing.T) {
	generator, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("complete workflow", func(t *testing.T) {
		// Create a profile with multiple patterns
		profile := &LoadProfile{
			Name:        "Integration Test",
			Description: "Complete workflow test",
			Patterns: []LoadPattern{
				{
					Type:     PatternRamp,
					Duration: 500 * time.Millisecond,
					Parameters: map[string]interface{}{
						"start_rate":    5.0,
						"end_rate":      20.0,
						"ramp_duration": 300 * time.Millisecond,
						"hold_duration": 200 * time.Millisecond,
					},
				},
				{
					Type:     PatternBurst,
					Duration: 500 * time.Millisecond,
					Parameters: map[string]interface{}{
						"burst_rate":     30.0,
						"burst_duration": 200 * time.Millisecond,
						"idle_duration":  300 * time.Millisecond,
					},
				},
			},
			Guardrails: Guardrails{
				MaxRate:  50,
				MaxTotal: 500,
			},
			QueueName: "integration-test",
		}

		// Save profile
		err := generator.SaveProfile(profile)
		require.NoError(t, err)

		// Clear queue
		client.Del(ctx, fmt.Sprintf("queue:%s", profile.QueueName))

		// Start generation with profile
		err = generator.Start(profile.ID)
		require.NoError(t, err)

		// Monitor for completion
		timeout := time.After(2 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("Generation did not complete in time")
			case <-ticker.C:
				status := generator.GetStatus()
				if !status.Running {
					// Completed
					assert.Greater(t, status.JobsGenerated, int64(0))
					assert.Less(t, status.JobsGenerated, int64(500))

					// Check queue
					length, _ := client.LLen(ctx, fmt.Sprintf("queue:%s", profile.QueueName)).Result()
					assert.Greater(t, length, int64(0))

					// Check metrics
					metrics := generator.GetMetrics(5 * time.Minute)
					assert.Greater(t, len(metrics), 0)

					return
				}
			}
		}
	})
}

func BenchmarkPatternCalculation(b *testing.B) {
	generator, _, cleanup := setupTest(&testing.T{})
	defer cleanup()

	pattern := LoadPattern{
		Type:     PatternSine,
		Duration: 60 * time.Second,
		Parameters: map[string]interface{}{
			"amplitude": 50.0,
			"baseline":  100.0,
			"period":    10 * time.Second,
			"phase":     0.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elapsed := time.Duration(i%60) * time.Second
		generator.calculateRate(pattern, elapsed)
	}
}

func BenchmarkJobGeneration(b *testing.B) {
	generator := &SimpleJobGenerator{
		Template: map[string]interface{}{
			"type": "benchmark",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateJob()
	}
}

func BenchmarkRateLimiting(b *testing.B) {
	limiter := NewRateLimiter(1000, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}