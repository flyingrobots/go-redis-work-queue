// Copyright 2025 James Ross
package policysimulator

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PolicySimulator provides what-if analysis for queue policy changes
type PolicySimulator struct {
	mu              sync.RWMutex
	config          *SimulatorConfig
	trafficPatterns map[string]*TrafficPattern
	policies        map[string]*PolicyConfig
	simulations     map[string]*SimulationResult
	changes         map[string]*PolicyChange
}

// NewPolicySimulator creates a new policy simulator
func NewPolicySimulator(config *SimulatorConfig) *PolicySimulator {
	if config == nil {
		config = DefaultSimulatorConfig()
	}

	return &PolicySimulator{
		config:          config,
		trafficPatterns: make(map[string]*TrafficPattern),
		policies:        make(map[string]*PolicyConfig),
		simulations:     make(map[string]*SimulationResult),
		changes:         make(map[string]*PolicyChange),
	}
}

// DefaultSimulatorConfig returns sensible default configuration
func DefaultSimulatorConfig() *SimulatorConfig {
	return &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        10,
		RedisPoolSize:     20,
	}
}

// DefaultPolicyConfig returns default policy configuration
func DefaultPolicyConfig() *PolicyConfig {
	return &PolicyConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffStrategy:   "exponential",
		MaxRatePerSecond:  100.0,
		BurstSize:         10,
		MaxConcurrency:    5,
		QueueSize:         1000,
		ProcessingTimeout: 30 * time.Second,
		AckTimeout:        5 * time.Second,
		DLQEnabled:        true,
		DLQThreshold:      3,
		DLQQueueName:      "dead-letter",
	}
}

// RunSimulation executes a policy simulation
func (ps *PolicySimulator) RunSimulation(ctx context.Context, req *SimulationRequest) (*SimulationResult, error) {
	if err := ps.validateRequest(req); err != nil {
		return nil, err
	}

	result := &SimulationResult{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Policies:    req.Policies,
		Pattern:     req.TrafficPattern,
		Timeline:    make([]TimelineSnapshot, 0),
		Warnings:    make([]string, 0),
		CreatedAt:   time.Now(),
		Status:      StatusRunning,
	}

	if req.Config == nil {
		result.Config = ps.config
	}

	// Store simulation
	ps.mu.Lock()
	ps.simulations[result.ID] = result
	ps.mu.Unlock()

	// Run simulation asynchronously
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ps.mu.Lock()
				result.Status = StatusFailed
				result.Warnings = append(result.Warnings, fmt.Sprintf("Simulation panic: %v", r))
				ps.mu.Unlock()
			}
		}()

		if err := ps.executeSimulation(ctx, result); err != nil {
			ps.mu.Lock()
			result.Status = StatusFailed
			result.Warnings = append(result.Warnings, fmt.Sprintf("Simulation error: %v", err))
			ps.mu.Unlock()
		} else {
			ps.mu.Lock()
			result.Status = StatusCompleted
			ps.mu.Unlock()
		}
	}()

	return result, nil
}

