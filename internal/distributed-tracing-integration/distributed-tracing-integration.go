// Copyright 2025 James Ross
// Package distributedtracing provides comprehensive distributed tracing integration
// for the go-redis-work-queue system using OpenTelemetry.
package distributedtracing

import (
	"context"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingIntegration manages distributed tracing for the work queue system.
type TracingIntegration struct {
	config   *config.Config
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	meter    metric.Meter
	logger   *zap.Logger

	// Metrics with exemplars
	enqueueDuration metric.Float64Histogram
	dequeueDuration metric.Float64Histogram
	processDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter
}

// New creates a new TracingIntegration instance.
func New(cfg *config.Config, logger *zap.Logger) (*TracingIntegration, error) {
	// Initialize tracing provider
	provider, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		return nil, err
	}

	// Get tracer and meter
	tracer := otel.Tracer("distributed-tracing-integration")
	meter := otel.Meter("distributed-tracing-integration")

	ti := &TracingIntegration{
		config:   cfg,
		provider: provider,
		tracer:   tracer,
		meter:    meter,
		logger:   logger,
	}

	// Initialize metrics with exemplar support
	if err := ti.initMetrics(); err != nil {
		return nil, err
	}

	return ti, nil
}

// initMetrics initializes metrics that will have trace exemplars attached.
func (ti *TracingIntegration) initMetrics() error {
	var err error

	ti.enqueueDuration, err = ti.meter.Float64Histogram(
		"queue.enqueue.duration",
		metric.WithDescription("Duration of enqueue operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	ti.dequeueDuration, err = ti.meter.Float64Histogram(
		"queue.dequeue.duration",
		metric.WithDescription("Duration of dequeue operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	ti.processDuration, err = ti.meter.Float64Histogram(
		"job.process.duration",
		metric.WithDescription("Duration of job processing"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	ti.errorCounter, err = ti.meter.Int64Counter(
		"job.errors",
		metric.WithDescription("Number of job processing errors"),
	)
	if err != nil {
		return err
	}

	return nil
}

// EnqueueWithTracing wraps job enqueue with tracing.
func (ti *TracingIntegration) EnqueueWithTracing(ctx context.Context, job *queue.Job, queueName string) (context.Context, error) {
	// Start enqueue span
	ctx, span := obs.StartEnqueueSpan(ctx, queueName, job.Priority)
	defer span.End()

	startTime := time.Now()

	// Add job attributes to span
	span.SetAttributes(
		attribute.String("job.id", job.ID),
		attribute.String("job.filepath", job.FilePath),
		attribute.Int64("job.filesize", job.FileSize),
		attribute.Int("job.retries", job.Retries),
	)

	// Extract trace context and inject into job
	traceID, spanID := obs.GetTraceAndSpanID(ctx)
	if traceID != "" {
		job.TraceID = traceID
		job.SpanID = spanID
	}

	// Record enqueue event
	obs.AddEvent(ctx, "job.enqueued",
		attribute.String("queue", queueName),
		attribute.String("job_id", job.ID),
	)

	// Record metric with exemplar
	duration := float64(time.Since(startTime).Milliseconds())
	ti.enqueueDuration.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("queue", queueName),
			attribute.String("priority", job.Priority),
		),
	)

	obs.SetSpanSuccess(ctx)
	return ctx, nil
}

// DequeueWithTracing wraps job dequeue with tracing.
func (ti *TracingIntegration) DequeueWithTracing(ctx context.Context, queueName string) (context.Context, *queue.Job, error) {
	// Start dequeue span
	ctx, span := obs.StartDequeueSpan(ctx, queueName)
	defer span.End()

	startTime := time.Now()

	// Add queue depth attribute if available
	span.SetAttributes(
		attribute.String("queue.name", queueName),
	)

	// Record dequeue event
	obs.AddEvent(ctx, "job.dequeuing",
		attribute.String("queue", queueName),
	)

	// Simulated dequeue - in real implementation, this would call the actual dequeue
	// For now, return a placeholder
	job := &queue.Job{
		ID:       "test-job",
		FilePath: "/test/path",
		Priority: "high",
	}

	// Record metric with exemplar
	duration := float64(time.Since(startTime).Milliseconds())
	ti.dequeueDuration.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("queue", queueName),
		),
	)

	obs.AddEvent(ctx, "job.dequeued",
		attribute.String("queue", queueName),
		attribute.String("job_id", job.ID),
	)

	obs.SetSpanSuccess(ctx)
	return ctx, job, nil
}

// ProcessWithTracing wraps job processing with tracing.
func (ti *TracingIntegration) ProcessWithTracing(ctx context.Context, job queue.Job, handler func(context.Context, queue.Job) error) error {
	// Create processing span with parent context from job
	ctx, span := obs.ContextWithJobSpan(ctx, job)
	defer span.End()

	startTime := time.Now()

	// Add processing started event
	obs.AddEvent(ctx, "job.processing.started",
		attribute.String("job_id", job.ID),
	)

	// Execute the handler
	err := handler(ctx, job)

	// Record processing duration
	duration := float64(time.Since(startTime).Milliseconds())
	ti.processDuration.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("priority", job.Priority),
			attribute.Bool("success", err == nil),
		),
	)

	if err != nil {
		// Record error
		obs.RecordError(ctx, err)
		ti.errorCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("error_type", getErrorType(err)),
				attribute.String("priority", job.Priority),
			),
		)

		obs.AddEvent(ctx, "job.processing.failed",
			attribute.String("job_id", job.ID),
			attribute.String("error", err.Error()),
		)
	} else {
		obs.SetSpanSuccess(ctx)
		obs.AddEvent(ctx, "job.processing.completed",
			attribute.String("job_id", job.ID),
		)
	}

	// Add final attributes
	span.SetAttributes(
		attribute.Float64("processing.duration_ms", duration),
		attribute.Bool("processing.success", err == nil),
	)

	return err
}

// InstrumentAdminOperation wraps admin operations with tracing.
func (ti *TracingIntegration) InstrumentAdminOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	ctx, span := ti.tracer.Start(ctx, "admin."+operation,
		trace.WithAttributes(
			attribute.String("admin.operation", operation),
		),
	)
	defer span.End()

	startTime := time.Now()

	err := fn(ctx)

	span.SetAttributes(
		attribute.Float64("operation.duration_ms", float64(time.Since(startTime).Milliseconds())),
		attribute.Bool("operation.success", err == nil),
	)

	if err != nil {
		obs.RecordError(ctx, err)
	} else {
		obs.SetSpanSuccess(ctx)
	}

	return err
}

// ExtractContextFromMetadata extracts trace context from job metadata.
func (ti *TracingIntegration) ExtractContextFromMetadata(ctx context.Context, metadata map[string]string) context.Context {
	return obs.ExtractTraceContext(ctx, metadata)
}

// InjectContextToMetadata injects trace context into job metadata.
func (ti *TracingIntegration) InjectContextToMetadata(ctx context.Context, metadata map[string]string) {
	traceData := obs.InjectTraceContext(ctx)
	for k, v := range traceData {
		metadata[k] = v
	}
}

// Shutdown gracefully shuts down the tracing integration.
func (ti *TracingIntegration) Shutdown(ctx context.Context) error {
	if ti.provider != nil {
		return obs.TracerShutdown(ctx, ti.provider)
	}
	return nil
}

// getErrorType categorizes errors for metrics.
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}
	// Add error type detection logic here
	// For now, return a generic type
	return "generic"
}