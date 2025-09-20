//go:build chaos_harness_tests
// +build chaos_harness_tests

// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFaultInjectorManager(t *testing.T) {
	logger := zap.NewNop()
	config := &InjectorConfig{
		Enabled:    true,
		DefaultTTL: 1 * time.Minute,
		MaxTTL:     5 * time.Minute,
	}

	fim := NewFaultInjectorManager(logger, config)
	defer fim.Stop()

	t.Run("add_and_remove_injector", func(t *testing.T) {
		injector := &FaultInjector{
			ID:          "test-1",
			Type:        InjectorLatency,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 0.5,
			Parameters: map[string]interface{}{
				"latency_ms": 100.0,
			},
		}

		// Add injector
		err := fim.AddInjector(injector)
		require.NoError(t, err)

		// Verify it's active
		active := fim.GetActiveInjectors()
		assert.Len(t, active, 1)
		assert.Equal(t, "test-1", active[0].ID)

		// Remove injector
		err = fim.RemoveInjector("test-1")
		require.NoError(t, err)

		// Verify it's removed
		active = fim.GetActiveInjectors()
		assert.Len(t, active, 0)
	})

	t.Run("validate_injector", func(t *testing.T) {
		// Invalid probability
		injector := &FaultInjector{
			ID:          "invalid-1",
			Type:        InjectorError,
			Scope:       ScopeGlobal,
			Probability: 1.5, // Invalid
		}

		err := fim.AddInjector(injector)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "probability")

		// Missing scope value
		injector = &FaultInjector{
			ID:          "invalid-2",
			Type:        InjectorError,
			Scope:       ScopeWorker,
			ScopeValue:  "", // Missing
			Probability: 0.5,
		}

		err = fim.AddInjector(injector)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scope value")
	})

	t.Run("should_inject", func(t *testing.T) {
		// Add injector with 100% probability
		injector := &FaultInjector{
			ID:          "always",
			Type:        InjectorError,
			Scope:       ScopeQueue,
			ScopeValue:  "test-queue",
			Enabled:     true,
			Probability: 1.0,
		}

		err := fim.AddInjector(injector)
		require.NoError(t, err)

		ctx := context.Background()

		// Should inject for matching scope
		should, inj := fim.ShouldInject(ctx, ScopeQueue, "test-queue", InjectorError)
		assert.True(t, should)
		assert.NotNil(t, inj)
		assert.Equal(t, "always", inj.ID)

		// Should not inject for different scope value
		should, _ = fim.ShouldInject(ctx, ScopeQueue, "other-queue", InjectorError)
		assert.False(t, should)

		// Should not inject for different type
		should, _ = fim.ShouldInject(ctx, ScopeQueue, "test-queue", InjectorLatency)
		assert.False(t, should)
	})

	t.Run("inject_latency", func(t *testing.T) {
		injector := &FaultInjector{
			ID:          "latency-test",
			Type:        InjectorLatency,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 1.0,
			Parameters: map[string]interface{}{
				"latency_ms": 50.0,
			},
		}

		err := fim.AddInjector(injector)
		require.NoError(t, err)

		ctx := context.Background()
		start := time.Now()
		latency := fim.InjectLatency(ctx, ScopeGlobal, "")
		elapsed := time.Since(start)

		assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
		assert.Equal(t, 50*time.Millisecond, latency)
	})

	t.Run("inject_error", func(t *testing.T) {
		injector := &FaultInjector{
			ID:          "error-test",
			Type:        InjectorError,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 1.0,
			Parameters: map[string]interface{}{
				"error_message": "test error",
			},
		}

		err := fim.AddInjector(injector)
		require.NoError(t, err)

		ctx := context.Background()
		err = fim.InjectError(ctx, ScopeGlobal, "")
		assert.Error(t, err)

		injErr, ok := err.(*InjectedError)
		assert.True(t, ok)
		assert.Equal(t, "test error", injErr.Message)
		assert.Equal(t, "error-test", injErr.InjectorID)
	})

	t.Run("inject_partial_fail", func(t *testing.T) {
		injector := &FaultInjector{
			ID:          "partial-test",
			Type:        InjectorPartialFail,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 1.0,
			Parameters: map[string]interface{}{
				"fail_rate": 0.5,
			},
		}

		err := fim.AddInjector(injector)
		require.NoError(t, err)

		ctx := context.Background()
		items := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		processed := fim.InjectPartialFail(ctx, ScopeGlobal, "", items)

		// Should process roughly 50% of items
		assert.Less(t, len(processed), len(items))
		assert.Greater(t, len(processed), 0)
	})

	t.Run("ttl_expiry", func(t *testing.T) {
		injector := &FaultInjector{
			ID:          "ttl-test",
			Type:        InjectorError,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 1.0,
			TTL:         100 * time.Millisecond,
		}

		err := fim.AddInjector(injector)
		require.NoError(t, err)

		// Should be active initially
		active := fim.GetActiveInjectors()
		assert.Len(t, active, 1)

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		// Should not inject after expiry
		ctx := context.Background()
		should, _ := fim.ShouldInject(ctx, ScopeGlobal, "", InjectorError)
		assert.False(t, should)
	})
}

