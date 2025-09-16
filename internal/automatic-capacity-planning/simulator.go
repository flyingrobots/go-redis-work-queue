// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Simulator interface defines what-if analysis capabilities
type Simulator interface {
	Simulate(ctx context.Context, scenario SimulationScenario) (*Simulation, error)
	ValidateScenario(scenario SimulationScenario) error
	EstimateRuntime(scenario SimulationScenario) time.Duration
}

// simulator implements discrete event simulation for capacity planning
type simulator struct {
	config PlannerConfig
	random *rand.Rand
}

// NewSimulator creates a new simulation engine
func NewSimulator(config PlannerConfig) Simulator {
	return &simulator{
		config: config,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Simulate runs a discrete event simulation of the capacity plan
func (s *simulator) Simulate(ctx context.Context, scenario SimulationScenario) (*Simulation, error) {
	if err := s.ValidateScenario(scenario); err != nil {
		return nil, NewPlannerError(ErrConfigInvalid, "invalid simulation scenario", err)
	}

	// Initialize simulation state
	state := s.initializeState(scenario)
	timeline := make([]SimulationPoint, 0)
	violationPeriods := make([]ViolationPeriod, 0)

	// Run simulation loop
	granularity := scenario.Granularity
	if granularity == 0 {
		granularity = 5 * time.Minute // Default granularity
	}

	steps := int(scenario.Duration / granularity)
	currentTime := time.Now()

	for step := 0; step < steps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Apply scaling actions from plan
		s.applyScalingActions(state, currentTime, scenario.Plan)

		// Generate traffic according to pattern
		arrivalRate := s.generateTraffic(currentTime, scenario.TrafficPattern)

		// Calculate service rate based on current workers
		serviceRate := s.calculateServiceRate(state, scenario)

		// Update queue state
		s.updateQueueState(state, arrivalRate, serviceRate, granularity)

		// Calculate current latency
		latency := s.calculateLatency(state, arrivalRate, serviceRate)

		// Check for SLO violations
		sloViolation := s.checkSLOViolation(state, latency, scenario)

		// Calculate cost
		cost := s.calculateCost(state, scenario, granularity)

		// Record simulation point
		point := SimulationPoint{
			Timestamp:    currentTime,
			Workers:      state.workers,
			ArrivalRate:  arrivalRate,
			ServiceRate:  serviceRate,
			Backlog:      state.backlog,
			Latency:      latency,
			Utilization:  arrivalRate / serviceRate,
			Cost:         cost,
			SLOViolation: sloViolation,
		}

		timeline = append(timeline, point)

		// Track violation periods
		s.trackViolations(&violationPeriods, point, scenario)

		// Advance time
		currentTime = currentTime.Add(granularity)
	}

	// Generate summary and analysis
	summary := s.generateSummary(timeline)
	sloAnalysis := s.analyzeSLOCompliance(timeline, violationPeriods, scenario)
	costAnalysis := s.analyzeCosts(timeline, scenario)

	return &Simulation{
		ID:           generateSimulationID(),
		Scenario:     scenario,
		Timeline:     timeline,
		Summary:      summary,
		SLOAnalysis:  sloAnalysis,
		CostAnalysis: costAnalysis,
		Duration:     scenario.Duration,
		CreatedAt:    time.Now(),
	}, nil
}

// simulationState tracks the evolving state during simulation
type simulationState struct {
	workers         int
	backlog         int
	lastScalingTime time.Time
	totalCost       float64
	activeJobs      int
	utilization     float64
}

// initializeState sets up initial simulation state
func (s *simulator) initializeState(scenario SimulationScenario) *simulationState {
	return &simulationState{
		workers:         scenario.Plan.CurrentWorkers,
		backlog:         0,
		lastScalingTime: time.Time{},
		totalCost:       0,
		activeJobs:      0,
		utilization:     0,
	}
}

// applyScalingActions applies scaling steps from the capacity plan
func (s *simulator) applyScalingActions(state *simulationState, currentTime time.Time, plan CapacityPlan) {
	for _, step := range plan.Steps {
		if step.ScheduledAt.After(currentTime.Add(-time.Minute)) &&
		   step.ScheduledAt.Before(currentTime.Add(time.Minute)) {
			if currentTime.After(step.CooldownUntil) {
				state.workers = step.ToWorkers
				state.lastScalingTime = currentTime
			}
		}
	}
}

// generateTraffic creates arrival rate based on traffic pattern
func (s *simulator) generateTraffic(t time.Time, pattern TrafficPattern) float64 {
	baseRate := pattern.BaseRate

	switch pattern.Type {
	case PatternConstant:
		return baseRate + s.addNoise(baseRate, pattern.Noise)

	case PatternSinusoidal:
		// Sinusoidal pattern: base + amplitude * sin(2π * t / period)
		periodSeconds := pattern.Period.Seconds()
		timeSeconds := float64(t.Unix()) // Use Unix timestamp for deterministic results
		cyclePosition := math.Mod(timeSeconds, periodSeconds) / periodSeconds
		sinValue := math.Sin(2 * math.Pi * cyclePosition)
		rate := baseRate + pattern.Amplitude*sinValue
		return math.Max(0, rate+s.addNoise(rate, pattern.Noise))

	case PatternLinear:
		// Linear trend over time
		hoursElapsed := time.Since(t.Truncate(24*time.Hour)).Hours()
		rate := baseRate + pattern.Trend*hoursElapsed
		return math.Max(0, rate+s.addNoise(rate, pattern.Noise))

	case PatternDaily:
		// Daily pattern based on hour of day
		hour := t.Hour()
		dailyMultiplier := s.getDailyMultiplier(hour)
		rate := baseRate * dailyMultiplier
		return math.Max(0, rate+s.addNoise(rate, pattern.Noise))

	case PatternWeekly:
		// Weekly pattern based on day of week
		weekday := t.Weekday()
		weeklyMultiplier := s.getWeeklyMultiplier(weekday)
		rate := baseRate * weeklyMultiplier
		return math.Max(0, rate+s.addNoise(rate, pattern.Noise))

	case PatternSpiky:
		// Apply spikes on top of base rate
		rate := baseRate
		for _, spike := range pattern.Spikes {
			if t.After(spike.StartTime) && t.Before(spike.StartTime.Add(spike.Duration)) {
				spikeMultiplier := s.calculateSpikeMultiplier(t, spike)
				rate *= spikeMultiplier
			}
		}
		return math.Max(0, rate+s.addNoise(rate, pattern.Noise))

	default:
		return baseRate + s.addNoise(baseRate, pattern.Noise)
	}
}

// addNoise adds random variation to the traffic rate
func (s *simulator) addNoise(rate, noiseLevel float64) float64 {
	if noiseLevel <= 0 {
		return 0
	}
	// Gaussian noise with standard deviation = rate * noiseLevel
	noise := s.random.NormFloat64() * rate * noiseLevel
	return noise
}

// getDailyMultiplier returns traffic multiplier based on hour of day
func (s *simulator) getDailyMultiplier(hour int) float64 {
	// Typical daily pattern: low at night, peak during business hours
	pattern := []float64{
		0.3, 0.2, 0.1, 0.1, 0.2, 0.4, // 0-5 AM
		0.7, 1.0, 1.3, 1.5, 1.4, 1.2, // 6-11 AM
		1.1, 1.3, 1.4, 1.5, 1.3, 1.1, // 12-5 PM
		0.9, 0.8, 0.7, 0.6, 0.5, 0.4, // 6-11 PM
	}
	if hour >= 0 && hour < len(pattern) {
		return pattern[hour]
	}
	return 1.0
}

// getWeeklyMultiplier returns traffic multiplier based on day of week
func (s *simulator) getWeeklyMultiplier(weekday time.Weekday) float64 {
	// Typical weekly pattern: lower on weekends
	pattern := map[time.Weekday]float64{
		time.Sunday:    0.6,
		time.Monday:    1.2,
		time.Tuesday:   1.3,
		time.Wednesday: 1.4,
		time.Thursday:  1.3,
		time.Friday:    1.1,
		time.Saturday:  0.7,
	}
	return pattern[weekday]
}

// calculateSpikeMultiplier computes spike intensity based on shape and timing
func (s *simulator) calculateSpikeMultiplier(t time.Time, spike TrafficSpike) float64 {
	elapsed := t.Sub(spike.StartTime)
	progress := elapsed.Seconds() / spike.Duration.Seconds()

	switch spike.Shape {
	case SpikeInstant:
		return spike.Magnitude

	case SpikeLinear:
		if progress < 0.5 {
			// Ramp up
			return 1.0 + (spike.Magnitude-1.0)*progress*2
		} else {
			// Ramp down
			return spike.Magnitude - (spike.Magnitude-1.0)*(progress-0.5)*2
		}

	case SpikeExp:
		// Exponential growth and decay
		if progress < 0.5 {
			return 1.0 + (spike.Magnitude-1.0)*math.Pow(progress*2, 2)
		} else {
			decay := (progress - 0.5) * 2
			return spike.Magnitude * math.Exp(-4*decay)
		}

	case SpikeBell:
		// Bell curve (Gaussian)
		center := 0.5
		sigma := 0.2
		bellValue := math.Exp(-0.5 * math.Pow((progress-center)/sigma, 2))
		return 1.0 + (spike.Magnitude-1.0)*bellValue

	default:
		return spike.Magnitude
	}
}

// calculateServiceRate determines effective service rate based on workers
func (s *simulator) calculateServiceRate(state *simulationState, scenario SimulationScenario) float64 {
	if state.workers <= 0 {
		return 0
	}

	// Base service rate per worker (μ)
	baseServiceRate := 1.0 // jobs per second per worker
	if scenario.Plan.ForecastWindow > 0 {
		// Use estimated service rate from metrics if available
		// This would come from the actual metrics in a real implementation
		baseServiceRate = 0.5 // Conservative estimate
	}

	// Total service rate = workers × μ
	totalServiceRate := float64(state.workers) * baseServiceRate

	// Apply efficiency degradation for high utilization
	utilization := state.utilization
	if utilization > 0.8 {
		// Efficiency drops as utilization approaches 1.0
		efficiency := 1.0 - 0.3*math.Pow((utilization-0.8)/0.2, 2)
		totalServiceRate *= math.Max(0.5, efficiency)
	}

	return totalServiceRate
}

// updateQueueState evolves the queue state based on arrivals and service
func (s *simulator) updateQueueState(state *simulationState, arrivalRate, serviceRate float64, granularity time.Duration) {
	// Convert rates to job counts for this time period
	intervalSeconds := granularity.Seconds()
	arrivals := int(arrivalRate * intervalSeconds)

	// Add Poisson noise to arrivals
	if arrivalRate > 0 {
		arrivals = s.poissonRandom(arrivalRate * intervalSeconds)
	}

	// Service capacity for this interval
	serviceCapacity := int(serviceRate * intervalSeconds)

	// Process jobs: can't serve more than available
	jobsToProcess := min(state.backlog+arrivals, serviceCapacity)

	// Update backlog
	state.backlog = state.backlog + arrivals - jobsToProcess
	if state.backlog < 0 {
		state.backlog = 0
	}

	// Update active jobs (currently being processed)
	state.activeJobs = min(jobsToProcess, state.workers)

	// Update utilization
	if serviceRate > 0 {
		state.utilization = arrivalRate / serviceRate
	} else {
		state.utilization = 1.0
	}
}

// poissonRandom generates a Poisson-distributed random number
func (s *simulator) poissonRandom(lambda float64) int {
	if lambda <= 0 {
		return 0
	}

	// Use Knuth's algorithm for small lambda
	if lambda < 30 {
		l := math.Exp(-lambda)
		k := 0
		p := 1.0

		for p > l {
			k++
			p *= s.random.Float64()
		}
		return k - 1
	}

	// For large lambda, use normal approximation
	return int(math.Max(0, s.random.NormFloat64()*math.Sqrt(lambda)+lambda))
}

// calculateLatency estimates current system latency using queueing theory
func (s *simulator) calculateLatency(state *simulationState, arrivalRate, serviceRate float64) time.Duration {
	if serviceRate <= 0 || arrivalRate >= serviceRate {
		return time.Hour // High latency for unstable system
	}

	// Use M/M/c approximation for latency
	utilization := arrivalRate / serviceRate
	if utilization >= 1.0 {
		return time.Hour
	}

	// Wait time in queue (W_q) using M/M/c formula approximation
	servers := state.workers
	if servers <= 0 {
		return time.Hour
	}

	// Simplified M/M/c wait time calculation
	mu := serviceRate / float64(servers) // Service rate per server
	rho := arrivalRate / (float64(servers) * mu)

	if rho >= 1.0 {
		return time.Hour
	}

	// Approximate queue wait time
	waitTimeSeconds := rho / (mu * (1 - rho))

	// Add service time
	serviceTimeSeconds := 1.0 / mu
	totalLatency := waitTimeSeconds + serviceTimeSeconds

	return time.Duration(totalLatency * float64(time.Second))
}

// checkSLOViolation determines if current state violates SLO
func (s *simulator) checkSLOViolation(state *simulationState, latency time.Duration, scenario SimulationScenario) bool {
	slo := scenario.SLOOverride
	if slo == nil {
		// Use default SLO if not overridden
		slo = &SLO{
			P95Latency: 5 * time.Second,
			MaxBacklog: 1000,
		}
	}

	// Check latency violation
	if latency > slo.P95Latency {
		return true
	}

	// Check backlog violation
	if state.backlog > slo.MaxBacklog {
		return true
	}

	return false
}

// calculateCost computes cost for this simulation interval
func (s *simulator) calculateCost(state *simulationState, scenario SimulationScenario, granularity time.Duration) float64 {
	hoursElapsed := granularity.Hours()
	workerCost := float64(state.workers) * 0.50 * hoursElapsed // $0.50/worker/hour default

	// Add violation cost if SLO is breached
	violationCost := 0.0
	if s.checkSLOViolation(state, s.calculateLatency(state, 0, 1), scenario) {
		violationCost = 100.0 * hoursElapsed // $100/hour violation cost
	}

	return workerCost + violationCost
}

// trackViolations maintains a record of SLO violation periods
func (s *simulator) trackViolations(violationPeriods *[]ViolationPeriod, point SimulationPoint, scenario SimulationScenario) {
	if !point.SLOViolation {
		// End any ongoing violation
		if len(*violationPeriods) > 0 {
			lastViolation := &(*violationPeriods)[len(*violationPeriods)-1]
			if lastViolation.End.IsZero() {
				lastViolation.End = point.Timestamp
				lastViolation.Duration = lastViolation.End.Sub(lastViolation.Start)
			}
		}
		return
	}

	// Start new violation or continue existing one
	if len(*violationPeriods) == 0 || !(*violationPeriods)[len(*violationPeriods)-1].End.IsZero() {
		// Start new violation period
		violationType := "latency"
		severity := "minor"

		slo := scenario.SLOOverride
		if slo != nil && point.Backlog > slo.MaxBacklog {
			violationType = "backlog"
			if point.Backlog > slo.MaxBacklog*2 {
				severity = "major"
			}
		}

		if point.Latency > 30*time.Second {
			severity = "critical"
		}

		violation := ViolationPeriod{
			Start:    point.Timestamp,
			Type:     violationType,
			Severity: severity,
			MaxValue: math.Max(point.Latency.Seconds(), float64(point.Backlog)),
			Impact:   "SLO breach detected",
		}
		*violationPeriods = append(*violationPeriods, violation)
	} else {
		// Update ongoing violation
		lastViolation := &(*violationPeriods)[len(*violationPeriods)-1]
		lastViolation.MaxValue = math.Max(lastViolation.MaxValue,
			math.Max(point.Latency.Seconds(), float64(point.Backlog)))
	}
}

// generateSummary creates aggregate statistics from simulation timeline
func (s *simulator) generateSummary(timeline []SimulationPoint) SimulationSummary {
	if len(timeline) == 0 {
		return SimulationSummary{}
	}

	var totalBacklog, totalLatency, totalCost, totalUtilization float64
	var maxBacklog int
	var sloViolations int
	var latencies []time.Duration

	for _, point := range timeline {
		totalBacklog += float64(point.Backlog)
		totalLatency += point.Latency.Seconds()
		totalCost += point.Cost
		totalUtilization += point.Utilization
		latencies = append(latencies, point.Latency)

		if point.Backlog > maxBacklog {
			maxBacklog = point.Backlog
		}

		if point.SLOViolation {
			sloViolations++
		}
	}

	// Calculate P95 latency
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p95Index := int(0.95 * float64(len(latencies)))
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}

	var p95Latency time.Duration
	if len(latencies) > 0 {
		p95Latency = latencies[p95Index]
	}

	count := float64(len(timeline))
	sloAchievement := 1.0 - (float64(sloViolations) / count)

	// Efficiency score: SLO achievement / cost ratio
	efficiencyScore := sloAchievement / math.Max(totalCost, 1.0) * 1000 // Scale for readability

	return SimulationSummary{
		AvgBacklog:      totalBacklog / count,
		MaxBacklog:      maxBacklog,
		AvgLatency:      time.Duration(totalLatency/count) * time.Second,
		P95Latency:      p95Latency,
		SLOViolations:   sloViolations,
		SLOAchievement:  sloAchievement,
		TotalCost:       totalCost,
		AvgUtilization:  totalUtilization / count,
		EfficiencyScore: efficiencyScore,
	}
}

