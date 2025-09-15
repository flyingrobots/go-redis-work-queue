package dlqremediationui

import (
	"encoding/json"
	"time"
)

// DLQEntry represents a single entry in the dead letter queue
type DLQEntry struct {
	ID          string                 `json:"id"`
	JobID       string                 `json:"job_id"`
	Type        string                 `json:"type"`
	Queue       string                 `json:"queue"`
	Payload     json.RawMessage        `json:"payload"`
	Error       ErrorDetails           `json:"error"`
	Metadata    JobMetadata            `json:"metadata"`
	Attempts    []AttemptRecord        `json:"attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	FailedAt    time.Time              `json:"failed_at"`
	LastRetryAt *time.Time             `json:"last_retry_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Priority    int                    `json:"priority"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Size        int64                  `json:"size"` // Payload size in bytes
}

// ErrorDetails contains detailed information about the job failure
type ErrorDetails struct {
	Type        string            `json:"type"`
	Message     string            `json:"message"`
	StackTrace  string            `json:"stack_trace,omitempty"`
	Code        string            `json:"code,omitempty"`
	Category    string            `json:"category,omitempty"`
	Retryable   bool              `json:"retryable"`
	Context     map[string]string `json:"context,omitempty"`
	Fingerprint string            `json:"fingerprint"` // For pattern grouping
}

// JobMetadata contains execution context and system information
type JobMetadata struct {
	WorkerID       string                 `json:"worker_id"`
	WorkerVersion  string                 `json:"worker_version"`
	ProcessingTime time.Duration          `json:"processing_time"`
	MemoryUsed     int64                  `json:"memory_used"`
	StartedAt      time.Time              `json:"started_at"`
	EndedAt        time.Time              `json:"ended_at"`
	Headers        map[string]string      `json:"headers,omitempty"`
	Trace          TraceInfo              `json:"trace,omitempty"`
	Environment    string                 `json:"environment,omitempty"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// AttemptRecord tracks each retry attempt
type AttemptRecord struct {
	Number     int           `json:"number"`
	StartedAt  time.Time     `json:"started_at"`
	EndedAt    time.Time     `json:"ended_at"`
	Duration   time.Duration `json:"duration"`
	WorkerID   string        `json:"worker_id"`
	Error      string        `json:"error,omitempty"`
	Success    bool          `json:"success"`
	RetryDelay time.Duration `json:"retry_delay,omitempty"`
}

// TraceInfo contains distributed tracing information
type TraceInfo struct {
	TraceID  string `json:"trace_id,omitempty"`
	SpanID   string `json:"span_id,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
}

// DLQFilter defines filtering criteria for DLQ queries
type DLQFilter struct {
	// Time-based filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	FailedAfter   *time.Time `json:"failed_after,omitempty"`
	FailedBefore  *time.Time `json:"failed_before,omitempty"`

	// Job-based filters
	JobID      string   `json:"job_id,omitempty"`
	JobTypes   []string `json:"job_types,omitempty"`
	Queues     []string `json:"queues,omitempty"`
	TenantIDs  []string `json:"tenant_ids,omitempty"`
	Tags       []string `json:"tags,omitempty"`

	// Error-based filters
	ErrorTypes      []string `json:"error_types,omitempty"`
	ErrorMessages   []string `json:"error_messages,omitempty"`
	ErrorCategories []string `json:"error_categories,omitempty"`
	Retryable       *bool    `json:"retryable,omitempty"`

	// Worker-based filters
	WorkerIDs      []string `json:"worker_ids,omitempty"`
	WorkerVersions []string `json:"worker_versions,omitempty"`

	// Retry-based filters
	MinRetryCount *int `json:"min_retry_count,omitempty"`
	MaxRetryCount *int `json:"max_retry_count,omitempty"`

	// Priority filters
	MinPriority *int `json:"min_priority,omitempty"`
	MaxPriority *int `json:"max_priority,omitempty"`

	// Size filters
	MinSize *int64 `json:"min_size,omitempty"`
	MaxSize *int64 `json:"max_size,omitempty"`

	// Full-text search
	SearchQuery string `json:"search_query,omitempty"`

	// Pattern-based filters
	ErrorFingerprints []string `json:"error_fingerprints,omitempty"`
	SimilarTo         string   `json:"similar_to,omitempty"` // Job ID to find similar failures
}

// DLQSortField defines available sorting options
type DLQSortField string

