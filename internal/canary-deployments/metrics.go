package canary_deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMetricsCollector implements the MetricsCollector interface using Redis
type RedisMetricsCollector struct {
	redis  *redis.Client
	logger *slog.Logger
}

// NewRedisMetricsCollector creates a new Redis-based metrics collector
func NewRedisMetricsCollector(redis *redis.Client, logger *slog.Logger) *RedisMetricsCollector {
	return &RedisMetricsCollector{
		redis:  redis,
		logger: logger,
	}
}

// CollectSnapshot collects a metrics snapshot for a specific queue and version
func (rmc *RedisMetricsCollector) CollectSnapshot(ctx context.Context, queue string, version string, window time.Duration) (*MetricsSnapshot, error) {
	windowStart := time.Now().Add(-window)
	windowEnd := time.Now()

	// Collect metrics from Redis
	metrics, err := rmc.collectMetricsFromRedis(ctx, queue, version, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics from Redis: %w", err)
	}

	// Calculate derived metrics
	snapshot := &MetricsSnapshot{
		Timestamp:   time.Now(),
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		Version:     version,
	}

	if len(metrics) > 0 {
		rmc.calculateSnapshot(snapshot, metrics)
	}

	// Get additional queue metrics
	if err := rmc.enrichWithQueueMetrics(ctx, queue, version, snapshot); err != nil {
		rmc.logger.Warn("Failed to enrich with queue metrics", "error", err)
	}

	rmc.logger.Debug("Collected metrics snapshot",
		"queue", queue,
		"version", version,
		"window", window,
		"job_count", snapshot.JobCount,
		"error_rate", snapshot.ErrorRate)

	return snapshot, nil
}

// GetHistoricalMetrics returns historical metrics for a queue and version
func (rmc *RedisMetricsCollector) GetHistoricalMetrics(ctx context.Context, queue string, version string, since time.Time) ([]*MetricsSnapshot, error) {
	pattern := fmt.Sprintf("canary:metrics:%s:%s:*", queue, version)
	keys, err := rmc.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list metric keys: %w", err)
	}

	snapshots := make([]*MetricsSnapshot, 0)

	for _, key := range keys {
		data, err := rmc.redis.Get(ctx, key).Result()
		if err != nil {
			rmc.logger.Warn("Failed to load metrics", "key", key, "error", err)
			continue
		}

		var snapshot MetricsSnapshot
		if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
			rmc.logger.Warn("Failed to unmarshal metrics", "key", key, "error", err)
			continue
		}

		if snapshot.Timestamp.After(since) {
			snapshots = append(snapshots, &snapshot)
		}
	}

	// Sort by timestamp
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

// StoreJobMetrics stores metrics for a completed job
func (rmc *RedisMetricsCollector) StoreJobMetrics(ctx context.Context, job *Job, metrics *JobExecutionMetrics) error {
	// Create a job metrics entry
	jobMetric := &JobMetric{
		JobID:          job.ID,
		Queue:          job.Queue,
		Type:           job.Type,
		Version:        job.Version,
		Lane:           job.Lane,
		TenantID:       job.TenantID,
		WorkerID:       job.WorkerID,
		Success:        metrics.Success,
		ProcessingTime: metrics.ProcessingTime,
		MemoryUsage:    metrics.MemoryUsage,
		PayloadSize:    int64(len(fmt.Sprintf("%v", job.Payload))),
		StartTime:      metrics.StartTime,
		EndTime:        metrics.EndTime,
		ErrorMessage:   metrics.ErrorMessage,
	}

	// Store in Redis sorted set for time-based queries
	key := fmt.Sprintf("canary:job_metrics:%s:%s", job.Queue, job.Version)
	score := float64(jobMetric.EndTime.Unix())

	data, err := json.Marshal(jobMetric)
	if err != nil {
		return fmt.Errorf("failed to marshal job metric: %w", err)
	}

	if err := rmc.redis.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store job metric: %w", err)
	}

	// Set expiration to prevent unlimited growth
	rmc.redis.Expire(ctx, key, 7*24*time.Hour) // 7 days

	return nil
}