// executeSimulation runs the actual simulation logic
func (ps *PolicySimulator) executeSimulation(ctx context.Context, result *SimulationResult) error {
	// Initialize queueing model based on policies and traffic pattern
	model := ps.createQueueingModel(result.Policies, result.Pattern)

	// Create simulation state
	state := &SimulationState{
		QueueDepth:     0,
		ActiveWorkers:  0,
		TotalProcessed: 0,
		TotalFailed:    0,
		TotalRetries:   0,
		StartTime:      time.Now(),
		MemoryUsage:    0,
		CPUUsage:       0,
	}

	// Run simulation over time
	duration := result.Config.SimulationDuration
	timeStep := result.Config.TimeStep
	steps := int(duration / timeStep)

	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Calculate current time in simulation
			currentTime := time.Duration(i) * timeStep

			// Get arrival rate for this time step
			arrivalRate := ps.calculateArrivalRate(result.Pattern, currentTime)

			// Update model with current conditions
			model.ArrivalRate = arrivalRate

			// Run one simulation step
			stepResult := ps.simulateTimeStep(model, result.Policies, timeStep)

			// Update state
			ps.updateSimulationState(state, stepResult)

			// Record timeline snapshot
			snapshot := TimelineSnapshot{
				Timestamp:      result.CreatedAt.Add(currentTime),
				QueueDepth:     state.QueueDepth,
				ActiveWorkers:  state.ActiveWorkers,
				ProcessingRate: stepResult.ProcessingRate,
				FailureRate:    stepResult.FailureRate,
				MemoryUsage:    state.MemoryUsage,
				CPUUsage:       state.CPUUsage,
			}

			result.Timeline = append(result.Timeline, snapshot)
		}
	}

	// Calculate final metrics
	result.Metrics = ps.calculateFinalMetrics(state, result.Timeline)

	// Add model assumptions and warnings
	assumptions := ps.getModelAssumptions(model)
	result.Warnings = append(result.Warnings, assumptions.Limitations...)

	return nil
}

// SimulationState tracks the current state of the simulation
type SimulationState struct {
	QueueDepth     int
	ActiveWorkers  int
	TotalProcessed int
	TotalFailed    int
	TotalRetries   int
	StartTime      time.Time
	MemoryUsage    float64
	CPUUsage       float64
}

// TimeStepResult contains the results of a single simulation step
type TimeStepResult struct {
	ArrivalsThisStep int
	DeparturesThisStep int
	ProcessingRate   float64
	FailureRate      float64
	QueueChange      int
	WorkerUtilization float64
}

// createQueueingModel creates a queueing model based on configuration
func (ps *PolicySimulator) createQueueingModel(policy *PolicyConfig, pattern *TrafficPattern) *QueueingModel {
	// Use M/M/c model (Markovian arrival/service, c servers)
	model := &QueueingModel{
		Type:        ModelMMC,
		ServiceRate: 1.0 / policy.ProcessingTimeout.Seconds(), // messages per second per worker
		ArrivalRate: pattern.BaseRate,
		Servers:     policy.MaxConcurrency,
		Capacity:    policy.QueueSize,
		Parameters:  make(map[string]float64),
	}

	// Add model parameters
	model.Parameters["failure_rate"] = 0.05 // 5% base failure rate
	model.Parameters["retry_multiplier"] = 1.5 // Retry overhead factor

	return model
}

// calculateArrivalRate determines the arrival rate at a given time
func (ps *PolicySimulator) calculateArrivalRate(pattern *TrafficPattern, currentTime time.Duration) float64 {
	baseRate := pattern.BaseRate

	// Apply variations
	for _, variation := range pattern.Variations {
		if currentTime >= variation.StartTime && currentTime <= variation.EndTime {
			baseRate *= variation.Multiplier
		}
	}

	// Add pattern-specific modifications
	switch pattern.Type {
	case TrafficSpike:
		// Add random spikes
		if rand.Float64() < 0.01 { // 1% chance per step
			baseRate *= 3.0
		}
	case TrafficBursty:
		// Add bursts with random intervals
		if rand.Float64() < 0.05 { // 5% chance per step
			baseRate *= 2.0
		}
	case TrafficSeasonal:
		// Sinusoidal pattern
		cycleDuration := pattern.Duration.Seconds()
		if cycleDuration > 0 {
			phase := 2 * math.Pi * currentTime.Seconds() / cycleDuration
			seasonal := 1.0 + 0.5*math.Sin(phase)
			baseRate *= seasonal
		}
	}

	return math.Max(0, baseRate)
}

