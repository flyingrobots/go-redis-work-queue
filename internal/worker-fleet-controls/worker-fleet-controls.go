package workerfleetcontrols

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type WorkerFleetManagerImpl struct {
	redis         *redis.Client
	registry      WorkerRegistry
	controller    WorkerController
	signalHandler WorkerSignalHandler
	auditLogger   AuditLogger
	safetyChecker SafetyChecker
	config        Config
	logger        *slog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func NewWorkerFleetManager(redisClient *redis.Client, config Config, logger *slog.Logger) *WorkerFleetManagerImpl {
	ctx, cancel := context.WithCancel(context.Background())

	registry := NewRedisWorkerRegistry(redisClient, config, logger)
	signalHandler := NewRedisSignalHandler(redisClient, config, logger)
	auditLogger := NewRedisAuditLogger(redisClient, config, logger)
	safetyChecker := NewSafetyChecker(registry, config, logger)
	controller := NewWorkerController(registry, signalHandler, auditLogger, safetyChecker, config, logger)

	return &WorkerFleetManagerImpl{
		redis:         redisClient,
		registry:      registry,
		controller:    controller,
		signalHandler: signalHandler,
		auditLogger:   auditLogger,
		safetyChecker: safetyChecker,
		config:        config,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (m *WorkerFleetManagerImpl) Registry() WorkerRegistry {
	return m.registry
}

func (m *WorkerFleetManagerImpl) Controller() WorkerController {
	return m.controller
}

func (m *WorkerFleetManagerImpl) SignalHandler() WorkerSignalHandler {
	return m.signalHandler
}

func (m *WorkerFleetManagerImpl) AuditLogger() AuditLogger {
	return m.auditLogger
}

func (m *WorkerFleetManagerImpl) SafetyChecker() SafetyChecker {
	return m.safetyChecker
}

func (m *WorkerFleetManagerImpl) Start() error {
	m.logger.Info("Starting Worker Fleet Manager")

	m.wg.Add(1)
	go m.heartbeatMonitor()

	m.wg.Add(1)
	go m.cleanupRoutine()

	return nil
}

func (m *WorkerFleetManagerImpl) Stop() error {
	m.logger.Info("Stopping Worker Fleet Manager")
	m.cancel()
	m.wg.Wait()
	return nil
}

func (m *WorkerFleetManagerImpl) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.redis.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

func (m *WorkerFleetManagerImpl) heartbeatMonitor() {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkOfflineWorkers()
		}
	}
}

func (m *WorkerFleetManagerImpl) checkOfflineWorkers() {
	cutoff := time.Now().Add(-m.config.HeartbeatTimeout)

	filter := WorkerFilter{
		MaxHeartbeat: &cutoff,
		States:       []WorkerState{WorkerStateRunning, WorkerStatePaused, WorkerStateDraining},
	}

	request := WorkerListRequest{
		Filter: filter,
		Pagination: Pagination{
			Page:     1,
			PageSize: 1000,
		},
	}

	response, err := m.registry.ListWorkers(request)
	if err != nil {
		m.logger.Error("Failed to list workers for offline check", "error", err)
		return
	}

	for _, worker := range response.Workers {
		err := m.registry.SetWorkerState(worker.ID, WorkerStateOffline)
		if err != nil {
			m.logger.Error("Failed to mark worker as offline", "worker_id", worker.ID, "error", err)
		} else {
			m.logger.Info("Marked worker as offline", "worker_id", worker.ID, "last_heartbeat", worker.LastHeartbeat)
		}
	}
}

func (m *WorkerFleetManagerImpl) cleanupRoutine() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupOldAuditLogs()
		}
	}
}

func (m *WorkerFleetManagerImpl) cleanupOldAuditLogs() {
	cutoff := time.Now().Add(-m.config.AuditLogRetention)

	filter := AuditLogFilter{
		EndTime: &cutoff,
		Limit:   1000,
	}

	logs, err := m.auditLogger.GetAuditLogs(filter)
	if err != nil {
		m.logger.Error("Failed to get old audit logs for cleanup", "error", err)
		return
	}

	m.logger.Info("Cleaning up old audit logs", "count", len(logs), "cutoff", cutoff)
}

type RedisWorkerRegistry struct {
	redis  *redis.Client
	config Config
	logger *slog.Logger
}

func NewRedisWorkerRegistry(redisClient *redis.Client, config Config, logger *slog.Logger) *RedisWorkerRegistry {
	return &RedisWorkerRegistry{
		redis:  redisClient,
		config: config,
		logger: logger,
	}
}

