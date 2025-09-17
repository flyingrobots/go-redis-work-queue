# Anomaly Radar + SLO Budget API Documentation

## Overview

The Anomaly Radar + SLO Budget system provides real-time monitoring of queue health with Service Level Objective (SLO) tracking and error budget management. This system implements industry-standard SRE practices for proactive incident detection and reliability management. A draft OpenAPI specification is available at [docs/api/anomaly-radar-openapi.yaml](anomaly-radar-openapi.yaml).

## Versioning & Deprecation

- **Supported versions**: `v1` (current). Minor/patch updates must remain backward compatible; breaking changes require a new major version.
- **Compatibility guarantees**: Request/response schemas additively evolve; existing fields are never repurposed.
- **Breaking-change policy**: Proposed breaking changes require SRE approval and a published migration plan.
- **Deprecation timeline**: Minimum 90 days’ notice before removing or altering deprecated fields or endpoints.
- **Change log**: Updates are recorded in `docs/changelog/anomaly-radar-slo-budget.md` and release notes.
- **Client migration**: Consumers should pin to versioned routes (e.g., `/api/v1/…`) and watch for `Deprecation`/`Sunset` headers; see types such as `SLOConfig`, `BurnRateThresholds`, and `Alert` for field-level guidance.

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
import ars "github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"
```

### Basic Usage

```go
// Create configuration
config := ars.DefaultConfig()

// Customize for your environment
config.SLO.AvailabilityTarget = 0.999  // 99.9%
config.SLO.LatencyThresholdMs = 500    // 500ms
config.Thresholds.ErrorRateWarning = 0.01  // 1%

// Create metrics collector (see QueueMetricsCollector example below)
collector := &QueueMetricsCollector{Client: metricsClient}

// Create and start anomaly radar
radar := ars.New(config, collector)
err := radar.Start(context.Background())
if err != nil {
    log.Fatal(err)
}

// Register alert callback
radar.RegisterAlertCallback(func(alert ars.Alert) {
    log.Printf("Alert: %s - %s", alert.Type.String(), alert.Message)
})

// Set up HTTP endpoints
handler := ars.NewHTTPHandler(radar)
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

type BurnRateThresholds struct {
    FastBurnRate    float64       // Fast-path burn rate (budget/hour)
    FastBurnWindow  time.Duration // Evaluation window for fast burn (e.g., 5m)
    SlowBurnRate    float64       // Slow-path burn rate (budget/hour)
    SlowBurnWindow  time.Duration // Evaluation window for slow burn (e.g., 6h)
}
```

`budget_utilization` is a fraction [0,1], `current_burn_rate` measures budget consumed per hour, and `time_to_exhaustion` uses Go's `time.Duration` encoding (e.g., `89h30m0s`).

### Anomaly Thresholds

```go
type AnomalyThresholds struct {
    BacklogGrowthWarning  float64  // Warning threshold (items/second)
    BacklogGrowthCritical float64  // Critical threshold (items/second)
    ErrorRateWarning      float64  // Warning threshold (fraction 0–1)
    ErrorRateCritical     float64  // Critical threshold (fraction 0–1)
    LatencyP95Warning     float64  // Warning threshold (milliseconds)
    LatencyP95Critical    float64  // Critical threshold (milliseconds)
}

| Field                   | Unit            | Valid Range |
|-------------------------|-----------------|-------------|
| BacklogGrowthWarning    | items/second    | ≥0          |
| BacklogGrowthCritical   | items/second    | ≥ BacklogGrowthWarning |
| ErrorRateWarning        | fraction (0–1)  | 0 ≤ x ≤ 1   |
| ErrorRateCritical       | fraction (0–1)  | ErrorRateCritical ≥ ErrorRateWarning |
| LatencyP95Warning       | milliseconds    | ≥0          |
| LatencyP95Critical      | milliseconds    | ≥ LatencyP95Warning |
```

### Recommended Configurations

#### High-Criticality Systems
```go
config := ars.GetRecommendedConfig(
    5000.0,              // expectedQPS
    500*time.Millisecond, // targetLatency
    "critical",          // systemCriticality
)
```

#### Medium-Criticality Systems
```go
config := ars.GetRecommendedConfig(
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

Durations use Go's `time.ParseDuration` syntax (e.g., `30m`, `1h`, `7h30m`).

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

`budget_utilization` is a fraction between 0 and 1, `current_burn_rate` expresses budget consumed per hour, and `time_to_exhaustion` is encoded using Go's `time.Duration`/RFC3339 duration representation (e.g., `89h30m`).

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

Durations use Go's `time.ParseDuration` format (e.g., `30m`, `6h`, `7h30m`).
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
      "p90_latency_ms": 320.4,
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

Durations use Go's `time.ParseDuration` format (e.g., `30m`, `6h`, `7h30m`).

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
- `500 Internal Server Error`: Unexpected failure while evaluating system health
- `429 Too Many Requests`: Health checks throttled (caller should back off)
- `206 Partial Content`: Health data available but degraded (e.g., metrics backend unreachable)

### Start/Stop Operations

**POST** `/api/v1/anomaly-radar/start`

Starts the anomaly radar monitoring.

- **Auth**: Requires bearer token with `slo_admin` (or `anomaly_radar:manage`) permission.
- **Idempotency**: Returns `200 OK` with `{ "status": "already_started" }` if radar is already running; otherwise returns `202 Accepted` while startup completes.
- **Errors**: `401` for missing auth, `403` for insufficient scope, `500` for startup failures.

Example response when already running:

```json
{
  "status": "already_started",
  "timestamp": "2025-09-14T19:46:15Z"
}
```

**POST** `/api/v1/anomaly-radar/stop`

Stops the anomaly radar monitoring.

- **Auth**: Requires bearer token with `slo_admin` (or `anomaly_radar:manage`) permission.
- **Idempotency**: Returns `200 OK` with `{ "status": "already_stopped" }` if radar is not running; otherwise returns `202 Accepted` while shutdown completes.
- **Errors**: `401` for missing auth, `403` for insufficient scope, `500` for shutdown failures.

Example response when already stopped:

```json
{
  "status": "already_stopped",
  "timestamp": "2025-09-14T19:46:15Z"
}
```

Concurrent start/stop requests are handled safely; the service guarantees consistent state is returned in the response body.

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

## Examples

### Metrics Collector Interface

```go
type MetricsCollector interface {
    CollectMetrics(ctx context.Context) (ars.MetricSnapshot, error)
}
```

### Custom Metrics Collector

```go
type QueueMetricsCollector struct {
    queueService *QueueService
    statsCache   *StatsCache
}

func (c *QueueMetricsCollector) CollectMetrics(ctx context.Context) (ars.MetricSnapshot, error) {
    stats, err := c.queueService.GetStats(ctx)
    if err != nil {
        return ars.MetricSnapshot{}, err
    }

    latencies := c.statsCache.GetRecentLatencies()

    return ars.MetricSnapshot{
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
radar.RegisterAlertCallback(func(alert ars.Alert) {
    switch alert.Severity {
    case ars.AlertLevelCritical:
        pagerduty.TriggerIncident(alert.Message)
    case ars.AlertLevelWarning:
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

    // Export alert counts by severity (idempotent)
    counts := map[string]float64{}
    for _, alert := range anomalyStatus.ActiveAlerts {
        counts[alert.Severity.String()]++
    }
    for severity, value := range counts {
        alertCountVec.WithLabelValues(severity).Set(value)
    }
    for _, severity := range []string{"info", "warning", "critical"} {
        if _, ok := counts[severity]; !ok {
            alertCountVec.WithLabelValues(severity).Set(0)
        }
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
