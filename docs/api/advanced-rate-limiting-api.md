# Advanced Rate Limiting API Documentation

## Overview

The Advanced Rate Limiting system provides token-bucket based rate limiting with priority fairness for the go-redis-work-queue. It supports global limits, per-tenant limits, priority-based token allocation, and starvation prevention.

## Core Components

### RateLimiter

The main rate limiting component that manages token buckets using Redis for atomic operations.

#### Configuration

```go
type Config struct {
    // Global limits
    GlobalRatePerSecond int64
    GlobalBurstSize     int64
    
    // Default per-tenant limits
    DefaultRatePerSecond int64
    DefaultBurstSize     int64
    
    // Priority weights (higher = more tokens)
    PriorityWeights map[string]float64
    
    // Refill interval
    RefillInterval time.Duration
    
    // TTL for rate limit keys
    KeyTTL time.Duration
    
    // Dry run mode
    DryRun bool
}
```

#### Methods

##### NewRateLimiter

```go
func NewRateLimiter(redis *redis.Client, logger *zap.Logger, config *Config) *RateLimiter
```

Creates a new rate limiter instance.

**Parameters:**
- `redis`: Redis client for storage
- `logger`: Zap logger for debugging
- `config`: Rate limiter configuration

**Returns:** Configured RateLimiter instance

##### Consume

```go
func (rl *RateLimiter) Consume(ctx context.Context, scope string, tokens int64, priority string) (*ConsumeResult, error)
```

Attempts to consume tokens from the rate limiter.

**Parameters:**
- `ctx`: Context for cancellation
- `scope`: Tenant or scope identifier
- `tokens`: Number of tokens to consume
- `priority`: Priority level ("critical", "high", "normal", "low")

**Returns:**
- `ConsumeResult`: Contains allowed status, remaining tokens, and retry information
- `error`: Any errors during consumption

**Example:**
```go
result, err := rl.Consume(ctx, "tenant-123", 10, "high")
if err != nil {
    return err
}
if !result.Allowed {
    time.Sleep(result.RetryAfter)
    return fmt.Errorf("rate limited")
}
```

##### Refill

```go
func (rl *RateLimiter) Refill(ctx context.Context, scope string, tokens int64) (int64, error)
```

Manually adds tokens to a bucket.

**Parameters:**
- `ctx`: Context for cancellation
- `scope`: Tenant or scope identifier
- `tokens`: Number of tokens to add

**Returns:**
- `int64`: New token count
- `error`: Any errors during refill

##### GetStatus

```go
func (rl *RateLimiter) GetStatus(ctx context.Context, scope string) (*Status, error)
```

Returns the current status of a rate limiter scope.

**Parameters:**
- `ctx`: Context for cancellation
- `scope`: Tenant or scope identifier

**Returns:**
- `Status`: Current bucket status including available tokens, capacity, and refill rate
- `error`: Any errors retrieving status

##### Reset

```go
func (rl *RateLimiter) Reset(ctx context.Context, scope string) error
```

Clears the rate limit state for a scope.

**Parameters:**
- `ctx`: Context for cancellation
- `scope`: Tenant or scope identifier

**Returns:** Error if reset fails

### PriorityFairness

Implements weighted fair queuing for rate limiting to prevent starvation.

#### Configuration

```go
type FairnessConfig struct {
    // Base weights for each priority level
    Weights map[string]float64
    
    // Starvation prevention
    MinGuaranteedShare float64       // Minimum share per priority (e.g., 0.05 = 5%)
    MaxWaitTime        time.Duration // Maximum wait before forcing allocation
    
    // Adaptive fairness
    EnableAdaptive bool
    AdaptiveWindow time.Duration
    
    // Burst allowance
    BurstMultiplier float64 // Allow bursts up to this multiple of fair share
}
```

#### Methods

##### NewPriorityFairness

```go
func NewPriorityFairness(redis *redis.Client, logger *zap.Logger, config *FairnessConfig) *PriorityFairness
```

Creates a new priority fairness scheduler.

##### AllocateTokens

```go
func (pf *PriorityFairness) AllocateTokens(ctx context.Context, availableTokens int64, demands map[string]int64) (map[string]int64, error)
```

Distributes available tokens among priorities fairly.

**Parameters:**
- `ctx`: Context for cancellation
- `availableTokens`: Total tokens available for distribution
- `demands`: Map of priority to requested tokens

**Returns:**
- `map[string]int64`: Allocated tokens per priority
- `error`: Any errors during allocation

**Algorithm:**
1. Allocates minimum guaranteed share to all priorities
2. Distributes remaining tokens by weighted fair share
3. Applies starvation prevention for priorities waiting too long

##### CheckFairness

```go
func (pf *PriorityFairness) CheckFairness(ctx context.Context, priority string, requestedTokens int64) (*FairnessDecision, error)
```

Evaluates if current consumption is fair.

**Parameters:**
- `ctx`: Context for cancellation
- `priority`: Priority level to check
- `requestedTokens`: Number of tokens requested

**Returns:**
- `FairnessDecision`: Contains allowed status, fair share info, and suggested delay
- `error`: Any errors during check

## Data Types

### ConsumeResult

```go
type ConsumeResult struct {
    Allowed          bool          // Whether request was allowed
    Tokens           int64         // Tokens consumed
    Remaining        int64         // Tokens remaining
    RetryAfter       time.Duration // Wait time if denied
    ResetAt          time.Time     // When bucket refills
    DryRunWouldAllow bool          // Result if not in dry-run mode
}
```

