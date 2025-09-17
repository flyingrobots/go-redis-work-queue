package dlqremediationui

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   1, // Use different DB for tests
	})

	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}

	client.FlushDB(ctx)

	t.Cleanup(func() {
		client.FlushDB(ctx)
		client.Close()
	})

	return client
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))
}

func createTestDLQEntry(id, queue, jobType, errorMsg string) DLQEntry {
	return DLQEntry{
		ID:      id,
		JobID:   "job-" + id,
		Type:    jobType,
		Queue:   queue,
		Payload: json.RawMessage(`{"data": "test payload"}`),
		Error: ErrorDetails{
			Message: errorMsg,
			Code:    "TEST_ERROR",
			Context: map[string]interface{}{
				"test": true,
			},
		},
		Metadata: JobMetadata{
			Source:      "test",
			SubmittedAt: time.Now().Add(-time.Hour),
			Priority:    1,
		},
		Attempts: []AttemptRecord{
			{
				AttemptNumber: 1,
				StartedAt:     time.Now().Add(-30 * time.Minute),
				FailedAt:      time.Now().Add(-29 * time.Minute),
				Error:         errorMsg,
				WorkerID:      "worker-1",
			},
		},
		CreatedAt: time.Now().Add(-time.Hour),
		FailedAt:  time.Now().Add(-29 * time.Minute),
	}
}

func seedTestData(t *testing.T, client *redis.Client) {
	ctx := context.Background()
	dlqKey := "dlq:entries"

	entries := []DLQEntry{
		createTestDLQEntry("entry1", "queue1", "job_type_a", "Connection timeout"),
		createTestDLQEntry("entry2", "queue1", "job_type_a", "Connection timeout"),
		createTestDLQEntry("entry3", "queue1", "job_type_b", "Validation error"),
		createTestDLQEntry("entry4", "queue2", "job_type_a", "Database connection failed"),
		createTestDLQEntry("entry5", "queue2", "job_type_c", "Authentication failed"),
	}

	for _, entry := range entries {
		data, err := json.Marshal(entry)
		require.NoError(t, err)
		err = client.HSet(ctx, dlqKey, entry.ID, data).Err()
		require.NoError(t, err)
	}
}

func TestDLQManager_ListEntries(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{
		MaxPageSize:     1000,
		DefaultPageSize: 50,
	}

	manager := NewDLQManager(client, config, logger)
	seedTestData(t, client)

	ctx := context.Background()

	t.Run("list all entries", func(t *testing.T) {
		filter := DLQFilter{}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 5, response.TotalCount)
		assert.Equal(t, 5, len(response.Entries))
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 1, response.TotalPages)
		assert.False(t, response.HasNext)
		assert.False(t, response.HasPrevious)
	})

	t.Run("filter by queue", func(t *testing.T) {
		filter := DLQFilter{Queue: "queue1"}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 3, response.TotalCount)
		assert.Equal(t, 3, len(response.Entries))

		for _, entry := range response.Entries {
			assert.Equal(t, "queue1", entry.Queue)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		filter := DLQFilter{Type: "job_type_a"}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 3, response.TotalCount)
		assert.Equal(t, 3, len(response.Entries))

		for _, entry := range response.Entries {
			assert.Equal(t, "job_type_a", entry.Type)
		}
	})

	t.Run("filter by error pattern", func(t *testing.T) {
		filter := DLQFilter{ErrorPattern: "timeout"}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, response.TotalCount)
		assert.Equal(t, 2, len(response.Entries))

		for _, entry := range response.Entries {
			assert.Contains(t, entry.Error.Message, "timeout")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filter := DLQFilter{}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 2,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 5, response.TotalCount)
		assert.Equal(t, 2, len(response.Entries))
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 3, response.TotalPages)
		assert.True(t, response.HasNext)
		assert.False(t, response.HasPrevious)

		pagination.Page = 2
		response, err = manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, len(response.Entries))
		assert.Equal(t, 2, response.Page)
		assert.True(t, response.HasNext)
		assert.True(t, response.HasPrevious)

		pagination.Page = 3
		response, err = manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, len(response.Entries))
		assert.Equal(t, 3, response.Page)
		assert.False(t, response.HasNext)
		assert.True(t, response.HasPrevious)
	})

	t.Run("include patterns", func(t *testing.T) {
		filter := DLQFilter{IncludePatterns: true}
		pagination := PaginationRequest{
			Page:     1,
			PageSize: 10,
		}

		response, err := manager.ListEntries(ctx, filter, pagination)
		require.NoError(t, err)
		assert.Equal(t, 5, response.TotalCount)
		assert.Greater(t, len(response.Patterns), 0)

		foundTimeoutPattern := false
		for _, pattern := range response.Patterns {
			if pattern.Count == 2 && len(pattern.AffectedQueues) == 1 && pattern.AffectedQueues[0] == "queue1" {
				foundTimeoutPattern = true
				break
			}
		}
		assert.True(t, foundTimeoutPattern, "Should find timeout pattern affecting queue1")
	})
}

