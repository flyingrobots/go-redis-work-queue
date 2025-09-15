package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBackend implements QueueBackend for testing
type MockBackend struct {
	jobs         map[string]*Job
	capabilities BackendCapabilities
	stats        *BackendStats
	health       HealthStatus
	closed       bool
}

func NewMockBackend(capabilities BackendCapabilities) *MockBackend {
	return &MockBackend{
		jobs:         make(map[string]*Job),
		capabilities: capabilities,
		stats: &BackendStats{
			EnqueueRate: 0,
			DequeueRate: 0,
			ErrorRate:   0,
			QueueDepth:  0,
		},
		health: HealthStatus{
			Status:    HealthStatusHealthy,
			CheckedAt: time.Now(),
		},
	}
}

func (m *MockBackend) Enqueue(ctx context.Context, job *Job) error {
	if m.closed {
		return ErrConnectionFailed
	}
	m.jobs[job.ID] = job
	m.stats.EnqueueRate++
	m.stats.QueueDepth++
	return nil
}

func (m *MockBackend) Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error) {
	if m.closed {
		return nil, ErrConnectionFailed
	}
	for id, job := range m.jobs {
		delete(m.jobs, id)
		m.stats.DequeueRate++
		m.stats.QueueDepth--
		return job, nil
	}
	return nil, nil // No jobs available
}

func (m *MockBackend) Ack(ctx context.Context, jobID string) error {
	if m.closed {
		return ErrConnectionFailed
	}
	return nil
}

func (m *MockBackend) Nack(ctx context.Context, jobID string, requeue bool) error {
	if m.closed {
		return ErrConnectionFailed
	}
	return nil
}

func (m *MockBackend) Length(ctx context.Context) (int64, error) {
	if m.closed {
		return 0, ErrConnectionFailed
	}
	return int64(len(m.jobs)), nil
}

func (m *MockBackend) Peek(ctx context.Context, offset int64) (*Job, error) {
	if m.closed {
		return nil, ErrConnectionFailed
	}
	i := 0
	for _, job := range m.jobs {
		if int64(i) == offset {
			return job, nil
		}
		i++
	}
	return nil, nil
}

func (m *MockBackend) Move(ctx context.Context, jobID string, targetQueue string) error {
	if m.closed {
		return ErrConnectionFailed
	}
	if _, exists := m.jobs[jobID]; !exists {
		return ErrJobNotFound
	}
	return nil
}

func (m *MockBackend) Iter(ctx context.Context, opts IterOptions) (Iterator, error) {
	if m.closed {
		return nil, ErrConnectionFailed
	}
	jobs := make([]*Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return NewJobIterator(jobs), nil
}

func (m *MockBackend) Capabilities() BackendCapabilities {
	return m.capabilities
}

func (m *MockBackend) Stats(ctx context.Context) (*BackendStats, error) {
	if m.closed {
		return nil, ErrConnectionFailed
	}
	return m.stats, nil
}

func (m *MockBackend) Health(ctx context.Context) HealthStatus {
	if m.closed {
		return HealthStatus{
			Status:    HealthStatusUnhealthy,
			Message:   "Backend is closed",
			CheckedAt: time.Now(),
		}
	}
	return m.health
}

func (m *MockBackend) Close() error {
	m.closed = true
	return nil
}

// MockFactory creates mock backends
type MockFactory struct {
	capabilities BackendCapabilities
}

func (f *MockFactory) Create(config interface{}) (QueueBackend, error) {
	return NewMockBackend(f.capabilities), nil
}

func (f *MockFactory) Validate(config interface{}) error {
	return nil
}

func TestBackendRegistry(t *testing.T) {
	registry := NewBackendRegistry()

	// Test registration
	factory := &MockFactory{
		capabilities: BackendCapabilities{
			AtomicAck:          true,
			ConsumerGroups:     true,
			Replay:             true,
			Persistence:        true,
			BatchOperations:    true,
		},
	}

	registry.Register("mock", factory)

	// Test list
	backends := registry.List()
	assert.Contains(t, backends, "mock")

	// Test creation
	backend, err := registry.Create("mock", nil)
	require.NoError(t, err)
	assert.NotNil(t, backend)

	// Test capabilities
	caps := backend.Capabilities()
	assert.True(t, caps.AtomicAck)
	assert.True(t, caps.ConsumerGroups)
	assert.True(t, caps.Replay)

	// Test validation
	err = registry.Validate("mock", nil)
	assert.NoError(t, err)

	// Test non-existent backend
	_, err = registry.Create("nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestBackendManager(t *testing.T) {
	registry := NewBackendRegistry()
	factory := &MockFactory{
		capabilities: BackendCapabilities{
			Persistence: true,
		},
	}
	registry.Register("mock", factory)

	manager := NewBackendManager(registry)

	// Test adding backend
	config := BackendConfig{
		Type: "mock",
		Name: "test-backend",
		URL:  "mock://localhost",
	}

	err := manager.AddBackend("test-queue", config)
	require.NoError(t, err)

	// Test getting backend
	backend, err := manager.GetBackend("test-queue")
	require.NoError(t, err)
	assert.NotNil(t, backend)

	// Test listing queues
	queues := manager.ListQueues()
	assert.Contains(t, queues, "test-queue")

	// Test health check
	ctx := context.Background()
	health := manager.HealthCheck(ctx)
	assert.Contains(t, health, "test-queue")
	assert.Equal(t, HealthStatusHealthy, health["test-queue"].Status)

	// Test stats
	stats, err := manager.Stats(ctx)
	require.NoError(t, err)
	assert.Contains(t, stats, "test-queue")

	// Test removing backend
	err = manager.RemoveBackend("test-queue")
	require.NoError(t, err)

	// Verify removal
	_, err = manager.GetBackend("test-queue")
	assert.Error(t, err)

	// Test close
	err = manager.Close()
	assert.NoError(t, err)
}

func TestJobIterator(t *testing.T) {
	jobs := []*Job{
		{ID: "1", Type: "test", Queue: "q1"},
		{ID: "2", Type: "test", Queue: "q1"},
		{ID: "3", Type: "test", Queue: "q1"},
	}

	iter := NewJobIterator(jobs)

	// Test iteration
	count := 0
	for iter.Next() {
		job := iter.Job()
		assert.NotNil(t, job)
		assert.Equal(t, jobs[count].ID, job.ID)
		count++
	}

	assert.Equal(t, 3, count)
	assert.NoError(t, iter.Error())

	// Test close
	err := iter.Close()
	assert.NoError(t, err)

	// Test iteration after close
	assert.False(t, iter.Next())
}

func TestMockBackendOperations(t *testing.T) {
	backend := NewMockBackend(BackendCapabilities{
		AtomicAck:       true,
		Persistence:     true,
		BatchOperations: true,
	})

	ctx := context.Background()

	// Test enqueue
	job := &Job{
		ID:    "test-job-1",
		Type:  "test",
		Queue: "test-queue",
		Payload: map[string]interface{}{
			"message": "hello world",
		},
		Priority:   1,
		CreatedAt:  time.Now(),
		RetryCount: 0,
		MaxRetries: 3,
	}

	err := backend.Enqueue(ctx, job)
	require.NoError(t, err)

	// Test length
	length, err := backend.Length(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), length)

	// Test peek
	peekedJob, err := backend.Peek(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, job.ID, peekedJob.ID)

	// Test dequeue
	dequeuedJob, err := backend.Dequeue(ctx, DequeueOptions{})
	require.NoError(t, err)
	assert.Equal(t, job.ID, dequeuedJob.ID)

	// Test length after dequeue
	length, err = backend.Length(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), length)

	// Test ack
	err = backend.Ack(ctx, job.ID)
	assert.NoError(t, err)

	// Test nack
	err = backend.Nack(ctx, job.ID, true)
	assert.NoError(t, err)

	// Test stats
	stats, err := backend.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, float64(1), stats.EnqueueRate)
	assert.Equal(t, float64(1), stats.DequeueRate)

	// Test health
	health := backend.Health(ctx)
	assert.Equal(t, HealthStatusHealthy, health.Status)

	// Test capabilities
	caps := backend.Capabilities()
	assert.True(t, caps.AtomicAck)
	assert.True(t, caps.Persistence)
	assert.True(t, caps.BatchOperations)
}

