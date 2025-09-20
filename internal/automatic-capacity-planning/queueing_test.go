//go:build capacity_planning_tests
// +build capacity_planning_tests

// Copyright 2025 James Ross
package capacityplanning

import (
	"math"
	"testing"
	"time"
)

func TestNewQueueingCalculator(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mmc",
	}

	calc := NewQueueingCalculator(config)
	if calc == nil {
		t.Fatal("NewQueueingCalculator returned nil")
	}

	// Test interface compliance
	var _ QueueingCalculator = calc
}

func TestCalculateMM1(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mm1",
	}

	calc := NewQueueingCalculator(config)

	tests := []struct {
		name       string
		lambda     float64 // Arrival rate
		mu         float64 // Service rate
		servers    int     // Ignored for M/M/1
		want       *QueueingResult
		wantStable bool
	}{
		{
			name:       "stable system",
			lambda:     5.0,
			mu:         10.0,
			servers:    1,
			wantStable: true,
		},
		{
			name:       "unstable system",
			lambda:     15.0,
			mu:         10.0,
			servers:    1,
			wantStable: false,
		},
		{
			name:       "boundary case",
			lambda:     9.9,
			mu:         10.0,
			servers:    1,
			wantStable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Metrics{
				ServiceTime: time.Duration(1.0/tt.mu) * time.Second,
			}

			result := calc.Calculate(tt.lambda, tt.mu, tt.servers, metrics)
			if result == nil {
				t.Fatal("Calculate returned nil")
			}

			if result.Model != "M/M/1" {
				t.Errorf("Model = %v, want M/M/1", result.Model)
			}

			if tt.wantStable {
				if math.IsInf(result.QueueLength, 1) {
					t.Error("Expected stable system but got infinite queue length")
				}
				if result.Utilization >= 1.0 {
					t.Errorf("Utilization = %v, want < 1.0 for stable system", result.Utilization)
				}
				if result.Confidence <= 0 {
					t.Errorf("Confidence = %v, want > 0 for stable system", result.Confidence)
				}
			} else {
				if !math.IsInf(result.QueueLength, 1) {
					t.Error("Expected unstable system but got finite queue length")
				}
				if result.Utilization < 1.0 {
					t.Errorf("Utilization = %v, want >= 1.0 for unstable system", result.Utilization)
				}
			}

			// Check assumptions
			expectedAssumptions := []string{"Poisson arrivals", "Exponential service times", "Single server", "FIFO discipline"}
			if len(result.Assumptions) < len(expectedAssumptions) {
				t.Errorf("Missing assumptions, got %d, want at least %d", len(result.Assumptions), len(expectedAssumptions))
			}
		})
	}
}

func TestCalculateMMC(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mmc",
	}

	calc := NewQueueingCalculator(config)

	tests := []struct {
		name       string
		lambda     float64
		mu         float64
		servers    int
		wantStable bool
	}{
		{
			name:       "stable with multiple servers",
			lambda:     15.0,
			mu:         10.0,
			servers:    2,
			wantStable: true, // Total capacity = 2 * 10 = 20 > 15
		},
		{
			name:       "unstable even with multiple servers",
			lambda:     25.0,
			mu:         10.0,
			servers:    2,
			wantStable: false, // Total capacity = 2 * 10 = 20 < 25
		},
		{
			name:       "single server (equivalent to M/M/1)",
			lambda:     5.0,
			mu:         10.0,
			servers:    1,
			wantStable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Metrics{
				ServiceTime: time.Duration(1.0/tt.mu) * time.Second,
			}

			result := calc.Calculate(tt.lambda, tt.mu, tt.servers, metrics)
			if result == nil {
				t.Fatal("Calculate returned nil")
			}

			if result.Model != "M/M/c" {
				t.Errorf("Model = %v, want M/M/c", result.Model)
			}

			if result.Capacity != tt.servers {
				t.Errorf("Capacity = %v, want %v", result.Capacity, tt.servers)
			}

			expectedUtilization := tt.lambda / (float64(tt.servers) * tt.mu)
			if tt.wantStable {
				if math.IsInf(result.QueueLength, 1) {
					t.Error("Expected stable system but got infinite queue length")
				}
				if math.Abs(result.Utilization-expectedUtilization) > 0.001 {
					t.Errorf("Utilization = %v, want %v", result.Utilization, expectedUtilization)
				}
			} else {
				if !math.IsInf(result.QueueLength, 1) {
					t.Error("Expected unstable system but got finite queue length")
				}
			}
		})
	}
}

