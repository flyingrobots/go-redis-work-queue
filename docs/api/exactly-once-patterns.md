# Exactly-Once Patterns API Documentation

This document describes the exactly-once processing patterns implemented for the go-redis-work-queue system. These patterns provide idempotency guarantees and reliable event publishing through deduplication storage and the transactional outbox pattern.

## Overview

The exactly-once patterns module provides three main capabilities:

1. **Idempotency**: Ensures job processing happens exactly once by using deduplication keys
2. **Outbox Pattern**: Reliable event publishing with transactional guarantees
3. **Admin API**: Monitoring and management endpoints for operational visibility

## Configuration

### Basic Configuration

```yaml
exactly_once:
  idempotency:
    enabled: true
    default_ttl: 24h
    key_prefix: "idempotency:"
    cleanup_interval: 1h
    batch_size: 100
    storage:
      type: "redis"  # redis, memory, database
      redis:
        key_pattern: "{queue}:idempotency:{tenant}:{key}"
        hash_key_pattern: "{queue}:idempotency:{tenant}"
        use_hashes: true
        compression: false

  outbox:
    enabled: false  # Requires database setup
    storage_type: "redis"  # redis, database
    poll_interval: 5s
    batch_size: 50
    max_retries: 5
    cleanup_interval: 24h
    cleanup_after: 168h  # 7 days
    retry_backoff:
      initial_delay: 1s
      max_delay: 30m
      multiplier: 2.0
      jitter: true

  metrics:
    enabled: true
    collection_interval: 30s
    histogram_buckets: [0.001, 0.01, 0.1, 1, 10]
    cardinality_limit: 1000
```

### Storage Types

#### Redis Storage (Default)
```yaml
idempotency:
  storage:
    type: "redis"
    redis:
      key_pattern: "{queue}:idempotency:{tenant}:{key}"
      use_hashes: true  # More memory efficient
      compression: false
```

#### Memory Storage (Testing/Development)
```yaml
idempotency:
  storage:
    type: "memory"
    memory:
      max_keys: 10000
      eviction_policy: "fifo"  # fifo, lru
```

#### Database Storage
```yaml
idempotency:
  storage:
    type: "database"
    database:
      table_name: "idempotency_keys"
      batch_size: 100
      max_connections: 10
      transaction_timeout: 30s
```

## Go API Reference

### Manager

The `Manager` is the main interface for exactly-once patterns:

```go
import "github.com/flyingrobots/go-redis-work-queue/internal/exactly-once-patterns"

// Create manager with default configuration
cfg := exactlyonce.DefaultConfig()
manager := exactlyonce.NewManager(cfg, redisClient, logger)
```

#### Core Methods

##### ProcessWithIdempotency

Process a job with idempotency guarantees:

```go
func (m *Manager) ProcessWithIdempotency(
    ctx context.Context,
    key IdempotencyKey,
    processor func() (interface{}, error)
) (interface{}, error)
```

**Example:**
```go
// Generate or receive idempotency key
key := manager.GenerateIdempotencyKey("file-processing", "", jobID)

// Process with idempotency
result, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
    // Your job processing logic here
    return processFile(jobData)
})
```

##### GenerateIdempotencyKey

Create unique idempotency keys:

```go
func (m *Manager) GenerateIdempotencyKey(queueName, tenantID string, customSuffix ...string) IdempotencyKey
```

**Example:**
```go
// Basic key
key := manager.GenerateIdempotencyKey("user-queue", "tenant-123")

// With custom suffix
key := manager.GenerateIdempotencyKey("user-queue", "", "user-created", userID)
```

##### Outbox Operations

Store and publish outbox events:

```go
// Store event in outbox
event := exactlyonce.OutboxEvent{
    AggregateID: "user-123",
    EventType:   "user.created",
    Payload:     json.RawMessage(`{"user_id": "123"}`),
}

err := manager.StoreInOutbox(ctx, event)

// Publish pending events
err = manager.PublishOutboxEvents(ctx)
```

### Types

