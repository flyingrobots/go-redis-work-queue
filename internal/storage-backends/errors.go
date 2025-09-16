package storage

import (
	"errors"
	"fmt"
)

var (
	// ErrBackendNotFound is returned when a backend is not registered
	ErrBackendNotFound = errors.New("backend not found")

	// ErrQueueNotFound is returned when a queue configuration is not found
	ErrQueueNotFound = errors.New("queue not found")

	// ErrJobNotFound is returned when a job cannot be found
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyAcked is returned when trying to ack an already acknowledged job
	ErrJobAlreadyAcked = errors.New("job already acknowledged")

	// ErrJobProcessing is returned when a job is currently being processed
	ErrJobProcessing = errors.New("job is currently being processed")

	// ErrInvalidConfiguration is returned for invalid backend configurations
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrConnectionFailed is returned when backend connection fails
	ErrConnectionFailed = errors.New("connection failed")

	// ErrOperationNotSupported is returned when an operation is not supported by the backend
	ErrOperationNotSupported = errors.New("operation not supported")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")

	// ErrQueueEmpty is returned when trying to dequeue from an empty queue
	ErrQueueEmpty = errors.New("queue is empty")

	// ErrMigrationInProgress is returned when a migration is already in progress
	ErrMigrationInProgress = errors.New("migration already in progress")

	// ErrMigrationFailed is returned when a migration fails
	ErrMigrationFailed = errors.New("migration failed")

	// ErrConsumerGroupExists is returned when trying to create an existing consumer group
	ErrConsumerGroupExists = errors.New("consumer group already exists")

	// ErrStreamNotFound is returned when a stream doesn't exist
	ErrStreamNotFound = errors.New("stream not found")

	// ErrInvalidJobData is returned when job data cannot be parsed
	ErrInvalidJobData = errors.New("invalid job data")

	// ErrCircuitBreakerOpen is returned when circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limited")
)

// BackendError wraps backend-specific errors with additional context
type BackendError struct {
	Backend   string
	Operation string
	Err       error
}

func (e *BackendError) Error() string {
	return fmt.Sprintf("backend %s: operation %s failed: %v", e.Backend, e.Operation, e.Err)
}

func (e *BackendError) Unwrap() error {
	return e.Err
}

// NewBackendError creates a new backend error
func NewBackendError(backend, operation string, err error) *BackendError {
	return &BackendError{
		Backend:   backend,
		Operation: operation,
		Err:       err,
	}
}

// ConfigurationError represents configuration validation errors
type ConfigurationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in field %s (value: %v): %s", e.Field, e.Value, e.Message)
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(field string, value interface{}, message string) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// MigrationError represents migration-specific errors
type MigrationError struct {
	Phase   string
	Source  string
	Target  string
	JobID   string
	Message string
	Err     error
}

