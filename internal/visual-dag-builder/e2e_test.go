// Copyright 2025 James Ross
package visual_dag_builder

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// E2E tests that simulate real-world usage scenarios

func TestRealWorldWorkflow_E2E(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Create a realistic data processing workflow
	workflow := &WorkflowDefinition{
		ID:          "data-pipeline-v1",
		Name:        "Customer Data Processing Pipeline",
		Version:     "1.2.0",
		Description: "Processes customer data through validation, enrichment, and analytics",
		Config: WorkflowConfig{
			Timeout:             60 * time.Minute,
			ConcurrencyLimit:    10,
			EnableCompensation:  true,
			EnableTracing:       true,
			FailureStrategy:     "compensate",
		},
		Nodes: []Node{
			{
				ID:   "start",
				Type: StartNode,
				Name: "Pipeline Start",
				Position: Position{X: 50, Y: 300},
				Tags: []string{"entry-point"},
			},
			{
				ID:   "validate_input",
				Type: TaskNode,
				Name: "Validate Customer Data",
				Position: Position{X: 200, Y: 300},
				Job: &JobConfig{
					Queue:    "validation",
					Type:     "validate_customer_data",
					Priority: "high",
					Timeout:  5 * time.Minute,
				},
				Retry: &RetryPolicy{
					Strategy:     "exponential",
					MaxAttempts:  3,
					InitialDelay: 1 * time.Second,
					MaxDelay:     30 * time.Second,
					Multiplier:   2.0,
					Jitter:       true,
				},
				Tags: []string{"validation", "critical"},
			},
			{
				ID:   "quality_check",
				Type: DecisionNode,
				Name: "Data Quality Decision",
				Position: Position{X: 400, Y: 300},
				Conditions: []DecisionCondition{
					{
						Expression: "validation_score >= 0.95",
						Target:     "enrich_data",
						Label:      "High Quality",
					},
				},
				DefaultTarget: "reject_data",
			},
			{
				ID:   "enrich_data",
				Type: ParallelNode,
				Name: "Data Enrichment",
				Position: Position{X: 600, Y: 200},
				Parallel: &ParallelConfig{
					WaitFor:          "all",
					Branches:         []string{"demographic_enrichment"},
					ConcurrencyLimit: 1,
				},
			},
			{
				ID:   "demographic_enrichment",
				Type: TaskNode,
				Name: "Demographic Enrichment",
				Position: Position{X: 800, Y: 150},
				Job: &JobConfig{
					Queue:    "enrichment",
					Type:     "enrich_demographics",
					Priority: "normal",
					Timeout:  10 * time.Minute,
				},
			},
			{
				ID:   "reject_data",
				Type: TaskNode,
				Name: "Data Rejection Handler",
				Position: Position{X: 600, Y: 500},
				Job: &JobConfig{
					Queue:    "rejection",
					Type:     "handle_rejection",
					Priority: "low",
					Timeout:  2 * time.Minute,
				},
			},
			{
				ID:   "pipeline_end",
				Type: EndNode,
				Name: "Pipeline Complete",
				Position: Position{X: 1000, Y: 200},
			},
			{
				ID:   "failure_end",
				Type: EndNode,
				Name: "Pipeline Failed",
				Position: Position{X: 800, Y: 500},
			},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "validate_input", Type: SequentialEdge},
			{ID: "e2", From: "validate_input", To: "quality_check", Type: SequentialEdge},
			{ID: "e3", From: "quality_check", To: "enrich_data", Type: ConditionalEdge, Condition: "validation_score >= 0.95"},
			{ID: "e4", From: "quality_check", To: "reject_data", Type: SequentialEdge},
			{ID: "e5", From: "enrich_data", To: "demographic_enrichment", Type: SequentialEdge},
			{ID: "e6", From: "demographic_enrichment", To: "pipeline_end", Type: SequentialEdge},
			{ID: "e7", From: "reject_data", To: "failure_end", Type: SequentialEdge},
		},
		CreatedAt: time.Now(),
		CreatedBy: "integration_test",
		UpdatedAt: time.Now(),
		UpdatedBy: "integration_test",
		Tags:      []string{"integration", "e2e"},
	}

	// Test validation
	t.Run("WorkflowValidation", func(t *testing.T) {
		result := builder.ValidateDAG(workflow)
		if !result.Valid {
			t.Fatalf("E2E workflow should be valid, errors: %v", result.Errors)
		}

		if len(workflow.Nodes) < 5 {
			t.Errorf("Expected complex workflow with many nodes, got %d", len(workflow.Nodes))
		}

		// Verify key node types are present
		nodeTypes := make(map[NodeType]bool)
		for _, node := range workflow.Nodes {
			nodeTypes[node.Type] = true
		}

		expectedTypes := []NodeType{StartNode, EndNode, TaskNode, DecisionNode, ParallelNode}
		for _, expectedType := range expectedTypes {
			if !nodeTypes[expectedType] {
				t.Errorf("Expected to find node type %s in workflow", expectedType)
			}
		}
	})

	// Test serialization
	t.Run("Serialization", func(t *testing.T) {
		jsonData, err := workflow.ToJSON()
		if err != nil {
			t.Fatalf("Failed to serialize workflow: %v", err)
		}

		var deserializedWorkflow WorkflowDefinition
		err = json.Unmarshal([]byte(jsonData), &deserializedWorkflow)
		if err != nil {
			t.Fatalf("Failed to deserialize workflow: %v", err)
		}

		if deserializedWorkflow.ID != workflow.ID {
			t.Errorf("Workflow ID mismatch after serialization")
		}

		// Validate deserialized workflow
		result := builder.ValidateDAG(&deserializedWorkflow)
		if !result.Valid {
			t.Errorf("Deserialized workflow should be valid, errors: %v", result.Errors)
		}
	})

	// Test topological sorting
	t.Run("TopologicalSort", func(t *testing.T) {
		sortedNodes, err := builder.TopologicalSort(workflow)
		if err != nil {
			t.Fatalf("Failed to topologically sort workflow: %v", err)
		}

		if len(sortedNodes) != len(workflow.Nodes) {
			t.Errorf("Expected %d nodes in topological sort, got %d", len(workflow.Nodes), len(sortedNodes))
		}

		// Verify start node comes first
		if sortedNodes[0] != "start" {
			t.Errorf("Expected start node first in topological sort, got %s", sortedNodes[0])
		}
	})
}

