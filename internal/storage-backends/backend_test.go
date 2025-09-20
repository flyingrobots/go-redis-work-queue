//go:build storage_backends_tests
// +build storage_backends_tests

package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BackendTestSuite provides a comprehensive test suite for any QueueBackend implementation
type BackendTestSuite struct {
	backend QueueBackend
	ctx     context.Context
}

// NewBackendTestSuite creates a new test suite for a backend
func NewBackendTestSuite(backend QueueBackend) *BackendTestSuite {
	return &BackendTestSuite{
		backend: backend,
		ctx:     context.Background(),
	}
}

// TestInterface verifies that the backend implements the QueueBackend interface correctly
func (suite *BackendTestSuite) TestInterface(t *testing.T) {
	t.Run("Interface Implementation", func(t *testing.T) {
		assert.Implements(t, (*QueueBackend)(nil), suite.backend)
	})
}

// TestCapabilities verifies that capabilities are properly defined
func (suite *BackendTestSuite) TestCapabilities(t *testing.T) {
	t.Run("Capabilities Defined", func(t *testing.T) {
		caps := suite.backend.Capabilities()

		// Capabilities should be deterministic
		caps2 := suite.backend.Capabilities()
		assert.Equal(t, caps, caps2, "Capabilities should be consistent")
	})
}

// TestHealth verifies health check functionality
func (suite *BackendTestSuite) TestHealth(t *testing.T) {
	t.Run("Health Check", func(t *testing.T) {
		health := suite.backend.Health(suite.ctx)

		assert.NotEmpty(t, health.Status, "Health status should not be empty")
		assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, health.Status,
			"Health status should be one of: healthy, degraded, unhealthy")
		assert.False(t, health.CheckedAt.IsZero(), "CheckedAt should be set")
	})
}

// TestStats verifies stats collection
func (suite *BackendTestSuite) TestStats(t *testing.T) {
	t.Run("Stats Collection", func(t *testing.T) {
		stats, err := suite.backend.Stats(suite.ctx)
		assert.NoError(t, err, "Stats should not return error")
		assert.NotNil(t, stats, "Stats should not be nil")

		// Stats should have reasonable default values
		assert.GreaterOrEqual(t, stats.EnqueueRate, float64(0), "Enqueue rate should be non-negative")
		assert.GreaterOrEqual(t, stats.DequeueRate, float64(0), "Dequeue rate should be non-negative")
		assert.GreaterOrEqual(t, stats.ErrorRate, float64(0), "Error rate should be non-negative")
		assert.GreaterOrEqual(t, stats.QueueDepth, int64(0), "Queue depth should be non-negative")
	})
}

// TestBasicEnqueueDequeue verifies basic job flow
func (suite *BackendTestSuite) TestBasicEnqueueDequeue(t *testing.T) {
	t.Run("Basic Enqueue/Dequeue", func(t *testing.T) {
		job := &Job{
			ID:        "test-job-1",
			Type:      "test",
			Queue:     "test-queue",
			Payload:   map[string]interface{}{"message": "hello world"},
			Priority:  5,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"test": true},
			Tags:      []string{"test", "basic"},
		}

		// Test enqueue
		err := suite.backend.Enqueue(suite.ctx, job)
		assert.NoError(t, err, "Enqueue should succeed")

		// Verify queue length increased
		length, err := suite.backend.Length(suite.ctx)
		assert.NoError(t, err, "Length should not error")
		assert.GreaterOrEqual(t, length, int64(1), "Queue should have at least 1 job")

		// Test dequeue
		opts := DequeueOptions{
			Timeout: 1 * time.Second,
			Count:   1,
		}

		dequeuedJob, err := suite.backend.Dequeue(suite.ctx, opts)
		assert.NoError(t, err, "Dequeue should succeed")
		assert.NotNil(t, dequeuedJob, "Dequeued job should not be nil")

		// Verify job data
		assert.Equal(t, job.ID, dequeuedJob.ID, "Job ID should match")
		assert.Equal(t, job.Type, dequeuedJob.Type, "Job type should match")
		assert.Equal(t, job.Queue, dequeuedJob.Queue, "Job queue should match")
		assert.Equal(t, job.Priority, dequeuedJob.Priority, "Job priority should match")
	})
}

