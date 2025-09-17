package canary_deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// WorkerRegistry manages worker registration and health tracking
type WorkerRegistry struct {
	redis   *redis.Client
	logger  *slog.Logger
	workers map[string]*WorkerInfo
	mu      sync.RWMutex

	// Health monitoring
	healthCheckInterval time.Duration
	workerTimeout       time.Duration
}

// NewWorkerRegistry creates a new worker registry
func NewWorkerRegistry(redis *redis.Client, logger *slog.Logger) *WorkerRegistry {
	return &WorkerRegistry{
		redis:               redis,
		logger:              logger,
		workers:             make(map[string]*WorkerInfo),
		healthCheckInterval: 30 * time.Second,
		workerTimeout:       2 * time.Minute,
	}
}

// RegisterWorker registers a new worker or updates an existing one
func (wr *WorkerRegistry) RegisterWorker(info *WorkerInfo) error {
	if info.ID == "" {
		return NewValidationError("worker_id", "worker ID is required")
	}

	if info.Version == "" {
		return NewValidationError("version", "worker version is required")
	}

	if info.Lane == "" {
		info.Lane = "stable" // Default to stable lane
	}

	info.LastSeen = time.Now()
	info.Status = WorkerHealthy

	wr.mu.Lock()
	wr.workers[info.ID] = info
	wr.mu.Unlock()

	// Persist to Redis
	if err := wr.saveWorkerToRedis(info); err != nil {
		return fmt.Errorf("failed to save worker to Redis: %w", err)
	}

	wr.logger.Info("Worker registered",
		"worker_id", info.ID,
		"version", info.Version,
		"lane", info.Lane,
		"queues", info.Queues)

	return nil
}

// UpdateWorkerStatus updates a worker's status
func (wr *WorkerRegistry) UpdateWorkerStatus(workerID string, status WorkerStatus) error {
	wr.mu.Lock()
	worker, exists := wr.workers[workerID]
	if !exists {
		wr.mu.Unlock()
		return NewWorkerNotFoundError(workerID)
	}

	worker.Status = status
	worker.LastSeen = time.Now()
	wr.mu.Unlock()

	// Update in Redis
	if err := wr.saveWorkerToRedis(worker); err != nil {
		return fmt.Errorf("failed to update worker in Redis: %w", err)
	}

	wr.logger.Info("Worker status updated",
		"worker_id", workerID,
		"status", status)

	return nil
}

// UpdateWorkerMetrics updates a worker's performance metrics
func (wr *WorkerRegistry) UpdateWorkerMetrics(workerID string, metrics WorkerMetrics) error {
	wr.mu.Lock()
	worker, exists := wr.workers[workerID]
	if !exists {
		wr.mu.Unlock()
		return NewWorkerNotFoundError(workerID)
	}

	worker.Metrics = metrics
	worker.LastSeen = time.Now()
	wr.mu.Unlock()

	// Update in Redis
	if err := wr.saveWorkerToRedis(worker); err != nil {
		return fmt.Errorf("failed to update worker metrics in Redis: %w", err)
	}

	wr.logger.Debug("Worker metrics updated",
		"worker_id", workerID,
		"jobs_processed", metrics.JobsProcessed,
		"success_rate", float64(metrics.JobsSucceeded)/float64(metrics.JobsProcessed)*100)

	return nil
}

// GetWorker returns information about a specific worker
func (wr *WorkerRegistry) GetWorker(workerID string) (*WorkerInfo, error) {
	wr.mu.RLock()
	worker, exists := wr.workers[workerID]
	wr.mu.RUnlock()

	if !exists {
		return nil, NewWorkerNotFoundError(workerID)
	}

	// Return a copy to prevent external modification
	return wr.copyWorkerInfo(worker), nil
}

// GetWorkersByLane returns all workers in a specific lane
func (wr *WorkerRegistry) GetWorkersByLane(lane string) []*WorkerInfo {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	workers := make([]*WorkerInfo, 0)
	for _, worker := range wr.workers {
		if worker.Lane == lane {
			workers = append(workers, wr.copyWorkerInfo(worker))
		}
	}

	return workers
}

// GetWorkersByQueue returns all workers that can process a specific queue
func (wr *WorkerRegistry) GetWorkersByQueue(queue string) []*WorkerInfo {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	workers := make([]*WorkerInfo, 0)
	for _, worker := range wr.workers {
		for _, workerQueue := range worker.Queues {
			if workerQueue == queue {
				workers = append(workers, wr.copyWorkerInfo(worker))
				break
			}
		}
	}

	return workers
}

