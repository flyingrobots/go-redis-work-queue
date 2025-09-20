//go:build integration_tests
// +build integration_tests

package integration

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storage "github.com/flyingrobots/go-redis-work-queue/internal/storage-backends"
)

// BenchmarkConfig holds benchmark configuration
type BenchmarkConfig struct {
	NumJobs        int
	NumWorkers     int
	JobSizeBytes   int
	TestDuration   time.Duration
	WarmupDuration time.Duration
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	Backend           string                `json:"backend"`
	Config            BenchmarkConfig       `json:"config"`
	EnqueueOpsPerSec  float64               `json:"enqueue_ops_per_sec"`
	DequeueOpsPerSec  float64               `json:"dequeue_ops_per_sec"`
	AvgLatencyMs      float64               `json:"avg_latency_ms"`
	P95LatencyMs      float64               `json:"p95_latency_ms"`
	P99LatencyMs      float64               `json:"p99_latency_ms"`
	ErrorRate         float64               `json:"error_rate"`
	ThroughputMBps    float64               `json:"throughput_mbps"`
	ConcurrentWorkers int                   `json:"concurrent_workers"`
	TotalJobs         int                   `json:"total_jobs"`
	TestDuration      time.Duration         `json:"test_duration"`
	BackendStats      *storage.BackendStats `json:"backend_stats,omitempty"`
}

// TestStorageBackendPerformance runs comprehensive performance tests
func TestStorageBackendPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	configs := []BenchmarkConfig{
		{
			NumJobs:        1000,
			NumWorkers:     1,
			JobSizeBytes:   100,
			TestDuration:   10 * time.Second,
			WarmupDuration: 2 * time.Second,
		},
		{
			NumJobs:        5000,
			NumWorkers:     5,
			JobSizeBytes:   500,
			TestDuration:   15 * time.Second,
			WarmupDuration: 3 * time.Second,
		},
		{
			NumJobs:        10000,
			NumWorkers:     10,
			JobSizeBytes:   1024,
			TestDuration:   20 * time.Second,
			WarmupDuration: 5 * time.Second,
		},
	}

	for _, config := range configs {
		t.Run(fmt.Sprintf("Performance_%dJobs_%dWorkers_%dBytes", config.NumJobs, config.NumWorkers, config.JobSizeBytes), func(t *testing.T) {
			result := runPerformanceBenchmark(t, config)
			validatePerformanceResult(t, result, config)
			logPerformanceResult(t, result)
		})
	}
}

// TestThroughputBenchmark measures sustained throughput
func TestThroughputBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping throughput tests in short mode")
	}

	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()
	duration := 30 * time.Second
	warmup := 5 * time.Second

	t.Logf("Starting throughput benchmark for %v (warmup: %v)", duration, warmup)

	// Warmup phase
	warmupCtx, warmupCancel := context.WithTimeout(ctx, warmup)
	defer warmupCancel()

	runThroughputPhase(t, backend, warmupCtx, "warmup", 100, 2)

	// Main benchmark phase
	benchCtx, benchCancel := context.WithTimeout(ctx, duration)
	defer benchCancel()

	result := runThroughputPhase(t, backend, benchCtx, "benchmark", 1000, 5)

	// Validate results
	assert.Greater(t, result.EnqueueOpsPerSec, 100.0, "Enqueue throughput should be > 100 ops/sec")
	assert.Greater(t, result.DequeueOpsPerSec, 100.0, "Dequeue throughput should be > 100 ops/sec")
	assert.Less(t, result.ErrorRate, 0.01, "Error rate should be < 1%")

	t.Logf("Throughput Results: Enqueue=%.2f ops/sec, Dequeue=%.2f ops/sec, ErrorRate=%.4f%%",
		result.EnqueueOpsPerSec, result.DequeueOpsPerSec, result.ErrorRate*100)
}

