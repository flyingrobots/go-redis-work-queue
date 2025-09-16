// Copyright 2025 James Ross
package chaosharness

import (
	"math/rand"
	"sync"
	"time"
)

// MetricsCollector collects metrics during chaos testing
type MetricsCollector struct {
	// Simulated metrics for demonstration
	// In production, this would integrate with actual metrics systems
	mu     sync.RWMutex
	random *rand.Rand

	// Current metrics
	requests       int64
	successful     int64
	failed         int64
	faultsInjected int64
	backlogSize    int64
	latencies      []float64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
		latencies: make([]float64, 0, 1000),
	}
}

// Collect collects current metrics
func (mc *MetricsCollector) Collect() TimeSeriesPoint {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Simulate metrics collection
	mc.requests += int64(mc.random.Intn(100))
	mc.successful += int64(mc.random.Intn(90))
	mc.failed = mc.requests - mc.successful
	mc.faultsInjected += int64(mc.random.Intn(10))
	mc.backlogSize = int64(mc.random.Intn(1000))

	// Generate latency data
	for i := 0; i < 100; i++ {
		latency := 10.0 + mc.random.Float64()*90.0 // 10-100ms
		mc.latencies = append(mc.latencies, latency)
	}

	// Keep only recent latencies
	if len(mc.latencies) > 1000 {
		mc.latencies = mc.latencies[len(mc.latencies)-1000:]
	}

	// Calculate percentiles
	p50 := mc.percentile(mc.latencies, 0.5)
	p95 := mc.percentile(mc.latencies, 0.95)
	p99 := mc.percentile(mc.latencies, 0.99)

	errorRate := 0.0
	if mc.requests > 0 {
		errorRate = float64(mc.failed) / float64(mc.requests)
	}

	return TimeSeriesPoint{
		Timestamp: time.Now(),
		Metrics: map[string]float64{
			"requests":         float64(mc.requests),
			"successful":       float64(mc.successful),
			"failed":           float64(mc.failed),
			"faults_injected":  float64(mc.faultsInjected),
			"backlog_size":     float64(mc.backlogSize),
			"error_rate":       errorRate,
			"latency_p50_ms":   p50,
			"latency_p95_ms":   p95,
			"latency_p99_ms":   p99,
		},
	}
}

// percentile calculates percentile from sorted data
func (mc *MetricsCollector) percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Simple percentile calculation (not exact but good enough for demo)
	index := int(float64(len(data)) * p)
	if index >= len(data) {
		index = len(data) - 1
	}

	return data[index]
}

// Reset resets the metrics collector
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requests = 0
	mc.successful = 0
	mc.failed = 0
	mc.faultsInjected = 0
	mc.backlogSize = 0
	mc.latencies = make([]float64, 0, 1000)
}
