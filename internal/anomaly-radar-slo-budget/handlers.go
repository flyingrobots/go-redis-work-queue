package anomalyradarslobudget

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultMaxSamples = 1000
	maxMaxSamples     = 5000
)

// HTTPHandler provides HTTP endpoints for anomaly radar management
type HTTPHandler struct {
	radar        *AnomalyRadar
	scopeChecker ScopeChecker
	now          nowFunc
}

// NewHTTPHandler creates a new HTTP handler for anomaly radar
func NewHTTPHandler(radar *AnomalyRadar, opts ...HandlerOption) *HTTPHandler {
	h := &HTTPHandler{
		radar: radar,
		now:   time.Now,
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.scopeChecker == nil {
		h.scopeChecker = func(ctx context.Context, required string) bool {
			if required == "" {
				return true
			}
			return hasScope(scopesFromContext(ctx), required)
		}
	}
	return h
}

// StatusRequest represents a request for current status
type StatusRequest struct {
	IncludeMetrics bool          `json:"include_metrics"`
	MetricWindow   time.Duration `json:"metric_window"`
}

// StatusResponse represents the response with current status
type StatusResponse struct {
	AnomalyStatus AnomalyStatus    `json:"anomaly_status"`
	SLOBudget     SLOBudget        `json:"slo_budget"`
	Metrics       []MetricSnapshot `json:"metrics,omitempty"`
	Timestamp     time.Time        `json:"timestamp"`
}

// ConfigResponse represents configuration information
type ConfigResponse struct {
	Config    Config    `json:"config"`
	Summary   string    `json:"summary"`
	IsValid   bool      `json:"is_valid"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricsRequest represents a request for historical metrics
type MetricsRequest struct {
	Window     time.Duration `json:"window"`
	MaxSamples int           `json:"max_samples"`
	Cursor     string        `json:"next_cursor"`
}

// MetricsResponse represents historical metrics data
type MetricsResponse struct {
	Metrics    []MetricSnapshot `json:"metrics"`
	Window     time.Duration    `json:"window"`
	Count      int              `json:"count"`
	Timestamp  time.Time        `json:"timestamp"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

// AlertsResponse represents active alerts
type AlertsResponse struct {
	Alerts    []Alert   `json:"alerts"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	IsRunning    bool          `json:"is_running"`
	Status       MetricStatus  `json:"status"`
	AlertLevel   AlertLevel    `json:"alert_level"`
	ActiveAlerts int           `json:"active_alerts"`
	LastUpdated  time.Time     `json:"last_updated"`
	Uptime       time.Duration `json:"uptime"`
	Timestamp    time.Time     `json:"timestamp"`
}

type StartStopResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *HTTPHandler) nowTime() time.Time {
	if h.now != nil {
		return h.now().UTC()
	}
	return time.Now().UTC()
}

func (h *HTTPHandler) requireScope(w http.ResponseWriter, r *http.Request, scope string) bool {
	if scope == "" {
		return true
	}
	if h.scopeChecker == nil {
		return true
	}
	if h.scopeChecker(r.Context(), scope) {
		return true
	}
	writeJSONError(w, r, http.StatusForbidden, "ACCESS_DENIED", fmt.Sprintf("required scope '%s' not granted", scope), nil)
	return false
}

func startStopPayload(status string, ts time.Time) StartStopResponse {
	return StartStopResponse{Status: status, Timestamp: ts}
}

// GetStatus returns the current anomaly and SLO status
func (h *HTTPHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
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
		Timestamp:     h.nowTime(),
	}

	// Include metrics if requested
	if req.IncludeMetrics {
		metrics, _ := h.radar.GetMetricsPage(req.MetricWindow, 0, 0)
		response.Metrics = metrics
	}

	writeJSON(w, http.StatusOK, response)
}

// GetConfig returns the current configuration
func (h *HTTPHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
	config := h.radar.GetConfig()
	summary := h.radar.GetConfigSummary()

	response := ConfigResponse{
		Config:    config,
		Summary:   summary,
		IsValid:   ValidateConfig(config) == nil,
		Timestamp: h.nowTime(),
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateConfig updates the anomaly radar configuration
func (h *HTTPHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeAdmin) {
		return
	}
	var newConfig Config

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&newConfig); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("invalid JSON: %v", err), nil)
		return
	}

	if err := h.radar.UpdateConfig(newConfig); err != nil {
		writeJSONError(w, r, http.StatusUnprocessableEntity, "CONFIG_INVALID", err.Error(), nil)
		return
	}

	config := h.radar.GetConfig()
	summary := h.radar.GetConfigSummary()
	response := ConfigResponse{
		Config:    config,
		Summary:   summary,
		IsValid:   true,
		Timestamp: h.nowTime(),
	}
	writeJSON(w, http.StatusOK, response)
}

// GetMetrics returns historical metrics data
func (h *HTTPHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
	var req MetricsRequest

	// Parse window parameter
	if windowStr := r.URL.Query().Get("window"); windowStr != "" {
		window, err := time.ParseDuration(windowStr)
		if err != nil {
			writeJSONError(w, r, http.StatusBadRequest, "INVALID_WINDOW", "window must use Go duration syntax", nil)
			return
		}
		req.Window = window
	} else {
		req.Window = 24 * time.Hour
	}

	if maxSamplesStr := r.URL.Query().Get("max_samples"); maxSamplesStr != "" {
		maxSamples, err := strconv.Atoi(maxSamplesStr)
		if err != nil || maxSamples <= 0 {
			writeJSONError(w, r, http.StatusBadRequest, "INVALID_MAX_SAMPLES", "max_samples must be a positive integer", nil)
			return
		}
		req.MaxSamples = maxSamples
	}

	req.Cursor = r.URL.Query().Get("next_cursor")
	cursorIdx, err := decodeCursor(req.Cursor)
	if err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "INVALID_CURSOR", "next_cursor must be a non-negative integer", nil)
		return
	}

	limit := req.MaxSamples
	if limit <= 0 {
		limit = defaultMaxSamples
	}
	if limit > maxMaxSamples {
		limit = maxMaxSamples
	}

	metrics, nextIndex := h.radar.GetMetricsPage(req.Window, limit, cursorIdx)
	nextCursor := ""
	if nextIndex >= 0 {
		nextCursor = encodeCursor(nextIndex)
	}

	response := MetricsResponse{
		Metrics:    metrics,
		Window:     req.Window,
		Count:      len(metrics),
		Timestamp:  h.nowTime(),
		NextCursor: nextCursor,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetAlerts returns active alerts
func (h *HTTPHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
	anomalyStatus, _ := h.radar.GetCurrentStatus()

	response := AlertsResponse{
		Alerts:    anomalyStatus.ActiveAlerts,
		Count:     len(anomalyStatus.ActiveAlerts),
		Timestamp: h.nowTime(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetHealth returns the health status of the anomaly radar
func (h *HTTPHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
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
		Timestamp:    h.nowTime(),
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

	writeJSON(w, statusCode, response)
}

// Start starts the anomaly radar monitoring
func (h *HTTPHandler) Start(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeAdmin) {
		return
	}

	if err := h.radar.Start(r.Context()); err != nil {
		if errors.Is(err, ErrAlreadyRunning) {
			writeJSON(w, http.StatusOK, startStopPayload("already_started", h.nowTime()))
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "START_FAILED", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusAccepted, startStopPayload("started", h.nowTime()))
}

// Stop stops the anomaly radar monitoring
func (h *HTTPHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeAdmin) {
		return
	}

	if err := h.radar.Stop(); err != nil {
		if errors.Is(err, ErrNotRunning) {
			writeJSON(w, http.StatusOK, startStopPayload("already_stopped", h.nowTime()))
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "STOP_FAILED", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusAccepted, startStopPayload("stopped", h.nowTime()))
}

// GetSLOBudget returns detailed SLO budget information
func (h *HTTPHandler) GetSLOBudget(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
	_, sloBudget := h.radar.GetCurrentStatus()

	// Calculate additional budget insights
	budgetInsights := map[string]interface{}{
		"budget_exhausted_percentage": sloBudget.BudgetUtilization * 100,
		"budget_remaining_percentage": (1.0 - sloBudget.BudgetUtilization) * 100,
		"is_budget_healthy":           sloBudget.IsHealthy,
		"days_since_window_start":     sloBudget.Config.Window.Hours() / 24,
	}

	// Add burn rate projections
	if sloBudget.CurrentBurnRate > 0 {
		budgetInsights["hours_to_exhaustion"] = sloBudget.TimeToExhaustion.Hours()
		budgetInsights["projected_exhaustion_date"] = time.Now().Add(sloBudget.TimeToExhaustion).Format(time.RFC3339)
	}

	response := map[string]interface{}{
		"slo_budget": sloBudget,
		"insights":   budgetInsights,
		"timestamp":  h.nowTime(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetPercentiles returns latency percentiles for a given window
func (h *HTTPHandler) GetPercentiles(w http.ResponseWriter, r *http.Request) {
	if !h.requireScope(w, r, ScopeReader) {
		return
	}
	// Parse window parameter
	windowStr := r.URL.Query().Get("window")
	window := time.Hour // Default to 1 hour
	if windowStr != "" {
		if parsedWindow, err := time.ParseDuration(windowStr); err == nil {
			window = parsedWindow
		} else {
			writeJSONError(w, r, http.StatusBadRequest, "INVALID_WINDOW", "window must use Go duration syntax", nil)
			return
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
		"timestamp":   h.nowTime(),
	}

	writeJSON(w, http.StatusOK, response)
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
