package canary_deployments

import (
	"fmt"
	"time"
)

// Config holds the overall configuration for the canary deployment system
type Config struct {
	// Redis configuration
	RedisAddr     string `json:"redis_addr" yaml:"redis_addr"`
	RedisPassword string `json:"redis_password" yaml:"redis_password"`
	RedisDB       int    `json:"redis_db" yaml:"redis_db"`

	// Deployment defaults
	DefaultConfig CanaryConfig `json:"default_config" yaml:"default_config"`

	// Monitoring configuration
	MetricsInterval    time.Duration `json:"metrics_interval" yaml:"metrics_interval"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	WorkerTimeout      time.Duration `json:"worker_timeout" yaml:"worker_timeout"`

	// Performance tuning
	MaxConcurrentDeployments int           `json:"max_concurrent_deployments" yaml:"max_concurrent_deployments"`
	MetricsRetention        time.Duration `json:"metrics_retention" yaml:"metrics_retention"`
	EventRetention          time.Duration `json:"event_retention" yaml:"event_retention"`

	// Safety limits
	MaxCanaryPercentage     int           `json:"max_canary_percentage" yaml:"max_canary_percentage"`
	MinMetricsSamples       int           `json:"min_metrics_samples" yaml:"min_metrics_samples"`
	EmergencyRollbackDelay  time.Duration `json:"emergency_rollback_delay" yaml:"emergency_rollback_delay"`

	// Alerting
	WebhookURLs      []string      `json:"webhook_urls" yaml:"webhook_urls"`
	AlertCooldown    time.Duration `json:"alert_cooldown" yaml:"alert_cooldown"`
	SlackChannel     string        `json:"slack_channel" yaml:"slack_channel"`
	PagerDutyKey     string        `json:"pagerduty_key" yaml:"pagerduty_key"`

	// TUI configuration
	EnableTUI        bool   `json:"enable_tui" yaml:"enable_tui"`
	TUIUpdateInterval time.Duration `json:"tui_update_interval" yaml:"tui_update_interval"`

	// API configuration
	EnableAPI        bool   `json:"enable_api" yaml:"enable_api"`
	APIListenAddr    string `json:"api_listen_addr" yaml:"api_listen_addr"`
	APIAuthToken     string `json:"api_auth_token" yaml:"api_auth_token"`
}

// Validate checks the configuration for common errors
func (c *Config) Validate() error {
	if c.RedisAddr == "" {
		return fmt.Errorf("redis_addr is required")
	}

	if c.MetricsInterval < time.Second {
		return fmt.Errorf("metrics_interval must be at least 1 second")
	}

	if c.HealthCheckInterval < time.Second {
		return fmt.Errorf("health_check_interval must be at least 1 second")
	}

	if c.WorkerTimeout < 10*time.Second {
		return fmt.Errorf("worker_timeout must be at least 10 seconds")
	}

	if c.MaxConcurrentDeployments <= 0 {
		return fmt.Errorf("max_concurrent_deployments must be positive")
	}

	if c.MaxCanaryPercentage < 1 || c.MaxCanaryPercentage > 100 {
		return fmt.Errorf("max_canary_percentage must be between 1 and 100")
	}

	if c.MinMetricsSamples < 1 {
		return fmt.Errorf("min_metrics_samples must be at least 1")
	}

	if c.EnableAPI && c.APIListenAddr == "" {
		return fmt.Errorf("api_listen_addr is required when API is enabled")
	}

	return c.DefaultConfig.Validate()
}

// SetDefaults sets reasonable default values for the configuration
func (c *Config) SetDefaults() {
	if c.RedisAddr == "" {
		c.RedisAddr = "localhost:6379"
	}

	if c.MetricsInterval == 0 {
		c.MetricsInterval = 30 * time.Second
	}

	if c.HealthCheckInterval == 0 {
		c.HealthCheckInterval = 15 * time.Second
	}

	if c.WorkerTimeout == 0 {
		c.WorkerTimeout = 2 * time.Minute
	}

	if c.MaxConcurrentDeployments == 0 {
		c.MaxConcurrentDeployments = 10
	}

	if c.MetricsRetention == 0 {
		c.MetricsRetention = 24 * time.Hour
	}

	if c.EventRetention == 0 {
		c.EventRetention = 7 * 24 * time.Hour
	}

	if c.MaxCanaryPercentage == 0 {
		c.MaxCanaryPercentage = 50
	}

	if c.MinMetricsSamples == 0 {
		c.MinMetricsSamples = 10
	}

	if c.EmergencyRollbackDelay == 0 {
		c.EmergencyRollbackDelay = 30 * time.Second
	}

	if c.AlertCooldown == 0 {
		c.AlertCooldown = 5 * time.Minute
	}

	if c.TUIUpdateInterval == 0 {
		c.TUIUpdateInterval = 2 * time.Second
	}

	if c.APIListenAddr == "" {
		c.APIListenAddr = ":8080"
	}

	c.DefaultConfig.SetDefaults()
}

// Validate checks the canary configuration for errors
func (cc *CanaryConfig) Validate() error {
	if cc.MaxCanaryDuration <= 0 {
		return fmt.Errorf("max_canary_duration must be positive")
	}

	if cc.MinCanaryDuration <= 0 {
		return fmt.Errorf("min_canary_duration must be positive")
	}

	if cc.MinCanaryDuration >= cc.MaxCanaryDuration {
		return fmt.Errorf("min_canary_duration must be less than max_canary_duration")
	}

	if cc.DrainTimeout <= 0 {
		return fmt.Errorf("drain_timeout must be positive")
	}

	if cc.MetricsWindow <= 0 {
		return fmt.Errorf("metrics_window must be positive")
	}

	// Validate routing strategy
	switch cc.RoutingStrategy {
	case SplitQueueStrategy, StreamGroupStrategy, HashRingStrategy:
		// Valid strategies
	default:
		return fmt.Errorf("invalid routing_strategy: %s", cc.RoutingStrategy)
	}

	// Validate promotion stages
	for i, stage := range cc.PromotionStages {
		if stage.Percentage < 0 || stage.Percentage > 100 {
			return fmt.Errorf("promotion_stages[%d].percentage must be between 0 and 100", i)
		}
		if stage.Duration <= 0 {
			return fmt.Errorf("promotion_stages[%d].duration must be positive", i)
		}
		if err := stage.Conditions.Validate(); err != nil {
			return fmt.Errorf("promotion_stages[%d].conditions: %w", i, err)
		}
	}

	// Validate thresholds
	if err := cc.RollbackThresholds.Validate(); err != nil {
		return fmt.Errorf("rollback_thresholds: %w", err)
	}

	return nil
}

// SetDefaults sets reasonable default values for canary configuration
func (cc *CanaryConfig) SetDefaults() {
	if cc.RoutingStrategy == "" {
		cc.RoutingStrategy = SplitQueueStrategy
	}

	if cc.MaxCanaryDuration == 0 {
		cc.MaxCanaryDuration = 2 * time.Hour
	}

	if cc.MinCanaryDuration == 0 {
		cc.MinCanaryDuration = 5 * time.Minute
	}

	if cc.DrainTimeout == 0 {
		cc.DrainTimeout = 5 * time.Minute
	}

	if cc.MetricsWindow == 0 {
		cc.MetricsWindow = 5 * time.Minute
	}

	// Set default promotion stages if none specified
	if len(cc.PromotionStages) == 0 && cc.AutoPromotion {
		cc.PromotionStages = []PromotionStage{
			{
				Percentage:  5,
				Duration:    10 * time.Minute,
				AutoPromote: true,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  2.0,
					MaxLatencyIncrease:    20.0,
					MaxThroughputDecrease: 10.0,
					MinSuccessRate:        98.0,
					RequiredSampleSize:    50,
				},
			},
			{
				Percentage:  20,
				Duration:    15 * time.Minute,
				AutoPromote: true,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  1.5,
					MaxLatencyIncrease:    15.0,
					MaxThroughputDecrease: 5.0,
					MinSuccessRate:        99.0,
					RequiredSampleSize:    100,
				},
			},
			{
				Percentage:  50,
				Duration:    20 * time.Minute,
				AutoPromote: false, // Manual approval for 50%+
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  1.0,
					MaxLatencyIncrease:    10.0,
					MaxThroughputDecrease: 3.0,
					MinSuccessRate:        99.5,
					RequiredSampleSize:    200,
				},
			},
		}
	}

	cc.RollbackThresholds.SetDefaults()
}

// Validate checks SLO thresholds for reasonable values
func (st *SLOThresholds) Validate() error {
	if st.MaxErrorRateIncrease < 0 {
		return fmt.Errorf("max_error_rate_increase cannot be negative")
	}

	if st.MaxLatencyIncrease < 0 {
		return fmt.Errorf("max_latency_increase cannot be negative")
	}

	if st.MaxThroughputDecrease < 0 {
		return fmt.Errorf("max_throughput_decrease cannot be negative")
	}

	if st.MinSuccessRate < 0 || st.MinSuccessRate > 100 {
		return fmt.Errorf("min_success_rate must be between 0 and 100")
	}

	if st.RequiredSampleSize < 1 {
		return fmt.Errorf("required_sample_size must be at least 1")
	}

	if st.MaxMemoryIncrease < 0 {
		return fmt.Errorf("max_memory_increase cannot be negative")
	}

	return nil
}

// SetDefaults sets reasonable default values for SLO thresholds
func (st *SLOThresholds) SetDefaults() {
	if st.MaxErrorRateIncrease == 0 {
		st.MaxErrorRateIncrease = 5.0 // 5 percentage points
	}

	if st.MaxLatencyIncrease == 0 {
		st.MaxLatencyIncrease = 50.0 // 50% increase
	}

	if st.MaxThroughputDecrease == 0 {
		st.MaxThroughputDecrease = 20.0 // 20% decrease
	}

	if st.MinSuccessRate == 0 {
		st.MinSuccessRate = 95.0 // 95% success rate
	}

	if st.RequiredSampleSize == 0 {
		st.RequiredSampleSize = 20 // 20 jobs minimum
	}

	if st.MaxMemoryIncrease == 0 {
		st.MaxMemoryIncrease = 100.0 // 100% memory increase
	}
}

// DefaultCanaryConfig returns a configuration suitable for most use cases
func DefaultCanaryConfig() *CanaryConfig {
	config := &CanaryConfig{
		RoutingStrategy:   SplitQueueStrategy,
		StickyRouting:     true,
		AutoPromotion:     false, // Conservative default
		MaxCanaryDuration: 2 * time.Hour,
		MinCanaryDuration: 5 * time.Minute,
		DrainTimeout:      5 * time.Minute,
		MetricsWindow:     5 * time.Minute,
		RollbackThresholds: SLOThresholds{
			MaxErrorRateIncrease:  10.0, // 10 percentage points
			MaxLatencyIncrease:    100.0, // 100% increase
			MaxThroughputDecrease: 50.0,  // 50% decrease
			MinSuccessRate:        80.0,  // 80% success rate
			RequiredSampleSize:    10,    // 10 jobs minimum
			MaxMemoryIncrease:     200.0, // 200% memory increase
		},
	}

	config.SetDefaults()
	return config
}

// ConservativeCanaryConfig returns a very safe configuration for critical systems
func ConservativeCanaryConfig() *CanaryConfig {
	config := &CanaryConfig{
		RoutingStrategy:   SplitQueueStrategy,
		StickyRouting:     true,
		AutoPromotion:     false,
		MaxCanaryDuration: 4 * time.Hour,
		MinCanaryDuration: 15 * time.Minute,
		DrainTimeout:      10 * time.Minute,
		MetricsWindow:     10 * time.Minute,
		PromotionStages: []PromotionStage{
			{
				Percentage:  2,
				Duration:    20 * time.Minute,
				AutoPromote: false,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  0.5,
					MaxLatencyIncrease:    5.0,
					MaxThroughputDecrease: 2.0,
					MinSuccessRate:        99.5,
					RequiredSampleSize:    100,
				},
			},
			{
				Percentage:  5,
				Duration:    30 * time.Minute,
				AutoPromote: false,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  0.3,
					MaxLatencyIncrease:    3.0,
					MaxThroughputDecrease: 1.0,
					MinSuccessRate:        99.7,
					RequiredSampleSize:    200,
				},
			},
		},
		RollbackThresholds: SLOThresholds{
			MaxErrorRateIncrease:  2.0,
			MaxLatencyIncrease:    20.0,
			MaxThroughputDecrease: 10.0,
			MinSuccessRate:        95.0,
			RequiredSampleSize:    20,
			MaxMemoryIncrease:     50.0,
		},
	}

	return config
}

// AggressiveCanaryConfig returns a configuration for fast iteration
func AggressiveCanaryConfig() *CanaryConfig {
	config := &CanaryConfig{
		RoutingStrategy:   SplitQueueStrategy,
		StickyRouting:     false, // Allow more random distribution
		AutoPromotion:     true,
		MaxCanaryDuration: 30 * time.Minute,
		MinCanaryDuration: 2 * time.Minute,
		DrainTimeout:      2 * time.Minute,
		MetricsWindow:     2 * time.Minute,
		PromotionStages: []PromotionStage{
			{
				Percentage:  10,
				Duration:    5 * time.Minute,
				AutoPromote: true,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  5.0,
					MaxLatencyIncrease:    50.0,
					MaxThroughputDecrease: 20.0,
					MinSuccessRate:        90.0,
					RequiredSampleSize:    20,
				},
			},
			{
				Percentage:  50,
				Duration:    10 * time.Minute,
				AutoPromote: true,
				Conditions: SLOThresholds{
					MaxErrorRateIncrease:  3.0,
					MaxLatencyIncrease:    30.0,
					MaxThroughputDecrease: 15.0,
					MinSuccessRate:        95.0,
					RequiredSampleSize:    50,
				},
			},
		},
		RollbackThresholds: SLOThresholds{
			MaxErrorRateIncrease:  20.0,
			MaxLatencyIncrease:    200.0,
			MaxThroughputDecrease: 75.0,
			MinSuccessRate:        70.0,
			RequiredSampleSize:    5,
			MaxMemoryIncrease:     500.0,
		},
	}

	return config
}

// GetConfigByProfile returns a pre-configured canary config based on a profile name
func GetConfigByProfile(profile string) *CanaryConfig {
	switch profile {
	case "conservative":
		return ConservativeCanaryConfig()
	case "aggressive":
		return AggressiveCanaryConfig()
	case "default":
		fallthrough
	default:
		return DefaultCanaryConfig()
	}
}