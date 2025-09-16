// Copyright 2025 James Ross
package dlqremediation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RemediationPipeline is the main pipeline for DLQ remediation
type RemediationPipeline struct {
	redis           *redis.Client
	logger          *zap.Logger
	config          *Config
	classifier      *ClassificationEngine
	actionExecutor  *ActionExecutor
	auditLogger     *AuditLogger
	rateLimiter     *RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	idempotency     *IdempotencyTracker
	rules           []RemediationRule
	state           *PipelineState
	stopCh          chan struct{}
	wg              sync.WaitGroup
	mu              sync.RWMutex
	metrics         *PipelineMetrics
}

// NewRemediationPipeline creates a new remediation pipeline
func NewRemediationPipeline(config *Config, logger *zap.Logger) (*RemediationPipeline, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Apply defaults and validate config
	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.Redis.Addr,
		Password:     config.Redis.Password,
		DB:           config.Redis.DB,
		MaxRetries:   config.Redis.MaxRetries,
		DialTimeout:  config.Redis.DialTimeout,
		ReadTimeout:  config.Redis.ReadTimeout,
		WriteTimeout: config.Redis.WriteTimeout,
		PoolSize:     config.Redis.PoolSize,
		MinIdleConns: config.Redis.MinIdleConns,
		MaxConnAge:   config.Redis.MaxConnAge,
		PoolTimeout:  config.Redis.PoolTimeout,
		IdleTimeout:  config.Redis.IdleTimeout,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	pipeline := &RemediationPipeline{
		redis:  redisClient,
		logger: logger,
		config: config,
		classifier: NewClassificationEngine(redisClient, &config.Pipeline, logger),
		actionExecutor: NewActionExecutor(redisClient, &config.Pipeline, logger),
		auditLogger: NewAuditLogger(redisClient, &config.Pipeline, logger),
		rateLimiter: &RateLimiter{
			MaxPerMinute: config.Pipeline.GlobalSafetyLimits.MaxPerMinute,
			MaxTotal:     config.Pipeline.GlobalSafetyLimits.MaxTotalPerRun,
			BurstSize:    config.Pipeline.BatchSize,
		},
		circuitBreakers: make(map[string]*CircuitBreaker),
		idempotency:     NewIdempotencyTracker(config.Pipeline.RetentionPolicy.ProcessedJobsTTL),
		stopCh:          make(chan struct{}),
		state: &PipelineState{
			Status:    StatusStopped,
			StartedAt: time.Now(),
		},
		metrics: &PipelineMetrics{},
	}

	// Initialize circuit breakers for different operations
	pipeline.initializeCircuitBreakers()

	// Load existing rules
	if err := pipeline.loadRules(context.Background()); err != nil {
		logger.Warn("Failed to load existing rules", zap.Error(err))
	}

	logger.Info("DLQ remediation pipeline created successfully")

	return pipeline, nil
}

// Start starts the remediation pipeline
func (rp *RemediationPipeline) Start(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.state.Status == StatusRunning {
		return fmt.Errorf("pipeline is already running")
	}

	if !rp.config.Pipeline.Enabled {
		return fmt.Errorf("pipeline is disabled in configuration")
	}

	rp.state.Status = StatusRunning
	rp.state.StartedAt = time.Now()
	rp.state.LastError = ""
	rp.state.LastErrorAt = time.Time{}

	// Store state in Redis
	rp.saveState(ctx)

	// Start processing loop
	rp.wg.Add(1)
	go rp.processingLoop(ctx)

	// Start cleanup routine
	rp.wg.Add(1)
	go rp.cleanupLoop(ctx)

	rp.logger.Info("DLQ remediation pipeline started")

	return rp.auditLogger.LogPipelineState(ctx, "pipeline_started", rp.state, "system")
}

