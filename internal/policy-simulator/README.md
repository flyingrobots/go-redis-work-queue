# Policy Simulator API Documentation

## Overview

The Policy Simulator provides a "what-if" analysis system for testing queue policy changes before applying them to production. It uses queueing theory models to predict the impact of configuration changes on queue performance.

## Core Components

### PolicySimulator

The main simulator service that provides simulation and policy change management capabilities.

#### Configuration

```go
type SimulatorConfig struct {
    SimulationDuration time.Duration `json:"simulation_duration"` // How long to simulate
    TimeStep          time.Duration `json:"time_step"`           // Granularity of simulation
    MaxWorkers        int           `json:"max_workers"`         // Maximum concurrent workers
    RedisPoolSize     int           `json:"redis_pool_size"`     // Redis connection pool size
}
```

#### Policy Configuration

```go
type PolicyConfig struct {
    // Retry policies
    MaxRetries      int           `json:"max_retries"`
    InitialBackoff  time.Duration `json:"initial_backoff"`
    MaxBackoff      time.Duration `json:"max_backoff"`
    BackoffStrategy string        `json:"backoff_strategy"` // exponential, linear, constant

    // Rate limiting
    MaxRatePerSecond float64 `json:"max_rate_per_second"`
    BurstSize        int     `json:"burst_size"`

    // Concurrency controls
    MaxConcurrency int `json:"max_concurrency"`
    QueueSize      int `json:"queue_size"`

    // Timeout settings
    ProcessingTimeout time.Duration `json:"processing_timeout"`
    AckTimeout        time.Duration `json:"ack_timeout"`

    // Dead letter queue
    DLQEnabled    bool   `json:"dlq_enabled"`
    DLQThreshold  int    `json:"dlq_threshold"`
    DLQQueueName  string `json:"dlq_queue_name"`
}
```

## REST API Endpoints

### Simulations

#### Create Simulation
- **POST** `/api/policy-simulator/simulations`
- **Description**: Runs a new simulation with specified policies and traffic patterns
- **Request Body**:
```json
{
    "name": "string",
    "description": "string",
    "policies": PolicyConfig,
    "traffic_pattern": TrafficPattern,
    "config": SimulatorConfig (optional)
}
```
- **Response**: `SimulationResult` with status 201

#### List Simulations
- **GET** `/api/policy-simulator/simulations`
- **Query Parameters**:
  - `limit`: Maximum number of results
  - `status`: Filter by simulation status
- **Response**: Array of `SimulationResult`

#### Get Simulation
- **GET** `/api/policy-simulator/simulations/{id}`
- **Response**: `SimulationResult`

#### Get Simulation Charts
- **GET** `/api/policy-simulator/simulations/{id}/charts`
- **Response**: Chart data for visualization
- **Requirements**: Simulation must be completed

### Policy Changes

#### Create Policy Change
- **POST** `/api/policy-simulator/changes`
- **Request Body**:
```json
{
    "description": "string",
    "changes": {
        "field_name": "new_value"
    }
}
```
- **Headers**: `X-User-ID` (optional, defaults to "anonymous")
- **Response**: `PolicyChange`

#### Apply Policy Change
- **POST** `/api/policy-simulator/changes/{id}/apply`
- **Request Body**:
```json
{
    "reason": "string (optional)"
}
```
- **Headers**: `X-User-ID` required
- **Response**: Success message with application details

#### Rollback Policy Change
- **POST** `/api/policy-simulator/changes/{id}/rollback`
- **Request Body**:
```json
{
    "reason": "string (optional)"
}
```
- **Headers**: `X-User-ID` required
- **Response**: Success message with rollback details

### Presets

#### Get Policy Presets
- **GET** `/api/policy-simulator/presets/policies`
- **Response**: Predefined policy configurations (conservative, aggressive, balanced)

#### Get Traffic Presets
- **GET** `/api/policy-simulator/presets/traffic`
- **Response**: Predefined traffic patterns (steady, spike, seasonal, bursty)

## Traffic Patterns

### Types
- **constant**: Steady state load
- **linear**: Linear increase/decrease
- **spike**: Sudden burst
- **seasonal**: Periodic patterns
- **bursty**: Random bursts
- **exponential**: Exponential growth/decay

### TrafficPattern Structure
```go
type TrafficPattern struct {
    Name        string                `json:"name"`
    Type        TrafficPatternType    `json:"type"`
    BaseRate    float64               `json:"base_rate"`    // Messages per second baseline
    Variations  []TrafficVariation    `json:"variations"`   // Spikes, drops, seasonal patterns
    Duration    time.Duration         `json:"duration"`     // How long this pattern lasts
    Probability float64               `json:"probability"`  // Likelihood of this pattern occurring
}
```

