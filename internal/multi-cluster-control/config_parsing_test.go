//go:build multi_cluster_control_tests
// +build multi_cluster_control_tests

package multicluster

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)

	// Test default values
	assert.Len(t, cfg.Clusters, 1)
	assert.Equal(t, "local", cfg.DefaultCluster)
	assert.Equal(t, "local", cfg.Clusters[0].Name)
	assert.Equal(t, "localhost:6379", cfg.Clusters[0].Endpoint)
	assert.True(t, cfg.Clusters[0].Enabled)

	// Test polling config
	assert.Equal(t, 5*time.Second, cfg.Polling.Interval)
	assert.Equal(t, 1*time.Second, cfg.Polling.Jitter)
	assert.Equal(t, 3*time.Second, cfg.Polling.Timeout)
	assert.True(t, cfg.Polling.Enabled)

	// Test cache config
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, 30*time.Second, cfg.Cache.TTL)
	assert.Equal(t, 1000, cfg.Cache.MaxEntries)

	// Test actions config
	assert.True(t, cfg.Actions.RequireConfirmation)
	assert.Equal(t, 5, cfg.Actions.MaxConcurrent)
	assert.Contains(t, cfg.Actions.AllowedActions, ActionTypePurgeDLQ)
	assert.Contains(t, cfg.Actions.AllowedActions, ActionTypeBenchmark)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *Config
		wantError string
	}{
		{
			name: "valid config",
			setupFunc: func() *Config {
				return DefaultConfig()
			},
			wantError: "",
		},
		{
			name: "no clusters",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Clusters = []ClusterConfig{}
				return cfg
			},
			wantError: "at least one cluster must be configured",
		},
		{
			name: "empty cluster name",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Clusters[0].Name = ""
				return cfg
			},
			wantError: "cluster name cannot be empty",
		},
		{
			name: "duplicate cluster names",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Clusters = append(cfg.Clusters, ClusterConfig{
					Name:     "local", // Same as default
					Endpoint: "localhost:6380",
					Enabled:  true,
				})
				return cfg
			},
			wantError: "duplicate cluster name: local",
		},
		{
			name: "empty endpoint",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Clusters[0].Endpoint = ""
				return cfg
			},
			wantError: "cluster local: endpoint cannot be empty",
		},
		{
			name: "invalid default cluster",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.DefaultCluster = "nonexistent"
				return cfg
			},
			wantError: "default cluster nonexistent not found in clusters",
		},
		{
			name: "polling interval too short",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Polling.Interval = 500 * time.Millisecond
				return cfg
			},
			wantError: "polling interval must be at least 1 second",
		},
		{
			name: "polling timeout exceeds interval",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Polling.Timeout = 10 * time.Second
				cfg.Polling.Interval = 5 * time.Second
				return cfg
			},
			wantError: "polling timeout cannot exceed polling interval",
		},
		{
			name: "invalid max concurrent actions",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Actions.MaxConcurrent = 0
				return cfg
			},
			wantError: "max concurrent actions must be at least 1",
		},
		{
			name: "negative max retry attempts",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Actions.RetryPolicy.MaxAttempts = -1
				return cfg
			},
			wantError: "max retry attempts cannot be negative",
		},
		{
			name: "negative max cache entries",
			setupFunc: func() *Config {
				cfg := DefaultConfig()
				cfg.Cache.MaxEntries = -1
				return cfg
			},
			wantError: "max cache entries cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()
			err := cfg.Validate()

			if tt.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}

func TestConfigClusterManagement(t *testing.T) {
	cfg := DefaultConfig()

	// Test GetCluster
	cluster, err := cfg.GetCluster("local")
	require.NoError(t, err)
	assert.Equal(t, "local", cluster.Name)

	// Test GetCluster with non-existent cluster
	_, err = cfg.GetCluster("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster nonexistent not found")

	// Test AddCluster
	newCluster := ClusterConfig{
		Name:     "test",
		Label:    "Test Cluster",
		Endpoint: "localhost:6380",
		Enabled:  true,
	}
	err = cfg.AddCluster(newCluster)
	require.NoError(t, err)
	assert.Len(t, cfg.Clusters, 2)

	// Test AddCluster with duplicate name
	err = cfg.AddCluster(newCluster)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster test already exists")

	// Test RemoveCluster
	err = cfg.RemoveCluster("test")
	require.NoError(t, err)
	assert.Len(t, cfg.Clusters, 1)

	// Test RemoveCluster with non-existent cluster
	err = cfg.RemoveCluster("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster nonexistent not found")

	// Test removing default cluster updates default
	err = cfg.RemoveCluster("local")
	require.NoError(t, err)
	assert.Empty(t, cfg.DefaultCluster)
}

