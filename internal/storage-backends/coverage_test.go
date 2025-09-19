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
	RegisterBackend("test-backend", &coverageMockFactory{})
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

	// Test error conditions and edge cases
	testErrorConditions(t, registry, manager)
	testIteratorEdgeCases(t)
	testErrorCodeCoverage(t)
	testBackendConfigAndCRUD(t, manager)
}

func testErrorConditions(t *testing.T, registry *BackendRegistry, manager *BackendManager) {
	// Test registry with non-existent backend
	_, err := registry.Create("non-existent", nil)
	if err == nil {
		t.Error("Creating non-existent backend should fail")
	}

	err = registry.Validate("non-existent", nil)
	if err == nil {
		t.Error("Validating non-existent backend should fail")
	}

	// Test manager with no backend
	_, err = manager.GetBackend("non-existent-queue")
	if err == nil {
		t.Error("Getting non-existent backend should fail")
	}

	err = manager.RemoveBackend("non-existent-queue")
	if err == nil {
		t.Error("Removing non-existent backend should fail")
	}

	// Test manager operations
	queues := manager.ListQueues()
	if queues == nil {
		t.Error("Queues list should not be nil")
	}

	healthResults := manager.HealthCheck(context.Background())
	if healthResults == nil {
		t.Error("Health results should not be nil")
	}

	stats, err := manager.Stats(context.Background())
	if stats == nil && err != nil {
		t.Error("Stats should be accessible when no backends exist")
	}

	// Test manager close
	err = manager.Close()
	if err != nil {
		t.Errorf("Manager close should not fail: %v", err)
	}
}

func testIteratorEdgeCases(t *testing.T) {
	// Test empty iterator
	emptyIter := NewJobIterator([]*Job{})
	if emptyIter.Next() {
		t.Error("Empty iterator should not have next")
	}
	if emptyIter.Job() != nil {
		t.Error("Empty iterator job should be nil")
	}
	emptyIter.Close()

	// Test iterator with nil jobs
	jobs := []*Job{nil, {ID: "1"}, nil}
	iter := NewJobIterator(jobs)
	count := 0
	for iter.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}
}

