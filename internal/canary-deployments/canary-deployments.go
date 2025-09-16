package canary_deployments

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// Manager implements the CanaryManager interface
type Manager struct {
	config     *Config
	redis      *redis.Client
	logger     *slog.Logger

	// Internal components
	router        Router
	collector     MetricsCollector
	alerter       Alerter
	workers       *WorkerRegistry

	// State management
	deployments   map[string]*CanaryDeployment
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// Channels for internal communication
	alertChan     chan *Alert
	eventChan     chan *DeploymentEvent
}

// NewManager creates a new canary deployment manager
func NewManager(config *Config, redis *redis.Client, logger *slog.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:      config,
		redis:       redis,
		logger:      logger,
		deployments: make(map[string]*CanaryDeployment),
		ctx:         ctx,
		cancel:      cancel,
		alertChan:   make(chan *Alert, 100),
		eventChan:   make(chan *DeploymentEvent, 100),
	}

	// Initialize components
	manager.router = NewRedisRouter(redis, logger)
	manager.collector = NewRedisMetricsCollector(redis, logger)
	manager.alerter = NewWebhookAlerter(config.WebhookURLs, logger)
	manager.workers = NewWorkerRegistry(redis, logger)

	return manager
}

// Start begins the canary deployment manager
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting canary deployment manager")

	// Load existing deployments from Redis
	if err := m.loadDeployments(ctx); err != nil {
		return fmt.Errorf("failed to load deployments: %w", err)
	}

	// Start background goroutines
	m.wg.Add(4)
	go m.monitorDeployments()
	go m.processAlerts()
	go m.processEvents()
	go m.cleanupExpiredData()

	m.logger.Info("Canary deployment manager started")
	return nil
}

// Stop gracefully shuts down the canary deployment manager
func (m *Manager) Stop(ctx context.Context) error {
	m.logger.Info("Stopping canary deployment manager")

	m.cancel()

	// Wait for background goroutines to finish
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Canary deployment manager stopped gracefully")
	case <-ctx.Done():
		m.logger.Warn("Canary deployment manager shutdown timed out")
	}

	return nil
}

// CreateDeployment creates a new canary deployment
func (m *Manager) CreateDeployment(ctx context.Context, config *CanaryConfig) (*CanaryDeployment, error) {
	if err := config.Validate(); err != nil {
		return nil, NewValidationError("config", err.Error())
	}

	// Check concurrency limits
	m.mu.RLock()
	activeCount := 0
	for _, dep := range m.deployments {
		if dep.Status == StatusActive || dep.Status == StatusPromoting {
			activeCount++
		}
	}
	m.mu.RUnlock()

	if activeCount >= m.config.MaxConcurrentDeployments {
		return nil, NewConcurrencyLimitError(m.config.MaxConcurrentDeployments)
	}

	// Generate deployment ID
	deploymentID := "canary_" + uuid.New().String()

	// Create deployment object
	deployment := &CanaryDeployment{
		ID:             deploymentID,
		QueueName:      "", // Will be set by caller
		Status:         StatusActive,
		StartTime:      time.Now(),
		LastUpdate:     time.Now(),
		Config:         config,
		CurrentPercent: 0,
		TargetPercent:  5, // Start with 5%
	}

	// Store deployment
	m.mu.Lock()
	m.deployments[deploymentID] = deployment
	m.mu.Unlock()

	// Persist to Redis
	if err := m.saveDeployment(ctx, deployment); err != nil {
		m.mu.Lock()
		delete(m.deployments, deploymentID)
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}

	// Emit event
	m.emitEvent(deployment, "deployment_created", "Canary deployment created")

	m.logger.Info("Created canary deployment",
		"deployment_id", deploymentID,
		"queue", deployment.QueueName,
		"strategy", config.RoutingStrategy)

	return deployment, nil
}

// GetDeployment retrieves a deployment by ID
func (m *Manager) GetDeployment(ctx context.Context, id string) (*CanaryDeployment, error) {
	m.mu.RLock()
	deployment, exists := m.deployments[id]
	m.mu.RUnlock()

	if !exists {
		return nil, NewDeploymentNotFoundError(id)
	}

	// Return a copy to prevent external modification
	return m.copyDeployment(deployment), nil
}

