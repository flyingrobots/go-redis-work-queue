// Copyright 2025 James Ross
package smartretry

import (
	"errors"
	"fmt"
)

// Custom errors for smart retry strategies
var (
	// Configuration errors
	ErrInvalidConfig        = errors.New("invalid configuration")
	ErrMissingRedisAddr     = errors.New("Redis address is required")
	ErrInvalidThreshold     = errors.New("invalid Bayesian threshold")
	ErrInvalidSampleRate    = errors.New("invalid sample rate")

	// Policy errors
	ErrPolicyNotFound       = errors.New("policy not found")
	ErrDuplicatePolicy      = errors.New("policy already exists")
	ErrInvalidPolicy        = errors.New("invalid policy configuration")
	ErrNoPoliciesAvailable  = errors.New("no policies available")

	// Model errors
	ErrModelNotFound        = errors.New("model not found")
	ErrModelNotTrained      = errors.New("model not trained")
	ErrInsufficientData     = errors.New("insufficient training data")
	ErrModelDeployFailed    = errors.New("model deployment failed")
	ErrMLNotEnabled         = errors.New("ML is not enabled")

	// Recommendation errors
	ErrNoRecommendation     = errors.New("no recommendation available")
	ErrMaxAttemptsReached   = errors.New("maximum attempts reached")
	ErrInvalidFeatures      = errors.New("invalid features provided")

	// Data errors
	ErrDataNotFound         = errors.New("data not found")
	ErrStorageFailed        = errors.New("storage operation failed")
	ErrCacheExpired         = errors.New("cache entry expired")

	// Guardrail errors
	ErrGuardrailViolation   = errors.New("guardrail violation")
	ErrBudgetExceeded       = errors.New("retry budget exceeded")
	ErrEmergencyStop        = errors.New("emergency stop activated")
)

// RetryError wraps an error with additional context
type RetryError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// Error implements the error interface
func (e *RetryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *RetryError) Unwrap() error {
	return e.Cause
}

// NewRetryError creates a new retry error
func NewRetryError(code, message string, cause error) *RetryError {
	return &RetryError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
		Timestamp: timeNow().Unix(),
	}
}

// WithContext adds context to the error
func (e *RetryError) WithContext(key string, value interface{}) *RetryError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Configuration error constructors
func NewConfigError(message string, cause error) *RetryError {
	return NewRetryError("CONFIG_ERROR", message, cause)
}

func NewPolicyError(message string, cause error) *RetryError {
	return NewRetryError("POLICY_ERROR", message, cause)
}

func NewModelError(message string, cause error) *RetryError {
	return NewRetryError("MODEL_ERROR", message, cause)
}

func NewRecommendationError(message string, cause error) *RetryError {
	return NewRetryError("RECOMMENDATION_ERROR", message, cause)
}

func NewDataError(message string, cause error) *RetryError {
	return NewRetryError("DATA_ERROR", message, cause)
}

func NewGuardrailError(message string, cause error) *RetryError {
	return NewRetryError("GUARDRAIL_ERROR", message, cause)
}

// Specific error constructors

func ErrPolicyAlreadyExists(name string) *RetryError {
	return NewPolicyError(fmt.Sprintf("Policy '%s' already exists", name), ErrDuplicatePolicy).
		WithContext("policy_name", name)
}

func ErrPolicyNameRequired() *RetryError {
	return NewPolicyError("Policy name is required", ErrInvalidPolicy)
}

func ErrInvalidPolicyPriority(priority int) *RetryError {
	return NewPolicyError(fmt.Sprintf("Invalid policy priority: %d", priority), ErrInvalidPolicy).
		WithContext("priority", priority)
}

func ErrModelVersionMismatch(expected, actual string) *RetryError {
	return NewModelError(fmt.Sprintf("Model version mismatch: expected %s, got %s", expected, actual), ErrModelNotFound).
		WithContext("expected_version", expected).
		WithContext("actual_version", actual)
}

func ErrInsufficientTrainingData(required, actual int) *RetryError {
	return NewModelError(fmt.Sprintf("Insufficient training data: need %d, got %d", required, actual), ErrInsufficientData).
		WithContext("required_samples", required).
		WithContext("actual_samples", actual)
}