// GetHealthyWorkers returns all healthy workers
func (wr *WorkerRegistry) GetHealthyWorkers() []*WorkerInfo {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	workers := make([]*WorkerInfo, 0)
	for _, worker := range wr.workers {
		if worker.IsHealthy() {
			workers = append(workers, wr.copyWorkerInfo(worker))
		}
	}

	return workers
}

// RemoveWorker removes a worker from the registry
func (wr *WorkerRegistry) RemoveWorker(workerID string) error {
	wr.mu.Lock()
	_, exists := wr.workers[workerID]
	if !exists {
		wr.mu.Unlock()
		return NewWorkerNotFoundError(workerID)
	}

	delete(wr.workers, workerID)
	wr.mu.Unlock()

	// Remove from Redis
	if err := wr.removeWorkerFromRedis(workerID); err != nil {
		return fmt.Errorf("failed to remove worker from Redis: %w", err)
	}

	wr.logger.Info("Worker removed", "worker_id", workerID)
	return nil
}

// StartHealthMonitoring starts the health monitoring background process
func (wr *WorkerRegistry) StartHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(wr.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wr.performHealthCheck()
		}
	}
}

// LoadWorkersFromRedis loads worker information from Redis
func (wr *WorkerRegistry) LoadWorkersFromRedis(ctx context.Context) error {
	pattern := "canary:worker:*"
	keys, err := wr.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to list worker keys: %w", err)
	}

	wr.mu.Lock()
	defer wr.mu.Unlock()

	loadedCount := 0
	for _, key := range keys {
		data, err := wr.redis.Get(ctx, key).Result()
		if err != nil {
			wr.logger.Warn("Failed to load worker", "key", key, "error", err)
			continue
		}

		var worker WorkerInfo
		if err := json.Unmarshal([]byte(data), &worker); err != nil {
			wr.logger.Warn("Failed to unmarshal worker", "key", key, "error", err)
			continue
		}

		// Check if worker is still alive based on last seen timestamp
		if time.Since(worker.LastSeen) > wr.workerTimeout {
			worker.Status = WorkerUnreachable
		}

		wr.workers[worker.ID] = &worker
		loadedCount++
	}

	wr.logger.Info("Loaded workers from Redis",
		"total_keys", len(keys),
		"loaded_count", loadedCount)

	return nil
}

// GetWorkerStatistics returns overall worker statistics
func (wr *WorkerRegistry) GetWorkerStatistics() *WorkerStatistics {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	stats := &WorkerStatistics{
		TotalWorkers: len(wr.workers),
		ByStatus:     make(map[WorkerStatus]int),
		ByLane:       make(map[string]int),
		ByVersion:    make(map[string]int),
	}

	for _, worker := range wr.workers {
		stats.ByStatus[worker.Status]++
		stats.ByLane[worker.Lane]++
		stats.ByVersion[worker.Version]++
	}

	stats.HealthyWorkers = stats.ByStatus[WorkerHealthy]
	stats.UnhealthyWorkers = stats.TotalWorkers - stats.HealthyWorkers

	return stats
}

// Private methods

func (wr *WorkerRegistry) performHealthCheck() {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	now := time.Now()
	staleWorkers := make([]string, 0)

	for workerID, worker := range wr.workers {
		timeSinceLastSeen := now.Sub(worker.LastSeen)

		if timeSinceLastSeen > wr.workerTimeout {
			if worker.Status != WorkerUnreachable {
				worker.Status = WorkerUnreachable
				wr.logger.Warn("Worker marked as unreachable",
					"worker_id", workerID,
					"last_seen", worker.LastSeen,
					"timeout", wr.workerTimeout)

				// Update in Redis
				go func(w *WorkerInfo) {
					if err := wr.saveWorkerToRedis(w); err != nil {
						wr.logger.Error("Failed to update unreachable worker status",
							"worker_id", w.ID,
							"error", err)
					}
				}(worker)
			}

			// Mark for removal if unreachable for too long
			if timeSinceLastSeen > 2*wr.workerTimeout {
				staleWorkers = append(staleWorkers, workerID)
			}
		}
	}

	// Remove stale workers
	for _, workerID := range staleWorkers {
		delete(wr.workers, workerID)
		wr.logger.Info("Removed stale worker", "worker_id", workerID)

		// Remove from Redis
		go func(id string) {
			if err := wr.removeWorkerFromRedis(id); err != nil {
				wr.logger.Error("Failed to remove stale worker from Redis",
					"worker_id", id,
					"error", err)
			}
		}(workerID)
	}
}

func (wr *WorkerRegistry) saveWorkerToRedis(worker *WorkerInfo) error {
	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker: %w", err)
	}

	key := fmt.Sprintf("canary:worker:%s", worker.ID)
	ctx := context.Background()

	if err := wr.redis.Set(ctx, key, data, 3*wr.workerTimeout).Err(); err != nil {
		return fmt.Errorf("failed to save worker to Redis: %w", err)
	}

	return nil
}