// ListDeployments returns all deployments
func (m *Manager) ListDeployments(ctx context.Context) ([]*CanaryDeployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deployments := make([]*CanaryDeployment, 0, len(m.deployments))
	for _, deployment := range m.deployments {
		deployments = append(deployments, m.copyDeployment(deployment))
	}

	// Sort by start time (newest first)
	sort.Slice(deployments, func(i, j int) bool {
		return deployments[i].StartTime.After(deployments[j].StartTime)
	})

	return deployments, nil
}

// UpdateDeploymentPercentage updates the traffic split percentage
func (m *Manager) UpdateDeploymentPercentage(ctx context.Context, id string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return NewInvalidPercentageError(percentage)
	}

	m.mu.Lock()
	deployment, exists := m.deployments[id]
	if !exists {
		m.mu.Unlock()
		return NewDeploymentNotFoundError(id)
	}

	if deployment.Status != StatusActive && deployment.Status != StatusPromoting {
		m.mu.Unlock()
		return NewCanaryError(CodeDeploymentNotActive, "deployment is not active")
	}

	// Check if percentage exceeds configured maximum
	if percentage > m.config.MaxCanaryPercentage {
		m.mu.Unlock()
		return NewCanaryError(CodeInvalidPercentage,
			fmt.Sprintf("percentage exceeds maximum allowed (%d%%)", m.config.MaxCanaryPercentage))
	}

	deployment.CurrentPercent = percentage
	deployment.TargetPercent = percentage
	deployment.LastUpdate = time.Now()
	m.mu.Unlock()

	// Update routing
	if err := m.router.UpdateRoutingPercentage(ctx, deployment.QueueName, percentage); err != nil {
		return fmt.Errorf("failed to update routing: %w", err)
	}

	// Save updated deployment
	if err := m.saveDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to save deployment: %w", err)
	}

	// Emit event
	m.emitEvent(deployment, "percentage_updated",
		fmt.Sprintf("Traffic split updated to %d%%", percentage))

	m.logger.Info("Updated deployment percentage",
		"deployment_id", id,
		"percentage", percentage)

	return nil
}

// PromoteDeployment promotes a canary to 100%
func (m *Manager) PromoteDeployment(ctx context.Context, id string) error {
	m.mu.Lock()
	deployment, exists := m.deployments[id]
	if !exists {
		m.mu.Unlock()
		return NewDeploymentNotFoundError(id)
	}

	if deployment.Status != StatusActive && deployment.Status != StatusPromoting {
		m.mu.Unlock()
		return NewCanaryError(CodeDeploymentNotActive, "deployment is not active")
	}

	deployment.Status = StatusPromoting
	deployment.LastUpdate = time.Now()
	m.mu.Unlock()

	// Set to 100%
	if err := m.UpdateDeploymentPercentage(ctx, id, 100); err != nil {
		return fmt.Errorf("failed to set 100%% traffic: %w", err)
	}

	// Mark as completed
	m.mu.Lock()
	deployment.Status = StatusCompleted
	now := time.Now()
	deployment.CompletedAt = &now
	deployment.LastUpdate = now
	m.mu.Unlock()

	// Save deployment
	if err := m.saveDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to save deployment: %w", err)
	}

	// Emit event
	m.emitEvent(deployment, "deployment_promoted", "Canary deployment promoted to 100%")

	m.logger.Info("Promoted canary deployment", "deployment_id", id)
	return nil
}