func testErrorCodeCoverage(t *testing.T) {
	// Test all error codes
	testCases := []struct {
		err  error
		code string
	}{
		{ErrBackendNotFound, "BACKEND_NOT_FOUND"},
		{ErrQueueNotFound, "QUEUE_NOT_FOUND"},
		{ErrJobNotFound, "JOB_NOT_FOUND"},
		{ErrJobAlreadyAcked, "JOB_ALREADY_ACKED"},
		{ErrJobProcessing, "JOB_PROCESSING"},
		{ErrInvalidConfiguration, "INVALID_CONFIGURATION"},
		{ErrConnectionFailed, "CONNECTION_FAILED"},
		{ErrOperationNotSupported, "OPERATION_NOT_SUPPORTED"},
		{ErrTimeout, "TIMEOUT"},
		{ErrQueueEmpty, "QUEUE_EMPTY"},
		{ErrMigrationInProgress, "MIGRATION_IN_PROGRESS"},
		{ErrMigrationFailed, "MIGRATION_FAILED"},
		{ErrConsumerGroupExists, "CONSUMER_GROUP_EXISTS"},
		{ErrStreamNotFound, "STREAM_NOT_FOUND"},
		{ErrInvalidJobData, "INVALID_JOB_DATA"},
		{ErrCircuitBreakerOpen, "CIRCUIT_BREAKER_OPEN"},
		{ErrRateLimited, "RATE_LIMITED"},
	}

	for _, tc := range testCases {
		code := ErrorCode(tc.err)
		if code != tc.code {
			t.Errorf("Expected error code %s, got %s", tc.code, code)
		}
	}

	// Test wrapped errors
	backendErr := NewBackendError("test", "op", ErrTimeout)
	code := ErrorCode(backendErr)
	// BackendError wrapping ErrTimeout should still return TIMEOUT due to errors.Is() check
	if code != "TIMEOUT" {
		t.Errorf("Backend error wrapping timeout should have TIMEOUT code, got %s", code)
	}

	// Test backend error with nil underlying error
	pureBackendErr := NewBackendError("test", "op", nil)
	code = ErrorCode(pureBackendErr)
	if code != "BACKEND_ERROR" {
		t.Errorf("Pure backend error should have BACKEND_ERROR code, got %s", code)
	}

	configErr := NewConfigurationError("field", "value", "msg")
	if ErrorCode(configErr) != "CONFIGURATION_ERROR" {
		t.Error("Config error should have CONFIGURATION_ERROR code")
	}

	migrationErr := NewMigrationError("phase", "src", "tgt", "job", "msg", nil)
	if ErrorCode(migrationErr) != "MIGRATION_ERROR" {
		t.Error("Migration error should have MIGRATION_ERROR code")
	}

	connErr := NewConnectionError("backend", "url", nil)
	if ErrorCode(connErr) != "CONNECTION_ERROR" {
		t.Error("Connection error should have CONNECTION_ERROR code")
	}

	validationErr := NewValidationError("backend", "field", "value", "rule", "msg")
	if ErrorCode(validationErr) != "VALIDATION_ERROR" {
		t.Error("Validation error should have VALIDATION_ERROR code")
	}

	opErr := NewOperationError("backend", "queue", "op", "job", nil)
	if ErrorCode(opErr) != "OPERATION_ERROR" {
		t.Error("Operation error should have OPERATION_ERROR code")
	}

	// Test retryability
	if !IsRetryable(ErrTimeout) {
		t.Error("Timeout should be retryable")
	}

	if !IsRetryable(ErrConnectionFailed) {
		t.Error("Connection failed should be retryable")
	}

	if !IsRetryable(ErrRateLimited) {
		t.Error("Rate limited should be retryable")
	}

	if IsRetryable(ErrCircuitBreakerOpen) {
		t.Error("Circuit breaker open should not be retryable")
	}

	if IsRetryable(ErrJobNotFound) {
		t.Error("Job not found should not be retryable")
	}

	// Test permanence
	if !IsPermanent(ErrJobNotFound) {
		t.Error("Job not found should be permanent")
	}

	if !IsPermanent(ErrInvalidConfiguration) {
		t.Error("Invalid configuration should be permanent")
	}

	if IsPermanent(ErrTimeout) {
		t.Error("Timeout should not be permanent")
	}

	// Test temporariness
	if !IsTemporary(ErrConnectionFailed) {
		t.Error("Connection failed should be temporary")
	}

	if IsTemporary(ErrJobNotFound) {
		t.Error("Job not found should not be temporary")
	}
}

func testBackendConfigAndCRUD(t *testing.T, manager *BackendManager) {
	// Test backend registration and creation
	factory := &coverageMockFactory{}
	config := MockConfig{URL: "test://localhost"}

	// Test factory create
	backend, err := factory.Create(config)
	if err != nil {
		t.Errorf("Factory create should not fail: %v", err)
	}
	if backend == nil {
		t.Error("Created backend should not be nil")
	}

	// Test factory validate
	err = factory.Validate(config)
	if err != nil {
		t.Errorf("Factory validate should not fail: %v", err)
	}

	// Test backend capabilities
	caps := backend.Capabilities()
	if caps.AtomicAck {
		t.Error("Mock backend should not support atomic ack")
	}

	// Test backend stats
	stats, err := backend.Stats(context.Background())
	if err != nil {
		t.Errorf("Backend stats should not fail: %v", err)
	}
	if stats == nil {
		t.Error("Backend stats should not be nil")
	}

	// Test backend health
	health := backend.Health(context.Background())
	if health.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}

	// Test backend close
	err = backend.Close()
	if err != nil {
		t.Errorf("Backend close should not fail: %v", err)
	}

	// Test error message formatting
	testErrorMessages(t)
	testMigrationEdgeCases(t, manager)
}

