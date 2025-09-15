# Trace Drilldown & Log Tail API Documentation

## Overview

The Trace Drilldown & Log Tail system provides comprehensive distributed tracing and real-time log tailing capabilities with backpressure protection. It enables rapid root cause analysis by correlating traces with logs, supporting multiple tracing providers, and offering powerful filtering and search capabilities.

## Core Components

### TraceManager

Manages distributed tracing with support for multiple providers (Jaeger, Zipkin, Datadog).

```go
traceManager := NewTraceManager(config, redisClient, logger)
```

#### Configuration

```go
config := &TracingConfig{
    Enabled:      true,
    Provider:     "jaeger",              // jaeger, zipkin, datadog
    Endpoint:     "http://jaeger:14268", // Tracing backend endpoint
    ServiceName:  "my-service",
    SamplingRate: 0.1,                   // Sample 10% of traces

    // URL template for external trace viewing
    URLTemplate: "http://jaeger.local/trace/{trace_id}",

    // Headers to propagate
    PropagateHeaders: []string{"X-Request-ID", "X-User-ID"},

    // Authentication
    AuthToken: "secret-token",
}
```

### LogTailer

Provides real-time log tailing with filtering and backpressure protection.

```go
logTailer := NewLogTailer(config, redisClient, logger)
```

#### Configuration

```go
config := &LoggingConfig{
    Enabled:         true,
    RetentionPeriod: 7 * 24 * time.Hour,  // Keep logs for 7 days
    MaxStorageSize:  10 * 1024 * 1024 * 1024, // 10GB max

    // Fields to index for fast searching
    IndexFields: []string{"trace_id", "job_id", "worker_id"},

    // Log parsing formats
    ParseFormats: []string{"json", "logfmt", "syslog"},

    // Log sources
    Sources: []LogSource{
        {
            Name:    "application",
            Type:    "redis",
            Enabled: true,
            Config: map[string]string{
                "channel": "logs:app",
            },
        },
        {
            Name:    "system",
            Type:    "file",
            Enabled: true,
            Config: map[string]string{
                "path": "/var/log/system.log",
            },
        },
    },
}
```

## Tracing Operations

### Starting a Trace

```go
ctx := context.Background()
traceCtx, newCtx := traceManager.StartTrace(ctx, "process-job")

// Use newCtx for all operations within this trace
defer traceManager.EndTrace(newCtx, "success")
```

### Adding Trace Logs

```go
// Add structured logs to the trace
traceManager.AddTraceLog(newCtx, "info", "Processing started", map[string]interface{}{
    "job_id": "job-123",
    "queue":  "high-priority",
})

// Log errors with context
if err != nil {
    traceManager.AddTraceLog(newCtx, "error", err.Error(), map[string]interface{}{
        "error_type": "database",
        "retry_count": 3,
    })
    traceManager.EndTrace(newCtx, "error")
}
```

### Getting Trace Information

```go
// Retrieve trace details
trace, err := traceManager.GetTrace(traceID)
if err == nil {
    fmt.Printf("Operation: %s\n", trace.OperationName)
    fmt.Printf("Duration: %v\n", trace.Duration)
    fmt.Printf("Status: %s\n", trace.Status)

    // Display trace logs
    for _, log := range trace.Logs {
        fmt.Printf("[%s] %s: %s\n",
            log.Timestamp, log.Level, log.Message)
    }
}

// Get external trace link
link, err := traceManager.GetTraceLink(traceID)
if err == nil {
    fmt.Printf("View trace: %s\n", link.URL)
}
```

### Span Summary

```go
// Get comprehensive span information
summary, err := traceManager.GetSpanSummary(ctx, traceID)
if err == nil {
    fmt.Printf("Total Spans: %d\n", summary.TotalSpans)
    fmt.Printf("Duration: %v\n", summary.Duration)
    fmt.Printf("Errors: %d\n", summary.ErrorCount)

    // Display timeline
    for _, event := range summary.Timeline {
        fmt.Printf("%s: %s - %s\n",
            event.Timestamp, event.Operation, event.EventType)
    }
}
```

### Trace Propagation

#### HTTP Requests

```go
// Propagate trace to downstream service
req, _ := http.NewRequest("GET", "http://api/endpoint", nil)
traceManager.PropagateTrace(ctx, req.Header)

// Headers added:
// X-Trace-Id: <trace-id>
// X-Span-Id: <span-id>
// Provider-specific headers (uber-trace-id, X-B3-TraceId, etc.)
```

