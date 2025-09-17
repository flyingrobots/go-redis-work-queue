# Anomaly Radar + SLO Budget API Documentation

## Overview

The Anomaly Radar + SLO Budget system provides real-time monitoring of queue health with Service Level Objective (SLO) tracking and error budget management. This system implements industry-standard SRE practices for proactive incident detection and reliability management.

## Features

- **Real-time Anomaly Detection**: Monitors backlog growth, error rates, and latency percentiles
- **SLO Budget Tracking**: Calculates and tracks error budget consumption with burn rate alerts
- **Priority-based Alerting**: Configurable warning and critical thresholds for all metrics
- **Rolling Window Analysis**: Maintains historical metrics for trend analysis
- **Burn Rate Monitoring**: Fast and slow burn rate detection with time-to-exhaustion calculations
- **Lightweight Footprint**: Optimized for minimal CPU and memory usage

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Metrics        │    │  Anomaly Radar   │    │  Alert System  │
│  Collector      │────│  Engine          │────│  & Callbacks   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         v                       v                       v
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Data Sources   │    │  SLO Budget      │    │  HTTP API       │
│  (Queue Stats)  │    │  Calculator      │    │  Endpoints      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Getting Started

### Installation

```go
import anomalyradarslobudget "github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"
```

### Basic Usage

```go
// Create configuration
config := anomalyradarslobudget.DefaultConfig()

// Customize for your environment
config.SLO.AvailabilityTarget = 0.999  // 99.9%
config.SLO.LatencyThresholdMs = 500    // 500ms
config.Thresholds.ErrorRateWarning = 0.01  // 1%

// Create metrics collector
collector := anomalyradarslobudget.NewSimpleMetricsCollector(
    func() int64 { return getBacklogSize() },
    func() int64 { return getRequestCount() },
    func() int64 { return getErrorCount() },
    func() (float64, float64, float64) { return getLatencyPercentiles() },
)

// Create and start anomaly radar
radar := anomalyradarslobudget.New(config, collector)
err := radar.Start(context.Background())
if err != nil {
    log.Fatal(err)
}

// Register alert callback
radar.RegisterAlertCallback(func(alert anomalyradarslobudget.Alert) {
    log.Printf("Alert: %s - %s", alert.Type.String(), alert.Message)
})

// Set up HTTP endpoints
handler := anomalyradarslobudget.NewHTTPHandler(radar)
mux := http.NewServeMux()
handler.RegisterRoutes(mux, "/api/v1/anomaly-radar")
```

## Configuration

### SLO Configuration

```go
type SLOConfig struct {
    AvailabilityTarget    float64           // Target availability (0.995 = 99.5%)
    LatencyPercentile     float64           // Percentile to monitor (0.95 = p95)
    LatencyThresholdMs    int64             // Latency threshold in milliseconds
    Window               time.Duration      // SLO measurement window
    BurnRateThresholds   BurnRateThresholds // Alert thresholds
}
```

### Anomaly Thresholds

```go
type AnomalyThresholds struct {
    BacklogGrowthWarning  float64  // Items/second warning threshold
    BacklogGrowthCritical float64  // Items/second critical threshold
    ErrorRateWarning      float64  // Error rate warning (0-1)
    ErrorRateCritical     float64  // Error rate critical (0-1)
    LatencyP95Warning     float64  // P95 latency warning (ms)
    LatencyP95Critical    float64  // P95 latency critical (ms)
}
```

### Recommended Configurations

#### High-Criticality Systems
```go
config := anomalyradarslobudget.GetRecommendedConfig(
    5000.0,              // expectedQPS
    500*time.Millisecond, // targetLatency
    "critical",          // systemCriticality
)
```

#### Medium-Criticality Systems
```go
config := anomalyradarslobudget.GetRecommendedConfig(
    1000.0,              // expectedQPS
    time.Second,         // targetLatency
    "medium",            // systemCriticality
)
```

## HTTP API Endpoints

### Get Current Status

**GET** `/api/v1/anomaly-radar/status`

Returns current anomaly detection status and SLO budget information.

