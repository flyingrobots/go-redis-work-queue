# Forecasting API Documentation

## Overview

The Forecasting system provides time-series analysis and predictive capabilities for queue metrics, enabling proactive capacity planning and operational intelligence. It combines statistical models (EWMA, Holt-Winters) with actionable recommendations for scaling, maintenance, and SLO management.

## Core Components

### ForecastingEngine

The main orchestrator that manages models, storage, and recommendations.

```go
engine := forecasting.NewForecastingEngine(config, logger)
```

#### Configuration

```go
config := &forecasting.ForecastConfig{
    EWMAConfig: &forecasting.EWMAConfig{
        Alpha:              0.3,    // Smoothing parameter
        AutoAdjust:         true,   // Auto-tune based on accuracy
        MinObservations:    5,      // Minimum data points
        ConfidenceInterval: 0.95,   // 95% confidence bounds
    },
    HoltWintersConfig: &forecasting.HoltWintersConfig{
        Alpha:            0.3,      // Level smoothing
        Beta:             0.1,      // Trend smoothing  
        Gamma:            0.1,      // Seasonal smoothing
        SeasonLength:     24,       // Hours in a day
        SeasonalMethod:   "additive",
        AutoDetectSeason: true,     // Auto-detect patterns
    },
    StorageConfig: &forecasting.StorageConfig{
        RetentionDuration: 7 * 24 * time.Hour,
        SamplingInterval:  1 * time.Minute,
        MaxDataPoints:     10080,   // 7 days of minute data
        PersistToDisk:     true,
        StoragePath:       "/var/lib/forecasting",
    },
    EngineConfig: &forecasting.EngineConfig{
        Enabled:        true,
        UpdateInterval: 1 * time.Minute,
        Thresholds: map[string]float64{
            "critical_backlog":    1000,
            "high_backlog":        500,
            "critical_error_rate": 0.1,
            "high_error_rate":     0.05,
        },
    },
}
```

### Time Series Models

#### EWMA (Exponentially Weighted Moving Average)

Simple, responsive model for short-term predictions:

```go
ewma := forecasting.NewEWMAForecaster(&forecasting.EWMAConfig{
    Alpha:              0.3,
    AutoAdjust:         true,
    MinObservations:    5,
    ConfidenceInterval: 0.95,
})

// Update with observations
ewma.Update(100.0)
ewma.Update(110.0)
ewma.Update(105.0)

// Generate forecast
forecast, err := ewma.Forecast(60) // 60-minute horizon
```

#### Holt-Winters Triple Exponential Smoothing

Advanced model capturing trend and seasonality:

```go
hw := forecasting.NewHoltWintersForecaster(&forecasting.HoltWintersConfig{
    Alpha:            0.3,
    Beta:             0.1,
    Gamma:            0.1,
    SeasonLength:     24,
    SeasonalMethod:   "multiplicative",
    AutoDetectSeason: true,
})

// Needs 2+ seasons of data for initialization
for i := 0; i < 48; i++ {
    hw.Update(data[i])
}

forecast, err := hw.Forecast(120) // 2-hour horizon
```

### Recommendation Engine

Translates forecasts into actionable guidance:

```go
engine := forecasting.NewRecommendationEngine(config, logger)

forecasts := map[forecasting.MetricType]*forecasting.ForecastResult{
    forecasting.MetricBacklog: backlogForecast,
    forecasting.MetricThroughput: throughputForecast,
    forecasting.MetricErrorRate: errorForecast,
}

currentMetrics := &forecasting.QueueMetrics{
    Backlog:       500,
    Throughput:    10.5,
    ErrorRate:     0.02,
    ActiveWorkers: 5,
}

recommendations := engine.GenerateRecommendations(forecasts, currentMetrics)
```

## API Methods

### UpdateMetrics

```go
func (fe *ForecastingEngine) UpdateMetrics(metrics *QueueMetrics) error
```

Updates the engine with new metric observations.

**Example:**
```go
metrics := &forecasting.QueueMetrics{
    Timestamp:     time.Now(),
    Backlog:       750,
    Throughput:    12.5,
    ErrorRate:     0.015,
    LatencyP50:    45.0,
    LatencyP95:    120.0,
    LatencyP99:    250.0,
    ActiveWorkers: 8,
    QueueName:     "high-priority",
}

err := engine.UpdateMetrics(metrics)
```

### GetForecasts

