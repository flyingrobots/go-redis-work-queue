// Copyright 2025 James Ross
package smartretry

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DefaultConfig returns a default configuration for smart retry strategies
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		Strategy: RetryStrategy{
			Name:              "default",
			Enabled:           true,
			Policies:          defaultPolicies(),
			BayesianThreshold: 0.7,
			MLEnabled:         false,
			Guardrails: PolicyGuardrails{
				MaxAttempts:       10,
				MaxDelayMs:        300000, // 5 minutes
				MaxBudgetPercent:  20.0,
				PerTenantLimits:   true,
				EmergencyStop:     false,
				ExplainabilityReq: true,
			},
			DataCollection: DataCollectionConfig{
				Enabled:             true,
				SampleRate:          1.0,
				RetentionDays:       30,
				AggregationInterval: 5 * time.Minute,
				FeatureExtraction:   true,
			},
		},
		DataCollection: DataCollectionConfig{
			Enabled:             true,
			SampleRate:          1.0,
			RetentionDays:       30,
			AggregationInterval: 5 * time.Minute,
			FeatureExtraction:   true,
		},
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        5 * time.Minute,
			MaxEntries: 1000,
		},
		API: APIConfig{
			Enabled: true,
			Port:    8080,
			Path:    "/api/v1/retry",
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing fields
	if config.Strategy.Policies == nil {
		config.Strategy.Policies = defaultPolicies()
	}

	if config.Cache.TTL == 0 {
		config.Cache.TTL = 5 * time.Minute
	}

	if config.Cache.MaxEntries == 0 {
		config.Cache.MaxEntries = 1000
	}

	if config.API.Port == 0 {
		config.API.Port = 8080
	}

	if config.API.Path == "" {
		config.API.Path = "/api/v1/retry"
	}

	return &config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.RedisAddr == "" {
		return fmt.Errorf("redis_addr is required")
	}

	if c.Strategy.BayesianThreshold < 0 || c.Strategy.BayesianThreshold > 1 {
		return fmt.Errorf("bayesian_threshold must be between 0 and 1")
	}

	if c.Strategy.Guardrails.MaxAttempts <= 0 {
		return fmt.Errorf("max_attempts must be positive")
	}

	if c.Strategy.Guardrails.MaxDelayMs <= 0 {
		return fmt.Errorf("max_delay_ms must be positive")
	}

	if c.DataCollection.SampleRate < 0 || c.DataCollection.SampleRate > 1 {
		return fmt.Errorf("sample_rate must be between 0 and 1")
	}

	if c.DataCollection.RetentionDays <= 0 {
		return fmt.Errorf("retention_days must be positive")
	}

	if c.Cache.MaxEntries <= 0 {
		return fmt.Errorf("cache max_entries must be positive")
	}

	if c.API.Port <= 0 || c.API.Port > 65535 {
		return fmt.Errorf("api port must be between 1 and 65535")
	}

	// Validate policies
	for i, policy := range c.Strategy.Policies {
		if err := policy.Validate(); err != nil {
			return fmt.Errorf("policy %d invalid: %w", i, err)
		}
	}

	return nil
}

// Validate validates a retry policy
func (p *RetryPolicy) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if p.MaxAttempts <= 0 {
		return fmt.Errorf("max_attempts must be positive")
	}

	if p.BaseDelayMs < 0 {
		return fmt.Errorf("base_delay_ms must be non-negative")
	}

	if p.MaxDelayMs < p.BaseDelayMs {
		return fmt.Errorf("max_delay_ms must be >= base_delay_ms")
	}

	if p.BackoffMultiplier <= 0 {
		return fmt.Errorf("backoff_multiplier must be positive")
	}

	if p.JitterPercent < 0 || p.JitterPercent > 100 {
		return fmt.Errorf("jitter_percent must be between 0 and 100")
	}

	return nil
}

