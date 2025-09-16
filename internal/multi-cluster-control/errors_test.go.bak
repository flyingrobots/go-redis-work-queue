package multicluster

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiClusterError(t *testing.T) {
	err := &MultiClusterError{
		Code:    "CLUSTER_NOT_FOUND",
		Message: "Cluster 'test' not found",
		Details: "The specified cluster does not exist in the configuration",
		Cause:   errors.New("original error"),
	}

	assert.Equal(t, "CLUSTER_NOT_FOUND", err.GetCode())
	assert.Equal(t, "Cluster 'test' not found", err.GetMessage())
	assert.Equal(t, "The specified cluster does not exist in the configuration", err.GetDetails())
	assert.Equal(t, "original error", err.Unwrap().Error())
	assert.Contains(t, err.Error(), "Cluster 'test' not found")
	assert.Contains(t, err.Error(), "The specified cluster does not exist in the configuration")
}

func TestNewMultiClusterError(t *testing.T) {
	err := NewMultiClusterError("TEST_CODE", "Test message", "Test details")

	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, "Test message", err.Message)
	assert.Equal(t, "Test details", err.Details)
	assert.Nil(t, err.Cause)
}

func TestWrapMultiClusterError(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapMultiClusterError("WRAP_CODE", "Wrapped message", originalErr)

	assert.Equal(t, "WRAP_CODE", err.Code)
	assert.Equal(t, "Wrapped message", err.Message)
	assert.Equal(t, originalErr, err.Cause)
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestConnectionError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := NewConnectionError("test-cluster", "localhost:6379", originalErr)

	assert.Equal(t, "CONNECTION_FAILED", err.Code)
	assert.Contains(t, err.Message, "test-cluster")
	assert.Contains(t, err.Message, "localhost:6379")
	assert.Equal(t, originalErr, err.Cause)
}

func TestClusterError(t *testing.T) {
	originalErr := errors.New("operation failed")
	err := NewClusterError("test-cluster", "stats", originalErr)

	assert.Equal(t, "CLUSTER_OPERATION_FAILED", err.Code)
	assert.Contains(t, err.Message, "test-cluster")
	assert.Contains(t, err.Message, "stats")
	assert.Equal(t, originalErr, err.Cause)
}

func TestActionError(t *testing.T) {
	originalErr := errors.New("validation failed")
	err := NewActionError("action-123", ActionTypePurgeDLQ, "test-cluster", "validation", originalErr)

	assert.Equal(t, "ACTION_FAILED", err.Code)
	assert.Contains(t, err.Message, "action-123")
	assert.Contains(t, err.Message, "purge_dlq")
	assert.Contains(t, err.Message, "test-cluster")
	assert.Contains(t, err.Message, "validation")
	assert.Equal(t, originalErr, err.Cause)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{"ErrNoEnabledClusters", ErrNoEnabledClusters, "NO_ENABLED_CLUSTERS"},
		{"ErrInvalidConfiguration", ErrInvalidConfiguration, "INVALID_CONFIGURATION"},
		{"ErrClusterNotFound", ErrClusterNotFound, "CLUSTER_NOT_FOUND"},
		{"ErrClusterAlreadyExists", ErrClusterAlreadyExists, "CLUSTER_ALREADY_EXISTS"},
		{"ErrClusterDisconnected", ErrClusterDisconnected, "CLUSTER_DISCONNECTED"},
		{"ErrInsufficientClusters", ErrInsufficientClusters, "INSUFFICIENT_CLUSTERS"},
		{"ErrActionNotAllowed", ErrActionNotAllowed, "ACTION_NOT_ALLOWED"},
		{"ErrConfirmationRequired", ErrConfirmationRequired, "CONFIRMATION_REQUIRED"},
		{"ErrCacheExpired", ErrCacheExpired, "CACHE_EXPIRED"},
		{"ErrOperationTimeout", ErrOperationTimeout, "OPERATION_TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mcErr, ok := tt.err.(*MultiClusterError); ok {
				assert.Equal(t, tt.code, mcErr.Code)
				assert.NotEmpty(t, mcErr.Message)
			} else {
				t.Errorf("Expected *MultiClusterError, got %T", tt.err)
			}
		})
	}
}

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		severity ErrorSeverity
	}{
		{"nil error", nil, SeverityInfo},
		{"no enabled clusters", ErrNoEnabledClusters, SeverityCritical},
		{"invalid configuration", ErrInvalidConfiguration, SeverityCritical},
		{"cluster not found", ErrClusterNotFound, SeverityError},
		{"action not allowed", ErrActionNotAllowed, SeverityError},
		{"cluster disconnected", ErrClusterDisconnected, SeverityWarning},
		{"cache expired", ErrCacheExpired, SeverityWarning},
		{"unknown error", errors.New("unknown"), SeverityError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := ClassifyError(tt.err)
			assert.Equal(t, tt.severity, severity)
		})
	}
}

func TestErrorSeverity(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.severity))
		})
	}
}

