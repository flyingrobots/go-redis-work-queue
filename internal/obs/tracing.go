// Copyright 2025 James Ross
package obs

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// MaybeInitTracing optionally initializes a global tracer provider with sampling and propagation.
func MaybeInitTracing(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Observability.Tracing.Enabled || cfg.Observability.Tracing.Endpoint == "" {
		return nil, nil
	}

	// Create OTLP exporter
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.Observability.Tracing.Endpoint),
		otlptracehttp.WithInsecure(),
	))
	if err != nil {
		return nil, err
	}

	// Get hostname for resource attributes
	hostname, _ := os.Hostname()

	// Create resource with service information
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("go-redis-work-queue"),
		semconv.ServiceVersionKey.String("1.0.0"),
		semconv.HostNameKey.String(hostname),
		attribute.String("environment", cfg.Observability.Tracing.Environment),
	)

	// Configure sampler based on config
	var sampler sdktrace.Sampler
	switch cfg.Observability.Tracing.SamplingStrategy {
	case "always":
		sampler = sdktrace.AlwaysSample()
	case "never":
		sampler = sdktrace.NeverSample()
	case "probabilistic":
		sampler = sdktrace.TraceIDRatioBased(cfg.Observability.Tracing.SamplingRate)
	default:
		// Default to probabilistic with configured rate
		sampler = sdktrace.TraceIDRatioBased(cfg.Observability.Tracing.SamplingRate)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for W3C Trace Context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// ContextWithJobSpan creates a span for processing a job, attempting to honor
// job-provided TraceID/SpanID as a remote parent when present and valid.
func ContextWithJobSpan(ctx context.Context, job queue.Job) (context.Context, trace.Span) {
	tracer := otel.Tracer("worker")

	// Try to parse provided IDs for context propagation
	if tid, err := trace.TraceIDFromHex(job.TraceID); err == nil {
		if sid, err2 := trace.SpanIDFromHex(job.SpanID); err2 == nil {
			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    tid,
				SpanID:     sid,
				TraceFlags: trace.FlagsSampled, // Mark as sampled
				Remote:     true,
			})
			ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}

	// Start span with attributes
	ctx, span := tracer.Start(ctx, "job.process",
		trace.WithAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.filepath", job.FilePath),
			attribute.Int64("job.filesize", job.FileSize),
			attribute.String("job.priority", job.Priority),
			attribute.Int("job.retries", job.Retries),
			attribute.String("job.creation_time", job.CreationTime),
			attribute.String("queue.type", "worker"),
		),
	)

	return ctx, span
}

// StartEnqueueSpan creates a span for enqueueing a job.
func StartEnqueueSpan(ctx context.Context, queueName string, priority string) (context.Context, trace.Span) {
	tracer := otel.Tracer("producer")
	return tracer.Start(ctx, "queue.enqueue",
		trace.WithAttributes(
			attribute.String("queue.name", queueName),
			attribute.String("queue.priority", priority),
			attribute.String("queue.operation", "enqueue"),
		),
	)
}

// StartDequeueSpan creates a span for dequeuing a job.
func StartDequeueSpan(ctx context.Context, queueName string) (context.Context, trace.Span) {
	tracer := otel.Tracer("worker")
	return tracer.Start(ctx, "queue.dequeue",
		trace.WithAttributes(
			attribute.String("queue.name", queueName),
			attribute.String("queue.operation", "dequeue"),
		),
	)
}

// RecordError records an error on the span if one exists in the context.
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() && err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanSuccess marks the span as successful.
func SetSpanSuccess(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(codes.Ok, "success")
	}
}

// ExtractTraceContext extracts trace context from a map (for job metadata).
func ExtractTraceContext(ctx context.Context, carrier map[string]string) context.Context {
	prop := otel.GetTextMapPropagator()
	return prop.Extract(ctx, propagation.MapCarrier(carrier))
}

// InjectTraceContext injects trace context into a map (for job metadata).
func InjectTraceContext(ctx context.Context) map[string]string {
	carrier := make(map[string]string)
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, propagation.MapCarrier(carrier))
	return carrier
}

// GetTraceAndSpanID extracts the current trace and span IDs from context.
func GetTraceAndSpanID(ctx context.Context) (traceID string, spanID string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		sc := span.SpanContext()
		if sc.IsValid() {
			return sc.TraceID().String(), sc.SpanID().String()
		}
	}
	return "", ""
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// AddSpanAttributes adds attributes to the current span.
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// TracerShutdown gracefully shuts down the tracer provider.
func TracerShutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

// KeyValue creates an attribute key-value pair for use in spans and events.
func KeyValue(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}
