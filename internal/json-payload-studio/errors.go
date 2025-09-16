package jsonpayloadstudio

import (
	"fmt"
)

// Common error types for JSON Payload Studio
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeSyntax        ErrorType = "syntax"
	ErrorTypeSchema        ErrorType = "schema"
	ErrorTypeSize          ErrorType = "size"
	ErrorTypeDepth         ErrorType = "depth"
	ErrorTypeFieldCount    ErrorType = "field_count"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeSession       ErrorType = "session"
	ErrorTypeTemplate      ErrorType = "template"
	ErrorTypeSnippet       ErrorType = "snippet"
	ErrorTypeHistory       ErrorType = "history"
	ErrorTypeEnqueue       ErrorType = "enqueue"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeUnsupported   ErrorType = "unsupported"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
)

// StudioError represents a structured error from the JSON Payload Studio
type StudioError struct {
	Type     ErrorType   `json:"type"`
	Message  string      `json:"message"`
	Details  interface{} `json:"details,omitempty"`
	Position *Position   `json:"position,omitempty"`
	Path     string      `json:"path,omitempty"`
	Code     string      `json:"code,omitempty"`
}

func (e *StudioError) Error() string {
	if e.Position != nil {
		return fmt.Sprintf("%s error at line %d, column %d: %s", e.Type, e.Position.Line, e.Position.Column, e.Message)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s error at path '%s': %s", e.Type, e.Path, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, position *Position) *StudioError {
	return &StudioError{
		Type:     ErrorTypeValidation,
		Message:  message,
		Position: position,
	}
}

// NewSyntaxError creates a new syntax error
func NewSyntaxError(message string, line, column int) *StudioError {
	return &StudioError{
		Type:    ErrorTypeSyntax,
		Message: message,
		Position: &Position{
			Line:   line,
			Column: column,
		},
	}
}

// NewSchemaError creates a new schema validation error
func NewSchemaError(message string, path string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeSchema,
		Message: message,
		Path:    path,
	}
}

// NewSizeError creates a new size limit error
func NewSizeError(message string, details interface{}) *StudioError {
	return &StudioError{
		Type:    ErrorTypeSize,
		Message: message,
		Details: details,
	}
}

// NewDepthError creates a new nesting depth error
func NewDepthError(maxDepth int, actualDepth int) *StudioError {
	return &StudioError{
		Type:    ErrorTypeDepth,
		Message: fmt.Sprintf("JSON nesting depth %d exceeds maximum allowed depth of %d", actualDepth, maxDepth),
		Details: map[string]int{
			"max_depth":    maxDepth,
			"actual_depth": actualDepth,
		},
	}
}

// NewFieldCountError creates a new field count error
func NewFieldCountError(maxFields int, actualFields int) *StudioError {
	return &StudioError{
		Type:    ErrorTypeFieldCount,
		Message: fmt.Sprintf("JSON contains %d fields, exceeding maximum of %d", actualFields, maxFields),
		Details: map[string]int{
			"max_fields":    maxFields,
			"actual_fields": actualFields,
		},
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resourceType string, id string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s with ID '%s' not found", resourceType, id),
		Details: map[string]string{
			"resource_type": resourceType,
			"resource_id":   id,
		},
	}
}

// NewSessionError creates a new session error
func NewSessionError(message string, sessionID string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeSession,
		Message: message,
		Details: map[string]string{
			"session_id": sessionID,
		},
	}
}

// NewTemplateError creates a new template error
func NewTemplateError(message string, templateID string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeTemplate,
		Message: message,
		Details: map[string]string{
			"template_id": templateID,
		},
	}
}

// NewSnippetError creates a new snippet error
func NewSnippetError(message string, trigger string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeSnippet,
		Message: message,
		Details: map[string]string{
			"trigger": trigger,
		},
	}
}

// NewHistoryError creates a new history error
func NewHistoryError(message string, action string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeHistory,
		Message: message,
		Details: map[string]string{
			"action": action,
		},
	}
}

