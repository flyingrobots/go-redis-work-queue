# Job Budgeting API Documentation

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| High | Cost / Governance | Metrics export, tenant labels, RL integration | Misestimation, incentive gaming, PII | ~400–700 | Medium | 5–8 (Fib) | High |

## Overview

The Job Budgeting system provides comprehensive cost tracking, budget management, and enforcement for job queue operations. This system transforms queue management from resource guesswork into precise financial planning through real-time cost tracking, intelligent forecasting, and graduated enforcement that prevents budget overruns while maintaining service quality.

## Core Components

### Cost Calculation Engine

The cost calculation engine uses a weighted formula that captures the true resource consumption of each job:

```go
type CostModel struct {
    CPUTimeWeight         float64 // $/second
    MemoryWeight          float64 // $/MB·second
    PayloadWeight         float64 // $/KB
    RedisOpsWeight        float64 // $/operation
    NetworkWeight         float64 // $/MB transferred
    BaseJobWeight         float64 // Fixed cost per job
    EnvironmentMultiplier float64 // Production vs staging
}
```

#### Usage Example

```go
// Initialize cost engine
model := budgeting.DefaultCostModel()
engine := budgeting.NewCostCalculationEngine(model)

// Calculate job cost
metrics := budgeting.JobMetrics{
    JobID:           "job-123",
    TenantID:        "acme-corp",
    QueueName:       "ml-training",
    CPUTime:         15.2,  // 15.2 seconds
    MemoryMBSeconds: 2048,  // 2GB for 1 second
    PayloadBytes:    51200, // 50KB
    RedisOps:        25,    // 25 operations
    NetworkBytes:    10240, // 10KB network transfer
    JobType:         "ml-training",
    Priority:        5,
}

cost, err := engine.CalculateJobCost(metrics)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job cost: $%.4f\n", cost.TotalCost)
fmt.Printf("Breakdown: CPU=$%.4f, Memory=$%.4f, Payload=$%.4f\n",
    cost.CostBreakdown.CPUCost,
    cost.CostBreakdown.MemoryCost,
    cost.CostBreakdown.PayloadCost)
```

### Budget Management

Budgets operate on a hierarchical model supporting both tenant-level and queue-level allocations:

```go
type Budget struct {
    ID                string
    TenantID          string
    QueueName         string  // Empty = tenant-wide budget
    Period            BudgetPeriod
    Amount            float64
    WarningThreshold  float64 // 0.75 = 75%
    ThrottleThreshold float64 // 0.90 = 90%
    BlockThreshold    float64 // 1.00 = 100%
    EnforcementPolicy EnforcementPolicy
    Notifications     []NotificationChannel
}
```

#### Creating Budgets

```go
// Initialize budget service
config := budgeting.Config{
    DatabaseURL:   "postgresql://localhost/budgets",
    DefaultTenant: "acme-corp",
    CostModel:     budgeting.ProductionCostModel(),
}

service, err := budgeting.NewBudgetService(config)
if err != nil {
    log.Fatal(err)
}

// Create monthly budget
budget := &budgeting.Budget{
    TenantID:  "acme-corp",
    QueueName: "ml-training", // Queue-specific budget
    Amount:    5000.0,        // $5000/month
    Currency:  "USD",
    Period: budgeting.BudgetPeriod{
        Type:      "monthly",
        StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
    },
    WarningThreshold:  0.75, // Warn at 75%
    ThrottleThreshold: 0.90, // Throttle at 90%
    BlockThreshold:    1.00, // Block at 100%
    EnforcementPolicy: budgeting.EnforcementPolicy{
        WarnOnly:         false,
        ThrottleFactor:   0.5,  // 50% capacity when throttling
        BlockNewJobs:     true,
        AllowEmergency:   true, // High priority bypass
        GracePeriodHours: 24,   // 24hr grace period
    },
    Notifications: []budgeting.NotificationChannel{
        {
            Type:   "email",
            Target: "ops@acme-corp.com",
            Events: []string{"warning", "throttle", "block"},
        },
        {
            Type:   "slack",
            Target: "https://hooks.slack.com/services/...",
            Events: []string{"throttle", "block"},
        },
    },
}

err = service.CreateBudget(budget)
if err != nil {
    log.Fatal(err)
}
```

