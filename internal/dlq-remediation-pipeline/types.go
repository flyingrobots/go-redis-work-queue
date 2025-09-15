// Copyright 2025 James Ross
package dlqremediation

import (
	"encoding/json"
	"fmt"
	"time"
)

// DLQJob represents a job in the dead letter queue
type DLQJob struct {
	ID          string                 `json:"id" redis:"id"`
	JobID       string                 `json:"job_id" redis:"job_id"`
	Queue       string                 `json:"queue" redis:"queue"`
	JobType     string                 `json:"job_type" redis:"job_type"`
	Payload     json.RawMessage        `json:"payload" redis:"payload"`
	Error       string                 `json:"error" redis:"error"`
	ErrorType   string                 `json:"error_type" redis:"error_type"`
	RetryCount  int                    `json:"retry_count" redis:"retry_count"`
	FailedAt    time.Time              `json:"failed_at" redis:"failed_at"`
	CreatedAt   time.Time              `json:"created_at" redis:"created_at"`
	Metadata    map[string]interface{} `json:"metadata" redis:"metadata"`
	PayloadSize int64                  `json:"payload_size" redis:"payload_size"`
	WorkerID    string                 `json:"worker_id" redis:"worker_id"`
	TraceID     string                 `json:"trace_id" redis:"trace_id"`
}

// RemediationRule defines how to classify and remediate DLQ jobs
type RemediationRule struct {
	ID          string        `json:"id" redis:"id"`
	Name        string        `json:"name" redis:"name"`
	Description string        `json:"description" redis:"description"`
	Priority    int           `json:"priority" redis:"priority"`
	Enabled     bool          `json:"enabled" redis:"enabled"`
	CreatedAt   time.Time     `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" redis:"updated_at"`
	CreatedBy   string        `json:"created_by" redis:"created_by"`
	Matcher     RuleMatcher   `json:"matcher" redis:"matcher"`
	Actions     []Action      `json:"actions" redis:"actions"`
	Safety      SafetyLimits  `json:"safety" redis:"safety"`
	Tags        []string      `json:"tags" redis:"tags"`
	Statistics  RuleStats     `json:"statistics" redis:"statistics"`
}

// RuleMatcher defines conditions for matching DLQ jobs
type RuleMatcher struct {
	ErrorPattern    string            `json:"error_pattern,omitempty"`
	ErrorType       string            `json:"error_type,omitempty"`
	JobType         string            `json:"job_type,omitempty"`
	SourceQueue     string            `json:"source_queue,omitempty"`
	RetryCount      string            `json:"retry_count,omitempty"`      // e.g., "> 3", "= 0", "< 5"
	PayloadSize     string            `json:"payload_size,omitempty"`     // e.g., "> 1MB", "< 100KB"
	TimePattern     string            `json:"time_pattern,omitempty"`     // e.g., "business_hours", "weekends"
	PayloadMatchers []PayloadMatcher  `json:"payload_matchers,omitempty"`
	MetadataFilters map[string]string `json:"metadata_filters,omitempty"`
	AgeThreshold    string            `json:"age_threshold,omitempty"`    // e.g., "> 1h", "< 24h"
}

// PayloadMatcher matches specific fields in job payload
type PayloadMatcher struct {
	JSONPath  string      `json:"json_path"`
	Operator  string      `json:"operator"` // equals, contains, regex, exists, not_exists, gt, lt
	Value     interface{} `json:"value"`
	CaseInsensitive bool  `json:"case_insensitive,omitempty"`
}