// simulateTimeStep runs one step of the simulation
func (ps *PolicySimulator) simulateTimeStep(model *QueueingModel, policy *PolicyConfig, timeStep time.Duration) *TimeStepResult {
	// Calculate arrivals using Poisson distribution approximation
	expectedArrivals := model.ArrivalRate * timeStep.Seconds()
	arrivals := ps.poissonSample(expectedArrivals)

	// Calculate service capacity
	serviceRate := model.ServiceRate * float64(model.Servers)
	maxDepartures := int(serviceRate * timeStep.Seconds())

	// Account for retries and failures
	failureRate := model.Parameters["failure_rate"]
	retryMultiplier := model.Parameters["retry_multiplier"]

	// Effective processing rate considering failures and retries
	effectiveRate := serviceRate * (1.0 - failureRate) / retryMultiplier
	departures := int(math.Min(float64(maxDepartures), effectiveRate*timeStep.Seconds()))

	// Calculate utilization
	utilization := math.Min(1.0, model.ArrivalRate/serviceRate)

	return &TimeStepResult{
		ArrivalsThisStep:   arrivals,
		DeparturesThisStep: departures,
		ProcessingRate:     effectiveRate,
		FailureRate:        failureRate,
		QueueChange:        arrivals - departures,
		WorkerUtilization:  utilization,
	}
}

// poissonSample approximates Poisson sampling for small lambda
func (ps *PolicySimulator) poissonSample(lambda float64) int {
	if lambda < 0.1 {
		// For very small lambda, use exact Poisson
		if rand.Float64() < lambda {
			return 1
		}
		return 0
	}

	// For larger lambda, use normal approximation
	return int(math.Max(0, rand.NormFloat64()*math.Sqrt(lambda) + lambda))
}

// updateSimulationState updates the simulation state with step results
func (ps *PolicySimulator) updateSimulationState(state *SimulationState, stepResult *TimeStepResult) {
	// Update queue depth
	state.QueueDepth = int(math.Max(0, float64(state.QueueDepth + stepResult.QueueChange)))

	// Update worker count (simplified)
	state.ActiveWorkers = int(math.Min(float64(state.QueueDepth), stepResult.WorkerUtilization*10))

	// Update counters
	state.TotalProcessed += stepResult.DeparturesThisStep
	state.TotalFailed += int(float64(stepResult.DeparturesThisStep) * stepResult.FailureRate)

	// Estimate resource usage (simplified model)
	state.MemoryUsage = float64(state.QueueDepth)*0.1 + float64(state.ActiveWorkers)*50 // MB
	state.CPUUsage = stepResult.WorkerUtilization * 100 // Percentage
}

// calculateFinalMetrics computes final simulation metrics
func (ps *PolicySimulator) calculateFinalMetrics(state *SimulationState, timeline []TimelineSnapshot) *SimulationMetrics {
	if len(timeline) == 0 {
		return &SimulationMetrics{}
	}

	// Calculate averages and percentiles
	queueDepths := make([]float64, len(timeline))
	waitTimes := make([]float64, len(timeline))
	memoryUsages := make([]float64, len(timeline))

	var totalQueueDepth, totalMemory, totalCPU float64
	var maxQueueDepth int

	for i, snapshot := range timeline {
		queueDepths[i] = float64(snapshot.QueueDepth)
		totalQueueDepth += float64(snapshot.QueueDepth)
		totalMemory += snapshot.MemoryUsage
		totalCPU += snapshot.CPUUsage

		if snapshot.QueueDepth > maxQueueDepth {
			maxQueueDepth = snapshot.QueueDepth
		}

		// Estimate wait time using Little's Law: W = L / λ
		if snapshot.ProcessingRate > 0 {
			waitTimes[i] = float64(snapshot.QueueDepth) / snapshot.ProcessingRate * 1000 // Convert to ms
		}
	}

	// Sort for percentile calculations
	sort.Float64s(waitTimes)

	duration := timeline[len(timeline)-1].Timestamp.Sub(timeline[0].Timestamp).Seconds()

	return &SimulationMetrics{
		AvgQueueDepth:     totalQueueDepth / float64(len(timeline)),
		MaxQueueDepth:     maxQueueDepth,
		AvgWaitTime:       ps.calculateMean(waitTimes),
		P95WaitTime:       ps.calculatePercentile(waitTimes, 0.95),
		P99WaitTime:       ps.calculatePercentile(waitTimes, 0.99),
		MessagesProcessed: state.TotalProcessed,
		ProcessingRate:    float64(state.TotalProcessed) / duration,
		Utilization:       totalCPU / float64(len(timeline)),
		FailureRate:       float64(state.TotalFailed) / float64(state.TotalProcessed) * 100,
		RetryRate:         float64(state.TotalRetries) / float64(state.TotalProcessed) * 100,
		AvgMemoryUsage:    totalMemory / float64(len(timeline)),
		PeakMemoryUsage:   ps.calculateMax(memoryUsages),
		AvgCPUUsage:       totalCPU / float64(len(timeline)),
		SimulationStart:   timeline[0].Timestamp,
		SimulationEnd:     timeline[len(timeline)-1].Timestamp,
		Duration:          duration,
	}
}