func TestBackendAfterClose(t *testing.T) {
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	// Close the backend
	err := backend.Close()
	require.NoError(t, err)

	// Test operations after close
	job := &Job{ID: "test", Type: "test", Queue: "test"}

	err = backend.Enqueue(ctx, job)
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionFailed, err)

	_, err = backend.Dequeue(ctx, DequeueOptions{})
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionFailed, err)

	_, err = backend.Length(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionFailed, err)

	_, err = backend.Stats(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionFailed, err)

	health := backend.Health(ctx)
	assert.Equal(t, HealthStatusUnhealthy, health.Status)
}

func TestIteratorOperations(t *testing.T) {
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	// Add some jobs
	jobs := []*Job{
		{ID: "1", Type: "test", Queue: "q1"},
		{ID: "2", Type: "test", Queue: "q1"},
		{ID: "3", Type: "test", Queue: "q1"},
	}

	for _, job := range jobs {
		err := backend.Enqueue(ctx, job)
		require.NoError(t, err)
	}

	// Test iteration
	iter, err := backend.Iter(ctx, IterOptions{})
	require.NoError(t, err)

	var iteratedJobs []*Job
	for iter.Next() {
		job := iter.Job()
		iteratedJobs = append(iteratedJobs, job)
	}

	assert.NoError(t, iter.Error())
	assert.Len(t, iteratedJobs, 3)

	err = iter.Close()
	assert.NoError(t, err)
}

func TestMoveOperation(t *testing.T) {
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	// Add a job
	job := &Job{ID: "test-job", Type: "test", Queue: "source-queue"}
	err := backend.Enqueue(ctx, job)
	require.NoError(t, err)

	// Test move
	err = backend.Move(ctx, job.ID, "target-queue")
	assert.NoError(t, err)

	// Test move non-existent job
	err = backend.Move(ctx, "nonexistent", "target-queue")
	assert.Error(t, err)
	assert.Equal(t, ErrJobNotFound, err)
}

func BenchmarkMockBackendEnqueue(b *testing.B) {
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	job := &Job{
		ID:    "bench-job",
		Type:  "benchmark",
		Queue: "bench-queue",
		Payload: map[string]interface{}{
			"data": "benchmark data",
		},
		CreatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job.ID = string(rune(i))
		err := backend.Enqueue(ctx, job)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMockBackendDequeue(b *testing.B) {
	backend := NewMockBackend(BackendCapabilities{})
	ctx := context.Background()

	// Pre-populate with jobs
	for i := 0; i < b.N; i++ {
		job := &Job{
			ID:        string(rune(i)),
			Type:      "benchmark",
			Queue:     "bench-queue",
			CreatedAt: time.Now(),
		}
		err := backend.Enqueue(ctx, job)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := backend.Dequeue(ctx, DequeueOptions{})
		if err != nil {
			b.Fatal(err)
		}
	}
}