func TestCalculateMGC(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mgc",
	}

	calc := NewQueueingCalculator(config)

	lambda := 10.0
	mu := 5.0
	servers := 3

	// Test with different service time variability
	tests := []struct {
		name           string
		serviceTimeStd time.Duration
		wantHigherWait bool // Expect higher wait time than M/M/c
	}{
		{
			name:           "low variability",
			serviceTimeStd: 50 * time.Millisecond,
			wantHigherWait: false, // Similar to M/M/c
		},
		{
			name:           "high variability",
			serviceTimeStd: 500 * time.Millisecond,
			wantHigherWait: true, // Should be higher than M/M/c
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Metrics{
				ServiceTime:    time.Duration(1.0/mu) * time.Second,
				ServiceTimeStd: tt.serviceTimeStd,
			}

			result := calc.Calculate(lambda, mu, servers, metrics)
			if result == nil {
				t.Fatal("Calculate returned nil")
			}

			if result.Model != "M/G/c" {
				t.Errorf("Model = %v, want M/G/c", result.Model)
			}

			// Compare with M/M/c result
			mmcConfig := PlannerConfig{QueueingModel: "mmc"}
			mmcCalc := NewQueueingCalculator(mmcConfig)
			mmcResult := mmcCalc.Calculate(lambda, mu, servers, metrics)

			if tt.wantHigherWait {
				if result.WaitTime <= mmcResult.WaitTime {
					t.Errorf("M/G/c wait time %v should be higher than M/M/c %v for high variability",
						result.WaitTime, mmcResult.WaitTime)
				}
			}

			// Check that assumptions include general service time
			foundGeneralService := false
			for _, assumption := range result.Assumptions {
				if assumption == "General service time distribution" {
					foundGeneralService = true
					break
				}
			}
			if !foundGeneralService {
				t.Error("M/G/c assumptions should include general service time distribution")
			}
		})
	}
}

func TestCalculateCapacity(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mmc",
	}

	calc := NewQueueingCalculator(config)

	tests := []struct {
		name           string
		lambda         float64
		mu             float64
		targetLatency  time.Duration
		wantMinServers int
		wantMaxServers int
	}{
		{
			name:           "low load",
			lambda:         5.0,
			mu:             10.0,
			targetLatency:  1 * time.Second,
			wantMinServers: 1,
			wantMaxServers: 2,
		},
		{
			name:           "high load",
			lambda:         50.0,
			mu:             10.0,
			targetLatency:  1 * time.Second,
			wantMinServers: 6, // Need at least 6 servers for stability (50/10 = 5, plus safety)
			wantMaxServers: 20,
		},
		{
			name:           "strict latency",
			lambda:         20.0,
			mu:             10.0,
			targetLatency:  100 * time.Millisecond,
			wantMinServers: 3,
			wantMaxServers: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := calc.CalculateCapacity(tt.lambda, tt.mu, tt.targetLatency)

			if capacity < tt.wantMinServers {
				t.Errorf("CalculateCapacity() = %v, want >= %v", capacity, tt.wantMinServers)
			}

			if capacity > tt.wantMaxServers {
				t.Errorf("CalculateCapacity() = %v, want <= %v", capacity, tt.wantMaxServers)
			}

			// Verify that the calculated capacity actually meets the target
			result := calc.Calculate(tt.lambda, tt.mu, capacity, Metrics{})
			if !math.IsInf(result.ResponseTime.Seconds(), 1) && result.ResponseTime > tt.targetLatency {
				t.Errorf("Calculated capacity %d doesn't meet target latency: got %v, want <= %v",
					capacity, result.ResponseTime, tt.targetLatency)
			}
		})
	}
}

