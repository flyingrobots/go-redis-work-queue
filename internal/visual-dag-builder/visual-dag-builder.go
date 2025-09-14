// Copyright 2025 James Ross
package visual_dag_builder

import (
	"fmt"
	"strings"
	"time"
)

// DAGBuilder provides core DAG manipulation and validation functionality
type DAGBuilder struct {
	config Config
}

// NewDAGBuilder creates a new DAG builder
func NewDAGBuilder(config Config) *DAGBuilder {
	return &DAGBuilder{
		config: config,
	}
}

// ValidateDAG performs comprehensive validation of a workflow DAG
func (d *DAGBuilder) ValidateDAG(workflow *WorkflowDefinition) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Check for empty workflow
	if len(workflow.Nodes) == 0 {
		result.addError(NewValidationError(MissingRequiredError, "workflow must have at least one node"))
		return result
	}

	// Validate individual components
	d.validateNodes(workflow, result)
	d.validateEdges(workflow, result)
	d.validateConnectivity(workflow, result)
	d.validateCycles(workflow, result)
	d.validateReachability(workflow, result)

	result.Valid = len(result.Errors) == 0
	return result
}

// validateNodes checks all nodes for validity
func (d *DAGBuilder) validateNodes(workflow *WorkflowDefinition, result *ValidationResult) {
	nodeIDs := make(map[string]bool)

	for _, node := range workflow.Nodes {
		// Check for duplicate IDs
		if nodeIDs[node.ID] {
			result.addError(NewNodeValidationError(DuplicateIDError,
				fmt.Sprintf("duplicate node ID: %s", node.ID), node.ID))
			continue
		}
		nodeIDs[node.ID] = true

		// Validate node ID
		if node.ID == "" {
			result.addError(NewNodeValidationError(MissingRequiredError,
				"node ID is required", node.ID))
		}

		// Validate node type
		if !d.isValidNodeType(node.Type) {
			result.addError(NewNodeValidationError(InvalidNodeTypeError,
				fmt.Sprintf("invalid node type: %s", node.Type), node.ID))
		}

		// Type-specific validation
		switch node.Type {
		case TaskNode:
			d.validateTaskNode(&node, result)
		case DecisionNode:
			d.validateDecisionNode(&node, result)
		case ParallelNode:
			d.validateParallelNode(&node, result)
		case LoopNode:
			d.validateLoopNode(&node, result)
		case CompensateNode:
			d.validateCompensateNode(&node, result)
		}

		// Validate retry policy if present
		if node.Retry != nil {
			d.validateRetryPolicy(node.Retry, node.ID, result)
		}
	}
}

// validateEdges checks all edges for validity
func (d *DAGBuilder) validateEdges(workflow *WorkflowDefinition, result *ValidationResult) {
	edgeIDs := make(map[string]bool)
	nodeIDs := make(map[string]bool)

	// Build node ID map
	for _, node := range workflow.Nodes {
		nodeIDs[node.ID] = true
	}

	for _, edge := range workflow.Edges {
		// Check for duplicate edge IDs
		if edgeIDs[edge.ID] {
			result.addError(NewEdgeValidationError(DuplicateIDError,
				fmt.Sprintf("duplicate edge ID: %s", edge.ID), edge.ID))
			continue
		}
		edgeIDs[edge.ID] = true

		// Validate edge ID
		if edge.ID == "" {
			result.addError(NewEdgeValidationError(MissingRequiredError,
				"edge ID is required", edge.ID))
		}

		// Validate from/to nodes exist
		if !nodeIDs[edge.From] {
			result.addError(NewEdgeValidationError(InvalidReferenceError,
				fmt.Sprintf("edge references non-existent source node: %s", edge.From), edge.ID))
		}

		if !nodeIDs[edge.To] {
			result.addError(NewEdgeValidationError(InvalidReferenceError,
				fmt.Sprintf("edge references non-existent target node: %s", edge.To), edge.ID))
		}

		// Validate edge type
		if !d.isValidEdgeType(edge.Type) {
			result.addError(NewEdgeValidationError(InvalidEdgeTypeError,
				fmt.Sprintf("invalid edge type: %s", edge.Type), edge.ID))
		}

		// Validate conditional edges have conditions
		if edge.Type == ConditionalEdge && edge.Condition == "" {
			result.addError(NewEdgeValidationError(MissingRequiredError,
				"conditional edge requires condition expression", edge.ID))
		}
	}
}

