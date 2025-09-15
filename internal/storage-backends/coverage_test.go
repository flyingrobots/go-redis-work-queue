package storage

import (
	"context"
	"testing"
)

// TestCoverageBasics tests basic type creation and validation to improve coverage
func TestCoverageBasics(t *testing.T) {
	// Test backend registry
	registry := NewBackendRegistry()
	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	// Test backend manager
	manager := NewBackendManager(registry)
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	// Test error functions
	backendErr := NewBackendError("test-backend", "test-op", nil)
	if backendErr == nil {
		t.Fatal("Backend error should not be nil")
	}

	configErr := NewConfigurationError("test-field", "test-value", "test message")
	if configErr == nil {
		t.Fatal("Config error should not be nil")
	}

	migrationErr := NewMigrationError("test-phase", "src", "tgt", "job1", "test msg", nil)
	if migrationErr == nil {
		t.Fatal("Migration error should not be nil")
	}

	connErr := NewConnectionError("test-backend", "redis://localhost", nil)
	if connErr == nil {
		t.Fatal("Connection error should not be nil")
	}

	validationErr := NewValidationError("test-backend", "field", "value", "rule", "msg")
	if validationErr == nil {
		t.Fatal("Validation error should not be nil")
	}

	opErr := NewOperationError("backend", "queue", "op", "job1", nil)
	if opErr == nil {
		t.Fatal("Operation error should not be nil")
	}

	// Test error checking functions
	if IsRetryable(nil) {
		t.Error("nil should not be retryable")
	}

	if IsPermanent(nil) {
		t.Error("nil should not be permanent")
	}

	if IsTemporary(nil) {
		t.Error("nil should not be temporary")
	}

	code := ErrorCode(nil)
	if code != "UNKNOWN_ERROR" {
		t.Errorf("Expected UNKNOWN_ERROR, got %s", code)
	}

	// Test predefined errors
	code = ErrorCode(ErrBackendNotFound)
	if code != "BACKEND_NOT_FOUND" {
		t.Errorf("Expected BACKEND_NOT_FOUND, got %s", code)
	}

	// Test iterator
	jobs := []*Job{
		{ID: "1", Type: "test"},
		{ID: "2", Type: "test"},
	}
	iter := NewJobIterator(jobs)
	if iter == nil {
		t.Fatal("Iterator should not be nil")
	}

	count := 0
	for iter.Next() {
		job := iter.Job()
		if job == nil {
			t.Error("Job should not be nil")
		}
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 jobs, got %d", count)
	}

	if iter.Error() != nil {
		t.Errorf("Iterator error should be nil, got %v", iter.Error())
	}

	iter.Close()

	// Test migration manager and tool
	migrationManager := NewMigrationManager(manager)
	if migrationManager == nil {
		t.Fatal("Migration manager should not be nil")
	}

	tool := NewMigrationTool(manager)
	if tool == nil {
		t.Fatal("Migration tool should not be nil")
	}

	// Test registry functions
	backends := registry.List()
	if backends == nil {
		t.Error("Backends list should not be nil")
	}

	// Test the default registry
	defaultReg := DefaultRegistry()
	if defaultReg == nil {
		t.Fatal("Default registry should not be nil")
	}

	// Test global functions
	RegisterBackend("test-backend", &MockFactory{})
	allBackends := defaultReg.List()
	found := false
	for _, name := range allBackends {
		if name == "test-backend" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test-backend should be registered")
	}
}

// MockFactory for testing
type MockFactory struct{}

func (f *MockFactory) Create(config interface{}) (QueueBackend, error) {
	return &MockBackend{}, nil
}

func (f *MockFactory) Validate(config interface{}) error {
	return nil
}

// MockBackend for testing
type MockBackend struct{}

func (m *MockBackend) Enqueue(ctx context.Context, job *Job) error { return nil }
func (m *MockBackend) Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error) { return nil, nil }
func (m *MockBackend) Ack(ctx context.Context, jobID string) error { return nil }
func (m *MockBackend) Nack(ctx context.Context, jobID string, requeue bool) error { return nil }
func (m *MockBackend) Length(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockBackend) Peek(ctx context.Context, offset int64) (*Job, error) { return nil, nil }
func (m *MockBackend) Move(ctx context.Context, jobID string, targetQueue string) error { return nil }
func (m *MockBackend) Iter(ctx context.Context, opts IterOptions) (Iterator, error) { return NewJobIterator([]*Job{}), nil }
func (m *MockBackend) Capabilities() BackendCapabilities { return BackendCapabilities{} }
func (m *MockBackend) Stats(ctx context.Context) (*BackendStats, error) { return &BackendStats{}, nil }
func (m *MockBackend) Health(ctx context.Context) HealthStatus { return HealthStatus{Status: HealthStatusHealthy} }
func (m *MockBackend) Close() error { return nil }