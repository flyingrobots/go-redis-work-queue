// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

// E2E tests verify complete trace propagation from enqueue through processing
func TestEndToEnd_TracePropagatThroughWorkflow(t *testing.T) {
	endpoint := os.Getenv("OTEL_ENDPOINT")
	if endpoint == "" {
		t.Skip("Skipping E2E test: OTEL_ENDPOINT not set")
	}

	// Setup real OTLP exporter
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	require.NoError(t, err)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tp.ForceFlush(ctx)
		tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)
	tracer := otel.Tracer("e2e-workflow-test")

	ti := NewWithDefaults()

	t.Run("full_enqueue_to_process_workflow", func(t *testing.T) {
		// Step 1: Client initiates request (root span)
		rootCtx, rootSpan := tracer.Start(context.Background(), "client-request")
		rootSpan.SetAttributes(
			attribute.String("client.request_id", "req-12345"),
			attribute.String("operation", "file-processing-request"),
		)

		// Step 2: Enqueue operation
		enqueueCtx, enqueueSpan := tracer.Start(rootCtx, "enqueue-job")
		enqueueSpan.SetAttributes(
			attribute.String("queue.operation", "enqueue"),
			attribute.String("queue.name", "file-processing"),
		)

		// Create job with trace context
		jobPayload, err := ti.CreateJobWithTracing(enqueueCtx, "e2e-job-001", "/data/test-file.dat", 4096, "high")
		require.NoError(t, err)
		assert.NotEmpty(t, jobPayload)

		enqueueSpan.End()

		// Step 3: Simulate job being picked up by worker
		traceableJob, err := ParseJobWithTrace(jobPayload)
		require.NoError(t, err)
		assert.Equal(t, "e2e-job-001", traceableJob.ID)
		assert.NotEmpty(t, traceableJob.TraceID)
		assert.NotEmpty(t, traceableJob.SpanID)

		// Step 4: Worker processing (child span)
		processCtx, processSpan := tracer.Start(rootCtx, "process-job")
		processSpan.SetAttributes(
			attribute.String("queue.operation", "process"),
			attribute.String("job.id", traceableJob.ID),
			attribute.String("job.filepath", traceableJob.FilePath),
			attribute.Int64("job.filesize", traceableJob.FileSize),
			attribute.String("job.priority", traceableJob.Priority),
			attribute.String("worker.id", "worker-e2e-001"),
		)

		// Simulate processing work
		time.Sleep(50 * time.Millisecond)

		// Record processing events
		processSpan.AddEvent("job.started",
			attribute.String("job.id", traceableJob.ID),
		)

		time.Sleep(25 * time.Millisecond)

		processSpan.AddEvent("job.completed", trace.WithAttributes(
			attribute.String("job.id", traceableJob.ID),
			attribute.Int64("processing.duration_ms", 75),
		))

		processSpan.End()

		// Step 5: Complete workflow
		rootSpan.SetAttributes(
			attribute.String("workflow.status", "completed"),
			attribute.Int64("workflow.total_jobs", 1),
		)
		rootSpan.End()

		// Step 6: Verify trace continuity
		assert.Equal(t, traceableJob.TraceID, rootSpan.SpanContext().TraceID().String())

		// Step 7: Test TUI integration
		enhancedResult, err := ti.EnhancePeekWithTracing("file-processing", []string{jobPayload})
		require.NoError(t, err)
		assert.Len(t, enhancedResult.TraceJobs, 1)
		assert.Contains(t, enhancedResult.TraceActions, traceableJob.ID)

		// Verify trace actions
		actions := enhancedResult.TraceActions[traceableJob.ID]
		assert.NotEmpty(t, actions)

		var hasOpenAction, hasCopyAction bool
		for _, action := range actions {
			switch action.Type {
			case "open":
				hasOpenAction = true
				assert.Contains(t, action.URL, traceableJob.TraceID)
			case "copy":
				hasCopyAction = true
				assert.Contains(t, action.Command, traceableJob.TraceID)
			}
		}
		assert.True(t, hasOpenAction)
		assert.True(t, hasCopyAction)
	})

	t.Run("batch_processing_workflow", func(t *testing.T) {
		// Test processing multiple jobs in a batch with shared parent trace
		batchCtx, batchSpan := tracer.Start(context.Background(), "batch-processing")
		batchSpan.SetAttributes(
			attribute.String("batch.id", "batch-001"),
			attribute.Int("batch.size", 3),
		)

		var jobs []*TraceableJob
		var jobPayloads []string

		// Create 3 jobs in the same batch
		for i := 0; i < 3; i++ {
			jobCtx, jobSpan := tracer.Start(batchCtx, "create-batch-job")
			jobSpan.SetAttributes(
				attribute.Int("batch.job_index", i),
			)

			jobID := fmt.Sprintf("batch-job-%03d", i)
			filePath := fmt.Sprintf("/data/batch/file-%03d.dat", i)
			fileSize := int64(1024 * (i + 1))

			payload, err := ti.CreateJobWithTracing(jobCtx, jobID, filePath, fileSize, "batch")
			require.NoError(t, err)
			jobPayloads = append(jobPayloads, payload)

			job, err := ParseJobWithTrace(payload)
			require.NoError(t, err)
			jobs = append(jobs, job)

			jobSpan.End()
		}

		batchSpan.End()

		// Verify all jobs share the same trace ID (from batch parent)
		expectedTraceID := jobs[0].TraceID
		for _, job := range jobs {
			assert.Equal(t, expectedTraceID, job.TraceID, "All batch jobs should share the same trace ID")
		}

		// Test enhanced peek with batch
		enhancedResult, err := ti.EnhancePeekWithTracing("batch-queue", jobPayloads)
		require.NoError(t, err)
		assert.Len(t, enhancedResult.TraceJobs, 3)
		assert.Len(t, enhancedResult.TraceActions, 3)

		// Verify TUI formatting
		formatted := ti.FormatJobsForTUIDisplay(jobs)
		require.Len(t, formatted, 3)

		for i, formattedJob := range formatted {
			assert.Contains(t, formattedJob, fmt.Sprintf("batch-job-%03d", i))
			assert.Contains(t, formattedJob, "priority: batch")
			// All should show the same trace ID prefix (first 8 chars)
			assert.Contains(t, formattedJob, fmt.Sprintf("[Trace: %s]", expectedTraceID[:8]))
		}
	})

	t.Run("error_handling_with_traces", func(t *testing.T) {
		// Test error scenarios while maintaining trace context
		errorCtx, errorSpan := tracer.Start(context.Background(), "error-handling-test")
		errorSpan.SetAttributes(
			attribute.String("test.scenario", "error-handling"),
		)

		// Create a job that will encounter processing errors
		jobPayload, err := ti.CreateJobWithTracing(errorCtx, "error-job", "/invalid/path.dat", 0, "high")
		require.NoError(t, err)

		job, err := ParseJobWithTrace(jobPayload)
		require.NoError(t, err)

		// Simulate processing with errors
		processCtx, processSpan := tracer.Start(errorCtx, "process-with-errors")
		processSpan.SetAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.filepath", job.FilePath),
		)

		// Record error events
		processSpan.AddEvent("error.file_not_found", trace.WithAttributes(
			attribute.String("error.path", job.FilePath),
			attribute.String("error.type", "file_not_found"),
		))

		processSpan.SetAttributes(
			attribute.String("error.category", "file_access"),
			attribute.Bool("job.failed", true),
		)

		// Mark span as error (in real implementation, this would be done by error recording)
		processSpan.RecordError(fmt.Errorf("file not found: %s", job.FilePath))

		processSpan.End()
		errorSpan.End()

		// Verify trace information is still preserved
		assert.Equal(t, job.TraceID, errorSpan.SpanContext().TraceID().String())

		// Test that error jobs still work with TUI
		enhancedResult, err := ti.EnhancePeekWithTracing("error-queue", []string{jobPayload})
		require.NoError(t, err)
		assert.Len(t, enhancedResult.TraceJobs, 1)

		actions := enhancedResult.TraceActions[job.ID]
		assert.NotEmpty(t, actions)

		// Should still generate trace URL even for failed jobs
		traceURL := ti.GetTraceURL(job.TraceID)
		assert.Contains(t, traceURL, job.TraceID)
	})
}

