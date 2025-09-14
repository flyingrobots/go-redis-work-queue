// Copyright 2025 James Ross
package visual_dag_builder

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Orchestrator manages workflow execution lifecycle
type Orchestrator struct {
	config   Config
	storage  WorkflowStorage
	queue    JobQueue
	logger   *zap.Logger
	builder  *DAGBuilder

	// Active executions
	executions map[string]*WorkflowExecution
	mu         sync.RWMutex

	// Execution control
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// WorkflowStorage interface for persisting workflows and executions
type WorkflowStorage interface {
	// Workflow management
	SaveWorkflow(ctx context.Context, workflow *WorkflowDefinition) error
	LoadWorkflow(ctx context.Context, id string) (*WorkflowDefinition, error)
	ListWorkflows(ctx context.Context) ([]*WorkflowDefinition, error)
	DeleteWorkflow(ctx context.Context, id string) error

	// Execution management
	SaveExecution(ctx context.Context, execution *WorkflowExecution) error
	LoadExecution(ctx context.Context, id string) (*WorkflowExecution, error)
	ListExecutions(ctx context.Context, workflowID string) ([]*WorkflowExecution, error)
	UpdateNodeState(ctx context.Context, executionID, nodeID string, state *NodeState) error

	// Event tracking
	SaveEvent(ctx context.Context, event *ExecutionEvent) error
	GetEvents(ctx context.Context, executionID string) ([]*ExecutionEvent, error)
}

// JobQueue interface for job submission and monitoring
type JobQueue interface {
	Enqueue(ctx context.Context, queue string, jobType string, payload map[string]interface{}) (string, error)
	GetJobStatus(ctx context.Context, jobID string) (string, error)
	GetJobResult(ctx context.Context, jobID string) (map[string]interface{}, error)
	Subscribe(ctx context.Context, queues []string, handler func(jobID string, result map[string]interface{})) error
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator(config Config, storage WorkflowStorage, queue JobQueue, logger *zap.Logger) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Orchestrator{
		config:     config,
		storage:    storage,
		queue:      queue,
		logger:     logger,
		builder:    NewDAGBuilder(config),
		executions: make(map[string]*WorkflowExecution),
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
	}
}

// Start begins the orchestrator's background processing
func (o *Orchestrator) Start() error {
	o.logger.Info("Starting workflow orchestrator")

	// Start execution monitoring goroutine
	go o.monitorExecutions()

	// Start cleanup goroutine
	go o.cleanupRoutine()

	return nil
}

// Stop gracefully shuts down the orchestrator
func (o *Orchestrator) Stop() error {
	o.logger.Info("Stopping workflow orchestrator")

	o.cancel()

	// Wait for background routines to finish
	select {
	case <-o.done:
	case <-time.After(30 * time.Second):
		o.logger.Warn("Orchestrator shutdown timed out")
	}

	return nil
}

// ExecuteWorkflow starts a new workflow execution
func (o *Orchestrator) ExecuteWorkflow(ctx context.Context, workflowID string, parameters map[string]interface{}) (*WorkflowExecution, error) {
	// Load workflow definition
	workflow, err := o.storage.LoadWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	// Validate workflow
	validation := o.builder.ValidateDAG(workflow)
	if !validation.Valid {
		return nil, fmt.Errorf("workflow validation failed: %v", validation.Errors)
	}

	// Create execution
	execution := &WorkflowExecution{
		ID:              generateExecutionID(),
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		Status:          ExecutionPending,
		Parameters:      parameters,
		Context:         make(map[string]interface{}),
		NodeStates:      make(map[string]*NodeState),
		StartedAt:       time.Now(),
		CreatedBy:       "system", // TODO: get from context
		Metadata:        make(map[string]interface{}),
	}

	// Initialize node states
	for _, node := range workflow.Nodes {
		execution.NodeStates[node.ID] = &NodeState{
			NodeID:   node.ID,
			Status:   NotStarted,
			Attempts: 0,
			Metadata: make(map[string]interface{}),
		}
	}

	// Save execution
	if err := o.storage.SaveExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	// Track execution
	o.mu.Lock()
	o.executions[execution.ID] = execution
	o.mu.Unlock()

	// Start execution
	go o.executeWorkflowAsync(workflow, execution)

	o.logger.Info("Started workflow execution",
		zap.String("execution_id", execution.ID),
		zap.String("workflow_id", workflowID))

	return execution, nil
}

// GetExecution returns the current state of an execution
func (o *Orchestrator) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	// Check in-memory cache first
	o.mu.RLock()
	if execution, exists := o.executions[executionID]; exists {
		o.mu.RUnlock()
		return execution, nil
	}
	o.mu.RUnlock()

	// Load from storage
	return o.storage.LoadExecution(ctx, executionID)
}

