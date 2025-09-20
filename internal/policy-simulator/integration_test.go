//go:build policy_simulator_tests
// +build policy_simulator_tests

// Copyright 2025 James Ross
package policysimulator

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullWorkflow tests the complete policy simulator workflow
func TestFullWorkflow(t *testing.T) {
	// Setup
	config := &SimulatorConfig{
		SimulationDuration: 60 * time.Second,
		TimeStep:           1 * time.Second,
		MaxWorkers:         5,
		RedisPoolSize:      3,
	}

	simulator := NewPolicySimulator(config)
	handlers := NewPolicySimulatorHandlers(simulator)
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	// Step 1: Create a simulation
	simReq := SimulationRequest{
		Name:        "Integration Test Simulation",
		Description: "Full workflow test",
		Policies: &PolicyConfig{
			MaxRetries:        5,
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        30 * time.Second,
			BackoffStrategy:   "exponential",
			MaxRatePerSecond:  75.0,
			BurstSize:         8,
			MaxConcurrency:    4,
			QueueSize:         800,
			ProcessingTimeout: 25 * time.Second,
			AckTimeout:        4 * time.Second,
			DLQEnabled:        true,
			DLQThreshold:      5,
			DLQQueueName:      "test-dlq",
		},
		TrafficPattern: &TrafficPattern{
			Name:     "Test Pattern",
			Type:     TrafficSpike,
			BaseRate: 40.0,
			Duration: 60 * time.Second,
			Variations: []TrafficVariation{
				{
					StartTime:   15 * time.Second,
					EndTime:     45 * time.Second,
					Multiplier:  2.5,
					Description: "2.5x spike for 30 seconds",
				},
			},
			Probability: 1.0,
		},
	}

	reqBody, err := json.Marshal(simReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/policy-simulator/simulations", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var simResult SimulationResult
	err = json.NewDecoder(w.Body).Decode(&simResult)
	require.NoError(t, err)

	simulationID := simResult.ID
	assert.NotEmpty(t, simulationID)
	assert.Equal(t, StatusCompleted, simResult.Status)

	// Step 2: Retrieve the simulation
	req = httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+simulationID, nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedSim SimulationResult
	err = json.NewDecoder(w.Body).Decode(&retrievedSim)
	require.NoError(t, err)

	assert.Equal(t, simulationID, retrievedSim.ID)
	assert.Equal(t, "Integration Test Simulation", retrievedSim.Name)
	assert.NotNil(t, retrievedSim.Metrics)
	assert.NotEmpty(t, retrievedSim.Timeline)

	// Step 3: Get charts for the simulation
	req = httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+simulationID+"/charts", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var chartsResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&chartsResp)
	require.NoError(t, err)

	charts := chartsResp["charts"].([]interface{})
	assert.Len(t, charts, 3) // Queue depth, processing rate, resource usage

	// Step 4: Create a policy change
	policyChangeReq := CreatePolicyChangeRequest{
		Description: "Increase retry limit and rate",
		Changes: map[string]interface{}{
			"max_retries":         8,
			"max_rate_per_second": 120.0,
		},
	}

	reqBody, err = json.Marshal(policyChangeReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "integration-test-user")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var policyChange PolicyChange
	err = json.NewDecoder(w.Body).Decode(&policyChange)
	require.NoError(t, err)

	changeID := policyChange.ID
	assert.NotEmpty(t, changeID)
	assert.Equal(t, ChangeStatusProposed, policyChange.Status)

	// Step 5: Apply the policy change
	applyReq := ApplyPolicyChangeRequest{
		Reason: "Integration test application",
	}

	reqBody, err = json.Marshal(applyReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes/"+changeID+"/apply", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "admin-user")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var applyResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&applyResp)
	require.NoError(t, err)

	assert.Equal(t, "Policy change applied successfully", applyResp["message"])
	assert.Equal(t, changeID, applyResp["change_id"])

	// Step 6: Rollback the policy change
	rollbackReq := RollbackPolicyChangeRequest{
		Reason: "Integration test rollback",
	}

	reqBody, err = json.Marshal(rollbackReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes/"+changeID+"/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "admin-user")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var rollbackResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&rollbackResp)
	require.NoError(t, err)

	assert.Equal(t, "Policy change rolled back successfully", rollbackResp["message"])
	assert.Equal(t, changeID, rollbackResp["change_id"])

	// Step 7: List all simulations
	req = httptest.NewRequest("GET", "/api/policy-simulator/simulations", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&listResp)
	require.NoError(t, err)

	simulations := listResp["simulations"].([]interface{})
	assert.Len(t, simulations, 1)
	assert.Equal(t, float64(1), listResp["total"])

	// Step 8: Test presets
	req = httptest.NewRequest("GET", "/api/policy-simulator/presets/policies", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var policyPresets map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&policyPresets)
	require.NoError(t, err)

	presets := policyPresets["presets"].(map[string]interface{})
	assert.Contains(t, presets, "conservative")
	assert.Contains(t, presets, "aggressive")
	assert.Contains(t, presets, "balanced")

	req = httptest.NewRequest("GET", "/api/policy-simulator/presets/traffic", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var trafficPresets map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&trafficPresets)
	require.NoError(t, err)

	trafficPresetsData := trafficPresets["presets"].(map[string]interface{})
	assert.Contains(t, trafficPresetsData, "steady")
	assert.Contains(t, trafficPresetsData, "spike")
}

