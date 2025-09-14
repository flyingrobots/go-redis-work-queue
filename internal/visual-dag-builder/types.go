// Copyright 2025 James Ross
package visual_dag_builder

import (
	"encoding/json"
	"time"
)

// NodeType represents the type of a workflow node
type NodeType string

const (
	TaskNode        NodeType = "task"
	DecisionNode    NodeType = "decision"
	ParallelNode    NodeType = "parallel"
	LoopNode        NodeType = "loop"
	CompensateNode  NodeType = "compensate"
	DelayNode       NodeType = "delay"
	StartNode       NodeType = "start"
	EndNode         NodeType = "end"
)

// NodeStatus represents the current status of a node during execution
type NodeStatus string

const (
	NotStarted    NodeStatus = "not_started"
	Queued        NodeStatus = "queued"
	Running       NodeStatus = "running"
	Completed     NodeStatus = "completed"
	Failed        NodeStatus = "failed"
	Retrying      NodeStatus = "retrying"
	Compensating  NodeStatus = "compensating"
	Compensated   NodeStatus = "compensated"
	Skipped       NodeStatus = "skipped"
)

// EdgeType represents the type of connection between nodes
type EdgeType string

const (
	SequentialEdge   EdgeType = "sequential"
	ConditionalEdge  EdgeType = "conditional"
	CompensationEdge EdgeType = "compensation"
	LoopbackEdge     EdgeType = "loopback"
)

// Position represents a 2D coordinate on the canvas
type Position struct {
	X int `json:"x" yaml:"x"`
	Y int `json:"y" yaml:"y"`
}

// RetryPolicy defines how a node should retry on failure
type RetryPolicy struct {
	Strategy    string        `json:"strategy" yaml:"strategy"`       // exponential, fixed, linear
	MaxAttempts int           `json:"max_attempts" yaml:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay" yaml:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay" yaml:"max_delay"`
	Multiplier   float64       `json:"multiplier" yaml:"multiplier"`
	Jitter       bool          `json:"jitter" yaml:"jitter"`
}

// JobConfig defines the job to execute for a task node
type JobConfig struct {
	Queue       string                 `json:"queue" yaml:"queue"`
	Type        string                 `json:"type" yaml:"type"`
	Payload     map[string]interface{} `json:"payload" yaml:"payload"`
	Priority    string                 `json:"priority" yaml:"priority"`
	Timeout     time.Duration          `json:"timeout" yaml:"timeout"`
}

// DecisionCondition defines a condition and its target for decision nodes
type DecisionCondition struct {
	Expression string `json:"expression" yaml:"expression"`
	Target     string `json:"target" yaml:"target"`
	Label      string `json:"label" yaml:"label"`
}

// ParallelConfig defines parallel execution settings
type ParallelConfig struct {
	WaitFor     string   `json:"wait_for" yaml:"wait_for"`         // all, any, count
	Count       int      `json:"count" yaml:"count"`               // if wait_for == "count"
	Branches    []string `json:"branches" yaml:"branches"`
	ConcurrencyLimit int `json:"concurrency_limit" yaml:"concurrency_limit"`
}

// LoopConfig defines loop execution settings
type LoopConfig struct {
	Iterator      string `json:"iterator" yaml:"iterator"`           // field containing array/collection
	Parallel      bool   `json:"parallel" yaml:"parallel"`
	MaxIterations int    `json:"max_iterations" yaml:"max_iterations"`
	BreakCondition string `json:"break_condition" yaml:"break_condition"`
}

