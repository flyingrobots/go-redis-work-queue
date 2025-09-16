package anomalyradarslobudget

import (
	"context"
	"testing"
	"time"
)

// MockMetricsCollector for testing
type MockMetricsCollector struct {
	snapshots []MetricSnapshot
	index     int
	errorRate float64
}

func NewMockMetricsCollector(snapshots []MetricSnapshot) *MockMetricsCollector {
	return &MockMetricsCollector{
		snapshots: snapshots,
		index:     0,
	}
}

func (m *MockMetricsCollector) CollectMetrics(ctx context.Context) (MetricSnapshot, error) {
	if m.index >= len(m.snapshots) {
		// Return the last snapshot for continued operation
		return m.snapshots[len(m.snapshots)-1], nil
	}

	snapshot := m.snapshots[m.index]
	m.index++
	return snapshot, nil
}

func TestAnomalyRadarBasicOperation(t *testing.T) {
	config := DefaultConfig()
	config.MonitoringInterval = 100 * time.Millisecond
	config.MaxSnapshots = 10

	// Create test snapshots with increasing error rates
	snapshots := []MetricSnapshot{
		{
			Timestamp:    time.Now(),
			BacklogSize:  100,
			RequestCount: 1000,
			ErrorCount:   5,
			P95LatencyMs: 200,
		},
		{
			Timestamp:    time.Now().Add(time.Second),
			BacklogSize:  120,
			RequestCount: 1000,
			ErrorCount:   10,
			P95LatencyMs: 250,
		},
		{
			Timestamp:    time.Now().Add(2 * time.Second),
			BacklogSize:  150,
			RequestCount: 1000,
			ErrorCount:   60, // Triggers critical error rate
			P95LatencyMs: 300,
		},
	}

	collector := NewMockMetricsCollector(snapshots)
	radar := New(config, collector)

	// Test starting the radar
	err := radar.Start(context.Background())
	if err != nil {
		t.Fatalf("Failed to start radar: %v", err)
	}

	// Wait for a few monitoring cycles
	time.Sleep(500 * time.Millisecond)

	// Get current status
	anomalyStatus, sloBudget := radar.GetCurrentStatus()

	// Verify we have collected metrics
	if len(radar.GetMetrics(time.Hour)) == 0 {
		t.Error("Expected metrics to be collected")
	}

	// Verify SLO budget calculations
	if sloBudget.TotalBudget <= 0 {
		t.Error("Expected positive total budget")
	}

	if sloBudget.ConsumedBudget <= 0 {
		t.Error("Expected some budget consumption")
	}

	// Verify anomaly detection is working
	if anomalyStatus.LastUpdated.IsZero() {
		t.Error("Expected anomaly status to be updated")
	}

	// Stop the radar
	err = radar.Stop()
	if err != nil {
		t.Fatalf("Failed to stop radar: %v", err)
	}
}

func TestRollingWindowOperations(t *testing.T) {
	window := NewRollingWindow(time.Hour, 5)

	// Test adding snapshots
	now := time.Now()
	for i := 0; i < 10; i++ {
		snapshot := MetricSnapshot{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			BacklogSize:  int64(100 + i*10),
			RequestCount: 1000,
			ErrorCount:   int64(i),
		}
		window.AddSnapshot(snapshot)
	}

	// Should be limited to max snapshots
	if len(window.Snapshots) != 5 {
		t.Errorf("Expected 5 snapshots, got %d", len(window.Snapshots))
	}

	// Should contain the most recent snapshots
	if window.Snapshots[0].BacklogSize != 150 { // Snapshot index 5
		t.Errorf("Expected oldest snapshot to have backlog 150, got %d", window.Snapshots[0].BacklogSize)
	}

	if window.Snapshots[4].BacklogSize != 190 { // Snapshot index 9
		t.Errorf("Expected newest snapshot to have backlog 190, got %d", window.Snapshots[4].BacklogSize)
	}
}