// NewEnqueueError creates a new enqueue error
func NewEnqueueError(message string, queue string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeEnqueue,
		Message: message,
		Details: map[string]string{
			"queue": queue,
		},
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *StudioError {
	details := map[string]string{
		"error": err.Error(),
	}
	return &StudioError{
		Type:    ErrorTypeInternal,
		Message: message,
		Details: details,
	}
}

// NewUnsupportedError creates a new unsupported feature error
func NewUnsupportedError(feature string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeUnsupported,
		Message: fmt.Sprintf("Feature '%s' is not supported", feature),
		Details: map[string]string{
			"feature": feature,
		},
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, duration string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("Operation '%s' timed out after %s", operation, duration),
		Details: map[string]string{
			"operation": operation,
			"duration":  duration,
		},
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(limit int, window string) *StudioError {
	return &StudioError{
		Type:    ErrorTypeRateLimit,
		Message: fmt.Sprintf("Rate limit exceeded: %d requests per %s", limit, window),
		Details: map[string]interface{}{
			"limit":  limit,
			"window": window,
		},
	}
}

// ErrorCollection represents a collection of errors
type ErrorCollection struct {
	Errors   []*StudioError `json:"errors"`
	Warnings []*StudioError `json:"warnings,omitempty"`
	Info     []*StudioError `json:"info,omitempty"`
}

// NewErrorCollection creates a new error collection
func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		Errors:   make([]*StudioError, 0),
		Warnings: make([]*StudioError, 0),
		Info:     make([]*StudioError, 0),
	}
}

// AddError adds an error to the collection
func (ec *ErrorCollection) AddError(err *StudioError) {
	ec.Errors = append(ec.Errors, err)
}

// AddWarning adds a warning to the collection
func (ec *ErrorCollection) AddWarning(err *StudioError) {
	ec.Warnings = append(ec.Warnings, err)
}

// AddInfo adds an info message to the collection
func (ec *ErrorCollection) AddInfo(err *StudioError) {
	ec.Info = append(ec.Info, err)
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (ec *ErrorCollection) HasWarnings() bool {
	return len(ec.Warnings) > 0
}

// Count returns the total number of errors, warnings, and info messages
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors) + len(ec.Warnings) + len(ec.Info)
}

// Clear removes all errors, warnings, and info messages
func (ec *ErrorCollection) Clear() {
	ec.Errors = make([]*StudioError, 0)
	ec.Warnings = make([]*StudioError, 0)
	ec.Info = make([]*StudioError, 0)
}

// ToValidationErrors converts the error collection to validation errors
func (ec *ErrorCollection) ToValidationErrors() []ValidationError {
	var result []ValidationError

	for _, err := range ec.Errors {
		ve := ValidationError{
			Type:     string(err.Type),
			Message:  err.Message,
			Severity: "error",
		}
		if err.Position != nil {
			ve.Line = err.Position.Line
			ve.Column = err.Position.Column
		}
		if err.Path != "" {
			ve.Path = err.Path
		}
		result = append(result, ve)
	}

	for _, warn := range ec.Warnings {
		ve := ValidationError{
			Type:     string(warn.Type),
			Message:  warn.Message,
			Severity: "warning",
		}
		if warn.Position != nil {
			ve.Line = warn.Position.Line
			ve.Column = warn.Position.Column
		}
		if warn.Path != "" {
			ve.Path = warn.Path
		}
		result = append(result, ve)
	}

	for _, info := range ec.Info {
		ve := ValidationError{
			Type:     string(info.Type),
			Message:  info.Message,
			Severity: "info",
		}
		if info.Position != nil {
			ve.Line = info.Position.Line
			ve.Column = info.Position.Column
		}
		if info.Path != "" {
			ve.Path = info.Path
		}
		result = append(result, ve)
	}

	return result
}

// Error implements the error interface for ErrorCollection
func (ec *ErrorCollection) Error() string {
	if !ec.HasErrors() {
		return "no errors"
	}

	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}

	return fmt.Sprintf("%d errors occurred", len(ec.Errors))
}