## Simulation Results

### SimulationResult
```go
type SimulationResult struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Config      *SimulatorConfig    `json:"config"`
    Policies    *PolicyConfig       `json:"policies"`
    Pattern     *TrafficPattern     `json:"pattern"`
    Metrics     *SimulationMetrics  `json:"metrics"`
    Timeline    []TimelineSnapshot  `json:"timeline"`
    Warnings    []string            `json:"warnings"`
    CreatedAt   time.Time           `json:"created_at"`
    Status      SimulationStatus    `json:"status"`
}
```

### Key Metrics
- **Queue Depth**: Average and maximum queue depth
- **Wait Times**: Average, P95, and P99 wait times
- **Throughput**: Messages per second processing rate
- **Utilization**: Worker utilization percentage
- **Error Rates**: Failure, retry, and DLQ rates
- **Resource Usage**: Memory and CPU estimates

## Interactive UI

The system includes a terminal-based UI for interactive policy configuration and simulation:

```bash
go run cmd/policy-simulator/main.go
```

### UI Features
- **Policy Tab**: Configure retry policies, rate limits, concurrency
- **Traffic Tab**: Define traffic patterns and variations
- **Simulation Tab**: Run simulations and view progress
- **Results Tab**: View detailed metrics and warnings
- **Charts Tab**: Visualize queue depth, processing rates, and resource usage

### UI Controls
- `←→` or `tab`: Navigate between tabs
- `↑↓`: Navigate between fields
- `enter`: Confirm selections
- `s`: Run simulation (when on Simulation tab)
- `r`: Reset to defaults
- `q`: Quit application

## Queueing Models

The simulator supports multiple queueing theory models:

### M/M/1 Model
- Markovian arrivals and service times
- Single server
- Suitable for simple scenarios

### M/M/c Model
- Markovian arrivals and service times
- Multiple servers (c workers)
- Good for multi-worker scenarios

### Simplified Model
- Based on Little's Law
- Faster computation
- Good for quick estimates

## Error Handling

### Error Types
- `INVALID_CONFIG`: Configuration validation errors
- `INVALID_POLICY`: Policy configuration errors
- `INVALID_TRAFFIC_PATTERN`: Traffic pattern errors
- `SIMULATION_FAILED`: Runtime simulation errors
- `CHANGE_NOT_FOUND`: Policy change not found
- `APPLY_FAILED`: Failed to apply policy change
- `ROLLBACK_FAILED`: Failed to rollback policy change

### Response Format
```json
{
    "error": "Error message",
    "status": 400,
    "timestamp": "2025-01-14T...",
    "details": "Detailed error information"
}
```

## Limitations and Assumptions

### Model Limitations
- Does not account for cold starts or warmup time
- Simplified failure model
- No modeling of Redis network latency
- Worker startup/shutdown time not modeled
- Memory and CPU estimates are approximations

### Accuracy
- ±20% for steady-state metrics under normal conditions
- Best for relative comparison of policies
- Not recommended for precise absolute predictions

### Recommended Use Cases
- Steady-state analysis
- Relative comparison of policies
- Capacity planning
- Performance optimization

### Not Recommended For
- Precise absolute predictions
- Modeling of rare failure modes
- Cold start performance analysis
- Network partition scenarios

## Health Check

- **GET** `/health`
- **Response**: Service health status

```json
{
    "status": "healthy",
    "service": "policy-simulator",
    "timestamp": "2025-01-14T...",
    "version": "1.0.0"
}
```

## Example Usage

### Basic Simulation
```bash
curl -X POST http://localhost:8080/api/policy-simulator/simulations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Simulation",
    "description": "Testing new retry policy",
    "policies": {
      "max_retries": 5,
      "initial_backoff": "1s",
      "max_backoff": "30s",
      "backoff_strategy": "exponential",
      "max_rate_per_second": 100,
      "max_concurrency": 10
    },
    "traffic_pattern": {
      "name": "Steady Load",
      "type": "constant",
      "base_rate": 50,
      "duration": "5m"
    }
  }'
```

### Policy Change Workflow
```bash
# Create policy change
curl -X POST http://localhost:8080/api/policy-simulator/changes \
  -H "Content-Type: application/json" \
  -H "X-User-ID: admin" \
  -d '{
    "description": "Increase retry limit",
    "changes": {
      "max_retries": 8
    }
  }'

# Apply change (requires approval in production)
curl -X POST http://localhost:8080/api/policy-simulator/changes/{id}/apply \
  -H "Content-Type: application/json" \
  -H "X-User-ID: admin" \
  -d '{"reason": "Performance improvement"}'
```