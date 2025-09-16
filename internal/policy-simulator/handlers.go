// Copyright 2025 James Ross
package policysimulator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// PolicySimulatorHandlers provides HTTP handlers for the policy simulator
type PolicySimulatorHandlers struct {
	simulator *PolicySimulator
}

// NewPolicySimulatorHandlers creates new policy simulator handlers
func NewPolicySimulatorHandlers(simulator *PolicySimulator) *PolicySimulatorHandlers {
	return &PolicySimulatorHandlers{
		simulator: simulator,
	}
}

// RegisterRoutes registers all policy simulator routes
func (h *PolicySimulatorHandlers) RegisterRoutes(router *mux.Router) {
	// Simulation endpoints
	router.HandleFunc("/api/policy-simulator/simulations", h.CreateSimulation).Methods("POST")
	router.HandleFunc("/api/policy-simulator/simulations", h.ListSimulations).Methods("GET")
	router.HandleFunc("/api/policy-simulator/simulations/{id}", h.GetSimulation).Methods("GET")

	// Policy change endpoints
	router.HandleFunc("/api/policy-simulator/changes", h.CreatePolicyChange).Methods("POST")
	router.HandleFunc("/api/policy-simulator/changes", h.ListPolicyChanges).Methods("GET")
	router.HandleFunc("/api/policy-simulator/changes/{id}", h.GetPolicyChange).Methods("GET")
	router.HandleFunc("/api/policy-simulator/changes/{id}/apply", h.ApplyPolicyChange).Methods("POST")
	router.HandleFunc("/api/policy-simulator/changes/{id}/rollback", h.RollbackPolicyChange).Methods("POST")

	// Configuration endpoints
	router.HandleFunc("/api/policy-simulator/presets/policies", h.GetPolicyPresets).Methods("GET")
	router.HandleFunc("/api/policy-simulator/presets/traffic", h.GetTrafficPresets).Methods("GET")

	// Chart data endpoints
	router.HandleFunc("/api/policy-simulator/simulations/{id}/charts", h.GetSimulationCharts).Methods("GET")
}

// CreateSimulation handles POST /api/policy-simulator/simulations
func (h *PolicySimulatorHandlers) CreateSimulation(w http.ResponseWriter, r *http.Request) {
	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.simulator.RunSimulation(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to create simulation", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, result)
}

// ListSimulations handles GET /api/policy-simulator/simulations
func (h *PolicySimulatorHandlers) ListSimulations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	limitStr := query.Get("limit")
	statusFilter := query.Get("status")

	simulations := h.simulator.ListSimulations()

	// Apply status filter
	if statusFilter != "" {
		filtered := make([]*SimulationResult, 0)
		for _, sim := range simulations {
			if string(sim.Status) == statusFilter {
				filtered = append(filtered, sim)
			}
		}
		simulations = filtered
	}

	// Apply limit
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(simulations) {
			simulations = simulations[:limit]
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"simulations": simulations,
		"total":       len(simulations),
	})
}

// GetSimulation handles GET /api/policy-simulator/simulations/{id}
func (h *PolicySimulatorHandlers) GetSimulation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.simulator.GetSimulation(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Simulation not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// CreatePolicyChangeRequest represents the request to create a policy change
type CreatePolicyChangeRequest struct {
	Description string                 `json:"description"`
	Changes     map[string]interface{} `json:"changes"`
}

// CreatePolicyChange handles POST /api/policy-simulator/changes
func (h *PolicySimulatorHandlers) CreatePolicyChange(w http.ResponseWriter, r *http.Request) {
	var req CreatePolicyChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user from context (simplified - in real implementation, use proper auth)
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = "anonymous"
	}

	change, err := h.simulator.CreatePolicyChange(req.Description, req.Changes, user)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to create policy change", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, change)
}

// ListPolicyChanges handles GET /api/policy-simulator/changes
func (h *PolicySimulatorHandlers) ListPolicyChanges(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch from the simulator
	// For now, return empty list as the simulator doesn't expose this method
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"changes": []PolicyChange{},
		"total":   0,
	})
}

// GetPolicyChange handles GET /api/policy-simulator/changes/{id}
func (h *PolicySimulatorHandlers) GetPolicyChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// In a real implementation, this would fetch the specific change
	h.writeError(w, http.StatusNotFound, "Policy change not found", fmt.Errorf("change %s not found", id))
}

