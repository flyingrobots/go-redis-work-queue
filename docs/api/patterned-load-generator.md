# Patterned Load Generator API Documentation

## Overview

The Patterned Load Generator provides sophisticated load testing capabilities with various patterns including sine waves, bursts, ramps, and custom patterns. It includes safety guardrails, real-time metrics, and profile management for reproducible testing scenarios.

## Core Components

### LoadGenerator

The main load generation engine that orchestrates pattern execution.

```go
generator := NewLoadGenerator(config, redisClient, logger)
```

#### Configuration

```go
config := &GeneratorConfig{
    DefaultGuardrails: Guardrails{
        MaxRate:           1000,              // Max jobs per second
        MaxTotal:          1000000,           // Max total jobs
        MaxDuration:       1 * time.Hour,     // Max run duration
        MaxQueueDepth:     10000,             // Stop if queue exceeds
        EmergencyStopFile: "/tmp/stop",       // Emergency stop trigger
        RateLimitWindow:   1 * time.Second,   // Rate limit window
    },
    MetricsInterval:   1 * time.Second,       // Metrics collection rate
    ProfilesPath:      "./profiles",          // Profile storage directory
    EnableCharts:      true,                  // Enable chart data
    ChartUpdateRate:   500 * time.Millisecond,// Chart update frequency
    MaxHistoryPoints:  1000,                  // Max metrics history
}
```

### Load Patterns

#### Sine Wave Pattern

Generates load following a sine wave pattern.

```go
pattern := LoadPattern{
    Type:     PatternSine,
    Duration: 60 * time.Second,
    Parameters: map[string]interface{}{
        "amplitude": 50.0,           // Peak variation from baseline
        "baseline":  100.0,          // Center line (jobs/sec)
        "period":    10 * time.Second, // Wave period
        "phase":     0.0,            // Phase shift (radians)
    },
}
```

#### Burst Pattern

Generates periodic bursts of high load.

```go
pattern := LoadPattern{
    Type:     PatternBurst,
    Duration: 120 * time.Second,
    Parameters: map[string]interface{}{
        "burst_rate":     500.0,             // Jobs/sec during burst
        "burst_duration": 5 * time.Second,   // Burst length
        "idle_duration":  10 * time.Second,  // Time between bursts
        "burst_count":    10,                // Number of bursts (0=infinite)
    },
}
```

#### Ramp Pattern

Gradually increases/decreases load.

```go
pattern := LoadPattern{
    Type:     PatternRamp,
    Duration: 300 * time.Second,
    Parameters: map[string]interface{}{
        "start_rate":    10.0,               // Initial rate
        "end_rate":      500.0,              // Final rate
        "ramp_duration": 60 * time.Second,   // Ramp time
        "hold_duration": 180 * time.Second,  // Hold at end rate
        "ramp_down":     true,               // Ramp back down
    },
}
```

#### Step Pattern

Steps through different load levels.

```go
pattern := LoadPattern{
    Type:     PatternStep,
    Duration: 180 * time.Second,
    Parameters: map[string]interface{}{
        "steps": []StepLevel{
            {Rate: 10, Duration: 30 * time.Second},
            {Rate: 50, Duration: 30 * time.Second},
            {Rate: 100, Duration: 30 * time.Second},
            {Rate: 200, Duration: 30 * time.Second},
        },
        "step_duration": 30 * time.Second, // Default duration
        "repeat":        true,              // Loop pattern
    },
}
```

#### Custom Pattern

Define arbitrary load patterns with data points.

```go
pattern := LoadPattern{
    Type:     PatternCustom,
    Duration: 120 * time.Second,
    Parameters: map[string]interface{}{
        "points": []DataPoint{
            {Time: 0, Rate: 10},
            {Time: 30 * time.Second, Rate: 100},
            {Time: 60 * time.Second, Rate: 50},
            {Time: 90 * time.Second, Rate: 200},
            {Time: 120 * time.Second, Rate: 10},
        },
        "loop": false, // Whether to loop the pattern
    },
}
```

### Guardrails

Safety limits to prevent runaway load generation.