### Status

```go
type Status struct {
    Scope      string    // Scope identifier
    Available  int64     // Available tokens
    Capacity   int64     // Maximum capacity
    RefillRate int64     // Tokens per second
    LastRefill time.Time // Last refill time
    NextRefill time.Time // Next refill time
    Priority   string    // Priority level
    Weight     float64   // Priority weight
}
```

### FairnessDecision

```go
type FairnessDecision struct {
    Priority          string        // Priority level
    Allowed           bool          // Whether request allowed
    FairShare         int64         // Fair share allocation
    CurrentUsage      int64         // Current token usage
    BurstLimit        int64         // Burst capacity
    IsWithinFairShare bool          // Within fair allocation
    IsStarving        bool          // Priority is starving
    SuggestedDelay    time.Duration // Recommended wait time
}
```

## Usage Examples

### Basic Rate Limiting

```go
// Initialize rate limiter
config := ratelimiting.DefaultConfig()
rl := ratelimiting.NewRateLimiter(redisClient, logger, config)

// Consume tokens
result, err := rl.Consume(ctx, "api-client-1", 5, "normal")
if err != nil {
    return fmt.Errorf("rate limit error: %w", err)
}

if !result.Allowed {
    // Rate limited - wait and retry
    time.Sleep(result.RetryAfter)
    return fmt.Errorf("rate limited, retry after %v", result.RetryAfter)
}

// Process request
processRequest()
```

### Priority-Based Rate Limiting

```go
// High priority request gets more tokens
result, err := rl.Consume(ctx, "tenant-vip", 100, "critical")
if err != nil {
    return err
}

if result.Allowed {
    // Process critical request
    processCriticalRequest()
} else {
    // Even critical requests can be limited
    log.Warnf("Critical request rate limited, retry in %v", result.RetryAfter)
}
```

### Fair Token Allocation

```go
// Setup priority fairness
fairnessConfig := ratelimiting.DefaultFairnessConfig()
pf := ratelimiting.NewPriorityFairness(redisClient, logger, fairnessConfig)

// Allocate tokens fairly among priorities
demands := map[string]int64{
    "critical": 500,
    "high":     300,
    "normal":   200,
    "low":      100,
}

allocations, err := pf.AllocateTokens(ctx, 1000, demands)
if err != nil {
    return err
}

// Use allocations to process requests
for priority, tokens := range allocations {
    processRequestsForPriority(priority, tokens)
}
```

### Monitoring Rate Limits

```go
// Get current status
status, err := rl.GetStatus(ctx, "tenant-123")
if err != nil {
    return err
}

log.Infof("Tenant %s: %d/%d tokens available, refill rate: %d/sec",
    status.Scope,
    status.Available,
    status.Capacity,
    status.RefillRate)

// Check if approaching limit
if float64(status.Available)/float64(status.Capacity) < 0.2 {
    log.Warnf("Tenant %s approaching rate limit", status.Scope)
}
```

### Dry-Run Mode

```go
// Enable dry-run for testing
config := ratelimiting.DefaultConfig()
config.DryRun = true
rl := ratelimiting.NewRateLimiter(redisClient, logger, config)

// Test rate limits without enforcement
result, err := rl.Consume(ctx, "test-tenant", 1000, "normal")
if err != nil {
    return err
}

if result.Allowed {
    log.Infof("Request allowed (dry-run)")
    if !result.DryRunWouldAllow {
        log.Warnf("Would be denied in production")
    }
}
```

## Best Practices

1. **Set Appropriate Burst Sizes**: Burst size should be 2-3x the rate to handle traffic spikes
2. **Use Priority Weights**: Assign higher weights to critical services
3. **Monitor Rate Limit Metrics**: Track allowed/denied ratios to tune limits
4. **Implement Backoff**: Use exponential backoff when rate limited
5. **Configure TTLs**: Set appropriate key TTLs to prevent memory bloat
6. **Test with Dry-Run**: Always test configuration changes in dry-run mode first
7. **Prevent Starvation**: Enable minimum guaranteed share for all priorities
8. **Handle Errors Gracefully**: Implement fallback behavior when rate limiting fails

## Performance Considerations

- **Atomic Operations**: All token consumption uses Lua scripts for atomicity
- **Redis Load**: Each consume operation is a single Redis round-trip
- **Memory Usage**: O(n) where n is number of active scopes
- **Refill Overhead**: Refill calculations happen inline during consumption
- **Benchmarks**: ~10,000 ops/sec per Redis instance (single tenant)

## Error Handling

Common errors and handling:

```go
result, err := rl.Consume(ctx, scope, tokens, priority)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Timeout - fail open or closed based on policy
        return handleTimeout()
    }
    if errors.Is(err, redis.Nil) {
        // Key doesn't exist - first request
        return handleFirstRequest()
    }
    // Other error - log and fail
    return fmt.Errorf("rate limit error: %w", err)
}
```

## Migration Guide

### From Fixed Window to Token Bucket

1. Deploy with dry-run enabled
2. Monitor DryRunWouldAllow to understand impact
3. Adjust burst sizes based on observed patterns
4. Disable dry-run when confident

### Adding Priority Fairness

1. Define priority weights based on SLAs
2. Enable with conservative MinGuaranteedShare (0.05)
3. Monitor starvation metrics
4. Tune weights based on actual usage patterns