// TestConcurrentOperations tests concurrent API operations
func TestConcurrentOperations(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:           1 * time.Second,
		MaxWorkers:         3,
		RedisPoolSize:      2,
	}

	simulator := NewPolicySimulator(config)
	handlers := NewPolicySimulatorHandlers(simulator)
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	const numConcurrentOps = 10
	results := make(chan string, numConcurrentOps)
	errors := make(chan error, numConcurrentOps)

	// Run multiple simulations concurrently
	for i := 0; i < numConcurrentOps; i++ {
		go func(id int) {
			simReq := SimulationRequest{
				Name:           "Concurrent Test " + string(rune('A'+id)),
				Description:    "Concurrent integration test",
				Policies:       DefaultPolicyConfig(),
				TrafficPattern: DefaultTrafficPattern(),
			}

			reqBody, err := json.Marshal(simReq)
			if err != nil {
				errors <- err
				return
			}

			req := httptest.NewRequest("POST", "/api/policy-simulator/simulations", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				errors <- assert.AnError
				return
			}

			var result SimulationResult
			err = json.NewDecoder(w.Body).Decode(&result)
			if err != nil {
				errors <- err
				return
			}

			results <- result.ID
		}(i)
	}

	// Collect results
	var simulationIDs []string
	for i := 0; i < numConcurrentOps; i++ {
		select {
		case id := <-results:
			simulationIDs = append(simulationIDs, id)
		case err := <-errors:
			t.Errorf("Concurrent operation failed: %v", err)
		case <-time.After(60 * time.Second):
			t.Error("Concurrent operation timed out")
		}
	}

	assert.Len(t, simulationIDs, numConcurrentOps)

	// Verify all simulations are accessible
	for _, id := range simulationIDs {
		req := httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+id, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 10 * time.Second,
		TimeStep:           1 * time.Second,
		MaxWorkers:         3,
		RedisPoolSize:      2,
	}

	simulator := NewPolicySimulator(config)
	handlers := NewPolicySimulatorHandlers(simulator)
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	// Test invalid simulation request
	invalidReq := map[string]interface{}{
		"name": "Invalid Test",
		"policies": map[string]interface{}{
			"max_retries": -1, // Invalid value
		},
	}

	reqBody, err := json.Marshal(invalidReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/policy-simulator/simulations", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test getting non-existent simulation
	req = httptest.NewRequest("GET", "/api/policy-simulator/simulations/nonexistent", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test getting charts for non-existent simulation
	req = httptest.NewRequest("GET", "/api/policy-simulator/simulations/nonexistent/charts", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test applying non-existent policy change
	applyReq := ApplyPolicyChangeRequest{
		Reason: "Test",
	}

	reqBody, err = json.Marshal(applyReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes/nonexistent/apply", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test rolling back non-existent policy change
	rollbackReq := RollbackPolicyChangeRequest{
		Reason: "Test",
	}

	reqBody, err = json.Marshal(rollbackReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes/nonexistent/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLongRunningSimulation tests simulation with longer duration
func TestLongRunningSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	config := &SimulatorConfig{
		SimulationDuration: 2 * time.Minute,
		TimeStep:           2 * time.Second,
		MaxWorkers:         8,
		RedisPoolSize:      4,
	}

	simulator := NewPolicySimulator(config)

	complexPattern := &TrafficPattern{
		Name:     "Complex Pattern",
		Type:     TrafficSeasonal,
		BaseRate: 60.0,
		Duration: 2 * time.Minute,
		Variations: []TrafficVariation{
			{
				StartTime:   0,
				EndTime:     30 * time.Second,
				Multiplier:  1.5,
				Description: "Initial ramp up",
			},
			{
				StartTime:   30 * time.Second,
				EndTime:     90 * time.Second,
				Multiplier:  3.0,
				Description: "Peak load",
			},
			{
				StartTime:   90 * time.Second,
				EndTime:     2 * time.Minute,
				Multiplier:  0.7,
				Description: "Wind down",
			},
		},
		Probability: 1.0,
	}

	complexPolicy := &PolicyConfig{
		MaxRetries:        8,
		InitialBackoff:    200 * time.Millisecond,
		MaxBackoff:        45 * time.Second,
		BackoffStrategy:   "exponential",
		MaxRatePerSecond:  150.0,
		BurstSize:         20,
		MaxConcurrency:    8,
		QueueSize:         2000,
		ProcessingTimeout: 45 * time.Second,
		AckTimeout:        6 * time.Second,
		DLQEnabled:        true,
		DLQThreshold:      8,
		DLQQueueName:      "complex-dlq",
	}

	req := &SimulationRequest{
		Name:           "Long Running Test",
		Description:    "Testing with complex patterns and longer duration",
		Policies:       complexPolicy,
		TrafficPattern: complexPattern,
	}

	ctx := context.Background()
	start := time.Now()

	result, err := simulator.RunSimulation(ctx, req)

	elapsed := time.Since(start)
	t.Logf("Simulation completed in %v", elapsed)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusCompleted, result.Status)
	assert.NotNil(t, result.Metrics)
	assert.NotEmpty(t, result.Timeline)

	// Should have many timeline snapshots for a 2-minute simulation
	assert.True(t, len(result.Timeline) >= 50)

	// Verify metrics make sense for this complex scenario
	assert.True(t, result.Metrics.MessagesProcessed > 0)
	assert.True(t, result.Metrics.ProcessingRate > 0)
	assert.True(t, result.Metrics.Duration > 0)
	assert.True(t, result.Metrics.MaxQueueDepth >= 0)

	// Should have processed more messages due to longer duration and higher rates
	assert.True(t, result.Metrics.MessagesProcessed > 100)
}

// TestMemoryAndPerformance tests memory usage and performance characteristics
func TestMemoryAndPerformance(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 60 * time.Second,
		TimeStep:           500 * time.Millisecond,
		MaxWorkers:         10,
		RedisPoolSize:      5,
	}

	simulator := NewPolicySimulator(config)

	highVolumePattern := &TrafficPattern{
		Name:        "High Volume",
		Type:        TrafficConstant,
		BaseRate:    500.0, // Very high rate
		Duration:    60 * time.Second,
		Variations:  []TrafficVariation{},
		Probability: 1.0,
	}

	req := &SimulationRequest{
		Name:           "Performance Test",
		Description:    "Testing high volume simulation",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: highVolumePattern,
	}

	ctx := context.Background()
	start := time.Now()

	result, err := simulator.RunSimulation(ctx, req)

	elapsed := time.Since(start)
	t.Logf("High volume simulation completed in %v", elapsed)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusCompleted, result.Status)

	// Should complete within reasonable time (not more than 30 seconds for simulation)
	assert.True(t, elapsed < 30*time.Second, "Simulation took too long: %v", elapsed)

	// Should have processed many messages
	assert.True(t, result.Metrics.MessagesProcessed > 1000)

	// Timeline should be substantial but not excessive
	assert.True(t, len(result.Timeline) > 50)
	assert.True(t, len(result.Timeline) < 500) // Reasonable upper bound
}

// TestPolicyValidationIntegration tests policy validation in the full flow
func TestPolicyValidationIntegration(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:           1 * time.Second,
		MaxWorkers:         5,
		RedisPoolSize:      3,
	}

	simulator := NewPolicySimulator(config)
	handlers := NewPolicySimulatorHandlers(simulator)
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	// Test various invalid configurations
	invalidConfigs := []struct {
		name   string
		policy *PolicyConfig
	}{
		{
			name: "Negative retries",
			policy: &PolicyConfig{
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
				DLQQueueName:      "dlq",
			},
		},
		{
			name: "Zero backoff",
			policy: &PolicyConfig{
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
				DLQQueueName:      "dlq",
			},
		},
		{
			name: "Invalid strategy",
			policy: &PolicyConfig{
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
				DLQQueueName:      "dlq",
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			simReq := SimulationRequest{
				Name:           "Invalid Test - " + tc.name,
				Description:    "Testing invalid configuration",
				Policies:       tc.policy,
				TrafficPattern: DefaultTrafficPattern(),
			}

			reqBody, err := json.Marshal(simReq)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/policy-simulator/simulations", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResp map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp["error"], "Failed to create simulation")
		})
	}
}