```go
guardrails := Guardrails{
    MaxRate:           1000,              // Max jobs per second
    MaxTotal:          100000,            // Max total jobs
    MaxDuration:       30 * time.Minute,  // Max run duration
    MaxQueueDepth:     5000,              // Stop if queue exceeds
    EmergencyStopFile: "/tmp/emergency",  // Touch file to stop
    RateLimitWindow:   1 * time.Second,   // Rate limit window
}
```

## API Methods

### Start/Stop Operations

#### Start with Profile

```go
err := generator.Start("profile-id-123")
```

#### Start with Pattern

```go
pattern := &LoadPattern{
    Type:     PatternSine,
    Duration: 60 * time.Second,
    Parameters: /* ... */,
}

err := generator.StartPattern(pattern, &guardrails)
```

#### Stop Generation

```go
err := generator.Stop()
```

#### Pause/Resume

```go
// Pause generation
err := generator.Pause()

// Resume generation
err := generator.Resume()
```

### Status and Metrics

#### Get Current Status

```go
status := generator.GetStatus()

fmt.Printf("Running: %v\n", status.Running)
fmt.Printf("Pattern: %s\n", status.Pattern)
fmt.Printf("Jobs Generated: %d\n", status.JobsGenerated)
fmt.Printf("Current Rate: %.2f jobs/sec\n", status.CurrentRate)
fmt.Printf("Target Rate: %.2f jobs/sec\n", status.TargetRate)
fmt.Printf("Errors: %d\n", status.Errors)
```

#### Get Metrics History

```go
// Get last 5 minutes of metrics
metrics := generator.GetMetrics(5 * time.Minute)

for _, m := range metrics {
    fmt.Printf("%s: Target=%.2f, Actual=%.2f, Queue=%d\n",
        m.Timestamp, m.TargetRate, m.ActualRate, m.QueueDepth)
}
```

#### Get Chart Data

```go
chartData := generator.GetChartData()

// Use for visualization
for i, t := range chartData.TimePoints {
    fmt.Printf("%s: Target=%.2f, Actual=%.2f\n",
        t, chartData.TargetRates[i], chartData.ActualRates[i])
}
```

### Profile Management

#### Save Profile

```go
profile := &LoadProfile{
    Name:        "Peak Hour Test",
    Description: "Simulates peak hour traffic patterns",
    Patterns: []LoadPattern{
        // Gradual ramp up
        {
            Type:     PatternRamp,
            Duration: 5 * time.Minute,
            Parameters: /* ... */,
        },
        // Sustained high load with variations
        {
            Type:     PatternSine,
            Duration: 20 * time.Minute,
            Parameters: /* ... */,
        },
        // Gradual ramp down
        {
            Type:     PatternRamp,
            Duration: 5 * time.Minute,
            Parameters: /* ... */,
        },
    },
    Guardrails: guardrails,
    QueueName:  "production",
    Tags:       []string{"peak", "stress-test"},
}

err := generator.SaveProfile(profile)
fmt.Printf("Profile saved with ID: %s\n", profile.ID)
```

#### Load Profile

```go
profile, err := generator.LoadProfile("profile-id-123")
```

#### List Profiles

```go
profiles, err := generator.ListProfiles()

for _, p := range profiles {
    fmt.Printf("%s: %s (%d patterns)\n",
        p.ID, p.Name, len(p.Patterns))
}
```

#### Delete Profile

```go
err := generator.DeleteProfile("profile-id-123")
```

### Event Monitoring

```go
eventCh := generator.GetEventChannel()

go func() {
    for event := range eventCh {
        switch event.Type {
        case EventStarted:
            fmt.Printf("Started: %s\n", event.Message)
        case EventStopped:
            fmt.Printf("Stopped: %s\n", event.Message)
        case EventGuardrailHit:
            fmt.Printf("Guardrail: %s\n", event.Message)
        case EventError:
            fmt.Printf("Error: %s\n", event.Message)
        case EventMetricsUpdate:
            // Handle metrics update
        }
    }
}()
```

## Usage Examples

### Basic Load Test

