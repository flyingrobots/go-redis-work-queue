package timetraveldebugger

import (
	"encoding/json"
	"time"
)

// EventType represents different types of job state transitions
type EventType string

const (
	EventEnqueued   EventType = "ENQUEUED"
	EventDequeued   EventType = "DEQUEUED"
	EventProcessing EventType = "PROCESSING"
	EventRetrying   EventType = "RETRYING"
	EventFailed     EventType = "FAILED"
	EventCompleted  EventType = "COMPLETED"
	EventDLQ        EventType = "MOVED_TO_DLQ"
	EventScheduled  EventType = "SCHEDULED"
	EventCancelled  EventType = "CANCELLED"
)

// Event represents a single job state transition
type Event struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        EventType              `json:"type"`
	JobID       string                 `json:"job_id"`
	WorkerID    string                 `json:"worker_id,omitempty"`
	QueueName   string                 `json:"queue_name"`
	StateChange StateDiff              `json:"state_change"`
	Context     map[string]interface{} `json:"context,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// StateDiff represents changes in job or system state
type StateDiff struct {
	JobStateBefore  *JobState     `json:"job_state_before,omitempty"`
	JobStateAfter   *JobState     `json:"job_state_after,omitempty"`
	SystemChanges   []SystemChange `json:"system_changes,omitempty"`
	PerformanceData *PerformanceSnapshot `json:"performance_data,omitempty"`
}

// JobState represents the complete state of a job at a point in time
type JobState struct {
	ID           string                 `json:"id"`
	Priority     string                 `json:"priority"`
	Retries      int                    `json:"retries"`
	MaxRetries   int                    `json:"max_retries"`
	Status       string                 `json:"status"`
	Payload      map[string]interface{} `json:"payload"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	FailedAt     *time.Time             `json:"failed_at,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// SystemChange represents a change in system state
type SystemChange struct {
	Component string      `json:"component"` // e.g., "queue", "worker", "redis"
	Key       string      `json:"key"`       // e.g., "length", "health", "memory_usage"
	Before    interface{} `json:"before"`
	After     interface{} `json:"after"`
}

// PerformanceSnapshot captures performance metrics at event time
type PerformanceSnapshot struct {
	ProcessingTime   time.Duration `json:"processing_time,omitempty"`
	QueueLength      int64         `json:"queue_length"`
	WorkerCount      int           `json:"worker_count"`
	MemoryUsageMB    float64       `json:"memory_usage_mb"`
	CPUUsagePercent  float64       `json:"cpu_usage_percent"`
	RedisConnections int           `json:"redis_connections"`
	ErrorRate        float64       `json:"error_rate"`
}

// StateSnapshot represents complete system state at a specific time
type StateSnapshot struct {
	Timestamp    time.Time              `json:"timestamp"`
	JobState     *JobState              `json:"job_state"`
	WorkerState  *WorkerState           `json:"worker_state,omitempty"`
	QueueState   *QueueState            `json:"queue_state"`
	RedisKeys    map[string]interface{} `json:"redis_keys,omitempty"`
	SystemMetrics *PerformanceSnapshot  `json:"system_metrics"`
}

// WorkerState represents the state of a worker at a point in time
type WorkerState struct {
	ID              string            `json:"id"`
	Health          string            `json:"health"`
	ProcessingJobID string            `json:"processing_job_id,omitempty"`
	StartTime       time.Time         `json:"start_time"`
	LastSeen        time.Time         `json:"last_seen"`
	ResourceUsage   map[string]float64 `json:"resource_usage"`
	QueueAssignment []string          `json:"queue_assignment"`
}

// QueueState represents the state of a queue at a point in time
type QueueState struct {
	Name          string    `json:"name"`
	Length        int64     `json:"length"`
	ProcessingCount int64   `json:"processing_count"`
	PendingCount  int64     `json:"pending_count"`
	DLQCount      int64     `json:"dlq_count"`
	Rate          float64   `json:"rate"` // jobs per second
	BackPressure  bool      `json:"back_pressure"`
	LastUpdated   time.Time `json:"last_updated"`
}

// ExecutionRecord represents a complete recording of a job's execution
type ExecutionRecord struct {
	JobID       string                    `json:"job_id"`
	StartTime   time.Time                 `json:"start_time"`
	EndTime     *time.Time                `json:"end_time,omitempty"`
	Events      []Event                   `json:"events"`
	Snapshots   map[string]StateSnapshot  `json:"snapshots"` // keyed by timestamp
	Metadata    RecordMetadata            `json:"metadata"`
	Compressed  bool                      `json:"compressed"`
	Tags        []string                  `json:"tags,omitempty"`
}

// RecordMetadata contains information about the recording itself
type RecordMetadata struct {
	RecordID     string            `json:"record_id"`
	CreatedAt    time.Time         `json:"created_at"`
	RecordedBy   string            `json:"recorded_by"` // system component that triggered recording
	Reason       string            `json:"reason"`      // why this job was recorded
	Importance   int               `json:"importance"`  // 1-10 scale for retention priority
	Size         int64             `json:"size"`        // compressed size in bytes
	EventCount   int               `json:"event_count"`
	SnapshotCount int              `json:"snapshot_count"`
	Retention    time.Duration     `json:"retention"`   // how long to keep this recording
	Annotations  map[string]string `json:"annotations,omitempty"` // user annotations
}

// TimelinePosition represents a position in the replay timeline
type TimelinePosition struct {
	EventIndex   int       `json:"event_index"`
	Timestamp    time.Time `json:"timestamp"`
	Description  string    `json:"description"`
	IsBookmark   bool      `json:"is_bookmark"`
	IsBreakpoint bool      `json:"is_breakpoint"`
}

// ReplaySession represents an active replay session
type ReplaySession struct {
	ID           string              `json:"id"`
	RecordID     string              `json:"record_id"`
	UserID       string              `json:"user_id"`
	StartTime    time.Time           `json:"start_time"`
	CurrentPos   TimelinePosition    `json:"current_position"`
	PlaybackSpeed float64            `json:"playback_speed"`
	IsPlaying    bool                `json:"is_playing"`
	Bookmarks    []TimelinePosition  `json:"bookmarks"`
	Annotations  map[string]string   `json:"annotations"`
	ComparisonID string              `json:"comparison_id,omitempty"` // for comparing with another recording
}

// CaptureConfig controls what and how we capture events
type CaptureConfig struct {
	Enabled            bool          `json:"enabled"`
	SamplingRate       float64       `json:"sampling_rate"`        // 0.0-1.0, what fraction of jobs to record
	ForceOnFailure     bool          `json:"force_on_failure"`     // always record failed jobs
	ForceOnRetry       bool          `json:"force_on_retry"`       // always record jobs that retry
	MaxEvents          int           `json:"max_events"`           // max events per recording
	SnapshotInterval   time.Duration `json:"snapshot_interval"`    // how often to take full snapshots
	CompressionEnabled bool          `json:"compression_enabled"`
	RetentionPolicy    RetentionPolicy `json:"retention_policy"`
	SensitiveFields    []string      `json:"sensitive_fields"`     // fields to redact in payloads
}

// RetentionPolicy defines how long to keep recordings
type RetentionPolicy struct {
	FailedJobs     time.Duration `json:"failed_jobs"`      // e.g., 7 days
	SuccessfulJobs time.Duration `json:"successful_jobs"`  // e.g., 24 hours
	ImportantJobs  time.Duration `json:"important_jobs"`   // e.g., 30 days (user-marked)
	MaxRecordings  int           `json:"max_recordings"`   // total limit before pruning oldest
}

// ExportRequest represents a request to export a replay session
type ExportRequest struct {
	RecordID    string            `json:"record_id"`
	Format      ExportFormat      `json:"format"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
}

