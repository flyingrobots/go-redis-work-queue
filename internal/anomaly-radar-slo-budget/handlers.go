package anomalyradarslobudget

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// HTTPHandler provides HTTP endpoints for anomaly radar management
type HTTPHandler struct {
	radar *AnomalyRadar
}

// NewHTTPHandler creates a new HTTP handler for anomaly radar
func NewHTTPHandler(radar *AnomalyRadar) *HTTPHandler {
	return &HTTPHandler{
		radar: radar,
	}
}

// StatusRequest represents a request for current status
type StatusRequest struct {
	IncludeMetrics bool          `json:"include_metrics"`
	MetricWindow   time.Duration `json:"metric_window"`
}

// StatusResponse represents the response with current status
type StatusResponse struct {
	AnomalyStatus AnomalyStatus   `json:"anomaly_status"`
	SLOBudget     SLOBudget       `json:"slo_budget"`
	Metrics       []MetricSnapshot `json:"metrics,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ConfigResponse represents configuration information
type ConfigResponse struct {
	Config    Config `json:"config"`
	Summary   string `json:"summary"`
	IsValid   bool   `json:"is_valid"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricsRequest represents a request for historical metrics
type MetricsRequest struct {
	Window     time.Duration `json:"window"`
	MaxSamples int          `json:"max_samples"`
}

// MetricsResponse represents historical metrics data
type MetricsResponse struct {
	Metrics   []MetricSnapshot `json:"metrics"`
	Window    time.Duration    `json:"window"`
	Count     int             `json:"count"`
	Timestamp time.Time       `json:"timestamp"`
}

// AlertsResponse represents active alerts
type AlertsResponse struct {
	Alerts    []Alert   `json:"alerts"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	IsRunning     bool          `json:"is_running"`
	Status        MetricStatus  `json:"status"`
	AlertLevel    AlertLevel    `json:"alert_level"`
	ActiveAlerts  int          `json:"active_alerts"`
	LastUpdated   time.Time    `json:"last_updated"`
	Uptime        time.Duration `json:"uptime"`
	Timestamp     time.Time    `json:"timestamp"`
}

// GetStatus returns the current anomaly and SLO status
func (h *HTTPHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest

	// Parse query parameters
	if includeMetrics := r.URL.Query().Get("include_metrics"); includeMetrics == "true" {
		req.IncludeMetrics = true
	}

	if windowStr := r.URL.Query().Get("metric_window"); windowStr != "" {
		if window, err := time.ParseDuration(windowStr); err == nil {
			req.MetricWindow = window
		} else {
			req.MetricWindow = time.Hour // Default to 1 hour
		}
	} else {
		req.MetricWindow = time.Hour
	}

	// Get current status
	anomalyStatus, sloBudget := h.radar.GetCurrentStatus()

	response := StatusResponse{
		AnomalyStatus: anomalyStatus,
		SLOBudget:     sloBudget,
		Timestamp:     time.Now(),
	}

	// Include metrics if requested
	if req.IncludeMetrics {
		response.Metrics = h.radar.GetMetrics(req.MetricWindow)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetConfig returns the current configuration
func (h *HTTPHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.radar.GetConfig()
	summary := h.radar.GetConfigSummary()

	response := ConfigResponse{
		Config:    config,
		Summary:   summary,
		IsValid:   ValidateConfig(config) == nil,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateConfig updates the anomaly radar configuration
func (h *HTTPHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig Config

	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.radar.UpdateConfig(newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Configuration update failed: %v", err), http.StatusBadRequest)
		return
	}

	// Return updated configuration
	h.GetConfig(w, r)
}

// GetMetrics returns historical metrics data
func (h *HTTPHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	var req MetricsRequest

	// Parse query parameters
	if windowStr := r.URL.Query().Get("window"); windowStr != "" {
		if window, err := time.ParseDuration(windowStr); err == nil {
			req.Window = window
		} else {
			req.Window = 24 * time.Hour // Default to 24 hours
		}
	} else {
		req.Window = 24 * time.Hour
	}

	if maxSamplesStr := r.URL.Query().Get("max_samples"); maxSamplesStr != "" {
		if maxSamples, err := strconv.Atoi(maxSamplesStr); err == nil && maxSamples > 0 {
			req.MaxSamples = maxSamples
		}
	}

	// Get metrics
	metrics := h.radar.GetMetrics(req.Window)

	// Limit samples if requested
	if req.MaxSamples > 0 && len(metrics) > req.MaxSamples {
		// Sample evenly across the available metrics
		step := len(metrics) / req.MaxSamples
		if step < 1 {
			step = 1
		}

		sampledMetrics := make([]MetricSnapshot, 0, req.MaxSamples)
		for i := 0; i < len(metrics); i += step {
			sampledMetrics = append(sampledMetrics, metrics[i])
		}
		metrics = sampledMetrics
	}

	response := MetricsResponse{
		Metrics:   metrics,
		Window:    req.Window,
		Count:     len(metrics),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAlerts returns active alerts
func (h *HTTPHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	anomalyStatus, _ := h.radar.GetCurrentStatus()

	response := AlertsResponse{
		Alerts:    anomalyStatus.ActiveAlerts,
		Count:     len(anomalyStatus.ActiveAlerts),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetHealth returns the health status of the anomaly radar
func (h *HTTPHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	anomalyStatus, sloBudget := h.radar.GetCurrentStatus()

	// Determine if the radar is running
	isRunning := !anomalyStatus.LastUpdated.IsZero() &&
		time.Since(anomalyStatus.LastUpdated) < 2*h.radar.GetConfig().MonitoringInterval

	// Calculate uptime (approximation based on last update)
	var uptime time.Duration
	if isRunning {
		uptime = time.Since(anomalyStatus.LastUpdated)
	}

	response := HealthResponse{
		IsRunning:    isRunning,
		Status:       anomalyStatus.OverallStatus,
		AlertLevel:   sloBudget.AlertLevel,
		ActiveAlerts: len(anomalyStatus.ActiveAlerts),
		LastUpdated:  anomalyStatus.LastUpdated,
		Uptime:       uptime,
		Timestamp:    time.Now(),
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if !isRunning {
		statusCode = http.StatusServiceUnavailable
	} else if anomalyStatus.OverallStatus == MetricStatusCritical {
		statusCode = http.StatusServiceUnavailable
	} else if anomalyStatus.OverallStatus == MetricStatusWarning {
		statusCode = http.StatusOK // Warning is still OK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Start starts the anomaly radar monitoring
func (h *HTTPHandler) Start(w http.ResponseWriter, r *http.Request) {
	if err := h.radar.Start(context.Background()); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start anomaly radar: %v", err), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "started",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Stop stops the anomaly radar monitoring
func (h *HTTPHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.radar.Stop(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop anomaly radar: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "stopped",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetSLOBudget returns detailed SLO budget information
func (h *HTTPHandler) GetSLOBudget(w http.ResponseWriter, r *http.Request) {
	_, sloBudget := h.radar.GetCurrentStatus()

	// Calculate additional budget insights
	budgetInsights := map[string]interface{}{
		"budget_exhausted_percentage": sloBudget.BudgetUtilization * 100,
		"budget_remaining_percentage": (1.0 - sloBudget.BudgetUtilization) * 100,
		"is_budget_healthy": sloBudget.IsHealthy,
		"days_since_window_start": sloBudget.Config.Window.Hours() / 24,
	}

	// Add burn rate projections
	if sloBudget.CurrentBurnRate > 0 {
		budgetInsights["hours_to_exhaustion"] = sloBudget.TimeToExhaustion.Hours()
		budgetInsights["projected_exhaustion_date"] = time.Now().Add(sloBudget.TimeToExhaustion).Format(time.RFC3339)
	}

	response := map[string]interface{}{
		"slo_budget": sloBudget,
		"insights":   budgetInsights,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetPercentiles returns latency percentiles for a given window
func (h *HTTPHandler) GetPercentiles(w http.ResponseWriter, r *http.Request) {
	// Parse window parameter
	windowStr := r.URL.Query().Get("window")
	window := time.Hour // Default to 1 hour
	if windowStr != "" {
		if parsedWindow, err := time.ParseDuration(windowStr); err == nil {
			window = parsedWindow
		}
	}

	// Calculate percentiles
	percentiles := map[string]float64{
		"p50": h.radar.GetPercentile(0.50, window),
		"p90": h.radar.GetPercentile(0.90, window),
		"p95": h.radar.GetPercentile(0.95, window),
		"p99": h.radar.GetPercentile(0.99, window),
	}

	response := map[string]interface{}{
		"percentiles": percentiles,
		"window":      window.String(),
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all HTTP routes for the anomaly radar
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/status", h.GetStatus)
	mux.HandleFunc(prefix+"/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetConfig(w, r)
		case http.MethodPut:
			h.UpdateConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(prefix+"/metrics", h.GetMetrics)
	mux.HandleFunc(prefix+"/alerts", h.GetAlerts)
	mux.HandleFunc(prefix+"/health", h.GetHealth)
	mux.HandleFunc(prefix+"/slo-budget", h.GetSLOBudget)
	mux.HandleFunc(prefix+"/percentiles", h.GetPercentiles)
	mux.HandleFunc(prefix+"/start", h.Start)
	mux.HandleFunc(prefix+"/stop", h.Stop)
}

// SimpleMetricsCollector provides a basic implementation of MetricsCollector
type SimpleMetricsCollector struct {
	getBacklogSize  func() int64
	getRequestCount func() int64
	getErrorCount   func() int64
	getLatencies    func() (p50, p95, p99 float64)
}

// NewSimpleMetricsCollector creates a new simple metrics collector
func NewSimpleMetricsCollector(
	getBacklogSize func() int64,
	getRequestCount func() int64,
	getErrorCount func() int64,
	getLatencies func() (p50, p95, p99 float64),
) *SimpleMetricsCollector {
	return &SimpleMetricsCollector{
		getBacklogSize:  getBacklogSize,
		getRequestCount: getRequestCount,
		getErrorCount:   getErrorCount,
		getLatencies:    getLatencies,
	}
}

// CollectMetrics implements the MetricsCollector interface
func (c *SimpleMetricsCollector) CollectMetrics(ctx context.Context) (MetricSnapshot, error) {
	snapshot := MetricSnapshot{
		Timestamp: time.Now(),
	}

	// Collect basic metrics
	if c.getBacklogSize != nil {
		snapshot.BacklogSize = c.getBacklogSize()
	}

	if c.getRequestCount != nil {
		snapshot.RequestCount = c.getRequestCount()
	}

	if c.getErrorCount != nil {
		snapshot.ErrorCount = c.getErrorCount()
	}

	// Calculate error rate
	if snapshot.RequestCount > 0 {
		snapshot.ErrorRate = float64(snapshot.ErrorCount) / float64(snapshot.RequestCount)
	}

	// Collect latency metrics
	if c.getLatencies != nil {
		p50, p95, p99 := c.getLatencies()
		snapshot.P50LatencyMs = p50
		snapshot.P95LatencyMs = p95
		snapshot.P99LatencyMs = p99
	}

	return snapshot, nil
}