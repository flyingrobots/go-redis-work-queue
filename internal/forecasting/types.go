// Copyright 2025 James Ross
package forecasting

import (
	"sync"
	"time"
)

// MetricType represents the type of metric being forecasted
type MetricType string

const (
	MetricBacklog    MetricType = "backlog"
	MetricThroughput MetricType = "throughput"
	MetricErrorRate  MetricType = "error_rate"
	MetricLatency    MetricType = "latency"
	MetricWorkers    MetricType = "workers"
)

// ForecastResult contains the forecast output
type ForecastResult struct {
	Points         []float64     `json:"points"`
	UpperBounds    []float64     `json:"upper_bounds"`
	LowerBounds    []float64     `json:"lower_bounds"`
	Confidence     float64       `json:"confidence"`
	ModelUsed      string        `json:"model_used"`
	GeneratedAt    time.Time     `json:"generated_at"`
	HorizonMinutes int           `json:"horizon_minutes"`
	MetricType     MetricType    `json:"metric_type"`
}

// QueueMetrics represents current queue state
type QueueMetrics struct {
	Timestamp      time.Time `json:"timestamp"`
	Backlog        int64     `json:"backlog"`
	Throughput     float64   `json:"throughput"`
	ErrorRate      float64   `json:"error_rate"`
	LatencyP50     float64   `json:"latency_p50"`
	LatencyP95     float64   `json:"latency_p95"`
	LatencyP99     float64   `json:"latency_p99"`
	ActiveWorkers  int       `json:"active_workers"`
	QueueName      string    `json:"queue_name"`
}

// DataPoint represents a single time series observation
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeries represents a collection of data points
type TimeSeries struct {
	Name       string      `json:"name"`
	MetricType MetricType  `json:"metric_type"`
	Points     []DataPoint `json:"points"`
	mu         sync.RWMutex
}