// StoreWorkerMetrics stores metrics for a worker
func (rmc *RedisMetricsCollector) StoreWorkerMetrics(ctx context.Context, workerID string, metrics *WorkerMetrics) error {
	key := fmt.Sprintf("canary:worker_metrics:%s", workerID)

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal worker metrics: %w", err)
	}

	if err := rmc.redis.HSet(ctx, key, "metrics", data, "timestamp", time.Now().Unix()).Err(); err != nil {
		return fmt.Errorf("failed to store worker metrics: %w", err)
	}

	// Set expiration
	rmc.redis.Expire(ctx, key, 1*time.Hour)

	return nil
}

// CreatePeriodicSnapshot creates and stores a periodic metrics snapshot
func (rmc *RedisMetricsCollector) CreatePeriodicSnapshot(ctx context.Context, queue string, version string, window time.Duration) error {
	snapshot, err := rmc.CollectSnapshot(ctx, queue, version, window)
	if err != nil {
		return fmt.Errorf("failed to collect snapshot: %w", err)
	}

	// Store snapshot in Redis
	key := fmt.Sprintf("canary:metrics:%s:%s:%d", queue, version, time.Now().Unix())

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := rmc.redis.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}

	return nil
}

// Private methods

func (rmc *RedisMetricsCollector) collectMetricsFromRedis(ctx context.Context, queue string, version string, start, end time.Time) ([]*JobMetric, error) {
	key := fmt.Sprintf("canary:job_metrics:%s:%s", queue, version)

	// Query sorted set by score range (timestamp)
	startScore := float64(start.Unix())
	endScore := float64(end.Unix())

	results, err := rmc.redis.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", startScore),
		Max: fmt.Sprintf("%f", endScore),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	metrics := make([]*JobMetric, 0, len(results))
	for _, result := range results {
		var metric JobMetric
		if err := json.Unmarshal([]byte(result), &metric); err != nil {
			rmc.logger.Warn("Failed to unmarshal job metric", "error", err)
			continue
		}
		metrics = append(metrics, &metric)
	}

	return metrics, nil
}

func (rmc *RedisMetricsCollector) calculateSnapshot(snapshot *MetricsSnapshot, metrics []*JobMetric) {
	if len(metrics) == 0 {
		return
	}

	snapshot.JobCount = int64(len(metrics))

	var successCount int64
	var totalLatency float64
	var latencies []float64
	var totalMemory float64
	var workerIDs = make(map[string]bool)

	for _, metric := range metrics {
		if metric.Success {
			successCount++
		}

		latency := float64(metric.ProcessingTime.Milliseconds())
		totalLatency += latency
		latencies = append(latencies, latency)

		totalMemory += metric.MemoryUsage

		if metric.WorkerID != "" {
			workerIDs[metric.WorkerID] = true
		}
	}

	// Basic metrics
	snapshot.SuccessCount = successCount
	snapshot.ErrorCount = snapshot.JobCount - successCount

	if snapshot.JobCount > 0 {
		snapshot.SuccessRate = float64(successCount) / float64(snapshot.JobCount) * 100
		snapshot.ErrorRate = float64(snapshot.ErrorCount) / float64(snapshot.JobCount) * 100
	}

	// Latency metrics
	if len(latencies) > 0 {
		sort.Float64s(latencies)

		snapshot.AvgLatency = totalLatency / float64(len(latencies))
		snapshot.P50Latency = rmc.percentile(latencies, 0.5)
		snapshot.P95Latency = rmc.percentile(latencies, 0.95)
		snapshot.P99Latency = rmc.percentile(latencies, 0.99)
		snapshot.MaxLatency = latencies[len(latencies)-1]
	}

	// Throughput calculation
	windowDuration := snapshot.WindowEnd.Sub(snapshot.WindowStart)
	if windowDuration > 0 {
		snapshot.JobsPerSecond = float64(snapshot.JobCount) / windowDuration.Seconds()
	}

	// Resource metrics
	if snapshot.JobCount > 0 {
		snapshot.AvgMemoryMB = totalMemory / float64(snapshot.JobCount)
		snapshot.PeakMemoryMB = rmc.findMaxMemory(metrics)
	}

	snapshot.WorkerCount = len(workerIDs)
}

