package timetraveldebugger

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewReplayEngine(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)

	engine := NewReplayEngine(rdb, logger)
	require.NotNil(t, engine)

	assert.Equal(t, rdb, engine.redis)
	assert.Equal(t, logger, engine.logger)
	assert.NotNil(t, engine.sessions)
}

func TestReplayEngineLoadRecording(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create a test recording
	recording := &ExecutionRecord{
		JobID:     "test_job_123",
		StartTime: time.Now(),
		Events: []Event{
			{
				ID:        "evt_1",
				Timestamp: time.Now(),
				Type:      EventEnqueued,
				JobID:     "test_job_123",
			},
		},
		Snapshots: make(map[string]StateSnapshot),
		Metadata: RecordMetadata{
			RecordID:   "rec_123",
			EventCount: 1,
		},
	}

	// Store the recording in Redis
	data, err := json.Marshal(recording)
	require.NoError(t, err)

	key := "time_travel:recording:rec_123"
	err = rdb.Set(context.Background(), key, data, time.Hour).Err()
	require.NoError(t, err)

	// Test loading the recording
	loadedRecording, err := engine.LoadRecording("rec_123")
	require.NoError(t, err)

	assert.Equal(t, recording.JobID, loadedRecording.JobID)
	assert.Equal(t, recording.Metadata.RecordID, loadedRecording.Metadata.RecordID)
	assert.Equal(t, len(recording.Events), len(loadedRecording.Events))
}

func TestReplayEngineLoadRecordingNotFound(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Test loading non-existent recording
	_, err = engine.LoadRecording("nonexistent")
	assert.Equal(t, ErrRecordingNotFound, err)
}

func TestReplayEngineLoadRecordingByJobID(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	jobID := "test_job_456"
	recordID := "rec_456"

	// Create recording
	recording := &ExecutionRecord{
		JobID: jobID,
		Events: []Event{
			{ID: "evt_1", JobID: jobID, Type: EventEnqueued},
		},
		Metadata: RecordMetadata{RecordID: recordID},
	}

	// Store recording and job index
	data, err := json.Marshal(recording)
	require.NoError(t, err)

	recordKey := "time_travel:recording:" + recordID
	err = rdb.Set(context.Background(), recordKey, data, time.Hour).Err()
	require.NoError(t, err)

	jobKey := "time_travel:job_index:" + jobID
	err = rdb.Set(context.Background(), jobKey, recordID, time.Hour).Err()
	require.NoError(t, err)

	// Test loading by job ID
	loadedRecording, err := engine.LoadRecordingByJobID(jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, loadedRecording.JobID)
	assert.Equal(t, recordID, loadedRecording.Metadata.RecordID)
}

func TestReplayEngineCreateReplaySession(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create and store a recording
	recording := createTestRecording("rec_session_test", "job_session_test")
	storeTestRecording(t, rdb, recording)

	// Create replay session
	session, err := engine.CreateReplaySession("rec_session_test", "user_123")
	require.NoError(t, err)

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "rec_session_test", session.RecordID)
	assert.Equal(t, "user_123", session.UserID)
	assert.Equal(t, 0, session.CurrentPos.EventIndex)
	assert.Equal(t, 1.0, session.PlaybackSpeed)
	assert.False(t, session.IsPlaying)
	assert.NotNil(t, session.Bookmarks)
	assert.NotNil(t, session.Annotations)

	// Verify session is stored
	retrievedSession, err := engine.GetReplaySession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrievedSession.ID)
}

func TestReplayEngineSeekTo(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with multiple events
	recording := createTestRecordingWithEvents("rec_seek_test", "job_seek_test", 5)
	storeTestRecording(t, rdb, recording)

	// Create session
	session, err := engine.CreateReplaySession("rec_seek_test", "user_123")
	require.NoError(t, err)

	// Test seeking to position 2
	position := TimelinePosition{
		EventIndex:  2,
		Timestamp:   recording.Events[2].Timestamp,
		Description: "Test position",
	}

	state, err := engine.SeekTo(session.ID, position)
	require.NoError(t, err)

	assert.Equal(t, 2, state.Position.EventIndex)
	assert.Equal(t, recording.Events[2].Timestamp, state.Position.Timestamp)

	// Verify session position was updated
	updatedSession, err := engine.GetReplaySession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updatedSession.CurrentPos.EventIndex)
}