### Enforcement System

The enforcement system provides graduated responses that maintain system stability while controlling costs:

```go
// Check budget compliance before processing job
action, err := service.CheckBudgetCompliance("acme-corp", "ml-training", 5)
if err != nil {
    log.Fatal(err)
}

switch action.Type {
case "block":
    fmt.Printf("Job blocked: %s\n", action.Message)
    return fmt.Errorf("budget enforcement: %s", action.Message)

case "throttle":
    fmt.Printf("Job throttled to %.0f%% capacity: %s\n",
        action.Factor*100, action.Message)
    // Apply throttling to job processing

case "warn":
    fmt.Printf("Budget warning: %s\n", action.Message)
    // Continue processing but log warning

case "allow":
    // Process job normally
}
```

### Forecasting and Analytics

The forecasting system uses linear regression with seasonal adjustments to predict month-end spending:

```go
// Generate spending forecast
forecast, err := service.GetForecast("acme-corp", "ml-training")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Forecast for period ending %s:\n",
    forecast.PeriodEnd.Format("2006-01-02"))
fmt.Printf("  Predicted spend: $%.2f\n", forecast.PredictedSpend)
fmt.Printf("  Budget utilization: %.1f%%\n", forecast.BudgetUtilization*100)
fmt.Printf("  Trend: %s\n", forecast.TrendDirection)
fmt.Printf("  Recommendation: %s\n", forecast.Recommendation)

if forecast.DaysUntilOverrun != nil {
    fmt.Printf("  Days until budget overrun: %d\n", *forecast.DaysUntilOverrun)
}
```

### Cost Aggregation and Reporting

The system aggregates costs into daily summaries for efficient analysis:

```go
// Generate comprehensive budget report
period := budgeting.BudgetPeriod{
    Type:      "monthly",
    StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
}

report, err := service.GenerateReport("acme-corp", period)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Budget Report for %s\n", report.TenantID)
fmt.Printf("Period: %s to %s\n",
    report.Period.StartDate.Format("2006-01-02"),
    report.Period.EndDate.Format("2006-01-02"))
fmt.Printf("Total Spend: $%.2f / $%.2f (%.1f%%)\n",
    report.TotalSpend, report.BudgetAmount, report.Utilization*100)

fmt.Println("\nTop Cost Drivers:")
for i, driver := range report.TopDrivers[:5] {
    fmt.Printf("%d. %s: $%.2f (%.1f%%) - %d jobs @ $%.4f avg\n",
        i+1, driver.QueueName, driver.TotalCost, driver.Percentage,
        driver.JobCount, driver.AvgCostPerJob)
}

fmt.Println("\nQueue Breakdown:")
for queue, cost := range report.QueueBreakdown {
    percentage := cost / report.TotalSpend * 100
    fmt.Printf("  %s: $%.2f (%.1f%%)\n", queue, cost, percentage)
}

fmt.Println("\nRecommendations:")
for _, rec := range report.Recommendations {
    fmt.Printf("  • %s\n", rec)
}
```

## Terminal User Interface (TUI)

The system includes a comprehensive TUI for budget management and monitoring:

```go
// Start TUI for interactive budget management
err = service.StartTUI()
if err != nil {
    log.Fatal(err)
}
```

### TUI Features

| Key | Action | Description |
|-----|--------|-------------|
| `1` | Overview | Budget summary and cost drivers |
| `2` | Trends | Daily spending charts and forecasts |
| `3` | Controls | Budget creation and management |
| `4` | Alerts | Budget violations and notifications |
| `r` | Refresh | Update all data in current view |
| `q` | Quit | Exit application |