// ApplyPolicyChangeRequest represents the request to apply a policy change
type ApplyPolicyChangeRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ApplyPolicyChange handles POST /api/policy-simulator/changes/{id}/apply
func (h *PolicySimulatorHandlers) ApplyPolicyChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req ApplyPolicyChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user from context
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = "anonymous"
	}

	err := h.simulator.ApplyPolicyChange(id, user)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to apply policy change", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Policy change applied successfully",
		"change_id":  id,
		"applied_by": user,
		"applied_at": time.Now().UTC(),
	})
}

// RollbackPolicyChangeRequest represents the request to rollback a policy change
type RollbackPolicyChangeRequest struct {
	Reason string `json:"reason,omitempty"`
}

// RollbackPolicyChange handles POST /api/policy-simulator/changes/{id}/rollback
func (h *PolicySimulatorHandlers) RollbackPolicyChange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req RollbackPolicyChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user from context
	user := r.Header.Get("X-User-ID")
	if user == "" {
		user = "anonymous"
	}

	err := h.simulator.RollbackPolicyChange(id, user)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to rollback policy change", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Policy change rolled back successfully",
		"change_id":      id,
		"rolled_back_by": user,
		"rolled_back_at": time.Now().UTC(),
	})
}

// GetPolicyPresets handles GET /api/policy-simulator/presets/policies
func (h *PolicySimulatorHandlers) GetPolicyPresets(w http.ResponseWriter, r *http.Request) {
	presets := map[string]*PolicyConfig{
		"conservative": {
			MaxRetries:        5,
			InitialBackoff:    2 * time.Second,
			MaxBackoff:        60 * time.Second,
			BackoffStrategy:   "exponential",
			MaxRatePerSecond:  50.0,
			BurstSize:         5,
			MaxConcurrency:    3,
			QueueSize:         500,
			ProcessingTimeout: 60 * time.Second,
			AckTimeout:        10 * time.Second,
			DLQEnabled:        true,
			DLQThreshold:      5,
			DLQQueueName:      "dead-letter",
		},
		"aggressive": {
			MaxRetries:        2,
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        10 * time.Second,
			BackoffStrategy:   "exponential",
			MaxRatePerSecond:  200.0,
			BurstSize:         20,
			MaxConcurrency:    10,
			QueueSize:         2000,
			ProcessingTimeout: 15 * time.Second,
			AckTimeout:        3 * time.Second,
			DLQEnabled:        true,
			DLQThreshold:      2,
			DLQQueueName:      "dead-letter",
		},
		"balanced": DefaultPolicyConfig(),
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"presets": presets,
	})
}

// GetTrafficPresets handles GET /api/policy-simulator/presets/traffic
func (h *PolicySimulatorHandlers) GetTrafficPresets(w http.ResponseWriter, r *http.Request) {
	presets := map[string]*TrafficPattern{
		"steady": {
			Name:        "Steady Load",
			Type:        TrafficConstant,
			BaseRate:    50.0,
			Duration:    5 * time.Minute,
			Variations:  []TrafficVariation{},
			Probability: 1.0,
		},
		"spike": {
			Name:     "Traffic Spike",
			Type:     TrafficSpike,
			BaseRate: 30.0,
			Duration: 10 * time.Minute,
			Variations: []TrafficVariation{
				{
					StartTime:   3 * time.Minute,
					EndTime:     5 * time.Minute,
					Multiplier:  5.0,
					Description: "5x spike for 2 minutes",
				},
			},
			Probability: 1.0,
		},
		"seasonal": {
			Name:     "Seasonal Pattern",
			Type:     TrafficSeasonal,
			BaseRate: 40.0,
			Duration: 60 * time.Minute,
			Variations: []TrafficVariation{
				{
					StartTime:   0,
					EndTime:     60 * time.Minute,
					Multiplier:  1.0,
					Description: "Sinusoidal pattern over 1 hour",
				},
			},
			Probability: 1.0,
		},
		"bursty": {
			Name:        "Bursty Load",
			Type:        TrafficBursty,
			BaseRate:    25.0,
			Duration:    15 * time.Minute,
			Variations:  []TrafficVariation{},
			Probability: 1.0,
		},
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"presets": presets,
	})
}

