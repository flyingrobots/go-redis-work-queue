package anomalyradarslobudget

import (
	"time"
)

// SLOConfig defines Service Level Objective configuration
type SLOConfig struct {
	// Target availability (e.g., 0.995 for 99.5%)
	AvailabilityTarget float64 `json:"availability_target" yaml:"availability_target"`

	// Target response time percentile (e.g., 0.95 for p95)
	LatencyPercentile float64 `json:"latency_percentile" yaml:"latency_percentile"`

	// Target latency threshold in milliseconds
	LatencyThresholdMs int64 `json:"latency_threshold_ms" yaml:"latency_threshold_ms"`

	// SLO measurement window (e.g., 30 days)
	Window time.Duration `json:"window" yaml:"window"`

	// Burn rate alert thresholds
	BurnRateThresholds BurnRateThresholds `json:"burn_rate_thresholds" yaml:"burn_rate_thresholds"`
}

// BurnRateThresholds defines when to alert on SLO budget consumption
type BurnRateThresholds struct {
	// Fast burn: 1% of budget consumed in 1 hour
	FastBurnRate float64 `json:"fast_burn_rate" yaml:"fast_burn_rate"`
	FastBurnWindow time.Duration `json:"fast_burn_window" yaml:"fast_burn_window"`

	// Slow burn: 5% of budget consumed in 6 hours
	SlowBurnRate float64 `json:"slow_burn_rate" yaml:"slow_burn_rate"`
	SlowBurnWindow time.Duration `json:"slow_burn_window" yaml:"slow_burn_window"`
}

// MetricSnapshot represents a point-in-time measurement
type MetricSnapshot struct {
	Timestamp time.Time `json:"timestamp"`

	// Queue metrics
	BacklogSize int64 `json:"backlog_size"`
	BacklogGrowthRate float64 `json:"backlog_growth_rate"` // items/second

	// Performance metrics
	RequestCount int64 `json:"request_count"`
	ErrorCount int64 `json:"error_count"`
	ErrorRate float64 `json:"error_rate"` // 0-1

	// Latency metrics
	P50LatencyMs float64 `json:"p50_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`
}

// RollingWindow maintains metrics over a sliding time window
type RollingWindow struct {
	WindowSize time.Duration `json:"window_size"`
	Snapshots []MetricSnapshot `json:"snapshots"`
	maxSnapshots int
}

// SLOBudget tracks error budget consumption
type SLOBudget struct {
	// Budget configuration
	Config SLOConfig `json:"config"`

	// Current budget status
	TotalBudget float64 `json:"total_budget"` // Total allowed errors in window
	ConsumedBudget float64 `json:"consumed_budget"` // Errors consumed so far
	RemainingBudget float64 `json:"remaining_budget"` // Budget left
	BudgetUtilization float64 `json:"budget_utilization"` // 0-1

	// Burn rate analysis
	CurrentBurnRate float64 `json:"current_burn_rate"` // budget/hour
	TimeToExhaustion time.Duration `json:"time_to_exhaustion"` // Time until budget = 0

	// Status
	IsHealthy bool `json:"is_healthy"`
	AlertLevel AlertLevel `json:"alert_level"`
	LastCalculated time.Time `json:"last_calculated"`
}

// AlertLevel represents SLO budget alert severity
type AlertLevel int

const (
	AlertLevelNone AlertLevel = iota
	AlertLevelInfo
	AlertLevelWarning
	AlertLevelCritical
)

