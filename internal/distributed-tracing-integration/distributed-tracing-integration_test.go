// Copyright 2025 James Ross
package distributedtracing

import (
	"context"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
)

func TestTracingIntegration_New(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "localhost:4317",
				Environment:      "test",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	logger := zap.NewNop()

	ti, err := New(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	// Cleanup
	if ti != nil {
		err = ti.Shutdown(context.Background())
		assert.NoError(t, err)
	}
}

func TestTracingIntegration_EnqueueWithTracing(t *testing.T) {
	// Create in-memory span exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create tracer provider with in-memory exporter
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "localhost:4317",
				Environment:      "test",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	logger := zap.NewNop()

	ti := &TracingIntegration{
		config:   cfg,
		provider: nil, // Using global provider for this test
		tracer:   otel.Tracer("test"),
		meter:    otel.Meter("test"),
		logger:   logger,
	}

	// Initialize metrics
	err := ti.initMetrics()
	require.NoError(t, err)

	job := queue.Job{
		ID:           "test-job-123",
		FilePath:     "/test/file.txt",
		FileSize:     1024,
		Priority:     "high",
		Retries:      0,
		CreationTime: time.Now(),
	}

	ctx := context.Background()
	ctx, err = ti.EnqueueWithTracing(ctx, &job, "test-queue")
	assert.NoError(t, err)

	// Force flush to ensure spans are exported
	tp.ForceFlush(context.Background())

	// Check that spans were created
	spans := exporter.GetSpans()
	assert.Greater(t, len(spans), 0)

	// Verify the enqueue span
	var enqueueSpan *tracetest.SpanStub
	for _, span := range spans {
		if span.Name == "queue.enqueue" {
			enqueueSpan = &span
			break
		}
	}

	require.NotNil(t, enqueueSpan, "Enqueue span not found")

	// Check span attributes
	attrs := enqueueSpan.Attributes
	hasQueueName := false
	hasPriority := false

	for _, attr := range attrs {
		if string(attr.Key) == "queue.name" && attr.Value.AsString() == "test-queue" {
			hasQueueName = true
		}
		if string(attr.Key) == "queue.priority" && attr.Value.AsString() == "high" {
			hasPriority = true
		}
	}

	assert.True(t, hasQueueName, "Queue name attribute not found")
	assert.True(t, hasPriority, "Priority attribute not found")

	// Check that trace ID was set on the job
	assert.NotEmpty(t, job.TraceID)
	assert.NotEmpty(t, job.SpanID)
}

func TestTracingIntegration_ProcessWithTracing(t *testing.T) {
	// Create in-memory span exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create tracer provider with in-memory exporter
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "localhost:4317",
				Environment:      "test",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	logger := zap.NewNop()

	ti := &TracingIntegration{
		config:   cfg,
		provider: nil,
		tracer:   otel.Tracer("test"),
		meter:    otel.Meter("test"),
		logger:   logger,
	}

	// Initialize metrics
	err := ti.initMetrics()
	require.NoError(t, err)

	job := queue.Job{
		ID:           "test-job-456",
		FilePath:     "/test/file.txt",
		FileSize:     2048,
		Priority:     "low",
		Retries:      0,
		CreationTime: time.Now(),
		TraceID:      "0123456789abcdef0123456789abcdef",
		SpanID:       "0123456789abcdef",
	}

	// Test successful processing
	t.Run("SuccessfulProcessing", func(t *testing.T) {
		exporter.Reset()

		handler := func(ctx context.Context, j queue.Job) error {
			// Simulate some processing
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		ctx := context.Background()
		err := ti.ProcessWithTracing(ctx, job, handler)
		assert.NoError(t, err)

		// Force flush to ensure spans are exported
		tp.ForceFlush(context.Background())

		// Check that spans were created
		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		// Verify the process span
		var processSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "job.process" {
				processSpan = &span
				break
			}
		}

		require.NotNil(t, processSpan, "Process span not found")

		// Check that the span has the correct parent context
		assert.Equal(t, job.TraceID, processSpan.SpanContext.TraceID().String())

		// Check span events
		hasStartedEvent := false
		hasCompletedEvent := false

		for _, event := range processSpan.Events {
			if event.Name == "job.processing.started" {
				hasStartedEvent = true
			}
			if event.Name == "job.processing.completed" {
				hasCompletedEvent = true
			}
		}

		assert.True(t, hasStartedEvent, "Processing started event not found")
		assert.True(t, hasCompletedEvent, "Processing completed event not found")
	})

	// Test failed processing
	t.Run("FailedProcessing", func(t *testing.T) {
		exporter.Reset()

		handler := func(ctx context.Context, j queue.Job) error {
			return assert.AnError
		}

		ctx := context.Background()
		err := ti.ProcessWithTracing(ctx, job, handler)
		assert.Error(t, err)

		// Force flush to ensure spans are exported
		tp.ForceFlush(context.Background())

		// Check that spans were created
		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		// Verify the process span
		var processSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "job.process" {
				processSpan = &span
				break
			}
		}

		require.NotNil(t, processSpan, "Process span not found")

		// Check span events
		hasStartedEvent := false
		hasFailedEvent := false

		for _, event := range processSpan.Events {
			if event.Name == "job.processing.started" {
				hasStartedEvent = true
			}
			if event.Name == "job.processing.failed" {
				hasFailedEvent = true
			}
		}

		assert.True(t, hasStartedEvent, "Processing started event not found")
		assert.True(t, hasFailedEvent, "Processing failed event not found")

		// Check that error was recorded
		assert.Len(t, processSpan.Events, 2) // started and failed events
	})
}