// analyzeSLOCompliance provides detailed SLO analysis
func (s *simulator) analyzeSLOCompliance(timeline []SimulationPoint, violationPeriods []ViolationPeriod, scenario SimulationScenario) SLOAnalysis {
	if len(timeline) == 0 {
		return SLOAnalysis{}
	}

	slo := scenario.SLOOverride
	if slo == nil {
		slo = &SLO{
			P95Latency:   5 * time.Second,
			MaxBacklog:   1000,
			Availability: 0.99,
		}
	}

	var latencyCompliant, backlogCompliant int
	totalPoints := len(timeline)

	for _, point := range timeline {
		if point.Latency <= slo.P95Latency {
			latencyCompliant++
		}
		if point.Backlog <= slo.MaxBacklog {
			backlogCompliant++
		}
	}

	latencyCompliance := float64(latencyCompliant) / float64(totalPoints)
	backlogCompliance := float64(backlogCompliant) / float64(totalPoints)

	// Calculate availability (% of time without violations)
	violationTime := time.Duration(0)
	for _, violation := range violationPeriods {
		if !violation.End.IsZero() {
			violationTime += violation.Duration
		}
	}

	totalTime := scenario.Duration
	availabilityAchieved := 1.0 - (violationTime.Seconds() / totalTime.Seconds())

	// Error budget usage
	errorBudgetUsed := (1.0 - availabilityAchieved) / (1.0 - slo.Availability)

	// Risk score based on violation severity and frequency
	riskScore := s.calculateRiskScore(violationPeriods, totalTime)

	return SLOAnalysis{
		LatencyCompliance:    latencyCompliance,
		BacklogCompliance:    backlogCompliance,
		AvailabilityAchieved: availabilityAchieved,
		ErrorBudgetUsed:      errorBudgetUsed,
		ViolationPeriods:     violationPeriods,
		RiskScore:            riskScore,
	}
}