// Helper functions for statistical calculations
func (ps *PolicySimulator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (ps *PolicySimulator) calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	index := int(percentile * float64(len(sortedValues)))
	if index >= len(sortedValues) {
		index = len(sortedValues) - 1
	}
	return sortedValues[index]
}

func (ps *PolicySimulator) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

// GetSimulation retrieves a simulation result by ID
func (ps *PolicySimulator) GetSimulation(id string) (*SimulationResult, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result, exists := ps.simulations[id]
	if !exists {
		return nil, ErrSimulationFailed.WithDetails("simulation not found: " + id)
	}

	return result, nil
}

// ListSimulations returns all simulation results
func (ps *PolicySimulator) ListSimulations() []*SimulationResult {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	results := make([]*SimulationResult, 0, len(ps.simulations))
	for _, result := range ps.simulations {
		results = append(results, result)
	}

	// Sort by creation time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results
}

// CreatePolicyChange creates a new policy change proposal
func (ps *PolicySimulator) CreatePolicyChange(description string, changes map[string]interface{}, user string) (*PolicyChange, error) {
	change := &PolicyChange{
		ID:          uuid.New().String(),
		Description: description,
		Changes:     changes,
		AppliedBy:   user,
		Status:      ChangeStatusProposed,
		AuditLog: []AuditEntry{
			{
				Timestamp: time.Now(),
				Action:    "created",
				User:      user,
				Details:   changes,
			},
		},
	}

	ps.mu.Lock()
	ps.changes[change.ID] = change
	ps.mu.Unlock()

	return change, nil
}

// ApplyPolicyChange applies a policy change
func (ps *PolicySimulator) ApplyPolicyChange(changeID string, user string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	change, exists := ps.changes[changeID]
	if !exists {
		return ErrChangeNotFound.WithDetails(changeID)
	}

	if change.Status != ChangeStatusApproved {
		return ErrApplyFailed.WithDetails("change must be approved before applying")
	}

	// Store previous values for rollback
	change.PreviousValues = make(map[string]interface{})
	// TODO: Implement actual policy retrieval and update

	// Mark as applied
	now := time.Now()
	change.AppliedAt = &now
	change.Status = ChangeStatusApplied

	// Add audit entry
	change.AuditLog = append(change.AuditLog, AuditEntry{
		Timestamp: now,
		Action:    "applied",
		User:      user,
		Details:   change.Changes,
	})

	return nil
}

