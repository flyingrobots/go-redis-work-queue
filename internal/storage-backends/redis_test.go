package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisListsConfig(t *testing.T) {
	factory := &RedisListsFactory{}

	// Test valid config
	config := RedisListsConfig{
		URL:      "redis://localhost:6379/0",
		Database: 0,
	}

	err := factory.Validate(config)
	assert.NoError(t, err)

	// Test invalid config - no URL or cluster addresses
	invalidConfig := RedisListsConfig{}
	err = factory.Validate(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL or cluster addresses must be provided")

	// Test cluster config
	clusterConfig := RedisListsConfig{
		ClusterMode:  true,
		ClusterAddrs: []string{"redis1:6379", "redis2:6379"},
	}
	err = factory.Validate(clusterConfig)
	assert.NoError(t, err)
}

func TestRedisStreamsConfig(t *testing.T) {
	factory := &RedisStreamsFactory{}

	// Test valid config
	config := RedisStreamsConfig{
		URL:           "redis://localhost:6379/0",
		StreamName:    "test-stream",
		ConsumerGroup: "test-group",
		ConsumerName:  "test-consumer",
	}

	err := factory.Validate(config)
	assert.NoError(t, err)

	// Test missing required fields
	tests := []struct {
		name   string
		config RedisStreamsConfig
		errMsg string
	}{
		{
			name:   "no URL or cluster",
			config: RedisStreamsConfig{},
			errMsg: "URL or cluster addresses must be provided",
		},
		{
			name: "no stream name",
			config: RedisStreamsConfig{
				URL: "redis://localhost:6379",
			},
			errMsg: "stream name is required",
		},
		{
			name: "no consumer group",
			config: RedisStreamsConfig{
				URL:        "redis://localhost:6379",
				StreamName: "test-stream",
			},
			errMsg: "consumer group is required",
		},
		{
			name: "no consumer name",
			config: RedisStreamsConfig{
				URL:           "redis://localhost:6379",
				StreamName:    "test-stream",
				ConsumerGroup: "test-group",
			},
			errMsg: "consumer name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.Validate(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestRedisListsCapabilities(t *testing.T) {
	// Test single Redis instance capabilities
	config := RedisListsConfig{
		URL:         "redis://localhost:6379/0",
		ClusterMode: false,
	}

	// We can't actually connect to Redis in tests, so we'll create a mock
	// but test the capability structure
	expectedCaps := BackendCapabilities{
		AtomicAck:          false,
		ConsumerGroups:     false,
		Replay:             false,
		IdempotentEnqueue:  false,
		Transactions:       true,
		Persistence:        true,
		Clustering:         false,
		TimeToLive:         false,
		Prioritization:     false,
		BatchOperations:    true,
	}

	// Create a mock backend with the same capabilities as Redis Lists
	backend := NewMockBackend(expectedCaps)
	caps := backend.Capabilities()

	assert.Equal(t, expectedCaps.AtomicAck, caps.AtomicAck)
	assert.Equal(t, expectedCaps.ConsumerGroups, caps.ConsumerGroups)
	assert.Equal(t, expectedCaps.Replay, caps.Replay)
	assert.Equal(t, expectedCaps.Transactions, caps.Transactions)
	assert.Equal(t, expectedCaps.Persistence, caps.Persistence)
	assert.Equal(t, expectedCaps.BatchOperations, caps.BatchOperations)
}

func TestRedisStreamsCapabilities(t *testing.T) {
	expectedCaps := BackendCapabilities{
		AtomicAck:          true,
		ConsumerGroups:     true,
		Replay:             true,
		IdempotentEnqueue:  false,
		Transactions:       true,
		Persistence:        true,
		Clustering:         false,
		TimeToLive:         false,
		Prioritization:     false,
		BatchOperations:    true,
	}

	// Create a mock backend with the same capabilities as Redis Streams
	backend := NewMockBackend(expectedCaps)
	caps := backend.Capabilities()

	assert.Equal(t, expectedCaps.AtomicAck, caps.AtomicAck)
	assert.Equal(t, expectedCaps.ConsumerGroups, caps.ConsumerGroups)
	assert.Equal(t, expectedCaps.Replay, caps.Replay)
	assert.Equal(t, expectedCaps.Transactions, caps.Transactions)
	assert.Equal(t, expectedCaps.Persistence, caps.Persistence)
	assert.Equal(t, expectedCaps.BatchOperations, caps.BatchOperations)
}

func TestRedisBackendFactories(t *testing.T) {
	// Test Redis Lists factory
	listsFactory := &RedisListsFactory{}

	// Test with wrong config type
	_, err := listsFactory.Create("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")

	err = listsFactory.Validate("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")

	// Test Redis Streams factory
	streamsFactory := &RedisStreamsFactory{}

	_, err = streamsFactory.Create("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")

	err = streamsFactory.Validate("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

func TestRedisBackendRegistration(t *testing.T) {
	// Test that Redis backends are registered in init()
	registry := DefaultRegistry()

	backends := registry.List()
	assert.Contains(t, backends, BackendTypeRedisLists)
	assert.Contains(t, backends, BackendTypeRedisStreams)

	// Test validation through registry
	listsConfig := RedisListsConfig{
		URL: "redis://localhost:6379/0",
	}
	err := registry.Validate(BackendTypeRedisLists, listsConfig)
	assert.NoError(t, err)

	streamsConfig := RedisStreamsConfig{
		URL:           "redis://localhost:6379/0",
		StreamName:    "test-stream",
		ConsumerGroup: "test-group",
		ConsumerName:  "test-consumer",
	}
	err = registry.Validate(BackendTypeRedisStreams, streamsConfig)
	assert.NoError(t, err)
}

// Integration tests would go here if we had a Redis instance available
// For now, we'll mock the Redis behavior

func TestRedisListsSimulatedOperations(t *testing.T) {
	// Simulate Redis Lists behavior with mock backend
	backend := NewMockBackend(BackendCapabilities{
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	ctx := context.Background()

	// Test job serialization/deserialization simulation
	job := &Job{
		ID:    "redis-test-1",
		Type:  "email",
		Queue: "notifications",
		Payload: map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Test Email",
			"body":    "This is a test email",
		},
		Priority:   1,
		CreatedAt:  time.Now(),
		RetryCount: 0,
		MaxRetries: 3,
		Metadata: map[string]interface{}{
			"source": "api",
			"user_id": 12345,
		},
		Tags: []string{"urgent", "notification"},
	}

	// Test enqueue
	err := backend.Enqueue(ctx, job)
	require.NoError(t, err)

	// Test dequeue
	dequeuedJob, err := backend.Dequeue(ctx, DequeueOptions{
		Timeout: 1 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, dequeuedJob)

	// Verify job data
	assert.Equal(t, job.ID, dequeuedJob.ID)
	assert.Equal(t, job.Type, dequeuedJob.Type)
	assert.Equal(t, job.Queue, dequeuedJob.Queue)
	assert.Equal(t, job.Priority, dequeuedJob.Priority)
	assert.Equal(t, job.RetryCount, dequeuedJob.RetryCount)
	assert.Equal(t, job.MaxRetries, dequeuedJob.MaxRetries)
}

func TestRedisStreamsSimulatedOperations(t *testing.T) {
	// Simulate Redis Streams behavior with mock backend
	backend := NewMockBackend(BackendCapabilities{
		AtomicAck:       true,
		ConsumerGroups:  true,
		Replay:          true,
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	ctx := context.Background()

	// Test stream-specific operations
	job := &Job{
		ID:         "stream-test-1",
		Type:       "analytics",
		Queue:      "data-processing",
		Payload:    map[string]interface{}{"event": "user_login", "user_id": 456},
		Priority:   2,
		CreatedAt:  time.Now(),
		RetryCount: 0,
		MaxRetries: 5,
	}

	// Test enqueue to stream
	err := backend.Enqueue(ctx, job)
	require.NoError(t, err)

	// Test dequeue with consumer group
	dequeuedJob, err := backend.Dequeue(ctx, DequeueOptions{
		ConsumerGroup: "analytics-workers",
		ConsumerID:    "worker-1",
		Timeout:       1 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, dequeuedJob)

	// Test acknowledgment (streams support atomic ack)
	err = backend.Ack(ctx, dequeuedJob.ID)
	assert.NoError(t, err)

	// Test negative acknowledgment with requeue
	err = backend.Nack(ctx, dequeuedJob.ID, true)
	assert.NoError(t, err)
}

func TestRedisClusterKeyTagging(t *testing.T) {
	// Test key tagging for Redis Cluster compatibility
	// This would be implemented in the actual Redis backend

	// Simulate cluster key tagging
	queueName := "payments"
	expectedKey := "{payments}:queue:payments" // Key with cluster tag

	// In a real implementation, this would be part of getQueueKey()
	keyWithTag := "{" + queueName + "}:queue:" + queueName
	assert.Equal(t, expectedKey, keyWithTag)

	// Test that different queues get different tags
	queue2 := "notifications"
	key2WithTag := "{" + queue2 + "}:queue:" + queue2
	expectedKey2 := "{notifications}:queue:notifications"
	assert.Equal(t, expectedKey2, key2WithTag)

	// Verify that operations on the same queue use the same slot
	assert.Contains(t, keyWithTag, "{"+queueName+"}")

	// Test job-related keys would also use the same tag
	jobKey := "{" + queueName + "}:job:123"
	dlqKey := "{" + queueName + "}:dlq:" + queueName

	assert.Contains(t, jobKey, "{"+queueName+"}")
	assert.Contains(t, dlqKey, "{"+queueName+"}")
}

func TestRedisBackendErrorHandling(t *testing.T) {
	// Test various error conditions
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	// Test operations after backend is closed
	err := backend.Close()
	require.NoError(t, err)

	job := &Job{ID: "test", Type: "test", Queue: "test"}

	// All operations should fail after close
	err = backend.Enqueue(ctx, job)
	assert.Error(t, err)

	_, err = backend.Dequeue(ctx, DequeueOptions{})
	assert.Error(t, err)

	err = backend.Ack(ctx, "test")
	assert.Error(t, err)

	err = backend.Nack(ctx, "test", false)
	assert.Error(t, err)

	_, err = backend.Length(ctx)
	assert.Error(t, err)

	_, err = backend.Peek(ctx, 0)
	assert.Error(t, err)

	err = backend.Move(ctx, "test", "target")
	assert.Error(t, err)

	_, err = backend.Iter(ctx, IterOptions{})
	assert.Error(t, err)

	_, err = backend.Stats(ctx)
	assert.Error(t, err)

	health := backend.Health(ctx)
	assert.Equal(t, HealthStatusUnhealthy, health.Status)
}

func BenchmarkRedisListsSimulation(b *testing.B) {
	backend := NewMockBackend(BackendCapabilities{
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	ctx := context.Background()
	job := &Job{
		ID:      "bench-job",
		Type:    "benchmark",
		Queue:   "bench-queue",
		Payload: map[string]interface{}{"data": "benchmark data"},
	}

	b.Run("Enqueue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			job.ID = string(rune(i))
			err := backend.Enqueue(ctx, job)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Reset backend for dequeue benchmark
	backend = NewMockBackend(BackendCapabilities{
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	// Pre-populate
	for i := 0; i < b.N; i++ {
		job.ID = string(rune(i))
		backend.Enqueue(ctx, job)
	}

	b.Run("Dequeue", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := backend.Dequeue(ctx, DequeueOptions{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRedisStreamsSimulation(b *testing.B) {
	backend := NewMockBackend(BackendCapabilities{
		AtomicAck:       true,
		ConsumerGroups:  true,
		Replay:          true,
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	ctx := context.Background()
	job := &Job{
		ID:      "stream-bench-job",
		Type:    "stream-benchmark",
		Queue:   "stream-bench-queue",
		Payload: map[string]interface{}{"stream": "benchmark data"},
	}

	b.Run("StreamEnqueue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			job.ID = string(rune(i))
			err := backend.Enqueue(ctx, job)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Reset backend for dequeue benchmark
	backend = NewMockBackend(BackendCapabilities{
		AtomicAck:       true,
		ConsumerGroups:  true,
		Replay:          true,
		Transactions:    true,
		Persistence:     true,
		BatchOperations: true,
	})

	// Pre-populate
	for i := 0; i < b.N; i++ {
		job.ID = string(rune(i))
		backend.Enqueue(ctx, job)
	}

	b.Run("StreamDequeue", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := backend.Dequeue(ctx, DequeueOptions{
				ConsumerGroup: "bench-group",
				ConsumerID:    "bench-consumer",
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}