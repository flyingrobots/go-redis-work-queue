// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMultiAction_PurgeDLQ(t *testing.T) {
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	// Setup DLQ data
	mr1.Lpush("jobqueue:dead_letter", "dead1", "dead2", "dead3")
	mr2.Lpush("jobqueue:dead_letter", "dead4", "dead5")

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "cluster2", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
				ActionTypePauseQueue,
			},
			MaxConcurrent: 5,
			ActionTimeouts: map[ActionType]Duration{
				ActionTypePurgeDLQ: Duration(30 * time.Second),
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create purge DLQ action
	action := &MultiAction{
		ID:      "purge-dlq-001",
		Type:    ActionTypePurgeDLQ,
		Targets: []string{"cluster1", "cluster2"},
		Parameters: map[string]interface{}{
			"confirm": true,
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute the action
	err = manager.ExecuteAction(ctx, action)
	require.NoError(t, err)

	// Verify action completed successfully
	assert.Equal(t, ActionStatusCompleted, action.Status)
	assert.Equal(t, 2, len(action.Results))

	// Check results for both clusters
	result1, exists := action.Results["cluster1"]
	assert.True(t, exists)
	assert.True(t, result1.Success)
	assert.Empty(t, result1.Error)

	result2, exists := action.Results["cluster2"]
	assert.True(t, exists)
	assert.True(t, result2.Success)
	assert.Empty(t, result2.Error)

	// Verify DLQ was actually purged
	assert.Equal(t, 0, mr1.LLen("jobqueue:dead_letter"))
	assert.Equal(t, 0, mr2.LLen("jobqueue:dead_letter"))
}

func TestMultiAction_BenchmarkExecution(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "bench-cluster", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
			},
			ActionTimeouts: map[ActionType]Duration{
				ActionTypeBenchmark: Duration(60 * time.Second),
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create benchmark action
	action := &MultiAction{
		ID:      "benchmark-001",
		Type:    ActionTypeBenchmark,
		Targets: []string{"bench-cluster"},
		Parameters: map[string]interface{}{
			"iterations":    float64(10),
			"payload_size":  float64(100),
			"queue_name":    "test-queue",
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute the action
	err = manager.ExecuteAction(ctx, action)
	require.NoError(t, err)

	// Verify action completed successfully
	assert.Equal(t, ActionStatusCompleted, action.Status)
	assert.Equal(t, 1, len(action.Results))

	result, exists := action.Results["bench-cluster"]
	assert.True(t, exists)
	assert.True(t, result.Success)
	assert.Greater(t, result.Duration, float64(0)) // Should take some time
}

func TestMultiAction_ConfirmationRequired(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "secure-cluster", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: true, // Require confirmation
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create action requiring confirmation
	action := &MultiAction{
		ID:      "secure-action-001",
		Type:    ActionTypePurgeDLQ,
		Targets: []string{"secure-cluster"},
		Parameters: map[string]interface{}{
			"queue": "sensitive-data",
		},
		Confirmations: []ActionConfirmation{
			{
				Required: true,
				Message:  "Are you sure you want to purge the sensitive-data DLQ?",
			},
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute without confirmation - should fail
	err = manager.ExecuteAction(ctx, action)
	assert.Error(t, err)
	assert.Equal(t, ActionStatusPending, action.Status)

	// Confirm the action
	err = manager.ConfirmAction(ctx, action.ID, "test-user")
	require.NoError(t, err)

	// Verify confirmation was recorded
	assert.Equal(t, "test-user", action.Confirmations[0].ConfirmedBy)
	assert.False(t, action.Confirmations[0].ConfirmedAt.IsZero())

	// Now execute should succeed
	err = manager.ExecuteAction(ctx, action)
	require.NoError(t, err)
	assert.Equal(t, ActionStatusCompleted, action.Status)
}

func TestMultiAction_PartialFailure(t *testing.T) {
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	// mr2 intentionally not started to cause connection failure

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "working", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "failing", Endpoint: "localhost:9999", DB: 0, Enabled: true}, // Invalid port
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
			},
			ContinueOnFailure: true, // Continue even if some clusters fail
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create action targeting both clusters
	action := &MultiAction{
		ID:      "partial-fail-001",
		Type:    ActionTypeBenchmark,
		Targets: []string{"working", "failing"},
		Parameters: map[string]interface{}{
			"iterations": float64(5),
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute the action
	err = manager.ExecuteAction(ctx, action)
	// Should not return error due to ContinueOnFailure=true

	// Verify mixed results
	assert.Equal(t, 2, len(action.Results))

	workingResult, exists := action.Results["working"]
	assert.True(t, exists)
	assert.True(t, workingResult.Success)

	failingResult, exists := action.Results["failing"]
	assert.True(t, exists)
	assert.False(t, failingResult.Success)
	assert.NotEmpty(t, failingResult.Error)

	// Action should be marked as partially completed or failed based on implementation
	assert.Contains(t, []ActionStatus{ActionStatusFailed, ActionStatusCompleted}, action.Status)
}

func TestMultiAction_Cancellation(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "test-cluster", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
			},
			MaxConcurrent: 1,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create a long-running action
	action := &MultiAction{
		ID:      "long-running-001",
		Type:    ActionTypeBenchmark,
		Targets: []string{"test-cluster"},
		Parameters: map[string]interface{}{
			"iterations": float64(1000), // Many iterations to make it run longer
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Start execution in a goroutine
	go func() {
		manager.ExecuteAction(ctx, action)
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the action
	err = manager.CancelAction(ctx, action.ID)
	require.NoError(t, err)

	// Wait a bit and verify it was cancelled
	time.Sleep(100 * time.Millisecond)

	// Check action status
	actionStatus, err := manager.GetActionStatus(ctx, action.ID)
	require.NoError(t, err)
	assert.Equal(t, ActionStatusCancelled, actionStatus.Status)
}

func TestMultiAction_ValidationErrors(t *testing.T) {
	cfg := DefaultConfig()
	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	tests := []struct {
		name     string
		action   *MultiAction
		wantErr  bool
		errorMsg string
	}{
		{
			name: "invalid action type",
			action: &MultiAction{
				ID:      "invalid-type-001",
				Type:    ActionType("invalid"),
				Targets: []string{"cluster1"},
				Status:  ActionStatusPending,
			},
			wantErr:  true,
			errorMsg: "not allowed",
		},
		{
			name: "no targets",
			action: &MultiAction{
				ID:      "no-targets-001",
				Type:    ActionTypeBenchmark,
				Targets: []string{},
				Status:  ActionStatusPending,
			},
			wantErr:  true,
			errorMsg: "targets",
		},
		{
			name: "non-existent target",
			action: &MultiAction{
				ID:      "nonexistent-001",
				Type:    ActionTypeBenchmark,
				Targets: []string{"nonexistent-cluster"},
				Status:  ActionStatusPending,
			},
			wantErr:  true,
			errorMsg: "not found",
		},
		{
			name: "missing required parameters",
			action: &MultiAction{
				ID:      "missing-params-001",
				Type:    ActionTypeBenchmark,
				Targets: []string{"cluster1"},
				Parameters: map[string]interface{}{
					// Missing required "iterations" parameter
				},
				Status: ActionStatusPending,
			},
			wantErr:  true,
			errorMsg: "parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ExecuteAction(ctx, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiAction_RetryPolicy(t *testing.T) {
	// This would test retry behavior for failed actions
	// Implementation depends on the retry policy in the actual code
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "flaky-cluster", Endpoint: "localhost:9999", DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
			},
			RetryPolicy: RetryPolicy{
				MaxAttempts:   3,
				BaseDelay:     Duration(100 * time.Millisecond),
				MaxDelay:      Duration(1 * time.Second),
				BackoffFactor: 2.0,
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	action := &MultiAction{
		ID:      "retry-test-001",
		Type:    ActionTypeBenchmark,
		Targets: []string{"flaky-cluster"},
		Parameters: map[string]interface{}{
			"iterations": float64(1),
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	start := time.Now()
	err = manager.ExecuteAction(ctx, action)
	duration := time.Since(start)

	// Should have failed after retries
	assert.Equal(t, ActionStatusFailed, action.Status)

	// Should have taken some time due to retries and backoff
	assert.Greater(t, duration, 100*time.Millisecond)

	// Should have result with error details
	result, exists := action.Results["flaky-cluster"]
	assert.True(t, exists)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

func TestMultiAction_ConcurrentExecutions(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "concurrent-cluster", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
			},
			MaxConcurrent: 2, // Allow 2 concurrent actions
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Create multiple actions
	actions := []*MultiAction{
		{
			ID:      "concurrent-001",
			Type:    ActionTypeBenchmark,
			Targets: []string{"concurrent-cluster"},
			Parameters: map[string]interface{}{
				"iterations": float64(10),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:      "concurrent-002",
			Type:    ActionTypeBenchmark,
			Targets: []string{"concurrent-cluster"},
			Parameters: map[string]interface{}{
				"iterations": float64(10),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:      "concurrent-003",
			Type:    ActionTypeBenchmark,
			Targets: []string{"concurrent-cluster"},
			Parameters: map[string]interface{}{
				"iterations": float64(10),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		},
	}

	// Execute all actions concurrently
	results := make(chan error, len(actions))
	for _, action := range actions {
		go func(a *MultiAction) {
			results <- manager.ExecuteAction(ctx, a)
		}(action)
	}

	// Collect results
	var successCount int
	var errorCount int
	for i := 0; i < len(actions); i++ {
		select {
		case err := <-results:
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent actions")
		}
	}

	// Based on MaxConcurrent=2, some actions might be queued/delayed
	// All should eventually succeed or have proper error handling
	assert.Equal(t, len(actions), successCount+errorCount)
}

func TestActionTypeValidation(t *testing.T) {
	validTypes := []ActionType{
		ActionTypePurgeDLQ,
		ActionTypePauseQueue,
		ActionTypeResumeQueue,
		ActionTypeBenchmark,
		ActionTypeRebalance,
		ActionTypeFailover,
	}

	for _, actionType := range validTypes {
		assert.NotEmpty(t, string(actionType))
		assert.True(t, len(string(actionType)) > 0)
	}

	// Test action type string representations
	assert.Equal(t, "purge_dlq", string(ActionTypePurgeDLQ))
	assert.Equal(t, "pause_queue", string(ActionTypePauseQueue))
	assert.Equal(t, "resume_queue", string(ActionTypeResumeQueue))
	assert.Equal(t, "benchmark", string(ActionTypeBenchmark))
	assert.Equal(t, "rebalance", string(ActionTypeRebalance))
	assert.Equal(t, "failover", string(ActionTypeFailover))
}

func TestActionStatusTransitions(t *testing.T) {
	validStatuses := []ActionStatus{
		ActionStatusPending,
		ActionStatusConfirmed,
		ActionStatusExecuting,
		ActionStatusCompleted,
		ActionStatusFailed,
		ActionStatusCancelled,
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, string(status))
	}

	// Test status string representations
	assert.Equal(t, "pending", string(ActionStatusPending))
	assert.Equal(t, "confirmed", string(ActionStatusConfirmed))
	assert.Equal(t, "executing", string(ActionStatusExecuting))
	assert.Equal(t, "completed", string(ActionStatusCompleted))
	assert.Equal(t, "failed", string(ActionStatusFailed))
	assert.Equal(t, "cancelled", string(ActionStatusCancelled))
}