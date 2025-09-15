# Storage Backends API Documentation

## Overview

The Storage Backends module provides a pluggable architecture for queue storage, enabling support for multiple Redis-compatible backends including Redis Lists, Redis Streams, KeyDB, and Dragonfly. This abstraction allows operators to choose optimal storage strategies based on their specific requirements.

## Core Interface

### QueueBackend Interface

All storage backends implement the `QueueBackend` interface:

```go
type QueueBackend interface {
    // Core operations
    Enqueue(ctx context.Context, job *Job) error
    Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error)
    Ack(ctx context.Context, jobID string) error
    Nack(ctx context.Context, jobID string, requeue bool) error

    // Inspection operations
    Length(ctx context.Context) (int64, error)
    Peek(ctx context.Context, offset int64) (*Job, error)

    // DLQ management
    Move(ctx context.Context, jobID string, targetQueue string) error

    // Advanced operations
    Iter(ctx context.Context, opts IterOptions) (Iterator, error)

    // Metadata and management
    Capabilities() BackendCapabilities
    Stats(ctx context.Context) (*BackendStats, error)
    Health(ctx context.Context) HealthStatus
    Close() error
}
```

### Backend Capabilities

Each backend exposes its capabilities through the `BackendCapabilities` structure:

```go
type BackendCapabilities struct {
    AtomicAck          bool // Guaranteed single processing
    ConsumerGroups     bool // Multiple consumer support
    Replay             bool // Historical job access
    IdempotentEnqueue  bool // Duplicate detection
    Transactions       bool // Multi-operation atomicity
    Persistence        bool // Survives restarts
    Clustering         bool // Distributed operation
    TimeToLive         bool // Automatic expiration
    Prioritization     bool // Priority queues
    BatchOperations    bool // Bulk enqueue/dequeue
}
```

## Available Backends

### Redis Lists Backend

The default backend using Redis Lists for backward compatibility.

**Capabilities:**
- ✅ Transactions (via Lua scripts)
- ✅ Persistence
- ✅ Clustering (with key tagging)
- ✅ Batch operations
- ❌ Atomic acknowledgment
- ❌ Consumer groups
- ❌ Replay

**Configuration:**
```go
config := RedisListsConfig{
    URL:              "redis://localhost:6379/0",
    Database:         0,
    Password:         "",
    KeyPrefix:        "queue:",
    MaxConnections:   10,
    ConnTimeout:      30 * time.Second,
    ReadTimeout:      1 * time.Second,
    WriteTimeout:     1 * time.Second,
    PoolTimeout:      4 * time.Second,
    IdleTimeout:      5 * time.Minute,
    MaxRetries:       3,
    ClusterMode:      false,
    ClusterAddrs:     []string{},
    TLS:              false,
}
```

**Use Cases:**
- Simple job queues
- Backward compatibility
- Environments requiring minimal Redis features

### Redis Streams Backend

Advanced backend using Redis Streams for consumer groups and replay capabilities.

**Capabilities:**
- ✅ Atomic acknowledgment (XACK)
- ✅ Consumer groups (XGROUP)
- ✅ Replay (historical XREAD)
- ✅ Transactions
- ✅ Persistence
- ✅ Clustering
- ✅ Batch operations
- ❌ Idempotent enqueue (application level)
- ❌ TTL
- ❌ Prioritization

**Configuration:**
```go
config := RedisStreamsConfig{
    URL:              "redis://localhost:6379/0",
    Database:         0,
    Password:         "",
    StreamName:       "job-stream",
    ConsumerGroup:    "workers",
    ConsumerName:     "worker-1",
    MaxLength:        10000,
    BlockTimeout:     1 * time.Second,
    ClaimMinIdle:     30 * time.Second,
    ClaimCount:       100,
    MaxConnections:   10,
    ConnTimeout:      30 * time.Second,
    ReadTimeout:      1 * time.Second,
    WriteTimeout:     1 * time.Second,
    PoolTimeout:      4 * time.Second,
    IdleTimeout:      5 * time.Minute,
    MaxRetries:       3,
    ClusterMode:      false,
    ClusterAddrs:     []string{},
    TLS:              false,
}
```

**Use Cases:**
- Analytics and audit queues requiring replay
- Multi-consumer scenarios
- High-reliability job processing

## Backend Management

### BackendRegistry

The registry manages available backend types:

```go
// Get default registry
registry := storage.DefaultRegistry()

// Register custom backend
registry.Register("my-backend", &MyBackendFactory{})

// Create backend instance
backend, err := registry.Create("redis-lists", config)

// Validate configuration
err := registry.Validate("redis-streams", config)
```

