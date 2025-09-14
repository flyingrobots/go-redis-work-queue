// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyKey(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		now := time.Now()
		key := IdempotencyKey{
			ID:        "test-123",
			QueueName: "processing-queue",
			TenantID:  "tenant-456",
			CreatedAt: now,
			TTL:       time.Hour,
		}

		assert.Equal(t, "test-123", key.ID)
		assert.Equal(t, "processing-queue", key.QueueName)
		assert.Equal(t, "tenant-456", key.TenantID)
		assert.Equal(t, now, key.CreatedAt)
		assert.Equal(t, time.Hour, key.TTL)
	})

	t.Run("without tenant", func(t *testing.T) {
		key := IdempotencyKey{
			ID:        "test-123",
			QueueName: "processing-queue",
			CreatedAt: time.Now(),
			TTL:       time.Hour,
		}

		assert.Empty(t, key.TenantID)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		key := IdempotencyKey{
			ID:        "test-123",
			QueueName: "processing-queue",
			TenantID:  "tenant-456",
			CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			TTL:       time.Hour,
		}

		data, err := json.Marshal(key)
		require.NoError(t, err)

		var unmarshalled IdempotencyKey
		err = json.Unmarshal(data, &unmarshalled)
		require.NoError(t, err)

		assert.Equal(t, key.ID, unmarshalled.ID)
		assert.Equal(t, key.QueueName, unmarshalled.QueueName)
		assert.Equal(t, key.TenantID, unmarshalled.TenantID)
		assert.True(t, key.CreatedAt.Equal(unmarshalled.CreatedAt))
		assert.Equal(t, key.TTL, unmarshalled.TTL)
	})
}

func TestDedupEntry(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(time.Hour)

	entry := DedupEntry{
		Key:       "dedup-key-123",
		Value:     "cached-result",
		QueueName: "test-queue",
		TenantID:  "test-tenant",
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	assert.Equal(t, "dedup-key-123", entry.Key)
	assert.Equal(t, "cached-result", entry.Value)
	assert.Equal(t, "test-queue", entry.QueueName)
	assert.Equal(t, "test-tenant", entry.TenantID)
	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, expiresAt, entry.ExpiresAt)

	// Test JSON serialization
	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var unmarshalled DedupEntry
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, entry.Key, unmarshalled.Key)
	assert.Equal(t, entry.Value, unmarshalled.Value)
}

func TestOutboxEvent(t *testing.T) {
	t.Run("basic event", func(t *testing.T) {
		payload := json.RawMessage(`{"user_id": "123", "action": "created"}`)
		headers := map[string]string{
			"source":      "user-service",
			"content-type": "application/json",
		}

		event := OutboxEvent{
			ID:          "event-123",
			AggregateID: "user-123",
			EventType:   "user.created",
			Payload:     payload,
			Headers:     headers,
			CreatedAt:   time.Now(),
			Retries:     0,
			MaxRetries:  3,
		}

		assert.Equal(t, "event-123", event.ID)
		assert.Equal(t, "user-123", event.AggregateID)
		assert.Equal(t, "user.created", event.EventType)
		assert.Equal(t, payload, event.Payload)
		assert.Equal(t, headers, event.Headers)
		assert.Equal(t, 0, event.Retries)
		assert.Equal(t, 3, event.MaxRetries)
		assert.Nil(t, event.PublishedAt)
		assert.Nil(t, event.NextRetryAt)
	})

	t.Run("published event", func(t *testing.T) {
		now := time.Now()
		event := OutboxEvent{
			ID:          "event-123",
			EventType:   "test.event",
			PublishedAt: &now,
		}

		assert.NotNil(t, event.PublishedAt)
		assert.Equal(t, now, *event.PublishedAt)
	})

	t.Run("failed event with retry", func(t *testing.T) {
		nextRetry := time.Now().Add(time.Minute)
		event := OutboxEvent{
			ID:          "event-123",
			EventType:   "test.event",
			Retries:     2,
			MaxRetries:  3,
			NextRetryAt: &nextRetry,
		}

		assert.Equal(t, 2, event.Retries)
		assert.NotNil(t, event.NextRetryAt)
		assert.Equal(t, nextRetry, *event.NextRetryAt)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		payload := json.RawMessage(`{"test": true}`)
		event := OutboxEvent{
			ID:          "event-123",
			AggregateID: "agg-456",
			EventType:   "test.event",
			Payload:     payload,
			CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Retries:     1,
			MaxRetries:  3,
		}

		data, err := json.Marshal(event)
		require.NoError(t, err)

		var unmarshalled OutboxEvent
		err = json.Unmarshal(data, &unmarshalled)
		require.NoError(t, err)

		assert.Equal(t, event.ID, unmarshalled.ID)
		assert.Equal(t, event.AggregateID, unmarshalled.AggregateID)
		assert.Equal(t, event.EventType, unmarshalled.EventType)
		assert.Equal(t, event.Payload, unmarshalled.Payload)
		assert.Equal(t, event.Retries, unmarshalled.Retries)
	})
}