// Stop stops the remediation pipeline
func (rp *RemediationPipeline) Stop(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.state.Status != StatusRunning {
		return fmt.Errorf("pipeline is not running")
	}

	close(rp.stopCh)
	rp.wg.Wait()

	rp.state.Status = StatusStopped
	rp.saveState(ctx)

	rp.logger.Info("DLQ remediation pipeline stopped")

	return rp.auditLogger.LogPipelineState(ctx, "pipeline_stopped", rp.state, "system")
}

// Pause pauses the remediation pipeline
func (rp *RemediationPipeline) Pause(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.state.Status != StatusRunning {
		return fmt.Errorf("pipeline is not running")
	}

	rp.state.Status = StatusPaused
	rp.saveState(ctx)

	rp.logger.Info("DLQ remediation pipeline paused")

	return rp.auditLogger.LogPipelineState(ctx, "pipeline_paused", rp.state, "system")
}

// Resume resumes the remediation pipeline
func (rp *RemediationPipeline) Resume(ctx context.Context) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.state.Status != StatusPaused {
		return fmt.Errorf("pipeline is not paused")
	}

	rp.state.Status = StatusRunning
	rp.saveState(ctx)

	rp.logger.Info("DLQ remediation pipeline resumed")

	return rp.auditLogger.LogPipelineState(ctx, "pipeline_resumed", rp.state, "system")
}

// ProcessBatch processes a batch of DLQ jobs
func (rp *RemediationPipeline) ProcessBatch(ctx context.Context, dryRun bool) (*BatchResult, error) {
	startTime := time.Now()

	result := &BatchResult{
		StartedAt: startTime,
		Results:   make([]ProcessingResult, 0),
	}

	// Fetch DLQ jobs
	jobs, err := rp.fetchDLQJobs(ctx, rp.config.Pipeline.BatchSize)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch DLQ jobs: %v", err))
		result.CompletedAt = time.Now()
		return result, err
	}

	result.TotalJobs = len(jobs)

	if len(jobs) == 0 {
		rp.logger.Debug("No DLQ jobs to process")
		result.CompletedAt = time.Now()
		return result, nil
	}

	rp.logger.Info("Processing DLQ batch",
		zap.Int("job_count", len(jobs)),
		zap.Bool("dry_run", dryRun))

	// Process each job
	for _, job := range jobs {
		if !rp.shouldProcessJob(job) {
			result.SkippedJobs++
			continue
		}

		// Check rate limits
		if !rp.rateLimiter.CanProcess() {
			rp.logger.Warn("Rate limit reached, stopping batch processing")
			break
		}

		jobResult := rp.processJob(ctx, job, dryRun)
		result.Results = append(result.Results, *jobResult)
		result.ProcessedJobs++

		if jobResult.Success {
			result.SuccessfulJobs++
			rp.rateLimiter.RecordProcessed()
		} else {
			result.FailedJobs++
			result.Errors = append(result.Errors, fmt.Sprintf("Job %s: %s", jobResult.JobID, jobResult.Error))
		}

		// Record metrics
		rp.updateMetrics(jobResult)
	}

	result.CompletedAt = time.Now()

	// Update pipeline state
	rp.mu.Lock()
	rp.state.LastRunAt = time.Now()
	rp.state.NextRunAt = time.Now().Add(rp.config.Pipeline.PollInterval)
	rp.state.TotalProcessed += int64(result.ProcessedJobs)
	rp.state.TotalSuccessful += int64(result.SuccessfulJobs)
	rp.state.TotalFailed += int64(result.FailedJobs)
	rp.state.CurrentBatchSize = len(jobs)
	rp.mu.Unlock()

	rp.saveState(ctx)

	// Log batch operation
	err = rp.auditLogger.LogBulkOperation(ctx, "batch_processing", result, "system")
	if err != nil {
		rp.logger.Warn("Failed to log batch operation", zap.Error(err))
	}

	rp.logger.Info("Batch processing completed",
		zap.Int("total_jobs", result.TotalJobs),
		zap.Int("successful_jobs", result.SuccessfulJobs),
		zap.Int("failed_jobs", result.FailedJobs),
		zap.Int("skipped_jobs", result.SkippedJobs),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)))

	return result, nil
}

