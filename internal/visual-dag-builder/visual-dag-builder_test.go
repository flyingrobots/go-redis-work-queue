// Copyright 2025 James Ross
package visual_dag_builder

import (
	"testing"
	"time"
)

func TestNewDAGBuilder(t *testing.T) {
	config := Config{
		Storage: StorageConfig{
			Type:   "memory",
			Prefix: "test:",
		},
	}

	builder := NewDAGBuilder(config)
	if builder == nil {
		t.Fatal("NewDAGBuilder should return a non-nil builder")
	}

	if builder.config.Storage.Type != "memory" {
		t.Errorf("Expected storage type 'memory', got %s", builder.config.Storage.Type)
	}
}

func TestValidateDAG_EmptyWorkflow(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "empty-workflow",
		Name:    "Empty Workflow",
		Version: "1.0",
		Nodes:   []Node{},
		Edges:   []Edge{},
	}

	result := builder.ValidateDAG(workflow)

	if result.Valid {
		t.Error("Empty workflow should be invalid")
	}

	if len(result.Errors) == 0 {
		t.Error("Empty workflow should have validation errors")
	}

	foundEmptyError := false
	for _, err := range result.Errors {
		if err.Type == string(MissingRequiredError) && err.Message == "workflow must have at least one node" {
			foundEmptyError = true
			break
		}
	}

	if !foundEmptyError {
		t.Error("Should have error about empty workflow")
	}
}

func TestValidateDAG_ValidSimpleWorkflow(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "simple-workflow",
		Name:    "Simple Workflow",
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
				Job: &JobConfig{
					Queue:    "default",
					Type:     "process_file",
					Priority: "normal",
					Timeout:  5 * time.Minute,
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 200, Y: 0},
			},
		},
		Edges: []Edge{
			{
				ID:     "edge1",
				From:   "start",
				To:     "task1",
				Type:   SequentialEdge,
			},
			{
				ID:     "edge2",
				From:   "task1",
				To:     "end",
				Type:   SequentialEdge,
			},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Simple workflow should be valid, errors: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Valid workflow should have no errors, got: %v", result.Errors)
	}
}

func TestValidateDAG_DecisionNode(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "decision-workflow",
		Name:    "Decision Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:   "decision",
				Type: DecisionNode,
				Name: "Decision",
				Position: Position{X: 100, Y: 0},
				Conditions: []DecisionCondition{
					{
						Expression: "data.status == 'approved'",
						Target:     "approved_task",
						Label:      "Approved",
					},
					{
						Expression: "data.status == 'rejected'",
						Target:     "rejected_task",
						Label:      "Rejected",
					},
				},
				DefaultTarget: "default_task",
			},
			{
				ID:   "approved_task",
				Type: TaskNode,
				Name: "Approved Task",
				Position: Position{X: 200, Y: -50},
				Job: &JobConfig{
					Queue:    "approved",
					Type:     "process_approved",
					Priority: "high",
				},
			},
			{
				ID:   "rejected_task",
				Type: TaskNode,
				Name: "Rejected Task",
				Position: Position{X: 200, Y: 50},
				Job: &JobConfig{
					Queue:    "rejected",
					Type:     "process_rejected",
					Priority: "low",
				},
			},
			{
				ID:   "default_task",
				Type: TaskNode,
				Name: "Default Task",
				Position: Position{X: 200, Y: 0},
				Job: &JobConfig{
					Queue:    "default",
					Type:     "process_default",
					Priority: "normal",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 300, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "decision", Type: SequentialEdge},
			{ID: "e2", From: "decision", To: "approved_task", Type: ConditionalEdge, Condition: "data.status == 'approved'"},
			{ID: "e3", From: "decision", To: "rejected_task", Type: ConditionalEdge, Condition: "data.status == 'rejected'"},
			{ID: "e4", From: "decision", To: "default_task", Type: SequentialEdge},
			{ID: "e5", From: "approved_task", To: "end", Type: SequentialEdge},
			{ID: "e6", From: "rejected_task", To: "end", Type: SequentialEdge},
			{ID: "e7", From: "default_task", To: "end", Type: SequentialEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Decision workflow should be valid, errors: %v", result.Errors)
	}
}