func TestEstimateServiceRate(t *testing.T) {
	config := PlannerConfig{}
	calc := NewQueueingCalculator(config)

	tests := []struct {
		name        string
		serviceTime time.Duration
		wantRate    float64
	}{
		{
			name:        "1 second service time",
			serviceTime: 1 * time.Second,
			wantRate:    1.0,
		},
		{
			name:        "2 second service time",
			serviceTime: 2 * time.Second,
			wantRate:    0.5,
		},
		{
			name:        "100ms service time",
			serviceTime: 100 * time.Millisecond,
			wantRate:    10.0,
		},
		{
			name:        "zero service time",
			serviceTime: 0,
			wantRate:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Metrics{
				ServiceTime: tt.serviceTime,
			}

			rate := calc.EstimateServiceRate(metrics)
			if math.Abs(rate-tt.wantRate) > 0.001 {
				t.Errorf("EstimateServiceRate() = %v, want %v", rate, tt.wantRate)
			}
		})
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want float64
	}{
		{"0!", 0, 1},
		{"1!", 1, 1},
		{"2!", 2, 2},
		{"3!", 3, 6},
		{"4!", 4, 24},
		{"5!", 5, 120},
		{"negative", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := factorial(tt.n)
			if got != tt.want {
				t.Errorf("factorial(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}

	// Test caching behavior
	factorial(10) // This should cache factorials 0-10
	factorial(8)  // This should use cached value

	// Verify cache contains expected values
	expectedValues := map[int]float64{
		6: 720,
		7: 5040,
		8: 40320,
	}

	for n, expected := range expectedValues {
		if factorialCache[n] != expected {
			t.Errorf("factorial cache[%d] = %v, want %v", n, factorialCache[n], expected)
		}
	}
}

func TestErlangC(t *testing.T) {
	tests := []struct {
		name    string
		lambda  float64
		mu      float64
		servers int
		want    float64
		wantMax float64 // Upper bound for reasonable results
	}{
		{
			name:    "low utilization",
			lambda:  5.0,
			mu:      10.0,
			servers: 2,
			want:    0.0, // Very low probability of waiting
			wantMax: 0.5,
		},
		{
			name:    "moderate utilization",
			lambda:  15.0,
			mu:      10.0,
			servers: 2,
			want:    0.5, // Moderate probability
			wantMax: 0.9,
		},
		{
			name:    "overloaded system",
			lambda:  25.0,
			mu:      10.0,
			servers: 2,
			want:    1.0, // System overloaded
			wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := erlangC(tt.lambda, tt.mu, tt.servers)

			if got < 0 || got > 1 {
				t.Errorf("erlangC() = %v, want probability between 0 and 1", got)
			}

			if got > tt.wantMax {
				t.Errorf("erlangC() = %v, want <= %v", got, tt.wantMax)
			}

			// For overloaded systems, should return 1.0
			if tt.lambda >= float64(tt.servers)*tt.mu {
				if got != 1.0 {
					t.Errorf("erlangC() = %v, want 1.0 for overloaded system", got)
				}
			}
		})
	}
}

func TestLittlesLaw(t *testing.T) {
	tests := []struct {
		name     string
		lambda   float64
		waitTime time.Duration
		want     float64
	}{
		{
			name:     "basic case",
			lambda:   10.0,
			waitTime: 2 * time.Second,
			want:     20.0,
		},
		{
			name:     "zero arrival rate",
			lambda:   0.0,
			waitTime: 5 * time.Second,
			want:     0.0,
		},
		{
			name:     "zero wait time",
			lambda:   10.0,
			waitTime: 0,
			want:     0.0,
		},
		{
			name:     "millisecond precision",
			lambda:   5.0,
			waitTime: 100 * time.Millisecond,
			want:     0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := littlesLaw(tt.lambda, tt.waitTime)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("littlesLaw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelConfidence(t *testing.T) {
	config := PlannerConfig{
		QueueingModel: "mmc",
	}

	calc := NewQueueingCalculator(config).(*queueingCalculator)

	tests := []struct {
		name        string
		model       string
		lambda      float64
		mu          float64
		servers     int
		metrics     Metrics
		wantMinConf float64
		wantMaxConf float64
	}{
		{
			name:        "low utilization M/M/c",
			model:       "M/M/c",
			lambda:      10.0,
			mu:          10.0,
			servers:     2,
			metrics:     Metrics{ServiceTimeStd: 100 * time.Millisecond},
			wantMinConf: 0.7,
			wantMaxConf: 1.0,
		},
		{
			name:        "high utilization M/M/c",
			model:       "M/M/c",
			lambda:      18.0,
			mu:          10.0,
			servers:     2,
			metrics:     Metrics{ServiceTimeStd: 100 * time.Millisecond},
			wantMinConf: 0.5,
			wantMaxConf: 0.8,
		},
		{
			name:        "M/M/1 with multiple workers",
			model:       "M/M/1",
			lambda:      10.0,
			mu:          10.0,
			servers:     3, // Mismatch should reduce confidence
			metrics:     Metrics{ServiceTimeStd: 100 * time.Millisecond},
			wantMinConf: 0.4,
			wantMaxConf: 0.7,
		},
		{
			name:    "M/G/c with high variability",
			model:   "M/G/c",
			lambda:  10.0,
			mu:      10.0,
			servers: 2,
			metrics: Metrics{
				ServiceTime:    100 * time.Millisecond,
				ServiceTimeStd: 200 * time.Millisecond, // High CV
			},
			wantMinConf: 0.8,
			wantMaxConf: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := calc.calculateModelConfidence(tt.model, tt.lambda, tt.mu, tt.servers, tt.metrics)

			if conf < 0.1 || conf > 1.0 {
				t.Errorf("confidence = %v, want between 0.1 and 1.0", conf)
			}

			if conf < tt.wantMinConf || conf > tt.wantMaxConf {
				t.Errorf("confidence = %v, want between %v and %v", conf, tt.wantMinConf, tt.wantMaxConf)
			}
		})
	}
}

func TestQueueingDefaultModel(t *testing.T) {
	// Test default model selection when queueing model is not specified or invalid
	config := PlannerConfig{
		QueueingModel: "invalid_model",
	}

	calc := NewQueueingCalculator(config)
	lambda := 10.0
	mu := 5.0
	servers := 3
	metrics := Metrics{ServiceTime: 200 * time.Millisecond}

	result := calc.Calculate(lambda, mu, servers, metrics)

	// Should default to M/M/c
	if result.Model != "M/M/c" {
		t.Errorf("Default model = %v, want M/M/c", result.Model)
	}
}

func TestUtilizationEdgeCases(t *testing.T) {
	config := PlannerConfig{QueueingModel: "mmc"}
	calc := NewQueueingCalculator(config)

	// Test exactly at stability boundary
	lambda := 20.0
	mu := 10.0
	servers := 2 // Total capacity = 2 * 10 = 20, exactly equal to lambda

	metrics := Metrics{ServiceTime: 100 * time.Millisecond}
	result := calc.Calculate(lambda, mu, servers, metrics)

	// System should be considered unstable at boundary
	if !math.IsInf(result.QueueLength, 1) {
		t.Error("System at stability boundary should be considered unstable")
	}

	// Test just below boundary
	lambda = 19.9
	result = calc.Calculate(lambda, mu, servers, metrics)

	if math.IsInf(result.QueueLength, 1) {
		t.Error("System just below stability boundary should be stable")
	}
}
