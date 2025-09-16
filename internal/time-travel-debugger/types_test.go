package timetraveldebugger

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventMarshaling(t *testing.T) {
	event := Event{
		ID:        "evt_123",
		Timestamp: time.Now(),
		Type:      EventEnqueued,
		JobID:     "job_456",
		WorkerID:  "worker_789",
		QueueName: "test_queue",
		StateChange: StateDiff{
			JobStateAfter: &JobState{
				ID:       "job_456",
				Priority: "high",
				Status:   "enqueued",
				Retries:  0,
			},
		},
		Context: map[string]interface{}{
			"source": "test",
		},
		TraceID: "trace_123",
		SpanID:  "span_456",
	}

	// Test marshaling
	data, err := event.Marshal()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled Event
	err = unmarshaled.Unmarshal(data)
	require.NoError(t, err)

	assert.Equal(t, event.ID, unmarshaled.ID)
	assert.Equal(t, event.Type, unmarshaled.Type)
	assert.Equal(t, event.JobID, unmarshaled.JobID)
	assert.Equal(t, event.WorkerID, unmarshaled.WorkerID)
	assert.Equal(t, event.QueueName, unmarshaled.QueueName)
	assert.Equal(t, event.TraceID, unmarshaled.TraceID)
	assert.Equal(t, event.SpanID, unmarshaled.SpanID)
}

func TestExecutionRecordDuration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)

	// Test with end time
	record := ExecutionRecord{
		StartTime: startTime,
		EndTime:   &endTime,
	}

	duration := record.Duration()
	assert.Equal(t, 5*time.Minute, duration)

	// Test without end time (ongoing)
	recordOngoing := ExecutionRecord{
		StartTime: startTime,
		EndTime:   nil,
	}

	durationOngoing := recordOngoing.Duration()
	assert.True(t, durationOngoing > 0)
}

func TestExecutionRecordEventAccess(t *testing.T) {
	events := []Event{
		{ID: "evt_1", Timestamp: time.Now(), Type: EventEnqueued},
		{ID: "evt_2", Timestamp: time.Now().Add(time.Second), Type: EventDequeued},
		{ID: "evt_3", Timestamp: time.Now().Add(2 * time.Second), Type: EventCompleted},
	}

	record := ExecutionRecord{
		Events: events,
	}

	// Test valid index
	event := record.EventAtIndex(1)
	require.NotNil(t, event)
	assert.Equal(t, "evt_2", event.ID)
	assert.Equal(t, EventDequeued, event.Type)

	// Test invalid indices
	assert.Nil(t, record.EventAtIndex(-1))
	assert.Nil(t, record.EventAtIndex(10))
}

func TestExecutionRecordFindEventByTimestamp(t *testing.T) {
	baseTime := time.Now()
	events := []Event{
		{ID: "evt_1", Timestamp: baseTime, Type: EventEnqueued},
		{ID: "evt_2", Timestamp: baseTime.Add(time.Second), Type: EventDequeued},
		{ID: "evt_3", Timestamp: baseTime.Add(2 * time.Second), Type: EventCompleted},
	}

	record := ExecutionRecord{
		Events: events,
	}

	// Test exact timestamp match
	event := record.FindEventByTimestamp(baseTime.Add(time.Second))
	require.NotNil(t, event)
	assert.Equal(t, "evt_2", event.ID)

	// Test closest timestamp
	event = record.FindEventByTimestamp(baseTime.Add(1500 * time.Millisecond))
	require.NotNil(t, event)
	assert.Equal(t, "evt_2", event.ID) // Should be closer to evt_2

	// Test with empty events
	emptyRecord := ExecutionRecord{Events: []Event{}}
	assert.Nil(t, emptyRecord.FindEventByTimestamp(baseTime))
}

func TestExecutionRecordGetEventsInTimeRange(t *testing.T) {
	baseTime := time.Now()
	events := []Event{
		{ID: "evt_1", Timestamp: baseTime, Type: EventEnqueued},
		{ID: "evt_2", Timestamp: baseTime.Add(time.Second), Type: EventDequeued},
		{ID: "evt_3", Timestamp: baseTime.Add(2 * time.Second), Type: EventProcessing},
		{ID: "evt_4", Timestamp: baseTime.Add(3 * time.Second), Type: EventCompleted},
	}

	record := ExecutionRecord{
		Events: events,
	}

	// Test range that includes some events
	rangeEvents := record.GetEventsInTimeRange(
		baseTime.Add(500*time.Millisecond),
		baseTime.Add(2500*time.Millisecond),
	)

	assert.Len(t, rangeEvents, 2)
	assert.Equal(t, "evt_2", rangeEvents[0].ID)
	assert.Equal(t, "evt_3", rangeEvents[1].ID)

	// Test range with no events
	noEvents := record.GetEventsInTimeRange(
		baseTime.Add(10*time.Second),
		baseTime.Add(20*time.Second),
	)
	assert.Len(t, noEvents, 0)
}

