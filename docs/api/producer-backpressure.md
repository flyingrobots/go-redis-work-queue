# Producer Backpressure API Documentation

## Overview

The Producer Backpressure system provides intelligent flow control for job producers, preventing system overload through adaptive throttling and circuit breaking. The system monitors queue backlogs and provides real-time throttling recommendations to producers.

## Architecture

The backpressure system consists of several key components:

- **BackpressureController**: Main interface for throttling decisions
- **CircuitBreaker**: Prevents cascade failures during extreme load
- **StatsProvider**: Provides queue statistics for decision making
- **MetricsCollector**: Exports monitoring data

## Core Interfaces

### BackpressureController

The main interface for backpressure operations:

```go
type BackpressureController interface {
    // SuggestThrottle returns recommended delay for given priority and queue
    SuggestThrottle(ctx context.Context, priority Priority, queueName string) (*ThrottleDecision, error)

    // Run executes work function with automatic throttling
    Run(ctx context.Context, priority Priority, queueName string, work func() error) error

    // ProcessBatch processes multiple jobs with backpressure awareness
    ProcessBatch(ctx context.Context, jobs []BatchJob) error

    // GetCircuitState returns current circuit breaker state for queue
    GetCircuitState(queueName string) CircuitState

    // SetManualOverride enables/disables manual override mode
    SetManualOverride(enabled bool)

    // Start begins background polling and maintenance
    Start(ctx context.Context) error

    // Stop shuts down the controller gracefully
    Stop() error

    // Health returns controller health status
    Health() map[string]interface{}
}
```

### StatsProvider

Interface for retrieving queue statistics:

```go
type StatsProvider interface {
    GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error)
    GetAllQueueStats(ctx context.Context) (map[string]*QueueStats, error)
}
```

## Configuration

### BackpressureConfig

Complete configuration for the backpressure system:

```go
type BackpressureConfig struct {
    Thresholds BacklogThresholds `json:"thresholds"`
    Circuit    CircuitConfig     `json:"circuit"`
    Polling    PollingConfig     `json:"polling"`
    Recovery   RecoveryStrategy  `json:"recovery"`
}
```

### BacklogThresholds

Defines throttling thresholds for different priority levels:

```go
type BacklogThresholds struct {
    HighPriority   BacklogWindow `json:"high_priority"`
    MediumPriority BacklogWindow `json:"medium_priority"`
    LowPriority    BacklogWindow `json:"low_priority"`
}

type BacklogWindow struct {
    Green  int `json:"green_max"`  // 0-Green: no throttling
    Yellow int `json:"yellow_max"` // Green-Yellow: light throttling
    Red    int `json:"red_max"`    // Yellow-Red: heavy throttling/shedding
}
```

**Default Values:**
- High Priority: Green=1000, Yellow=5000, Red=10000
- Medium Priority: Green=500, Yellow=2000, Red=5000
- Low Priority: Green=100, Yellow=500, Red=1000

### CircuitConfig

Circuit breaker configuration:

```go
type CircuitConfig struct {
    FailureThreshold  int           `json:"failure_threshold"`  // Trip after N failures
    RecoveryThreshold int           `json:"recovery_threshold"` // Close after N successes
    TripWindow        time.Duration `json:"trip_window"`        // Time window for failure counting
    RecoveryTimeout   time.Duration `json:"recovery_timeout"`   // Wait before half-open
    ProbeInterval     time.Duration `json:"probe_interval"`     // Half-open probe frequency
}
```

**Default Values:**
- FailureThreshold: 5
- RecoveryThreshold: 3
- TripWindow: 30 seconds
- RecoveryTimeout: 60 seconds
- ProbeInterval: 5 seconds

### PollingConfig

Statistics polling configuration:

```go
type PollingConfig struct {
    Interval   time.Duration `json:"interval"`      // Base polling interval
    Jitter     time.Duration `json:"jitter"`        // Jitter to prevent thundering herd
    Timeout    time.Duration `json:"timeout"`       // API call timeout
    MaxBackoff time.Duration `json:"max_backoff"`   // Maximum backoff on failures
    CacheTTL   time.Duration `json:"cache_ttl"`     // How long to cache throttle decisions
    Enabled    bool          `json:"enabled"`       // Enable/disable polling
}
```

**Default Values:**
- Interval: 5 seconds
- Jitter: 1 second
- Timeout: 3 seconds
- MaxBackoff: 60 seconds
- CacheTTL: 30 seconds
- Enabled: true

## Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/flyingrobots/go-redis-work-queue/internal/producer-backpressure"
    "go.uber.org/zap"
)

