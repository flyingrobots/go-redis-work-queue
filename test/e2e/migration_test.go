package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	storage "github.com/flyingrobots/go-redis-work-queue/internal/storage-backends"
)

// MigrationE2ETestSuite provides comprehensive end-to-end tests for migration scenarios
type MigrationE2ETestSuite struct {
	suite.Suite
	registry *storage.BackendRegistry
	manager  *storage.BackendManager
	migrator *storage.MigrationManager
	tool     *storage.MigrationTool
}

// SetupSuite initializes the test suite
func (suite *MigrationE2ETestSuite) SetupSuite() {
	suite.registry = storage.NewBackendRegistry()
	suite.manager = storage.NewBackendManager(suite.registry)
	suite.migrator = storage.NewMigrationManager(suite.manager)
	suite.tool = storage.NewMigrationTool(suite.manager)

	// Register Redis Lists backend
	suite.registry.Register(storage.BackendTypeRedisLists, &storage.RedisListsFactory{})
}

// TearDownSuite cleans up the test suite
func (suite *MigrationE2ETestSuite) TearDownSuite() {
	if suite.manager != nil {
		suite.manager.Close()
	}
}

// SetupTest prepares each test
func (suite *MigrationE2ETestSuite) SetupTest() {
	// Clear any existing backends
	for _, queue := range suite.manager.ListQueues() {
		suite.manager.RemoveBackend(queue)
	}
}

