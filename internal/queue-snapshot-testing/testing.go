// Copyright 2025 James Ross
package queuesnapshotesting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TestHelper provides snapshot testing utilities for Go tests
type TestHelper struct {
	manager      *SnapshotManager
	updateMode   bool
	snapshotDir  string
	testName     string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T, redis *redis.Client) *TestHelper {
	// Check for update mode
	updateMode := os.Getenv("UPDATE_SNAPSHOTS") == "true"

	// Create test-specific snapshot directory
	snapshotDir := filepath.Join("testdata", "snapshots", t.Name())
	os.MkdirAll(snapshotDir, 0755)

	config := &SnapshotConfig{
		StoragePath:      snapshotDir,
		IgnoreTimestamps: true,
		IgnoreIDs:        true,
		IgnoreWorkerIDs:  true,
		CompressLevel:    0, // No compression for test snapshots
	}

	logger := zap.NewNop()
	manager, _ := NewSnapshotManager(config, redis, logger)

	return &TestHelper{
		manager:     manager,
		updateMode:  updateMode,
		snapshotDir: snapshotDir,
		testName:    t.Name(),
	}
}

// AssertSnapshot asserts that the current state matches a snapshot
func (th *TestHelper) AssertSnapshot(t *testing.T, name string) {
	t.Helper()

	ctx := context.Background()
	snapshotID := th.getSnapshotID(name)

	if th.updateMode {
		// Update mode: capture and save new snapshot
		snapshot, err := th.manager.CaptureSnapshot(ctx, name, "Test snapshot", []string{"test"})
		if err != nil {
			t.Fatalf("Failed to capture snapshot: %v", err)
		}

		// Save with deterministic ID
		snapshot.ID = snapshotID
		if err := th.manager.storage.Save(snapshot); err != nil {
			t.Fatalf("Failed to save snapshot: %v", err)
		}

		t.Logf("Updated snapshot: %s", name)
		return
	}

	// Normal mode: compare with existing snapshot
	if !th.manager.storage.Exists(snapshotID) {
		t.Fatalf("Snapshot '%s' does not exist. Run with UPDATE_SNAPSHOTS=true to create.", name)
	}

	result, err := th.manager.AssertSnapshot(ctx, snapshotID)
	if err != nil {
		t.Fatalf("Failed to assert snapshot: %v", err)
	}

	if !result.Passed {
		// Format differences for display
		diff := th.formatDifferences(result.Differences)
		t.Errorf("Snapshot assertion failed:\n%s\n\nDifferences:\n%s\n\nRun with UPDATE_SNAPSHOTS=true to update.",
			result.Message, diff)
	}
}

// CaptureSnapshot captures a named snapshot
func (th *TestHelper) CaptureSnapshot(t *testing.T, name string) *Snapshot {
	t.Helper()

	ctx := context.Background()
	snapshot, err := th.manager.CaptureSnapshot(ctx, name, "Test snapshot", []string{"test"})
	if err != nil {
		t.Fatalf("Failed to capture snapshot: %v", err)
	}

	return snapshot
}

// CompareSnapshots compares two named snapshots
func (th *TestHelper) CompareSnapshots(t *testing.T, leftName, rightName string) *DiffResult {
	t.Helper()

	leftID := th.getSnapshotID(leftName)
	rightID := th.getSnapshotID(rightName)

	diff, err := th.manager.CompareSnapshots(leftID, rightID)
	if err != nil {
		t.Fatalf("Failed to compare snapshots: %v", err)
	}

	return diff
}

// RestoreSnapshot restores a named snapshot
func (th *TestHelper) RestoreSnapshot(t *testing.T, name string) {
	t.Helper()

	ctx := context.Background()
	snapshotID := th.getSnapshotID(name)

	if err := th.manager.RestoreSnapshot(ctx, snapshotID); err != nil {
		t.Fatalf("Failed to restore snapshot: %v", err)
	}
}

// Cleanup removes test snapshots
func (th *TestHelper) Cleanup() {
	// Keep snapshots for debugging unless explicitly requested
	if os.Getenv("CLEANUP_SNAPSHOTS") == "true" {
		os.RemoveAll(th.snapshotDir)
	}
}

