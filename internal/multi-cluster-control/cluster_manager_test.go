// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClusterManager_SwitchCluster(t *testing.T) {
	// Setup multiple Redis instances
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "cluster1",
				Label:    "Cluster 1",
				Color:    "#ff0000",
				Endpoint: mr1.Addr(),
				DB:       0,
				Enabled:  true,
			},
			{
				Name:     "cluster2",
				Label:    "Cluster 2",
				Color:    "#00ff00",
				Endpoint: mr2.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "cluster1",
		Polling: PollingConfig{
			Enabled: false,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test initial active cluster
	assert.Equal(t, "cluster1", manager.activeTab)

	// Test switching to cluster2
	err = manager.SwitchCluster(ctx, "cluster2")
	require.NoError(t, err)
	assert.Equal(t, "cluster2", manager.activeTab)

	// Test switching to non-existent cluster
	err = manager.SwitchCluster(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrClusterNotFound, err)

	// Verify active cluster didn't change after failed switch
	assert.Equal(t, "cluster2", manager.activeTab)
}

func TestClusterManager_CompareMode(t *testing.T) {
	cfg := DefaultConfig()
	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test enabling compare mode with no clusters
	err = manager.SetCompareMode(ctx, true, []string{})
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientClusters, err)

	// Test enabling compare mode with single cluster
	err = manager.SetCompareMode(ctx, true, []string{"cluster1"})
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientClusters, err)

	// Test enabling compare mode with multiple clusters
	err = manager.SetCompareMode(ctx, true, []string{"cluster1", "cluster2"})
	require.NoError(t, err)
	assert.True(t, manager.compareMode)
	assert.Equal(t, []string{"cluster1", "cluster2"}, manager.compareClusters)

	// Test disabling compare mode
	err = manager.SetCompareMode(ctx, false, nil)
	require.NoError(t, err)
	assert.False(t, manager.compareMode)
	assert.Empty(t, manager.compareClusters)

	// Test enabling compare mode with too many clusters (>5)
	manyClusters := []string{"c1", "c2", "c3", "c4", "c5", "c6"}
	err = manager.SetCompareMode(ctx, true, manyClusters)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum")
}

func TestClusterManager_TabConfig(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "prod-east", Label: "Prod East", Color: "#ff0000", Enabled: true},
			{Name: "prod-west", Label: "Prod West", Color: "#00ff00", Enabled: true},
			{Name: "staging", Label: "Staging", Color: "#0000ff", Enabled: true},
		},
		DefaultCluster: "prod-east",
		Polling:        PollingConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Get initial tab config
	tabConfig, err := manager.GetTabConfig(ctx)
	require.NoError(t, err)

	assert.Equal(t, 3, len(tabConfig.Tabs))
	assert.Equal(t, 0, tabConfig.ActiveTab) // Should be index of "prod-east"
	assert.False(t, tabConfig.CompareMode)

	// Check first tab
	tab1 := tabConfig.Tabs[0]
	assert.Equal(t, 0, tab1.Index)
	assert.Equal(t, "prod-east", tab1.ClusterName)
	assert.Equal(t, "Prod East", tab1.Label)
	assert.Equal(t, "#ff0000", tab1.Color)
	assert.Equal(t, "1", tab1.Shortcut)

	// Enable compare mode
	err = manager.SetCompareMode(ctx, true, []string{"prod-east", "staging"})
	require.NoError(t, err)

	// Get updated tab config
	tabConfig, err = manager.GetTabConfig(ctx)
	require.NoError(t, err)
	assert.True(t, tabConfig.CompareMode)
	assert.Equal(t, []string{"prod-east", "staging"}, tabConfig.CompareWith)
}

func TestClusterManager_StatsCollection(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	// Setup test data in Redis
	mr.Lpush("jobqueue:queue:high", "job1")
	mr.Lpush("jobqueue:queue:high", "job2")
	mr.Lpush("jobqueue:queue:normal", "job3")
	mr.Lpush("jobqueue:queue:normal", "job4")
	mr.Lpush("jobqueue:queue:normal", "job5")
	mr.Lpush("jobqueue:processing", "processing1")
	mr.Lpush("jobqueue:dead_letter", "dead1")
	mr.Lpush("jobqueue:dead_letter", "dead2")
	mr.Set("jobqueue:workers:count", "5")

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "test",
				Label:    "Test Cluster",
				Color:    "#blue",
				Endpoint: mr.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "test",
		Polling: PollingConfig{
			Enabled: false,
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     30 * time.Second,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test getting stats
	stats, err := manager.GetStats(ctx, "test")
	require.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, "test", stats.ClusterName)
	assert.Equal(t, int64(2), stats.QueueSizes["high"])
	assert.Equal(t, int64(3), stats.QueueSizes["normal"])
	assert.Equal(t, int64(1), stats.ProcessingCount)
	assert.Equal(t, int64(2), stats.DeadLetterCount)

	// Test getting all stats
	allStats, err := manager.GetAllStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(allStats))
	assert.Contains(t, allStats, "test")
}