func TestDLQManager_PeekEntry(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{}

	manager := NewDLQManager(client, config, logger)
	seedTestData(t, client)

	ctx := context.Background()

	t.Run("peek existing entry", func(t *testing.T) {
		entry, err := manager.PeekEntry(ctx, "entry1")
		require.NoError(t, err)
		assert.Equal(t, "entry1", entry.ID)
		assert.Equal(t, "job-entry1", entry.JobID)
		assert.Equal(t, "queue1", entry.Queue)
		assert.Equal(t, "job_type_a", entry.Type)
		assert.Equal(t, "Connection timeout", entry.Error.Message)
	})

	t.Run("peek non-existing entry", func(t *testing.T) {
		_, err := manager.PeekEntry(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDLQManager_RequeueEntries(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{BulkOperationLimit: 100}

	manager := NewDLQManager(client, config, logger)
	seedTestData(t, client)

	ctx := context.Background()

	t.Run("requeue single entry", func(t *testing.T) {
		result, err := manager.RequeueEntries(ctx, []string{"entry1"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalRequested)
		assert.Equal(t, 1, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))
		assert.Contains(t, result.Successful, "entry1")

		_, err = manager.PeekEntry(ctx, "entry1")
		assert.Error(t, err, "Entry should be removed from DLQ")

		queueKey := "queue:queue1"
		queueLen := client.LLen(ctx, queueKey).Val()
		assert.Equal(t, int64(1), queueLen, "Job should be added back to queue")
	})

	t.Run("requeue multiple entries", func(t *testing.T) {
		result, err := manager.RequeueEntries(ctx, []string{"entry2", "entry3"})
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalRequested)
		assert.Equal(t, 2, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))
	})

	t.Run("requeue non-existing entry", func(t *testing.T) {
		result, err := manager.RequeueEntries(ctx, []string{"nonexistent"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalRequested)
		assert.Equal(t, 0, len(result.Successful))
		assert.Equal(t, 1, len(result.Failed))
		assert.Equal(t, "nonexistent", result.Failed[0].ID)
	})

	t.Run("bulk operation limit", func(t *testing.T) {
		config := Config{BulkOperationLimit: 1}
		limitedManager := NewDLQManager(client, config, logger)

		_, err := limitedManager.RequeueEntries(ctx, []string{"entry4", "entry5"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bulk operation limit exceeded")
	})
}

func TestDLQManager_PurgeEntries(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{BulkOperationLimit: 100}

	manager := NewDLQManager(client, config, logger)

	ctx := context.Background()

	t.Run("purge single entry", func(t *testing.T) {
		seedTestData(t, client)

		result, err := manager.PurgeEntries(ctx, []string{"entry1"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalRequested)
		assert.Equal(t, 1, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))

		_, err = manager.PeekEntry(ctx, "entry1")
		assert.Error(t, err, "Entry should be removed from DLQ")
	})

	t.Run("purge multiple entries", func(t *testing.T) {
		result, err := manager.PurgeEntries(ctx, []string{"entry2", "entry3"})
		require.NoError(t, err)
		assert.Equal(t, 2, result.TotalRequested)
		assert.Equal(t, 2, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))
	})

	t.Run("purge non-existing entry", func(t *testing.T) {
		result, err := manager.PurgeEntries(ctx, []string{"nonexistent"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalRequested)
		assert.Equal(t, 1, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))
	})
}

func TestDLQManager_PurgeAll(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{BulkOperationLimit: 100}

	manager := NewDLQManager(client, config, logger)
	seedTestData(t, client)

	ctx := context.Background()

	t.Run("purge all with no filter", func(t *testing.T) {
		filter := DLQFilter{}
		result, err := manager.PurgeAll(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalRequested)
		assert.Equal(t, 5, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))

		stats, err := manager.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, stats.TotalEntries)
	})

	t.Run("purge all with filter", func(t *testing.T) {
		seedTestData(t, client)

		filter := DLQFilter{Queue: "queue1"}
		result, err := manager.PurgeAll(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 3, result.TotalRequested)
		assert.Equal(t, 3, len(result.Successful))
		assert.Equal(t, 0, len(result.Failed))

		stats, err := manager.GetStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, stats.TotalEntries)
	})
}

func TestDLQManager_GetStats(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := Config{}

	manager := NewDLQManager(client, config, logger)
	seedTestData(t, client)

	ctx := context.Background()

	stats, err := manager.GetStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, 5, stats.TotalEntries)
	assert.Equal(t, 2, len(stats.ByQueue))
	assert.Equal(t, 3, stats.ByQueue["queue1"])
	assert.Equal(t, 2, stats.ByQueue["queue2"])

	assert.Equal(t, 3, len(stats.ByType))
	assert.Equal(t, 3, stats.ByType["job_type_a"])
	assert.Equal(t, 1, stats.ByType["job_type_b"])
	assert.Equal(t, 1, stats.ByType["job_type_c"])
}

