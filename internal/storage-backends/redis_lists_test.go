package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RedisListsTestSuite tests the Redis Lists backend implementation
type RedisListsTestSuite struct {
	suite.Suite
	redis   *miniredis.Miniredis
	backend *RedisListsBackend
	ctx     context.Context
}

// SetupTest initializes the test environment
func (s *RedisListsTestSuite) SetupTest() {
	// Start in-memory Redis
	s.redis = miniredis.NewMiniRedis()
	err := s.redis.Start()
	s.Require().NoError(err)

	// Create backend
	config := RedisListsConfig{
		URL:            "redis://" + s.redis.Addr(),
		Database:       0,
		KeyPrefix:      "test:",
		MaxConnections: 10,
		ConnTimeout:    5 * time.Second,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		PoolTimeout:    4 * time.Second,
		IdleTimeout:    5 * time.Minute,
		MaxRetries:     3,
		ClusterMode:    false,
	}

	backend, err := createRedisListsBackend(config, "test-queue")
	s.Require().NoError(err)

	s.backend = backend
	s.ctx = context.Background()
}

// TearDownTest cleans up the test environment
func (s *RedisListsTestSuite) TearDownTest() {
	if s.backend != nil {
		s.backend.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
}

// TestRedisListsFactory tests the factory implementation
func (s *RedisListsTestSuite) TestRedisListsFactory() {
	factory := &RedisListsFactory{}

	// Test valid configuration
	config := RedisListsConfig{
		URL:       "redis://localhost:6379",
		Database:  0,
		KeyPrefix: "test:",
	}

	err := factory.Validate(config)
	s.NoError(err, "Valid configuration should pass validation")

	// Test invalid configuration
	invalidConfig := RedisListsConfig{
		URL: "", // Empty URL should be invalid
	}

	err = factory.Validate(invalidConfig)
	s.Error(err, "Invalid configuration should fail validation")
}

// TestCapabilities verifies the Redis Lists backend capabilities
func (s *RedisListsTestSuite) TestCapabilities() {
	caps := s.backend.Capabilities()

	// Verify expected capabilities for Redis Lists
	s.False(caps.AtomicAck, "Redis Lists should not have atomic ack")
	s.False(caps.ConsumerGroups, "Redis Lists should not support consumer groups")
	s.False(caps.Replay, "Redis Lists should not support replay")
	s.False(caps.IdempotentEnqueue, "Redis Lists should not have idempotent enqueue")
	s.True(caps.Transactions, "Redis Lists should support transactions via Lua")
	s.True(caps.Persistence, "Redis Lists should be persistent")
	s.True(caps.Clustering, "Redis Lists should support clustering with key tagging")
	s.False(caps.TimeToLive, "Redis Lists should not have TTL support")
	s.False(caps.Prioritization, "Redis Lists should not have native prioritization")
	s.True(caps.BatchOperations, "Redis Lists should support batch operations")
}

// TestBasicOperations tests basic enqueue/dequeue operations
func (s *RedisListsTestSuite) TestBasicOperations() {
	job := &Job{
		ID:      "redis-test-1",
		Type:    "test-job",
		Queue:   "test-queue",
		Payload: "test payload",
		Priority: 5,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"source": "test",
		},
		Tags: []string{"redis", "test"},
	}

	// Test enqueue
	err := s.backend.Enqueue(s.ctx, job)
	s.NoError(err, "Enqueue should succeed")

	// Verify queue length
	length, err := s.backend.Length(s.ctx)
	s.NoError(err, "Length should not error")
	s.Equal(int64(1), length, "Queue should have 1 job")

	// Test dequeue
	opts := DequeueOptions{
		Timeout: 1 * time.Second,
		Count:   1,
	}

	dequeuedJob, err := s.backend.Dequeue(s.ctx, opts)
	s.NoError(err, "Dequeue should succeed")
	s.NotNil(dequeuedJob, "Dequeued job should not be nil")

	// Verify job data
	s.Equal(job.ID, dequeuedJob.ID)
	s.Equal(job.Type, dequeuedJob.Type)
	s.Equal(job.Queue, dequeuedJob.Queue)
	s.Equal(job.Payload, dequeuedJob.Payload)
	s.Equal(job.Priority, dequeuedJob.Priority)
	s.Equal(job.Metadata, dequeuedJob.Metadata)
	s.Equal(job.Tags, dequeuedJob.Tags)
}

