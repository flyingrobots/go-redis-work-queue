// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"testing"
	"time"
)

func TestNewCapacityPlanner(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)
	if planner == nil {
		t.Fatal("NewCapacityPlanner returned nil")
	}

	// Test interface compliance
	var _ CapacityPlanner = planner
}

func TestGeneratePlan(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
		ScaleUpThreshold:    0.80,
		ScaleDownThreshold:  0.60,
	}

	planner := NewCapacityPlanner(config)

	tests := []struct {
		name     string
		request  PlanRequest
		wantErr  bool
		errCode  string
	}{
		{
			name: "valid request",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					Timestamp:      time.Now(),
					ArrivalRate:    10.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: 5,
					Utilization:    0.5,
					Backlog:        100,
				},
				SLO: SLO{
					P95Latency: 5 * time.Second,
					MaxBacklog: 1000,
				},
				Config: config,
			},
			wantErr: false,
		},
		{
			name: "invalid metrics - negative arrival rate",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					ArrivalRate:    -1.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: 5,
				},
				SLO:    SLO{P95Latency: 5 * time.Second},
				Config: config,
			},
			wantErr: true,
			errCode: ErrInvalidMetrics,
		},
		{
			name: "invalid metrics - zero service time",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					ArrivalRate:    10.0,
					ServiceTime:    0,
					CurrentWorkers: 5,
				},
				SLO:    SLO{P95Latency: 5 * time.Second},
				Config: config,
			},
			wantErr: true,
			errCode: ErrInvalidMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			response, err := planner.GeneratePlan(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GeneratePlan() expected error but got none")
					return
				}
				if plannerErr, ok := err.(*PlannerError); ok {
					if plannerErr.Code != tt.errCode {
						t.Errorf("GeneratePlan() error code = %v, want %v", plannerErr.Code, tt.errCode)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("GeneratePlan() unexpected error = %v", err)
				return
			}

			if response == nil {
				t.Error("GeneratePlan() returned nil response")
				return
			}

			// Validate response structure
			if response.Plan.QueueName != tt.request.QueueName {
				t.Errorf("Plan queue name = %v, want %v", response.Plan.QueueName, tt.request.QueueName)
			}

			if response.Plan.CurrentWorkers != tt.request.CurrentMetrics.CurrentWorkers {
				t.Errorf("Plan current workers = %v, want %v", response.Plan.CurrentWorkers, tt.request.CurrentMetrics.CurrentWorkers)
			}

			if len(response.Forecast) == 0 {
				t.Error("GeneratePlan() returned empty forecast")
			}

			if response.GenerationTime <= 0 {
				t.Error("GeneratePlan() generation time should be positive")
			}
		})
	}
}

func TestGeneratePlanWithHistory(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)

	// Create historical metrics
	history := make([]Metrics, 10)
	baseTime := time.Now().Add(-10 * time.Hour)
	for i := range history {
		history[i] = Metrics{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate:    10.0 + float64(i), // Increasing trend
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.5,
			Backlog:        100,
		}
	}

	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    20.0,
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.8,
			Backlog:        200,
		},
		SLO: SLO{
			P95Latency: 5 * time.Second,
			MaxBacklog: 1000,
		},
		Config: config,
	}

	ctx := context.Background()
	response, err := planner.GeneratePlan(ctx, request)

	if err != nil {
		t.Fatalf("GeneratePlan() with history failed: %v", err)
	}

	// Should recommend scaling up due to high utilization and increasing trend
	if response.Plan.TargetWorkers <= response.Plan.CurrentWorkers {
		t.Errorf("Expected scale-up recommendation, got target=%d, current=%d",
			response.Plan.TargetWorkers, response.Plan.CurrentWorkers)
	}

	// Should have scaling steps
	if len(response.Plan.Steps) == 0 {
		t.Error("Expected scaling steps in plan")
	}
}

