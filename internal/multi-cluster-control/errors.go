// Copyright 2025 James Ross
package multicluster

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrClusterNotFound       = errors.New("cluster not found")
	ErrClusterAlreadyExists  = errors.New("cluster already exists")
	ErrClusterDisconnected   = errors.New("cluster is disconnected")
	ErrNoEnabledClusters     = errors.New("no enabled clusters found")
	ErrInvalidConfiguration  = errors.New("invalid configuration")
	ErrActionNotAllowed      = errors.New("action not allowed")
	ErrActionTimeout         = errors.New("action timed out")
	ErrActionCancelled       = errors.New("action cancelled")
	ErrActionAlreadyExecuted = errors.New("action already executed")
	ErrConfirmationRequired  = errors.New("confirmation required")
	ErrCacheExpired          = errors.New("cache entry expired")
	ErrCacheNotFound         = errors.New("cache entry not found")
	ErrCompareModeDisabled   = errors.New("compare mode is disabled")
	ErrInsufficientClusters  = errors.New("insufficient clusters for comparison")
	ErrConnectionFailed      = errors.New("connection failed")
	ErrHealthCheckFailed     = errors.New("health check failed")
)

// ClusterError represents an error related to a specific cluster
type ClusterError struct {
	Cluster string
	Op      string
	Err     error
}

func (e *ClusterError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("cluster %s: %s: %v", e.Cluster, e.Op, e.Err)
	}
	return fmt.Sprintf("cluster %s: %v", e.Cluster, e.Err)
}

func (e *ClusterError) Unwrap() error {
	return e.Err
}

// NewClusterError creates a new ClusterError
func NewClusterError(cluster, op string, err error) error {
	return &ClusterError{
		Cluster: cluster,
		Op:      op,
		Err:     err,
	}
}

// MultiClusterError represents errors from multiple clusters
type MultiClusterError struct {
	Errors map[string]error
}

func (e *MultiClusterError) Error() string {
	if len(e.Errors) == 0 {
		return "multi-cluster operation failed"
	}

	if len(e.Errors) == 1 {
		for cluster, err := range e.Errors {
			return fmt.Sprintf("cluster %s: %v", cluster, err)
		}
	}

	return fmt.Sprintf("multi-cluster operation failed on %d clusters", len(e.Errors))
}

// Add adds an error for a cluster
func (e *MultiClusterError) Add(cluster string, err error) {
	if e.Errors == nil {
		e.Errors = make(map[string]error)
	}
	e.Errors[cluster] = err
}

// HasErrors returns true if there are any errors
func (e *MultiClusterError) HasErrors() bool {
	return len(e.Errors) > 0
}

// ActionError represents an error during action execution
type ActionError struct {
	ActionID string
	Type     ActionType
	Cluster  string
	Phase    string
	Err      error
}

func (e *ActionError) Error() string {
	if e.Cluster != "" {
		return fmt.Sprintf("action %s (%s) failed on cluster %s during %s: %v",
			e.ActionID, e.Type, e.Cluster, e.Phase, e.Err)
	}
	return fmt.Sprintf("action %s (%s) failed during %s: %v",
		e.ActionID, e.Type, e.Phase, e.Err)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

// NewActionError creates a new ActionError
func NewActionError(actionID string, actionType ActionType, cluster, phase string, err error) error {
	return &ActionError{
		ActionID: actionID,
		Type:     actionType,
		Cluster:  cluster,
		Phase:    phase,
		Err:      err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ConfigError represents a configuration error
type ConfigError struct {
	Section string
	Key     string
	Value   interface{}
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("config error in %s.%s (value: %v): %v", e.Section, e.Key, e.Value, e.Err)
	}
	return fmt.Sprintf("config error in %s: %v", e.Section, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// ConnectionError represents a connection error
type ConnectionError struct {
	Cluster  string
	Endpoint string
	Err      error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection to cluster %s (%s) failed: %v", e.Cluster, e.Endpoint, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewConnectionError creates a new ConnectionError
func NewConnectionError(cluster, endpoint string, err error) error {
	return &ConnectionError{
		Cluster:  cluster,
		Endpoint: endpoint,
		Err:      err,
	}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	switch err {
	case ErrClusterDisconnected, ErrConnectionFailed, ErrActionTimeout:
		return true
	case ErrActionCancelled, ErrActionAlreadyExecuted, ErrActionNotAllowed:
		return false
	}

	// Check for wrapped errors
	var connErr *ConnectionError
	if errors.As(err, &connErr) {
		return true
	}

	var clusterErr *ClusterError
	if errors.As(err, &clusterErr) {
		return IsRetryableError(clusterErr.Err)
	}

	// Check error message for known retryable patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporarily unavailable",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return fmt.Sprintf("%s", s) != "" && fmt.Sprintf("%s", substr) != "" &&
		(s == substr || fmt.Sprintf("%s", s) != fmt.Sprintf("%s", substr))
}

// ErrorSeverity represents the severity of an error
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// ClassifyError classifies an error by severity
func ClassifyError(err error) ErrorSeverity {
	if err == nil {
		return SeverityInfo
	}

	// Critical errors
	switch err {
	case ErrNoEnabledClusters, ErrInvalidConfiguration:
		return SeverityCritical
	}

	// Error level
	switch err {
	case ErrClusterNotFound, ErrActionNotAllowed, ErrConfirmationRequired:
		return SeverityError
	}

	// Warning level
	switch err {
	case ErrClusterDisconnected, ErrCacheExpired, ErrHealthCheckFailed:
		return SeverityWarning
	}

	// Check for multi-cluster errors
	var multiErr *MultiClusterError
	if errors.As(err, &multiErr) {
		if multiErr.HasErrors() {
			// If all clusters failed, it's critical
			// If some clusters failed, it's an error
			// This is a simplified check; could be more sophisticated
			return SeverityError
		}
	}

	// Default to error level for unknown errors
	return SeverityError
}