// TestFIFOOrdering verifies FIFO ordering of jobs
func (s *RedisListsTestSuite) TestFIFOOrdering() {
	jobs := []*Job{
		{ID: "fifo-1", Type: "test", Queue: "fifo-test", Payload: "first", CreatedAt: time.Now()},
		{ID: "fifo-2", Type: "test", Queue: "fifo-test", Payload: "second", CreatedAt: time.Now().Add(1 * time.Millisecond)},
		{ID: "fifo-3", Type: "test", Queue: "fifo-test", Payload: "third", CreatedAt: time.Now().Add(2 * time.Millisecond)},
	}

	// Enqueue all jobs
	for _, job := range jobs {
		err := s.backend.Enqueue(s.ctx, job)
		s.Require().NoError(err, "Failed to enqueue job %s", job.ID)
	}

	// Verify queue length
	length, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(3), length, "Queue should have 3 jobs")

	// Dequeue and verify FIFO order
	opts := DequeueOptions{Timeout: 1 * time.Second, Count: 1}

	for i, expectedJob := range jobs {
		dequeuedJob, err := s.backend.Dequeue(s.ctx, opts)
		s.Require().NoError(err, "Failed to dequeue job %d", i)
		s.Require().NotNil(dequeuedJob, "Dequeued job %d should not be nil", i)

		s.Equal(expectedJob.ID, dequeuedJob.ID, "Job %d should maintain FIFO order", i)
	}

	// Queue should be empty now
	finalLength, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(0), finalLength, "Queue should be empty after dequeuing all jobs")
}

// TestPeekOperation tests the peek functionality
func (s *RedisListsTestSuite) TestPeekOperation() {
	job := &Job{
		ID:      "peek-test",
		Type:    "test",
		Queue:   "peek-queue",
		Payload: "peek payload",
		CreatedAt: time.Now(),
	}

	// Enqueue job
	err := s.backend.Enqueue(s.ctx, job)
	s.Require().NoError(err)

	// Peek should return the job without removing it
	peekedJob, err := s.backend.Peek(s.ctx, 0)
	s.NoError(err, "Peek should not error")
	s.NotNil(peekedJob, "Peeked job should not be nil")
	s.Equal(job.ID, peekedJob.ID, "Peeked job should match enqueued job")

	// Verify job is still in queue
	length, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(1), length, "Job should still be in queue after peek")

	// Peek with offset beyond queue size should return error
	_, err = s.backend.Peek(s.ctx, 1)
	s.Error(err, "Peek beyond queue size should return error")
}

// TestEmptyQueueBehavior tests behavior with empty queue
func (s *RedisListsTestSuite) TestEmptyQueueBehavior() {
	// Verify queue is empty
	length, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(0), length, "Queue should be empty initially")

	// Test dequeue with short timeout
	opts := DequeueOptions{
		Timeout: 100 * time.Millisecond,
		Count:   1,
	}

	start := time.Now()
	job, err := s.backend.Dequeue(s.ctx, opts)
	elapsed := time.Since(start)

	// Should timeout or return nil job
	if err == nil {
		s.Nil(job, "Empty queue should return nil job when no error")
	}
	s.True(elapsed >= 100*time.Millisecond, "Should respect timeout")

	// Test peek on empty queue
	_, err = s.backend.Peek(s.ctx, 0)
	s.Error(err, "Peek on empty queue should return error")
}

// TestStats verifies stats collection
func (s *RedisListsTestSuite) TestStats() {
	// Get initial stats
	stats, err := s.backend.Stats(s.ctx)
	s.NoError(err)
	s.NotNil(stats)

	initialEnqueueRate := stats.EnqueueRate
	initialDequeueRate := stats.DequeueRate

	// Perform some operations
	job := &Job{
		ID:      "stats-test",
		Type:    "test",
		Queue:   "stats-queue",
		Payload: "stats payload",
		CreatedAt: time.Now(),
	}

	err = s.backend.Enqueue(s.ctx, job)
	s.NoError(err)

	_, err = s.backend.Dequeue(s.ctx, DequeueOptions{Timeout: 1 * time.Second})
	s.NoError(err)

	// Get updated stats
	updatedStats, err := s.backend.Stats(s.ctx)
	s.NoError(err)
	s.NotNil(updatedStats)

	// Stats should reflect operations (might be rate-limited updates)
	s.GreaterOrEqual(updatedStats.EnqueueRate, initialEnqueueRate)
	s.GreaterOrEqual(updatedStats.DequeueRate, initialDequeueRate)
}