// Helper methods

func (th *TestHelper) getSnapshotID(name string) string {
	// Generate deterministic ID from test name and snapshot name
	return fmt.Sprintf("%s_%s", th.sanitizeName(th.testName), th.sanitizeName(name))
}

func (th *TestHelper) sanitizeName(name string) string {
	// Replace special characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		" ", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

func (th *TestHelper) formatDifferences(differences []Change) string {
	if len(differences) == 0 {
		return "No differences"
	}

	var b strings.Builder
	for i, diff := range differences {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, diff.Description))
		b.WriteString(fmt.Sprintf("   Path: %s\n", diff.Path))
		b.WriteString(fmt.Sprintf("   Type: %s\n", diff.Type))
		if diff.OldValue != nil {
			b.WriteString(fmt.Sprintf("   Old: %v\n", diff.OldValue))
		}
		if diff.NewValue != nil {
			b.WriteString(fmt.Sprintf("   New: %v\n", diff.NewValue))
		}
		if diff.Impact != "" {
			b.WriteString(fmt.Sprintf("   Impact: %s\n", diff.Impact))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// SnapshotMatcher provides custom matchers for snapshot testing
type SnapshotMatcher struct {
	config *SnapshotConfig
}

// NewSnapshotMatcher creates a new snapshot matcher
func NewSnapshotMatcher() *SnapshotMatcher {
	return &SnapshotMatcher{
		config: &SnapshotConfig{
			IgnoreTimestamps: true,
			IgnoreIDs:        true,
		},
	}
}

// IgnoreTimestamps configures the matcher to ignore timestamps
func (sm *SnapshotMatcher) IgnoreTimestamps(ignore bool) *SnapshotMatcher {
	sm.config.IgnoreTimestamps = ignore
	return sm
}

// IgnoreIDs configures the matcher to ignore IDs
func (sm *SnapshotMatcher) IgnoreIDs(ignore bool) *SnapshotMatcher {
	sm.config.IgnoreIDs = ignore
	return sm
}

// WithCustomIgnores adds custom ignore patterns
func (sm *SnapshotMatcher) WithCustomIgnores(patterns ...string) *SnapshotMatcher {
	sm.config.CustomIgnores = append(sm.config.CustomIgnores, patterns...)
	return sm
}

// Match performs the matching
func (sm *SnapshotMatcher) Match(expected, actual *Snapshot) (bool, string) {
	differ := NewDiffer(sm.config)
	diff, err := differ.Compare(expected, actual)
	if err != nil {
		return false, fmt.Sprintf("Failed to compare: %v", err)
	}

	if diff.TotalChanges == 0 {
		return true, "Snapshots match"
	}

	return false, sm.formatDiff(diff)
}

func (sm *SnapshotMatcher) formatDiff(diff *DiffResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Found %d differences:\n", diff.TotalChanges))
	b.WriteString(fmt.Sprintf("  Added: %d\n", diff.Added))
	b.WriteString(fmt.Sprintf("  Removed: %d\n", diff.Removed))
	b.WriteString(fmt.Sprintf("  Modified: %d\n", diff.Modified))

	if len(diff.SemanticChanges) > 0 {
		b.WriteString("\nSemantic Changes:\n")
		for _, sc := range diff.SemanticChanges {
			b.WriteString(fmt.Sprintf("  - %s: %s (severity: %s)\n",
				sc.Type, sc.Description, sc.Severity))
		}
	}

	return b.String()
}

// Fixtures provides pre-defined snapshot scenarios
type Fixtures struct {
	manager *SnapshotManager
}

// NewFixtures creates a new fixtures helper
func NewFixtures(manager *SnapshotManager) *Fixtures {
	return &Fixtures{
		manager: manager,
	}
}

// LoadScenario loads a predefined scenario
func (f *Fixtures) LoadScenario(ctx context.Context, scenario string) error {
	switch scenario {
	case "empty":
		return f.loadEmptyScenario(ctx)
	case "simple":
		return f.loadSimpleScenario(ctx)
	case "complex":
		return f.loadComplexScenario(ctx)
	case "error":
		return f.loadErrorScenario(ctx)
	default:
		return fmt.Errorf("unknown scenario: %s", scenario)
	}
}

func (f *Fixtures) loadEmptyScenario(ctx context.Context) error {
	// Clear all state
	return f.manager.clearCurrentState(ctx)
}

func (f *Fixtures) loadSimpleScenario(ctx context.Context) error {
	// Create a simple queue with a few jobs
	redis := f.manager.redis

	// Create queue
	redis.Del(ctx, "queue:simple")

	// Add jobs
	for i := 0; i < 5; i++ {
		job := JobState{
			ID:        fmt.Sprintf("job-%d", i),
			QueueName: "simple",
			Status:    "pending",
			CreatedAt: time.Now(),
			Payload: map[string]interface{}{
				"task": fmt.Sprintf("task-%d", i),
			},
		}

		jobData, _ := json.Marshal(job)
		redis.RPush(ctx, "queue:simple", string(jobData))
	}

	// Add a worker
	redis.HSet(ctx, "worker:worker-1", map[string]interface{}{
		"status":    "active",
		"last_seen": time.Now().Format(time.RFC3339),
	})

	return nil
}

func (f *Fixtures) loadComplexScenario(ctx context.Context) error {
	redis := f.manager.redis

	// Create multiple queues
	queues := []string{"high-priority", "normal", "low-priority", "dead-letter"}

	for _, queue := range queues {
		key := fmt.Sprintf("queue:%s", queue)
		redis.Del(ctx, key)

		// Add varying numbers of jobs
		jobCount := 10 + len(queue)
		for i := 0; i < jobCount; i++ {
			job := JobState{
				ID:        fmt.Sprintf("%s-job-%d", queue, i),
				QueueName: queue,
				Status:    "pending",
				Priority:  len(queue),
				CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
				Payload: map[string]interface{}{
					"queue": queue,
					"index": i,
				},
			}

			jobData, _ := json.Marshal(job)
			redis.RPush(ctx, key, string(jobData))
		}

		// Add queue config
		configKey := fmt.Sprintf("queue:config:%s", queue)
		redis.HSet(ctx, configKey, map[string]interface{}{
			"max_retries":    "3",
			"timeout":        "300",
			"rate_limit":     "100",
		})
	}

	// Add multiple workers
	for i := 0; i < 5; i++ {
		workerKey := fmt.Sprintf("worker:worker-%d", i)
		status := "idle"
		if i < 2 {
			status = "active"
		}

		redis.HSet(ctx, workerKey, map[string]interface{}{
			"status":          status,
			"last_seen":       time.Now().Format(time.RFC3339),
			"processed_count": fmt.Sprintf("%d", i*100),
			"error_count":     fmt.Sprintf("%d", i*5),
		})
	}

	// Add metrics
	redis.Set(ctx, "metrics:total_processed", "1000", 0)
	redis.Set(ctx, "metrics:total_failed", "50", 0)
	redis.Set(ctx, "metrics:avg_latency", "250", 0)

	return nil
}

func (f *Fixtures) loadErrorScenario(ctx context.Context) error {
	redis := f.manager.redis

	// Create queue with failed jobs
	redis.Del(ctx, "queue:error-test")

	for i := 0; i < 10; i++ {
		status := "failed"
		if i%3 == 0 {
			status = "pending"
		}

		job := JobState{
			ID:        fmt.Sprintf("error-job-%d", i),
			QueueName: "error-test",
			Status:    status,
			Attempts:  3,
			MaxRetries: 3,
			Error:     "simulated error",
			CreatedAt: time.Now(),
			Payload: map[string]interface{}{
				"error": true,
			},
		}

		jobData, _ := json.Marshal(job)
		redis.RPush(ctx, "queue:error-test", string(jobData))
	}

	// Add error metrics
	redis.Set(ctx, "metrics:total_failed", "100", 0)
	redis.Set(ctx, "metrics:error_rate", "0.1", 0)

	return nil
}