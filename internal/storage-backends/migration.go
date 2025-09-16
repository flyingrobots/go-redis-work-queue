package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MigrationManager handles queue migrations between backends
type MigrationManager struct {
	backendManager *BackendManager
	activeMigrations map[string]*MigrationStatus
	mu             sync.RWMutex
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(backendManager *BackendManager) *MigrationManager {
	return &MigrationManager{
		backendManager:   backendManager,
		activeMigrations: make(map[string]*MigrationStatus),
	}
}

// StartMigration begins a migration from one backend to another
func (m *MigrationManager) StartMigration(ctx context.Context, queueName string, opts MigrationOptions) (*MigrationStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if migration is already in progress
	if _, exists := m.activeMigrations[queueName]; exists {
		return nil, ErrMigrationInProgress
	}

	// Get source backend
	sourceBackend, err := m.backendManager.GetBackend(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source backend: %w", err)
	}

	// Validate source backend capabilities
	sourceCaps := sourceBackend.Capabilities()
	if !sourceCaps.Persistence {
		return nil, fmt.Errorf("source backend does not support persistence")
	}

	// Create migration status
	status := &MigrationStatus{
		Phase:        MigrationPhaseValidation,
		TotalJobs:    0,
		MigratedJobs: 0,
		FailedJobs:   0,
		Progress:     0.0,
		StartedAt:    time.Now(),
	}

	// Store active migration
	m.activeMigrations[queueName] = status

	// Start migration in background
	go m.runMigration(ctx, queueName, sourceBackend, opts, status)

	return status, nil
}

// GetMigrationStatus returns the status of an ongoing migration
func (m *MigrationManager) GetMigrationStatus(queueName string) (*MigrationStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.activeMigrations[queueName]
	if !exists {
		return nil, fmt.Errorf("no migration in progress for queue %q", queueName)
	}

	return status, nil
}

// CancelMigration cancels an ongoing migration
func (m *MigrationManager) CancelMigration(queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.activeMigrations[queueName]
	if !exists {
		return fmt.Errorf("no migration in progress for queue %q", queueName)
	}

	status.Phase = MigrationPhaseFailed
	status.LastError = fmt.Errorf("migration cancelled by user")

	delete(m.activeMigrations, queueName)
	return nil
}

// ListActiveMigrations returns all active migrations
func (m *MigrationManager) ListActiveMigrations() map[string]*MigrationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*MigrationStatus)
	for queue, status := range m.activeMigrations {
		// Return a copy to avoid race conditions
		statusCopy := *status
		result[queue] = &statusCopy
	}

	return result
}

// runMigration executes the migration process
func (m *MigrationManager) runMigration(ctx context.Context, queueName string, sourceBackend QueueBackend, opts MigrationOptions, status *MigrationStatus) {
	defer func() {
		m.mu.Lock()
		delete(m.activeMigrations, queueName)
		m.mu.Unlock()
	}()

	// Phase 1: Validation
	if err := m.validateMigration(ctx, sourceBackend, opts, status); err != nil {
		status.Phase = MigrationPhaseFailed
		status.LastError = err
		return
	}

	// Phase 2: Get total job count
	totalJobs, err := sourceBackend.Length(ctx)
	if err != nil {
		status.Phase = MigrationPhaseFailed
		status.LastError = fmt.Errorf("failed to get queue length: %w", err)
		return
	}

	status.TotalJobs = totalJobs
	status.Phase = MigrationPhaseDraining

	// Phase 3: Drain source if requested
	if opts.DrainFirst {
		if err := m.drainSource(ctx, sourceBackend, status); err != nil {
			status.Phase = MigrationPhaseFailed
			status.LastError = err
			return
		}
	}

	// Phase 4: Copy jobs
	status.Phase = MigrationPhaseCopying
	if err := m.copyJobs(ctx, queueName, sourceBackend, opts, status); err != nil {
		status.Phase = MigrationPhaseFailed
		status.LastError = err
		return
	}

	// Phase 5: Verification
	status.Phase = MigrationPhaseVerifying
	if opts.VerifyData {
		if err := m.verifyMigration(ctx, queueName, opts, status); err != nil {
			status.Phase = MigrationPhaseFailed
			status.LastError = err
			return
		}
	}

	// Complete migration
	status.Phase = MigrationPhaseCompleted
	status.Progress = 100.0
}

