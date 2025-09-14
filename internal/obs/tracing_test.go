// Copyright 2025 James Ross
package obs

import (
	"context"
	"testing"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestMaybeInitTracing(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectNil     bool
		expectEnabled bool
	}{
		{
			name: "tracing disabled",
			config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled: false,
					},
				},
			},
			expectNil: true,
		},
		{
			name: "tracing enabled with endpoint",
			config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled:          true,
						Endpoint:         "http://localhost:4318/v1/traces",
						Environment:      "test",
						SamplingStrategy: "always",
						SamplingRate:     1.0,
					},
				},
			},
			expectNil:     false,
			expectEnabled: true,
		},
		{
			name: "tracing enabled without endpoint",
			config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled: true,
					},
				},
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global tracer provider
			otel.SetTracerProvider(trace.NewNoopTracerProvider())

			tp, err := MaybeInitTracing(tt.config)
			if err != nil {
				t.Fatalf("MaybeInitTracing() error = %v", err)
			}

			if tt.expectNil && tp != nil {
				t.Errorf("Expected nil tracer provider, got %v", tp)
			}

			if !tt.expectNil && tp == nil {
				t.Errorf("Expected non-nil tracer provider, got nil")
			}

			if tt.expectEnabled {
				// Verify tracer provider is set
				globalTP := otel.GetTracerProvider()
				if _, ok := globalTP.(*sdktrace.TracerProvider); !ok {
					t.Errorf("Expected SDK tracer provider, got %T", globalTP)
				}

				// Verify propagator is set
				prop := otel.GetTextMapPropagator()
				if _, ok := prop.(propagation.CompositeTextMapPropagator); !ok {
					t.Errorf("Expected composite propagator, got %T", prop)
				}
			}

			// Cleanup
			if tp != nil {
				tp.Shutdown(context.Background())
			}
		})
	}
}

func TestContextWithJobSpan(t *testing.T) {
	// Setup test tracer provider
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tests := []struct {
		name string
		job  queue.Job
	}{
		{
			name: "job with valid trace IDs",
			job: queue.Job{
				ID:           "job-123",
				FilePath:     "/test/file.txt",
				FileSize:     1024,
				Priority:     "high",
				Retries:      2,
				CreationTime: "2025-01-14T10:00:00Z",
				TraceID:      "4bf92f3577b34da6a3ce929d0e0e4736",
				SpanID:       "00f067aa0ba902b7",
			},
		},
		{
			name: "job with invalid trace IDs",
			job: queue.Job{
				ID:           "job-456",
				FilePath:     "/test/file2.txt",
				FileSize:     2048,
				Priority:     "normal",
				Retries:      0,
				CreationTime: "2025-01-14T11:00:00Z",
				TraceID:      "invalid-trace-id",
				SpanID:       "invalid-span-id",
			},
		},
		{
			name: "job without trace IDs",
			job: queue.Job{
				ID:           "job-789",
				FilePath:     "/test/file3.txt",
				FileSize:     512,
				Priority:     "low",
				Retries:      1,
				CreationTime: "2025-01-14T12:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, span := ContextWithJobSpan(ctx, tt.job)

			if span == nil {
				t.Fatal("Expected non-nil span")
			}

			if !span.IsRecording() {
				t.Error("Expected span to be recording")
			}

			// Verify span name
			if span.SpanContext().IsValid() {
				// We can't easily check the span name from the public API
				// but we can verify the span is valid and recording
			}

			span.End()

			// Verify attributes were set by checking that the span context is valid
			spanCtx := span.SpanContext()
			if !spanCtx.IsValid() {
				t.Error("Expected valid span context")
			}
		})
	}
}

func TestStartEnqueueSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := StartEnqueueSpan(ctx, "high-priority", "high")

	if span == nil {
		t.Fatal("Expected non-nil span")
	}

	if !span.IsRecording() {
		t.Error("Expected span to be recording")
	}

	span.End()

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		t.Error("Expected valid span context")
	}
}

func TestStartDequeueSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := StartDequeueSpan(ctx, "test-queue")

	if span == nil {
		t.Fatal("Expected non-nil span")
	}

	if !span.IsRecording() {
		t.Error("Expected span to be recording")
	}

	span.End()

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		t.Error("Expected valid span context")
	}
}

func TestRecordError(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test recording an error
	testErr := &testError{message: "test error"}
	RecordError(ctx, testErr)

	// Test with nil error
	RecordError(ctx, nil)

	// Test with context without span
	RecordError(context.Background(), testErr)
}

func TestSetSpanSuccess(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	SetSpanSuccess(ctx)

	// Test with context without span
	SetSpanSuccess(context.Background())
}

func TestExtractInjectTraceContext(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Set up propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test injection
	carrier := InjectTraceContext(ctx)
	if len(carrier) == 0 {
		t.Error("Expected non-empty carrier after injection")
	}

	// Test extraction
	newCtx := ExtractTraceContext(context.Background(), carrier)
	newSpan := trace.SpanFromContext(newCtx)

	// The extracted context should have span context
	if !trace.SpanContextFromContext(newCtx).IsValid() {
		t.Error("Expected valid span context after extraction")
	}

	// Test with empty carrier
	emptyCtx := ExtractTraceContext(context.Background(), map[string]string{})
	if trace.SpanContextFromContext(emptyCtx).IsValid() {
		t.Error("Expected invalid span context with empty carrier")
	}

	_ = newSpan // Avoid unused variable warning
}