func TestSLOBudgetCalculation(t *testing.T) {
	config := DefaultConfig()
	config.SLO.AvailabilityTarget = 0.99 // 99% availability
	config.SLO.Window = time.Hour

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Manually add snapshots to test budget calculation
	now := time.Now()
	snapshots := []MetricSnapshot{
		{
			Timestamp:    now.Add(-30 * time.Minute),
			RequestCount: 1000,
			ErrorCount:   5, // 0.5% error rate
			P95LatencyMs: 200,
		},
		{
			Timestamp:    now.Add(-15 * time.Minute),
			RequestCount: 1000,
			ErrorCount:   10, // 1% error rate
			P95LatencyMs: 300,
		},
		{
			Timestamp:    now,
			RequestCount: 1000,
			ErrorCount:   15, // 1.5% error rate
			P95LatencyMs: 250,
		},
	}

	for _, snapshot := range snapshots {
		radar.window.AddSnapshot(snapshot)
	}

	// Manually trigger SLO budget calculation
	radar.updateSLOBudget()

	// Verify budget calculations
	totalRequests := 3000.0
	expectedTotalBudget := totalRequests * (1.0 - config.SLO.AvailabilityTarget) // 1% of 3000 = 30
	expectedConsumedBudget := 30.0                                                // Total errors

	if radar.budget.TotalBudget != expectedTotalBudget {
		t.Errorf("Expected total budget %.2f, got %.2f", expectedTotalBudget, radar.budget.TotalBudget)
	}

	if radar.budget.ConsumedBudget != expectedConsumedBudget {
		t.Errorf("Expected consumed budget %.2f, got %.2f", expectedConsumedBudget, radar.budget.ConsumedBudget)
	}

	expectedUtilization := expectedConsumedBudget / expectedTotalBudget // Should be 1.0 (100%)
	if radar.budget.BudgetUtilization != expectedUtilization {
		t.Errorf("Expected budget utilization %.2f, got %.2f", expectedUtilization, radar.budget.BudgetUtilization)
	}

	// Budget should be exhausted
	if radar.budget.IsHealthy {
		t.Error("Expected budget to be unhealthy when fully consumed")
	}
}

func TestAnomalyDetection(t *testing.T) {
	config := DefaultConfig()
	config.Thresholds.ErrorRateWarning = 0.01   // 1%
	config.Thresholds.ErrorRateCritical = 0.05  // 5%
	config.Thresholds.LatencyP95Warning = 500   // 500ms
	config.Thresholds.LatencyP95Critical = 1000 // 1s

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Test normal conditions
	normalSnapshot := MetricSnapshot{
		Timestamp:         time.Now(),
		RequestCount:      1000,
		ErrorCount:        5,    // 0.5% error rate
		ErrorRate:         0.005, // 0.5%
		P95LatencyMs:      300,   // Below warning threshold
		BacklogGrowthRate: 5,     // Below warning threshold
	}

	radar.window.AddSnapshot(normalSnapshot)
	radar.detectAnomalies()

	if radar.anomalies.OverallStatus != MetricStatusHealthy {
		t.Errorf("Expected healthy status for normal conditions, got %s", radar.anomalies.OverallStatus.String())
	}

	// Test warning conditions
	warningSnapshot := MetricSnapshot{
		Timestamp:         time.Now(),
		RequestCount:      1000,
		ErrorCount:        15,   // 1.5% error rate
		ErrorRate:         0.015, // 1.5%
		P95LatencyMs:      600,   // Above warning threshold
		BacklogGrowthRate: 15,    // Above warning threshold
	}

	radar.window.AddSnapshot(warningSnapshot)
	radar.detectAnomalies()

	if radar.anomalies.ErrorRateStatus != MetricStatusWarning {
		t.Errorf("Expected warning status for error rate, got %s", radar.anomalies.ErrorRateStatus.String())
	}

	if radar.anomalies.LatencyStatus != MetricStatusWarning {
		t.Errorf("Expected warning status for latency, got %s", radar.anomalies.LatencyStatus.String())
	}

	// Test critical conditions
	criticalSnapshot := MetricSnapshot{
		Timestamp:         time.Now(),
		RequestCount:      1000,
		ErrorCount:        60,   // 6% error rate
		ErrorRate:         0.06, // 6%
		P95LatencyMs:      1200, // Above critical threshold
		BacklogGrowthRate: 60,   // Above critical threshold
	}

	radar.window.AddSnapshot(criticalSnapshot)
	radar.detectAnomalies()

	if radar.anomalies.ErrorRateStatus != MetricStatusCritical {
		t.Errorf("Expected critical status for error rate, got %s", radar.anomalies.ErrorRateStatus.String())
	}

	if radar.anomalies.LatencyStatus != MetricStatusCritical {
		t.Errorf("Expected critical status for latency, got %s", radar.anomalies.LatencyStatus.String())
	}

	if radar.anomalies.OverallStatus != MetricStatusCritical {
		t.Errorf("Expected critical overall status, got %s", radar.anomalies.OverallStatus.String())
	}
}

func TestAlertGeneration(t *testing.T) {
	config := DefaultConfig()
	config.Thresholds.ErrorRateWarning = 0.01
	config.Thresholds.ErrorRateCritical = 0.05

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Track alerts received
	alertsReceived := make([]Alert, 0)
	radar.RegisterAlertCallback(func(alert Alert) {
		alertsReceived = append(alertsReceived, alert)
	})

	// Add snapshot that triggers an alert
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

	// Should have generated an alert
	if len(radar.anomalies.ActiveAlerts) == 0 {
		t.Error("Expected active alerts to be generated")
	}

	// Check that alert callback was called
	if len(alertsReceived) == 0 {
		t.Error("Expected alert callback to be called")
	}

	// Verify alert details
	found := false
	for _, alert := range radar.anomalies.ActiveAlerts {
		if alert.Type == AlertTypeErrorRate {
			found = true
			if alert.Severity != AlertLevelWarning {
				t.Errorf("Expected warning alert severity, got %s", alert.Severity.String())
			}
			if alert.Value != 0.02 {
				t.Errorf("Expected alert value 0.02, got %.3f", alert.Value)
			}
		}
	}

	if !found {
		t.Error("Expected to find error rate alert")
	}
}

