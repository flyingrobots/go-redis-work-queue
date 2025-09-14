// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"go.opentelemetry.io/otel/trace"
)

// TracingIntegration provides high-level distributed tracing integration
type TracingIntegration struct {
	config TracingUIConfig
}

// New creates a new TracingIntegration instance
func New(config TracingUIConfig) *TracingIntegration {
	return &TracingIntegration{
		config: config,
	}
}

// NewWithDefaults creates a TracingIntegration with default configuration
func NewWithDefaults() *TracingIntegration {
	return &TracingIntegration{
		config: DefaultTracingUIConfig(),
	}
}

// EnhancedPeekResult extends the basic PeekResult with trace information
type EnhancedPeekResult struct {
	Queue        string                     `json:"queue"`
	Items        []string                   `json:"items"`
	TraceJobs    []*TraceableJob            `json:"trace_jobs,omitempty"`
	TraceActions map[string][]TraceAction   `json:"trace_actions,omitempty"`
}

// EnhancePeekWithTracing enhances a peek result with tracing information
func (t *TracingIntegration) EnhancePeekWithTracing(queue string, items []string) (*EnhancedPeekResult, error) {
	result := &EnhancedPeekResult{
		Queue:        queue,
		Items:        items,
		TraceJobs:    make([]*TraceableJob, 0),
		TraceActions: make(map[string][]TraceAction),
	}

	// Parse jobs and extract trace information
	for _, item := range items {
		job, err := ParseJobWithTrace(item)
		if err != nil {
			// Skip items that can't be parsed, but continue processing
			continue
		}

		result.TraceJobs = append(result.TraceJobs, job)

		// Generate trace actions for jobs with trace information
		if job.TraceID != "" {
			actions := GenerateTraceActions(job.TraceID, t.config)
			result.TraceActions[job.ID] = actions
		}
	}

	return result, nil
}

// GetTraceURL returns the trace URL for a given trace ID
func (t *TracingIntegration) GetTraceURL(traceID string) string {
	return t.config.GetTraceURL(traceID)
}

// GetConfig returns the tracing configuration
func (t *TracingIntegration) GetConfig() TracingUIConfig {
	return t.config
}

// GetCurrentTraceInfo extracts trace information from the current context
func (t *TracingIntegration) GetCurrentTraceInfo(ctx context.Context) *TraceInfo {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	sc := span.SpanContext()
	if !sc.IsValid() {
		return nil
	}

	return &TraceInfo{
		TraceID: sc.TraceID().String(),
		SpanID:  sc.SpanID().String(),
		Sampled: sc.IsSampled(),
	}
}

// InjectTraceIntoJobPayload injects trace context into a job payload
func (t *TracingIntegration) InjectTraceIntoJobPayload(ctx context.Context, jobPayload map[string]interface{}) error {
	traceInfo := t.GetCurrentTraceInfo(ctx)
	if traceInfo == nil {
		// No active trace, set empty values
		jobPayload["trace_id"] = ""
		jobPayload["span_id"] = ""
		return nil
	}

	jobPayload["trace_id"] = traceInfo.TraceID
	jobPayload["span_id"] = traceInfo.SpanID
	return nil
}

// CreateJobWithTracing creates a job payload with tracing information
func (t *TracingIntegration) CreateJobWithTracing(ctx context.Context, jobID, filePath string, fileSize int64, priority string) (string, error) {
	jobPayload := map[string]interface{}{
		"id":            jobID,
		"filepath":      filePath,
		"filesize":      fileSize,
		"priority":      priority,
		"retries":       0,
		"creation_time": time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Inject trace information
	if err := t.InjectTraceIntoJobPayload(ctx, jobPayload); err != nil {
		return "", fmt.Errorf("failed to inject trace into job payload: %w", err)
	}

	// Convert to JSON
	jobJSON, err := json.Marshal(jobPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job payload: %w", err)
	}

	return string(jobJSON), nil
}

// ValidateTracingSetup checks if tracing is properly configured
func (t *TracingIntegration) ValidateTracingSetup(cfg *config.Config) error {
	if !cfg.Observability.Tracing.Enabled {
		return ErrTracingDisabled
	}

	if cfg.Observability.Tracing.Endpoint == "" {
		return fmt.Errorf("tracing endpoint not configured")
	}

	// Check if at least one trace viewer is configured
	if t.config.JaegerBaseURL == "" && t.config.ZipkinBaseURL == "" && t.config.CustomTraceURL == "" {
		return ErrTracingUINotConfigured
	}

	return nil
}

// FormatJobsForTUIDisplay formats multiple jobs for TUI display with trace information
func (t *TracingIntegration) FormatJobsForTUIDisplay(jobs []*TraceableJob) []string {
	formatted := make([]string, len(jobs))

	for i, job := range jobs {
		base := fmt.Sprintf("Job %s: %s (size: %d, priority: %s, retries: %d)",
			job.ID, job.FilePath, job.FileSize, job.Priority, job.Retries)

		if job.TraceID != "" {
			base += fmt.Sprintf(" [Trace: %s]", job.TraceID[:8]) // Show first 8 chars of trace ID
		}

		formatted[i] = base
	}

	return formatted
}