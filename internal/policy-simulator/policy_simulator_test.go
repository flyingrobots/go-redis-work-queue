// Copyright 2025 James Ross
package policysimulator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicySimulator(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        10,
		RedisPoolSize:     5,
	}

	simulator := NewPolicySimulator(config)

	assert.NotNil(t, simulator)
	assert.Equal(t, config, simulator.config)
	assert.NotNil(t, simulator.trafficPatterns)
	assert.NotNil(t, simulator.policies)
	assert.NotNil(t, simulator.simulations)
	assert.NotNil(t, simulator.changes)
}

func TestDefaultPolicyConfig(t *testing.T) {
	config := DefaultPolicyConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, time.Second, config.InitialBackoff)
	assert.Equal(t, 30*time.Second, config.MaxBackoff)
	assert.Equal(t, "exponential", config.BackoffStrategy)
	assert.Equal(t, 100.0, config.MaxRatePerSecond)
	assert.Equal(t, 10, config.BurstSize)
	assert.Equal(t, 5, config.MaxConcurrency)
	assert.Equal(t, 1000, config.QueueSize)
	assert.Equal(t, 30*time.Second, config.ProcessingTimeout)
	assert.Equal(t, 5*time.Second, config.AckTimeout)
	assert.True(t, config.DLQEnabled)
	assert.Equal(t, 3, config.DLQThreshold)
	assert.Equal(t, "dead-letter", config.DLQQueueName)
}

func TestDefaultTrafficPattern(t *testing.T) {
	pattern := DefaultTrafficPattern()

	assert.Equal(t, "Default Load", pattern.Name)
	assert.Equal(t, TrafficConstant, pattern.Type)
	assert.Equal(t, 50.0, pattern.BaseRate)
	assert.Equal(t, 5*time.Minute, pattern.Duration)
	assert.Empty(t, pattern.Variations)
	assert.Equal(t, 1.0, pattern.Probability)
}

func TestRunSimulation(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)

	req := &SimulationRequest{
		Name:           "Test Simulation",
		Description:    "Testing basic simulation",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	ctx := context.Background()
	result, err := simulator.RunSimulation(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Simulation", result.Name)
	assert.Equal(t, "Testing basic simulation", result.Description)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, StatusCompleted, result.Status)
	assert.NotNil(t, result.Metrics)
	assert.NotEmpty(t, result.Timeline)
}

func TestSimulationWithTrafficSpike(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 60 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        3,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	spikePattern := &TrafficPattern{
		Name:     "Spike Pattern",
		Type:     TrafficSpike,
		BaseRate: 30.0,
		Duration: 60 * time.Second,
		Variations: []TrafficVariation{
			{
				StartTime:   20 * time.Second,
				EndTime:     40 * time.Second,
				Multiplier:  3.0,
				Description: "3x spike for 20 seconds",
			},
		},
		Probability: 1.0,
	}

	req := &SimulationRequest{
		Name:           "Spike Test",
		Description:    "Testing traffic spike handling",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: spikePattern,
	}

	ctx := context.Background()
	result, err := simulator.RunSimulation(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusCompleted, result.Status)

	// Check that we have timeline snapshots
	assert.NotEmpty(t, result.Timeline)
	assert.True(t, len(result.Timeline) >= 50) // Should have snapshots for most time steps

	// Verify metrics make sense
	assert.NotNil(t, result.Metrics)
	assert.True(t, result.Metrics.MessagesProcessed > 0)
	assert.True(t, result.Metrics.ProcessingRate > 0)
	assert.True(t, result.Metrics.MaxQueueDepth >= 0)
}

func TestSimulationWithHighRetryPolicy(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        2,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	highRetryPolicy := DefaultPolicyConfig()
	highRetryPolicy.MaxRetries = 10
	highRetryPolicy.InitialBackoff = 100 * time.Millisecond
	highRetryPolicy.MaxBackoff = 5 * time.Second

	req := &SimulationRequest{
		Name:           "High Retry Test",
		Description:    "Testing high retry configuration",
		Policies:       highRetryPolicy,
		TrafficPattern: DefaultTrafficPattern(),
	}

	ctx := context.Background()
	result, err := simulator.RunSimulation(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusCompleted, result.Status)
	assert.NotNil(t, result.Metrics)
}