func TestReplayEngineSeekToInvalidPosition(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with 3 events
	recording := createTestRecordingWithEvents("rec_invalid_seek", "job_invalid_seek", 3)
	storeTestRecording(t, rdb, recording)

	// Create session
	session, err := engine.CreateReplaySession("rec_invalid_seek", "user_123")
	require.NoError(t, err)

	// Test seeking to invalid position (out of bounds)
	position := TimelinePosition{
		EventIndex: 10, // Beyond available events
	}

	_, err = engine.SeekTo(session.ID, position)
	require.Error(t, err)

	replayErr, ok := err.(*ReplayError)
	require.True(t, ok)
	assert.Contains(t, replayErr.Message, "event index out of bounds")
}

func TestReplayEngineStepForwardBackward(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with multiple events
	recording := createTestRecordingWithEvents("rec_step_test", "job_step_test", 5)
	storeTestRecording(t, rdb, recording)

	// Create session
	session, err := engine.CreateReplaySession("rec_step_test", "user_123")
	require.NoError(t, err)

	// Start at position 0, step forward
	state, err := engine.StepForward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, state.Position.EventIndex)

	// Step forward again
	state, err = engine.StepForward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, state.Position.EventIndex)

	// Step backward
	state, err = engine.StepBackward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, state.Position.EventIndex)

	// Step backward to beginning
	state, err = engine.StepBackward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, state.Position.EventIndex)

	// Try to step backward beyond beginning
	_, err = engine.StepBackward(session.ID)
	require.Error(t, err)

	replayErr, ok := err.(*ReplayError)
	require.True(t, ok)
	assert.Contains(t, replayErr.Message, "already at beginning")
}

func TestReplayEngineJumpToNextError(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with error events
	baseTime := time.Now()
	recording := &ExecutionRecord{
		JobID:     "job_error_test",
		StartTime: baseTime,
		Events: []Event{
			{ID: "evt_1", Type: EventEnqueued, Timestamp: baseTime},
			{ID: "evt_2", Type: EventProcessing, Timestamp: baseTime.Add(time.Second)},
			{ID: "evt_3", Type: EventFailed, Timestamp: baseTime.Add(2 * time.Second)}, // Error event
			{ID: "evt_4", Type: EventRetrying, Timestamp: baseTime.Add(3 * time.Second)},
			{ID: "evt_5", Type: EventDLQ, Timestamp: baseTime.Add(4 * time.Second)}, // Another error event
		},
		Snapshots: make(map[string]StateSnapshot),
		Metadata:  RecordMetadata{RecordID: "rec_error_test"},
	}
	storeTestRecording(t, rdb, recording)

	// Create session
	session, err := engine.CreateReplaySession("rec_error_test", "user_123")
	require.NoError(t, err)

	// Jump to next error (should go to position 2 - EventFailed)
	state, err := engine.JumpToNextError(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, state.Position.EventIndex)
	assert.Equal(t, EventFailed, recording.Events[2].Type)

	// Jump to next error again (should go to position 4 - EventDLQ)
	state, err = engine.JumpToNextError(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, state.Position.EventIndex)
	assert.Equal(t, EventDLQ, recording.Events[4].Type)

	// Try to jump to next error when there are none
	_, err = engine.JumpToNextError(session.ID)
	require.Error(t, err)

	replayErr, ok := err.(*ReplayError)
	require.True(t, ok)
	assert.Contains(t, replayErr.Message, "no error events found")
}

func TestReplayEngineJumpToNextRetry(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with retry events
	baseTime := time.Now()
	recording := &ExecutionRecord{
		JobID:     "job_retry_test",
		StartTime: baseTime,
		Events: []Event{
			{ID: "evt_1", Type: EventEnqueued, Timestamp: baseTime},
			{ID: "evt_2", Type: EventProcessing, Timestamp: baseTime.Add(time.Second)},
			{ID: "evt_3", Type: EventRetrying, Timestamp: baseTime.Add(2 * time.Second)}, // Retry event
			{ID: "evt_4", Type: EventProcessing, Timestamp: baseTime.Add(3 * time.Second)},
			{ID: "evt_5", Type: EventRetrying, Timestamp: baseTime.Add(4 * time.Second)}, // Another retry event
		},
		Snapshots: make(map[string]StateSnapshot),
		Metadata:  RecordMetadata{RecordID: "rec_retry_test"},
	}
	storeTestRecording(t, rdb, recording)

	// Create session
	session, err := engine.CreateReplaySession("rec_retry_test", "user_123")
	require.NoError(t, err)

	// Jump to next retry (should go to position 2)
	state, err := engine.JumpToNextRetry(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, state.Position.EventIndex)
	assert.Equal(t, EventRetrying, recording.Events[2].Type)
}

