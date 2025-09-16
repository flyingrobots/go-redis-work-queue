# Chaos Harness API Documentation

## Overview

The Chaos Harness provides controlled fault injection and chaos testing capabilities for the go-redis-work-queue system. It allows you to inject failures, test resilience, and validate recovery behaviors.

## Core Components

### Fault Injectors

Fault injectors introduce controlled failures into the system:

- **Latency Injection**: Adds artificial delays to operations
- **Error Injection**: Forces operations to return errors
- **Panic Injection**: Triggers panics in workers (with recovery)
- **Partial Failure**: Fails a percentage of items in batch operations
- **Resource Hogging**: Simulates resource exhaustion

### Chaos Scenarios

Scenarios orchestrate multiple fault injections over time:

- **Multi-stage execution**: Different faults at different stages
- **Load generation**: Concurrent load during chaos testing
- **Guardrails**: Safety limits to prevent damage
- **Metrics collection**: Track system behavior during chaos

## REST API Endpoints

### Injector Management

#### List Injectors
```http
GET /api/v1/chaos/injectors
```

Returns all active fault injectors.

**Response:**
```json
{
  "injectors": [
    {
      "id": "latency-global-1",
      "type": "latency",
      "scope": "global",
      "enabled": true,
      "probability": 0.3,
      "parameters": {
        "latency_ms": 500,
        "jitter_ms": 100
      },
      "ttl": "5m",
      "expires_at": "2025-01-14T10:30:00Z"
    }
  ],
  "count": 1
}
```

#### Create Injector
```http
POST /api/v1/chaos/injectors
```

Creates a new fault injector.

**Request:**
```json
{
  "type": "error",
  "scope": "worker",
  "scope_value": "worker-1",
  "enabled": true,
  "probability": 0.1,
  "parameters": {
    "error_message": "simulated database error"
  },
  "ttl": "10m"
}
```

**Response:**
```json
{
  "id": "error-worker-1234567890",
  "type": "error",
  "scope": "worker",
  "scope_value": "worker-1",
  "enabled": true,
  "probability": 0.1,
  "parameters": {
    "error_message": "simulated database error"
  },
  "ttl": "10m",
  "expires_at": "2025-01-14T10:15:00Z",
  "created_at": "2025-01-14T10:05:00Z",
  "created_by": "api"
}
```

#### Delete Injector
```http
DELETE /api/v1/chaos/injectors/{id}
```

Removes a fault injector.

#### Toggle Injector
```http
POST /api/v1/chaos/injectors/{id}/toggle
```

Enables or disables an injector.

**Request:**
```json
{
  "enabled": false
}
```

### Scenario Management

#### List Scenarios
```http
GET /api/v1/chaos/scenarios
```

Returns available and running scenarios.

**Response:**
```json
{
  "scenarios": [
    {
      "id": "latency-test",
      "name": "Latency Injection Test",
      "description": "Tests system behavior under increased latency",
      "duration": "5m",
      "stages": [
        {
          "name": "Baseline",
          "duration": "1m",
          "load_config": {
            "rps": 100,
            "pattern": "constant"
          }
        },
        {
          "name": "Inject Latency",
          "duration": "2m",
          "injectors": [...]
        }
      ]
    }
  ],
  "running": []
}
```

#### Run Scenario
```http
POST /api/v1/chaos/scenarios/{id}/run
```

Executes a chaos scenario.

**Response:**
```json
{
  "status": "started",
  "scenario_id": "latency-test"
}
```

#### Abort Scenario
```http
POST /api/v1/chaos/scenarios/{id}/abort
```

Aborts a running scenario.

#### Get Scenario Report
```http
GET /api/v1/chaos/scenarios/{id}/report
```

Returns execution report for a scenario.

**Response:**
```json
{
  "scenario_id": "latency-test",
  "scenario_name": "Latency Injection Test",
  "executed_at": "2025-01-14T10:00:00Z",
  "duration": "5m",
  "result": "passed",
  "metrics": {
    "total_requests": 30000,
    "successful_requests": 29500,
    "failed_requests": 500,
    "injected_faults": 9000,
    "recovery_time": "45s",
    "error_rate": 0.017,
    "latency_p50": "25ms",
    "latency_p95": "500ms",
    "latency_p99": "1200ms"
  },
  "findings": [
    {
      "severity": "medium",
      "type": "slow_recovery",
      "description": "System took longer than expected to recover",
      "impact": "Increased latency for 45 seconds after fault removal"
    }
  ],
  "recommendations": [
    "Implement circuit breaker for downstream services",
    "Add connection pooling with health checks",
    "Reduce timeout values for faster failure detection"
  ]
}
```

