// Copyright 2025 James Ross
package policysimulator

import (
	"time"
)


// SimulatorConfig configures the simulation parameters
type SimulatorConfig struct {
	SimulationDuration time.Duration `json:"simulation_duration"` // How long to simulate
	TimeStep          time.Duration `json:"time_step"`           // Granularity of simulation
	MaxWorkers        int           `json:"max_workers"`         // Maximum concurrent workers
	RedisPoolSize     int           `json:"redis_pool_size"`     // Redis connection pool size
}

// PolicyConfig represents configurable queue policies
type PolicyConfig struct {
	// Retry policies
	MaxRetries      int           `json:"max_retries"`
	InitialBackoff  time.Duration `json:"initial_backoff"`
	MaxBackoff      time.Duration `json:"max_backoff"`
	BackoffStrategy string        `json:"backoff_strategy"` // exponential, linear, constant

	// Rate limiting
	MaxRatePerSecond float64 `json:"max_rate_per_second"`
	BurstSize        int     `json:"burst_size"`

	// Concurrency controls
	MaxConcurrency int `json:"max_concurrency"`
	QueueSize      int `json:"queue_size"`

	// Timeout settings
	ProcessingTimeout time.Duration `json:"processing_timeout"`
	AckTimeout        time.Duration `json:"ack_timeout"`

	// Dead letter queue
	DLQEnabled    bool   `json:"dlq_enabled"`
	DLQThreshold  int    `json:"dlq_threshold"`
	DLQQueueName  string `json:"dlq_queue_name"`
}

// TrafficPattern defines expected load patterns
type TrafficPattern struct {
	Name        string                `json:"name"`
	Type        TrafficPatternType    `json:"type"`
	BaseRate    float64               `json:"base_rate"`    // Messages per second baseline
	Variations  []TrafficVariation    `json:"variations"`   // Spikes, drops, seasonal patterns
	Duration    time.Duration         `json:"duration"`     // How long this pattern lasts
	Probability float64               `json:"probability"`  // Likelihood of this pattern occurring
	Metadata    map[string]interface{} `json:"metadata"`    // Additional pattern-specific data
}

// TrafficPatternType defines different types of load patterns
type TrafficPatternType string

const (
	TrafficConstant     TrafficPatternType = "constant"      // Steady state
	TrafficLinear       TrafficPatternType = "linear"        // Linear increase/decrease
	TrafficSpike        TrafficPatternType = "spike"         // Sudden burst
	TrafficSeasonal     TrafficPatternType = "seasonal"      // Periodic patterns
	TrafficBursty       TrafficPatternType = "bursty"        // Random bursts
	TrafficExponential  TrafficPatternType = "exponential"   // Exponential growth/decay
)

// TrafficVariation represents changes in traffic over time
type TrafficVariation struct {
	StartTime    time.Duration `json:"start_time"`    // When variation begins
	EndTime      time.Duration `json:"end_time"`      // When variation ends
	Multiplier   float64       `json:"multiplier"`    // Rate multiplier (1.0 = no change)
	Description  string        `json:"description"`   // Human readable description
}

// SimulationMetrics contains predicted performance metrics
type SimulationMetrics struct {
	// Queue metrics
	AvgQueueDepth     float64 `json:"avg_queue_depth"`
	MaxQueueDepth     int     `json:"max_queue_depth"`
	AvgWaitTime       float64 `json:"avg_wait_time_ms"`
	P95WaitTime       float64 `json:"p95_wait_time_ms"`
	P99WaitTime       float64 `json:"p99_wait_time_ms"`

	// Throughput metrics
	MessagesProcessed int     `json:"messages_processed"`
	ProcessingRate    float64 `json:"processing_rate"`    // Messages per second
	Utilization       float64 `json:"utilization"`        // Worker utilization %

	// Error metrics
	FailureRate       float64 `json:"failure_rate"`       // Percentage of failed jobs
	RetryRate         float64 `json:"retry_rate"`         // Percentage requiring retries
	DLQRate           float64 `json:"dlq_rate"`           // Percentage sent to DLQ

	// Resource metrics
	AvgMemoryUsage    float64 `json:"avg_memory_usage_mb"`
	PeakMemoryUsage   float64 `json:"peak_memory_usage_mb"`
	AvgCPUUsage       float64 `json:"avg_cpu_usage_percent"`
	RedisConnections  int     `json:"redis_connections"`

	// Timestamps
	SimulationStart   time.Time `json:"simulation_start"`
	SimulationEnd     time.Time `json:"simulation_end"`
	Duration          float64   `json:"duration_seconds"`
}

