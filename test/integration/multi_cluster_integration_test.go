//go:build integration_tests
// +build integration_tests

// Copyright 2025 James Ross
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	multicluster "github.com/flyingrobots/go-redis-work-queue/internal/multi-cluster-control"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// TestMultiClusterIntegration tests multi-cluster functionality with real Redis containers
func TestMultiClusterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Redis containers
	redis1, endpoint1 := startRedisContainer(t, ctx)
	defer redis1.Terminate(ctx)

	redis2, endpoint2 := startRedisContainer(t, ctx)
	defer redis2.Terminate(ctx)

	// Configure multi-cluster manager
	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{
				Name:     "redis1",
				Label:    "Redis Cluster 1",
				Color:    "#ff0000",
				Endpoint: endpoint1,
				DB:       0,
				Enabled:  true,
			},
			{
				Name:     "redis2",
				Label:    "Redis Cluster 2",
				Color:    "#00ff00",
				Endpoint: endpoint2,
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "redis1",
		Polling: multicluster.PollingConfig{
			Enabled:  true,
			Interval: 1 * time.Second,
			Timeout:  5 * time.Second,
			Jitter:   200 * time.Millisecond,
		},
		Cache: multicluster.CacheConfig{
			Enabled:    true,
			TTL:        30 * time.Second,
			MaxEntries: 1000,
		},
		Actions: multicluster.ActionsConfig{
			RequireConfirmation: false,
			MaxConcurrent:       5,
			AllowedActions: []multicluster.ActionType{
				multicluster.ActionTypePurgeDLQ,
				multicluster.ActionTypeBenchmark,
				multicluster.ActionTypePauseQueue,
			},
		},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	t.Run("BasicConnectivity", func(t *testing.T) {
		testBasicConnectivity(t, manager, ctx)
	})

	t.Run("StatsCollection", func(t *testing.T) {
		testStatsCollection(t, manager, ctx, endpoint1, endpoint2)
	})

	t.Run("CompareMode", func(t *testing.T) {
		testCompareMode(t, manager, ctx, endpoint1, endpoint2)
	})

	t.Run("MultiClusterActions", func(t *testing.T) {
		testMultiClusterActions(t, manager, ctx)
	})

	t.Run("EventStreaming", func(t *testing.T) {
		testEventStreaming(t, manager, ctx)
	})

	t.Run("FailureRecovery", func(t *testing.T) {
		testFailureRecovery(t, manager, ctx, redis1, redis2)
	})
}

func testBasicConnectivity(t *testing.T, manager multicluster.Manager, ctx context.Context) {
	// Test listing clusters
	clusters, err := manager.ListClusters(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(clusters))

	// Test getting individual cluster connections
	conn1, err := manager.GetCluster(ctx, "redis1")
	require.NoError(t, err)
	assert.NotNil(t, conn1)
	assert.True(t, conn1.Status.Connected)

	conn2, err := manager.GetCluster(ctx, "redis2")
	require.NoError(t, err)
	assert.NotNil(t, conn2)
	assert.True(t, conn2.Status.Connected)

	// Test switching between clusters
	err = manager.SwitchCluster(ctx, "redis1")
	require.NoError(t, err)

	err = manager.SwitchCluster(ctx, "redis2")
	require.NoError(t, err)

	// Test health checks
	health1, err := manager.GetHealth(ctx, "redis1")
	require.NoError(t, err)
	assert.True(t, health1.Healthy)

	health2, err := manager.GetHealth(ctx, "redis2")
	require.NoError(t, err)
	assert.True(t, health2.Healthy)
}

