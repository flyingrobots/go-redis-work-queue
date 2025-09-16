// Copyright 2025 James Ross
package capacityplanning

import (
	"time"
)

// Metrics represents the current state of the queue system
type Metrics struct {
	Timestamp      time.Time     `json:"timestamp"`
	ArrivalRate    float64       `json:"arrival_rate"`     // Jobs per second (λ)
	ServiceTime    time.Duration `json:"service_time"`     // Mean service time (1/μ)
	ServiceTimeP95 time.Duration `json:"service_time_p95"` // 95th percentile service time
	ServiceTimeStd time.Duration `json:"service_time_std"` // Standard deviation
	CurrentWorkers int           `json:"current_workers"`  // Current worker count (c)
	Utilization    float64       `json:"utilization"`      // Current utilization (ρ)
	Backlog        int           `json:"backlog"`          // Current queue length
	ActiveJobs     int           `json:"active_jobs"`      // Jobs currently being processed
	TotalCapacity  int           `json:"total_capacity"`   // Total processing slots
	QueueName      string        `json:"queue_name"`       // Queue identifier
}

// SLO defines service level objectives
type SLO struct {
	P95Latency    time.Duration `json:"p95_latency"`    // Target 95th percentile latency
	MaxBacklog    int           `json:"max_backlog"`    // Maximum allowed backlog
	ErrorBudget   float64       `json:"error_budget"`   // Acceptable error rate (0.0-1.0)
	DrainTime     time.Duration `json:"drain_time"`     // Time to drain after burst
	Availability  float64       `json:"availability"`   // Target availability (0.99 = 99%)
}

// CapacityPlan represents a scaling recommendation
type CapacityPlan struct {
	ID              string        `json:"id"`
	GeneratedAt     time.Time     `json:"generated_at"`
	CurrentWorkers  int           `json:"current_workers"`
	TargetWorkers   int           `json:"target_workers"`
	Delta           int           `json:"delta"` // TargetWorkers - CurrentWorkers
	Steps           []ScalingStep `json:"steps"`
	Confidence      float64       `json:"confidence"`       // 0.0-1.0
	CostImpact      CostAnalysis  `json:"cost_impact"`
	SLOAchievable   bool          `json:"slo_achievable"`
	Rationale       string        `json:"rationale"`
	ForecastWindow  time.Duration `json:"forecast_window"`
	SafetyMargin    float64       `json:"safety_margin"`
	ValidUntil      time.Time     `json:"valid_until"`
	QueueName       string        `json:"queue_name"`
}

// ScalingStep represents a single scaling action
type ScalingStep struct {
	Sequence      int           `json:"sequence"`
	ScheduledAt   time.Time     `json:"scheduled_at"`
	Action        ScalingAction `json:"action"`
	FromWorkers   int           `json:"from_workers"`
	ToWorkers     int           `json:"to_workers"`
	Delta         int           `json:"delta"`
	Rationale     string        `json:"rationale"`
	EstimatedCost float64       `json:"estimated_cost"` // $/hour
	Confidence    float64       `json:"confidence"`
	CooldownUntil time.Time     `json:"cooldown_until"`
}

// ScalingAction defines the type of scaling operation
type ScalingAction string

const (
	ScaleUp   ScalingAction = "scale_up"
	ScaleDown ScalingAction = "scale_down"
	NoChange  ScalingAction = "no_change"
)

// CostAnalysis provides financial impact information
type CostAnalysis struct {
	CurrentCostPerHour  float64 `json:"current_cost_per_hour"`
	ProjectedCostPerHour float64 `json:"projected_cost_per_hour"`
	DeltaCostPerHour    float64 `json:"delta_cost_per_hour"`
	MonthlyCostDelta    float64 `json:"monthly_cost_delta"`
	ViolationCostRisk   float64 `json:"violation_cost_risk"`
	NetBenefit          float64 `json:"net_benefit"` // Savings minus violation risk
	PaybackPeriod       string  `json:"payback_period"`
}

// Forecast represents future arrival rate predictions
type Forecast struct {
	Timestamp   time.Time `json:"timestamp"`
	ArrivalRate float64   `json:"arrival_rate"`
	Confidence  float64   `json:"confidence"`
	Lower       float64   `json:"lower_bound"` // Confidence interval
	Upper       float64   `json:"upper_bound"`
	Model       string    `json:"model"` // "ewma", "holt_winters", etc.
}

// Simulation represents what-if analysis results
type Simulation struct {
	ID           string              `json:"id"`
	Scenario     SimulationScenario  `json:"scenario"`
	Timeline     []SimulationPoint   `json:"timeline"`
	Summary      SimulationSummary   `json:"summary"`
	SLOAnalysis  SLOAnalysis         `json:"slo_analysis"`
	CostAnalysis CostAnalysis        `json:"cost_analysis"`
	Duration     time.Duration       `json:"duration"`
	CreatedAt    time.Time           `json:"created_at"`
}

