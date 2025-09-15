package multicluster

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestE2E_ProductionScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup realistic production-like environment
	prodEast := miniredis.RunT(t)
	defer prodEast.Close()
	prodWest := miniredis.RunT(t)
	defer prodWest.Close()
	staging := miniredis.RunT(t)
	defer staging.Close()

	// Simulate production workload on east
	for i := 0; i < 50; i++ {
		prodEast.Lpush("jobqueue:queue:high", fmt.Sprintf("urgent-job-%d", i))
	}
	for i := 0; i < 100; i++ {
		prodEast.Lpush("jobqueue:queue:normal", fmt.Sprintf("normal-job-%d", i))
	}
	// Add some failed jobs
	for i := 0; i < 5; i++ {
		prodEast.Lpush("jobqueue:dead_letter", fmt.Sprintf("failed-job-%d", i))
	}
	// Simulate active workers
	for i := 1; i <= 10; i++ {
		prodEast.Set(fmt.Sprintf("worker:heartbeat:%d", i), "alive", time.Hour)
	}

	// Simulate production workload on west (lighter)
	for i := 0; i < 20; i++ {
		prodWest.Lpush("jobqueue:queue:high", fmt.Sprintf("urgent-job-west-%d", i))
	}
	for i := 0; i < 40; i++ {
		prodWest.Lpush("jobqueue:queue:normal", fmt.Sprintf("normal-job-west-%d", i))
	}
	// Fewer workers in west
	for i := 1; i <= 5; i++ {
		prodWest.Set(fmt.Sprintf("worker:heartbeat:west-%d", i), "alive", time.Hour)
	}

	// Staging has minimal load
	staging.Lpush("jobqueue:queue:normal", "test-job-1", "test-job-2")
	staging.Set("worker:heartbeat:staging-1", "alive", time.Hour)

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:        "prod-us-east",
				Label:       "Production US East",
				Color:       "#ff0000",
				Environment: "production",
				Region:      "us-east-1",
				Endpoint:    prodEast.Addr(),
				Enabled:     true,
				Tags:        []string{"production", "primary"},
			},
			{
				Name:        "prod-us-west",
				Label:       "Production US West",
				Color:       "#ff6600",
				Environment: "production",
				Region:      "us-west-2",
				Endpoint:    prodWest.Addr(),
				Enabled:     true,
				Tags:        []string{"production", "secondary"},
			},
			{
				Name:        "staging",
				Label:       "Staging Environment",
				Color:       "#00ff00",
				Environment: "staging",
				Region:      "us-east-1",
				Endpoint:    staging.Addr(),
				Enabled:     true,
				Tags:        []string{"staging"},
			},
		},
		DefaultCluster: "prod-us-east",
		Polling: PollingConfig{
			Enabled:  true,
			Interval: Duration(5 * time.Second),
			Jitter:   Duration(1 * time.Second),
			Timeout:  Duration(3 * time.Second),
		},
		Cache: CacheConfig{
			Enabled:         true,
			TTL:             Duration(2 * time.Minute),
			MaxEntries:      1000,
			CleanupInterval: Duration(30 * time.Second),
		},
		Actions: ActionsConfig{
			RequireConfirmation: true,
			MaxConcurrent:       5,
			DefaultTimeout:      Duration(30 * time.Second),
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
				ActionTypePauseQueue,
				ActionTypeResumeQueue,
			},
		},
		CompareMode: CompareModeConfig{
			Enabled:        true,
			DeltaThreshold: 25.0,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// E2E Test 1: Multi-cluster overview and health check
	t.Run("MultiClusterOverview", func(t *testing.T) {
		allStats, err := manager.GetAllStats(ctx)
		require.NoError(t, err)
		assert.Len(t, allStats, 3)

		// Verify production east has high load
		eastStats := allStats["prod-us-east"]
		require.NotNil(t, eastStats)
		assert.Equal(t, 10, eastStats.WorkerCount)

		// Verify production west has medium load
		westStats := allStats["prod-us-west"]
		require.NotNil(t, westStats)
		assert.Equal(t, 5, westStats.WorkerCount)

		// Verify staging has low load
		stagingStats := allStats["staging"]
		require.NotNil(t, stagingStats)
		assert.Equal(t, 1, stagingStats.WorkerCount)

		// Check health of all clusters
		for clusterName := range allStats {
			health, err := manager.GetHealth(ctx, clusterName)
			require.NoError(t, err)

			if clusterName == "prod-us-east" {
				// East should have issues due to DLQ
				assert.Contains(t, health.Issues, "High dead letter count")
			} else {
				// Other clusters should be healthy
				assert.True(t, health.Healthy, "Cluster %s should be healthy", clusterName)
			}
		}
	})

	// E2E Test 2: Cross-cluster comparison and anomaly detection
	t.Run("CrossClusterComparison", func(t *testing.T) {
		// Compare production clusters
		result, err := manager.CompareClusters(ctx, []string{"prod-us-east", "prod-us-west"})
		require.NoError(t, err)

		// Should detect worker count difference
		workerMetric, exists := result.Metrics["worker_count"]
		assert.True(t, exists)
		assert.Equal(t, 10.0, workerMetric.Values["prod-us-east"])
		assert.Equal(t, 5.0, workerMetric.Values["prod-us-west"])
		assert.Equal(t, 5.0, workerMetric.Delta)

		// Should detect anomalies due to significant differences
		assert.NotEmpty(t, result.Anomalies)

		// Enable compare mode
		err = manager.SetCompareMode(ctx, true, []string{"prod-us-east", "prod-us-west"})
		require.NoError(t, err)

		tabConfig, err := manager.GetTabConfig(ctx)
		require.NoError(t, err)
		assert.True(t, tabConfig.CompareMode)
		assert.Len(t, tabConfig.CompareWith, 2)
	})

	// E2E Test 3: Emergency DLQ cleanup across production clusters
	t.Run("EmergencyDLQCleanup", func(t *testing.T) {
		// Verify DLQ exists in prod-east
		eastStats, err := manager.GetStats(ctx, "prod-us-east")
		require.NoError(t, err)
		assert.Greater(t, eastStats.DeadLetterCount, 0)

		// Execute DLQ purge on production clusters
		action := &MultiAction{
			ID:      "emergency-dlq-cleanup",
			Type:    ActionTypePurgeDLQ,
			Targets: []string{"prod-us-east", "prod-us-west"},
			Parameters: map[string]interface{}{
				"reason": "Emergency cleanup due to high DLQ count",
			},
			Status:    ActionStatusConfirmed, // Skip confirmation for test
			CreatedAt: time.Now(),
		}

		err = manager.ExecuteAction(ctx, action)
		require.NoError(t, err)
		assert.Equal(t, ActionStatusCompleted, action.Status)

		// Verify all targets succeeded
		for _, target := range action.Targets {
			result, exists := action.Results[target]
			assert.True(t, exists, "Missing result for %s", target)
			assert.True(t, result.Success, "Action failed for %s: %s", target, result.Error)
		}

		// Verify DLQ was actually cleared
		eastStatsAfter, err := manager.GetStats(ctx, "prod-us-east")
		require.NoError(t, err)
		assert.Equal(t, 0, eastStatsAfter.DeadLetterCount)
	})

	// E2E Test 4: Staged deployment simulation
	t.Run("StagedDeployment", func(t *testing.T) {
		// Step 1: Test on staging first
		stagingAction := &MultiAction{
			ID:      "staging-test",
			Type:    ActionTypeBenchmark,
			Targets: []string{"staging"},
			Parameters: map[string]interface{}{
				"iterations": float64(10),
			},
			Status: ActionStatusConfirmed,
		}

		err := manager.ExecuteAction(ctx, stagingAction)
		require.NoError(t, err)
		assert.Equal(t, ActionStatusCompleted, stagingAction.Status)

		// Step 2: If staging succeeds, deploy to secondary production
		if stagingAction.Status == ActionStatusCompleted {
			westAction := &MultiAction{
				ID:      "west-deployment",
				Type:    ActionTypeBenchmark,
				Targets: []string{"prod-us-west"},
				Parameters: map[string]interface{}{
					"iterations": float64(15),
				},
				Status: ActionStatusConfirmed,
			}

			err := manager.ExecuteAction(ctx, westAction)
			require.NoError(t, err)
			assert.Equal(t, ActionStatusCompleted, westAction.Status)

			// Step 3: Finally deploy to primary production
			eastAction := &MultiAction{
				ID:      "east-deployment",
				Type:    ActionTypeBenchmark,
				Targets: []string{"prod-us-east"},
				Parameters: map[string]interface{}{
					"iterations": float64(20),
				},
				Status: ActionStatusConfirmed,
			}

			err := manager.ExecuteAction(ctx, eastAction)
			require.NoError(t, err)
			assert.Equal(t, ActionStatusCompleted, eastAction.Status)
		}
	})

	// E2E Test 5: Cluster switching and navigation
	t.Run("ClusterNavigation", func(t *testing.T) {
		// Switch between clusters
		clusters := []string{"prod-us-east", "prod-us-west", "staging"}

		for _, cluster := range clusters {
			err := manager.SwitchCluster(ctx, cluster)
			require.NoError(t, err)

			// Verify we can get data from the active cluster
			stats, err := manager.GetStats(ctx, cluster)
			require.NoError(t, err)
			assert.Equal(t, cluster, stats.ClusterName)

			// Get tab configuration
			tabConfig, err := manager.GetTabConfig(ctx)
			require.NoError(t, err)
			assert.Len(t, tabConfig.Tabs, 3)

			// Find active tab
			var activeCluster string
			for _, tab := range tabConfig.Tabs {
				if tab.Index-1 == tabConfig.ActiveTab {
					activeCluster = tab.ClusterName
					break
				}
			}
			assert.Equal(t, cluster, activeCluster)
		}
	})
}

