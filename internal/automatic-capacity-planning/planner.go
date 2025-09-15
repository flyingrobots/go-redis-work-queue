// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// CapacityPlanner is the main interface for capacity planning
type CapacityPlanner interface {
	GeneratePlan(ctx context.Context, req PlanRequest) (*PlanResponse, error)
	UpdateConfig(config PlannerConfig) error
	GetState() PlannerState
	ApplyPlan(ctx context.Context, planID string) error
	SimulateWhatIf(ctx context.Context, scenario SimulationScenario) (*Simulation, error)
}

// planner implements the CapacityPlanner interface
type planner struct {
	config     PlannerConfig
	state      PlannerState
	forecaster Forecaster
	simulator  Simulator
	queueing   QueueingCalculator
}

// NewCapacityPlanner creates a new capacity planner instance
func NewCapacityPlanner(config PlannerConfig) CapacityPlanner {
	return &planner{
		config:     config,
		state:      PlannerState{},
		forecaster: NewForecaster(config),
		simulator:  NewSimulator(config),
		queueing:   NewQueueingCalculator(config),
	}
}

// GeneratePlan creates a capacity plan based on current metrics and SLO
func (p *planner) GeneratePlan(ctx context.Context, req PlanRequest) (*PlanResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := p.validateRequest(req); err != nil {
		return nil, NewPlannerError(ErrConfigInvalid, "Invalid request", err)
	}

	// Check if we're in cooldown period
	if time.Now().Before(p.state.CooldownUntil) && !req.ForceRegen {
		return nil, NewPlannerError(ErrCooldownActive, "Planner is in cooldown period", nil)
	}

	// Detect anomalies
	if p.detectAnomalies(req.CurrentMetrics) {
		p.state.AnomalyDetected = true
		p.state.AnomalyStart = time.Now()
		return nil, NewPlannerError(ErrAnomalyDetected, "Traffic anomaly detected, pausing auto-scaling", nil)
	}

	// Update history
	p.updateHistory(req.CurrentMetrics)

	// Generate forecast
	forecast, err := p.forecaster.Predict(ctx, p.state.RecentHistory, req.Config.ForecastWindow)
	if err != nil {
		return nil, NewPlannerError(ErrForecastFailed, "Failed to generate forecast", err)
	}

	// Calculate required capacity using queueing theory
	queueingResult, err := p.calculateCapacity(req.CurrentMetrics, req.SLO, forecast)
	if err != nil {
		return nil, err
	}

	// Generate scaling plan
	plan, err := p.generatePlan(req, queueingResult, forecast)
	if err != nil {
		return nil, err
	}

	// Update state
	p.state.LastPlan = plan
	p.state.LastUpdate = time.Now()

	// Generate recommendations and warnings
	recommendations := p.generateRecommendations(plan, queueingResult)
	warnings := p.generateWarnings(plan, req.CurrentMetrics)

	response := &PlanResponse{
		Plan:           *plan,
		Forecast:       forecast,
		QueueingResult: *queueingResult,
		Recommendations: recommendations,
		Warnings:       warnings,
		GenerationTime: time.Since(startTime),
		CacheHit:       false,
	}

	return response, nil
}

// calculateCapacity uses queueing theory to determine required capacity
func (p *planner) calculateCapacity(metrics Metrics, slo SLO, forecast []Forecast) (*QueueingResult, error) {
	// Use the maximum forecasted arrival rate for safety
	maxLambda := metrics.ArrivalRate
	for _, f := range forecast {
		if f.ArrivalRate > maxLambda {
			maxLambda = f.ArrivalRate
		}
	}

	// Calculate service rate (μ) from service time
	mu := 1.0 / metrics.ServiceTime.Seconds()

	// Start with current workers and iterate to find optimal capacity
	targetCapacity := metrics.CurrentWorkers

	for c := 1; c <= p.config.MaxWorkers; c++ {
		result := p.queueing.Calculate(maxLambda, mu, c, metrics)

		// Check if this capacity meets SLO
		if p.meetsSLO(result, slo) {
			targetCapacity = c
			break
		}

		// If we've exceeded max workers, use max but mark as unachievable
		if c == p.config.MaxWorkers {
			targetCapacity = p.config.MaxWorkers
			result.Confidence = 0.5 // Lower confidence if SLO not achievable
			result.Assumptions = append(result.Assumptions, "SLO may not be achievable with max capacity")
		}
	}

	// Apply safety margin
	safeCapacity := int(float64(targetCapacity) * (1.0 + p.config.SafetyMargin))
	if safeCapacity > p.config.MaxWorkers {
		safeCapacity = p.config.MaxWorkers
	}
	if safeCapacity < p.config.MinWorkers {
		safeCapacity = p.config.MinWorkers
	}

	// Final calculation with safe capacity
	result := p.queueing.Calculate(maxLambda, mu, safeCapacity, metrics)
	result.Capacity = safeCapacity

	return result, nil
}