### BackendManager

The manager orchestrates multiple backends across queues:

```go
// Create manager
manager := storage.NewBackendManager(registry)

// Add backend for specific queue
config := storage.BackendConfig{
    Type: "redis-lists",
    Name: "default-backend",
    URL:  "redis://localhost:6379/0",
}
err := manager.AddBackend("my-queue", config)

// Get backend for queue
backend, err := manager.GetBackend("my-queue")

// Health check all backends
health := manager.HealthCheck(ctx)

// Get stats for all backends
stats, err := manager.Stats(ctx)
```

## Configuration

### Complete Configuration Example

```yaml
# config/storage.yaml
backends:
  redis-default:
    type: redis-lists
    url: redis://localhost:6379/0
    options:
      key_prefix: "queue:"
      max_connections: 10

  redis-streams:
    type: redis-streams
    url: redis://localhost:6379/1
    options:
      stream_name: "analytics-stream"
      consumer_group: "processors"
      consumer_name: "worker-1"
      max_length: 10000

  keydb-cluster:
    type: keydb
    cluster_mode: true
    cluster_addrs:
      - keydb1:6379
      - keydb2:6379
      - keydb3:6379
    options:
      pipeline_size: 1000
      max_connections: 50

queues:
  default:
    backend: redis-default

  analytics:
    backend: redis-streams

  bulk-processing:
    backend: keydb-cluster

defaults:
  backend: redis-lists
  max_retries: 3
  timeout: 30s
  batch_size: 100
  pool_size: 10
  idle_timeout: 5m
  read_timeout: 1s
  write_timeout: 1s
```

### Loading Configuration

```go
// Load from file
config, err := storage.LoadConfigFromFile("config/storage.yaml")
if err != nil {
    log.Fatal(err)
}

// Apply defaults
config.ApplyDefaults()

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal(err)
}

// Create backend manager
registry := storage.DefaultRegistry()
manager := storage.NewBackendManager(registry)

// Configure backends
for queueName, queueConfig := range config.Queues {
    backendConfig, err := config.GetBackendConfig(queueConfig.Backend)
    if err != nil {
        log.Fatal(err)
    }

    if err := manager.AddBackend(queueName, backendConfig); err != nil {
        log.Fatal(err)
    }
}
```

## Migration

### MigrationManager

Migrate jobs between backends safely:

```go
// Create migration manager
migrationManager := storage.NewMigrationManager(backendManager)

// Start migration
opts := storage.MigrationOptions{
    SourceBackend: "redis-lists",
    TargetBackend: "redis-streams",
    DrainFirst:    true,
    Timeout:       30 * time.Minute,
    BatchSize:     100,
    VerifyData:    true,
    DryRun:        false,
}

status, err := migrationManager.StartMigration(ctx, "my-queue", opts)
if err != nil {
    log.Fatal(err)
}

// Monitor migration progress
for {
    status, err := migrationManager.GetMigrationStatus("my-queue")
    if err != nil {
        break
    }

    fmt.Printf("Migration progress: %.1f%% (%d/%d jobs)\n",
        status.Progress, status.MigratedJobs, status.TotalJobs)

    if status.Phase == storage.MigrationPhaseCompleted {
        break
    }

    time.Sleep(1 * time.Second)
}
```

### Migration Tool

High-level migration utilities:

```go
// Create migration tool
tool := storage.NewMigrationTool(backendManager)

// Plan migration
plan, err := tool.PlanMigration(ctx, "my-queue", opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Migration plan:\n")
fmt.Printf("  Queue: %s\n", plan.QueueName)
fmt.Printf("  Jobs: %d\n", plan.JobCount)
fmt.Printf("  Estimated duration: %v\n", plan.EstimatedDuration)
fmt.Printf("  Batch size: %d\n", plan.BatchSize)

for _, warning := range plan.Warnings {
    fmt.Printf("  WARNING: %s\n", warning)
}

for _, rec := range plan.Recommendations {
    fmt.Printf("  RECOMMENDATION: %s\n", rec)
}

// Execute migration
status, err := tool.ExecuteMigration(ctx, "my-queue", opts)
if err != nil {
    log.Fatal(err)
}

// Quick migration with defaults
status, err = tool.QuickMigrate(ctx, "my-queue", "redis-streams")
```

## Error Handling

### Error Types

The module provides structured error types:

```go
// Backend-specific errors
var backendErr *storage.BackendError
if errors.As(err, &backendErr) {
    fmt.Printf("Backend %s operation %s failed: %v\n",
        backendErr.Backend, backendErr.Operation, backendErr.Err)
}

// Configuration errors
var configErr *storage.ConfigurationError
if errors.As(err, &configErr) {
    fmt.Printf("Configuration error in field %s: %s\n",
        configErr.Field, configErr.Message)
}

// Migration errors
var migrationErr *storage.MigrationError
if errors.As(err, &migrationErr) {
    fmt.Printf("Migration error in phase %s: %s\n",
        migrationErr.Phase, migrationErr.Message)
}
```

### Error Classification

```go
// Check if error is retryable
if storage.IsRetryable(err) {
    // Retry operation
}

// Check if error is permanent
if storage.IsPermanent(err) {
    // Don't retry, handle failure
}

// Get stable error code
code := storage.ErrorCode(err)
switch code {
case "BACKEND_NOT_FOUND":
    // Handle missing backend
case "CONNECTION_FAILED":
    // Handle connection issues
case "MIGRATION_FAILED":
    // Handle migration failure
}
```

## Monitoring and Observability

### Health Checks

```go
// Check individual backend health
health := backend.Health(ctx)
fmt.Printf("Backend status: %s\n", health.Status)
if health.Message != "" {
    fmt.Printf("Message: %s\n", health.Message)
}

// Check all backends
healthMap := manager.HealthCheck(ctx)
for queue, health := range healthMap {
    fmt.Printf("Queue %s: %s\n", queue, health.Status)
}
```

### Statistics

```go
// Get backend statistics
stats, err := backend.Stats(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Queue depth: %d\n", stats.QueueDepth)
fmt.Printf("Enqueue rate: %.2f/sec\n", stats.EnqueueRate)
fmt.Printf("Dequeue rate: %.2f/sec\n", stats.DequeueRate)
fmt.Printf("Error rate: %.2f%%\n", stats.ErrorRate*100)

// Streams-specific metrics
if stats.StreamLength != nil {
    fmt.Printf("Stream length: %d\n", *stats.StreamLength)
}
if stats.ConsumerLag != nil {
    fmt.Printf("Consumer lag: %d\n", *stats.ConsumerLag)
}

// Connection pool stats
if stats.ConnectionPool != nil {
    fmt.Printf("Pool - Active: %d, Idle: %d\n",
        stats.ConnectionPool.Active, stats.ConnectionPool.Idle)
}
```

## Best Practices

### Backend Selection

1. **Redis Lists**: Use for simple queues and backward compatibility
2. **Redis Streams**: Use for analytics, audit trails, and multi-consumer scenarios
3. **KeyDB**: Use for high-throughput scenarios requiring performance
4. **Dragonfly**: Use for memory-efficient, high-performance deployments

### Configuration Guidelines

1. **Connection Pooling**: Set appropriate pool sizes based on concurrency
2. **Timeouts**: Configure timeouts based on network latency and SLAs
3. **Clustering**: Use key tagging for Redis Cluster deployments
4. **Monitoring**: Enable comprehensive health checks and metrics

### Migration Best Practices

1. **Plan First**: Always run migration planning before execution
2. **Drain Source**: For large migrations, consider draining the source first
3. **Batch Size**: Tune batch size based on job size and network capacity
4. **Verification**: Always enable data verification for critical migrations
5. **Rollback**: Have rollback procedures ready before starting migration

### Performance Optimization

1. **Pipeline Operations**: Use batch operations when supported
2. **Connection Reuse**: Configure appropriate connection pooling
3. **Compression**: Enable compression for large payloads (KeyDB/Dragonfly)
4. **Monitoring**: Track key performance metrics continuously

## Troubleshooting

### Common Issues

1. **Connection Failures**: Check Redis connectivity and authentication
2. **High Error Rates**: Monitor backend health and resource usage
3. **Migration Stalls**: Check source/target backend health and network
4. **Performance Issues**: Review connection pool settings and batch sizes

### Debugging

```go
// Enable debug logging
import "log"

// Check backend capabilities
caps := backend.Capabilities()
log.Printf("Backend capabilities: %+v", caps)

// Monitor health continuously
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        health := backend.Health(ctx)
        if health.Status != storage.HealthStatusHealthy {
            log.Printf("Backend unhealthy: %s", health.Message)
        }
    }
}()

// Track migration progress
go func() {
    for {
        status, err := migrationManager.GetMigrationStatus("queue")
        if err != nil {
            break
        }
        log.Printf("Migration: %s %.1f%%", status.Phase, status.Progress)
        time.Sleep(5 * time.Second)
    }
}()
```