**Query Parameters:**
- `include_metrics` (boolean): Include recent metrics in response
- `metric_window` (duration): Time window for included metrics (default: 1h)

**Response:**
```json
{
  "anomaly_status": {
    "backlog_status": "healthy",
    "error_rate_status": "warning",
    "latency_status": "healthy",
    "overall_status": "warning",
    "active_alerts": [
      {
        "id": "error_rate",
        "type": "error_rate",
        "severity": "warning",
        "message": "error_rate is warning: 0.02 (threshold: 0.01)",
        "value": 0.02,
        "threshold": 0.01,
        "created_at": "2025-09-14T19:45:30Z",
        "updated_at": "2025-09-14T19:46:15Z"
      }
    ],
    "last_updated": "2025-09-14T19:46:15Z"
  },
  "slo_budget": {
    "config": {
      "availability_target": 0.995,
      "latency_percentile": 0.95,
      "latency_threshold_ms": 1000,
      "window": "720h0m0s"
    },
    "total_budget": 150.0,
    "consumed_budget": 85.5,
    "remaining_budget": 64.5,
    "budget_utilization": 0.57,
    "current_burn_rate": 0.012,
    "time_to_exhaustion": "89h30m",
    "is_healthy": true,
    "alert_level": "warning",
    "last_calculated": "2025-09-14T19:46:15Z"
  },
"timestamp": "2025-09-14T19:46:15Z"
}
```

Durations are encoded using Go’s `time.Duration` format (e.g., `72h`, `720h0m0s`, `30m`, `1h30m`, `1500ms`, or negative offsets like `-5m`). Clients should parse these values accordingly, including optional fractional seconds.

### Get Configuration

**GET** `/api/v1/anomaly-radar/config`

**Authentication/Authorization**

This endpoint requires Admin API credentials. Supply an `Authorization: Bearer <token>` header that carries an admin-scoped JWT (scope: `admin` or the equivalent capability for anomaly radar management). Requests missing valid tokens receive `401 Unauthorized`; tokens lacking required scope receive `403 Forbidden`.

Returns current configuration.

**Response:**
```json
{
  "config": {
    "slo": {
      "availability_target": 0.995,
      "latency_percentile": 0.95,
      "latency_threshold_ms": 1000,
      "window": "720h0m0s",
      "burn_rate_thresholds": {
        "fast_burn_rate": 0.01,
        "fast_burn_window": "1h30m",
        "slow_burn_rate": 0.05,
        "slow_burn_window": "6h"
      }
    },
    "thresholds": {
      "backlog_growth_warning": 10.0,
      "backlog_growth_critical": 50.0,
      "error_rate_warning": 0.01,
      "error_rate_critical": 0.05,
      "latency_p95_warning": 500.0,
      "latency_p95_critical": 1000.0
    },
    "monitoring_interval": "10s",
    "metric_retention": "24h0m0s",
    "max_snapshots": 8640,
    "sampling_rate": 1.0
  },
  "summary": "Anomaly Radar Configuration:\n  SLO Target: 99.50% availability, p95 < 1000ms\n  ...",
  "is_valid": true,
  "timestamp": "2025-09-14T19:46:15Z"
}
```

### Update Configuration

**PUT** `/api/v1/anomaly-radar/config`

Updates the configuration.

**Request Body:**
```json
{
  "slo": {
    "availability_target": 0.999,
    "latency_threshold_ms": 500
  },
  "thresholds": {
    "error_rate_warning": 0.005,
    "error_rate_critical": 0.02
  }
}
```

### Get Historical Metrics

**GET** `/api/v1/anomaly-radar/metrics`

Returns historical metric snapshots.

**Query Parameters:**
- `window` (duration): Time window for metrics (default: 24h)
- `max_samples` (integer): Maximum number of samples to return

**Response:**
```json
{
  "metrics": [
    {
      "timestamp": "2025-09-14T19:45:30Z",
      "backlog_size": 1250,
      "backlog_growth_rate": 12.5,
      "request_count": 5000,
      "error_count": 25,
      "error_rate": 0.005,
      "p50_latency_ms": 150.5,
      "p95_latency_ms": 485.2,
      "p99_latency_ms": 890.1
    }
  ],
  "window": "24h0m0s",
  "count": 1,
  "timestamp": "2025-09-14T19:46:15Z"
}
```