func TestWorkflowTemplates_E2E(t *testing.T) {
	builder := NewDAGBuilder(Config{})

	// Simple ETL template
	etlTemplate := &WorkflowDefinition{
		ID:          "template-simple-etl",
		Name:        "Simple ETL Template",
		Version:     "1.0.0",
		Description: "Basic Extract-Transform-Load pattern",
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start"},
			{ID: "extract", Type: TaskNode, Name: "Extract Data", Job: &JobConfig{Queue: "etl", Type: "extract"}},
			{ID: "transform", Type: TaskNode, Name: "Transform Data", Job: &JobConfig{Queue: "etl", Type: "transform"}},
			{ID: "load", Type: TaskNode, Name: "Load Data", Job: &JobConfig{Queue: "etl", Type: "load"}},
			{ID: "end", Type: EndNode, Name: "End"},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "extract", Type: SequentialEdge},
			{ID: "e2", From: "extract", To: "transform", Type: SequentialEdge},
			{ID: "e3", From: "transform", To: "load", Type: SequentialEdge},
			{ID: "e4", From: "load", To: "end", Type: SequentialEdge},
		},
		Tags: []string{"template", "etl", "simple"},
	}

	t.Run("TemplateValidation", func(t *testing.T) {
		result := builder.ValidateDAG(etlTemplate)
		if !result.Valid {
			t.Errorf("ETL template should be valid, errors: %v", result.Errors)
		}
	})

	t.Run("TemplateSerialization", func(t *testing.T) {
		tempDir := os.TempDir()
		filename := tempDir + "/etl-template.json"

		jsonData, err := etlTemplate.ToJSON()
		if err != nil {
			t.Fatalf("Failed to serialize template: %v", err)
		}

		err = os.WriteFile(filename, []byte(jsonData), 0644)
		if err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}
		defer os.Remove(filename)

		// Read back and verify
		fileData, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read template file: %v", err)
		}

		var loadedTemplate WorkflowDefinition
		err = json.Unmarshal(fileData, &loadedTemplate)
		if err != nil {
			t.Fatalf("Failed to deserialize template: %v", err)
		}

		if loadedTemplate.ID != etlTemplate.ID {
			t.Errorf("Template ID mismatch after file round-trip")
		}
	})

	t.Run("TemplateInstantiation", func(t *testing.T) {
		// Customize the template
		customWorkflow := *etlTemplate
		customWorkflow.ID = "custom-etl-001"
		customWorkflow.Name = "Customer Data ETL"

		// Validate customized workflow
		result := builder.ValidateDAG(&customWorkflow)
		if !result.Valid {
			t.Errorf("Customized workflow should be valid, errors: %v", result.Errors)
		}
	})
}

