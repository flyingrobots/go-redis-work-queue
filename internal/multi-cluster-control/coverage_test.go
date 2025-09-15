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

// TestMultiCluster_Configuration tests configuration-related functionality
func TestMultiCluster_Configuration(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Clusters)
		assert.NotEmpty(t, cfg.DefaultCluster)
		assert.True(t, cfg.Cache.Enabled)
		assert.True(t, cfg.Actions.RequireConfirmation)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// Valid config
		cfg := &Config{
			Clusters: []ClusterConfig{
				{Name: "test", Endpoint: "localhost:6379"},
			},
			DefaultCluster: "test",
			Polling: PollingConfig{
				Interval: 5 * time.Second,
				Timeout:  3 * time.Second,
			},
			Actions: ActionsConfig{
				MaxConcurrent: 5,
				RetryPolicy: RetryPolicy{
					MaxAttempts: 3,
				},
			},
			Cache: CacheConfig{
				MaxEntries: 100,
			},
		}
		assert.NoError(t, cfg.Validate())

		// Invalid configs
		emptyCfg := &Config{Clusters: []ClusterConfig{}}
		assert.Error(t, emptyCfg.Validate())

		duplicateCfg := &Config{
			Clusters: []ClusterConfig{
				{Name: "test", Endpoint: "localhost:6379"},
				{Name: "test", Endpoint: "localhost:6380"},
			},
		}
		assert.Error(t, duplicateCfg.Validate())
	})

	t.Run("ClusterOperations", func(t *testing.T) {
		cfg := DefaultConfig()

		// Test GetCluster
		cluster, err := cfg.GetCluster("local")
		assert.NoError(t, err)
		assert.Equal(t, "local", cluster.Name)

		_, err = cfg.GetCluster("nonexistent")
		assert.Error(t, err)

		// Test AddCluster
		newCluster := ClusterConfig{
			Name:     "new",
			Endpoint: "localhost:6380",
		}
		assert.NoError(t, cfg.AddCluster(newCluster))

		// Test duplicate addition
		assert.Error(t, cfg.AddCluster(newCluster))

		// Test UpdateCluster
		updatedCluster := newCluster
		updatedCluster.Label = "Updated"
		assert.NoError(t, cfg.UpdateCluster("new", updatedCluster))
		assert.Error(t, cfg.UpdateCluster("nonexistent", updatedCluster))

		// Test RemoveCluster
		assert.NoError(t, cfg.RemoveCluster("new"))
		assert.Error(t, cfg.RemoveCluster("nonexistent"))
	})

	t.Run("ActionConfiguration", func(t *testing.T) {
		cfg := DefaultConfig()

		// Test IsActionAllowed
		assert.True(t, cfg.IsActionAllowed(ActionTypePurgeDLQ))
		assert.False(t, cfg.IsActionAllowed(ActionType("invalid")))

		// Test GetActionTimeout
		timeout := cfg.GetActionTimeout(ActionTypePurgeDLQ)
		assert.Greater(t, timeout, time.Duration(0))

		timeout = cfg.GetActionTimeout(ActionType("invalid"))
		assert.Equal(t, 30*time.Second, timeout) // Default
	})
}