// SimulationResult contains the complete simulation output
type SimulationResult struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Config      *SimulatorConfig    `json:"config"`
	Policies    *PolicyConfig       `json:"policies"`
	Pattern     *TrafficPattern     `json:"pattern"`
	Metrics     *SimulationMetrics  `json:"metrics"`
	Timeline    []TimelineSnapshot  `json:"timeline"`
	Warnings    []string            `json:"warnings"`
	CreatedAt   time.Time           `json:"created_at"`
	Status      SimulationStatus    `json:"status"`
}

// TimelineSnapshot captures metrics at a point in time
type TimelineSnapshot struct {
	Timestamp       time.Time `json:"timestamp"`
	QueueDepth      int       `json:"queue_depth"`
	ActiveWorkers   int       `json:"active_workers"`
	ProcessingRate  float64   `json:"processing_rate"`
	FailureRate     float64   `json:"failure_rate"`
	MemoryUsage     float64   `json:"memory_usage_mb"`
	CPUUsage        float64   `json:"cpu_usage_percent"`
}

// SimulationStatus represents the current state of a simulation
type SimulationStatus string

const (
	StatusPending   SimulationStatus = "pending"
	StatusRunning   SimulationStatus = "running"
	StatusCompleted SimulationStatus = "completed"
	StatusFailed    SimulationStatus = "failed"
	StatusCancelled SimulationStatus = "cancelled"
)

// PolicyChange represents a proposed change to apply
type PolicyChange struct {
	ID              string                 `json:"id"`
	Description     string                 `json:"description"`
	Changes         map[string]interface{} `json:"changes"`        // Field -> new value
	PreviousValues  map[string]interface{} `json:"previous_values"` // For rollback
	AppliedAt       *time.Time             `json:"applied_at,omitempty"`
	RolledBackAt    *time.Time             `json:"rolled_back_at,omitempty"`
	AppliedBy       string                 `json:"applied_by"`
	Status          ChangeStatus           `json:"status"`
	AuditLog        []AuditEntry           `json:"audit_log"`
}

// ChangeStatus represents the status of a policy change
type ChangeStatus string

const (
	ChangeStatusProposed    ChangeStatus = "proposed"
	ChangeStatusSimulated   ChangeStatus = "simulated"
	ChangeStatusApproved    ChangeStatus = "approved"
	ChangeStatusApplied     ChangeStatus = "applied"
	ChangeStatusRolledBack  ChangeStatus = "rolled_back"
	ChangeStatusFailed      ChangeStatus = "failed"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Action    string      `json:"action"`
	User      string      `json:"user"`
	Details   interface{} `json:"details"`
}

// SimulationRequest represents a request to run a simulation
type SimulationRequest struct {
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Policies       *PolicyConfig    `json:"policies"`
	TrafficPattern *TrafficPattern  `json:"traffic_pattern"`
	Config         *SimulatorConfig `json:"config,omitempty"`
}

