// Copyright 2025 James Ross
package obs

import (
	"context"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// MaybeInitTracing optionally initializes a global tracer provider.
func MaybeInitTracing(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Observability.Tracing.Enabled || cfg.Observability.Tracing.Endpoint == "" {
		return nil, nil
	}
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.Observability.Tracing.Endpoint),
		otlptracehttp.WithInsecure(),
	))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-redis-work-queue"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// ContextWithJobSpan creates a span for processing a job, attempting to honor
// job-provided TraceID/SpanID as a remote parent when present and valid.
func ContextWithJobSpan(ctx context.Context, job queue.Job) (context.Context, trace.Span) {
	tracer := otel.Tracer("worker")
	// Try to parse provided IDs
	if tid, err := trace.TraceIDFromHex(job.TraceID); err == nil {
		if sid, err2 := trace.SpanIDFromHex(job.SpanID); err2 == nil {
			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    tid,
				SpanID:     sid,
				TraceFlags: 0,
				Remote:     true,
			})
			ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}
	ctx, span := tracer.Start(ctx, "process_job")
	return ctx, span
}