func main() {
    // Create configuration
    config := backpressure.DefaultConfig()

    // Create stats provider (implement based on your queue system)
    statsProvider := NewYourStatsProvider()

    // Create logger
    logger, _ := zap.NewProduction()

    // Create controller
    controller, err := backpressure.NewController(config, statsProvider, logger)
    if err != nil {
        log.Fatal(err)
    }

    // Start controller
    ctx := context.Background()
    if err := controller.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer controller.Stop()

    // Use controller...
}
```

### Simple Throttling

```go
// Get throttling recommendation
decision, err := controller.SuggestThrottle(ctx, backpressure.MediumPriority, "email-queue")
if err != nil {
    return err
}

// Handle recommendation
if decision.ShouldShed {
    return backpressure.ErrJobShed
}

if decision.Delay > 0 {
    time.Sleep(decision.Delay)
}

// Proceed with work
return enqueueJob(payload)
```

### Automatic Throttling

```go
// Use Run() for automatic throttling
err := controller.Run(ctx, backpressure.HighPriority, "payments-queue", func() error {
    return enqueuePayment(payment)
})

if backpressure.IsShedError(err) {
    // Job was shed - handle gracefully
    return requeueForLater(payment)
}

return err
```

### Batch Processing

```go
jobs := []backpressure.BatchJob{
    {
        Priority:  backpressure.HighPriority,
        QueueName: "payments-queue",
        Work: func() error {
            return enqueuePayment(payment1)
        },
    },
    {
        Priority:  backpressure.LowPriority,
        QueueName: "analytics-queue",
        Work: func() error {
            return enqueueAnalytics(event1)
        },
    },
}

err := controller.ProcessBatch(ctx, jobs)
// Some jobs might be shed, but critical ones will be processed
```

## Priority Levels

The system supports three priority levels with different throttling behaviors:

### HighPriority
- **Use for**: Critical operations (payments, authentication, high-value user actions)
- **Throttling**: Minimal throttling even under high load
- **Shedding**: Never shed, always processed
- **Delay scaling**: 50% of base throttling delay

### MediumPriority
- **Use for**: Standard operations (notifications, reports, user requests)
- **Throttling**: Standard throttling based on backlog
- **Shedding**: Throttled but not shed
- **Delay scaling**: 100% of base throttling delay

### LowPriority
- **Use for**: Background operations (analytics, cleanup, bulk processing)
- **Throttling**: Aggressive throttling under load
- **Shedding**: Shed when system is overloaded
- **Delay scaling**: 150% of base throttling delay

## Throttling Algorithm

The system uses a three-zone throttling algorithm:

### Green Zone (Healthy)
- **Condition**: Backlog ≤ Green threshold
- **Action**: No throttling
- **Delay**: 0ms

### Yellow Zone (Warning)
- **Condition**: Green < Backlog ≤ Yellow threshold
- **Action**: Light throttling
- **Delay**: 10ms to 500ms (linear scaling)

### Red Zone (Critical)
- **Condition**: Backlog > Yellow threshold
- **Action**: Heavy throttling/shedding
- **Delay**: 500ms to 5s (with priority scaling)
- **Shedding**: Low priority jobs shed when ratio > 80%

## Circuit Breaker States

### Closed (Normal)
- **Behavior**: All requests allowed
- **Transition**: Opens after failure threshold reached

### Open (Blocked)
- **Behavior**: All requests blocked/shed
- **Transition**: Moves to half-open after recovery timeout

### Half-Open (Probing)
- **Behavior**: Limited probes allowed
- **Transition**: Closes on success threshold or opens on failure

## Error Handling

### Error Types

```go
// Common errors
var (
    ErrJobShed              = errors.New("job shed due to backpressure")
    ErrCircuitOpen          = errors.New("circuit breaker is open")
    ErrControllerNotStarted = errors.New("backpressure controller not started")
    ErrStatsUnavailable     = errors.New("queue statistics unavailable")
)

// Check error types
if backpressure.IsShedError(err) {
    // Handle job shedding
}

if backpressure.IsCircuitOpenError(err) {
    // Handle circuit breaker open
}

if backpressure.IsRetryable(err) {
    // Retry operation
}
```

### Fallback Strategies

When queue statistics are unavailable, the system can:

1. **Conservative Fallback**: Apply moderate throttling
2. **Permissive Fallback**: Allow all requests through
3. **Circuit Protection**: Use circuit breaker state only

Configure via `RecoveryStrategy.FallbackMode`.

## Monitoring and Metrics

### Prometheus Metrics

The system exports comprehensive Prometheus metrics:

```go
// Throttle events by priority and queue
backpressure_throttle_events_total{priority="medium",queue="email-queue"}

// Job shed events by priority and queue
backpressure_shed_events_total{priority="low",queue="analytics-queue"}