func (e *MigrationError) Error() string {
	msg := fmt.Sprintf("migration error in phase %s (source: %s, target: %s): %s",
		e.Phase, e.Source, e.Target, e.Message)
	if e.JobID != "" {
		msg += fmt.Sprintf(" (job: %s)", e.JobID)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *MigrationError) Unwrap() error {
	return e.Err
}

// NewMigrationError creates a new migration error
func NewMigrationError(phase, source, target, jobID, message string, err error) *MigrationError {
	return &MigrationError{
		Phase:   phase,
		Source:  source,
		Target:  target,
		JobID:   jobID,
		Message: message,
		Err:     err,
	}
}

// ConnectionError represents connection-related errors
type ConnectionError struct {
	Backend string
	URL     string
	Err     error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error for backend %s (url: %s): %v", e.Backend, e.URL, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewConnectionError creates a new connection error
func NewConnectionError(backend, url string, err error) *ConnectionError {
	return &ConnectionError{
		Backend: backend,
		URL:     url,
		Err:     err,
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Backend string
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for backend %s, field %s (value: %v, rule: %s): %s",
		e.Backend, e.Field, e.Value, e.Rule, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(backend, field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Backend: backend,
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// OperationError represents operation-specific errors
type OperationError struct {
	Backend   string
	Queue     string
	Operation string
	JobID     string
	Err       error
}

func (e *OperationError) Error() string {
	msg := fmt.Sprintf("operation %s failed on backend %s", e.Operation, e.Backend)
	if e.Queue != "" {
		msg += fmt.Sprintf(" (queue: %s)", e.Queue)
	}
	if e.JobID != "" {
		msg += fmt.Sprintf(" (job: %s)", e.JobID)
	}
	if e.Err != nil {
		msg += fmt.Sprintf(": %v", e.Err)
	}
	return msg
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// NewOperationError creates a new operation error
func NewOperationError(backend, queue, operation, jobID string, err error) *OperationError {
	return &OperationError{
		Backend:   backend,
		Queue:     queue,
		Operation: operation,
		JobID:     jobID,
		Err:       err,
	}
}

// IsRetryable returns true if the error indicates a retryable condition
func IsRetryable(err error) bool {
	switch {
	case errors.Is(err, ErrTimeout):
		return true
	case errors.Is(err, ErrConnectionFailed):
		return true
	case errors.Is(err, ErrCircuitBreakerOpen):
		return false // Don't retry when circuit breaker is open
	case errors.Is(err, ErrRateLimited):
		return true
	case errors.Is(err, ErrQueueEmpty):
		return false // Not an error condition
	case errors.Is(err, ErrJobNotFound):
		return false // Permanent condition
	case errors.Is(err, ErrJobAlreadyAcked):
		return false // Permanent condition
	case errors.Is(err, ErrInvalidJobData):
		return false // Permanent condition
	default:
		// Check if it's a backend error and examine the underlying error
		var backendErr *BackendError
		if errors.As(err, &backendErr) {
			return IsRetryable(backendErr.Err)
		}

		// Check if it's an operation error
		var opErr *OperationError
		if errors.As(err, &opErr) {
			return IsRetryable(opErr.Err)
		}

		// Default to not retryable for unknown errors
		return false
	}
}

// IsPermanent returns true if the error indicates a permanent failure
func IsPermanent(err error) bool {
	switch {
	case errors.Is(err, ErrJobNotFound):
		return true
	case errors.Is(err, ErrJobAlreadyAcked):
		return true
	case errors.Is(err, ErrInvalidJobData):
		return true
	case errors.Is(err, ErrInvalidConfiguration):
		return true
	case errors.Is(err, ErrOperationNotSupported):
		return true
	default:
		return false
	}
}

// IsTemporary returns true if the error indicates a temporary failure
func IsTemporary(err error) bool {
	return !IsPermanent(err) && IsRetryable(err)
}

// ErrorCode returns a stable error code for the error
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrBackendNotFound):
		return "BACKEND_NOT_FOUND"
	case errors.Is(err, ErrQueueNotFound):
		return "QUEUE_NOT_FOUND"
	case errors.Is(err, ErrJobNotFound):
		return "JOB_NOT_FOUND"
	case errors.Is(err, ErrJobAlreadyAcked):
		return "JOB_ALREADY_ACKED"
	case errors.Is(err, ErrJobProcessing):
		return "JOB_PROCESSING"
	case errors.Is(err, ErrInvalidConfiguration):
		return "INVALID_CONFIGURATION"
	case errors.Is(err, ErrConnectionFailed):
		return "CONNECTION_FAILED"
	case errors.Is(err, ErrOperationNotSupported):
		return "OPERATION_NOT_SUPPORTED"
	case errors.Is(err, ErrTimeout):
		return "TIMEOUT"
	case errors.Is(err, ErrQueueEmpty):
		return "QUEUE_EMPTY"
	case errors.Is(err, ErrMigrationInProgress):
		return "MIGRATION_IN_PROGRESS"
	case errors.Is(err, ErrMigrationFailed):
		return "MIGRATION_FAILED"
	case errors.Is(err, ErrConsumerGroupExists):
		return "CONSUMER_GROUP_EXISTS"
	case errors.Is(err, ErrStreamNotFound):
		return "STREAM_NOT_FOUND"
	case errors.Is(err, ErrInvalidJobData):
		return "INVALID_JOB_DATA"
	case errors.Is(err, ErrCircuitBreakerOpen):
		return "CIRCUIT_BREAKER_OPEN"
	case errors.Is(err, ErrRateLimited):
		return "RATE_LIMITED"
	default:
		// Check for typed errors
		var backendErr *BackendError
		if errors.As(err, &backendErr) {
			return "BACKEND_ERROR"
		}

		var configErr *ConfigurationError
		if errors.As(err, &configErr) {
			return "CONFIGURATION_ERROR"
		}

		var migrationErr *MigrationError
		if errors.As(err, &migrationErr) {
			return "MIGRATION_ERROR"
		}

		var connErr *ConnectionError
		if errors.As(err, &connErr) {
			return "CONNECTION_ERROR"
		}

		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return "VALIDATION_ERROR"
		}

		var opErr *OperationError
		if errors.As(err, &opErr) {
			return "OPERATION_ERROR"
		}

		return "UNKNOWN_ERROR"
	}
}