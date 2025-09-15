package storage

import (
	"fmt"
	"time"
)

// Config represents the overall storage configuration
type Config struct {
	Backends map[string]BackendConfig `json:"backends" yaml:"backends"`
	Queues   map[string]QueueConfig   `json:"queues" yaml:"queues"`
	Defaults DefaultsConfig           `json:"defaults" yaml:"defaults"`
}

// QueueConfig configures a specific queue
type QueueConfig struct {
	Backend string                 `json:"backend" yaml:"backend"`
	Options map[string]interface{} `json:"options" yaml:"options"`
}

// DefaultsConfig provides default configuration values
type DefaultsConfig struct {
	Backend      string        `json:"backend" yaml:"backend"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	ClusterMode  bool          `json:"cluster_mode" yaml:"cluster_mode"`
	TLS          bool          `json:"tls" yaml:"tls"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate that all queues reference valid backends
	for queueName, queueConfig := range c.Queues {
		if _, exists := c.Backends[queueConfig.Backend]; !exists {
			return fmt.Errorf("queue %q references non-existent backend %q", queueName, queueConfig.Backend)
		}
	}

	// Validate individual backend configs
	registry := DefaultRegistry()
	for backendName, backendConfig := range c.Backends {
		if err := registry.Validate(backendConfig.Type, backendConfig); err != nil {
			return fmt.Errorf("invalid configuration for backend %q: %w", backendName, err)
		}
	}

	return nil
}

// GetBackendConfig returns the configuration for a specific backend
func (c *Config) GetBackendConfig(backendName string) (BackendConfig, error) {
	config, exists := c.Backends[backendName]
	if !exists {
		return BackendConfig{}, fmt.Errorf("backend %q not found", backendName)
	}
	return config, nil
}

// GetQueueConfig returns the configuration for a specific queue
func (c *Config) GetQueueConfig(queueName string) (QueueConfig, error) {
	config, exists := c.Queues[queueName]
	if !exists {
		// Return default configuration
		return QueueConfig{
			Backend: c.Defaults.Backend,
			Options: make(map[string]interface{}),
		}, nil
	}
	return config, nil
}

// ApplyDefaults applies default values to backend configurations
func (c *Config) ApplyDefaults() {
	for name, backend := range c.Backends {
		// Apply defaults based on backend type
		switch backend.Type {
		case BackendTypeRedisLists:
			c.applyRedisListsDefaults(&backend)
		case BackendTypeRedisStreams:
			c.applyRedisStreamsDefaults(&backend)
		case BackendTypeKeyDB:
			c.applyKeyDBDefaults(&backend)
		case BackendTypeDragonfly:
			c.applyDragonflyDefaults(&backend)
		}
		c.Backends[name] = backend
	}
}

func (c *Config) applyRedisListsDefaults(config *BackendConfig) {
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}

	setDefault := func(key string, value interface{}) {
		if _, exists := config.Options[key]; !exists {
			config.Options[key] = value
		}
	}

	setDefault("max_connections", c.Defaults.PoolSize)
	setDefault("conn_timeout", c.Defaults.Timeout)
	setDefault("read_timeout", c.Defaults.ReadTimeout)
	setDefault("write_timeout", c.Defaults.WriteTimeout)
	setDefault("idle_timeout", c.Defaults.IdleTimeout)
	setDefault("max_retries", c.Defaults.MaxRetries)
	setDefault("cluster_mode", c.Defaults.ClusterMode)
	setDefault("tls", c.Defaults.TLS)
	setDefault("key_prefix", "queue:")
}

func (c *Config) applyRedisStreamsDefaults(config *BackendConfig) {
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}

	setDefault := func(key string, value interface{}) {
		if _, exists := config.Options[key]; !exists {
			config.Options[key] = value
		}
	}

	setDefault("max_connections", c.Defaults.PoolSize)
	setDefault("conn_timeout", c.Defaults.Timeout)
	setDefault("read_timeout", c.Defaults.ReadTimeout)
	setDefault("write_timeout", c.Defaults.WriteTimeout)
	setDefault("idle_timeout", c.Defaults.IdleTimeout)
	setDefault("max_retries", c.Defaults.MaxRetries)
	setDefault("cluster_mode", c.Defaults.ClusterMode)
	setDefault("tls", c.Defaults.TLS)
	setDefault("block_timeout", "1s")
	setDefault("claim_min_idle", "30s")
	setDefault("claim_count", int64(100))
	setDefault("max_length", int64(10000))
}

func (c *Config) applyKeyDBDefaults(config *BackendConfig) {
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}

	setDefault := func(key string, value interface{}) {
		if _, exists := config.Options[key]; !exists {
			config.Options[key] = value
		}
	}

	setDefault("max_connections", c.Defaults.PoolSize*2) // KeyDB can handle more connections
	setDefault("conn_timeout", c.Defaults.Timeout)
	setDefault("read_timeout", time.Duration(100)*time.Millisecond) // Aggressive timeouts
	setDefault("write_timeout", time.Duration(200)*time.Millisecond)
	setDefault("idle_timeout", c.Defaults.IdleTimeout)
	setDefault("max_retries", c.Defaults.MaxRetries)
	setDefault("cluster_mode", c.Defaults.ClusterMode)
	setDefault("tls", c.Defaults.TLS)
	setDefault("key_prefix", "queue:")
	setDefault("pipeline_size", c.Defaults.BatchSize)
}