#### Overview Tab
- Real-time budget utilization by queue
- Current spending vs. budget amounts
- Status indicators (OK, Warning, Throttle, Block)
- Top cost drivers with job counts and averages

#### Trends Tab
- 30-day daily spending chart (text-based)
- Cost breakdown by component (CPU, Memory, Payload, Redis, Network)
- Trend analysis and seasonal patterns

#### Controls Tab
- Budget creation and editing forms
- Threshold configuration (warning, throttle, block)
- Enforcement policy settings
- Notification channel management

#### Alerts Tab
- Active budget violations
- Alert history and acknowledgment
- Notification delivery status

## Cost Model Calibration

The system supports cost model calibration against real infrastructure costs:

```go
// Initialize calibrator
calibrator := budgeting.NewModelCalibrator()

// Add benchmark data from actual infrastructure costs
benchmark := budgeting.BenchmarkResult{
    JobType:            "ml-training",
    ActualCost:         0.45,   // Actual AWS/GCP cost
    CalculatedCost:     0.42,   // Model calculation
    CPUTime:            30.0,
    MemoryMBSeconds:    4096,
    PayloadBytes:       102400,
    RedisOps:           50,
    NetworkBytes:       20480,
    InfrastructureCost: 0.45,
    Timestamp:          time.Now(),
}

calibrator.AddBenchmark(benchmark)

// Calibrate model after collecting sufficient benchmarks (>10)
currentModel := service.GetCostModel()
calibratedModel, err := calibrator.CalibrateModel(currentModel)
if err != nil {
    log.Fatal(err)
}

// Apply calibrated model
err = service.UpdateCostModel(calibratedModel)
if err != nil {
    log.Fatal(err)
}

// Check calibration accuracy
report := calibrator.GetCalibrationReport()
fmt.Printf("Model accuracy: %.1f%% (%s)\n",
    report.Accuracy*100, report.RecommendedAction)
```

## Integration with Rate Limiting

Budget information feeds directly into the rate limiting system:

```go
// Example integration with rate limiter
type BudgetAwareRateLimiter struct {
    baseLimiter    *RateLimiter
    budgetService  *budgeting.BudgetService
}

func (r *BudgetAwareRateLimiter) CheckRateLimit(tenantID, queueName string) error {
    // Check budget compliance first
    action, err := r.budgetService.CheckBudgetCompliance(tenantID, queueName, 5)
    if err != nil {
        return err
    }

    switch action.Type {
    case "block":
        return fmt.Errorf("rate limited: %s", action.Message)
    case "throttle":
        // Apply budget-based throttling
        return r.applyThrottling(action.Factor)
    default:
        // Apply normal rate limiting
        return r.baseLimiter.CheckLimit(tenantID, queueName)
    }
}
```

## Notification Channels

### Email Notifications

```go
emailChannel := budgeting.NotificationChannel{
    Type:   "email",
    Target: "ops@company.com",
    Events: []string{"warning", "throttle", "block"},
    Enabled: true,
    Metadata: map[string]string{
        "subject_prefix": "[BUDGET]",
        "template":       "budget_alert_template",
    },
}
```

### Slack Integration

```go
slackChannel := budgeting.NotificationChannel{
    Type:   "slack",
    Target: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
    Events: []string{"throttle", "block"},
    Enabled: true,
    Metadata: map[string]string{
        "channel":  "#ops-alerts",
        "username": "Budget Monitor",
        "icon":     ":moneybag:",
    },
}
```

### Webhook Notifications

```go
webhookChannel := budgeting.NotificationChannel{
    Type:   "webhook",
    Target: "https://your-api.com/budget-webhook",
    Events: []string{"warning", "throttle", "block", "reset"},
    Enabled: true,
    Metadata: map[string]string{
        "auth_header": "Bearer your-api-token",
        "timeout":     "30s",
        "retry_count": "3",
    },
}
```