// validateConnectivity ensures proper DAG connectivity
func (d *DAGBuilder) validateConnectivity(workflow *WorkflowDefinition, result *ValidationResult) {
	// Find start and end nodes
	startNodes := d.findStartNodes(workflow)
	endNodes := d.findEndNodes(workflow)

	if len(startNodes) == 0 {
		result.addWarning(NewValidationError(MissingRequiredError,
			"workflow has no start nodes (nodes with no incoming edges)"))
	}

	if len(endNodes) == 0 {
		result.addWarning(NewValidationError(MissingRequiredError,
			"workflow has no end nodes (nodes with no outgoing edges)"))
	}

	// Check for isolated nodes
	for _, node := range workflow.Nodes {
		incoming := workflow.GetIncomingEdges(node.ID)
		outgoing := workflow.GetOutgoingEdges(node.ID)

		if len(incoming) == 0 && len(outgoing) == 0 {
			result.addWarning(NewNodeValidationError(UnreachableNodeError,
				"node is isolated (no incoming or outgoing edges)", node.ID))
		}
	}
}

// validateCycles detects cycles in the DAG using DFS
func (d *DAGBuilder) validateCycles(workflow *WorkflowDefinition, result *ValidationResult) {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var dfs func(nodeID string, path []string) bool
	dfs = func(nodeID string, path []string) bool {
		if recursionStack[nodeID] {
			// Cycle detected
			cycleStart := -1
			for i, id := range path {
				if id == nodeID {
					cycleStart = i
					break
				}
			}
			cyclePath := append(path[cycleStart:], nodeID)
			result.addError(NewValidationError(CyclicDependencyError,
				fmt.Sprintf("cycle detected: %s", strings.Join(cyclePath, " -> "))))
			return true
		}

		if visited[nodeID] {
			return false
		}

		visited[nodeID] = true
		recursionStack[nodeID] = true
		path = append(path, nodeID)

		// Visit all adjacent nodes
		for _, edge := range workflow.GetOutgoingEdges(nodeID) {
			if dfs(edge.To, path) {
				return true
			}
		}

		recursionStack[nodeID] = false
		return false
	}

	// Check all nodes as potential cycle entry points
	for _, node := range workflow.Nodes {
		if !visited[node.ID] {
			dfs(node.ID, []string{})
		}
	}
}

// validateReachability ensures all nodes are reachable from start nodes
func (d *DAGBuilder) validateReachability(workflow *WorkflowDefinition, result *ValidationResult) {
	startNodes := d.findStartNodes(workflow)
	if len(startNodes) == 0 {
		return // No start nodes, skip reachability check
	}

	reachable := make(map[string]bool)

	// BFS from all start nodes
	var bfs func(nodeID string)
	bfs = func(nodeID string) {
		if reachable[nodeID] {
			return
		}
		reachable[nodeID] = true

		for _, edge := range workflow.GetOutgoingEdges(nodeID) {
			bfs(edge.To)
		}
	}

	for _, startNode := range startNodes {
		bfs(startNode.ID)
	}

	// Check for unreachable nodes
	for _, node := range workflow.Nodes {
		if !reachable[node.ID] {
			result.addWarning(NewNodeValidationError(UnreachableNodeError,
				"node is not reachable from start nodes", node.ID))
		}
	}
}