// RollbackDeployment rolls back a canary to 0%
func (m *Manager) RollbackDeployment(ctx context.Context, id string, reason string) error {
	m.mu.Lock()
	deployment, exists := m.deployments[id]
	if !exists {
		m.mu.Unlock()
		return NewDeploymentNotFoundError(id)
	}

	if deployment.Status == StatusCompleted {
		m.mu.Unlock()
		return NewCanaryError(CodeDeploymentCompleted, "deployment already completed")
	}

	deployment.Status = StatusRollingBack
	deployment.LastUpdate = time.Now()
	m.mu.Unlock()

	// Set to 0%
	if err := m.router.UpdateRoutingPercentage(ctx, deployment.QueueName, 0); err != nil {
		return fmt.Errorf("failed to set 0%% traffic: %w", err)
	}

	// Drain canary jobs if using split queue strategy
	if deployment.Config.RoutingStrategy == SplitQueueStrategy {
		if err := m.drainCanaryQueue(ctx, deployment); err != nil {
			m.logger.Warn("Failed to drain canary queue", "error", err)
		}
	}

	// Mark as failed
	m.mu.Lock()
	deployment.Status = StatusFailed
	deployment.CurrentPercent = 0
	deployment.TargetPercent = 0
	now := time.Now()
	deployment.CompletedAt = &now
	deployment.LastUpdate = now
	m.mu.Unlock()

	// Save deployment
	if err := m.saveDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to save deployment: %w", err)
	}

	// Emit event
	m.emitEvent(deployment, "deployment_rolled_back",
		fmt.Sprintf("Canary deployment rolled back: %s", reason))

	// Send alert
	alert := &Alert{
		ID:           "alert_" + uuid.New().String(),
		DeploymentID: id,
		Level:        CriticalAlert,
		Message:      fmt.Sprintf("Canary deployment rolled back: %s", reason),
		Action:       NoAction,
		Timestamp:    time.Now(),
	}
	m.alertChan <- alert

	m.logger.Warn("Rolled back canary deployment",
		"deployment_id", id,
		"reason", reason)

	return nil
}

// DeleteDeployment removes a deployment
func (m *Manager) DeleteDeployment(ctx context.Context, id string) error {
	m.mu.Lock()
	deployment, exists := m.deployments[id]
	if !exists {
		m.mu.Unlock()
		return NewDeploymentNotFoundError(id)
	}

	// Can only delete completed or failed deployments
	if deployment.Status == StatusActive || deployment.Status == StatusPromoting {
		m.mu.Unlock()
		return NewCanaryError(CodeDeploymentInProgress, "cannot delete active deployment")
	}

	delete(m.deployments, id)
	m.mu.Unlock()

	// Remove from Redis
	if err := m.deleteDeploymentFromRedis(ctx, id); err != nil {
		return fmt.Errorf("failed to delete from Redis: %w", err)
	}

	m.logger.Info("Deleted canary deployment", "deployment_id", id)
	return nil
}

// GetDeploymentHealth returns the current health status of a deployment
func (m *Manager) GetDeploymentHealth(ctx context.Context, id string) (*CanaryHealthStatus, error) {
	deployment, err := m.GetDeployment(ctx, id)
	if err != nil {
		return nil, err
	}

	if deployment.Status != StatusActive && deployment.Status != StatusPromoting {
		return &CanaryHealthStatus{
			OverallStatus:  UnknownCanary,
			LastEvaluation: time.Now(),
		}, nil
	}

	// Collect current metrics
	stableMetrics, canaryMetrics, err := m.GetDeploymentMetrics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Evaluate health based on thresholds
	health := m.evaluateHealth(deployment, stableMetrics, canaryMetrics)
	return health, nil
}

// GetDeploymentMetrics returns current metrics for stable and canary versions
func (m *Manager) GetDeploymentMetrics(ctx context.Context, id string) (*MetricsSnapshot, *MetricsSnapshot, error) {
	deployment, err := m.GetDeployment(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	window := deployment.Config.MetricsWindow

	stableMetrics, err := m.collector.CollectSnapshot(ctx, deployment.QueueName, deployment.StableVersion, window)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect stable metrics: %w", err)
	}

	canaryMetrics, err := m.collector.CollectSnapshot(ctx, deployment.QueueName, deployment.CanaryVersion, window)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect canary metrics: %w", err)
	}

	return stableMetrics, canaryMetrics, nil
}

// GetDeploymentEvents returns events for a deployment
func (m *Manager) GetDeploymentEvents(ctx context.Context, id string) ([]*DeploymentEvent, error) {
	events, err := m.loadEventsFromRedis(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Sort by timestamp (newest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	return events, nil
}

// RegisterWorker registers a new worker
func (m *Manager) RegisterWorker(ctx context.Context, info *WorkerInfo) error {
	return m.workers.RegisterWorker(info)
}

// GetWorkers returns workers for a specific lane
func (m *Manager) GetWorkers(ctx context.Context, lane string) ([]*WorkerInfo, error) {
	return m.workers.GetWorkersByLane(lane), nil
}

// UpdateWorkerStatus updates a worker's status
func (m *Manager) UpdateWorkerStatus(ctx context.Context, workerID string, status WorkerStatus) error {
	return m.workers.UpdateWorkerStatus(workerID, status)
}

// Private methods

func (m *Manager) monitorDeployments() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkActiveDeployments()
		}
	}
}

