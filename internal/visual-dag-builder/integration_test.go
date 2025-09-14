// Copyright 2025 James Ross
package visual_dag_builder

import (
	"encoding/json"
	"testing"
	"time"
)

// Integration tests that test multiple components working together

func TestWorkflowSerialization_Integration(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create a complex workflow
	workflow := &WorkflowDefinition{
		ID:          "complex-workflow",
		Name:        "Complex Integration Workflow",
		Version:     "1.0",
		Description: "Tests serialization of complex workflow",
		Config: WorkflowConfig{
			Timeout:             30 * time.Minute,
			ConcurrencyLimit:    5,
			EnableCompensation:  true,
			EnableTracing:       true,
			FailureStrategy:     "compensate",
		},
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
				Tags: []string{"entry"},
			},
			{
				ID:   "decision",
				Type: DecisionNode,
				Name: "Route Decision",
				Position: Position{X: 100, Y: 0},
				Conditions: []DecisionCondition{
					{Expression: "data.priority == 'high'", Target: "parallel_high", Label: "High Priority"},
					{Expression: "data.priority == 'low'", Target: "delay", Label: "Low Priority"},
				},
				DefaultTarget: "task_normal",
			},
			{
				ID:   "parallel_high",
				Type: ParallelNode,
				Name: "High Priority Processing",
				Position: Position{X: 200, Y: -50},
				Parallel: &ParallelConfig{
					WaitFor:          "all",
					Branches:         []string{"task_urgent", "task_validate"},
					ConcurrencyLimit: 2,
				},
			},
			{
				ID:   "task_urgent",
				Type: TaskNode,
				Name: "Urgent Task",
				Position: Position{X: 300, Y: -80},
				Job: &JobConfig{
					Queue:    "urgent",
					Type:     "process_urgent",
					Priority: "high",
					Timeout:  2 * time.Minute,
					Payload: map[string]interface{}{
						"priority_level": 5,
						"escalation":     true,
					},
				},
				Retry: &RetryPolicy{
					Strategy:     "exponential",
					MaxAttempts:  3,
					InitialDelay: 500 * time.Millisecond,
					MaxDelay:     10 * time.Second,
					Multiplier:   2.0,
					Jitter:       true,
				},
				CompensationJob: &JobConfig{
					Queue: "compensation",
					Type:  "rollback_urgent",
				},
				Tags: []string{"urgent", "high-priority"},
			},
			{
				ID:   "task_validate",
				Type: TaskNode,
				Name: "Validation Task",
				Position: Position{X: 300, Y: -20},
				Job: &JobConfig{
					Queue:    "validation",
					Type:     "validate_data",
					Priority: "normal",
					Timeout:  1 * time.Minute,
				},
			},
			{
				ID:   "delay",
				Type: DelayNode,
				Name: "Delay Processing",
				Position: Position{X: 200, Y: 50},
				DelayConfig: &DelayConfig{
					Duration:   5 * time.Minute,
					Dynamic:    false,
					Expression: "",
				},
			},
			{
				ID:   "task_normal",
				Type: TaskNode,
				Name: "Normal Task",
				Position: Position{X: 200, Y: 0},
				Job: &JobConfig{
					Queue:    "normal",
					Type:     "process_normal",
					Priority: "normal",
					Timeout:  5 * time.Minute,
				},
			},
			{
				ID:   "loop",
				Type: LoopNode,
				Name: "Process Loop",
				Position: Position{X: 400, Y: 0},
				Loop: &LoopConfig{
					Iterator:       "items",
					Parallel:       false,
					MaxIterations:  10,
					BreakCondition: "item.processed == true",
				},
			},
			{
				ID:   "loop_task",
				Type: TaskNode,
				Name: "Loop Processing",
				Position: Position{X: 500, Y: 0},
				Job: &JobConfig{
					Queue: "loop_processing",
					Type:  "process_item",
				},
			},
			{
				ID:   "compensate",
				Type: CompensateNode,
				Name: "Compensate Failed Task",
				Position: Position{X: 400, Y: 100},
				CompensationJob: &JobConfig{
					Queue:    "compensation",
					Type:     "compensate_failure",
					Priority: "high",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 600, Y: 0},
				Tags: []string{"exit"},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "decision", Type: SequentialEdge, Label: "Start Processing"},
			{ID: "e2", From: "decision", To: "parallel_high", Type: ConditionalEdge, Condition: "data.priority == 'high'"},
			{ID: "e3", From: "decision", To: "delay", Type: ConditionalEdge, Condition: "data.priority == 'low'"},
			{ID: "e4", From: "decision", To: "task_normal", Type: SequentialEdge},
			{ID: "e5", From: "parallel_high", To: "task_urgent", Type: SequentialEdge},
			{ID: "e6", From: "parallel_high", To: "task_validate", Type: SequentialEdge},
			{ID: "e7", From: "delay", To: "task_normal", Type: SequentialEdge, Delay: 1 * time.Second},
			{ID: "e8", From: "task_urgent", To: "loop", Type: SequentialEdge},
			{ID: "e9", From: "task_validate", To: "loop", Type: SequentialEdge},
			{ID: "e10", From: "task_normal", To: "loop", Type: SequentialEdge},
			{ID: "e11", From: "loop", To: "loop_task", Type: SequentialEdge},
			{ID: "e12", From: "loop_task", To: "loop", Type: LoopbackEdge, Label: "Continue Loop"},
			{ID: "e13", From: "loop", To: "end", Type: SequentialEdge},
			{ID: "e14", From: "task_urgent", To: "compensate", Type: CompensationEdge, Label: "On Failure"},
		},
		CreatedAt: time.Now(),
		CreatedBy: "integration_test",
		UpdatedAt: time.Now(),
		UpdatedBy: "integration_test",
		Tags:      []string{"integration", "complex"},
	}

	// Test validation
	result := builder.ValidateDAG(workflow)
	if !result.Valid {
		t.Fatalf("Complex workflow should be valid, errors: %v", result.Errors)
	}

	// Test JSON serialization
	jsonData, err := workflow.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize workflow to JSON: %v", err)
	}

	// Test JSON deserialization
	var deserializedWorkflow WorkflowDefinition
	err = json.Unmarshal([]byte(jsonData), &deserializedWorkflow)
	if err != nil {
		t.Fatalf("Failed to deserialize workflow from JSON: %v", err)
	}

	// Verify deserialized workflow
	if deserializedWorkflow.ID != workflow.ID {
		t.Errorf("Expected ID %s, got %s", workflow.ID, deserializedWorkflow.ID)
	}

	if len(deserializedWorkflow.Nodes) != len(workflow.Nodes) {
		t.Errorf("Expected %d nodes, got %d", len(workflow.Nodes), len(deserializedWorkflow.Nodes))
	}

	if len(deserializedWorkflow.Edges) != len(workflow.Edges) {
		t.Errorf("Expected %d edges, got %d", len(workflow.Edges), len(deserializedWorkflow.Edges))
	}

	// Validate the deserialized workflow
	deserializedResult := builder.ValidateDAG(&deserializedWorkflow)
	if !deserializedResult.Valid {
		t.Errorf("Deserialized workflow should be valid, errors: %v", deserializedResult.Errors)
	}
}