### Control Endpoints

#### Get Status
```http
GET /api/v1/chaos/status
```

Returns chaos harness status.

**Response:**
```json
{
  "status": "active",
  "active_injectors": 3,
  "running_scenarios": 1,
  "config": {
    "enabled": true,
    "allow_production": false
  }
}
```

#### Clear All
```http
POST /api/v1/chaos/clear
```

Removes all injectors and stops all scenarios.

## Programmatic API

### Using the Chaos Harness

```go
import "internal/chaos-harness"

// Initialize chaos harness
config := chaosharness.DefaultConfig()
ch := chaosharness.NewChaosHarness(logger, config)

// Register API routes
router := mux.NewRouter()
ch.RegisterRoutes(router)

// Inject faults programmatically
ctx := context.Background()

// Inject latency
latency := ch.InjectLatency(ctx, chaosharness.ScopeWorker, "worker-1")

// Inject error
err := ch.InjectError(ctx, chaosharness.ScopeQueue, "high-priority")

// Check if fault should be injected
should, injector := ch.ShouldInject(ctx, 
    chaosharness.ScopeGlobal, "", 
    chaosharness.InjectorLatency)
```

### Creating Custom Scenarios

```go
scenario := &chaosharness.ChaosScenario{
    ID:          "custom-chaos",
    Name:        "Custom Chaos Test",
    Description: "Tests multiple failure modes",
    Duration:    10 * time.Minute,
    Stages: []chaosharness.ScenarioStage{
        {
            Name:     "Warm-up",
            Duration: 2 * time.Minute,
            LoadConfig: &chaosharness.LoadConfig{
                RPS:     50,
                Pattern: chaosharness.LoadLinear,
            },
        },
        {
            Name:     "Inject Chaos",
            Duration: 5 * time.Minute,
            Injectors: []chaosharness.FaultInjector{
                {
                    Type:        chaosharness.InjectorLatency,
                    Scope:       chaosharness.ScopeGlobal,
                    Probability: 0.3,
                    Parameters: map[string]interface{}{
                        "latency_ms": 1000,
                    },
                },
                {
                    Type:        chaosharness.InjectorError,
                    Scope:       chaosharness.ScopeWorker,
                    ScopeValue:  "worker-2",
                    Probability: 0.1,
                },
            },
            LoadConfig: &chaosharness.LoadConfig{
                RPS:       100,
                Pattern:   chaosharness.LoadSpike,
                BurstSize: 500,
            },
        },
        {
            Name:     "Recovery",
            Duration: 3 * time.Minute,
            LoadConfig: &chaosharness.LoadConfig{
                RPS:     50,
                Pattern: chaosharness.LoadConstant,
            },
        },
    },
    Guardrails: chaosharness.ScenarioGuardrails{
        MaxErrorRate:     0.25,
        MaxLatencyP99:    5 * time.Second,
        MaxBacklogSize:   10000,
        RequireConfirm:   true,
        AutoAbortOnPanic: true,
    },
}

// Run scenario
err := ch.RunScenario(ctx, scenario)
```

## Injector Types

### Latency Injection

**Type:** `latency`

**Parameters:**
- `latency_ms`: Base latency in milliseconds
- `jitter_ms`: Random jitter range (optional)

**Example:**
```json
{
  "type": "latency",
  "parameters": {
    "latency_ms": 500,
    "jitter_ms": 100
  }
}
```

### Error Injection

**Type:** `error`

**Parameters:**
- `error_message`: Error message to return

**Example:**
```json
{
  "type": "error",
  "parameters": {
    "error_message": "connection refused"
  }
}
```

### Panic Injection

**Type:** `panic`

**Parameters:**
- `panic_message`: Panic message (optional)

**Example:**
```json
{
  "type": "panic",
  "parameters": {
    "panic_message": "simulated panic"
  }
}
```

### Partial Failure

**Type:** `partial_fail`

**Parameters:**
- `fail_rate`: Percentage of items to fail (0.0-1.0)

**Example:**
```json
{
  "type": "partial_fail",
  "parameters": {
    "fail_rate": 0.3
  }
}
```

## Scopes

### Global Scope

**Scope:** `global`