// GetSimulationCharts handles GET /api/policy-simulator/simulations/{id}/charts
func (h *PolicySimulatorHandlers) GetSimulationCharts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.simulator.GetSimulation(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Simulation not found", err)
		return
	}

	if result.Status != StatusCompleted {
		h.writeError(w, http.StatusBadRequest, "Simulation not completed", fmt.Errorf("simulation status: %s", result.Status))
		return
	}

	// Generate chart data from simulation timeline
	charts := h.generateChartData(result)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"charts": charts,
	})
}

// generateChartData creates chart data from simulation results
func (h *PolicySimulatorHandlers) generateChartData(result *SimulationResult) []ChartData {
	var charts []ChartData

	if len(result.Timeline) == 0 {
		return charts
	}

	// Queue Depth Chart
	queueChart := ChartData{
		Title: "Queue Depth Over Time",
		Type:  ChartLine,
		XAxis: AxisConfig{
			Label:  "Time",
			Unit:   "seconds",
			Format: "timestamp",
		},
		YAxis: AxisConfig{
			Label: "Queue Depth",
			Unit:  "messages",
		},
		Series: []ChartSeries{
			{
				Name:  "Queue Depth",
				Color: "#3b82f6",
				Points: make([]ChartPoint, len(result.Timeline)),
			},
		},
	}

	// Processing Rate Chart
	rateChart := ChartData{
		Title: "Processing Rate Over Time",
		Type:  ChartLine,
		XAxis: AxisConfig{
			Label:  "Time",
			Unit:   "seconds",
			Format: "timestamp",
		},
		YAxis: AxisConfig{
			Label: "Rate",
			Unit:  "msg/sec",
		},
		Series: []ChartSeries{
			{
				Name:  "Processing Rate",
				Color: "#10b981",
				Points: make([]ChartPoint, len(result.Timeline)),
			},
		},
	}

	// Resource Usage Chart
	resourceChart := ChartData{
		Title: "Resource Usage Over Time",
		Type:  ChartLine,
		XAxis: AxisConfig{
			Label:  "Time",
			Unit:   "seconds",
			Format: "timestamp",
		},
		YAxis: AxisConfig{
			Label: "Usage",
			Unit:  "%",
		},
		Series: []ChartSeries{
			{
				Name:  "CPU Usage",
				Color: "#f59e0b",
				Points: make([]ChartPoint, len(result.Timeline)),
			},
			{
				Name:  "Memory Usage",
				Color: "#ef4444",
				Points: make([]ChartPoint, len(result.Timeline)),
			},
		},
	}

	// Populate chart data points
	startTime := result.Timeline[0].Timestamp
	for i, snapshot := range result.Timeline {
		timeOffset := snapshot.Timestamp.Sub(startTime).Seconds()

		// Queue depth points
		queueChart.Series[0].Points[i] = ChartPoint{
			X: timeOffset,
			Y: float64(snapshot.QueueDepth),
		}

		// Processing rate points
		rateChart.Series[0].Points[i] = ChartPoint{
			X: timeOffset,
			Y: snapshot.ProcessingRate,
		}

		// Resource usage points
		resourceChart.Series[0].Points[i] = ChartPoint{
			X: timeOffset,
			Y: snapshot.CPUUsage,
		}
		resourceChart.Series[1].Points[i] = ChartPoint{
			X: timeOffset,
			Y: snapshot.MemoryUsage / 100, // Convert to percentage for visualization
		}
	}

	charts = append(charts, queueChart, rateChart, resourceChart)

	// Add annotations for significant events
	if result.Metrics != nil {
		// Add annotation for peak queue depth
		for _, chart := range charts {
			if chart.Title == "Queue Depth Over Time" {
				peakAnnotation := Annotation{
					Type:        AnnotationPoint,
					Y:           float64(result.Metrics.MaxQueueDepth),
					Text:        fmt.Sprintf("Peak: %d", result.Metrics.MaxQueueDepth),
					Color:       "#dc2626",
					Description: "Maximum queue depth reached during simulation",
				}
				chart.Annotations = append(chart.Annotations, peakAnnotation)
			}
		}
	}

	return charts
}

// Helper methods for HTTP responses
func (h *PolicySimulatorHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *PolicySimulatorHandlers) writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error":   message,
		"status":  status,
		"timestamp": time.Now().UTC(),
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// HealthCheckHandler provides a health check endpoint for the policy simulator
func (h *PolicySimulatorHandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "policy-simulator",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}