#### Extract Trace from Request

```go
// In receiving service
func handler(w http.ResponseWriter, r *http.Request) {
    traceCtx := traceManager.ExtractTrace(r.Header)
    if traceCtx != nil {
        ctx := context.WithValue(r.Context(), "trace", traceCtx)
        // Continue with traced context
    }
}
```

### Searching Traces

```go
// Search with filters
filter := &LogFilter{
    StartTime:  time.Now().Add(-1 * time.Hour),
    EndTime:    time.Now(),
    TraceIDs:   []string{"trace-123", "trace-456"},
    SearchText: "error",
}

result, err := traceManager.SearchTraces(ctx, filter)
for _, trace := range result.Traces {
    fmt.Printf("Trace %s: %s (%v)\n",
        trace.TraceID, trace.OperationName, trace.Duration)
}
```

## Log Operations

### Writing Logs

```go
// Write structured log entry
err := logTailer.WriteLog(&LogEntry{
    Timestamp:  time.Now(),
    Level:      "error",
    Message:    "Database connection failed",
    Source:     "db-connector",
    JobID:      "job-789",
    WorkerID:   "worker-2",
    QueueName:  "critical",
    TraceID:    traceCtx.TraceID,
    SpanID:     traceCtx.SpanID,
    Fields: map[string]interface{}{
        "host":     "db-primary",
        "port":     5432,
        "attempts": 3,
    },
    StackTrace: debug.Stack(),
})
```

### Searching Logs

```go
// Complex log search
filter := &LogFilter{
    StartTime:    time.Now().Add(-24 * time.Hour),
    Levels:       []string{"error", "warning"},
    JobIDs:       []string{"job-789"},
    WorkerIDs:    []string{"worker-1", "worker-2"},
    QueueNames:   []string{"critical", "high"},
    SearchText:   "connection",
    MaxResults:   100,
    IncludeStack: true,
}

result, err := logTailer.SearchLogs(ctx, filter)
if err == nil {
    fmt.Printf("Found %d logs (total: %d)\n",
        len(result.Logs), result.TotalCount)

    for _, log := range result.Logs {
        fmt.Printf("[%s] %s: %s\n",
            log.Timestamp, log.Level, log.Message)

        if log.StackTrace != "" {
            fmt.Printf("Stack: %s\n", log.StackTrace)
        }
    }
}
```

### Real-time Log Tailing

```go
// Configure tail session
config := &TailConfig{
    Follow:            true,  // Continue tailing new logs
    BufferSize:        1000,  // Internal buffer size
    MaxLinesPerSecond: 100,   // Rate limiting
    BackpressureLimit: 5000,  // Trigger backpressure
    FlushInterval:     100 * time.Millisecond,

    Filter: &LogFilter{
        Levels:     []string{"error", "warning"},
        WorkerIDs:  []string{"worker-1"},
        TraceIDs:   []string{traceID},
    },
}

// Start tailing
session, eventCh, err := logTailer.StartTail(config)
if err != nil {
    log.Fatal(err)
}

// Process events
go func() {
    for event := range eventCh {
        switch event.Type {
        case "log":
            log := event.Data.(LogEntry)
            fmt.Printf("[%s] %s: %s\n",
                log.Timestamp, log.Level, log.Message)

        case "backpressure":
            status := event.Data.(BackpressureStatus)
            fmt.Printf("Backpressure active! Dropped: %d lines\n",
                status.DroppedLines)

        case "error":
            fmt.Printf("Error: %v\n", event.Data)

        case "status":
            stats := event.Data.(map[string]interface{})
            fmt.Printf("Processed: %d lines\n",
                stats["lines_processed"])
        }
    }
}()

// Stop when done
defer logTailer.StopTail(session.ID)
```

### Log Statistics

```go
stats, err := logTailer.GetLogStats(ctx)
if err == nil {
    fmt.Printf("Total Lines: %d\n", stats.TotalLines)
    fmt.Printf("Rate: %.2f lines/sec\n", stats.LinesPerSecond)
    fmt.Printf("Errors: %d\n", stats.ErrorCount)
    fmt.Printf("Warnings: %d\n", stats.WarningCount)
    fmt.Printf("Unique Traces: %d\n", stats.UniqueTraces)
    fmt.Printf("Storage Used: %d bytes\n", stats.StorageUsed)

    // Level breakdown
    for level, count := range stats.LevelBreakdown {
        fmt.Printf("%s: %d\n", level, count)
    }
}
```