// calculateRiskScore assesses overall risk based on violations
func (s *simulator) calculateRiskScore(violationPeriods []ViolationPeriod, totalTime time.Duration) float64 {
	if len(violationPeriods) == 0 {
		return 0.0
	}

	score := 0.0
	for _, violation := range violationPeriods {
		duration := violation.Duration.Seconds()

		// Base score from duration
		durationScore := duration / totalTime.Seconds()

		// Severity multiplier
		severityMultiplier := 1.0
		switch violation.Severity {
		case "minor":
			severityMultiplier = 1.0
		case "major":
			severityMultiplier = 2.0
		case "critical":
			severityMultiplier = 3.0
		}

		score += durationScore * severityMultiplier
	}

	// Normalize to 0-1 scale
	return math.Min(score, 1.0)
}

// analyzeCosts provides cost analysis for the simulation
func (s *simulator) analyzeCosts(timeline []SimulationPoint, scenario SimulationScenario) CostAnalysis {
	if len(timeline) == 0 {
		return CostAnalysis{}
	}

	totalCost := 0.0
	for _, point := range timeline {
		totalCost += point.Cost
	}

	duration := scenario.Duration.Hours()
	currentWorkers := scenario.Plan.CurrentWorkers
	targetWorkers := scenario.Plan.TargetWorkers

	// Calculate cost components
	currentCostPerHour := float64(currentWorkers) * 0.50
	projectedCostPerHour := float64(targetWorkers) * 0.50
	deltaCostPerHour := projectedCostPerHour - currentCostPerHour
	monthlyCostDelta := deltaCostPerHour * 24 * 30

	// Estimate violation cost risk
	violationCostRisk := 0.0
	for _, point := range timeline {
		if point.SLOViolation {
			violationCostRisk += 100.0 // $100/hour violation cost
		}
	}
	violationCostRisk /= duration

	// Net benefit calculation
	netBenefit := -deltaCostPerHour - violationCostRisk

	// Payback period
	paybackPeriod := "N/A"
	if netBenefit > 0 {
		paybackPeriod = "Immediate"
	} else if deltaCostPerHour < 0 {
		months := math.Abs(monthlyCostDelta) / 1000 // Assume $1000 implementation cost
		paybackPeriod = fmt.Sprintf("%.1f months", months)
	}

	return CostAnalysis{
		CurrentCostPerHour:   currentCostPerHour,
		ProjectedCostPerHour: projectedCostPerHour,
		DeltaCostPerHour:     deltaCostPerHour,
		MonthlyCostDelta:     monthlyCostDelta,
		ViolationCostRisk:    violationCostRisk,
		NetBenefit:           netBenefit,
		PaybackPeriod:        paybackPeriod,
	}
}

// ValidateScenario checks if simulation parameters are valid
func (s *simulator) ValidateScenario(scenario SimulationScenario) error {
	if scenario.Duration <= 0 {
		return NewPlannerError(ErrConfigInvalid, "simulation duration must be positive", nil)
	}

	if scenario.Granularity <= 0 {
		scenario.Granularity = 5 * time.Minute // Default
	}

	if scenario.Granularity > scenario.Duration {
		return NewPlannerError(ErrConfigInvalid, "granularity cannot exceed duration", nil)
	}

	if scenario.TrafficPattern.BaseRate < 0 {
		return NewPlannerError(ErrConfigInvalid, "base traffic rate must be non-negative", nil)
	}

	return nil
}

// EstimateRuntime provides an estimate of simulation execution time
func (s *simulator) EstimateRuntime(scenario SimulationScenario) time.Duration {
	steps := int(scenario.Duration / scenario.Granularity)

	// Estimate 100 microseconds per simulation step
	estimatedMicros := steps * 100

	return time.Duration(estimatedMicros) * time.Microsecond
}

// Utility functions

func generateSimulationID() string {
	return fmt.Sprintf("sim_%d", time.Now().UnixNano())
}