func TestUpdateCluster(t *testing.T) {
	cfg := DefaultConfig()

	// Add another cluster first
	err := cfg.AddCluster(ClusterConfig{
		Name:     "test",
		Endpoint: "localhost:6380",
		Enabled:  true,
	})
	require.NoError(t, err)

	// Test UpdateCluster
	updatedCluster := ClusterConfig{
		Name:     "test-updated",
		Label:    "Updated Test",
		Endpoint: "localhost:6381",
		Enabled:  false,
	}
	err = cfg.UpdateCluster("test", updatedCluster)
	require.NoError(t, err)

	cluster, err := cfg.GetCluster("test-updated")
	require.NoError(t, err)
	assert.Equal(t, "Updated Test", cluster.Label)
	assert.Equal(t, "localhost:6381", cluster.Endpoint)
	assert.False(t, cluster.Enabled)

	// Test UpdateCluster with duplicate name
	err = cfg.UpdateCluster("test-updated", ClusterConfig{
		Name:     "local", // Already exists
		Endpoint: "localhost:6382",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster local already exists")

	// Test UpdateCluster with non-existent cluster
	err = cfg.UpdateCluster("nonexistent", ClusterConfig{
		Name:     "new",
		Endpoint: "localhost:6383",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster nonexistent not found")
}

func TestGetEnabledClusters(t *testing.T) {
	cfg := DefaultConfig()

	// Add disabled cluster
	err := cfg.AddCluster(ClusterConfig{
		Name:     "disabled",
		Endpoint: "localhost:6380",
		Enabled:  false,
	})
	require.NoError(t, err)

	// Add enabled cluster
	err = cfg.AddCluster(ClusterConfig{
		Name:     "enabled",
		Endpoint: "localhost:6381",
		Enabled:  true,
	})
	require.NoError(t, err)

	enabled := cfg.GetEnabledClusters()
	assert.Len(t, enabled, 2) // local + enabled

	var enabledNames []string
	for _, cluster := range enabled {
		enabledNames = append(enabledNames, cluster.Name)
	}
	assert.Contains(t, enabledNames, "local")
	assert.Contains(t, enabledNames, "enabled")
	assert.NotContains(t, enabledNames, "disabled")
}

func TestIsActionAllowed(t *testing.T) {
	cfg := DefaultConfig()

	// Test allowed actions
	assert.True(t, cfg.IsActionAllowed(ActionTypePurgeDLQ))
	assert.True(t, cfg.IsActionAllowed(ActionTypeBenchmark))

	// Test disallowed action
	assert.False(t, cfg.IsActionAllowed(ActionTypeRebalance))

	// Modify allowed actions
	cfg.Actions.AllowedActions = []ActionType{ActionTypeRebalance}
	assert.False(t, cfg.IsActionAllowed(ActionTypePurgeDLQ))
	assert.True(t, cfg.IsActionAllowed(ActionTypeRebalance))
}

func TestGetActionTimeout(t *testing.T) {
	cfg := DefaultConfig()

	// Test configured timeout
	timeout := cfg.GetActionTimeout(ActionTypePurgeDLQ)
	assert.Equal(t, 30*time.Second, timeout)

	// Test unconfigured timeout (should return default)
	timeout = cfg.GetActionTimeout(ActionTypeRebalance)
	assert.Equal(t, 30*time.Second, timeout) // Default timeout
}

func TestLoadSaveConfig(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Test loading non-existent file (should return default config)
	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "local", cfg.DefaultCluster)

	// Modify config
	cfg.DefaultCluster = "test"
	cfg.Clusters = append(cfg.Clusters, ClusterConfig{
		Name:     "test",
		Endpoint: "localhost:6380",
		Enabled:  true,
	})

	// Save config
	err = cfg.SaveConfig(configPath)
	require.NoError(t, err)

	// Load saved config
	loadedCfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test", loadedCfg.DefaultCluster)
	assert.Len(t, loadedCfg.Clusters, 2)

	// Test loading invalid JSON
	invalidPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("{invalid json"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")

	// Test loading invalid config (validation fails)
	invalidCfg := DefaultConfig()
	invalidCfg.Clusters = []ClusterConfig{} // Invalid: no clusters
	invalidConfigPath := filepath.Join(tempDir, "invalid_config.json")
	err = invalidCfg.SaveConfig(invalidConfigPath)
	require.NoError(t, err)

	_, err = LoadConfig(invalidConfigPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestDurationMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		jsonStr  string
	}{
		{
			name:     "seconds",
			duration: Duration(30 * time.Second),
			jsonStr:  `"30s"`,
		},
		{
			name:     "minutes",
			duration: Duration(5 * time.Minute),
			jsonStr:  `"5m0s"`,
		},
		{
			name:     "hours",
			duration: Duration(2 * time.Hour),
			jsonStr:  `"2h0m0s"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := tt.duration.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.jsonStr, string(data))

			// Test unmarshaling
			var d Duration
			err = d.UnmarshalJSON([]byte(tt.jsonStr))
			require.NoError(t, err)
			assert.Equal(t, tt.duration, d)
		})
	}

	// Test invalid duration string
	var d Duration
	err := d.UnmarshalJSON([]byte(`"invalid"`))
	assert.Error(t, err)
}

func TestConfigConcurrency(t *testing.T) {
	cfg := DefaultConfig()

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_, err := cfg.GetCluster("local")
				assert.NoError(t, err)
				enabled := cfg.GetEnabledClusters()
				assert.NotEmpty(t, enabled)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent writes
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			clusterName := fmt.Sprintf("test-%d", id)
			err := cfg.AddCluster(ClusterConfig{
				Name:     clusterName,
				Endpoint: fmt.Sprintf("localhost:638%d", id),
				Enabled:  true,
			})
			// May or may not succeed due to race conditions, but shouldn't panic
			_ = err
		}(i)
	}

	// Wait for write goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}