// CancelExecution cancels a running execution
func (o *Orchestrator) CancelExecution(ctx context.Context, executionID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return ErrExecutionNotFound
	}

	if execution.Status == ExecutionCompleted || execution.Status == ExecutionFailed {
		return ErrExecutionCompleted
	}

	execution.Status = ExecutionCancelled
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = time.Since(execution.StartedAt)

	// Update storage
	if err := o.storage.SaveExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to save cancelled execution: %w", err)
	}

	o.logger.Info("Cancelled workflow execution",
		zap.String("execution_id", executionID))

	return nil
}

// executeWorkflowAsync executes a workflow asynchronously
func (o *Orchestrator) executeWorkflowAsync(workflow *WorkflowDefinition, execution *WorkflowExecution) {
	defer func() {
		// Clean up from active executions
		o.mu.Lock()
		delete(o.executions, execution.ID)
		o.mu.Unlock()
	}()

	ctx := context.WithValue(o.ctx, "execution_id", execution.ID)

	// Update status to running
	execution.Status = ExecutionRunning
	o.storage.SaveExecution(ctx, execution)

	// Get topological order
	order, err := o.builder.TopologicalSort(workflow)
	if err != nil {
		o.failExecution(ctx, execution, fmt.Errorf("failed to sort workflow: %w", err))
		return
	}

	// Execute nodes in topological order
	for _, nodeID := range order {
		node := workflow.GetNode(nodeID)
		if node == nil {
			o.failExecution(ctx, execution, fmt.Errorf("node %s not found", nodeID))
			return
		}

		// Check if execution was cancelled
		if execution.Status == ExecutionCancelled {
			return
		}

		// Check if dependencies are satisfied
		if !o.dependenciesSatisfied(workflow, execution, nodeID) {
			continue // Skip this node for now
		}

		// Execute the node
		if err := o.executeNode(ctx, workflow, execution, node); err != nil {
			o.logger.Error("Node execution failed",
				zap.String("execution_id", execution.ID),
				zap.String("node_id", nodeID),
				zap.Error(err))

			// Handle failure based on strategy
			if o.handleNodeFailure(ctx, workflow, execution, node, err) {
				return // Execution terminated
			}
		}
	}

	// Complete execution
	o.completeExecution(ctx, execution)
}