```go
// Simple constant load test
pattern := &LoadPattern{
    Type:     PatternConstant,
    Duration: 5 * time.Minute,
    Parameters: map[string]interface{}{
        "rate": 100.0, // 100 jobs/sec
    },
}

err := generator.StartPattern(pattern, nil)
if err != nil {
    log.Fatal(err)
}

// Monitor progress
ticker := time.NewTicker(10 * time.Second)
for range ticker.C {
    status := generator.GetStatus()
    fmt.Printf("Generated: %d jobs, Rate: %.2f/sec\n",
        status.JobsGenerated, status.CurrentRate)

    if !status.Running {
        break
    }
}
```

### Stress Test with Increasing Load

```go
profile := &LoadProfile{
    Name: "Stress Test",
    Patterns: []LoadPattern{
        // Start gentle
        {
            Type:     PatternConstant,
            Duration: 1 * time.Minute,
            Parameters: map[string]interface{}{"rate": 10.0},
        },
        // Ramp up
        {
            Type:     PatternRamp,
            Duration: 5 * time.Minute,
            Parameters: map[string]interface{}{
                "start_rate": 10.0,
                "end_rate":   1000.0,
                "ramp_duration": 5 * time.Minute,
            },
        },
        // Sustain peak
        {
            Type:     PatternConstant,
            Duration: 10 * time.Minute,
            Parameters: map[string]interface{}{"rate": 1000.0},
        },
    },
    Guardrails: Guardrails{
        MaxRate:       1500,
        MaxQueueDepth: 10000,
    },
}

generator.SaveProfile(profile)
generator.Start(profile.ID)
```

### Realistic Traffic Simulation

```go
// Simulate daily traffic pattern
dailyPattern := &LoadProfile{
    Name: "Daily Traffic Pattern",
    Patterns: []LoadPattern{
        // Night (low traffic)
        {
            Type:     PatternConstant,
            Duration: 6 * time.Hour,
            Parameters: map[string]interface{}{"rate": 5.0},
        },
        // Morning ramp up
        {
            Type:     PatternRamp,
            Duration: 2 * time.Hour,
            Parameters: map[string]interface{}{
                "start_rate": 5.0,
                "end_rate":   100.0,
            },
        },
        // Day with variations
        {
            Type:     PatternSine,
            Duration: 8 * time.Hour,
            Parameters: map[string]interface{}{
                "baseline":  100.0,
                "amplitude": 30.0,
                "period":    1 * time.Hour,
            },
        },
        // Evening peak
        {
            Type:     PatternBurst,
            Duration: 2 * time.Hour,
            Parameters: map[string]interface{}{
                "burst_rate":     200.0,
                "burst_duration": 15 * time.Minute,
                "idle_duration":  15 * time.Minute,
            },
        },
        // Night ramp down
        {
            Type:     PatternRamp,
            Duration: 6 * time.Hour,
            Parameters: map[string]interface{}{
                "start_rate": 100.0,
                "end_rate":   5.0,
            },
        },
    },
}
```

### Chaos Testing

```go
// Unpredictable load patterns
chaosPattern := &LoadPattern{
    Type:     PatternCustom,
    Duration: 30 * time.Minute,
    Parameters: map[string]interface{}{
        "points": []DataPoint{
            {Time: 0, Rate: 10},
            {Time: 2 * time.Minute, Rate: 500},
            {Time: 3 * time.Minute, Rate: 10},
            {Time: 5 * time.Minute, Rate: 1000},
            {Time: 7 * time.Minute, Rate: 50},
            {Time: 10 * time.Minute, Rate: 800},
            {Time: 12 * time.Minute, Rate: 0},
            {Time: 15 * time.Minute, Rate: 1500},
        },
        "loop": true,
    },
}

// With aggressive guardrails
guardrails := Guardrails{
    MaxRate:       2000,
    MaxQueueDepth: 5000,
    EmergencyStopFile: "/tmp/chaos-stop",
}

generator.StartPattern(chaosPattern, &guardrails)
```

## Performance Considerations

### Rate Limiting

The generator uses a token bucket algorithm for precise rate limiting:
- Tokens refill at the specified rate
- Burst capacity equals rate × window
- Smooth traffic generation without spikes

### Resource Usage

- **CPU**: Minimal overhead (~1-2% for 1000 jobs/sec)
- **Memory**: ~10MB base + metrics history
- **Network**: Depends on Redis latency

