// Copyright 2025 James Ross
package dlqremediation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
)

func TestNewRemediationPipeline(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	if pipeline == nil {
		t.Fatal("Pipeline should not be nil")
	}

	state := pipeline.GetState()
	if state.Status != StatusStopped {
		t.Errorf("Expected status %s, got %s", StatusStopped, state.Status)
	}
}

func TestPipelineStartStop(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Start pipeline
	err = pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	state := pipeline.GetState()
	if state.Status != StatusRunning {
		t.Errorf("Expected status %s, got %s", StatusRunning, state.Status)
	}

	// Stop pipeline
	err = pipeline.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop pipeline: %v", err)
	}

	state = pipeline.GetState()
	if state.Status != StatusStopped {
		t.Errorf("Expected status %s, got %s", StatusStopped, state.Status)
	}
}

func TestPipelinePauseResume(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Start pipeline
	err = pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	// Pause pipeline
	err = pipeline.Pause(ctx)
	if err != nil {
		t.Fatalf("Failed to pause pipeline: %v", err)
	}

	state := pipeline.GetState()
	if state.Status != StatusPaused {
		t.Errorf("Expected status %s, got %s", StatusPaused, state.Status)
	}

	// Resume pipeline
	err = pipeline.Resume(ctx)
	if err != nil {
		t.Fatalf("Failed to resume pipeline: %v", err)
	}

	state = pipeline.GetState()
	if state.Status != StatusRunning {
		t.Errorf("Expected status %s, got %s", StatusRunning, state.Status)
	}

	// Stop pipeline
	pipeline.Stop(ctx)
}

func TestRuleManagement(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Create test rule
	rule := RemediationRule{
		Name:        "Test Rule",
		Description: "A test rule for validation errors",
		Priority:    100,
		Enabled:     true,
		Matcher: RuleMatcher{
			ErrorPattern: "validation.*failed",
			JobType:      "user_registration",
		},
		Actions: []Action{
			{
				Type: ActionRequeue,
				Parameters: map[string]interface{}{
					"target_queue": "user_registration_retry",
					"delay":        "5m",
				},
			},
		},
		Safety: SafetyLimits{
			MaxPerMinute:       10,
			MaxTotalPerRun:     100,
			ErrorRateThreshold: 0.1,
		},
		Tags: []string{"validation", "user"},
	}

	// Add rule
	err = pipeline.AddRule(ctx, rule, "test_user")
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// Get rules
	rules := pipeline.GetRules()
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	addedRule := rules[0]
	if addedRule.Name != rule.Name {
		t.Errorf("Expected rule name %s, got %s", rule.Name, addedRule.Name)
	}

	if addedRule.ID == "" {
		t.Error("Rule ID should not be empty")
	}

	// Update rule
	updatedRule := addedRule
	updatedRule.Description = "Updated description"
	err = pipeline.UpdateRule(ctx, addedRule.ID, updatedRule, "test_user")
	if err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}

	rules = pipeline.GetRules()
	if rules[0].Description != "Updated description" {
		t.Errorf("Expected updated description, got %s", rules[0].Description)
	}

	// Delete rule
	err = pipeline.DeleteRule(ctx, addedRule.ID, "test_user")
	if err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	rules = pipeline.GetRules()
	if len(rules) != 0 {
		t.Fatalf("Expected 0 rules after deletion, got %d", len(rules))
	}
}