// RollbackPolicyChange rolls back a policy change
func (ps *PolicySimulator) RollbackPolicyChange(changeID string, user string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	change, exists := ps.changes[changeID]
	if !exists {
		return ErrChangeNotFound.WithDetails(changeID)
	}

	if change.Status != ChangeStatusApplied {
		return ErrRollbackFailed.WithDetails("change must be applied before rollback")
	}

	// TODO: Implement actual policy rollback

	// Mark as rolled back
	now := time.Now()
	change.RolledBackAt = &now
	change.Status = ChangeStatusRolledBack

	// Add audit entry
	change.AuditLog = append(change.AuditLog, AuditEntry{
		Timestamp: now,
		Action:    "rolled_back",
		User:      user,
		Details:   change.PreviousValues,
	})

	return nil
}

// validateRequest validates a simulation request
func (ps *PolicySimulator) validateRequest(req *SimulationRequest) error {
	if req == nil {
		return ErrInvalidConfig.WithDetails("request cannot be nil")
	}

	if req.Name == "" {
		return ErrInvalidConfig.WithDetails("simulation name is required")
	}

	if req.Policies == nil {
		return ErrInvalidPolicy.WithDetails("policy configuration is required")
	}

	if req.TrafficPattern == nil {
		return ErrInvalidTrafficPattern.WithDetails("traffic pattern is required")
	}

	// Validate policy configuration
	if err := ps.validatePolicyConfig(req.Policies); err != nil {
		return err
	}

	// Validate traffic pattern
	if err := ps.validateTrafficPattern(req.TrafficPattern); err != nil {
		return err
	}

	return nil
}

// validatePolicyConfig validates policy configuration
func (ps *PolicySimulator) validatePolicyConfig(policy *PolicyConfig) error {
	if policy.MaxRetries < 0 {
		return ErrInvalidPolicy.WithDetails("max_retries cannot be negative")
	}

	if policy.InitialBackoff <= 0 {
		return ErrInvalidPolicy.WithDetails("initial_backoff must be positive")
	}

	if policy.MaxBackoff < policy.InitialBackoff {
		return ErrInvalidPolicy.WithDetails("max_backoff must be >= initial_backoff")
	}

	if policy.MaxRatePerSecond <= 0 {
		return ErrInvalidPolicy.WithDetails("max_rate_per_second must be positive")
	}

	if policy.MaxConcurrency <= 0 {
		return ErrInvalidPolicy.WithDetails("max_concurrency must be positive")
	}

	return nil
}

// validateTrafficPattern validates traffic pattern configuration
func (ps *PolicySimulator) validateTrafficPattern(pattern *TrafficPattern) error {
	if pattern.BaseRate < 0 {
		return ErrInvalidTrafficPattern.WithDetails("base_rate cannot be negative")
	}

	validTypes := map[TrafficPatternType]bool{
		TrafficConstant:    true,
		TrafficLinear:      true,
		TrafficSpike:       true,
		TrafficSeasonal:    true,
		TrafficBursty:      true,
		TrafficExponential: true,
	}

	if !validTypes[pattern.Type] {
		return ErrInvalidTrafficPattern.WithDetails("invalid traffic pattern type: " + string(pattern.Type))
	}

	return nil
}

// getModelAssumptions returns assumptions and limitations for the model
func (ps *PolicySimulator) getModelAssumptions(model *QueueingModel) *SimulationAssumptions {
	return &SimulationAssumptions{
		ModelType: model.Type,
		Assumptions: []string{
			"Markovian arrival and service processes",
			"Independent service times",
			"Infinite population of potential arrivals",
			"FIFO queue discipline",
			"Homogeneous servers",
		},
		Limitations: []string{
			"Does not account for cold starts or warmup time",
			"Simplified failure model",
			"No modeling of Redis network latency",
			"Worker startup/shutdown time not modeled",
			"Memory and CPU estimates are approximations",
		},
		AccuracyEstimate:   "±20% for steady-state metrics under normal conditions",
		RecommendedUseCase: "Steady-state analysis and relative comparison of policies",
		NotRecommendedFor: []string{
			"Precise absolute predictions",
			"Modeling of rare failure modes",
			"Cold start performance",
			"Network partition scenarios",
		},
	}
}