const (
	SortByCreatedAt    DLQSortField = "created_at"
	SortByFailedAt     DLQSortField = "failed_at"
	SortByJobType      DLQSortField = "job_type"
	SortByQueue        DLQSortField = "queue"
	SortByRetryCount   DLQSortField = "retry_count"
	SortByPriority     DLQSortField = "priority"
	SortBySize         DLQSortField = "size"
	SortByErrorType    DLQSortField = "error_type"
	SortByWorkerID     DLQSortField = "worker_id"
)

// SortOrder defines sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// DLQSort defines sorting criteria
type DLQSort struct {
	Field DLQSortField `json:"field"`
	Order SortOrder    `json:"order"`
}

// PageInfo contains pagination information
type PageInfo struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	Total    int  `json:"total"`
	HasNext  bool `json:"has_next"`
	HasPrev  bool `json:"has_prev"`
}

// DLQListResponse represents the response from listing DLQ entries
type DLQListResponse struct {
	Entries    []*DLQEntry      `json:"entries"`
	PageInfo   PageInfo         `json:"page_info"`
	Filter     *DLQFilter       `json:"filter,omitempty"`
	Sort       *DLQSort         `json:"sort,omitempty"`
	Stats      *DLQStats        `json:"stats,omitempty"`
	Patterns   []*ErrorPattern  `json:"patterns,omitempty"`
	Suggestions []*ActionSuggestion `json:"suggestions,omitempty"`
}

// DLQStats provides aggregate statistics
type DLQStats struct {
	TotalEntries     int                       `json:"total_entries"`
	TotalSize        int64                     `json:"total_size"`
	EntryCountByType map[string]int            `json:"entry_count_by_type"`
	EntryCountByQueue map[string]int           `json:"entry_count_by_queue"`
	ErrorTypeCount   map[string]int            `json:"error_type_count"`
	AvgRetryCount    float64                   `json:"avg_retry_count"`
	OldestEntry      *time.Time                `json:"oldest_entry,omitempty"`
	NewestEntry      *time.Time                `json:"newest_entry,omitempty"`
	GrowthRate       *DLQGrowthRate            `json:"growth_rate,omitempty"`
	TopErrors        []*ErrorFrequency         `json:"top_errors,omitempty"`
}

// DLQGrowthRate tracks DLQ growth patterns
type DLQGrowthRate struct {
	LastHour      int     `json:"last_hour"`
	Last24Hours   int     `json:"last_24_hours"`
	Last7Days     int     `json:"last_7_days"`
	HourlyRate    float64 `json:"hourly_rate"`
	PredictedFull *time.Time `json:"predicted_full,omitempty"`
}

// ErrorFrequency represents error frequency data
type ErrorFrequency struct {
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Trend       string  `json:"trend"` // "increasing", "decreasing", "stable"
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Count       int       `json:"count"`
	Percentage  float64   `json:"percentage"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Fingerprint string    `json:"fingerprint"`
	Confidence  float64   `json:"confidence"`
	RootCause   string    `json:"root_cause,omitempty"`
	Examples    []string  `json:"examples,omitempty"` // Job IDs
	AffectedTypes []string `json:"affected_types,omitempty"`
	AffectedQueues []string `json:"affected_queues,omitempty"`
}

// ActionSuggestion represents a suggested remediation action
type ActionSuggestion struct {
	ID          string                 `json:"id"`
	Type        ActionType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Impact      string                 `json:"impact"`
	Risk        string                 `json:"risk"`
	Effort      string                 `json:"effort"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	AppliesTo   []string               `json:"applies_to,omitempty"` // Job IDs or patterns
}

// ActionType defines types of remediation actions
type ActionType string

const (
	ActionRetry           ActionType = "retry"
	ActionRetryWithDelay  ActionType = "retry_with_delay"
	ActionModifyAndRetry  ActionType = "modify_and_retry"
	ActionPurge           ActionType = "purge"
	ActionEscalate        ActionType = "escalate"
	ActionIgnore          ActionType = "ignore"
	ActionBulkRetry       ActionType = "bulk_retry"
	ActionBulkPurge       ActionType = "bulk_purge"
)

