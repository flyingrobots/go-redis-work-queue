// Copyright 2025 James Ross
package backpressure

import (
	"errors"
	"fmt"
)

// Common backpressure errors
var (
	ErrJobShed              = errors.New("job shed due to backpressure")
	ErrCircuitOpen          = errors.New("circuit breaker is open")
	ErrControllerNotStarted = errors.New("backpressure controller not started")
	ErrControllerStopped    = errors.New("backpressure controller stopped")
	ErrInvalidPriority      = errors.New("invalid priority level")
	ErrInvalidQueue         = errors.New("invalid queue name")
	ErrStatsUnavailable     = errors.New("queue statistics unavailable")
	ErrConfigInvalid        = errors.New("invalid configuration")
	ErrThresholdExceeded    = errors.New("backlog threshold exceeded")
	ErrPollingFailed        = errors.New("failed to poll queue statistics")
)

// BackpressureError wraps errors with additional context
type BackpressureError struct {
	Op       string // Operation that failed
	Queue    string // Queue name (if applicable)
	Priority Priority
	Err      error
}

func (e *BackpressureError) Error() string {
	if e.Queue != "" && e.Priority >= 0 {
		return fmt.Sprintf("backpressure %s failed for queue %s priority %s: %v",
			e.Op, e.Queue, e.Priority.String(), e.Err)
	} else if e.Queue != "" {
		return fmt.Sprintf("backpressure %s failed for queue %s: %v", e.Op, e.Queue, e.Err)
	}
	return fmt.Sprintf("backpressure %s failed: %v", e.Op, e.Err)
}

func (e *BackpressureError) Unwrap() error {
	return e.Err
}

// NewBackpressureError creates a new BackpressureError
func NewBackpressureError(op, queue string, priority Priority, err error) *BackpressureError {
	return &BackpressureError{
		Op:       op,
		Queue:    queue,
		Priority: priority,
		Err:      err,
	}
}

// CircuitBreakerError represents circuit breaker related errors
type CircuitBreakerError struct {
	Queue   string
	State   CircuitState
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker for queue %s is %s: %s",
		e.Queue, e.State.String(), e.Message)
}

// NewCircuitBreakerError creates a new CircuitBreakerError
func NewCircuitBreakerError(queue string, state CircuitState, message string) *CircuitBreakerError {
	return &CircuitBreakerError{
		Queue:   queue,
		State:   state,
		Message: message,
	}
}

// ConfigurationError represents configuration validation errors
type ConfigurationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in field %s (value: %v): %s",
		e.Field, e.Value, e.Message)
}

// NewConfigurationError creates a new ConfigurationError
func NewConfigurationError(field string, value interface{}, message string) *ConfigurationError {
	return &ConfigurationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// PollingError represents errors during queue statistics polling
type PollingError struct {
	Queue     string
	Attempt   int
	LastError error
	Backoff   bool
}

func (e *PollingError) Error() string {
	status := ""
	if e.Backoff {
		status = " (backing off)"
	}
	return fmt.Sprintf("polling failed for queue %s (attempt %d)%s: %v",
		e.Queue, e.Attempt, status, e.LastError)
}

func (e *PollingError) Unwrap() error {
	return e.LastError
}

// NewPollingError creates a new PollingError
func NewPollingError(queue string, attempt int, err error, backoff bool) *PollingError {
	return &PollingError{
		Queue:     queue,
		Attempt:   attempt,
		LastError: err,
		Backoff:   backoff,
	}
}

// IsBackpressureError checks if error is a backpressure-related error
func IsBackpressureError(err error) bool {
	var bpErr *BackpressureError
	var cbErr *CircuitBreakerError
	var cfgErr *ConfigurationError
	var polErr *PollingError

	return errors.As(err, &bpErr) ||
		   errors.As(err, &cbErr) ||
		   errors.As(err, &cfgErr) ||
		   errors.As(err, &polErr)
}

// IsRetryable returns true if the error might be resolved by retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Specific non-retryable errors
	switch {
	case errors.Is(err, ErrJobShed):
		return false
	case errors.Is(err, ErrControllerStopped):
		return false
	case errors.Is(err, ErrInvalidPriority):
		return false
	case errors.Is(err, ErrInvalidQueue):
		return false
	case errors.Is(err, ErrConfigInvalid):
		return false
	}

	// Circuit breaker errors - retryable based on state
	var cbErr *CircuitBreakerError
	if errors.As(err, &cbErr) {
		return cbErr.State != Open // Closed and HalfOpen are retryable
	}

	// Polling errors are generally retryable
	var polErr *PollingError
	if errors.As(err, &polErr) {
		return true
	}

	// Other backpressure errors might be retryable
	var bpErr *BackpressureError
	if errors.As(err, &bpErr) {
		return IsRetryable(bpErr.Err)
	}

	// Unknown errors are assumed retryable
	return true
}

// IsShedError returns true if the error indicates job shedding
func IsShedError(err error) bool {
	return errors.Is(err, ErrJobShed)
}

// IsCircuitOpenError returns true if the error is due to an open circuit breaker
func IsCircuitOpenError(err error) bool {
	if errors.Is(err, ErrCircuitOpen) {
		return true
	}

	var cbErr *CircuitBreakerError
	if errors.As(err, &cbErr) {
		return cbErr.State == Open
	}

	return false
}