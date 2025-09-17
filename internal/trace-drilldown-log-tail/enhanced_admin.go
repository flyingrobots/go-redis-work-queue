// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/distributed-tracing-integration"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EnhancedAdmin provides admin functionality with trace integration
type EnhancedAdmin struct {
	tracingIntegration *distributed_tracing_integration.TracingIntegration
	traceManager       *TraceManager
	logger             *zap.Logger
}

// NewEnhancedAdmin creates a new enhanced admin
func NewEnhancedAdmin(tracingIntegration *distributed_tracing_integration.TracingIntegration, traceManager *TraceManager, logger *zap.Logger) *EnhancedAdmin {
	return &EnhancedAdmin{
		tracingIntegration: tracingIntegration,
		traceManager:       traceManager,
		logger:             logger,
	}
}

// EnhancedPeekResult extends PeekResult with trace information
type EnhancedPeekResult struct {
	Queue         string                                                   `json:"queue"`
	Items         []string                                                 `json:"items"`
	JobsWithTrace []JobWithTraceInfo                                       `json:"jobs_with_trace"`
	TraceActions  map[string][]distributed_tracing_integration.TraceAction `json:"trace_actions"`
	Summary       EnhancedPeekSummary                                      `json:"summary"`
}

// EnhancedPeekSummary provides summary information
type EnhancedPeekSummary struct {
	TotalJobs      int `json:"total_jobs"`
	JobsWithTraces int `json:"jobs_with_traces"`
	UniqueTraces   int `json:"unique_traces"`
}

// JobWithTraceInfo represents a job with trace information
type JobWithTraceInfo struct {
	JobID        string                                        `json:"job_id"`
	FilePath     string                                        `json:"file_path"`
	Priority     string                                        `json:"priority"`
	Retries      int                                           `json:"retries"`
	CreationTime string                                        `json:"creation_time"`
	TraceID      string                                        `json:"trace_id"`
	SpanID       string                                        `json:"span_id"`
	TraceInfo    *TraceInfo                                    `json:"trace_info,omitempty"`
	TraceActions []distributed_tracing_integration.TraceAction `json:"trace_actions,omitempty"`
	RawJobData   string                                        `json:"raw_job_data"`
}

// EnhancedPeek performs a peek operation with trace information
func (ea *EnhancedAdmin) EnhancedPeek(ctx context.Context, cfg *config.Config, rdb *redis.Client, queueAlias string, n int64) (*EnhancedPeekResult, error) {
	// First, perform the standard peek
	standardPeek, err := admin.Peek(ctx, cfg, rdb, queueAlias, n)
	if err != nil {
		return nil, fmt.Errorf("failed to peek queue: %w", err)
	}

	result := &EnhancedPeekResult{
		Queue:         standardPeek.Queue,
		Items:         standardPeek.Items,
		JobsWithTrace: make([]JobWithTraceInfo, 0),
		TraceActions:  make(map[string][]distributed_tracing_integration.TraceAction),
		Summary: EnhancedPeekSummary{
			TotalJobs: len(standardPeek.Items),
		},
	}

	uniqueTraces := make(map[string]bool)

	// Process each job to extract trace information
	for _, item := range standardPeek.Items {
		jobInfo, err := ea.parseJobWithTrace(item)
		if err != nil {
			ea.logger.Warn("Failed to parse job with trace", zap.Error(err), zap.String("job_data", item))
			continue
		}

		result.JobsWithTrace = append(result.JobsWithTrace, *jobInfo)

		// Track unique traces
		if jobInfo.TraceID != "" {
			uniqueTraces[jobInfo.TraceID] = true
			result.Summary.JobsWithTraces++

			// Generate trace actions
			if ea.tracingIntegration != nil {
				actions := distributed_tracing_integration.GenerateTraceActions(jobInfo.TraceID, ea.tracingIntegration.GetConfig())
				result.TraceActions[jobInfo.JobID] = actions
				jobInfo.TraceActions = actions
			}

			// Try to get additional trace information
			if ea.traceManager != nil {
				if traceInfo, err := ea.traceManager.GetTrace(jobInfo.TraceID); err == nil {
					jobInfo.TraceInfo = traceInfo
				}
			}
		}
	}

	result.Summary.UniqueTraces = len(uniqueTraces)

	return result, nil
}

