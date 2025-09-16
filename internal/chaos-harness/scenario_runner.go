// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ScenarioRunner executes chaos scenarios
type ScenarioRunner struct {
	injectorManager *FaultInjectorManager
	loadGenerator   *LoadGenerator
	metricsCollector *MetricsCollector
	logger          *zap.Logger

	// Running scenarios
	running map[string]*runningScenario
	mu      sync.RWMutex
}

// runningScenario tracks a running scenario
type runningScenario struct {
	scenario   *ChaosScenario
	cancel     context.CancelFunc
	done       chan struct{}
	metrics    *ScenarioMetrics
	aborted    atomic.Bool
}

// NewScenarioRunner creates a new scenario runner
func NewScenarioRunner(injectorManager *FaultInjectorManager, logger *zap.Logger) *ScenarioRunner {
	return &ScenarioRunner{
		injectorManager:  injectorManager,
		loadGenerator:    NewLoadGenerator(logger),
		metricsCollector: NewMetricsCollector(),
		logger:          logger,
		running:         make(map[string]*runningScenario),
	}
}

// RunScenario executes a chaos scenario
func (sr *ScenarioRunner) RunScenario(ctx context.Context, scenario *ChaosScenario) error {
	// Validate scenario
	if err := sr.validateScenario(scenario); err != nil {
		return fmt.Errorf("invalid scenario: %w", err)
	}

	// Check if already running
	sr.mu.Lock()
	if _, exists := sr.running[scenario.ID]; exists {
		sr.mu.Unlock()
		return fmt.Errorf("scenario %s is already running", scenario.ID)
	}

	// Check guardrails
	if scenario.Guardrails.RequireConfirm {
		// In a real implementation, this would prompt for confirmation
		sr.logger.Warn("Scenario requires confirmation",
			zap.String("scenario_id", scenario.ID),
			zap.String("name", scenario.Name))
	}

	// Create running scenario
	ctx, cancel := context.WithTimeout(ctx, scenario.Duration)
	rs := &runningScenario{
		scenario: scenario,
		cancel:   cancel,
		done:     make(chan struct{}),
		metrics:  &ScenarioMetrics{
			TimeSeriesData: make([]TimeSeriesPoint, 0),
		},
	}

	sr.running[scenario.ID] = rs
	sr.mu.Unlock()

	// Update scenario status
	now := time.Now()
	scenario.Status = StatusRunning
	scenario.StartedAt = &now
	scenario.Metrics = rs.metrics

	sr.logger.Info("Starting chaos scenario",
		zap.String("scenario_id", scenario.ID),
		zap.String("name", scenario.Name),
		zap.Duration("duration", scenario.Duration),
		zap.Int("stages", len(scenario.Stages)))

	// Run scenario in goroutine
	go sr.executeScenario(ctx, rs)

	// Wait for completion or timeout
	select {
	case <-rs.done:
		// Scenario completed
	case <-ctx.Done():
		// Context cancelled or timeout
		rs.aborted.Store(true)
	}

	// Clean up
	sr.mu.Lock()
	delete(sr.running, scenario.ID)
	sr.mu.Unlock()

	// Update final status
	endTime := time.Now()
	scenario.EndedAt = &endTime

	if rs.aborted.Load() {
		scenario.Status = StatusAborted
		return fmt.Errorf("scenario aborted")
	}

	// Check if scenario passed based on metrics
	if sr.evaluateScenario(scenario) {
		scenario.Status = StatusCompleted
	} else {
		scenario.Status = StatusFailed
	}

	return nil
}