// processingLoop is the main processing loop
func (rp *RemediationPipeline) processingLoop(ctx context.Context) {
	defer rp.wg.Done()

	ticker := time.NewTicker(rp.config.Pipeline.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rp.stopCh:
			return
		case <-ticker.C:
			rp.mu.RLock()
			if rp.state.Status == StatusRunning {
				rp.mu.RUnlock()

				_, err := rp.ProcessBatch(ctx, rp.config.Pipeline.DryRun)
				if err != nil {
					rp.logger.Error("Batch processing failed", zap.Error(err))
					rp.handleProcessingError(ctx, err)
				}
			} else {
				rp.mu.RUnlock()
			}
		}
	}
}

// cleanupLoop performs periodic cleanup tasks
func (rp *RemediationPipeline) cleanupLoop(ctx context.Context) {
	defer rp.wg.Done()

	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rp.stopCh:
			return
		case <-ticker.C:
			rp.performCleanup(ctx)
		}
	}
}

// processJob processes a single DLQ job
func (rp *RemediationPipeline) processJob(ctx context.Context, job *DLQJob, dryRun bool) *ProcessingResult {
	startTime := time.Now()

	// Check if already processed
	if rp.idempotency.IsProcessed(job.JobID) {
		return &ProcessingResult{
			JobID:   job.JobID,
			Success: true,
			Duration: time.Since(startTime),
			DryRun:  dryRun,
		}
	}

	// Classify the job
	classification, err := rp.classifier.Classify(ctx, job, rp.rules)
	if err != nil {
		return &ProcessingResult{
			JobID:   job.JobID,
			Success: false,
			Error:   fmt.Sprintf("Classification failed: %v", err),
			Duration: time.Since(startTime),
			DryRun:  dryRun,
		}
	}

	// Find matching rule
	rule := rp.findRuleByID(classification.RuleID)
	if rule == nil && classification.RuleID != "" {
		return &ProcessingResult{
			JobID:   job.JobID,
			Success: false,
			Error:   fmt.Sprintf("Rule %s not found", classification.RuleID),
			Duration: time.Since(startTime),
			DryRun:  dryRun,
		}
	}

	// If no rule matches, skip processing
	if rule == nil {
		return &ProcessingResult{
			JobID:   job.JobID,
			Success: true,
			Duration: time.Since(startTime),
			DryRun:  dryRun,
		}
	}

	// Check safety limits for the rule
	if !rp.canProcessRule(rule) {
		return &ProcessingResult{
			JobID:   job.JobID,
			Success: false,
			Error:   "Safety limits exceeded for rule",
			Duration: time.Since(startTime),
			DryRun:  dryRun,
		}
	}

	// Execute actions
	result, err := rp.actionExecutor.Execute(ctx, job, rule.Actions, dryRun)
	if err != nil {
		rp.recordRuleFailure(rule)
		return result
	}

	// Mark as processed if not dry run
	if !dryRun {
		rp.idempotency.MarkProcessed(job.JobID)
	}

	// Update rule statistics
	rp.updateRuleStatistics(rule, result.Success, result.Duration)

	// Log the remediation
	err = rp.auditLogger.LogRemediation(ctx, job, rule, classification, result, "system")
	if err != nil {
		rp.logger.Warn("Failed to log remediation", zap.Error(err))
	}

	return result
}

// shouldProcessJob checks if a job should be processed
func (rp *RemediationPipeline) shouldProcessJob(job *DLQJob) bool {
	// Skip if already processed
	if rp.idempotency.IsProcessed(job.JobID) {
		return false
	}

	// Add any other business logic for skipping jobs
	return true
}

// canProcessRule checks if a rule can be processed based on safety limits
func (rp *RemediationPipeline) canProcessRule(rule *RemediationRule) bool {
	// Check circuit breaker
	cb := rp.getCircuitBreaker(rule.ID)
	if !cb.CanExecute() {
		return false
	}

	// Check rate limits
	// Could implement per-rule rate limiting here
	return true
}