func (wr *WorkerRegistry) removeWorkerFromRedis(workerID string) error {
	key := fmt.Sprintf("canary:worker:%s", workerID)
	ctx := context.Background()

	return wr.redis.Del(ctx, key).Err()
}

func (wr *WorkerRegistry) copyWorkerInfo(worker *WorkerInfo) *WorkerInfo {
	copy := *worker

	// Deep copy slices
	if worker.Queues != nil {
		copy.Queues = make([]string, len(worker.Queues))
		for i, queue := range worker.Queues {
			copy.Queues[i] = queue
		}
	}

	if worker.Metadata != nil {
		copy.Metadata = make(map[string]string)
		for k, v := range worker.Metadata {
			copy.Metadata[k] = v
		}
	}

	return &copy
}

// WorkerStatistics represents overall worker statistics
type WorkerStatistics struct {
	TotalWorkers     int                  `json:"total_workers"`
	HealthyWorkers   int                  `json:"healthy_workers"`
	UnhealthyWorkers int                  `json:"unhealthy_workers"`
	ByStatus         map[WorkerStatus]int `json:"by_status"`
	ByLane           map[string]int       `json:"by_lane"`
	ByVersion        map[string]int       `json:"by_version"`
}

// WorkerHealthChecker provides advanced health checking capabilities
type WorkerHealthChecker struct {
	registry *WorkerRegistry
	redis    *redis.Client
	logger   *slog.Logger
}

// NewWorkerHealthChecker creates a new worker health checker
func NewWorkerHealthChecker(registry *WorkerRegistry, redis *redis.Client, logger *slog.Logger) *WorkerHealthChecker {
	return &WorkerHealthChecker{
		registry: registry,
		redis:    redis,
		logger:   logger,
	}
}

// CheckWorkerHealth performs a comprehensive health check on a worker
func (whc *WorkerHealthChecker) CheckWorkerHealth(ctx context.Context, workerID string) (*WorkerHealthReport, error) {
	worker, err := whc.registry.GetWorker(workerID)
	if err != nil {
		return nil, err
	}

	report := &WorkerHealthReport{
		WorkerID:  workerID,
		Timestamp: time.Now(),
		Checks:    make(map[string]HealthCheckResult),
	}

	// Connectivity check
	report.Checks["connectivity"] = whc.checkConnectivity(ctx, worker)

	// Performance check
	report.Checks["performance"] = whc.checkPerformance(ctx, worker)

	// Resource utilization check
	report.Checks["resources"] = whc.checkResourceUtilization(ctx, worker)

	// Queue processing check
	report.Checks["queue_processing"] = whc.checkQueueProcessing(ctx, worker)

	// Overall health determination
	report.OverallHealth = whc.determineOverallHealth(report.Checks)

	return report, nil
}

// CheckLaneHealth checks the overall health of a specific lane
func (whc *WorkerHealthChecker) CheckLaneHealth(ctx context.Context, lane string) (*LaneHealthReport, error) {
	workers := whc.registry.GetWorkersByLane(lane)

	report := &LaneHealthReport{
		Lane:          lane,
		Timestamp:     time.Now(),
		TotalWorkers:  len(workers),
		WorkerReports: make([]*WorkerHealthReport, 0),
	}

	for _, worker := range workers {
		workerReport, err := whc.CheckWorkerHealth(ctx, worker.ID)
		if err != nil {
			whc.logger.Warn("Failed to check worker health",
				"worker_id", worker.ID,
				"error", err)
			continue
		}
		report.WorkerReports = append(report.WorkerReports, workerReport)

		switch workerReport.OverallHealth {
		case "healthy":
			report.HealthyWorkers++
		case "degraded":
			report.DegradedWorkers++
		case "unhealthy":
			report.UnhealthyWorkers++
		}
	}

	// Calculate overall lane health
	if report.TotalWorkers == 0 {
		report.OverallHealth = "unknown"
	} else {
		healthyPercentage := float64(report.HealthyWorkers) / float64(report.TotalWorkers) * 100
		if healthyPercentage >= 80 {
			report.OverallHealth = "healthy"
		} else if healthyPercentage >= 50 {
			report.OverallHealth = "degraded"
		} else {
			report.OverallHealth = "unhealthy"
		}
	}

	return report, nil
}

// Private health check methods

