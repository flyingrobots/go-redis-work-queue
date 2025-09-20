//go:build capacity_planning_tests
// +build capacity_planning_tests

// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"testing"
	"time"
)

func TestNewSimulator(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config)
	if sim == nil {
		t.Fatal("NewSimulator returned nil")
	}

	// Test interface compliance
	var _ Simulator = sim
}

func TestSimulateBasic(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config)

	scenario := SimulationScenario{
		Name:        "basic test",
		Description: "Test basic simulation",
		Plan: CapacityPlan{
			CurrentWorkers: 5,
			TargetWorkers:  7,
			Steps: []ScalingStep{
				{
					Sequence:    1,
					ScheduledAt: time.Now().Add(30 * time.Minute),
					Action:      ScaleUp,
					FromWorkers: 5,
					ToWorkers:   7,
					Delta:       2,
				},
			},
		},
		TrafficPattern: TrafficPattern{
			Type:     PatternConstant,
			BaseRate: 10.0,
			Noise:    0.1,
		},
		Duration:    2 * time.Hour,
		Granularity: 5 * time.Minute,
	}

	ctx := context.Background()
	result, err := sim.Simulate(ctx, scenario)

	if err != nil {
		t.Fatalf("Simulate() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Simulate() returned nil result")
	}

	// Verify basic structure
	if result.Scenario.Name != scenario.Name {
		t.Errorf("Result scenario name = %v, want %v", result.Scenario.Name, scenario.Name)
	}

	if len(result.Timeline) == 0 {
		t.Error("Simulate() returned empty timeline")
	}

	expectedPoints := int(scenario.Duration / scenario.Granularity)
	if len(result.Timeline) != expectedPoints {
		t.Errorf("Timeline length = %d, want %d", len(result.Timeline), expectedPoints)
	}

	// Verify timeline progression
	for i, point := range result.Timeline {
		if i > 0 {
			prevPoint := result.Timeline[i-1]
			if !point.Timestamp.After(prevPoint.Timestamp) {
				t.Errorf("Timeline[%d] timestamp not after previous", i)
			}
		}

		if point.Workers < 0 {
			t.Errorf("Timeline[%d] workers = %d, want >= 0", i, point.Workers)
		}

		if point.ArrivalRate < 0 {
			t.Errorf("Timeline[%d] arrival rate = %v, want >= 0", i, point.ArrivalRate)
		}

		if point.ServiceRate < 0 {
			t.Errorf("Timeline[%d] service rate = %v, want >= 0", i, point.ServiceRate)
		}

		if point.Cost < 0 {
			t.Errorf("Timeline[%d] cost = %v, want >= 0", i, point.Cost)
		}
	}

	// Verify scaling happened
	foundScaling := false
	for _, point := range result.Timeline {
		if point.Workers == 7 {
			foundScaling = true
			break
		}
	}
	if !foundScaling {
		t.Error("Expected to find scaling to 7 workers in timeline")
	}
}

func TestTrafficPatternConstant(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	pattern := TrafficPattern{
		Type:     PatternConstant,
		BaseRate: 15.0,
		Noise:    0.0, // No noise for predictable testing
	}

	rate := sim.generateTraffic(time.Now(), pattern)
	if rate != 15.0 {
		t.Errorf("Constant pattern rate = %v, want 15.0", rate)
	}

	// Test with noise
	pattern.Noise = 0.1
	rate = sim.generateTraffic(time.Now(), pattern)
	if rate < 10.0 || rate > 20.0 {
		t.Errorf("Constant pattern with noise rate = %v, want ~15.0 ± reasonable variance", rate)
	}
}

func TestTrafficPatternSinusoidal(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	pattern := TrafficPattern{
		Type:      PatternSinusoidal,
		BaseRate:  10.0,
		Amplitude: 5.0,
		Period:    24 * time.Hour,
		Noise:     0.0,
	}

	baseTime := time.Now().Truncate(24 * time.Hour) // Start of day

	// Test at different times of day
	times := []struct {
		offset    time.Duration
		expected  float64
		tolerance float64
	}{
		{0 * time.Hour, 10.0, 1.0},  // Start: base rate
		{6 * time.Hour, 15.0, 1.0},  // Quarter cycle: base + amplitude
		{12 * time.Hour, 10.0, 1.0}, // Half cycle: base rate
		{18 * time.Hour, 5.0, 1.0},  // Three-quarter cycle: base - amplitude
	}

	for _, tc := range times {
		testTime := baseTime.Add(tc.offset)
		rate := sim.generateTraffic(testTime, pattern)

		if rate < 0 {
			t.Errorf("Sinusoidal rate at %v = %v, want >= 0", tc.offset, rate)
		}

		// Allow some tolerance for sinusoidal approximation
		if rate < tc.expected-tc.tolerance || rate > tc.expected+tc.tolerance {
			t.Errorf("Sinusoidal rate at %v = %v, want %v ± %v", tc.offset, rate, tc.expected, tc.tolerance)
		}
	}
}