func (rmc *RedisMetricsCollector) enrichWithQueueMetrics(ctx context.Context, queue string, version string, snapshot *MetricsSnapshot) error {
	// Get queue depth
	queueKey := queue
	if version != "" {
		// Try version-specific queue first
		versionQueue := fmt.Sprintf("%s@%s", queue, version)
		depth, err := rmc.redis.LLen(ctx, versionQueue).Result()
		if err == nil {
			snapshot.QueueDepth = depth
		} else {
			// Fall back to main queue
			depth, err := rmc.redis.LLen(ctx, queueKey).Result()
			if err != nil {
				return fmt.Errorf("failed to get queue depth: %w", err)
			}
			snapshot.QueueDepth = depth
		}
	} else {
		depth, err := rmc.redis.LLen(ctx, queueKey).Result()
		if err != nil {
			return fmt.Errorf("failed to get queue depth: %w", err)
		}
		snapshot.QueueDepth = depth
	}

	// Get dead letter queue metrics
	dlqKey := fmt.Sprintf("%s:dlq", queue)
	dlqDepth, err := rmc.redis.LLen(ctx, dlqKey).Result()
	if err != nil {
		// DLQ might not exist
		dlqDepth = 0
	}
	snapshot.DeadLetters = dlqDepth

	return nil
}

func (rmc *RedisMetricsCollector) percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (rmc *RedisMetricsCollector) findMaxMemory(metrics []*JobMetric) float64 {
	var max float64
	for _, metric := range metrics {
		if metric.MemoryUsage > max {
			max = metric.MemoryUsage
		}
	}
	return max
}

// JobMetric represents metrics for a single job execution
type JobMetric struct {
	JobID          string        `json:"job_id"`
	Queue          string        `json:"queue"`
	Type           string        `json:"type"`
	Version        string        `json:"version"`
	Lane           string        `json:"lane"`
	TenantID       string        `json:"tenant_id"`
	WorkerID       string        `json:"worker_id"`
	Success        bool          `json:"success"`
	ProcessingTime time.Duration `json:"processing_time"`
	MemoryUsage    float64       `json:"memory_usage_mb"`
	PayloadSize    int64         `json:"payload_size_bytes"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	ErrorMessage   string        `json:"error_message,omitempty"`
}

// JobExecutionMetrics represents metrics collected during job execution
type JobExecutionMetrics struct {
	Success        bool          `json:"success"`
	ProcessingTime time.Duration `json:"processing_time"`
	MemoryUsage    float64       `json:"memory_usage_mb"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	ErrorMessage   string        `json:"error_message,omitempty"`
}

// PerformanceAnalyzer provides advanced metrics analysis
type PerformanceAnalyzer struct {
	collector *RedisMetricsCollector
	logger    *slog.Logger
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(collector *RedisMetricsCollector, logger *slog.Logger) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		collector: collector,
		logger:    logger,
	}
}

// ComparePerformance compares performance between stable and canary versions
func (pa *PerformanceAnalyzer) ComparePerformance(ctx context.Context, queue string, stableVersion, canaryVersion string, window time.Duration) (*PerformanceComparison, error) {
	stableSnapshot, err := pa.collector.CollectSnapshot(ctx, queue, stableVersion, window)
	if err != nil {
		return nil, fmt.Errorf("failed to collect stable metrics: %w", err)
	}

	canarySnapshot, err := pa.collector.CollectSnapshot(ctx, queue, canaryVersion, window)
	if err != nil {
		return nil, fmt.Errorf("failed to collect canary metrics: %w", err)
	}

	comparison := &PerformanceComparison{
		Queue:         queue,
		StableVersion: stableVersion,
		CanaryVersion: canaryVersion,
		Window:        window,
		Timestamp:     time.Now(),
		Stable:        stableSnapshot,
		Canary:        canarySnapshot,
	}

	// Calculate deltas
	comparison.calculateDeltas()

	return comparison, nil
}