func TestPatternAnalyzer(t *testing.T) {
	logger := createTestLogger()
	analyzer := NewPatternAnalyzer(100, logger)

	entries := []DLQEntry{
		createTestDLQEntry("1", "queue1", "type_a", "Connection timeout after 30 seconds"),
		createTestDLQEntry("2", "queue1", "type_a", "Connection timeout after 45 seconds"),
		createTestDLQEntry("3", "queue1", "type_a", "Connection timeout after 60 seconds"),
		createTestDLQEntry("4", "queue2", "type_b", "Validation error: field 'name' is required"),
		createTestDLQEntry("5", "queue2", "type_b", "Validation error: field 'email' is required"),
		createTestDLQEntry("6", "queue3", "type_c", "Database connection failed"),
	}

	ctx := context.Background()

	patterns, err := analyzer.AnalyzeEntries(ctx, entries)
	require.NoError(t, err)

	assert.Greater(t, len(patterns), 0)

	timeoutPattern := findPatternByCount(patterns, 3)
	require.NotNil(t, timeoutPattern, "Should find timeout pattern with 3 occurrences")
	assert.Equal(t, 3, timeoutPattern.Count)
	assert.Contains(t, timeoutPattern.AffectedQueues, "queue1")
	assert.Contains(t, timeoutPattern.AffectedTypes, "type_a")

	validationPattern := findPatternByCount(patterns, 2)
	require.NotNil(t, validationPattern, "Should find validation pattern with 2 occurrences")
	assert.Equal(t, 2, validationPattern.Count)
	assert.Contains(t, validationPattern.AffectedQueues, "queue2")
	assert.Contains(t, validationPattern.AffectedTypes, "type_b")
}

func findPatternByCount(patterns []ErrorPattern, count int) *ErrorPattern {
	for _, pattern := range patterns {
		if pattern.Count == count {
			return &pattern
		}
	}
	return nil
}

func TestRemediationEngine(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	engine := NewRemediationEngine(client, logger)

	ctx := context.Background()

	t.Run("requeue entry", func(t *testing.T) {
		client.FlushDB(ctx)

		entry := createTestDLQEntry("test1", "queue1", "type_a", "Test error")
		data, err := json.Marshal(entry)
		require.NoError(t, err)

		err = client.HSet(ctx, "dlq:entries", entry.ID, data).Err()
		require.NoError(t, err)

		err = engine.Requeue(ctx, entry.ID)
		require.NoError(t, err)

		exists := client.HExists(ctx, "dlq:entries", entry.ID).Val()
		assert.False(t, exists, "Entry should be removed from DLQ")

		queueLen := client.LLen(ctx, "queue:queue1").Val()
		assert.Equal(t, int64(1), queueLen, "Job should be added to queue")
	})

	t.Run("purge entry", func(t *testing.T) {
		client.FlushDB(ctx)

		entry := createTestDLQEntry("test2", "queue1", "type_a", "Test error")
		data, err := json.Marshal(entry)
		require.NoError(t, err)

		err = client.HSet(ctx, "dlq:entries", entry.ID, data).Err()
		require.NoError(t, err)

		err = engine.Purge(ctx, entry.ID)
		require.NoError(t, err)

		exists := client.HExists(ctx, "dlq:entries", entry.ID).Val()
		assert.False(t, exists, "Entry should be removed from DLQ")
	})
}