// TestFIFOOrdering verifies FIFO job ordering (if supported)
func (suite *BackendTestSuite) TestFIFOOrdering(t *testing.T) {
	t.Run("FIFO Ordering", func(t *testing.T) {
		jobs := []*Job{
			{ID: "job-1", Type: "test", Queue: "fifo-test", Payload: "first", CreatedAt: time.Now()},
			{ID: "job-2", Type: "test", Queue: "fifo-test", Payload: "second", CreatedAt: time.Now().Add(1 * time.Millisecond)},
			{ID: "job-3", Type: "test", Queue: "fifo-test", Payload: "third", CreatedAt: time.Now().Add(2 * time.Millisecond)},
		}

		// Enqueue all jobs
		for _, job := range jobs {
			err := suite.backend.Enqueue(suite.ctx, job)
			require.NoError(t, err, "Failed to enqueue job %s", job.ID)
		}

		// Dequeue and verify order
		opts := DequeueOptions{Timeout: 1 * time.Second, Count: 1}

		for i, expectedJob := range jobs {
			dequeuedJob, err := suite.backend.Dequeue(suite.ctx, opts)
			require.NoError(t, err, "Failed to dequeue job %d", i)
			require.NotNil(t, dequeuedJob, "Dequeued job %d should not be nil", i)

			assert.Equal(t, expectedJob.ID, dequeuedJob.ID, "Job %d should maintain FIFO order", i)
		}
	})
}

// TestEmptyQueueBehavior verifies behavior when queue is empty
func (suite *BackendTestSuite) TestEmptyQueueBehavior(t *testing.T) {
	t.Run("Empty Queue Behavior", func(t *testing.T) {
		// Verify queue is empty
		length, err := suite.backend.Length(suite.ctx)
		assert.NoError(t, err, "Length should not error")

		if length > 0 {
			t.Skip("Queue is not empty, skipping empty queue test")
		}

		// Test dequeue with timeout
		opts := DequeueOptions{
			Timeout: 100 * time.Millisecond,
			Count:   1,
		}

		job, err := suite.backend.Dequeue(suite.ctx, opts)
		// Empty queue should either return nil job or specific error
		if err == nil {
			assert.Nil(t, job, "Empty queue should return nil job when no error")
		}

		// Test peek on empty queue
		_, err = suite.backend.Peek(suite.ctx, 0)
		// Peek on empty queue should handle gracefully
		assert.Error(t, err, "Peek on empty queue should return error")
	})
}

// TestPeekOperation verifies peek functionality
func (suite *BackendTestSuite) TestPeekOperation(t *testing.T) {
	t.Run("Peek Operation", func(t *testing.T) {
		job := &Job{
			ID:        "peek-test-job",
			Type:      "test",
			Queue:     "peek-test",
			Payload:   "peek payload",
			CreatedAt: time.Now(),
		}

		// Enqueue job
		err := suite.backend.Enqueue(suite.ctx, job)
		require.NoError(t, err, "Failed to enqueue job for peek test")

		// Peek should return the job without removing it
		peekedJob, err := suite.backend.Peek(suite.ctx, 0)
		assert.NoError(t, err, "Peek should not error")
		assert.NotNil(t, peekedJob, "Peeked job should not be nil")
		assert.Equal(t, job.ID, peekedJob.ID, "Peeked job should match enqueued job")

		// Verify job is still in queue
		length, err := suite.backend.Length(suite.ctx)
		assert.NoError(t, err, "Length should not error")
		assert.GreaterOrEqual(t, length, int64(1), "Job should still be in queue after peek")

		// Clean up
		_, err = suite.backend.Dequeue(suite.ctx, DequeueOptions{Timeout: 1 * time.Second})
		assert.NoError(t, err, "Cleanup dequeue should succeed")
	})
}