// fetchDLQJobs fetches jobs from the DLQ
func (rp *RemediationPipeline) fetchDLQJobs(ctx context.Context, limit int) ([]*DLQJob, error) {
	// Fetch from Redis DLQ stream
	entries, err := rp.redis.LRange(ctx, rp.config.Storage.DLQStreamKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DLQ jobs: %w", err)
	}

	jobs := make([]*DLQJob, 0, len(entries))
	for _, entry := range entries {
		var job DLQJob
		if err := json.Unmarshal([]byte(entry), &job); err != nil {
			rp.logger.Warn("Failed to unmarshal DLQ job", zap.Error(err))
			continue
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// findRuleByID finds a rule by its ID
func (rp *RemediationPipeline) findRuleByID(ruleID string) *RemediationRule {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	for i := range rp.rules {
		if rp.rules[i].ID == ruleID {
			return &rp.rules[i]
		}
	}
	return nil
}

// getCircuitBreaker gets or creates a circuit breaker for a rule
func (rp *RemediationPipeline) getCircuitBreaker(ruleID string) *CircuitBreaker {
	if cb, exists := rp.circuitBreakers[ruleID]; exists {
		return cb
	}

	cb := &CircuitBreaker{
		ErrorThreshold:  rp.config.Pipeline.GlobalSafetyLimits.ErrorRateThreshold,
		MinRequests:     5,
		RecoveryTimeout: 5 * time.Minute,
		State:          CircuitClosed,
	}

	rp.circuitBreakers[ruleID] = cb
	return cb
}

// recordRuleFailure records a failure for a rule
func (rp *RemediationPipeline) recordRuleFailure(rule *RemediationRule) {
	cb := rp.getCircuitBreaker(rule.ID)
	cb.RecordFailure()

	// Update rule statistics
	rp.mu.Lock()
	for i := range rp.rules {
		if rp.rules[i].ID == rule.ID {
			rp.rules[i].Statistics.FailedActions++
			rp.rules[i].Statistics.LastFailureAt = time.Now()
			rp.rules[i].Statistics.SuccessRate = float64(rp.rules[i].Statistics.SuccessfulActions) /
				float64(rp.rules[i].Statistics.SuccessfulActions + rp.rules[i].Statistics.FailedActions)
			break
		}
	}
	rp.mu.Unlock()
}

// updateRuleStatistics updates statistics for a rule
func (rp *RemediationPipeline) updateRuleStatistics(rule *RemediationRule, success bool, duration time.Duration) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	for i := range rp.rules {
		if rp.rules[i].ID == rule.ID {
			rp.rules[i].Statistics.TotalMatches++
			rp.rules[i].Statistics.LastMatchedAt = time.Now()

			if success {
				rp.rules[i].Statistics.SuccessfulActions++
				rp.rules[i].Statistics.LastSuccessAt = time.Now()

				// Record success in circuit breaker
				cb := rp.getCircuitBreaker(rule.ID)
				cb.RecordSuccess()
			} else {
				rp.rules[i].Statistics.FailedActions++
				rp.rules[i].Statistics.LastFailureAt = time.Now()
			}

			// Update success rate
			total := rp.rules[i].Statistics.SuccessfulActions + rp.rules[i].Statistics.FailedActions
			if total > 0 {
				rp.rules[i].Statistics.SuccessRate = float64(rp.rules[i].Statistics.SuccessfulActions) / float64(total)
			}

			// Update average latency
			if rp.rules[i].Statistics.AverageLatency == 0 {
				rp.rules[i].Statistics.AverageLatency = duration.Seconds() * 1000 // Convert to ms
			} else {
				// Simple moving average
				rp.rules[i].Statistics.AverageLatency = (rp.rules[i].Statistics.AverageLatency + duration.Seconds()*1000) / 2
			}

			break
		}
	}
}

// updateMetrics updates pipeline metrics
func (rp *RemediationPipeline) updateMetrics(result *ProcessingResult) {
	rp.metrics.Timestamp = time.Now()
	rp.metrics.JobsProcessed++

	if result.Success {
		rp.metrics.ActionsSuccessful++
	} else {
		rp.metrics.ActionsFailed++
	}

	rp.metrics.EndToEndTime = result.Duration.Seconds() * 1000 // Convert to ms
	rp.metrics.ActionsExecuted += int64(len(result.Actions))
}

// handleProcessingError handles processing errors
func (rp *RemediationPipeline) handleProcessingError(ctx context.Context, err error) {
	rp.mu.Lock()
	rp.state.LastError = err.Error()
	rp.state.LastErrorAt = time.Now()
	rp.mu.Unlock()

	// Could implement circuit breaker for the entire pipeline here
	// For now, just log and continue
	rp.logger.Error("Processing error", zap.Error(err))
}

// performCleanup performs periodic cleanup tasks
func (rp *RemediationPipeline) performCleanup(ctx context.Context) {
	rp.logger.Debug("Performing periodic cleanup")

	// Cleanup idempotency tracker
	rp.idempotency.Cleanup()

	// Cleanup audit log
	if err := rp.auditLogger.CleanupExpiredEntries(ctx); err != nil {
		rp.logger.Warn("Failed to cleanup audit entries", zap.Error(err))
	}

	// Could add more cleanup tasks here
}

// initializeCircuitBreakers initializes circuit breakers
func (rp *RemediationPipeline) initializeCircuitBreakers() {
	// Circuit breakers are created on-demand in getCircuitBreaker
}

// loadRules loads rules from storage
func (rp *RemediationPipeline) loadRules(ctx context.Context) error {
	entries, err := rp.redis.HGetAll(ctx, rp.config.Storage.RulesKey).Result()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	rules := make([]RemediationRule, 0, len(entries))
	for _, ruleData := range entries {
		var rule RemediationRule
		if err := json.Unmarshal([]byte(ruleData), &rule); err != nil {
			rp.logger.Warn("Failed to unmarshal rule", zap.Error(err))
			continue
		}
		rules = append(rules, rule)
	}

	rp.mu.Lock()
	rp.rules = rules
	rp.state.RulesEnabled = 0
	rp.state.RulesDisabled = 0
	for _, rule := range rules {
		if rule.Enabled {
			rp.state.RulesEnabled++
		} else {
			rp.state.RulesDisabled++
		}
	}
	rp.mu.Unlock()

	rp.logger.Info("Rules loaded successfully",
		zap.Int("total_rules", len(rules)),
		zap.Int("enabled_rules", rp.state.RulesEnabled),
		zap.Int("disabled_rules", rp.state.RulesDisabled))

	return nil
}

// saveState saves pipeline state to Redis
func (rp *RemediationPipeline) saveState(ctx context.Context) error {
	stateData, err := json.Marshal(rp.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return rp.redis.Set(ctx, rp.config.Storage.StateKey, string(stateData), 0).Err()
}

// GetState returns the current pipeline state
func (rp *RemediationPipeline) GetState() *PipelineState {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// Return a copy to avoid race conditions
	state := *rp.state
	return &state
}

// GetMetrics returns current pipeline metrics
func (rp *RemediationPipeline) GetMetrics() *PipelineMetrics {
	// Return a copy to avoid race conditions
	metrics := *rp.metrics
	return &metrics
}

// GetRules returns current rules
func (rp *RemediationPipeline) GetRules() []RemediationRule {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// Return a copy to avoid race conditions
	rules := make([]RemediationRule, len(rp.rules))
	copy(rules, rp.rules)
	return rules
}

// AddRule adds a new remediation rule
func (rp *RemediationPipeline) AddRule(ctx context.Context, rule RemediationRule, userID string) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.CreatedBy = userID

	// Store in Redis
	ruleData, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal rule: %w", err)
	}

	err = rp.redis.HSet(ctx, rp.config.Storage.RulesKey, rule.ID, string(ruleData)).Err()
	if err != nil {
		return fmt.Errorf("failed to store rule: %w", err)
	}

	// Update local rules
	rp.mu.Lock()
	rp.rules = append(rp.rules, rule)
	if rule.Enabled {
		rp.state.RulesEnabled++
	} else {
		rp.state.RulesDisabled++
	}
	rp.mu.Unlock()

	// Log the change
	err = rp.auditLogger.LogRuleChange(ctx, rule.ID, "rule_created", userID, nil, &rule)
	if err != nil {
		rp.logger.Warn("Failed to log rule creation", zap.Error(err))
	}

	rp.logger.Info("Rule added successfully",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
		zap.String("user_id", userID))

	return nil
}

// UpdateRule updates an existing remediation rule
func (rp *RemediationPipeline) UpdateRule(ctx context.Context, ruleID string, updated RemediationRule, userID string) error {
	// Get existing rule for audit
	rp.mu.RLock()
	var existing *RemediationRule
	for i := range rp.rules {
		if rp.rules[i].ID == ruleID {
			existing = &rp.rules[i]
			break
		}
	}
	rp.mu.RUnlock()

	if existing == nil {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	updated.ID = ruleID
	updated.CreatedAt = existing.CreatedAt
	updated.CreatedBy = existing.CreatedBy
	updated.UpdatedAt = time.Now()

	// Store in Redis
	ruleData, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("failed to marshal rule: %w", err)
	}

	err = rp.redis.HSet(ctx, rp.config.Storage.RulesKey, ruleID, string(ruleData)).Err()
	if err != nil {
		return fmt.Errorf("failed to store rule: %w", err)
	}

	// Update local rules
	rp.mu.Lock()
	for i := range rp.rules {
		if rp.rules[i].ID == ruleID {
			// Update enabled/disabled counts
			if existing.Enabled && !updated.Enabled {
				rp.state.RulesEnabled--
				rp.state.RulesDisabled++
			} else if !existing.Enabled && updated.Enabled {
				rp.state.RulesDisabled--
				rp.state.RulesEnabled++
			}

			rp.rules[i] = updated
			break
		}
	}
	rp.mu.Unlock()

	// Log the change
	err = rp.auditLogger.LogRuleChange(ctx, ruleID, "rule_updated", userID, existing, &updated)
	if err != nil {
		rp.logger.Warn("Failed to log rule update", zap.Error(err))
	}

	rp.logger.Info("Rule updated successfully",
		zap.String("rule_id", ruleID),
		zap.String("user_id", userID))

	return nil
}

// DeleteRule deletes a remediation rule
func (rp *RemediationPipeline) DeleteRule(ctx context.Context, ruleID, userID string) error {
	// Get existing rule for audit
	rp.mu.RLock()
	var existing *RemediationRule
	index := -1
	for i := range rp.rules {
		if rp.rules[i].ID == ruleID {
			existing = &rp.rules[i]
			index = i
			break
		}
	}
	rp.mu.RUnlock()

	if existing == nil {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	// Remove from Redis
	err := rp.redis.HDel(ctx, rp.config.Storage.RulesKey, ruleID).Err()
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	// Remove from local rules
	rp.mu.Lock()
	if index >= 0 {
		rp.rules = append(rp.rules[:index], rp.rules[index+1:]...)
		if existing.Enabled {
			rp.state.RulesEnabled--
		} else {
			rp.state.RulesDisabled--
		}
	}
	rp.mu.Unlock()

	// Log the change
	err = rp.auditLogger.LogRuleChange(ctx, ruleID, "rule_deleted", userID, existing, nil)
	if err != nil {
		rp.logger.Warn("Failed to log rule deletion", zap.Error(err))
	}

	rp.logger.Info("Rule deleted successfully",
		zap.String("rule_id", ruleID),
		zap.String("user_id", userID))

	return nil
}

// Close closes the pipeline and releases resources
func (rp *RemediationPipeline) Close() error {
	if rp.state.Status == StatusRunning {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rp.Stop(ctx)
	}

	return rp.redis.Close()
}