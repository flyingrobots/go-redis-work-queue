package multicluster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Clusters)
	assert.Empty(t, cfg.DefaultCluster)
	assert.True(t, cfg.Polling.Enabled)
	assert.Equal(t, 30*time.Second, cfg.Polling.Interval.Duration())
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.Cache.TTL.Duration())
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal config",
			config: &Config{
				Clusters: []ClusterConfig{
					{
						Name:     "test",
						Endpoint: "localhost:6379",
						Enabled:  true,
					},
				},
				DefaultCluster: "test",
			},
			wantErr: false,
		},
		{
			name: "empty clusters",
			config: &Config{
				Clusters: []ClusterConfig{},
			},
			wantErr: true,
			errMsg:  "no clusters configured",
		},
		{
			name: "duplicate cluster names",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "test", Endpoint: "localhost:6379"},
					{Name: "test", Endpoint: "localhost:6380"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate cluster name",
		},
		{
			name: "invalid default cluster",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "cluster1", Endpoint: "localhost:6379"},
				},
				DefaultCluster: "cluster2",
			},
			wantErr: true,
			errMsg:  "default cluster 'cluster2' not found",
		},
		{
			name: "missing cluster name",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "", Endpoint: "localhost:6379"},
				},
			},
			wantErr: true,
			errMsg:  "cluster name cannot be empty",
		},
		{
			name: "missing cluster endpoint",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "test", Endpoint: ""},
				},
			},
			wantErr: true,
			errMsg:  "cluster endpoint cannot be empty",
		},
		{
			name: "invalid polling interval",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "test", Endpoint: "localhost:6379"},
				},
				Polling: PollingConfig{
					Interval: Duration(0),
				},
			},
			wantErr: true,
			errMsg:  "polling interval must be positive",
		},
		{
			name: "invalid cache TTL",
			config: &Config{
				Clusters: []ClusterConfig{
					{Name: "test", Endpoint: "localhost:6379"},
				},
				Cache: CacheConfig{
					Enabled: true,
					TTL:     Duration(0),
				},
			},
			wantErr: true,
			errMsg:  "cache TTL must be positive when cache is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigAddCluster(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "existing", Endpoint: "localhost:6379"},
		},
	}

	// Test adding new cluster
	newCluster := ClusterConfig{
		Name:     "new",
		Endpoint: "localhost:6380",
		Enabled:  true,
	}

	err := cfg.AddCluster(newCluster)
	assert.NoError(t, err)
	assert.Len(t, cfg.Clusters, 2)
	assert.Equal(t, "new", cfg.Clusters[1].Name)

	// Test adding duplicate cluster
	err = cfg.AddCluster(newCluster)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConfigRemoveCluster(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: "localhost:6379"},
			{Name: "cluster2", Endpoint: "localhost:6380"},
		},
		DefaultCluster: "cluster1",
	}

	// Test removing existing cluster
	err := cfg.RemoveCluster("cluster2")
	assert.NoError(t, err)
	assert.Len(t, cfg.Clusters, 1)
	assert.Equal(t, "cluster1", cfg.Clusters[0].Name)

	// Test removing non-existent cluster
	err = cfg.RemoveCluster("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test removing default cluster (should fail)
	err = cfg.RemoveCluster("cluster1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove default cluster")
}

func TestConfigGetEnabledClusters(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "enabled1", Endpoint: "localhost:6379", Enabled: true},
			{Name: "disabled", Endpoint: "localhost:6380", Enabled: false},
			{Name: "enabled2", Endpoint: "localhost:6381", Enabled: true},
		},
	}

	enabled := cfg.GetEnabledClusters()
	assert.Len(t, enabled, 2)
	assert.Equal(t, "enabled1", enabled[0].Name)
	assert.Equal(t, "enabled2", enabled[1].Name)
}