func TestBurnRateCalculation(t *testing.T) {
	config := DefaultConfig()
	config.SLO.AvailabilityTarget = 0.99 // 99%
	config.SLO.BurnRateThresholds.FastBurnRate = 0.01
	config.SLO.BurnRateThresholds.SlowBurnRate = 0.05

	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Add snapshots over time to simulate burn rate
	now := time.Now()
	baseTime := now.Add(-2 * time.Hour)

	for i := 0; i < 120; i++ { // 2 hours of data at 1-minute intervals
		errorCount := int64(10) // Consistent 1% error rate
		if i >= 60 {           // Last hour has higher error rate
			errorCount = 30 // 3% error rate
		}

		snapshot := MetricSnapshot{
			Timestamp:    baseTime.Add(time.Duration(i) * time.Minute),
			RequestCount: 1000,
			ErrorCount:   errorCount,
			P95LatencyMs: 200,
		}

		radar.window.AddSnapshot(snapshot)
	}

	// Calculate burn rate
	radar.updateSLOBudget()

	// Should detect elevated burn rate in the last hour
	if radar.budget.CurrentBurnRate == 0 {
		t.Error("Expected non-zero burn rate")
	}

	// Should trigger appropriate alert level
	if radar.budget.AlertLevel == AlertLevelNone {
		t.Error("Expected burn rate alert to be triggered")
	}

	// Time to exhaustion should be calculated
	if radar.budget.TimeToExhaustion == 0 {
		t.Error("Expected time to exhaustion to be calculated")
	}
}

func TestConfigurationValidation(t *testing.T) {
	// Test valid configuration
	validConfig := DefaultConfig()
	if err := ValidateConfig(validConfig); err != nil {
		t.Errorf("Expected valid default config, got error: %v", err)
	}

	// Test invalid availability target
	invalidConfig := validConfig
	invalidConfig.SLO.AvailabilityTarget = 1.5 // > 1.0
	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid availability target")
	}

	// Test invalid error rate thresholds
	invalidConfig = validConfig
	invalidConfig.Thresholds.ErrorRateCritical = invalidConfig.Thresholds.ErrorRateWarning
	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid error rate thresholds")
	}

	// Test invalid monitoring interval
	invalidConfig = validConfig
	invalidConfig.MonitoringInterval = -time.Second
	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Expected error for negative monitoring interval")
	}
}

func TestPercentileCalculation(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Add snapshots with known latency values
	latencies := []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
	now := time.Now()

	for i, latency := range latencies {
		snapshot := MetricSnapshot{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			P95LatencyMs: latency,
		}
		radar.window.AddSnapshot(snapshot)
	}

	// Test percentile calculations
	p50 := radar.GetPercentile(0.50, time.Hour)
	p95 := radar.GetPercentile(0.95, time.Hour)
	p99 := radar.GetPercentile(0.99, time.Hour)

	// With 10 values, p50 should be around 550 (average of 5th and 6th values)
	if p50 < 500 || p50 > 600 {
		t.Errorf("Expected p50 around 550, got %.2f", p50)
	}

	// p95 should be the 9th value (900) or 10th value (1000)
	if p95 < 900 {
		t.Errorf("Expected p95 >= 900, got %.2f", p95)
	}

	// p99 should be the highest value (1000)
	if p99 != 1000 {
		t.Errorf("Expected p99 = 1000, got %.2f", p99)
	}
}

func TestMetricsRetrieval(t *testing.T) {
	config := DefaultConfig()
	collector := NewMockMetricsCollector([]MetricSnapshot{})
	radar := New(config, collector)

	// Add snapshots across different time periods
	now := time.Now()
	for i := 0; i < 10; i++ {
		snapshot := MetricSnapshot{
			Timestamp:    now.Add(time.Duration(-i) * time.Hour),
			BacklogSize:  int64(100 + i),
			RequestCount: 1000,
		}
		radar.window.AddSnapshot(snapshot)
	}

	// Test retrieving metrics for different windows
	metrics1h := radar.GetMetrics(time.Hour)
	metrics6h := radar.GetMetrics(6 * time.Hour)
	metrics24h := radar.GetMetrics(24 * time.Hour)

	// Should have different counts based on window
	if len(metrics1h) == 0 {
		t.Error("Expected metrics within 1 hour window")
	}

	if len(metrics6h) <= len(metrics1h) {
		t.Error("Expected more metrics in 6 hour window than 1 hour window")
	}

	if len(metrics24h) != len(metrics6h) {
		// In this test, all metrics are within 10 hours, so 24h should equal 6h+ window
		if len(metrics24h) < len(metrics6h) {
			t.Error("Expected at least as many metrics in 24 hour window")
		}
	}
}