func (r *RedisWorkerRegistry) RegisterWorker(worker *Worker) error {
	ctx := context.Background()

	worker.LastHeartbeat = time.Now()
	if worker.StartedAt.IsZero() {
		worker.StartedAt = time.Now()
	}
	if worker.State == "" {
		worker.State = WorkerStateRunning
	}

	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker: %w", err)
	}

	key := fmt.Sprintf("worker:registry:%s", worker.ID)
	err = r.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	err = r.redis.SAdd(ctx, "workers:active", worker.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to add worker to active set: %w", err)
	}

	r.logger.Info("Worker registered", "worker_id", worker.ID, "state", worker.State)
	return nil
}

func (r *RedisWorkerRegistry) UpdateWorker(workerID string, updates *Worker) error {
	ctx := context.Background()

	existing, err := r.GetWorker(workerID)
	if err != nil {
		return fmt.Errorf("failed to get existing worker: %w", err)
	}

	if updates.State != "" {
		existing.State = updates.State
	}
	if !updates.LastHeartbeat.IsZero() {
		existing.LastHeartbeat = updates.LastHeartbeat
	}
	if updates.CurrentJob != nil {
		existing.CurrentJob = updates.CurrentJob
	}
	if updates.Stats.JobsProcessed > 0 {
		existing.Stats = updates.Stats
	}
	if updates.Health.Status != "" {
		existing.Health = updates.Health
	}

	data, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("failed to marshal updated worker: %w", err)
	}

	key := fmt.Sprintf("worker:registry:%s", workerID)
	err = r.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update worker: %w", err)
	}

	return nil
}

func (r *RedisWorkerRegistry) GetWorker(workerID string) (*Worker, error) {
	ctx := context.Background()

	key := fmt.Sprintf("worker:registry:%s", workerID)
	data, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("worker not found: %s", workerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get worker: %w", err)
	}

	var worker Worker
	err = json.Unmarshal([]byte(data), &worker)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker: %w", err)
	}

	return &worker, nil
}

func (r *RedisWorkerRegistry) ListWorkers(request WorkerListRequest) (*WorkerListResponse, error) {
	ctx := context.Background()

	workerIDs, err := r.redis.SMembers(ctx, "workers:active").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active workers: %w", err)
	}

	workers := make([]Worker, 0)
	for _, workerID := range workerIDs {
		worker, err := r.GetWorker(workerID)
		if err != nil {
			r.logger.Warn("Failed to get worker details", "worker_id", workerID, "error", err)
			continue
		}
		workers = append(workers, *worker)
	}

	filteredWorkers := r.applyFilter(workers, request.Filter)
	sortedWorkers := r.applySorting(filteredWorkers, request.SortBy, request.SortOrder)

	totalCount := len(sortedWorkers)

	if request.Pagination.PageSize <= 0 {
		request.Pagination.PageSize = 50
	}
	if request.Pagination.Page <= 0 {
		request.Pagination.Page = 1
	}

	start := (request.Pagination.Page - 1) * request.Pagination.PageSize
	end := start + request.Pagination.PageSize

	if start >= totalCount {
		sortedWorkers = []Worker{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		sortedWorkers = sortedWorkers[start:end]
	}

	totalPages := (totalCount + request.Pagination.PageSize - 1) / request.Pagination.PageSize
	hasNext := request.Pagination.Page < totalPages
	hasPrev := request.Pagination.Page > 1

	summary := r.generateFleetSummary(workers)

	return &WorkerListResponse{
		Workers:     sortedWorkers,
		TotalCount:  totalCount,
		Page:        request.Pagination.Page,
		PageSize:    request.Pagination.PageSize,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrev,
		Filter:      request.Filter,
		Summary:     summary,
	}, nil
}

func (r *RedisWorkerRegistry) RemoveWorker(workerID string) error {
	ctx := context.Background()

	key := fmt.Sprintf("worker:registry:%s", workerID)
	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove worker: %w", err)
	}

	err = r.redis.SRem(ctx, "workers:active", workerID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove worker from active set: %w", err)
	}

	r.logger.Info("Worker removed", "worker_id", workerID)
	return nil
}

func (r *RedisWorkerRegistry) UpdateHeartbeat(workerID string, heartbeat time.Time, currentJob *ActiveJob) error {
	worker, err := r.GetWorker(workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker for heartbeat update: %w", err)
	}

	worker.LastHeartbeat = heartbeat
	worker.CurrentJob = currentJob

	return r.UpdateWorker(workerID, worker)
}

func (r *RedisWorkerRegistry) GetFleetSummary() (*FleetSummary, error) {
	request := WorkerListRequest{
		Pagination: Pagination{
			Page:     1,
			PageSize: 10000,
		},
	}

	response, err := r.ListWorkers(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers for summary: %w", err)
	}

	return &response.Summary, nil
}

