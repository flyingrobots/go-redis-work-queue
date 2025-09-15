// Copyright 2025 James Ross
package visual_dag_builder

import (
	"testing"
	"time"
)

// Benchmark for DAG validation with small workflows
func BenchmarkValidateDAG_Small(b *testing.B) {
	builder := NewDAGBuilder(Config{})

	workflow := &WorkflowDefinition{
		ID:   "small-benchmark",
		Name: "Small Benchmark Workflow",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.ValidateDAG(workflow)
	}
}

// Benchmark for DAG validation with medium workflows
func BenchmarkValidateDAG_Medium(b *testing.B) {
	builder := NewDAGBuilder(Config{})

	// Create a workflow with 20 nodes and 25 edges
	nodes := []Node{
		{ID: "start", Type: StartNode, Name: "Start"},
		{ID: "end", Type: EndNode, Name: "End"},
	}

	// Add 18 task nodes
	for i := 1; i <= 18; i++ {
		nodes = append(nodes, Node{
			ID:   "task" + string(rune('0'+i)),
			Type: TaskNode,
			Name: "Task " + string(rune('0'+i)),
			Job:  &JobConfig{Queue: "default", Type: "test"},
		})
	}

	edges := []Edge{
		{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
	}

	// Create sequential chain
	for i := 1; i < 18; i++ {
		edges = append(edges, Edge{
			ID:   "e" + string(rune('1'+i)),
			From: "task" + string(rune('0'+i)),
			To:   "task" + string(rune('1'+i)),
			Type: SequentialEdge,
		})
	}

	// Add some parallel branches
	edges = append(edges,
		Edge{ID: "e19", From: "task5", To: "task8", Type: SequentialEdge},
		Edge{ID: "e20", From: "task5", To: "task9", Type: SequentialEdge},
		Edge{ID: "e21", From: "task8", To: "task12", Type: SequentialEdge},
		Edge{ID: "e22", From: "task9", To: "task12", Type: SequentialEdge},
		Edge{ID: "e23", From: "task18", To: "end", Type: SequentialEdge},
	)

	workflow := &WorkflowDefinition{
		ID:    "medium-benchmark",
		Name:  "Medium Benchmark Workflow",
		Nodes: nodes,
		Edges: edges,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.ValidateDAG(workflow)
	}
}

// Benchmark for cycle detection
func BenchmarkHasCycle_Large(b *testing.B) {
	builder := NewDAGBuilder(Config{})

	// Create a large workflow without cycles
	nodes := []Node{}
	edges := []Edge{}

	// Create 100 nodes in a chain
	for i := 0; i < 100; i++ {
		nodes = append(nodes, Node{
			ID:   "node" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Type: TaskNode,
			Name: "Node " + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Job:  &JobConfig{Queue: "default", Type: "test"},
		})

		if i > 0 {
			edges = append(edges, Edge{
				ID:   "e" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
				From: "node" + string(rune('0'+(i-1)%10)) + string(rune('0'+((i-1)/10)%10)),
				To:   "node" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
				Type: SequentialEdge,
			})
		}
	}

	workflow := &WorkflowDefinition{
		ID:    "large-cycle-benchmark",
		Name:  "Large Cycle Benchmark",
		Nodes: nodes,
		Edges: edges,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.hasCycle(workflow)
	}
}

// Benchmark for topological sorting
func BenchmarkTopologicalSort_Large(b *testing.B) {
	builder := NewDAGBuilder(Config{})

	// Create a complex DAG with multiple paths
	nodes := []Node{
		{ID: "start", Type: StartNode, Name: "Start"},
		{ID: "end", Type: EndNode, Name: "End"},
	}

	// Add 50 task nodes
	for i := 1; i <= 50; i++ {
		nodes = append(nodes, Node{
			ID:   "task" + string(rune('0'+(i%10))) + string(rune('0'+(i/10))),
			Type: TaskNode,
			Name: "Task " + string(rune('0'+(i%10))) + string(rune('0'+(i/10))),
			Job:  &JobConfig{Queue: "default", Type: "test"},
		})
	}

	edges := []Edge{
		{ID: "e_start", From: "start", To: "task01", Type: SequentialEdge},
	}

	// Create multiple parallel branches and convergence points
	for i := 1; i <= 48; i++ {
		fromNode := "task" + string(rune('0'+(i%10))) + string(rune('0'+(i/10)))
		toNode := "task" + string(rune('0'+((i+1)%10))) + string(rune('0'+((i+1)/10)))

		edges = append(edges, Edge{
			ID:   "e" + string(rune('0'+(i%10))) + string(rune('0'+(i/10))),
			From: fromNode,
			To:   toNode,
			Type: SequentialEdge,
		})

		// Add some parallel branches every 10 nodes
		if i%10 == 0 && i < 40 {
			parallelNode := "task" + string(rune('0'+((i+5)%10))) + string(rune('0'+((i+5)/10)))
			edges = append(edges, Edge{
				ID:   "p" + string(rune('0'+(i%10))) + string(rune('0'+(i/10))),
				From: fromNode,
				To:   parallelNode,
				Type: SequentialEdge,
			})
		}
	}

	edges = append(edges, Edge{ID: "e_end", From: "task50", To: "end", Type: SequentialEdge})

	workflow := &WorkflowDefinition{
		ID:    "large-topo-benchmark",
		Name:  "Large Topological Sort Benchmark",
		Nodes: nodes,
		Edges: edges,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.TopologicalSort(workflow)
	}
}

// Benchmark for workflow serialization
func BenchmarkWorkflowSerialization(b *testing.B) {
	// Create a moderately complex workflow
	workflow := &WorkflowDefinition{
		ID:      "serialization-benchmark",
		Name:    "Serialization Benchmark Workflow",
		Version: "1.0",
		Nodes: []Node{
			{ID: "start", Type: StartNode, Name: "Start"},
			{
				ID:   "task1",
				Type: TaskNode,
				Name: "Task 1",
				Job:  &JobConfig{Queue: "default", Type: "test", Timeout: 30 * time.Second},
				Retry: &RetryPolicy{
					Strategy:     "exponential",
					MaxAttempts:  3,
					InitialDelay: 1 * time.Second,
					MaxDelay:     5 * time.Minute,
					Multiplier:   2.0,
					Jitter:       true,
				},
			},
			{
				ID:   "decision",
				Type: DecisionNode,
				Name: "Decision",
				Conditions: []DecisionCondition{
					{Expression: "result == 'success'", Target: "task2"},
					{Expression: "result == 'retry'", Target: "task1"},
				},
				DefaultTarget: "end",
			},
			{
				ID:   "task2",
				Type: TaskNode,
				Name: "Task 2",
				Job:  &JobConfig{Queue: "priority", Type: "process"},
			},
			{ID: "end", Type: EndNode, Name: "End"},
		},
		Edges: []Edge{
			{ID: "e1", From: "start", To: "task1", Type: SequentialEdge},
			{ID: "e2", From: "task1", To: "decision", Type: SequentialEdge},
			{ID: "e3", From: "decision", To: "task2", Type: ConditionalEdge},
			{ID: "e4", From: "decision", To: "task1", Type: ConditionalEdge},
			{ID: "e5", From: "decision", To: "end", Type: ConditionalEdge},
			{ID: "e6", From: "task2", To: "end", Type: SequentialEdge},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, err := workflow.ToJSON()
		if err != nil {
			b.Fatal(err)
		}

		var newWorkflow WorkflowDefinition
		err = newWorkflow.FromJSON(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

// Benchmark for workflow creation and node addition
func BenchmarkWorkflowBuilding(b *testing.B) {
	builder := NewDAGBuilder(Config{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workflow := builder.CreateWorkflow("benchmark-workflow", "Benchmark workflow")

		// Add nodes
		for j := 1; j <= 10; j++ {
			node := Node{
				ID:   "task" + string(rune('0'+j%10)),
				Type: TaskNode,
				Name: "Task " + string(rune('0'+j%10)),
				Job:  &JobConfig{Queue: "default", Type: "test"},
			}
			builder.AddNode(workflow, node)
		}

		// Add edges
		for j := 1; j < 10; j++ {
			edge := Edge{
				ID:   "e" + string(rune('0'+j%10)),
				From: "task" + string(rune('0'+j%10)),
				To:   "task" + string(rune('0'+(j+1)%10)),
				Type: SequentialEdge,
			}
			builder.AddEdge(workflow, edge)
		}
	}
}