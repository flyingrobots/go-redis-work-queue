// Copyright 2025 James Ross
package smartretry

import (
	"time"
)

// AttemptHistory represents the historical data for a job attempt
type AttemptHistory struct {
	JobID           string            `json:"job_id"`
	JobType         string            `json:"job_type"`
	AttemptNumber   int               `json:"attempt_number"`
	ErrorClass      string            `json:"error_class,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	Status          string            `json:"status"`
	Queue           string            `json:"queue"`
	Tenant          string            `json:"tenant,omitempty"`
	PayloadSize     int64             `json:"payload_size"`
	TimeOfDay       int               `json:"time_of_day"` // Hour of day 0-23
	WorkerVersion   string            `json:"worker_version"`
	Health          map[string]float64 `json:"health,omitempty"` // Downstream health signals
	DelayMs         int64             `json:"delay_ms"`
	Success         bool              `json:"success"`
	Timestamp       time.Time         `json:"timestamp"`
	ProcessingTime  time.Duration     `json:"processing_time"`
}

// RetryFeatures represents the features extracted for retry decision making
type RetryFeatures struct {
	JobType         string            `json:"job_type"`
	ErrorClass      string            `json:"error_class"`
	ErrorCode       string            `json:"error_code"`
	AttemptNumber   int               `json:"attempt_number"`
	Queue           string            `json:"queue"`
	Tenant          string            `json:"tenant,omitempty"`
	PayloadSize     int64             `json:"payload_size"`
	TimeOfDay       int               `json:"time_of_day"`
	WorkerVersion   string            `json:"worker_version"`
	Health          map[string]float64 `json:"health,omitempty"`
	SinceLastFailure time.Duration    `json:"since_last_failure"`
	RecentFailures  int               `json:"recent_failures"`
	AvgProcessingTime time.Duration   `json:"avg_processing_time"`
}

// RetryRecommendation represents a recommendation for retry timing and policy
type RetryRecommendation struct {
	ShouldRetry       bool          `json:"should_retry"`
	DelayMs           int64         `json:"delay_ms"`
	MaxAttempts       int           `json:"max_attempts"`
	Confidence        float64       `json:"confidence"`
	Rationale         string        `json:"rationale"`
	Method            string        `json:"method"` // "rules", "bayesian", "ml"
	EstimatedSuccess  float64       `json:"estimated_success"`
	NextEvaluation    time.Time     `json:"next_evaluation"`
	PolicyGuardrails  []string      `json:"policy_guardrails,omitempty"`
}

// RetryPolicy represents a retry policy configuration
type RetryPolicy struct {
	Name              string        `json:"name"`
	ErrorPatterns     []string      `json:"error_patterns"`
	JobTypePatterns   []string      `json:"job_type_patterns,omitempty"`
	MaxAttempts       int           `json:"max_attempts"`
	BaseDelayMs       int64         `json:"base_delay_ms"`
	MaxDelayMs        int64         `json:"max_delay_ms"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	JitterPercent     float64       `json:"jitter_percent"`
	StopOnValidation  bool          `json:"stop_on_validation"`
	Priority          int           `json:"priority"` // Higher priority = evaluated first
}

// BayesianModel represents the Bayesian success probability model
type BayesianModel struct {
	JobType        string                    `json:"job_type"`
	ErrorClass     string                    `json:"error_class"`
	Buckets        []BayesianBucket          `json:"buckets"`
	LastUpdated    time.Time                 `json:"last_updated"`
	SampleCount    int                       `json:"sample_count"`
	Confidence     float64                   `json:"confidence"`
	Metadata       map[string]interface{}    `json:"metadata,omitempty"`
}

// BayesianBucket represents a delay range bucket with success statistics
type BayesianBucket struct {
	DelayMinMs    int64   `json:"delay_min_ms"`
	DelayMaxMs    int64   `json:"delay_max_ms"`
	Successes     int     `json:"successes"`     // Alpha parameter
	Failures      int     `json:"failures"`      // Beta parameter
	Probability   float64 `json:"probability"`   // Beta distribution mean
	UpperBound    float64 `json:"upper_bound"`   // 95% confidence upper bound
	LowerBound    float64 `json:"lower_bound"`   // 95% confidence lower bound
}

