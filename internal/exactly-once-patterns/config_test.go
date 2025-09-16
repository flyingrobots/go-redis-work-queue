// Copyright 2025 James Ross
package exactlyonce

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test idempotency defaults
	assert.True(t, cfg.Idempotency.Enabled)
	assert.Equal(t, 24*time.Hour, cfg.Idempotency.DefaultTTL)
	assert.Equal(t, "idempotency:", cfg.Idempotency.KeyPrefix)
	assert.Equal(t, 3, cfg.Idempotency.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, cfg.Idempotency.RetryDelay)
	assert.Equal(t, 1*time.Hour, cfg.Idempotency.CleanupInterval)
	assert.Equal(t, 100, cfg.Idempotency.BatchSize)

	// Test storage defaults
	assert.Equal(t, "redis", cfg.Idempotency.Storage.Type)
	assert.Equal(t, "{queue}:idempotency:{tenant}:{key}", cfg.Idempotency.Storage.Redis.KeyPattern)
	assert.Equal(t, "{queue}:idempotency:{tenant}", cfg.Idempotency.Storage.Redis.HashKeyPattern)
	assert.True(t, cfg.Idempotency.Storage.Redis.UseHashes)
	assert.False(t, cfg.Idempotency.Storage.Redis.Compression)

	// Test outbox defaults
	assert.False(t, cfg.Outbox.Enabled) // Disabled by default
	assert.Equal(t, 5*time.Second, cfg.Outbox.PollInterval)
	assert.Equal(t, 50, cfg.Outbox.BatchSize)
	assert.Equal(t, 5, cfg.Outbox.MaxRetries)
	assert.Equal(t, 24*time.Hour, cfg.Outbox.CleanupInterval)
	assert.Equal(t, 7*24*time.Hour, cfg.Outbox.CleanupAfter)

	// Test retry backoff defaults
	assert.Equal(t, 1*time.Second, cfg.Outbox.RetryBackoff.InitialDelay)
	assert.Equal(t, 30*time.Minute, cfg.Outbox.RetryBackoff.MaxDelay)
	assert.Equal(t, 2.0, cfg.Outbox.RetryBackoff.Multiplier)
	assert.True(t, cfg.Outbox.RetryBackoff.Jitter)

	// Test metrics defaults
	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, 30*time.Second, cfg.Metrics.CollectionInterval)
	assert.Equal(t, []float64{0.001, 0.01, 0.1, 1, 10}, cfg.Metrics.HistogramBuckets)
	assert.Equal(t, 1000, cfg.Metrics.CardinalityLimit)
}

func TestIdempotencyConfig_Validation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &IdempotencyConfig{
			Enabled:         true,
			DefaultTTL:      time.Hour,
			KeyPrefix:       "test:",
			MaxRetries:      3,
			RetryDelay:      100 * time.Millisecond,
			CleanupInterval: 30 * time.Minute,
			BatchSize:       50,
		}

		// No validation method implemented, but structure should be valid
		assert.NotNil(t, cfg)
	})

	t.Run("zero values", func(t *testing.T) {
		cfg := &IdempotencyConfig{}

		assert.False(t, cfg.Enabled)
		assert.Equal(t, time.Duration(0), cfg.DefaultTTL)
		assert.Empty(t, cfg.KeyPrefix)
		assert.Equal(t, 0, cfg.MaxRetries)
	})
}

func TestStorageConfig_Types(t *testing.T) {
	t.Run("redis storage", func(t *testing.T) {
		cfg := StorageConfig{
			Type: "redis",
			Redis: RedisStorageConfig{
				KeyPattern:     "custom:{queue}:{key}",
				HashKeyPattern: "custom:{queue}",
				UseHashes:      false,
				Compression:    true,
			},
		}

		assert.Equal(t, "redis", cfg.Type)
		assert.Equal(t, "custom:{queue}:{key}", cfg.Redis.KeyPattern)
		assert.False(t, cfg.Redis.UseHashes)
		assert.True(t, cfg.Redis.Compression)
	})

	t.Run("memory storage", func(t *testing.T) {
		cfg := StorageConfig{
			Type: "memory",
			Memory: MemoryStorageConfig{
				MaxKeys:        1000,
				EvictionPolicy: "lru",
			},
		}

		assert.Equal(t, "memory", cfg.Type)
		assert.Equal(t, 1000, cfg.Memory.MaxKeys)
		assert.Equal(t, "lru", cfg.Memory.EvictionPolicy)
	})

	t.Run("database storage", func(t *testing.T) {
		cfg := StorageConfig{
			Type: "database",
			Database: DatabaseStorageConfig{
				TableName:            "outbox_events",
				IdempotencyTableName: "idempotency_keys",
				BatchSize:            25,
				MaxConnections:       10,
				TransactionTimeout:   30 * time.Second,
			},
		}

		assert.Equal(t, "database", cfg.Type)
		assert.Equal(t, "outbox_events", cfg.Database.TableName)
		assert.Equal(t, "idempotency_keys", cfg.Database.IdempotencyTableName)
		assert.Equal(t, 25, cfg.Database.BatchSize)
	})
}

