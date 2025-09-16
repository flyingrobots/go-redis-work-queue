package canary_deployments

import (
	"context"
	"time"
)

// DeploymentStatus represents the current state of a canary deployment
type DeploymentStatus string

const (
	StatusActive      DeploymentStatus = "active"
	StatusPromoting   DeploymentStatus = "promoting"
	StatusRollingBack DeploymentStatus = "rolling_back"
	StatusCompleted   DeploymentStatus = "completed"
	StatusFailed      DeploymentStatus = "failed"
	StatusPaused      DeploymentStatus = "paused"
)

// CanaryHealth represents the overall health assessment of a canary deployment
type CanaryHealth string

const (
	HealthyCanary   CanaryHealth = "healthy"
	WarningCanary   CanaryHealth = "warning"
	FailingCanary   CanaryHealth = "failing"
	UnknownCanary   CanaryHealth = "unknown"
)

// WorkerStatus represents the current status of a worker
type WorkerStatus string

const (
	WorkerHealthy     WorkerStatus = "healthy"
	WorkerDegraded    WorkerStatus = "degraded"
	WorkerUnhealthy   WorkerStatus = "unhealthy"
	WorkerUnreachable WorkerStatus = "unreachable"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	InfoAlert     AlertLevel = "info"
	WarningAlert  AlertLevel = "warning"
	CriticalAlert AlertLevel = "critical"
)

// AlertAction represents recommended actions for alerts
type AlertAction string

const (
	NoAction        AlertAction = "no_action"
	SuggestPause    AlertAction = "suggest_pause"
	SuggestRollback AlertAction = "suggest_rollback"
	ForceRollback   AlertAction = "force_rollback"
)

// RoutingStrategy defines how traffic is split between stable and canary
type RoutingStrategy string

const (
	SplitQueueStrategy RoutingStrategy = "split_queue"
	StreamGroupStrategy RoutingStrategy = "stream_group"
	HashRingStrategy   RoutingStrategy = "hash_ring"
)

