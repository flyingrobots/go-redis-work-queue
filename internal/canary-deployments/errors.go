package canary_deployments

import (
	"errors"
	"fmt"
)

// Predefined error variables for common error conditions
var (
	// Deployment errors
	ErrDeploymentNotFound     = errors.New("deployment not found")
	ErrDeploymentExists       = errors.New("deployment already exists")
	ErrDeploymentNotActive    = errors.New("deployment is not active")
	ErrDeploymentInProgress   = errors.New("deployment operation already in progress")
	ErrDeploymentCompleted    = errors.New("deployment already completed")
	ErrDeploymentFailed       = errors.New("deployment has failed")

	// Configuration errors
	ErrInvalidConfiguration   = errors.New("invalid configuration")
	ErrInvalidPercentage      = errors.New("invalid percentage value")
	ErrInvalidDuration        = errors.New("invalid duration value")
	ErrInvalidThreshold       = errors.New("invalid threshold value")
	ErrInvalidRoutingStrategy = errors.New("invalid routing strategy")

	// Worker errors
	ErrWorkerNotFound         = errors.New("worker not found")
	ErrWorkerUnhealthy        = errors.New("worker is unhealthy")
	ErrNoHealthyWorkers       = errors.New("no healthy workers available")
	ErrWorkerVersionMismatch  = errors.New("worker version mismatch")

	// Metrics errors
	ErrInsufficientMetrics    = errors.New("insufficient metrics data")
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	ErrInvalidMetricsWindow   = errors.New("invalid metrics window")

	// Queue/Routing errors
	ErrQueueNotFound          = errors.New("queue not found")
	ErrRoutingFailed          = errors.New("job routing failed")
	ErrDrainTimeout           = errors.New("drain operation timed out")
	ErrQueueEmpty             = errors.New("queue is empty")

	// System errors
	ErrSystemNotReady         = errors.New("system is not ready")
	ErrRedisConnectionFailed  = errors.New("Redis connection failed")
	ErrOperationTimeout       = errors.New("operation timed out")
	ErrConcurrencyLimit       = errors.New("concurrency limit exceeded")

	// Validation errors
	ErrValidationFailed       = errors.New("validation failed")
	ErrMissingRequiredField   = errors.New("missing required field")
	ErrInvalidFieldValue      = errors.New("invalid field value")
)

// CanaryError represents a structured error from the canary deployment system
type CanaryError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Underlying error             `json:"-"`
}

// Error implements the error interface
func (ce *CanaryError) Error() string {
	if ce.Underlying != nil {
		return fmt.Sprintf("%s: %s (%v)", ce.Code, ce.Message, ce.Underlying)
	}
	return fmt.Sprintf("%s: %s", ce.Code, ce.Message)
}

// Unwrap returns the underlying error for error chain inspection
func (ce *CanaryError) Unwrap() error {
	return ce.Underlying
}

// Is checks if the error matches a target error
func (ce *CanaryError) Is(target error) bool {
	if ce.Underlying != nil {
		return errors.Is(ce.Underlying, target)
	}
	return ce.Error() == target.Error()
}

// NewCanaryError creates a new CanaryError with the given code and message
func NewCanaryError(code, message string) *CanaryError {
	return &CanaryError{
		Code:    code,
		Message: message,
		Details: make(map[string]string),
	}
}