func TestE2E_DisasterRecoveryScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup primary and backup clusters
	primary := miniredis.RunT(t)
	defer primary.Close()
	backup := miniredis.RunT(t)
	defer backup.Close()

	// Primary has production load
	for i := 0; i < 100; i++ {
		primary.Lpush("jobqueue:queue:critical", fmt.Sprintf("critical-job-%d", i))
	}
	for i := 1; i <= 8; i++ {
		primary.Set(fmt.Sprintf("worker:heartbeat:%d", i), "alive", time.Hour)
	}

	// Backup is on standby
	backup.Set("worker:heartbeat:backup-1", "alive", time.Hour)

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:        "primary",
				Label:       "Primary Production",
				Color:       "#ff0000",
				Environment: "production",
				Endpoint:    primary.Addr(),
				Enabled:     true,
				Tags:        []string{"production", "primary"},
			},
			{
				Name:        "backup",
				Label:       "Backup Cluster",
				Color:       "#ff9900",
				Environment: "production",
				Endpoint:    backup.Addr(),
				Enabled:     true,
				Tags:        []string{"production", "backup"},
			},
		},
		DefaultCluster: "primary",
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("PreFailureState", func(t *testing.T) {
		// Verify primary is healthy and handling load
		primaryStats, err := manager.GetStats(ctx, "primary")
		require.NoError(t, err)
		assert.Equal(t, 8, primaryStats.WorkerCount)

		primaryHealth, err := manager.GetHealth(ctx, "primary")
		require.NoError(t, err)
		assert.True(t, primaryHealth.Healthy)

		// Verify backup is ready
		backupHealth, err := manager.GetHealth(ctx, "backup")
		require.NoError(t, err)
		assert.True(t, backupHealth.Healthy)
	})

	t.Run("FailoverDetection", func(t *testing.T) {
		// Subscribe to events to monitor failover
		eventCh, err := manager.SubscribeEvents(ctx)
		require.NoError(t, err)

		// Simulate primary failure
		primary.Close()

		// Wait for health check to detect failure
		time.Sleep(200 * time.Millisecond)

		// Switch to backup cluster
		err = manager.SwitchCluster(ctx, "backup")
		require.NoError(t, err)

		// Verify backup is accessible
		backupStats, err := manager.GetStats(ctx, "backup")
		require.NoError(t, err)
		assert.Equal(t, "backup", backupStats.ClusterName)

		// Clean up event subscription
		manager.UnsubscribeEvents(ctx, eventCh)
	})

	t.Run("RecoveryValidation", func(t *testing.T) {
		// Verify we're operating on backup cluster
		tabConfig, err := manager.GetTabConfig(ctx)
		require.NoError(t, err)

		// Find active cluster name
		var activeCluster string
		for _, tab := range tabConfig.Tabs {
			if tab.Index-1 == tabConfig.ActiveTab {
				activeCluster = tab.ClusterName
				break
			}
		}
		assert.Equal(t, "backup", activeCluster)

		// Execute operations on backup to verify it's working
		action := &MultiAction{
			ID:      "recovery-test",
			Type:    ActionTypeBenchmark,
			Targets: []string{"backup"},
			Parameters: map[string]interface{}{
				"iterations": float64(5),
			},
			Status: ActionStatusConfirmed,
		}

		err = manager.ExecuteAction(ctx, action)
		require.NoError(t, err)
		assert.Equal(t, ActionStatusCompleted, action.Status)
	})
}

