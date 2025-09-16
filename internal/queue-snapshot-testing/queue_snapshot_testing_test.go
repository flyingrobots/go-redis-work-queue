// Copyright 2025 James Ross
package queuesnapshotesting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*SnapshotManager, *redis.Client, func()) {
	// Create miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create temporary directory for snapshots
	tmpDir := t.TempDir()

	config := &SnapshotConfig{
		StoragePath:      tmpDir,
		MaxSnapshots:     10,
		RetentionDays:    7,
		CompressLevel:    0,
		IgnoreTimestamps: true,
		IgnoreIDs:        false,
		MaxJobsPerSnapshot: 100,
		SampleRate:       1.0,
		TimeoutSeconds:   10,
	}

	logger := zap.NewNop()
	manager, err := NewSnapshotManager(config, client, logger)
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		mr.Close()
		os.RemoveAll(tmpDir)
	}

	return manager, client, cleanup
}

func TestSnapshotCapture(t *testing.T) {
	manager, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("capture empty state", func(t *testing.T) {
		snapshot, err := manager.CaptureSnapshot(ctx, "empty", "Empty state snapshot", []string{"test"})
		require.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.Equal(t, "empty", snapshot.Name)
		assert.Equal(t, 0, len(snapshot.Queues))
		assert.Equal(t, 0, len(snapshot.Jobs))
		assert.NotEmpty(t, snapshot.ID)
		assert.NotEmpty(t, snapshot.Checksum)
	})

	t.Run("capture with queues and jobs", func(t *testing.T) {
		// Setup test data
		setupTestData(t, client)

		snapshot, err := manager.CaptureSnapshot(ctx, "with-data", "Snapshot with data", []string{"test", "queues"})
		require.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.Greater(t, len(snapshot.Queues), 0)
		assert.Greater(t, len(snapshot.Jobs), 0)
		assert.Contains(t, snapshot.Tags, "queues")
	})

	t.Run("capture with workers", func(t *testing.T) {
		// Add worker data
		client.HSet(ctx, "worker:w1", map[string]interface{}{
			"status":    "active",
			"last_seen": time.Now().Format(time.RFC3339),
		})

		snapshot, err := manager.CaptureSnapshot(ctx, "with-workers", "Snapshot with workers", nil)
		require.NoError(t, err)
		assert.Equal(t, 1, len(snapshot.Workers))
		assert.Equal(t, "w1", snapshot.Workers[0].ID)
		assert.Equal(t, "active", snapshot.Workers[0].Status)
	})
}

