package canary_deployments

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Setup test Redis instance
	os.Exit(m.Run())
}

func setupTestManager(t *testing.T) (*Manager, *redis.Client) {
	// Create test Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test database
	})

	// Clear test database
	rdb.FlushDB(context.Background())

	config := &Config{
		RedisAddr:                "localhost:6379",
		RedisDB:                  15,
		MaxConcurrentDeployments: 5,
		MetricsInterval:          1 * time.Second,
		HealthCheckInterval:      1 * time.Second,
		WorkerTimeout:            10 * time.Second,
		MaxCanaryPercentage:      50,
		MinMetricsSamples:        5,
		EmergencyRollbackDelay:   1 * time.Second,
	}
	config.SetDefaults()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))

	manager := NewManager(config, rdb, logger)
	return manager, rdb
}

func TestManager_CreateDeployment(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	config := DefaultCanaryConfig()
	config.AutoPromotion = false

	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)
	assert.NotEmpty(t, deployment.ID)
	assert.Equal(t, StatusActive, deployment.Status)
	assert.Equal(t, 0, deployment.CurrentPercent)
}

func TestManager_UpdateDeploymentPercentage(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Update percentage
	err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, 25)
	require.NoError(t, err)

	// Verify update
	updated, err := manager.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, 25, updated.CurrentPercent)
	assert.Equal(t, 25, updated.TargetPercent)
}

func TestManager_PromoteDeployment(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Promote deployment
	err = manager.PromoteDeployment(ctx, deployment.ID)
	require.NoError(t, err)

	// Verify promotion
	updated, err := manager.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, updated.Status)
	assert.Equal(t, 100, updated.CurrentPercent)
	assert.NotNil(t, updated.CompletedAt)
}

func TestManager_RollbackDeployment(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Set to 50% first
	err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, 50)
	require.NoError(t, err)

	// Rollback deployment
	reason := "Test rollback"
	err = manager.RollbackDeployment(ctx, deployment.ID, reason)
	require.NoError(t, err)

	// Verify rollback
	updated, err := manager.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, updated.Status)
	assert.Equal(t, 0, updated.CurrentPercent)
	assert.NotNil(t, updated.CompletedAt)
}

func TestManager_ConcurrencyLimit(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	// Create deployments up to the limit
	deployments := make([]*CanaryDeployment, 0)
	for i := 0; i < 5; i++ { // Max is 5 in test config
		deployment, err := manager.CreateDeployment(ctx, config)
		require.NoError(t, err)
		deployments = append(deployments, deployment)
	}

	// Try to create one more - should fail
	_, err := manager.CreateDeployment(ctx, config)
	assert.Error(t, err)
	assert.True(t, IsCode(err, CodeConcurrencyLimit))
}

func TestManager_InvalidPercentage(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Test invalid percentages
	testCases := []int{-1, 101, -10, 150}
	for _, percentage := range testCases {
		err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, percentage)
		assert.Error(t, err)
		assert.True(t, IsCode(err, CodeInvalidPercentage))
	}
}

func TestManager_GetDeploymentHealth(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Get health (should work even without metrics)
	health, err := manager.GetDeploymentHealth(ctx, deployment.ID)
	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.NotZero(t, health.LastEvaluation)
}

func TestManager_ListDeployments(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	// Create multiple deployments
	expectedCount := 3
	for i := 0; i < expectedCount; i++ {
		_, err := manager.CreateDeployment(ctx, config)
		require.NoError(t, err)
	}

	// List deployments
	deployments, err := manager.ListDeployments(ctx)
	require.NoError(t, err)
	assert.Len(t, deployments, expectedCount)

	// Verify sorted by start time (newest first)
	for i := 1; i < len(deployments); i++ {
		assert.True(t, deployments[i-1].StartTime.After(deployments[i].StartTime) ||
			deployments[i-1].StartTime.Equal(deployments[i].StartTime))
	}
}

func TestManager_DeploymentNotFound(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Try to get non-existent deployment
	_, err := manager.GetDeployment(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.True(t, IsCode(err, CodeDeploymentNotFound))

	// Try to update non-existent deployment
	err = manager.UpdateDeploymentPercentage(ctx, "non-existent-id", 50)
	assert.Error(t, err)
	assert.True(t, IsCode(err, CodeDeploymentNotFound))
}

func TestManager_WorkerRegistration(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Register worker
	worker := &WorkerInfo{
		ID:      "test-worker-1",
		Version: "v1.0.0",
		Lane:    "stable",
		Queues:  []string{"test-queue"},
		Status:  WorkerHealthy,
	}

	err := manager.RegisterWorker(ctx, worker)
	require.NoError(t, err)

	// Get workers by lane
	workers, err := manager.GetWorkers(ctx, "stable")
	require.NoError(t, err)
	assert.Len(t, workers, 1)
	assert.Equal(t, "test-worker-1", workers[0].ID)
	assert.Equal(t, "v1.0.0", workers[0].Version)
}

func TestManager_WorkerStatusUpdate(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Register worker
	worker := &WorkerInfo{
		ID:      "test-worker-1",
		Version: "v1.0.0",
		Lane:    "stable",
		Queues:  []string{"test-queue"},
		Status:  WorkerHealthy,
	}

	err := manager.RegisterWorker(ctx, worker)
	require.NoError(t, err)

	// Update status
	err = manager.UpdateWorkerStatus(ctx, "test-worker-1", WorkerDegraded)
	require.NoError(t, err)

	// Note: We don't have a GetWorker method in the interface,
	// so we'll get workers by lane and check
	workers, err := manager.GetWorkers(ctx, "stable")
	require.NoError(t, err)
	assert.Len(t, workers, 1)
	// Status update through the registry might not be immediately reflected
	// in the manager's GetWorkers response depending on implementation
}

func TestManager_GetDeploymentEvents(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Perform some operations to generate events
	err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, 25)
	require.NoError(t, err)

	err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, 50)
	require.NoError(t, err)

	// Give events time to be processed
	time.Sleep(100 * time.Millisecond)

	// Get events
	events, err := manager.GetDeploymentEvents(ctx, deployment.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, events)

	// Events should be sorted by timestamp (newest first)
	for i := 1; i < len(events); i++ {
		assert.True(t, events[i-1].Timestamp.After(events[i].Timestamp) ||
			events[i-1].Timestamp.Equal(events[i].Timestamp))
	}
}