### Get Active Alerts

**GET** `/api/v1/anomaly-radar/alerts`

Returns currently active alerts.

**Response:**
```json
{
  "alerts": [
    {
      "id": "error_rate",
      "type": "error_rate",
      "severity": "warning",
      "message": "error_rate is warning: 0.02 (threshold: 0.01)",
      "value": 0.02,
      "threshold": 0.01,
      "created_at": "2025-09-14T19:45:30Z",
      "updated_at": "2025-09-14T19:46:15Z"
    }
  ],
  "count": 1,
  "timestamp": "2025-09-14T19:46:15Z"
}
```

### Get SLO Budget Details

**GET** `/api/v1/anomaly-radar/slo-budget`

Returns detailed SLO budget information with insights.

**Response:**
```json
{
  "slo_budget": {
    "total_budget": 150.0,
    "consumed_budget": 85.5,
    "remaining_budget": 64.5,
    "budget_utilization": 0.57,
    "current_burn_rate": 0.012,
    "time_to_exhaustion": "89h30m0s",
    "is_healthy": true,
    "alert_level": "warning"
  },
  "insights": {
    "budget_exhausted_percentage": 57.0,
    "budget_remaining_percentage": 43.0,
    "is_budget_healthy": true,
    "days_since_window_start": 30,
    "hours_to_exhaustion": 89.5,
    "projected_exhaustion_date": "2025-09-18T13:16:15Z"
  },
  "timestamp": "2025-09-14T19:46:15Z"
}
```

### Get Latency Percentiles

**GET** `/api/v1/anomaly-radar/percentiles`

Returns latency percentiles for a given time window.

**Query Parameters:**
- `window` (duration): Time window for calculation (default: 1h)

**Response:**
```json
{
  "percentiles": {
    "p50": 145.2,
    "p90": 385.7,
    "p95": 485.2,
    "p99": 890.1
  },
  "window": "1h0m0s",
  "timestamp": "2025-09-14T19:46:15Z"
}
```

### Health Check

**GET** `/api/v1/anomaly-radar/health`

Returns health status of the anomaly radar system.

**Response:**
```json
{
  "is_running": true,
  "status": "warning",
  "alert_level": "warning",
  "active_alerts": 1,
  "last_updated": "2025-09-14T19:46:15Z",
  "uptime": "2h15m30s",
  "timestamp": "2025-09-14T19:46:15Z"
}
```

**HTTP Status Codes:**
- `200 OK`: System is healthy or has warnings
- `503 Service Unavailable`: System is not running or has critical issues

### Start/Stop Operations

**POST** `/api/v1/anomaly-radar/start`

Starts the anomaly radar monitoring.

**POST** `/api/v1/anomaly-radar/stop`

Stops the anomaly radar monitoring.

## Alert Types

### Backlog Growth Alerts
Triggered when queue backlog grows faster than configured thresholds.

### Error Rate Alerts
Triggered when error rate exceeds warning or critical thresholds.

### Latency Alerts
Triggered when P95 latency exceeds configured thresholds.

### Burn Rate Alerts
Triggered when SLO budget consumption rate indicates budget exhaustion risk.

## Metrics and Monitoring

### Key Metrics

- **Backlog Size**: Current number of jobs in queue
- **Backlog Growth Rate**: Rate of queue growth (items/second)
- **Request Count**: Total requests processed
- **Error Count**: Total errors encountered
- **Error Rate**: Percentage of requests that resulted in errors
- **Latency Percentiles**: P50, P95, P99 latency measurements

### SLO Budget Metrics

- **Total Budget**: Maximum allowed errors in SLO window
- **Consumed Budget**: Errors consumed so far
- **Remaining Budget**: Budget remaining
- **Budget Utilization**: Percentage of budget consumed (0-1)
- **Burn Rate**: Rate of budget consumption (budget/hour)
- **Time to Exhaustion**: Estimated time until budget is fully consumed

