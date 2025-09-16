package anomalyradarslobudget

import (
	"fmt"
	"time"
)

// ValidateConfig validates the anomaly radar configuration
func ValidateConfig(config Config) error {
	// Validate SLO configuration
	if err := validateSLOConfig(config.SLO); err != nil {
		return fmt.Errorf("invalid SLO config: %w", err)
	}

	// Validate thresholds
	if err := validateThresholds(config.Thresholds); err != nil {
		return fmt.Errorf("invalid thresholds: %w", err)
	}

	// Validate monitoring settings
	if config.MonitoringInterval <= 0 {
		return fmt.Errorf("monitoring interval must be positive")
	}

	if config.MetricRetention <= 0 {
		return fmt.Errorf("metric retention must be positive")
	}

	if config.MaxSnapshots <= 0 {
		return fmt.Errorf("max snapshots must be positive")
	}

	if config.SamplingRate <= 0 || config.SamplingRate > 1.0 {
		return fmt.Errorf("sampling rate must be between 0 and 1")
	}

	return nil
}

// validateSLOConfig validates SLO configuration parameters
func validateSLOConfig(slo SLOConfig) error {
	// Validate availability target
	if slo.AvailabilityTarget <= 0 || slo.AvailabilityTarget >= 1.0 {
		return fmt.Errorf("availability target must be between 0 and 1")
	}

	// Validate latency percentile
	if slo.LatencyPercentile <= 0 || slo.LatencyPercentile >= 1.0 {
		return fmt.Errorf("latency percentile must be between 0 and 1")
	}

	// Validate latency threshold
	if slo.LatencyThresholdMs <= 0 {
		return fmt.Errorf("latency threshold must be positive")
	}

	// Validate window
	if slo.Window <= 0 {
		return fmt.Errorf("SLO window must be positive")
	}

	// Validate burn rate thresholds
	if slo.BurnRateThresholds.FastBurnRate <= 0 || slo.BurnRateThresholds.FastBurnRate >= 1.0 {
		return fmt.Errorf("fast burn rate must be between 0 and 1")
	}

	if slo.BurnRateThresholds.SlowBurnRate <= 0 || slo.BurnRateThresholds.SlowBurnRate >= 1.0 {
		return fmt.Errorf("slow burn rate must be between 0 and 1")
	}

	if slo.BurnRateThresholds.FastBurnWindow <= 0 {
		return fmt.Errorf("fast burn window must be positive")
	}

	if slo.BurnRateThresholds.SlowBurnWindow <= 0 {
		return fmt.Errorf("slow burn window must be positive")
	}

	// Fast burn should be more aggressive than slow burn
	if slo.BurnRateThresholds.FastBurnRate >= slo.BurnRateThresholds.SlowBurnRate {
		return fmt.Errorf("fast burn rate should be less than slow burn rate")
	}

	return nil
}

// validateThresholds validates anomaly detection thresholds
func validateThresholds(thresholds AnomalyThresholds) error {
	// Validate backlog growth thresholds
	if thresholds.BacklogGrowthWarning < 0 {
		return fmt.Errorf("backlog growth warning threshold must be non-negative")
	}

	if thresholds.BacklogGrowthCritical <= thresholds.BacklogGrowthWarning {
		return fmt.Errorf("backlog growth critical threshold must be greater than warning")
	}

	// Validate error rate thresholds
	if thresholds.ErrorRateWarning < 0 || thresholds.ErrorRateWarning >= 1.0 {
		return fmt.Errorf("error rate warning threshold must be between 0 and 1")
	}

	if thresholds.ErrorRateCritical <= thresholds.ErrorRateWarning || thresholds.ErrorRateCritical >= 1.0 {
		return fmt.Errorf("error rate critical threshold must be greater than warning and less than 1")
	}

	// Validate latency thresholds
	if thresholds.LatencyP95Warning <= 0 {
		return fmt.Errorf("latency p95 warning threshold must be positive")
	}

	if thresholds.LatencyP95Critical <= thresholds.LatencyP95Warning {
		return fmt.Errorf("latency p95 critical threshold must be greater than warning")
	}

	return nil
}