func TestTrafficPatternDaily(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	pattern := TrafficPattern{
		Type:     PatternDaily,
		BaseRate: 10.0,
		Noise:    0.0,
	}

	// Test business hours vs night hours
	businessHour := time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC) // 2 PM
	nightHour := time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC)     // 3 AM

	businessRate := sim.generateTraffic(businessHour, pattern)
	nightRate := sim.generateTraffic(nightHour, pattern)

	if businessRate <= nightRate {
		t.Errorf("Business hour rate %v should be higher than night rate %v",
			businessRate, nightRate)
	}

	if businessRate <= 0 || nightRate <= 0 {
		t.Errorf("Rates should be positive: business=%v, night=%v", businessRate, nightRate)
	}
}

func TestTrafficPatternSpiky(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	spikeStart := time.Now()
	pattern := TrafficPattern{
		Type:     PatternSpiky,
		BaseRate: 10.0,
		Spikes: []TrafficSpike{
			{
				StartTime: spikeStart,
				Duration:  1 * time.Hour,
				Magnitude: 3.0, // 3x multiplier
				Shape:     SpikeInstant,
			},
		},
		Noise: 0.0,
	}

	// Test before spike
	beforeSpike := spikeStart.Add(-10 * time.Minute)
	rate := sim.generateTraffic(beforeSpike, pattern)
	if rate != 10.0 {
		t.Errorf("Rate before spike = %v, want 10.0", rate)
	}

	// Test during spike
	duringSpike := spikeStart.Add(30 * time.Minute)
	rate = sim.generateTraffic(duringSpike, pattern)
	if rate != 30.0 { // 10.0 * 3.0
		t.Errorf("Rate during spike = %v, want 30.0", rate)
	}

	// Test after spike
	afterSpike := spikeStart.Add(2 * time.Hour)
	rate = sim.generateTraffic(afterSpike, pattern)
	if rate != 10.0 {
		t.Errorf("Rate after spike = %v, want 10.0", rate)
	}
}

func TestSpikeShapes(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	spikeStart := time.Now()
	spikeDuration := 1 * time.Hour

	shapes := []SpikeShape{SpikeInstant, SpikeLinear, SpikeExp, SpikeBell}

	for _, shape := range shapes {
		spike := TrafficSpike{
			StartTime: spikeStart,
			Duration:  spikeDuration,
			Magnitude: 2.0,
			Shape:     shape,
		}

		// Test at spike midpoint
		midPoint := spikeStart.Add(spikeDuration / 2)
		multiplier := sim.calculateSpikeMultiplier(midPoint, spike)

		if multiplier < 1.0 {
			t.Errorf("Spike multiplier for %v = %v, want >= 1.0", shape, multiplier)
		}

		if shape == SpikeInstant && multiplier != 2.0 {
			t.Errorf("Instant spike multiplier = %v, want 2.0", multiplier)
		}

		// Test at spike start
		startMultiplier := sim.calculateSpikeMultiplier(spikeStart, spike)
		if startMultiplier < 1.0 {
			t.Errorf("Spike multiplier at start for %v = %v, want >= 1.0", shape, startMultiplier)
		}

		// Test at spike end
		endTime := spikeStart.Add(spikeDuration)
		endMultiplier := sim.calculateSpikeMultiplier(endTime, spike)
		if shape != SpikeInstant && endMultiplier > 1.5 {
			t.Errorf("Spike multiplier at end for %v = %v, want decay", shape, endMultiplier)
		}
	}
}

func TestSLOViolationDetection(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	slo := &SLO{
		P95Latency: 2 * time.Second,
		MaxBacklog: 100,
	}

	scenario := SimulationScenario{
		SLOOverride: slo,
	}

	state := &simulationState{
		workers: 5,
		backlog: 50,
	}

	// Test within SLO
	latency := 1 * time.Second
	violation := sim.checkSLOViolation(state, latency, scenario)
	if violation {
		t.Error("Expected no SLO violation for normal conditions")
	}

	// Test latency violation
	latency = 5 * time.Second
	violation = sim.checkSLOViolation(state, latency, scenario)
	if !violation {
		t.Error("Expected SLO violation for high latency")
	}

	// Test backlog violation
	latency = 1 * time.Second
	state.backlog = 200
	violation = sim.checkSLOViolation(state, latency, scenario)
	if !violation {
		t.Error("Expected SLO violation for high backlog")
	}
}