// parseJobWithTrace parses a job JSON and extracts trace information
func (ea *EnhancedAdmin) parseJobWithTrace(jobData string) (*JobWithTraceInfo, error) {
	var jobMap map[string]interface{}
	if err := json.Unmarshal([]byte(jobData), &jobMap); err != nil {
		return nil, fmt.Errorf("failed to parse job JSON: %w", err)
	}

	jobInfo := &JobWithTraceInfo{
		RawJobData: jobData,
	}

	// Extract standard job fields
	if id, ok := jobMap["id"].(string); ok {
		jobInfo.JobID = id
	}
	if filepath, ok := jobMap["filepath"].(string); ok {
		jobInfo.FilePath = filepath
	}
	if priority, ok := jobMap["priority"].(string); ok {
		jobInfo.Priority = priority
	}
	if retries, ok := jobMap["retries"].(float64); ok {
		jobInfo.Retries = int(retries)
	}
	if creationTime, ok := jobMap["creation_time"].(string); ok {
		jobInfo.CreationTime = creationTime
	}

	// Extract trace fields
	if traceID, ok := jobMap["trace_id"].(string); ok {
		jobInfo.TraceID = traceID
	}
	if spanID, ok := jobMap["span_id"].(string); ok {
		jobInfo.SpanID = spanID
	}

	return jobInfo, nil
}

// GetJobTraceActions returns available trace actions for a job
func (ea *EnhancedAdmin) GetJobTraceActions(jobData string) ([]distributed_tracing_integration.TraceAction, error) {
	jobInfo, err := ea.parseJobWithTrace(jobData)
	if err != nil {
		return nil, err
	}

	if jobInfo.TraceID == "" {
		return nil, fmt.Errorf("no trace ID found in job")
	}

	if ea.tracingIntegration == nil {
		return nil, fmt.Errorf("tracing integration not configured")
	}

	actions := distributed_tracing_integration.GenerateTraceActions(jobInfo.TraceID, ea.tracingIntegration.GetConfig())
	return actions, nil
}

// OpenJobTrace opens a job's trace in the configured trace viewer
func (ea *EnhancedAdmin) OpenJobTrace(jobData string) (*TraceActionResult, error) {
	jobInfo, err := ea.parseJobWithTrace(jobData)
	if err != nil {
		return nil, err
	}

	if jobInfo.TraceID == "" {
		return nil, fmt.Errorf("no trace ID found in job")
	}

	if ea.tracingIntegration == nil {
		return nil, fmt.Errorf("tracing integration not configured")
	}

	traceURL := ea.tracingIntegration.GetTraceURL(jobInfo.TraceID)
	if traceURL == "" {
		return nil, fmt.Errorf("no trace URL configured")
	}

	return &TraceActionResult{
		JobID:        jobInfo.JobID,
		TraceID:      jobInfo.TraceID,
		Action:       "open",
		URL:          traceURL,
		Success:      true,
		Message:      "Trace URL generated successfully",
		Instructions: "Open this URL in your browser to view the trace",
	}, nil
}

// SearchJobsByTrace searches for jobs by trace ID
func (ea *EnhancedAdmin) SearchJobsByTrace(ctx context.Context, cfg *config.Config, rdb *redis.Client, traceID string) (*TraceJobSearchResult, error) {
	result := &TraceJobSearchResult{
		TraceID: traceID,
		Jobs:    make([]JobWithTraceInfo, 0),
	}

	// Search through all queues for jobs with this trace ID
	queues := []string{}
	for _, queue := range cfg.Worker.Queues {
		queues = append(queues, queue)
	}
	queues = append(queues, cfg.Worker.CompletedList, cfg.Worker.DeadLetterList)

	// Search processing lists
	processingKeys, err := rdb.Keys(ctx, "jobqueue:worker:*:processing").Result()
	if err == nil {
		queues = append(queues, processingKeys...)
	}

	for _, queue := range queues {
		// Get all items from this queue
		items, err := rdb.LRange(ctx, queue, 0, -1).Result()
		if err != nil {
			continue
		}

		for _, item := range items {
			jobInfo, err := ea.parseJobWithTrace(item)
			if err != nil {
				continue
			}

			if jobInfo.TraceID == traceID {
				jobInfo.TraceInfo = &TraceInfo{} // Add queue context
				jobInfo.TraceInfo.Tags = map[string]string{"found_in_queue": queue}
				result.Jobs = append(result.Jobs, *jobInfo)
			}
		}
	}

	result.TotalFound = len(result.Jobs)
	return result, nil
}