### Optimization Tips

1. **Batch Job Generation**: For rates > 1000/sec, consider batching
2. **Metrics Interval**: Increase interval for lower overhead
3. **Chart Data**: Disable if not needed for visualization
4. **Profile Storage**: Use SSD for profile I/O

## Best Practices

### Pattern Design

1. **Start Gradual**: Always ramp up load gradually
2. **Include Recovery**: Test system recovery with idle periods
3. **Vary Patterns**: Mix patterns for realistic scenarios
4. **Set Guardrails**: Always define safety limits

### Testing Strategy

1. **Baseline First**: Establish baseline with constant load
2. **Incremental Testing**: Gradually increase complexity
3. **Monitor Everything**: Track both generator and system metrics
4. **Document Results**: Save profiles with test results

### Production Safety

1. **Emergency Stop**: Always configure emergency stop file
2. **Queue Monitoring**: Set appropriate MaxQueueDepth
3. **Rate Limits**: Start conservative, increase gradually
4. **Isolation**: Test in isolated environments first

## Troubleshooting

### Common Issues

#### Jobs Not Generated

```go
// Check status
status := generator.GetStatus()
if status.Errors > 0 {
    fmt.Printf("Errors: %d, Last: %s\n",
        status.Errors, status.LastError)
}

// Check guardrails
if !status.Running {
    // May have hit guardrail
    events := generator.GetEventChannel()
    // Check for EventGuardrailHit
}
```

#### Rate Mismatch

```go
// Compare target vs actual
metrics := generator.GetMetrics(1 * time.Minute)
for _, m := range metrics {
    deviation := math.Abs(m.TargetRate - m.ActualRate)
    if deviation > m.TargetRate * 0.1 { // >10% deviation
        fmt.Printf("Rate mismatch at %s: Target=%.2f, Actual=%.2f\n",
            m.Timestamp, m.TargetRate, m.ActualRate)
    }
}
```

#### Memory Growth

```go
// Limit metrics history
config.MaxHistoryPoints = 100 // Reduce from default 1000

// Disable charts if not needed
config.EnableCharts = false
```

## Advanced Features

### Custom Job Generators

```go
type CustomJobGenerator struct {
    userIDs []string
    actions []string
}

func (g *CustomJobGenerator) GenerateJob() (interface{}, error) {
    return map[string]interface{}{
        "user_id": g.userIDs[rand.Intn(len(g.userIDs))],
        "action":  g.actions[rand.Intn(len(g.actions))],
        "timestamp": time.Now().Unix(),
        "metadata": generateMetadata(),
    }, nil
}

// Use custom generator
generator.jobGenerator = &CustomJobGenerator{
    userIDs: []string{"user1", "user2", "user3"},
    actions: []string{"login", "purchase", "browse"},
}
```

### Multi-Queue Testing

```go
// Test multiple queues with different patterns
queues := []string{"high-priority", "normal", "batch"}

for i, queue := range queues {
    go func(q string, delay time.Duration) {
        time.Sleep(delay) // Stagger starts

        gen := NewLoadGenerator(config, redis, logger)
        pattern := &LoadPattern{
            Type:     PatternSine,
            Duration: 10 * time.Minute,
            Parameters: map[string]interface{}{
                "baseline":  float64(100 * (i + 1)),
                "amplitude": float64(20 * (i + 1)),
            },
        }

        profile := &LoadProfile{
            Patterns:  []LoadPattern{*pattern},
            QueueName: q,
        }

        gen.SaveProfile(profile)
        gen.Start(profile.ID)
    }(queue, time.Duration(i) * 10 * time.Second)
}
```

### Coordinated Load Testing

```go
// Coordinate multiple generators
type TestCoordinator struct {
    generators []*LoadGenerator
}

func (tc *TestCoordinator) StartWave() {
    for i, gen := range tc.generators {
        // Stagger with increasing delay
        time.AfterFunc(time.Duration(i)*5*time.Second, func() {
            gen.Start("wave-profile")
        })
    }
}

func (tc *TestCoordinator) Emergency stop() {
    for _, gen := range tc.generators {
        gen.Stop()
    }
}
```