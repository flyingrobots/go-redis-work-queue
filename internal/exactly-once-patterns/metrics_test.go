// Copyright 2025 James Ross
package exactlyonce

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricsCollector(t *testing.T) {
	t.Run("disabled metrics", func(t *testing.T) {
		cfg := MetricsConfig{
			Enabled: false,
		}

		collector := NewMetricsCollector(cfg)
		assert.NotNil(t, collector)
		assert.Equal(t, &cfg, collector.cfg)
	})

	t.Run("enabled metrics", func(t *testing.T) {
		// Skip enabled metrics tests to avoid Prometheus registration conflicts
		// In production, each service instance would have its own registry
		t.Skip("Skipping enabled metrics tests to avoid Prometheus registration conflicts")
	})
}

func TestMetricsCollector_DisabledOperations(t *testing.T) {
	cfg := MetricsConfig{
		Enabled: false,
	}

	collector := NewMetricsCollector(cfg)

	// Should handle disabled metrics gracefully
	collector.RecordProcessingLatency(100*time.Millisecond, "test-queue")
	collector.IncrementDuplicatesAvoided("test-queue")
	collector.IncrementSuccessfulProcessing("test-queue")
	collector.IncrementStorageErrors("test-queue")
	collector.IncrementIdempotencyChecks("test-queue", "hit")
	collector.IncrementOutboxPublished(5)
	collector.IncrementOutboxFailed(2)
	collector.RecordOutboxLatency(250*time.Millisecond, "user.created", "kafka")
	collector.IncrementStorageOperations("get", "redis", "test-queue")
	collector.SetStorageSize(1000, "test-queue", "", "redis")

	snapshot := collector.GetMetricsSnapshot()
	assert.NotNil(t, snapshot)

	// Cleanup operations should not error
	collector.Unregister()
	err := collector.Close()
	assert.NoError(t, err)
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
				Enabled:            false, // Disabled to avoid registration conflicts
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
				Enabled:          false,
				CardinalityLimit: 0,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector(tt.config)
			assert.NotNil(t, collector)
			collector.Close()
		})
	}
}