func TestValidateDAG_ParallelNode(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "parallel-workflow",
		Name:    "Parallel Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:   "parallel",
				Type: ParallelNode,
				Name: "Parallel Processing",
				Position: Position{X: 100, Y: 0},
				Parallel: &ParallelConfig{
					WaitFor:          "all",
					Branches:         []string{"branch1", "branch2"},
					ConcurrencyLimit: 2,
				},
			},
			{
				ID:   "branch1",
				Type: TaskNode,
				Name: "Branch 1 Task",
				Position: Position{X: 200, Y: -50},
				Job: &JobConfig{
					Queue:    "parallel1",
					Type:     "process_branch1",
					Priority: "normal",
				},
			},
			{
				ID:   "branch2",
				Type: TaskNode,
				Name: "Branch 2 Task",
				Position: Position{X: 200, Y: 50},
				Job: &JobConfig{
					Queue:    "parallel2",
					Type:     "process_branch2",
					Priority: "normal",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 300, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "parallel", Type: SequentialEdge},
			{ID: "e2", From: "parallel", To: "branch1", Type: SequentialEdge},
			{ID: "e3", From: "parallel", To: "branch2", Type: SequentialEdge},
			{ID: "e4", From: "branch1", To: "end", Type: SequentialEdge},
			{ID: "e5", From: "branch2", To: "end", Type: SequentialEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Parallel workflow should be valid, errors: %v", result.Errors)
	}
}

func TestValidateDAG_LoopNode(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "loop-workflow",
		Name:    "Loop Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:   "loop",
				Type: LoopNode,
				Name: "Process Items",
				Position: Position{X: 100, Y: 0},
				Loop: &LoopConfig{
					Iterator:       "items",
					Parallel:       false,
					MaxIterations:  100,
					BreakCondition: "item.processed == true",
				},
			},
			{
				ID:   "task",
				Type: TaskNode,
				Name: "Process Item",
				Position: Position{X: 200, Y: 0},
				Job: &JobConfig{
					Queue:    "items",
					Type:     "process_item",
					Priority: "normal",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 300, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "loop", Type: SequentialEdge},
			{ID: "e2", From: "loop", To: "task", Type: SequentialEdge},
			{ID: "e3", From: "task", To: "loop", Type: LoopbackEdge},
			{ID: "e4", From: "loop", To: "end", Type: SequentialEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Loop workflow should be valid, errors: %v", result.Errors)
	}
}

func TestValidateDAG_CyclicDependency(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "cyclic-workflow",
		Name:    "Cyclic Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Task 1",
				Position: Position{X: 0, Y: 0},
				Job: &JobConfig{Queue: "default", Type: "task1"},
			},
			{
				ID:   "task2",
				Type: TaskNode,
				Name: "Task 2",
				Position: Position{X: 100, Y: 0},
				Job: &JobConfig{Queue: "default", Type: "task2"},
			},
			{
				ID:   "task3",
				Type: TaskNode,
				Name: "Task 3",
				Position: Position{X: 200, Y: 0},
				Job: &JobConfig{Queue: "default", Type: "task3"},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "task1", To: "task2", Type: SequentialEdge},
			{ID: "e2", From: "task2", To: "task3", Type: SequentialEdge},
			{ID: "e3", From: "task3", To: "task1", Type: SequentialEdge}, // Creates cycle
		},
	}

	result := builder.ValidateDAG(workflow)

	if result.Valid {
		t.Error("Cyclic workflow should be invalid")
	}

	// Should have cycle detection error
	foundCycleError := false
	for _, err := range result.Errors {
		if err.Type == string(CyclicDependencyError) {
			foundCycleError = true
			break
		}
	}

	if !foundCycleError {
		t.Error("Should detect cyclic dependency")
	}
}

