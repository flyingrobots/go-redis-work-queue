package context_menus

import (
	"fmt"
)

// ErrorCode represents different types of context menu errors
type ErrorCode string

const (
	// Zone-related errors
	ErrCodeEmptyZoneID     ErrorCode = "EMPTY_ZONE_ID"
	ErrCodeZoneNotFound    ErrorCode = "ZONE_NOT_FOUND"
	ErrCodeInvalidZoneBounds ErrorCode = "INVALID_ZONE_BOUNDS"
	ErrCodeDuplicateZone   ErrorCode = "DUPLICATE_ZONE"

	// Action-related errors
	ErrCodeActionNotFound    ErrorCode = "ACTION_NOT_FOUND"
	ErrCodeInvalidAction     ErrorCode = "INVALID_ACTION"
	ErrCodeActionNotAllowed  ErrorCode = "ACTION_NOT_ALLOWED"
	ErrCodeMissingHandler    ErrorCode = "MISSING_HANDLER"

	// Context-related errors
	ErrCodeInvalidContext    ErrorCode = "INVALID_CONTEXT"
	ErrCodeMissingContext    ErrorCode = "MISSING_CONTEXT"
	ErrCodeContextMismatch   ErrorCode = "CONTEXT_MISMATCH"

	// Configuration errors
	ErrCodeInvalidConfig     ErrorCode = "INVALID_CONFIG"
	ErrCodeInvalidMinWidth   ErrorCode = "INVALID_MIN_WIDTH"
	ErrCodeInvalidMaxWidth   ErrorCode = "INVALID_MAX_WIDTH"
	ErrCodeInvalidDuration   ErrorCode = "INVALID_DURATION"
	ErrCodeInvalidTimeout    ErrorCode = "INVALID_TIMEOUT"

	// Menu-related errors
	ErrCodeMenuNotVisible    ErrorCode = "MENU_NOT_VISIBLE"
	ErrCodeMenuAlreadyVisible ErrorCode = "MENU_ALREADY_VISIBLE"
	ErrCodeInvalidMenuPosition ErrorCode = "INVALID_MENU_POSITION"

	// System errors
	ErrCodeSystemDisabled    ErrorCode = "SYSTEM_DISABLED"
	ErrCodeInitializationFailed ErrorCode = "INITIALIZATION_FAILED"
)

// ContextMenuError represents an error in the context menu system
type ContextMenuError struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
	Cause   error
}

// Error implements the error interface
func (e *ContextMenuError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ContextMenuError) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *ContextMenuError) WithDetails(key string, value interface{}) *ContextMenuError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause adds a cause error
func (e *ContextMenuError) WithCause(cause error) *ContextMenuError {
	e.Cause = cause
	return e
}

// NewError creates a new ContextMenuError
func NewError(code ErrorCode, message string) *ContextMenuError {
	return &ContextMenuError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Pre-defined error instances
var (
	ErrEmptyZoneID = NewError(ErrCodeEmptyZoneID, "zone ID cannot be empty")
	ErrZoneNotFound = NewError(ErrCodeZoneNotFound, "bubble zone not found")
	ErrActionNotFound = NewError(ErrCodeActionNotFound, "action not found in registry")
	ErrInvalidContext = NewError(ErrCodeInvalidContext, "invalid context for menu action")
	ErrSystemDisabled = NewError(ErrCodeSystemDisabled, "context menu system is disabled")
	ErrMenuNotVisible = NewError(ErrCodeMenuNotVisible, "menu is not currently visible")
)

// Zone-specific errors
func NewZoneNotFoundError(zoneID string) *ContextMenuError {
	return NewError(ErrCodeZoneNotFound, fmt.Sprintf("zone '%s' not found", zoneID)).
		WithDetails("zoneID", zoneID)
}

func NewInvalidZoneBoundsError(x, y, width, height int) *ContextMenuError {
	return NewError(ErrCodeInvalidZoneBounds, "invalid zone bounds").
		WithDetails("x", x).
		WithDetails("y", y).
		WithDetails("width", width).
		WithDetails("height", height)
}

func NewDuplicateZoneError(zoneID string) *ContextMenuError {
	return NewError(ErrCodeDuplicateZone, fmt.Sprintf("zone '%s' already exists", zoneID)).
		WithDetails("zoneID", zoneID)
}

// Action-specific errors
func NewActionNotFoundError(actionID string, contextType ContextType) *ContextMenuError {
	return NewError(ErrCodeActionNotFound, fmt.Sprintf("action '%s' not found for context type %d", actionID, int(contextType))).
		WithDetails("actionID", actionID).
		WithDetails("contextType", contextType)
}

func NewInvalidActionError(actionID string, reason string) *ContextMenuError {
	return NewError(ErrCodeInvalidAction, fmt.Sprintf("invalid action '%s': %s", actionID, reason)).
		WithDetails("actionID", actionID).
		WithDetails("reason", reason)
}

func NewActionNotAllowedError(actionID string, context MenuContext) *ContextMenuError {
	return NewError(ErrCodeActionNotAllowed, fmt.Sprintf("action '%s' not allowed in current context", actionID)).
		WithDetails("actionID", actionID).
		WithDetails("context", context)
}

func NewMissingHandlerError(actionID string) *ContextMenuError {
	return NewError(ErrCodeMissingHandler, fmt.Sprintf("no handler registered for action '%s'", actionID)).
		WithDetails("actionID", actionID)
}

// Context-specific errors
func NewInvalidContextError(contextType ContextType, reason string) *ContextMenuError {
	return NewError(ErrCodeInvalidContext, fmt.Sprintf("invalid context type %d: %s", int(contextType), reason)).
		WithDetails("contextType", contextType).
		WithDetails("reason", reason)
}

func NewMissingContextError(field string) *ContextMenuError {
	return NewError(ErrCodeMissingContext, fmt.Sprintf("missing required context field: %s", field)).
		WithDetails("field", field)
}

func NewContextMismatchError(expected, actual ContextType) *ContextMenuError {
	return NewError(ErrCodeContextMismatch, fmt.Sprintf("context type mismatch: expected %d, got %d", int(expected), int(actual))).
		WithDetails("expected", expected).
		WithDetails("actual", actual)
}

// Configuration errors
func NewInvalidConfigError(field string, value interface{}, reason string) *ContextMenuError {
	return NewError(ErrCodeInvalidConfig, fmt.Sprintf("invalid configuration for %s: %s", field, reason)).
		WithDetails("field", field).
		WithDetails("value", value).
		WithDetails("reason", reason)
}

// Menu-specific errors
func NewInvalidMenuPositionError(x, y int, maxWidth, maxHeight int) *ContextMenuError {
	return NewError(ErrCodeInvalidMenuPosition, "menu position would be outside screen bounds").
		WithDetails("x", x).
		WithDetails("y", y).
		WithDetails("maxWidth", maxWidth).
		WithDetails("maxHeight", maxHeight)
}

// System errors
func NewInitializationFailedError(component string, cause error) *ContextMenuError {
	return NewError(ErrCodeInitializationFailed, fmt.Sprintf("failed to initialize %s", component)).
		WithDetails("component", component).
		WithCause(cause)
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if cmErr, ok := err.(*ContextMenuError); ok {
		return cmErr.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if cmErr, ok := err.(*ContextMenuError); ok {
		return cmErr.Code
	}
	return ""
}

// GetErrorDetails extracts the error details from an error
func GetErrorDetails(err error) map[string]interface{} {
	if cmErr, ok := err.(*ContextMenuError); ok {
		return cmErr.Details
	}
	return nil
}