// TestLatencyBenchmark measures operation latencies
func TestLatencyBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency tests in short mode")
	}

	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()
	numOperations := 1000

	// Measure enqueue latencies
	enqueueLatencies := measureEnqueueLatencies(t, backend, ctx, numOperations)
	enqueueStats := calculateLatencyStats(enqueueLatencies)

	// Measure dequeue latencies
	dequeueLatencies := measureDequeueLatencies(t, backend, ctx, numOperations)
	dequeueStats := calculateLatencyStats(dequeueLatencies)

	// Validate latencies
	assert.Less(t, enqueueStats.P99, 100*time.Millisecond, "Enqueue P99 latency should be < 100ms")
	assert.Less(t, dequeueStats.P99, 100*time.Millisecond, "Dequeue P99 latency should be < 100ms")
	assert.Less(t, enqueueStats.Average, 50*time.Millisecond, "Average enqueue latency should be < 50ms")
	assert.Less(t, dequeueStats.Average, 50*time.Millisecond, "Average dequeue latency should be < 50ms")

	t.Logf("Enqueue Latencies: avg=%.2fms, p95=%.2fms, p99=%.2fms",
		float64(enqueueStats.Average.Nanoseconds())/1e6,
		float64(enqueueStats.P95.Nanoseconds())/1e6,
		float64(enqueueStats.P99.Nanoseconds())/1e6)

	t.Logf("Dequeue Latencies: avg=%.2fms, p95=%.2fms, p99=%.2fms",
		float64(dequeueStats.Average.Nanoseconds())/1e6,
		float64(dequeueStats.P95.Nanoseconds())/1e6,
		float64(dequeueStats.P99.Nanoseconds())/1e6)
}

// TestConcurrencyBenchmark tests concurrent access patterns
func TestConcurrencyBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency tests in short mode")
	}

	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()
	concurrencyLevels := []int{1, 2, 5, 10, 20}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			result := runConcurrencyBenchmark(t, backend, ctx, concurrency, 1000, 10*time.Second)

			// Validate concurrency scaling
			assert.Greater(t, result.EnqueueOpsPerSec, float64(concurrency)*10,
				"Throughput should scale with concurrency")
			assert.Less(t, result.ErrorRate, 0.05, "Error rate should be < 5% under concurrency")

			t.Logf("Concurrency %d: %.2f enqueue ops/sec, %.2f dequeue ops/sec, %.2f%% errors",
				concurrency, result.EnqueueOpsPerSec, result.DequeueOpsPerSec, result.ErrorRate*100)
		})
	}
}

// TestMemoryUsageBenchmark measures memory efficiency
func TestMemoryUsageBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory tests in short mode")
	}

	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()
	jobSizes := []int{100, 1024, 10240, 102400} // 100B, 1KB, 10KB, 100KB

	for _, jobSize := range jobSizes {
		t.Run(fmt.Sprintf("JobSize_%dBytes", jobSize), func(t *testing.T) {
			result := runMemoryBenchmark(t, backend, ctx, jobSize, 1000)

			// Log memory usage patterns
			if result.BackendStats != nil && result.BackendStats.MemoryUsage != nil {
				memoryPerJobKB := float64(*result.BackendStats.MemoryUsage) / float64(1000) / 1024
				t.Logf("Job size: %d bytes, Memory per job: %.2f KB, Efficiency: %.2f%%",
					jobSize, memoryPerJobKB, float64(jobSize)/float64(memoryPerJobKB*1024)*100)
			}

			// Validate reasonable memory usage
			assert.Greater(t, result.ThroughputMBps, 0.1, "Should achieve minimum throughput")
		})
	}
}

// TestStressTest runs extended stress testing
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()
	duration := 2 * time.Minute
	numWorkers := 10
	jobsPerSecond := 100

	t.Logf("Starting stress test for %v with %d workers at %d jobs/sec", duration, numWorkers, jobsPerSecond)

	result := runStressTest(t, backend, ctx, duration, numWorkers, jobsPerSecond)

	// Validate system stability under stress
	assert.Less(t, result.ErrorRate, 0.02, "Error rate should be < 2% under stress")
	assert.Greater(t, result.EnqueueOpsPerSec, float64(jobsPerSecond)*0.8,
		"Should maintain 80% of target throughput")

	// Verify backend health after stress
	health := backend.Health(ctx)
	assert.Equal(t, "healthy", health.Status, "Backend should remain healthy after stress test")

	t.Logf("Stress test completed: %.2f ops/sec, %.4f%% errors, backend status: %s",
		result.EnqueueOpsPerSec, result.ErrorRate*100, health.Status)
}

// Helper functions