// TestBulkOperations verifies batch operations (if supported)
func (suite *BackendTestSuite) TestBulkOperations(t *testing.T) {
	caps := suite.backend.Capabilities()

	if !caps.BatchOperations {
		t.Skip("Backend does not support batch operations")
	}

	t.Run("Bulk Operations", func(t *testing.T) {
		jobs := make([]*Job, 10)
		for i := 0; i < 10; i++ {
			jobs[i] = &Job{
				ID:        fmt.Sprintf("bulk-job-%d", i),
				Type:      "bulk-test",
				Queue:     "bulk-test",
				Payload:   fmt.Sprintf("bulk payload %d", i),
				CreatedAt: time.Now(),
			}
		}

		// Bulk enqueue (if supported by implementation)
		for _, job := range jobs {
			err := suite.backend.Enqueue(suite.ctx, job)
			require.NoError(t, err, "Bulk enqueue should succeed for job %s", job.ID)
		}

		// Verify all jobs were enqueued
		length, err := suite.backend.Length(suite.ctx)
		assert.NoError(t, err, "Length should not error")
		assert.GreaterOrEqual(t, length, int64(len(jobs)), "All jobs should be enqueued")

		// Bulk dequeue (if supported)
		opts := DequeueOptions{
			Timeout: 1 * time.Second,
			Count:   5, // Try to dequeue 5 jobs at once
		}

		dequeuedJob, err := suite.backend.Dequeue(suite.ctx, opts)
		assert.NoError(t, err, "Bulk dequeue should succeed")
		assert.NotNil(t, dequeuedJob, "Should get at least one job")
	})
}

// TestAckNackOperations verifies acknowledgment operations
func (suite *BackendTestSuite) TestAckNackOperations(t *testing.T) {
	caps := suite.backend.Capabilities()

	if !caps.AtomicAck {
		t.Skip("Backend does not support atomic acknowledgments")
	}

	t.Run("Ack/Nack Operations", func(t *testing.T) {
		job := &Job{
			ID:        "ack-test-job",
			Type:      "test",
			Queue:     "ack-test",
			Payload:   "ack payload",
			CreatedAt: time.Now(),
		}

		// Enqueue and dequeue
		err := suite.backend.Enqueue(suite.ctx, job)
		require.NoError(t, err, "Failed to enqueue job")

		dequeuedJob, err := suite.backend.Dequeue(suite.ctx, DequeueOptions{Timeout: 1 * time.Second})
		require.NoError(t, err, "Failed to dequeue job")
		require.NotNil(t, dequeuedJob, "Dequeued job should not be nil")

		// Test successful acknowledgment
		err = suite.backend.Ack(suite.ctx, dequeuedJob.ID)
		assert.NoError(t, err, "Ack should succeed")

		// Test negative acknowledgment with requeue
		job2 := &Job{
			ID:        "nack-test-job",
			Type:      "test",
			Queue:     "nack-test",
			Payload:   "nack payload",
			CreatedAt: time.Now(),
		}

		err = suite.backend.Enqueue(suite.ctx, job2)
		require.NoError(t, err, "Failed to enqueue job2")

		dequeuedJob2, err := suite.backend.Dequeue(suite.ctx, DequeueOptions{Timeout: 1 * time.Second})
		require.NoError(t, err, "Failed to dequeue job2")

		// Nack with requeue
		err = suite.backend.Nack(suite.ctx, dequeuedJob2.ID, true)
		assert.NoError(t, err, "Nack with requeue should succeed")
	})
}

// TestConcurrentOperations verifies thread safety
func (suite *BackendTestSuite) TestConcurrentOperations(t *testing.T) {
	t.Run("Concurrent Operations", func(t *testing.T) {
		numGoroutines := 10
		jobsPerGoroutine := 5

		// Use channels to coordinate goroutines
		enqueueErrors := make(chan error, numGoroutines*jobsPerGoroutine)
		dequeueErrors := make(chan error, numGoroutines*jobsPerGoroutine)

		// Concurrent enqueue
		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				for j := 0; j < jobsPerGoroutine; j++ {
					job := &Job{
						ID:        fmt.Sprintf("concurrent-job-%d-%d", goroutineID, j),
						Type:      "concurrent-test",
						Queue:     "concurrent-test",
						Payload:   fmt.Sprintf("payload-%d-%d", goroutineID, j),
						CreatedAt: time.Now(),
					}

					err := suite.backend.Enqueue(suite.ctx, job)
					enqueueErrors <- err
				}
			}(g)
		}

		// Wait for all enqueue operations to complete
		for i := 0; i < numGoroutines*jobsPerGoroutine; i++ {
			err := <-enqueueErrors
			assert.NoError(t, err, "Concurrent enqueue should not fail")
		}

		// Concurrent dequeue
		for g := 0; g < numGoroutines; g++ {
			go func() {
				for j := 0; j < jobsPerGoroutine; j++ {
					_, err := suite.backend.Dequeue(suite.ctx, DequeueOptions{Timeout: 2 * time.Second})
					dequeueErrors <- err
				}
			}()
		}

		// Wait for all dequeue operations to complete
		for i := 0; i < numGoroutines*jobsPerGoroutine; i++ {
			err := <-dequeueErrors
			assert.NoError(t, err, "Concurrent dequeue should not fail")
		}
	})
}