// NewCanaryErrorWithDetails creates a new CanaryError with details
func NewCanaryErrorWithDetails(code, message string, details map[string]string) *CanaryError {
	return &CanaryError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError wraps an existing error with canary context
func WrapError(err error, code, message string) *CanaryError {
	return &CanaryError{
		Code:       code,
		Message:    message,
		Details:    make(map[string]string),
		Underlying: err,
	}
}

// WithDetail adds a detail to the error
func (ce *CanaryError) WithDetail(key, value string) *CanaryError {
	if ce.Details == nil {
		ce.Details = make(map[string]string)
	}
	ce.Details[key] = value
	return ce
}

// Error code constants for structured error handling
const (
	// Deployment error codes
	CodeDeploymentNotFound     = "DEPLOYMENT_NOT_FOUND"
	CodeDeploymentExists       = "DEPLOYMENT_EXISTS"
	CodeDeploymentNotActive    = "DEPLOYMENT_NOT_ACTIVE"
	CodeDeploymentInProgress   = "DEPLOYMENT_IN_PROGRESS"
	CodeDeploymentCompleted    = "DEPLOYMENT_COMPLETED"
	CodeDeploymentFailed       = "DEPLOYMENT_FAILED"

	// Configuration error codes
	CodeInvalidConfiguration   = "INVALID_CONFIGURATION"
	CodeInvalidPercentage      = "INVALID_PERCENTAGE"
	CodeInvalidDuration        = "INVALID_DURATION"
	CodeInvalidThreshold       = "INVALID_THRESHOLD"
	CodeInvalidRoutingStrategy = "INVALID_ROUTING_STRATEGY"

	// Worker error codes
	CodeWorkerNotFound         = "WORKER_NOT_FOUND"
	CodeWorkerUnhealthy        = "WORKER_UNHEALTHY"
	CodeNoHealthyWorkers       = "NO_HEALTHY_WORKERS"
	CodeWorkerVersionMismatch  = "WORKER_VERSION_MISMATCH"

	// Metrics error codes
	CodeInsufficientMetrics    = "INSUFFICIENT_METRICS"
	CodeMetricsCollectionFailed = "METRICS_COLLECTION_FAILED"
	CodeInvalidMetricsWindow   = "INVALID_METRICS_WINDOW"

	// Queue/Routing error codes
	CodeQueueNotFound          = "QUEUE_NOT_FOUND"
	CodeRoutingFailed          = "ROUTING_FAILED"
	CodeDrainTimeout           = "DRAIN_TIMEOUT"
	CodeQueueEmpty             = "QUEUE_EMPTY"

	// System error codes
	CodeSystemNotReady         = "SYSTEM_NOT_READY"
	CodeRedisConnectionFailed  = "REDIS_CONNECTION_FAILED"
	CodeOperationTimeout       = "OPERATION_TIMEOUT"
	CodeConcurrencyLimit       = "CONCURRENCY_LIMIT"

	// Validation error codes
	CodeValidationFailed       = "VALIDATION_FAILED"
	CodeMissingRequiredField   = "MISSING_REQUIRED_FIELD"
	CodeInvalidFieldValue      = "INVALID_FIELD_VALUE"

	// Health check error codes
	CodeHealthCheckFailed      = "HEALTH_CHECK_FAILED"
	CodeSLOViolation           = "SLO_VIOLATION"
	CodePromotionBlocked       = "PROMOTION_BLOCKED"
	CodeRollbackTriggered      = "ROLLBACK_TRIGGERED"

	// Alert error codes
	CodeAlertFailed            = "ALERT_FAILED"
	CodeWebhookFailed          = "WEBHOOK_FAILED"
	CodeNotificationFailed     = "NOTIFICATION_FAILED"
)

// Convenience functions for creating common errors

// NewDeploymentNotFoundError creates an error for missing deployments
func NewDeploymentNotFoundError(deploymentID string) *CanaryError {
	return NewCanaryError(CodeDeploymentNotFound, "deployment not found").
		WithDetail("deployment_id", deploymentID)
}

// NewDeploymentExistsError creates an error for duplicate deployments
func NewDeploymentExistsError(deploymentID string) *CanaryError {
	return NewCanaryError(CodeDeploymentExists, "deployment already exists").
		WithDetail("deployment_id", deploymentID)
}

// NewWorkerNotFoundError creates an error for missing workers
func NewWorkerNotFoundError(workerID string) *CanaryError {
	return NewCanaryError(CodeWorkerNotFound, "worker not found").
		WithDetail("worker_id", workerID)
}

// NewQueueNotFoundError creates an error for missing queues
func NewQueueNotFoundError(queueName string) *CanaryError {
	return NewCanaryError(CodeQueueNotFound, "queue not found").
		WithDetail("queue_name", queueName)
}

// NewInvalidPercentageError creates an error for invalid percentage values
func NewInvalidPercentageError(percentage int) *CanaryError {
	return NewCanaryError(CodeInvalidPercentage, "percentage must be between 0 and 100").
		WithDetail("percentage", fmt.Sprintf("%d", percentage))
}

// NewInsufficientMetricsError creates an error for insufficient metrics data
func NewInsufficientMetricsError(required, actual int) *CanaryError {
	return NewCanaryError(CodeInsufficientMetrics, "insufficient metrics data").
		WithDetail("required_samples", fmt.Sprintf("%d", required)).
		WithDetail("actual_samples", fmt.Sprintf("%d", actual))
}

// NewSLOViolationError creates an error for SLO threshold violations
func NewSLOViolationError(metric string, threshold, actual float64) *CanaryError {
	return NewCanaryError(CodeSLOViolation, fmt.Sprintf("%s SLO violated", metric)).
		WithDetail("metric", metric).
		WithDetail("threshold", fmt.Sprintf("%.2f", threshold)).
		WithDetail("actual", fmt.Sprintf("%.2f", actual))
}

// NewPromotionBlockedError creates an error when promotion is blocked
func NewPromotionBlockedError(reason string) *CanaryError {
	return NewCanaryError(CodePromotionBlocked, "promotion blocked").
		WithDetail("reason", reason)
}

// NewRollbackTriggeredError creates an error when rollback is triggered
func NewRollbackTriggeredError(reason string) *CanaryError {
	return NewCanaryError(CodeRollbackTriggered, "rollback triggered").
		WithDetail("reason", reason)
}

// NewValidationError creates a validation error with field details
func NewValidationError(field, reason string) *CanaryError {
	return NewCanaryError(CodeValidationFailed, "validation failed").
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// NewConcurrencyLimitError creates an error for concurrency limit violations
func NewConcurrencyLimitError(limit int) *CanaryError {
	return NewCanaryError(CodeConcurrencyLimit, "too many concurrent deployments").
		WithDetail("limit", fmt.Sprintf("%d", limit))
}

// NewOperationTimeoutError creates an error for operation timeouts
func NewOperationTimeoutError(operation string, timeout string) *CanaryError {
	return NewCanaryError(CodeOperationTimeout, fmt.Sprintf("%s operation timed out", operation)).
		WithDetail("operation", operation).
		WithDetail("timeout", timeout)
}

// IsCanaryError checks if an error is a CanaryError
func IsCanaryError(err error) bool {
	var canaryErr *CanaryError
	return errors.As(err, &canaryErr)
}

// GetCanaryError extracts a CanaryError from an error chain
func GetCanaryError(err error) *CanaryError {
	var canaryErr *CanaryError
	if errors.As(err, &canaryErr) {
		return canaryErr
	}
	return nil
}

// IsCode checks if an error has a specific canary error code
func IsCode(err error, code string) bool {
	if canaryErr := GetCanaryError(err); canaryErr != nil {
		return canaryErr.Code == code
	}
	return false
}

// IsTemporary checks if an error is considered temporary and retryable
func IsTemporary(err error) bool {
	canaryErr := GetCanaryError(err)
	if canaryErr == nil {
		return false
	}

	// These error codes indicate temporary conditions
	temporaryCodes := map[string]bool{
		CodeRedisConnectionFailed:  true,
		CodeOperationTimeout:       true,
		CodeMetricsCollectionFailed: true,
		CodeSystemNotReady:         true,
		CodeQueueEmpty:             true,
		CodeWorkerUnhealthy:        true,
	}

	return temporaryCodes[canaryErr.Code]
}

// IsPermanent checks if an error is considered permanent and non-retryable
func IsPermanent(err error) bool {
	canaryErr := GetCanaryError(err)
	if canaryErr == nil {
		return false
	}

	// These error codes indicate permanent conditions
	permanentCodes := map[string]bool{
		CodeDeploymentNotFound:     true,
		CodeWorkerNotFound:         true,
		CodeQueueNotFound:          true,
		CodeInvalidConfiguration:   true,
		CodeInvalidPercentage:      true,
		CodeInvalidDuration:        true,
		CodeInvalidThreshold:       true,
		CodeInvalidRoutingStrategy: true,
		CodeValidationFailed:       true,
		CodeMissingRequiredField:   true,
		CodeInvalidFieldValue:      true,
		CodeDeploymentCompleted:    true,
	}

	return permanentCodes[canaryErr.Code]
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
	TraceID string            `json:"trace_id,omitempty"`
}

// ToErrorResponse converts a CanaryError to an API error response
func (ce *CanaryError) ToErrorResponse(traceID string) *ErrorResponse {
	return &ErrorResponse{
		Error:   ce.Message,
		Code:    ce.Code,
		Details: ce.Details,
		TraceID: traceID,
	}
}

// NewErrorResponse creates an ErrorResponse from any error
func NewErrorResponse(err error, traceID string) *ErrorResponse {
	if canaryErr := GetCanaryError(err); canaryErr != nil {
		return canaryErr.ToErrorResponse(traceID)
	}

	return &ErrorResponse{
		Error:   err.Error(),
		Code:    "INTERNAL_ERROR",
		TraceID: traceID,
	}
}