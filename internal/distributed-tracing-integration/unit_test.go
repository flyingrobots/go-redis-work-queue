// Copyright 2025 James Ross
package distributedtracing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "none",
		},
		{
			name:     "generic error",
			err:      errors.New("test error"),
			expected: "generic",
		},
		{
			name:     "timeout error",
			err:      context.DeadlineExceeded,
			expected: "generic", // Current implementation doesn't categorize
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorType(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTracingIntegration_InitMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled: true,
			},
		},
	}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	err := ti.initMetrics()
	assert.NoError(t, err)

	// Verify all metrics are initialized
	assert.NotNil(t, ti.enqueueDuration)
	assert.NotNil(t, ti.dequeueDuration)
	assert.NotNil(t, ti.processDuration)
	assert.NotNil(t, ti.errorCounter)
}

func TestTracingIntegration_PropagationHelpers(t *testing.T) {
	// Setup test tracer
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	t.Run("InjectContextToMetadata", func(t *testing.T) {
		// Create a span to get trace context
		ctx, span := tp.Tracer("test").Start(context.Background(), "test-span")
		defer span.End()

		metadata := make(map[string]string)
		ti.InjectContextToMetadata(ctx, metadata)

		// Verify trace context was injected
		assert.NotEmpty(t, metadata)

		// Should have W3C trace context headers
		found := false
		for key := range metadata {
			if key == "traceparent" || key == "tracestate" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected W3C trace context headers in metadata")
	})

	t.Run("ExtractContextFromMetadata", func(t *testing.T) {
		// Create metadata with trace context
		metadata := map[string]string{
			"traceparent": "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01",
		}

		ctx := ti.ExtractContextFromMetadata(context.Background(), metadata)

		// Verify context was extracted
		span := oteltrace.SpanFromContext(ctx)
		assert.NotNil(t, span)

		spanContext := span.SpanContext()
		assert.True(t, spanContext.IsValid())
	})

	t.Run("RoundTripPropagation", func(t *testing.T) {
		// Create original context with span
		originalCtx, span := tp.Tracer("test").Start(context.Background(), "original-span")
		originalTraceID := span.SpanContext().TraceID()
		span.End()

		// Inject into metadata
		metadata := make(map[string]string)
		ti.InjectContextToMetadata(originalCtx, metadata)

		// Extract from metadata
		extractedCtx := ti.ExtractContextFromMetadata(context.Background(), metadata)

		// Verify the trace ID is preserved
		extractedSpan := oteltrace.SpanFromContext(extractedCtx)
		if extractedSpan.SpanContext().IsValid() {
			assert.Equal(t, originalTraceID, extractedSpan.SpanContext().TraceID())
		}
	})
}

func TestTracingIntegration_AttributeSets(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	err := ti.initMetrics()
	require.NoError(t, err)

	t.Run("EnqueueAttributeSet", func(t *testing.T) {
		exporter.Reset()

		job := &queue.Job{
			ID:           "test-job-123",
			FilePath:     "/test/file.txt",
			FileSize:     1024,
			Priority:     "high",
			Retries:      2,
			CreationTime: time.Now(),
		}

		ctx := context.Background()
		ctx, err := ti.EnqueueWithTracing(ctx, job, "test-queue")
		assert.NoError(t, err)

		tp.ForceFlush(context.Background())

		// Verify span attributes
		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var enqueueSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "queue.enqueue" {
				enqueueSpan = &span
				break
			}
		}

		require.NotNil(t, enqueueSpan)

		// Check required attributes
		expectedAttrs := map[string]interface{}{
			"job.id":       "test-job-123",
			"job.filepath": "/test/file.txt",
			"job.filesize": int64(1024),
			"job.retries":  2,
		}

		actualAttrs := make(map[string]interface{})
		for _, attr := range enqueueSpan.Attributes {
			actualAttrs[string(attr.Key)] = attr.Value.AsInterface()
		}

		for key, expected := range expectedAttrs {
			actual, exists := actualAttrs[key]
			assert.True(t, exists, "Attribute %s not found", key)
			assert.Equal(t, expected, actual, "Attribute %s value mismatch", key)
		}
	})

	t.Run("ProcessingAttributeSet", func(t *testing.T) {
		exporter.Reset()

		job := queue.Job{
			ID:           "process-job-456",
			FilePath:     "/process/file.txt",
			FileSize:     2048,
			Priority:     "low",
			Retries:      0,
			CreationTime: time.Now(),
			TraceID:      "0123456789abcdef0123456789abcdef",
			SpanID:       "0123456789abcdef",
		}

		handler := func(ctx context.Context, j queue.Job) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		ctx := context.Background()
		err := ti.ProcessWithTracing(ctx, job, handler)
		assert.NoError(t, err)

		tp.ForceFlush(context.Background())

		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var processSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "job.process" {
				processSpan = &span
				break
			}
		}

		require.NotNil(t, processSpan)

		// Check processing-specific attributes
		hasSuccess := false
		hasDuration := false

		for _, attr := range processSpan.Attributes {
			if string(attr.Key) == "processing.success" {
				hasSuccess = true
				assert.True(t, attr.Value.AsBool())
			}
			if string(attr.Key) == "processing.duration_ms" {
				hasDuration = true
				assert.Greater(t, attr.Value.AsFloat64(), 0.0)
			}
		}

		assert.True(t, hasSuccess, "Success attribute not found")
		assert.True(t, hasDuration, "Duration attribute not found")
	})
}

