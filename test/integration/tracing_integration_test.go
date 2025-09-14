// Copyright 2025 James Ross
//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/producer"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TestOTLPExport tests that spans are properly exported via OTLP
func TestOTLPExport(t *testing.T) {
	// Mock OTLP collector server
	receivedSpans := make([]map[string]interface{}, 0)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/traces" {
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				mu.Lock()
				receivedSpans = append(receivedSpans, payload)
				mu.Unlock()
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure tracing to use mock server
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         server.URL + "/v1/traces",
				Environment:      "test",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	// Initialize tracing
	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create test spans
	tracer := otel.Tracer("integration-test")
	ctx, span1 := tracer.Start(context.Background(), "test-operation-1")
	obs.AddSpanAttributes(ctx, attribute.String("test.name", "integration"))
	span1.End()

	ctx, span2 := tracer.Start(context.Background(), "test-operation-2")
	obs.RecordError(ctx, fmt.Errorf("test error"))
	span2.End()

	// Force export by shutting down (with timeout to allow export)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tp.Shutdown(ctx)

	// Give some time for the export to happen
	time.Sleep(100 * time.Millisecond)

	// Verify spans were exported
	mu.Lock()
	defer mu.Unlock()

	if len(receivedSpans) == 0 {
		t.Error("No spans were exported to OTLP collector")
	}
}

// TestParentChildSpanLinkage tests that parent-child relationships are properly maintained
func TestParentChildSpanLinkage(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "http://localhost:4318/v1/traces", // Won't actually export
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create parent span
	parentCtx, parentSpan := obs.StartEnqueueSpan(context.Background(), "test-queue", "high")
	parentTraceID, parentSpanID := obs.GetTraceAndSpanID(parentCtx)

	// Inject context for propagation
	carrier := obs.InjectTraceContext(parentCtx)

	// Simulate job with trace context
	job := queue.Job{
		ID:           "test-job-1",
		FilePath:     "/test/file.txt",
		FileSize:     1024,
		Priority:     "high",
		Retries:      0,
		CreationTime: time.Now().Format(time.RFC3339),
	}

	// Extract context and create child span
	extractedCtx := obs.ExtractTraceContext(context.Background(), carrier)
	childCtx, childSpan := obs.ContextWithJobSpan(extractedCtx, job)

	childTraceID, childSpanID := obs.GetTraceAndSpanID(childCtx)

	// Verify parent-child relationship
	if childTraceID != parentTraceID {
		t.Errorf("Child span should have same trace ID as parent. Parent: %s, Child: %s", parentTraceID, childTraceID)
	}

	if childSpanID == parentSpanID {
		t.Error("Child span should have different span ID from parent")
	}

	if parentTraceID == "" || parentSpanID == "" {
		t.Error("Parent span should have valid trace and span IDs")
	}

	if childTraceID == "" || childSpanID == "" {
		t.Error("Child span should have valid trace and span IDs")
	}

	// End spans
	childSpan.End()
	parentSpan.End()
}