func TestScenarioRunner(t *testing.T) {
	logger := zap.NewNop()
	config := &InjectorConfig{
		Enabled:    true,
		DefaultTTL: 1 * time.Minute,
		MaxTTL:     5 * time.Minute,
	}

	fim := NewFaultInjectorManager(logger, config)
	defer fim.Stop()

	sr := NewScenarioRunner(fim, logger)

	t.Run("validate_scenario", func(t *testing.T) {
		// Invalid scenario - no ID
		scenario := &ChaosScenario{
			Duration: 1 * time.Minute,
		}

		err := sr.RunScenario(context.Background(), scenario)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")

		// Invalid scenario - no stages
		scenario = &ChaosScenario{
			ID:       "test",
			Duration: 1 * time.Minute,
			Stages:   []ScenarioStage{},
		}

		err = sr.RunScenario(context.Background(), scenario)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one stage")
	})

	t.Run("run_simple_scenario", func(t *testing.T) {
		scenario := &ChaosScenario{
			ID:       "simple-test",
			Name:     "Simple Test",
			Duration: 500 * time.Millisecond,
			Stages: []ScenarioStage{
				{
					Name:     "Test Stage",
					Duration: 200 * time.Millisecond,
					Injectors: []FaultInjector{
						{
							ID:          "stage-injector",
							Type:        InjectorLatency,
							Scope:       ScopeGlobal,
							Enabled:     true,
							Probability: 0.5,
							Parameters: map[string]interface{}{
								"latency_ms": 10.0,
							},
						},
					},
				},
			},
			Guardrails: ScenarioGuardrails{
				MaxErrorRate:     0.5,
				AutoAbortOnPanic: true,
			},
		}

		err := sr.RunScenario(context.Background(), scenario)
		require.NoError(t, err)

		// Verify scenario completed
		assert.Equal(t, StatusCompleted, scenario.Status)
		assert.NotNil(t, scenario.StartedAt)
		assert.NotNil(t, scenario.EndedAt)
		assert.NotNil(t, scenario.Metrics)
	})

	t.Run("abort_scenario", func(t *testing.T) {
		scenario := &ChaosScenario{
			ID:       "abort-test",
			Name:     "Abort Test",
			Duration: 5 * time.Second,
			Stages: []ScenarioStage{
				{
					Name:     "Long Stage",
					Duration: 5 * time.Second,
				},
			},
		}

		// Start scenario in background
		go sr.RunScenario(context.Background(), scenario)

		// Wait for it to start
		time.Sleep(100 * time.Millisecond)

		// Verify it's running
		running := sr.GetRunningScenarios()
		assert.Len(t, running, 1)

		// Abort it
		err := sr.AbortScenario("abort-test")
		require.NoError(t, err)

		// Wait for abort to complete
		time.Sleep(100 * time.Millisecond)

		// Verify it's no longer running
		running = sr.GetRunningScenarios()
		assert.Len(t, running, 0)
	})

	t.Run("guardrail_violation", func(t *testing.T) {
		scenario := &ChaosScenario{
			ID:       "guardrail-test",
			Name:     "Guardrail Test",
			Duration: 1 * time.Second,
			Stages: []ScenarioStage{
				{
					Name:     "Violation Stage",
					Duration: 500 * time.Millisecond,
				},
			},
			Guardrails: ScenarioGuardrails{
				MaxErrorRate:     0.01, // Very low threshold
				AutoAbortOnPanic: true,
			},
		}

		// Inject high error rate to trigger guardrail
		scenario.Metrics = &ScenarioMetrics{
			ErrorRate: 0.5, // 50% error rate
		}

		err := sr.RunScenario(context.Background(), scenario)
		// Should abort due to guardrail violation
		assert.Error(t, err)
	})
}

func TestLoadGenerator(t *testing.T) {
	logger := zap.NewNop()
	lg := NewLoadGenerator(logger)

	t.Run("generate_constant_load", func(t *testing.T) {
		config := &LoadConfig{
			RPS:     100,
			Pattern: LoadConstant,
		}

		metrics := &ScenarioMetrics{}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		lg.Start(ctx, config, metrics)
		time.Sleep(100 * time.Millisecond)
		lg.Stop()

		stats := lg.GetStats()
		assert.Greater(t, stats["total_requests"], int64(0))
	})

	t.Run("calculate_rps_patterns", func(t *testing.T) {
		tests := []struct {
			name     string
			pattern  LoadPattern
			elapsed  float64
			expected int
		}{
			{"constant", LoadConstant, 10, 100},
			{"linear_start", LoadLinear, 0, 0},
			{"linear_mid", LoadLinear, 30, 50},
			{"linear_end", LoadLinear, 60, 100},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &LoadConfig{
					RPS:     100,
					Pattern: tt.pattern,
				}

				rps := lg.calculateRPS(config, tt.elapsed)
				assert.Equal(t, tt.expected, rps)
			})
		}
	})
}

func TestChaosHarness(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()

	ch := NewChaosHarness(logger, config)
	defer ch.Stop()

	t.Run("get_status", func(t *testing.T) {
		status := ch.GetStatus()
		assert.True(t, status["enabled"].(bool))
		assert.False(t, status["allow_production"].(bool))
		assert.Equal(t, 0, status["active_injectors"].(int))
		assert.Equal(t, 0, status["running_scenarios"].(int))
	})

	t.Run("inject_when_enabled", func(t *testing.T) {
		ctx := context.Background()

		// Add injector
		injector := &FaultInjector{
			ID:          "harness-test",
			Type:        InjectorError,
			Scope:       ScopeGlobal,
			Enabled:     true,
			Probability: 1.0,
			Parameters: map[string]interface{}{
				"error_message": "harness error",
			},
		}

		ch.injectorManager.AddInjector(injector)

		// Should inject error
		err := ch.InjectError(ctx, ScopeGlobal, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "harness error")
	})

	t.Run("no_inject_when_disabled", func(t *testing.T) {
		// Disable chaos harness
		ch.config.Enabled = false

		ctx := context.Background()

		// Should not inject
		err := ch.InjectError(ctx, ScopeGlobal, "")
		assert.NoError(t, err)

		latency := ch.InjectLatency(ctx, ScopeGlobal, "")
		assert.Equal(t, time.Duration(0), latency)

		// Re-enable for other tests
		ch.config.Enabled = true
	})
}