func (c *Config) applyDragonflyDefaults(config *BackendConfig) {
	if config.Options == nil {
		config.Options = make(map[string]interface{})
	}

	setDefault := func(key string, value interface{}) {
		if _, exists := config.Options[key]; !exists {
			config.Options[key] = value
		}
	}

	setDefault("max_connections", c.Defaults.PoolSize*3) // Dragonfly can handle even more
	setDefault("conn_timeout", c.Defaults.Timeout)
	setDefault("read_timeout", time.Duration(50)*time.Millisecond) // Very aggressive
	setDefault("write_timeout", time.Duration(100)*time.Millisecond)
	setDefault("idle_timeout", c.Defaults.IdleTimeout)
	setDefault("max_retries", c.Defaults.MaxRetries)
	setDefault("cluster_mode", c.Defaults.ClusterMode)
	setDefault("tls", c.Defaults.TLS)
	setDefault("key_prefix", "queue:")
	setDefault("pipeline_size", c.Defaults.BatchSize*2)
	setDefault("compression", true)
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Backends: map[string]BackendConfig{
			"redis-default": {
				Type: BackendTypeRedisLists,
				Name: "redis-default",
				URL:  "redis://localhost:6379/0",
				Options: map[string]interface{}{
					"key_prefix": "queue:",
				},
			},
		},
		Queues: map[string]QueueConfig{
			"default": {
				Backend: "redis-default",
				Options: make(map[string]interface{}),
			},
		},
		Defaults: DefaultsConfig{
			Backend:      BackendTypeRedisLists,
			MaxRetries:   3,
			Timeout:      30 * time.Second,
			BatchSize:    100,
			ClusterMode:  false,
			TLS:          false,
			PoolSize:     10,
			IdleTimeout:  5 * time.Minute,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		},
	}
}

// ExampleConfigs provides example configurations for different scenarios
func ExampleConfigs() map[string]*Config {
	return map[string]*Config{
		"simple": {
			Backends: map[string]BackendConfig{
				"redis": {
					Type: BackendTypeRedisLists,
					Name: "redis",
					URL:  "redis://localhost:6379/0",
				},
			},
			Queues: map[string]QueueConfig{
				"default": {Backend: "redis"},
				"high-priority": {Backend: "redis"},
			},
			Defaults: DefaultsConfig{
				Backend: BackendTypeRedisLists,
			},
		},
		"streams": {
			Backends: map[string]BackendConfig{
				"redis-streams": {
					Type: BackendTypeRedisStreams,
					Name: "redis-streams",
					URL:  "redis://localhost:6379/0",
					Options: map[string]interface{}{
						"stream_name":    "job-stream",
						"consumer_group": "workers",
						"consumer_name":  "worker-1",
					},
				},
			},
			Queues: map[string]QueueConfig{
				"analytics": {Backend: "redis-streams"},
				"reporting": {Backend: "redis-streams"},
			},
			Defaults: DefaultsConfig{
				Backend: BackendTypeRedisStreams,
			},
		},
		"high-performance": {
			Backends: map[string]BackendConfig{
				"keydb": {
					Type: BackendTypeKeyDB,
					Name: "keydb",
					URL:  "redis://keydb-cluster:6379",
					Options: map[string]interface{}{
						"cluster_mode":   true,
						"cluster_addrs":  []string{"keydb1:6379", "keydb2:6379", "keydb3:6379"},
						"pipeline_size":  1000,
						"max_connections": 50,
					},
				},
			},
			Queues: map[string]QueueConfig{
				"bulk-processing": {Backend: "keydb"},
				"real-time": {Backend: "keydb"},
			},
			Defaults: DefaultsConfig{
				Backend: BackendTypeKeyDB,
			},
		},
		"mixed": {
			Backends: map[string]BackendConfig{
				"redis-lists": {
					Type: BackendTypeRedisLists,
					Name: "redis-lists",
					URL:  "redis://localhost:6379/0",
				},
				"redis-streams": {
					Type: BackendTypeRedisStreams,
					Name: "redis-streams",
					URL:  "redis://localhost:6379/1",
					Options: map[string]interface{}{
						"stream_name":    "audit-stream",
						"consumer_group": "auditors",
					},
				},
				"dragonfly": {
					Type: BackendTypeDragonfly,
					Name: "dragonfly",
					URL:  "redis://dragonfly:6379",
					Options: map[string]interface{}{
						"compression": true,
						"pipeline_size": 2000,
					},
				},
			},
			Queues: map[string]QueueConfig{
				"default":     {Backend: "redis-lists"},
				"audit":       {Backend: "redis-streams"},
				"bulk":        {Backend: "dragonfly"},
				"analytics":   {Backend: "redis-streams"},
				"high-volume": {Backend: "dragonfly"},
			},
			Defaults: DefaultsConfig{
				Backend: BackendTypeRedisLists,
			},
		},
	}
}