// Copyright 2025 James Ross
package visual_dag_builder

import (
	"testing"
	"time"
)

// Additional tests to improve code coverage

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Storage.Type == "" {
		t.Error("Default config should have storage type")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Storage: StorageConfig{
					Type:   "redis",
					Prefix: "test:",
				},
				Execution: ExecutionConfig{
					MaxConcurrentNodes: 5,
				},
				UI: UIConfig{
					GridSize: 20,
					MaxZoom:  2.0,
					MinZoom:  0.5,
				},
				API: APIConfig{
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid storage type",
			config: Config{
				Storage: StorageConfig{
					Type: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkflowFromJSON(t *testing.T) {
	// Create a simple workflow
	workflow := &WorkflowDefinition{
		ID:      "json-test",
		Name:    "JSON Test Workflow",
		Version: "1.0",
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start"},
			{ID: "end", Type: EndNode, Name: "End"},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "end", Type: SequentialEdge},
		},
	}

	// Convert to JSON
	jsonData, err := workflow.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Test FromJSON method
	var newWorkflow WorkflowDefinition
	err = newWorkflow.FromJSON(jsonData)
	if err != nil {
		t.Errorf("FromJSON failed: %v", err)
	}

	if newWorkflow.ID != workflow.ID {
		t.Errorf("Expected ID %s, got %s", workflow.ID, newWorkflow.ID)
	}

	// Test with invalid JSON
	err = newWorkflow.FromJSON("invalid json")
	if err == nil {
		t.Error("FromJSON should fail with invalid JSON")
	}
}

func TestExecutionError(t *testing.T) {
	// Test ExecutionError
	execErr := NewExecutionError("exec-001", "node-001", "test error", nil)

	if execErr.ExecutionID != "exec-001" {
		t.Errorf("Expected execution ID exec-001, got %s", execErr.ExecutionID)
	}

	if execErr.NodeID != "node-001" {
		t.Errorf("Expected node ID node-001, got %s", execErr.NodeID)
	}

	// Test Error method
	errorMsg := execErr.Error()
	expectedMsg := "exec-001:node-001 - test error"
	if errorMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errorMsg)
	}

	// Test Unwrap method (should return nil since no cause was provided)
	if execErr.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause is provided")
	}

	// Test without node ID
	execErrNoNode := &ExecutionError{
		ExecutionID: "exec-002",
		Message:     "general error",
	}

	expectedMsgNoNode := "exec-002 - general error"
	if execErrNoNode.Error() != expectedMsgNoNode {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsgNoNode, execErrNoNode.Error())
	}
}

func TestNodeExecutionError(t *testing.T) {
	nodeErr := &NodeExecutionError{
		NodeID:   "node-001",
		NodeType: TaskNode,
		Message:  "node failed",
	}

	// Test Error method
	expectedMsg := "node-001 (task) - node failed"
	if nodeErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, nodeErr.Error())
	}

	// Test Unwrap method
	if nodeErr.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause is provided")
	}
}

func TestCompensationError(t *testing.T) {
	compErr := &CompensationError{
		NodeID:  "node-001",
		Message: "compensation failed",
	}

	// Test Error method
	expectedMsg := "compensation failed for node-001: compensation failed"
	if compErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, compErr.Error())
	}

	// Test Unwrap method
	if compErr.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause is provided")
	}
}

func TestCreateWorkflow(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := builder.CreateWorkflow("Test Workflow", "A test workflow")

	if workflow.Name != "Test Workflow" {
		t.Errorf("Expected name 'Test Workflow', got %s", workflow.Name)
	}

	if len(workflow.Nodes) != 0 {
		t.Errorf("New workflow should have no nodes, got %d", len(workflow.Nodes))
	}

	if len(workflow.Edges) != 0 {
		t.Errorf("New workflow should have no edges, got %d", len(workflow.Edges))
	}
}