// Recommendation represents an actionable recommendation
type Recommendation struct {
	ID          string                 `json:"id"`
	Priority    RecommendationPriority `json:"priority"`
	Category    RecommendationCategory `json:"category"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Timing      time.Duration          `json:"timing"`
	Confidence  float64                `json:"confidence"`
	CreatedAt   time.Time              `json:"created_at"`
}

// RecommendationPriority defines priority levels
type RecommendationPriority int

const (
	PriorityCritical RecommendationPriority = iota
	PriorityHigh
	PriorityMedium
	PriorityLow
	PriorityInfo
)

// RecommendationCategory defines recommendation types
type RecommendationCategory string

const (
	CategoryCapacityScaling     RecommendationCategory = "capacity_scaling"
	CategorySLOManagement       RecommendationCategory = "slo_management"
	CategoryMaintenanceScheduling RecommendationCategory = "maintenance_scheduling"
	CategoryAnomaly             RecommendationCategory = "anomaly"
	CategoryPerformance         RecommendationCategory = "performance"
)

// AccuracyMetrics tracks model accuracy
type AccuracyMetrics struct {
	MAE            float64   `json:"mae"`             // Mean Absolute Error
	RMSE           float64   `json:"rmse"`            // Root Mean Square Error
	MAPE           float64   `json:"mape"`            // Mean Absolute Percentage Error
	PredictionBias float64   `json:"prediction_bias"` // Average prediction - actual
	R2Score        float64   `json:"r2_score"`        // Coefficient of determination
	SampleSize     int       `json:"sample_size"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ModelConfig defines model configuration
type ModelConfig struct {
	ModelType      string                 `json:"model_type"`
	Parameters     map[string]interface{} `json:"parameters"`
	UpdateInterval time.Duration          `json:"update_interval"`
	Horizon        time.Duration          `json:"horizon"`
	Enabled        bool                   `json:"enabled"`
}

// ForecastConfig defines forecasting configuration
type ForecastConfig struct {
	EWMAConfig       *EWMAConfig       `json:"ewma_config"`
	HoltWintersConfig *HoltWintersConfig `json:"holt_winters_config"`
	StorageConfig    *StorageConfig    `json:"storage_config"`
	EngineConfig     *EngineConfig     `json:"engine_config"`
}

// EWMAConfig defines EWMA model configuration
type EWMAConfig struct {
	Alpha              float64       `json:"alpha"`                // Smoothing parameter (0 < α < 1)
	AutoAdjust         bool          `json:"auto_adjust"`          // Auto-adjust alpha based on accuracy
	MinObservations    int           `json:"min_observations"`     // Minimum observations before forecasting
	ConfidenceInterval float64       `json:"confidence_interval"`  // Confidence interval (e.g., 0.95 for 95%)
}

// HoltWintersConfig defines Holt-Winters model configuration
type HoltWintersConfig struct {
	Alpha              float64       `json:"alpha"`                // Level smoothing
	Beta               float64       `json:"beta"`                 // Trend smoothing
	Gamma              float64       `json:"gamma"`                // Seasonal smoothing
	SeasonLength       int           `json:"season_length"`        // Length of seasonal cycle
	SeasonalMethod     string        `json:"seasonal_method"`      // "additive" or "multiplicative"
	AutoDetectSeason   bool          `json:"auto_detect_season"`   // Auto-detect seasonality
}

// StorageConfig defines data storage configuration
type StorageConfig struct {
	RetentionDuration  time.Duration `json:"retention_duration"`   // How long to keep historical data
	SamplingInterval   time.Duration `json:"sampling_interval"`     // How often to sample metrics
	MaxDataPoints      int           `json:"max_data_points"`       // Maximum data points to store
	PersistToDisk      bool          `json:"persist_to_disk"`       // Persist data to disk
	StoragePath        string        `json:"storage_path"`          // Path for persistent storage
}

// EngineConfig defines recommendation engine configuration
type EngineConfig struct {
	Enabled                bool                   `json:"enabled"`
	UpdateInterval         time.Duration          `json:"update_interval"`
	Thresholds             map[string]float64     `json:"thresholds"`
	ScalingPolicy          ScalingPolicy          `json:"scaling_policy"`
	MaintenancePreferences MaintenancePreferences `json:"maintenance_preferences"`
}

// ScalingPolicy defines how to recommend scaling
type ScalingPolicy struct {
	MinWorkers         int     `json:"min_workers"`
	MaxWorkers         int     `json:"max_workers"`
	ScaleUpThreshold   float64 `json:"scale_up_threshold"`
	ScaleDownThreshold float64 `json:"scale_down_threshold"`
	CooldownPeriod     time.Duration `json:"cooldown_period"`
}

// MaintenancePreferences defines maintenance window preferences
type MaintenancePreferences struct {
	PreferredDays      []time.Weekday `json:"preferred_days"`
	PreferredStartHour int            `json:"preferred_start_hour"`
	PreferredEndHour   int            `json:"preferred_end_hour"`
	MinimumDuration    time.Duration  `json:"minimum_duration"`
	MaximumDuration    time.Duration  `json:"maximum_duration"`
}

// MaintenanceWindow represents an optimal maintenance window
type MaintenanceWindow struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Impact     float64   `json:"impact"`      // Estimated jobs affected
	Confidence float64   `json:"confidence"`
}

// SLOBudget tracks SLO budget consumption
type SLOBudget struct {
	Target           float64   `json:"target"`            // SLO target (e.g., 0.999)
	CurrentBurn      float64   `json:"current_burn"`      // Current budget consumption
	WeeklyBurnRate   float64   `json:"weekly_burn_rate"`  // Burn rate this week
	MonthlyBurnRate  float64   `json:"monthly_burn_rate"` // Burn rate this month
	RemainingBudget  float64   `json:"remaining_budget"`  // Remaining error budget
	ProjectedBurn    float64   `json:"projected_burn"`    // Projected burn based on forecast
	TimeToExhaustion time.Duration `json:"time_to_exhaustion"` // Time until budget exhausted
	LastUpdated      time.Time `json:"last_updated"`
}

// PredictionRecord tracks predictions for accuracy evaluation
type PredictionRecord struct {
	Timestamp   time.Time  `json:"timestamp"`
	MetricType  MetricType `json:"metric_type"`
	Predicted   float64    `json:"predicted"`
	Actual      float64    `json:"actual"`
	ModelUsed   string     `json:"model_used"`
	Horizon     time.Duration `json:"horizon"`
	Error       float64    `json:"error"`
	ErrorPercent float64   `json:"error_percent"`
}