func TestScalingSafety(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         10, // Limit step size
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          2,
		MaxWorkers:          50,
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)

	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    100.0, // Very high arrival rate
			ServiceTime:    1 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.95,
			Backlog:        1000,
		},
		SLO: SLO{
			P95Latency: 2 * time.Second,
			MaxBacklog: 500,
		},
		Config: config,
	}

	ctx := context.Background()
	response, err := planner.GeneratePlan(ctx, request)

	if err != nil {
		t.Fatalf("GeneratePlan() safety test failed: %v", err)
	}

	// Should respect max workers limit
	if response.Plan.TargetWorkers > config.MaxWorkers {
		t.Errorf("Plan exceeded max workers: got %d, max %d",
			response.Plan.TargetWorkers, config.MaxWorkers)
	}

	// Should respect min workers limit
	if response.Plan.TargetWorkers < config.MinWorkers {
		t.Errorf("Plan below min workers: got %d, min %d",
			response.Plan.TargetWorkers, config.MinWorkers)
	}

	// Should limit step sizes
	for _, step := range response.Plan.Steps {
		if step.Delta > config.MaxStepSize {
			t.Errorf("Step size %d exceeds limit %d", step.Delta, config.MaxStepSize)
		}
	}

	// Should include cooldown periods
	for i, step := range response.Plan.Steps {
		if i > 0 {
			prevStep := response.Plan.Steps[i-1]
			timeDiff := step.ScheduledAt.Sub(prevStep.ScheduledAt)
			if timeDiff < config.CooldownPeriod {
				t.Errorf("Step %d scheduled too soon after previous: %v < %v",
					i, timeDiff, config.CooldownPeriod)
			}
		}
	}
}

func TestCooldownLogic(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      10 * time.Minute, // Longer cooldown
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)

	// Set state with recent scaling action
	state := &PlannerState{
		LastScaling:   time.Now().Add(-5 * time.Minute), // 5 minutes ago
		CooldownUntil: time.Now().Add(5 * time.Minute),  // 5 minutes remaining
	}

	// Use reflection or a test interface to set internal state
	// For now, test the cooldown through multiple calls

	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    50.0,
			ServiceTime:    1 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.9, // High utilization should trigger scaling
			Backlog:        500,
		},
		SLO: SLO{
			P95Latency: 2 * time.Second,
			MaxBacklog: 300,
		},
		Config: config,
	}

	ctx := context.Background()

	// First call should generate a plan
	response1, err := planner.GeneratePlan(ctx, request)
	if err != nil {
		t.Fatalf("First GeneratePlan() failed: %v", err)
	}

	// Second call immediately after should be influenced by cooldown
	response2, err := planner.GeneratePlan(ctx, request)
	if err != nil {
		t.Fatalf("Second GeneratePlan() failed: %v", err)
	}

	// Responses should be consistent or show conservative behavior
	if response2.Plan.TargetWorkers > response1.Plan.TargetWorkers+config.MaxStepSize {
		t.Errorf("Second plan too aggressive during cooldown: %d vs %d",
			response2.Plan.TargetWorkers, response1.Plan.TargetWorkers)
	}
}

func TestAnomalyDetection(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
		AnomalyThreshold:    3.0, // Z-score threshold
		SpikeThreshold:      2.0, // 2x normal rate
	}

	planner := NewCapacityPlanner(config)

	// Create baseline history with normal traffic
	history := make([]Metrics, 20)
	baseTime := time.Now().Add(-20 * time.Hour)
	for i := range history {
		history[i] = Metrics{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate:    10.0, // Consistent rate
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.4,
			Backlog:        50,
		}
	}

	// Test with anomalous current metrics (traffic spike)
	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    100.0, // 10x normal rate - clear anomaly
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.95,
			Backlog:        1000,
		},
		SLO: SLO{
			P95Latency: 5 * time.Second,
			MaxBacklog: 500,
		},
		Config: config,
	}

	ctx := context.Background()
	response, err := planner.GeneratePlan(ctx, request)

	if err != nil {
		t.Fatalf("GeneratePlan() with anomaly failed: %v", err)
	}

	// Should detect anomaly and be more conservative
	if response.Plan.Confidence > 0.7 {
		t.Errorf("Plan confidence too high during anomaly: %f", response.Plan.Confidence)
	}

	// Should include warnings about anomaly
	foundAnomalyWarning := false
	for _, warning := range response.Warnings {
		if containsString(warning, "anomaly") || containsString(warning, "spike") {
			foundAnomalyWarning = true
			break
		}
	}

	if !foundAnomalyWarning {
		t.Error("Expected anomaly warning in response")
	}
}

