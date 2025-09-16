// Copyright 2025 James Ross
package archives

import (
	"context"
	"time"
)

// ArchiveJob represents a completed job in the archive system
type ArchiveJob struct {
	JobID           string                 `json:"job_id" ch:"job_id"`
	Queue           string                 `json:"queue" ch:"queue"`
	Priority        int                    `json:"priority" ch:"priority"`
	EnqueuedAt      time.Time              `json:"enqueued_at" ch:"enqueued_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty" ch:"started_at"`
	CompletedAt     time.Time              `json:"completed_at" ch:"completed_at"`
	Outcome         JobOutcome             `json:"outcome" ch:"outcome"`
	RetryCount      int                    `json:"retry_count" ch:"retry_count"`
	WorkerID        string                 `json:"worker_id" ch:"worker_id"`
	PayloadSize     int64                  `json:"payload_size" ch:"payload_size"`
	TraceID         string                 `json:"trace_id,omitempty" ch:"trace_id"`
	ErrorMessage    string                 `json:"error_message,omitempty" ch:"error_message"`
	ErrorCode       string                 `json:"error_code,omitempty" ch:"error_code"`
	ProcessingTime  int64                  `json:"processing_time_ms" ch:"processing_time_ms"`
	PayloadHash     string                 `json:"payload_hash,omitempty" ch:"payload_hash"`
	PayloadSnapshot []byte                 `json:"payload_snapshot,omitempty" ch:"payload_snapshot"`
	Tags            map[string]string      `json:"tags,omitempty" ch:"tags"`
	SchemaVersion   int                    `json:"schema_version" ch:"schema_version"`
	ArchivedAt      time.Time              `json:"archived_at" ch:"archived_at"`
	Tenant          string                 `json:"tenant,omitempty" ch:"tenant"`
	JobType         string                 `json:"job_type,omitempty" ch:"job_type"`
}

// JobOutcome represents the outcome of a job execution
type JobOutcome string

const (
	OutcomeSuccess JobOutcome = "success"
	OutcomeFailed  JobOutcome = "failed"
	OutcomeTimeout JobOutcome = "timeout"
	OutcomeCanceled JobOutcome = "canceled"
	OutcomeRetry   JobOutcome = "retry"
)

// ArchiveConfig represents the configuration for the archive system
type ArchiveConfig struct {
	Enabled          bool                 `json:"enabled"`
	SamplingRate     float64              `json:"sampling_rate"`
	RedisStreamKey   string               `json:"redis_stream_key"`
	BatchSize        int                  `json:"batch_size"`
	ExportInterval   time.Duration        `json:"export_interval"`
	ClickHouse       ClickHouseConfig     `json:"clickhouse"`
	S3               S3Config             `json:"s3"`
	Retention        RetentionConfig      `json:"retention"`
	SchemaVersion    int                  `json:"schema_version"`
	PayloadHandling  PayloadHandlingConfig `json:"payload_handling"`
}

// ClickHouseConfig represents ClickHouse connection and settings
type ClickHouseConfig struct {
	Enabled       bool          `json:"enabled"`
	DSN           string        `json:"dsn"`
	Database      string        `json:"database"`
	Table         string        `json:"table"`
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	MaxOpenConns  int           `json:"max_open_conns"`
	MaxIdleConns  int           `json:"max_idle_conns"`
	ConnMaxLife   time.Duration `json:"conn_max_life"`
	Compression   string        `json:"compression"`
	Async         bool          `json:"async"`
}