// executeScenario runs the scenario stages
func (sr *ScenarioRunner) executeScenario(ctx context.Context, rs *runningScenario) {
	defer close(rs.done)

	// Start metrics collection
	metricsCtx, metricsCancel := context.WithCancel(ctx)
	defer metricsCancel()

	go sr.collectMetrics(metricsCtx, rs)

	// Execute stages
	for i, stage := range rs.scenario.Stages {
		if rs.aborted.Load() {
			break
		}

		sr.logger.Info("Starting scenario stage",
			zap.String("scenario_id", rs.scenario.ID),
			zap.Int("stage", i+1),
			zap.String("name", stage.Name),
			zap.Duration("duration", stage.Duration))

		// Apply injectors for this stage
		for _, injector := range stage.Injectors {
			if err := sr.injectorManager.AddInjector(&injector); err != nil {
				sr.logger.Error("Failed to add injector",
					zap.String("injector_id", injector.ID),
					zap.Error(err))
			}
		}

		// Start load generation if configured
		if stage.LoadConfig != nil {
			sr.loadGenerator.Start(ctx, stage.LoadConfig, rs.metrics)
		}

		// Run stage for its duration
		stageCtx, stageCancel := context.WithTimeout(ctx, stage.Duration)
		sr.monitorStage(stageCtx, rs)
		stageCancel()

		// Stop load generation
		if stage.LoadConfig != nil {
			sr.loadGenerator.Stop()
		}

		// Remove stage-specific injectors
		for _, injector := range stage.Injectors {
			sr.injectorManager.RemoveInjector(injector.ID)
		}
	}

	// Calculate recovery metrics
	sr.calculateRecoveryMetrics(rs)
}

// monitorStage monitors a stage for guardrail violations
func (sr *ScenarioRunner) monitorStage(ctx context.Context, rs *runningScenario) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check guardrails
			if violation := sr.checkGuardrails(rs); violation != nil {
				sr.logger.Error("Guardrail violation detected",
					zap.String("scenario_id", rs.scenario.ID),
					zap.Error(violation))

				if rs.scenario.Guardrails.AutoAbortOnPanic {
					rs.aborted.Store(true)
					rs.cancel()
					return
				}
			}
		}
	}
}

// collectMetrics collects metrics during scenario execution
func (sr *ScenarioRunner) collectMetrics(ctx context.Context, rs *runningScenario) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Collect current metrics
			point := sr.metricsCollector.Collect()

			// Get active faults
			activeFaults := make([]string, 0)
			for _, injector := range sr.injectorManager.GetActiveInjectors() {
				activeFaults = append(activeFaults, injector.ID)
			}
			point.ActiveFaults = activeFaults

			// Add to time series
			rs.metrics.mu.Lock()
			rs.metrics.TimeSeriesData = append(rs.metrics.TimeSeriesData, point)

			// Update aggregate metrics
			sr.updateAggregateMetrics(rs.metrics, point)
			rs.metrics.mu.Unlock()
		}
	}
}

// updateAggregateMetrics updates aggregate metrics from point data
func (sr *ScenarioRunner) updateAggregateMetrics(metrics *ScenarioMetrics, point TimeSeriesPoint) {
	if requests, ok := point.Metrics["requests"]; ok {
		metrics.TotalRequests = int64(requests)
	}
	if successful, ok := point.Metrics["successful"]; ok {
		metrics.SuccessfulRequests = int64(successful)
	}
	if failed, ok := point.Metrics["failed"]; ok {
		metrics.FailedRequests = int64(failed)
	}
	if faults, ok := point.Metrics["faults_injected"]; ok {
		metrics.InjectedFaults = int64(faults)
	}
	if backlog, ok := point.Metrics["backlog_size"]; ok {
		metrics.BacklogSize = int64(backlog)
	}

	// Calculate error rate
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// Update latency percentiles
	if p50, ok := point.Metrics["latency_p50_ms"]; ok {
		metrics.LatencyP50 = time.Duration(p50) * time.Millisecond
	}
	if p95, ok := point.Metrics["latency_p95_ms"]; ok {
		metrics.LatencyP95 = time.Duration(p95) * time.Millisecond
	}
	if p99, ok := point.Metrics["latency_p99_ms"]; ok {
		metrics.LatencyP99 = time.Duration(p99) * time.Millisecond
	}
}