## Usage Examples

### Complete Job Tracing

```go
func processJob(ctx context.Context, job *Job) error {
    // Start trace
    traceCtx, tracedCtx := traceManager.StartTrace(ctx, "process-job")
    defer func() {
        if r := recover(); r != nil {
            traceManager.AddTraceLog(tracedCtx, "panic", fmt.Sprintf("%v", r), nil)
            traceManager.EndTrace(tracedCtx, "panic")
            panic(r)
        }
    }()

    // Log job start
    logTailer.WriteLog(&LogEntry{
        Level:    "info",
        Message:  "Starting job processing",
        JobID:    job.ID,
        TraceID:  traceCtx.TraceID,
    })

    // Process with tracing
    for i, step := range job.Steps {
        traceManager.AddTraceLog(tracedCtx, "info",
            fmt.Sprintf("Processing step %d", i),
            map[string]interface{}{"step": step.Name})

        if err := step.Execute(tracedCtx); err != nil {
            // Log error with trace
            traceManager.AddTraceLog(tracedCtx, "error", err.Error(), nil)
            logTailer.WriteLog(&LogEntry{
                Level:      "error",
                Message:    fmt.Sprintf("Step %d failed: %v", i, err),
                JobID:      job.ID,
                TraceID:    traceCtx.TraceID,
                StackTrace: string(debug.Stack()),
            })

            traceManager.EndTrace(tracedCtx, "error")
            return err
        }
    }

    traceManager.EndTrace(tracedCtx, "success")
    return nil
}
```

### Root Cause Analysis

```go
func analyzeFailure(jobID string) {
    // Find related traces
    logs, _ := logTailer.SearchLogs(ctx, &LogFilter{
        JobIDs: []string{jobID},
        Levels: []string{"error"},
    })

    if len(logs.Logs) > 0 {
        errorLog := logs.Logs[0]

        // Get full trace
        if errorLog.TraceID != "" {
            trace, _ := traceManager.GetTrace(errorLog.TraceID)
            summary, _ := traceManager.GetSpanSummary(ctx, errorLog.TraceID)

            fmt.Printf("Error occurred in trace: %s\n", errorLog.TraceID)
            fmt.Printf("Operation: %s\n", trace.OperationName)
            fmt.Printf("Total duration: %v\n", summary.Duration)

            // Get trace link for detailed view
            link, _ := traceManager.GetTraceLink(errorLog.TraceID)
            fmt.Printf("View in %s: %s\n", link.Type, link.URL)

            // Get all logs for this trace
            traceLogs, _ := logTailer.SearchLogs(ctx, &LogFilter{
                TraceIDs: []string{errorLog.TraceID},
            })

            fmt.Println("\nTrace Timeline:")
            for _, log := range traceLogs.Logs {
                fmt.Printf("[%s] %s: %s\n",
                    log.Timestamp, log.Level, log.Message)
            }
        }
    }
}
```

### Live Debugging

```go
func debugWorker(workerID string) {
    // Start tailing worker logs
    config := &TailConfig{
        Follow: true,
        Filter: &LogFilter{
            WorkerIDs: []string{workerID},
        },
    }

    session, eventCh, _ := logTailer.StartTail(config)
    defer logTailer.StopTail(session.ID)

    fmt.Printf("Tailing logs for worker: %s\n", workerID)

    for event := range eventCh {
        if event.Type == "log" {
            log := event.Data.(LogEntry)

            // Color code by level
            color := ""
            switch log.Level {
            case "error":
                color = "\033[31m" // Red
            case "warning":
                color = "\033[33m" // Yellow
            case "info":
                color = "\033[32m" // Green
            }

            fmt.Printf("%s[%s] %s: %s\033[0m\n",
                color, log.Timestamp.Format("15:04:05"),
                log.Level, log.Message)

            // If error with trace, show link
            if log.Level == "error" && log.TraceID != "" {
                link, _ := traceManager.GetTraceLink(log.TraceID)
                fmt.Printf("  → Trace: %s\n", link.URL)
            }
        }
    }
}
```

### Cross-Service Tracing

