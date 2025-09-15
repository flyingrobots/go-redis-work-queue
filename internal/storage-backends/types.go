package storage

import (
	"context"
	"time"
)

// Job represents a job in the queue system
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Queue       string                 `json:"queue"`
	Payload     interface{}            `json:"payload"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// QueueBackend defines the interface for queue storage implementations
type QueueBackend interface {
	// Core operations
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error)
	Ack(ctx context.Context, jobID string) error
	Nack(ctx context.Context, jobID string, requeue bool) error

	// Inspection operations
	Length(ctx context.Context) (int64, error)
	Peek(ctx context.Context, offset int64) (*Job, error)

	// DLQ management
	Move(ctx context.Context, jobID string, targetQueue string) error

	// Advanced operations (capability-gated)
	Iter(ctx context.Context, opts IterOptions) (Iterator, error)

	// Metadata and management
	Capabilities() BackendCapabilities
	Stats(ctx context.Context) (*BackendStats, error)
	Health(ctx context.Context) HealthStatus
	Close() error
}

// BackendCapabilities describes what features a backend supports
type BackendCapabilities struct {
	AtomicAck          bool `json:"atomic_ack"`           // Guaranteed single processing
	ConsumerGroups     bool `json:"consumer_groups"`      // Multiple consumer support
	Replay             bool `json:"replay"`               // Historical job access
	IdempotentEnqueue  bool `json:"idempotent_enqueue"`   // Duplicate detection
	Transactions       bool `json:"transactions"`         // Multi-operation atomicity
	Persistence        bool `json:"persistence"`          // Survives restarts
	Clustering         bool `json:"clustering"`           // Distributed operation
	TimeToLive         bool `json:"time_to_live"`         // Automatic expiration
	Prioritization     bool `json:"prioritization"`       // Priority queues
	BatchOperations    bool `json:"batch_operations"`     // Bulk enqueue/dequeue
}

// DequeueOptions configures dequeue behavior
type DequeueOptions struct {
	Timeout       time.Duration `json:"timeout"`
	ConsumerID    string        `json:"consumer_id,omitempty"`
	ConsumerGroup string        `json:"consumer_group,omitempty"`
	Count         int           `json:"count,omitempty"`
}

// IterOptions configures iteration behavior
type IterOptions struct {
	StartID   string    `json:"start_id,omitempty"`
	EndID     string    `json:"end_id,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Count     int64     `json:"count,omitempty"`
	Reverse   bool      `json:"reverse"`
}

// Iterator provides sequential access to jobs
type Iterator interface {
	Next() bool
	Job() *Job
	Error() error
	Close() error
}

// BackendStats provides metrics about backend performance
type BackendStats struct {
	// Universal metrics
	EnqueueRate     float64       `json:"enqueue_rate"`
	DequeueRate     float64       `json:"dequeue_rate"`
	ErrorRate       float64       `json:"error_rate"`
	QueueDepth      int64         `json:"queue_depth"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P99Latency      time.Duration `json:"p99_latency"`

	// Backend-specific metrics
	StreamLength    *int64     `json:"stream_length,omitempty"`    // Streams only
	ConsumerLag     *int64     `json:"consumer_lag,omitempty"`     // Streams only
	ClusterShards   *int       `json:"cluster_shards,omitempty"`   // Cluster only
	MemoryUsage     *int64     `json:"memory_usage,omitempty"`     // KeyDB/Dragonfly
	ConnectionPool  *PoolStats `json:"connection_pool,omitempty"`

	// Timing metrics
	LastEnqueue time.Time `json:"last_enqueue"`
	LastDequeue time.Time `json:"last_dequeue"`
	LastError   time.Time `json:"last_error"`
}

// PoolStats provides connection pool metrics
type PoolStats struct {
	Active   int           `json:"active"`
	Idle     int           `json:"idle"`
	Total    int           `json:"total"`
	MaxOpen  int           `json:"max_open"`
	MaxIdle  int           `json:"max_idle"`
	WaitTime time.Duration `json:"wait_time"`
}

// HealthStatus describes backend health
type HealthStatus struct {
	Status    string            `json:"status"` // healthy, degraded, unhealthy
	Message   string            `json:"message,omitempty"`
	Error     error             `json:"error,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CheckedAt time.Time         `json:"checked_at"`
}

// BackendFactory creates backend instances
type BackendFactory interface {
	Create(config interface{}) (QueueBackend, error)
	Validate(config interface{}) error
}

// BackendConfig holds common configuration for all backends
type BackendConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Name     string                 `json:"name" yaml:"name"`
	URL      string                 `json:"url" yaml:"url"`
	Database int                    `json:"database" yaml:"database"`
	Options  map[string]interface{} `json:"options" yaml:"options"`
}

// MigrationOptions configures queue migration between backends
type MigrationOptions struct {
	SourceBackend string        `json:"source_backend"`
	TargetBackend string        `json:"target_backend"`
	DrainFirst    bool          `json:"drain_first"`
	Timeout       time.Duration `json:"timeout"`
	BatchSize     int           `json:"batch_size"`
	VerifyData    bool          `json:"verify_data"`
	DryRun        bool          `json:"dry_run"`
}

// MigrationStatus tracks migration progress
type MigrationStatus struct {
	Phase         string    `json:"phase"`
	TotalJobs     int64     `json:"total_jobs"`
	MigratedJobs  int64     `json:"migrated_jobs"`
	FailedJobs    int64     `json:"failed_jobs"`
	Progress      float64   `json:"progress"`
	StartedAt     time.Time `json:"started_at"`
	EstimatedETA  time.Time `json:"estimated_eta,omitempty"`
	LastError     error     `json:"last_error,omitempty"`
}

// OutboxEvent represents an event for Kafka outbox pattern
type OutboxEvent struct {
	JobID     string      `json:"job_id"`
	Queue     string      `json:"queue"`
	Operation string      `json:"operation"` // enqueue, ack, nack, move
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// Backend type constants
const (
	BackendTypeRedisLists   = "redis-lists"
	BackendTypeRedisStreams = "redis-streams"
	BackendTypeKeyDB        = "keydb"
	BackendTypeDragonfly    = "dragonfly"
)

// Health status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// Migration phase constants
const (
	MigrationPhaseValidation = "validation"
	MigrationPhaseDraining   = "draining"
	MigrationPhaseCopying    = "copying"
	MigrationPhaseVerifying  = "verifying"
	MigrationPhaseCompleted  = "completed"
	MigrationPhaseFailed     = "failed"
)