// generatePlan creates a scaling plan with steps and cooldowns
func (p *planner) generatePlan(req PlanRequest, queueingResult *QueueingResult, forecast []Forecast) (*CapacityPlan, error) {
	currentWorkers := req.CurrentMetrics.CurrentWorkers
	targetWorkers := queueingResult.Capacity
	delta := targetWorkers - currentWorkers

	plan := &CapacityPlan{
		ID:             generatePlanID(),
		GeneratedAt:    time.Now(),
		CurrentWorkers: currentWorkers,
		TargetWorkers:  targetWorkers,
		Delta:          delta,
		Confidence:     queueingResult.Confidence,
		SLOAchievable:  p.meetsSLO(queueingResult, req.SLO),
		ForecastWindow: req.Config.ForecastWindow,
		SafetyMargin:   req.Config.SafetyMargin,
		ValidUntil:     time.Now().Add(req.Config.ForecastWindow),
		QueueName:      req.QueueName,
	}

	// Generate rationale
	plan.Rationale = p.generateRationale(req, queueingResult, forecast)

	// Generate scaling steps
	steps, err := p.generateScalingSteps(currentWorkers, targetWorkers, req.Config)
	if err != nil {
		return nil, err
	}
	plan.Steps = steps

	// Calculate cost impact
	plan.CostImpact = p.calculateCostImpact(currentWorkers, targetWorkers, req.Config)

	return plan, nil
}

// generateScalingSteps creates a sequence of scaling actions with cooldowns
func (p *planner) generateScalingSteps(current, target int, config PlannerConfig) ([]ScalingStep, error) {
	var steps []ScalingStep

	if current == target {
		// No scaling needed
		steps = append(steps, ScalingStep{
			Sequence:      1,
			ScheduledAt:   time.Now(),
			Action:        NoChange,
			FromWorkers:   current,
			ToWorkers:     current,
			Delta:         0,
			Rationale:     "Current capacity meets requirements",
			EstimatedCost: 0,
			Confidence:    1.0,
			CooldownUntil: time.Now(),
		})
		return steps, nil
	}

	delta := target - current
	action := ScaleUp
	if delta < 0 {
		action = ScaleDown
		delta = -delta
	}

	// Break large changes into steps
	maxStep := config.MaxStepSize
	stepsNeeded := int(math.Ceil(float64(delta) / float64(maxStep)))

	currentLevel := current
	scheduledTime := time.Now()

	for i := 0; i < stepsNeeded; i++ {
		stepSize := maxStep
		if i == stepsNeeded-1 {
			// Last step - use remaining delta
			stepSize = target - currentLevel
			if action == ScaleDown {
				stepSize = currentLevel - target
			}
		}

		var newLevel int
		if action == ScaleUp {
			newLevel = currentLevel + stepSize
		} else {
			newLevel = currentLevel - stepSize
			stepSize = -stepSize // Make delta negative for scale down
		}

		step := ScalingStep{
			Sequence:      i + 1,
			ScheduledAt:   scheduledTime,
			Action:        action,
			FromWorkers:   currentLevel,
			ToWorkers:     newLevel,
			Delta:         stepSize,
			Rationale:     p.generateStepRationale(action, stepSize, i+1, stepsNeeded),
			EstimatedCost: config.WorkerCostPerHour * float64(stepSize),
			Confidence:    p.calculateStepConfidence(i, stepsNeeded),
			CooldownUntil: scheduledTime.Add(config.CooldownPeriod),
		}

		steps = append(steps, step)
		currentLevel = newLevel
		scheduledTime = scheduledTime.Add(config.CooldownPeriod)
	}

	return steps, nil
}