## API Reference

### BudgetService Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `ProcessJobCost` | Process job cost metrics | `JobMetrics` | `error` |
| `CheckBudgetCompliance` | Check budget compliance | `tenantID, queueName, priority` | `EnforcementAction, error` |
| `CreateBudget` | Create new budget | `*Budget` | `error` |
| `GetBudgetStatus` | Get budget status | `budgetID` | `*BudgetStatus, error` |
| `GetForecast` | Generate forecast | `tenantID, queueName` | `*Forecast, error` |
| `GenerateReport` | Generate budget report | `tenantID, period` | `*BudgetReport, error` |
| `EstimateJobCost` | Estimate job cost | `jobType, cpuTime, payloadKB` | `float64` |

### Error Handling

The system provides structured error types for different failure modes:

```go
// Check for specific error types
var budgetErr *budgeting.BudgetError
if errors.As(err, &budgetErr) {
    fmt.Printf("Budget operation %s failed for tenant %s: %v\n",
        budgetErr.Operation, budgetErr.TenantID, budgetErr.Err)
}

var enforcementErr *budgeting.EnforcementError
if errors.As(err, &enforcementErr) {
    fmt.Printf("Budget enforcement %s: $%.2f/$%.2f (%.1f%%)\n",
        enforcementErr.Action, enforcementErr.CurrentSpend,
        enforcementErr.BudgetAmount, enforcementErr.Utilization*100)
}

// Check error classification
if budgeting.IsRetryable(err) {
    // Retry operation
    time.Sleep(time.Second)
    return retryOperation()
}

if budgeting.IsPermanent(err) {
    // Don't retry, handle failure
    return handlePermanentFailure(err)
}

// Get stable error code for logging/monitoring
code := budgeting.ErrorCode(err)
log.Printf("Budget error [%s]: %v", code, err)
```

## Database Schema

The system uses PostgreSQL with the following schema:

```sql
-- Daily cost aggregates
CREATE TABLE daily_costs (
    tenant_id VARCHAR(50) NOT NULL,
    queue_name VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    total_jobs INTEGER NOT NULL DEFAULT 0,
    total_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    cpu_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    memory_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    payload_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    redis_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    network_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    avg_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    max_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    min_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    p95_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, queue_name, date)
);

-- Budget definitions
CREATE TABLE budgets (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    queue_name VARCHAR(100) NOT NULL DEFAULT '',
    period_type VARCHAR(20) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    warning_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.75,
    throttle_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.90,
    block_threshold DECIMAL(3,2) NOT NULL DEFAULT 1.00,
    enforcement_policy JSONB NOT NULL DEFAULT '{}',
    notifications JSONB NOT NULL DEFAULT '[]',
    tags JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) NOT NULL DEFAULT 'system'
);
```

## Configuration

### Service Configuration

```yaml
# config/budget.yaml
database_url: "postgresql://user:pass@localhost/budgets"
flush_interval: "5m"
retention_days: 365
default_tenant: "default"
enable_tui: true

cost_model:
  cpu_time_weight: 0.0001
  memory_weight: 0.00001
  payload_weight: 0.00002
  redis_ops_weight: 0.000001
  network_weight: 0.01
  base_job_weight: 0.001
  environment_multiplier: 1.0

notification_channels:
  - type: "email"
    target: "ops@company.com"
    events: ["warning", "throttle", "block"]
  - type: "slack"
    target: "https://hooks.slack.com/services/..."
    events: ["throttle", "block"]
```

### Environment-Specific Models

```go
// Production environment
prodConfig := budgeting.Config{
    CostModel: budgeting.ProductionCostModel(), // 2x multiplier
    DatabaseURL: "postgresql://prod-db/budgets",
}

// Staging environment
stagingConfig := budgeting.Config{
    CostModel: budgeting.StagingCostModel(), // 0.5x multiplier
    DatabaseURL: "postgresql://staging-db/budgets",
}
```