// Action defines a remediation action to apply to matching jobs
type Action struct {
	Type        ActionType             `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Conditions  []ActionCondition      `json:"conditions,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ActionType represents the type of remediation action
type ActionType string

const (
	ActionRequeue    ActionType = "requeue"
	ActionTransform  ActionType = "transform"
	ActionRedact     ActionType = "redact"
	ActionDrop       ActionType = "drop"
	ActionRoute      ActionType = "route"
	ActionDelay      ActionType = "delay"
	ActionTag        ActionType = "tag"
	ActionNotify     ActionType = "notify"
)

// ActionCondition defines when an action should be executed
type ActionCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SafetyLimits defines rate limiting and safety constraints
type SafetyLimits struct {
	MaxPerMinute        int     `json:"max_per_minute"`
	MaxTotalPerRun      int     `json:"max_total_per_run"`
	ErrorRateThreshold  float64 `json:"error_rate_threshold"`
	BackoffOnFailure    bool    `json:"backoff_on_failure"`
	RequireConfirmation bool    `json:"require_confirmation,omitempty"`
	DryRunOnly          bool    `json:"dry_run_only,omitempty"`
}

// RuleStats tracks statistics for a remediation rule
type RuleStats struct {
	TotalMatches       int64     `json:"total_matches" redis:"total_matches"`
	SuccessfulActions  int64     `json:"successful_actions" redis:"successful_actions"`
	FailedActions      int64     `json:"failed_actions" redis:"failed_actions"`
	LastMatchedAt      time.Time `json:"last_matched_at" redis:"last_matched_at"`
	LastSuccessAt      time.Time `json:"last_success_at" redis:"last_success_at"`
	LastFailureAt      time.Time `json:"last_failure_at" redis:"last_failure_at"`
	SuccessRate        float64   `json:"success_rate" redis:"success_rate"`
	AverageLatency     float64   `json:"average_latency" redis:"average_latency"`
}

// Classification represents the result of classifying a DLQ job
type Classification struct {
	JobID       string                 `json:"job_id"`
	Category    string                 `json:"category"`
	Confidence  float64                `json:"confidence"`
	RuleID      string                 `json:"rule_id,omitempty"`
	Actions     []string               `json:"suggested_actions"`
	Reason      string                 `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ExternalClassifier defines configuration for external classification service
type ExternalClassifier struct {
	Enabled    bool              `json:"enabled"`
	Endpoint   string            `json:"endpoint"`
	Timeout    time.Duration     `json:"timeout"`
	Headers    map[string]string `json:"headers,omitempty"`
	RetryCount int               `json:"retry_count"`
	CacheTTL   time.Duration     `json:"cache_ttl"`
}

// ClassificationRequest sent to external classifier
type ClassificationRequest struct {
	JobID      string                 `json:"job_id"`
	Error      string                 `json:"error"`
	ErrorType  string                 `json:"error_type"`
	Payload    json.RawMessage        `json:"payload"`
	Queue      string                 `json:"queue"`
	JobType    string                 `json:"job_type"`
	RetryCount int                    `json:"retry_count"`
	Metadata   map[string]interface{} `json:"metadata"`
	FailedAt   time.Time              `json:"failed_at"`
}

// ClassificationResponse from external classifier
type ClassificationResponse struct {
	Category    string                 `json:"category"`
	Confidence  float64                `json:"confidence"`
	Actions     []string               `json:"suggested_actions"`
	Reason      string                 `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CacheFor    time.Duration          `json:"cache_for,omitempty"`
}

// PipelineConfig contains configuration for the remediation pipeline
type PipelineConfig struct {
	Enabled             bool                `json:"enabled"`
	PollInterval        time.Duration       `json:"poll_interval"`
	BatchSize           int                 `json:"batch_size"`
	MaxConcurrentRules  int                 `json:"max_concurrent_rules"`
	DryRun              bool                `json:"dry_run"`
	RedisStreamKey      string              `json:"redis_stream_key"`
	MetricsEnabled      bool                `json:"metrics_enabled"`
	AuditEnabled        bool                `json:"audit_enabled"`
	ExternalClassifier  ExternalClassifier  `json:"external_classifier"`
	GlobalSafetyLimits  SafetyLimits        `json:"global_safety_limits"`
	RetentionPolicy     RetentionPolicy     `json:"retention_policy"`
}

// RetentionPolicy defines how long to keep various types of data
type RetentionPolicy struct {
	AuditLogTTL        time.Duration `json:"audit_log_ttl"`
	MetricsTTL         time.Duration `json:"metrics_ttl"`
	ClassificationTTL  time.Duration `json:"classification_ttl"`
	ProcessedJobsTTL   time.Duration `json:"processed_jobs_ttl"`
}

// PipelineState represents the current state of the pipeline
type PipelineState struct {
	Status           PipelineStatus `json:"status"`
	StartedAt        time.Time      `json:"started_at"`
	LastRunAt        time.Time      `json:"last_run_at"`
	NextRunAt        time.Time      `json:"next_run_at"`
	TotalProcessed   int64          `json:"total_processed"`
	TotalSuccessful  int64          `json:"total_successful"`
	TotalFailed      int64          `json:"total_failed"`
	RulesEnabled     int            `json:"rules_enabled"`
	RulesDisabled    int            `json:"rules_disabled"`
	CurrentBatchSize int            `json:"current_batch_size"`
	LastError        string         `json:"last_error,omitempty"`
	LastErrorAt      time.Time      `json:"last_error_at,omitempty"`
}

// PipelineStatus represents the current status of the pipeline
type PipelineStatus string

const (
	StatusStopped PipelineStatus = "stopped"
	StatusRunning PipelineStatus = "running"
	StatusPaused  PipelineStatus = "paused"
	StatusError   PipelineStatus = "error"
)

// AuditLogEntry records actions taken by the pipeline
type AuditLogEntry struct {
	ID          string                 `json:"id" redis:"id"`
	Timestamp   time.Time              `json:"timestamp" redis:"timestamp"`
	JobID       string                 `json:"job_id" redis:"job_id"`
	RuleID      string                 `json:"rule_id" redis:"rule_id"`
	RuleName    string                 `json:"rule_name" redis:"rule_name"`
	Action      ActionType             `json:"action" redis:"action"`
	Parameters  map[string]interface{} `json:"parameters" redis:"parameters"`
	Result      string                 `json:"result" redis:"result"`
	Error       string                 `json:"error,omitempty" redis:"error"`
	DryRun      bool                   `json:"dry_run" redis:"dry_run"`
	UserID      string                 `json:"user_id,omitempty" redis:"user_id"`
	Duration    time.Duration          `json:"duration" redis:"duration"`
	BeforeState json.RawMessage        `json:"before_state,omitempty" redis:"before_state"`
	AfterState  json.RawMessage        `json:"after_state,omitempty" redis:"after_state"`
}

// PipelineMetrics contains various metrics about pipeline performance
type PipelineMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	JobsProcessed       int64     `json:"jobs_processed"`
	JobsMatched         int64     `json:"jobs_matched"`
	ActionsExecuted     int64     `json:"actions_executed"`
	ActionsSuccessful   int64     `json:"actions_successful"`
	ActionsFailed       int64     `json:"actions_failed"`
	ClassificationTime  float64   `json:"classification_time_ms"`
	ActionTime          float64   `json:"action_time_ms"`
	EndToEndTime        float64   `json:"end_to_end_time_ms"`
	RateLimitHits       int64     `json:"rate_limit_hits"`
	CircuitBreakerTrips int64     `json:"circuit_breaker_trips"`
	CacheHitRate        float64   `json:"cache_hit_rate"`
}

