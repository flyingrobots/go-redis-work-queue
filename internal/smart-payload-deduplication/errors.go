// Copyright 2025 James Ross
package deduplication

import (
	"errors"
	"fmt"
	"sync"
)

// Predefined error instances for common cases
var (
	// Configuration errors
	ErrInvalidConfig        = errors.New("invalid deduplication configuration")
	ErrMissingRedisAddr     = errors.New("redis address is required")
	ErrInvalidChunkSize     = errors.New("invalid chunk size configuration")
	ErrInvalidThreshold     = errors.New("invalid similarity threshold")

	// Storage errors
	ErrChunkNotFound        = errors.New("chunk not found in storage")
	ErrPayloadNotFound      = errors.New("payload map not found")
	ErrStorageFull          = errors.New("storage capacity exceeded")
	ErrStorageUnavailable   = errors.New("storage backend unavailable")
	ErrChecksumMismatch     = errors.New("payload checksum verification failed")

	// Compression errors
	ErrCompressionFailed    = errors.New("data compression failed")
	ErrDecompressionFailed  = errors.New("data decompression failed")
	ErrDictionaryBuildFailed = errors.New("compression dictionary build failed")

	// Reference counting errors
	ErrReferenceCorruption  = errors.New("reference count corruption detected")
	ErrReferenceOverflow    = errors.New("reference count overflow")
	ErrOrphanedChunk        = errors.New("orphaned chunk detected")

	// Garbage collection errors
	ErrGCFailed             = errors.New("garbage collection failed")
	ErrGCInProgress         = errors.New("garbage collection already in progress")

	// Integrity errors
	ErrIntegrityViolation   = errors.New("data integrity violation")
	ErrDataCorruption       = errors.New("data corruption detected")
)

// DeduplicationErrorType represents different categories of errors
type DeduplicationErrorType string

const (
	ErrorTypeConfiguration  DeduplicationErrorType = "configuration"
	ErrorTypeStorage        DeduplicationErrorType = "storage"
	ErrorTypeCompression    DeduplicationErrorType = "compression"
	ErrorTypeReference      DeduplicationErrorType = "reference"
	ErrorTypeGarbageCollection DeduplicationErrorType = "garbage_collection"
	ErrorTypeIntegrity      DeduplicationErrorType = "integrity"
	ErrorTypeTimeout        DeduplicationErrorType = "timeout"
	ErrorTypeNetwork        DeduplicationErrorType = "network"
)

// ExtendedDeduplicationError provides detailed error information
type ExtendedDeduplicationError struct {
	Type        DeduplicationErrorType `json:"type"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Cause       error                  `json:"cause,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	Recoverable bool                   `json:"recoverable"`
	RetryAfter  int64                  `json:"retry_after,omitempty"` // Seconds
}

// Error implements the error interface
func (e *ExtendedDeduplicationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Type, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ExtendedDeduplicationError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for errors.Is()
func (e *ExtendedDeduplicationError) Is(target error) bool {
	if t, ok := target.(*ExtendedDeduplicationError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// NewExtendedError creates a new extended deduplication error
func NewExtendedError(errorType DeduplicationErrorType, code, message string) *ExtendedDeduplicationError {
	return &ExtendedDeduplicationError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		Context:     make(map[string]interface{}),
		Timestamp:   getCurrentTimestamp(),
		Recoverable: isRecoverableError(errorType, code),
	}
}

// WithCause adds a cause to the error
func (e *ExtendedDeduplicationError) WithCause(cause error) *ExtendedDeduplicationError {
	e.Cause = cause
	return e
}

// WithContext adds context information to the error
func (e *ExtendedDeduplicationError) WithContext(key string, value interface{}) *ExtendedDeduplicationError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRetryAfter sets the retry delay for recoverable errors
func (e *ExtendedDeduplicationError) WithRetryAfter(seconds int64) *ExtendedDeduplicationError {
	e.RetryAfter = seconds
	return e
}

// Error factory functions for common error patterns

// NewConfigurationError creates a configuration-related error
func NewConfigurationError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeConfiguration, code, message)
}

