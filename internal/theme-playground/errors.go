// Copyright 2025 James Ross
package themeplayground

import (
	"fmt"
)

// ThemePlaygroundError represents errors specific to the theme playground
type ThemePlaygroundError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Component string `json:"component,omitempty"`
}

func (e *ThemePlaygroundError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewThemePlaygroundError creates a new theme playground error
func NewThemePlaygroundError(code, message string) *ThemePlaygroundError {
	return &ThemePlaygroundError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *ThemePlaygroundError) WithDetails(details string) *ThemePlaygroundError {
	return &ThemePlaygroundError{
		Code:      e.Code,
		Message:   e.Message,
		Details:   details,
		Component: e.Component,
	}
}

// WithComponent adds component context to the error
func (e *ThemePlaygroundError) WithComponent(component string) *ThemePlaygroundError {
	return &ThemePlaygroundError{
		Code:      e.Code,
		Message:   e.Message,
		Details:   e.Details,
		Component: component,
	}
}

// Error constants for theme playground operations
var (
	// Theme management errors
	ErrPlaygroundNotInitialized = NewThemePlaygroundError(
		"PLAYGROUND_NOT_INITIALIZED",
		"theme playground has not been initialized",
	)

	ErrThemeRegistrationFailed = NewThemePlaygroundError(
		"THEME_REGISTRATION_FAILED",
		"failed to register theme in playground",
	)

	ErrThemeUpdateFailed = NewThemePlaygroundError(
		"THEME_UPDATE_FAILED",
		"failed to update theme",
	)

	// Validation errors
	ErrThemeValidationFailed = NewThemePlaygroundError(
		"THEME_VALIDATION_FAILED",
		"theme validation failed",
	)

	ErrInvalidThemeFormat = NewThemePlaygroundError(
		"INVALID_THEME_FORMAT",
		"theme format is invalid",
	)

	ErrIncompatibleThemeVersion = NewThemePlaygroundError(
		"INCOMPATIBLE_THEME_VERSION",
		"theme version is incompatible",
	)

	// File system errors
	ErrConfigDirectoryNotFound = NewThemePlaygroundError(
		"CONFIG_DIRECTORY_NOT_FOUND",
		"configuration directory not found",
	)

	ErrThemeFileCorrupted = NewThemePlaygroundError(
		"THEME_FILE_CORRUPTED",
		"theme file is corrupted or unreadable",
	)

	ErrInsufficientPermissions = NewThemePlaygroundError(
		"INSUFFICIENT_PERMISSIONS",
		"insufficient permissions to access theme files",
	)

	// Runtime errors
	ErrThemeNotActive = NewThemePlaygroundError(
		"THEME_NOT_ACTIVE",
		"no theme is currently active",
	)

	ErrCircularThemeDependency = NewThemePlaygroundError(
		"CIRCULAR_THEME_DEPENDENCY",
		"circular dependency detected in theme hierarchy",
	)

	ErrThemeConflict = NewThemePlaygroundError(
		"THEME_CONFLICT",
		"theme conflicts with existing configuration",
	)

	// Import/Export errors
	ErrThemeImportFailed = NewThemePlaygroundError(
		"THEME_IMPORT_FAILED",
		"failed to import theme",
	)

	ErrThemeExportFailed = NewThemePlaygroundError(
		"THEME_EXPORT_FAILED",
		"failed to export theme",
	)

	ErrUnsupportedThemeFormat = NewThemePlaygroundError(
		"UNSUPPORTED_THEME_FORMAT",
		"theme format is not supported",
	)

	// Cache errors
	ErrThemeCacheCorrupted = NewThemePlaygroundError(
		"THEME_CACHE_CORRUPTED",
		"theme cache is corrupted",
	)

	ErrCacheWriteFailed = NewThemePlaygroundError(
		"CACHE_WRITE_FAILED",
		"failed to write to theme cache",
	)

	// Network/API errors
	ErrAPIEndpointNotAvailable = NewThemePlaygroundError(
		"API_ENDPOINT_NOT_AVAILABLE",
		"theme API endpoint is not available",
	)

	ErrInvalidAPIRequest = NewThemePlaygroundError(
		"INVALID_API_REQUEST",
		"invalid API request for theme operation",
	)

	// Resource errors
	ErrThemeResourceNotFound = NewThemePlaygroundError(
		"THEME_RESOURCE_NOT_FOUND",
		"theme resource not found",
	)

	ErrResourceLimitExceeded = NewThemePlaygroundError(
		"RESOURCE_LIMIT_EXCEEDED",
		"theme resource limit exceeded",
	)

	// Integration errors
	ErrIntegrationNotSupported = NewThemePlaygroundError(
		"INTEGRATION_NOT_SUPPORTED",
		"theme integration is not supported",
	)

	ErrExternalServiceUnavailable = NewThemePlaygroundError(
		"EXTERNAL_SERVICE_UNAVAILABLE",
		"external theme service is unavailable",
	)
)

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	if tpErr, ok := err.(*ThemePlaygroundError); ok {
		recoverableCodes := map[string]bool{
			"THEME_NOT_FOUND":              true,
			"THEME_VALIDATION_FAILED":      true,
			"INVALID_THEME_FORMAT":         true,
			"THEME_FILE_CORRUPTED":         true,
			"THEME_CACHE_CORRUPTED":        true,
			"API_ENDPOINT_NOT_AVAILABLE":   true,
			"EXTERNAL_SERVICE_UNAVAILABLE": true,
		}
		return recoverableCodes[tpErr.Code]
	}
	return false
}