func TestWorkflowManipulation_Integration(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Start with a simple workflow
	workflow := &WorkflowDefinition{
		ID:      "manipulation-test",
		Name:    "Workflow Manipulation Test",
		Version: "1.0",
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start", Position: Position{X: 0, Y: 0}},
			{ID: "end", Type: EndNode, Name: "End", Position: Position{X: 200, Y: 0}},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "end", Type: SequentialEdge},
		},
	}

	// Validate initial workflow
	result := builder.ValidateDAG(workflow)
	if !result.Valid {
		t.Fatalf("Initial workflow should be valid, errors: %v", result.Errors)
	}

	// Add a task node in the middle
	taskNode := Node{
		ID:   "task1",
		Type: TaskNode,
		Name: "Middle Task",
		Position: Position{X: 100, Y: 0},
		Job: &JobConfig{
			Queue: "default",
			Type:  "process_data",
		},
	}
	workflow.AddNode(taskNode)

	// Remove the direct edge from start to end
	workflow.RemoveEdge("e1")

	// Add new edges through the task
	workflow.AddEdge(Edge{ID: "e2", From: "start", To: "task1", Type: SequentialEdge})
	workflow.AddEdge(Edge{ID: "e3", From: "task1", To: "end", Type: SequentialEdge})

	// Validate modified workflow
	result = builder.ValidateDAG(workflow)
	if !result.Valid {
		t.Errorf("Modified workflow should be valid, errors: %v", result.Errors)
	}

	// Test node retrieval
	retrievedNode := workflow.GetNode("task1")
	if retrievedNode == nil {
		t.Error("Should be able to retrieve added node")
	}
	if retrievedNode.Name != "Middle Task" {
		t.Errorf("Expected node name 'Middle Task', got '%s'", retrievedNode.Name)
	}

	// Test edge retrieval
	retrievedEdge := workflow.GetEdge("e2")
	if retrievedEdge == nil {
		t.Error("Should be able to retrieve added edge")
	}
	if retrievedEdge.From != "start" || retrievedEdge.To != "task1" {
		t.Errorf("Edge should connect start to task1, got %s to %s", retrievedEdge.From, retrievedEdge.To)
	}

	// Test incoming/outgoing edges
	incomingEdges := workflow.GetIncomingEdges("task1")
	if len(incomingEdges) != 1 {
		t.Errorf("Expected 1 incoming edge to task1, got %d", len(incomingEdges))
	}

	outgoingEdges := workflow.GetOutgoingEdges("task1")
	if len(outgoingEdges) != 1 {
		t.Errorf("Expected 1 outgoing edge from task1, got %d", len(outgoingEdges))
	}

	// Test node removal (should remove connected edges too)
	workflow.RemoveNode("task1")

	// Verify node and edges are removed
	if workflow.GetNode("task1") != nil {
		t.Error("Node should be removed")
	}

	if workflow.GetEdge("e2") != nil || workflow.GetEdge("e3") != nil {
		t.Error("Connected edges should be removed when node is removed")
	}

	// Workflow should now have isolated node warnings
	result = builder.ValidateDAG(workflow)
	if result.Valid && len(result.Warnings) == 0 {
		t.Error("Workflow should have warnings about isolated nodes")
	}

	// Should have warnings about isolated nodes
	isolatedWarnings := 0
	for _, warning := range result.Warnings {
		if warning.Type == string(UnreachableNodeError) {
			isolatedWarnings++
		}
	}

	if isolatedWarnings != 2 {
		t.Errorf("Expected 2 isolated node warnings, got %d", isolatedWarnings)
	}
}