func (m *Manager) checkActiveDeployments() {
	m.mu.RLock()
	activeDeployments := make([]*CanaryDeployment, 0)
	for _, deployment := range m.deployments {
		if deployment.Status == StatusActive || deployment.Status == StatusPromoting {
			activeDeployments = append(activeDeployments, m.copyDeployment(deployment))
		}
	}
	m.mu.RUnlock()

	for _, deployment := range activeDeployments {
		m.checkDeploymentHealth(deployment)
		m.checkAutoPromotion(deployment)
		m.checkTimeout(deployment)
	}
}

func (m *Manager) checkDeploymentHealth(deployment *CanaryDeployment) {
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	health, err := m.GetDeploymentHealth(ctx, deployment.ID)
	if err != nil {
		m.logger.Error("Failed to check deployment health",
			"deployment_id", deployment.ID,
			"error", err)
		return
	}

	// Check for rollback conditions
	if health.OverallStatus == FailingCanary {
		reason := health.GetFailureReason()
		if err := m.RollbackDeployment(ctx, deployment.ID, reason); err != nil {
			m.logger.Error("Failed to rollback deployment",
				"deployment_id", deployment.ID,
				"error", err)
		}
	}
}

func (m *Manager) checkAutoPromotion(deployment *CanaryDeployment) {
	if !deployment.Config.AutoPromotion {
		return
	}

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	stableMetrics, canaryMetrics, err := m.GetDeploymentMetrics(ctx, deployment.ID)
	if err != nil {
		m.logger.Error("Failed to get metrics for auto-promotion check",
			"deployment_id", deployment.ID,
			"error", err)
		return
	}

	// Check promotion stages
	for _, stage := range deployment.Config.PromotionStages {
		if deployment.CurrentPercent < stage.Percentage {
			if m.evaluatePromotionConditions(stableMetrics, canaryMetrics, stage.Conditions) {
				if err := m.UpdateDeploymentPercentage(ctx, deployment.ID, stage.Percentage); err != nil {
					m.logger.Error("Failed to auto-promote deployment",
						"deployment_id", deployment.ID,
						"target_percentage", stage.Percentage,
						"error", err)
				} else {
					m.logger.Info("Auto-promoted deployment",
						"deployment_id", deployment.ID,
						"percentage", stage.Percentage)
				}
			}
			break // Only check the next stage
		}
	}
}

func (m *Manager) checkTimeout(deployment *CanaryDeployment) {
	if time.Since(deployment.StartTime) > deployment.Config.MaxCanaryDuration {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		if err := m.RollbackDeployment(ctx, deployment.ID, "deployment timeout"); err != nil {
			m.logger.Error("Failed to rollback timed-out deployment",
				"deployment_id", deployment.ID,
				"error", err)
		}
	}
}