func TestEndToEnd_TracingConfigurationValidation(t *testing.T) {
	endpoint := os.Getenv("OTEL_ENDPOINT")
	if endpoint == "" {
		t.Skip("Skipping E2E config test: OTEL_ENDPOINT not set")
	}

	tests := []struct {
		name           string
		tracingConfig  config.TracingConfig
		uiConfig       TracingUIConfig
		expectedValid  bool
		expectedErrors []error
	}{
		{
			name: "complete_valid_configuration",
			tracingConfig: config.TracingConfig{
				Enabled:  true,
				Endpoint: endpoint,
			},
			uiConfig: TracingUIConfig{
				JaegerBaseURL:      "http://localhost:16686",
				ZipkinBaseURL:      "http://localhost:9411",
				CustomTraceURL:     "https://traces.example.com/trace/{traceID}",
				EnableCopyActions:  true,
				EnableOpenActions:  true,
				DefaultTraceViewer: "jaeger",
			},
			expectedValid: true,
		},
		{
			name: "tracing_disabled",
			tracingConfig: config.TracingConfig{
				Enabled:  false,
				Endpoint: endpoint,
			},
			uiConfig:       DefaultTracingUIConfig(),
			expectedValid:  false,
			expectedErrors: []error{ErrTracingDisabled},
		},
		{
			name: "no_ui_configured",
			tracingConfig: config.TracingConfig{
				Enabled:  true,
				Endpoint: endpoint,
			},
			uiConfig: TracingUIConfig{
				JaegerBaseURL:      "",
				ZipkinBaseURL:      "",
				CustomTraceURL:     "",
				EnableCopyActions:  false,
				EnableOpenActions:  false,
				DefaultTraceViewer: "jaeger",
			},
			expectedValid:  false,
			expectedErrors: []error{ErrTracingUINotConfigured},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: tt.tracingConfig,
				},
			}

			ti := New(tt.uiConfig)
			err := ti.ValidateTracingSetup(cfg)

			if tt.expectedValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				for _, expectedErr := range tt.expectedErrors {
					assert.ErrorIs(t, err, expectedErr)
				}
			}
		})
	}
}