func TestExecutionRecordGetSnapshotNearTimestamp(t *testing.T) {
	baseTime := time.Now()
	snapshots := map[string]StateSnapshot{
		baseTime.Format(time.RFC3339Nano): {
			Timestamp: baseTime,
			JobState:  &JobState{ID: "job_1"},
		},
		baseTime.Add(2 * time.Second).Format(time.RFC3339Nano): {
			Timestamp: baseTime.Add(2 * time.Second),
			JobState:  &JobState{ID: "job_2"},
		},
	}

	record := ExecutionRecord{
		Snapshots: snapshots,
	}

	// Test finding closest snapshot
	snapshot := record.GetSnapshotNearTimestamp(baseTime.Add(time.Second))
	require.NotNil(t, snapshot)
	assert.Equal(t, "job_1", snapshot.JobState.ID)

	// Test finding later snapshot
	snapshot = record.GetSnapshotNearTimestamp(baseTime.Add(3 * time.Second))
	require.NotNil(t, snapshot)
	assert.Equal(t, "job_2", snapshot.JobState.ID)

	// Test with no snapshots
	emptyRecord := ExecutionRecord{Snapshots: make(map[string]StateSnapshot)}
	assert.Nil(t, emptyRecord.GetSnapshotNearTimestamp(baseTime))
}

func TestJobStateJSONMarshaling(t *testing.T) {
	createdAt := time.Now()
	jobState := &JobState{
		ID:          "job_123",
		Priority:    "high",
		Retries:     2,
		MaxRetries:  5,
		Status:      "processing",
		Payload:     map[string]interface{}{"key": "value"},
		Metadata:    map[string]interface{}{"source": "test"},
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt.Add(time.Minute),
		StartedAt:   &createdAt,
		ErrorMessage: "test error",
	}

	// Marshal to JSON
	data, err := json.Marshal(jobState)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled JobState
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, jobState.ID, unmarshaled.ID)
	assert.Equal(t, jobState.Priority, unmarshaled.Priority)
	assert.Equal(t, jobState.Retries, unmarshaled.Retries)
	assert.Equal(t, jobState.Status, unmarshaled.Status)
	assert.Equal(t, jobState.ErrorMessage, unmarshaled.ErrorMessage)
	assert.NotNil(t, unmarshaled.StartedAt)
}

func TestEventTypeConstants(t *testing.T) {
	// Ensure all event types are defined correctly
	eventTypes := []EventType{
		EventEnqueued,
		EventDequeued,
		EventProcessing,
		EventRetrying,
		EventFailed,
		EventCompleted,
		EventDLQ,
		EventScheduled,
		EventCancelled,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType))
	}

	// Test specific values
	assert.Equal(t, EventType("ENQUEUED"), EventEnqueued)
	assert.Equal(t, EventType("FAILED"), EventFailed)
	assert.Equal(t, EventType("COMPLETED"), EventCompleted)
}

func TestPerformanceSnapshotValidation(t *testing.T) {
	perfSnapshot := &PerformanceSnapshot{
		ProcessingTime:   time.Second,
		QueueLength:      100,
		WorkerCount:      5,
		MemoryUsageMB:    512.5,
		CPUUsagePercent:  75.2,
		RedisConnections: 10,
		ErrorRate:        0.05,
	}

	// Marshal and unmarshal to ensure proper JSON handling
	data, err := json.Marshal(perfSnapshot)
	require.NoError(t, err)

	var unmarshaled PerformanceSnapshot
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, perfSnapshot.ProcessingTime, unmarshaled.ProcessingTime)
	assert.Equal(t, perfSnapshot.QueueLength, unmarshaled.QueueLength)
	assert.Equal(t, perfSnapshot.WorkerCount, unmarshaled.WorkerCount)
	assert.Equal(t, perfSnapshot.MemoryUsageMB, unmarshaled.MemoryUsageMB)
	assert.Equal(t, perfSnapshot.CPUUsagePercent, unmarshaled.CPUUsagePercent)
	assert.Equal(t, perfSnapshot.RedisConnections, unmarshaled.RedisConnections)
	assert.Equal(t, perfSnapshot.ErrorRate, unmarshaled.ErrorRate)
}

func TestTimelinePositionBookmarkAndBreakpoint(t *testing.T) {
	position := TimelinePosition{
		EventIndex:   10,
		Timestamp:    time.Now(),
		Description:  "Test position",
		IsBookmark:   true,
		IsBreakpoint: false,
	}

	// Test JSON marshaling
	data, err := json.Marshal(position)
	require.NoError(t, err)

	var unmarshaled TimelinePosition
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, position.EventIndex, unmarshaled.EventIndex)
	assert.Equal(t, position.Description, unmarshaled.Description)
	assert.Equal(t, position.IsBookmark, unmarshaled.IsBookmark)
	assert.Equal(t, position.IsBreakpoint, unmarshaled.IsBreakpoint)
}

func TestExportFormatConstants(t *testing.T) {
	formats := []ExportFormat{
		ExportJSON,
		ExportMarkdown,
		ExportBundle,
		ExportTestCase,
		ExportVideo,
	}

	for _, format := range formats {
		assert.NotEmpty(t, string(format))
	}

	// Test specific values
	assert.Equal(t, ExportFormat("json"), ExportJSON)
	assert.Equal(t, ExportFormat("markdown"), ExportMarkdown)
	assert.Equal(t, ExportFormat("bundle"), ExportBundle)
}