func ErrFeatureExtractionFailed(feature string, cause error) *RetryError {
	return NewRecommendationError(fmt.Sprintf("Failed to extract feature '%s'", feature), cause).
		WithContext("feature", feature)
}

func ErrMaxAttemptsExceeded(attempts, maxAttempts int) *RetryError {
	return NewGuardrailError(fmt.Sprintf("Max attempts exceeded: %d/%d", attempts, maxAttempts), ErrMaxAttemptsReached).
		WithContext("attempts", attempts).
		WithContext("max_attempts", maxAttempts)
}

func ErrMaxDelayExceeded(delay, maxDelay int64) *RetryError {
	return NewGuardrailError(fmt.Sprintf("Max delay exceeded: %dms > %dms", delay, maxDelay), ErrGuardrailViolation).
		WithContext("delay_ms", delay).
		WithContext("max_delay_ms", maxDelay)
}

func ErrBudgetLimitExceeded(used, limit float64) *RetryError {
	return NewGuardrailError(fmt.Sprintf("Budget limit exceeded: %.1f%% > %.1f%%", used, limit), ErrBudgetExceeded).
		WithContext("used_percent", used).
		WithContext("limit_percent", limit)
}

// Error classification functions

// IsConfigError checks if the error is a configuration error
func IsConfigError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "CONFIG_ERROR"
	}
	return errors.Is(err, ErrInvalidConfig) ||
		   errors.Is(err, ErrMissingRedisAddr) ||
		   errors.Is(err, ErrInvalidThreshold) ||
		   errors.Is(err, ErrInvalidSampleRate)
}

// IsPolicyError checks if the error is a policy error
func IsPolicyError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "POLICY_ERROR"
	}
	return errors.Is(err, ErrPolicyNotFound) ||
		   errors.Is(err, ErrDuplicatePolicy) ||
		   errors.Is(err, ErrInvalidPolicy) ||
		   errors.Is(err, ErrNoPoliciesAvailable)
}

// IsModelError checks if the error is a model error
func IsModelError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "MODEL_ERROR"
	}
	return errors.Is(err, ErrModelNotFound) ||
		   errors.Is(err, ErrModelNotTrained) ||
		   errors.Is(err, ErrInsufficientData) ||
		   errors.Is(err, ErrModelDeployFailed) ||
		   errors.Is(err, ErrMLNotEnabled)
}

// IsRecommendationError checks if the error is a recommendation error
func IsRecommendationError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "RECOMMENDATION_ERROR"
	}
	return errors.Is(err, ErrNoRecommendation) ||
		   errors.Is(err, ErrInvalidFeatures)
}

// IsDataError checks if the error is a data error
func IsDataError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "DATA_ERROR"
	}
	return errors.Is(err, ErrDataNotFound) ||
		   errors.Is(err, ErrStorageFailed) ||
		   errors.Is(err, ErrCacheExpired)
}

// IsGuardrailError checks if the error is a guardrail error
func IsGuardrailError(err error) bool {
	var retryErr *RetryError
	if errors.As(err, &retryErr) {
		return retryErr.Code == "GUARDRAIL_ERROR"
	}
	return errors.Is(err, ErrGuardrailViolation) ||
		   errors.Is(err, ErrBudgetExceeded) ||
		   errors.Is(err, ErrEmergencyStop) ||
		   errors.Is(err, ErrMaxAttemptsReached)
}

// IsRetriable checks if an error indicates the operation should be retried
func IsRetriable(err error) bool {
	// Data errors and some model errors are retriable
	return IsDataError(err) || errors.Is(err, ErrCacheExpired)
}

// GetErrorSeverity returns the severity level of an error
func GetErrorSeverity(err error) string {
	if IsConfigError(err) {
		return "critical"
	}
	if IsGuardrailError(err) {
		return "warning"
	}
	if IsModelError(err) {
		return "error"
	}
	if IsRecommendationError(err) || IsPolicyError(err) {
		return "error"
	}
	if IsDataError(err) {
		return "warning"
	}
	return "info"
}

// Mock time function for testing
var timeNow = func() mockTime {
	return mockTime{}
}

type mockTime struct{}

func (t mockTime) Unix() int64 {
	return 1642680000 // Mock timestamp
}