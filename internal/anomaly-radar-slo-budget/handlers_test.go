//go:build anomaly_radar_tests
// +build anomaly_radar_tests

package anomalyradarslobudget

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var allowAllScopes = WithScopeChecker(func(ctx context.Context, required string) bool { return true })

func contextWithScope(ctx context.Context, scope string) context.Context {
	return ContextWithScopes(ctx, []string{scope})
}

func TestHTTPHandlerStatus(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{
		{
			Timestamp:    time.Now(),
			BacklogSize:  100,
			RequestCount: 1000,
			ErrorCount:   5,
			P95LatencyMs: 200,
		},
	})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Start radar to populate some data
	if err := radar.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := radar.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	// Test status endpoint
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(contextWithScope(req.Context(), ScopeReader))
	handler.GetStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	// Test with include_metrics parameter
	req = httptest.NewRequest("GET", "/status?include_metrics=true&metric_window=1h", nil)
	req = req.WithContext(contextWithScope(req.Context(), ScopeReader))
	w = httptest.NewRecorder()

	handler.GetStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Metrics) == 0 {
		t.Error("Expected metrics to be included when requested")
	}
}

func TestHTTPHandlerConfig(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Test get config
	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.IsValid {
		t.Error("Expected config to be valid")
	}

	if response.Summary == "" {
		t.Error("Expected config summary to be provided")
	}

	// Test update config
	newConfig := config
	newConfig.Thresholds.ErrorRateWarning = 0.02 // Change warning threshold

	configJSON, _ := json.Marshal(newConfig)
	req = httptest.NewRequest("PUT", "/config", bytes.NewBuffer(configJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	req = req.WithContext(contextWithScope(req.Context(), ScopeAdmin))
	handler.UpdateConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify config was updated
	updatedConfig := radar.GetConfig()
	if updatedConfig.Thresholds.ErrorRateWarning != 0.02 {
		t.Errorf("Expected updated error rate warning threshold 0.02, got %.3f", updatedConfig.Thresholds.ErrorRateWarning)
	}

	// Test invalid config update
	invalidConfig := config
	invalidConfig.SLO.AvailabilityTarget = 1.5 // Invalid value

	configJSON, _ = json.Marshal(invalidConfig)
	req = httptest.NewRequest("PUT", "/config", bytes.NewBuffer(configJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.UpdateConfig(w, req.WithContext(contextWithScope(req.Context(), ScopeAdmin)))

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status 422 for invalid config, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	if errResp.Code != "CONFIG_INVALID" {
		t.Errorf("Expected error code CONFIG_INVALID, got %s", errResp.Code)
	}
}

func TestHTTPHandlerMetrics(t *testing.T) {
	config := DefaultConfig()

	// Create snapshots for testing
	snapshots := make([]MetricSnapshot, 10)
	now := time.Now()
	for i := 0; i < 10; i++ {
		snapshots[i] = MetricSnapshot{
			Timestamp:    now.Add(time.Duration(-i) * time.Hour),
			BacklogSize:  int64(100 + i),
			RequestCount: 1000,
			ErrorCount:   int64(i),
			P95LatencyMs: float64(200 + i*10),
		}
	}

	collector := NewMockMetricsCollector(snapshots)
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Add some metrics data
	for _, snapshot := range snapshots {
		radar.window.AddSnapshot(snapshot)
	}

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.GetMetrics(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response MetricsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Metrics) == 0 {
		t.Error("Expected metrics to be returned")
	}

	if response.Count != len(response.Metrics) {
		t.Errorf("Expected count %d to match metrics length %d", response.Count, len(response.Metrics))
	}

	// Test with window parameter
	req = httptest.NewRequest("GET", "/metrics?window=6h", nil)
	w = httptest.NewRecorder()
	handler.GetMetrics(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with max_samples parameter
	req = httptest.NewRequest("GET", "/metrics?max_samples=5", nil)
	w = httptest.NewRecorder()
	handler.GetMetrics(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Metrics) > 5 {
		t.Errorf("Expected max 5 metrics when max_samples=5, got %d", len(response.Metrics))
	}

	if response.NextCursor == "" {
		t.Error("Expected next_cursor to be provided when more metrics remain")
	}

	// Fetch next page using cursor
	req = httptest.NewRequest("GET", "/metrics?max_samples=5&next_cursor="+response.NextCursor, nil)
	w = httptest.NewRecorder()
	handler.GetMetrics(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for second page, got %d", w.Code)
	}

	var second MetricsResponse
	if err := json.NewDecoder(w.Body).Decode(&second); err != nil {
		t.Fatalf("Failed to decode second page: %v", err)
	}

	if second.NextCursor != "" && second.NextCursor == response.NextCursor {
		t.Error("Expected next_cursor to advance between pages")
	}

	// Invalid cursor should return 400
	req = httptest.NewRequest("GET", "/metrics?next_cursor=invalid", nil)
	w = httptest.NewRecorder()
	handler.GetMetrics(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid cursor, got %d", w.Code)
	}
}

func TestHTTPHandlerAlerts(t *testing.T) {
	config := DefaultConfig()
	config.Thresholds.ErrorRateWarning = 0.01

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Add snapshot that triggers alert
	alertSnapshot := MetricSnapshot{
		Timestamp:    time.Now(),
		RequestCount: 1000,
		ErrorCount:   20,   // 2% error rate - triggers warning
		ErrorRate:    0.02, // 2%
		P95LatencyMs: 300,
	}

	radar.window.AddSnapshot(alertSnapshot)
	radar.detectAnomalies()
	radar.updateAlerts()

	// Test alerts endpoint
	req := httptest.NewRequest("GET", "/alerts", nil)
	w := httptest.NewRecorder()
	handler.GetAlerts(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response AlertsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Count == 0 {
		t.Error("Expected alerts to be returned")
	}

	if len(response.Alerts) != response.Count {
		t.Errorf("Expected alerts count %d to match array length %d", response.Count, len(response.Alerts))
	}

	// Verify alert content
	found := false
	for _, alert := range response.Alerts {
		if alert.Type == AlertTypeErrorRate {
			found = true
			if alert.Severity != AlertLevelWarning {
				t.Errorf("Expected warning severity, got %s", alert.Severity.String())
			}
		}
	}

	if !found {
		t.Error("Expected to find error rate alert")
	}
}

func TestHTTPHandlerHealth(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Test health when not running
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.GetHealth(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 when not running, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.IsRunning {
		t.Error("Expected IsRunning to be false when radar is not started")
	}

	// Start radar and test health
	if err := radar.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // Allow time for metrics collection

	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	handler.GetHealth(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when running, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.IsRunning {
		t.Error("Expected IsRunning to be true when radar is started")
	}

	if err := radar.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestHTTPHandlerSLOBudget(t *testing.T) {
	config := DefaultConfig()
	config.SLO.AvailabilityTarget = 0.99

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Add snapshots to calculate budget
	now := time.Now()
	snapshots := []MetricSnapshot{
		{
			Timestamp:    now.Add(-30 * time.Minute),
			RequestCount: 1000,
			ErrorCount:   10,
			P95LatencyMs: 200,
		},
		{
			Timestamp:    now,
			RequestCount: 1000,
			ErrorCount:   5,
			P95LatencyMs: 250,
		},
	}

	for _, snapshot := range snapshots {
		radar.window.AddSnapshot(snapshot)
	}

	radar.updateSLOBudget()

	// Test SLO budget endpoint
	req := httptest.NewRequest("GET", "/slo-budget", nil)
	w := httptest.NewRecorder()
	handler.GetSLOBudget(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := response["slo_budget"]; !exists {
		t.Error("Expected slo_budget in response")
	}

	if _, exists := response["insights"]; !exists {
		t.Error("Expected insights in response")
	}

	insights := response["insights"].(map[string]interface{})
	if _, exists := insights["budget_exhausted_percentage"]; !exists {
		t.Error("Expected budget_exhausted_percentage in insights")
	}
}

func TestHTTPHandlerPercentiles(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar)

	// Add snapshots with known latency values
	latencies := []float64{100, 200, 300, 400, 500}
	now := time.Now()

	for i, latency := range latencies {
		snapshot := MetricSnapshot{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			P95LatencyMs: latency,
		}
		radar.window.AddSnapshot(snapshot)
	}

	// Test percentiles endpoint
	req := httptest.NewRequest("GET", "/percentiles", nil)
	w := httptest.NewRecorder()

	handler.GetPercentiles(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	percentiles := response["percentiles"].(map[string]interface{})
	if len(percentiles) == 0 {
		t.Error("Expected percentiles to be calculated")
	}

	expectedPercentiles := []string{"p50", "p90", "p95", "p99"}
	for _, p := range expectedPercentiles {
		if _, exists := percentiles[p]; !exists {
			t.Errorf("Expected percentile %s to be present", p)
		}
	}

	// Test with custom window
	req = httptest.NewRequest("GET", "/percentiles?window=30m", nil)
	w = httptest.NewRecorder()
	handler.GetPercentiles(w, req.WithContext(contextWithScope(req.Context(), ScopeReader)))

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandlerStartStop(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, allowAllScopes)

	// Test start
	req := httptest.NewRequest(http.MethodPost, "/start", nil)
	w := httptest.NewRecorder()
	handler.Start(w, req.WithContext(contextWithScope(req.Context(), ScopeAdmin)))

	if w.Code != http.StatusAccepted {
		t.Fatalf("Expected status 202 for start, got %d", w.Code)
	}

	var startResp StartStopResponse
	if err := json.NewDecoder(w.Body).Decode(&startResp); err != nil {
		t.Fatalf("Failed to decode start response: %v", err)
	}
	if startResp.Status != "started" {
		t.Errorf("Expected status 'started', got '%s'", startResp.Status)
	}

	// Starting again should be idempotent
	req = httptest.NewRequest(http.MethodPost, "/start", nil)
	w = httptest.NewRecorder()
	handler.Start(w, req.WithContext(contextWithScope(req.Context(), ScopeAdmin)))

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for already started, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&startResp); err != nil {
		t.Fatalf("Failed to decode second start response: %v", err)
	}
	if startResp.Status != "already_started" {
		t.Errorf("Expected status 'already_started', got '%s'", startResp.Status)
	}

	// Stop should succeed
	req = httptest.NewRequest(http.MethodPost, "/stop", nil)
	w = httptest.NewRecorder()
	handler.Stop(w, req.WithContext(contextWithScope(req.Context(), ScopeAdmin)))

	if w.Code != http.StatusAccepted {
		t.Fatalf("Expected status 202 for stop, got %d", w.Code)
	}
	var stopResp StartStopResponse
	if err := json.NewDecoder(w.Body).Decode(&stopResp); err != nil {
		t.Fatalf("Failed to decode stop response: %v", err)
	}
	if stopResp.Status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", stopResp.Status)
	}

	// Second stop should be idempotent
	req = httptest.NewRequest(http.MethodPost, "/stop", nil)
	w = httptest.NewRecorder()
	handler.Stop(w, req.WithContext(contextWithScope(req.Context(), ScopeAdmin)))

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for already stopped, got %d", w.Code)
	}
	if err := json.NewDecoder(w.Body).Decode(&stopResp); err != nil {
		t.Fatalf("Failed to decode second stop response: %v", err)
	}
	if stopResp.Status != "already_stopped" {
		t.Errorf("Expected status 'already_stopped', got '%s'", stopResp.Status)
	}
}

func TestHTTPHandlerScopeDenied(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar, WithScopeChecker(func(ctx context.Context, required string) bool { return false }))

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()

	handler.GetStatus(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403 when scope missing, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	if errResp.Code != "ACCESS_DENIED" {
		t.Errorf("Expected error code ACCESS_DENIED, got %s", errResp.Code)
	}
}

func TestSimpleMetricsCollector(t *testing.T) {
	// Test metrics collector with all functions provided
	collector := NewSimpleMetricsCollector(
		func() int64 { return 100 },
		func() int64 { return 1000 },
		func() int64 { return 10 },
		func() (float64, float64, float64) { return 100, 200, 300 },
	)

	snapshot, err := collector.CollectMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if snapshot.BacklogSize != 100 {
		t.Errorf("Expected backlog size 100, got %d", snapshot.BacklogSize)
	}

	if snapshot.RequestCount != 1000 {
		t.Errorf("Expected request count 1000, got %d", snapshot.RequestCount)
	}

	if snapshot.ErrorCount != 10 {
		t.Errorf("Expected error count 10, got %d", snapshot.ErrorCount)
	}

	if snapshot.ErrorRate != 0.01 {
		t.Errorf("Expected error rate 0.01, got %.3f", snapshot.ErrorRate)
	}

	if snapshot.P50LatencyMs != 100 {
		t.Errorf("Expected P50 latency 100, got %.2f", snapshot.P50LatencyMs)
	}

	if snapshot.P95LatencyMs != 200 {
		t.Errorf("Expected P95 latency 200, got %.2f", snapshot.P95LatencyMs)
	}

	if snapshot.P99LatencyMs != 300 {
		t.Errorf("Expected P99 latency 300, got %.2f", snapshot.P99LatencyMs)
	}

	// Test with nil functions (should not panic)
	nilCollector := NewSimpleMetricsCollector(nil, nil, nil, nil)
	snapshot, err = nilCollector.CollectMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error with nil functions, got %v", err)
	}

	if snapshot.BacklogSize != 0 {
		t.Errorf("Expected backlog size 0 with nil function, got %d", snapshot.BacklogSize)
	}
}