// GetRecommendedConfig returns configuration recommendations based on system characteristics
func GetRecommendedConfig(expectedQPS float64, targetLatency time.Duration, systemCriticality string) Config {
	config := DefaultConfig()

	// Adjust based on expected QPS
	if expectedQPS > 1000 {
		// High-throughput system
		config.MonitoringInterval = 5 * time.Second
		config.MaxSnapshots = 17280 // 24 hours at 5-second intervals
		config.Thresholds.BacklogGrowthWarning = 50.0
		config.Thresholds.BacklogGrowthCritical = 200.0
	} else if expectedQPS > 100 {
		// Medium-throughput system
		config.MonitoringInterval = 10 * time.Second
		config.Thresholds.BacklogGrowthWarning = 20.0
		config.Thresholds.BacklogGrowthCritical = 100.0
	} else {
		// Low-throughput system
		config.MonitoringInterval = 30 * time.Second
		config.MaxSnapshots = 2880 // 24 hours at 30-second intervals
		config.Thresholds.BacklogGrowthWarning = 5.0
		config.Thresholds.BacklogGrowthCritical = 25.0
	}

	// Adjust latency thresholds based on target
	latencyMs := float64(targetLatency.Milliseconds())
	config.SLO.LatencyThresholdMs = int64(latencyMs)
	config.Thresholds.LatencyP95Warning = latencyMs * 0.8
	config.Thresholds.LatencyP95Critical = latencyMs

	// Adjust SLO targets based on criticality
	switch systemCriticality {
	case "critical":
		config.SLO.AvailabilityTarget = 0.9999 // 99.99%
		config.Thresholds.ErrorRateWarning = 0.001 // 0.1%
		config.Thresholds.ErrorRateCritical = 0.01 // 1%
		config.SLO.BurnRateThresholds.FastBurnRate = 0.005 // 0.5% in 1 hour
		config.SLO.BurnRateThresholds.SlowBurnRate = 0.02  // 2% in 6 hours
	case "high":
		config.SLO.AvailabilityTarget = 0.999 // 99.9%
		config.Thresholds.ErrorRateWarning = 0.005 // 0.5%
		config.Thresholds.ErrorRateCritical = 0.02 // 2%
		config.SLO.BurnRateThresholds.FastBurnRate = 0.01 // 1% in 1 hour
		config.SLO.BurnRateThresholds.SlowBurnRate = 0.05 // 5% in 6 hours
	case "medium":
		// Use default values (99.5%)
	case "low":
		config.SLO.AvailabilityTarget = 0.99 // 99%
		config.Thresholds.ErrorRateWarning = 0.02 // 2%
		config.Thresholds.ErrorRateCritical = 0.05 // 5%
		config.SLO.BurnRateThresholds.FastBurnRate = 0.02 // 2% in 1 hour
		config.SLO.BurnRateThresholds.SlowBurnRate = 0.1  // 10% in 6 hours
	}

	return config
}

// UpdateConfig safely updates configuration while preserving runtime state
func (ar *AnomalyRadar) UpdateConfig(newConfig Config) error {
	// Validate new configuration
	if err := ValidateConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Update configuration
	oldConfig := ar.config
	ar.config = newConfig

	// Update SLO budget config
	ar.budget.Config = newConfig.SLO

	// Adjust rolling window if retention changed
	if newConfig.MetricRetention != oldConfig.MetricRetention {
		ar.window.WindowSize = newConfig.MetricRetention
		// Trim snapshots if retention decreased
		ar.trimWindowToRetention()
	}

	// Update max snapshots if changed
	if newConfig.MaxSnapshots != oldConfig.MaxSnapshots {
		ar.window.maxSnapshots = newConfig.MaxSnapshots
		// Trim snapshots if limit decreased
		ar.trimWindowToMaxSnapshots()
	}

	return nil
}

// trimWindowToRetention removes snapshots outside the retention window
func (ar *AnomalyRadar) trimWindowToRetention() {
	cutoff := time.Now().Add(-ar.window.WindowSize)
	validSnapshots := make([]MetricSnapshot, 0, len(ar.window.Snapshots))

	for _, snapshot := range ar.window.Snapshots {
		if snapshot.Timestamp.After(cutoff) {
			validSnapshots = append(validSnapshots, snapshot)
		}
	}

	ar.window.Snapshots = validSnapshots
}

// trimWindowToMaxSnapshots enforces the maximum snapshots limit
func (ar *AnomalyRadar) trimWindowToMaxSnapshots() {
	if len(ar.window.Snapshots) > ar.window.maxSnapshots {
		excess := len(ar.window.Snapshots) - ar.window.maxSnapshots
		ar.window.Snapshots = ar.window.Snapshots[excess:]
	}
}

// GetConfig returns the current configuration
func (ar *AnomalyRadar) GetConfig() Config {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return ar.config
}

// GetConfigSummary returns a human-readable summary of the current configuration
func (ar *AnomalyRadar) GetConfigSummary() string {
	config := ar.GetConfig()

	summary := fmt.Sprintf(`Anomaly Radar Configuration:
  SLO Target: %.2f%% availability, p95 < %dms
  SLO Window: %v
  Monitoring Interval: %v
  Metric Retention: %v

  Thresholds:
    Backlog Growth: Warning %.1f/s, Critical %.1f/s
    Error Rate: Warning %.1f%%, Critical %.1f%%
    Latency P95: Warning %.0fms, Critical %.0fms

  Burn Rate Alerts:
    Fast: %.1f%% in %v
    Slow: %.1f%% in %v`,
		config.SLO.AvailabilityTarget*100,
		config.SLO.LatencyThresholdMs,
		config.SLO.Window,
		config.MonitoringInterval,
		config.MetricRetention,
		config.Thresholds.BacklogGrowthWarning,
		config.Thresholds.BacklogGrowthCritical,
		config.Thresholds.ErrorRateWarning*100,
		config.Thresholds.ErrorRateCritical*100,
		config.Thresholds.LatencyP95Warning,
		config.Thresholds.LatencyP95Critical,
		config.SLO.BurnRateThresholds.FastBurnRate*100,
		config.SLO.BurnRateThresholds.FastBurnWindow,
		config.SLO.BurnRateThresholds.SlowBurnRate*100,
		config.SLO.BurnRateThresholds.SlowBurnWindow,
	)

	return summary
}