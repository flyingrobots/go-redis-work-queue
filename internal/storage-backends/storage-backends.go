package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BackendRegistry manages available queue backend implementations
type BackendRegistry struct {
	backends map[string]BackendFactory
	mu       sync.RWMutex
}

// NewBackendRegistry creates a new backend registry
func NewBackendRegistry() *BackendRegistry {
	return &BackendRegistry{
		backends: make(map[string]BackendFactory),
	}
}

// Register adds a backend factory to the registry
func (r *BackendRegistry) Register(name string, factory BackendFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[name] = factory
}

// Create instantiates a backend by name
func (r *BackendRegistry) Create(name string, config interface{}) (QueueBackend, error) {
	r.mu.RLock()
	factory, exists := r.backends[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("backend %q not registered", name)
	}

	return factory.Create(config)
}

// List returns all registered backend names
func (r *BackendRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.backends))
	for name := range r.backends {
		names = append(names, name)
	}
	return names
}

// Validate checks if a backend configuration is valid
func (r *BackendRegistry) Validate(name string, config interface{}) error {
	r.mu.RLock()
	factory, exists := r.backends[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("backend %q not registered", name)
	}

	return factory.Validate(config)
}

// BackendManager orchestrates backend operations across multiple queues
type BackendManager struct {
	registry *BackendRegistry
	backends map[string]QueueBackend
	configs  map[string]BackendConfig
	mu       sync.RWMutex
}

// NewBackendManager creates a new backend manager
func NewBackendManager(registry *BackendRegistry) *BackendManager {
	return &BackendManager{
		registry: registry,
		backends: make(map[string]QueueBackend),
		configs:  make(map[string]BackendConfig),
	}
}

// AddBackend configures and adds a backend for a specific queue
func (m *BackendManager) AddBackend(queueName string, config BackendConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate configuration
	if err := m.registry.Validate(config.Type, config); err != nil {
		return fmt.Errorf("invalid config for backend %s: %w", config.Type, err)
	}

	// Create backend instance
	backend, err := m.registry.Create(config.Type, config)
	if err != nil {
		return fmt.Errorf("failed to create backend %s: %w", config.Type, err)
	}

	// Store backend and config
	m.backends[queueName] = backend
	m.configs[queueName] = config

	return nil
}

// GetBackend returns the backend for a specific queue
func (m *BackendManager) GetBackend(queueName string) (QueueBackend, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backend, exists := m.backends[queueName]
	if !exists {
		return nil, fmt.Errorf("no backend configured for queue %q", queueName)
	}

	return backend, nil
}

// RemoveBackend removes a backend for a queue
func (m *BackendManager) RemoveBackend(queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	backend, exists := m.backends[queueName]
	if !exists {
		return fmt.Errorf("no backend configured for queue %q", queueName)
	}

	// Close the backend
	if err := backend.Close(); err != nil {
		return fmt.Errorf("failed to close backend for queue %q: %w", queueName, err)
	}

	// Remove from maps
	delete(m.backends, queueName)
	delete(m.configs, queueName)

	return nil
}

// ListQueues returns all configured queue names
func (m *BackendManager) ListQueues() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queues := make([]string, 0, len(m.backends))
	for queue := range m.backends {
		queues = append(queues, queue)
	}
	return queues
}

// HealthCheck performs health checks on all backends
func (m *BackendManager) HealthCheck(ctx context.Context) map[string]HealthStatus {
	m.mu.RLock()
	backends := make(map[string]QueueBackend, len(m.backends))
	for queue, backend := range m.backends {
		backends[queue] = backend
	}
	m.mu.RUnlock()

	results := make(map[string]HealthStatus)
	for queue, backend := range backends {
		status := backend.Health(ctx)
		status.CheckedAt = time.Now()
		results[queue] = status
	}

	return results
}

// Stats retrieves statistics from all backends
func (m *BackendManager) Stats(ctx context.Context) (map[string]*BackendStats, error) {
	m.mu.RLock()
	backends := make(map[string]QueueBackend, len(m.backends))
	for queue, backend := range m.backends {
		backends[queue] = backend
	}
	m.mu.RUnlock()

	results := make(map[string]*BackendStats)
	for queue, backend := range backends {
		stats, err := backend.Stats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for queue %q: %w", queue, err)
		}
		results[queue] = stats
	}

	return results, nil
}

// Close gracefully shuts down all backends
func (m *BackendManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for queue, backend := range m.backends {
		if err := backend.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close backend for queue %q: %w", queue, err))
		}
	}

	// Clear all backends
	m.backends = make(map[string]QueueBackend)
	m.configs = make(map[string]BackendConfig)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing backends: %v", errs)
	}

	return nil
}

// MigrationManager handles queue migrations between backends
type MigrationManager struct {
	backendManager *BackendManager
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(backendManager *BackendManager) *MigrationManager {
	return &MigrationManager{
		backendManager: backendManager,
	}
}

// Migrate moves jobs from one backend to another
func (m *MigrationManager) Migrate(ctx context.Context, queueName string, opts MigrationOptions) (*MigrationStatus, error) {
	// Get source and target backends
	sourceBackend, err := m.backendManager.GetBackend(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source backend: %w", err)
	}

	// For now, return a basic migration status
	// This would be expanded with actual migration logic
	status := &MigrationStatus{
		Phase:        MigrationPhaseValidation,
		TotalJobs:    0,
		MigratedJobs: 0,
		FailedJobs:   0,
		Progress:     0.0,
		StartedAt:    time.Now(),
	}

	// Validate compatibility
	caps := sourceBackend.Capabilities()
	if !caps.Persistence {
		return status, fmt.Errorf("source backend does not support persistence")
	}

	// Get queue length
	length, err := sourceBackend.Length(ctx)
	if err != nil {
		status.Phase = MigrationPhaseFailed
		status.LastError = err
		return status, fmt.Errorf("failed to get queue length: %w", err)
	}

	status.TotalJobs = length
	status.Phase = MigrationPhaseCompleted

	return status, nil
}

// Iterator implementations

// jobIterator implements Iterator interface
type jobIterator struct {
	jobs   []*Job
	index  int
	err    error
	closed bool
}

// NewJobIterator creates a new job iterator from a slice of jobs
func NewJobIterator(jobs []*Job) Iterator {
	return &jobIterator{
		jobs:  jobs,
		index: -1,
	}
}

// Next advances to the next job
func (i *jobIterator) Next() bool {
	if i.closed || i.err != nil {
		return false
	}

	i.index++
	return i.index < len(i.jobs)
}

// Job returns the current job
func (i *jobIterator) Job() *Job {
	if i.index < 0 || i.index >= len(i.jobs) {
		return nil
	}
	return i.jobs[i.index]
}

// Error returns any iteration error
func (i *jobIterator) Error() error {
	return i.err
}

// Close closes the iterator
func (i *jobIterator) Close() error {
	i.closed = true
	return nil
}

// Global registry instance
var defaultRegistry = NewBackendRegistry()

// DefaultRegistry returns the global backend registry
func DefaultRegistry() *BackendRegistry {
	return defaultRegistry
}

// RegisterBackend registers a backend factory in the global registry
func RegisterBackend(name string, factory BackendFactory) {
	defaultRegistry.Register(name, factory)
}

// CreateBackend creates a backend using the global registry
func CreateBackend(name string, config interface{}) (QueueBackend, error) {
	return defaultRegistry.Create(name, config)
}