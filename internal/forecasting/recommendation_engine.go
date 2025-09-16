// Copyright 2025 James Ross
package forecasting

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RecommendationEngine generates actionable recommendations from forecasts
type RecommendationEngine struct {
	config       *EngineConfig
	sloTracker   *SLOTracker
	logger       *zap.Logger
	mu           sync.RWMutex

	// Recent recommendations to avoid duplicates
	recentRecs map[string]time.Time
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(config *EngineConfig, logger *zap.Logger) *RecommendationEngine {
	if config == nil {
		config = &EngineConfig{
			Enabled:        true,
			UpdateInterval: 5 * time.Minute,
			Thresholds: map[string]float64{
				"critical_backlog":     1000,
				"high_backlog":         500,
				"critical_error_rate":  0.1,
				"high_error_rate":      0.05,
				"low_throughput":       10,
				"high_latency_p99":     5000, // ms
			},
			ScalingPolicy: ScalingPolicy{
				MinWorkers:         1,
				MaxWorkers:         100,
				ScaleUpThreshold:   0.8,
				ScaleDownThreshold: 0.3,
				CooldownPeriod:     5 * time.Minute,
			},
			MaintenancePreferences: MaintenancePreferences{
				PreferredDays:      []time.Weekday{time.Tuesday, time.Wednesday, time.Thursday},
				PreferredStartHour: 2,  // 2 AM
				PreferredEndHour:   6,  // 6 AM
				MinimumDuration:    30 * time.Minute,
				MaximumDuration:    4 * time.Hour,
			},
		}
	}

	return &RecommendationEngine{
		config:     config,
		sloTracker: NewSLOTracker(),
		logger:     logger,
		recentRecs: make(map[string]time.Time),
	}
}

// GenerateRecommendations generates recommendations from forecasts and metrics
func (re *RecommendationEngine) GenerateRecommendations(
	forecasts map[MetricType]*ForecastResult,
	currentMetrics *QueueMetrics) []Recommendation {

	if !re.config.Enabled {
		return nil
	}

	re.mu.Lock()
	defer re.mu.Unlock()

	// Clean up old recommendations
	re.cleanupRecentRecommendations()

	recommendations := []Recommendation{}

	// Capacity scaling recommendations
	if recs := re.generateCapacityRecommendations(forecasts, currentMetrics); recs != nil {
		recommendations = append(recommendations, recs...)
	}

	// SLO management recommendations
	if recs := re.generateSLORecommendations(forecasts, currentMetrics); recs != nil {
		recommendations = append(recommendations, recs...)
	}

	// Maintenance window recommendations
	if rec := re.generateMaintenanceRecommendation(forecasts); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Anomaly detection recommendations
	if recs := re.generateAnomalyRecommendations(forecasts, currentMetrics); recs != nil {
		recommendations = append(recommendations, recs...)
	}

	// Performance optimization recommendations
	if recs := re.generatePerformanceRecommendations(forecasts, currentMetrics); recs != nil {
		recommendations = append(recommendations, recs...)
	}

	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority < recommendations[j].Priority
	})

	// Mark as recent to avoid duplicates
	for i := range recommendations {
		recommendations[i].CreatedAt = time.Now()
		re.recentRecs[recommendations[i].ID] = time.Now()
	}

	return recommendations
}