```go
func (fe *ForecastingEngine) GetForecasts(horizonMinutes int) (map[MetricType]*ForecastResult, error)
```

Generates forecasts for the specified time horizon.

**Example:**
```go
forecasts, err := engine.GetForecasts(120) // 2-hour forecast

for metricType, forecast := range forecasts {
    fmt.Printf("%s forecast: %.2f (confidence: %.2f)\n",
        metricType, forecast.Points[0], forecast.Confidence)
}
```

### GetRecommendations

```go
func (fe *ForecastingEngine) GetRecommendations() []Recommendation
```

Returns current recommendations based on latest forecasts.

**Example:**
```go
recs := engine.GetRecommendations()

for _, rec := range recs {
    switch rec.Priority {
    case forecasting.PriorityCritical:
        // Immediate action required
        executeAction(rec.Action)
    case forecasting.PriorityHigh:
        // Schedule action
        scheduleAction(rec.Action, rec.Timing)
    }
}
```

### GetHistoricalData

```go
func (fe *ForecastingEngine) GetHistoricalData(
    metricType MetricType,
    queueName string,
    duration time.Duration) []DataPoint
```

Retrieves historical metric data.

**Example:**
```go
history := engine.GetHistoricalData(
    forecasting.MetricBacklog,
    "high-priority",
    24 * time.Hour,
)

for _, point := range history {
    fmt.Printf("%s: %.2f\n", point.Timestamp, point.Value)
}
```

## Data Types

### QueueMetrics

```go
type QueueMetrics struct {
    Timestamp      time.Time
    Backlog        int64
    Throughput     float64
    ErrorRate      float64
    LatencyP50     float64
    LatencyP95     float64
    LatencyP99     float64
    ActiveWorkers  int
    QueueName      string
}
```

### ForecastResult

```go
type ForecastResult struct {
    Points         []float64     // Predicted values
    UpperBounds    []float64     // Upper confidence bounds
    LowerBounds    []float64     // Lower confidence bounds
    Confidence     float64       // Model confidence (0-1)
    ModelUsed      string        // Model identifier
    GeneratedAt    time.Time
    HorizonMinutes int
    MetricType     MetricType
}
```

### Recommendation

```go
type Recommendation struct {
    ID          string
    Priority    RecommendationPriority
    Category    RecommendationCategory
    Title       string
    Description string
    Action      string                 // Actionable command
    Timing      time.Duration          // When to act
    Confidence  float64
    CreatedAt   time.Time
}
```

### AccuracyMetrics

```go
type AccuracyMetrics struct {
    MAE            float64   // Mean Absolute Error
    RMSE           float64   // Root Mean Square Error
    MAPE           float64   // Mean Absolute Percentage Error
    PredictionBias float64   // Systematic over/under prediction
    R2Score        float64   // Coefficient of determination
    SampleSize     int
    LastUpdated    time.Time
}
```

## Usage Examples

### Basic Forecasting

```go
// Initialize engine
engine := forecasting.NewForecastingEngine(config, logger)
defer engine.Stop()

// Feed metrics
for {
    metrics := collectMetrics()
    engine.UpdateMetrics(metrics)
    
    // Get forecasts
    forecasts, _ := engine.GetForecasts(60)
    
    // Check for critical conditions
    if backlog := forecasts[forecasting.MetricBacklog]; backlog != nil {
        if backlog.Points[30] > 1000 { // 30 minutes ahead
            scaleWorkers(calculateNeededWorkers(backlog))
        }
    }
    
    time.Sleep(1 * time.Minute)
}
```

### Automated Scaling

```go
func autoScale(engine *forecasting.ForecastingEngine) {
    recs := engine.GetRecommendations()
    
    for _, rec := range recs {
        if rec.Category == forecasting.CategoryCapacityScaling {
            if rec.Priority == forecasting.PriorityCritical {
                // Execute scaling immediately
                cmd := exec.Command("sh", "-c", rec.Action)
                if err := cmd.Run(); err != nil {
                    log.Error("Scaling failed", err)
                } else {
                    log.Info("Scaled workers", rec.Description)
                }
            }
        }
    }
}
```

### SLO Monitoring