func TestConfigIsActionAllowed(t *testing.T) {
	cfg := &Config{
		Actions: ActionsConfig{
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
			},
		},
	}

	assert.True(t, cfg.IsActionAllowed(ActionTypePurgeDLQ))
	assert.True(t, cfg.IsActionAllowed(ActionTypeBenchmark))
	assert.False(t, cfg.IsActionAllowed(ActionTypePauseQueue))
}

func TestConfigGetActionTimeout(t *testing.T) {
	cfg := &Config{
		Actions: ActionsConfig{
			ActionTimeouts: map[ActionType]Duration{
				ActionTypePurgeDLQ:  Duration(30 * time.Second),
				ActionTypeBenchmark: Duration(60 * time.Second),
			},
			DefaultTimeout: Duration(10 * time.Second),
		},
	}

	// Test specific timeout
	timeout := cfg.GetActionTimeout(ActionTypePurgeDLQ)
	assert.Equal(t, 30*time.Second, timeout)

	// Test default timeout
	timeout = cfg.GetActionTimeout(ActionTypePauseQueue)
	assert.Equal(t, 10*time.Second, timeout)
}

func TestClusterConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cluster ClusterConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cluster",
			cluster: ClusterConfig{
				Name:     "test",
				Endpoint: "localhost:6379",
				Enabled:  true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			cluster: ClusterConfig{
				Name:     "",
				Endpoint: "localhost:6379",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "empty endpoint",
			cluster: ClusterConfig{
				Name:     "test",
				Endpoint: "",
			},
			wantErr: true,
			errMsg:  "endpoint cannot be empty",
		},
		{
			name: "invalid DB number",
			cluster: ClusterConfig{
				Name:     "test",
				Endpoint: "localhost:6379",
				DB:       -1,
			},
			wantErr: true,
			errMsg:  "DB must be non-negative",
		},
		{
			name: "valid with all fields",
			cluster: ClusterConfig{
				Name:        "production",
				Label:       "Production Cluster",
				Color:       "#ff0000",
				Environment: "prod",
				Region:      "us-east-1",
				Endpoint:    "prod.redis.example.com:6379",
				Password:    "secret",
				DB:          0,
				Enabled:     true,
				Tags:        []string{"production", "critical"},
				Metadata: map[string]string{
					"owner": "platform-team",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cluster.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPollingConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		polling PollingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid polling config",
			polling: PollingConfig{
				Enabled:  true,
				Interval: Duration(30 * time.Second),
				Jitter:   Duration(5 * time.Second),
				Timeout:  Duration(10 * time.Second),
			},
			wantErr: false,
		},
		{
			name: "disabled polling",
			polling: PollingConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "zero interval",
			polling: PollingConfig{
				Enabled:  true,
				Interval: Duration(0),
			},
			wantErr: true,
			errMsg:  "interval must be positive",
		},
		{
			name: "negative timeout",
			polling: PollingConfig{
				Enabled:  true,
				Interval: Duration(30 * time.Second),
				Timeout:  Duration(-1 * time.Second),
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "jitter larger than interval",
			polling: PollingConfig{
				Enabled:  true,
				Interval: Duration(10 * time.Second),
				Jitter:   Duration(20 * time.Second),
			},
			wantErr: true,
			errMsg:  "jitter cannot be larger than interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.polling.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCacheConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cache   CacheConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cache config",
			cache: CacheConfig{
				Enabled:         true,
				TTL:             Duration(5 * time.Minute),
				MaxEntries:      1000,
				CleanupInterval: Duration(1 * time.Minute),
			},
			wantErr: false,
		},
		{
			name: "disabled cache",
			cache: CacheConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "zero TTL when enabled",
			cache: CacheConfig{
				Enabled: true,
				TTL:     Duration(0),
			},
			wantErr: true,
			errMsg:  "TTL must be positive when cache is enabled",
		},
		{
			name: "zero max entries",
			cache: CacheConfig{
				Enabled:    true,
				TTL:        Duration(5 * time.Minute),
				MaxEntries: 0,
			},
			wantErr: true,
			errMsg:  "max entries must be positive",
		},
		{
			name: "zero cleanup interval",
			cache: CacheConfig{
				Enabled:         true,
				TTL:             Duration(5 * time.Minute),
				MaxEntries:      1000,
				CleanupInterval: Duration(0),
			},
			wantErr: true,
			errMsg:  "cleanup interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cache.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestActionsConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		actions ActionsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid actions config",
			actions: ActionsConfig{
				RequireConfirmation: true,
				MaxConcurrent:       5,
				DefaultTimeout:      Duration(30 * time.Second),
				AllowedActions: []ActionType{
					ActionTypePurgeDLQ,
					ActionTypeBenchmark,
				},
				RetryPolicy: RetryPolicy{
					MaxAttempts: 3,
					Backoff:     Duration(1 * time.Second),
				},
			},
			wantErr: false,
		},
		{
			name: "zero max concurrent",
			actions: ActionsConfig{
				MaxConcurrent: 0,
			},
			wantErr: true,
			errMsg:  "max concurrent must be positive",
		},
		{
			name: "zero default timeout",
			actions: ActionsConfig{
				MaxConcurrent:  5,
				DefaultTimeout: Duration(0),
			},
			wantErr: true,
			errMsg:  "default timeout must be positive",
		},
		{
			name: "zero retry max attempts",
			actions: ActionsConfig{
				MaxConcurrent:  5,
				DefaultTimeout: Duration(30 * time.Second),
				RetryPolicy: RetryPolicy{
					MaxAttempts: 0,
				},
			},
			wantErr: true,
			errMsg:  "retry max attempts must be positive",
		},
		{
			name: "negative retry backoff",
			actions: ActionsConfig{
				MaxConcurrent:  5,
				DefaultTimeout: Duration(30 * time.Second),
				RetryPolicy: RetryPolicy{
					MaxAttempts: 3,
					Backoff:     Duration(-1 * time.Second),
				},
			},
			wantErr: true,
			errMsg:  "retry backoff cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.actions.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	d := Duration(5 * time.Second)
	assert.Equal(t, 5*time.Second, d.Duration())

	// Test JSON marshaling/unmarshaling
	data, err := d.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"5s"`, string(data))

	var unmarshaled Duration
	err = unmarshaled.UnmarshalJSON(data)
	require.NoError(t, err)
	assert.Equal(t, d, unmarshaled)

	// Test invalid JSON
	err = unmarshaled.UnmarshalJSON([]byte(`"invalid"`))
	assert.Error(t, err)
}

func TestConfigMerge(t *testing.T) {
	base := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: "localhost:6379"},
		},
		DefaultCluster: "cluster1",
		Polling: PollingConfig{
			Enabled:  true,
			Interval: Duration(30 * time.Second),
		},
	}

	override := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster2", Endpoint: "localhost:6380"},
		},
		Polling: PollingConfig{
			Interval: Duration(60 * time.Second),
		},
	}

	merged := base.Merge(override)

	// Should have both clusters
	assert.Len(t, merged.Clusters, 2)
	assert.Equal(t, "cluster1", merged.Clusters[0].Name)
	assert.Equal(t, "cluster2", merged.Clusters[1].Name)

	// Should use override's polling interval but base's enabled
	assert.True(t, merged.Polling.Enabled)
	assert.Equal(t, Duration(60*time.Second), merged.Polling.Interval)

	// Should keep base's default cluster
	assert.Equal(t, "cluster1", merged.DefaultCluster)
}

func TestConfigGetCluster(t *testing.T) {
	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "cluster1", Endpoint: "localhost:6379"},
			{Name: "cluster2", Endpoint: "localhost:6380"},
		},
	}

	// Test getting existing cluster
	cluster, err := cfg.GetCluster("cluster1")
	assert.NoError(t, err)
	assert.Equal(t, "cluster1", cluster.Name)

	// Test getting non-existent cluster
	_, err = cfg.GetCluster("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}