// Copyright 2025 James Ross
package chaosharness

import (
	"sync"
	"time"
)

// InjectorType defines the type of fault injection
type InjectorType string

const (
	InjectorLatency      InjectorType = "latency"
	InjectorError        InjectorType = "error"
	InjectorPanic        InjectorType = "panic"
	InjectorPartialFail  InjectorType = "partial_fail"
	InjectorResourceHog  InjectorType = "resource_hog"
	InjectorRedisLatency InjectorType = "redis_latency"
	InjectorRedisDrop    InjectorType = "redis_drop"
)

// InjectorScope defines the scope of injection
type InjectorScope string

const (
	ScopeGlobal  InjectorScope = "global"
	ScopeWorker  InjectorScope = "worker"
	ScopeQueue   InjectorScope = "queue"
	ScopeTenant  InjectorScope = "tenant"
)

// FaultInjector represents a configured fault injection
type FaultInjector struct {
	ID          string        `json:"id"`
	Type        InjectorType  `json:"type"`
	Scope       InjectorScope `json:"scope"`
	ScopeValue  string        `json:"scope_value,omitempty"`
	Enabled     bool          `json:"enabled"`
	Probability float64       `json:"probability"` // 0.0 to 1.0
	Parameters  map[string]interface{} `json:"parameters"`
	TTL         time.Duration `json:"ttl,omitempty"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	CreatedBy   string        `json:"created_by"`
	mu          sync.RWMutex
}

// ChaosScenario defines a chaos testing scenario
type ChaosScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Duration    time.Duration          `json:"duration"`
	Stages      []ScenarioStage        `json:"stages"`
	Metrics     *ScenarioMetrics       `json:"metrics,omitempty"`
	Status      ScenarioStatus         `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	EndedAt     *time.Time             `json:"ended_at,omitempty"`
	Guardrails  ScenarioGuardrails     `json:"guardrails"`
}

// ScenarioStage represents a stage in a chaos scenario
type ScenarioStage struct {
	Name       string          `json:"name"`
	Duration   time.Duration   `json:"duration"`
	Injectors  []FaultInjector `json:"injectors"`
	LoadConfig *LoadConfig     `json:"load_config,omitempty"`
}

// LoadConfig defines load generation parameters
type LoadConfig struct {
	RPS         int           `json:"rps"`
	Pattern     LoadPattern   `json:"pattern"`
	BurstSize   int           `json:"burst_size,omitempty"`
	BurstDelay  time.Duration `json:"burst_delay,omitempty"`
}

// LoadPattern defines the pattern of load generation
type LoadPattern string

const (
	LoadConstant LoadPattern = "constant"
	LoadLinear   LoadPattern = "linear"
	LoadSine     LoadPattern = "sine"
	LoadSpike    LoadPattern = "spike"
	LoadRandom   LoadPattern = "random"
)

// ScenarioStatus represents the status of a scenario
type ScenarioStatus string

const (
	StatusPending   ScenarioStatus = "pending"
	StatusRunning   ScenarioStatus = "running"
	StatusCompleted ScenarioStatus = "completed"
	StatusFailed    ScenarioStatus = "failed"
	StatusAborted   ScenarioStatus = "aborted"
)

// ScenarioMetrics tracks metrics during scenario execution
type ScenarioMetrics struct {
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	InjectedFaults     int64                  `json:"injected_faults"`
	RecoveryTime       time.Duration          `json:"recovery_time"`
	BacklogSize        int64                  `json:"backlog_size"`
	ErrorRate          float64                `json:"error_rate"`
	LatencyP50         time.Duration          `json:"latency_p50"`
	LatencyP95         time.Duration          `json:"latency_p95"`
	LatencyP99         time.Duration          `json:"latency_p99"`
	TimeSeriesData     []TimeSeriesPoint      `json:"time_series_data"`
	mu                 sync.RWMutex
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp    time.Time              `json:"timestamp"`
	Metrics      map[string]float64     `json:"metrics"`
	ActiveFaults []string               `json:"active_faults"`
}

// ScenarioGuardrails defines safety limits for scenarios
type ScenarioGuardrails struct {
	MaxErrorRate      float64       `json:"max_error_rate"`      // Abort if exceeded
	MaxLatencyP99     time.Duration `json:"max_latency_p99"`      // Abort if exceeded
	MaxBacklogSize    int64         `json:"max_backlog_size"`     // Abort if exceeded
	RequireConfirm    bool          `json:"require_confirm"`      // Require user confirmation
	AllowProduction   bool          `json:"allow_production"`     // Allow in production
	AutoAbortOnPanic  bool          `json:"auto_abort_on_panic"`  // Auto-abort on panic
}

// InjectorConfig holds runtime configuration for injectors
type InjectorConfig struct {
	Enabled         bool          `json:"enabled"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxTTL          time.Duration `json:"max_ttl"`
	AllowProduction bool          `json:"allow_production"`
}

// ChaosReport represents a chaos testing report
type ChaosReport struct {
	ScenarioID   string                 `json:"scenario_id"`
	ScenarioName string                 `json:"scenario_name"`
	ExecutedAt   time.Time              `json:"executed_at"`
	Duration     time.Duration          `json:"duration"`
	Result       ScenarioResult         `json:"result"`
	Metrics      *ScenarioMetrics       `json:"metrics"`
	Findings     []Finding              `json:"findings"`
	Recommends   []string               `json:"recommendations"`
}

// ScenarioResult represents the outcome of a scenario
type ScenarioResult string

const (
	ResultPassed ScenarioResult = "passed"
	ResultFailed ScenarioResult = "failed"
	ResultPartial ScenarioResult = "partial"
)

// Finding represents an issue discovered during chaos testing
type Finding struct {
	Severity    FindingSeverity `json:"severity"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Impact      string          `json:"impact"`
	Evidence    interface{}     `json:"evidence,omitempty"`
}

// FindingSeverity represents the severity of a finding
type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "critical"
	SeverityHigh     FindingSeverity = "high"
	SeverityMedium   FindingSeverity = "medium"
	SeverityLow      FindingSeverity = "low"
	SeverityInfo     FindingSeverity = "info"
)

// WorkerInjectorState tracks injector state per worker
type WorkerInjectorState struct {
	WorkerID         string                 `json:"worker_id"`
	ActiveInjectors  map[string]*FaultInjector `json:"active_injectors"`
	FaultsInjected   int64                  `json:"faults_injected"`
	LastFaultAt      *time.Time             `json:"last_fault_at,omitempty"`
	mu               sync.RWMutex
}
