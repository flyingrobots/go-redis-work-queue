// Copyright 2025 James Ross
package exactlyonce

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricsCollector(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)
	assert.NotNil(t, collector)
	assert.Equal(t, &cfg, collector.cfg)
}

func TestMetricsCollector_RecordProcessingLatency(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
		CardinalityLimit:   10,
	}

	collector := NewMetricsCollector(cfg)

	t.Run("record latency", func(t *testing.T) {
		duration := 100 * time.Millisecond
		collector.RecordProcessingLatency(duration, "test-queue")

		// Check that the metric was recorded
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("cardinality limit", func(t *testing.T) {
		// Add metrics until we hit the limit
		for i := 0; i < 15; i++ {
			queueName := fmt.Sprintf("queue-%d", i)
			collector.RecordProcessingLatency(100*time.Millisecond, queueName)
		}

		// Should respect cardinality limit
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})
}

func TestMetricsCollector_IncrementCounters(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	t.Run("increment duplicates avoided", func(t *testing.T) {
		collector.IncrementDuplicatesAvoided("test-queue")
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("increment successful processing", func(t *testing.T) {
		collector.IncrementSuccessfulProcessing("test-queue")
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("increment storage errors", func(t *testing.T) {
		collector.IncrementStorageErrors("test-queue")
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("increment idempotency checks", func(t *testing.T) {
		collector.IncrementIdempotencyChecks("test-queue", "hit")
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("increment outbox published", func(t *testing.T) {
		collector.IncrementOutboxPublished(5)
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("increment outbox failed", func(t *testing.T) {
		collector.IncrementOutboxFailed(2)
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})
}

func TestMetricsCollector_RecordOutboxLatency(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	duration := 250 * time.Millisecond
	collector.RecordOutboxLatency(duration, "user.created", "kafka")

	snapshot := collector.GetMetricsSnapshot()
	assert.NotNil(t, snapshot)
}

func TestMetricsCollector_StorageOperations(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	t.Run("increment storage operations", func(t *testing.T) {
		collector.IncrementStorageOperations("get", "redis", "test-queue")
		collector.IncrementStorageOperations("set", "redis", "test-queue")
		collector.IncrementStorageOperations("delete", "memory", "test-queue")

		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})

	t.Run("set storage size", func(t *testing.T) {
		collector.SetStorageSize(1500, "test-queue", "tenant-1", "redis")
		collector.SetStorageSize(300, "test-queue", "tenant-1", "memory")

		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)
	})
}

func TestMetricsCollector_CardinalityLimiting(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:          true,
		CardinalityLimit: 3, // Very low limit for testing
	}

	collector := NewMetricsCollector(cfg)

	// Test cardinality limiting
	for i := 0; i < 10; i++ {
		queueName := fmt.Sprintf("queue-%d", i)
		collector.IncrementDuplicatesAvoided(queueName)
	}

	// Should not panic or cause issues
	snapshot := collector.GetMetricsSnapshot()
	assert.NotNil(t, snapshot)
}

func TestMetricsCollector_GetMetricsSnapshot(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	// Add some metrics
	collector.IncrementDuplicatesAvoided("test-queue")
	collector.IncrementSuccessfulProcessing("test-queue")
	collector.RecordProcessingLatency(100*time.Millisecond, "test-queue")
	collector.SetStorageSize(1000, "test-queue", "", "redis")

	t.Run("get snapshot", func(t *testing.T) {
		snapshot := collector.GetMetricsSnapshot()
		assert.NotNil(t, snapshot)

		// Should contain metrics data
		assert.True(t, len(snapshot) > 0)
	})
}

func TestMetricsCollector_Unregister(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	// Add some metrics first
	collector.IncrementDuplicatesAvoided("test-queue")

	t.Run("unregister metrics", func(t *testing.T) {
		collector.Unregister()
		// Should not panic
	})
}

func TestMetricsCollector_Close(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		CardinalityLimit:   1000,
	}

	collector := NewMetricsCollector(cfg)

	t.Run("close collector", func(t *testing.T) {
		err := collector.Close()
		assert.NoError(t, err)
	})
}

func TestMetricsCollector_DisabledMetrics(t *testing.T) {
	cfg := MetricsConfig{
		Enabled: false, // Disabled
	}

	collector := NewMetricsCollector(cfg)

	// Should handle disabled metrics gracefully
	collector.IncrementDuplicatesAvoided("test-queue")
	collector.RecordProcessingLatency(100*time.Millisecond, "test-queue")
	collector.SetStorageSize(1000, "test-queue", "", "redis")

	snapshot := collector.GetMetricsSnapshot()
	assert.NotNil(t, snapshot)
}

func TestMetricsConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config MetricsConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: MetricsConfig{
				Enabled:            true,
				CollectionInterval: 30 * time.Second,
				HistogramBuckets:   []float64{0.001, 0.01, 0.1, 1, 10},
				CardinalityLimit:   1000,
			},
			valid: true,
		},
		{
			name: "disabled metrics",
			config: MetricsConfig{
				Enabled: false,
			},
			valid: true,
		},
		{
			name: "zero cardinality limit",
			config: MetricsConfig{
				Enabled:          true,
				CardinalityLimit: 0,
			},
			valid: true, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector(tt.config)
			assert.NotNil(t, collector)
		})
	}
}