// TopologicalSort returns nodes in topologically sorted order
func (d *DAGBuilder) TopologicalSort(workflow *WorkflowDefinition) ([]string, error) {
	// Check for cycles first
	validation := d.ValidateDAG(workflow)
	for _, err := range validation.Errors {
		if err.Type == string(CyclicDependencyError) {
			return nil, ErrCyclicDependency
		}
	}

	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize in-degree and adjacency list
	for _, node := range workflow.Nodes {
		inDegree[node.ID] = 0
		adjList[node.ID] = []string{}
	}

	// Build adjacency list and calculate in-degrees
	for _, edge := range workflow.Edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
		inDegree[edge.To]++
	}

	// Kahn's algorithm
	var queue []string
	var result []string

	// Find nodes with no incoming edges
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		// Remove edges from this node
		for _, neighbor := range adjList[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all nodes were processed
	if len(result) != len(workflow.Nodes) {
		return nil, ErrCyclicDependency
	}

	return result, nil
}

// CreateWorkflow creates a new workflow with basic validation
func (d *DAGBuilder) CreateWorkflow(name, description string) *WorkflowDefinition {
	return &WorkflowDefinition{
		ID:          generateWorkflowID(),
		Name:        name,
		Version:     "1.0.0",
		Description: description,
		Nodes:       []Node{},
		Edges:       []Edge{},
		Config:      WorkflowConfig{
			Timeout:             30 * time.Minute,
			ConcurrencyLimit:    10,
			EnableCompensation:  true,
			EnableTracing:       true,
			FailureStrategy:     "fail_fast",
		},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        []string{},
		Metadata:    make(map[string]interface{}),
	}
}

// AddNode adds a node to the workflow with validation
func (d *DAGBuilder) AddNode(workflow *WorkflowDefinition, node Node) error {
	// Check for duplicate ID
	if workflow.GetNode(node.ID) != nil {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	// Basic node validation
	if node.ID == "" {
		return ErrMissingDependency
	}

	if !d.isValidNodeType(node.Type) {
		return ErrInvalidNodeType
	}

	workflow.AddNode(node)
	return nil
}

// AddEdge adds an edge to the workflow with validation
func (d *DAGBuilder) AddEdge(workflow *WorkflowDefinition, edge Edge) error {
	// Check for duplicate ID
	if workflow.GetEdge(edge.ID) != nil {
		return fmt.Errorf("edge with ID %s already exists", edge.ID)
	}

	// Check that nodes exist
	if workflow.GetNode(edge.From) == nil {
		return fmt.Errorf("source node %s does not exist", edge.From)
	}

	if workflow.GetNode(edge.To) == nil {
		return fmt.Errorf("target node %s does not exist", edge.To)
	}

	// Check for self-loops
	if edge.From == edge.To {
		return fmt.Errorf("self-loops are not allowed")
	}

	workflow.AddEdge(edge)

	// Check for cycles after adding edge
	if d.hasCycle(workflow) {
		// Remove the edge that created the cycle
		workflow.RemoveEdge(edge.ID)
		return ErrCyclicDependency
	}

	return nil
}

// Helper methods

func (d *DAGBuilder) isValidNodeType(nodeType NodeType) bool {
	validTypes := []NodeType{
		TaskNode, DecisionNode, ParallelNode, LoopNode,
		CompensateNode, DelayNode, StartNode, EndNode,
	}
	for _, valid := range validTypes {
		if nodeType == valid {
			return true
		}
	}
	return false
}

func (d *DAGBuilder) isValidEdgeType(edgeType EdgeType) bool {
	validTypes := []EdgeType{
		SequentialEdge, ConditionalEdge, CompensationEdge, LoopbackEdge,
	}
	for _, valid := range validTypes {
		if edgeType == valid {
			return true
		}
	}
	return false
}

func (d *DAGBuilder) validateTaskNode(node *Node, result *ValidationResult) {
	if node.Job == nil {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"task node requires job configuration", node.ID))
		return
	}

	if node.Job.Queue == "" {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"task node requires queue name", node.ID))
	}

	if node.Job.Type == "" {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"task node requires job type", node.ID))
	}
}

func (d *DAGBuilder) validateDecisionNode(node *Node, result *ValidationResult) {
	if len(node.Conditions) == 0 {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"decision node requires at least one condition", node.ID))
	}

	for i, condition := range node.Conditions {
		if condition.Expression == "" {
			result.addError(NewNodeValidationError(MissingRequiredError,
				fmt.Sprintf("condition %d requires expression", i), node.ID))
		}
		if condition.Target == "" {
			result.addError(NewNodeValidationError(MissingRequiredError,
				fmt.Sprintf("condition %d requires target node", i), node.ID))
		}
	}
}