// ExportFormat defines available export formats
type ExportFormat string

const (
	ExportJSON       ExportFormat = "json"        // raw JSON recording
	ExportMarkdown   ExportFormat = "markdown"    // human-readable report
	ExportBundle     ExportFormat = "bundle"      // shareable replay package
	ExportTestCase   ExportFormat = "test_case"   // generated unit test
	ExportVideo      ExportFormat = "video"       // MP4 recording (future)
)

// ExportResult contains the exported data
type ExportResult struct {
	Data        []byte            `json:"data"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Marshal provides JSON marshaling for events
func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal provides JSON unmarshaling for events
func (e *Event) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}

// Duration returns the total duration of the execution recording
func (er ExecutionRecord) Duration() time.Duration {
	if er.EndTime == nil {
		return time.Since(er.StartTime)
	}
	return er.EndTime.Sub(er.StartTime)
}

// EventAtIndex returns the event at the specified index, or nil if out of bounds
func (er ExecutionRecord) EventAtIndex(index int) *Event {
	if index < 0 || index >= len(er.Events) {
		return nil
	}
	return &er.Events[index]
}

// FindEventByTimestamp finds the event closest to the specified timestamp
func (er ExecutionRecord) FindEventByTimestamp(timestamp time.Time) *Event {
	if len(er.Events) == 0 {
		return nil
	}

	// Binary search for closest timestamp
	left, right := 0, len(er.Events)-1
	closest := 0
	minDiff := time.Duration(1<<63 - 1) // max duration

	for left <= right {
		mid := (left + right) / 2
		diff := timestamp.Sub(er.Events[mid].Timestamp)
		if diff < 0 {
			diff = -diff
		}

		if diff < minDiff {
			minDiff = diff
			closest = mid
		}

		if er.Events[mid].Timestamp.Before(timestamp) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return &er.Events[closest]
}

// GetEventsInTimeRange returns events within the specified time range
func (er ExecutionRecord) GetEventsInTimeRange(start, end time.Time) []Event {
	var result []Event
	for _, event := range er.Events {
		if !event.Timestamp.Before(start) && !event.Timestamp.After(end) {
			result = append(result, event)
		}
	}
	return result
}

// GetSnapshotNearTimestamp returns the snapshot closest to the specified timestamp
func (er ExecutionRecord) GetSnapshotNearTimestamp(timestamp time.Time) *StateSnapshot {
	var closest *StateSnapshot
	minDiff := time.Duration(1<<63 - 1) // max duration

	for _, snapshot := range er.Snapshots {
		diff := timestamp.Sub(snapshot.Timestamp)
		if diff < 0 {
			diff = -diff
		}

		if diff < minDiff {
			minDiff = diff
			snap := snapshot // copy to avoid pointer issues
			closest = &snap
		}
	}

	return closest
}