func TestClassificationEngine(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Create test rules
	rules := []RemediationRule{
		{
			ID:      "rule1",
			Name:    "Validation Error Rule",
			Enabled: true,
			Priority: 100,
			Matcher: RuleMatcher{
				ErrorPattern: "validation.*failed",
				JobType:      "user_registration",
			},
		},
		{
			ID:      "rule2",
			Name:    "Timeout Error Rule",
			Enabled: true,
			Priority: 90,
			Matcher: RuleMatcher{
				ErrorType: "timeout",
				RetryCount: "> 2",
			},
		},
	}

	// Test jobs
	testJobs := []*DLQJob{
		{
			JobID:   "job1",
			JobType: "user_registration",
			Error:   "validation failed: email required",
			ErrorType: "validation_error",
			RetryCount: 1,
		},
		{
			JobID:   "job2",
			JobType: "payment_processing",
			Error:   "connection timeout",
			ErrorType: "timeout",
			RetryCount: 3,
		},
		{
			JobID:   "job3",
			JobType: "notification",
			Error:   "unknown error",
			ErrorType: "unknown",
			RetryCount: 0,
		},
	}

	// Test classification
	for i, job := range testJobs {
		classification, err := pipeline.classifier.Classify(ctx, job, rules)
		if err != nil {
			t.Fatalf("Classification failed for job %d: %v", i, err)
		}

		if classification == nil {
			t.Fatalf("Classification should not be nil for job %d", i)
		}

		// Job 1 should match rule1
		if i == 0 {
			if classification.RuleID != "rule1" {
				t.Errorf("Job 1 should match rule1, got %s", classification.RuleID)
			}
		}

		// Job 2 should match rule2
		if i == 1 {
			if classification.RuleID != "rule2" {
				t.Errorf("Job 2 should match rule2, got %s", classification.RuleID)
			}
		}

		// Job 3 should not match any rule
		if i == 2 {
			if classification.Category != "unclassified" {
				t.Errorf("Job 3 should be unclassified, got %s", classification.Category)
			}
		}
	}
}

func TestActionExecution(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Test job
	job := &DLQJob{
		JobID:   "test_job",
		JobType: "user_registration",
		Queue:   "user_queue",
		Payload: json.RawMessage(`{"user_id": 123, "email": "test@example.com", "password": "secret"}`),
		Error:   "validation failed",
		ErrorType: "validation_error",
		PayloadSize: 100,
	}

	tests := []struct {
		name    string
		actions []Action
		dryRun  bool
		wantErr bool
	}{
		{
			name: "requeue action",
			actions: []Action{
				{
					Type: ActionRequeue,
					Parameters: map[string]interface{}{
						"target_queue": "retry_queue",
						"delay":        "5m",
					},
				},
			},
			dryRun:  true,
			wantErr: false,
		},
		{
			name: "transform action",
			actions: []Action{
				{
					Type: ActionTransform,
					Parameters: map[string]interface{}{
						"set": map[string]interface{}{
							"processed_at": "2024-01-15T10:00:00Z",
						},
						"remove": []interface{}{"password"},
					},
				},
			},
			dryRun:  true,
			wantErr: false,
		},
		{
			name: "redact action",
			actions: []Action{
				{
					Type: ActionRedact,
					Parameters: map[string]interface{}{
						"fields":      []interface{}{"password", "email"},
						"replacement": "[REDACTED]",
					},
				},
			},
			dryRun:  true,
			wantErr: false,
		},
		{
			name: "drop action",
			actions: []Action{
				{
					Type: ActionDrop,
					Parameters: map[string]interface{}{
						"reason": "Test drop",
					},
				},
			},
			dryRun:  true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pipeline.actionExecutor.Execute(ctx, job, tt.actions, tt.dryRun)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if result.DryRun != tt.dryRun {
				t.Errorf("Expected dry_run %v, got %v", tt.dryRun, result.DryRun)
			}

			if len(result.Actions) != len(tt.actions) {
				t.Errorf("Expected %d actions, got %d", len(tt.actions), len(result.Actions))
			}
		})
	}
}