func (r *RedisWorkerRegistry) SetWorkerState(workerID string, state WorkerState) error {
	worker, err := r.GetWorker(workerID)
	if err != nil {
		return fmt.Errorf("failed to get worker for state update: %w", err)
	}

	worker.State = state
	return r.UpdateWorker(workerID, worker)
}

func (r *RedisWorkerRegistry) GetWorkersByState(state WorkerState) ([]Worker, error) {
	filter := WorkerFilter{
		States: []WorkerState{state},
	}

	request := WorkerListRequest{
		Filter: filter,
		Pagination: Pagination{
			Page:     1,
			PageSize: 10000,
		},
	}

	response, err := r.ListWorkers(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers by state: %w", err)
	}

	return response.Workers, nil
}

func (r *RedisWorkerRegistry) applyFilter(workers []Worker, filter WorkerFilter) []Worker {
	if r.isFilterEmpty(filter) {
		return workers
	}

	filtered := make([]Worker, 0)
	for _, worker := range workers {
		if r.matchesFilter(worker, filter) {
			filtered = append(filtered, worker)
		}
	}
	return filtered
}

func (r *RedisWorkerRegistry) isFilterEmpty(filter WorkerFilter) bool {
	return len(filter.States) == 0 &&
		len(filter.Labels) == 0 &&
		len(filter.Capabilities) == 0 &&
		len(filter.HealthStatus) == 0 &&
		filter.MinHeartbeat == nil &&
		filter.MaxHeartbeat == nil &&
		filter.HasCurrentJob == nil &&
		filter.Version == "" &&
		filter.Hostname == ""
}

func (r *RedisWorkerRegistry) matchesFilter(worker Worker, filter WorkerFilter) bool {
	if len(filter.States) > 0 {
		found := false
		for _, state := range filter.States {
			if worker.State == state {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.HealthStatus) > 0 {
		found := false
		for _, status := range filter.HealthStatus {
			if worker.Health.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if filter.MinHeartbeat != nil && worker.LastHeartbeat.Before(*filter.MinHeartbeat) {
		return false
	}

	if filter.MaxHeartbeat != nil && worker.LastHeartbeat.After(*filter.MaxHeartbeat) {
		return false
	}

	if filter.HasCurrentJob != nil {
		hasJob := worker.CurrentJob != nil
		if *filter.HasCurrentJob != hasJob {
			return false
		}
	}

	if filter.Version != "" && worker.Version != filter.Version {
		return false
	}

	if filter.Hostname != "" && worker.Hostname != filter.Hostname {
		return false
	}

	for key, value := range filter.Labels {
		if worker.Labels[key] != value {
			return false
		}
	}

	if len(filter.Capabilities) > 0 {
		for _, requiredCap := range filter.Capabilities {
			found := false
			for _, workerCap := range worker.Capabilities {
				if workerCap == requiredCap {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

func (r *RedisWorkerRegistry) applySorting(workers []Worker, sortBy string, order SortOrder) []Worker {
	if sortBy == "" {
		sortBy = "last_heartbeat"
	}

	sort.Slice(workers, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "id":
			less = workers[i].ID < workers[j].ID
		case "state":
			less = workers[i].State < workers[j].State
		case "last_heartbeat":
			less = workers[i].LastHeartbeat.Before(workers[j].LastHeartbeat)
		case "started_at":
			less = workers[i].StartedAt.Before(workers[j].StartedAt)
		case "hostname":
			less = workers[i].Hostname < workers[j].Hostname
		case "version":
			less = workers[i].Version < workers[j].Version
		case "jobs_processed":
			less = workers[i].Stats.JobsProcessed < workers[j].Stats.JobsProcessed
		case "health":
			less = workers[i].Health.Status < workers[j].Health.Status
		default:
			less = workers[i].LastHeartbeat.Before(workers[j].LastHeartbeat)
		}

		if order == SortOrderDesc {
			return !less
		}
		return less
	})

	return workers
}

func (r *RedisWorkerRegistry) generateFleetSummary(workers []Worker) FleetSummary {
	summary := FleetSummary{
		TotalWorkers:       len(workers),
		StateDistribution:  make(map[WorkerState]int),
		HealthDistribution: make(map[HealthStatus]int),
		UpdatedAt:          time.Now(),
	}

	var totalLoad float64
	activeJobs := 0

	for _, worker := range workers {
		summary.StateDistribution[worker.State]++
		summary.HealthDistribution[worker.Health.Status]++

		if worker.CurrentJob != nil {
			activeJobs++
		}

		totalLoad += worker.Stats.CPUUsage
	}

	summary.ActiveJobs = activeJobs
	if len(workers) > 0 {
		summary.AverageLoad = totalLoad / float64(len(workers))
	}

	return summary
}