func (a AlertLevel) String() string {
	switch a {
	case AlertLevelNone:
		return "none"
	case AlertLevelInfo:
		return "info"
	case AlertLevelWarning:
		return "warning"
	case AlertLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AnomalyThresholds define when metrics indicate anomalies
type AnomalyThresholds struct {
	// Backlog growth thresholds
	BacklogGrowthWarning float64 `json:"backlog_growth_warning"` // items/second
	BacklogGrowthCritical float64 `json:"backlog_growth_critical"`

	// Error rate thresholds
	ErrorRateWarning float64 `json:"error_rate_warning"` // 0-1
	ErrorRateCritical float64 `json:"error_rate_critical"`

	// Latency thresholds
	LatencyP95Warning float64 `json:"latency_p95_warning"` // milliseconds
	LatencyP95Critical float64 `json:"latency_p95_critical"`
}

// AnomalyStatus represents detected anomalies
type AnomalyStatus struct {
	// Individual metric statuses
	BacklogStatus MetricStatus `json:"backlog_status"`
	ErrorRateStatus MetricStatus `json:"error_rate_status"`
	LatencyStatus MetricStatus `json:"latency_status"`

	// Overall status
	OverallStatus MetricStatus `json:"overall_status"`

	// Active alerts
	ActiveAlerts []Alert `json:"active_alerts"`

	LastUpdated time.Time `json:"last_updated"`
}

// MetricStatus represents the health status of a metric
type MetricStatus int

const (
	MetricStatusHealthy MetricStatus = iota
	MetricStatusWarning
	MetricStatusCritical
)

func (m MetricStatus) String() string {
	switch m {
	case MetricStatusHealthy:
		return "healthy"
	case MetricStatusWarning:
		return "warning"
	case MetricStatusCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Alert represents an active anomaly alert
type Alert struct {
	ID string `json:"id"`
	Type AlertType `json:"type"`
	Severity AlertLevel `json:"severity"`
	Message string `json:"message"`
	Value float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertType categorizes different types of alerts
type AlertType int

const (
	AlertTypeBacklogGrowth AlertType = iota
	AlertTypeErrorRate
	AlertTypeLatency
	AlertTypeBurnRate
)

func (a AlertType) String() string {
	switch a {
	case AlertTypeBacklogGrowth:
		return "backlog_growth"
	case AlertTypeErrorRate:
		return "error_rate"
	case AlertTypeLatency:
		return "latency"
	case AlertTypeBurnRate:
		return "burn_rate"
	default:
		return "unknown"
	}
}

// Config represents the complete configuration for anomaly radar
type Config struct {
	// SLO configuration
	SLO SLOConfig `json:"slo" yaml:"slo"`

	// Anomaly detection thresholds
	Thresholds AnomalyThresholds `json:"thresholds" yaml:"thresholds"`

	// Monitoring configuration
	MonitoringInterval time.Duration `json:"monitoring_interval" yaml:"monitoring_interval"`
	MetricRetention time.Duration `json:"metric_retention" yaml:"metric_retention"`

	// Performance settings
	MaxSnapshots int `json:"max_snapshots" yaml:"max_snapshots"`
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		SLO: SLOConfig{
			AvailabilityTarget: 0.995, // 99.5%
			LatencyPercentile: 0.95,   // p95
			LatencyThresholdMs: 1000,  // 1 second
			Window: 30 * 24 * time.Hour, // 30 days
			BurnRateThresholds: BurnRateThresholds{
				FastBurnRate: 0.01, // 1% in 1 hour = budget exhausted in ~4 days
				FastBurnWindow: time.Hour,
				SlowBurnRate: 0.05, // 5% in 6 hours = budget exhausted in ~5 days
				SlowBurnWindow: 6 * time.Hour,
			},
		},
		Thresholds: AnomalyThresholds{
			BacklogGrowthWarning: 10.0,   // 10 items/second
			BacklogGrowthCritical: 50.0,  // 50 items/second
			ErrorRateWarning: 0.01,       // 1%
			ErrorRateCritical: 0.05,      // 5%
			LatencyP95Warning: 500.0,     // 500ms
			LatencyP95Critical: 1000.0,   // 1s
		},
		MonitoringInterval: 10 * time.Second,
		MetricRetention: 24 * time.Hour,
		MaxSnapshots: 8640, // 24 hours at 10-second intervals
		SamplingRate: 1.0,  // 100% sampling
	}
}