func TestBatchProcessing(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := testConfig(mr.Addr())
	pipeline, err := NewRemediationPipeline(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	defer pipeline.Close()

	ctx := context.Background()

	// Add test jobs to DLQ
	testJobs := []DLQJob{
		{
			ID:      "dlq_job_1",
			JobID:   "job_1",
			JobType: "user_registration",
			Error:   "validation failed",
			ErrorType: "validation_error",
			Payload: json.RawMessage(`{"user_id": 1}`),
		},
		{
			ID:      "dlq_job_2",
			JobID:   "job_2",
			JobType: "payment",
			Error:   "timeout",
			ErrorType: "timeout",
			Payload: json.RawMessage(`{"amount": 100}`),
		},
	}

	// Add jobs to Redis DLQ
	for _, job := range testJobs {
		jobData, _ := json.Marshal(job)
		mr.Lpush(config.Storage.DLQStreamKey, string(jobData))
	}

	// Add a test rule
	rule := RemediationRule{
		Name:    "Test Rule",
		Enabled: true,
		Priority: 100,
		Matcher: RuleMatcher{
			ErrorPattern: "validation.*failed",
		},
		Actions: []Action{
			{
				Type: ActionRequeue,
				Parameters: map[string]interface{}{
					"target_queue": "retry_queue",
				},
			},
		},
		Safety: SafetyLimits{
			MaxPerMinute:   100,
			MaxTotalPerRun: 1000,
		},
	}

	err = pipeline.AddRule(ctx, rule, "test_user")
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// Process batch
	result, err := pipeline.ProcessBatch(ctx, true) // Dry run
	if err != nil {
		t.Fatalf("Batch processing failed: %v", err)
	}

	if result.TotalJobs != 2 {
		t.Errorf("Expected 2 total jobs, got %d", result.TotalJobs)
	}

	if result.ProcessedJobs == 0 {
		t.Error("Expected some jobs to be processed")
	}

	if len(result.Results) == 0 {
		t.Error("Expected some results")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := &RateLimiter{
		MaxPerMinute: 5,
		MaxTotal:     10,
		BurstSize:    2,
	}

	// Should allow initial requests
	for i := 0; i < 5; i++ {
		if !rl.CanProcess() {
			t.Errorf("Should allow request %d", i+1)
		}
		rl.RecordProcessed()
	}

	// Should deny after hitting per-minute limit
	if rl.CanProcess() {
		t.Error("Should deny request after hitting per-minute limit")
	}

	// Reset to new minute
	rl.CurrentMinute = time.Now().Add(-time.Minute)
	rl.CountThisMinute = 0

	// Should allow again
	for i := 0; i < 5; i++ {
		if !rl.CanProcess() {
			t.Errorf("Should allow request %d after minute reset", i+1)
		}
		rl.RecordProcessed()
	}

	// Should deny after hitting total limit
	if rl.CanProcess() {
		t.Error("Should deny request after hitting total limit")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := &CircuitBreaker{
		ErrorThreshold:  0.5,
		MinRequests:     5,
		RecoveryTimeout: time.Minute,
		State:          CircuitClosed,
	}

	// Should allow initial requests
	if !cb.CanExecute() {
		t.Error("Should allow initial request")
	}

	// Record successes
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}

	// Record failures to trip circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Should be open now
	if cb.State != CircuitOpen {
		t.Errorf("Expected circuit to be open, got %s", cb.State)
	}

	// Should deny requests when open
	if cb.CanExecute() {
		t.Error("Should deny request when circuit is open")
	}

	// Simulate recovery timeout
	cb.LastFailure = time.Now().Add(-2 * time.Minute)

	// Should allow one request (half-open)
	if !cb.CanExecute() {
		t.Error("Should allow request after recovery timeout")
	}

	if cb.State != CircuitHalfOpen {
		t.Errorf("Expected circuit to be half-open, got %s", cb.State)
	}

	// Record success to close circuit
	cb.RecordSuccess()

	if cb.State != CircuitClosed {
		t.Errorf("Expected circuit to be closed after success, got %s", cb.State)
	}
}

func TestIdempotencyTracker(t *testing.T) {
	tracker := NewIdempotencyTracker(time.Minute)

	jobID := "test_job_123"

	// Should not be processed initially
	if tracker.IsProcessed(jobID) {
		t.Error("Job should not be processed initially")
	}

	// Mark as processed
	tracker.MarkProcessed(jobID)

	// Should be processed now
	if !tracker.IsProcessed(jobID) {
		t.Error("Job should be processed after marking")
	}

	// Simulate expiry
	tracker.ProcessedJobs[jobID] = time.Now().Add(-2 * time.Minute)

	// Should not be processed after expiry
	if tracker.IsProcessed(jobID) {
		t.Error("Job should not be processed after expiry")
	}

	// Test cleanup
	tracker.ProcessedJobs["old_job"] = time.Now().Add(-2 * time.Minute)
	tracker.ProcessedJobs["new_job"] = time.Now()

	tracker.Cleanup()

	if _, exists := tracker.ProcessedJobs["old_job"]; exists {
		t.Error("Old job should be cleaned up")
	}

	if _, exists := tracker.ProcessedJobs["new_job"]; !exists {
		t.Error("New job should not be cleaned up")
	}
}

// Helper function to create test configuration
func testConfig(redisAddr string) *Config {
	config := DefaultConfig()
	config.Redis.Addr = redisAddr
	config.Pipeline.Enabled = true
	config.Pipeline.DryRun = true
	config.Pipeline.BatchSize = 10
	config.Pipeline.PollInterval = time.Second
	return config
}