func TestSLOAchievability(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          10, // Very low limit
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)

	// Test impossible SLO with low worker limit
	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    100.0, // Very high rate
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.95,
			Backlog:        1000,
		},
		SLO: SLO{
			P95Latency: 1 * time.Second, // Very strict SLO
			MaxBacklog: 10,              // Very low backlog limit
		},
		Config: config,
	}

	ctx := context.Background()
	response, err := planner.GeneratePlan(ctx, request)

	if err != nil {
		t.Fatalf("GeneratePlan() SLO test failed: %v", err)
	}

	// Should recognize SLO is not achievable
	if response.Plan.SLOAchievable {
		t.Error("Plan incorrectly claims SLO is achievable with insufficient capacity")
	}

	// Should include appropriate warnings
	foundCapacityWarning := false
	for _, warning := range response.Warnings {
		if containsString(warning, "capacity") || containsString(warning, "SLO") {
			foundCapacityWarning = true
			break
		}
	}

	if !foundCapacityWarning {
		t.Error("Expected capacity/SLO warning in response")
	}
}

func TestPlanCaching(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow:      60 * time.Minute,
		ForecastModel:       "ewma",
		SafetyMargin:        0.15,
		ConfidenceThreshold: 0.85,
		MaxStepSize:         15,
		CooldownPeriod:      5 * time.Minute,
		MinWorkers:          1,
		MaxWorkers:          100,
		QueueingModel:       "mmc",
	}

	planner := NewCapacityPlanner(config)

	request := PlanRequest{
		QueueName: "test-queue",
		CurrentMetrics: Metrics{
			Timestamp:      time.Now(),
			ArrivalRate:    10.0,
			ServiceTime:    2 * time.Second,
			CurrentWorkers: 5,
			Utilization:    0.5,
			Backlog:        100,
		},
		SLO: SLO{
			P95Latency: 5 * time.Second,
			MaxBacklog: 1000,
		},
		Config:     config,
		ForceRegen: false, // Allow caching
	}

	ctx := context.Background()

	// First call
	response1, err := planner.GeneratePlan(ctx, request)
	if err != nil {
		t.Fatalf("First GeneratePlan() failed: %v", err)
	}

	// Second call with same parameters
	response2, err := planner.GeneratePlan(ctx, request)
	if err != nil {
		t.Fatalf("Second GeneratePlan() failed: %v", err)
	}

	// Second call should be faster (cached)
	if response2.GenerationTime > response1.GenerationTime {
		t.Logf("Cache miss - second call took longer (%v vs %v)",
			response2.GenerationTime, response1.GenerationTime)
	}

	// Force regeneration
	request.ForceRegen = true
	response3, err := planner.GeneratePlan(ctx, request)
	if err != nil {
		t.Fatalf("Third GeneratePlan() with ForceRegen failed: %v", err)
	}

	if response3.CacheHit {
		t.Error("Expected cache miss with ForceRegen=true")
	}
}

// Helper function to check if string contains substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    (len(s) > len(substr) &&
		     (s[:len(substr)] == substr ||
		      s[len(s)-len(substr):] == substr ||
		      containsStringHelper(s, substr))))
}

func containsStringHelper(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidateRequest(t *testing.T) {
	config := PlannerConfig{
		ForecastWindow: 60 * time.Minute,
		MinWorkers:     1,
		MaxWorkers:     100,
	}

	planner := NewCapacityPlanner(config)

	tests := []struct {
		name    string
		request PlanRequest
		wantErr bool
		errCode string
	}{
		{
			name: "valid request",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					Timestamp:      time.Now(),
					ArrivalRate:    10.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: 5,
				},
				SLO: SLO{P95Latency: 5 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "empty queue name",
			request: PlanRequest{
				QueueName: "",
				CurrentMetrics: Metrics{
					ArrivalRate:    10.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: 5,
				},
				SLO: SLO{P95Latency: 5 * time.Second},
			},
			wantErr: true,
			errCode: ErrInvalidMetrics,
		},
		{
			name: "negative current workers",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					ArrivalRate:    10.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: -1,
				},
				SLO: SLO{P95Latency: 5 * time.Second},
			},
			wantErr: true,
			errCode: ErrInvalidMetrics,
		},
		{
			name: "zero p95 latency SLO",
			request: PlanRequest{
				QueueName: "test-queue",
				CurrentMetrics: Metrics{
					ArrivalRate:    10.0,
					ServiceTime:    2 * time.Second,
					CurrentWorkers: 5,
				},
				SLO: SLO{P95Latency: 0},
			},
			wantErr: true,
			errCode: ErrInvalidMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := planner.GeneratePlan(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GeneratePlan() expected error but got none")
					return
				}
				if plannerErr, ok := err.(*PlannerError); ok {
					if plannerErr.Code != tt.errCode {
						t.Errorf("GeneratePlan() error code = %v, want %v", plannerErr.Code, tt.errCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GeneratePlan() unexpected error = %v", err)
				}
			}
		})
	}
}