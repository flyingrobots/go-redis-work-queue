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
	handler := NewHTTPHandler(radar)

	// Start radar to populate some data
	radar.Start(context.Background())
	time.Sleep(200 * time.Millisecond)
	radar.Stop()

	// Test status endpoint
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

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
	handler := NewHTTPHandler(radar)

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

	handler.UpdateConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid config, got %d", w.Code)
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
	handler := NewHTTPHandler(radar)

	// Add some metrics data
	for _, snapshot := range snapshots {
		radar.window.AddSnapshot(snapshot)
	}

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.GetMetrics(w, req)

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

	handler.GetMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with max_samples parameter
	req = httptest.NewRequest("GET", "/metrics?max_samples=5", nil)
	w = httptest.NewRecorder()

	handler.GetMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Metrics) > 5 {
		t.Errorf("Expected max 5 metrics when max_samples=5, got %d", len(response.Metrics))
	}
}

func TestHTTPHandlerAlerts(t *testing.T) {
	config := DefaultConfig()
	config.Thresholds.ErrorRateWarning = 0.01

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar)

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

	handler.GetAlerts(w, req)

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
	handler := NewHTTPHandler(radar)

	// Test health when not running
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealth(w, req)

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
	radar.Start(context.Background())
	time.Sleep(200 * time.Millisecond) // Allow time for metrics collection

	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()

	handler.GetHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when running, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.IsRunning {
		t.Error("Expected IsRunning to be true when radar is started")
	}

	radar.Stop()
}

func TestHTTPHandlerSLOBudget(t *testing.T) {
	config := DefaultConfig()
	config.SLO.AvailabilityTarget = 0.99

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar)

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

	handler.GetSLOBudget(w, req)

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

	handler.GetPercentiles(w, req)

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

	handler.GetPercentiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandlerStartStop(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)
	handler := NewHTTPHandler(radar)

	// Test start
	req := httptest.NewRequest("POST", "/start", nil)
	w := httptest.NewRecorder()

	handler.Start(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for start, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode start response: %v", err)
	}

	if response["status"] != "started" {
		t.Errorf("Expected status 'started', got '%s'", response["status"])
	}

	// Test start when already running (should fail)
	req = httptest.NewRequest("POST", "/start", nil)
	w = httptest.NewRecorder()

	handler.Start(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 when already running, got %d", w.Code)
	}

	// Test stop
	req = httptest.NewRequest("POST", "/stop", nil)
	w = httptest.NewRecorder()

	handler.Stop(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for stop, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode stop response: %v", err)
	}

	if response["status"] != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", response["status"])
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