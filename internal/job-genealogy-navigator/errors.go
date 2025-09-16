// Copyright 2025 James Ross
package genealogy

import (
	"errors"
	"fmt"
)

// Common genealogy errors
var (
	ErrJobNotFound          = errors.New("job not found")
	ErrRelationshipNotFound = errors.New("relationship not found")
	ErrCyclicRelationship   = errors.New("cyclic relationship detected")
	ErrInvalidViewMode      = errors.New("invalid view mode")
	ErrInvalidLayoutMode    = errors.New("invalid layout mode")
	ErrTreeNotLoaded        = errors.New("genealogy tree not loaded")
	ErrNodeNotSelected      = errors.New("no node selected")
	ErrNavigationLimit      = errors.New("navigation limit reached")
	ErrRenderingFailed      = errors.New("tree rendering failed")
	ErrLayoutComputation    = errors.New("layout computation failed")
	ErrCacheFailure         = errors.New("cache operation failed")
	ErrStorageFailure       = errors.New("storage operation failed")
)

// GenealogyError represents genealogy-specific errors with context
type GenealogyError struct {
	Op      string // Operation that failed
	JobID   string // Job ID (if applicable)
	Message string // Human-readable message
	Err     error  // Underlying error
}

func (e *GenealogyError) Error() string {
	if e.JobID != "" {
		return fmt.Sprintf("genealogy %s failed for job %s: %s", e.Op, e.JobID, e.Message)
	}
	return fmt.Sprintf("genealogy %s failed: %s", e.Op, e.Message)
}

func (e *GenealogyError) Unwrap() error {
	return e.Err
}

// NewGenealogyError creates a new genealogy error
func NewGenealogyError(op, jobID, message string, err error) *GenealogyError {
	return &GenealogyError{
		Op:      op,
		JobID:   jobID,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents configuration or input validation errors
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s (value: %v): %s",
		e.Field, e.Value, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// RenderingError represents errors during tree rendering
type RenderingError struct {
	LayoutMode  LayoutMode
	ViewMode    ViewMode
	TreeSize    int
	Message     string
	Err         error
}

func (e *RenderingError) Error() string {
	return fmt.Sprintf("rendering failed (layout: %s, view: %s, size: %d): %s",
		e.LayoutMode, e.ViewMode, e.TreeSize, e.Message)
}

func (e *RenderingError) Unwrap() error {
	return e.Err
}

// NewRenderingError creates a new rendering error
func NewRenderingError(layoutMode LayoutMode, viewMode ViewMode, treeSize int, message string, err error) *RenderingError {
	return &RenderingError{
		LayoutMode: layoutMode,
		ViewMode:   viewMode,
		TreeSize:   treeSize,
		Message:    message,
		Err:        err,
	}
}

// NavigationError represents navigation-related errors
type NavigationError struct {
	Direction   string
	CurrentNode string
	Message     string
}

func (e *NavigationError) Error() string {
	return fmt.Sprintf("navigation failed: %s from node %s (%s)",
		e.Direction, e.CurrentNode, e.Message)
}

// NewNavigationError creates a new navigation error
func NewNavigationError(direction, currentNode, message string) *NavigationError {
	return &NavigationError{
		Direction:   direction,
		CurrentNode: currentNode,
		Message:     message,
	}
}

// StorageError represents storage backend errors
type StorageError struct {
	Operation string
	Key       string
	Message   string
	Err       error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage %s failed for key %s: %s",
		e.Operation, e.Key, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new storage error
func NewStorageError(operation, key, message string, err error) *StorageError {
	return &StorageError{
		Operation: operation,
		Key:       key,
		Message:   message,
		Err:       err,
	}
}

// Error classification functions

// IsJobNotFound returns true if the error indicates a job was not found
func IsJobNotFound(err error) bool {
	return errors.Is(err, ErrJobNotFound)
}

// IsValidationError returns true if the error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsRenderingError returns true if the error is a rendering error
func IsRenderingError(err error) bool {
	var renderingErr *RenderingError
	return errors.As(err, &renderingErr)
}

// IsNavigationError returns true if the error is a navigation error
func IsNavigationError(err error) bool {
	var navErr *NavigationError
	return errors.As(err, &navErr)
}

// IsStorageError returns true if the error is a storage error
func IsStorageError(err error) bool {
	var storageErr *StorageError
	return errors.As(err, &storageErr)
}

// IsRetryable returns true if the error might be resolved by retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Non-retryable errors
	switch {
	case errors.Is(err, ErrJobNotFound):
		return false
	case errors.Is(err, ErrInvalidViewMode):
		return false
	case errors.Is(err, ErrInvalidLayoutMode):
		return false
	case errors.Is(err, ErrCyclicRelationship):
		return false
	case IsValidationError(err):
		return false
	}

	// Navigation errors are usually not retryable
	if IsNavigationError(err) {
		return false
	}

	// Storage and rendering errors might be retryable
	return true
}

// GetErrorCode returns a code for programmatic error handling
func GetErrorCode(err error) string {
	if err == nil {
		return "OK"
	}

	switch {
	case errors.Is(err, ErrJobNotFound):
		return "JOB_NOT_FOUND"
	case errors.Is(err, ErrRelationshipNotFound):
		return "RELATIONSHIP_NOT_FOUND"
	case errors.Is(err, ErrCyclicRelationship):
		return "CYCLIC_RELATIONSHIP"
	case errors.Is(err, ErrInvalidViewMode):
		return "INVALID_VIEW_MODE"
	case errors.Is(err, ErrInvalidLayoutMode):
		return "INVALID_LAYOUT_MODE"
	case errors.Is(err, ErrTreeNotLoaded):
		return "TREE_NOT_LOADED"
	case errors.Is(err, ErrNodeNotSelected):
		return "NODE_NOT_SELECTED"
	case errors.Is(err, ErrNavigationLimit):
		return "NAVIGATION_LIMIT"
	case errors.Is(err, ErrRenderingFailed):
		return "RENDERING_FAILED"
	case errors.Is(err, ErrLayoutComputation):
		return "LAYOUT_COMPUTATION_FAILED"
	case errors.Is(err, ErrCacheFailure):
		return "CACHE_FAILURE"
	case errors.Is(err, ErrStorageFailure):
		return "STORAGE_FAILURE"
	case IsValidationError(err):
		return "VALIDATION_ERROR"
	case IsRenderingError(err):
		return "RENDERING_ERROR"
	case IsNavigationError(err):
		return "NAVIGATION_ERROR"
	case IsStorageError(err):
		return "STORAGE_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}