func testErrorMessages(t *testing.T) {
	// Test all error message methods
	backendErr := NewBackendError("test-backend", "test-operation", ErrTimeout)
	msg := backendErr.Error()
	if msg == "" {
		t.Error("Backend error message should not be empty")
	}
	if backendErr.Unwrap() != ErrTimeout {
		t.Error("Backend error should unwrap to timeout")
	}

	configErr := NewConfigurationError("test-field", "test-value", "test message")
	msg = configErr.Error()
	if msg == "" {
		t.Error("Configuration error message should not be empty")
	}

	migrationErr := NewMigrationError("test-phase", "src", "tgt", "job123", "test message", ErrTimeout)
	msg = migrationErr.Error()
	if msg == "" {
		t.Error("Migration error message should not be empty")
	}
	if migrationErr.Unwrap() != ErrTimeout {
		t.Error("Migration error should unwrap to timeout")
	}

	connErr := NewConnectionError("test-backend", "redis://localhost", ErrConnectionFailed)
	msg = connErr.Error()
	if msg == "" {
		t.Error("Connection error message should not be empty")
	}
	if connErr.Unwrap() != ErrConnectionFailed {
		t.Error("Connection error should unwrap to connection failed")
	}

	validationErr := NewValidationError("test-backend", "field", "value", "rule", "validation message")
	msg = validationErr.Error()
	if msg == "" {
		t.Error("Validation error message should not be empty")
	}

	opErr := NewOperationError("backend", "queue", "operation", "job123", ErrJobNotFound)
	msg = opErr.Error()
	if msg == "" {
		t.Error("Operation error message should not be empty")
	}
	if opErr.Unwrap() != ErrJobNotFound {
		t.Error("Operation error should unwrap to job not found")
	}
}

func testMigrationEdgeCases(t *testing.T, manager *BackendManager) {
	// Test migration manager edge cases
	migrationManager := NewMigrationManager(manager)

	// Test getting status of non-existent migration
	_, err := migrationManager.GetMigrationStatus("non-existent")
	if err == nil {
		t.Error("Getting non-existent migration status should fail")
	}

	// Test cancelling non-existent migration
	err = migrationManager.CancelMigration("non-existent")
	if err == nil {
		t.Error("Cancelling non-existent migration should fail")
	}

	// Test list active migrations when none exist
	active := migrationManager.ListActiveMigrations()
	if active == nil {
		t.Error("Active migrations list should not be nil")
	}
	if len(active) != 0 {
		t.Error("Active migrations list should be empty initially")
	}

	// Test migration tool planning with non-existent queue
	tool := NewMigrationTool(manager)
	_, err = tool.PlanMigration(context.Background(), "non-existent", MigrationOptions{})
	if err == nil {
		t.Error("Planning migration for non-existent queue should fail")
	}

	// Test migration tool execution with non-existent queue
	_, err = tool.ExecuteMigration(context.Background(), "non-existent", MigrationOptions{})
	if err == nil {
		t.Error("Executing migration for non-existent queue should fail")
	}

	// Test monitoring non-existent migration
	_, err = tool.MonitorMigration("non-existent")
	if err == nil {
		t.Error("Monitoring non-existent migration should fail")
	}

	// Test quick migrate with non-existent queue
	_, err = tool.QuickMigrate(context.Background(), "non-existent", "target")
	if err == nil {
		t.Error("Quick migrate for non-existent queue should fail")
	}
}

// MockConfig for testing
type MockConfig struct {
	URL string
}

// coverageMockFactory for testing
type coverageMockFactory struct{}

func (f *coverageMockFactory) Create(config interface{}) (QueueBackend, error) {
	return &coverageMockBackend{}, nil
}

func (f *coverageMockFactory) Validate(config interface{}) error {
	return nil
}

// coverageMockBackend for testing
type coverageMockBackend struct{}

func (m *coverageMockBackend) Enqueue(ctx context.Context, job *Job) error { return nil }
func (m *coverageMockBackend) Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error) {
	return nil, nil
}
func (m *coverageMockBackend) Ack(ctx context.Context, jobID string) error                { return nil }
func (m *coverageMockBackend) Nack(ctx context.Context, jobID string, requeue bool) error { return nil }
func (m *coverageMockBackend) Length(ctx context.Context) (int64, error)                  { return 0, nil }
func (m *coverageMockBackend) Peek(ctx context.Context, offset int64) (*Job, error)       { return nil, nil }
func (m *coverageMockBackend) Move(ctx context.Context, jobID string, targetQueue string) error {
	return nil
}
func (m *coverageMockBackend) Iter(ctx context.Context, opts IterOptions) (Iterator, error) {
	return NewJobIterator([]*Job{}), nil
}
func (m *coverageMockBackend) Capabilities() BackendCapabilities { return BackendCapabilities{} }
func (m *coverageMockBackend) Stats(ctx context.Context) (*BackendStats, error) {
	return &BackendStats{}, nil
}
func (m *coverageMockBackend) Health(ctx context.Context) HealthStatus {
	return HealthStatus{Status: HealthStatusHealthy}
}
func (m *coverageMockBackend) Close() error { return nil }