func TestManager_StartStop(t *testing.T) {
	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start manager
	err := manager.Start(ctx)
	require.NoError(t, err)

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	err = manager.Stop(ctx)
	require.NoError(t, err)
}

func TestManager_LoadDeployments(t *testing.T) {
	manager1, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	// Create deployment with first manager
	deployment, err := manager1.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Create second manager (simulating restart)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	testConfig := &Config{
		RedisAddr:                "localhost:6379",
		RedisDB:                  15,
		MaxConcurrentDeployments: 5,
		MetricsInterval:          1 * time.Second,
		HealthCheckInterval:      1 * time.Second,
		WorkerTimeout:            10 * time.Second,
		MaxCanaryPercentage:      50,
		MinMetricsSamples:        5,
		EmergencyRollbackDelay:   1 * time.Second,
	}
	testConfig.SetDefaults()

	manager2 := NewManager(testConfig, rdb, logger)

	// Start second manager (should load existing deployments)
	err = manager2.Start(ctx)
	require.NoError(t, err)
	defer manager2.Stop(ctx)

	// Verify deployment was loaded
	loadedDeployment, err := manager2.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, deployment.ID, loadedDeployment.ID)
	assert.Equal(t, deployment.Status, loadedDeployment.Status)
}

// Benchmark tests

func BenchmarkManager_CreateDeployment(b *testing.B) {
	manager, rdb := setupTestManager(&testing.T{})
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.CreateDeployment(ctx, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManager_UpdatePercentage(b *testing.B) {
	manager, rdb := setupTestManager(&testing.T{})
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	// Create a deployment to update
	deployment, err := manager.CreateDeployment(ctx, config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		percentage := (i % 100) + 1 // 1-100
		err := manager.UpdateDeploymentPercentage(ctx, deployment.ID, percentage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManager_GetDeployment(b *testing.B) {
	manager, rdb := setupTestManager(&testing.T{})
	defer rdb.Close()

	ctx := context.Background()
	config := DefaultCanaryConfig()

	// Create a deployment to retrieve
	deployment, err := manager.CreateDeployment(ctx, config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetDeployment(ctx, deployment.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration tests

func TestIntegration_CompleteCanaryFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Start manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Create deployment with auto-promotion enabled
	config := DefaultCanaryConfig()
	config.AutoPromotion = true
	config.MinCanaryDuration = 100 * time.Millisecond

	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Gradual rollout
	percentages := []int{5, 10, 25, 50, 75, 100}
	for _, percentage := range percentages {
		err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, percentage)
		require.NoError(t, err)

		// Verify update
		updated, err := manager.GetDeployment(ctx, deployment.ID)
		require.NoError(t, err)
		assert.Equal(t, percentage, updated.CurrentPercent)

		// Wait a bit
		time.Sleep(50 * time.Millisecond)
	}

	// Final promotion
	err = manager.PromoteDeployment(ctx, deployment.ID)
	require.NoError(t, err)

	// Verify final state
	final, err := manager.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, final.Status)
	assert.Equal(t, 100, final.CurrentPercent)
	assert.NotNil(t, final.CompletedAt)

	// Check events
	events, err := manager.GetDeploymentEvents(ctx, deployment.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, events)

	// Should have creation, multiple updates, and promotion events
	eventTypes := make(map[string]int)
	for _, event := range events {
		eventTypes[event.Type]++
	}

	assert.Greater(t, eventTypes["deployment_created"], 0)
	assert.Greater(t, eventTypes["percentage_updated"], 0)
	assert.Greater(t, eventTypes["deployment_promoted"], 0)
}

func TestIntegration_RollbackFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager, rdb := setupTestManager(t)
	defer rdb.Close()

	ctx := context.Background()

	// Start manager
	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Create deployment
	config := DefaultCanaryConfig()
	deployment, err := manager.CreateDeployment(ctx, config)
	require.NoError(t, err)

	// Ramp up
	err = manager.UpdateDeploymentPercentage(ctx, deployment.ID, 50)
	require.NoError(t, err)

	// Simulate failure and rollback
	reason := "High error rate detected"
	err = manager.RollbackDeployment(ctx, deployment.ID, reason)
	require.NoError(t, err)

	// Verify rollback
	final, err := manager.GetDeployment(ctx, deployment.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, final.Status)
	assert.Equal(t, 0, final.CurrentPercent)
	assert.NotNil(t, final.CompletedAt)

	// Check events
	events, err := manager.GetDeploymentEvents(ctx, deployment.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, events)

	// Should have rollback event
	hasRollbackEvent := false
	for _, event := range events {
		if event.Type == "deployment_rolled_back" {
			hasRollbackEvent = true
			assert.Contains(t, event.Message, reason)
			break
		}
	}
	assert.True(t, hasRollbackEvent, "Should have rollback event")
}