#### IdempotencyKey

```go
type IdempotencyKey struct {
    ID        string        // Unique identifier
    QueueName string        // Queue/service name
    TenantID  string        // Optional tenant isolation
    CreatedAt time.Time     // Creation timestamp
    TTL       time.Duration // Time to live
}
```

#### OutboxEvent

```go
type OutboxEvent struct {
    ID             string                 // Unique event ID
    AggregateID    string                 // Entity ID
    AggregateType  string                 // Entity type
    EventType      string                 // Event type
    Payload        json.RawMessage        // Event data
    Headers        map[string]string      // Optional headers
    Metadata       map[string]interface{} // Additional metadata
    CreatedAt      time.Time             // Creation time
    PublishedAt    *time.Time            // Publication time
    RetryCount     int                   // Current retry count
    MaxRetries     int                   // Maximum retries
    Status         string                // Event status
}
```

#### DedupStats

```go
type DedupStats struct {
    QueueName         string    // Queue name
    TenantID          string    // Tenant ID
    TotalKeys         int64     // Number of keys stored
    HitRate           float64   // Cache hit rate (0.0-1.0)
    TotalRequests     int64     // Total requests processed
    DuplicatesAvoided int64     // Number of duplicates prevented
    LastUpdated       time.Time // Last update time
}
```

### Processing Hooks

Implement custom hooks for processing lifecycle events:

```go
type LoggingHook struct {
    log *zap.Logger
}

func (h *LoggingHook) BeforeProcessing(ctx context.Context, jobID string, key IdempotencyKey) error {
    h.log.Info("Starting job", zap.String("job_id", jobID))
    return nil
}

func (h *LoggingHook) AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error {
    if err != nil {
        h.log.Error("Job failed", zap.String("job_id", jobID), zap.Error(err))
    }
    return nil
}

func (h *LoggingHook) OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error {
    h.log.Info("Duplicate job detected", zap.String("job_id", jobID))
    return nil
}

// Register the hook
manager.RegisterHook(&LoggingHook{log: logger})
```

## Admin API

The admin API provides HTTP endpoints for monitoring and management:

### Setup

```go
handler := exactlyonce.NewAdminHandler(manager, logger)
mux := http.NewServeMux()
handler.RegisterRoutes(mux)
```

### Endpoints

#### GET /api/v1/exactly-once/stats

Get deduplication statistics:

**Query Parameters:**
- `queue` (required): Queue name
- `tenant` (optional): Tenant ID

**Example:**
```bash
curl "http://localhost:8080/api/v1/exactly-once/stats?queue=file-processing&tenant=acme"
```

**Response:**
```json
{
  "queue_name": "file-processing",
  "tenant_id": "acme",
  "total_keys": 1500,
  "hit_rate": 0.85,
  "total_requests": 10000,
  "duplicates_avoided": 8500,
  "last_updated": "2023-01-15T10:30:00Z"
}
```

#### GET /api/v1/exactly-once/idempotency

Check if an idempotency key exists:

**Query Parameters:**
- `queue` (required): Queue name
- `key` (required): Idempotency key
- `tenant` (optional): Tenant ID

**Example:**
```bash
curl "http://localhost:8080/api/v1/exactly-once/idempotency?queue=orders&key=order-123"
```

#### POST /api/v1/exactly-once/idempotency

Create an idempotency key:

**Request Body:**
```json
{
  "queue_name": "orders",
  "tenant_id": "acme",
  "key_id": "order-123",
  "value": {"status": "processed"},
  "ttl": "24h"
}
```

#### DELETE /api/v1/exactly-once/idempotency

Delete an idempotency key:

**Query Parameters:**
- `queue` (required): Queue name
- `key` (required): Idempotency key
- `tenant` (optional): Tenant ID

#### POST /api/v1/exactly-once/outbox

Trigger outbox event publishing:

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/exactly-once/outbox"
```

#### POST /api/v1/exactly-once/cleanup

Trigger cleanup operations:

**Query Parameters:**
- `type` (optional): Cleanup type ("idempotency", "outbox", "all")

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/exactly-once/cleanup?type=idempotency"
```