func TestAddNodeToBuilder(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create a workflow first
	workflow := builder.CreateWorkflow("test", "Test")

	// Test adding a valid node
	node := Node{
		ID:   "task1",
		Type: TaskNode,
		Name: "Task 1",
		Job:  &JobConfig{Queue: "default", Type: "test"},
	}

	err := builder.AddNode(workflow, node)
	if err != nil {
		t.Errorf("AddNode failed: %v", err)
	}

	if len(workflow.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(workflow.Nodes))
	}

	// Test adding duplicate node ID
	duplicateNode := Node{
		ID:   "task1",
		Type: TaskNode,
		Name: "Duplicate Task",
		Job:  &JobConfig{Queue: "default", Type: "test"},
	}

	err = builder.AddNode(workflow, duplicateNode)
	if err == nil {
		t.Error("AddNode should fail with duplicate node ID")
	}

	// Test adding invalid node (missing job for task node)
	invalidNode := Node{
		ID:   "invalid",
		Type: TaskNode,
		Name: "Invalid Task",
		// Missing Job config
	}

	err = builder.AddNode(workflow, invalidNode)
	// Note: AddNode only checks for duplicate IDs, not node validity
	// Validation happens during ValidateDAG
	if err != nil {
		t.Logf("AddNode validation: %v", err)
	}
}

func TestAddEdgeToBuilder(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create a workflow with nodes
	workflow := builder.CreateWorkflow("test", "Test")

	// Add nodes first
	node1 := Node{ID: "start", Type: StartNode, Name: "Start"}
	node2 := Node{ID: "end", Type: EndNode, Name: "End"}

	err := builder.AddNode(workflow, node1)
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	err = builder.AddNode(workflow, node2)
	if err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}

	// Test adding valid edge
	edge := Edge{
		ID:   "e1",
		From: "start",
		To:   "end",
		Type: SequentialEdge,
	}

	err = builder.AddEdge(workflow, edge)
	if err != nil {
		t.Errorf("AddEdge failed: %v", err)
	}

	if len(workflow.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(workflow.Edges))
	}

	// Test adding edge with missing nodes
	invalidEdge := Edge{
		ID:   "e2",
		From: "missing",
		To:   "end",
		Type: SequentialEdge,
	}

	err = builder.AddEdge(workflow, invalidEdge)
	if err == nil {
		t.Error("AddEdge should fail with missing source node")
	}
}

func TestHasCycle(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create workflow without cycles
	workflow := &WorkflowDefinition{
		ID: "test",
		Nodes: []Node{
			{ID: "start", Type: StartNode},
			{ID: "end", Type: EndNode},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "end", Type: SequentialEdge},
		},
	}

	if builder.hasCycle(workflow) {
		t.Error("Workflow without cycles should not have cycles")
	}

	// Create workflow with cycle
	cyclicWorkflow := &WorkflowDefinition{
		ID: "cyclic",
		Nodes: []Node{
			{ID: "task1", Type: TaskNode, Job: &JobConfig{Queue: "default", Type: "test"}},
			{ID: "task2", Type: TaskNode, Job: &JobConfig{Queue: "default", Type: "test"}},
		},
		Edges: []Edge{
			{ID: "e1", From: "task1", To: "task2", Type: SequentialEdge},
			{ID: "e2", From: "task2", To: "task1", Type: SequentialEdge},
		},
	}

	if !builder.hasCycle(cyclicWorkflow) {
		t.Error("Workflow with cycles should detect cycles")
	}
}

func TestGenerateWorkflowID(t *testing.T) {
	id1 := generateWorkflowID()

	// Add a small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	id2 := generateWorkflowID()

	if id1 == id2 {
		t.Error("Generated workflow IDs should be unique")
	}

	if id1 == "" || id2 == "" {
		t.Error("Generated workflow IDs should not be empty")
	}
}