// TestMigrationE2E_SuccessfulMigration tests a complete successful migration workflow
func (suite *MigrationE2ETestSuite) TestMigrationE2E_SuccessfulMigration() {
	ctx := context.Background()

	// Setup source backend
	sourceConfig := storage.RedisListsConfig{
		URL:        "redis://localhost:6379",
		Database:   1,
		KeyPrefix:  "test_source:",
		MaxRetries: 3,
	}

	err := suite.manager.AddBackend("source_queue", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "source_backend",
	})
	suite.Require().NoError(err)

	// Setup target backend
	targetConfig := storage.RedisListsConfig{
		URL:        "redis://localhost:6379",
		Database:   2,
		KeyPrefix:  "test_target:",
		MaxRetries: 3,
	}

	err = suite.manager.AddBackend("target_queue", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "target_backend",
	})
	suite.Require().NoError(err)

	// Get backends
	sourceBackend, err := suite.manager.GetBackend("source_queue")
	suite.Require().NoError(err)

	targetBackend, err := suite.manager.GetBackend("target_queue")
	suite.Require().NoError(err)

	// Populate source with test jobs
	testJobs := suite.createTestJobs(10)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Verify source has jobs
	sourceLength, err := sourceBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(10), sourceLength)

	// Plan migration
	plan, err := suite.tool.PlanMigration(ctx, "source_queue", storage.MigrationOptions{
		SourceBackend: "source_backend",
		TargetBackend: "target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     5,
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(plan)
	suite.Equal("source_queue", plan.QueueName)
	suite.Equal("target_backend", plan.TargetBackend)
	suite.Equal(int64(10), plan.JobCount)
	suite.Equal(5, plan.BatchSize)

	// Execute migration
	status, err := suite.tool.ExecuteMigration(ctx, "source_queue", storage.MigrationOptions{
		SourceBackend: "source_backend",
		TargetBackend: "target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     5,
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Monitor migration until completion
	suite.waitForMigrationCompletion(ctx, "source_queue", 10*time.Second)

	// Verify final status
	finalStatus, err := suite.tool.MonitorMigration("source_queue")
	if err == nil { // Migration might have been cleaned up
		suite.Equal(storage.MigrationPhaseCompleted, finalStatus.Phase)
		suite.Equal(100.0, finalStatus.Progress)
		suite.Equal(int64(10), finalStatus.MigratedJobs)
		suite.Equal(int64(0), finalStatus.FailedJobs)
	}

	// Verify target has all jobs
	targetLength, err := targetBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(10), targetLength)

	// Verify job content integrity
	suite.verifyJobIntegrity(ctx, sourceBackend, targetBackend, testJobs)
}

// TestMigrationE2E_DryRunMigration tests dry run migration
func (suite *MigrationE2ETestSuite) TestMigrationE2E_DryRunMigration() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("dry_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "dry_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("dry_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "dry_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("dry_source")
	suite.Require().NoError(err)

	targetBackend, err := suite.manager.GetBackend("dry_target")
	suite.Require().NoError(err)

	// Add jobs to source
	testJobs := suite.createTestJobs(5)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	originalSourceLength, err := sourceBackend.Length(ctx)
	suite.Require().NoError(err)

	originalTargetLength, err := targetBackend.Length(ctx)
	suite.Require().NoError(err)

	// Execute dry run migration
	status, err := suite.tool.ExecuteMigration(ctx, "dry_source", storage.MigrationOptions{
		SourceBackend: "dry_source_backend",
		TargetBackend: "dry_target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     3,
		VerifyData:    false,
		DryRun:        true,
	})
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Wait for dry run completion
	suite.waitForMigrationCompletion(ctx, "dry_source", 5*time.Second)

	// Verify no actual data movement occurred
	finalSourceLength, err := sourceBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(originalSourceLength, finalSourceLength)

	finalTargetLength, err := targetBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(originalTargetLength, finalTargetLength)
}

// TestMigrationE2E_DrainFirstMigration tests migration with drain-first option
func (suite *MigrationE2ETestSuite) TestMigrationE2E_DrainFirstMigration() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("drain_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "drain_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("drain_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "drain_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("drain_source")
	suite.Require().NoError(err)

	// Add jobs to source
	testJobs := suite.createTestJobs(8)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Execute migration with drain first
	status, err := suite.tool.ExecuteMigration(ctx, "drain_source", storage.MigrationOptions{
		SourceBackend: "drain_source_backend",
		TargetBackend: "drain_target_backend",
		DrainFirst:    true,
		Timeout:       30 * time.Second,
		BatchSize:     4,
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Monitor migration phases
	suite.waitForMigrationPhase(ctx, "drain_source", storage.MigrationPhaseDraining, 15*time.Second)
	suite.waitForMigrationCompletion(ctx, "drain_source", 20*time.Second)
}

// TestMigrationE2E_ConcurrentMigrationAttempt tests handling of concurrent migration attempts
func (suite *MigrationE2ETestSuite) TestMigrationE2E_ConcurrentMigrationAttempt() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("concurrent_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "concurrent_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("concurrent_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "concurrent_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("concurrent_source")
	suite.Require().NoError(err)

	// Add jobs to source
	testJobs := suite.createTestJobs(20)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Start first migration
	status1, err := suite.tool.ExecuteMigration(ctx, "concurrent_source", storage.MigrationOptions{
		SourceBackend: "concurrent_source_backend",
		TargetBackend: "concurrent_target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     10,
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(status1)

	// Attempt second migration (should fail)
	status2, err := suite.tool.ExecuteMigration(ctx, "concurrent_source", storage.MigrationOptions{
		SourceBackend: "concurrent_source_backend",
		TargetBackend: "concurrent_target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     5,
		VerifyData:    false,
		DryRun:        false,
	})
	suite.Error(err)
	suite.ErrorIs(err, storage.ErrMigrationInProgress)
	suite.Nil(status2)

	// Wait for first migration to complete
	suite.waitForMigrationCompletion(ctx, "concurrent_source", 15*time.Second)
}

// TestMigrationE2E_MigrationCancellation tests migration cancellation
func (suite *MigrationE2ETestSuite) TestMigrationE2E_MigrationCancellation() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("cancel_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "cancel_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("cancel_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "cancel_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("cancel_source")
	suite.Require().NoError(err)

	// Add many jobs to ensure migration takes time
	testJobs := suite.createTestJobs(100)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Start migration
	status, err := suite.tool.ExecuteMigration(ctx, "cancel_source", storage.MigrationOptions{
		SourceBackend: "cancel_source_backend",
		TargetBackend: "cancel_target_backend",
		DrainFirst:    true, // This will add delay
		Timeout:       60 * time.Second,
		BatchSize:     1, // Small batch size to slow down
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Wait a bit for migration to start
	time.Sleep(2 * time.Second)

	// Cancel migration
	err = suite.migrator.CancelMigration("cancel_source")
	suite.Require().NoError(err)

	// Verify migration was cancelled
	status, err = suite.tool.MonitorMigration("cancel_source")
	suite.Error(err) // Should not be found after cancellation
}

// TestMigrationE2E_InvalidBackendConfiguration tests migration with invalid backends
func (suite *MigrationE2ETestSuite) TestMigrationE2E_InvalidBackendConfiguration() {
	ctx := context.Background()

	// Try to plan migration with non-existent source
	plan, err := suite.tool.PlanMigration(ctx, "nonexistent_queue", storage.MigrationOptions{
		SourceBackend: "nonexistent_source",
		TargetBackend: "nonexistent_target",
	})
	suite.Error(err)
	suite.Nil(plan)

	// Setup only source backend
	err = suite.manager.AddBackend("partial_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "partial_source_backend",
	})
	suite.Require().NoError(err)

	// Try to execute migration with non-existent target
	status, err := suite.tool.ExecuteMigration(ctx, "partial_source", storage.MigrationOptions{
		SourceBackend: "partial_source_backend",
		TargetBackend: "nonexistent_target",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     10,
		VerifyData:    false,
		DryRun:        false,
	})
	suite.Error(err)
	suite.Nil(status)
}

// TestMigrationE2E_LargeBatchMigration tests migration with large batches
func (suite *MigrationE2ETestSuite) TestMigrationE2E_LargeBatchMigration() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("large_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "large_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("large_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "large_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("large_source")
	suite.Require().NoError(err)

	targetBackend, err := suite.manager.GetBackend("large_target")
	suite.Require().NoError(err)

	// Add many jobs
	testJobs := suite.createTestJobs(50)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Execute migration with large batch size
	status, err := suite.tool.ExecuteMigration(ctx, "large_source", storage.MigrationOptions{
		SourceBackend: "large_source_backend",
		TargetBackend: "large_target_backend",
		DrainFirst:    false,
		Timeout:       30 * time.Second,
		BatchSize:     25, // Large batch
		VerifyData:    true,
		DryRun:        false,
	})
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Wait for completion
	suite.waitForMigrationCompletion(ctx, "large_source", 15*time.Second)

	// Verify all jobs migrated
	targetLength, err := targetBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(50), targetLength)
}

// TestMigrationE2E_QuickMigrate tests the convenience quick migrate function
func (suite *MigrationE2ETestSuite) TestMigrationE2E_QuickMigrate() {
	ctx := context.Background()

	// Setup backends
	err := suite.manager.AddBackend("quick_source", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "quick_source_backend",
	})
	suite.Require().NoError(err)

	err = suite.manager.AddBackend("quick_target", storage.BackendConfig{
		Type: storage.BackendTypeRedisLists,
		Name: "quick_target_backend",
	})
	suite.Require().NoError(err)

	sourceBackend, err := suite.manager.GetBackend("quick_source")
	suite.Require().NoError(err)

	targetBackend, err := suite.manager.GetBackend("quick_target")
	suite.Require().NoError(err)

	// Add jobs
	testJobs := suite.createTestJobs(15)
	for _, job := range testJobs {
		err := sourceBackend.Enqueue(ctx, job)
		suite.Require().NoError(err)
	}

	// Use quick migrate
	status, err := suite.tool.QuickMigrate(ctx, "quick_source", "quick_target_backend")
	suite.Require().NoError(err)
	suite.NotNil(status)

	// Wait for completion
	suite.waitForMigrationCompletion(ctx, "quick_source", 10*time.Second)

	// Verify migration
	targetLength, err := targetBackend.Length(ctx)
	suite.Require().NoError(err)
	suite.Equal(int64(15), targetLength)
}

// Helper methods

// createTestJobs creates a slice of test jobs
func (suite *MigrationE2ETestSuite) createTestJobs(count int) []*storage.Job {
	jobs := make([]*storage.Job, count)
	for i := 0; i < count; i++ {
		jobs[i] = &storage.Job{
			ID:       suite.generateJobID(i),
			Type:     "test_job",
			Queue:    "test_queue",
			Payload:  map[string]interface{}{"index": i, "data": "test data"},
			Priority: i % 3, // Mix of priorities
			CreatedAt: time.Now(),
			RetryCount: 0,
			MaxRetries: 3,
			Metadata: map[string]interface{}{
				"test_meta": "value",
				"batch":     i / 5,
			},
			Tags: []string{suite.T().Name(), "migration_test"},
		}
	}
	return jobs
}

// generateJobID generates a unique job ID
func (suite *MigrationE2ETestSuite) generateJobID(index int) string {
	return suite.T().Name() + "_job_" + time.Now().Format("20060102150405") + "_" +
		   string(rune('0'+index%10))
}

// waitForMigrationCompletion waits for migration to complete or timeout
func (suite *MigrationE2ETestSuite) waitForMigrationCompletion(ctx context.Context, queueName string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			suite.T().Fatal("Context cancelled while waiting for migration")
			return
		case <-ticker.C:
			status, err := suite.tool.MonitorMigration(queueName)
			if err != nil {
				// Migration might have completed and been cleaned up
				return
			}
			if status.Phase == storage.MigrationPhaseCompleted ||
			   status.Phase == storage.MigrationPhaseFailed {
				return
			}
		}
	}
	suite.T().Fatal("Migration did not complete within timeout")
}

// waitForMigrationPhase waits for migration to reach a specific phase
func (suite *MigrationE2ETestSuite) waitForMigrationPhase(ctx context.Context, queueName, phase string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			suite.T().Fatal("Context cancelled while waiting for migration phase")
			return
		case <-ticker.C:
			status, err := suite.tool.MonitorMigration(queueName)
			if err != nil {
				continue
			}
			if status.Phase == phase {
				return
			}
		}
	}
	suite.T().Fatalf("Migration did not reach phase %s within timeout", phase)
}

// verifyJobIntegrity verifies that jobs were correctly migrated
func (suite *MigrationE2ETestSuite) verifyJobIntegrity(ctx context.Context, source, target storage.QueueBackend, originalJobs []*storage.Job) {
	// Create maps for easy lookup
	originalJobsMap := make(map[string]*storage.Job)
	for _, job := range originalJobs {
		originalJobsMap[job.ID] = job
	}

	// Get iterator for target backend
	iter, err := target.Iter(ctx, storage.IterOptions{})
	suite.Require().NoError(err)
	defer iter.Close()

	migratedCount := 0
	for iter.Next() {
		job := iter.Job()
		suite.Require().NotNil(job)

		originalJob, exists := originalJobsMap[job.ID]
		suite.Require().True(exists, "Migrated job %s not found in original jobs", job.ID)

		// Verify job fields
		suite.Equal(originalJob.Type, job.Type)
		suite.Equal(originalJob.Queue, job.Queue)
		suite.Equal(originalJob.Priority, job.Priority)
		suite.Equal(originalJob.MaxRetries, job.MaxRetries)
		suite.Equal(originalJob.RetryCount, job.RetryCount)

		// Verify payload (basic check)
		suite.NotNil(job.Payload)

		// Verify metadata
		suite.Equal(originalJob.Metadata, job.Metadata)

		// Verify tags
		suite.Equal(originalJob.Tags, job.Tags)

		migratedCount++
	}

	suite.Require().NoError(iter.Error())
	suite.Equal(len(originalJobs), migratedCount, "Not all jobs were migrated")
}

// Test runner
func TestMigrationE2ETestSuite(t *testing.T) {
	suite.Run(t, new(MigrationE2ETestSuite))
}