// CanaryDeployment represents a single canary deployment
type CanaryDeployment struct {
	ID              string            `json:"id"`
	QueueName       string            `json:"queue_name"`
	TenantID        string            `json:"tenant_id,omitempty"`
	StableVersion   string            `json:"stable_version"`
	CanaryVersion   string            `json:"canary_version"`
	CurrentPercent  int               `json:"current_percent"`
	TargetPercent   int               `json:"target_percent"`
	Status          DeploymentStatus  `json:"status"`
	StartTime       time.Time         `json:"start_time"`
	LastUpdate      time.Time         `json:"last_update"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`

	// Metrics
	StableMetrics   *MetricsSnapshot  `json:"stable_metrics,omitempty"`
	CanaryMetrics   *MetricsSnapshot  `json:"canary_metrics,omitempty"`

	// Configuration
	Config          *CanaryConfig     `json:"config"`
	Rules           []PromotionRule   `json:"rules"`

	// Metadata
	CreatedBy       string            `json:"created_by,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// CanaryConfig holds configuration for a canary deployment
type CanaryConfig struct {
	RoutingStrategy     RoutingStrategy   `json:"routing_strategy"`
	StickyRouting       bool              `json:"sticky_routing"`
	AutoPromotion       bool              `json:"auto_promotion"`
	MaxCanaryDuration   time.Duration     `json:"max_canary_duration"`
	MinCanaryDuration   time.Duration     `json:"min_canary_duration"`
	PromotionStages     []PromotionStage  `json:"promotion_stages,omitempty"`
	RollbackThresholds  SLOThresholds     `json:"rollback_thresholds"`
	DrainTimeout        time.Duration     `json:"drain_timeout"`
	MetricsWindow       time.Duration     `json:"metrics_window"`
	AlertWebhooks       []string          `json:"alert_webhooks,omitempty"`
	Exemptions          []string          `json:"exemptions,omitempty"`
}

// PromotionStage defines a stage in automatic promotion
type PromotionStage struct {
	Percentage   int           `json:"percentage"`
	Duration     time.Duration `json:"duration"`
	AutoPromote  bool          `json:"auto_promote"`
	Conditions   SLOThresholds `json:"conditions"`
}

// SLOThresholds defines the thresholds for SLO monitoring
type SLOThresholds struct {
	MaxErrorRateIncrease    float64       `json:"max_error_rate_increase"`    // Percentage points
	MaxLatencyIncrease      float64       `json:"max_latency_increase"`       // Percentage
	MaxThroughputDecrease   float64       `json:"max_throughput_decrease"`    // Percentage
	MinSuccessRate          float64       `json:"min_success_rate"`           // Percentage
	MaxMemoryIncrease       float64       `json:"max_memory_increase"`        // Percentage
	RequiredSampleSize      int           `json:"required_sample_size"`       // Minimum jobs to evaluate
}

// MetricsSnapshot represents a point-in-time metrics snapshot
type MetricsSnapshot struct {
	Timestamp       time.Time     `json:"timestamp"`
	WindowStart     time.Time     `json:"window_start"`
	WindowEnd       time.Time     `json:"window_end"`

	// Core metrics
	JobCount        int64         `json:"job_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	ErrorRate       float64       `json:"error_rate"`         // 0-100
	SuccessRate     float64       `json:"success_rate"`       // 0-100

	// Latency metrics (in milliseconds)
	AvgLatency      float64       `json:"avg_latency"`
	P50Latency      float64       `json:"p50_latency"`
	P95Latency      float64       `json:"p95_latency"`
	P99Latency      float64       `json:"p99_latency"`
	MaxLatency      float64       `json:"max_latency"`

	// Throughput metrics
	JobsPerSecond   float64       `json:"jobs_per_second"`

	// Resource metrics
	AvgMemoryMB     float64       `json:"avg_memory_mb"`
	PeakMemoryMB    float64       `json:"peak_memory_mb"`
	AvgCPUPercent   float64       `json:"avg_cpu_percent"`

	// Queue metrics
	QueueDepth      int64         `json:"queue_depth"`
	DeadLetters     int64         `json:"dead_letters"`

	// Additional context
	WorkerCount     int           `json:"worker_count"`
	Version         string        `json:"version"`
}

// PromotionRule interface for defining promotion conditions
type PromotionRule interface {
	Evaluate(stable, canary *MetricsSnapshot) bool
	Description() string
	Type() string
}

// ErrorRateRule implements promotion rule based on error rate
type ErrorRateRule struct {
	MaxIncrease float64 `json:"max_increase"` // Percentage points
}

func (err *ErrorRateRule) Evaluate(stable, canary *MetricsSnapshot) bool {
	if canary.JobCount < 10 { // Need minimum sample size
		return false
	}
	increase := canary.ErrorRate - stable.ErrorRate
	return increase <= err.MaxIncrease
}

func (err *ErrorRateRule) Description() string {
	return "Error rate increase must not exceed threshold"
}

func (err *ErrorRateRule) Type() string {
	return "error_rate"
}

// LatencyRule implements promotion rule based on latency
type LatencyRule struct {
	MaxIncrease float64 `json:"max_increase"` // Percentage
}

func (lr *LatencyRule) Evaluate(stable, canary *MetricsSnapshot) bool {
	if stable.P95Latency == 0 || canary.JobCount < 10 {
		return canary.P95Latency < 5000 // 5s fallback
	}
	increase := (canary.P95Latency - stable.P95Latency) / stable.P95Latency * 100
	return increase <= lr.MaxIncrease
}

func (lr *LatencyRule) Description() string {
	return "P95 latency increase must not exceed threshold"
}

func (lr *LatencyRule) Type() string {
	return "latency"
}

// DurationRule implements promotion rule based on minimum duration
type DurationRule struct {
	MinDuration time.Duration `json:"min_duration"`
}

func (dr *DurationRule) Evaluate(stable, canary *MetricsSnapshot) bool {
	return time.Since(canary.WindowStart) >= dr.MinDuration
}