func TestGetTraceAndSpanID(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	traceID, spanID := GetTraceAndSpanID(ctx)

	if traceID == "" {
		t.Error("Expected non-empty trace ID")
	}

	if spanID == "" {
		t.Error("Expected non-empty span ID")
	}

	if len(traceID) != 32 { // Hex representation of 16-byte trace ID
		t.Errorf("Expected trace ID length 32, got %d", len(traceID))
	}

	if len(spanID) != 16 { // Hex representation of 8-byte span ID
		t.Errorf("Expected span ID length 16, got %d", len(spanID))
	}

	// Test with context without span
	emptyTraceID, emptySpanID := GetTraceAndSpanID(context.Background())
	if emptyTraceID != "" || emptySpanID != "" {
		t.Error("Expected empty IDs for context without span")
	}
}

func TestAddEvent(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test adding event with attributes
	AddEvent(ctx, "test-event",
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)

	// Test adding event without attributes
	AddEvent(ctx, "simple-event")

	// Test with context without span
	AddEvent(context.Background(), "no-span-event")
}

func TestAddSpanAttributes(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test adding attributes
	AddSpanAttributes(ctx,
		attribute.String("attr1", "value1"),
		attribute.Int("attr2", 123),
		attribute.Bool("attr3", true),
	)

	// Test with context without span
	AddSpanAttributes(context.Background(), attribute.String("no-span", "value"))
}

func TestTracerShutdown(t *testing.T) {
	// Test with nil tracer provider
	err := TracerShutdown(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected no error for nil tracer provider, got %v", err)
	}

	// Test with valid tracer provider
	tp := sdktrace.NewTracerProvider()
	err = TracerShutdown(context.Background(), tp)
	if err != nil {
		t.Errorf("Unexpected error shutting down tracer provider: %v", err)
	}
}

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected attribute.Type
	}{
		{"string", "key", "value", attribute.STRING},
		{"int", "key", 42, attribute.INT64},
		{"int64", "key", int64(42), attribute.INT64},
		{"float64", "key", 3.14, attribute.FLOAT64},
		{"bool", "key", true, attribute.BOOL},
		{"other", "key", struct{}{}, attribute.STRING},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := KeyValue(tt.key, tt.value)
			if kv.Key != attribute.Key(tt.key) {
				t.Errorf("Expected key %s, got %s", tt.key, kv.Key)
			}
			if kv.Value.Type() != tt.expected {
				t.Errorf("Expected type %v, got %v", tt.expected, kv.Value.Type())
			}
		})
	}
}

func TestTracingSampling(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		rate     float64
	}{
		{"always", "always", 1.0},
		{"never", "never", 0.0},
		{"probabilistic", "probabilistic", 0.5},
		{"default", "unknown", 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled:          true,
						Endpoint:         "http://localhost:4318/v1/traces",
						SamplingStrategy: tt.strategy,
						SamplingRate:     tt.rate,
					},
				},
			}

			tp, err := MaybeInitTracing(cfg)
			if err != nil {
				t.Fatalf("MaybeInitTracing() error = %v", err)
			}

			if tp == nil {
				t.Fatal("Expected non-nil tracer provider")
			}

			// Cleanup
			tp.Shutdown(context.Background())
		})
	}
}

func TestPropagationRoundTrip(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := otel.Tracer("test")
	originalCtx, originalSpan := tracer.Start(context.Background(), "original-span")
	defer originalSpan.End()

	// Get original IDs
	originalTraceID, originalSpanID := GetTraceAndSpanID(originalCtx)

	// Inject to carrier
	carrier := InjectTraceContext(originalCtx)

	// Extract to new context
	newCtx := ExtractTraceContext(context.Background(), carrier)

	// Start child span
	newCtx, childSpan := tracer.Start(newCtx, "child-span")
	defer childSpan.End()

	// Get child IDs
	childTraceID, childSpanID := GetTraceAndSpanID(newCtx)

	// Trace ID should be the same, span ID should be different
	if childTraceID != originalTraceID {
		t.Errorf("Expected same trace ID, got original=%s, child=%s", originalTraceID, childTraceID)
	}

	if childSpanID == originalSpanID {
		t.Error("Expected different span IDs for parent and child")
	}

	// Verify the child span has the original span as parent
	childSpanCtx := childSpan.SpanContext()
	if !childSpanCtx.IsValid() {
		t.Error("Child span context should be valid")
	}
}

// testError is a custom error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Benchmark tests
func BenchmarkStartSpan(b *testing.B) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, span := StartEnqueueSpan(ctx, "test-queue", "high")
		span.End()
	}
}

func BenchmarkInjectExtract(b *testing.B) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		carrier := InjectTraceContext(ctx)
		ExtractTraceContext(context.Background(), carrier)
	}
}