// calculateRecoveryMetrics calculates recovery time after chaos
func (sr *ScenarioRunner) calculateRecoveryMetrics(rs *runningScenario) {
	rs.metrics.mu.Lock()
	defer rs.metrics.mu.Unlock()

	if len(rs.metrics.TimeSeriesData) < 2 {
		return
	}

	// Find when all faults were removed
	var faultsEndTime time.Time
	for i := len(rs.metrics.TimeSeriesData) - 1; i >= 0; i-- {
		if len(rs.metrics.TimeSeriesData[i].ActiveFaults) > 0 {
			faultsEndTime = rs.metrics.TimeSeriesData[i].Timestamp
			break
		}
	}

	if faultsEndTime.IsZero() {
		return
	}

	// Find when metrics returned to normal
	baselineErrorRate := rs.metrics.TimeSeriesData[0].Metrics["error_rate"]
	for i := len(rs.metrics.TimeSeriesData) - 1; i >= 0; i-- {
		point := rs.metrics.TimeSeriesData[i]
		if point.Timestamp.After(faultsEndTime) {
			errorRate := point.Metrics["error_rate"]
			if errorRate <= baselineErrorRate*1.1 { // Within 10% of baseline
				rs.metrics.RecoveryTime = point.Timestamp.Sub(faultsEndTime)
				break
			}
		}
	}
}

// checkGuardrails checks if guardrails are violated
func (sr *ScenarioRunner) checkGuardrails(rs *runningScenario) error {
	rs.metrics.mu.RLock()
	defer rs.metrics.mu.RUnlock()

	guardrails := rs.scenario.Guardrails

	// Check error rate
	if guardrails.MaxErrorRate > 0 && rs.metrics.ErrorRate > guardrails.MaxErrorRate {
		return &GuardrailViolation{
			ScenarioID: rs.scenario.ID,
			Guardrail:  "max_error_rate",
			Current:    rs.metrics.ErrorRate,
			Limit:      guardrails.MaxErrorRate,
		}
	}

	// Check latency
	if guardrails.MaxLatencyP99 > 0 && rs.metrics.LatencyP99 > guardrails.MaxLatencyP99 {
		return &GuardrailViolation{
			ScenarioID: rs.scenario.ID,
			Guardrail:  "max_latency_p99",
			Current:    rs.metrics.LatencyP99,
			Limit:      guardrails.MaxLatencyP99,
		}
	}

	// Check backlog size
	if guardrails.MaxBacklogSize > 0 && rs.metrics.BacklogSize > guardrails.MaxBacklogSize {
		return &GuardrailViolation{
			ScenarioID: rs.scenario.ID,
			Guardrail:  "max_backlog_size",
			Current:    rs.metrics.BacklogSize,
			Limit:      guardrails.MaxBacklogSize,
		}
	}

	return nil
}

// evaluateScenario evaluates if scenario passed
func (sr *ScenarioRunner) evaluateScenario(scenario *ChaosScenario) bool {
	if scenario.Metrics == nil {
		return false
	}

	scenario.Metrics.mu.RLock()
	defer scenario.Metrics.mu.RUnlock()

	// Check if recovery was successful
	if scenario.Metrics.RecoveryTime == 0 {
		return false // Never recovered
	}

	// Check final error rate
	if scenario.Metrics.ErrorRate > 0.1 { // More than 10% errors
		return false
	}

	return true
}

// validateScenario validates a scenario configuration
func (sr *ScenarioRunner) validateScenario(scenario *ChaosScenario) error {
	if scenario.ID == "" {
		return fmt.Errorf("scenario ID is required")
	}

	if scenario.Duration <= 0 {
		return fmt.Errorf("scenario duration must be positive")
	}

	if len(scenario.Stages) == 0 {
		return fmt.Errorf("scenario must have at least one stage")
	}

	for i, stage := range scenario.Stages {
		if stage.Duration <= 0 {
			return fmt.Errorf("stage %d duration must be positive", i)
		}
	}

	return nil
}

// AbortScenario aborts a running scenario
func (sr *ScenarioRunner) AbortScenario(scenarioID string) error {
	sr.mu.RLock()
	rs, exists := sr.running[scenarioID]
	sr.mu.RUnlock()

	if !exists {
		return fmt.Errorf("scenario %s is not running", scenarioID)
	}

	rs.aborted.Store(true)
	rs.cancel()

	sr.logger.Info("Aborted scenario", zap.String("scenario_id", scenarioID))
	return nil
}

// GetRunningScenarios returns currently running scenarios
func (sr *ScenarioRunner) GetRunningScenarios() []*ChaosScenario {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var scenarios []*ChaosScenario
	for _, rs := range sr.running {
		scenarios = append(scenarios, rs.scenario)
	}

	return scenarios
}