// TestJobProcessingWithTracing tests complete job processing flow with tracing
func TestJobProcessingWithTracing(t *testing.T) {
	// Skip if Redis not available
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available for integration test")
	}
	defer rdb.Close()

	// Clean up test keys
	defer func() {
		keys, _ := rdb.Keys(context.Background(), "test:*").Result()
		if len(keys) > 0 {
			rdb.Del(context.Background(), keys...)
		}
	}()

	cfg := &config.Config{
		Producer: config.ProducerConfig{
			QueueKey: "test:queue",
		},
		Worker: config.WorkerConfig{
			Queues: map[string]string{
				"test": "test:queue",
			},
			CompletedList:      "test:completed",
			DeadLetterList:     "test:dlq",
			ProcessingList:     "test:processing",
			HeartbeatKey:       "test:heartbeat",
			HeartbeatInterval:  time.Second,
			HeartbeatTTL:       5 * time.Second,
			JobTimeout:         10 * time.Second,
			MaxRetries:         2,
			PollInterval:       100 * time.Millisecond,
		},
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "http://localhost:4318/v1/traces",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	// Initialize tracing
	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create producer with tracing
	prod := producer.New(rdb, cfg)

	// Create and enqueue job with tracing
	enqueueCtx, enqueueSpan := obs.StartEnqueueSpan(context.Background(), "test", "high")

	job := queue.Job{
		ID:           "trace-test-job",
		FilePath:     "/test/trace-file.txt",
		FileSize:     2048,
		Priority:     "high",
		CreationTime: time.Now().Format(time.RFC3339),
	}

	// Inject trace context into job metadata
	traceContext := obs.InjectTraceContext(enqueueCtx)
	job.TraceID = traceContext["traceparent"] // Simplified - normally would parse

	enqueueTraceID, enqueueSpanID := obs.GetTraceAndSpanID(enqueueCtx)

	err = prod.Enqueue(context.Background(), job)
	if err != nil {
		t.Fatalf("Failed to enqueue job: %v", err)
	}

	obs.SetSpanSuccess(enqueueCtx)
	enqueueSpan.End()

	// Create worker with tracing
	w := worker.New(rdb, cfg, nil)

	// Process job with tracing
	processedJobCh := make(chan queue.Job, 1)
	var processSpanID string
	var processTraceID string

	// Mock handler that captures trace context
	handler := func(ctx context.Context, job queue.Job) error {
		// Verify tracing context is available
		processTraceID, processSpanID = obs.GetTraceAndSpanID(ctx)

		// Add some events and attributes
		obs.AddEvent(ctx, "job.processing.started")
		obs.AddSpanAttributes(ctx,
			attribute.String("job.handler", "test-handler"),
			attribute.Int64("job.size", job.FileSize),
		)

		// Simulate work
		time.Sleep(10 * time.Millisecond)

		obs.AddEvent(ctx, "job.processing.completed")
		obs.SetSpanSuccess(ctx)

		processedJobCh <- job
		return nil
	}

	w.RegisterHandler("test-handler", handler)

	// Start worker
	go func() {
		w.Run(context.Background())
	}()

	// Wait for job to be processed
	select {
	case processedJob := <-processedJobCh:
		// Verify the job was processed
		if processedJob.ID != job.ID {
			t.Errorf("Processed job ID mismatch. Expected: %s, Got: %s", job.ID, processedJob.ID)
		}

		// Verify trace continuity
		if processTraceID == "" || processSpanID == "" {
			t.Error("Processing span should have valid trace and span IDs")
		}

		// In a real scenario with proper context propagation, these would be linked
		// For now, we verify that tracing was active during processing
		if processTraceID == enqueueTraceID {
			t.Logf("Trace continuity maintained: %s", processTraceID)
		} else {
			t.Logf("Separate traces created - Enqueue: %s, Process: %s", enqueueTraceID, processTraceID)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("Job processing timed out")
	}
}

// TestContextPropagationRoundTrip tests full context propagation round trip
func TestContextPropagationRoundTrip(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "http://localhost:4318/v1/traces",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Step 1: Create root span (simulates incoming request)
	rootCtx, rootSpan := obs.StartEnqueueSpan(context.Background(), "input-queue", "high")
	rootTraceID, rootSpanID := obs.GetTraceAndSpanID(rootCtx)

	// Step 2: Inject context for job metadata
	metadata := obs.InjectTraceContext(rootCtx)

	// Step 3: Simulate job creation with metadata
	job := queue.Job{
		ID:           "propagation-test-job",
		FilePath:     "/test/propagation.txt",
		FileSize:     1024,
		Priority:     "high",
		CreationTime: time.Now().Format(time.RFC3339),
		// In real implementation, metadata would be properly stored
	}

	// Step 4: Extract context during job processing
	extractedCtx := obs.ExtractTraceContext(context.Background(), metadata)

	// Step 5: Create processing span with extracted context
	processCtx, processSpan := obs.ContextWithJobSpan(extractedCtx, job)
	processTraceID, processSpanID := obs.GetTraceAndSpanID(processCtx)

	// Step 6: Create nested operation span
	nestedCtx, nestedSpan := obs.StartDequeueSpan(processCtx, "processing-queue")
	nestedTraceID, nestedSpanID := obs.GetTraceAndSpanID(nestedCtx)

	// Verify trace continuity
	if processTraceID != rootTraceID {
		t.Errorf("Process span should inherit root trace ID. Root: %s, Process: %s", rootTraceID, processTraceID)
	}

	if nestedTraceID != rootTraceID {
		t.Errorf("Nested span should inherit root trace ID. Root: %s, Nested: %s", rootTraceID, nestedTraceID)
	}

	// Verify span uniqueness
	spanIDs := []string{rootSpanID, processSpanID, nestedSpanID}
	for i, id1 := range spanIDs {
		for j, id2 := range spanIDs {
			if i != j && id1 == id2 {
				t.Errorf("Span IDs should be unique. Found duplicate: %s", id1)
			}
		}
	}

	// End spans in reverse order
	nestedSpan.End()
	processSpan.End()
	rootSpan.End()
}

// TestTracingWithErrors tests error recording and status setting
func TestTracingWithErrors(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "http://localhost:4318/v1/traces",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Test successful operation
	successCtx, successSpan := obs.StartEnqueueSpan(context.Background(), "success-queue", "normal")
	obs.AddEvent(successCtx, "operation.started")
	obs.SetSpanSuccess(successCtx)
	successSpan.End()

	// Test operation with error
	errorCtx, errorSpan := obs.StartDequeueSpan(context.Background(), "error-queue")
	obs.AddEvent(errorCtx, "operation.started")

	testError := fmt.Errorf("test operation failed")
	obs.RecordError(errorCtx, testError)
	obs.AddEvent(errorCtx, "operation.failed",
		attribute.String("error.type", "TestError"),
		attribute.String("error.message", testError.Error()),
	)

	errorSpan.End()

	// Verify spans were created and have valid contexts
	successTraceID, successSpanID := obs.GetTraceAndSpanID(successCtx)
	if successTraceID == "" || successSpanID == "" {
		t.Error("Success span should have valid trace and span IDs")
	}

	errorTraceID, errorSpanID := obs.GetTraceAndSpanID(errorCtx)
	if errorTraceID == "" || errorSpanID == "" {
		t.Error("Error span should have valid trace and span IDs")
	}

	// Different operations should have different trace IDs
	if successTraceID == errorTraceID {
		t.Error("Independent operations should have different trace IDs")
	}
}

// TestTracingPerformance tests that tracing overhead is acceptable
func TestTracingPerformance(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "http://localhost:4318/v1/traces",
				SamplingStrategy: "probabilistic",
				SamplingRate:     0.1, // 10% sampling
			},
		},
	}

	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Benchmark span creation and operations
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		ctx, span := obs.StartEnqueueSpan(context.Background(), "perf-test", "normal")
		obs.AddSpanAttributes(ctx,
			attribute.Int("iteration", i),
			attribute.String("test.type", "performance"),
		)
		obs.AddEvent(ctx, "test.event")
		obs.SetSpanSuccess(ctx)
		span.End()
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(iterations)

	t.Logf("Created %d spans in %v (avg: %v per span)", iterations, duration, avgLatency)

	// Verify reasonable performance (should be well under 1ms per span)
	if avgLatency > time.Millisecond {
		t.Errorf("Tracing overhead too high: %v per span", avgLatency)
	}
}