// BulkOperation represents a bulk operation request
type BulkOperation struct {
	Type        BulkOperationType      `json:"type"`
	JobIDs      []string               `json:"job_ids,omitempty"`
	Filter      *DLQFilter             `json:"filter,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Confirmation string                `json:"confirmation,omitempty"`
	DryRun      bool                   `json:"dry_run,omitempty"`
}

// BulkOperationType defines types of bulk operations
type BulkOperationType string

const (
	BulkRetry           BulkOperationType = "retry"
	BulkRetryWithModify BulkOperationType = "retry_with_modify"
	BulkPurge           BulkOperationType = "purge"
	BulkMove            BulkOperationType = "move"
	BulkExport          BulkOperationType = "export"
)

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	OperationID   string                 `json:"operation_id"`
	Type          BulkOperationType      `json:"type"`
	Status        OperationStatus        `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	TotalJobs     int                    `json:"total_jobs"`
	ProcessedJobs int                    `json:"processed_jobs"`
	SuccessJobs   int                    `json:"success_jobs"`
	FailedJobs    int                    `json:"failed_jobs"`
	SkippedJobs   int                    `json:"skipped_jobs"`
	Errors        []string               `json:"errors,omitempty"`
	Progress      float64                `json:"progress"`
	EstimatedTime *time.Duration         `json:"estimated_time,omitempty"`
	Results       map[string]interface{} `json:"results,omitempty"`
}

// OperationStatus defines the status of an operation
type OperationStatus string

const (
	StatusPending    OperationStatus = "pending"
	StatusRunning    OperationStatus = "running"
	StatusPaused     OperationStatus = "paused"
	StatusCompleted  OperationStatus = "completed"
	StatusFailed     OperationStatus = "failed"
	StatusCancelled  OperationStatus = "cancelled"
)

// PayloadModification represents modifications to apply to job payloads
type PayloadModification struct {
	Set    map[string]interface{} `json:"set,omitempty"`
	Remove []string               `json:"remove,omitempty"`
	Transform map[string]string   `json:"transform,omitempty"` // JSONPath transformations
}

// DLQPreferences stores user preferences for the DLQ UI
type DLQPreferences struct {
	DefaultPageSize   int                   `json:"default_page_size"`
	DefaultSort       DLQSort               `json:"default_sort"`
	SavedFilters      map[string]*DLQFilter `json:"saved_filters"`
	HiddenColumns     []string              `json:"hidden_columns"`
	AutoRefresh       bool                  `json:"auto_refresh"`
	RefreshInterval   time.Duration         `json:"refresh_interval"`
	ConfirmPurge      bool                  `json:"confirm_purge"`
	ConfirmBulkOps    bool                  `json:"confirm_bulk_ops"`
	ShowPayloadPreview bool                 `json:"show_payload_preview"`
	MaxPayloadPreview int                   `json:"max_payload_preview"`
}

// DLQAnalytics provides analytical insights
type DLQAnalytics struct {
	TimeRange       TimeRange           `json:"time_range"`
	ErrorTrends     []*ErrorTrend       `json:"error_trends"`
	ResolutionStats *ResolutionStats    `json:"resolution_stats"`
	PatternHistory  []*PatternEvolution `json:"pattern_history"`
	Correlations    []*EventCorrelation `json:"correlations"`
	Predictions     []*GrowthPrediction `json:"predictions"`
}

// TimeRange defines a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ErrorTrend tracks error frequency over time
type ErrorTrend struct {
	Type       string      `json:"type"`
	Timestamps []time.Time `json:"timestamps"`
	Counts     []int       `json:"counts"`
	Trend      string      `json:"trend"`
	Regression *float64    `json:"regression,omitempty"`
}

// ResolutionStats tracks how failures are resolved
type ResolutionStats struct {
	TotalResolved     int                    `json:"total_resolved"`
	ResolutionMethods map[string]int         `json:"resolution_methods"`
	MTTR              time.Duration          `json:"mttr"` // Mean Time To Resolution
	MTTRByType        map[string]time.Duration `json:"mttr_by_type"`
	SuccessRates      map[string]float64     `json:"success_rates"`
}

// PatternEvolution tracks how error patterns change over time
type PatternEvolution struct {
	PatternID   string      `json:"pattern_id"`
	Name        string      `json:"name"`
	Timestamps  []time.Time `json:"timestamps"`
	Counts      []int       `json:"counts"`
	Confidence  []float64   `json:"confidence"`
	Lifecycle   string      `json:"lifecycle"` // "emerging", "active", "declining", "resolved"
}

// EventCorrelation represents correlation with external events
type EventCorrelation struct {
	EventType   string    `json:"event_type"`
	EventTime   time.Time `json:"event_time"`
	Description string    `json:"description"`
	Correlation float64   `json:"correlation"`
	Impact      int       `json:"impact"` // Number of failures potentially caused
}