func TestValidateDAG_MissingNode(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "missing-node-workflow",
		Name:    "Missing Node Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Task 1",
				Position: Position{X: 0, Y: 0},
				Job: &JobConfig{Queue: "default", Type: "task1"},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "task1", To: "missing_task", Type: SequentialEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if result.Valid {
		t.Error("Workflow with missing node should be invalid")
	}

	// Should have missing node error
	foundMissingError := false
	for _, err := range result.Errors {
		if err.Type == string(InvalidReferenceError) {
			foundMissingError = true
			break
		}
	}

	if !foundMissingError {
		t.Error("Should detect missing node reference")
	}
}

func TestValidateDAG_RetryPolicy(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	tests := []struct {
		name        string
		retry       *RetryPolicy
		shouldBeValid bool
	}{
		{
			name: "valid exponential retry",
			retry: &RetryPolicy{
				Strategy:     "exponential",
				MaxAttempts:  3,
				InitialDelay: 1 * time.Second,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			shouldBeValid: true,
		},
		{
			name: "valid fixed retry",
			retry: &RetryPolicy{
				Strategy:     "fixed",
				MaxAttempts:  5,
				InitialDelay: 2 * time.Second,
			},
			shouldBeValid: true,
		},
		{
			name: "invalid retry - zero max attempts",
			retry: &RetryPolicy{
				Strategy:     "exponential",
				MaxAttempts:  0,
				InitialDelay: 1 * time.Second,
			},
			shouldBeValid: false,
		},
		{
			name: "invalid retry - negative delay",
			retry: &RetryPolicy{
				Strategy:     "fixed",
				MaxAttempts:  3,
				InitialDelay: -1 * time.Second,
			},
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &WorkflowDefinition{
				ID:      "retry-test",
				Name:    "Retry Test",
				Version: "1.0",
				Nodes: []Node{
					{
						ID:   "task1",
						Type: TaskNode,
						Name: "Task with Retry",
						Position: Position{X: 0, Y: 0},
						Job: &JobConfig{
							Queue: "default",
							Type:  "retry_task",
						},
						Retry: tt.retry,
					},
				},
				Edges: []Edge{},
			}

			result := builder.ValidateDAG(workflow)

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid workflow, got errors: %v", result.Errors)
			}

			if !tt.shouldBeValid && result.Valid {
				t.Errorf("Expected invalid workflow, but got valid result")
			}
		})
	}
}

func TestValidateDAG_CompensationEdges(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "compensation-workflow",
		Name:    "Compensation Workflow",
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
				Name: "Main Task",
				Position: Position{X: 100, Y: 0},
				Job: &JobConfig{
					Queue: "main",
					Type:  "main_task",
				},
			},
			{
				ID:   "compensate1",
				Type: CompensateNode,
				Name: "Compensate Task",
				Position: Position{X: 100, Y: 100},
				CompensationJob: &JobConfig{
					Queue: "compensation",
					Type:  "compensate_task",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 200, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
			{ID: "e2", From: "task1", To: "end", Type: SequentialEdge},
			{ID: "e3", From: "task1", To: "compensate1", Type: CompensationEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Compensation workflow should be valid, errors: %v", result.Errors)
	}
}

func TestValidateDAG_DelayNode(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:      "delay-workflow",
		Name:    "Delay Workflow",
		Version: "1.0",
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Start",
				Position: Position{X: 0, Y: 0},
			},
			{
				ID:   "delay",
				Type: DelayNode,
				Name: "Wait",
				Position: Position{X: 100, Y: 0},
				DelayConfig: &DelayConfig{
					Duration:  5 * time.Minute,
					Dynamic:   false,
					Expression: "",
				},
			},
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Delayed Task",
				Position: Position{X: 200, Y: 0},
				Job: &JobConfig{
					Queue: "delayed",
					Type:  "delayed_task",
				},
			},
			{
				ID:   "end",
				Type: EndNode,
				Name: "End",
				Position: Position{X: 300, Y: 0},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "delay", Type: SequentialEdge},
			{ID: "e2", From: "delay", To: "task1", Type: SequentialEdge},
			{ID: "e3", From: "task1", To: "end", Type: SequentialEdge},
		},
	}

	result := builder.ValidateDAG(workflow)

	if !result.Valid {
		t.Errorf("Delay workflow should be valid, errors: %v", result.Errors)
	}
}