// validateMigration validates that the migration can proceed
func (m *MigrationManager) validateMigration(ctx context.Context, sourceBackend QueueBackend, opts MigrationOptions, status *MigrationStatus) error {
	// Check source backend health
	health := sourceBackend.Health(ctx)
	if health.Status == HealthStatusUnhealthy {
		return fmt.Errorf("source backend is unhealthy: %s", health.Message)
	}

	// Validate target backend exists
	if _, err := m.backendManager.GetBackend(opts.TargetBackend); err != nil {
		return fmt.Errorf("target backend not found: %w", err)
	}

	// Check capabilities compatibility
	// (This would be expanded based on specific backend requirements)

	status.Progress = 5.0
	return nil
}

// drainSource stops new jobs from being added to the source
func (m *MigrationManager) drainSource(ctx context.Context, sourceBackend QueueBackend, status *MigrationStatus) error {
	// In a real implementation, this might involve:
	// 1. Marking the queue as read-only
	// 2. Waiting for in-flight jobs to complete
	// 3. Ensuring no new jobs are added

	// For now, we'll simulate a drain period
	drainTime := 10 * time.Second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			elapsed := time.Since(start)
			if elapsed >= drainTime {
				status.Progress = 15.0
				return nil
			}
			// Update progress during drain
			drainProgress := float64(elapsed) / float64(drainTime) * 10.0
			status.Progress = 5.0 + drainProgress
		}
	}
}

// copyJobs copies jobs from source to target backend
func (m *MigrationManager) copyJobs(ctx context.Context, queueName string, sourceBackend QueueBackend, opts MigrationOptions, status *MigrationStatus) error {
	// Create target backend
	targetBackend, err := m.backendManager.GetBackend(opts.TargetBackend)
	if err != nil {
		return fmt.Errorf("failed to get target backend: %w", err)
	}

	// Get iterator for source jobs
	iter, err := sourceBackend.Iter(ctx, IterOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job iterator: %w", err)
	}
	defer iter.Close()

	// Process jobs in batches
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	var batch []*Job
	processed := int64(0)

	for iter.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		job := iter.Job()
		if job == nil {
			continue
		}

		batch = append(batch, job)

		// Process batch when full or at end
		if len(batch) >= batchSize {
			if err := m.processBatch(ctx, targetBackend, batch, opts, status); err != nil {
				return err
			}

			processed += int64(len(batch))
			batch = batch[:0] // Reset batch

			// Update progress
			if status.TotalJobs > 0 {
				progress := float64(processed) / float64(status.TotalJobs) * 70.0 // 70% of total progress for copying
				status.Progress = 15.0 + progress
				status.MigratedJobs = processed
			}

			// Update ETA
			if processed > 0 {
				elapsed := time.Since(status.StartedAt)
				rate := float64(processed) / elapsed.Seconds()
				remaining := status.TotalJobs - processed
				etaSeconds := float64(remaining) / rate
				status.EstimatedETA = time.Now().Add(time.Duration(etaSeconds) * time.Second)
			}
		}
	}

	// Process remaining jobs in batch
	if len(batch) > 0 {
		if err := m.processBatch(ctx, targetBackend, batch, opts, status); err != nil {
			return err
		}
		processed += int64(len(batch))
		status.MigratedJobs = processed
	}

	// Check for iteration errors
	if err := iter.Error(); err != nil {
		return fmt.Errorf("error during job iteration: %w", err)
	}

	status.Progress = 85.0
	return nil
}

// processBatch processes a batch of jobs
func (m *MigrationManager) processBatch(ctx context.Context, targetBackend QueueBackend, batch []*Job, opts MigrationOptions, status *MigrationStatus) error {
	for _, job := range batch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if opts.DryRun {
			// In dry run mode, just simulate the migration
			continue
		}

		// Copy job to target backend
		if err := targetBackend.Enqueue(ctx, job); err != nil {
			status.FailedJobs++
			status.LastError = NewMigrationError(
				status.Phase,
				opts.SourceBackend,
				opts.TargetBackend,
				job.ID,
				"failed to enqueue job to target",
				err,
			)
			// Continue with other jobs unless we exceed failure threshold
			if status.FailedJobs > status.TotalJobs/10 { // More than 10% failures
				return status.LastError
			}
		}
	}

	return nil
}