func TestReplayEngineAddBookmark(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording and session
	recording := createTestRecording("rec_bookmark_test", "job_bookmark_test")
	storeTestRecording(t, rdb, recording)

	session, err := engine.CreateReplaySession("rec_bookmark_test", "user_123")
	require.NoError(t, err)

	// Seek to position 1
	position := TimelinePosition{EventIndex: 1}
	_, err = engine.SeekTo(session.ID, position)
	require.NoError(t, err)

	// Add bookmark
	description := "Important point"
	err = engine.AddBookmark(session.ID, description)
	require.NoError(t, err)

	// Verify bookmark was added
	updatedSession, err := engine.GetReplaySession(session.ID)
	require.NoError(t, err)

	assert.Len(t, updatedSession.Bookmarks, 1)
	bookmark := updatedSession.Bookmarks[0]
	assert.Equal(t, 1, bookmark.EventIndex)
	assert.Equal(t, description, bookmark.Description)
	assert.True(t, bookmark.IsBookmark)
}

func TestReplayEngineGetTimeline(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording with events
	recording := createTestRecordingWithEvents("rec_timeline_test", "job_timeline_test", 3)
	storeTestRecording(t, rdb, recording)

	// Get timeline
	timeline, err := engine.GetTimeline("rec_timeline_test")
	require.NoError(t, err)

	assert.Len(t, timeline, 3)
	for i, position := range timeline {
		assert.Equal(t, i, position.EventIndex)
		assert.Equal(t, recording.Events[i].Timestamp, position.Timestamp)
		assert.Contains(t, position.Description, string(recording.Events[i].Type))
	}
}

func TestReplayEngineCloseSession(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	engine := NewReplayEngine(rdb, logger)

	// Create recording and session
	recording := createTestRecording("rec_close_test", "job_close_test")
	storeTestRecording(t, rdb, recording)

	session, err := engine.CreateReplaySession("rec_close_test", "user_123")
	require.NoError(t, err)

	// Verify session exists
	_, err = engine.GetReplaySession(session.ID)
	require.NoError(t, err)

	// Close session
	err = engine.CloseSession(session.ID)
	require.NoError(t, err)

	// Verify session no longer exists
	_, err = engine.GetReplaySession(session.ID)
	assert.Equal(t, ErrReplaySessionNotFound, err)
}

// Helper functions

func createTestRecording(recordID, jobID string) *ExecutionRecord {
	return &ExecutionRecord{
		JobID:     jobID,
		StartTime: time.Now(),
		Events: []Event{
			{
				ID:        "evt_1",
				Timestamp: time.Now(),
				Type:      EventEnqueued,
				JobID:     jobID,
			},
		},
		Snapshots: make(map[string]StateSnapshot),
		Metadata: RecordMetadata{
			RecordID:   recordID,
			EventCount: 1,
		},
	}
}

func createTestRecordingWithEvents(recordID, jobID string, eventCount int) *ExecutionRecord {
	baseTime := time.Now()
	events := make([]Event, eventCount)

	eventTypes := []EventType{
		EventEnqueued, EventDequeued, EventProcessing, EventCompleted, EventFailed,
	}

	for i := 0; i < eventCount; i++ {
		events[i] = Event{
			ID:        "evt_" + string(rune('1'+i)),
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Type:      eventTypes[i%len(eventTypes)],
			JobID:     jobID,
		}
	}

	return &ExecutionRecord{
		JobID:     jobID,
		StartTime: baseTime,
		Events:    events,
		Snapshots: make(map[string]StateSnapshot),
		Metadata: RecordMetadata{
			RecordID:   recordID,
			EventCount: eventCount,
		},
	}
}

func storeTestRecording(t *testing.T, rdb *redis.Client, recording *ExecutionRecord) {
	data, err := json.Marshal(recording)
	require.NoError(t, err)

	key := "time_travel:recording:" + recording.Metadata.RecordID
	err = rdb.Set(context.Background(), key, data, time.Hour).Err()
	require.NoError(t, err)

	// Also store job index if needed
	jobKey := "time_travel:job_index:" + recording.JobID
	err = rdb.Set(context.Background(), jobKey, recording.Metadata.RecordID, time.Hour).Err()
	require.NoError(t, err)
}