func TestValidateLoopNodeCoverage(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test loop node with missing configuration
	workflow := &WorkflowDefinition{
		ID: "loop-test",
		Nodes: []Node{
			{
				ID:   "loop1",
				Type: LoopNode,
				Name: "Loop without config",
				// Missing Loop config
			},
		},
	}

	result := builder.ValidateDAG(workflow)
	if result.Valid {
		t.Error("Loop node without configuration should be invalid")
	}

	// Test loop node with empty iterator
	workflow2 := &WorkflowDefinition{
		ID: "loop-test2",
		Nodes: []Node{
			{
				ID:   "loop2",
				Type: LoopNode,
				Name: "Loop with empty iterator",
				Loop: &LoopConfig{
					Iterator: "", // Empty iterator
				},
			},
		},
	}

	result2 := builder.ValidateDAG(workflow2)
	if result2.Valid {
		t.Error("Loop node with empty iterator should be invalid")
	}
}

func TestValidateParallelNodeBranches(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test parallel node with count wait_for but invalid count
	workflow := &WorkflowDefinition{
		ID: "parallel-test",
		Nodes: []Node{
			{
				ID:   "parallel1",
				Type: ParallelNode,
				Name: "Parallel with invalid count",
				Parallel: &ParallelConfig{
					WaitFor:  "count",
					Count:    0, // Invalid count
					Branches: []string{"branch1"},
				},
			},
		},
	}

	result := builder.ValidateDAG(workflow)
	if result.Valid {
		t.Error("Parallel node with invalid count should be invalid")
	}
}

func TestTopologicalSortEdgeCases(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test with cyclic workflow
	cyclicWorkflow := &WorkflowDefinition{
		ID: "cyclic",
		Nodes: []Node{
			{ID: "task1", Type: TaskNode, Job: &JobConfig{Queue: "default", Type: "test"}},
			{ID: "task2", Type: TaskNode, Job: &JobConfig{Queue: "default", Type: "test"}},
		},
		Edges: []Edge{
			{ID: "e1", From: "task1", To: "task2", Type: SequentialEdge},
			{ID: "e2", From: "task2", To: "task1", Type: SequentialEdge},
		},
	}

	_, err := builder.TopologicalSort(cyclicWorkflow)
	if err == nil {
		t.Error("TopologicalSort should fail with cyclic workflow")
	}
}

func TestValidateTaskNodeMissingFields(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test task node with missing queue
	workflow := &WorkflowDefinition{
		ID: "task-test",
		Nodes: []Node{
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Task with missing queue",
				Job: &JobConfig{
					// Missing Queue
					Type: "test_task",
				},
			},
		},
	}

	result := builder.ValidateDAG(workflow)
	if result.Valid {
		t.Error("Task node with missing queue should be invalid")
	}

	// Test task node with missing type
	workflow2 := &WorkflowDefinition{
		ID: "task-test2",
		Nodes: []Node{
			{
				ID:   "task2",
				Type: TaskNode,
				Name: "Task with missing type",
				Job: &JobConfig{
					Queue: "default",
					// Missing Type
				},
			},
		},
	}

	result2 := builder.ValidateDAG(workflow2)
	if result2.Valid {
		t.Error("Task node with missing type should be invalid")
	}
}

func TestValidateDecisionNodeMissingFields(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test decision node with condition missing expression
	workflow := &WorkflowDefinition{
		ID: "decision-test",
		Nodes: []Node{
			{
				ID:   "decision1",
				Type: DecisionNode,
				Name: "Decision with missing expression",
				Conditions: []DecisionCondition{
					{
						Expression: "", // Missing expression
						Target:     "target1",
					},
				},
			},
		},
	}

	result := builder.ValidateDAG(workflow)
	if result.Valid {
		t.Error("Decision node with missing expression should be invalid")
	}

	// Test decision node with condition missing target
	workflow2 := &WorkflowDefinition{
		ID: "decision-test2",
		Nodes: []Node{
			{
				ID:   "decision2",
				Type: DecisionNode,
				Name: "Decision with missing target",
				Conditions: []DecisionCondition{
					{
						Expression: "true",
						Target:     "", // Missing target
					},
				},
			},
		},
	}

	result2 := builder.ValidateDAG(workflow2)
	if result2.Valid {
		t.Error("Decision node with missing target should be invalid")
	}
}