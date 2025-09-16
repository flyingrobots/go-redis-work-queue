// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// TraceInfo represents trace information extracted from a job or context
type TraceInfo struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
	Sampled bool   `json:"sampled"`
}

// TraceAction represents an action available for a trace
type TraceAction struct {
	Type        string `json:"type"`        // "open", "copy", "view"
	Label       string `json:"label"`       // Human-readable label
	URL         string `json:"url"`         // URL for external trace viewer
	Command     string `json:"command"`     // Command to execute
	Description string `json:"description"` // Action description
}

// SpanAttributes represents structured span attributes for consistent tagging
type SpanAttributes struct {
	QueueName      string                 `json:"queue_name,omitempty"`
	QueueOperation string                 `json:"queue_operation,omitempty"`
	JobID          string                 `json:"job_id,omitempty"`
	JobFilePath    string                 `json:"job_filepath,omitempty"`
	JobFileSize    int64                  `json:"job_filesize,omitempty"`
	JobPriority    string                 `json:"job_priority,omitempty"`
	JobRetries     int                    `json:"job_retries,omitempty"`
	WorkerID       string                 `json:"worker_id,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time,omitempty"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// ToAttributes converts SpanAttributes to OpenTelemetry attributes
func (sa *SpanAttributes) ToAttributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	if sa.QueueName != "" {
		attrs = append(attrs, attribute.String("queue.name", sa.QueueName))
	}
	if sa.QueueOperation != "" {
		attrs = append(attrs, attribute.String("queue.operation", sa.QueueOperation))
	}
	if sa.JobID != "" {
		attrs = append(attrs, attribute.String("job.id", sa.JobID))
	}
	if sa.JobFilePath != "" {
		attrs = append(attrs, attribute.String("job.filepath", sa.JobFilePath))
	}
	if sa.JobFileSize > 0 {
		attrs = append(attrs, attribute.Int64("job.filesize", sa.JobFileSize))
	}
	if sa.JobPriority != "" {
		attrs = append(attrs, attribute.String("job.priority", sa.JobPriority))
	}
	if sa.JobRetries > 0 {
		attrs = append(attrs, attribute.Int("job.retries", sa.JobRetries))
	}
	if sa.WorkerID != "" {
		attrs = append(attrs, attribute.String("worker.id", sa.WorkerID))
	}
	if sa.ProcessingTime > 0 {
		attrs = append(attrs, attribute.Int64("processing.time_ms", sa.ProcessingTime.Milliseconds()))
	}

	// Add custom attributes
	for k, v := range sa.Custom {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	return attrs
}

// TraceableJob represents a job with tracing information
type TraceableJob struct {
	ID           string    `json:"id"`
	FilePath     string    `json:"filepath"`
	FileSize     int64     `json:"filesize"`
	Priority     string    `json:"priority"`
	Retries      int       `json:"retries"`
	CreationTime string    `json:"creation_time"`
	TraceID      string    `json:"trace_id"`
	SpanID       string    `json:"span_id"`
	TraceInfo    TraceInfo `json:"trace_info,omitempty"`
}