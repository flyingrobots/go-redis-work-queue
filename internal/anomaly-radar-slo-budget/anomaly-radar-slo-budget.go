package anomalyradarslobudget

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// AnomalyRadar monitors queue health and tracks SLO budgets
type AnomalyRadar struct {
	config Config
	window *RollingWindow
	budget *SLOBudget
	anomalies *AnomalyStatus
	alerts map[string]*Alert

	// Metrics collector interface
	metricsCollector MetricsCollector

	// State management
	mu sync.RWMutex
	running bool
	stopCh chan struct{}
	wg sync.WaitGroup

	// Alert callbacks
	alertCallbacks []AlertCallback
}

// MetricsCollector interface for gathering system metrics
type MetricsCollector interface {
	CollectMetrics(ctx context.Context) (MetricSnapshot, error)
}

// AlertCallback is called when alerts are triggered or resolved
type AlertCallback func(alert Alert)

// New creates a new AnomalyRadar instance
func New(config Config, collector MetricsCollector) *AnomalyRadar {
	return &AnomalyRadar{
		config: config,
		window: NewRollingWindow(config.MetricRetention, config.MaxSnapshots),
		budget: &SLOBudget{
			Config: config.SLO,
		},
		anomalies: &AnomalyStatus{
			ActiveAlerts: make([]Alert, 0),
		},
		alerts: make(map[string]*Alert),
		metricsCollector: collector,
		stopCh: make(chan struct{}),
		alertCallbacks: make([]AlertCallback, 0),
	}
}

// NewRollingWindow creates a new rolling window for metrics
func NewRollingWindow(retention time.Duration, maxSnapshots int) *RollingWindow {
	return &RollingWindow{
		WindowSize: retention,
		Snapshots: make([]MetricSnapshot, 0, maxSnapshots),
		maxSnapshots: maxSnapshots,
	}
}

// Start begins monitoring and SLO budget tracking
func (ar *AnomalyRadar) Start(ctx context.Context) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if ar.running {
		return fmt.Errorf("anomaly radar is already running")
	}

	ar.running = true
	ar.wg.Add(1)

	go ar.monitoringLoop(ctx)

	return nil
}

// Stop gracefully shuts down the anomaly radar
func (ar *AnomalyRadar) Stop() error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if !ar.running {
		return nil
	}

	ar.running = false
	close(ar.stopCh)
	ar.wg.Wait()

	return nil
}

// RegisterAlertCallback adds a callback for alert notifications
func (ar *AnomalyRadar) RegisterAlertCallback(callback AlertCallback) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.alertCallbacks = append(ar.alertCallbacks, callback)
}

// GetCurrentStatus returns the current anomaly and SLO status
func (ar *AnomalyRadar) GetCurrentStatus() (AnomalyStatus, SLOBudget) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	return *ar.anomalies, *ar.budget
}

// GetMetrics returns recent metric snapshots
func (ar *AnomalyRadar) GetMetrics(window time.Duration) []MetricSnapshot {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	metrics := make([]MetricSnapshot, 0)

	for _, snapshot := range ar.window.Snapshots {
		if snapshot.Timestamp.After(cutoff) {
			metrics = append(metrics, snapshot)
		}
	}

	return metrics
}

// monitoringLoop runs the main monitoring and analysis cycle
func (ar *AnomalyRadar) monitoringLoop(ctx context.Context) {
	defer ar.wg.Done()

	ticker := time.NewTicker(ar.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ar.stopCh:
			return
		case <-ticker.C:
			if err := ar.collectAndAnalyze(ctx); err != nil {
				// Log error but continue monitoring
				// In a real implementation, this would use a proper logger
				continue
			}
		}
	}
}

// collectAndAnalyze performs one cycle of metric collection and analysis
func (ar *AnomalyRadar) collectAndAnalyze(ctx context.Context) error {
	// Collect current metrics
	snapshot, err := ar.metricsCollector.CollectMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Add to rolling window
	ar.window.AddSnapshot(snapshot)

	// Calculate derived metrics
	ar.calculateDerivedMetrics()

	// Update SLO budget
	ar.updateSLOBudget()

	// Detect anomalies
	ar.detectAnomalies()

	// Update alert status
	ar.updateAlerts()

	return nil
}

// AddSnapshot adds a metric snapshot to the rolling window
func (rw *RollingWindow) AddSnapshot(snapshot MetricSnapshot) {
	// Add timestamp if not set
	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now()
	}

	// Add to snapshots
	rw.Snapshots = append(rw.Snapshots, snapshot)

	// Trim old snapshots
	cutoff := time.Now().Add(-rw.WindowSize)
	validSnapshots := make([]MetricSnapshot, 0, len(rw.Snapshots))

	for _, s := range rw.Snapshots {
		if s.Timestamp.After(cutoff) {
			validSnapshots = append(validSnapshots, s)
		}
	}

	rw.Snapshots = validSnapshots

	// Enforce max snapshots limit
	if len(rw.Snapshots) > rw.maxSnapshots {
		excess := len(rw.Snapshots) - rw.maxSnapshots
		rw.Snapshots = rw.Snapshots[excess:]
	}
}

