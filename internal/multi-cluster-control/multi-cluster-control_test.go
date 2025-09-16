// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	cfg := DefaultConfig()
	logger := zap.NewNop()

	manager, err := NewManager(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if len(manager.connections) == 0 {
		t.Skip("No connections established (expected for default config)")
	}
}

func TestAddRemoveCluster(t *testing.T) {
	// Start a mini Redis server for testing
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters:       []ClusterConfig{},
		DefaultCluster: "",
		Polling: PollingConfig{
			Interval: 5 * time.Second,
			Jitter:   1 * time.Second,
			Timeout:  3 * time.Second,
			Enabled:  false, // Disable for tests
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Test adding a cluster
	clusterCfg := ClusterConfig{
		Name:     "test-cluster",
		Label:    "Test Cluster",
		Color:    "blue",
		Endpoint: mr.Addr(),
		DB:       0,
		Enabled:  true,
	}

	err = manager.AddCluster(ctx, clusterCfg)
	if err != nil {
		t.Fatalf("Failed to add cluster: %v", err)
	}

	// Verify cluster was added
	clusters, err := manager.ListClusters(ctx)
	if err != nil {
		t.Fatalf("Failed to list clusters: %v", err)
	}

	if len(clusters) != 1 {
		t.Fatalf("Expected 1 cluster, got %d", len(clusters))
	}

	if clusters[0].Name != "test-cluster" {
		t.Fatalf("Expected cluster name 'test-cluster', got '%s'", clusters[0].Name)
	}

	// Test getting the cluster
	conn, err := manager.GetCluster(ctx, "test-cluster")
	if err != nil {
		t.Fatalf("Failed to get cluster: %v", err)
	}

	if conn == nil {
		t.Fatal("Connection should not be nil")
	}

	// Test removing the cluster
	err = manager.RemoveCluster(ctx, "test-cluster")
	if err != nil {
		t.Fatalf("Failed to remove cluster: %v", err)
	}

	// Verify cluster was removed
	clusters, err = manager.ListClusters(ctx)
	if err != nil {
		t.Fatalf("Failed to list clusters: %v", err)
	}

	if len(clusters) != 0 {
		t.Fatalf("Expected 0 clusters, got %d", len(clusters))
	}
}


func TestCompareMode(t *testing.T) {
	cfg := DefaultConfig()
	manager, err := NewManager(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Test enabling compare mode with insufficient clusters
	err = manager.SetCompareMode(ctx, true, []string{"cluster1"})
	if err != ErrInsufficientClusters {
		t.Errorf("Expected ErrInsufficientClusters, got %v", err)
	}

	// Test enabling compare mode with sufficient clusters
	err = manager.SetCompareMode(ctx, true, []string{"cluster1", "cluster2"})
	if err != nil {
		t.Errorf("Failed to set compare mode: %v", err)
	}

	// Get tab config to verify
	tabConfig, err := manager.GetTabConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get tab config: %v", err)
	}

	if !tabConfig.CompareMode {
		t.Error("Compare mode should be enabled")
	}

	if len(tabConfig.CompareWith) != 2 {
		t.Errorf("Expected 2 compare clusters, got %d", len(tabConfig.CompareWith))
	}

	// Test disabling compare mode
	err = manager.SetCompareMode(ctx, false, nil)
	if err != nil {
		t.Errorf("Failed to disable compare mode: %v", err)
	}
}



func TestCacheOperations(t *testing.T) {
	cache := NewClusterCache()

	// Test Set and Get
	cache.Set("test-key", "test-value", 1*time.Second)

	value, ok := cache.Get("test-key")
	if !ok {
		t.Error("Failed to get cached value")
	}

	if value != "test-value" {
		t.Errorf("Expected 'test-value', got %v", value)
	}

	// Test expiration
	time.Sleep(2 * time.Second)

	_, ok = cache.Get("test-key")
	if ok {
		t.Error("Cached value should have expired")
	}

	// Test Delete
	cache.Set("delete-key", "delete-value", 10*time.Second)
	cache.Delete("delete-key")

	_, ok = cache.Get("delete-key")
	if ok {
		t.Error("Cached value should have been deleted")
	}

	// Test Cleanup
	cache.Set("cleanup-key-1", "value1", 1*time.Millisecond)
	cache.Set("cleanup-key-2", "value2", 10*time.Second)

	time.Sleep(10 * time.Millisecond)

	cache.Cleanup()

	_, ok = cache.Get("cleanup-key-1")
	if ok {
		t.Error("Expired entry should have been cleaned up")
	}

	_, ok = cache.Get("cleanup-key-2")
	if !ok {
		t.Error("Non-expired entry should still exist")
	}
}