// IsCritical checks if an error is critical and requires immediate attention
func IsCritical(err error) bool {
	if tpErr, ok := err.(*ThemePlaygroundError); ok {
		criticalCodes := map[string]bool{
			"PLAYGROUND_NOT_INITIALIZED":   true,
			"CONFIG_DIRECTORY_NOT_FOUND":   true,
			"INSUFFICIENT_PERMISSIONS":     true,
			"CIRCULAR_THEME_DEPENDENCY":    true,
			"RESOURCE_LIMIT_EXCEEDED":      true,
		}
		return criticalCodes[tpErr.Code]
	}
	return false
}

// ErrorCategory categorizes errors for better handling
type ErrorCategory string

const (
	CategoryValidation   ErrorCategory = "validation"
	CategoryFileSystem   ErrorCategory = "filesystem"
	CategoryRuntime      ErrorCategory = "runtime"
	CategoryImportExport ErrorCategory = "import_export"
	CategoryCache        ErrorCategory = "cache"
	CategoryAPI          ErrorCategory = "api"
	CategoryResource     ErrorCategory = "resource"
	CategoryIntegration  ErrorCategory = "integration"
)

// GetErrorCategory returns the category of an error
func GetErrorCategory(err error) ErrorCategory {
	if tpErr, ok := err.(*ThemePlaygroundError); ok {
		categoryMap := map[string]ErrorCategory{
			"THEME_VALIDATION_FAILED":      CategoryValidation,
			"INVALID_THEME_FORMAT":         CategoryValidation,
			"INCOMPATIBLE_THEME_VERSION":   CategoryValidation,
			"CONFIG_DIRECTORY_NOT_FOUND":   CategoryFileSystem,
			"THEME_FILE_CORRUPTED":         CategoryFileSystem,
			"INSUFFICIENT_PERMISSIONS":     CategoryFileSystem,
			"THEME_NOT_ACTIVE":             CategoryRuntime,
			"CIRCULAR_THEME_DEPENDENCY":    CategoryRuntime,
			"THEME_CONFLICT":               CategoryRuntime,
			"THEME_IMPORT_FAILED":          CategoryImportExport,
			"THEME_EXPORT_FAILED":          CategoryImportExport,
			"UNSUPPORTED_THEME_FORMAT":     CategoryImportExport,
			"THEME_CACHE_CORRUPTED":        CategoryCache,
			"CACHE_WRITE_FAILED":           CategoryCache,
			"API_ENDPOINT_NOT_AVAILABLE":   CategoryAPI,
			"INVALID_API_REQUEST":          CategoryAPI,
			"THEME_RESOURCE_NOT_FOUND":     CategoryResource,
			"RESOURCE_LIMIT_EXCEEDED":      CategoryResource,
			"INTEGRATION_NOT_SUPPORTED":    CategoryIntegration,
			"EXTERNAL_SERVICE_UNAVAILABLE": CategoryIntegration,
		}

		if category, exists := categoryMap[tpErr.Code]; exists {
			return category
		}
	}

	return CategoryRuntime // Default category
}

// ErrorWithRecoveryAction provides error context with suggested recovery actions
type ErrorWithRecoveryAction struct {
	Err            error
	RecoveryAction string
	UserMessage    string
}

// Error implements the error interface
func (e *ErrorWithRecoveryAction) Error() string {
	return e.Err.Error()
}

// WrapWithRecovery wraps an error with recovery action information
func WrapWithRecovery(err error, action, userMessage string) *ErrorWithRecoveryAction {
	return &ErrorWithRecoveryAction{
		Err:            err,
		RecoveryAction: action,
		UserMessage:    userMessage,
	}
}

// GetRecoveryAction returns a suggested recovery action for common errors
func GetRecoveryAction(err error) string {
	if tpErr, ok := err.(*ThemePlaygroundError); ok {
		actions := map[string]string{
			"THEME_NOT_FOUND":              "Check theme name spelling or create the theme",
			"THEME_VALIDATION_FAILED":      "Review theme configuration and fix validation errors",
			"INVALID_THEME_FORMAT":         "Ensure theme follows the correct JSON schema",
			"CONFIG_DIRECTORY_NOT_FOUND":   "Create configuration directory or check permissions",
			"THEME_FILE_CORRUPTED":         "Delete corrupted file and re-import theme",
			"INSUFFICIENT_PERMISSIONS":     "Check file permissions or run with appropriate privileges",
			"THEME_CACHE_CORRUPTED":        "Clear theme cache and restart application",
			"RESOURCE_LIMIT_EXCEEDED":      "Remove unused themes or increase limits",
			"API_ENDPOINT_NOT_AVAILABLE":   "Check network connection and retry",
			"EXTERNAL_SERVICE_UNAVAILABLE": "Wait for service to recover and retry",
		}

		if action, exists := actions[tpErr.Code]; exists {
			return action
		}
	}

	return "Review error details and consult documentation"
}