func TestE2E_HighVolumeOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup multiple clusters for load testing
	clusters := make([]*miniredis.Miniredis, 5)
	clusterConfigs := make([]ClusterConfig, 5)

	for i := 0; i < 5; i++ {
		clusters[i] = miniredis.RunT(t)
		defer clusters[i].Close()

		// Add varying amounts of test data
		for j := 0; j < (i+1)*20; j++ {
			clusters[i].Lpush("jobqueue:queue:high", fmt.Sprintf("job-%d-%d", i, j))
		}
		// Add workers
		for w := 1; w <= (i+1)*2; w++ {
			clusters[i].Set(fmt.Sprintf("worker:heartbeat:cluster%d-worker%d", i, w), "alive", time.Hour)
		}

		clusterConfigs[i] = ClusterConfig{
			Name:     fmt.Sprintf("cluster-%d", i),
			Label:    fmt.Sprintf("Cluster %d", i),
			Endpoint: clusters[i].Addr(),
			Enabled:  true,
		}
	}

	cfg := &Config{
		Clusters:    clusterConfigs,
		DefaultCluster: "cluster-0",
		Actions: ActionsConfig{
			RequireConfirmation: false,
			MaxConcurrent:       10,
			AllowedActions:      []ActionType{ActionTypeBenchmark},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("MassStatsCollection", func(t *testing.T) {
		// Collect stats from all clusters simultaneously
		start := time.Now()
		allStats, err := manager.GetAllStats(ctx)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, allStats, 5)
		assert.Less(t, duration, 5*time.Second, "Stats collection took too long: %v", duration)

		// Verify stats are correct for each cluster
		for i, stats := range allStats {
			clusterName := fmt.Sprintf("cluster-%d", i)
			if clusterStats, exists := allStats[clusterName]; exists {
				expectedWorkers := (i + 1) * 2
				assert.Equal(t, expectedWorkers, clusterStats.WorkerCount, "Wrong worker count for %s", clusterName)
			}
		}
	})

	t.Run("ConcurrentActions", func(t *testing.T) {
		// Execute actions on all clusters concurrently
		const numActions = 20
		var wg sync.WaitGroup
		results := make(chan error, numActions)

		start := time.Now()

		for i := 0; i < numActions; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				action := &MultiAction{
					ID:      fmt.Sprintf("concurrent-action-%d", i),
					Type:    ActionTypeBenchmark,
					Targets: []string{fmt.Sprintf("cluster-%d", i%5)}, // Distribute across clusters
					Parameters: map[string]interface{}{
						"iterations": float64(10),
					},
					Status: ActionStatusPending,
				}

				err := manager.ExecuteAction(ctx, action)
				results <- err
			}(i)
		}

		wg.Wait()
		close(results)
		duration := time.Since(start)

		// Check all actions succeeded
		for err := range results {
			assert.NoError(t, err)
		}

		assert.Less(t, duration, 30*time.Second, "Concurrent actions took too long: %v", duration)
	})

	t.Run("LargeScaleComparison", func(t *testing.T) {
		// Compare all clusters
		clusterNames := make([]string, 5)
		for i := 0; i < 5; i++ {
			clusterNames[i] = fmt.Sprintf("cluster-%d", i)
		}

		start := time.Now()
		result, err := manager.CompareClusters(ctx, clusterNames)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result.Clusters, 5)
		assert.Less(t, duration, 10*time.Second, "Large scale comparison took too long: %v", duration)

		// Should detect differences between clusters
		assert.NotEmpty(t, result.Metrics)
		workerMetric, exists := result.Metrics["worker_count"]
		assert.True(t, exists)
		assert.Greater(t, workerMetric.Delta, 0.0, "Should detect worker count differences")
	})
}

