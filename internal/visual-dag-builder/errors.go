// Copyright 2025 James Ross
package visual_dag_builder

import "errors"

var (
	// ErrWorkflowNotFound indicates that a workflow was not found
	ErrWorkflowNotFound = errors.New("workflow not found")

	// ErrWorkflowExists indicates that a workflow already exists
	ErrWorkflowExists = errors.New("workflow already exists")

	// ErrInvalidWorkflow indicates that a workflow definition is invalid
	ErrInvalidWorkflow = errors.New("invalid workflow definition")

	// ErrNodeNotFound indicates that a node was not found
	ErrNodeNotFound = errors.New("node not found")

	// ErrEdgeNotFound indicates that an edge was not found
	ErrEdgeNotFound = errors.New("edge not found")

	// ErrCyclicDependency indicates that the DAG contains cycles
	ErrCyclicDependency = errors.New("cyclic dependency detected")

	// ErrUnreachableNode indicates that a node is unreachable
	ErrUnreachableNode = errors.New("unreachable node detected")

	// ErrMissingDependency indicates that a required dependency is missing
	ErrMissingDependency = errors.New("missing dependency")

	// ErrExecutionNotFound indicates that an execution was not found
	ErrExecutionNotFound = errors.New("execution not found")

	// ErrExecutionRunning indicates that an execution is already running
	ErrExecutionRunning = errors.New("execution already running")

	// ErrExecutionCompleted indicates that an execution has already completed
	ErrExecutionCompleted = errors.New("execution already completed")

	// ErrInvalidExecutionState indicates an invalid execution state transition
	ErrInvalidExecutionState = errors.New("invalid execution state")

	// ErrNodeExecutionFailed indicates that a node execution failed
	ErrNodeExecutionFailed = errors.New("node execution failed")

	// ErrCompensationFailed indicates that compensation failed
	ErrCompensationFailed = errors.New("compensation failed")

	// ErrTimeoutExceeded indicates that a timeout was exceeded
	ErrTimeoutExceeded = errors.New("timeout exceeded")

	// ErrMaxRetriesExceeded indicates that maximum retries were exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")

	// ErrInvalidNodeType indicates an invalid node type
	ErrInvalidNodeType = errors.New("invalid node type")

	// ErrInvalidEdgeType indicates an invalid edge type
	ErrInvalidEdgeType = errors.New("invalid edge type")

	// ErrInvalidConfiguration indicates invalid node or workflow configuration
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrStorageFailure indicates a storage operation failure
	ErrStorageFailure = errors.New("storage operation failed")

	// ErrQueueFailure indicates a queue operation failure
	ErrQueueFailure = errors.New("queue operation failed")

	// ErrCanvasBounds indicates coordinates are outside canvas bounds
	ErrCanvasBounds = errors.New("coordinates outside canvas bounds")

	// ErrInvalidPosition indicates an invalid position
	ErrInvalidPosition = errors.New("invalid position")

	// ErrParsingFailed indicates that parsing failed
	ErrParsingFailed = errors.New("parsing failed")

	// ErrSerializationFailed indicates that serialization failed
	ErrSerializationFailed = errors.New("serialization failed")

	// ErrConditionEvaluationFailed indicates that condition evaluation failed
	ErrConditionEvaluationFailed = errors.New("condition evaluation failed")

	// ErrParameterMissing indicates that a required parameter is missing
	ErrParameterMissing = errors.New("required parameter missing")

	// ErrPermissionDenied indicates that the operation is not permitted
	ErrPermissionDenied = errors.New("permission denied")
)

// ValidationErrorType represents the type of validation error
type ValidationErrorType string

const (
	CyclicDependencyError  ValidationErrorType = "cyclic_dependency"
	UnreachableNodeError   ValidationErrorType = "unreachable_node"
	MissingDependencyError ValidationErrorType = "missing_dependency"
	InvalidConfigError     ValidationErrorType = "invalid_config"
	DuplicateIDError       ValidationErrorType = "duplicate_id"
	InvalidNodeTypeError   ValidationErrorType = "invalid_node_type"
	InvalidEdgeTypeError   ValidationErrorType = "invalid_edge_type"
	MissingRequiredError   ValidationErrorType = "missing_required"
	InvalidReferenceError  ValidationErrorType = "invalid_reference"
	ConditionSyntaxError   ValidationErrorType = "condition_syntax"
)

// NewValidationError creates a new validation error
func NewValidationError(errorType ValidationErrorType, message string) ValidationError {
	return ValidationError{
		Type:    string(errorType),
		Message: message,
	}
}

// NewNodeValidationError creates a new validation error for a specific node
func NewNodeValidationError(errorType ValidationErrorType, message, nodeID string) ValidationError {
	return ValidationError{
		Type:     string(errorType),
		Message:  message,
		NodeID:   nodeID,
		Location: "node:" + nodeID,
	}
}

// NewEdgeValidationError creates a new validation error for a specific edge
func NewEdgeValidationError(errorType ValidationErrorType, message, edgeID string) ValidationError {
	return ValidationError{
		Type:     string(errorType),
		Message:  message,
		EdgeID:   edgeID,
		Location: "edge:" + edgeID,
	}
}

// ExecutionError represents an error that occurs during workflow execution
type ExecutionError struct {
	ExecutionID string            `json:"execution_id"`
	NodeID      string            `json:"node_id,omitempty"`
	Message     string            `json:"message"`
	Cause       error             `json:"cause,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   string            `json:"timestamp"`
	Recoverable bool              `json:"recoverable"`
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	if e.NodeID != "" {
		return e.ExecutionID + ":" + e.NodeID + " - " + e.Message
	}
	return e.ExecutionID + " - " + e.Message
}

// Unwrap returns the underlying error
func (e *ExecutionError) Unwrap() error {
	return e.Cause
}

// NewExecutionError creates a new execution error
func NewExecutionError(executionID, nodeID, message string, cause error) *ExecutionError {
	return &ExecutionError{
		ExecutionID: executionID,
		NodeID:      nodeID,
		Message:     message,
		Cause:       cause,
		Timestamp:   "timestamp",
		Recoverable: false,
	}
}

// NodeExecutionError represents an error specific to node execution
type NodeExecutionError struct {
	NodeID    string            `json:"node_id"`
	NodeType  NodeType          `json:"node_type"`
	JobID     string            `json:"job_id,omitempty"`
	Message   string            `json:"message"`
	Cause     error             `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp string            `json:"timestamp"`
	Attempt   int               `json:"attempt"`
}

// Error implements the error interface
func (e *NodeExecutionError) Error() string {
	return e.NodeID + " (" + string(e.NodeType) + ") - " + e.Message
}

// Unwrap returns the underlying error
func (e *NodeExecutionError) Unwrap() error {
	return e.Cause
}

// CompensationError represents an error during compensation
type CompensationError struct {
	NodeID          string `json:"node_id"`
	CompensationJobID string `json:"compensation_job_id"`
	Message         string `json:"message"`
	Cause           error  `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *CompensationError) Error() string {
	return "compensation failed for " + e.NodeID + ": " + e.Message
}

// Unwrap returns the underlying error
func (e *CompensationError) Unwrap() error {
	return e.Cause
}