// calculateDerivedMetrics computes rates and trends from raw metrics
func (ar *AnomalyRadar) calculateDerivedMetrics() {
	if len(ar.window.Snapshots) < 2 {
		return
	}

	// Get recent snapshots for rate calculation
	recent := ar.window.Snapshots[len(ar.window.Snapshots)-1]
	previous := ar.window.Snapshots[len(ar.window.Snapshots)-2]

	// Calculate time delta
	timeDelta := recent.Timestamp.Sub(previous.Timestamp).Seconds()
	if timeDelta <= 0 {
		return
	}

	// Calculate backlog growth rate
	backlogDelta := recent.BacklogSize - previous.BacklogSize
	recent.BacklogGrowthRate = float64(backlogDelta) / timeDelta

	// Calculate error rate
	if recent.RequestCount > 0 {
		recent.ErrorRate = float64(recent.ErrorCount) / float64(recent.RequestCount)
	}

	// Update the most recent snapshot
	ar.window.Snapshots[len(ar.window.Snapshots)-1] = recent
}

// updateSLOBudget calculates current SLO budget consumption
func (ar *AnomalyRadar) updateSLOBudget() {
	now := time.Now()
	windowStart := now.Add(-ar.config.SLO.Window)

	// Get snapshots within SLO window
	windowSnapshots := make([]MetricSnapshot, 0)
	for _, snapshot := range ar.window.Snapshots {
		if snapshot.Timestamp.After(windowStart) {
			windowSnapshots = append(windowSnapshots, snapshot)
		}
	}

	if len(windowSnapshots) == 0 {
		return
	}

	// Calculate total requests and errors in window
	totalRequests := int64(0)
	totalErrors := int64(0)
	latencyViolations := int64(0)

	for _, snapshot := range windowSnapshots {
		totalRequests += snapshot.RequestCount
		totalErrors += snapshot.ErrorCount

		// Count latency violations
		if snapshot.P95LatencyMs > float64(ar.config.SLO.LatencyThresholdMs) {
			latencyViolations += snapshot.RequestCount
		}
	}

	// Calculate error budget
	if totalRequests > 0 {
		// Total budget is the number of errors we can afford
		ar.budget.TotalBudget = float64(totalRequests) * (1.0 - ar.config.SLO.AvailabilityTarget)

		// Consumed budget includes both errors and latency violations
		ar.budget.ConsumedBudget = float64(totalErrors + latencyViolations)

		// Remaining budget
		ar.budget.RemainingBudget = ar.budget.TotalBudget - ar.budget.ConsumedBudget
		if ar.budget.RemainingBudget < 0 {
			ar.budget.RemainingBudget = 0
		}

		// Budget utilization (0-1)
		ar.budget.BudgetUtilization = ar.budget.ConsumedBudget / ar.budget.TotalBudget
		if ar.budget.BudgetUtilization > 1.0 {
			ar.budget.BudgetUtilization = 1.0
		}

		// Calculate burn rate (budget consumed per hour)
		ar.calculateBurnRate(windowSnapshots)

		// Determine health status
		ar.budget.IsHealthy = ar.budget.BudgetUtilization < 1.0
		ar.budget.AlertLevel = ar.calculateBudgetAlertLevel()
	}

	ar.budget.LastCalculated = now
}

// calculateBurnRate computes the current rate of budget consumption
func (ar *AnomalyRadar) calculateBurnRate(snapshots []MetricSnapshot) {
	if len(snapshots) < 2 {
		return
	}

	// Calculate burn rate over the last hour
	oneHourAgo := time.Now().Add(-time.Hour)
	recentSnapshots := make([]MetricSnapshot, 0)

	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(oneHourAgo) {
			recentSnapshots = append(recentSnapshots, snapshot)
		}
	}

	if len(recentSnapshots) < 2 {
		return
	}

	// Calculate errors in the last hour
	recentErrors := int64(0)
	recentRequests := int64(0)

	for _, snapshot := range recentSnapshots {
		recentErrors += snapshot.ErrorCount
		recentRequests += snapshot.RequestCount
	}

	// Calculate hourly burn rate
	if recentRequests > 0 {
		errorRate := float64(recentErrors) / float64(recentRequests)
		expectedBudget := float64(recentRequests) * (1.0 - ar.config.SLO.AvailabilityTarget)
		ar.budget.CurrentBurnRate = (errorRate * float64(recentRequests)) / expectedBudget
	}

	// Calculate time to exhaustion
	if ar.budget.CurrentBurnRate > 0 && ar.budget.RemainingBudget > 0 {
		hoursToExhaustion := ar.budget.RemainingBudget / ar.budget.CurrentBurnRate
		ar.budget.TimeToExhaustion = time.Duration(hoursToExhaustion * float64(time.Hour))
	} else {
		ar.budget.TimeToExhaustion = time.Duration(0)
	}
}