// SimulationScenario defines the parameters for simulation
type SimulationScenario struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Plan           CapacityPlan    `json:"plan"`
	SLOOverride    *SLO            `json:"slo_override,omitempty"`
	TrafficPattern TrafficPattern  `json:"traffic_pattern"`
	Duration       time.Duration   `json:"duration"`
	Granularity    time.Duration   `json:"granularity"`
}

// SimulationPoint represents a single point in time during simulation
type SimulationPoint struct {
	Timestamp     time.Time     `json:"timestamp"`
	Workers       int           `json:"workers"`
	ArrivalRate   float64       `json:"arrival_rate"`
	ServiceRate   float64       `json:"service_rate"`
	Backlog       int           `json:"backlog"`
	Latency       time.Duration `json:"latency"`
	Utilization   float64       `json:"utilization"`
	Cost          float64       `json:"cost"`
	SLOViolation  bool          `json:"slo_violation"`
}

// SimulationSummary provides aggregate results
type SimulationSummary struct {
	AvgBacklog      float64       `json:"avg_backlog"`
	MaxBacklog      int           `json:"max_backlog"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	SLOViolations   int           `json:"slo_violations"`
	SLOAchievement  float64       `json:"slo_achievement"` // 0.0-1.0
	TotalCost       float64       `json:"total_cost"`
	AvgUtilization  float64       `json:"avg_utilization"`
	EfficiencyScore float64       `json:"efficiency_score"` // Cost-effectiveness metric
}

// SLOAnalysis provides detailed SLO compliance analysis
type SLOAnalysis struct {
	LatencyCompliance    float64           `json:"latency_compliance"`    // % of time within target
	BacklogCompliance    float64           `json:"backlog_compliance"`    // % of time below max
	AvailabilityAchieved float64           `json:"availability_achieved"` // Actual availability
	ErrorBudgetUsed      float64           `json:"error_budget_used"`     // % of error budget consumed
	ViolationPeriods     []ViolationPeriod `json:"violation_periods"`
	RiskScore            float64           `json:"risk_score"` // Overall risk assessment
}

// ViolationPeriod represents a time range when SLO was violated
type ViolationPeriod struct {
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Duration   time.Duration `json:"duration"`
	Type       string        `json:"type"`        // "latency", "backlog", "availability"
	Severity   string        `json:"severity"`    // "minor", "major", "critical"
	MaxValue   float64       `json:"max_value"`   // Peak violation value
	Impact     string        `json:"impact"`      // Description of impact
}

// TrafficPattern defines arrival rate patterns for simulation
type TrafficPattern struct {
	Type       PatternType     `json:"type"`
	BaseRate   float64         `json:"base_rate"`
	Amplitude  float64         `json:"amplitude"`  // For sinusoidal patterns
	Period     time.Duration   `json:"period"`     // For periodic patterns
	Spikes     []TrafficSpike  `json:"spikes"`     // Discrete traffic events
	Trend      float64         `json:"trend"`      // Growth/decline rate
	Noise      float64         `json:"noise"`      // Random variation (0.0-1.0)
}

// PatternType defines the shape of traffic patterns
type PatternType string

const (
	PatternConstant   PatternType = "constant"
	PatternSinusoidal PatternType = "sinusoidal"
	PatternLinear     PatternType = "linear"
	PatternSpiky      PatternType = "spiky"
	PatternDaily      PatternType = "daily"
	PatternWeekly     PatternType = "weekly"
	PatternCustom     PatternType = "custom"
)

// TrafficSpike represents a sudden increase in arrival rate
type TrafficSpike struct {
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Magnitude float64       `json:"magnitude"` // Multiplier (2.0 = 2x normal rate)
	Shape     SpikeShape    `json:"shape"`
}

// SpikeShape defines how traffic spikes evolve over time
type SpikeShape string

const (
	SpikeInstant   SpikeShape = "instant"   // Immediate jump
	SpikeLinear    SpikeShape = "linear"    // Linear ramp up/down
	SpikeExp       SpikeShape = "exp"       // Exponential growth/decay
	SpikeBell      SpikeShape = "bell"      // Bell curve
)

// PlannerConfig configures the capacity planner behavior
type PlannerConfig struct {
	// Forecasting
	ForecastWindow    time.Duration `json:"forecast_window"`    // How far ahead to predict
	ForecastModel     string        `json:"forecast_model"`     // "ewma", "holt_winters"
	HistoryWindow     time.Duration `json:"history_window"`     // Historical data to use
	SeasonalPeriod    time.Duration `json:"seasonal_period"`    // Daily, weekly patterns

	// Safety
	SafetyMargin      float64       `json:"safety_margin"`      // Additional capacity (0.1 = 10%)
	ConfidenceThreshold float64     `json:"confidence_threshold"` // Min confidence for auto-apply
	MaxStepSize       int           `json:"max_step_size"`      // Max workers per step
	CooldownPeriod    time.Duration `json:"cooldown_period"`    // Min time between actions

	// Limits
	MinWorkers        int           `json:"min_workers"`        // Absolute minimum
	MaxWorkers        int           `json:"max_workers"`        // Absolute maximum

	// Cost
	WorkerCostPerHour float64       `json:"worker_cost_per_hour"` // $/worker/hour
	ViolationCostPerHour float64    `json:"violation_cost_per_hour"` // $/hour during SLO violation

	// Thresholds
	ScaleUpThreshold  float64       `json:"scale_up_threshold"`  // Utilization to trigger scale up
	ScaleDownThreshold float64      `json:"scale_down_threshold"` // Utilization to trigger scale down

	// Anomaly Detection
	AnomalyThreshold  float64       `json:"anomaly_threshold"`   // Z-score for anomaly detection
	SpikeThreshold    float64       `json:"spike_threshold"`     // Multiplier for spike detection

	// Model Parameters
	QueueingModel     string        `json:"queueing_model"`      // "mm1", "mmc", "mgc"
	ServiceTimeModel  string        `json:"service_time_model"`  // "exponential", "general"
}

// QueueingResult represents the output of queueing theory calculations
type QueueingResult struct {
	Utilization     float64       `json:"utilization"`      // ρ = λ/(c×μ)
	QueueLength     float64       `json:"queue_length"`     // L_q
	WaitTime        time.Duration `json:"wait_time"`        // W_q
	ResponseTime    time.Duration `json:"response_time"`    // W = W_q + service time
	Throughput      float64       `json:"throughput"`       // Effective job completion rate
	Capacity        int           `json:"capacity"`         // Workers needed for target performance
	Model           string        `json:"model"`            // Which model was used
	Confidence      float64       `json:"confidence"`       // Model confidence
	Assumptions     []string      `json:"assumptions"`      // Model assumptions
}

// PlannerState represents the internal state of the planner
type PlannerState struct {
	LastPlan        *CapacityPlan `json:"last_plan"`
	LastUpdate      time.Time     `json:"last_update"`
	LastScaling     time.Time     `json:"last_scaling"`
	CooldownUntil   time.Time     `json:"cooldown_until"`
	AnomalyDetected bool          `json:"anomaly_detected"`
	AnomalyStart    time.Time     `json:"anomaly_start"`
	BaselineMetrics Metrics       `json:"baseline_metrics"`
	RecentHistory   []Metrics     `json:"recent_history"`
	ForecastCache   []Forecast    `json:"forecast_cache"`
	ConfigVersion   int           `json:"config_version"`
}

// PlanRequest represents a request to generate a capacity plan
type PlanRequest struct {
	QueueName      string        `json:"queue_name"`
	CurrentMetrics Metrics       `json:"current_metrics"`
	SLO            SLO           `json:"slo"`
	Config         PlannerConfig `json:"config"`
	ForceRegen     bool          `json:"force_regen"`      // Ignore cache
	WhatIfScenario *Simulation   `json:"what_if_scenario"` // Optional what-if parameters
}

// PlanResponse contains the generated capacity plan and metadata
type PlanResponse struct {
	Plan           CapacityPlan    `json:"plan"`
	Forecast       []Forecast      `json:"forecast"`
	QueueingResult QueueingResult  `json:"queueing_result"`
	Recommendations []string       `json:"recommendations"`
	Warnings       []string        `json:"warnings"`
	GenerationTime time.Duration   `json:"generation_time"`
	CacheHit       bool            `json:"cache_hit"`
}

// Error types for capacity planning
type PlannerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *PlannerError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error codes
const (
	ErrInvalidMetrics      = "INVALID_METRICS"
	ErrInsufficientHistory = "INSUFFICIENT_HISTORY"
	ErrForecastFailed      = "FORECAST_FAILED"
	ErrModelNotSupported   = "MODEL_NOT_SUPPORTED"
	ErrConfigInvalid       = "CONFIG_INVALID"
	ErrSLOUnachievable     = "SLO_UNACHIEVABLE"
	ErrCapacityLimitExceeded = "CAPACITY_LIMIT_EXCEEDED"
	ErrCooldownActive      = "COOLDOWN_ACTIVE"
	ErrAnomalyDetected     = "ANOMALY_DETECTED"
)

// Helper functions for error creation
func NewPlannerError(code, message string, cause error) *PlannerError {
	return &PlannerError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Constants for default values
const (
	DefaultForecastWindow      = 60 * time.Minute
	DefaultHistoryWindow       = 24 * time.Hour
	DefaultSafetyMargin        = 0.15 // 15%
	DefaultConfidenceThreshold = 0.85 // 85%
	DefaultMaxStepSize         = 15   // workers
	DefaultCooldownPeriod      = 5 * time.Minute
	DefaultScaleUpThreshold    = 0.80 // 80%
	DefaultScaleDownThreshold  = 0.60 // 60%
	DefaultAnomalyThreshold    = 3.0  // Z-score
	DefaultSpikeThreshold      = 2.0  // 2x normal
)