func (d *DAGBuilder) validateParallelNode(node *Node, result *ValidationResult) {
	if node.Parallel == nil {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"parallel node requires parallel configuration", node.ID))
		return
	}

	if len(node.Parallel.Branches) == 0 {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"parallel node requires at least one branch", node.ID))
	}

	validWaitFor := []string{"all", "any", "count"}
	isValid := false
	for _, valid := range validWaitFor {
		if node.Parallel.WaitFor == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		result.addError(NewNodeValidationError(InvalidConfigError,
			"parallel node wait_for must be 'all', 'any', or 'count'", node.ID))
	}

	if node.Parallel.WaitFor == "count" && node.Parallel.Count <= 0 {
		result.addError(NewNodeValidationError(InvalidConfigError,
			"parallel node with wait_for=count requires positive count", node.ID))
	}
}

func (d *DAGBuilder) validateLoopNode(node *Node, result *ValidationResult) {
	if node.Loop == nil {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"loop node requires loop configuration", node.ID))
		return
	}

	if node.Loop.Iterator == "" {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"loop node requires iterator field", node.ID))
	}
}

func (d *DAGBuilder) validateCompensateNode(node *Node, result *ValidationResult) {
	if node.CompensationJob == nil {
		result.addError(NewNodeValidationError(MissingRequiredError,
			"compensation node requires compensation job configuration", node.ID))
	}
}

func (d *DAGBuilder) validateRetryPolicy(retry *RetryPolicy, nodeID string, result *ValidationResult) {
	validStrategies := []string{"exponential", "fixed", "linear"}
	isValid := false
	for _, valid := range validStrategies {
		if retry.Strategy == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		result.addError(NewNodeValidationError(InvalidConfigError,
			"retry strategy must be 'exponential', 'fixed', or 'linear'", nodeID))
	}

	if retry.MaxAttempts <= 0 {
		result.addError(NewNodeValidationError(InvalidConfigError,
			"retry max_attempts must be positive", nodeID))
	}

	if retry.InitialDelay <= 0 {
		result.addError(NewNodeValidationError(InvalidConfigError,
			"retry initial_delay must be positive", nodeID))
	}
}

func (d *DAGBuilder) findStartNodes(workflow *WorkflowDefinition) []Node {
	var startNodes []Node
	for _, node := range workflow.Nodes {
		incoming := workflow.GetIncomingEdges(node.ID)
		if len(incoming) == 0 {
			startNodes = append(startNodes, node)
		}
	}
	return startNodes
}

func (d *DAGBuilder) findEndNodes(workflow *WorkflowDefinition) []Node {
	var endNodes []Node
	for _, node := range workflow.Nodes {
		outgoing := workflow.GetOutgoingEdges(node.ID)
		if len(outgoing) == 0 {
			endNodes = append(endNodes, node)
		}
	}
	return endNodes
}

func (d *DAGBuilder) hasCycle(workflow *WorkflowDefinition) bool {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var dfs func(nodeID string) bool
	dfs = func(nodeID string) bool {
		if recursionStack[nodeID] {
			return true // Cycle detected
		}
		if visited[nodeID] {
			return false
		}

		visited[nodeID] = true
		recursionStack[nodeID] = true

		for _, edge := range workflow.GetOutgoingEdges(nodeID) {
			if dfs(edge.To) {
				return true
			}
		}

		recursionStack[nodeID] = false
		return false
	}

	for _, node := range workflow.Nodes {
		if !visited[node.ID] {
			if dfs(node.ID) {
				return true
			}
		}
	}
	return false
}

func generateWorkflowID() string {
	return fmt.Sprintf("workflow_%d", time.Now().UnixNano())
}

// Helper methods for ValidationResult
func (v *ValidationResult) addError(err ValidationError) {
	v.Errors = append(v.Errors, err)
	v.Valid = false
}

func (v *ValidationResult) addWarning(err ValidationError) {
	v.Warnings = append(v.Warnings, err)
}