// detectAnomalies checks for unusual traffic patterns
func (p *planner) detectAnomalies(metrics Metrics) bool {
	if len(p.state.RecentHistory) < 10 {
		return false // Need enough history for anomaly detection
	}

	// Calculate baseline statistics
	var sum, sumSq float64
	count := len(p.state.RecentHistory)

	for _, m := range p.state.RecentHistory {
		sum += m.ArrivalRate
		sumSq += m.ArrivalRate * m.ArrivalRate
	}

	mean := sum / float64(count)
	variance := (sumSq / float64(count)) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// Z-score based anomaly detection
	zscore := math.Abs(metrics.ArrivalRate - mean) / stdDev
	if zscore > p.config.AnomalyThreshold {
		return true
	}

	// Sudden spike detection
	recentAvg := p.calculateRecentAverage(5) // Last 5 measurements
	if metrics.ArrivalRate > recentAvg * p.config.SpikeThreshold {
		return true
	}

	return false
}

// meetsSLO checks if the queueing result satisfies the SLO
func (p *planner) meetsSLO(result *QueueingResult, slo SLO) bool {
	// Check latency requirement
	if result.ResponseTime > slo.P95Latency {
		return false
	}

	// Check utilization (should be under 100% with margin)
	if result.Utilization > 0.95 {
		return false
	}

	// Additional checks can be added here
	return true
}

// calculateCostImpact computes the financial impact of scaling
func (p *planner) calculateCostImpact(current, target int, config PlannerConfig) CostAnalysis {
	delta := target - current
	currentCost := float64(current) * config.WorkerCostPerHour
	projectedCost := float64(target) * config.WorkerCostPerHour
	deltaCost := float64(delta) * config.WorkerCostPerHour

	return CostAnalysis{
		CurrentCostPerHour:  currentCost,
		ProjectedCostPerHour: projectedCost,
		DeltaCostPerHour:    deltaCost,
		MonthlyCostDelta:    deltaCost * 24 * 30, // 30 days
		ViolationCostRisk:   p.estimateViolationCost(current, target, config),
		NetBenefit:          p.calculateNetBenefit(deltaCost, config),
		PaybackPeriod:       p.calculatePaybackPeriod(deltaCost, config),
	}
}

// Helper methods

func (p *planner) updateHistory(metrics Metrics) {
	p.state.RecentHistory = append(p.state.RecentHistory, metrics)

	// Keep only recent history within the window
	cutoff := time.Now().Add(-p.config.HistoryWindow)
	var filtered []Metrics
	for _, m := range p.state.RecentHistory {
		if m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}
	p.state.RecentHistory = filtered
}

func (p *planner) calculateRecentAverage(count int) float64 {
	if len(p.state.RecentHistory) == 0 {
		return 0
	}

	start := len(p.state.RecentHistory) - count
	if start < 0 {
		start = 0
	}

	var sum float64
	samples := 0
	for i := start; i < len(p.state.RecentHistory); i++ {
		sum += p.state.RecentHistory[i].ArrivalRate
		samples++
	}

	if samples == 0 {
		return 0
	}

	return sum / float64(samples)
}

func (p *planner) validateRequest(req PlanRequest) error {
	if req.CurrentMetrics.ArrivalRate < 0 {
		return fmt.Errorf("invalid arrival rate: %f", req.CurrentMetrics.ArrivalRate)
	}

	if req.CurrentMetrics.ServiceTime <= 0 {
		return fmt.Errorf("invalid service time: %v", req.CurrentMetrics.ServiceTime)
	}

	if req.CurrentMetrics.CurrentWorkers <= 0 {
		return fmt.Errorf("invalid current workers: %d", req.CurrentMetrics.CurrentWorkers)
	}

	if req.SLO.P95Latency <= 0 {
		return fmt.Errorf("invalid SLO latency: %v", req.SLO.P95Latency)
	}

	return nil
}

func (p *planner) generateRationale(req PlanRequest, result *QueueingResult, forecast []Forecast) string {
	maxForecast := req.CurrentMetrics.ArrivalRate
	for _, f := range forecast {
		if f.ArrivalRate > maxForecast {
			maxForecast = f.ArrivalRate
		}
	}

	return fmt.Sprintf(
		"Based on current arrival rate %.1f jobs/sec and forecasted peak %.1f jobs/sec, "+
		"queueing model recommends %d workers to maintain P95 latency below %v "+
		"(utilization: %.1f%%, safety margin: %.1f%%)",
		req.CurrentMetrics.ArrivalRate,
		maxForecast,
		result.Capacity,
		req.SLO.P95Latency,
		result.Utilization*100,
		p.config.SafetyMargin*100,
	)
}