// verifyMigration verifies that the migration was successful
func (m *MigrationManager) verifyMigration(ctx context.Context, queueName string, opts MigrationOptions, status *MigrationStatus) error {
	// Get both backends
	sourceBackend, err := m.backendManager.GetBackend(queueName)
	if err != nil {
		return fmt.Errorf("failed to get source backend: %w", err)
	}

	targetBackend, err := m.backendManager.GetBackend(opts.TargetBackend)
	if err != nil {
		return fmt.Errorf("failed to get target backend: %w", err)
	}

	// Compare queue lengths
	sourceLength, err := sourceBackend.Length(ctx)
	if err != nil {
		return fmt.Errorf("failed to get source queue length: %w", err)
	}

	targetLength, err := targetBackend.Length(ctx)
	if err != nil {
		return fmt.Errorf("failed to get target queue length: %w", err)
	}

	// In a real implementation, we might allow some tolerance for active processing
	if !opts.DrainFirst && targetLength < sourceLength {
		return fmt.Errorf("target queue length (%d) is less than source (%d)", targetLength, sourceLength)
	}

	// Additional verification could include:
	// - Spot checking specific jobs
	// - Verifying job data integrity
	// - Checking that critical jobs were migrated

	status.Progress = 95.0
	return nil
}

// MigrationTool provides command-line style migration utilities
type MigrationTool struct {
	manager *MigrationManager
}

// NewMigrationTool creates a new migration tool
func NewMigrationTool(backendManager *BackendManager) *MigrationTool {
	return &MigrationTool{
		manager: NewMigrationManager(backendManager),
	}
}

// PlanMigration analyzes a migration without executing it
func (t *MigrationTool) PlanMigration(ctx context.Context, queueName string, opts MigrationOptions) (*MigrationPlan, error) {
	// Get source backend info
	sourceBackend, err := t.manager.backendManager.GetBackend(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source backend: %w", err)
	}

	// Get queue length
	length, err := sourceBackend.Length(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue length: %w", err)
	}

	// Get source stats
	stats, err := sourceBackend.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}

	// Calculate estimates
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	// Estimate duration based on typical processing rates
	estimatedRate := float64(batchSize) * 10.0 // 10 batches per second
	estimatedDuration := time.Duration(float64(length)/estimatedRate) * time.Second

	plan := &MigrationPlan{
		QueueName:         queueName,
		SourceBackend:     opts.SourceBackend,
		TargetBackend:     opts.TargetBackend,
		JobCount:          length,
		EstimatedDuration: estimatedDuration,
		BatchSize:         batchSize,
		Requirements:      []string{},
		Warnings:          []string{},
		Recommendations:   []string{},
	}

	// Add requirements based on source backend
	sourceCaps := sourceBackend.Capabilities()
	if !sourceCaps.Persistence {
		plan.Warnings = append(plan.Warnings, "Source backend does not guarantee persistence")
	}

	// Add recommendations
	if length > 10000 {
		plan.Recommendations = append(plan.Recommendations, "Consider draining source first for large migrations")
	}

	if stats.ErrorRate > 0.01 {
		plan.Warnings = append(plan.Warnings, "Source backend has elevated error rate")
	}

	return plan, nil
}

// MigrationPlan describes a planned migration
type MigrationPlan struct {
	QueueName         string        `json:"queue_name"`
	SourceBackend     string        `json:"source_backend"`
	TargetBackend     string        `json:"target_backend"`
	JobCount          int64         `json:"job_count"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	BatchSize         int           `json:"batch_size"`
	Requirements      []string      `json:"requirements"`
	Warnings          []string      `json:"warnings"`
	Recommendations   []string      `json:"recommendations"`
}

// ExecuteMigration executes a planned migration
func (t *MigrationTool) ExecuteMigration(ctx context.Context, queueName string, opts MigrationOptions) (*MigrationStatus, error) {
	return t.manager.StartMigration(ctx, queueName, opts)
}

// MonitorMigration monitors an ongoing migration
func (t *MigrationTool) MonitorMigration(queueName string) (*MigrationStatus, error) {
	return t.manager.GetMigrationStatus(queueName)
}

// QuickMigrate performs a simple migration with sensible defaults
func (t *MigrationTool) QuickMigrate(ctx context.Context, queueName, targetBackend string) (*MigrationStatus, error) {
	opts := MigrationOptions{
		TargetBackend: targetBackend,
		DrainFirst:    false,
		Timeout:       30 * time.Minute,
		BatchSize:     100,
		VerifyData:    true,
		DryRun:        false,
	}

	return t.manager.StartMigration(ctx, queueName, opts)
}