// TestHealth verifies health check functionality
func (s *RedisListsTestSuite) TestHealth() {
	health := s.backend.Health(s.ctx)

	s.Equal("healthy", health.Status, "Backend should be healthy")
	s.False(health.CheckedAt.IsZero(), "CheckedAt should be set")
	s.NoError(health.Error, "Should not have error when healthy")

	// Test with Redis disconnected
	s.redis.Close()

	health = s.backend.Health(s.ctx)
	s.Equal("unhealthy", health.Status, "Backend should be unhealthy when Redis is down")
	s.NotNil(health.Error, "Should have error when unhealthy")
}

// TestConcurrentOperations tests thread safety
func (s *RedisListsTestSuite) TestConcurrentOperations() {
	numGoroutines := 5
	jobsPerGoroutine := 10

	enqueueErrors := make(chan error, numGoroutines*jobsPerGoroutine)
	dequeueErrors := make(chan error, numGoroutines*jobsPerGoroutine)

	// Concurrent enqueue
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for j := 0; j < jobsPerGoroutine; j++ {
				job := &Job{
					ID:      fmt.Sprintf("concurrent-%d-%d", goroutineID, j),
					Type:    "concurrent-test",
					Queue:   "concurrent-queue",
					Payload: fmt.Sprintf("payload-%d-%d", goroutineID, j),
					CreatedAt: time.Now(),
				}

				err := s.backend.Enqueue(s.ctx, job)
				enqueueErrors <- err
			}
		}(g)
	}

	// Wait for all enqueue operations
	for i := 0; i < numGoroutines*jobsPerGoroutine; i++ {
		err := <-enqueueErrors
		s.NoError(err, "Concurrent enqueue should not fail")
	}

	// Verify all jobs were enqueued
	length, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(numGoroutines*jobsPerGoroutine), length, "All jobs should be enqueued")

	// Concurrent dequeue
	for g := 0; g < numGoroutines; g++ {
		go func() {
			for j := 0; j < jobsPerGoroutine; j++ {
				_, err := s.backend.Dequeue(s.ctx, DequeueOptions{Timeout: 2 * time.Second})
				dequeueErrors <- err
			}
		}()
	}

	// Wait for all dequeue operations
	for i := 0; i < numGoroutines*jobsPerGoroutine; i++ {
		err := <-dequeueErrors
		s.NoError(err, "Concurrent dequeue should not fail")
	}

	// Queue should be empty
	finalLength, err := s.backend.Length(s.ctx)
	s.NoError(err)
	s.Equal(int64(0), finalLength, "Queue should be empty after concurrent operations")
}

// TestComplexJobSerialization tests serialization of complex job data
func (s *RedisListsTestSuite) TestComplexJobSerialization() {
	complexPayload := map[string]interface{}{
		"string":  "test value with unicode: 你好",
		"number":  42,
		"float":   3.14159265359,
		"boolean": true,
		"null":    nil,
		"array":   []interface{}{1, "two", 3.0, true, nil},
		"object": map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": "value",
				"count": 100,
			},
			"array_in_object": []interface{}{1, 2, 3},
		},
	}

	job := &Job{
		ID:        "complex-job",
		Type:      "complex-serialization-test",
		Queue:     "complex-queue",
		Payload:   complexPayload,
		Priority:  10,
		CreatedAt: time.Now().Truncate(time.Millisecond), // Remove nanoseconds
		Metadata: map[string]interface{}{
			"source":    "complex-test",
			"timestamp": time.Now().Unix(),
			"metadata_object": map[string]interface{}{
				"nested": "metadata",
			},
		},
		Tags: []string{"complex", "serialization", "unicode-test", "nested-data"},
	}

	// Enqueue complex job
	err := s.backend.Enqueue(s.ctx, job)
	s.Require().NoError(err, "Failed to enqueue complex job")

	// Dequeue and verify data integrity
	dequeuedJob, err := s.backend.Dequeue(s.ctx, DequeueOptions{Timeout: 1 * time.Second})
	s.Require().NoError(err, "Failed to dequeue complex job")
	s.Require().NotNil(dequeuedJob, "Dequeued job should not be nil")

	// Verify all fields are preserved
	s.Equal(job.ID, dequeuedJob.ID)
	s.Equal(job.Type, dequeuedJob.Type)
	s.Equal(job.Queue, dequeuedJob.Queue)
	s.Equal(job.Priority, dequeuedJob.Priority)
	s.Equal(job.Tags, dequeuedJob.Tags)

	// Verify complex payload structure
	s.Equal(job.Payload, dequeuedJob.Payload, "Complex payload should be perfectly preserved")

	// Verify metadata
	s.Equal(job.Metadata, dequeuedJob.Metadata, "Metadata should be preserved")
}