// Node represents a single node in the workflow DAG
type Node struct {
	ID           string                 `json:"id" yaml:"id"`
	Type         NodeType               `json:"type" yaml:"type"`
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	Position     Position               `json:"position" yaml:"position"`

	// Job configuration (for task nodes)
	Job          *JobConfig             `json:"job,omitempty" yaml:"job,omitempty"`

	// Retry configuration
	Retry        *RetryPolicy           `json:"retry,omitempty" yaml:"retry,omitempty"`

	// Decision configuration
	Conditions   []DecisionCondition    `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	DefaultTarget string                `json:"default_target,omitempty" yaml:"default_target,omitempty"`

	// Parallel configuration
	Parallel     *ParallelConfig        `json:"parallel,omitempty" yaml:"parallel,omitempty"`

	// Loop configuration
	Loop         *LoopConfig            `json:"loop,omitempty" yaml:"loop,omitempty"`

	// Compensation configuration
	CompensationJob *JobConfig          `json:"compensation_job,omitempty" yaml:"compensation_job,omitempty"`

	// Runtime metadata
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Edge represents a connection between two nodes
type Edge struct {
	ID        string                 `json:"id" yaml:"id"`
	From      string                 `json:"from" yaml:"from"`
	To        string                 `json:"to" yaml:"to"`
	Type      EdgeType               `json:"type" yaml:"type"`
	Condition string                 `json:"condition,omitempty" yaml:"condition,omitempty"`
	Label     string                 `json:"label,omitempty" yaml:"label,omitempty"`
	Priority  int                    `json:"priority,omitempty" yaml:"priority,omitempty"`
	Delay     time.Duration          `json:"delay,omitempty" yaml:"delay,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// WorkflowDefinition represents a complete workflow DAG
type WorkflowDefinition struct {
	ID          string                 `json:"id" yaml:"id"`
	Name        string                 `json:"name" yaml:"name"`
	Version     string                 `json:"version" yaml:"version"`
	Description string                 `json:"description" yaml:"description"`

	// DAG structure
	Nodes       []Node                 `json:"nodes" yaml:"nodes"`
	Edges       []Edge                 `json:"edges" yaml:"edges"`

	// Global configuration
	Config      WorkflowConfig         `json:"config" yaml:"config"`

	// Metadata
	CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`
	CreatedBy   string                 `json:"created_by" yaml:"created_by"`
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
	UpdatedBy   string                 `json:"updated_by" yaml:"updated_by"`
	Tags        []string               `json:"tags" yaml:"tags"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// WorkflowConfig defines global workflow settings
type WorkflowConfig struct {
	Timeout             time.Duration `json:"timeout" yaml:"timeout"`
	ConcurrencyLimit    int           `json:"concurrency_limit" yaml:"concurrency_limit"`
	RetryPolicy         *RetryPolicy  `json:"retry_policy,omitempty" yaml:"retry_policy,omitempty"`
	EnableCompensation  bool          `json:"enable_compensation" yaml:"enable_compensation"`
	EnableTracing       bool          `json:"enable_tracing" yaml:"enable_tracing"`
	FailureStrategy     string        `json:"failure_strategy" yaml:"failure_strategy"` // fail_fast, continue, compensate
}

// WorkflowExecution represents a running instance of a workflow
type WorkflowExecution struct {
	ID           string                 `json:"id" yaml:"id"`
	WorkflowID   string                 `json:"workflow_id" yaml:"workflow_id"`
	WorkflowVersion string              `json:"workflow_version" yaml:"workflow_version"`
	Status       ExecutionStatus        `json:"status" yaml:"status"`

	// Execution parameters
	Parameters   map[string]interface{} `json:"parameters" yaml:"parameters"`
	Context      map[string]interface{} `json:"context" yaml:"context"`

	// Node states
	NodeStates   map[string]*NodeState  `json:"node_states" yaml:"node_states"`

	// Timing information
	StartedAt    time.Time              `json:"started_at" yaml:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	Duration     time.Duration          `json:"duration" yaml:"duration"`

	// Tracing and observability
	TraceID      string                 `json:"trace_id" yaml:"trace_id"`
	SpanID       string                 `json:"span_id" yaml:"span_id"`

	// Results and errors
	Result       interface{}            `json:"result,omitempty" yaml:"result,omitempty"`
	Error        string                 `json:"error,omitempty" yaml:"error,omitempty"`

	// Metadata
	CreatedBy    string                 `json:"created_by" yaml:"created_by"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ExecutionStatus represents the status of a workflow execution
type ExecutionStatus string

const (
	ExecutionPending    ExecutionStatus = "pending"
	ExecutionRunning    ExecutionStatus = "running"
	ExecutionCompleted  ExecutionStatus = "completed"
	ExecutionFailed     ExecutionStatus = "failed"
	ExecutionCancelled  ExecutionStatus = "cancelled"
	ExecutionPaused     ExecutionStatus = "paused"
	ExecutionCompensating ExecutionStatus = "compensating"
)

// NodeState represents the runtime state of a node during execution
type NodeState struct {
	NodeID       string                 `json:"node_id" yaml:"node_id"`
	Status       NodeStatus             `json:"status" yaml:"status"`
	Attempts     int                    `json:"attempts" yaml:"attempts"`

	// Timing
	StartedAt    *time.Time             `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	Duration     time.Duration          `json:"duration" yaml:"duration"`

	// Job information
	JobID        string                 `json:"job_id,omitempty" yaml:"job_id,omitempty"`
	QueueName    string                 `json:"queue_name,omitempty" yaml:"queue_name,omitempty"`

	// Results and errors
	Input        map[string]interface{} `json:"input,omitempty" yaml:"input,omitempty"`
	Output       map[string]interface{} `json:"output,omitempty" yaml:"output,omitempty"`
	Error        string                 `json:"error,omitempty" yaml:"error,omitempty"`

	// Retry information
	NextRetryAt  *time.Time             `json:"next_retry_at,omitempty" yaml:"next_retry_at,omitempty"`

	// Compensation information
	CompensationJobID string            `json:"compensation_job_id,omitempty" yaml:"compensation_job_id,omitempty"`

	// Metadata
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ValidationError represents a DAG validation error
type ValidationError struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	NodeID   string `json:"node_id,omitempty"`
	EdgeID   string `json:"edge_id,omitempty"`
	Location string `json:"location,omitempty"`
}

// ValidationResult contains the results of DAG validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
}

// CanvasState represents the current state of the DAG editor canvas
type CanvasState struct {
	Workflow     *WorkflowDefinition `json:"workflow"`
	SelectedNode string              `json:"selected_node"`
	SelectedEdge string              `json:"selected_edge"`
	ViewOffset   Position            `json:"view_offset"`
	ZoomLevel    float64             `json:"zoom_level"`
	GridSize     int                 `json:"grid_size"`
	ShowGrid     bool                `json:"show_grid"`
	Mode         CanvasMode          `json:"mode"`
}

// CanvasMode represents the current editing mode
type CanvasMode string

const (
	SelectMode    CanvasMode = "select"
	AddNodeMode   CanvasMode = "add_node"
	ConnectMode   CanvasMode = "connect"
	PanMode       CanvasMode = "pan"
)

// ExecutionEvent represents an event during workflow execution
type ExecutionEvent struct {
	ID           string                 `json:"id"`
	ExecutionID  string                 `json:"execution_id"`
	NodeID       string                 `json:"node_id"`
	EventType    string                 `json:"event_type"`
	Status       NodeStatus             `json:"status"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
}

// Helper methods

// ToJSON converts any struct to JSON string
func (w *WorkflowDefinition) ToJSON() (string, error) {
	data, err := json.MarshalIndent(w, "", "  ")
	return string(data), err
}

// FromJSON loads a workflow from JSON string
func (w *WorkflowDefinition) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), w)
}