func (re *RecommendationEngine) generateCapacityRecommendations(
	forecasts map[MetricType]*ForecastResult,
	metrics *QueueMetrics) []Recommendation {

	recs := []Recommendation{}

	// Check backlog forecast
	backlogForecast, exists := forecasts[MetricBacklog]
	if !exists || backlogForecast == nil {
		return recs
	}

	// Find peak backlog in forecast
	maxBacklog, peakIndex := re.findPeak(backlogForecast.Points)
	peakTime := time.Duration(peakIndex) * time.Minute

	// Critical backlog level
	if maxBacklog > re.config.Thresholds["critical_backlog"] {
		workersNeeded := re.calculateWorkersNeeded(
			metrics.ActiveWorkers,
			metrics.Throughput,
			maxBacklog,
			peakTime,
		)

		if !re.isDuplicate("scale-critical") {
			recs = append(recs, Recommendation{
				ID:          "scale-critical",
				Priority:    PriorityCritical,
				Category:    CategoryCapacityScaling,
				Title:       "🚨 CRITICAL: Scale Workers Immediately",
				Description: fmt.Sprintf("Backlog will reach %.0f in %v. Scale workers +%d NOW!",
					maxBacklog, peakTime, workersNeeded),
				Action: fmt.Sprintf("kubectl scale deployment workers --replicas=%d",
					metrics.ActiveWorkers+workersNeeded),
				Timing:     peakTime - 5*time.Minute,
				Confidence: backlogForecast.Confidence,
			})
		}
	} else if maxBacklog > re.config.Thresholds["high_backlog"] {
		// High backlog level
		workersNeeded := re.calculateWorkersNeeded(
			metrics.ActiveWorkers,
			metrics.Throughput,
			maxBacklog,
			peakTime,
		)

		if !re.isDuplicate("scale-high") {
			recs = append(recs, Recommendation{
				ID:          "scale-high",
				Priority:    PriorityHigh,
				Category:    CategoryCapacityScaling,
				Title:       "⚠️ Scale Workers Recommended",
				Description: fmt.Sprintf("Backlog trending up to %.0f. Consider scaling +%d workers",
					maxBacklog, workersNeeded),
				Action: fmt.Sprintf("kubectl scale deployment workers --replicas=%d",
					metrics.ActiveWorkers+workersNeeded),
				Timing:     peakTime,
				Confidence: backlogForecast.Confidence,
			})
		}
	}

	// Check for over-provisioning
	if metrics.Backlog < 100 && metrics.ActiveWorkers > re.config.ScalingPolicy.MinWorkers*2 {
		if !re.isDuplicate("scale-down") {
			recs = append(recs, Recommendation{
				ID:          "scale-down",
				Priority:    PriorityLow,
				Category:    CategoryCapacityScaling,
				Title:       "💰 Consider Scaling Down",
				Description: fmt.Sprintf("Low backlog with %d workers. Could reduce to %d",
					metrics.ActiveWorkers, metrics.ActiveWorkers/2),
				Action:     fmt.Sprintf("kubectl scale deployment workers --replicas=%d", metrics.ActiveWorkers/2),
				Timing:     15 * time.Minute,
				Confidence: 0.75,
			})
		}
	}

	return recs
}

func (re *RecommendationEngine) generateSLORecommendations(
	forecasts map[MetricType]*ForecastResult,
	metrics *QueueMetrics) []Recommendation {

	recs := []Recommendation{}

	// Update SLO tracker
	re.sloTracker.Update(metrics.ErrorRate)

	// Check error rate forecast
	errorForecast, exists := forecasts[MetricErrorRate]
	if exists && errorForecast != nil {
		// Project budget burn
		budget := re.sloTracker.ProjectBudgetBurn(errorForecast.Points)

		if budget.WeeklyBurnRate > 0.9 {
			if !re.isDuplicate("slo-critical") {
				recs = append(recs, Recommendation{
					ID:       "slo-critical",
					Priority: PriorityCritical,
					Category: CategorySLOManagement,
					Title:    "🔴 SLO Budget Critical",
					Description: fmt.Sprintf("Only %.1f%% error budget remaining this week!",
						(1-budget.WeeklyBurnRate)*100),
					Action:     "Review error logs and implement fixes immediately",
					Timing:     0,
					Confidence: 0.9,
				})
			}
		} else if budget.WeeklyBurnRate > 0.7 {
			if !re.isDuplicate("slo-warning") {
				recs = append(recs, Recommendation{
					ID:       "slo-warning",
					Priority: PriorityHigh,
					Category: CategorySLOManagement,
					Title:    "⚠️ SLO Budget Warning",
					Description: fmt.Sprintf("%.1f%% error budget consumed. Time to exhaustion: %v",
						budget.WeeklyBurnRate*100, budget.TimeToExhaustion),
					Action:     "Monitor error trends closely",
					Timing:     1 * time.Hour,
					Confidence: 0.85,
				})
			}
		}
	}

	return recs
}

