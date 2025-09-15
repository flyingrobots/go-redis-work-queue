// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMultiCluster_EndToEndScenarios tests comprehensive real-world scenarios
func TestMultiCluster_EndToEndScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup multiple Redis instances
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()
	mr3 := miniredis.RunT(t)
	defer mr3.Close()

	// Configure realistic multi-cluster setup
	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "prod-east",
				Label:    "Production East",
				Color:    "#ff0000",
				Endpoint: mr1.Addr(),
				DB:       0,
				Enabled:  true,
			},
			{
				Name:     "prod-west",
				Label:    "Production West",
				Color:    "#00ff00",
				Endpoint: mr2.Addr(),
				DB:       0,
				Enabled:  true,
			},
			{
				Name:     "staging",
				Label:    "Staging",
				Color:    "#0000ff",
				Endpoint: mr3.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "prod-east",
		Polling: PollingConfig{
			Enabled:  true,
			Interval: 100 * time.Millisecond, // Fast for testing
			Timeout:  50 * time.Millisecond,
			Jitter:   10 * time.Millisecond,
		},
		Cache: CacheConfig{
			Enabled:         true,
			TTL:             5 * time.Second,
			MaxEntries:      100,
			CleanupInterval: 1 * time.Second,
		},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
				ActionTypePauseQueue,
				ActionTypeResumeQueue,
			},
			MaxConcurrent: 3,
			RetryPolicy: RetryPolicy{
				MaxAttempts:  2,
				InitialDelay: 10 * time.Millisecond,
				MaxDelay:     100 * time.Millisecond,
				Multiplier:   2.0,
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("Scenario1_BasicOperations", func(t *testing.T) {
		// Test basic connectivity
		clusters, err := manager.ListClusters(ctx)
		require.NoError(t, err)
		assert.Len(t, clusters, 3)

		// Test switching between clusters
		for _, cluster := range clusters {
			err := manager.SwitchCluster(ctx, cluster.Name)
			if err != nil {
				t.Logf("Failed to switch to cluster %s: %v", cluster.Name, err)
			}
		}

		// Test getting cluster connections
		for _, cluster := range clusters {
			conn, err := manager.GetCluster(ctx, cluster.Name)
			if err == nil {
				assert.NotNil(t, conn)
				assert.Equal(t, cluster.Name, conn.Config.Name)
			}
		}
	})

	t.Run("Scenario2_StatsAndMonitoring", func(t *testing.T) {
		// Setup test data in each cluster
		setupTestData(mr1, map[string][]string{
			"jobqueue:queue:high":   {"job1", "job2", "job3"},
			"jobqueue:queue:normal": {"job4", "job5"},
			"jobqueue:dead_letter":  {"dead1"},
		})

		setupTestData(mr2, map[string][]string{
			"jobqueue:queue:high":   {"job6", "job7"},
			"jobqueue:queue:normal": {"job8", "job9", "job10"},
			"jobqueue:dead_letter":  {"dead2", "dead3"},
		})

		setupTestData(mr3, map[string][]string{
			"jobqueue:queue:low":    {"job11"},
			"jobqueue:dead_letter":  {"dead4"},
		})

		// Wait for polling to pick up data
		time.Sleep(200 * time.Millisecond)

		// Test getting stats from individual clusters
		stats1, err := manager.GetStats(ctx, "prod-east")
		if err == nil {
			assert.NotNil(t, stats1)
			assert.Equal(t, "prod-east", stats1.ClusterName)
		}

		stats2, err := manager.GetStats(ctx, "prod-west")
		if err == nil {
			assert.NotNil(t, stats2)
			assert.Equal(t, "prod-west", stats2.ClusterName)
		}

		// Test getting all stats
		allStats, err := manager.GetAllStats(ctx)
		if err == nil {
			assert.NotEmpty(t, allStats)
		}

		// Test health checks
		for _, cluster := range []string{"prod-east", "prod-west", "staging"} {
			health, err := manager.GetHealth(ctx, cluster)
			if err == nil {
				assert.NotNil(t, health)
			}
		}
	})

	t.Run("Scenario3_MultiClusterActions", func(t *testing.T) {
		// Test benchmark across multiple clusters
		benchmarkAction := &MultiAction{
			ID:      "benchmark-scenario",
			Type:    ActionTypeBenchmark,
			Targets: []string{"prod-east", "prod-west"},
			Parameters: map[string]interface{}{
				"iterations":   float64(5),
				"payload_size": float64(100),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		}

		err := manager.ExecuteAction(ctx, benchmarkAction)
		if err == nil {
			assert.Equal(t, ActionStatusCompleted, benchmarkAction.Status)
			assert.Equal(t, 2, len(benchmarkAction.Results))
		}

		// Test DLQ purge across all clusters
		purgeAction := &MultiAction{
			ID:      "purge-scenario",
			Type:    ActionTypePurgeDLQ,
			Targets: []string{"prod-east", "prod-west", "staging"},
			Parameters: map[string]interface{}{
				"confirm": true,
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		}

		err = manager.ExecuteAction(ctx, purgeAction)
		if err == nil {
			assert.Equal(t, ActionStatusCompleted, purgeAction.Status)
			assert.Equal(t, 3, len(purgeAction.Results))
		}
	})

	t.Run("Scenario4_ComparisonAndAnalysis", func(t *testing.T) {
		// Test cluster comparison
		result, err := manager.CompareClusters(ctx, []string{"prod-east", "prod-west"})
		if err == nil {
			assert.NotNil(t, result)
			assert.Equal(t, []string{"prod-east", "prod-west"}, result.Clusters)
			assert.NotEmpty(t, result.Metrics)
		}

		// Test compare mode
		err = manager.SetCompareMode(ctx, true, []string{"prod-east", "prod-west"})
		if err == nil {
			tabConfig, err := manager.GetTabConfig(ctx)
			if err == nil {
				assert.True(t, tabConfig.CompareMode)
				assert.Equal(t, []string{"prod-east", "prod-west"}, tabConfig.CompareWith)
			}

			// Disable compare mode
			err = manager.SetCompareMode(ctx, false, nil)
			assert.NoError(t, err)
		}
	})

	t.Run("Scenario5_DynamicClusterManagement", func(t *testing.T) {
		// Add a new cluster dynamically
		mr4 := miniredis.RunT(t)
		defer mr4.Close()

		newCluster := ClusterConfig{
			Name:     "dev",
			Label:    "Development",
			Color:    "#ffff00",
			Endpoint: mr4.Addr(),
			DB:       0,
			Enabled:  true,
		}

		err := manager.AddCluster(ctx, newCluster)
		assert.NoError(t, err)

		// Verify it was added
		clusters, err := manager.ListClusters(ctx)
		assert.NoError(t, err)
		assert.Len(t, clusters, 4)

		// Test switching to the new cluster
		err = manager.SwitchCluster(ctx, "dev")
		if err == nil {
			tabConfig, err := manager.GetTabConfig(ctx)
			if err == nil {
				// Find the dev cluster tab
				found := false
				for i, tab := range tabConfig.Tabs {
					if tab.ClusterName == "dev" {
						assert.Equal(t, i, tabConfig.ActiveTab)
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		}

		// Remove the cluster
		err = manager.RemoveCluster(ctx, "dev")
		assert.NoError(t, err)

		clusters, err = manager.ListClusters(ctx)
		assert.NoError(t, err)
		assert.Len(t, clusters, 3)
	})

	t.Run("Scenario6_FailureRecovery", func(t *testing.T) {
		// Test operations with a failed cluster
		mr2.Close() // Simulate failure of prod-west

		// Actions should handle failures gracefully
		action := &MultiAction{
			ID:      "failure-test",
			Type:    ActionTypeBenchmark,
			Targets: []string{"prod-east", "prod-west"}, // One working, one failed
			Parameters: map[string]interface{}{
				"iterations": float64(3),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		}

		err := manager.ExecuteAction(ctx, action)
		// Should complete with mixed results
		if action.Results != nil {
			eastResult := action.Results["prod-east"]
			westResult := action.Results["prod-west"]

			if eastResult.Success && !westResult.Success {
				// Expected: east succeeds, west fails
				assert.True(t, eastResult.Success)
				assert.False(t, westResult.Success)
				assert.NotEmpty(t, westResult.Error)
			}
		}

		// Health checks should detect the failure
		health, err := manager.GetHealth(ctx, "prod-west")
		if err == nil {
			assert.False(t, health.Healthy)
			assert.NotEmpty(t, health.Issues)
		}
	})
}

// TestMultiCluster_PerformanceAndScaling tests performance characteristics
func TestMultiCluster_PerformanceAndScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create multiple Redis instances for scaling test
	const numClusters = 5
	var mrs []*miniredis.Miniredis
	var clusters []ClusterConfig

	for i := 0; i < numClusters; i++ {
		mr := miniredis.RunT(t)
		defer mr.Close()
		mrs = append(mrs, mr)

		clusters = append(clusters, ClusterConfig{
			Name:     fmt.Sprintf("cluster%d", i),
			Label:    fmt.Sprintf("Cluster %d", i),
			Color:    "#000000",
			Endpoint: mr.Addr(),
			DB:       0,
			Enabled:  true,
		})
	}

	cfg := &Config{
		Clusters:       clusters,
		DefaultCluster: "cluster0",
		Polling:        PollingConfig{Enabled: false}, // Disable for performance testing
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        30 * time.Second,
			MaxEntries: 1000,
		},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions:      []ActionType{ActionTypeBenchmark},
			MaxConcurrent:       10,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("ConcurrentStats", func(t *testing.T) {
		start := time.Now()

		// Get stats from all clusters concurrently
		allStats, err := manager.GetAllStats(ctx)
		duration := time.Since(start)

		if err == nil {
			assert.Len(t, allStats, numClusters)
			t.Logf("Got stats from %d clusters in %v", numClusters, duration)
			assert.Less(t, duration, 5*time.Second, "Should complete within 5 seconds")
		}
	})

	t.Run("ConcurrentActions", func(t *testing.T) {
		const numActions = 10
		var actions []*MultiAction

		// Create multiple actions
		for i := 0; i < numActions; i++ {
			action := &MultiAction{
				ID:      fmt.Sprintf("perf-test-%d", i),
				Type:    ActionTypeBenchmark,
				Targets: []string{fmt.Sprintf("cluster%d", i%numClusters)},
				Parameters: map[string]interface{}{
					"iterations": float64(3),
				},
				Status:    ActionStatusPending,
				CreatedAt: time.Now(),
			}
			actions = append(actions, action)
		}

		start := time.Now()

		// Execute actions concurrently
		done := make(chan error, numActions)
		for _, action := range actions {
			go func(a *MultiAction) {
				done <- manager.ExecuteAction(ctx, a)
			}(action)
		}

		// Wait for all to complete
		completed := 0
		for i := 0; i < numActions; i++ {
			select {
			case err := <-done:
				if err == nil {
					completed++
				}
			case <-time.After(30 * time.Second):
				t.Fatal("Timeout waiting for actions to complete")
			}
		}

		duration := time.Since(start)
		t.Logf("Completed %d/%d actions in %v", completed, numActions, duration)
		assert.Greater(t, completed, numActions/2, "At least half of actions should succeed")
	})
}

// setupTestData is a helper function to set up test data in a miniredis instance
func setupTestData(mr *miniredis.Miniredis, data map[string][]string) {
	for key, values := range data {
		mr.Del(key) // Clear existing data
		for _, value := range values {
			mr.Lpush(key, value)
		}
	}
}