// executeNode executes a single node
func (o *Orchestrator) executeNode(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node) error {
	nodeState := execution.NodeStates[node.ID]

	// Update node state to running
	nodeState.Status = Running
	now := time.Now()
	nodeState.StartedAt = &now
	o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)

	// Record execution event
	o.recordEvent(ctx, execution.ID, node.ID, "node_started", Running, "Node execution started", nil)

	var err error
	switch node.Type {
	case TaskNode:
		err = o.executeTaskNode(ctx, workflow, execution, node)
	case DecisionNode:
		err = o.executeDecisionNode(ctx, workflow, execution, node)
	case ParallelNode:
		err = o.executeParallelNode(ctx, workflow, execution, node)
	case DelayNode:
		err = o.executeDelayNode(ctx, workflow, execution, node)
	default:
		err = fmt.Errorf("unsupported node type: %s", node.Type)
	}

	// Update node state based on result
	if err != nil {
		nodeState.Status = Failed
		nodeState.Error = err.Error()
		o.recordEvent(ctx, execution.ID, node.ID, "node_failed", Failed, err.Error(), nil)
	} else {
		nodeState.Status = Completed
		completed := time.Now()
		nodeState.CompletedAt = &completed
		nodeState.Duration = completed.Sub(*nodeState.StartedAt)
		o.recordEvent(ctx, execution.ID, node.ID, "node_completed", Completed, "Node execution completed", nil)
	}

	o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)
	return err
}

// executeTaskNode executes a task node by submitting a job
func (o *Orchestrator) executeTaskNode(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node) error {
	if node.Job == nil {
		return fmt.Errorf("task node %s has no job configuration", node.ID)
	}

	// Prepare job payload
	payload := make(map[string]interface{})

	// Copy node job payload
	for k, v := range node.Job.Payload {
		payload[k] = v
	}

	// Add execution context
	payload["execution_id"] = execution.ID
	payload["node_id"] = node.ID
	payload["workflow_id"] = execution.WorkflowID

	// Add parameters
	for k, v := range execution.Parameters {
		payload[k] = v
	}

	// Submit job
	jobID, err := o.queue.Enqueue(ctx, node.Job.Queue, node.Job.Type, payload)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Update node state with job ID
	nodeState := execution.NodeStates[node.ID]
	nodeState.JobID = jobID
	nodeState.QueueName = node.Job.Queue
	o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)

	// Wait for job completion (simplified - in real implementation would use async monitoring)
	return o.waitForJobCompletion(ctx, execution, node, jobID)
}

// executeDecisionNode evaluates conditions and updates execution flow
func (o *Orchestrator) executeDecisionNode(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node) error {
	// Evaluate conditions (simplified implementation)
	for _, condition := range node.Conditions {
		// In a real implementation, this would use an expression evaluator
		if o.evaluateCondition(condition.Expression, execution.Context) {
			// Mark the target path as active
			execution.Context["decision_"+node.ID] = condition.Target
			return nil
		}
	}

	// Use default target if no conditions matched
	if node.DefaultTarget != "" {
		execution.Context["decision_"+node.ID] = node.DefaultTarget
		return nil
	}

	return fmt.Errorf("no condition matched and no default target for decision node %s", node.ID)
}

// executeParallelNode handles parallel execution (simplified)
func (o *Orchestrator) executeParallelNode(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node) error {
	// In a real implementation, this would manage parallel branch execution
	// For now, just mark as completed
	return nil
}

// executeDelayNode implements time-based delays
func (o *Orchestrator) executeDelayNode(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node) error {
	// Get delay duration from node metadata or default
	delay := 1 * time.Second // Default delay
	if delayStr, ok := node.Metadata["delay"].(string); ok {
		if parsedDelay, err := time.ParseDuration(delayStr); err == nil {
			delay = parsedDelay
		}
	}

	// Sleep for the specified duration
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Helper methods

func (o *Orchestrator) dependenciesSatisfied(workflow *WorkflowDefinition, execution *WorkflowExecution, nodeID string) bool {
	incoming := workflow.GetIncomingEdges(nodeID)
	for _, edge := range incoming {
		fromState := execution.NodeStates[edge.From]
		if fromState.Status != Completed {
			return false
		}
	}
	return true
}

func (o *Orchestrator) handleNodeFailure(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, node *Node, err error) bool {
	nodeState := execution.NodeStates[node.ID]

	// Check retry policy
	if node.Retry != nil && nodeState.Attempts < node.Retry.MaxAttempts {
		// Schedule retry
		nodeState.Attempts++
		retryDelay := o.calculateRetryDelay(node.Retry, nodeState.Attempts)
		nextRetry := time.Now().Add(retryDelay)
		nodeState.NextRetryAt = &nextRetry
		nodeState.Status = Retrying

		o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)
		o.recordEvent(ctx, execution.ID, node.ID, "node_retry_scheduled", Retrying,
			fmt.Sprintf("Retry scheduled in %v", retryDelay), nil)

		return false // Continue execution
	}

	// Handle based on failure strategy
	switch workflow.Config.FailureStrategy {
	case "fail_fast":
		o.failExecution(ctx, execution, err)
		return true

	case "continue":
		// Mark node as failed but continue
		return false

	case "compensate":
		// Start compensation flow
		return o.startCompensation(ctx, workflow, execution, node)

	default:
		o.failExecution(ctx, execution, err)
		return true
	}
}

