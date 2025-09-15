// Copyright 2025 James Ross
package dlqremediation

import (
	"time"
)

// Config holds configuration for the DLQ remediation pipeline
type Config struct {
	// Pipeline settings
	Pipeline PipelineConfig `json:"pipeline" yaml:"pipeline" toml:"pipeline"`

	// Redis connection settings
	Redis RedisConfig `json:"redis" yaml:"redis" toml:"redis"`

	// Storage settings
	Storage StorageConfig `json:"storage" yaml:"storage" toml:"storage"`

	// Logging settings
	Logging LoggingConfig `json:"logging" yaml:"logging" toml:"logging"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Addr         string        `json:"addr" yaml:"addr" toml:"addr"`
	Password     string        `json:"password" yaml:"password" toml:"password"`
	DB           int           `json:"db" yaml:"db" toml:"db"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries" toml:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout" toml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout" toml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" toml:"write_timeout"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size" toml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns" toml:"min_idle_conns"`
	MaxConnAge   time.Duration `json:"max_conn_age" yaml:"max_conn_age" toml:"max_conn_age"`
	PoolTimeout  time.Duration `json:"pool_timeout" yaml:"pool_timeout" toml:"pool_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout" toml:"idle_timeout"`
}

// StorageConfig contains storage-related configuration
type StorageConfig struct {
	RulesKey            string `json:"rules_key" yaml:"rules_key" toml:"rules_key"`
	AuditLogKey         string `json:"audit_log_key" yaml:"audit_log_key" toml:"audit_log_key"`
	MetricsKey          string `json:"metrics_key" yaml:"metrics_key" toml:"metrics_key"`
	StateKey            string `json:"state_key" yaml:"state_key" toml:"state_key"`
	ClassificationCache string `json:"classification_cache" yaml:"classification_cache" toml:"classification_cache"`
	IdempotencyKey      string `json:"idempotency_key" yaml:"idempotency_key" toml:"idempotency_key"`
	DLQStreamKey        string `json:"dlq_stream_key" yaml:"dlq_stream_key" toml:"dlq_stream_key"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level" toml:"level"`
	Format     string `json:"format" yaml:"format" toml:"format"`
	Output     string `json:"output" yaml:"output" toml:"output"`
	Structured bool   `json:"structured" yaml:"structured" toml:"structured"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Pipeline: PipelineConfig{
			Enabled:            true,
			PollInterval:       30 * time.Second,
			BatchSize:          50,
			MaxConcurrentRules: 10,
			DryRun:             false,
			RedisStreamKey:     "dlq:stream",
			MetricsEnabled:     true,
			AuditEnabled:       true,
			ExternalClassifier: ExternalClassifier{
				Enabled:    false,
				Timeout:    5 * time.Second,
				RetryCount: 3,
				CacheTTL:   5 * time.Minute,
			},
			GlobalSafetyLimits: SafetyLimits{
				MaxPerMinute:       1000,
				MaxTotalPerRun:     5000,
				ErrorRateThreshold: 0.1,
				BackoffOnFailure:   true,
			},
			RetentionPolicy: RetentionPolicy{
				AuditLogTTL:       7 * 24 * time.Hour,
				MetricsTTL:        30 * 24 * time.Hour,
				ClassificationTTL: 1 * time.Hour,
				ProcessedJobsTTL:  24 * time.Hour,
			},
		},
		Redis: RedisConfig{
			Addr:         "localhost:6379",
			DB:           0,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
			MaxConnAge:   1 * time.Hour,
			PoolTimeout:  30 * time.Second,
			IdleTimeout:  5 * time.Minute,
		},
		Storage: StorageConfig{
			RulesKey:            "dlq:remediation:rules",
			AuditLogKey:         "dlq:remediation:audit",
			MetricsKey:          "dlq:remediation:metrics",
			StateKey:            "dlq:remediation:state",
			ClassificationCache: "dlq:remediation:classification_cache",
			IdempotencyKey:      "dlq:remediation:processed",
			DLQStreamKey:        "dlq:stream",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			Structured: true,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []ValidationError

	// Validate pipeline config
	if c.Pipeline.BatchSize <= 0 {
		errors = append(errors, ValidationError{
			Field:   "pipeline.batch_size",
			Message: "must be greater than 0",
			Value:   c.Pipeline.BatchSize,
		})
	}

	if c.Pipeline.BatchSize > 1000 {
		errors = append(errors, ValidationError{
			Field:   "pipeline.batch_size",
			Message: "must be less than or equal to 1000",
			Value:   c.Pipeline.BatchSize,
		})
	}

	if c.Pipeline.PollInterval < time.Second {
		errors = append(errors, ValidationError{
			Field:   "pipeline.poll_interval",
			Message: "must be at least 1 second",
			Value:   c.Pipeline.PollInterval,
		})
	}

	if c.Pipeline.MaxConcurrentRules <= 0 {
		errors = append(errors, ValidationError{
			Field:   "pipeline.max_concurrent_rules",
			Message: "must be greater than 0",
			Value:   c.Pipeline.MaxConcurrentRules,
		})
	}

	// Validate safety limits
	if c.Pipeline.GlobalSafetyLimits.ErrorRateThreshold < 0 || c.Pipeline.GlobalSafetyLimits.ErrorRateThreshold > 1 {
		errors = append(errors, ValidationError{
			Field:   "pipeline.global_safety_limits.error_rate_threshold",
			Message: "must be between 0 and 1",
			Value:   c.Pipeline.GlobalSafetyLimits.ErrorRateThreshold,
		})
	}

	// Validate external classifier
	if c.Pipeline.ExternalClassifier.Enabled {
		if c.Pipeline.ExternalClassifier.Endpoint == "" {
			errors = append(errors, ValidationError{
				Field:   "pipeline.external_classifier.endpoint",
				Message: "endpoint is required when external classifier is enabled",
			})
		}

		if c.Pipeline.ExternalClassifier.Timeout <= 0 {
			errors = append(errors, ValidationError{
				Field:   "pipeline.external_classifier.timeout",
				Message: "timeout must be greater than 0",
				Value:   c.Pipeline.ExternalClassifier.Timeout,
			})
		}
	}

	// Validate Redis config
	if c.Redis.Addr == "" {
		errors = append(errors, ValidationError{
			Field:   "redis.addr",
			Message: "address is required",
		})
	}

	if c.Redis.DB < 0 {
		errors = append(errors, ValidationError{
			Field:   "redis.db",
			Message: "database number must be non-negative",
			Value:   c.Redis.DB,
		})
	}

	// Validate storage config
	if c.Storage.RulesKey == "" {
		errors = append(errors, ValidationError{
			Field:   "storage.rules_key",
			Message: "rules key is required",
		})
	}

	if c.Storage.DLQStreamKey == "" {
		errors = append(errors, ValidationError{
			Field:   "storage.dlq_stream_key",
			Message: "DLQ stream key is required",
		})
	}

	// Validate retention policy
	if c.Pipeline.RetentionPolicy.AuditLogTTL <= 0 {
		errors = append(errors, ValidationError{
			Field:   "pipeline.retention_policy.audit_log_ttl",
			Message: "audit log TTL must be greater than 0",
			Value:   c.Pipeline.RetentionPolicy.AuditLogTTL,
		})
	}

	if len(errors) > 0 {
		return MultiValidationError{Errors: errors}
	}

	return nil
}

// ApplyDefaults applies default values for any missing configuration
func (c *Config) ApplyDefaults() {
	defaults := DefaultConfig()

	// Apply pipeline defaults
	if c.Pipeline.PollInterval == 0 {
		c.Pipeline.PollInterval = defaults.Pipeline.PollInterval
	}
	if c.Pipeline.BatchSize == 0 {
		c.Pipeline.BatchSize = defaults.Pipeline.BatchSize
	}
	if c.Pipeline.MaxConcurrentRules == 0 {
		c.Pipeline.MaxConcurrentRules = defaults.Pipeline.MaxConcurrentRules
	}
	if c.Pipeline.RedisStreamKey == "" {
		c.Pipeline.RedisStreamKey = defaults.Pipeline.RedisStreamKey
	}

	// Apply Redis defaults
	if c.Redis.Addr == "" {
		c.Redis.Addr = defaults.Redis.Addr
	}
	if c.Redis.MaxRetries == 0 {
		c.Redis.MaxRetries = defaults.Redis.MaxRetries
	}
	if c.Redis.DialTimeout == 0 {
		c.Redis.DialTimeout = defaults.Redis.DialTimeout
	}
	if c.Redis.ReadTimeout == 0 {
		c.Redis.ReadTimeout = defaults.Redis.ReadTimeout
	}
	if c.Redis.WriteTimeout == 0 {
		c.Redis.WriteTimeout = defaults.Redis.WriteTimeout
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = defaults.Redis.PoolSize
	}

	// Apply storage defaults
	if c.Storage.RulesKey == "" {
		c.Storage.RulesKey = defaults.Storage.RulesKey
	}
	if c.Storage.AuditLogKey == "" {
		c.Storage.AuditLogKey = defaults.Storage.AuditLogKey
	}
	if c.Storage.MetricsKey == "" {
		c.Storage.MetricsKey = defaults.Storage.MetricsKey
	}
	if c.Storage.StateKey == "" {
		c.Storage.StateKey = defaults.Storage.StateKey
	}
	if c.Storage.ClassificationCache == "" {
		c.Storage.ClassificationCache = defaults.Storage.ClassificationCache
	}
	if c.Storage.IdempotencyKey == "" {
		c.Storage.IdempotencyKey = defaults.Storage.IdempotencyKey
	}
	if c.Storage.DLQStreamKey == "" {
		c.Storage.DLQStreamKey = defaults.Storage.DLQStreamKey
	}

	// Apply logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = defaults.Logging.Level
	}
	if c.Logging.Format == "" {
		c.Logging.Format = defaults.Logging.Format
	}
	if c.Logging.Output == "" {
		c.Logging.Output = defaults.Logging.Output
	}

	// Apply external classifier defaults
	if c.Pipeline.ExternalClassifier.Enabled && c.Pipeline.ExternalClassifier.Timeout == 0 {
		c.Pipeline.ExternalClassifier.Timeout = defaults.Pipeline.ExternalClassifier.Timeout
	}
	if c.Pipeline.ExternalClassifier.Enabled && c.Pipeline.ExternalClassifier.RetryCount == 0 {
		c.Pipeline.ExternalClassifier.RetryCount = defaults.Pipeline.ExternalClassifier.RetryCount
	}
	if c.Pipeline.ExternalClassifier.Enabled && c.Pipeline.ExternalClassifier.CacheTTL == 0 {
		c.Pipeline.ExternalClassifier.CacheTTL = defaults.Pipeline.ExternalClassifier.CacheTTL
	}

	// Apply safety limit defaults
	if c.Pipeline.GlobalSafetyLimits.MaxPerMinute == 0 {
		c.Pipeline.GlobalSafetyLimits.MaxPerMinute = defaults.Pipeline.GlobalSafetyLimits.MaxPerMinute
	}
	if c.Pipeline.GlobalSafetyLimits.MaxTotalPerRun == 0 {
		c.Pipeline.GlobalSafetyLimits.MaxTotalPerRun = defaults.Pipeline.GlobalSafetyLimits.MaxTotalPerRun
	}
	if c.Pipeline.GlobalSafetyLimits.ErrorRateThreshold == 0 {
		c.Pipeline.GlobalSafetyLimits.ErrorRateThreshold = defaults.Pipeline.GlobalSafetyLimits.ErrorRateThreshold
	}

	// Apply retention policy defaults
	if c.Pipeline.RetentionPolicy.AuditLogTTL == 0 {
		c.Pipeline.RetentionPolicy.AuditLogTTL = defaults.Pipeline.RetentionPolicy.AuditLogTTL
	}
	if c.Pipeline.RetentionPolicy.MetricsTTL == 0 {
		c.Pipeline.RetentionPolicy.MetricsTTL = defaults.Pipeline.RetentionPolicy.MetricsTTL
	}
	if c.Pipeline.RetentionPolicy.ClassificationTTL == 0 {
		c.Pipeline.RetentionPolicy.ClassificationTTL = defaults.Pipeline.RetentionPolicy.ClassificationTTL
	}
	if c.Pipeline.RetentionPolicy.ProcessedJobsTTL == 0 {
		c.Pipeline.RetentionPolicy.ProcessedJobsTTL = defaults.Pipeline.RetentionPolicy.ProcessedJobsTTL
	}
}