// MLModel represents an optional machine learning model
type MLModel struct {
	Version         string                 `json:"version"`
	ModelType       string                 `json:"model_type"` // "logistic", "gradient_boost", etc.
	Features        []string               `json:"features"`
	ModelData       []byte                 `json:"model_data"`
	TrainedAt       time.Time              `json:"trained_at"`
	Accuracy        float64                `json:"accuracy"`
	F1Score         float64                `json:"f1_score"`
	ValidationSet   string                 `json:"validation_set"`
	Enabled         bool                   `json:"enabled"`
	CanaryPercent   float64                `json:"canary_percent"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RetryStrategy represents the main retry strategy configuration
type RetryStrategy struct {
	Name              string          `json:"name"`
	Enabled           bool            `json:"enabled"`
	Policies          []RetryPolicy   `json:"policies"`
	BayesianThreshold float64         `json:"bayesian_threshold"`
	MLEnabled         bool            `json:"ml_enabled"`
	MLModel           *MLModel        `json:"ml_model,omitempty"`
	Guardrails        PolicyGuardrails `json:"guardrails"`
	DataCollection    DataCollectionConfig `json:"data_collection"`
}

// PolicyGuardrails represents hard limits and safety constraints
type PolicyGuardrails struct {
	MaxAttempts       int           `json:"max_attempts"`
	MaxDelayMs        int64         `json:"max_delay_ms"`
	MaxBudgetPercent  float64       `json:"max_budget_percent"`
	PerTenantLimits   bool          `json:"per_tenant_limits"`
	EmergencyStop     bool          `json:"emergency_stop"`
	ExplainabilityReq bool          `json:"explainability_required"`
}

// DataCollectionConfig represents configuration for data collection
type DataCollectionConfig struct {
	Enabled             bool          `json:"enabled"`
	SampleRate          float64       `json:"sample_rate"`
	RetentionDays       int           `json:"retention_days"`
	AggregationInterval time.Duration `json:"aggregation_interval"`
	FeatureExtraction   bool          `json:"feature_extraction"`
}

// RetryStats represents aggregated retry statistics
type RetryStats struct {
	JobType           string            `json:"job_type"`
	ErrorClass        string            `json:"error_class"`
	TotalAttempts     int64             `json:"total_attempts"`
	SuccessfulRetries int64             `json:"successful_retries"`
	FailedRetries     int64             `json:"failed_retries"`
	AvgDelayMs        float64           `json:"avg_delay_ms"`
	SuccessRate       float64           `json:"success_rate"`
	LastUpdated       time.Time         `json:"last_updated"`
	WindowStart       time.Time         `json:"window_start"`
	WindowEnd         time.Time         `json:"window_end"`
}

// RetryPreview represents a preview of recommended retry schedule
type RetryPreview struct {
	JobID         string                `json:"job_id"`
	CurrentAttempt int                  `json:"current_attempt"`
	Features      RetryFeatures         `json:"features"`
	Recommendations []RetryRecommendation `json:"recommendations"`
	Timeline      []RetryTimelineEntry  `json:"timeline"`
	GeneratedAt   time.Time             `json:"generated_at"`
}

// RetryTimelineEntry represents a point in the retry timeline
type RetryTimelineEntry struct {
	AttemptNumber    int       `json:"attempt_number"`
	ScheduledTime    time.Time `json:"scheduled_time"`
	EstimatedSuccess float64   `json:"estimated_success"`
	DelayMs          int64     `json:"delay_ms"`
	Method           string    `json:"method"`
	Rationale        string    `json:"rationale"`
}

// Manager interface defines the smart retry strategies operations
type Manager interface {
	// Retry recommendations
	GetRecommendation(features RetryFeatures) (*RetryRecommendation, error)
	PreviewRetrySchedule(features RetryFeatures, maxAttempts int) (*RetryPreview, error)

	// Data collection
	RecordAttempt(attempt AttemptHistory) error
	GetStats(jobType, errorClass string, window time.Duration) (*RetryStats, error)

	// Model management
	UpdateBayesianModel(jobType, errorClass string) error
	TrainMLModel(config MLTrainingConfig) (*MLModel, error)
	DeployMLModel(model *MLModel, canaryPercent float64) error
	RollbackMLModel() error

	// Policy management
	AddPolicy(policy RetryPolicy) error
	RemovePolicy(name string) error
	UpdateGuardrails(guardrails PolicyGuardrails) error

	// Configuration
	GetStrategy() (*RetryStrategy, error)
	UpdateStrategy(strategy *RetryStrategy) error
}

// MLTrainingConfig represents configuration for ML model training
type MLTrainingConfig struct {
	ModelType      string            `json:"model_type"`
	Features       []string          `json:"features"`
	TrainingPeriod time.Duration     `json:"training_period"`
	ValidationSet  float64           `json:"validation_set"` // Percentage for validation
	CrossValidation int              `json:"cross_validation"` // K-fold CV
	Hyperparameters map[string]interface{} `json:"hyperparameters,omitempty"`
}

// RetryEvent represents an event in the retry system
type RetryEvent struct {
	ID        string                 `json:"id"`
	Type      RetryEventType         `json:"type"`
	JobID     string                 `json:"job_id"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// RetryEventType represents the type of retry event
type RetryEventType string

const (
	EventTypeRecommendationGenerated RetryEventType = "recommendation_generated"
	EventTypeAttemptRecorded         RetryEventType = "attempt_recorded"
	EventTypeBayesianUpdated         RetryEventType = "bayesian_updated"
	EventTypeMLModelTrained          RetryEventType = "ml_model_trained"
	EventTypeMLModelDeployed         RetryEventType = "ml_model_deployed"
	EventTypeGuardrailTriggered      RetryEventType = "guardrail_triggered"
	EventTypePolicyUpdated           RetryEventType = "policy_updated"
)

// Config represents the configuration for the smart retry strategies module
type Config struct {
	Enabled         bool                 `json:"enabled"`
	RedisAddr       string               `json:"redis_addr"`
	RedisPassword   string               `json:"redis_password,omitempty"`
	RedisDB         int                  `json:"redis_db"`
	Strategy        RetryStrategy        `json:"strategy"`
	DataCollection  DataCollectionConfig `json:"data_collection"`
	Cache           CacheConfig          `json:"cache"`
	API             APIConfig            `json:"api"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl"`
	MaxEntries int           `json:"max_entries"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Path    string `json:"path"`
}