func testStatsCollection(t *testing.T, manager multicluster.Manager, ctx context.Context, endpoint1, endpoint2 string) {
	// Setup test data in both clusters
	setupTestData(t, endpoint1, map[string][]string{
		"jobqueue:queue:high":   {"job1", "job2", "job3"},
		"jobqueue:queue:normal": {"job4", "job5"},
		"jobqueue:dead_letter":  {"dead1"},
	})

	setupTestData(t, endpoint2, map[string][]string{
		"jobqueue:queue:high":   {"job6", "job7"},
		"jobqueue:queue:normal": {"job8", "job9", "job10", "job11"},
		"jobqueue:dead_letter":  {"dead2", "dead3"},
	})

	// Give some time for data to settle
	time.Sleep(100 * time.Millisecond)

	// Test getting stats from individual clusters
	stats1, err := manager.GetStats(ctx, "redis1")
	require.NoError(t, err)
	assert.NotNil(t, stats1)
	assert.Equal(t, "redis1", stats1.ClusterName)
	assert.Equal(t, int64(3), stats1.QueueSizes["high"])
	assert.Equal(t, int64(2), stats1.QueueSizes["normal"])
	assert.Equal(t, int64(1), stats1.DeadLetterCount)

	stats2, err := manager.GetStats(ctx, "redis2")
	require.NoError(t, err)
	assert.NotNil(t, stats2)
	assert.Equal(t, "redis2", stats2.ClusterName)
	assert.Equal(t, int64(2), stats2.QueueSizes["high"])
	assert.Equal(t, int64(4), stats2.QueueSizes["normal"])
	assert.Equal(t, int64(2), stats2.DeadLetterCount)

	// Test getting all stats at once
	allStats, err := manager.GetAllStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(allStats))
	assert.Contains(t, allStats, "redis1")
	assert.Contains(t, allStats, "redis2")
}