func TestGetSimulationCore(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 10 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        3,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	req := &SimulationRequest{
		Name:           "Get Test",
		Description:    "Testing get simulation",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	ctx := context.Background()
	result, err := simulator.RunSimulation(ctx, req)
	require.NoError(t, err)

	// Test getting the simulation
	retrieved, err := simulator.GetSimulation(result.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, result.ID, retrieved.ID)
	assert.Equal(t, result.Name, retrieved.Name)

	// Test getting non-existent simulation
	_, err = simulator.GetSimulation("non-existent")
	assert.Error(t, err)
}

func TestListSimulationsCore(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        2,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	// Initially empty
	simulations := simulator.ListSimulations()
	assert.Empty(t, simulations)

	// Run a few simulations
	req1 := &SimulationRequest{
		Name:           "Sim 1",
		Description:    "First simulation",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	req2 := &SimulationRequest{
		Name:           "Sim 2",
		Description:    "Second simulation",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	ctx := context.Background()
	_, err := simulator.RunSimulation(ctx, req1)
	require.NoError(t, err)

	_, err = simulator.RunSimulation(ctx, req2)
	require.NoError(t, err)

	// Should have 2 simulations now
	simulations = simulator.ListSimulations()
	assert.Len(t, simulations, 2)
}

func TestCreatePolicyChangeCore(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        2,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	changes := map[string]interface{}{
		"max_retries":        5,
		"max_rate_per_second": 200.0,
	}

	change, err := simulator.CreatePolicyChange("Increase retries and rate", changes, "test-user")

	assert.NoError(t, err)
	assert.NotNil(t, change)
	assert.NotEmpty(t, change.ID)
	assert.Equal(t, "Increase retries and rate", change.Description)
	assert.Equal(t, changes, change.Changes)
	assert.Equal(t, "test-user", change.AppliedBy)
	assert.Equal(t, ChangeStatusProposed, change.Status)
	assert.Len(t, change.AuditLog, 1)
}

func TestApplyPolicyChangeCore(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        2,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	changes := map[string]interface{}{
		"max_retries": 7,
	}

	change, err := simulator.CreatePolicyChange("Test change", changes, "test-user")
	require.NoError(t, err)

	err = simulator.ApplyPolicyChange(change.ID, "admin-user")
	assert.NoError(t, err)

	// Verify change was applied
	retrieved := simulator.changes[change.ID]
	assert.Equal(t, ChangeStatusApplied, retrieved.Status)
	assert.NotNil(t, retrieved.AppliedAt)
	assert.Len(t, retrieved.AuditLog, 2) // Create + Apply
}

func TestRollbackPolicyChangeCore(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        2,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	changes := map[string]interface{}{
		"max_retries": 8,
	}

	change, err := simulator.CreatePolicyChange("Test rollback", changes, "test-user")
	require.NoError(t, err)

	// Apply first
	err = simulator.ApplyPolicyChange(change.ID, "admin-user")
	require.NoError(t, err)

	// Then rollback
	err = simulator.RollbackPolicyChange(change.ID, "admin-user")
	assert.NoError(t, err)

	// Verify rollback
	retrieved := simulator.changes[change.ID]
	assert.Equal(t, ChangeStatusRolledBack, retrieved.Status)
	assert.NotNil(t, retrieved.RolledBackAt)
	assert.Len(t, retrieved.AuditLog, 3) // Create + Apply + Rollback
}

func TestValidatePolicyConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *PolicyConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultPolicyConfig(),
			wantErr: false,
		},
		{
			name: "negative max retries",
			config: &PolicyConfig{
				MaxRetries:        -1,
				InitialBackoff:    time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffStrategy:   "exponential",
				MaxRatePerSecond:  100.0,
				BurstSize:         10,
				MaxConcurrency:    5,
				QueueSize:         1000,
				ProcessingTimeout: 30 * time.Second,
				AckTimeout:        5 * time.Second,
				DLQEnabled:        true,
				DLQThreshold:      3,
				DLQQueueName:      "dead-letter",
			},
			wantErr: true,
		},
		{
			name: "zero initial backoff",
			config: &PolicyConfig{
				MaxRetries:        3,
				InitialBackoff:    0,
				MaxBackoff:        30 * time.Second,
				BackoffStrategy:   "exponential",
				MaxRatePerSecond:  100.0,
				BurstSize:         10,
				MaxConcurrency:    5,
				QueueSize:         1000,
				ProcessingTimeout: 30 * time.Second,
				AckTimeout:        5 * time.Second,
				DLQEnabled:        true,
				DLQThreshold:      3,
				DLQQueueName:      "dead-letter",
			},
			wantErr: true,
		},
		{
			name: "invalid backoff strategy",
			config: &PolicyConfig{
				MaxRetries:        3,
				InitialBackoff:    time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffStrategy:   "invalid",
				MaxRatePerSecond:  100.0,
				BurstSize:         10,
				MaxConcurrency:    5,
				QueueSize:         1000,
				ProcessingTimeout: 30 * time.Second,
				AckTimeout:        5 * time.Second,
				DLQEnabled:        true,
				DLQThreshold:      3,
				DLQQueueName:      "dead-letter",
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			config: &PolicyConfig{
				MaxRetries:        3,
				InitialBackoff:    time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffStrategy:   "exponential",
				MaxRatePerSecond:  -10.0,
				BurstSize:         10,
				MaxConcurrency:    5,
				QueueSize:         1000,
				ProcessingTimeout: 30 * time.Second,
				AckTimeout:        5 * time.Second,
				DLQEnabled:        true,
				DLQThreshold:      3,
				DLQQueueName:      "dead-letter",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePolicyConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTrafficPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern *TrafficPattern
		wantErr bool
	}{
		{
			name:    "valid pattern",
			pattern: DefaultTrafficPattern(),
			wantErr: false,
		},
		{
			name: "negative base rate",
			pattern: &TrafficPattern{
				Name:        "Invalid",
				Type:        TrafficConstant,
				BaseRate:    -10.0,
				Duration:    5 * time.Minute,
				Variations:  []TrafficVariation{},
				Probability: 1.0,
			},
			wantErr: true,
		},
		{
			name: "zero duration",
			pattern: &TrafficPattern{
				Name:        "Invalid",
				Type:        TrafficConstant,
				BaseRate:    50.0,
				Duration:    0,
				Variations:  []TrafficVariation{},
				Probability: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid probability",
			pattern: &TrafficPattern{
				Name:        "Invalid",
				Type:        TrafficConstant,
				BaseRate:    50.0,
				Duration:    5 * time.Minute,
				Variations:  []TrafficVariation{},
				Probability: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTrafficPattern(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQueueingModel(t *testing.T) {
	// Test M/M/1 model
	model := &QueueingModel{
		Type:        ModelMM1,
		ServiceRate: 10.0,
		ArrivalRate: 8.0,
		Servers:     1,
		Capacity:    0,
		Parameters:  make(map[string]float64),
	}

	metrics := calculateQueueingMetrics(model)
	assert.NotNil(t, metrics)
	assert.True(t, metrics.Utilization > 0)
	assert.True(t, metrics.Utilization < 1)

	// Test M/M/c model
	model.Type = ModelMMC
	model.Servers = 3
	metrics = calculateQueueingMetrics(model)
	assert.NotNil(t, metrics)
	assert.True(t, metrics.Utilization > 0)
}

func TestConcurrentSimulations(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 10 * time.Second,
		TimeStep:          500 * time.Millisecond,
		MaxWorkers:        3,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)

	// Run multiple simulations concurrently
	const numSims = 5
	results := make(chan *SimulationResult, numSims)
	errors := make(chan error, numSims)

	for i := 0; i < numSims; i++ {
		go func(id int) {
			req := &SimulationRequest{
				Name:           "Concurrent Test " + string(rune('A'+id)),
				Description:    "Testing concurrent execution",
				Policies:       DefaultPolicyConfig(),
				TrafficPattern: DefaultTrafficPattern(),
			}

			ctx := context.Background()
			result, err := simulator.RunSimulation(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(i)
	}

	// Collect results
	for i := 0; i < numSims; i++ {
		select {
		case result := <-results:
			assert.NotNil(t, result)
			assert.Equal(t, StatusCompleted, result.Status)
		case err := <-errors:
			t.Errorf("Concurrent simulation failed: %v", err)
		case <-time.After(30 * time.Second):
			t.Error("Simulation timed out")
		}
	}

	// Verify all simulations are stored
	simulations := simulator.ListSimulations()
	assert.Len(t, simulations, numSims)
}