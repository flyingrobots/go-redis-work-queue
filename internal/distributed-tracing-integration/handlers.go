// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"encoding/json"
	"fmt"
)

// ParseJobWithTrace parses a job JSON and extracts trace information
func ParseJobWithTrace(jobJSON string) (*TraceableJob, error) {
	var job TraceableJob
	if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
		return nil, fmt.Errorf("failed to parse job JSON: %w", err)
	}

	// Extract trace info from the job
	job.TraceInfo = TraceInfo{
		TraceID: job.TraceID,
		SpanID:  job.SpanID,
		Sampled: job.TraceID != "" && job.SpanID != "",
	}

	return &job, nil
}

// GenerateTraceActions creates available actions for a trace
func GenerateTraceActions(traceID string, config TracingUIConfig) []TraceAction {
	var actions []TraceAction

	if traceID == "" {
		return actions
	}

	// Copy trace ID action
	if config.EnableCopyActions {
		actions = append(actions, TraceAction{
			Type:        "copy",
			Label:       "Copy Trace ID",
			Command:     fmt.Sprintf("echo '%s' | pbcopy", traceID),
			Description: "Copy trace ID to clipboard",
		})
	}

	// Open trace in browser action
	if config.EnableOpenActions {
		traceURL := config.GetTraceURL(traceID)
		if traceURL != "" {
			actions = append(actions, TraceAction{
				Type:        "open",
				Label:       "Open Trace",
				URL:         traceURL,
				Command:     fmt.Sprintf("open '%s'", traceURL),
				Description: fmt.Sprintf("Open trace in %s", config.DefaultTraceViewer),
			})
		}
	}

	// View trace ID action (always available)
	actions = append(actions, TraceAction{
		Type:        "view",
		Label:       "View Trace ID",
		Description: fmt.Sprintf("Trace ID: %s", traceID),
	})

	return actions
}

// FormatTraceForDisplay formats trace information for display in TUI
func FormatTraceForDisplay(job *TraceableJob, config TracingUIConfig) string {
	if job.TraceID == "" {
		return "No trace information available"
	}

	output := fmt.Sprintf("Trace ID: %s", job.TraceID)
	if job.SpanID != "" {
		output += fmt.Sprintf("\nSpan ID: %s", job.SpanID)
	}

	// Add trace URL if available
	traceURL := config.GetTraceURL(job.TraceID)
	if traceURL != "" {
		output += fmt.Sprintf("\nTrace URL: %s", traceURL)
	}

	return output
}

// ExtractTraceInfoFromJobs processes multiple job JSONs and extracts trace information
func ExtractTraceInfoFromJobs(jobJSONs []string) ([]*TraceableJob, error) {
	jobs := make([]*TraceableJob, 0, len(jobJSONs))

	for _, jobJSON := range jobJSONs {
		job, err := ParseJobWithTrace(jobJSON)
		if err != nil {
			// Continue processing other jobs, but log the error
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}