// S3Config represents S3/Parquet export configuration
type S3Config struct {
	Enabled         bool          `json:"enabled"`
	Bucket          string        `json:"bucket"`
	Region          string        `json:"region"`
	KeyPrefix       string        `json:"key_prefix"`
	AccessKeyID     string        `json:"access_key_id"`
	SecretAccessKey string        `json:"secret_access_key"`
	Endpoint        string        `json:"endpoint,omitempty"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	CompressionType string        `json:"compression_type"`
	PartitionBy     string        `json:"partition_by"`
}

// RetentionConfig represents data retention settings
type RetentionConfig struct {
	RedisStreamTTL  time.Duration `json:"redis_stream_ttl"`
	ArchiveWindow   time.Duration `json:"archive_window"`
	DeleteAfter     time.Duration `json:"delete_after"`
	GDPRCompliant   bool          `json:"gdpr_compliant"`
	DeleteHookURL   string        `json:"delete_hook_url,omitempty"`
}

// PayloadHandlingConfig represents how job payloads should be handled
type PayloadHandlingConfig struct {
	IncludePayload    bool     `json:"include_payload"`
	MaxPayloadSize    int64    `json:"max_payload_size"`
	RedactFields      []string `json:"redact_fields"`
	HashOnly          bool     `json:"hash_only"`
	CompressionType   string   `json:"compression_type"`
}

// ExportStatus represents the status of an export operation
type ExportStatus struct {
	ID              string              `json:"id"`
	Type            ExportType          `json:"type"`
	Status          ExportStatusType    `json:"status"`
	StartedAt       time.Time           `json:"started_at"`
	CompletedAt     *time.Time          `json:"completed_at,omitempty"`
	RecordsTotal    int64               `json:"records_total"`
	RecordsExported int64               `json:"records_exported"`
	RecordsFailed   int64               `json:"records_failed"`
	BatchesTotal    int                 `json:"batches_total"`
	BatchesExported int                 `json:"batches_exported"`
	ErrorMessage    string              `json:"error_message,omitempty"`
	LastExportAt    time.Time           `json:"last_export_at"`
	NextExportAt    time.Time           `json:"next_export_at"`
	Metrics         ExportMetrics       `json:"metrics"`
}

// ExportType represents the type of export destination
type ExportType string

const (
	ExportTypeClickHouse ExportType = "clickhouse"
	ExportTypeS3         ExportType = "s3"
)

// ExportStatusType represents the status of an export operation
type ExportStatusType string

const (
	ExportStatusPending    ExportStatusType = "pending"
	ExportStatusRunning    ExportStatusType = "running"
	ExportStatusCompleted  ExportStatusType = "completed"
	ExportStatusFailed     ExportStatusType = "failed"
	ExportStatusCanceled   ExportStatusType = "canceled"
)

// ExportMetrics represents metrics for export operations
type ExportMetrics struct {
	AvgBatchSize      float64 `json:"avg_batch_size"`
	AvgExportTime     float64 `json:"avg_export_time_ms"`
	SuccessRate       float64 `json:"success_rate"`
	ErrorRate         float64 `json:"error_rate"`
	TotalSize         int64   `json:"total_size_bytes"`
	CompressionRatio  float64 `json:"compression_ratio"`
	LastLagTime       float64 `json:"last_lag_time_ms"`
}

// ArchiveBatch represents a batch of jobs for export
type ArchiveBatch struct {
	ID          string       `json:"id"`
	Jobs        []ArchiveJob `json:"jobs"`
	CreatedAt   time.Time    `json:"created_at"`
	Size        int64        `json:"size_bytes"`
	Compressed  bool         `json:"compressed"`
	Checksum    string       `json:"checksum"`
}

// SchemaEvolution represents schema version changes
type SchemaEvolution struct {
	Version     int                    `json:"version"`
	Description string                 `json:"description"`
	Changes     []SchemaChange         `json:"changes"`
	CreatedAt   time.Time              `json:"created_at"`
	Backward    bool                   `json:"backward_compatible"`
	Migration   *MigrationInfo         `json:"migration,omitempty"`
}

// SchemaChange represents a single schema change
type SchemaChange struct {
	Type        ChangeType             `json:"type"`
	Field       string                 `json:"field"`
	OldType     string                 `json:"old_type,omitempty"`
	NewType     string                 `json:"new_type,omitempty"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default,omitempty"`
}

// ChangeType represents the type of schema change
type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeRemove ChangeType = "remove"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeRename ChangeType = "rename"
)

// MigrationInfo represents information about data migration
type MigrationInfo struct {
	Required      bool              `json:"required"`
	Script        string            `json:"script,omitempty"`
	EstimatedTime time.Duration     `json:"estimated_time"`
	RollbackScript string           `json:"rollback_script,omitempty"`
}

