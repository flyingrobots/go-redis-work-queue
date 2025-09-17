# Distributed Tracing Integration

## Overview

The go-redis-work-queue system provides comprehensive distributed tracing capabilities using OpenTelemetry. This integration enables end-to-end visibility into job processing, from enqueue through dequeue to completion, with automatic context propagation and performance metrics.

## Features

### Core Capabilities

- **Automatic Span Creation**: Spans are created for all major operations (enqueue, dequeue, process)
- **Context Propagation**: W3C Trace Context propagation through job metadata
- **Trace Linking**: Jobs maintain trace continuity across system boundaries
- **Performance Metrics**: Histograms with trace exemplars for latency analysis
- **Error Tracking**: Automatic error recording and categorization
- **Event Logging**: Rich event stream within spans for debugging
- **Sampling Strategies**: Configurable sampling (always, never, probabilistic, adaptive)

### Instrumented Operations

#### Producer Operations
- `queue.enqueue`: Traces job enqueue operations with queue and priority attributes
- Automatic trace ID generation or context propagation
- Queue depth and enqueue latency metrics

#### Worker Operations
- `queue.dequeue`: Traces job dequeue from Redis queues
- `job.process`: Full processing lifecycle with parent context restoration
- Retry tracking with backoff timing
- Dead letter queue operations

#### Admin Operations
- `admin.*`: Instrumented admin operations (stats, peek, purge)
- Audit trail with trace correlation
- Destructive operation tracking

## Configuration

### Basic Configuration

```yaml
observability:
  tracing:
    enabled: true
    endpoint: "localhost:4317"  # OTLP endpoint
    environment: "production"
    sampling_strategy: "probabilistic"
    sampling_rate: 0.1  # Sample 10% of traces
    insecure: false  # TLS by default; set true only for local development
```

> **Note:** The tracer validates endpoint schemes—if the endpoint implies TLS (e.g., `https://`, standard OTLP TLS ports) while `insecure: true`, startup fails fast with a clear error so operators do not accidentally ship plaintext.

### Advanced Configuration

```yaml
observability:
  tracing:
    enabled: true
    endpoint: "otel-collector.observability.svc.cluster.local:4317"
    environment: "production"

    # Sampling configuration
    sampling_strategy: "adaptive"  # always, never, probabilistic, adaptive
    sampling_rate: 0.01  # 1% baseline sampling

    # Export configuration
    batch_timeout: 5s
    max_export_batch_size: 512

    # Authentication headers
    headers:
      Authorization: "Bearer ${OTEL_AUTH_TOKEN}"

    # Security
    insecure: false
    redact_sensitive: true

    # Propagation format
    propagation_format: "w3c"  # w3c, b3, jaeger

    # Attribute filtering
    attribute_allowlist:
      - "job.id"
      - "queue.name"
      - "worker.id"

    # Metrics integration
    enable_metric_exemplars: true
```

## Usage

### Enabling Tracing

Tracing is automatically initialized when the application starts if enabled in configuration:

```go
// In main.go or initialization code
tp, err := obs.MaybeInitTracing(cfg)
if err != nil {
    log.Fatal("Failed to initialize tracing", err)
}
defer tp.Shutdown(context.Background())
```

### Manual Instrumentation

For custom operations, use the tracing integration directly:

```go
import "github.com/flyingrobots/go-redis-work-queue/internal/distributed-tracing-integration"

// Initialize tracing integration
ti, err := distributedtracing.New(cfg, logger)
if err != nil {
    return err
}

// Instrument custom operation
err = ti.InstrumentAdminOperation(ctx, "custom_operation", func(ctx context.Context) error {
    // Your operation here
    return nil
})
```

### Context Propagation

The system automatically propagates trace context through job metadata:

```go
// Producer side - context is automatically injected
ctx, span := obs.StartEnqueueSpan(ctx, queueName, priority)
traceID, spanID := obs.GetTraceAndSpanID(ctx)
job.TraceID = traceID
job.SpanID = spanID

// Worker side - context is automatically extracted
ctx, span := obs.ContextWithJobSpan(ctx, job)
// Process job with trace context
```