func TestDedupStats(t *testing.T) {
	now := time.Now()
	stats := DedupStats{
		QueueName:         "test-queue",
		TenantID:          "test-tenant",
		TotalKeys:         100,
		HitRate:          0.85,
		TotalRequests:     1000,
		DuplicatesAvoided: 850,
		LastUpdated:       now,
	}

	assert.Equal(t, "test-queue", stats.QueueName)
	assert.Equal(t, "test-tenant", stats.TenantID)
	assert.Equal(t, int64(100), stats.TotalKeys)
	assert.Equal(t, 0.85, stats.HitRate)
	assert.Equal(t, int64(1000), stats.TotalRequests)
	assert.Equal(t, int64(850), stats.DuplicatesAvoided)
	assert.Equal(t, now, stats.LastUpdated)

	// Test JSON serialization
	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var unmarshalled DedupStats
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, stats.QueueName, unmarshalled.QueueName)
	assert.Equal(t, stats.HitRate, unmarshalled.HitRate)
}

func TestIdempotencyResult(t *testing.T) {
	t.Run("first time processing", func(t *testing.T) {
		result := IdempotencyResult{
			IsFirstTime: true,
			Key:         "test-key",
		}

		assert.True(t, result.IsFirstTime)
		assert.Nil(t, result.ExistingValue)
		assert.Equal(t, "test-key", result.Key)
	})

	t.Run("duplicate processing", func(t *testing.T) {
		existingValue := map[string]interface{}{
			"result": "success",
			"timestamp": "2023-01-01T12:00:00Z",
		}

		result := IdempotencyResult{
			IsFirstTime:   false,
			ExistingValue: existingValue,
			Key:          "test-key",
		}

		assert.False(t, result.IsFirstTime)
		assert.Equal(t, existingValue, result.ExistingValue)
		assert.Equal(t, "test-key", result.Key)
	})
}

func TestProcessingStatus(t *testing.T) {
	testCases := []struct {
		status   ProcessingStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusProcessing, "processing"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{ProcessingStatus(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestProcessingStatus_Value(t *testing.T) {
	status := StatusCompleted

	value, err := status.Value()
	require.NoError(t, err)

	assert.Equal(t, int64(StatusCompleted), value)

	// Test that the value implements driver.Valuer
	var _ driver.Valuer = status
}

func TestProcessingStatus_DatabaseCompatibility(t *testing.T) {
	statuses := []ProcessingStatus{
		StatusPending,
		StatusProcessing,
		StatusCompleted,
		StatusFailed,
	}

	for i, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			value, err := status.Value()
			require.NoError(t, err)
			assert.Equal(t, int64(i), value)
		})
	}
}

// Test interfaces compilation
func TestInterfaces(t *testing.T) {
	t.Run("IdempotencyStorage interface", func(t *testing.T) {
		// This test ensures that our implementations satisfy the interface
		var _ IdempotencyStorage = (*MemoryIdempotencyStorage)(nil)
		var _ IdempotencyStorage = (*RedisIdempotencyStorage)(nil)
	})

	t.Run("ProcessingHook interface", func(t *testing.T) {
		// Mock implementation for testing
		hook := &mockProcessingHook{}
		var _ ProcessingHook = hook
	})
}

// mockProcessingHook for interface testing
type mockProcessingHook struct{}

func (m *mockProcessingHook) BeforeProcessing(ctx context.Context, jobID string, idempotencyKey IdempotencyKey) error {
	return nil
}

func (m *mockProcessingHook) AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error {
	return nil
}

func (m *mockProcessingHook) OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error {
	return nil
}

// Benchmark type operations
func BenchmarkProcessingStatus_String(b *testing.B) {
	status := StatusCompleted
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

func BenchmarkProcessingStatus_Value(b *testing.B) {
	status := StatusCompleted
	for i := 0; i < b.N; i++ {
		_, err := status.Value()
		if err != nil {
			b.Fatal(err)
		}
	}
}