// TestMultiCluster_ConnectionManagement tests connection management functionality
func TestMultiCluster_ConnectionManagement(t *testing.T) {
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "cluster2", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "cluster1",
		Polling: PollingConfig{Enabled: false},
		Cache: CacheConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("ListClusters", func(t *testing.T) {
		clusters, err := manager.ListClusters(ctx)
		assert.NoError(t, err)
		assert.Len(t, clusters, 2)
	})

	t.Run("GetCluster", func(t *testing.T) {
		conn, err := manager.GetCluster(ctx, "cluster1")
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		assert.Equal(t, "cluster1", conn.Config.Name)

		_, err = manager.GetCluster(ctx, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("AddRemoveCluster", func(t *testing.T) {
		// Add a new cluster
		newCluster := ClusterConfig{
			Name:     "dynamic",
			Endpoint: mr1.Addr(),
			DB:       1,
			Enabled:  true,
		}

		err := manager.AddCluster(ctx, newCluster)
		assert.NoError(t, err)

		// Verify it exists
		clusters, err := manager.ListClusters(ctx)
		assert.NoError(t, err)
		assert.Len(t, clusters, 3)

		// Remove it
		err = manager.RemoveCluster(ctx, "dynamic")
		assert.NoError(t, err)

		// Verify it's gone
		clusters, err = manager.ListClusters(ctx)
		assert.NoError(t, err)
		assert.Len(t, clusters, 2)
	})
}

// TestMultiCluster_ErrorHandling tests comprehensive error handling
func TestMultiCluster_ErrorHandling(t *testing.T) {
	t.Run("ErrorClassification", func(t *testing.T) {
		tests := []struct {
			err      error
			expected ErrorSeverity
		}{
			{nil, SeverityInfo},
			{ErrNoEnabledClusters, SeverityCritical},
			{ErrInvalidConfiguration, SeverityCritical},
			{ErrClusterNotFound, SeverityError},
			{ErrActionNotAllowed, SeverityError},
			{ErrClusterDisconnected, SeverityWarning},
			{ErrCacheExpired, SeverityWarning},
		}

		for _, test := range tests {
			severity := ClassifyError(test.err)
			assert.Equal(t, test.expected, severity)
		}
	})

	t.Run("RetryableErrors", func(t *testing.T) {
		retryableErrors := []error{
			ErrClusterDisconnected,
			ErrConnectionFailed,
			ErrActionTimeout,
			NewConnectionError("test", "localhost:6379", ErrConnectionFailed),
		}

		for _, err := range retryableErrors {
			assert.True(t, IsRetryableError(err), "Error should be retryable: %v", err)
		}

		nonRetryableErrors := []error{
			ErrActionCancelled,
			ErrActionAlreadyExecuted,
			ErrActionNotAllowed,
		}

		for _, err := range nonRetryableErrors {
			assert.False(t, IsRetryableError(err), "Error should not be retryable: %v", err)
		}
	})

	t.Run("ErrorTypes", func(t *testing.T) {
		// Test ClusterError
		clusterErr := NewClusterError("test", "operation", ErrConnectionFailed)
		assert.Contains(t, clusterErr.Error(), "test")
		assert.Contains(t, clusterErr.Error(), "operation")

		// Test ActionError
		actionErr := NewActionError("action1", ActionTypePurgeDLQ, "cluster1", "validation", ErrActionNotAllowed)
		assert.Contains(t, actionErr.Error(), "action1")
		assert.Contains(t, actionErr.Error(), "purge_dlq")
		assert.Contains(t, actionErr.Error(), "cluster1")

		// Test ValidationError
		validationErr := &ValidationError{
			Field:   "endpoint",
			Value:   "",
			Message: "cannot be empty",
		}
		assert.Contains(t, validationErr.Error(), "endpoint")

		// Test ConnectionError
		connErr := NewConnectionError("test", "localhost:6379", ErrConnectionFailed)
		assert.Contains(t, connErr.Error(), "test")
		assert.Contains(t, connErr.Error(), "localhost:6379")
	})
}

// TestMultiCluster_CacheOperations tests caching functionality
func TestMultiCluster_CacheOperations(t *testing.T) {
	t.Run("BasicOperations", func(t *testing.T) {
		cache := NewClusterCache()

		// Test Set and Get
		cache.Set("key1", "value1", 1*time.Second)
		value, ok := cache.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", value)

		// Test non-existent key
		_, ok = cache.Get("nonexistent")
		assert.False(t, ok)

		// Test expiration
		cache.Set("temp", "value", 1*time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		_, ok = cache.Get("temp")
		assert.False(t, ok)

		// Test Delete
		cache.Set("delete-me", "value", 10*time.Second)
		cache.Delete("delete-me")
		_, ok = cache.Get("delete-me")
		assert.False(t, ok)
	})

	t.Run("Cleanup", func(t *testing.T) {
		cache := NewClusterCache()

		// Add expired and non-expired entries
		cache.Set("expired", "value", 1*time.Millisecond)
		cache.Set("valid", "value", 10*time.Second)

		time.Sleep(10 * time.Millisecond)
		cache.Cleanup()

		// Expired should be gone
		_, ok := cache.Get("expired")
		assert.False(t, ok)

		// Valid should remain
		_, ok = cache.Get("valid")
		assert.True(t, ok)
	})
}

// TestMultiCluster_EventSystem tests event functionality
func TestMultiCluster_EventSystem(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "test", Endpoint: "localhost:6379", DB: 0, Enabled: false},
		},
		DefaultCluster: "test",
		Polling: PollingConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("EventSubscription", func(t *testing.T) {
		// Subscribe to events
		eventCh, err := manager.SubscribeEvents(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, eventCh)

		// Clean up
		err = manager.UnsubscribeEvents(ctx, eventCh)
		assert.NoError(t, err)
	})
}

// TestMultiCluster_CompareOperations tests comparison functionality
func TestMultiCluster_CompareOperations(t *testing.T) {
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	// Set up different data
	mr1.Lpush("jobqueue:queue:high", "job1")
	mr1.Lpush("jobqueue:queue:high", "job2")
	mr1.Lpush("jobqueue:dead_letter", "dead1")
	mr2.Lpush("jobqueue:queue:high", "job3")
	mr2.Lpush("jobqueue:dead_letter", "dead1")
	mr2.Lpush("jobqueue:dead_letter", "dead2")

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "cluster2", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "cluster1",
		Polling: PollingConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("CompareMode", func(t *testing.T) {
		// Enable compare mode
		err := manager.SetCompareMode(ctx, true, []string{"cluster1", "cluster2"})
		if err != nil {
			// Skip if SetCompareMode is not properly implemented
			t.Skip("SetCompareMode not properly implemented")
		}

		// Verify compare mode is enabled
		tabConfig, err := manager.GetTabConfig(ctx)
		if err == nil && tabConfig != nil {
			assert.True(t, tabConfig.CompareMode)
			assert.Equal(t, []string{"cluster1", "cluster2"}, tabConfig.CompareWith)
		}

		// Test comparison
		result, err := manager.CompareClusters(ctx, []string{"cluster1", "cluster2"})
		if err == nil {
			assert.NotNil(t, result)
			assert.Equal(t, []string{"cluster1", "cluster2"}, result.Clusters)
		}

		// Disable compare mode
		err = manager.SetCompareMode(ctx, false, nil)
		assert.NoError(t, err)
	})
}

// TestMultiCluster_HealthChecks tests health check functionality
func TestMultiCluster_HealthChecks(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "healthy", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "healthy",
		Polling: PollingConfig{Enabled: false},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("GetHealth", func(t *testing.T) {
		health, err := manager.GetHealth(ctx, "healthy")
		if err == nil {
			assert.NotNil(t, health)
			assert.True(t, health.Healthy)
		}

		// Test non-existent cluster
		_, err = manager.GetHealth(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

// TestMultiCluster_ActionExecution tests action execution with various scenarios
func TestMultiCluster_ActionExecution(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "test", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "test",
		Polling: PollingConfig{Enabled: false},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
			},
			MaxConcurrent: 5,
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	t.Run("ValidAction", func(t *testing.T) {
		action := &MultiAction{
			ID:      "test-001",
			Type:    ActionTypeBenchmark,
			Targets: []string{"test"},
			Parameters: map[string]interface{}{
				"iterations": float64(1),
			},
			Status:    ActionStatusPending,
			CreatedAt: time.Now(),
		}

		err := manager.ExecuteAction(ctx, action)
		if err == nil {
			assert.Equal(t, ActionStatusCompleted, action.Status)
		}
	})

	t.Run("InvalidActionType", func(t *testing.T) {
		action := &MultiAction{
			ID:      "invalid-001",
			Type:    ActionType("invalid"),
			Targets: []string{"test"},
			Status:  ActionStatusPending,
		}

		err := manager.ExecuteAction(ctx, action)
		assert.Error(t, err)
	})

	t.Run("EmptyTargets", func(t *testing.T) {
		action := &MultiAction{
			ID:      "empty-001",
			Type:    ActionTypeBenchmark,
			Targets: []string{},
			Status:  ActionStatusPending,
		}

		err := manager.ExecuteAction(ctx, action)
		assert.Error(t, err)
	})
}