func TestTracingIntegration_InstrumentAdminOperation(t *testing.T) {
	// Create in-memory span exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	// Create tracer provider with in-memory exporter
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			Tracing: config.TracingConfig{
				Enabled:          true,
				Endpoint:         "localhost:4317",
				Environment:      "test",
				SamplingStrategy: "always",
				SamplingRate:     1.0,
			},
		},
	}

	logger := zap.NewNop()

	ti := &TracingIntegration{
		config:   cfg,
		provider: nil,
		tracer:   otel.Tracer("test"),
		meter:    otel.Meter("test"),
		logger:   logger,
	}

	t.Run("SuccessfulOperation", func(t *testing.T) {
		exporter.Reset()

		operation := func(ctx context.Context) error {
			// Simulate admin operation
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		ctx := context.Background()
		err := ti.InstrumentAdminOperation(ctx, "purge_queue", operation)
		assert.NoError(t, err)

		// Force flush to ensure spans are exported
		tp.ForceFlush(context.Background())

		// Check that spans were created
		spans := exporter.GetSpans()
		assert.Greater(t, len(spans), 0)

		// Verify the admin span
		var adminSpan *tracetest.SpanStub
		for _, span := range spans {
			if span.Name == "admin.purge_queue" {
				adminSpan = &span
				break
			}
		}

		require.NotNil(t, adminSpan, "Admin span not found")

		// Check span attributes
		hasOperation := false
		hasSuccess := false

		for _, attr := range adminSpan.Attributes {
			if string(attr.Key) == "admin.operation" && attr.Value.AsString() == "purge_queue" {
				hasOperation = true
			}
			if string(attr.Key) == "operation.success" && attr.Value.AsBool() {
				hasSuccess = true
			}
		}

		assert.True(t, hasOperation, "Operation attribute not found")
		assert.True(t, hasSuccess, "Success attribute not found")
	})
}

func TestTracingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TracingConfig
		wantErr error
	}{
		{
			name: "Valid config",
			config: &TracingConfig{
				SamplingRate:      0.5,
				SamplingStrategy:  "probabilistic",
				PropagationFormat: "w3c",
			},
			wantErr: nil,
		},
		{
			name: "Invalid sampling rate too low",
			config: &TracingConfig{
				SamplingRate:      -0.1,
				SamplingStrategy:  "probabilistic",
				PropagationFormat: "w3c",
			},
			wantErr: ErrInvalidSamplingRate,
		},
		{
			name: "Invalid sampling rate too high",
			config: &TracingConfig{
				SamplingRate:      1.1,
				SamplingStrategy:  "probabilistic",
				PropagationFormat: "w3c",
			},
			wantErr: ErrInvalidSamplingRate,
		},
		{
			name: "Invalid sampling strategy",
			config: &TracingConfig{
				SamplingRate:      0.5,
				SamplingStrategy:  "invalid",
				PropagationFormat: "w3c",
			},
			wantErr: ErrInvalidSamplingStrategy,
		},
		{
			name: "Invalid propagation format",
			config: &TracingConfig{
				SamplingRate:      0.5,
				SamplingStrategy:  "probabilistic",
				PropagationFormat: "invalid",
			},
			wantErr: ErrInvalidPropagationFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}