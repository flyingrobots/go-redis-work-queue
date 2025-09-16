// Copyright 2025 James Ross
package patternedloadgenerator

import (
	"time"
)

// PatternType defines types of load patterns
type PatternType string

const (
	PatternSine     PatternType = "sine"
	PatternBurst    PatternType = "burst"
	PatternRamp     PatternType = "ramp"
	PatternConstant PatternType = "constant"
	PatternStep     PatternType = "step"
	PatternCustom   PatternType = "custom"
)

// LoadPattern defines a load generation pattern
type LoadPattern struct {
	Type        PatternType            `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Duration    time.Duration          `json:"duration"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// SineParameters defines parameters for sine wave pattern
type SineParameters struct {
	Amplitude  float64       `json:"amplitude"`  // Peak variation from baseline
	Baseline   float64       `json:"baseline"`   // Center line of wave
	Period     time.Duration `json:"period"`     // Time for one complete cycle
	Phase      float64       `json:"phase"`      // Phase shift in radians
}

// BurstParameters defines parameters for burst pattern
type BurstParameters struct {
	BurstRate     float64       `json:"burst_rate"`     // Jobs per second during burst
	BurstDuration time.Duration `json:"burst_duration"` // How long each burst lasts
	IdleDuration  time.Duration `json:"idle_duration"`  // Time between bursts
	BurstCount    int           `json:"burst_count"`    // Number of bursts (0 = infinite)
}

// RampParameters defines parameters for ramp pattern
type RampParameters struct {
	StartRate    float64       `json:"start_rate"`    // Initial jobs per second
	EndRate      float64       `json:"end_rate"`      // Final jobs per second
	RampDuration time.Duration `json:"ramp_duration"` // Time to ramp from start to end
	HoldDuration time.Duration `json:"hold_duration"` // Time to hold at end rate
	RampDown     bool          `json:"ramp_down"`     // Whether to ramp back down
}

// StepParameters defines parameters for step pattern
type StepParameters struct {
	Steps        []StepLevel   `json:"steps"`         // Step levels
	StepDuration time.Duration `json:"step_duration"` // Duration of each step
	Repeat       bool          `json:"repeat"`        // Whether to repeat the pattern
}

// StepLevel defines a single step in step pattern
type StepLevel struct {
	Rate     float64       `json:"rate"`     // Jobs per second for this step
	Duration time.Duration `json:"duration"` // Override duration for this step
}

// CustomParameters defines parameters for custom pattern
type CustomParameters struct {
	Points []DataPoint `json:"points"` // Time series of rates
	Loop   bool        `json:"loop"`   // Whether to loop the pattern
}

// DataPoint represents a point in time with a rate
type DataPoint struct {
	Time time.Duration `json:"time"` // Time offset from start
	Rate float64       `json:"rate"` // Jobs per second at this time
}

// Guardrails defines safety limits for load generation
type Guardrails struct {
	MaxRate           float64       `json:"max_rate"`            // Maximum jobs per second
	MaxTotal          int64         `json:"max_total"`           // Maximum total jobs to generate
	MaxDuration       time.Duration `json:"max_duration"`        // Maximum run duration
	MaxQueueDepth     int64         `json:"max_queue_depth"`     // Stop if queue exceeds this
	EmergencyStopFile string        `json:"emergency_stop_file"` // File to check for emergency stop
	RateLimitWindow   time.Duration `json:"rate_limit_window"`   // Window for rate limiting
}

// LoadProfile defines a complete load generation profile
type LoadProfile struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Patterns    []LoadPattern          `json:"patterns"`
	Guardrails  Guardrails             `json:"guardrails"`
	JobTemplate map[string]interface{} `json:"job_template"`
	QueueName   string                 `json:"queue_name"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     string                 `json:"version"`
}

// GeneratorStatus represents the current status of the generator
type GeneratorStatus struct {
	Running       bool          `json:"running"`
	Pattern       PatternType   `json:"pattern"`
	StartedAt     time.Time     `json:"started_at"`
	Duration      time.Duration `json:"duration"`
	JobsGenerated int64         `json:"jobs_generated"`
	CurrentRate   float64       `json:"current_rate"`
	TargetRate    float64       `json:"target_rate"`
	Errors        int64         `json:"errors"`
	LastError     string        `json:"last_error,omitempty"`
}

// MetricsSnapshot captures metrics at a point in time
type MetricsSnapshot struct {
	Timestamp     time.Time `json:"timestamp"`
	TargetRate    float64   `json:"target_rate"`
	ActualRate    float64   `json:"actual_rate"`
	JobsGenerated int64     `json:"jobs_generated"`
	QueueDepth    int64     `json:"queue_depth"`
	Errors        int64     `json:"errors"`
}

// GeneratorConfig defines configuration for the load generator
type GeneratorConfig struct {
	DefaultGuardrails Guardrails    `json:"default_guardrails"`
	MetricsInterval   time.Duration `json:"metrics_interval"`
	ProfilesPath      string        `json:"profiles_path"`
	EnableCharts      bool          `json:"enable_charts"`
	ChartUpdateRate   time.Duration `json:"chart_update_rate"`
	MaxHistoryPoints  int           `json:"max_history_points"`
}

// JobGenerator defines the interface for generating jobs
type JobGenerator interface {
	GenerateJob() (interface{}, error)
}

// SimpleJobGenerator generates simple test jobs
type SimpleJobGenerator struct {
	Template map[string]interface{}
	Counter  int64
}

// ChartData represents data for visualization
type ChartData struct {
	TimePoints   []time.Time `json:"time_points"`
	TargetRates  []float64   `json:"target_rates"`
	ActualRates  []float64   `json:"actual_rates"`
	QueueDepths  []int64     `json:"queue_depths"`
	ErrorCounts  []int64     `json:"error_counts"`
}

// ControlCommand represents a control command for the generator
type ControlCommand string

const (
	CommandStart  ControlCommand = "start"
	CommandStop   ControlCommand = "stop"
	CommandPause  ControlCommand = "pause"
	CommandResume ControlCommand = "resume"
	CommandReset  ControlCommand = "reset"
)

// ControlMessage represents a control message
type ControlMessage struct {
	Command   ControlCommand         `json:"command"`
	ProfileID string                 `json:"profile_id,omitempty"`
	Pattern   *LoadPattern           `json:"pattern,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// GeneratorEvent represents an event from the generator
type GeneratorEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EventType defines types of generator events
const (
	EventStarted         = "started"
	EventStopped         = "stopped"
	EventPaused          = "paused"
	EventResumed         = "resumed"
	EventPatternChanged  = "pattern_changed"
	EventGuardrailHit    = "guardrail_hit"
	EventError           = "error"
	EventMetricsUpdate   = "metrics_update"
)