func TestE2E_RealWorldWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup development, staging, and production environments
	dev := miniredis.RunT(t)
	defer dev.Close()
	staging := miniredis.RunT(t)
	defer staging.Close()
	prod := miniredis.RunT(t)
	defer prod.Close()

	// Setup different workloads
	dev.Lpush("jobqueue:queue:normal", "dev-job-1", "dev-job-2")
	dev.Set("worker:heartbeat:dev-1", "alive", time.Hour)

	staging.Lpush("jobqueue:queue:normal", "staging-job-1", "staging-job-2", "staging-job-3")
	staging.Set("worker:heartbeat:staging-1", "alive", time.Hour)
	staging.Set("worker:heartbeat:staging-2", "alive", time.Hour)

	prod.Lpush("jobqueue:queue:critical", "prod-job-1", "prod-job-2", "prod-job-3", "prod-job-4")
	prod.Lpush("jobqueue:queue:normal", "prod-normal-1", "prod-normal-2")
	prod.Set("worker:heartbeat:prod-1", "alive", time.Hour)
	prod.Set("worker:heartbeat:prod-2", "alive", time.Hour)
	prod.Set("worker:heartbeat:prod-3", "alive", time.Hour)

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:        "development",
				Label:       "Development",
				Color:       "#00ff00",
				Environment: "development",
				Endpoint:    dev.Addr(),
				Enabled:     true,
				Tags:        []string{"dev"},
			},
			{
				Name:        "staging",
				Label:       "Staging",
				Color:       "#ffff00",
				Environment: "staging",
				Endpoint:    staging.Addr(),
				Enabled:     true,
				Tags:        []string{"staging"},
			},
			{
				Name:        "production",
				Label:       "Production",
				Color:       "#ff0000",
				Environment: "production",
				Endpoint:    prod.Addr(),
				Enabled:     true,
				Tags:        []string{"prod"},
			},
		},
		DefaultCluster: "development",
		Actions: ActionsConfig{
			RequireConfirmation: true,
			AllowedActions: []ActionType{
				ActionTypeBenchmark,
				ActionTypePurgeDLQ,
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Workflow 1: Development to Production Pipeline
	t.Run("DevToProdPipeline", func(t *testing.T) {
		pipeline := []string{"development", "staging", "production"}

		for _, env := range pipeline {
			// Switch to environment
			err := manager.SwitchCluster(ctx, env)
			require.NoError(t, err)

			// Check health
			health, err := manager.GetHealth(ctx, env)
			require.NoError(t, err)
			assert.True(t, health.Healthy, "Environment %s should be healthy", env)

			// Run validation benchmark
			action := &MultiAction{
				ID:      fmt.Sprintf("pipeline-validation-%s", env),
				Type:    ActionTypeBenchmark,
				Targets: []string{env},
				Parameters: map[string]interface{}{
					"iterations": float64(5),
				},
				Status: ActionStatusConfirmed,
			}

			err = manager.ExecuteAction(ctx, action)
			require.NoError(t, err)
			assert.Equal(t, ActionStatusCompleted, action.Status)

			result := action.Results[env]
			assert.True(t, result.Success, "Validation failed for %s: %s", env, result.Error)
		}
	})

	// Workflow 2: Environment Comparison and Drift Detection
	t.Run("EnvironmentDriftDetection", func(t *testing.T) {
		// Compare staging vs production
		result, err := manager.CompareClusters(ctx, []string{"staging", "production"})
		require.NoError(t, err)

		// Should detect worker count differences
		workerMetric, exists := result.Metrics["worker_count"]
		assert.True(t, exists)
		assert.Equal(t, 2.0, workerMetric.Values["staging"])
		assert.Equal(t, 3.0, workerMetric.Values["production"])

		// Log differences for operations team
		if len(result.Anomalies) > 0 {
			t.Logf("Detected %d anomalies between staging and production", len(result.Anomalies))
			for _, anomaly := range result.Anomalies {
				t.Logf("Anomaly: %s on %s - %s", anomaly.Type, anomaly.Cluster, anomaly.Description)
			}
		}
	})

	// Workflow 3: Emergency Response Simulation
	t.Run("EmergencyResponse", func(t *testing.T) {
		// Subscribe to events for monitoring
		eventCh, err := manager.SubscribeEvents(ctx)
		require.NoError(t, err)
		defer manager.UnsubscribeEvents(ctx, eventCh)

		// Simulate emergency: development environment is compromised
		// Switch all traffic handling to staging for testing
		err = manager.SwitchCluster(ctx, "staging")
		require.NoError(t, err)

		// Verify staging can handle the load
		stagingHealth, err := manager.GetHealth(ctx, "staging")
		require.NoError(t, err)
		assert.True(t, stagingHealth.Healthy)

		// Run emergency benchmark to test capacity
		emergencyAction := &MultiAction{
			ID:      "emergency-capacity-test",
			Type:    ActionTypeBenchmark,
			Targets: []string{"staging"},
			Parameters: map[string]interface{}{
				"iterations": float64(20), // Higher load test
			},
			Status: ActionStatusConfirmed,
		}

		err = manager.ExecuteAction(ctx, emergencyAction)
		require.NoError(t, err)
		assert.Equal(t, ActionStatusCompleted, emergencyAction.Status)
	})

	// Workflow 4: Maintenance Window Operations
	t.Run("MaintenanceWindow", func(t *testing.T) {
		// Simulate maintenance on development environment
		// Temporarily disable (simulate by removing worker)
		dev.Del("worker:heartbeat:dev-1")

		// Verify it's detected as unhealthy
		devHealth, err := manager.GetHealth(ctx, "development")
		require.NoError(t, err)
		assert.False(t, devHealth.Healthy)
		assert.Contains(t, devHealth.Issues, "No active workers")

		// During maintenance, ensure other environments are still working
		for _, env := range []string{"staging", "production"} {
			health, err := manager.GetHealth(ctx, env)
			require.NoError(t, err)
			assert.True(t, health.Healthy, "Environment %s should remain healthy during maintenance", env)
		}

		// Restore development environment
		dev.Set("worker:heartbeat:dev-1", "alive", time.Hour)

		// Verify recovery
		devHealthAfter, err := manager.GetHealth(ctx, "development")
		require.NoError(t, err)
		assert.True(t, devHealthAfter.Healthy)
	})
}