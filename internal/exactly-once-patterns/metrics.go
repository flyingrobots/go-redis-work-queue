// Copyright 2025 James Ross
package exactlyonce

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector manages metrics for exactly-once patterns
type MetricsCollector struct {
	cfg *MetricsConfig

	// Idempotency metrics
	processingLatency    *prometheus.HistogramVec
	duplicatesAvoided    *prometheus.CounterVec
	successfulProcessing *prometheus.CounterVec
	storageErrors        *prometheus.CounterVec
	idempotencyChecks    *prometheus.CounterVec

	// Outbox metrics
	outboxPublished *prometheus.CounterVec
	outboxFailed    *prometheus.CounterVec
	outboxLatency   *prometheus.HistogramVec

	// Storage metrics
	storageOperations *prometheus.CounterVec
	storageSize       *prometheus.GaugeVec

	// Cardinality tracking
	cardinalityLimiter sync.Map
	mu                 sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(cfg MetricsConfig) *MetricsCollector {
	if !cfg.Enabled {
		return &MetricsCollector{cfg: &cfg}
	}

	buckets := cfg.HistogramBuckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	mc := &MetricsCollector{
		cfg: &cfg,

		processingLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "exactly_once_processing_duration_seconds",
				Help:    "Time spent processing jobs with idempotency checks",
				Buckets: buckets,
			},
			[]string{"queue", "tenant"},
		),

		duplicatesAvoided: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_duplicates_avoided_total",
				Help: "Total number of duplicate job executions avoided",
			},
			[]string{"queue", "tenant"},
		),

		successfulProcessing: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_successful_processing_total",
				Help: "Total number of successful job processing with idempotency",
			},
			[]string{"queue", "tenant"},
		),

		storageErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_storage_errors_total",
				Help: "Total number of idempotency storage errors",
			},
			[]string{"queue", "tenant", "operation"},
		),

		idempotencyChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_idempotency_checks_total",
				Help: "Total number of idempotency checks performed",
			},
			[]string{"queue", "tenant", "result"},
		),

		outboxPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_outbox_published_total",
				Help: "Total number of outbox events published successfully",
			},
			[]string{"event_type", "publisher"},
		),

		outboxFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_outbox_failed_total",
				Help: "Total number of outbox events that failed to publish",
			},
			[]string{"event_type", "publisher", "error"},
		),

		outboxLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "exactly_once_outbox_processing_duration_seconds",
				Help:    "Time spent processing outbox events",
				Buckets: buckets,
			},
			[]string{"event_type", "publisher"},
		),

		storageOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "exactly_once_storage_operations_total",
				Help: "Total number of storage operations performed",
			},
			[]string{"operation", "storage_type", "result"},
		),

		storageSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "exactly_once_storage_size",
				Help: "Current size of idempotency storage",
			},
			[]string{"queue", "tenant", "storage_type"},
		),
	}

	return mc
}

// RecordProcessingLatency records the time taken for processing with idempotency
func (m *MetricsCollector) RecordProcessingLatency(duration time.Duration, queueName string) {
	if !m.cfg.Enabled || m.processingLatency == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, ""})
	m.processingLatency.WithLabelValues(labels...).Observe(duration.Seconds())
}

// IncrementDuplicatesAvoided increments the counter for duplicates avoided
func (m *MetricsCollector) IncrementDuplicatesAvoided(queueName string) {
	if !m.cfg.Enabled || m.duplicatesAvoided == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, ""})
	m.duplicatesAvoided.WithLabelValues(labels...).Inc()
}

// IncrementSuccessfulProcessing increments the counter for successful processing
func (m *MetricsCollector) IncrementSuccessfulProcessing(queueName string) {
	if !m.cfg.Enabled || m.successfulProcessing == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, ""})
	m.successfulProcessing.WithLabelValues(labels...).Inc()
}

// IncrementStorageErrors increments the counter for storage errors
func (m *MetricsCollector) IncrementStorageErrors(queueName string) {
	if !m.cfg.Enabled || m.storageErrors == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, "", "check"})
	m.storageErrors.WithLabelValues(labels...).Inc()
}

// IncrementIdempotencyChecks increments the counter for idempotency checks
func (m *MetricsCollector) IncrementIdempotencyChecks(queueName, result string) {
	if !m.cfg.Enabled || m.idempotencyChecks == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, "", result})
	m.idempotencyChecks.WithLabelValues(labels...).Inc()
}

// IncrementOutboxPublished increments the counter for successful outbox publications
func (m *MetricsCollector) IncrementOutboxPublished(count int) {
	if !m.cfg.Enabled || m.outboxPublished == nil {
		return
	}

	labels := m.limitCardinality([]string{"unknown", "default"})
	m.outboxPublished.WithLabelValues(labels...).Add(float64(count))
}

// IncrementOutboxFailed increments the counter for failed outbox publications
func (m *MetricsCollector) IncrementOutboxFailed(count int) {
	if !m.cfg.Enabled || m.outboxFailed == nil {
		return
	}

	labels := m.limitCardinality([]string{"unknown", "default", "publish_error"})
	m.outboxFailed.WithLabelValues(labels...).Add(float64(count))
}

// RecordOutboxLatency records the time taken for outbox processing
func (m *MetricsCollector) RecordOutboxLatency(duration time.Duration, eventType, publisher string) {
	if !m.cfg.Enabled || m.outboxLatency == nil {
		return
	}

	labels := m.limitCardinality([]string{eventType, publisher})
	m.outboxLatency.WithLabelValues(labels...).Observe(duration.Seconds())
}

// IncrementStorageOperations increments the counter for storage operations
func (m *MetricsCollector) IncrementStorageOperations(operation, storageType, result string) {
	if !m.cfg.Enabled || m.storageOperations == nil {
		return
	}

	labels := m.limitCardinality([]string{operation, storageType, result})
	m.storageOperations.WithLabelValues(labels...).Inc()
}

// SetStorageSize sets the current size of the storage
func (m *MetricsCollector) SetStorageSize(size int64, queueName, tenantID, storageType string) {
	if !m.cfg.Enabled || m.storageSize == nil {
		return
	}

	labels := m.limitCardinality([]string{queueName, tenantID, storageType})
	m.storageSize.WithLabelValues(labels...).Set(float64(size))
}

// limitCardinality prevents metric explosion by limiting label cardinality
func (m *MetricsCollector) limitCardinality(labels []string) []string {
	if m.cfg.CardinalityLimit <= 0 {
		return labels
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// This is a simplified cardinality limiter
	// In production, you'd want more sophisticated logic
	result := make([]string, len(labels))
	for i, label := range labels {
		if label == "" {
			result[i] = ""
			continue
		}

		key := label
		if count, exists := m.cardinalityLimiter.Load(key); exists {
			if count.(int) >= m.cfg.CardinalityLimit {
				result[i] = "other" // Replace high-cardinality labels
			} else {
				result[i] = label
				m.cardinalityLimiter.Store(key, count.(int)+1)
			}
		} else {
			result[i] = label
			m.cardinalityLimiter.Store(key, 1)
		}
	}

	return result
}

// GetMetricsSnapshot returns current metrics values for monitoring
func (m *MetricsCollector) GetMetricsSnapshot() map[string]interface{} {
	if !m.cfg.Enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	return map[string]interface{}{
		"enabled": true,
		// TODO: Add actual metric values extraction if needed
		// This would require walking through the prometheus metrics
	}
}

// Close cleans up the metrics collector
func (m *MetricsCollector) Close() error {
	// TODO: Unregister metrics if needed
	return nil
}