// Throttle delay distribution
backpressure_throttle_delay_seconds{priority="high",queue="payments-queue"}

// Circuit breaker states (0=closed, 1=open, 2=half-open)
backpressure_circuit_breaker_state{queue="email-queue"}

// Current queue backlog sizes
backpressure_queue_backlog_size{queue="email-queue"}

// Producer compliance rates
backpressure_producer_compliance_ratio{queue="email-queue"}

// Polling errors
backpressure_polling_errors_total{error_type="timeout"}

// Cache hit rates
backpressure_cache_hit_rate
```

### Health Monitoring

```go
health := controller.Health()

// Returns:
// {
//   "started": true,
//   "stopped": false,
//   "manual_override": false,
//   "emergency_mode": false,
//   "cache_hit_rate": 0.85,
//   "cache_size": 42,
//   "circuit_states": {
//     "email-queue": "closed",
//     "analytics-queue": "half-open"
//   },
//   "polling_enabled": true,
//   "last_fallback": "2025-01-15T10:30:00Z"
// }
```

## Manual Controls

### Override Mode

Disable all throttling temporarily:

```go
// Disable throttling for emergency maintenance
controller.SetManualOverride(true)

// Re-enable throttling
controller.SetManualOverride(false)
```

### Circuit Breaker Controls

```go
// Check circuit state
state := controller.GetCircuitState("email-queue")

// Get detailed circuit stats
cb := controller.getOrCreateCircuitBreaker("email-queue")
stats := cb.GetStats()

// Manual circuit control
cb.ForceOpen()  // Force circuit open
cb.ForceClose() // Force circuit closed
cb.Reset()      // Reset to clean state
```

## Performance Characteristics

### Latency
- **Throttle Decision**: <1ms typical
- **Cache Hit**: <0.1ms
- **Cache Miss**: 1-5ms (depending on stats provider)

### Memory Usage
- **Controller Overhead**: ~50KB base
- **Per Queue**: ~1KB (circuit breaker + cache entries)
- **Cache**: ~100 bytes per cached decision

### CPU Usage
- **Steady State**: <0.1% CPU
- **High Load**: <1% CPU (with 1000+ RPS)

### Network
- **Polling**: 1 API call per interval (default: 5s)
- **Bandwidth**: <1KB per poll

## Best Practices

### Configuration
1. **Start Conservative**: Begin with default thresholds and adjust
2. **Monitor Metrics**: Watch shed rates and compliance
3. **Environment-Specific**: Different thresholds for prod vs staging

### Integration
1. **Graceful Degradation**: Handle shed errors appropriately
2. **Retry Logic**: Implement exponential backoff for retries
3. **Priority Classification**: Correctly classify job priorities

### Monitoring
1. **Alert on High Shed Rates**: >10 sheds/minute for critical queues
2. **Alert on Circuit Trips**: Any circuit breaker opening
3. **Monitor Compliance**: Producer compliance should be >90%

### Testing
1. **Load Testing**: Verify behavior under various load conditions
2. **Failure Testing**: Test circuit breaker behavior
3. **Integration Testing**: Verify with real queue systems

## Migration Guide

### From No Backpressure

1. **Start with Monitoring**: Deploy with all thresholds very high
2. **Establish Baselines**: Monitor normal queue depths for 1 week
3. **Set Conservative Thresholds**: Set thresholds 2x normal peak
4. **Gradually Tighten**: Reduce thresholds weekly while monitoring

### Configuration Evolution

```go
// Week 1: Monitoring only
config.Thresholds.LowPriority.Red = 10000  // Very high
config.Recovery.FallbackMode = false       // Allow all through

// Week 2: Conservative throttling
config.Thresholds.LowPriority.Red = 2000   // Based on monitoring

// Week 3: Normal operation
config.Thresholds = backpressure.DefaultThresholds()
```

## Troubleshooting

### Common Issues

**High Shed Rates**
- Check if thresholds are too aggressive
- Verify queue processing capacity
- Monitor for seasonal traffic patterns

**Circuit Breaker Tripping**
- Investigate underlying queue/worker health
- Check for configuration issues
- Verify stats provider connectivity

**Poor Performance**
- Check cache hit rates (should be >80%)
- Verify polling interval isn't too frequent
- Monitor stats provider latency

**Unexpected Behavior**
- Verify manual override is disabled
- Check emergency mode status
- Validate configuration with `config.Validate()`

### Debug Information

```go
// Enable debug logging
logger = logger.With(zap.String("component", "backpressure"))

// Check health status
health := controller.Health()
logger.Info("Controller health", zap.Any("health", health))

// Monitor specific queue
state := controller.GetCircuitState("problematic-queue")
logger.Info("Circuit state", zap.String("state", state.String()))
```