func TestIsMultiClusterError(t *testing.T) {
	mcErr := NewMultiClusterError("TEST", "Test error", "")
	regularErr := errors.New("regular error")

	assert.True(t, IsMultiClusterError(mcErr))
	assert.False(t, IsMultiClusterError(regularErr))
	assert.False(t, IsMultiClusterError(nil))
}

func TestGetErrorCode(t *testing.T) {
	mcErr := NewMultiClusterError("TEST_CODE", "Test error", "")
	regularErr := errors.New("regular error")

	assert.Equal(t, "TEST_CODE", GetErrorCode(mcErr))
	assert.Equal(t, "UNKNOWN", GetErrorCode(regularErr))
	assert.Equal(t, "UNKNOWN", GetErrorCode(nil))
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"connection error", NewConnectionError("test", "localhost", errors.New("refused")), true},
		{"operation timeout", ErrOperationTimeout, true},
		{"cache expired", ErrCacheExpired, true},
		{"cluster not found", ErrClusterNotFound, false},
		{"action not allowed", ErrActionNotAllowed, false},
		{"invalid configuration", ErrInvalidConfiguration, false},
		{"regular error", errors.New("regular"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retryable, IsRetryableError(tt.err))
		})
	}
}

func TestIsCriticalError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		critical bool
	}{
		{"no enabled clusters", ErrNoEnabledClusters, true},
		{"invalid configuration", ErrInvalidConfiguration, true},
		{"cluster not found", ErrClusterNotFound, false},
		{"cluster disconnected", ErrClusterDisconnected, false},
		{"regular error", errors.New("regular"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.critical, IsCriticalError(tt.err))
		})
	}
}

func TestErrorChaining(t *testing.T) {
	rootErr := errors.New("root cause")
	wrappedErr := WrapMultiClusterError("WRAPPED", "Wrapped error", rootErr)
	doubleWrapped := WrapMultiClusterError("DOUBLE", "Double wrapped", wrappedErr)

	// Test error unwrapping
	assert.Equal(t, wrappedErr, doubleWrapped.Unwrap())
	assert.Equal(t, rootErr, wrappedErr.Unwrap())

	// Test error.Is
	assert.True(t, errors.Is(doubleWrapped, wrappedErr))
	assert.True(t, errors.Is(doubleWrapped, rootErr))
	assert.False(t, errors.Is(doubleWrapped, errors.New("different")))
}

func TestErrorContext(t *testing.T) {
	err := NewMultiClusterError("TEST", "Test error", "")

	contextErr := err.WithContext("cluster", "test-cluster").
		WithContext("operation", "stats").
		WithContext("user", "admin")

	assert.Contains(t, contextErr.Error(), "cluster=test-cluster")
	assert.Contains(t, contextErr.Error(), "operation=stats")
	assert.Contains(t, contextErr.Error(), "user=admin")
}

func TestErrorLogging(t *testing.T) {
	err := NewMultiClusterError("TEST", "Test error", "Test details")

	logFields := err.GetLogFields()
	assert.Equal(t, "TEST", logFields["error_code"])
	assert.Equal(t, "Test error", logFields["error_message"])
	assert.Equal(t, "Test details", logFields["error_details"])

	// Test with cause
	errWithCause := WrapMultiClusterError("WRAPPED", "Wrapped", errors.New("cause"))
	logFields = errWithCause.GetLogFields()
	assert.Equal(t, "cause", logFields["error_cause"])
}

func TestValidationError(t *testing.T) {
	validation := NewValidationError()
	validation.AddError("cluster_name", "cannot be empty")
	validation.AddError("endpoint", "invalid format")

	assert.True(t, validation.HasErrors())
	assert.Len(t, validation.GetErrors(), 2)

	err := validation.ToError()
	assert.Contains(t, err.Error(), "cluster_name")
	assert.Contains(t, err.Error(), "endpoint")

	// Test empty validation
	emptyValidation := NewValidationError()
	assert.False(t, emptyValidation.HasErrors())
	assert.Nil(t, emptyValidation.ToError())
}

func TestErrorRecovery(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		canRecover bool
	}{
		{"cache expired", ErrCacheExpired, true},
		{"operation timeout", ErrOperationTimeout, true},
		{"cluster disconnected", ErrClusterDisconnected, true},
		{"cluster not found", ErrClusterNotFound, false},
		{"invalid configuration", ErrInvalidConfiguration, false},
		{"action not allowed", ErrActionNotAllowed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canRecover, CanRecoverFromError(tt.err))
		})
	}
}

func TestErrorMetrics(t *testing.T) {
	err := NewMultiClusterError("TEST", "Test error", "")

	metrics := err.GetMetrics()
	assert.Equal(t, "TEST", metrics["code"])
	assert.Equal(t, "error", metrics["severity"])
	assert.Equal(t, "false", metrics["retryable"])

	// Test with retryable error
	retryableErr := NewConnectionError("test", "localhost", errors.New("refused"))
	metrics = retryableErr.GetMetrics()
	assert.Equal(t, "true", metrics["retryable"])
	assert.Equal(t, "warning", metrics["severity"])
}