// Copyright 2025 James Ross
package eventhooks

import (
	"errors"
	"fmt"
)

// Common event hooks errors
var (
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrInvalidSubscriptionID    = errors.New("invalid subscription ID")
	ErrInvalidWebhookURL        = errors.New("invalid webhook URL")
	ErrInvalidEventType         = errors.New("invalid event type")
	ErrInvalidRetryPolicy       = errors.New("invalid retry policy")
	ErrMaxRetriesExceeded       = errors.New("maximum retries exceeded")
	ErrSubscriptionDisabled     = errors.New("subscription is disabled")
	ErrRateLimitExceeded        = errors.New("rate limit exceeded")
	ErrCircuitBreakerOpen       = errors.New("circuit breaker is open")
	ErrDeliveryTimeout          = errors.New("delivery timeout")
	ErrInvalidSignature         = errors.New("invalid HMAC signature")
	ErrEventBusShutdown         = errors.New("event bus is shutting down")
	ErrDuplicateSubscription    = errors.New("subscription with this name already exists")
	ErrInvalidFilterCriteria    = errors.New("invalid filter criteria")
)

// DeliveryError represents an error during webhook delivery
type DeliveryError struct {
	SubscriptionID string
	EventID        string
	AttemptNumber  int
	StatusCode     int
	Message        string
	Retryable      bool
	Err            error
}

func (e *DeliveryError) Error() string {
	return fmt.Sprintf("delivery failed for subscription %s (attempt %d): %s",
		e.SubscriptionID, e.AttemptNumber, e.Message)
}

func (e *DeliveryError) Unwrap() error {
	return e.Err
}

func (e *DeliveryError) IsRetryable() bool {
	return e.Retryable
}

// NewDeliveryError creates a new delivery error
func NewDeliveryError(subscriptionID, eventID string, attemptNumber, statusCode int, message string, retryable bool, err error) *DeliveryError {
	return &DeliveryError{
		SubscriptionID: subscriptionID,
		EventID:        eventID,
		AttemptNumber:  attemptNumber,
		StatusCode:     statusCode,
		Message:        message,
		Retryable:      retryable,
		Err:            err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)",
		e.Field, e.Message, e.Value)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ConfigurationError represents a configuration error
type ConfigurationError struct {
	Component string
	Message   string
	Err       error
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error in %s: %s", e.Component, e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(component, message string, err error) *ConfigurationError {
	return &ConfigurationError{
		Component: component,
		Message:   message,
		Err:       err,
	}
}

// RetryableError indicates whether an error should trigger a retry
func IsRetryableError(err error) bool {
	var deliveryErr *DeliveryError
	if errors.As(err, &deliveryErr) {
		return deliveryErr.IsRetryable()
	}

	// Check for specific error types
	switch err {
	case ErrDeliveryTimeout, ErrCircuitBreakerOpen:
		return true
	case ErrInvalidSignature, ErrInvalidWebhookURL, ErrSubscriptionDisabled:
		return false
	default:
		// Default to retryable for unknown errors
		return true
	}
}

// IsTemporaryError checks if an error is temporary and should be retried
func IsTemporaryError(statusCode int) bool {
	switch {
	case statusCode >= 500 && statusCode < 600: // 5xx server errors
		return true
	case statusCode == 408: // Request timeout
		return true
	case statusCode == 429: // Too many requests
		return true
	case statusCode >= 400 && statusCode < 500: // 4xx client errors
		return false
	default:
		return false
	}
}