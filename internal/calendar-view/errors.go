package calendarview

import (
	"fmt"
	"time"
)

// CalendarError represents a custom error type for calendar operations
type CalendarError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
}

// ErrorCode represents different types of calendar errors
type ErrorCode int

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeInvalidTimeRange
	ErrorCodeInvalidCronSpec
	ErrorCodeTimezoneNotFound
	ErrorCodeEventNotFound
	ErrorCodeRuleNotFound
	ErrorCodeRescheduleConflict
	ErrorCodeRuleValidation
	ErrorCodeDatabaseError
	ErrorCodePermissionDenied
	ErrorCodeRateLimited
	ErrorCodeInvalidFilter
	ErrorCodeMaxEventsExceeded
)

// Error implements the error interface
func (e *CalendarError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap returns the underlying cause if present
func (e *CalendarError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches another error
func (e *CalendarError) Is(target error) bool {
	if other, ok := target.(*CalendarError); ok {
		return e.Code == other.Code
	}
	return false
}

// NewCalendarError creates a new calendar error
func NewCalendarError(code ErrorCode, message string, details ...string) *CalendarError {
	err := &CalendarError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// WrapCalendarError wraps an existing error with calendar context
func WrapCalendarError(code ErrorCode, message string, cause error) *CalendarError {
	return &CalendarError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Predefined error constructors

// ErrInvalidTimeRange creates an error for invalid time ranges
func ErrInvalidTimeRange(start, end time.Time) *CalendarError {
	return NewCalendarError(
		ErrorCodeInvalidTimeRange,
		"invalid time range",
		fmt.Sprintf("start %v must be before end %v", start, end),
	)
}

// ErrInvalidCronSpec creates an error for invalid cron specifications
func ErrInvalidCronSpec(spec string) *CalendarError {
	return NewCalendarError(
		ErrorCodeInvalidCronSpec,
		"invalid cron specification",
		fmt.Sprintf("cron spec '%s' is not valid", spec),
	)
}

// ErrTimezoneNotFound creates an error for unknown timezones
func ErrTimezoneNotFound(timezone string) *CalendarError {
	return NewCalendarError(
		ErrorCodeTimezoneNotFound,
		"timezone not found",
		fmt.Sprintf("timezone '%s' is not recognized", timezone),
	)
}

// ErrEventNotFound creates an error for missing events
func ErrEventNotFound(eventID string) *CalendarError {
	return NewCalendarError(
		ErrorCodeEventNotFound,
		"event not found",
		fmt.Sprintf("event with ID '%s' does not exist", eventID),
	)
}

// ErrRuleNotFound creates an error for missing recurring rules
func ErrRuleNotFound(ruleID string) *CalendarError {
	return NewCalendarError(
		ErrorCodeRuleNotFound,
		"recurring rule not found",
		fmt.Sprintf("rule with ID '%s' does not exist", ruleID),
	)
}

// ErrRescheduleConflict creates an error for reschedule conflicts
func ErrRescheduleConflict(eventID string, newTime time.Time) *CalendarError {
	return NewCalendarError(
		ErrorCodeRescheduleConflict,
		"reschedule conflict",
		fmt.Sprintf("event '%s' cannot be rescheduled to %v due to conflicts", eventID, newTime),
	)
}

// ErrRuleValidation creates an error for rule validation failures
func ErrRuleValidation(field, reason string) *CalendarError {
	return NewCalendarError(
		ErrorCodeRuleValidation,
		"rule validation failed",
		fmt.Sprintf("field '%s': %s", field, reason),
	)
}

// ErrDatabaseError creates an error for database operation failures
func ErrDatabaseError(operation string, cause error) *CalendarError {
	return WrapCalendarError(
		ErrorCodeDatabaseError,
		fmt.Sprintf("database error during %s", operation),
		cause,
	)
}

// ErrPermissionDenied creates an error for permission issues
func ErrPermissionDenied(userID, action string) *CalendarError {
	return NewCalendarError(
		ErrorCodePermissionDenied,
		"permission denied",
		fmt.Sprintf("user '%s' cannot perform action '%s'", userID, action),
	)
}

// ErrRateLimited creates an error for rate limiting
func ErrRateLimited(userID string, retryAfter time.Duration) *CalendarError {
	return NewCalendarError(
		ErrorCodeRateLimited,
		"rate limited",
		fmt.Sprintf("user '%s' is rate limited, retry after %v", userID, retryAfter),
	)
}

// ErrInvalidFilter creates an error for invalid filter parameters
func ErrInvalidFilter(field, value string) *CalendarError {
	return NewCalendarError(
		ErrorCodeInvalidFilter,
		"invalid filter",
		fmt.Sprintf("filter field '%s' has invalid value '%s'", field, value),
	)
}

// ErrMaxEventsExceeded creates an error for too many events
func ErrMaxEventsExceeded(count, max int) *CalendarError {
	return NewCalendarError(
		ErrorCodeMaxEventsExceeded,
		"maximum events exceeded",
		fmt.Sprintf("requested %d events exceeds maximum limit of %d", count, max),
	)
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if calErr, ok := err.(*CalendarError); ok {
		switch calErr.Code {
		case ErrorCodeDatabaseError, ErrorCodeRateLimited:
			return true
		default:
			return false
		}
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if calErr, ok := err.(*CalendarError); ok {
		return calErr.Code
	}
	return ErrorCodeUnknown
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error implements the error interface for ValidationErrors
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}
	if len(ve.Errors) == 1 {
		return fmt.Sprintf("validation failed: %s", ve.Errors[0].Message)
	}
	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string, value ...string) {
	err := ValidationError{
		Field:   field,
		Message: message,
	}
	if len(value) > 0 {
		err.Value = value[0]
	}
	ve.Errors = append(ve.Errors, err)
}

// HasErrors returns true if there are any validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// NewValidationErrors creates a new ValidationErrors instance
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}