func (dr *DurationRule) Description() string {
	return "Minimum canary duration must be met"
}

func (dr *DurationRule) Type() string {
	return "duration"
}

// ThroughputRule implements promotion rule based on throughput
type ThroughputRule struct {
	MaxDecrease float64 `json:"max_decrease"` // Percentage
}

func (tr *ThroughputRule) Evaluate(stable, canary *MetricsSnapshot) bool {
	if stable.JobsPerSecond == 0 || canary.JobCount < 10 {
		return true // Can't compare without stable baseline
	}
	decrease := (stable.JobsPerSecond - canary.JobsPerSecond) / stable.JobsPerSecond * 100
	return decrease <= tr.MaxDecrease
}

func (tr *ThroughputRule) Description() string {
	return "Throughput decrease must not exceed threshold"
}

func (tr *ThroughputRule) Type() string {
	return "throughput"
}

// WorkerInfo represents information about a worker
type WorkerInfo struct {
	ID          string            `json:"id"`
	Version     string            `json:"version"`
	Lane        string            `json:"lane"` // "stable" or "canary"
	Queues      []string          `json:"queues"`
	LastSeen    time.Time         `json:"last_seen"`
	Status      WorkerStatus      `json:"status"`
	Metrics     WorkerMetrics     `json:"metrics"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// WorkerMetrics holds performance metrics for a worker
type WorkerMetrics struct {
	JobsProcessed   int64     `json:"jobs_processed"`
	JobsSucceeded   int64     `json:"jobs_succeeded"`
	JobsFailed      int64     `json:"jobs_failed"`
	AvgProcessingTime float64 `json:"avg_processing_time"`
	MemoryUsageMB   float64   `json:"memory_usage_mb"`
	CPUUsagePercent float64   `json:"cpu_usage_percent"`
	LastJobAt       time.Time `json:"last_job_at"`
}

// IsHealthy checks if a worker is considered healthy
func (wi *WorkerInfo) IsHealthy() bool {
	return wi.Status == WorkerHealthy &&
		time.Since(wi.LastSeen) < 2*time.Minute
}

// QueueSplitter handles split queue routing strategy
type QueueSplitter struct {
	StableQueue string `json:"stable_queue"`
	CanaryQueue string `json:"canary_queue"`
	Percentage  int    `json:"percentage"` // 0-100, percentage going to canary
	StickyHash  bool   `json:"sticky_hash"` // Use job ID hash for consistency
}

// StreamCanaryConfig handles stream group routing strategy
type StreamCanaryConfig struct {
	StreamKey      string  `json:"stream_key"`
	StableGroup    string  `json:"stable_group"`
	CanaryGroup    string  `json:"canary_group"`
	CanaryWeight   float64 `json:"canary_weight"` // 0.0-1.0
}

// Alert represents a canary deployment alert
type Alert struct {
	ID          string      `json:"id"`
	DeploymentID string     `json:"deployment_id"`
	Level       AlertLevel  `json:"level"`
	Message     string      `json:"message"`
	Details     interface{} `json:"details,omitempty"`
	Action      AlertAction `json:"action"`
	Timestamp   time.Time   `json:"timestamp"`
	Resolved    bool        `json:"resolved"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Name        string    `json:"name"`
	Passing     bool      `json:"passing"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
}

// CanaryHealthStatus represents the overall health of a canary
type CanaryHealthStatus struct {
	OverallStatus   CanaryHealth    `json:"overall_status"`
	ErrorRateCheck  HealthCheck     `json:"error_rate_check"`
	LatencyCheck    HealthCheck     `json:"latency_check"`
	ThroughputCheck HealthCheck     `json:"throughput_check"`
	DurationCheck   HealthCheck     `json:"duration_check"`
	SampleSizeCheck HealthCheck     `json:"sample_size_check"`
	LastEvaluation  time.Time       `json:"last_evaluation"`
}

// AllChecksPass returns true if all health checks are passing
func (chs *CanaryHealthStatus) AllChecksPass() bool {
	return chs.ErrorRateCheck.Passing &&
		chs.LatencyCheck.Passing &&
		chs.ThroughputCheck.Passing &&
		chs.DurationCheck.Passing &&
		chs.SampleSizeCheck.Passing
}

// GetFailureReason returns a human-readable failure reason
func (chs *CanaryHealthStatus) GetFailureReason() string {
	if chs.AllChecksPass() {
		return "All checks passing"
	}

	if !chs.ErrorRateCheck.Passing {
		return chs.ErrorRateCheck.Message
	}
	if !chs.LatencyCheck.Passing {
		return chs.LatencyCheck.Message
	}
	if !chs.ThroughputCheck.Passing {
		return chs.ThroughputCheck.Message
	}
	if !chs.DurationCheck.Passing {
		return chs.DurationCheck.Message
	}
	if !chs.SampleSizeCheck.Passing {
		return chs.SampleSizeCheck.Message
	}

	return "Unknown failure"
}

// DeploymentEvent represents an event in the deployment lifecycle
type DeploymentEvent struct {
	ID           string            `json:"id"`
	DeploymentID string            `json:"deployment_id"`
	Type         string            `json:"type"`
	Message      string            `json:"message"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Job represents a job in the queue system
type Job struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Queue    string                 `json:"queue"`
	Payload  map[string]interface{} `json:"payload"`
	Priority int                    `json:"priority"`
	TenantID string                 `json:"tenant_id,omitempty"`

	// Timing
	CreatedAt   time.Time  `json:"created_at"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Execution context
	WorkerID    string `json:"worker_id,omitempty"`
	Version     string `json:"version,omitempty"`
	Lane        string `json:"lane,omitempty"`

	// Retry information
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	LastError   string    `json:"last_error,omitempty"`

	// Metadata
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CanaryManager interface defines the main canary deployment management operations
type CanaryManager interface {
	// Deployment lifecycle
	CreateDeployment(ctx context.Context, config *CanaryConfig) (*CanaryDeployment, error)
	GetDeployment(ctx context.Context, id string) (*CanaryDeployment, error)
	ListDeployments(ctx context.Context) ([]*CanaryDeployment, error)
	UpdateDeploymentPercentage(ctx context.Context, id string, percentage int) error
	PromoteDeployment(ctx context.Context, id string) error
	RollbackDeployment(ctx context.Context, id string, reason string) error
	DeleteDeployment(ctx context.Context, id string) error

	// Monitoring and health
	GetDeploymentHealth(ctx context.Context, id string) (*CanaryHealthStatus, error)
	GetDeploymentMetrics(ctx context.Context, id string) (*MetricsSnapshot, *MetricsSnapshot, error)
	GetDeploymentEvents(ctx context.Context, id string) ([]*DeploymentEvent, error)

	// Worker management
	RegisterWorker(ctx context.Context, info *WorkerInfo) error
	GetWorkers(ctx context.Context, lane string) ([]*WorkerInfo, error)
	UpdateWorkerStatus(ctx context.Context, workerID string, status WorkerStatus) error

	// Control operations
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// MetricsCollector interface for collecting metrics from workers and queues
type MetricsCollector interface {
	CollectSnapshot(ctx context.Context, queue string, version string, window time.Duration) (*MetricsSnapshot, error)
	GetHistoricalMetrics(ctx context.Context, queue string, version string, since time.Time) ([]*MetricsSnapshot, error)
}

// Router interface for routing jobs to appropriate queues
type Router interface {
	RouteJob(ctx context.Context, job *Job) (string, error)
	UpdateRoutingPercentage(ctx context.Context, queue string, percentage int) error
	GetRoutingStats(ctx context.Context, queue string) (map[string]int64, error)
}

// Alerter interface for sending alerts about canary deployments
type Alerter interface {
	SendAlert(ctx context.Context, alert *Alert) error
	GetAlerts(ctx context.Context, deploymentID string) ([]*Alert, error)
	ResolveAlert(ctx context.Context, alertID string) error
}