// GrowthPrediction predicts future DLQ growth
type GrowthPrediction struct {
	Timestamp       time.Time `json:"timestamp"`
	PredictedCount  int       `json:"predicted_count"`
	ConfidenceInterval struct {
		Lower int `json:"lower"`
		Upper int `json:"upper"`
	} `json:"confidence_interval"`
	Factors []string `json:"factors"` // Contributing factors
}

// DLQManager interface defines the main operations
type DLQManager interface {
	// Listing and filtering
	ListEntries(filter *DLQFilter, sort *DLQSort, page, pageSize int) (*DLQListResponse, error)
	GetEntry(id string) (*DLQEntry, error)
	GetStats(filter *DLQFilter) (*DLQStats, error)

	// Pattern analysis
	AnalyzePatterns(filter *DLQFilter) ([]*ErrorPattern, error)
	GetSuggestions(filter *DLQFilter) ([]*ActionSuggestion, error)

	// Individual operations
	PeekEntry(id string) (*DLQEntry, error)
	RetryEntry(id string, modifications *PayloadModification) error
	PurgeEntry(id string) error

	// Bulk operations
	StartBulkOperation(operation *BulkOperation) (*BulkOperationResult, error)
	GetBulkOperationStatus(operationID string) (*BulkOperationResult, error)
	CancelBulkOperation(operationID string) error

	// Analytics
	GetAnalytics(timeRange TimeRange, filter *DLQFilter) (*DLQAnalytics, error)

	// Preferences
	GetPreferences(userID string) (*DLQPreferences, error)
	SavePreferences(userID string, prefs *DLQPreferences) error
}

// DLQStorage interface defines storage operations
type DLQStorage interface {
	// Basic CRUD
	Add(entry *DLQEntry) error
	Get(id string) (*DLQEntry, error)
	Update(id string, entry *DLQEntry) error
	Delete(id string) error

	// Querying
	List(filter *DLQFilter, sort *DLQSort, offset, limit int) ([]*DLQEntry, int, error)
	Count(filter *DLQFilter) (int, error)

	// Bulk operations
	BulkDelete(ids []string) error
	BulkUpdate(updates map[string]*DLQEntry) error

	// Analytics
	GetAggregates(filter *DLQFilter) (*DLQStats, error)
	GetTimeSeries(filter *DLQFilter, interval time.Duration) ([][2]interface{}, error)
}

// PatternAnalyzer interface defines pattern detection
type PatternAnalyzer interface {
	AnalyzePatterns(entries []*DLQEntry) ([]*ErrorPattern, error)
	DetectAnomalies(entries []*DLQEntry) ([]*ErrorPattern, error)
	SuggestActions(patterns []*ErrorPattern) ([]*ActionSuggestion, error)
	UpdatePattern(pattern *ErrorPattern) error
}

// RemediationEngine interface defines remediation operations
type RemediationEngine interface {
	RetryJob(entry *DLQEntry, modifications *PayloadModification) error
	PurgeJob(entry *DLQEntry) error
	BulkRetry(entries []*DLQEntry, modifications *PayloadModification) (*BulkOperationResult, error)
	BulkPurge(entries []*DLQEntry) (*BulkOperationResult, error)
	ValidateModifications(entry *DLQEntry, modifications *PayloadModification) error
}

// Configuration for the DLQ system
type Config struct {
	// Storage configuration
	RedisAddr     string        `json:"redis_addr"`
	RedisPassword string        `json:"redis_password"`
	RedisDB       int           `json:"redis_db"`

	// Performance settings
	MaxPageSize       int           `json:"max_page_size"`
	DefaultPageSize   int           `json:"default_page_size"`
	MaxPayloadSize    int64         `json:"max_payload_size"`
	CacheTimeout      time.Duration `json:"cache_timeout"`

	// Pattern analysis settings
	MinPatternOccurrences int     `json:"min_pattern_occurrences"`
	PatternConfidenceThreshold float64 `json:"pattern_confidence_threshold"`
	AnalysisWindow    time.Duration `json:"analysis_window"`

	// Safety settings
	RequireConfirmation bool `json:"require_confirmation"`
	MaxBulkOperationSize int `json:"max_bulk_operation_size"`
	PurgeRetentionDays   int `json:"purge_retention_days"`

	// UI settings
	AutoRefreshInterval time.Duration `json:"auto_refresh_interval"`
	ShowStackTraces     bool          `json:"show_stack_traces"`
	EnableAnalytics     bool          `json:"enable_analytics"`
}