// calculateBudgetAlertLevel determines the appropriate alert level for budget consumption
func (ar *AnomalyRadar) calculateBudgetAlertLevel() AlertLevel {
	thresholds := ar.config.SLO.BurnRateThresholds

	// Check fast burn rate
	if ar.budget.CurrentBurnRate >= thresholds.FastBurnRate {
		return AlertLevelCritical
	}

	// Check slow burn rate
	if ar.budget.CurrentBurnRate >= thresholds.SlowBurnRate {
		return AlertLevelWarning
	}

	// Check budget utilization
	if ar.budget.BudgetUtilization >= 0.9 {
		return AlertLevelCritical
	} else if ar.budget.BudgetUtilization >= 0.75 {
		return AlertLevelWarning
	} else if ar.budget.BudgetUtilization >= 0.5 {
		return AlertLevelInfo
	}

	return AlertLevelNone
}

// detectAnomalies analyzes current metrics for anomalous conditions
func (ar *AnomalyRadar) detectAnomalies() {
	if len(ar.window.Snapshots) == 0 {
		return
	}

	latest := ar.window.Snapshots[len(ar.window.Snapshots)-1]
	thresholds := ar.config.Thresholds

	// Check backlog growth
	ar.anomalies.BacklogStatus = ar.evaluateThreshold(
		latest.BacklogGrowthRate,
		thresholds.BacklogGrowthWarning,
		thresholds.BacklogGrowthCritical,
	)

	// Check error rate
	ar.anomalies.ErrorRateStatus = ar.evaluateThreshold(
		latest.ErrorRate,
		thresholds.ErrorRateWarning,
		thresholds.ErrorRateCritical,
	)

	// Check latency
	ar.anomalies.LatencyStatus = ar.evaluateThreshold(
		latest.P95LatencyMs,
		thresholds.LatencyP95Warning,
		thresholds.LatencyP95Critical,
	)

	// Determine overall status
	ar.anomalies.OverallStatus = ar.calculateOverallStatus()
	ar.anomalies.LastUpdated = time.Now()
}

// evaluateThreshold determines metric status based on warning/critical thresholds
func (ar *AnomalyRadar) evaluateThreshold(value, warning, critical float64) MetricStatus {
	if value >= critical {
		return MetricStatusCritical
	} else if value >= warning {
		return MetricStatusWarning
	}
	return MetricStatusHealthy
}

// calculateOverallStatus determines the worst status among all metrics
func (ar *AnomalyRadar) calculateOverallStatus() MetricStatus {
	statuses := []MetricStatus{
		ar.anomalies.BacklogStatus,
		ar.anomalies.ErrorRateStatus,
		ar.anomalies.LatencyStatus,
	}

	// Add budget status
	switch ar.budget.AlertLevel {
	case AlertLevelCritical:
		statuses = append(statuses, MetricStatusCritical)
	case AlertLevelWarning:
		statuses = append(statuses, MetricStatusWarning)
	}

	// Return the worst status
	worstStatus := MetricStatusHealthy
	for _, status := range statuses {
		if status > worstStatus {
			worstStatus = status
		}
	}

	return worstStatus
}

// updateAlerts manages active alerts based on current conditions
func (ar *AnomalyRadar) updateAlerts() {
	newAlerts := make(map[string]*Alert)

	// Check for backlog growth alerts
	ar.checkMetricAlert(
		newAlerts,
		"backlog_growth",
		AlertTypeBacklogGrowth,
		ar.anomalies.BacklogStatus,
		ar.getLatestBacklogGrowthRate(),
		ar.config.Thresholds.BacklogGrowthWarning,
		ar.config.Thresholds.BacklogGrowthCritical,
	)

	// Check for error rate alerts
	ar.checkMetricAlert(
		newAlerts,
		"error_rate",
		AlertTypeErrorRate,
		ar.anomalies.ErrorRateStatus,
		ar.getLatestErrorRate(),
		ar.config.Thresholds.ErrorRateWarning,
		ar.config.Thresholds.ErrorRateCritical,
	)

	// Check for latency alerts
	ar.checkMetricAlert(
		newAlerts,
		"latency_p95",
		AlertTypeLatency,
		ar.anomalies.LatencyStatus,
		ar.getLatestP95Latency(),
		ar.config.Thresholds.LatencyP95Warning,
		ar.config.Thresholds.LatencyP95Critical,
	)

	// Check for burn rate alerts
	ar.checkBurnRateAlert(newAlerts)

	// Update active alerts list
	ar.anomalies.ActiveAlerts = make([]Alert, 0, len(newAlerts))
	for _, alert := range newAlerts {
		ar.anomalies.ActiveAlerts = append(ar.anomalies.ActiveAlerts, *alert)
	}

	// Notify callbacks for new/resolved alerts
	ar.notifyAlertChanges(newAlerts)

	// Update alerts map
	ar.alerts = newAlerts
}