// RateLimiter implements rate limiting for pipeline operations
type RateLimiter struct {
	MaxPerMinute    int       `json:"max_per_minute"`
	MaxTotal        int       `json:"max_total"`
	BurstSize       int       `json:"burst_size"`
	CurrentMinute   time.Time `json:"current_minute"`
	CountThisMinute int       `json:"count_this_minute"`
	TotalProcessed  int       `json:"total_processed"`
}

// CanProcess checks if the rate limiter allows processing
func (rl *RateLimiter) CanProcess() bool {
	now := time.Now()
	currentMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())

	// Reset counter if we're in a new minute
	if !currentMinute.Equal(rl.CurrentMinute) {
		rl.CurrentMinute = currentMinute
		rl.CountThisMinute = 0
	}

	// Check limits
	if rl.MaxPerMinute > 0 && rl.CountThisMinute >= rl.MaxPerMinute {
		return false
	}

	if rl.MaxTotal > 0 && rl.TotalProcessed >= rl.MaxTotal {
		return false
	}

	return true
}

// RecordProcessed increments the counters
func (rl *RateLimiter) RecordProcessed() {
	rl.CountThisMinute++
	rl.TotalProcessed++
}

// CircuitBreaker implements circuit breaker pattern for safety
type CircuitBreaker struct {
	ErrorThreshold   float64       `json:"error_threshold"`
	MinRequests      int           `json:"min_requests"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	State            CircuitState  `json:"state"`
	ErrorCount       int           `json:"error_count"`
	RequestCount     int           `json:"request_count"`
	LastFailure      time.Time     `json:"last_failure"`
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	now := time.Now()

	switch cb.State {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if now.Sub(cb.LastFailure) > cb.RecoveryTimeout {
			cb.State = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.RequestCount++
	if cb.State == CircuitHalfOpen {
		cb.State = CircuitClosed
		cb.ErrorCount = 0
		cb.RequestCount = 0
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.RequestCount++
	cb.ErrorCount++
	cb.LastFailure = time.Now()

	if cb.RequestCount >= cb.MinRequests {
		errorRate := float64(cb.ErrorCount) / float64(cb.RequestCount)
		if errorRate >= cb.ErrorThreshold {
			cb.State = CircuitOpen
		}
	}
}

// IdempotencyTracker prevents duplicate processing of jobs
type IdempotencyTracker struct {
	ProcessedJobs map[string]time.Time `json:"processed_jobs"`
	TTL           time.Duration        `json:"ttl"`
}

// NewIdempotencyTracker creates a new idempotency tracker
func NewIdempotencyTracker(ttl time.Duration) *IdempotencyTracker {
	return &IdempotencyTracker{
		ProcessedJobs: make(map[string]time.Time),
		TTL:           ttl,
	}
}

// IsProcessed checks if a job has already been processed
func (it *IdempotencyTracker) IsProcessed(jobID string) bool {
	timestamp, exists := it.ProcessedJobs[jobID]
	if !exists {
		return false
	}

	// Check if entry has expired
	if time.Since(timestamp) > it.TTL {
		delete(it.ProcessedJobs, jobID)
		return false
	}

	return true
}

// MarkProcessed marks a job as processed
func (it *IdempotencyTracker) MarkProcessed(jobID string) {
	it.ProcessedJobs[jobID] = time.Now()
}

// Cleanup removes expired entries
func (it *IdempotencyTracker) Cleanup() {
	now := time.Now()
	for jobID, timestamp := range it.ProcessedJobs {
		if now.Sub(timestamp) > it.TTL {
			delete(it.ProcessedJobs, jobID)
		}
	}
}

// ProcessingResult represents the result of processing a job
type ProcessingResult struct {
	JobID       string        `json:"job_id"`
	RuleID      string        `json:"rule_id"`
	Success     bool          `json:"success"`
	Actions     []ActionType  `json:"actions"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	DryRun      bool          `json:"dry_run"`
	BeforeState *DLQJob       `json:"before_state,omitempty"`
	AfterState  *DLQJob       `json:"after_state,omitempty"`
}

// BatchResult represents the result of processing a batch of jobs
type BatchResult struct {
	StartedAt    time.Time          `json:"started_at"`
	CompletedAt  time.Time          `json:"completed_at"`
	TotalJobs    int                `json:"total_jobs"`
	ProcessedJobs int               `json:"processed_jobs"`
	SuccessfulJobs int              `json:"successful_jobs"`
	FailedJobs   int                `json:"failed_jobs"`
	SkippedJobs  int                `json:"skipped_jobs"`
	Results      []ProcessingResult `json:"results"`
	Errors       []string           `json:"errors,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", ve.Field, ve.Message)
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (mve MultiValidationError) Error() string {
	if len(mve.Errors) == 1 {
		return mve.Errors[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(mve.Errors))
}

// HasErrors returns true if there are validation errors
func (mve MultiValidationError) HasErrors() bool {
	return len(mve.Errors) > 0
}