func TestClusterManager_CompareResults(t *testing.T) {
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	// Setup different data in each cluster
	mr1.Lpush("jobqueue:queue:high", "job1")
	mr1.Lpush("jobqueue:queue:high", "job2")
	mr1.Lpush("jobqueue:queue:high", "job3")
	mr1.Lpush("jobqueue:dead_letter", "dead1")
	mr2.Lpush("jobqueue:queue:high", "job1")
	mr2.Lpush("jobqueue:queue:high", "job2")
	mr2.Lpush("jobqueue:dead_letter", "dead1")
	mr2.Lpush("jobqueue:dead_letter", "dead2")
	mr2.Lpush("jobqueue:dead_letter", "dead3")

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Label: "Cluster 1", Color: "#red", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "cluster2", Label: "Cluster 2", Color: "#blue", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "cluster1",
		Polling: PollingConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test comparing clusters
	result, err := manager.CompareClusters(ctx, []string{"cluster1", "cluster2"})
	require.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, []string{"cluster1", "cluster2"}, result.Clusters)
	assert.NotEmpty(t, result.Metrics)

	// Should have metrics for dead letter count showing difference
	if dlqMetric, exists := result.Metrics["dead_letter_count"]; exists {
		assert.Contains(t, dlqMetric.Values, "cluster1")
		assert.Contains(t, dlqMetric.Values, "cluster2")
		assert.Equal(t, float64(1), dlqMetric.Values["cluster1"])
		assert.Equal(t, float64(3), dlqMetric.Values["cluster2"])
	}

	// Should detect anomalies for significant differences
	assert.NotEmpty(t, result.Anomalies)
}

func TestClusterManager_HealthChecks(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "healthy",
				Label:    "Healthy Cluster",
				Color:    "#green",
				Endpoint: mr.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "healthy",
		Polling: PollingConfig{
			Enabled:  true,
			Interval: 100 * time.Millisecond, // Fast polling for test
			Timeout:  50 * time.Millisecond,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test getting health status
	health, err := manager.GetHealth(ctx, "healthy")
	require.NoError(t, err)
	assert.NotNil(t, health)

	assert.True(t, health.Healthy)
	assert.Empty(t, health.Issues)
	assert.NotEmpty(t, health.Metrics)
	assert.Contains(t, health.Metrics, "latency_ms")

	// Test health for non-existent cluster
	_, err = manager.GetHealth(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrClusterNotFound, err)
}

func TestClusterManager_EventSubscription(t *testing.T) {
	cfg := DefaultConfig()
	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Subscribe to events
	eventCh, err := manager.SubscribeEvents(ctx)
	require.NoError(t, err)

	// Add a cluster to generate an event
	clusterCfg := ClusterConfig{
		Name:     "new-cluster",
		Label:    "New Cluster",
		Color:    "#purple",
		Endpoint: "localhost:6379",
		DB:       0,
		Enabled:  true,
	}

	// Should fail to connect (no Redis running), but should still generate an event
	err = manager.AddCluster(ctx, clusterCfg)
	// Error expected since Redis isn't running

	// Should still receive connection attempt event
	select {
	case event := <-eventCh:
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, "new-cluster", event.Cluster)
		assert.NotEmpty(t, event.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive an event within 1 second")
	}

	// Unsubscribe from events
	err = manager.UnsubscribeEvents(ctx, eventCh)
	assert.NoError(t, err)
}

func TestClusterManager_ConcurrentAccess(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "concurrent",
				Label:    "Concurrent Test",
				Color:    "#orange",
				Endpoint: mr.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "concurrent",
		Polling: PollingConfig{
			Enabled: false,
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     1 * time.Second,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test concurrent stats requests
	const numGoroutines = 10
	results := make(chan *ClusterStats, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			stats, err := manager.GetStats(ctx, "concurrent")
			if err != nil {
				errors <- err
			} else {
				results <- stats
			}
		}()
	}

	// Collect results
	var statsCount int
	var errorCount int

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-results:
			statsCount++
		case <-errors:
			errorCount++
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine results")
		}
	}

	// All requests should succeed
	assert.Equal(t, numGoroutines, statsCount)
	assert.Equal(t, 0, errorCount)
}

func TestClusterManager_ErrorHandling(t *testing.T) {
	// Test with invalid cluster configuration
	cfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "invalid",
				Label:    "Invalid Cluster",
				Color:    "#red",
				Endpoint: "invalid-host:9999",
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "invalid",
		Polling: PollingConfig{
			Enabled: false,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Test getting stats from disconnected cluster
	_, err = manager.GetStats(ctx, "invalid")
	assert.Error(t, err)

	// Test getting health from disconnected cluster
	health, err := manager.GetHealth(ctx, "invalid")
	if err == nil {
		// If no error, health should indicate unhealthy state
		assert.False(t, health.Healthy)
		assert.NotEmpty(t, health.Issues)
	}

	// Test operations on non-existent cluster
	_, err = manager.GetStats(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrClusterNotFound, err)
}

func TestClusterManager_ConfigurationReload(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	initialCfg := &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "initial",
				Label:    "Initial Cluster",
				Color:    "#blue",
				Endpoint: mr.Addr(),
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "initial",
		Polling: PollingConfig{Enabled: false},
	}

	manager, err := NewManager(initialCfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// Verify initial cluster exists
	clusters, err := manager.ListClusters(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, "initial", clusters[0].Name)

	// Add another cluster dynamically
	newCluster := ClusterConfig{
		Name:     "dynamic",
		Label:    "Dynamic Cluster",
		Color:    "#green",
		Endpoint: mr.Addr(),
		DB:       1, // Different DB
		Enabled:  true,
	}

	err = manager.AddCluster(ctx, newCluster)
	require.NoError(t, err)

	// Verify both clusters exist
	clusters, err = manager.ListClusters(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(clusters))

	clusterNames := make([]string, len(clusters))
	for i, c := range clusters {
		clusterNames[i] = c.Name
	}
	assert.Contains(t, clusterNames, "initial")
	assert.Contains(t, clusterNames, "dynamic")
}