## Best Practices

### Configuration

1. **Set Realistic SLO Targets**: Choose availability targets based on business requirements
2. **Calibrate Thresholds**: Adjust warning/critical thresholds based on baseline metrics
3. **Monitor Burn Rates**: Set appropriate burn rate thresholds for early warning
4. **Regular Review**: Periodically review and adjust thresholds based on system evolution

### Integration

1. **Implement Proper Metrics Collection**: Ensure accurate and timely metrics
2. **Set Up Alert Callbacks**: Integrate with your alerting system (PagerDuty, Slack, etc.)
3. **Dashboard Integration**: Display key metrics in operational dashboards
4. **Incident Response**: Include SLO budget status in incident response procedures

### Performance

1. **Optimize Sampling**: Adjust sampling rate based on system load
2. **Manage Retention**: Balance historical data needs with memory usage
3. **Monitor Resource Usage**: Track CPU and memory usage of the monitoring system
4. **Batch Operations**: Use batch endpoints for efficient data retrieval

## Examples

### Custom Metrics Collector

```go
type QueueMetricsCollector struct {
    queueService *QueueService
    statsCache   *StatsCache
}

func (c *QueueMetricsCollector) CollectMetrics(ctx context.Context) (anomalyradarslobudget.MetricSnapshot, error) {
    stats, err := c.queueService.GetStats(ctx)
    if err != nil {
        return anomalyradarslobudget.MetricSnapshot{}, err
    }

    latencies := c.statsCache.GetRecentLatencies()

    return anomalyradarslobudget.MetricSnapshot{
        Timestamp:    time.Now(),
        BacklogSize:  stats.QueueLength,
        RequestCount: stats.ProcessedCount,
        ErrorCount:   stats.ErrorCount,
        P50LatencyMs: latencies.P50,
        P95LatencyMs: latencies.P95,
        P99LatencyMs: latencies.P99,
    }, nil
}
```

### Alert Integration

```go
radar.RegisterAlertCallback(func(alert anomalyradarslobudget.Alert) {
    switch alert.Severity {
    case anomalyradarslobudget.AlertLevelCritical:
        pagerduty.TriggerIncident(alert.Message)
    case anomalyradarslobudget.AlertLevelWarning:
        slack.SendAlert(alert.Message)
    }
})
```

### Dashboard Integration

```go
// Prometheus metrics export
func (r *AnomalyRadar) ExportPrometheusMetrics() {
    anomalyStatus, sloBudget := r.GetCurrentStatus()

    // Export budget utilization
    budgetUtilizationGauge.Set(sloBudget.BudgetUtilization)

    // Export burn rate
    burnRateGauge.Set(sloBudget.CurrentBurnRate)

    // Export alert counts by severity
    for _, alert := range anomalyStatus.ActiveAlerts {
        alertCountVec.WithLabelValues(alert.Severity.String()).Inc()
    }
}
```

## Troubleshooting

### Common Issues

1. **No Metrics Collected**: Check metrics collector implementation and error handling
2. **Inaccurate SLO Calculations**: Verify time window configuration and data retention
3. **False Alerts**: Adjust thresholds based on baseline system behavior
4. **High Memory Usage**: Reduce max snapshots or metric retention period
5. **Performance Impact**: Increase monitoring interval or reduce sampling rate

### Debug Information

Use the debug endpoint (if implemented) to inspect internal state:

```bash
curl http://localhost:8080/api/v1/anomaly-radar/debug
```

### Logging

Enable debug logging to troubleshoot issues:

```go
// In your metrics collector
log.Printf("Collected metrics: backlog=%d, errors=%d, latency=%.2f",
    snapshot.BacklogSize, snapshot.ErrorCount, snapshot.P95LatencyMs)
```

## Security Considerations

1. **Authentication**: Secure API endpoints with appropriate authentication
2. **Authorization**: Limit access to configuration endpoints
3. **Input Validation**: Validate all configuration inputs
4. **Resource Limits**: Set appropriate limits on API response sizes
5. **Audit Logging**: Log configuration changes and administrative actions