// GetNode returns a node by ID
func (w *WorkflowDefinition) GetNode(id string) *Node {
	for i := range w.Nodes {
		if w.Nodes[i].ID == id {
			return &w.Nodes[i]
		}
	}
	return nil
}

// GetEdge returns an edge by ID
func (w *WorkflowDefinition) GetEdge(id string) *Edge {
	for i := range w.Edges {
		if w.Edges[i].ID == id {
			return &w.Edges[i]
		}
	}
	return nil
}

// GetIncomingEdges returns all edges that target the given node
func (w *WorkflowDefinition) GetIncomingEdges(nodeID string) []Edge {
	var edges []Edge
	for _, edge := range w.Edges {
		if edge.To == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetOutgoingEdges returns all edges that originate from the given node
func (w *WorkflowDefinition) GetOutgoingEdges(nodeID string) []Edge {
	var edges []Edge
	for _, edge := range w.Edges {
		if edge.From == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// AddNode adds a new node to the workflow
func (w *WorkflowDefinition) AddNode(node Node) {
	w.Nodes = append(w.Nodes, node)
	w.UpdatedAt = time.Now()
}

// AddEdge adds a new edge to the workflow
func (w *WorkflowDefinition) AddEdge(edge Edge) {
	w.Edges = append(w.Edges, edge)
	w.UpdatedAt = time.Now()
}

// RemoveNode removes a node and all connected edges
func (w *WorkflowDefinition) RemoveNode(nodeID string) {
	// Remove the node
	for i, node := range w.Nodes {
		if node.ID == nodeID {
			w.Nodes = append(w.Nodes[:i], w.Nodes[i+1:]...)
			break
		}
	}

	// Remove all connected edges
	var remainingEdges []Edge
	for _, edge := range w.Edges {
		if edge.From != nodeID && edge.To != nodeID {
			remainingEdges = append(remainingEdges, edge)
		}
	}
	w.Edges = remainingEdges
	w.UpdatedAt = time.Now()
}

// RemoveEdge removes an edge from the workflow
func (w *WorkflowDefinition) RemoveEdge(edgeID string) {
	for i, edge := range w.Edges {
		if edge.ID == edgeID {
			w.Edges = append(w.Edges[:i], w.Edges[i+1:]...)
			break
		}
	}
	w.UpdatedAt = time.Now()
}