func (whc *WorkerHealthChecker) checkConnectivity(ctx context.Context, worker *WorkerInfo) HealthCheckResult {
	timeSinceLastSeen := time.Since(worker.LastSeen)
	threshold := 2 * time.Minute

	if timeSinceLastSeen <= threshold {
		return HealthCheckResult{
			Status:  "healthy",
			Message: fmt.Sprintf("Last seen %v ago", timeSinceLastSeen.Truncate(time.Second)),
			Score:   100,
		}
	} else if timeSinceLastSeen <= 2*threshold {
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Last seen %v ago (threshold: %v)", timeSinceLastSeen.Truncate(time.Second), threshold),
			Score:   50,
		}
	} else {
		return HealthCheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Last seen %v ago (threshold: %v)", timeSinceLastSeen.Truncate(time.Second), threshold),
			Score:   0,
		}
	}
}

func (whc *WorkerHealthChecker) checkPerformance(ctx context.Context, worker *WorkerInfo) HealthCheckResult {
	metrics := worker.Metrics

	if metrics.JobsProcessed == 0 {
		return HealthCheckResult{
			Status:  "unknown",
			Message: "No jobs processed yet",
			Score:   50,
		}
	}

	successRate := float64(metrics.JobsSucceeded) / float64(metrics.JobsProcessed) * 100

	if successRate >= 95 {
		return HealthCheckResult{
			Status:  "healthy",
			Message: fmt.Sprintf("Success rate: %.1f%%", successRate),
			Score:   100,
		}
	} else if successRate >= 80 {
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Success rate: %.1f%% (below 95%%)", successRate),
			Score:   70,
		}
	} else {
		return HealthCheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Success rate: %.1f%% (below 80%%)", successRate),
			Score:   30,
		}
	}
}

func (whc *WorkerHealthChecker) checkResourceUtilization(ctx context.Context, worker *WorkerInfo) HealthCheckResult {
	metrics := worker.Metrics

	// Check memory usage
	if metrics.MemoryUsageMB > 1000 { // 1GB threshold
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("High memory usage: %.1f MB", metrics.MemoryUsageMB),
			Score:   60,
		}
	}

	// Check CPU usage
	if metrics.CPUUsagePercent > 80 {
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("High CPU usage: %.1f%%", metrics.CPUUsagePercent),
			Score:   60,
		}
	}

	return HealthCheckResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Memory: %.1f MB, CPU: %.1f%%", metrics.MemoryUsageMB, metrics.CPUUsagePercent),
		Score:   100,
	}
}

func (whc *WorkerHealthChecker) checkQueueProcessing(ctx context.Context, worker *WorkerInfo) HealthCheckResult {
	metrics := worker.Metrics

	// Check if worker is actively processing jobs
	timeSinceLastJob := time.Since(metrics.LastJobAt)
	if timeSinceLastJob > 10*time.Minute {
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("No jobs processed in %v", timeSinceLastJob.Truncate(time.Second)),
			Score:   40,
		}
	}

	// Check average processing time
	if metrics.AvgProcessingTime > 10.0 { // 10 seconds threshold
		return HealthCheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Slow processing: avg %.1fs", metrics.AvgProcessingTime),
			Score:   60,
		}
	}

	return HealthCheckResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Active processing, avg %.1fs", metrics.AvgProcessingTime),
		Score:   100,
	}
}

func (whc *WorkerHealthChecker) determineOverallHealth(checks map[string]HealthCheckResult) string {
	totalScore := 0
	checkCount := 0

	for _, check := range checks {
		if check.Status != "unknown" {
			totalScore += check.Score
			checkCount++
		}
	}

	if checkCount == 0 {
		return "unknown"
	}

	avgScore := totalScore / checkCount

	if avgScore >= 80 {
		return "healthy"
	} else if avgScore >= 50 {
		return "degraded"
	} else {
		return "unhealthy"
	}
}

// Health report types

type WorkerHealthReport struct {
	WorkerID      string                       `json:"worker_id"`
	Timestamp     time.Time                    `json:"timestamp"`
	OverallHealth string                       `json:"overall_health"`
	Checks        map[string]HealthCheckResult `json:"checks"`
}

type LaneHealthReport struct {
	Lane             string                `json:"lane"`
	Timestamp        time.Time             `json:"timestamp"`
	OverallHealth    string                `json:"overall_health"`
	TotalWorkers     int                   `json:"total_workers"`
	HealthyWorkers   int                   `json:"healthy_workers"`
	DegradedWorkers  int                   `json:"degraded_workers"`
	UnhealthyWorkers int                   `json:"unhealthy_workers"`
	WorkerReports    []*WorkerHealthReport `json:"worker_reports"`
}

type HealthCheckResult struct {
	Status  string `json:"status"` // "healthy", "degraded", "unhealthy", "unknown"
	Message string `json:"message"`
	Score   int    `json:"score"` // 0-100
}