```go
tracker := forecasting.NewSLOTracker()
tracker.SetTarget(0.999) // 99.9% availability

// Update with error rates
tracker.Update(metrics.ErrorRate)

// Check budget
budget := tracker.GetCurrentBudget()
if budget.WeeklyBurnRate > 0.8 {
    alert("SLO budget critical: %.1f%% consumed", 
          budget.WeeklyBurnRate * 100)
}

// Project future burn
forecast := getErrorForecast()
futureBudget := tracker.ProjectBudgetBurn(forecast.Points)
if futureBudget.TimeToExhaustion < 24*time.Hour {
    alert("SLO budget will exhaust in %v", 
          futureBudget.TimeToExhaustion)
}
```

### Maintenance Window Planning

```go
func findMaintenanceWindow(engine *forecasting.ForecastingEngine) {
    recs := engine.GetRecommendations()
    
    for _, rec := range recs {
        if rec.Category == forecasting.CategoryMaintenanceScheduling {
            // Parse window from recommendation
            window := parseMaintenanceWindow(rec.Description)
            
            // Schedule maintenance
            scheduleMaintenance(window.Start, window.End)
            
            // Notify team
            notify("Maintenance scheduled: %s", rec.Description)
            break
        }
    }
}
```

## Model Selection Guide

### When to Use EWMA

- Short-term predictions (< 1 hour)
- Rapidly changing metrics
- Limited historical data
- Real-time responsiveness needed

### When to Use Holt-Winters

- Long-term predictions (> 2 hours)
- Clear seasonal patterns (daily, weekly)
- Stable trending metrics
- Sufficient historical data (2+ cycles)

### Model Accuracy Evaluation

```go
accuracy := engine.GetModelAccuracy()

for model, metrics := range accuracy {
    fmt.Printf("Model: %s\n", model)
    fmt.Printf("  MAPE: %.2f%%\n", metrics.MAPE)
    fmt.Printf("  RMSE: %.2f\n", metrics.RMSE)
    fmt.Printf("  R²: %.3f\n", metrics.R2Score)
    
    // Switch models based on accuracy
    if metrics.MAPE > 20 {
        // Consider alternative model
    }
}
```

## Performance Considerations

### Computational Complexity

- EWMA Update: O(1)
- EWMA Forecast: O(n) for n-step horizon
- Holt-Winters Update: O(1)
- Holt-Winters Forecast: O(n)
- Storage Query: O(log n) for time range

### Memory Usage

- Per time series: ~8KB for 1000 points
- Per model: < 1KB state
- Forecast cache: ~10KB per result
- Total for 100 queues: ~10MB

### Optimization Tips

1. **Batch Updates**: Update multiple metrics together
2. **Cache Forecasts**: Reuse forecasts within update interval
3. **Aggregate Historical Data**: Use appropriate granularity
4. **Limit Horizon**: Longer horizons have lower accuracy
5. **Selective Models**: Only run models for critical metrics

## Best Practices

### Data Quality

```go
// Validate metrics before updating
if metrics.Throughput < 0 || metrics.ErrorRate > 1 {
    log.Warn("Invalid metrics detected", metrics)
    return
}

// Handle missing data
if metrics.Backlog == 0 && previousBacklog > 1000 {
    // Likely data collection issue, not actual zero
    metrics.Backlog = previousBacklog
}
```

### Model Tuning

```go
// Start conservative
config := &EWMAConfig{
    Alpha: 0.1, // Slow adaptation
    AutoAdjust: true,
}

// Monitor accuracy
if accuracy.MAPE > 15 {
    config.Alpha = 0.3 // Increase responsiveness
}
```

### Recommendation Handling

```go
// Implement cooldown
lastAction := make(map[string]time.Time)

for _, rec := range recommendations {
    if last, ok := lastAction[rec.Category]; ok {
        if time.Since(last) < 5*time.Minute {
            continue // Skip to avoid thrashing
        }
    }
    
    executeRecommendation(rec)
    lastAction[rec.Category] = time.Now()
}
```

## Troubleshooting

### Inaccurate Forecasts

1. Check data quality and consistency
2. Verify sufficient historical data
3. Adjust model parameters (alpha, beta, gamma)
4. Consider seasonal patterns
5. Evaluate model selection

### Missing Recommendations

1. Verify thresholds are appropriate
2. Check forecast confidence levels
3. Ensure metrics are being updated
4. Review cooldown periods
5. Check engine configuration

### High Memory Usage

1. Reduce retention duration
2. Increase aggregation interval
3. Limit number of models
4. Clear old forecast cache
5. Disable disk persistence if not needed