func TestCostCalculation(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	state := &simulationState{
		workers: 10,
	}

	scenario := SimulationScenario{
		SLOOverride: &SLO{P95Latency: 2 * time.Second},
	}

	granularity := 1 * time.Hour

	// Test normal cost (no violations)
	cost := sim.calculateCost(state, scenario, granularity)
	expectedWorkerCost := 10 * 0.50 * 1.0 // 10 workers * $0.50/hour * 1 hour
	if cost != expectedWorkerCost {
		t.Errorf("Normal cost = %v, want %v", cost, expectedWorkerCost)
	}

	// Test with SLO violation (would need to modify to force violation)
	// This is a simplified test - in practice, violation detection is more complex
}

func TestSimulationSummary(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	timeline := []SimulationPoint{
		{Backlog: 50, Latency: 1 * time.Second, Cost: 10.0, Utilization: 0.5, SLOViolation: false},
		{Backlog: 75, Latency: 2 * time.Second, Cost: 12.0, Utilization: 0.7, SLOViolation: false},
		{Backlog: 100, Latency: 3 * time.Second, Cost: 15.0, Utilization: 0.8, SLOViolation: true},
		{Backlog: 60, Latency: 1.5 * time.Second, Cost: 11.0, Utilization: 0.6, SLOViolation: false},
	}

	summary := sim.generateSummary(timeline)

	// Test averages
	expectedAvgBacklog := (50 + 75 + 100 + 60) / 4.0
	if summary.AvgBacklog != expectedAvgBacklog {
		t.Errorf("AvgBacklog = %v, want %v", summary.AvgBacklog, expectedAvgBacklog)
	}

	expectedMaxBacklog := 100
	if summary.MaxBacklog != expectedMaxBacklog {
		t.Errorf("MaxBacklog = %v, want %v", summary.MaxBacklog, expectedMaxBacklog)
	}

	expectedTotalCost := 10.0 + 12.0 + 15.0 + 11.0
	if summary.TotalCost != expectedTotalCost {
		t.Errorf("TotalCost = %v, want %v", summary.TotalCost, expectedTotalCost)
	}

	expectedSLOViolations := 1
	if summary.SLOViolations != expectedSLOViolations {
		t.Errorf("SLOViolations = %v, want %v", summary.SLOViolations, expectedSLOViolations)
	}

	expectedSLOAchievement := 0.75 // 3 out of 4 points compliant
	if summary.SLOAchievement != expectedSLOAchievement {
		t.Errorf("SLOAchievement = %v, want %v", summary.SLOAchievement, expectedSLOAchievement)
	}

	// P95 latency should be the 95th percentile (3rd element when sorted)
	if summary.P95Latency != 3*time.Second {
		t.Errorf("P95Latency = %v, want %v", summary.P95Latency, 3*time.Second)
	}
}