func (m *Manager) evaluateHealth(deployment *CanaryDeployment, stable, canary *MetricsSnapshot) *CanaryHealthStatus {
	health := &CanaryHealthStatus{
		LastEvaluation: time.Now(),
	}

	thresholds := deployment.Config.RollbackThresholds

	// Error rate check
	if stable != nil && canary != nil {
		errorRateIncrease := canary.ErrorRate - stable.ErrorRate
		health.ErrorRateCheck = HealthCheck{
			Name:      "Error Rate",
			Passing:   errorRateIncrease <= thresholds.MaxErrorRateIncrease,
			Message:   fmt.Sprintf("Error rate increase: %.2f%% (threshold: %.2f%%)", errorRateIncrease, thresholds.MaxErrorRateIncrease),
			Timestamp: time.Now(),
		}

		// Latency check
		latencyIncrease := float64(0)
		if stable.P95Latency > 0 {
			latencyIncrease = (canary.P95Latency - stable.P95Latency) / stable.P95Latency * 100
		}
		health.LatencyCheck = HealthCheck{
			Name:      "P95 Latency",
			Passing:   latencyIncrease <= thresholds.MaxLatencyIncrease,
			Message:   fmt.Sprintf("P95 latency increase: %.2f%% (threshold: %.2f%%)", latencyIncrease, thresholds.MaxLatencyIncrease),
			Timestamp: time.Now(),
		}

		// Throughput check
		throughputDecrease := float64(0)
		if stable.JobsPerSecond > 0 {
			throughputDecrease = (stable.JobsPerSecond - canary.JobsPerSecond) / stable.JobsPerSecond * 100
		}
		health.ThroughputCheck = HealthCheck{
			Name:      "Throughput",
			Passing:   throughputDecrease <= thresholds.MaxThroughputDecrease,
			Message:   fmt.Sprintf("Throughput decrease: %.2f%% (threshold: %.2f%%)", throughputDecrease, thresholds.MaxThroughputDecrease),
			Timestamp: time.Now(),
		}

		// Sample size check
		health.SampleSizeCheck = HealthCheck{
			Name:      "Sample Size",
			Passing:   canary.JobCount >= int64(thresholds.RequiredSampleSize),
			Message:   fmt.Sprintf("Sample size: %d (required: %d)", canary.JobCount, thresholds.RequiredSampleSize),
			Timestamp: time.Now(),
		}
	}

	// Duration check
	elapsed := time.Since(deployment.StartTime)
	health.DurationCheck = HealthCheck{
		Name:      "Duration",
		Passing:   elapsed >= deployment.Config.MinCanaryDuration,
		Message:   fmt.Sprintf("Duration: %v (minimum: %v)", elapsed.Truncate(time.Second), deployment.Config.MinCanaryDuration),
		Timestamp: time.Now(),
	}

	// Overall status
	if health.AllChecksPass() {
		health.OverallStatus = HealthyCanary
	} else if health.ErrorRateCheck.Passing && health.LatencyCheck.Passing {
		health.OverallStatus = WarningCanary
	} else {
		health.OverallStatus = FailingCanary
	}

	return health
}

func (m *Manager) evaluatePromotionConditions(stable, canary *MetricsSnapshot, conditions SLOThresholds) bool {
	if stable == nil || canary == nil {
		return false
	}

	if canary.JobCount < int64(conditions.RequiredSampleSize) {
		return false
	}

	// Check error rate
	errorRateIncrease := canary.ErrorRate - stable.ErrorRate
	if errorRateIncrease > conditions.MaxErrorRateIncrease {
		return false
	}

	// Check success rate
	if canary.SuccessRate < conditions.MinSuccessRate {
		return false
	}

	// Check latency
	if stable.P95Latency > 0 {
		latencyIncrease := (canary.P95Latency - stable.P95Latency) / stable.P95Latency * 100
		if latencyIncrease > conditions.MaxLatencyIncrease {
			return false
		}
	}

	// Check throughput
	if stable.JobsPerSecond > 0 {
		throughputDecrease := (stable.JobsPerSecond - canary.JobsPerSecond) / stable.JobsPerSecond * 100
		if throughputDecrease > conditions.MaxThroughputDecrease {
			return false
		}
	}

	return true
}

func (m *Manager) processAlerts() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case alert := <-m.alertChan:
			if err := m.alerter.SendAlert(m.ctx, alert); err != nil {
				m.logger.Error("Failed to send alert", "alert_id", alert.ID, "error", err)
			}
		}
	}
}

func (m *Manager) processEvents() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case event := <-m.eventChan:
			if err := m.saveEventToRedis(m.ctx, event); err != nil {
				m.logger.Error("Failed to save event", "event_id", event.ID, "error", err)
			}
		}
	}
}

func (m *Manager) cleanupExpiredData() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupOldMetrics()
			m.cleanupOldEvents()
		}
	}
}

func (m *Manager) cleanupOldMetrics() {
	cutoff := time.Now().Add(-m.config.MetricsRetention)
	pattern := "canary:metrics:*"

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Minute)
	defer cancel()

	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		m.logger.Error("Failed to list metrics keys for cleanup", "error", err)
		return
	}

	for _, key := range keys {
		// Check if the key is old enough to delete
		result := m.redis.HGet(ctx, key, "timestamp")
		if result.Err() != nil {
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, result.Val())
		if err != nil {
			continue
		}

		if timestamp.Before(cutoff) {
			m.redis.Del(ctx, key)
		}
	}
}

