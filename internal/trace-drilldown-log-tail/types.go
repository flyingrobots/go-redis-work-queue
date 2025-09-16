// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"time"
)

// TraceInfo represents trace information for a job or operation
type TraceInfo struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	ServiceName  string            `json:"service_name"`
	OperationName string           `json:"operation_name"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time,omitempty"`
	Duration     time.Duration     `json:"duration,omitempty"`
	Status       string            `json:"status"`
	Tags         map[string]string `json:"tags,omitempty"`
	Logs         []TraceLog        `json:"logs,omitempty"`
	Links        []TraceLink       `json:"links,omitempty"`
}

// TraceLog represents a log entry within a trace
type TraceLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// TraceLink represents external links for trace viewing
type TraceLink struct {
	Type        string `json:"type"`        // jaeger, zipkin, datadog, etc.
	URL         string `json:"url"`         // Full URL to trace
	DisplayName string `json:"display_name"`
}

// LogEntry represents a log line from the system
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	JobID       string                 `json:"job_id,omitempty"`
	WorkerID    string                 `json:"worker_id,omitempty"`
	QueueName   string                 `json:"queue_name,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// LogFilter defines filtering criteria for logs
type LogFilter struct {
	StartTime    time.Time         `json:"start_time,omitempty"`
	EndTime      time.Time         `json:"end_time,omitempty"`
	Levels       []string          `json:"levels,omitempty"`
	Sources      []string          `json:"sources,omitempty"`
	JobIDs       []string          `json:"job_ids,omitempty"`
	WorkerIDs    []string          `json:"worker_ids,omitempty"`
	QueueNames   []string          `json:"queue_names,omitempty"`
	TraceIDs     []string          `json:"trace_ids,omitempty"`
	SearchText   string            `json:"search_text,omitempty"`
	MaxResults   int               `json:"max_results,omitempty"`
	IncludeStack bool              `json:"include_stack,omitempty"`
}

// TailConfig defines configuration for log tailing
type TailConfig struct {
	Follow            bool          `json:"follow"`
	BufferSize        int           `json:"buffer_size"`
	MaxLinesPerSecond int           `json:"max_lines_per_second"`
	BackpressureLimit int           `json:"backpressure_limit"`
	FlushInterval     time.Duration `json:"flush_interval"`
	Filter            *LogFilter    `json:"filter,omitempty"`
}

// TracingConfig defines configuration for tracing integration
type TracingConfig struct {
	Enabled       bool              `json:"enabled"`
	Provider      string            `json:"provider"` // jaeger, zipkin, datadog, etc.
	Endpoint      string            `json:"endpoint"`
	ServiceName   string            `json:"service_name"`
	SamplingRate  float64           `json:"sampling_rate"`
	PropagateHeaders []string       `json:"propagate_headers"`
	URLTemplate   string            `json:"url_template"` // Template for external trace URLs
	AuthToken     string            `json:"auth_token,omitempty"`
	ExtraConfig   map[string]string `json:"extra_config,omitempty"`
}

// LoggingConfig defines configuration for log collection
type LoggingConfig struct {
	Enabled         bool              `json:"enabled"`
	Sources         []LogSource       `json:"sources"`
	RetentionPeriod time.Duration     `json:"retention_period"`
	MaxStorageSize  int64             `json:"max_storage_size"`
	IndexFields     []string          `json:"index_fields"`
	ParseFormats    []string          `json:"parse_formats"` // json, logfmt, syslog, etc.
}

// LogSource defines a source of logs
type LogSource struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // file, redis, kafka, etc.
	Config   map[string]string `json:"config"`
	Enabled  bool              `json:"enabled"`
}

// SpanSummary provides a summary of spans in a trace
type SpanSummary struct {
	TraceID      string        `json:"trace_id"`
	TotalSpans   int           `json:"total_spans"`
	Duration     time.Duration `json:"duration"`
	Services     []string      `json:"services"`
	ErrorCount   int           `json:"error_count"`
	WarningCount int           `json:"warning_count"`
	Operations   []Operation   `json:"operations"`
	Timeline     []TimelineEvent `json:"timeline"`
}

// Operation represents an operation within a trace
type Operation struct {
	Name      string        `json:"name"`
	Service   string        `json:"service"`
	Count     int           `json:"count"`
	Duration  time.Duration `json:"duration"`
	ErrorRate float64       `json:"error_rate"`
}

// TimelineEvent represents an event in the trace timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	SpanID      string    `json:"span_id"`
	Operation   string    `json:"operation"`
	Service     string    `json:"service"`
	Duration    time.Duration `json:"duration,omitempty"`
	EventType   string    `json:"event_type"` // start, end, log, error
	Description string    `json:"description,omitempty"`
}

// BackpressureStatus represents the current backpressure state
type BackpressureStatus struct {
	Active         bool      `json:"active"`
	BufferUsage    float64   `json:"buffer_usage"` // Percentage 0-100
	DroppedLines   int64     `json:"dropped_lines"`
	LastActivated  time.Time `json:"last_activated,omitempty"`
	CurrentRate    int       `json:"current_rate"` // Lines per second
	MaxRate        int       `json:"max_rate"`
}

// TailSession represents an active log tailing session
type TailSession struct {
	ID              string             `json:"id"`
	Config          TailConfig         `json:"config"`
	StartedAt       time.Time          `json:"started_at"`
	LinesProcessed  int64              `json:"lines_processed"`
	BackpressureStatus BackpressureStatus `json:"backpressure_status"`
	Connected       bool               `json:"connected"`
	LastActivity    time.Time          `json:"last_activity"`
}

// TraceContext carries trace information through the system
type TraceContext struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	Baggage      map[string]string `json:"baggage,omitempty"`
	Sampled      bool              `json:"sampled"`
}

// LogStreamEvent represents an event in a log stream
type LogStreamEvent struct {
	Type      string     `json:"type"` // log, error, status, backpressure
	Timestamp time.Time  `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// LogStats provides statistics about log collection
type LogStats struct {
	TotalLines      int64         `json:"total_lines"`
	LinesPerSecond  float64       `json:"lines_per_second"`
	ErrorCount      int64         `json:"error_count"`
	WarningCount    int64         `json:"warning_count"`
	UniqueTraces    int           `json:"unique_traces"`
	UniqueJobs      int           `json:"unique_jobs"`
	UniqueWorkers   int           `json:"unique_workers"`
	OldestEntry     time.Time     `json:"oldest_entry"`
	NewestEntry     time.Time     `json:"newest_entry"`
	StorageUsed     int64         `json:"storage_used"`
	LevelBreakdown  map[string]int64 `json:"level_breakdown"`
}

// TraceSearchResult represents search results for traces
type TraceSearchResult struct {
	Traces      []TraceInfo `json:"traces"`
	TotalCount  int         `json:"total_count"`
	HasMore     bool        `json:"has_more"`
	NextCursor  string      `json:"next_cursor,omitempty"`
}

// LogSearchResult represents search results for logs
type LogSearchResult struct {
	Logs       []LogEntry `json:"logs"`
	TotalCount int        `json:"total_count"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor,omitempty"`
	Stats      *LogStats  `json:"stats,omitempty"`
}