## Span Attributes

### Standard Attributes

All spans include these standard attributes:
- `service.name`: "go-redis-work-queue"
- `service.version`: Application version
- `host.name`: Hostname
- `environment`: Deployment environment

### Operation-Specific Attributes

#### Enqueue Spans
- `queue.name`: Target queue name
- `queue.priority`: Job priority (high/low)
- `queue.operation`: "enqueue"
- `job.id`: Unique job identifier
- `job.filepath`: File being processed
- `job.filesize`: Size in bytes

#### Dequeue Spans
- `queue.name`: Source queue name
- `queue.operation`: "dequeue"
- `queue.wait_time_ms`: Time spent waiting

#### Process Spans
- `job.id`: Job identifier
- `job.filepath`: File path
- `job.filesize`: File size
- `job.priority`: Priority level
- `job.retries`: Retry count
- `job.creation_time`: Job creation timestamp
- `worker.id`: Worker identifier
- `queue.source`: Source queue
- `processing.duration_ms`: Processing time
- `processing.success`: Success status

## Events

Spans include events marking significant points in processing:

### Enqueue Events
- `enqueueing_job`: Before enqueue
- `job_enqueued`: After successful enqueue

### Dequeue Events
- `job_dequeuing`: Attempting dequeue
- `job_dequeued`: Successfully dequeued

### Processing Events
- `job.processing.started`: Processing begins
- `job.processing.completed`: Successful completion
- `job.processing.failed`: Processing failure
- `job.retrying`: Retry attempt
- `job.dead_lettered`: Sent to DLQ

## Metrics with Exemplars

When `enable_metric_exemplars` is enabled, trace IDs are attached to metrics:

- `queue.enqueue.duration`: Enqueue operation duration
- `queue.dequeue.duration`: Dequeue operation duration
- `job.process.duration`: Job processing duration
- `job.errors`: Error counter with error types

## Integration with Observability Backends

### Jaeger

```yaml
observability:
  tracing:
    endpoint: "jaeger-collector:4317"
    propagation_format: "jaeger"
```

### Zipkin

```yaml
observability:
  tracing:
    endpoint: "zipkin-collector:9411"
    propagation_format: "b3"
```

### AWS X-Ray

```yaml
observability:
  tracing:
    endpoint: "aws-otel-collector:4317"
    headers:
      X-Amz-Security-Token: "${AWS_SESSION_TOKEN}"
```

### Google Cloud Trace

```yaml
observability:
  tracing:
    endpoint: "google-cloud-trace:4317"
    headers:
      Authorization: "Bearer ${GOOGLE_APPLICATION_CREDENTIALS}"
```

## Sampling Strategies

### Always Sample
```yaml
sampling_strategy: "always"
```
Samples every trace. Use only in development.

### Never Sample
```yaml
sampling_strategy: "never"
```
Disables sampling. Useful for testing without overhead.

### Probabilistic Sampling
```yaml
sampling_strategy: "probabilistic"
sampling_rate: 0.01  # 1% sampling
```
Random sampling based on trace ID.

### Adaptive Sampling
```yaml
sampling_strategy: "adaptive"
sampling_rate: 0.001  # 0.1% baseline
```
Adjusts sampling based on traffic patterns and errors.

## TUI Integration

### Enhanced Admin Functions

The distributed tracing integration provides enhanced admin functions that display trace information:

#### PeekWithTracing

```go
// Enhanced peek with trace information
result, err := admin.PeekWithTracing(ctx, cfg, rdb, "high", 10)
if err != nil {
    log.Fatal(err)
}

// Display jobs with trace information
for _, job := range result.TraceJobs {
    fmt.Printf("Job %s: %s [Trace: %s]\n",
        job.ID, job.FilePath, job.TraceID[:8])
}

// Available trace actions for each job
for jobID, actions := range result.TraceActions {
    fmt.Printf("Trace actions for %s:\n", jobID)
    for _, action := range actions {
        switch action.Type {
        case "copy":
            fmt.Printf("  - %s: %s\n", action.Label, action.Command)
        case "open":
            fmt.Printf("  - %s: %s\n", action.Label, action.URL)
        case "view":
            fmt.Printf("  - %s: %s\n", action.Label, action.Description)
        }
    }
}
```