// PolicySimulatorError represents errors from the policy simulator
type PolicySimulatorError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *PolicySimulatorError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewPolicySimulatorError creates a new policy simulator error
func NewPolicySimulatorError(code, message string) *PolicySimulatorError {
	return &PolicySimulatorError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to a policy simulator error
func (e *PolicySimulatorError) WithDetails(details string) *PolicySimulatorError {
	return &PolicySimulatorError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// Predefined errors
var (
	ErrInvalidConfig        = NewPolicySimulatorError("INVALID_CONFIG", "invalid configuration")
	ErrInvalidPolicy        = NewPolicySimulatorError("INVALID_POLICY", "invalid policy configuration")
	ErrInvalidTrafficPattern = NewPolicySimulatorError("INVALID_TRAFFIC_PATTERN", "invalid traffic pattern")
	ErrSimulationFailed     = NewPolicySimulatorError("SIMULATION_FAILED", "simulation execution failed")
	ErrPolicyNotFound       = NewPolicySimulatorError("POLICY_NOT_FOUND", "policy configuration not found")
	ErrChangeNotFound       = NewPolicySimulatorError("CHANGE_NOT_FOUND", "policy change not found")
	ErrUnauthorized         = NewPolicySimulatorError("UNAUTHORIZED", "insufficient permissions")
	ErrApplyFailed          = NewPolicySimulatorError("APPLY_FAILED", "failed to apply policy change")
	ErrRollbackFailed       = NewPolicySimulatorError("ROLLBACK_FAILED", "failed to rollback policy change")
)

// QueueingModel represents a queueing theory model for simulation
type QueueingModel struct {
	Type        QueueingModelType `json:"type"`
	ServiceRate float64           `json:"service_rate"`    // μ (messages/second per worker)
	ArrivalRate float64           `json:"arrival_rate"`    // λ (messages/second)
	Servers     int               `json:"servers"`         // Number of workers
	Capacity    int               `json:"capacity"`        // Queue capacity (0 = unlimited)
	Parameters  map[string]float64 `json:"parameters"`     // Model-specific parameters
}

// QueueingModelType defines different queueing models
type QueueingModelType string

const (
	ModelMM1     QueueingModelType = "M/M/1"     // Markovian arrival/service, 1 server
	ModelMMC     QueueingModelType = "M/M/c"     // Markovian arrival/service, c servers
	ModelMM1K    QueueingModelType = "M/M/1/K"   // M/M/1 with finite capacity K
	ModelMMCK    QueueingModelType = "M/M/c/K"   // M/M/c with finite capacity K
	ModelMG1     QueueingModelType = "M/G/1"     // Markovian arrival, general service
	ModelGGC     QueueingModelType = "G/G/c"     // General arrival/service
	ModelSimple  QueueingModelType = "Simple"    // Simplified Little's Law model
)

// ChartData represents data for visualization
type ChartData struct {
	Title       string      `json:"title"`
	Type        ChartType   `json:"type"`
	XAxis       AxisConfig  `json:"x_axis"`
	YAxis       AxisConfig  `json:"y_axis"`
	Series      []ChartSeries `json:"series"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// ChartType defines different chart types
type ChartType string

const (
	ChartLine      ChartType = "line"
	ChartBar       ChartType = "bar"
	ChartArea      ChartType = "area"
	ChartScatter   ChartType = "scatter"
	ChartHistogram ChartType = "histogram"
)

// AxisConfig configures chart axes
type AxisConfig struct {
	Label  string  `json:"label"`
	Unit   string  `json:"unit,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Format string  `json:"format,omitempty"`
}

// ChartSeries represents a data series in a chart
type ChartSeries struct {
	Name   string        `json:"name"`
	Color  string        `json:"color,omitempty"`
	Points []ChartPoint  `json:"points"`
	Style  SeriesStyle   `json:"style,omitempty"`
}

// ChartPoint represents a single data point
type ChartPoint struct {
	X     float64     `json:"x"`
	Y     float64     `json:"y"`
	Label string      `json:"label,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
}

// SeriesStyle defines visual styling for chart series
type SeriesStyle struct {
	LineWidth   int    `json:"line_width,omitempty"`
	Dash        string `json:"dash,omitempty"`
	MarkerSize  int    `json:"marker_size,omitempty"`
	MarkerShape string `json:"marker_shape,omitempty"`
}

// Annotation adds contextual information to charts
type Annotation struct {
	Type        AnnotationType `json:"type"`
	X           float64        `json:"x,omitempty"`
	Y           float64        `json:"y,omitempty"`
	Text        string         `json:"text"`
	Color       string         `json:"color,omitempty"`
	Description string         `json:"description,omitempty"`
}

// AnnotationType defines different annotation types
type AnnotationType string

const (
	AnnotationPoint     AnnotationType = "point"
	AnnotationLine      AnnotationType = "line"
	AnnotationRectangle AnnotationType = "rectangle"
	AnnotationText      AnnotationType = "text"
)

// SimulationAssumptions documents model assumptions and limitations
type SimulationAssumptions struct {
	ModelType           QueueingModelType `json:"model_type"`
	Assumptions         []string          `json:"assumptions"`
	Limitations         []string          `json:"limitations"`
	AccuracyEstimate    string            `json:"accuracy_estimate"`
	RecommendedUseCase  string            `json:"recommended_use_case"`
	NotRecommendedFor   []string          `json:"not_recommended_for"`
}