func TestEndToEnd_TraceMetadataPropagation(t *testing.T) {
	endpoint := os.Getenv("OTEL_ENDPOINT")
	if endpoint == "" {
		t.Skip("Skipping E2E metadata test: OTEL_ENDPOINT not set")
	}

	// Setup tracing
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	require.NoError(t, err)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tp.ForceFlush(ctx)
		tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)
	tracer := otel.Tracer("metadata-propagation-test")

	ti := NewWithDefaults()

	t.Run("metadata_injection_extraction_roundtrip", func(t *testing.T) {
		// Create initial trace context
		ctx, span := tracer.Start(context.Background(), "metadata-test")
		span.SetAttributes(
			attribute.String("test.phase", "metadata-propagation"),
		)

		originalTraceID := span.SpanContext().TraceID().String()
		originalSpanID := span.SpanContext().SpanID().String()

		// Simulate serializing trace context to job metadata
		metadata := make(map[string]interface{})
		err := ti.InjectTraceIntoJobPayload(ctx, metadata)
		require.NoError(t, err)

		assert.Equal(t, originalTraceID, metadata["trace_id"])
		assert.Equal(t, originalSpanID, metadata["span_id"])

		// Create job with metadata
		jobPayload, err := ti.CreateJobWithTracing(ctx, "metadata-job", "/test/metadata.dat", 2048, "metadata")
		require.NoError(t, err)

		// Parse job and verify trace information roundtrip
		job, err := ParseJobWithTrace(jobPayload)
		require.NoError(t, err)

		assert.Equal(t, originalTraceID, job.TraceID)
		assert.Equal(t, originalSpanID, job.SpanID)
		assert.Equal(t, originalTraceID, job.TraceInfo.TraceID)
		assert.Equal(t, originalSpanID, job.TraceInfo.SpanID)
		assert.True(t, job.TraceInfo.Sampled)

		span.End()

		// Verify trace URL generation works
		traceURL := ti.GetTraceURL(job.TraceID)
		assert.Contains(t, traceURL, originalTraceID)

		// Verify trace actions
		actions := GenerateTraceActions(job.TraceID, ti.config)
		assert.NotEmpty(t, actions)

		var foundOpenAction bool
		for _, action := range actions {
			if action.Type == "open" {
				foundOpenAction = true
				assert.Equal(t, traceURL, action.URL)
				break
			}
		}
		assert.True(t, foundOpenAction)
	})
}