func TestValidateScenario(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config)

	tests := []struct {
		name     string
		scenario SimulationScenario
		wantErr  bool
	}{
		{
			name: "valid scenario",
			scenario: SimulationScenario{
				Duration:    2 * time.Hour,
				Granularity: 5 * time.Minute,
				TrafficPattern: TrafficPattern{
					BaseRate: 10.0,
				},
			},
			wantErr: false,
		},
		{
			name: "zero duration",
			scenario: SimulationScenario{
				Duration:    0,
				Granularity: 5 * time.Minute,
				TrafficPattern: TrafficPattern{
					BaseRate: 10.0,
				},
			},
			wantErr: true,
		},
		{
			name: "granularity exceeds duration",
			scenario: SimulationScenario{
				Duration:    1 * time.Hour,
				Granularity: 2 * time.Hour,
				TrafficPattern: TrafficPattern{
					BaseRate: 10.0,
				},
			},
			wantErr: true,
		},
		{
			name: "negative base rate",
			scenario: SimulationScenario{
				Duration:    2 * time.Hour,
				Granularity: 5 * time.Minute,
				TrafficPattern: TrafficPattern{
					BaseRate: -5.0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sim.ValidateScenario(tt.scenario)
			if tt.wantErr {
				if err == nil {
					t.Error("ValidateScenario() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateScenario() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEstimateRuntime(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config)

	scenario := SimulationScenario{
		Duration:    1 * time.Hour,
		Granularity: 5 * time.Minute,
	}

	runtime := sim.EstimateRuntime(scenario)

	expectedSteps := int(scenario.Duration / scenario.Granularity) // 12 steps
	if runtime <= 0 {
		t.Error("EstimateRuntime() should return positive duration")
	}

	// Should be proportional to number of steps
	if runtime > 10*time.Second {
		t.Errorf("EstimateRuntime() = %v, seems too high for %d steps", runtime, expectedSteps)
	}
}

func TestContextCancellation(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config)

	scenario := SimulationScenario{
		Name:        "cancellation test",
		Duration:    10 * time.Hour, // Long duration
		Granularity: 1 * time.Minute,
		TrafficPattern: TrafficPattern{
			Type:     PatternConstant,
			BaseRate: 10.0,
		},
		Plan: CapacityPlan{
			CurrentWorkers: 5,
			TargetWorkers:  5,
		},
	}

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := sim.Simulate(ctx, scenario)

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestQueueStateEvolution(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	state := &simulationState{
		workers: 5,
		backlog: 10,
	}

	arrivalRate := 15.0 // jobs/second
	serviceRate := 10.0 // jobs/second (total capacity)
	granularity := 1 * time.Second

	// Update state
	sim.updateQueueState(state, arrivalRate, serviceRate, granularity)

	// In 1 second: 15 arrivals, can process up to 10
	// New backlog should be roughly: 10 + 15 - 10 = 15
	// (allowing for Poisson variation)
	if state.backlog < 10 || state.backlog > 25 {
		t.Errorf("Backlog after update = %d, want ~15 ± reasonable variance", state.backlog)
	}

	if state.utilization <= 0 || state.utilization > 2.0 {
		t.Errorf("Utilization = %v, want reasonable positive value", state.utilization)
	}

	if state.activeJobs < 0 || state.activeJobs > state.workers {
		t.Errorf("ActiveJobs = %d, want 0 <= activeJobs <= %d", state.activeJobs, state.workers)
	}
}

func TestViolationTracking(t *testing.T) {
	config := PlannerConfig{}
	sim := NewSimulator(config).(*simulator)

	violationPeriods := make([]ViolationPeriod, 0)
	scenario := SimulationScenario{
		SLOOverride: &SLO{
			P95Latency: 2 * time.Second,
			MaxBacklog: 50,
		},
	}

	baseTime := time.Now()

	// Normal point (no violation)
	point1 := SimulationPoint{
		Timestamp:    baseTime,
		Latency:      1 * time.Second,
		Backlog:      30,
		SLOViolation: false,
	}
	sim.trackViolations(&violationPeriods, point1, scenario)

	if len(violationPeriods) != 0 {
		t.Error("Should not create violation period for compliant point")
	}

	// Violation starts
	point2 := SimulationPoint{
		Timestamp:    baseTime.Add(1 * time.Minute),
		Latency:      5 * time.Second,
		Backlog:      80,
		SLOViolation: true,
	}
	sim.trackViolations(&violationPeriods, point2, scenario)

	if len(violationPeriods) != 1 {
		t.Errorf("Should create 1 violation period, got %d", len(violationPeriods))
	}

	if violationPeriods[0].Start != point2.Timestamp {
		t.Error("Violation start time incorrect")
	}

	// Violation continues
	point3 := SimulationPoint{
		Timestamp:    baseTime.Add(2 * time.Minute),
		Latency:      4 * time.Second,
		Backlog:      90,
		SLOViolation: true,
	}
	sim.trackViolations(&violationPeriods, point3, scenario)

	if len(violationPeriods) != 1 {
		t.Error("Should not create new violation period for continuing violation")
	}

	// Violation ends
	point4 := SimulationPoint{
		Timestamp:    baseTime.Add(3 * time.Minute),
		Latency:      1 * time.Second,
		Backlog:      20,
		SLOViolation: false,
	}
	sim.trackViolations(&violationPeriods, point4, scenario)

	if len(violationPeriods) != 1 {
		t.Error("Should still have 1 violation period")
	}

	if violationPeriods[0].End != point4.Timestamp {
		t.Error("Violation end time incorrect")
	}

	expectedDuration := point4.Timestamp.Sub(point2.Timestamp)
	if violationPeriods[0].Duration != expectedDuration {
		t.Errorf("Violation duration = %v, want %v", violationPeriods[0].Duration, expectedDuration)
	}
}