Applies to all operations system-wide.

### Worker Scope

**Scope:** `worker`
**Scope Value:** Worker ID

Applies to specific worker instances.

### Queue Scope

**Scope:** `queue`
**Scope Value:** Queue name

Applies to operations on specific queues.

### Tenant Scope

**Scope:** `tenant`
**Scope Value:** Tenant ID

Applies to operations for specific tenants.

## Load Patterns

### Constant

Maintains steady RPS throughout.

### Linear

Ramps up linearly from 0 to target RPS.

### Sine

Oscillates in sine wave pattern.

### Spike

Normal load with periodic bursts.

### Random

Random variation around target RPS.

## Safety Features

### TTL (Time To Live)

All injectors have TTL to prevent forgotten chaos:

- Default: 5 minutes
- Maximum: 1 hour
- Auto-cleanup after expiry

### Guardrails

Scenarios can define safety limits:

- `max_error_rate`: Abort if error rate exceeds threshold
- `max_latency_p99`: Abort if P99 latency too high
- `max_backlog_size`: Abort if backlog grows too large
- `require_confirm`: Require user confirmation before running
- `allow_production`: Whether to allow in production environment
- `auto_abort_on_panic`: Automatically abort on panic detection

### Production Safety

By default, chaos harness:

- Disabled in production (`allow_production: false`)
- Shows "CHAOS MODE" banner in UI
- Requires typed confirmation for dangerous operations
- Logs all chaos activities

## Best Practices

1. **Start Small**: Begin with low probability and short duration
2. **Use TTLs**: Always set reasonable TTLs on injectors
3. **Monitor Metrics**: Watch system metrics during chaos testing
4. **Set Guardrails**: Define safety limits for all scenarios
5. **Test in Staging**: Thoroughly test scenarios in non-production first
6. **Document Findings**: Record all discoveries and improvements
7. **Automate Recovery**: Ensure system can recover automatically
8. **Regular Testing**: Run chaos tests regularly, not just once

## Example Scenarios

### Redis Failover Test

Tests system behavior during Redis failover:

```json
{
  "id": "redis-failover",
  "name": "Redis Failover Simulation",
  "duration": "10m",
  "stages": [
    {
      "name": "Normal Operations",
      "duration": "2m"
    },
    {
      "name": "Redis Latency",
      "duration": "2m",
      "injectors": [
        {
          "type": "redis_latency",
          "probability": 0.5,
          "parameters": {
            "latency_ms": 2000
          }
        }
      ]
    },
    {
      "name": "Redis Down",
      "duration": "30s",
      "injectors": [
        {
          "type": "redis_drop",
          "probability": 1.0
        }
      ]
    },
    {
      "name": "Recovery",
      "duration": "5m30s"
    }
  ]
}
```

### Worker Failure Test

Tests handling of worker failures:

```json
{
  "id": "worker-failures",
  "name": "Worker Failure Handling",
  "duration": "5m",
  "stages": [
    {
      "name": "Random Worker Errors",
      "duration": "2m",
      "injectors": [
        {
          "type": "error",
          "scope": "worker",
          "scope_value": "*",
          "probability": 0.2
        }
      ]
    },
    {
      "name": "Worker Panics",
      "duration": "1m",
      "injectors": [
        {
          "type": "panic",
          "scope": "worker",
          "probability": 0.05
        }
      ]
    }
  ]
}
```

## Metrics and Reporting

### Collected Metrics

- Total requests
- Successful/failed requests
- Injected faults count
- Recovery time
- Backlog size
- Error rate
- Latency percentiles (P50, P95, P99)
- Time series data with fault markers

### Report Generation

Reports include:

- Scenario execution summary
- Metrics analysis
- Findings (issues discovered)
- Recommendations for improvements
- Time series graphs with chaos events marked

## Troubleshooting

### Injector Not Working

1. Check if chaos harness is enabled
2. Verify injector probability (1.0 = always)
3. Check scope matches your operations
4. Ensure injector hasn't expired (TTL)
5. Verify injector is enabled

### Scenario Won't Start

1. Check for already running scenarios
2. Verify scenario configuration is valid
3. Ensure guardrails aren't too restrictive
4. Check chaos harness has required permissions

### System Not Recovering

1. Check if all injectors were removed
2. Verify no panic loops occurring
3. Check resource exhaustion
4. Review backlog accumulation
5. Ensure retry mechanisms are working