func setupTestBackend(t *testing.T) storage.QueueBackend {
	// Start in-memory Redis
	miniRedis := miniredis.NewMiniRedis()
	err := miniRedis.Start()
	require.NoError(t, err)

	// Create Redis Lists backend
	config := storage.RedisListsConfig{
		URL:       "redis://" + miniRedis.Addr(),
		Database:  0,
		KeyPrefix: "perftest:",
	}

	backend, err := createRedisListsBackend(config, "performance-test")
	require.NoError(t, err)

	// Clean up function
	t.Cleanup(func() {
		backend.Close()
		miniRedis.Close()
	})

	return backend
}

func runPerformanceBenchmark(t *testing.T, config BenchmarkConfig) BenchmarkResult {
	backend := setupTestBackend(t)
	defer backend.Close()

	ctx := context.Background()

	// Warmup phase
	if config.WarmupDuration > 0 {
		warmupCtx, cancel := context.WithTimeout(ctx, config.WarmupDuration)
		runThroughputPhase(t, backend, warmupCtx, "warmup", config.NumJobs/5, config.NumWorkers)
		cancel()
	}

	// Main benchmark
	benchCtx, cancel := context.WithTimeout(ctx, config.TestDuration)
	defer cancel()

	result := runThroughputPhase(t, backend, benchCtx, "benchmark", config.NumJobs, config.NumWorkers)

	// Collect backend stats
	stats, err := backend.Stats(ctx)
	if err == nil {
		result.BackendStats = stats
	}

	result.Config = config
	return result
}

func runThroughputPhase(t *testing.T, backend storage.QueueBackend, ctx context.Context, phase string, numJobs, numWorkers int) BenchmarkResult {
	var wg sync.WaitGroup
	var enqueueOps, dequeueOps, errors int64
	var totalLatency time.Duration
	var latencyMutex sync.Mutex

	startTime := time.Now()

	// Producer workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			jobCount := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					job := generateTestJob(fmt.Sprintf("%s-worker-%d-job-%d", phase, workerID, jobCount), 500)

					opStart := time.Now()
					err := backend.Enqueue(ctx, job)
					opLatency := time.Since(opStart)

					latencyMutex.Lock()
					totalLatency += opLatency
					latencyMutex.Unlock()

					if err != nil {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&enqueueOps, 1)
					}

					jobCount++
					if jobCount >= numJobs/numWorkers {
						return
					}
				}
			}
		}(i)
	}

	// Consumer workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := backend.Dequeue(ctx, storage.DequeueOptions{
						Timeout: 100 * time.Millisecond,
						Count:   1,
					})

					if err == nil {
						atomic.AddInt64(&dequeueOps, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	avgLatency := float64(totalLatency.Nanoseconds()) / float64(enqueueOps) / 1e6 // Convert to ms

	return BenchmarkResult{
		Backend:           "redis_lists",
		EnqueueOpsPerSec:  float64(enqueueOps) / elapsed.Seconds(),
		DequeueOpsPerSec:  float64(dequeueOps) / elapsed.Seconds(),
		AvgLatencyMs:      avgLatency,
		ErrorRate:         float64(errors) / float64(enqueueOps+errors),
		ConcurrentWorkers: numWorkers,
		TotalJobs:         int(enqueueOps),
		TestDuration:      elapsed,
	}
}

func measureEnqueueLatencies(t *testing.T, backend storage.QueueBackend, ctx context.Context, numOps int) []time.Duration {
	latencies := make([]time.Duration, numOps)

	for i := 0; i < numOps; i++ {
		job := generateTestJob(fmt.Sprintf("latency-test-%d", i), 200)

		start := time.Now()
		err := backend.Enqueue(ctx, job)
		latencies[i] = time.Since(start)

		require.NoError(t, err, "Enqueue should not fail during latency measurement")
	}

	return latencies
}

func measureDequeueLatencies(t *testing.T, backend storage.QueueBackend, ctx context.Context, numOps int) []time.Duration {
	// First, enqueue jobs to dequeue
	for i := 0; i < numOps; i++ {
		job := generateTestJob(fmt.Sprintf("dequeue-latency-test-%d", i), 200)
		err := backend.Enqueue(ctx, job)
		require.NoError(t, err)
	}

	latencies := make([]time.Duration, numOps)

	for i := 0; i < numOps; i++ {
		start := time.Now()
		_, err := backend.Dequeue(ctx, storage.DequeueOptions{
			Timeout: 1 * time.Second,
			Count:   1,
		})
		latencies[i] = time.Since(start)

		require.NoError(t, err, "Dequeue should not fail during latency measurement")
	}

	return latencies
}

func runConcurrencyBenchmark(t *testing.T, backend storage.QueueBackend, ctx context.Context, concurrency, jobsPerWorker int, duration time.Duration) BenchmarkResult {
	benchCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	return runThroughputPhase(t, backend, benchCtx, "concurrency", jobsPerWorker*concurrency, concurrency)
}

func runMemoryBenchmark(t *testing.T, backend storage.QueueBackend, ctx context.Context, jobSize, numJobs int) BenchmarkResult {
	var totalBytes int64

	start := time.Now()

	for i := 0; i < numJobs; i++ {
		job := generateTestJob(fmt.Sprintf("memory-test-%d", i), jobSize)
		err := backend.Enqueue(ctx, job)
		require.NoError(t, err)

		totalBytes += int64(jobSize)
	}

	elapsed := time.Since(start)
	throughputMBps := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024

	stats, _ := backend.Stats(ctx)

	return BenchmarkResult{
		Backend:        "redis_lists",
		ThroughputMBps: throughputMBps,
		TotalJobs:      numJobs,
		TestDuration:   elapsed,
		BackendStats:   stats,
	}
}

func runStressTest(t *testing.T, backend storage.QueueBackend, ctx context.Context, duration time.Duration, numWorkers, targetOpsPerSec int) BenchmarkResult {
	stressCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Calculate job rate per worker
	jobsPerWorker := int(duration.Seconds()) * targetOpsPerSec / numWorkers

	return runThroughputPhase(t, backend, stressCtx, "stress", jobsPerWorker*numWorkers, numWorkers)
}

func generateTestJob(id string, payloadSize int) *storage.Job {
	// Generate random payload of specified size
	payload := make([]byte, payloadSize)
	rand.Read(payload)

	return &storage.Job{
		ID:        id,
		Type:      "performance-test",
		Queue:     "perf-test-queue",
		Payload:   string(payload),
		Priority:  rand.Intn(10),
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"test_type": "performance",
			"size":      payloadSize,
		},
		Tags: []string{"performance", "benchmark"},
	}
}

