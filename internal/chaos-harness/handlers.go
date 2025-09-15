// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// APIHandler handles chaos harness API requests
type APIHandler struct {
	injectorManager  *FaultInjectorManager
	scenarioRunner   *ScenarioRunner
	logger          *zap.Logger
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(injectorManager *FaultInjectorManager, scenarioRunner *ScenarioRunner, logger *zap.Logger) *APIHandler {
	return &APIHandler{
		injectorManager: injectorManager,
		scenarioRunner:  scenarioRunner,
		logger:         logger,
	}
}

// RegisterRoutes registers chaos harness API routes
func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	// Injector endpoints
	router.HandleFunc("/chaos/injectors", h.ListInjectors).Methods("GET")
	router.HandleFunc("/chaos/injectors", h.CreateInjector).Methods("POST")
	router.HandleFunc("/chaos/injectors/{id}", h.GetInjector).Methods("GET")
	router.HandleFunc("/chaos/injectors/{id}", h.DeleteInjector).Methods("DELETE")
	router.HandleFunc("/chaos/injectors/{id}/toggle", h.ToggleInjector).Methods("POST")

	// Scenario endpoints
	router.HandleFunc("/chaos/scenarios", h.ListScenarios).Methods("GET")
	router.HandleFunc("/chaos/scenarios", h.CreateScenario).Methods("POST")
	router.HandleFunc("/chaos/scenarios/{id}/run", h.RunScenario).Methods("POST")
	router.HandleFunc("/chaos/scenarios/{id}/abort", h.AbortScenario).Methods("POST")
	router.HandleFunc("/chaos/scenarios/{id}/report", h.GetScenarioReport).Methods("GET")

	// Control endpoints
	router.HandleFunc("/chaos/status", h.GetStatus).Methods("GET")
	router.HandleFunc("/chaos/clear", h.ClearAll).Methods("POST")
}

// ListInjectors returns all fault injectors
func (h *APIHandler) ListInjectors(w http.ResponseWriter, r *http.Request) {
	injectors := h.injectorManager.GetActiveInjectors()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"injectors": injectors,
		"count":     len(injectors),
	})
}

// CreateInjector creates a new fault injector
func (h *APIHandler) CreateInjector(w http.ResponseWriter, r *http.Request) {
	var injector FaultInjector
	if err := json.NewDecoder(r.Body).Decode(&injector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if injector.ID == "" {
		injector.ID = fmt.Sprintf("%s-%s-%d", injector.Type, injector.Scope, time.Now().Unix())
	}

	// Set created by from header or default
	injector.CreatedBy = r.Header.Get("X-User-ID")
	if injector.CreatedBy == "" {
		injector.CreatedBy = "api"
	}

	if err := h.injectorManager.AddInjector(&injector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(injector)
}

// GetInjector returns a specific injector
func (h *APIHandler) GetInjector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Find injector
	for _, injector := range h.injectorManager.GetActiveInjectors() {
		if injector.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(injector)
			return
		}
	}

	http.Error(w, "Injector not found", http.StatusNotFound)
}

// DeleteInjector removes a fault injector
func (h *APIHandler) DeleteInjector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.injectorManager.RemoveInjector(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleInjector enables/disables an injector
func (h *APIHandler) ToggleInjector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find and update injector
	for _, injector := range h.injectorManager.GetActiveInjectors() {
		if injector.ID == id {
			injector.mu.Lock()
			injector.Enabled = req.Enabled
			injector.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      id,
				"enabled": req.Enabled,
			})
			return
		}
	}

	http.Error(w, "Injector not found", http.StatusNotFound)
}

// ListScenarios returns available scenarios
func (h *APIHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	// Return predefined scenarios
	scenarios := h.getPredefinedScenarios()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios": scenarios,
		"running":   h.scenarioRunner.GetRunningScenarios(),
	})
}

// CreateScenario creates a new scenario
func (h *APIHandler) CreateScenario(w http.ResponseWriter, r *http.Request) {
	var scenario ChaosScenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if scenario.ID == "" {
		scenario.ID = fmt.Sprintf("scenario-%d", time.Now().Unix())
	}

	scenario.Status = StatusPending

	// Store scenario (in production, this would persist)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scenario)
}

// RunScenario executes a chaos scenario
func (h *APIHandler) RunScenario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Find scenario (from predefined or custom)
	var scenario *ChaosScenario
	for _, s := range h.getPredefinedScenarios() {
		if s.ID == id {
			scenario = s
			break
		}
	}

	if scenario == nil {
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Run scenario asynchronously
	go func() {
		if err := h.scenarioRunner.RunScenario(context.Background(), scenario); err != nil {
			h.logger.Error("Failed to run scenario",
				zap.String("scenario_id", id),
				zap.Error(err))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"scenario_id": id,
	})
}

