// Copyright 2025 James Ross
package queuesnapshotesting

import (
	"time"
)

// MetricType defines types of metrics that can be captured
type MetricType string

const (
	MetricQueueLength   MetricType = "queue_length"
	MetricProcessedJobs MetricType = "processed_jobs"
	MetricFailedJobs    MetricType = "failed_jobs"
	MetricWorkerCount   MetricType = "worker_count"
	MetricAvgLatency    MetricType = "avg_latency"
)

// JobState represents the state of a job in the queue
type JobState struct {
	ID          string                 `json:"id"`
	QueueName   string                 `json:"queue_name"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Attempts    int                    `json:"attempts"`
	MaxRetries  int                    `json:"max_retries"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// WorkerState represents the state of a worker
type WorkerState struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	CurrentJobID string    `json:"current_job_id,omitempty"`
	LastSeen     time.Time `json:"last_seen"`
	ProcessedCount int64   `json:"processed_count"`
	ErrorCount   int64     `json:"error_count"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// QueueState represents the complete state of a queue
type QueueState struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Length          int64          `json:"length"`
	Config          map[string]interface{} `json:"config"`
	RateLimits      map[string]interface{} `json:"rate_limits,omitempty"`
	DeadLetterQueue string         `json:"dead_letter_queue,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// Snapshot represents a complete system state snapshot
type Snapshot struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	Tags        []string               `json:"tags,omitempty"`

	// State data
	Queues      []QueueState           `json:"queues"`
	Jobs        []JobState             `json:"jobs"`
	Workers     []WorkerState          `json:"workers"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`

	// Metadata
	Context     map[string]interface{} `json:"context,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Checksum    string                 `json:"checksum"`
	Compressed  bool                   `json:"compressed"`
	SizeBytes   int64                  `json:"size_bytes"`
}

// DiffResult represents the difference between two snapshots
type DiffResult struct {
	LeftID      string       `json:"left_id"`
	RightID     string       `json:"right_id"`
	Timestamp   time.Time    `json:"timestamp"`

	// Summary statistics
	TotalChanges int         `json:"total_changes"`
	Added        int         `json:"added"`
	Removed      int         `json:"removed"`
	Modified     int         `json:"modified"`

	// Detailed changes
	QueueChanges  []Change    `json:"queue_changes"`
	JobChanges    []Change    `json:"job_changes"`
	WorkerChanges []Change    `json:"worker_changes"`
	MetricChanges []Change    `json:"metric_changes"`

	// Semantic analysis
	SemanticChanges []SemanticChange `json:"semantic_changes,omitempty"`
}

// Change represents a single change between snapshots
type Change struct {
	Type        ChangeType             `json:"type"`
	Path        string                 `json:"path"`
	OldValue    interface{}            `json:"old_value,omitempty"`
	NewValue    interface{}            `json:"new_value,omitempty"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact,omitempty"`
}

// ChangeType defines types of changes
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeRemoved  ChangeType = "removed"
	ChangeModified ChangeType = "modified"
	ChangeMoved    ChangeType = "moved"
)

// SemanticChange represents a high-level semantic change
type SemanticChange struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Components  []string  `json:"components"`
}

// SnapshotConfig defines configuration for snapshot operations
type SnapshotConfig struct {
	// Storage settings
	StoragePath     string        `json:"storage_path"`
	MaxSnapshots    int           `json:"max_snapshots"`
	RetentionDays   int           `json:"retention_days"`
	CompressLevel   int           `json:"compress_level"`

	// Diff settings
	IgnoreTimestamps bool         `json:"ignore_timestamps"`
	IgnoreIDs        bool         `json:"ignore_ids"`
	IgnoreWorkerIDs  bool         `json:"ignore_worker_ids"`
	CustomIgnores    []string     `json:"custom_ignores,omitempty"`

	// Performance settings
	MaxJobsPerSnapshot int        `json:"max_jobs_per_snapshot"`
	SampleRate         float64    `json:"sample_rate"`
	TimeoutSeconds     int        `json:"timeout_seconds"`
}

// AssertionResult represents the result of a snapshot assertion
type AssertionResult struct {
	Passed      bool         `json:"passed"`
	Message     string       `json:"message"`
	Differences []Change     `json:"differences,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
}

// SnapshotMetadata provides metadata about stored snapshots
type SnapshotMetadata struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
	SizeBytes   int64        `json:"size_bytes"`
	Tags        []string     `json:"tags,omitempty"`
	Environment string       `json:"environment,omitempty"`
}

// SnapshotFilter defines filters for searching snapshots
type SnapshotFilter struct {
	Name        string       `json:"name,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Environment string       `json:"environment,omitempty"`
	CreatedAfter  time.Time  `json:"created_after,omitempty"`
	CreatedBefore time.Time  `json:"created_before,omitempty"`
	MaxResults  int          `json:"max_results,omitempty"`
}

// Album represents a collection of related snapshots
type Album struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	SnapshotIDs []string     `json:"snapshot_ids"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Tags        []string     `json:"tags,omitempty"`
}