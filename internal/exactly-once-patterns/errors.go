// Copyright 2025 James Ross
package exactlyonce

import (
	"fmt"
)

// ErrIdempotencyKeyExists indicates that an idempotency key already exists
type ErrIdempotencyKeyExists struct {
	Key           string
	ExistingValue interface{}
	QueueName     string
	TenantID      string
}

func (e *ErrIdempotencyKeyExists) Error() string {
	if e.TenantID != "" {
		return fmt.Sprintf("idempotency key '%s' already exists in queue '%s' for tenant '%s'", e.Key, e.QueueName, e.TenantID)
	}
	return fmt.Sprintf("idempotency key '%s' already exists in queue '%s'", e.Key, e.QueueName)
}

// ErrIdempotencyKeyNotFound indicates that an idempotency key was not found
type ErrIdempotencyKeyNotFound struct {
	Key       string
	QueueName string
	TenantID  string
}

func (e *ErrIdempotencyKeyNotFound) Error() string {
	if e.TenantID != "" {
		return fmt.Sprintf("idempotency key '%s' not found in queue '%s' for tenant '%s'", e.Key, e.QueueName, e.TenantID)
	}
	return fmt.Sprintf("idempotency key '%s' not found in queue '%s'", e.Key, e.QueueName)
}

// ErrStorageUnavailable indicates that the storage backend is not available
type ErrStorageUnavailable struct {
	StorageType string
	Cause       error
}

func (e *ErrStorageUnavailable) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("storage backend '%s' is unavailable: %v", e.StorageType, e.Cause)
	}
	return fmt.Sprintf("storage backend '%s' is unavailable", e.StorageType)
}

func (e *ErrStorageUnavailable) Unwrap() error {
	return e.Cause
}

// ErrInvalidConfiguration indicates that the configuration is invalid
type ErrInvalidConfiguration struct {
	Field   string
	Value   interface{}
	Reason  string
}

func (e *ErrInvalidConfiguration) Error() string {
	return fmt.Sprintf("invalid configuration for field '%s' (value: %v): %s", e.Field, e.Value, e.Reason)
}

// ErrOutboxEventNotFound indicates that an outbox event was not found
type ErrOutboxEventNotFound struct {
	EventID string
}

func (e *ErrOutboxEventNotFound) Error() string {
	return fmt.Sprintf("outbox event '%s' not found", e.EventID)
}

// ErrPublisherNotFound indicates that a publisher was not found
type ErrPublisherNotFound struct {
	PublisherName string
}

func (e *ErrPublisherNotFound) Error() string {
	return fmt.Sprintf("publisher '%s' not found", e.PublisherName)
}

// ErrMaxRetriesExceeded indicates that maximum retries have been exceeded
type ErrMaxRetriesExceeded struct {
	Operation    string
	MaxRetries   int
	CurrentRetry int
	LastError    error
}

func (e *ErrMaxRetriesExceeded) Error() string {
	if e.LastError != nil {
		return fmt.Sprintf("max retries (%d) exceeded for operation '%s' at retry %d: %v", e.MaxRetries, e.Operation, e.CurrentRetry, e.LastError)
	}
	return fmt.Sprintf("max retries (%d) exceeded for operation '%s' at retry %d", e.MaxRetries, e.Operation, e.CurrentRetry)
}

func (e *ErrMaxRetriesExceeded) Unwrap() error {
	return e.LastError
}

// ErrTransactionFailed indicates that a database transaction failed
type ErrTransactionFailed struct {
	Operation string
	Cause     error
}

func (e *ErrTransactionFailed) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("transaction failed for operation '%s': %v", e.Operation, e.Cause)
	}
	return fmt.Sprintf("transaction failed for operation '%s'", e.Operation)
}

func (e *ErrTransactionFailed) Unwrap() error {
	return e.Cause
}

// ErrInvalidIdempotencyKey indicates that an idempotency key is invalid
type ErrInvalidIdempotencyKey struct {
	Key    string
	Reason string
}

func (e *ErrInvalidIdempotencyKey) Error() string {
	return fmt.Sprintf("invalid idempotency key '%s': %s", e.Key, e.Reason)
}

// ErrStorageCorruption indicates that storage data is corrupted
type ErrStorageCorruption struct {
	Key         string
	StorageType string
	Details     string
}

func (e *ErrStorageCorruption) Error() string {
	return fmt.Sprintf("storage corruption detected in %s for key '%s': %s", e.StorageType, e.Key, e.Details)
}

// ErrCircuitBreakerOpen indicates that the circuit breaker is open
type ErrCircuitBreakerOpen struct {
	Operation string
	Duration  string
}

func (e *ErrCircuitBreakerOpen) Error() string {
	return fmt.Sprintf("circuit breaker is open for operation '%s' (for %s)", e.Operation, e.Duration)
}

// ErrTenantNotAllowed indicates that a tenant is not allowed to perform the operation
type ErrTenantNotAllowed struct {
	TenantID  string
	Operation string
}

func (e *ErrTenantNotAllowed) Error() string {
	return fmt.Sprintf("tenant '%s' is not allowed to perform operation '%s'", e.TenantID, e.Operation)
}

// ErrQuotaExceeded indicates that a quota has been exceeded
type ErrQuotaExceeded struct {
	TenantID    string
	QueueName   string
	Quota       int64
	Current     int64
	QuotaType   string
}

func (e *ErrQuotaExceeded) Error() string {
	if e.TenantID != "" {
		return fmt.Sprintf("%s quota exceeded for tenant '%s' in queue '%s': %d/%d", e.QuotaType, e.TenantID, e.QueueName, e.Current, e.Quota)
	}
	return fmt.Sprintf("%s quota exceeded for queue '%s': %d/%d", e.QuotaType, e.QueueName, e.Current, e.Quota)
}

// Common error variables for frequently used errors
var (
	ErrIdempotencyDisabled = fmt.Errorf("idempotency checking is disabled")
	ErrOutboxDisabled     = fmt.Errorf("outbox pattern is disabled")
	ErrMetricsDisabled    = fmt.Errorf("metrics collection is disabled")
	ErrEmptyKey           = fmt.Errorf("idempotency key cannot be empty")
	ErrEmptyQueueName     = fmt.Errorf("queue name cannot be empty")
	ErrInvalidTTL         = fmt.Errorf("TTL must be positive")
)