// checkMetricAlert creates or updates an alert for a specific metric
func (ar *AnomalyRadar) checkMetricAlert(alerts map[string]*Alert, id string, alertType AlertType, status MetricStatus, value, warningThreshold, criticalThreshold float64) {
	if status == MetricStatusHealthy {
		return
	}

	severity := AlertLevelWarning
	threshold := warningThreshold
	if status == MetricStatusCritical {
		severity = AlertLevelCritical
		threshold = criticalThreshold
	}

	message := fmt.Sprintf("%s is %s: %.2f (threshold: %.2f)",
		alertType.String(), status.String(), value, threshold)

	alert := &Alert{
		ID: id,
		Type: alertType,
		Severity: severity,
		Message: message,
		Value: value,
		Threshold: threshold,
		UpdatedAt: time.Now(),
	}

	// Set created time for new alerts
	if existing, exists := ar.alerts[id]; exists {
		alert.CreatedAt = existing.CreatedAt
	} else {
		alert.CreatedAt = time.Now()
	}

	alerts[id] = alert
}

// checkBurnRateAlert creates burn rate alerts
func (ar *AnomalyRadar) checkBurnRateAlert(alerts map[string]*Alert) {
	if ar.budget.AlertLevel == AlertLevelNone {
		return
	}

	id := "burn_rate"
	severity := ar.budget.AlertLevel
	threshold := ar.config.SLO.BurnRateThresholds.SlowBurnRate
	if severity == AlertLevelCritical {
		threshold = ar.config.SLO.BurnRateThresholds.FastBurnRate
	}

	message := fmt.Sprintf("SLO budget burn rate is %s: %.4f/hour (threshold: %.4f/hour)",
		severity.String(), ar.budget.CurrentBurnRate, threshold)

	alert := &Alert{
		ID: id,
		Type: AlertTypeBurnRate,
		Severity: severity,
		Message: message,
		Value: ar.budget.CurrentBurnRate,
		Threshold: threshold,
		UpdatedAt: time.Now(),
	}

	if existing, exists := ar.alerts[id]; exists {
		alert.CreatedAt = existing.CreatedAt
	} else {
		alert.CreatedAt = time.Now()
	}

	alerts[id] = alert
}

// notifyAlertChanges calls registered callbacks for alert changes
func (ar *AnomalyRadar) notifyAlertChanges(newAlerts map[string]*Alert) {
	for _, callback := range ar.alertCallbacks {
		// Notify about new alerts
		for id, alert := range newAlerts {
			if _, existed := ar.alerts[id]; !existed {
				callback(*alert)
			}
		}

		// Notify about resolved alerts (alerts that existed but are no longer active)
		// This would be implemented by creating "resolved" alert events
	}
}

// Helper methods to get latest metric values
func (ar *AnomalyRadar) getLatestBacklogGrowthRate() float64 {
	if len(ar.window.Snapshots) == 0 {
		return 0
	}
	return ar.window.Snapshots[len(ar.window.Snapshots)-1].BacklogGrowthRate
}

func (ar *AnomalyRadar) getLatestErrorRate() float64 {
	if len(ar.window.Snapshots) == 0 {
		return 0
	}
	return ar.window.Snapshots[len(ar.window.Snapshots)-1].ErrorRate
}

func (ar *AnomalyRadar) getLatestP95Latency() float64 {
	if len(ar.window.Snapshots) == 0 {
		return 0
	}
	return ar.window.Snapshots[len(ar.window.Snapshots)-1].P95LatencyMs
}

// GetPercentile calculates the specified percentile from recent latency data
func (ar *AnomalyRadar) GetPercentile(percentile float64, window time.Duration) float64 {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	latencies := make([]float64, 0)

	for _, snapshot := range ar.window.Snapshots {
		if snapshot.Timestamp.After(cutoff) {
			latencies = append(latencies, snapshot.P95LatencyMs)
		}
	}

	if len(latencies) == 0 {
		return 0
	}

	sort.Float64s(latencies)
	index := int(math.Ceil(float64(len(latencies))*percentile)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return latencies[index]
}