// AbortScenario aborts a running scenario
func (h *APIHandler) AbortScenario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.scenarioRunner.AbortScenario(id); err != nil {
		if strings.Contains(err.Error(), "not running") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetScenarioReport returns a scenario execution report
func (h *APIHandler) GetScenarioReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Generate report (simplified for demo)
	report := &ChaosReport{
		ScenarioID:   id,
		ScenarioName: "Demo Scenario",
		ExecutedAt:   time.Now(),
		Duration:     5 * time.Minute,
		Result:       ResultPassed,
		Findings: []Finding{
			{
				Severity:    SeverityMedium,
				Type:        "slow_recovery",
				Description: "System took longer than expected to recover from Redis failover",
				Impact:      "Increased latency for 30 seconds after failover",
			},
		},
		Recommends: []string{
			"Implement circuit breaker for Redis connections",
			"Add connection pooling with health checks",
			"Reduce timeout for Redis operations",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetStatus returns chaos harness status
func (h *APIHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	injectors := h.injectorManager.GetActiveInjectors()
	running := h.scenarioRunner.GetRunningScenarios()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "active",
		"active_injectors": len(injectors),
		"running_scenarios": len(running),
		"config": map[string]interface{}{
			"enabled":          h.injectorManager.config.Enabled,
			"allow_production": h.injectorManager.config.AllowProduction,
		},
	})
}

// ClearAll removes all injectors and stops scenarios
func (h *APIHandler) ClearAll(w http.ResponseWriter, r *http.Request) {
	// Abort all running scenarios
	for _, scenario := range h.scenarioRunner.GetRunningScenarios() {
		h.scenarioRunner.AbortScenario(scenario.ID)
	}

	// Clear all injectors
	h.injectorManager.ClearAll()

	w.WriteHeader(http.StatusNoContent)
}

// getPredefinedScenarios returns predefined chaos scenarios
func (h *APIHandler) getPredefinedScenarios() []*ChaosScenario {
	return []*ChaosScenario{
		{
			ID:          "latency-test",
			Name:        "Latency Injection Test",
			Description: "Tests system behavior under increased latency",
			Duration:    5 * time.Minute,
			Stages: []ScenarioStage{
				{
					Name:     "Baseline",
					Duration: 1 * time.Minute,
					LoadConfig: &LoadConfig{
						RPS:     100,
						Pattern: LoadConstant,
					},
				},
				{
					Name:     "Inject Latency",
					Duration: 2 * time.Minute,
					Injectors: []FaultInjector{
						{
							ID:          "latency-1",
							Type:        InjectorLatency,
							Scope:       ScopeGlobal,
							Enabled:     true,
							Probability: 0.3,
							Parameters: map[string]interface{}{
								"latency_ms": 500.0,
								"jitter_ms":  100.0,
							},
						},
					},
					LoadConfig: &LoadConfig{
						RPS:     100,
						Pattern: LoadConstant,
					},
				},
				{
					Name:     "Recovery",
					Duration: 2 * time.Minute,
					LoadConfig: &LoadConfig{
						RPS:     100,
						Pattern: LoadConstant,
					},
				},
			},
			Guardrails: ScenarioGuardrails{
				MaxErrorRate:     0.5,
				MaxLatencyP99:    2 * time.Second,
				MaxBacklogSize:   10000,
				AutoAbortOnPanic: true,
			},
		},
		{
			ID:          "error-injection",
			Name:        "Error Injection Test",
			Description: "Tests error handling and recovery",
			Duration:    3 * time.Minute,
			Stages: []ScenarioStage{
				{
					Name:     "Inject Errors",
					Duration: 1 * time.Minute,
					Injectors: []FaultInjector{
						{
							ID:          "error-1",
							Type:        InjectorError,
							Scope:       ScopeWorker,
							ScopeValue:  "worker-1",
							Enabled:     true,
							Probability: 0.1,
							Parameters: map[string]interface{}{
								"error_message": "simulated processing error",
							},
						},
					},
					LoadConfig: &LoadConfig{
						RPS:     50,
						Pattern: LoadConstant,
					},
				},
			},
			Guardrails: ScenarioGuardrails{
				MaxErrorRate:     0.3,
				AutoAbortOnPanic: true,
			},
		},
	}
}