// TestJobSerialization verifies job data integrity
func (suite *BackendTestSuite) TestJobSerialization(t *testing.T) {
	t.Run("Job Serialization", func(t *testing.T) {
		complexPayload := map[string]interface{}{
			"string":  "test value",
			"number":  42,
			"float":   3.14159,
			"boolean": true,
			"null":    nil,
			"array":   []interface{}{1, 2, 3, "four"},
			"object": map[string]interface{}{
				"nested": "value",
				"count":  100,
			},
		}

		job := &Job{
			ID:        "serialization-test",
			Type:      "complex-test",
			Queue:     "serialization-test",
			Payload:   complexPayload,
			Priority:  10,
			CreatedAt: time.Now().Truncate(time.Millisecond), // Remove nanoseconds for comparison
			Metadata: map[string]interface{}{
				"source":    "test",
				"timestamp": time.Now().Unix(),
			},
			Tags: []string{"serialization", "complex", "test"},
		}

		// Enqueue complex job
		err := suite.backend.Enqueue(suite.ctx, job)
		require.NoError(t, err, "Failed to enqueue complex job")

		// Dequeue and verify data integrity
		dequeuedJob, err := suite.backend.Dequeue(suite.ctx, DequeueOptions{Timeout: 1 * time.Second})
		require.NoError(t, err, "Failed to dequeue complex job")
		require.NotNil(t, dequeuedJob, "Dequeued job should not be nil")

		// Verify all fields
		assert.Equal(t, job.ID, dequeuedJob.ID, "Job ID should be preserved")
		assert.Equal(t, job.Type, dequeuedJob.Type, "Job type should be preserved")
		assert.Equal(t, job.Queue, dequeuedJob.Queue, "Job queue should be preserved")
		assert.Equal(t, job.Priority, dequeuedJob.Priority, "Job priority should be preserved")
		assert.Equal(t, job.Tags, dequeuedJob.Tags, "Job tags should be preserved")

		// Verify complex payload
		assert.Equal(t, job.Payload, dequeuedJob.Payload, "Complex payload should be preserved")

		// Verify metadata
		assert.Equal(t, job.Metadata, dequeuedJob.Metadata, "Metadata should be preserved")
	})
}

// TestErrorHandling verifies error conditions are handled properly
func (suite *BackendTestSuite) TestErrorHandling(t *testing.T) {
	t.Run("Error Handling", func(t *testing.T) {
		// Test with cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		job := &Job{
			ID:        "cancelled-context-test",
			Type:      "test",
			Queue:     "error-test",
			Payload:   "test",
			CreatedAt: time.Now(),
		}

		err := suite.backend.Enqueue(cancelledCtx, job)
		// Should handle cancelled context gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
		}

		// Test invalid job data
		invalidJob := &Job{
			// Missing required fields
			Payload: make(chan int), // Unserialisable data
		}

		err = suite.backend.Enqueue(suite.ctx, invalidJob)
		// Should handle invalid job data
		assert.Error(t, err, "Should reject invalid job data")

		// Test invalid operations
		err = suite.backend.Ack(suite.ctx, "non-existent-job-id")
		// Should handle non-existent job IDs gracefully
		if err != nil {
			assert.NotPanics(t, func() { err.Error() }, "Error should be safe to call")
		}
	})
}

// RunAllTests runs the complete test suite
func (suite *BackendTestSuite) RunAllTests(t *testing.T) {
	suite.TestInterface(t)
	suite.TestCapabilities(t)
	suite.TestHealth(t)
	suite.TestStats(t)
	suite.TestBasicEnqueueDequeue(t)
	suite.TestFIFOOrdering(t)
	suite.TestEmptyQueueBehavior(t)
	suite.TestPeekOperation(t)
	suite.TestBulkOperations(t)
	suite.TestAckNackOperations(t)
	suite.TestConcurrentOperations(t)
	suite.TestJobSerialization(t)
	suite.TestErrorHandling(t)
}