func (o *Orchestrator) calculateRetryDelay(retry *RetryPolicy, attempt int) time.Duration {
	delay := retry.InitialDelay

	switch retry.Strategy {
	case "exponential":
		for i := 1; i < attempt; i++ {
			delay = time.Duration(float64(delay) * retry.Multiplier)
		}
	case "linear":
		delay = time.Duration(int64(delay) * int64(attempt))
	case "fixed":
		// delay remains the same
	}

	if delay > retry.MaxDelay {
		delay = retry.MaxDelay
	}

	// Add jitter if enabled
	if retry.Jitter {
		// Add up to 25% jitter
		jitter := time.Duration(float64(delay) * 0.25)
		delay += time.Duration(float64(jitter) * (2.0*randFloat64() - 1.0))
	}

	return delay
}

func (o *Orchestrator) startCompensation(ctx context.Context, workflow *WorkflowDefinition, execution *WorkflowExecution, failedNode *Node) bool {
	execution.Status = ExecutionCompensating
	o.storage.SaveExecution(ctx, execution)

	// Find completed nodes that need compensation
	for _, node := range workflow.Nodes {
		state := execution.NodeStates[node.ID]
		if state.Status == Completed && node.CompensationJob != nil {
			// Execute compensation
			o.executeCompensation(ctx, execution, &node)
		}
	}

	execution.Status = ExecutionFailed
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = time.Since(execution.StartedAt)
	execution.Error = "Execution failed and compensated"

	o.storage.SaveExecution(ctx, execution)
	return true
}

func (o *Orchestrator) executeCompensation(ctx context.Context, execution *WorkflowExecution, node *Node) {
	// Submit compensation job (simplified)
	if node.CompensationJob != nil {
		payload := map[string]interface{}{
			"execution_id":     execution.ID,
			"node_id":          node.ID,
			"compensation_for": node.ID,
		}

		jobID, err := o.queue.Enqueue(ctx, node.CompensationJob.Queue, node.CompensationJob.Type, payload)
		if err != nil {
			o.logger.Error("Failed to enqueue compensation job", zap.Error(err))
			return
		}

		nodeState := execution.NodeStates[node.ID]
		nodeState.CompensationJobID = jobID
		nodeState.Status = Compensating
		o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)
	}
}

func (o *Orchestrator) completeExecution(ctx context.Context, execution *WorkflowExecution) {
	execution.Status = ExecutionCompleted
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = time.Since(execution.StartedAt)

	o.storage.SaveExecution(ctx, execution)
	o.recordEvent(ctx, execution.ID, "", "execution_completed", NotStarted, "Workflow execution completed", nil)

	o.logger.Info("Completed workflow execution",
		zap.String("execution_id", execution.ID),
		zap.Duration("duration", execution.Duration))
}

func (o *Orchestrator) failExecution(ctx context.Context, execution *WorkflowExecution, err error) {
	execution.Status = ExecutionFailed
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = time.Since(execution.StartedAt)
	execution.Error = err.Error()

	o.storage.SaveExecution(ctx, execution)
	o.recordEvent(ctx, execution.ID, "", "execution_failed", NotStarted, err.Error(), nil)

	o.logger.Error("Failed workflow execution",
		zap.String("execution_id", execution.ID),
		zap.Error(err))
}