func TestCanvasOperations_E2E(t *testing.T) {
	// Create a simple workflow for canvas testing
	workflow := &WorkflowDefinition{
		ID:   "canvas-test",
		Name: "Canvas Test Workflow",
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start", Position: Position{X: 100, Y: 200}},
			{ID: "task1", Type: TaskNode, Name: "Task 1", Position: Position{X: 300, Y: 200}, Job: &JobConfig{Queue: "default", Type: "task"}},
			{ID: "end", Type: EndNode, Name: "End", Position: Position{X: 500, Y: 200}},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
			{ID: "e2", From: "task1", To: "end", Type: SequentialEdge},
		},
	}

	// Create canvas state
	canvas := &CanvasState{
		Workflow:     workflow,
		SelectedNode: "",
		SelectedEdge: "",
		ViewOffset:   Position{X: 0, Y: 0},
		ZoomLevel:    1.0,
		GridSize:     20,
		ShowGrid:     true,
		Mode:         SelectMode,
	}

	t.Run("CanvasOperations", func(t *testing.T) {
		// Test node selection
		canvas.SelectedNode = "task1"
		if canvas.SelectedNode != "task1" {
			t.Error("Node selection failed")
		}

		// Test zoom operations
		canvas.ZoomLevel = 1.5
		if canvas.ZoomLevel != 1.5 {
			t.Error("Zoom level change failed")
		}

		// Test mode changes
		canvas.Mode = AddNodeMode
		if canvas.Mode != AddNodeMode {
			t.Error("Mode change failed")
		}
	})

	t.Run("NodePositioning", func(t *testing.T) {
		for _, node := range workflow.Nodes {
			if node.Position.X < 0 || node.Position.Y < 0 {
				t.Errorf("Node %s has invalid position: %+v", node.ID, node.Position)
			}
		}

		// Test position updates
		task1Node := workflow.GetNode("task1")
		if task1Node == nil {
			t.Fatal("Task1 node not found")
		}

		originalX := task1Node.Position.X
		task1Node.Position.X = 400

		if task1Node.Position.X != 400 {
			t.Error("Node position update failed")
		}

		// Restore original position
		task1Node.Position.X = originalX
	})

	t.Run("CanvasSerialization", func(t *testing.T) {
		canvasJSON, err := json.Marshal(canvas)
		if err != nil {
			t.Fatalf("Failed to serialize canvas state: %v", err)
		}

		var deserializedCanvas CanvasState
		err = json.Unmarshal(canvasJSON, &deserializedCanvas)
		if err != nil {
			t.Fatalf("Failed to deserialize canvas state: %v", err)
		}

		if deserializedCanvas.ZoomLevel != canvas.ZoomLevel {
			t.Error("Zoom level not preserved in serialization")
		}

		if deserializedCanvas.Mode != canvas.Mode {
			t.Error("Mode not preserved in serialization")
		}
	})
}