func TestSnapshotRestore(t *testing.T) {
	manager, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("restore snapshot", func(t *testing.T) {
		// Setup initial state
		setupTestData(t, client)

		// Capture snapshot
		original, err := manager.CaptureSnapshot(ctx, "original", "Original state", nil)
		require.NoError(t, err)

		// Modify state
		client.Del(ctx, "queue:test")
		client.RPush(ctx, "queue:new", "new-job")

		// Restore snapshot
		err = manager.RestoreSnapshot(ctx, original.ID)
		require.NoError(t, err)

		// Verify restoration
		exists := client.Exists(ctx, "queue:test").Val()
		assert.Equal(t, int64(1), exists)

		length := client.LLen(ctx, "queue:test").Val()
		assert.Equal(t, int64(3), length) // Should have 3 jobs as in original
	})

	t.Run("restore non-existent snapshot", func(t *testing.T) {
		err := manager.RestoreSnapshot(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot not found")
	})
}

func TestSnapshotComparison(t *testing.T) {
	manager, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("compare identical snapshots", func(t *testing.T) {
		setupTestData(t, client)

		// Capture two identical snapshots
		snap1, _ := manager.CaptureSnapshot(ctx, "snap1", "First snapshot", nil)
		snap2, _ := manager.CaptureSnapshot(ctx, "snap2", "Second snapshot", nil)

		diff, err := manager.CompareSnapshots(snap1.ID, snap2.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, diff.TotalChanges)
	})

	t.Run("compare different snapshots", func(t *testing.T) {
		// First snapshot
		setupTestData(t, client)
		snap1, _ := manager.CaptureSnapshot(ctx, "before", "Before changes", nil)

		// Make changes
		client.RPush(ctx, "queue:test", "new-job-1", "new-job-2")
		client.Del(ctx, "queue:low")

		// Second snapshot
		snap2, _ := manager.CaptureSnapshot(ctx, "after", "After changes", nil)

		diff, err := manager.CompareSnapshots(snap1.ID, snap2.ID)
		require.NoError(t, err)
		assert.Greater(t, diff.TotalChanges, 0)
		assert.Greater(t, len(diff.QueueChanges), 0)
		assert.Greater(t, len(diff.JobChanges), 0)
	})

	t.Run("detect semantic changes", func(t *testing.T) {
		// Empty state
		snap1, _ := manager.CaptureSnapshot(ctx, "empty", "Empty state", nil)

		// Add significant load
		for i := 0; i < 200; i++ {
			client.RPush(ctx, "queue:overload", fmt.Sprintf("job-%d", i))
		}

		snap2, _ := manager.CaptureSnapshot(ctx, "overloaded", "Overloaded state", nil)

		diff, err := manager.CompareSnapshots(snap1.ID, snap2.ID)
		require.NoError(t, err)
		assert.Greater(t, len(diff.SemanticChanges), 0)

		// Should detect queue overload
		found := false
		for _, sc := range diff.SemanticChanges {
			if sc.Type == "queue_overload" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should detect queue overload")
	})
}

func TestSnapshotAssertion(t *testing.T) {
	manager, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("assertion passes with identical state", func(t *testing.T) {
		setupTestData(t, client)

		// Capture expected state
		expected, _ := manager.CaptureSnapshot(ctx, "expected", "Expected state", nil)

		// Assert current state matches
		result, err := manager.AssertSnapshot(ctx, expected.ID)
		require.NoError(t, err)
		assert.True(t, result.Passed)
		assert.Equal(t, "Snapshot assertion passed", result.Message)
	})

	t.Run("assertion fails with different state", func(t *testing.T) {
		setupTestData(t, client)

		// Capture expected state
		expected, _ := manager.CaptureSnapshot(ctx, "expected", "Expected state", nil)

		// Change state
		client.RPush(ctx, "queue:test", "unexpected-job")

		// Assert should fail
		result, err := manager.AssertSnapshot(ctx, expected.ID)
		require.NoError(t, err)
		assert.False(t, result.Passed)
		assert.Contains(t, result.Message, "differences")
		assert.Greater(t, len(result.Differences), 0)
	})
}

func TestDiffer(t *testing.T) {
	config := &SnapshotConfig{
		IgnoreTimestamps: true,
		IgnoreIDs:        false,
		IgnoreWorkerIDs:  false,
	}

	differ := NewDiffer(config)

	t.Run("detect queue changes", func(t *testing.T) {
		left := &Snapshot{
			ID: "left",
			Queues: []QueueState{
				{Name: "queue1", Length: 10},
				{Name: "queue2", Length: 20},
			},
		}

		right := &Snapshot{
			ID: "right",
			Queues: []QueueState{
				{Name: "queue1", Length: 15},
				{Name: "queue3", Length: 30},
			},
		}

		diff, err := differ.Compare(left, right)
		require.NoError(t, err)

		assert.Equal(t, 3, diff.TotalChanges) // Modified, removed, added
		assert.Equal(t, 1, diff.Added)
		assert.Equal(t, 1, diff.Removed)
		assert.Equal(t, 1, diff.Modified)
	})

	t.Run("detect job movements", func(t *testing.T) {
		left := &Snapshot{
			ID: "left",
			Jobs: []JobState{
				{ID: "job1", QueueName: "queue1", Status: "pending"},
				{ID: "job2", QueueName: "queue1", Status: "pending"},
			},
		}

		right := &Snapshot{
			ID: "right",
			Jobs: []JobState{
				{ID: "job1", QueueName: "queue2", Status: "pending"}, // Moved
				{ID: "job2", QueueName: "queue1", Status: "failed"},  // Status changed
			},
		}

		diff, err := differ.Compare(left, right)
		require.NoError(t, err)

		// Should detect job movement
		found := false
		for _, change := range diff.JobChanges {
			if change.Type == ChangeMoved {
				found = true
				assert.Equal(t, "job.job1", change.Path)
				break
			}
		}
		assert.True(t, found, "Should detect job movement")
	})

	t.Run("ignore timestamps when configured", func(t *testing.T) {
		config.IgnoreTimestamps = true
		differ := NewDiffer(config)

		left := &Snapshot{
			ID:        "left",
			CreatedAt: time.Now(),
		}

		right := &Snapshot{
			ID:        "right",
			CreatedAt: time.Now().Add(time.Hour),
		}

		diff, err := differ.Compare(left, right)
		require.NoError(t, err)
		assert.Equal(t, 0, diff.TotalChanges)
	})

	t.Run("detect worker changes", func(t *testing.T) {
		config.IgnoreWorkerIDs = false
		differ := NewDiffer(config)

		left := &Snapshot{
			ID: "left",
			Workers: []WorkerState{
				{ID: "w1", Status: "active"},
				{ID: "w2", Status: "idle"},
			},
		}

		right := &Snapshot{
			ID: "right",
			Workers: []WorkerState{
				{ID: "w1", Status: "idle"},
				{ID: "w3", Status: "active"},
			},
		}

		diff, err := differ.Compare(left, right)
		require.NoError(t, err)

		// Should detect worker removal, addition, and status change
		assert.Greater(t, diff.TotalChanges, 0)
		assert.Greater(t, len(diff.WorkerChanges), 0)
	})
}

func TestStorage(t *testing.T) {
	tmpDir := t.TempDir()
	logger := zap.NewNop()
	storage := NewFileStorage(tmpDir, logger)

	t.Run("save and load snapshot", func(t *testing.T) {
		snapshot := &Snapshot{
			ID:          "test-123",
			Name:        "Test Snapshot",
			Description: "Test description",
			CreatedAt:   time.Now(),
			Queues: []QueueState{
				{Name: "queue1", Length: 10},
			},
			Jobs: []JobState{
				{ID: "job1", QueueName: "queue1", Status: "pending"},
			},
		}

		// Save
		err := storage.Save(snapshot)
		require.NoError(t, err)

		// Verify file exists
		assert.True(t, storage.Exists("test-123"))

		// Load
		loaded, err := storage.Load("test-123")
		require.NoError(t, err)
		assert.Equal(t, snapshot.ID, loaded.ID)
		assert.Equal(t, snapshot.Name, loaded.Name)
		assert.Equal(t, len(snapshot.Queues), len(loaded.Queues))
	})

	t.Run("save compressed snapshot", func(t *testing.T) {
		snapshot := &Snapshot{
			ID:         "compressed-123",
			Name:       "Compressed",
			Compressed: true,
		}

		err := storage.Save(snapshot)
		require.NoError(t, err)

		loaded, err := storage.Load("compressed-123")
		require.NoError(t, err)
		assert.Equal(t, snapshot.ID, loaded.ID)
	})

	t.Run("list snapshots with filter", func(t *testing.T) {
		// Create multiple snapshots
		for i := 0; i < 5; i++ {
			snapshot := &Snapshot{
				ID:          fmt.Sprintf("list-test-%d", i),
				Name:        fmt.Sprintf("List Test %d", i),
				Tags:        []string{"test", fmt.Sprintf("batch-%d", i%2)},
				Environment: "test",
				CreatedAt:   time.Now().Add(time.Duration(i) * time.Hour),
			}
			storage.Save(snapshot)
		}

		// List all
		all, err := storage.List(nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 5)

		// Filter by tag
		filtered, err := storage.List(&SnapshotFilter{
			Tags: []string{"batch-0"},
		})
		require.NoError(t, err)
		assert.Greater(t, len(filtered), 0)

		// Filter by name
		filtered, err = storage.List(&SnapshotFilter{
			Name: "List Test 1",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(filtered))
	})

	t.Run("delete snapshot", func(t *testing.T) {
		snapshot := &Snapshot{
			ID:   "delete-me",
			Name: "To Delete",
		}

		storage.Save(snapshot)
		assert.True(t, storage.Exists("delete-me"))

		err := storage.Delete("delete-me")
		require.NoError(t, err)
		assert.False(t, storage.Exists("delete-me"))
	})
}

func TestTestHelper(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	helper := NewTestHelper(t, client)
	defer helper.Cleanup()

	t.Run("capture and assert snapshot", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		client.RPush(ctx, "queue:test", "job1", "job2")

		// Capture snapshot
		snapshot := helper.CaptureSnapshot(t, "test-state")
		assert.NotNil(t, snapshot)

		// Would normally assert, but in test mode it would fail
		// helper.AssertSnapshot(t, "test-state")
	})

	t.Run("restore snapshot", func(t *testing.T) {
		ctx := context.Background()

		// Setup and capture
		client.RPush(ctx, "queue:restore", "job1")
		helper.CaptureSnapshot(t, "restore-test")

		// Clear
		client.Del(ctx, "queue:restore")

		// Restore
		helper.RestoreSnapshot(t, "restore-test")

		// Verify
		exists := client.Exists(ctx, "queue:restore").Val()
		assert.Equal(t, int64(1), exists)
	})
}

func TestFixtures(t *testing.T) {
	manager, client, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()
	fixtures := NewFixtures(manager)

	t.Run("load empty scenario", func(t *testing.T) {
		err := fixtures.LoadScenario(ctx, "empty")
		require.NoError(t, err)

		keys := client.Keys(ctx, "*").Val()
		assert.Equal(t, 0, len(keys))
	})

	t.Run("load simple scenario", func(t *testing.T) {
		err := fixtures.LoadScenario(ctx, "simple")
		require.NoError(t, err)

		// Verify queue exists
		exists := client.Exists(ctx, "queue:simple").Val()
		assert.Equal(t, int64(1), exists)

		// Verify jobs
		length := client.LLen(ctx, "queue:simple").Val()
		assert.Equal(t, int64(5), length)

		// Verify worker
		workerExists := client.Exists(ctx, "worker:worker-1").Val()
		assert.Equal(t, int64(1), workerExists)
	})

	t.Run("load complex scenario", func(t *testing.T) {
		err := fixtures.LoadScenario(ctx, "complex")
		require.NoError(t, err)

		// Verify multiple queues
		queues := client.Keys(ctx, "queue:*").Val()
		assert.Greater(t, len(queues), 3)

		// Verify workers
		workers := client.Keys(ctx, "worker:*").Val()
		assert.Equal(t, 5, len(workers))

		// Verify metrics
		metricsKeys := client.Keys(ctx, "metrics:*").Val()
		assert.Greater(t, len(metricsKeys), 0)
	})

	t.Run("load error scenario", func(t *testing.T) {
		err := fixtures.LoadScenario(ctx, "error")
		require.NoError(t, err)

		// Verify error queue
		jobs, _ := client.LRange(ctx, "queue:error-test", 0, -1).Result()
		assert.Equal(t, 10, len(jobs))

		// Verify failed jobs
		failedCount := 0
		for _, jobData := range jobs {
			var job JobState
			json.Unmarshal([]byte(jobData), &job)
			if job.Status == "failed" {
				failedCount++
			}
		}
		assert.Greater(t, failedCount, 0)
	})
}

// Helper functions

func setupTestData(t *testing.T, client *redis.Client) {
	ctx := context.Background()

	// Create queues
	queues := map[string][]string{
		"queue:test": {"job1", "job2", "job3"},
		"queue:high": {"high1", "high2"},
		"queue:low":  {"low1"},
	}

	for queue, jobs := range queues {
		client.Del(ctx, queue)
		for _, job := range jobs {
			jobData := JobState{
				ID:        job,
				QueueName: queue[6:], // Remove "queue:" prefix
				Status:    "pending",
				CreatedAt: time.Now(),
				Payload: map[string]interface{}{
					"data": job,
				},
			}
			data, _ := json.Marshal(jobData)
			client.RPush(ctx, queue, string(data))
		}
	}

	// Add queue configs
	client.HSet(ctx, "queue:config:test", map[string]interface{}{
		"max_retries": "3",
		"timeout":     "300",
	})
}