// NewStorageError creates a storage-related error
func NewStorageError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeStorage, code, message)
}

// NewCompressionError creates a compression-related error
func NewCompressionError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeCompression, code, message)
}

// NewReferenceError creates a reference counting error
func NewReferenceError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeReference, code, message)
}

// NewGCError creates a garbage collection error
func NewGCError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeGarbageCollection, code, message)
}

// NewIntegrityError creates an integrity violation error
func NewIntegrityError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeIntegrity, code, message)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeTimeout, code, message).WithRetryAfter(5)
}

// NewNetworkError creates a network-related error
func NewNetworkError(code, message string) *ExtendedDeduplicationError {
	return NewExtendedError(ErrorTypeNetwork, code, message).WithRetryAfter(10)
}

// Error classification helpers

// IsRecoverable returns true if the error can be recovered from with retry
func IsRecoverable(err error) bool {
	if dedupErr, ok := err.(*ExtendedDeduplicationError); ok {
		return dedupErr.Recoverable
	}

	if dedupErr, ok := err.(*DeduplicationError); ok {
		return isRecoverableErrorCode(dedupErr.Code)
	}

	return false
}

// IsTemporary returns true if the error is likely temporary
func IsTemporary(err error) bool {
	if dedupErr, ok := err.(*ExtendedDeduplicationError); ok {
		return dedupErr.Type == ErrorTypeTimeout ||
			dedupErr.Type == ErrorTypeNetwork ||
			(dedupErr.Type == ErrorTypeStorage && dedupErr.Code != ErrCodeDataCorruption)
	}

	return false
}

// GetRetryDelay returns the recommended retry delay in seconds
func GetRetryDelay(err error) int64 {
	if dedupErr, ok := err.(*ExtendedDeduplicationError); ok {
		if dedupErr.RetryAfter > 0 {
			return dedupErr.RetryAfter
		}
	}

	// Default retry delays based on error type
	if IsTemporary(err) {
		return 5 // 5 seconds for temporary errors
	}

	return 0 // No retry for non-recoverable errors
}

// Error wrapping utilities

// WrapStorageError wraps a storage error with additional context
func WrapStorageError(err error, operation string) error {
	if err == nil {
		return nil
	}

	return NewStorageError("STORAGE_OPERATION_FAILED",
		fmt.Sprintf("storage operation '%s' failed", operation)).
		WithCause(err).
		WithContext("operation", operation)
}

// WrapCompressionError wraps a compression error with additional context
func WrapCompressionError(err error, operation string, size int) error {
	if err == nil {
		return nil
	}

	return NewCompressionError("COMPRESSION_OPERATION_FAILED",
		fmt.Sprintf("compression operation '%s' failed", operation)).
		WithCause(err).
		WithContext("operation", operation).
		WithContext("data_size", size)
}

// WrapReferenceError wraps a reference counting error with chunk context
func WrapReferenceError(err error, chunkHash string, operation string) error {
	if err == nil {
		return nil
	}

	return NewReferenceError("REFERENCE_OPERATION_FAILED",
		fmt.Sprintf("reference operation '%s' failed for chunk", operation)).
		WithCause(err).
		WithContext("operation", operation).
		WithContext("chunk_hash", chunkHash)
}

// Error aggregation for batch operations