// DetectAnomalies detects performance anomalies in the metrics
func (pa *PerformanceAnalyzer) DetectAnomalies(ctx context.Context, queue string, version string) ([]*PerformanceAnomaly, error) {
	// Get recent metrics for trend analysis
	since := time.Now().Add(-2 * time.Hour)
	snapshots, err := pa.collector.GetHistoricalMetrics(ctx, queue, version, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical metrics: %w", err)
	}

	if len(snapshots) < 3 {
		// Not enough data for anomaly detection
		return []*PerformanceAnomaly{}, nil
	}

	anomalies := make([]*PerformanceAnomaly, 0)

	// Check for error rate spikes
	if errorRateAnomaly := pa.detectErrorRateAnomaly(snapshots); errorRateAnomaly != nil {
		anomalies = append(anomalies, errorRateAnomaly)
	}

	// Check for latency spikes
	if latencyAnomaly := pa.detectLatencyAnomaly(snapshots); latencyAnomaly != nil {
		anomalies = append(anomalies, latencyAnomaly)
	}

	// Check for throughput drops
	if throughputAnomaly := pa.detectThroughputAnomaly(snapshots); throughputAnomaly != nil {
		anomalies = append(anomalies, throughputAnomaly)
	}

	return anomalies, nil
}

// PerformanceComparison represents a comparison between stable and canary performance
type PerformanceComparison struct {
	Queue         string           `json:"queue"`
	StableVersion string           `json:"stable_version"`
	CanaryVersion string           `json:"canary_version"`
	Window        time.Duration    `json:"window"`
	Timestamp     time.Time        `json:"timestamp"`
	Stable        *MetricsSnapshot `json:"stable"`
	Canary        *MetricsSnapshot `json:"canary"`

	// Calculated deltas
	ErrorRateDelta   float64 `json:"error_rate_delta"`   // Percentage points
	LatencyDelta     float64 `json:"latency_delta"`      // Percentage change
	ThroughputDelta  float64 `json:"throughput_delta"`   // Percentage change
	MemoryDelta      float64 `json:"memory_delta"`       // Percentage change
	SuccessRateDelta float64 `json:"success_rate_delta"` // Percentage points
}

func (pc *PerformanceComparison) calculateDeltas() {
	if pc.Stable == nil || pc.Canary == nil {
		return
	}

	// Error rate delta (percentage points)
	pc.ErrorRateDelta = pc.Canary.ErrorRate - pc.Stable.ErrorRate

	// Success rate delta (percentage points)
	pc.SuccessRateDelta = pc.Canary.SuccessRate - pc.Stable.SuccessRate

	// Latency delta (percentage change)
	if pc.Stable.P95Latency > 0 {
		pc.LatencyDelta = (pc.Canary.P95Latency - pc.Stable.P95Latency) / pc.Stable.P95Latency * 100
	}

	// Throughput delta (percentage change)
	if pc.Stable.JobsPerSecond > 0 {
		pc.ThroughputDelta = (pc.Canary.JobsPerSecond - pc.Stable.JobsPerSecond) / pc.Stable.JobsPerSecond * 100
	}

	// Memory delta (percentage change)
	if pc.Stable.AvgMemoryMB > 0 {
		pc.MemoryDelta = (pc.Canary.AvgMemoryMB - pc.Stable.AvgMemoryMB) / pc.Stable.AvgMemoryMB * 100
	}
}

// PerformanceAnomaly represents a detected performance anomaly
type PerformanceAnomaly struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
}

