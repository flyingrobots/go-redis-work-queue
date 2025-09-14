// Copyright 2025 James Ross
package multicluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config represents the multi-cluster configuration
type Config struct {
	Clusters      []ClusterConfig `json:"clusters"`
	DefaultCluster string         `json:"default_cluster"`
	Polling       PollingConfig   `json:"polling"`
	CompareMode   CompareModeConfig `json:"compare_mode"`
	Actions       ActionsConfig   `json:"actions"`
	Cache         CacheConfig     `json:"cache"`
	UI            UIConfig        `json:"ui"`
	mu            sync.RWMutex
}

// CompareModeConfig represents compare mode configuration
type CompareModeConfig struct {
	Enabled          bool     `json:"enabled"`
	DefaultClusters  []string `json:"default_clusters"`
	HighlightDeltas  bool     `json:"highlight_deltas"`
	DeltaThreshold   float64  `json:"delta_threshold"`
	RefreshInterval  string   `json:"refresh_interval"`
}

// ActionsConfig represents multi-action configuration
type ActionsConfig struct {
	RequireConfirmation bool                    `json:"require_confirmation"`
	AllowedActions      []ActionType            `json:"allowed_actions"`
	ActionTimeouts      map[ActionType]Duration `json:"action_timeouts"`
	MaxConcurrent       int                     `json:"max_concurrent"`
	RetryPolicy         RetryPolicy             `json:"retry_policy"`
}