// QueryTemplate represents predefined queries for the archive
type QueryTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	SQL         string            `json:"sql"`
	Parameters  []QueryParameter  `json:"parameters"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
}

// QueryParameter represents a parameter in a query template
type QueryParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ArchiveStats represents statistics about the archive system
type ArchiveStats struct {
	TotalJobs       int64             `json:"total_jobs"`
	TotalSize       int64             `json:"total_size_bytes"`
	JobsByOutcome   map[JobOutcome]int64 `json:"jobs_by_outcome"`
	JobsByQueue     map[string]int64  `json:"jobs_by_queue"`
	AvgProcessTime  float64           `json:"avg_processing_time_ms"`
	OldestJob       time.Time         `json:"oldest_job"`
	NewestJob       time.Time         `json:"newest_job"`
	ExportLag       time.Duration     `json:"export_lag"`
	LastExportAt    time.Time         `json:"last_export_at"`
	ErrorRate       float64           `json:"error_rate"`
}

// GDPRDeleteRequest represents a GDPR deletion request
type GDPRDeleteRequest struct {
	ID          string    `json:"id"`
	JobID       string    `json:"job_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Reason      string    `json:"reason"`
	RequestedAt time.Time `json:"requested_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Status      string    `json:"status"`
	Records     int64     `json:"records_deleted"`
}

// Manager interface defines the long-term archives operations
type Manager interface {
	// Export operations
	ExportJobs(ctx context.Context, jobs []ArchiveJob, exportType ExportType) (*ExportStatus, error)
	GetExportStatus(ctx context.Context, id string) (*ExportStatus, error)
	ListExports(ctx context.Context, limit int, offset int) ([]ExportStatus, error)
	CancelExport(ctx context.Context, id string) error

	// Archive operations
	ArchiveJob(ctx context.Context, job ArchiveJob) error
	GetArchivedJob(ctx context.Context, jobID string) (*ArchiveJob, error)
	SearchJobs(ctx context.Context, query SearchQuery) ([]ArchiveJob, error)
	GetStats(ctx context.Context, window time.Duration) (*ArchiveStats, error)

	// Schema management
	GetSchemaVersion(ctx context.Context) (int, error)
	UpgradeSchema(ctx context.Context, newVersion int) error
	GetSchemaEvolution(ctx context.Context) ([]SchemaEvolution, error)

	// Retention management
	CleanupExpired(ctx context.Context) (int64, error)
	ProcessGDPRDelete(ctx context.Context, request GDPRDeleteRequest) error

	// Query templates
	AddQueryTemplate(ctx context.Context, template QueryTemplate) error
	GetQueryTemplates(ctx context.Context) ([]QueryTemplate, error)
	ExecuteQuery(ctx context.Context, templateName string, params map[string]interface{}) (interface{}, error)

	// System operations
	GetHealth(ctx context.Context) (map[string]interface{}, error)
	Close() error
}

// SearchQuery represents a search query for archived jobs
type SearchQuery struct {
	JobIDs      []string           `json:"job_ids,omitempty"`
	Queue       string             `json:"queue,omitempty"`
	Outcome     JobOutcome         `json:"outcome,omitempty"`
	WorkerID    string             `json:"worker_id,omitempty"`
	TraceID     string             `json:"trace_id,omitempty"`
	StartTime   *time.Time         `json:"start_time,omitempty"`
	EndTime     *time.Time         `json:"end_time,omitempty"`
	Tags        map[string]string  `json:"tags,omitempty"`
	Limit       int                `json:"limit"`
	Offset      int                `json:"offset"`
	OrderBy     string             `json:"order_by"`
	OrderDir    string             `json:"order_dir"`
}

// ExportRequest represents a request to export data
type ExportRequest struct {
	Type        ExportType         `json:"type"`
	Query       SearchQuery        `json:"query"`
	Format      string             `json:"format"`
	Compression string             `json:"compression"`
	Destination string             `json:"destination"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// Config represents the overall configuration
type Config struct {
	RedisAddr    string        `json:"redis_addr"`
	RedisDB      int           `json:"redis_db"`
	Archive      ArchiveConfig `json:"archive"`
	Monitoring   MonitoringConfig `json:"monitoring"`
	API          APIConfig     `json:"api"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled         bool          `json:"enabled"`
	MetricsInterval time.Duration `json:"metrics_interval"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds represents thresholds for alerting
type AlertThresholds struct {
	ExportLagMinutes    int     `json:"export_lag_minutes"`
	ErrorRatePercent    float64 `json:"error_rate_percent"`
	DiskUsagePercent    float64 `json:"disk_usage_percent"`
	MemoryUsagePercent  float64 `json:"memory_usage_percent"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Path    string `json:"path"`
}

// Event represents an event in the archive system
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType represents the type of archive event
type EventType string

const (
	EventTypeExportStarted   EventType = "export_started"
	EventTypeExportCompleted EventType = "export_completed"
	EventTypeExportFailed    EventType = "export_failed"
	EventTypeSchemaUpgraded  EventType = "schema_upgraded"
	EventTypeRetentionRun    EventType = "retention_run"
	EventTypeGDPRRequest     EventType = "gdpr_request"
)

// Exporter interface defines export operations
type Exporter interface {
	Export(ctx context.Context, batch ArchiveBatch) error
	GetStatus(ctx context.Context) (*ExportStatus, error)
	Close() error
}

// SchemaManager interface defines schema management operations
type SchemaManager interface {
	GetCurrentVersion(ctx context.Context) (int, error)
	Upgrade(ctx context.Context, targetVersion int) error
	IsBackwardCompatible(ctx context.Context, fromVersion, toVersion int) (bool, error)
	GetEvolution(ctx context.Context) ([]SchemaEvolution, error)
}

// RetentionManager interface defines retention management operations
type RetentionManager interface {
	Cleanup(ctx context.Context) (int64, error)
	ProcessGDPRDelete(ctx context.Context, request GDPRDeleteRequest) error
	GetRetentionPolicy(ctx context.Context) (*RetentionConfig, error)
	UpdateRetentionPolicy(ctx context.Context, policy RetentionConfig) error
}