func (re *RecommendationEngine) generateMaintenanceRecommendation(
	forecasts map[MetricType]*ForecastResult) *Recommendation {

	backlogForecast, exists := forecasts[MetricBacklog]
	if !exists || backlogForecast == nil {
		return nil
	}

	// Find optimal maintenance window
	window := re.findOptimalMaintenanceWindow(backlogForecast.Points)
	if window == nil {
		return nil
	}

	if re.isDuplicate("maintenance-window") {
		return nil
	}

	return &Recommendation{
		ID:       "maintenance-window",
		Priority: PriorityInfo,
		Category: CategoryMaintenanceScheduling,
		Title:    "🔧 Optimal Maintenance Window",
		Description: fmt.Sprintf("Best window: %s to %s (impact: ~%.0f jobs)",
			window.Start.Format("Mon 15:04"),
			window.End.Format("15:04"),
			window.Impact),
		Action:     "Schedule maintenance during this low-impact period",
		Timing:     time.Until(window.Start),
		Confidence: window.Confidence,
	}
}

func (re *RecommendationEngine) generateAnomalyRecommendations(
	forecasts map[MetricType]*ForecastResult,
	metrics *QueueMetrics) []Recommendation {

	recs := []Recommendation{}

	// Check for anomalies in current metrics vs forecasts
	for metricType, forecast := range forecasts {
		if forecast == nil || len(forecast.Points) == 0 {
			continue
		}

		var currentValue float64
		switch metricType {
		case MetricBacklog:
			currentValue = float64(metrics.Backlog)
		case MetricThroughput:
			currentValue = metrics.Throughput
		case MetricErrorRate:
			currentValue = metrics.ErrorRate
		case MetricLatency:
			currentValue = metrics.LatencyP99
		}

		// Check if current value is outside confidence bounds
		if currentValue > forecast.UpperBounds[0]*1.5 {
			if !re.isDuplicate(fmt.Sprintf("anomaly-%s", metricType)) {
				recs = append(recs, Recommendation{
					ID:       fmt.Sprintf("anomaly-%s", metricType),
					Priority: PriorityHigh,
					Category: CategoryAnomaly,
					Title:    fmt.Sprintf("🔍 Anomaly Detected: %s", metricType),
					Description: fmt.Sprintf("%s is %.1f%% above expected range",
						metricType, (currentValue/forecast.Points[0]-1)*100),
					Action:     "Investigate recent changes or incidents",
					Timing:     0,
					Confidence: 0.8,
				})
			}
		}
	}

	return recs
}

func (re *RecommendationEngine) generatePerformanceRecommendations(
	forecasts map[MetricType]*ForecastResult,
	metrics *QueueMetrics) []Recommendation {

	recs := []Recommendation{}

	// Check latency
	if metrics.LatencyP99 > re.config.Thresholds["high_latency_p99"] {
		if !re.isDuplicate("perf-latency") {
			recs = append(recs, Recommendation{
				ID:          "perf-latency",
				Priority:    PriorityMedium,
				Category:    CategoryPerformance,
				Title:       "🐌 High Latency Detected",
				Description: fmt.Sprintf("P99 latency is %.0fms. Consider optimization", metrics.LatencyP99),
				Action:      "Profile slow jobs and optimize processing",
				Timing:      30 * time.Minute,
				Confidence:  0.85,
			})
		}
	}

	// Check throughput
	if metrics.Throughput < re.config.Thresholds["low_throughput"] {
		if !re.isDuplicate("perf-throughput") {
			recs = append(recs, Recommendation{
				ID:          "perf-throughput",
				Priority:    PriorityMedium,
				Category:    CategoryPerformance,
				Title:       "📉 Low Throughput",
				Description: fmt.Sprintf("Throughput is only %.1f jobs/sec", metrics.Throughput),
				Action:      "Check for blocking operations or resource constraints",
				Timing:      1 * time.Hour,
				Confidence:  0.75,
			})
		}
	}

	return recs
}