#### GET /api/v1/exactly-once/health

Health check endpoint:

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-01-15T10:30:00Z",
  "features": {
    "idempotency": true,
    "outbox": false,
    "metrics": true
  }
}
```

## Metrics

The module exposes Prometheus metrics for monitoring:

### Idempotency Metrics

- `exactly_once_processing_duration_seconds`: Processing time histogram
- `exactly_once_duplicates_avoided_total`: Counter of avoided duplicates
- `exactly_once_successful_processing_total`: Counter of successful processing
- `exactly_once_storage_errors_total`: Counter of storage errors
- `exactly_once_idempotency_checks_total`: Counter of idempotency checks

### Outbox Metrics

- `exactly_once_outbox_published_total`: Counter of published events
- `exactly_once_outbox_failed_total`: Counter of failed publications
- `exactly_once_outbox_processing_duration_seconds`: Outbox processing time

### Storage Metrics

- `exactly_once_storage_operations_total`: Counter of storage operations
- `exactly_once_storage_size`: Gauge of current storage size

## Performance Tuning

### Memory Optimization

1. **Use Redis Hashes**: Enable `use_hashes: true` for better memory efficiency
2. **Set Appropriate TTLs**: Balance between reliability and memory usage
3. **Configure Cleanup**: Regular cleanup prevents unbounded growth

```yaml
idempotency:
  default_ttl: 4h      # Shorter TTL for high-volume queues
  cleanup_interval: 30m # More frequent cleanup
  storage:
    redis:
      use_hashes: true   # Use hashes for memory efficiency
```

### Throughput Optimization

1. **Batch Operations**: Increase batch sizes for bulk operations
2. **Pipeline Operations**: Redis operations are automatically pipelined
3. **Reduce Network Calls**: Use hashes to group related keys

```yaml
idempotency:
  batch_size: 500      # Larger batches for cleanup operations

outbox:
  batch_size: 100      # Process more events per batch
  poll_interval: 2s    # More frequent polling for high throughput
```

### Reliability Configuration

1. **Increase Retry Limits**: For unreliable networks
2. **Configure Circuit Breakers**: Prevent cascade failures
3. **Set Appropriate Timeouts**: Balance responsiveness and reliability

```yaml
outbox:
  max_retries: 10
  retry_backoff:
    initial_delay: 2s
    max_delay: 5m
    multiplier: 1.5
    jitter: true
```

## Troubleshooting

### Common Issues

#### High Memory Usage
- Check TTL settings
- Enable hash-based storage
- Verify cleanup is running
- Monitor key cardinality

#### High Latency
- Check Redis connection
- Monitor storage operations
- Review batch sizes
- Check for lock contention

#### Duplicate Processing
- Verify idempotency keys are unique
- Check TTL configuration
- Monitor storage errors
- Verify key generation logic

### Debugging Tools

```go
// Enable debug logging
cfg.Metrics.Enabled = true

// Get runtime statistics
stats, err := manager.GetDedupStats(ctx, queueName, tenantID)

// Monitor with custom hooks
manager.RegisterHook(&DebuggingHook{})
```

## Migration Guide

### From Non-Idempotent Processing

1. **Identify Critical Operations**: Determine which jobs need idempotency
2. **Add Key Generation**: Implement stable key generation
3. **Wrap Processing Logic**: Use `ProcessWithIdempotency`
4. **Configure Storage**: Set appropriate TTLs and cleanup
5. **Monitor and Tune**: Use metrics to optimize performance

### Example Migration

**Before:**
```go
func processJob(job Job) error {
    return doWork(job)
}
```

**After:**
```go
func processJob(job Job, manager *exactlyonce.Manager) error {
    key := manager.GenerateIdempotencyKey("job-queue", job.TenantID, job.ID)

    _, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
        return nil, doWork(job)
    })

    return err
}
```

This completes the API documentation and tuning guide for the exactly-once patterns module.