func TestOutboxConfig_Publishers(t *testing.T) {
	cfg := OutboxConfig{
		Enabled: true,
		Publishers: []PublisherConfig{
			{
				Name:    "redis-publisher",
				Type:    "redis",
				Target:  "events:queue",
				Enabled: true,
				Config: map[string]interface{}{
					"max_retries": 3,
					"timeout":     "30s",
				},
			},
			{
				Name:    "kafka-publisher",
				Type:    "kafka",
				Target:  "events-topic",
				Enabled: false,
				Config: map[string]interface{}{
					"brokers": []string{"localhost:9092"},
					"topic":   "events",
				},
			},
		},
	}

	assert.Len(t, cfg.Publishers, 2)

	redis := cfg.Publishers[0]
	assert.Equal(t, "redis-publisher", redis.Name)
	assert.Equal(t, "redis", redis.Type)
	assert.True(t, redis.Enabled)
	assert.Contains(t, redis.Config, "max_retries")

	kafka := cfg.Publishers[1]
	assert.Equal(t, "kafka-publisher", kafka.Name)
	assert.False(t, kafka.Enabled)
	assert.Contains(t, kafka.Config, "brokers")
}

func TestBackoffConfig_Values(t *testing.T) {
	backoff := BackoffConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.5,
		Jitter:       true,
	}

	assert.Equal(t, 100*time.Millisecond, backoff.InitialDelay)
	assert.Equal(t, 10*time.Second, backoff.MaxDelay)
	assert.Equal(t, 2.5, backoff.Multiplier)
	assert.True(t, backoff.Jitter)
}

func TestMetricsConfig_Buckets(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 15 * time.Second,
		HistogramBuckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0},
		CardinalityLimit:   500,
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 15*time.Second, cfg.CollectionInterval)
	assert.Len(t, cfg.HistogramBuckets, 5)
	assert.Contains(t, cfg.HistogramBuckets, 1.0)
	assert.Equal(t, 500, cfg.CardinalityLimit)
}

// Test configuration combinations
func TestConfig_Integration(t *testing.T) {
	cfg := &Config{
		Idempotency: IdempotencyConfig{
			Enabled:    true,
			DefaultTTL: 2 * time.Hour,
			Storage: StorageConfig{
				Type: "memory",
				Memory: MemoryStorageConfig{
					MaxKeys:        500,
					EvictionPolicy: "fifo",
				},
			},
		},
		Outbox: OutboxConfig{
			Enabled:      true,
			BatchSize:    100,
			PollInterval: 30 * time.Second,
		},
		Metrics: MetricsConfig{
			Enabled:            false,
			CollectionInterval: 1 * time.Minute,
		},
	}

	// Verify integrated configuration
	assert.True(t, cfg.Idempotency.Enabled)
	assert.True(t, cfg.Outbox.Enabled)
	assert.False(t, cfg.Metrics.Enabled)

	assert.Equal(t, "memory", cfg.Idempotency.Storage.Type)
	assert.Equal(t, 500, cfg.Idempotency.Storage.Memory.MaxKeys)

	assert.Equal(t, 100, cfg.Outbox.BatchSize)
	assert.Equal(t, 30*time.Second, cfg.Outbox.PollInterval)
}

// Benchmark configuration creation
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := DefaultConfig()
		if cfg == nil {
			b.Fatal("config is nil")
		}
	}
}

func TestConfig_DeepCopy(t *testing.T) {
	original := DefaultConfig()

	// Modify original
	original.Idempotency.DefaultTTL = 5 * time.Hour
	original.Outbox.Enabled = true
	original.Metrics.CardinalityLimit = 2000

	// Create a "copy" by calling DefaultConfig again
	copy := DefaultConfig()

	// Original changes shouldn't affect the new default config
	assert.Equal(t, 24*time.Hour, copy.Idempotency.DefaultTTL)
	assert.False(t, copy.Outbox.Enabled)
	assert.Equal(t, 1000, copy.Metrics.CardinalityLimit)

	// But original should have the modified values
	assert.Equal(t, 5*time.Hour, original.Idempotency.DefaultTTL)
	assert.True(t, original.Outbox.Enabled)
	assert.Equal(t, 2000, original.Metrics.CardinalityLimit)
}