#### InfoWithTracing

```go
// Get detailed job information with trace data
info, err := admin.InfoWithTracing(ctx, cfg, rdb, "high", 0)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job: %s\n", info.JobID)
fmt.Printf("Queue: %s\n", info.Queue)
if info.Job != nil && info.Job.TraceID != "" {
    fmt.Printf("Trace ID: %s\n", info.Job.TraceID)
    fmt.Printf("Span ID: %s\n", info.Job.SpanID)
    fmt.Printf("Trace URL: %s\n", info.TraceURL)
}
```

### Configuration for TUI Trace Actions

Configure trace viewer URLs for TUI integration:

```go
tracing := distributed_tracing_integration.New(distributed_tracing_integration.TracingUIConfig{
    JaegerBaseURL:      "http://localhost:16686",
    ZipkinBaseURL:      "http://localhost:9411",
    CustomTraceURL:     "https://my-trace-viewer.com/trace/{traceID}",
    EnableCopyActions:  true,
    EnableOpenActions:  true,
    DefaultTraceViewer: "jaeger",
})
```

### Available Trace Actions

The TUI provides these actions for each traced job:

1. **Copy Trace ID**: Copy trace ID to clipboard
   - Command: `echo 'abc123def456' | pbcopy`
   - Description: "Copy trace ID to clipboard"

2. **Open Trace**: Open trace in browser
   - URL: `http://localhost:16686/trace/abc123def456`
   - Command: `open 'http://localhost:16686/trace/abc123def456'`
   - Description: "Open trace in jaeger"

3. **View Trace ID**: Display trace information
   - Description: "Trace ID: abc123def456"

## Security Considerations

### Sensitive Data Redaction

When `redact_sensitive` is enabled:
- File paths are anonymized
- Personal identifiers are removed
- Credentials are never logged

### Attribute Filtering

Use `attribute_allowlist` to control which attributes are sent:

```yaml
attribute_allowlist:
  - "job.id"
  - "queue.name"
  # Explicitly exclude sensitive fields
```

### Network Security

For production, use TLS:

```yaml
observability:
  tracing:
    insecure: false
    endpoint: "otel-collector.example.com:4317"
```

## Performance Impact

Typical overhead with 1% sampling:
- CPU: < 1% increase
- Memory: ~10MB for tracer
- Network: ~1KB per sampled trace
- Latency: < 1ms per operation

## Troubleshooting

### No Traces Appearing

1. Check tracing is enabled:
   ```yaml
   observability:
     tracing:
       enabled: true
   ```

2. Verify endpoint connectivity:
   ```bash
   telnet localhost 4317
   ```

3. Check sampling rate isn't 0:
   ```yaml
   sampling_rate: 0.01  # At least 1%
   ```

### Missing Trace Context

Ensure jobs have trace IDs:
```go
log.Info("Job trace info",
  zap.String("trace_id", job.TraceID),
  zap.String("span_id", job.SpanID))
```

### High Memory Usage

Reduce batch size:
```yaml
max_export_batch_size: 128  # Smaller batches
batch_timeout: 1s  # More frequent exports
```

## Example Trace Flow

```
[Browser/CLI] → [Producer] → [Redis] → [Worker] → [Complete]
     |             |           |          |           |
  TraceID:123   SpanID:456     |     SpanID:789   SpanID:abc
     |             |           |          |           |
     └─────────────┴───────────┴──────────┴───────────┘
                    Single Trace (TraceID:123)
```

## Best Practices

1. **Use Sampling in Production**: Full tracing impacts performance
2. **Set Resource Limits**: Configure max batch sizes and timeouts
3. **Monitor Trace Volumes**: Track data egress costs
4. **Use Span Events**: Add events instead of creating many small spans
5. **Leverage Baggage**: Pass user/tenant IDs through baggage
6. **Implement SLOs**: Use traces to measure service level objectives
7. **Regular Cleanup**: Implement retention policies in your backend