func TestTracingIntegration_ErrorRecording(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	err := ti.initMetrics()
	require.NoError(t, err)

	t.Run("ProcessingError", func(t *testing.T) {
		exporter.Reset()

		job := queue.Job{
			ID:           "error-job-789",
			FilePath:     "/error/file.txt",
			Priority:     "high",
			Retries:      1,
			CreationTime: time.Now(),
		}

		testError := errors.New("processing failed")
		handler := func(ctx context.Context, j queue.Job) error {
			return testError
		}

		ctx := context.Background()
		err := ti.ProcessWithTracing(ctx, job, handler)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		tp.ForceFlush(context.Background())

		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var processSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "job.process" {
				processSpan = &span
				break
			}
		}

		require.NotNil(t, processSpan)

		// Check error recording
		assert.Equal(t, codes.Error, processSpan.Status.Code)
		assert.Contains(t, processSpan.Status.Description, "processing failed")

		// Check error event
		hasFailedEvent := false
		for _, event := range processSpan.Events {
			if event.Name == "job.processing.failed" {
				hasFailedEvent = true

				// Check event attributes
				hasError := false
				for _, attr := range event.Attributes {
					if string(attr.Key) == "error" && attr.Value.AsString() == "processing failed" {
						hasError = true
						break
					}
				}
				assert.True(t, hasError, "Error attribute not found in failed event")
			}
		}

		assert.True(t, hasFailedEvent, "Processing failed event not found")

		// Check success attribute is false
		hasSuccessAttr := false
		for _, attr := range processSpan.Attributes {
			if string(attr.Key) == "processing.success" {
				hasSuccessAttr = true
				assert.False(t, attr.Value.AsBool())
				break
			}
		}
		assert.True(t, hasSuccessAttr, "Success attribute not found")
	})

	t.Run("AdminOperationError", func(t *testing.T) {
		exporter.Reset()

		testError := errors.New("admin operation failed")
		operation := func(ctx context.Context) error {
			return testError
		}

		ctx := context.Background()
		err := ti.InstrumentAdminOperation(ctx, "test_operation", operation)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		tp.ForceFlush(context.Background())

		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var adminSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "admin.test_operation" {
				adminSpan = &span
				break
			}
		}

		require.NotNil(t, adminSpan)

		// Check error status
		assert.Equal(t, codes.Error, adminSpan.Status.Code)

		// Check operation success attribute is false
		hasSuccessAttr := false
		for _, attr := range adminSpan.Attributes {
			if string(attr.Key) == "operation.success" {
				hasSuccessAttr = true
				assert.False(t, attr.Value.AsBool())
				break
			}
		}
		assert.True(t, hasSuccessAttr, "Operation success attribute not found")
	})
}

