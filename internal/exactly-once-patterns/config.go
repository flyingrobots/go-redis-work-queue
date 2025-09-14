// Copyright 2025 James Ross
package exactlyonce

import (
	"time"
)

// Config holds configuration for exactly-once patterns
type Config struct {
	Idempotency IdempotencyConfig `mapstructure:"idempotency"`
	Outbox      OutboxConfig      `mapstructure:"outbox"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
}

// IdempotencyConfig configures idempotency key handling
type IdempotencyConfig struct {
	// Enabled controls whether idempotency checking is active
	Enabled bool `mapstructure:"enabled"`

	// DefaultTTL is the default time-to-live for idempotency keys
	DefaultTTL time.Duration `mapstructure:"default_ttl"`

	// KeyPrefix is prepended to all idempotency keys in storage
	KeyPrefix string `mapstructure:"key_prefix"`

	// Storage configures the backing storage for idempotency keys
	Storage StorageConfig `mapstructure:"storage"`

	// MaxRetries for storage operations
	MaxRetries int `mapstructure:"max_retries"`

	// RetryDelay between storage operation retries
	RetryDelay time.Duration `mapstructure:"retry_delay"`

	// CleanupInterval for expired keys cleanup
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`

	// BatchSize for bulk operations
	BatchSize int `mapstructure:"batch_size"`
}

// StorageConfig configures the storage backend for idempotency keys
type StorageConfig struct {
	// Type specifies the storage backend ("redis", "memory", "database")
	Type string `mapstructure:"type"`

	// Redis-specific configuration
	Redis RedisStorageConfig `mapstructure:"redis"`

	// Database-specific configuration (for outbox pattern)
	Database DatabaseStorageConfig `mapstructure:"database"`

	// Memory-specific configuration (for testing/development)
	Memory MemoryStorageConfig `mapstructure:"memory"`
}

// RedisStorageConfig configures Redis storage for idempotency keys
type RedisStorageConfig struct {
	// KeyPattern for storing idempotency keys (supports templating)
	KeyPattern string `mapstructure:"key_pattern"`

	// HashKeyPattern for hash-based storage
	HashKeyPattern string `mapstructure:"hash_key_pattern"`

	// UseHashes determines whether to use Redis hashes vs individual keys
	UseHashes bool `mapstructure:"use_hashes"`

	// Compression enables value compression for large payloads
	Compression bool `mapstructure:"compression"`
}

// DatabaseStorageConfig configures database storage for outbox pattern
type DatabaseStorageConfig struct {
	// TableName for storing outbox events
	TableName string `mapstructure:"table_name"`

	// IdempotencyTableName for storing idempotency keys
	IdempotencyTableName string `mapstructure:"idempotency_table_name"`

	// BatchSize for database operations
	BatchSize int `mapstructure:"batch_size"`

	// MaxConnections for the database connection pool
	MaxConnections int `mapstructure:"max_connections"`

	// TransactionTimeout for outbox operations
	TransactionTimeout time.Duration `mapstructure:"transaction_timeout"`
}

// MemoryStorageConfig configures in-memory storage (for testing)
type MemoryStorageConfig struct {
	// MaxKeys limits the number of keys stored in memory
	MaxKeys int `mapstructure:"max_keys"`

	// EvictionPolicy determines how keys are evicted ("lru", "fifo")
	EvictionPolicy string `mapstructure:"eviction_policy"`
}

// OutboxConfig configures the transactional outbox pattern
type OutboxConfig struct {
	// Enabled controls whether outbox pattern is active
	Enabled bool `mapstructure:"enabled"`

	// PollInterval for checking unpublished events
	PollInterval time.Duration `mapstructure:"poll_interval"`

	// BatchSize for processing outbox events
	BatchSize int `mapstructure:"batch_size"`

	// MaxRetries for publishing events
	MaxRetries int `mapstructure:"max_retries"`

	// RetryBackoff configuration
	RetryBackoff BackoffConfig `mapstructure:"retry_backoff"`

	// CleanupInterval for removing old processed events
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`

	// CleanupAfter determines how long to keep processed events
	CleanupAfter time.Duration `mapstructure:"cleanup_after"`

	// Publishers configures event publishing destinations
	Publishers []PublisherConfig `mapstructure:"publishers"`
}

// BackoffConfig configures retry backoff behavior
type BackoffConfig struct {
	// InitialDelay is the initial retry delay
	InitialDelay time.Duration `mapstructure:"initial_delay"`

	// MaxDelay is the maximum retry delay
	MaxDelay time.Duration `mapstructure:"max_delay"`

	// Multiplier for exponential backoff
	Multiplier float64 `mapstructure:"multiplier"`

	// Jitter adds randomness to backoff timing
	Jitter bool `mapstructure:"jitter"`
}

// PublisherConfig configures an outbox event publisher
type PublisherConfig struct {
	// Name identifies the publisher
	Name string `mapstructure:"name"`

	// Type specifies the publisher type ("redis", "kafka", "sqs", etc.)
	Type string `mapstructure:"type"`

	// Target specifies the destination (queue name, topic, etc.)
	Target string `mapstructure:"target"`

	// Config contains publisher-specific configuration
	Config map[string]interface{} `mapstructure:"config"`

	// Enabled controls whether this publisher is active
	Enabled bool `mapstructure:"enabled"`
}

// MetricsConfig configures metrics collection for exactly-once patterns
type MetricsConfig struct {
	// Enabled controls whether metrics collection is active
	Enabled bool `mapstructure:"enabled"`

	// CollectionInterval for gathering statistics
	CollectionInterval time.Duration `mapstructure:"collection_interval"`

	// HistogramBuckets for latency metrics
	HistogramBuckets []float64 `mapstructure:"histogram_buckets"`

	// CardinaltityLimit prevents metric explosion
	CardinalityLimit int `mapstructure:"cardinality_limit"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Idempotency: IdempotencyConfig{
			Enabled:         true,
			DefaultTTL:      24 * time.Hour,
			KeyPrefix:       "idempotency:",
			MaxRetries:      3,
			RetryDelay:      100 * time.Millisecond,
			CleanupInterval: 1 * time.Hour,
			BatchSize:       100,
			Storage: StorageConfig{
				Type: "redis",
				Redis: RedisStorageConfig{
					KeyPattern:     "{queue}:idempotency:{tenant}:{key}",
					HashKeyPattern: "{queue}:idempotency:{tenant}",
					UseHashes:      true,
					Compression:    false,
				},
			},
		},
		Outbox: OutboxConfig{
			Enabled:         false, // Disabled by default as it requires database setup
			PollInterval:    5 * time.Second,
			BatchSize:       50,
			MaxRetries:      5,
			CleanupInterval: 24 * time.Hour,
			CleanupAfter:    7 * 24 * time.Hour, // Keep events for 7 days
			RetryBackoff: BackoffConfig{
				InitialDelay: 1 * time.Second,
				MaxDelay:     30 * time.Minute,
				Multiplier:   2.0,
				Jitter:       true,
			},
			Publishers: []PublisherConfig{},
		},
		Metrics: MetricsConfig{
			Enabled:            true,
			CollectionInterval: 30 * time.Second,
			HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
			CardinalityLimit:   1000,
		},
	}
}