## Performance Considerations

- **Aggregation Efficiency**: O(1) lookups for current spend using in-memory cache with periodic DB sync
- **Forecasting Complexity**: O(n log n) for trend analysis where n is historical data points (typically 30-90 days)
- **Enforcement Latency**: < 1ms budget checks using pre-calculated thresholds and cached values
- **Storage Growth**: ~100MB per million jobs with daily aggregation and 1-year retention
- **Real-time Updates**: WebSocket-based updates for budget panels with 1-second refresh rate

## Monitoring and Observability

Key metrics to track system health and adoption:

```go
type BudgetMetrics struct {
    // Usage metrics
    ActiveBudgets        int64   `metric:"active_budgets_total"`
    BudgetChecksPerSec   float64 `metric:"budget_checks_per_second"`
    CostCalculationTime  float64 `metric:"cost_calculation_duration_ms"`

    // Budget compliance
    BudgetWarnings       int64   `metric:"budget_warnings_total"`
    BudgetThrottles      int64   `metric:"budget_throttles_total"`
    BudgetBlocks         int64   `metric:"budget_blocks_total"`

    // Accuracy metrics
    ForecastAccuracy     float64 `metric:"forecast_accuracy_percent"`
    CostModelDrift       float64 `metric:"cost_model_drift_percent"`

    // Financial impact
    TotalSpendTracked    float64 `metric:"total_spend_tracked_dollars"`
    CostSavingsFromThrottling float64 `metric:"cost_savings_dollars"`
}
```

## Best Practices

### Budget Design

1. **Start Conservative**: Begin with warning-only budgets to understand spending patterns
2. **Graduated Thresholds**: Use 75% warning, 90% throttle, 100% block for balanced control
3. **Queue-Specific Budgets**: Create separate budgets for different workload types
4. **Grace Periods**: Allow 24-48 hour grace periods for new budgets
5. **Emergency Bypass**: Always enable emergency bypass for critical high-priority jobs

### Cost Optimization

1. **Regular Calibration**: Calibrate cost models monthly against actual infrastructure costs
2. **Component Analysis**: Monitor cost breakdowns to identify optimization opportunities
3. **Seasonal Adjustment**: Account for business cycles in forecasting
4. **Trend Analysis**: Watch for cost trend changes that indicate efficiency gains or degradation

### Operational Guidelines

1. **Alert Fatigue**: Limit alerts to actionable events and use rate limiting
2. **Forecasting**: Review forecasts weekly and adjust budgets based on business needs
3. **Reporting**: Generate monthly reports for cost review and optimization planning
4. **Testing**: Test enforcement policies in staging before production deployment

## Troubleshooting

### Common Issues

1. **Inaccurate Cost Calculations**: Check cost model calibration and benchmark data
2. **False Budget Alerts**: Verify budget periods and seasonal adjustments
3. **Enforcement Not Working**: Check budget active status and threshold configuration
4. **Poor Forecast Accuracy**: Ensure sufficient historical data (>30 days)

### Debugging Tools

```go
// Enable debug logging
service.EnableDebugLogging(true)

// Check cost model accuracy
report := calibrator.GetCalibrationReport()
fmt.Printf("Model accuracy: %.1f%%\n", report.Accuracy*100)

// Validate budget configuration
err := service.ValidateBudget(budget)
if err != nil {
    fmt.Printf("Budget validation failed: %v\n", err)
}

// Test notification channels
for _, channel := range budget.Notifications {
    err := notifier.SendTestNotification(channel)
    if err != nil {
        fmt.Printf("Notification test failed for %s: %v\n", channel.Type, err)
    }
}
```

---

This comprehensive API documentation provides everything needed to implement and operate the Job Budgeting system effectively. The system transforms queue operations from resource guesswork into precise financial planning with real-time cost tracking, intelligent forecasting, and graduated enforcement.