// RetryPolicy represents retry configuration for actions
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled        bool          `json:"enabled"`
	TTL            time.Duration `json:"ttl"`
	MaxEntries     int           `json:"max_entries"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// UIConfig represents UI-specific configuration
type UIConfig struct {
	TabShortcuts    map[int]string `json:"tab_shortcuts"`
	Colors          ColorScheme    `json:"colors"`
	RefreshRate     time.Duration  `json:"refresh_rate"`
	ShowHealthBar   bool           `json:"show_health_bar"`
	CompactMode     bool           `json:"compact_mode"`
}

// ColorScheme represents color configuration for the UI
type ColorScheme struct {
	Healthy      string `json:"healthy"`
	Warning      string `json:"warning"`
	Error        string `json:"error"`
	Disconnected string `json:"disconnected"`
	Highlight    string `json:"highlight"`
}

// Duration is a custom type for JSON unmarshaling of duration strings
type Duration time.Duration

// UnmarshalJSON implements custom unmarshaling for Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// MarshalJSON implements custom marshaling for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Clusters: []ClusterConfig{
			{
				Name:     "local",
				Label:    "Local",
				Color:    "green",
				Endpoint: "localhost:6379",
				DB:       0,
				Enabled:  true,
			},
		},
		DefaultCluster: "local",
		Polling: PollingConfig{
			Interval: 5 * time.Second,
			Jitter:   1 * time.Second,
			Timeout:  3 * time.Second,
			Enabled:  true,
		},
		CompareMode: CompareModeConfig{
			Enabled:         false,
			HighlightDeltas: true,
			DeltaThreshold:  10.0,
			RefreshInterval: "5s",
		},
		Actions: ActionsConfig{
			RequireConfirmation: true,
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypePauseQueue,
				ActionTypeResumeQueue,
				ActionTypeBenchmark,
			},
			ActionTimeouts: map[ActionType]Duration{
				ActionTypePurgeDLQ:    Duration(30 * time.Second),
				ActionTypePauseQueue:  Duration(10 * time.Second),
				ActionTypeResumeQueue: Duration(10 * time.Second),
				ActionTypeBenchmark:   Duration(60 * time.Second),
			},
			MaxConcurrent: 5,
			RetryPolicy: RetryPolicy{
				MaxAttempts:  3,
				InitialDelay: 1 * time.Second,
				MaxDelay:     10 * time.Second,
				Multiplier:   2.0,
			},
		},
		Cache: CacheConfig{
			Enabled:         true,
			TTL:             30 * time.Second,
			MaxEntries:      1000,
			CleanupInterval: 60 * time.Second,
		},
		UI: UIConfig{
			TabShortcuts: map[int]string{
				1: "1",
				2: "2",
				3: "3",
				4: "4",
				5: "5",
				6: "6",
				7: "7",
				8: "8",
				9: "9",
			},
			Colors: ColorScheme{
				Healthy:      "green",
				Warning:      "yellow",
				Error:        "red",
				Disconnected: "gray",
				Highlight:    "cyan",
			},
			RefreshRate:   1 * time.Second,
			ShowHealthBar: true,
			CompactMode:   false,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.Clusters) == 0 {
		return errors.New("at least one cluster must be configured")
	}

	// Check for duplicate cluster names
	names := make(map[string]bool)
	for _, cluster := range c.Clusters {
		if cluster.Name == "" {
			return errors.New("cluster name cannot be empty")
		}
		if names[cluster.Name] {
			return fmt.Errorf("duplicate cluster name: %s", cluster.Name)
		}
		names[cluster.Name] = true

		if cluster.Endpoint == "" {
			return fmt.Errorf("cluster %s: endpoint cannot be empty", cluster.Name)
		}
	}

	// Validate default cluster exists
	if c.DefaultCluster != "" {
		if !names[c.DefaultCluster] {
			return fmt.Errorf("default cluster %s not found in clusters", c.DefaultCluster)
		}
	}

	// Validate polling config
	if c.Polling.Interval < time.Second {
		return errors.New("polling interval must be at least 1 second")
	}
	if c.Polling.Timeout > c.Polling.Interval {
		return errors.New("polling timeout cannot exceed polling interval")
	}

	// Validate actions config
	if c.Actions.MaxConcurrent < 1 {
		return errors.New("max concurrent actions must be at least 1")
	}
	if c.Actions.RetryPolicy.MaxAttempts < 0 {
		return errors.New("max retry attempts cannot be negative")
	}

	// Validate cache config
	if c.Cache.MaxEntries < 0 {
		return errors.New("max cache entries cannot be negative")
	}

	return nil
}

// GetCluster returns a cluster configuration by name
func (c *Config) GetCluster(name string) (*ClusterConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not found", name)
}

// AddCluster adds a new cluster configuration
func (c *Config) AddCluster(cluster ClusterConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cluster already exists
	for _, existing := range c.Clusters {
		if existing.Name == cluster.Name {
			return fmt.Errorf("cluster %s already exists", cluster.Name)
		}
	}

	c.Clusters = append(c.Clusters, cluster)
	return nil
}

// RemoveCluster removes a cluster configuration
func (c *Config) RemoveCluster(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, cluster := range c.Clusters {
		if cluster.Name == name {
			// Remove the cluster
			c.Clusters = append(c.Clusters[:i], c.Clusters[i+1:]...)

			// Update default cluster if necessary
			if c.DefaultCluster == name {
				if len(c.Clusters) > 0 {
					c.DefaultCluster = c.Clusters[0].Name
				} else {
					c.DefaultCluster = ""
				}
			}

			return nil
		}
	}
	return fmt.Errorf("cluster %s not found", name)
}

// UpdateCluster updates an existing cluster configuration
func (c *Config) UpdateCluster(name string, update ClusterConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, cluster := range c.Clusters {
		if cluster.Name == name {
			// If the name is changing, check for duplicates
			if update.Name != name {
				for _, other := range c.Clusters {
					if other.Name == update.Name {
						return fmt.Errorf("cluster %s already exists", update.Name)
					}
				}
			}

			c.Clusters[i] = update

			// Update default cluster if necessary
			if c.DefaultCluster == name && update.Name != name {
				c.DefaultCluster = update.Name
			}

			return nil
		}
	}
	return fmt.Errorf("cluster %s not found", name)
}

// GetEnabledClusters returns all enabled clusters
func (c *Config) GetEnabledClusters() []ClusterConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var enabled []ClusterConfig
	for _, cluster := range c.Clusters {
		if cluster.Enabled {
			enabled = append(enabled, cluster)
		}
	}
	return enabled
}

// IsActionAllowed checks if an action type is allowed
func (c *Config) IsActionAllowed(actionType ActionType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, allowed := range c.Actions.AllowedActions {
		if allowed == actionType {
			return true
		}
	}
	return false
}

// GetActionTimeout returns the timeout for an action type
func (c *Config) GetActionTimeout(actionType ActionType) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if timeout, ok := c.Actions.ActionTimeouts[actionType]; ok {
		return time.Duration(timeout)
	}
	// Default timeout
	return 30 * time.Second
}