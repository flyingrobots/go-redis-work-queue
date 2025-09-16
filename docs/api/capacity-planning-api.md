# Automatic Capacity Planning API

## Overview

The Automatic Capacity Planning API provides intelligent worker scaling recommendations based on queueing theory, time-series forecasting, and SLO targets. It enables both manual planning workflows and automated scaling through predictive models.

## Table of Contents

- [Core Interfaces](#core-interfaces)
- [Data Types](#data-types)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)
- [Configuration](#configuration)
- [Performance Considerations](#performance-considerations)

## Core Interfaces

### CapacityPlanner

The main interface for generating scaling recommendations.

```go
type CapacityPlanner interface {
    GeneratePlan(ctx context.Context, req PlanRequest) (*PlanResponse, error)
    UpdateConfig(config PlannerConfig) error
    GetState() PlannerState
    Reset() error
}
```

#### GeneratePlan

Generates a capacity plan based on current metrics, SLO targets, and forecasting models.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `req`: Plan request containing metrics, SLO, and configuration

**Returns:**
- `*PlanResponse`: Complete capacity plan with forecasts and recommendations
- `error`: Any validation or computation errors

**Example:**
```go
planner := capacityplanning.NewCapacityPlanner(config)

request := capacityplanning.PlanRequest{
    QueueName: "high-priority-jobs",
    CurrentMetrics: capacityplanning.Metrics{
        Timestamp:      time.Now(),
        ArrivalRate:    25.5,  // jobs/second
        ServiceTime:    2 * time.Second,
        CurrentWorkers: 10,
        Utilization:    0.75,
        Backlog:        150,
    },
    SLO: capacityplanning.SLO{
        P95Latency: 5 * time.Second,
        MaxBacklog: 500,
        ErrorBudget: 0.01,
    },
    Config: config,
}

response, err := planner.GeneratePlan(ctx, request)
if err != nil {
    return fmt.Errorf("planning failed: %w", err)
}

fmt.Printf("Recommended workers: %d → %d\n",
    response.Plan.CurrentWorkers,
    response.Plan.TargetWorkers)
```

### Forecaster

Interface for time-series forecasting of arrival rates.

```go
type Forecaster interface {
    Predict(ctx context.Context, history []Metrics, horizon time.Duration) ([]Forecast, error)
    SetModel(model string) error
    GetAccuracy() float64
}
```

#### Predict

Generates arrival rate forecasts using various time-series models.

**Supported Models:**
- `ewma`: Exponential Weighted Moving Average
- `holt_winters`: Triple exponential smoothing with trend and seasonality
- `linear`: Linear regression with trend analysis
- `seasonal`: Seasonal decomposition with pattern recognition

**Example:**
```go
forecaster := capacityplanning.NewForecaster(config)

// Historical metrics from the last 24 hours
history := []capacityplanning.Metrics{
    // ... metrics collected over time
}

forecasts, err := forecaster.Predict(ctx, history, 2*time.Hour)
if err != nil {
    return fmt.Errorf("forecasting failed: %w", err)
}

for _, forecast := range forecasts {
    fmt.Printf("At %v: %.1f jobs/sec (±%.1f confidence: %.0f%%)\n",
        forecast.Timestamp,
        forecast.ArrivalRate,
        forecast.Upper - forecast.Lower,
        forecast.Confidence * 100)
}
```

### QueueingCalculator

Interface for queueing theory calculations and capacity estimation.

```go
type QueueingCalculator interface {
    Calculate(lambda, mu float64, servers int, metrics Metrics) *QueueingResult
    CalculateCapacity(lambda, mu float64, targetLatency time.Duration) int
    EstimateServiceRate(metrics Metrics) float64
}
```

#### Calculate

Performs queueing theory analysis using M/M/c, M/M/1, or M/G/c models.

**Parameters:**
- `lambda`: Arrival rate (jobs/second)
- `mu`: Service rate per server (jobs/second)
- `servers`: Number of worker servers
- `metrics`: Current system metrics for confidence estimation

**Example:**
```go
calc := capacityplanning.NewQueueingCalculator(config)

result := calc.Calculate(
    20.0,  // λ = 20 jobs/sec arrival rate
    5.0,   // μ = 5 jobs/sec per worker service rate
    4,     // c = 4 workers
    metrics,
)

fmt.Printf("Utilization: %.1f%%\n", result.Utilization * 100)
fmt.Printf("Queue length: %.1f jobs\n", result.QueueLength)
fmt.Printf("Wait time: %v\n", result.WaitTime)
fmt.Printf("Response time: %v\n", result.ResponseTime)
fmt.Printf("Model confidence: %.1f%%\n", result.Confidence * 100)
```

### Simulator

Interface for what-if analysis and plan validation.

```go
type Simulator interface {
    Simulate(ctx context.Context, scenario SimulationScenario) (*Simulation, error)
    ValidateScenario(scenario SimulationScenario) error
    EstimateRuntime(scenario SimulationScenario) time.Duration
}
```

#### Simulate

Runs discrete event simulation to validate capacity plans under different scenarios.

**Example:**
```go
simulator := capacityplanning.NewSimulator(config)

scenario := capacityplanning.SimulationScenario{
    Name: "Black Friday Traffic Spike",
    Plan: capacityPlan,
    TrafficPattern: capacityplanning.TrafficPattern{
        Type: capacityplanning.PatternSpiky,
        BaseRate: 50.0,
        Spikes: []capacityplanning.TrafficSpike{
            {
                StartTime: time.Now().Add(30 * time.Minute),
                Duration:  2 * time.Hour,
                Magnitude: 5.0,  // 5x normal traffic
                Shape:     capacityplanning.SpikeBell,
            },
        },
    },
    Duration:    6 * time.Hour,
    Granularity: 5 * time.Minute,
}

simulation, err := simulator.Simulate(ctx, scenario)
if err != nil {
    return fmt.Errorf("simulation failed: %w", err)
}

fmt.Printf("SLO Achievement: %.1f%%\n", simulation.Summary.SLOAchievement * 100)
fmt.Printf("Max Backlog: %d jobs\n", simulation.Summary.MaxBacklog)
fmt.Printf("Total Cost: $%.2f\n", simulation.Summary.TotalCost)
```

## Data Types

### Core Metrics

```go
type Metrics struct {
    Timestamp      time.Time     `json:"timestamp"`
    ArrivalRate    float64       `json:"arrival_rate"`     // Jobs per second (λ)
    ServiceTime    time.Duration `json:"service_time"`     // Mean service time (1/μ)
    ServiceTimeP95 time.Duration `json:"service_time_p95"` // 95th percentile
    ServiceTimeStd time.Duration `json:"service_time_std"` // Standard deviation
    CurrentWorkers int           `json:"current_workers"`  // Current worker count (c)
    Utilization    float64       `json:"utilization"`      // Current utilization (ρ)
    Backlog        int           `json:"backlog"`          // Current queue length
    ActiveJobs     int           `json:"active_jobs"`      // Jobs being processed
    QueueName      string        `json:"queue_name"`       // Queue identifier
}
```

### Service Level Objectives

```go
type SLO struct {
    P95Latency   time.Duration `json:"p95_latency"`    // Target 95th percentile latency
    MaxBacklog   int           `json:"max_backlog"`    // Maximum allowed backlog
    ErrorBudget  float64       `json:"error_budget"`   // Acceptable error rate (0.0-1.0)
    DrainTime    time.Duration `json:"drain_time"`     // Time to drain after burst
    Availability float64       `json:"availability"`   // Target availability (0.99 = 99%)
}
```

### Capacity Plan

```go
type CapacityPlan struct {
    ID              string        `json:"id"`
    GeneratedAt     time.Time     `json:"generated_at"`
    CurrentWorkers  int           `json:"current_workers"`
    TargetWorkers   int           `json:"target_workers"`
    Delta           int           `json:"delta"`           // TargetWorkers - CurrentWorkers
    Steps           []ScalingStep `json:"steps"`           // Time-sequenced scaling actions
    Confidence      float64       `json:"confidence"`      // Plan confidence (0.0-1.0)
    CostImpact      CostAnalysis  `json:"cost_impact"`
    SLOAchievable   bool          `json:"slo_achievable"`
    Rationale       string        `json:"rationale"`
    ForecastWindow  time.Duration `json:"forecast_window"`
    SafetyMargin    float64       `json:"safety_margin"`
    ValidUntil      time.Time     `json:"valid_until"`
    QueueName       string        `json:"queue_name"`
}
```

### Scaling Step

```go
type ScalingStep struct {
    Sequence      int           `json:"sequence"`
    ScheduledAt   time.Time     `json:"scheduled_at"`
    Action        ScalingAction `json:"action"`          // scale_up, scale_down, no_change
    FromWorkers   int           `json:"from_workers"`
    ToWorkers     int           `json:"to_workers"`
    Delta         int           `json:"delta"`
    Rationale     string        `json:"rationale"`
    EstimatedCost float64       `json:"estimated_cost"`  // $/hour
    Confidence    float64       `json:"confidence"`
    CooldownUntil time.Time     `json:"cooldown_until"`
}
```

### Traffic Patterns

```go
type TrafficPattern struct {
    Type      PatternType     `json:"type"`      // constant, sinusoidal, linear, spiky, daily, weekly
    BaseRate  float64         `json:"base_rate"` // Baseline arrival rate
    Amplitude float64         `json:"amplitude"` // For sinusoidal patterns
    Period    time.Duration   `json:"period"`    // For periodic patterns
    Spikes    []TrafficSpike  `json:"spikes"`    // Discrete traffic events
    Trend     float64         `json:"trend"`     // Growth/decline rate
    Noise     float64         `json:"noise"`     // Random variation (0.0-1.0)
}

type TrafficSpike struct {
    StartTime time.Time     `json:"start_time"`
    Duration  time.Duration `json:"duration"`
    Magnitude float64       `json:"magnitude"` // Multiplier (2.0 = 2x normal rate)
    Shape     SpikeShape    `json:"shape"`     // instant, linear, exp, bell
}
```

## Usage Examples

### Basic Capacity Planning

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/yourorg/go-redis-work-queue/internal/automatic-capacity-planning"
)

func main() {
    config := capacityplanning.PlannerConfig{
        ForecastWindow:      60 * time.Minute,
        ForecastModel:       "ewma",
        SafetyMargin:        0.15,
        ConfidenceThreshold: 0.85,
        MaxStepSize:         10,
        CooldownPeriod:      5 * time.Minute,
        MinWorkers:          2,
        MaxWorkers:          50,
        QueueingModel:       "mmc",
        ScaleUpThreshold:    0.80,
        ScaleDownThreshold:  0.60,
    }

    planner := capacityplanning.NewCapacityPlanner(config)

    // Current system state
    metrics := capacityplanning.Metrics{
        Timestamp:      time.Now(),
        ArrivalRate:    15.3,
        ServiceTime:    3 * time.Second,
        CurrentWorkers: 8,
        Utilization:    0.72,
        Backlog:        120,
        QueueName:      "image-processing",
    }

    // Service level objectives
    slo := capacityplanning.SLO{
        P95Latency:   10 * time.Second,
        MaxBacklog:   500,
        ErrorBudget:  0.02,
        Availability: 0.995,
    }

    request := capacityplanning.PlanRequest{
        QueueName:      "image-processing",
        CurrentMetrics: metrics,
        SLO:           slo,
        Config:        config,
    }

    ctx := context.Background()
    response, err := planner.GeneratePlan(ctx, request)
    if err != nil {
        panic(err)
    }

    fmt.Printf("=== Capacity Plan ===\n")
    fmt.Printf("Current Workers: %d\n", response.Plan.CurrentWorkers)
    fmt.Printf("Target Workers:  %d\n", response.Plan.TargetWorkers)
    fmt.Printf("Delta: %+d\n", response.Plan.Delta)
    fmt.Printf("Confidence: %.1f%%\n", response.Plan.Confidence * 100)
    fmt.Printf("SLO Achievable: %v\n", response.Plan.SLOAchievable)

    fmt.Printf("\n=== Scaling Steps ===\n")
    for _, step := range response.Plan.Steps {
        fmt.Printf("%v: %s %d → %d workers (%s)\n",
            step.ScheduledAt.Format("15:04"),
            step.Action,
            step.FromWorkers,
            step.ToWorkers,
            step.Rationale)
    }

    fmt.Printf("\n=== Cost Analysis ===\n")
    fmt.Printf("Current Cost: $%.2f/hour\n", response.Plan.CostImpact.CurrentCostPerHour)
    fmt.Printf("Projected Cost: $%.2f/hour\n", response.Plan.CostImpact.ProjectedCostPerHour)
    fmt.Printf("Monthly Delta: $%.2f\n", response.Plan.CostImpact.MonthlyCostDelta)
}
```

### What-If Analysis

```go
func runWhatIfAnalysis() {
    simulator := capacityplanning.NewSimulator(config)

    // Test different scenarios
    scenarios := []capacityplanning.SimulationScenario{
        {
            Name: "Current Plan",
            Plan: currentPlan,
            TrafficPattern: capacityplanning.TrafficPattern{
                Type:     capacityplanning.PatternDaily,
                BaseRate: 20.0,
                Noise:    0.1,
            },
            Duration:    24 * time.Hour,
            Granularity: 15 * time.Minute,
        },
        {
            Name: "Conservative Plan (+20% workers)",
            Plan: conservativePlan,
            TrafficPattern: capacityplanning.TrafficPattern{
                Type:     capacityplanning.PatternDaily,
                BaseRate: 20.0,
                Noise:    0.1,
            },
            Duration:    24 * time.Hour,
            Granularity: 15 * time.Minute,
        },
    }

    for _, scenario := range scenarios {
        simulation, err := simulator.Simulate(ctx, scenario)
        if err != nil {
            fmt.Printf("Simulation failed: %v\n", err)
            continue
        }

        fmt.Printf("\n=== %s ===\n", scenario.Name)
        fmt.Printf("SLO Achievement: %.1f%%\n", simulation.Summary.SLOAchievement * 100)
        fmt.Printf("Avg Latency: %v\n", simulation.Summary.AvgLatency)
        fmt.Printf("P95 Latency: %v\n", simulation.Summary.P95Latency)
        fmt.Printf("Max Backlog: %d\n", simulation.Summary.MaxBacklog)
        fmt.Printf("Total Cost: $%.2f\n", simulation.Summary.TotalCost)
        fmt.Printf("Efficiency Score: %.1f\n", simulation.Summary.EfficiencyScore)

        if len(simulation.SLOAnalysis.ViolationPeriods) > 0 {
            fmt.Printf("Violations: %d periods\n", len(simulation.SLOAnalysis.ViolationPeriods))
            for _, violation := range simulation.SLOAnalysis.ViolationPeriods {
                fmt.Printf("  %v - %v (%s): %s\n",
                    violation.Start.Format("15:04"),
                    violation.End.Format("15:04"),
                    violation.Duration,
                    violation.Type)
            }
        }
    }
}
```

### Continuous Monitoring

```go
func continuousPlanning() {
    planner := capacityplanning.NewCapacityPlanner(config)
    ticker := time.NewTicker(5 * time.Minute)

    for {
        select {
        case <-ticker.C:
            metrics := collectCurrentMetrics()

            request := capacityplanning.PlanRequest{
                QueueName:      "production-queue",
                CurrentMetrics: metrics,
                SLO:           slo,
                Config:        config,
            }

            response, err := planner.GeneratePlan(ctx, request)
            if err != nil {
                log.Printf("Planning failed: %v", err)
                continue
            }

            // Auto-apply if confidence is high enough
            if response.Plan.Confidence >= config.ConfidenceThreshold {
                if shouldAutoApply(response.Plan) {
                    err := applyScalingPlan(response.Plan)
                    if err != nil {
                        log.Printf("Auto-scaling failed: %v", err)
                    } else {
                        log.Printf("Auto-scaled to %d workers", response.Plan.TargetWorkers)
                    }
                }
            } else {
                log.Printf("Plan confidence %.1f%% below threshold %.1f%%, manual review required",
                    response.Plan.Confidence * 100,
                    config.ConfidenceThreshold * 100)
            }

        case <-ctx.Done():
            return
        }
    }
}
```

## Error Handling

The API uses structured error types for clear error categorization:

```go
type PlannerError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"cause,omitempty"`
}
```

### Error Codes

| Code | Description | Action |
|------|-------------|---------|
| `INVALID_METRICS` | Invalid input metrics | Validate metrics before retry |
| `INSUFFICIENT_HISTORY` | Not enough historical data | Collect more data or use simpler model |
| `FORECAST_FAILED` | Forecasting engine error | Check model configuration |
| `MODEL_NOT_SUPPORTED` | Unsupported queueing/forecast model | Use supported model name |
| `CONFIG_INVALID` | Invalid configuration | Fix configuration parameters |
| `SLO_UNACHIEVABLE` | SLO cannot be met | Relax SLO or increase capacity limits |
| `CAPACITY_LIMIT_EXCEEDED` | Exceeds max worker limit | Increase max workers or optimize efficiency |
| `COOLDOWN_ACTIVE` | Still in cooldown period | Wait for cooldown to expire |
| `ANOMALY_DETECTED` | Traffic anomaly detected | Manual review recommended |

### Error Handling Example

```go
response, err := planner.GeneratePlan(ctx, request)
if err != nil {
    if plannerErr, ok := err.(*capacityplanning.PlannerError); ok {
        switch plannerErr.Code {
        case capacityplanning.ErrInsufficientHistory:
            // Fall back to simpler model
            config.ForecastModel = "ewma"
            return retry(config)

        case capacityplanning.ErrSLOUnachievable:
            // Suggest SLO adjustment
            return suggestSLORelaxation(request.SLO)

        case capacityplanning.ErrCooldownActive:
            // Wait and retry
            time.Sleep(config.CooldownPeriod)
            return retry(config)

        default:
            return fmt.Errorf("planning failed: %w", err)
        }
    }
    return fmt.Errorf("unexpected error: %w", err)
}
```

## Configuration

### PlannerConfig

```go
type PlannerConfig struct {
    // Forecasting
    ForecastWindow    time.Duration `json:"forecast_window"`    // How far ahead to predict
    ForecastModel     string        `json:"forecast_model"`     // "ewma", "holt_winters", "linear", "seasonal"
    HistoryWindow     time.Duration `json:"history_window"`     // Historical data window
    SeasonalPeriod    time.Duration `json:"seasonal_period"`    // Daily, weekly patterns

    // Safety
    SafetyMargin        float64       `json:"safety_margin"`        // Additional capacity (0.15 = 15%)
    ConfidenceThreshold float64       `json:"confidence_threshold"` // Min confidence for auto-apply
    MaxStepSize         int           `json:"max_step_size"`        // Max workers per scaling step
    CooldownPeriod      time.Duration `json:"cooldown_period"`      // Min time between actions

    // Limits
    MinWorkers          int           `json:"min_workers"`          // Absolute minimum
    MaxWorkers          int           `json:"max_workers"`          // Absolute maximum

    // Cost
    WorkerCostPerHour     float64     `json:"worker_cost_per_hour"`     // $/worker/hour
    ViolationCostPerHour  float64     `json:"violation_cost_per_hour"`  // $/hour during SLO violation

    // Thresholds
    ScaleUpThreshold      float64     `json:"scale_up_threshold"`       // Utilization to trigger scale up
    ScaleDownThreshold    float64     `json:"scale_down_threshold"`     // Utilization to trigger scale down

    // Anomaly Detection
    AnomalyThreshold      float64     `json:"anomaly_threshold"`        // Z-score for anomaly detection
    SpikeThreshold        float64     `json:"spike_threshold"`          // Multiplier for spike detection

    // Model Parameters
    QueueingModel         string      `json:"queueing_model"`           // "mm1", "mmc", "mgc"
    ServiceTimeModel      string      `json:"service_time_model"`       // "exponential", "general"
}
```

### Default Configuration

```go
func DefaultConfig() PlannerConfig {
    return PlannerConfig{
        ForecastWindow:        60 * time.Minute,
        ForecastModel:         "ewma",
        HistoryWindow:         24 * time.Hour,
        SafetyMargin:          0.15,
        ConfidenceThreshold:   0.85,
        MaxStepSize:           15,
        CooldownPeriod:        5 * time.Minute,
        MinWorkers:            1,
        MaxWorkers:            100,
        WorkerCostPerHour:     0.50,
        ViolationCostPerHour:  100.0,
        ScaleUpThreshold:      0.80,
        ScaleDownThreshold:    0.60,
        AnomalyThreshold:      3.0,
        SpikeThreshold:        2.0,
        QueueingModel:         "mmc",
        ServiceTimeModel:      "exponential",
    }
}
```

## Performance Considerations

### Computational Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| M/M/c Calculation | O(c) | Linear in server count |
| EWMA Forecasting | O(n) | Linear in history size |
| Holt-Winters | O(n) | Linear in history size |
| Simulation | O(s) | Linear in simulation steps |
| Pattern Extraction | O(n) | Linear in data points |

### Memory Usage

- **History Storage**: O(w×p) where w = history window, p = number of pools
- **Forecast Cache**: O(f×p) where f = forecast horizon, p = number of pools
- **Simulation State**: O(s) where s = simulation timeline length

### Optimization Tips

1. **Limit History Window**: Use 24-48 hours of data for most use cases
2. **Cache Forecast Results**: Reuse forecasts for multiple plans
3. **Batch Multiple Queues**: Process related queues together
4. **Async Processing**: Use goroutines for independent calculations
5. **Circuit Breakers**: Timeout long-running forecasts

### Performance Monitoring

```go
// Track key performance metrics
capacity_planner_generation_duration_seconds{queue="name"}
capacity_planner_forecast_accuracy{model="ewma"}
capacity_planner_cache_hit_ratio{}
capacity_planner_simulation_duration_seconds{}
```

## Integration Patterns

### HTTP API Integration

```go
func (h *CapacityHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
    var request capacityplanning.PlanRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    response, err := h.planner.GeneratePlan(r.Context(), request)
    if err != nil {
        if plannerErr, ok := err.(*capacityplanning.PlannerError); ok {
            switch plannerErr.Code {
            case capacityplanning.ErrInvalidMetrics:
                http.Error(w, plannerErr.Message, http.StatusBadRequest)
            case capacityplanning.ErrSLOUnachievable:
                http.Error(w, plannerErr.Message, http.StatusConflict)
            default:
                http.Error(w, "Internal error", http.StatusInternalServerError)
            }
        } else {
            http.Error(w, "Internal error", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### gRPC Integration

```protobuf
service CapacityPlannerService {
    rpc GeneratePlan(PlanRequest) returns (PlanResponse);
    rpc Simulate(SimulationRequest) returns (SimulationResponse);
    rpc GetForecasts(ForecastRequest) returns (ForecastResponse);
}
```

### Kubernetes Operator Integration

```go
func (r *WorkerPoolReconciler) reconcileCapacity(ctx context.Context, pool *v1.WorkerPool) error {
    metrics := r.collectMetrics(pool)

    request := capacityplanning.PlanRequest{
        QueueName:      pool.Spec.QueueName,
        CurrentMetrics: metrics,
        SLO:           pool.Spec.SLO,
        Config:        pool.Spec.CapacityConfig,
    }

    response, err := r.planner.GeneratePlan(ctx, request)
    if err != nil {
        return err
    }

    if response.Plan.Confidence >= pool.Spec.AutoScaleThreshold {
        return r.applyScalingPlan(ctx, pool, response.Plan)
    }

    return r.recordRecommendation(ctx, pool, response.Plan)
}
```

## Best Practices

1. **Start Conservative**: Begin with higher safety margins and lower confidence thresholds
2. **Monitor Accuracy**: Track forecast accuracy and adjust models accordingly
3. **Use Simulation**: Always validate plans with what-if analysis before auto-applying
4. **Gradual Rollout**: Scale incrementally with proper cooldown periods
5. **Cost Awareness**: Balance performance targets with cost constraints
6. **Anomaly Handling**: Pause auto-scaling during detected anomalies
7. **SLO Alignment**: Ensure SLOs reflect actual business requirements
8. **Historical Analysis**: Regularly review scaling decisions and outcomes