type LatencyStats struct {
	Average time.Duration
	P50     time.Duration
	P95     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
}

func calculateLatencyStats(latencies []time.Duration) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Sort latencies
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var total time.Duration
	for _, lat := range latencies {
		total += lat
	}

	return LatencyStats{
		Average: total / time.Duration(len(latencies)),
		P50:     latencies[len(latencies)*50/100],
		P95:     latencies[len(latencies)*95/100],
		P99:     latencies[len(latencies)*99/100],
		Min:     latencies[0],
		Max:     latencies[len(latencies)-1],
	}
}

func validatePerformanceResult(t *testing.T, result BenchmarkResult, config BenchmarkConfig) {
	// Minimum performance thresholds
	assert.Greater(t, result.EnqueueOpsPerSec, 50.0, "Enqueue throughput too low")
	assert.Greater(t, result.DequeueOpsPerSec, 50.0, "Dequeue throughput too low")
	assert.Less(t, result.ErrorRate, 0.05, "Error rate too high")
	assert.Less(t, result.AvgLatencyMs, 100.0, "Average latency too high")
}

func logPerformanceResult(t *testing.T, result BenchmarkResult) {
	t.Logf("Performance Results:")
	t.Logf("  Backend: %s", result.Backend)
	t.Logf("  Workers: %d", result.ConcurrentWorkers)
	t.Logf("  Duration: %v", result.TestDuration)
	t.Logf("  Enqueue: %.2f ops/sec", result.EnqueueOpsPerSec)
	t.Logf("  Dequeue: %.2f ops/sec", result.DequeueOpsPerSec)
	t.Logf("  Avg Latency: %.2f ms", result.AvgLatencyMs)
	t.Logf("  Error Rate: %.4f%%", result.ErrorRate*100)
	t.Logf("  Total Jobs: %d", result.TotalJobs)
}

// Helper function to create Redis Lists backend (copy from unit tests)
func createRedisListsBackend(config storage.RedisListsConfig, queueName string) (storage.QueueBackend, error) {
	// Implementation details would be in the actual storage package
	// This is a placeholder for the test
	return nil, fmt.Errorf("not implemented in test environment")
}