// ErrorAggregate collects multiple errors from batch operations
type ErrorAggregate struct {
	Errors      []error                `json:"errors"`
	Summary     string                 `json:"summary"`
	TotalErrors int                    `json:"total_errors"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (ea *ErrorAggregate) Error() string {
	if ea.Summary != "" {
		return fmt.Sprintf("batch operation failed: %s (%d errors)", ea.Summary, ea.TotalErrors)
	}
	return fmt.Sprintf("batch operation failed with %d errors", ea.TotalErrors)
}

// Add adds an error to the aggregate
func (ea *ErrorAggregate) Add(err error) {
	if err != nil {
		ea.Errors = append(ea.Errors, err)
		ea.TotalErrors = len(ea.Errors)
	}
}

// HasErrors returns true if the aggregate contains any errors
func (ea *ErrorAggregate) HasErrors() bool {
	return ea.TotalErrors > 0
}

// GetByType returns all errors of a specific type
func (ea *ErrorAggregate) GetByType(errorType DeduplicationErrorType) []*ExtendedDeduplicationError {
	var result []*ExtendedDeduplicationError

	for _, err := range ea.Errors {
		if dedupErr, ok := err.(*ExtendedDeduplicationError); ok {
			if dedupErr.Type == errorType {
				result = append(result, dedupErr)
			}
		}
	}

	return result
}

// NewErrorAggregate creates a new error aggregate
func NewErrorAggregate(summary string) *ErrorAggregate {
	return &ErrorAggregate{
		Errors:  make([]error, 0),
		Summary: summary,
		Context: make(map[string]interface{}),
	}
}

// Helper functions

func getCurrentTimestamp() int64 {
	return timeNow().Unix()
}

func isRecoverableError(errorType DeduplicationErrorType, code string) bool {
	switch errorType {
	case ErrorTypeTimeout, ErrorTypeNetwork:
		return true
	case ErrorTypeStorage:
		return code != ErrCodeDataCorruption && code != ErrCodeChecksumMismatch
	case ErrorTypeReference:
		return code != ErrCodeReferenceCorruption
	case ErrorTypeCompression:
		return code != ErrCodeDecompressionFailed
	default:
		return false
	}
}

func isRecoverableErrorCode(code string) bool {
	recoverableCodes := map[string]bool{
		ErrCodeStorageFull:     true,
		ErrCodeChunkNotFound:   false,
		ErrCodePayloadNotFound: false,
		ErrCodeCompressionFailed: true,
		ErrCodeDecompressionFailed: false,
		ErrCodeReferenceCorruption: false,
		ErrCodeGCFailed:        true,
	}

	recoverable, exists := recoverableCodes[code]
	return exists && recoverable
}

// Mock time function for testing
var timeNow = func() mockTime {
	return mockTime{}
}

type mockTime struct{}

func (t mockTime) Unix() int64 {
	return 1642680000 // Mock timestamp
}

// Error code constants for backward compatibility
const (
	ErrCodeChunkNotFound       = "CHUNK_NOT_FOUND"
	ErrCodePayloadNotFound     = "PAYLOAD_NOT_FOUND"
	ErrCodeChecksumMismatch    = "CHECKSUM_MISMATCH"
	ErrCodeCompressionFailed   = "COMPRESSION_FAILED"
	ErrCodeDecompressionFailed = "DECOMPRESSION_FAILED"
	ErrCodeInvalidConfig       = "INVALID_CONFIG"
	ErrCodeStorageFull         = "STORAGE_FULL"
	ErrCodeReferenceCorruption = "REFERENCE_CORRUPTION"
	ErrCodeGCFailed            = "GC_FAILED"
	ErrCodeDataCorruption      = "DATA_CORRUPTION"
)

// Error monitoring and reporting

// ErrorReporter provides error reporting capabilities
type ErrorReporter struct {
	errorCounts map[string]int64
	mu          sync.RWMutex
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter() *ErrorReporter {
	return &ErrorReporter{
		errorCounts: make(map[string]int64),
	}
}

// Report records an error occurrence
func (er *ErrorReporter) Report(err error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if dedupErr, ok := err.(*ExtendedDeduplicationError); ok {
		key := fmt.Sprintf("%s:%s", dedupErr.Type, dedupErr.Code)
		er.errorCounts[key]++
	} else {
		er.errorCounts["unknown"]++
	}
}

// GetErrorCounts returns error counts by type and code
func (er *ErrorReporter) GetErrorCounts() map[string]int64 {
	er.mu.RLock()
	defer er.mu.RUnlock()

	counts := make(map[string]int64, len(er.errorCounts))
	for k, v := range er.errorCounts {
		counts[k] = v
	}

	return counts
}

// Reset clears all error counts
func (er *ErrorReporter) Reset() {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.errorCounts = make(map[string]int64)
}