// TestRedisConnectionFailure tests behavior when Redis connection fails
func (s *RedisListsTestSuite) TestRedisConnectionFailure() {
	// Stop Redis
	s.redis.Close()

	job := &Job{
		ID:      "connection-failure-test",
		Type:    "test",
		Queue:   "failure-queue",
		Payload: "test",
		CreatedAt: time.Now(),
	}

	// Operations should fail gracefully
	err := s.backend.Enqueue(s.ctx, job)
	s.Error(err, "Enqueue should fail when Redis is down")

	_, err = s.backend.Dequeue(s.ctx, DequeueOptions{Timeout: 100 * time.Millisecond})
	s.Error(err, "Dequeue should fail when Redis is down")

	_, err = s.backend.Length(s.ctx)
	s.Error(err, "Length should fail when Redis is down")

	_, err = s.backend.Peek(s.ctx, 0)
	s.Error(err, "Peek should fail when Redis is down")
}

// TestConfigurationValidation tests configuration validation
func (s *RedisListsTestSuite) TestConfigurationValidation() {
	factory := &RedisListsFactory{}

	validConfigs := []RedisListsConfig{
		{
			URL:       "redis://localhost:6379",
			Database:  0,
			KeyPrefix: "test:",
		},
		{
			URL:          "rediss://secure.redis.com:6380",
			Database:     1,
			KeyPrefix:    "prod:",
			Password:     "secret",
			TLS:          true,
			ClusterMode:  true,
			ClusterAddrs: []string{"node1:6379", "node2:6379"},
		},
	}

	for i, config := range validConfigs {
		err := factory.Validate(config)
		s.NoError(err, "Valid config %d should pass validation", i)
	}

	invalidConfigs := []RedisListsConfig{
		{URL: ""}, // Empty URL
		{URL: "invalid-url"},
		{URL: "redis://localhost:6379", Database: -1}, // Negative database
	}

	for i, config := range invalidConfigs {
		err := factory.Validate(config)
		s.Error(err, "Invalid config %d should fail validation", i)
	}
}

// Helper function to create Redis Lists backend
func createRedisListsBackend(config RedisListsConfig, queueName string) (*RedisListsBackend, error) {
	var client redis.Cmdable

	if config.ClusterMode {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.ClusterAddrs,
			Password:     config.Password,
			DialTimeout:  config.ConnTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,
			IdleTimeout:  config.IdleTimeout,
			MaxRetries:   config.MaxRetries,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         config.URL[8:], // Remove redis:// prefix
			Password:     config.Password,
			DB:           config.Database,
			DialTimeout:  config.ConnTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,
			IdleTimeout:  config.IdleTimeout,
			MaxRetries:   config.MaxRetries,
		})
	}

	return &RedisListsBackend{
		client:    client,
		config:    config,
		queueName: queueName,
		keyPrefix: config.KeyPrefix,
		stats: &BackendStats{
			EnqueueRate: 0,
			DequeueRate: 0,
			ErrorRate:   0,
			QueueDepth:  0,
		},
	}, nil
}

// TestRedisListsTestSuite runs the test suite
func TestRedisListsTestSuite(t *testing.T) {
	suite.Run(t, new(RedisListsTestSuite))
}

// TestRedisListsBackendConformance runs the generic backend test suite
func TestRedisListsBackendConformance(t *testing.T) {
	// Start in-memory Redis
	miniRedis := miniredis.NewMiniRedis()
	err := miniRedis.Start()
	require.NoError(t, err)
	defer miniRedis.Close()

	// Create backend
	config := RedisListsConfig{
		URL:       "redis://" + miniRedis.Addr(),
		Database:  0,
		KeyPrefix: "conformance:",
	}

	backend, err := createRedisListsBackend(config, "conformance-test")
	require.NoError(t, err)
	defer backend.Close()

	// Run the generic test suite
	testSuite := NewBackendTestSuite(backend)
	testSuite.RunAllTests(t)
}