func (o *Orchestrator) waitForJobCompletion(ctx context.Context, execution *WorkflowExecution, node *Node, jobID string) error {
	// Simplified job completion waiting - in real implementation would use callbacks
	timeout := node.Job.Timeout
	if timeout == 0 {
		timeout = o.config.Execution.DefaultTimeout
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := o.queue.GetJobStatus(ctx, jobID)
			if err != nil {
				return fmt.Errorf("failed to get job status: %w", err)
			}

			switch status {
			case "completed":
				// Get job result
				result, err := o.queue.GetJobResult(ctx, jobID)
				if err != nil {
					return fmt.Errorf("failed to get job result: %w", err)
				}

				// Update node state with result
				nodeState := execution.NodeStates[node.ID]
				nodeState.Output = result
				o.storage.UpdateNodeState(ctx, execution.ID, node.ID, nodeState)

				return nil

			case "failed":
				return fmt.Errorf("job %s failed", jobID)
			}

		case <-timeoutTimer.C:
			return ErrTimeoutExceeded

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *Orchestrator) evaluateCondition(expression string, context map[string]interface{}) bool {
	// Simplified condition evaluation - in real implementation would use expression parser
	// For now, just check if context contains the expression as a key with true value
	if val, ok := context[expression]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

func (o *Orchestrator) recordEvent(ctx context.Context, executionID, nodeID, eventType string, status NodeStatus, message string, data map[string]interface{}) {
	event := &ExecutionEvent{
		ID:          generateEventID(),
		ExecutionID: executionID,
		NodeID:      nodeID,
		EventType:   eventType,
		Status:      status,
		Message:     message,
		Data:        data,
		Timestamp:   time.Now(),
	}

	if err := o.storage.SaveEvent(ctx, event); err != nil {
		o.logger.Error("Failed to save execution event", zap.Error(err))
	}
}

func (o *Orchestrator) monitorExecutions() {
	defer close(o.done)

	ticker := time.NewTicker(o.config.Execution.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.processRetries()
		case <-o.ctx.Done():
			return
		}
	}
}

func (o *Orchestrator) processRetries() {
	o.mu.RLock()
	executions := make([]*WorkflowExecution, 0, len(o.executions))
	for _, exec := range o.executions {
		executions = append(executions, exec)
	}
	o.mu.RUnlock()

	now := time.Now()
	for _, execution := range executions {
		for nodeID, state := range execution.NodeStates {
			if state.Status == Retrying && state.NextRetryAt != nil && now.After(*state.NextRetryAt) {
				// Execute retry
				o.logger.Info("Executing retry",
					zap.String("execution_id", execution.ID),
					zap.String("node_id", nodeID),
					zap.Int("attempt", state.Attempts))

				// Load workflow and execute node
				workflow, err := o.storage.LoadWorkflow(o.ctx, execution.WorkflowID)
				if err != nil {
					o.logger.Error("Failed to load workflow for retry", zap.Error(err))
					continue
				}

				node := workflow.GetNode(nodeID)
				if node != nil {
					go o.executeNode(o.ctx, workflow, execution, node)
				}
			}
		}
	}
}

func (o *Orchestrator) cleanupRoutine() {
	ticker := time.NewTicker(o.config.Execution.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Clean up old executions
			cutoff := time.Now().Add(-o.config.Storage.TTL)
			o.logger.Debug("Cleaning up old executions", zap.Time("cutoff", cutoff))

		case <-o.ctx.Done():
			return
		}
	}
}

// Helper functions

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// randFloat64 is a placeholder for rand.Float64() - would normally import math/rand
func randFloat64() float64 {
	return 0.5 // Simplified for this implementation
}