func TestComplexValidation_Integration(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Test workflow with multiple validation issues
	workflow := &WorkflowDefinition{
		ID:      "validation-test",
		Name:    "Complex Validation Test",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Task 1",
				Position: Position{X: 100, Y: 0},
				// Missing job config - should cause error
			},
			{
				ID:   "decision",
				Type: DecisionNode,
				Name: "Decision Node",
				Position: Position{X: 200, Y: 0},
				// Missing conditions - should cause error
			},
			{
				ID:   "parallel",
				Type: ParallelNode,
				Name: "Parallel Node",
				Position: Position{X: 300, Y: 0},
				Parallel: &ParallelConfig{
					WaitFor: "invalid", // Invalid wait_for value
					Branches: []string{}, // Empty branches
				},
			},
			{
				ID:   "loop",
				Type: LoopNode,
				Name: "Loop Node",
				Position: Position{X: 400, Y: 0},
				Loop: &LoopConfig{
					// Missing iterator - should cause error
					MaxIterations: -1, // Invalid value
				},
			},
			{
				ID:   "compensate",
				Type: CompensateNode,
				Name: "Compensate Node",
				Position: Position{X: 500, Y: 0},
				// Missing compensation job - should cause error
			},
			{
				ID:   "task2",
				Type: TaskNode,
				Name: "Task 2",
				Position: Position{X: 600, Y: 0},
				Job: &JobConfig{
					Queue: "test",
					Type:  "test_task",
				},
				Retry: &RetryPolicy{
					Strategy:     "invalid_strategy", // Invalid strategy
					MaxAttempts:  0, // Invalid value
					InitialDelay: -1 * time.Second, // Invalid value
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
			{ID: "e2", From: "task1", To: "decision", Type: SequentialEdge},
			{ID: "e3", From: "decision", To: "parallel", Type: SequentialEdge},
			{ID: "e4", From: "parallel", To: "loop", Type: SequentialEdge},
			{ID: "e5", From: "loop", To: "compensate", Type: SequentialEdge},
			{ID: "e6", From: "compensate", To: "task2", Type: SequentialEdge},
			{ID: "e7", From: "nonexistent", To: "task1", Type: SequentialEdge}, // Reference to non-existent node
			{ID: "e8", From: "task2", To: "missing", Type: SequentialEdge}, // Reference to non-existent node
		},
	}

	result := builder.ValidateDAG(workflow)

	// Should have multiple errors
	if result.Valid {
		t.Error("Workflow with multiple issues should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("Should have validation errors")
	}

	// Check for specific error types
	errorTypes := make(map[string]bool)
	for _, err := range result.Errors {
		errorTypes[err.Type] = true
	}

	expectedErrors := []string{
		string(MissingRequiredError),    // Missing job config, conditions, etc.
		string(InvalidConfigError),     // Invalid parallel config, retry policy
		string(InvalidReferenceError),  // Non-existent node references
	}

	for _, expectedType := range expectedErrors {
		if !errorTypes[expectedType] {
			t.Errorf("Expected to find error type %s", expectedType)
		}
	}

	t.Logf("Found %d validation errors (expected multiple)", len(result.Errors))
	for _, err := range result.Errors {
		t.Logf("Error: %s - %s (Location: %s)", err.Type, err.Message, err.Location)
	}
}

func TestWorkflowExecution_Integration(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create a workflow for execution testing
	workflow := &WorkflowDefinition{
		ID:      "execution-test",
		Name:    "Execution Test Workflow",
		Version: "1.0",
		Config: WorkflowConfig{
			Timeout:             10 * time.Minute,
			ConcurrencyLimit:    3,
			EnableCompensation:  true,
			EnableTracing:       true,
			FailureStrategy:     "fail_fast",
		},
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start"},
			{ID: "task1", Type: TaskNode, Name: "Task 1", Job: &JobConfig{Queue: "default", Type: "test"}},
			{ID: "task2", Type: TaskNode, Name: "Task 2", Job: &JobConfig{Queue: "default", Type: "test"}},
			{ID: "end", Type: EndNode, Name: "End"},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
			{ID: "e2", From: "task1", To: "task2", Type: SequentialEdge},
			{ID: "e3", From: "task2", To: "end", Type: SequentialEdge},
		},
	}

	// Validate the workflow
	result := builder.ValidateDAG(workflow)
	if !result.Valid {
		t.Fatalf("Execution test workflow should be valid, errors: %v", result.Errors)
	}

	// Create execution instance
	execution := &WorkflowExecution{
		ID:              "exec-001",
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		Status:          ExecutionPending,
		Parameters: map[string]interface{}{
			"input_data": "test_value",
			"batch_size": 100,
		},
		Context: map[string]interface{}{
			"user_id":      "test_user",
			"request_id":   "req-12345",
			"environment":  "test",
		},
		NodeStates: map[string]*NodeState{
			"start": {
				NodeID: "start",
				Status: NotStarted,
			},
			"task1": {
				NodeID: "task1",
				Status: NotStarted,
			},
			"task2": {
				NodeID: "task2",
				Status: NotStarted,
			},
			"end": {
				NodeID: "end",
				Status: NotStarted,
			},
		},
		StartedAt: time.Now(),
		TraceID:   "trace-12345",
		SpanID:    "span-67890",
		CreatedBy: "integration_test",
	}

	// Verify execution structure
	if execution.WorkflowID != workflow.ID {
		t.Errorf("Expected workflow ID %s, got %s", workflow.ID, execution.WorkflowID)
	}

	if len(execution.NodeStates) != len(workflow.Nodes) {
		t.Errorf("Expected %d node states, got %d", len(workflow.Nodes), len(execution.NodeStates))
	}

	// Test node state updates
	execution.NodeStates["task1"].Status = Running
	execution.NodeStates["task1"].StartedAt = &execution.StartedAt
	execution.NodeStates["task1"].Attempts = 1

	if execution.NodeStates["task1"].Status != Running {
		t.Error("Node state should be updated to Running")
	}

	// Test execution events
	event := ExecutionEvent{
		ID:          "event-001",
		ExecutionID: execution.ID,
		NodeID:      "task1",
		EventType:   "node_started",
		Status:      Running,
		Message:     "Task started execution",
		Data: map[string]interface{}{
			"attempt": 1,
			"queue":   "default",
		},
		Timestamp: time.Now(),
		TraceID:   execution.TraceID,
		SpanID:    "span-task1",
	}

	// Verify event structure
	if event.ExecutionID != execution.ID {
		t.Errorf("Expected execution ID %s, got %s", execution.ID, event.ExecutionID)
	}

	if event.NodeID != "task1" {
		t.Errorf("Expected node ID task1, got %s", event.NodeID)
	}
}