func (m *Manager) cleanupOldEvents() {
	cutoff := time.Now().Add(-m.config.EventRetention)
	pattern := "canary:events:*"

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Minute)
	defer cancel()

	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		m.logger.Error("Failed to list event keys for cleanup", "error", err)
		return
	}

	for _, key := range keys {
		// Remove old events from sorted sets
		m.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff.Unix()))
	}
}

func (m *Manager) emitEvent(deployment *CanaryDeployment, eventType, message string) {
	event := &DeploymentEvent{
		ID:           "event_" + uuid.New().String(),
		DeploymentID: deployment.ID,
		Type:         eventType,
		Message:      message,
		Timestamp:    time.Now(),
	}

	select {
	case m.eventChan <- event:
	default:
		// Channel full, log the event instead
		m.logger.Warn("Event channel full, dropping event",
			"event_type", eventType,
			"deployment_id", deployment.ID)
	}
}

func (m *Manager) copyDeployment(deployment *CanaryDeployment) *CanaryDeployment {
	// Create a deep copy to prevent external modification
	copy := *deployment
	if deployment.Config != nil {
		configCopy := *deployment.Config
		copy.Config = &configCopy
	}
	if deployment.StableMetrics != nil {
		metricsCopy := *deployment.StableMetrics
		copy.StableMetrics = &metricsCopy
	}
	if deployment.CanaryMetrics != nil {
		metricsCopy := *deployment.CanaryMetrics
		copy.CanaryMetrics = &metricsCopy
	}
	return &copy
}

func (m *Manager) drainCanaryQueue(ctx context.Context, deployment *CanaryDeployment) error {
	canaryQueue := deployment.QueueName + "@canary"
	stableQueue := deployment.QueueName

	timeout := deployment.Config.DrainTimeout
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Move one job from canary to stable queue
		result := m.redis.BRPopLPush(ctx, canaryQueue, stableQueue, time.Second)
		if result.Err() == redis.Nil {
			// Queue is empty
			break
		}
		if result.Err() != nil {
			return fmt.Errorf("failed to drain job: %w", result.Err())
		}
	}

	return nil
}

// Redis persistence methods

func (m *Manager) loadDeployments(ctx context.Context) error {
	pattern := "canary:deployment:*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to list deployment keys: %w", err)
	}

	for _, key := range keys {
		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			m.logger.Warn("Failed to load deployment", "key", key, "error", err)
			continue
		}

		var deployment CanaryDeployment
		if err := json.Unmarshal([]byte(data), &deployment); err != nil {
			m.logger.Warn("Failed to unmarshal deployment", "key", key, "error", err)
			continue
		}

		m.deployments[deployment.ID] = &deployment
	}

	m.logger.Info("Loaded deployments from Redis", "count", len(m.deployments))
	return nil
}

func (m *Manager) saveDeployment(ctx context.Context, deployment *CanaryDeployment) error {
	data, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	key := fmt.Sprintf("canary:deployment:%s", deployment.ID)
	if err := m.redis.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save deployment to Redis: %w", err)
	}

	return nil
}

func (m *Manager) deleteDeploymentFromRedis(ctx context.Context, id string) error {
	key := fmt.Sprintf("canary:deployment:%s", id)
	return m.redis.Del(ctx, key).Err()
}

func (m *Manager) saveEventToRedis(ctx context.Context, event *DeploymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := fmt.Sprintf("canary:events:%s", event.DeploymentID)
	score := float64(event.Timestamp.Unix())

	return m.redis.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

func (m *Manager) loadEventsFromRedis(ctx context.Context, deploymentID string) ([]*DeploymentEvent, error) {
	key := fmt.Sprintf("canary:events:%s", deploymentID)

	results, err := m.redis.ZRevRange(ctx, key, 0, 100).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	events := make([]*DeploymentEvent, 0, len(results))
	for _, result := range results {
		var event DeploymentEvent
		if err := json.Unmarshal([]byte(result), &event); err != nil {
			m.logger.Warn("Failed to unmarshal event", "error", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}