func (p *planner) generateStepRationale(action ScalingAction, delta int, step, total int) string {
	direction := "up"
	if action == ScaleDown {
		direction = "down"
		delta = -delta
	}

	return fmt.Sprintf("Step %d/%d: Scale %s by %d workers", step, total, direction, abs(delta))
}

func (p *planner) calculateStepConfidence(step, total int) float64 {
	// Higher confidence for earlier steps, lower for later steps
	return 1.0 - (float64(step) / float64(total) * 0.3)
}

func (p *planner) generateRecommendations(plan *CapacityPlan, result *QueueingResult) []string {
	var recommendations []string

	if result.Utilization > 0.9 {
		recommendations = append(recommendations, "High utilization detected - consider increasing capacity")
	}

	if plan.Delta > 0 {
		recommendations = append(recommendations, "Scale up recommended to meet SLO requirements")
	} else if plan.Delta < 0 {
		recommendations = append(recommendations, "Scale down possible while maintaining SLO")
	}

	if plan.Confidence < 0.7 {
		recommendations = append(recommendations, "Low confidence - consider manual review before applying")
	}

	return recommendations
}

func (p *planner) generateWarnings(plan *CapacityPlan, metrics Metrics) []string {
	var warnings []string

	if plan.TargetWorkers == p.config.MaxWorkers {
		warnings = append(warnings, "Hitting maximum worker limit - SLO may not be achievable")
	}

	if plan.TargetWorkers == p.config.MinWorkers {
		warnings = append(warnings, "At minimum worker limit - reduced headroom for traffic spikes")
	}

	if len(plan.Steps) > 3 {
		warnings = append(warnings, fmt.Sprintf("Large change requires %d steps - consider increasing max step size", len(plan.Steps)))
	}

	return warnings
}

func (p *planner) estimateViolationCost(current, target int, config PlannerConfig) float64 {
	// Simplified violation cost estimation
	if target > current {
		// Scaling up - lower violation risk
		return config.ViolationCostPerHour * 0.1
	}
	// Scaling down - higher violation risk
	return config.ViolationCostPerHour * 0.3
}

func (p *planner) calculateNetBenefit(deltaCost float64, config PlannerConfig) float64 {
	// Simplified net benefit calculation
	if deltaCost < 0 {
		// Saving money by scaling down
		return -deltaCost
	}
	// Spending money to scale up - benefit is avoiding violations
	return config.ViolationCostPerHour * 0.5 - deltaCost
}

func (p *planner) calculatePaybackPeriod(deltaCost float64, config PlannerConfig) string {
	if deltaCost <= 0 {
		return "Immediate"
	}

	// Simplified payback calculation
	hoursToPayback := deltaCost / (config.ViolationCostPerHour * 0.1)
	if hoursToPayback < 24 {
		return fmt.Sprintf("%.1f hours", hoursToPayback)
	}

	days := hoursToPayback / 24
	return fmt.Sprintf("%.1f days", days)
}

// UpdateConfig updates the planner configuration
func (p *planner) UpdateConfig(config PlannerConfig) error {
	p.config = config
	p.state.ConfigVersion++
	return nil
}

// GetState returns the current planner state
func (p *planner) GetState() PlannerState {
	return p.state
}

// ApplyPlan applies a generated plan (implementation depends on integration)
func (p *planner) ApplyPlan(ctx context.Context, planID string) error {
	// This would integrate with Kubernetes operator or local scaler
	// For now, just mark the time
	p.state.LastScaling = time.Now()
	p.state.CooldownUntil = time.Now().Add(p.config.CooldownPeriod)
	return nil
}

// SimulateWhatIf runs a what-if simulation
func (p *planner) SimulateWhatIf(ctx context.Context, scenario SimulationScenario) (*Simulation, error) {
	return p.simulator.Run(ctx, scenario)
}

// Utility functions

func generatePlanID() string {
	return fmt.Sprintf("plan-%d", time.Now().UnixNano())
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}