// Merge merges another config into this one, with the other config taking precedence
func (c *Config) Merge(other *Config) {
	if other.Enabled != c.Enabled {
		c.Enabled = other.Enabled
	}

	if other.RedisAddr != "" {
		c.RedisAddr = other.RedisAddr
	}

	if other.RedisPassword != "" {
		c.RedisPassword = other.RedisPassword
	}

	if other.RedisDB != 0 {
		c.RedisDB = other.RedisDB
	}

	// Merge strategy
	c.Strategy.Merge(&other.Strategy)

	// Merge data collection
	c.DataCollection.Merge(&other.DataCollection)

	// Merge cache
	c.Cache.Merge(&other.Cache)

	// Merge API
	c.API.Merge(&other.API)
}

// Merge merges another retry strategy into this one
func (s *RetryStrategy) Merge(other *RetryStrategy) {
	if other.Name != "" {
		s.Name = other.Name
	}

	if other.Enabled != s.Enabled {
		s.Enabled = other.Enabled
	}

	if len(other.Policies) > 0 {
		s.Policies = other.Policies
	}

	if other.BayesianThreshold != 0 {
		s.BayesianThreshold = other.BayesianThreshold
	}

	if other.MLEnabled != s.MLEnabled {
		s.MLEnabled = other.MLEnabled
	}

	if other.MLModel != nil {
		s.MLModel = other.MLModel
	}

	s.Guardrails.Merge(&other.Guardrails)
	s.DataCollection.Merge(&other.DataCollection)
}

// Merge merges another guardrails config into this one
func (g *PolicyGuardrails) Merge(other *PolicyGuardrails) {
	if other.MaxAttempts != 0 {
		g.MaxAttempts = other.MaxAttempts
	}

	if other.MaxDelayMs != 0 {
		g.MaxDelayMs = other.MaxDelayMs
	}

	if other.MaxBudgetPercent != 0 {
		g.MaxBudgetPercent = other.MaxBudgetPercent
	}

	if other.PerTenantLimits != g.PerTenantLimits {
		g.PerTenantLimits = other.PerTenantLimits
	}

	if other.EmergencyStop != g.EmergencyStop {
		g.EmergencyStop = other.EmergencyStop
	}

	if other.ExplainabilityReq != g.ExplainabilityReq {
		g.ExplainabilityReq = other.ExplainabilityReq
	}
}

// Merge merges another data collection config into this one
func (d *DataCollectionConfig) Merge(other *DataCollectionConfig) {
	if other.Enabled != d.Enabled {
		d.Enabled = other.Enabled
	}

	if other.SampleRate != 0 {
		d.SampleRate = other.SampleRate
	}

	if other.RetentionDays != 0 {
		d.RetentionDays = other.RetentionDays
	}

	if other.AggregationInterval != 0 {
		d.AggregationInterval = other.AggregationInterval
	}

	if other.FeatureExtraction != d.FeatureExtraction {
		d.FeatureExtraction = other.FeatureExtraction
	}
}

// Merge merges another cache config into this one
func (c *CacheConfig) Merge(other *CacheConfig) {
	if other.Enabled != c.Enabled {
		c.Enabled = other.Enabled
	}

	if other.TTL != 0 {
		c.TTL = other.TTL
	}

	if other.MaxEntries != 0 {
		c.MaxEntries = other.MaxEntries
	}
}

// Merge merges another API config into this one
func (a *APIConfig) Merge(other *APIConfig) {
	if other.Enabled != a.Enabled {
		a.Enabled = other.Enabled
	}

	if other.Port != 0 {
		a.Port = other.Port
	}

	if other.Path != "" {
		a.Path = other.Path
	}
}

// ToJSON converts the config to JSON string
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON loads config from JSON string
func (c *Config) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), c)
}

// Clone creates a deep copy of the config
func (c *Config) Clone() *Config {
	data, _ := json.Marshal(c)
	var clone Config
	json.Unmarshal(data, &clone)
	return &clone
}