```go
// Service A: Initiate trace
func serviceA(ctx context.Context) {
    traceCtx, tracedCtx := traceManager.StartTrace(ctx, "service-a-operation")
    defer traceManager.EndTrace(tracedCtx, "success")

    // Call Service B with trace propagation
    req, _ := http.NewRequest("POST", "http://service-b/api", body)
    traceManager.PropagateTrace(tracedCtx, req.Header)

    resp, err := http.DefaultClient.Do(req)
    // ...
}

// Service B: Continue trace
func serviceB(w http.ResponseWriter, r *http.Request) {
    // Extract trace from incoming request
    traceCtx := traceManager.ExtractTrace(r.Header)

    ctx := r.Context()
    if traceCtx != nil {
        ctx = context.WithValue(ctx, "trace", traceCtx)

        // Continue with same trace
        traceManager.AddTraceLog(ctx, "info", "Service B processing", nil)
    }

    // Process request...
}
```

## Performance Considerations

### Sampling Strategy

```go
// Adaptive sampling based on error rate
config := &TracingConfig{
    SamplingRate: 0.01, // 1% baseline
}

// Increase sampling for errors
if isError {
    // Force sampling for errors
    traceCtx.Sampled = true
}

// Sample all slow operations
if duration > 1*time.Second {
    traceCtx.Sampled = true
}
```

### Log Retention

```go
// Automatic cleanup of old logs
config := &LoggingConfig{
    RetentionPeriod: 7 * 24 * time.Hour,  // 7 days
    MaxStorageSize:  10 * 1024 * 1024 * 1024, // 10GB
}

// Manual cleanup if needed
func cleanupOldLogs() {
    cutoff := time.Now().Add(-30 * 24 * time.Hour)

    filter := &LogFilter{
        EndTime: cutoff,
    }

    // Delete logs older than 30 days
    logTailer.DeleteLogs(filter)
}
```

### Backpressure Management

```go
// Configure aggressive backpressure for high-volume scenarios
config := &TailConfig{
    BufferSize:        10000,  // Large buffer
    MaxLinesPerSecond: 1000,   // High rate limit
    BackpressureLimit: 20000,  // Trigger at 2x buffer

    // Drop old logs when overwhelmed
    DropPolicy: "oldest", // or "newest", "sample"
}
```

## Best Practices

### Structured Logging

```go
// Use consistent field names
log := &LogEntry{
    Level:   "info",
    Message: "Operation completed",
    Fields: map[string]interface{}{
        "duration_ms": duration.Milliseconds(),
        "item_count":  count,
        "success":     true,
    },
}

// Standard fields for correlation
log.TraceID = getTraceID(ctx)
log.JobID = getJobID(ctx)
log.WorkerID = getWorkerID()
```

### Error Context

```go
// Rich error logging
if err != nil {
    logTailer.WriteLog(&LogEntry{
        Level:      "error",
        Message:    err.Error(),
        TraceID:    traceID,
        StackTrace: string(debug.Stack()),
        Fields: map[string]interface{}{
            "error_type":   fmt.Sprintf("%T", err),
            "retry_count":  retries,
            "last_success": lastSuccess,
            "context":      errorContext,
        },
    })
}
```

### Trace Naming

```go
// Use descriptive operation names
good := []string{
    "process-payment",
    "validate-order",
    "send-notification",
}

// Include context in operation name
operationName := fmt.Sprintf("process-%s-job", job.Type)
```

## Troubleshooting

### Missing Traces

```go
// Check sampling rate
if trace == nil {
    // May not be sampled
    fmt.Printf("Sampling rate: %.2f%%\n",
        config.SamplingRate * 100)
}

// Force sampling for debugging
config.SamplingRate = 1.0 // Sample everything
```

### Log Overflow

```go
// Monitor backpressure
session, eventCh, _ := logTailer.StartTail(config)

for event := range eventCh {
    if event.Type == "backpressure" {
        status := event.Data.(BackpressureStatus)

        if status.DroppedLines > 100 {
            // Increase limits or add filtering
            config.MaxLinesPerSecond *= 2
            config.BufferSize *= 2
        }
    }
}
```

### Trace Propagation Issues

```go
// Debug headers
fmt.Printf("Trace headers:\n")
for key, values := range req.Header {
    if strings.Contains(strings.ToLower(key), "trace") {
        fmt.Printf("  %s: %v\n", key, values)
    }
}

// Verify provider configuration
if traceCtx == nil {
    fmt.Printf("Provider: %s\n", config.Provider)
    fmt.Printf("Expected headers for %s\n", config.Provider)
}
```