func testCompareMode(t *testing.T, manager multicluster.Manager, ctx context.Context, endpoint1, endpoint2 string) {
	// Enable compare mode
	err := manager.SetCompareMode(ctx, true, []string{"redis1", "redis2"})
	require.NoError(t, err)

	// Get tab config to verify compare mode is enabled
	tabConfig, err := manager.GetTabConfig(ctx)
	require.NoError(t, err)
	assert.True(t, tabConfig.CompareMode)
	assert.Equal(t, []string{"redis1", "redis2"}, tabConfig.CompareWith)

	// Test cluster comparison
	compareResult, err := manager.CompareClusters(ctx, []string{"redis1", "redis2"})
	require.NoError(t, err)
	assert.NotNil(t, compareResult)
	assert.Equal(t, []string{"redis1", "redis2"}, compareResult.Clusters)
	assert.NotEmpty(t, compareResult.Metrics)

	// Should have detected differences in queue sizes
	found := false
	for _, metric := range compareResult.Metrics {
		if len(metric.Values) == 2 {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have comparison metrics for both clusters")

	// Disable compare mode
	err = manager.SetCompareMode(ctx, false, nil)
	require.NoError(t, err)

	// Verify compare mode is disabled
	tabConfig, err = manager.GetTabConfig(ctx)
	require.NoError(t, err)
	assert.False(t, tabConfig.CompareMode)
}

func testMultiClusterActions(t *testing.T, manager multicluster.Manager, ctx context.Context) {
	// Test benchmark action across both clusters
	benchmarkAction := &multicluster.MultiAction{
		ID:      "integration-benchmark-001",
		Type:    multicluster.ActionTypeBenchmark,
		Targets: []string{"redis1", "redis2"},
		Parameters: map[string]interface{}{
			"iterations":   float64(10),
			"payload_size": float64(50),
			"queue_name":   "benchmark-queue",
		},
		Status:    multicluster.ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute benchmark
	err := manager.ExecuteAction(ctx, benchmarkAction)
	require.NoError(t, err)

	// Verify action completed successfully
	assert.Equal(t, multicluster.ActionStatusCompleted, benchmarkAction.Status)
	assert.Equal(t, 2, len(benchmarkAction.Results))

	// Check results for both clusters
	for _, target := range []string{"redis1", "redis2"} {
		result, exists := benchmarkAction.Results[target]
		assert.True(t, exists, "Should have result for cluster %s", target)
		assert.True(t, result.Success, "Benchmark should succeed for cluster %s", target)
		assert.Greater(t, result.Duration, float64(0), "Should have measurable duration for cluster %s", target)
	}

	// Test DLQ purge action
	purgeAction := &multicluster.MultiAction{
		ID:      "integration-purge-001",
		Type:    multicluster.ActionTypePurgeDLQ,
		Targets: []string{"redis1", "redis2"},
		Parameters: map[string]interface{}{
			"confirm": true,
		},
		Status:    multicluster.ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Execute purge
	err = manager.ExecuteAction(ctx, purgeAction)
	require.NoError(t, err)

	// Verify purge completed
	assert.Equal(t, multicluster.ActionStatusCompleted, purgeAction.Status)
	assert.Equal(t, 2, len(purgeAction.Results))

	for _, target := range []string{"redis1", "redis2"} {
		result, exists := purgeAction.Results[target]
		assert.True(t, exists)
		assert.True(t, result.Success)
	}
}

func testEventStreaming(t *testing.T, manager multicluster.Manager, ctx context.Context) {
	// Subscribe to events
	eventCh, err := manager.SubscribeEvents(ctx)
	require.NoError(t, err)

	// Trigger an event by adding a new cluster
	newCluster := multicluster.ClusterConfig{
		Name:     "dynamic-cluster",
		Label:    "Dynamic Cluster",
		Color:    "#0000ff",
		Endpoint: "localhost:9999", // Will fail to connect
		DB:       0,
		Enabled:  true,
	}

	// Add cluster (will generate connection events)
	err = manager.AddCluster(ctx, newCluster)
	// Error expected since Redis isn't running on port 9999

	// Should receive an event within reasonable time
	select {
	case event := <-eventCh:
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, "dynamic-cluster", event.Cluster)
		assert.NotEmpty(t, event.Message)
		assert.NotZero(t, event.Timestamp)
	case <-time.After(2 * time.Second):
		t.Fatal("Expected to receive an event within 2 seconds")
	}

	// Unsubscribe
	err = manager.UnsubscribeEvents(ctx, eventCh)
	assert.NoError(t, err)
}

func testFailureRecovery(t *testing.T, manager multicluster.Manager, ctx context.Context, redis1, redis2 testcontainers.Container) {
	// Verify both clusters are initially healthy
	health1, err := manager.GetHealth(ctx, "redis1")
	require.NoError(t, err)
	assert.True(t, health1.Healthy)

	health2, err := manager.GetHealth(ctx, "redis2")
	require.NoError(t, err)
	assert.True(t, health2.Healthy)

	// Stop one Redis container to simulate failure
	err = redis1.Stop(ctx, nil)
	require.NoError(t, err)

	// Wait for health check to detect the failure
	time.Sleep(3 * time.Second)

	// Verify that the failed cluster is marked as unhealthy
	health1, err = manager.GetHealth(ctx, "redis1")
	if err == nil {
		assert.False(t, health1.Healthy)
		assert.NotEmpty(t, health1.Issues)
	}

	// Verify that the healthy cluster is still healthy
	health2, err = manager.GetHealth(ctx, "redis2")
	require.NoError(t, err)
	assert.True(t, health2.Healthy)

	// Actions targeting the failed cluster should fail gracefully
	failAction := &multicluster.MultiAction{
		ID:      "fail-test-001",
		Type:    multicluster.ActionTypeBenchmark,
		Targets: []string{"redis1"},
		Parameters: map[string]interface{}{
			"iterations": float64(5),
		},
		Status:    multicluster.ActionStatusPending,
		CreatedAt: time.Now(),
	}

	err = manager.ExecuteAction(ctx, failAction)
	// Should handle the failure gracefully
	assert.Equal(t, multicluster.ActionStatusFailed, failAction.Status)

	result, exists := failAction.Results["redis1"]
	assert.True(t, exists)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)

	// Restart the container
	err = redis1.Start(ctx)
	require.NoError(t, err)

	// Wait for recovery
	time.Sleep(3 * time.Second)

	// Verify recovery
	health1, err = manager.GetHealth(ctx, "redis1")
	require.NoError(t, err)
	assert.True(t, health1.Healthy)
}

// Helper functions

func startRedisContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
		Env: map[string]string{
			"REDIS_DISABLE_COMMANDS": "FLUSHDB,FLUSHALL,DEBUG",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	return container, endpoint
}

func setupTestData(t *testing.T, endpoint string, data map[string][]string) {
	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	defer client.Close()

	ctx := context.Background()

	for key, values := range data {
		// Clear the key first
		client.Del(ctx, key)

		// Add values
		if len(values) > 0 {
			args := make([]interface{}, len(values))
			for i, v := range values {
				args[i] = v
			}
			err := client.LPush(ctx, key, args...).Err()
			require.NoError(t, err)
		}
	}
}

// Test environment setup
func TestMain(m *testing.M) {
	// Skip integration tests if not explicitly enabled
	if os.Getenv("INTEGRATION_TESTS") == "" && !testing.Short() {
		os.Setenv("INTEGRATION_TESTS", "1")
	}

	code := m.Run()
	os.Exit(code)
}