func (pa *PerformanceAnalyzer) detectErrorRateAnomaly(snapshots []*MetricsSnapshot) *PerformanceAnomaly {
	if len(snapshots) < 3 {
		return nil
	}

	latest := snapshots[len(snapshots)-1]

	// Calculate baseline from previous snapshots
	var baselineErrorRate float64
	count := 0
	for i := 0; i < len(snapshots)-1; i++ {
		if snapshots[i].JobCount > 0 {
			baselineErrorRate += snapshots[i].ErrorRate
			count++
		}
	}

	if count == 0 {
		return nil
	}

	baselineErrorRate /= float64(count)

	// Check for significant increase
	threshold := 5.0 // 5 percentage points
	if latest.ErrorRate-baselineErrorRate > threshold {
		severity := "warning"
		if latest.ErrorRate-baselineErrorRate > 10.0 {
			severity = "critical"
		}

		return &PerformanceAnomaly{
			Type:        "error_rate_spike",
			Severity:    severity,
			Description: fmt.Sprintf("Error rate increased from %.2f%% to %.2f%%", baselineErrorRate, latest.ErrorRate),
			Value:       latest.ErrorRate,
			Threshold:   baselineErrorRate + threshold,
			Timestamp:   latest.Timestamp,
		}
	}

	return nil
}

func (pa *PerformanceAnalyzer) detectLatencyAnomaly(snapshots []*MetricsSnapshot) *PerformanceAnomaly {
	if len(snapshots) < 3 {
		return nil
	}

	latest := snapshots[len(snapshots)-1]

	// Calculate baseline from previous snapshots
	var baselineLatency float64
	count := 0
	for i := 0; i < len(snapshots)-1; i++ {
		if snapshots[i].JobCount > 0 {
			baselineLatency += snapshots[i].P95Latency
			count++
		}
	}

	if count == 0 || baselineLatency == 0 {
		return nil
	}

	baselineLatency /= float64(count)

	// Check for significant increase (more than 50%)
	threshold := 0.5
	increase := (latest.P95Latency - baselineLatency) / baselineLatency

	if increase > threshold {
		severity := "warning"
		if increase > 1.0 { // 100% increase
			severity = "critical"
		}

		return &PerformanceAnomaly{
			Type:        "latency_spike",
			Severity:    severity,
			Description: fmt.Sprintf("P95 latency increased from %.2fms to %.2fms (%.1f%% increase)", baselineLatency, latest.P95Latency, increase*100),
			Value:       latest.P95Latency,
			Threshold:   baselineLatency * (1 + threshold),
			Timestamp:   latest.Timestamp,
		}
	}

	return nil
}

func (pa *PerformanceAnalyzer) detectThroughputAnomaly(snapshots []*MetricsSnapshot) *PerformanceAnomaly {
	if len(snapshots) < 3 {
		return nil
	}

	latest := snapshots[len(snapshots)-1]

	// Calculate baseline from previous snapshots
	var baselineThroughput float64
	count := 0
	for i := 0; i < len(snapshots)-1; i++ {
		if snapshots[i].JobCount > 0 {
			baselineThroughput += snapshots[i].JobsPerSecond
			count++
		}
	}

	if count == 0 || baselineThroughput == 0 {
		return nil
	}

	baselineThroughput /= float64(count)

	// Check for significant decrease (more than 30%)
	threshold := 0.3
	decrease := (baselineThroughput - latest.JobsPerSecond) / baselineThroughput

	if decrease > threshold {
		severity := "warning"
		if decrease > 0.5 { // 50% decrease
			severity = "critical"
		}

		return &PerformanceAnomaly{
			Type:        "throughput_drop",
			Severity:    severity,
			Description: fmt.Sprintf("Throughput decreased from %.2f jobs/sec to %.2f jobs/sec (%.1f%% decrease)", baselineThroughput, latest.JobsPerSecond, decrease*100),
			Value:       latest.JobsPerSecond,
			Threshold:   baselineThroughput * (1 - threshold),
			Timestamp:   latest.Timestamp,
		}
	}

	return nil
}