// Helper methods

func (re *RecommendationEngine) findPeak(values []float64) (float64, int) {
	if len(values) == 0 {
		return 0, 0
	}

	maxVal := values[0]
	maxIdx := 0

	for i, v := range values {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	return maxVal, maxIdx
}

func (re *RecommendationEngine) calculateWorkersNeeded(
	currentWorkers int,
	throughput float64,
	backlog float64,
	timeToResolve time.Duration) int {

	if throughput <= 0 {
		throughput = 1 // Avoid division by zero
	}

	// Calculate required throughput to handle backlog
	requiredThroughput := backlog / timeToResolve.Minutes() / 60

	// Calculate workers needed (assuming linear scaling)
	workersNeeded := int(math.Ceil(requiredThroughput / throughput * float64(currentWorkers)))

	// Apply limits
	additionalWorkers := workersNeeded - currentWorkers
	if additionalWorkers < 1 {
		additionalWorkers = 1
	}
	if currentWorkers+additionalWorkers > re.config.ScalingPolicy.MaxWorkers {
		additionalWorkers = re.config.ScalingPolicy.MaxWorkers - currentWorkers
	}

	return additionalWorkers
}

func (re *RecommendationEngine) findOptimalMaintenanceWindow(backlogForecast []float64) *MaintenanceWindow {
	now := time.Now()
	prefs := re.config.MaintenancePreferences

	var bestWindow *MaintenanceWindow
	lowestImpact := math.MaxFloat64

	// Look ahead 7 days
	for days := 0; days < 7; days++ {
		testDate := now.AddDate(0, 0, days)

		// Check if day is preferred
		isPrefDay := false
		for _, prefDay := range prefs.PreferredDays {
			if testDate.Weekday() == prefDay {
				isPrefDay = true
				break
			}
		}

		if !isPrefDay {
			continue
		}

		// Check preferred hours
		for hour := prefs.PreferredStartHour; hour < prefs.PreferredEndHour; hour++ {
			start := time.Date(testDate.Year(), testDate.Month(), testDate.Day(),
				hour, 0, 0, 0, testDate.Location())
			end := start.Add(prefs.MinimumDuration)

			// Calculate impact (jobs affected during window)
			minutesFromNow := int(start.Sub(now).Minutes())
			if minutesFromNow < 0 || minutesFromNow >= len(backlogForecast) {
				continue
			}

			impact := backlogForecast[minutesFromNow]

			if impact < lowestImpact {
				lowestImpact = impact
				bestWindow = &MaintenanceWindow{
					Start:      start,
					End:        end,
					Impact:     impact,
					Confidence: 0.7,
				}
			}
		}
	}

	return bestWindow
}

func (re *RecommendationEngine) isDuplicate(id string) bool {
	if lastTime, exists := re.recentRecs[id]; exists {
		// Don't repeat recommendations within cooldown period
		if time.Since(lastTime) < re.config.ScalingPolicy.CooldownPeriod {
			return true
		}
	}
	return false
}

func (re *RecommendationEngine) cleanupRecentRecommendations() {
	cutoff := time.Now().Add(-re.config.ScalingPolicy.CooldownPeriod)
	for id, timestamp := range re.recentRecs {
		if timestamp.Before(cutoff) {
			delete(re.recentRecs, id)
		}
	}
}
