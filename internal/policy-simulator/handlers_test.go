// Copyright 2025 James Ross
package policysimulator

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandlers() (*PolicySimulatorHandlers, *PolicySimulator) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	handlers := NewPolicySimulatorHandlers(simulator)

	return handlers, simulator
}

func TestCreateSimulation(t *testing.T) {
	handlers, _ := setupTestHandlers()

	req := SimulationRequest{
		Name:           "Test API Simulation",
		Description:    "Testing via API",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/simulations", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateSimulation(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var result SimulationResult
	err = json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Test API Simulation", result.Name)
	assert.Equal(t, "Testing via API", result.Description)
	assert.NotEmpty(t, result.ID)

	// Wait for simulation to complete (it runs asynchronously)
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		simResult, err := handlers.simulator.GetSimulation(result.ID)
		if err == nil && simResult.Status == StatusCompleted {
			break
		}
	}

	// Get final result
	finalResult, err := handlers.simulator.GetSimulation(result.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, finalResult.Status)
}

func TestCreateSimulationInvalidBody(t *testing.T) {
	handlers, _ := setupTestHandlers()

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/simulations", strings.NewReader("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateSimulation(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request body", errorResp["error"])
}

func TestListSimulations(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create some test simulations
	req := &SimulationRequest{
		Name:           "Test Sim 1",
		Description:    "First test",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	_, err := simulator.RunSimulation(context.Background(), req)
	require.NoError(t, err)

	req.Name = "Test Sim 2"
	req.Description = "Second test"
	_, err = simulator.RunSimulation(context.Background(), req)
	require.NoError(t, err)

	// Test listing without filters
	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations", nil)
	w := httptest.NewRecorder()

	handlers.ListSimulations(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	simulations := response["simulations"].([]interface{})
	assert.Len(t, simulations, 2)
	assert.Equal(t, float64(2), response["total"])
}

func TestListSimulationsWithLimit(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create test simulations
	for i := 0; i < 5; i++ {
		req := &SimulationRequest{
			Name:           "Test Sim " + string(rune('A'+i)),
			Description:    "Test simulation",
			Policies:       DefaultPolicyConfig(),
			TrafficPattern: DefaultTrafficPattern(),
		}

		_, err := simulator.RunSimulation(context.Background(), req)
		require.NoError(t, err)
	}

	// Test with limit
	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations?limit=3", nil)
	w := httptest.NewRecorder()

	handlers.ListSimulations(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	simulations := response["simulations"].([]interface{})
	assert.Len(t, simulations, 3)
}

func TestGetSimulation(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create a test simulation
	req := &SimulationRequest{
		Name:           "Get Test Sim",
		Description:    "For get testing",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	result, err := simulator.RunSimulation(context.Background(), req)
	require.NoError(t, err)

	// Test getting the simulation
	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+result.ID, nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": result.ID})
	w := httptest.NewRecorder()

	handlers.GetSimulation(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var retrieved SimulationResult
	err = json.NewDecoder(w.Body).Decode(&retrieved)
	require.NoError(t, err)

	assert.Equal(t, result.ID, retrieved.ID)
	assert.Equal(t, "Get Test Sim", retrieved.Name)
}

func TestGetSimulationNotFound(t *testing.T) {
	handlers, _ := setupTestHandlers()

	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations/nonexistent", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	handlers.GetSimulation(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Equal(t, "Simulation not found", errorResp["error"])
}

func TestCreatePolicyChange(t *testing.T) {
	handlers, _ := setupTestHandlers()

	req := CreatePolicyChangeRequest{
		Description: "Test policy change",
		Changes: map[string]interface{}{
			"max_retries":        float64(5), // JSON unmarshaling converts to float64
			"max_rate_per_second": 200.0,
		},
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/changes", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	handlers.CreatePolicyChange(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var change PolicyChange
	err = json.NewDecoder(w.Body).Decode(&change)
	require.NoError(t, err)

	assert.Equal(t, "Test policy change", change.Description)
	assert.Equal(t, req.Changes, change.Changes)
	assert.Equal(t, "test-user", change.AppliedBy)
	assert.NotEmpty(t, change.ID)
}

func TestCreatePolicyChangeAnonymousUser(t *testing.T) {
	handlers, _ := setupTestHandlers()

	req := CreatePolicyChangeRequest{
		Description: "Anonymous change",
		Changes: map[string]interface{}{
			"max_retries": 3,
		},
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/changes", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreatePolicyChange(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var change PolicyChange
	err = json.NewDecoder(w.Body).Decode(&change)
	require.NoError(t, err)

	assert.Equal(t, "anonymous", change.AppliedBy)
}

func TestApplyPolicyChange(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create a policy change first
	changes := map[string]interface{}{
		"max_retries": 7,
	}

	change, err := simulator.CreatePolicyChange("Test apply", changes, "test-user")
	require.NoError(t, err)

	// Manually approve the change for testing
	simulator.changes[change.ID].Status = ChangeStatusApproved

	// Apply the change
	req := ApplyPolicyChangeRequest{
		Reason: "Testing apply functionality",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/changes/"+change.ID+"/apply", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", "admin-user")
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": change.ID})
	w := httptest.NewRecorder()

	handlers.ApplyPolicyChange(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Policy change applied successfully", response["message"])
	assert.Equal(t, change.ID, response["change_id"])
	assert.Equal(t, "admin-user", response["applied_by"])
	assert.NotNil(t, response["applied_at"])
}

func TestApplyPolicyChangeNotFound(t *testing.T) {
	handlers, _ := setupTestHandlers()

	req := ApplyPolicyChangeRequest{
		Reason: "Test",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/changes/nonexistent/apply", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	handlers.ApplyPolicyChange(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRollbackPolicyChange(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create and apply a policy change
	changes := map[string]interface{}{
		"max_retries": 8,
	}

	change, err := simulator.CreatePolicyChange("Test rollback", changes, "test-user")
	require.NoError(t, err)

	// Manually approve the change for testing
	simulator.changes[change.ID].Status = ChangeStatusApproved

	err = simulator.ApplyPolicyChange(change.ID, "admin-user")
	require.NoError(t, err)

	// Rollback the change
	req := RollbackPolicyChangeRequest{
		Reason: "Testing rollback functionality",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/policy-simulator/changes/"+change.ID+"/rollback", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", "admin-user")
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": change.ID})
	w := httptest.NewRecorder()

	handlers.RollbackPolicyChange(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Policy change rolled back successfully", response["message"])
	assert.Equal(t, change.ID, response["change_id"])
	assert.Equal(t, "admin-user", response["rolled_back_by"])
	assert.NotNil(t, response["rolled_back_at"])
}

func TestGetPolicyPresets(t *testing.T) {
	handlers, _ := setupTestHandlers()

	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/presets/policies", nil)
	w := httptest.NewRecorder()

	handlers.GetPolicyPresets(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	presets := response["presets"].(map[string]interface{})
	assert.Contains(t, presets, "conservative")
	assert.Contains(t, presets, "aggressive")
	assert.Contains(t, presets, "balanced")

	// Check conservative preset structure
	conservative := presets["conservative"].(map[string]interface{})
	assert.Equal(t, float64(5), conservative["max_retries"])
	assert.Equal(t, float64(50), conservative["max_rate_per_second"])
}

func TestGetTrafficPresets(t *testing.T) {
	handlers, _ := setupTestHandlers()

	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/presets/traffic", nil)
	w := httptest.NewRecorder()

	handlers.GetTrafficPresets(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	presets := response["presets"].(map[string]interface{})
	assert.Contains(t, presets, "steady")
	assert.Contains(t, presets, "spike")
	assert.Contains(t, presets, "seasonal")
	assert.Contains(t, presets, "bursty")

	// Check spike preset structure
	spike := presets["spike"].(map[string]interface{})
	assert.Equal(t, "Traffic Spike", spike["name"])
	assert.Equal(t, "spike", spike["type"])
	assert.Equal(t, float64(30), spike["base_rate"])
}

func TestGetSimulationCharts(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create a completed simulation
	req := &SimulationRequest{
		Name:           "Chart Test Sim",
		Description:    "For chart testing",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	result, err := simulator.RunSimulation(context.Background(), req)
	require.NoError(t, err)

	// Wait for simulation to complete
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		simResult, err := simulator.GetSimulation(result.ID)
		if err == nil && simResult.Status == StatusCompleted {
			break
		}
	}

	// Get charts for the simulation
	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+result.ID+"/charts", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": result.ID})
	w := httptest.NewRecorder()

	handlers.GetSimulationCharts(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	charts := response["charts"].([]interface{})
	assert.Len(t, charts, 3) // Queue depth, processing rate, resource usage

	// Check chart structure
	queueChart := charts[0].(map[string]interface{})
	assert.Equal(t, "Queue Depth Over Time", queueChart["title"])
	assert.Equal(t, "line", queueChart["type"])
	assert.NotNil(t, queueChart["series"])
}

func TestGetSimulationChartsNotCompleted(t *testing.T) {
	handlers, simulator := setupTestHandlers()

	// Create a simulation but manually set it to running status
	req := &SimulationRequest{
		Name:           "Running Sim",
		Description:    "Still running",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	result, err := simulator.RunSimulation(context.Background(), req)
	require.NoError(t, err)

	// Manually change status to running
	simulator.simulations[result.ID].Status = StatusRunning

	httpReq := httptest.NewRequest("GET", "/api/policy-simulator/simulations/"+result.ID+"/charts", nil)
	httpReq = mux.SetURLVars(httpReq, map[string]string{"id": result.ID})
	w := httptest.NewRecorder()

	handlers.GetSimulationCharts(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Equal(t, "Simulation not completed", errorResp["error"])
}

func TestHealthCheckHandler(t *testing.T) {
	handlers, _ := setupTestHandlers()

	httpReq := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthCheckHandler(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "policy-simulator", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotNil(t, response["timestamp"])
}

func TestRegisterRoutes(t *testing.T) {
	handlers, _ := setupTestHandlers()

	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	// Test that routes are registered correctly by checking if they match
	req := httptest.NewRequest("GET", "/api/policy-simulator/simulations", nil)
	match := &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req = httptest.NewRequest("POST", "/api/policy-simulator/changes", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))

	req = httptest.NewRequest("GET", "/api/policy-simulator/presets/policies", nil)
	match = &mux.RouteMatch{}
	assert.True(t, router.Match(req, match))
}

func TestJSONResponseHelpers(t *testing.T) {
	handlers, _ := setupTestHandlers()

	// Test writeJSON
	w := httptest.NewRecorder()
	data := map[string]string{"test": "value"}
	handlers.writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "value", response["test"])

	// Test writeError
	w = httptest.NewRecorder()
	handlers.writeError(w, http.StatusBadRequest, "test error", assert.AnError)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Equal(t, "test error", errorResp["error"])
	assert.Equal(t, float64(400), errorResp["status"])
	assert.NotNil(t, errorResp["timestamp"])
	assert.NotNil(t, errorResp["details"])
}