// GetTraceTimeline gets a timeline of jobs for a specific trace
func (ea *EnhancedAdmin) GetTraceTimeline(ctx context.Context, cfg *config.Config, rdb *redis.Client, traceID string) (*TraceTimeline, error) {
	searchResult, err := ea.SearchJobsByTrace(ctx, cfg, rdb, traceID)
	if err != nil {
		return nil, err
	}

	timeline := &TraceTimeline{
		TraceID: traceID,
		Events:  make([]TraceTimelineEvent, 0),
	}

	for _, job := range searchResult.Jobs {
		event := TraceTimelineEvent{
			Timestamp:   job.CreationTime,
			EventType:   "job_created",
			JobID:       job.JobID,
			Description: fmt.Sprintf("Job %s created in queue", job.JobID),
			Metadata: map[string]interface{}{
				"file_path": job.FilePath,
				"priority":  job.Priority,
				"retries":   job.Retries,
			},
		}
		timeline.Events = append(timeline.Events, event)
	}

	timeline.TotalEvents = len(timeline.Events)
	return timeline, nil
}

// TraceActionResult represents the result of a trace action
type TraceActionResult struct {
	JobID        string `json:"job_id"`
	TraceID      string `json:"trace_id"`
	Action       string `json:"action"`
	URL          string `json:"url,omitempty"`
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Instructions string `json:"instructions,omitempty"`
}

// TraceJobSearchResult represents search results for jobs by trace ID
type TraceJobSearchResult struct {
	TraceID    string             `json:"trace_id"`
	TotalFound int                `json:"total_found"`
	Jobs       []JobWithTraceInfo `json:"jobs"`
}

// TraceTimeline represents a timeline of events for a trace
type TraceTimeline struct {
	TraceID     string               `json:"trace_id"`
	TotalEvents int                  `json:"total_events"`
	Events      []TraceTimelineEvent `json:"events"`
}

// TraceTimelineEvent represents an event in a trace timeline
type TraceTimelineEvent struct {
	Timestamp   string                 `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	JobID       string                 `json:"job_id,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FormatEnhancedPeekForTUI formats enhanced peek results for TUI display
func (ea *EnhancedAdmin) FormatEnhancedPeekForTUI(result *EnhancedPeekResult) []string {
	formatted := make([]string, 0, len(result.JobsWithTrace))

	for _, job := range result.JobsWithTrace {
		line := fmt.Sprintf("Job %s: %s (priority: %s, retries: %d)",
			job.JobID, job.FilePath, job.Priority, job.Retries)

		if job.TraceID != "" {
			// Show first 8 characters of trace ID for brevity
			shortTraceID := job.TraceID
			if len(shortTraceID) > 8 {
				shortTraceID = shortTraceID[:8] + "..."
			}
			line += fmt.Sprintf(" [Trace: %s]", shortTraceID)

			// Add action hint
			if len(job.TraceActions) > 0 {
				line += " (Press 't' to open trace)"
			}
		}

		formatted = append(formatted, line)
	}

	// Add summary information
	if result.Summary.TotalJobs > 0 {
		summaryLine := fmt.Sprintf("Summary: %d jobs, %d with traces (%d unique traces)",
			result.Summary.TotalJobs, result.Summary.JobsWithTraces, result.Summary.UniqueTraces)
		formatted = append(formatted, "", summaryLine)
	}

	return formatted
}