func TestTracingIntegration_Shutdown(t *testing.T) {
	t.Run("ShutdownWithProvider", func(t *testing.T) {
		tp := trace.NewTracerProvider()

		ti := &TracingIntegration{
			provider: tp,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := ti.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("ShutdownWithoutProvider", func(t *testing.T) {
		ti := &TracingIntegration{
			provider: nil,
		}

		ctx := context.Background()
		err := ti.Shutdown(ctx)
		assert.NoError(t, err)
	})
}

func TestTracingIntegration_SpanEvents(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	err := ti.initMetrics()
	require.NoError(t, err)

	t.Run("EnqueueEvents", func(t *testing.T) {
		exporter.Reset()

		job := &queue.Job{
			ID:       "event-test-job",
			Priority: "high",
		}

		ctx := context.Background()
		_, err := ti.EnqueueWithTracing(ctx, job, "event-queue")
		assert.NoError(t, err)

		tp.ForceFlush(context.Background())

		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var enqueueSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "queue.enqueue" {
				enqueueSpan = &span
				break
			}
		}

		require.NotNil(t, enqueueSpan)

		// Check for enqueued event
		hasEnqueuedEvent := false
		for _, event := range enqueueSpan.Events {
			if event.Name == "job.enqueued" {
				hasEnqueuedEvent = true

				// Check event attributes
				hasQueue := false
				hasJobID := false
				for _, attr := range event.Attributes {
					if string(attr.Key) == "queue" && attr.Value.AsString() == "event-queue" {
						hasQueue = true
					}
					if string(attr.Key) == "job_id" && attr.Value.AsString() == "event-test-job" {
						hasJobID = true
					}
				}

				assert.True(t, hasQueue, "Queue attribute not found in enqueued event")
				assert.True(t, hasJobID, "Job ID attribute not found in enqueued event")
				break
			}
		}

		assert.True(t, hasEnqueuedEvent, "Job enqueued event not found")
	})

	t.Run("ProcessingEvents", func(t *testing.T) {
		exporter.Reset()

		job := queue.Job{
			ID:           "processing-event-job",
			Priority:     "low",
			CreationTime: time.Now(),
		}

		handler := func(ctx context.Context, j queue.Job) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		ctx := context.Background()
		err := ti.ProcessWithTracing(ctx, job, handler)
		assert.NoError(t, err)

		tp.ForceFlush(context.Background())

		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		var processSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "job.process" {
				processSpan = &span
				break
			}
		}

		require.NotNil(t, processSpan)

		// Check for processing events
		events := map[string]bool{
			"job.processing.started":   false,
			"job.processing.completed": false,
		}

		for _, event := range processSpan.Events {
			if _, exists := events[event.Name]; exists {
				events[event.Name] = true

				// Check job_id attribute in events
				hasJobID := false
				for _, attr := range event.Attributes {
					if string(attr.Key) == "job_id" && attr.Value.AsString() == "processing-event-job" {
						hasJobID = true
						break
					}
				}
				assert.True(t, hasJobID, "Job ID attribute not found in %s event", event.Name)
			}
		}

		for eventName, found := range events {
			assert.True(t, found, "Event %s not found", eventName)
		}
	})
}

func TestTracingIntegration_DequeueWithTracing(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	ti := &TracingIntegration{
		config: cfg,
		tracer: otel.Tracer("test"),
		meter:  otel.Meter("test"),
		logger: logger,
	}

	err := ti.initMetrics()
	require.NoError(t, err)

	ctx := context.Background()
	ctx, job, err := ti.DequeueWithTracing(ctx, "test-dequeue-queue")
	assert.NoError(t, err)
	assert.NotNil(t, job)

	tp.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Greater(t, len(spans), 0)

	var dequeueSpan *tracetest.SpanStub
	for _, span := range spans {
		if span.Name == "queue.dequeue" {
			dequeueSpan = &span
			break
		}
	}

	require.NotNil(t, dequeueSpan, "Dequeue span not found")

	// Check span attributes
	hasQueueName := false
	for _, attr := range dequeueSpan.Attributes {
		if string(attr.Key) == "queue.name" && attr.Value.AsString() == "test-dequeue-queue" {
			hasQueueName = true
			break
		}
	}
	assert.True(t, hasQueueName, "Queue name attribute not found")

	// Check dequeue events
	events := map[string]bool{
		"job.dequeuing": false,
		"job.dequeued":  false,
	}

	for _, event := range dequeueSpan.Events {
		if _, exists := events[event.Name]; exists {
			events[event.Name] = true
		}
	}

	for eventName, found := range events {
		assert.True(t, found, "Event %s not found", eventName)
	}
}

func TestTracingConfig_UI(t *testing.T) {
	t.Run("DefaultTracingUIConfig", func(t *testing.T) {
		config := DefaultTracingUIConfig()

		assert.Equal(t, "http://localhost:16686", config.JaegerBaseURL)
		assert.Equal(t, "http://localhost:9411", config.ZipkinBaseURL)
		assert.True(t, config.EnableCopyActions)
		assert.True(t, config.EnableOpenActions)
		assert.Equal(t, "jaeger", config.DefaultTraceViewer)
	})

	t.Run("GetTraceURL_Jaeger", func(t *testing.T) {
		config := TracingUIConfig{
			JaegerBaseURL:      "http://localhost:16686",
			DefaultTraceViewer: "jaeger",
		}

		traceID := "0123456789abcdef0123456789abcdef"
		url := config.GetTraceURL(traceID)

		expected := "http://localhost:16686/trace/0123456789abcdef0123456789abcdef"
		assert.Equal(t, expected, url)
	})

	t.Run("GetTraceURL_Zipkin", func(t *testing.T) {
		config := TracingUIConfig{
			ZipkinBaseURL:      "http://localhost:9411",
			DefaultTraceViewer: "zipkin",
		}

		traceID := "0123456789abcdef0123456789abcdef"
		url := config.GetTraceURL(traceID)

		expected := "http://localhost:9411/zipkin/traces/0123456789abcdef0123456789abcdef"
		assert.Equal(t, expected, url)
	})

	t.Run("GetTraceURL_Custom", func(t *testing.T) {
		config := TracingUIConfig{
			CustomTraceURL:     "https://my-tracing.com/trace/{traceID}",
			DefaultTraceViewer: "custom",
		}

		traceID := "0123456789abcdef0123456789abcdef"
		url := config.GetTraceURL(traceID)

		expected := "https://my-tracing.com/trace/0123456789abcdef0123456789abcdef"
		assert.Equal(t, expected, url)
	})

	t.Run("GetTraceURL_EmptyTraceID", func(t *testing.T) {
		config := DefaultTracingUIConfig()

		url := config.GetTraceURL("")
		assert.Empty(t, url)
	})

	t.Run("ReplaceTraceID_VariousFormats", func(t *testing.T) {
		traceID := "abc123"

		tests := []struct {
			template string
			expected string
		}{
			{"{traceID}", "abc123"},
			{"{trace_id}", "abc123"},
			{"{{traceID}}", "abc123"},
			{"{{trace_id}}", "abc123"},
			{"https://example.com/trace/{traceID}/view", "https://example.com/trace/abc123/view"},
			{"Multiple {traceID} and {trace_id} placeholders", "Multiple abc123 and abc123 placeholders"},
		}

		for _, test := range tests {
			result := replaceTraceID(test.template, traceID)
			assert.Equal(t, test.expected, result)
		}
	})
}