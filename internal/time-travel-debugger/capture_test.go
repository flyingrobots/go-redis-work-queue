//go:build time_travel_debugger_tests
// +build time_travel_debugger_tests

package timetraveldebugger

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewEventCapture(t *testing.T) {
	// Setup miniredis
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	capture := NewEventCapture(config, rdb, logger)
	require.NotNil(t, capture)

	assert.Equal(t, config, capture.config)
	assert.Equal(t, rdb, capture.redis)
	assert.Equal(t, logger, capture.logger)
	assert.NotNil(t, capture.recordings)
	assert.NotNil(t, capture.eventChan)
	assert.NotNil(t, capture.snapChan)

	// Clean up
	err = capture.Close()
	assert.NoError(t, err)
}

func TestEventCaptureStartRecording(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	// Test starting recording
	jobID := "test_job_123"
	reason := "testing"
	importance := 5

	err = capture.StartRecording(jobID, reason, importance)
	require.NoError(t, err)

	// Verify recording was created
	activeRecordings := capture.GetActiveRecordings()
	assert.Contains(t, activeRecordings, jobID)

	// Verify recording details
	capture.mu.RLock()
	activeRec, exists := capture.recordings[jobID]
	capture.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, jobID, activeRec.record.JobID)
	assert.Equal(t, reason, activeRec.record.Metadata.Reason)
	assert.Equal(t, importance, activeRec.record.Metadata.Importance)
	assert.Equal(t, "time-travel-debugger", activeRec.record.Metadata.RecordedBy)
}

func TestEventCaptureStartRecordingDisabled(t *testing.T) {
	// Setup with disabled config
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	config.Enabled = false
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	// Test should return error when disabled
	err = capture.StartRecording("test_job", "testing", 5)
	assert.Equal(t, ErrCaptureDisabled, err)
}

func TestEventCaptureCaptureEvent(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	// Start recording first
	jobID := "test_job_123"
	err = capture.StartRecording(jobID, "testing", 5)
	require.NoError(t, err)

	// Create job state
	jobState := &JobState{
		ID:       jobID,
		Priority: "high",
		Status:   "processing",
		Retries:  1,
	}

	// Create system changes
	systemChanges := []SystemChange{
		{
			Component: "queue",
			Key:       "length",
			Before:    10,
			After:     9,
		},
	}

	// Create context
	context := map[string]interface{}{
		"worker_id": "worker_123",
		"source":    "test",
	}

	// Capture event
	err = capture.CaptureEvent(jobID, EventProcessing, jobState, systemChanges, context)
	require.NoError(t, err)

	// Verify event was captured
	capture.mu.RLock()
	activeRec, exists := capture.recordings[jobID]
	capture.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, 1, len(activeRec.record.Events))

	event := activeRec.record.Events[0]
	assert.Equal(t, EventProcessing, event.Type)
	assert.Equal(t, jobID, event.JobID)
	assert.Equal(t, jobState, event.StateChange.JobStateAfter)
	assert.Equal(t, systemChanges, event.StateChange.SystemChanges)
	assert.Equal(t, context, event.Context)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestEventCaptureAutoStartRecording(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	jobID := "test_job_failed"
	jobState := &JobState{
		ID:     jobID,
		Status: "failed",
	}

	// Should auto-start recording for failure events
	err = capture.CaptureEvent(jobID, EventFailed, jobState, nil, nil)
	require.NoError(t, err)

	// Verify recording was auto-started
	activeRecordings := capture.GetActiveRecordings()
	assert.Contains(t, activeRecordings, jobID)

	// Verify recording has higher importance for failures
	capture.mu.RLock()
	activeRec, exists := capture.recordings[jobID]
	capture.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, 7, activeRec.record.Metadata.Importance) // Higher importance for failures
	assert.Equal(t, "FAILED", activeRec.record.Metadata.Reason)
}

func TestEventCaptureMaxEvents(t *testing.T) {
	// Setup with small max events
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	config.MaxEvents = 2 // Very small for testing
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	jobID := "test_job_max"
	err = capture.StartRecording(jobID, "testing", 5)
	require.NoError(t, err)

	// Capture up to max events
	for i := 0; i < config.MaxEvents; i++ {
		err = capture.CaptureEvent(jobID, EventProcessing, &JobState{ID: jobID}, nil, nil)
		require.NoError(t, err)
	}

	// Next event should fail due to max events exceeded
	err = capture.CaptureEvent(jobID, EventCompleted, &JobState{ID: jobID}, nil, nil)
	require.Error(t, err)

	captureErr, ok := err.(*CaptureError)
	require.True(t, ok)
	assert.Contains(t, captureErr.Message, "max events exceeded")
}

func TestEventCaptureSnapshots(t *testing.T) {
	// Setup with fast snapshot interval
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	config.SnapshotInterval = 10 * time.Millisecond // Very fast for testing
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	jobID := "test_job_snapshots"
	err = capture.StartRecording(jobID, "testing", 5)
	require.NoError(t, err)

	// Capture first event
	err = capture.CaptureEvent(jobID, EventProcessing, &JobState{ID: jobID}, nil, nil)
	require.NoError(t, err)

	// Wait for snapshot interval
	time.Sleep(15 * time.Millisecond)

	// Capture second event (should trigger snapshot)
	err = capture.CaptureEvent(jobID, EventCompleted, &JobState{ID: jobID}, nil, nil)
	require.NoError(t, err)

	// Verify snapshot was created
	capture.mu.RLock()
	activeRec, exists := capture.recordings[jobID]
	capture.mu.RUnlock()

	require.True(t, exists)
	assert.True(t, len(activeRec.record.Snapshots) > 0)
	assert.True(t, activeRec.record.Metadata.SnapshotCount > 0)
}

func TestEventCaptureFinishRecording(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	jobID := "test_job_finish"
	err = capture.StartRecording(jobID, "testing", 5)
	require.NoError(t, err)

	// Capture some events
	err = capture.CaptureEvent(jobID, EventProcessing, &JobState{ID: jobID}, nil, nil)
	require.NoError(t, err)

	finalJobState := &JobState{
		ID:     jobID,
		Status: "completed",
	}

	// Finish recording
	err = capture.FinishRecording(jobID, finalJobState)
	require.NoError(t, err)

	// Verify recording is no longer active
	activeRecordings := capture.GetActiveRecordings()
	assert.NotContains(t, activeRecordings, jobID)

	// Give time for async storage
	time.Sleep(100 * time.Millisecond)

	// Verify recording was stored in Redis
	recordKey := "time_travel:recording:*"
	keys, err := rdb.Keys(context.Background(), recordKey).Result()
	require.NoError(t, err)
	assert.True(t, len(keys) > 0)

	// Verify job index was created
	jobKey := "time_travel:job_index:" + jobID
	exists, err := rdb.Exists(context.Background(), jobKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

func TestEventCaptureFinishRecordingNotFound(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	// Try to finish recording that was never started
	err = capture.FinishRecording("nonexistent_job", nil)
	require.Error(t, err)

	recordingErr, ok := err.(*RecordingError)
	require.True(t, ok)
	assert.Contains(t, recordingErr.Message, "no active recording found")
}

func TestEventCaptureClose(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)

	// Start some recording
	err = capture.StartRecording("test_job", "testing", 5)
	require.NoError(t, err)

	// Close should succeed
	err = capture.Close()
	assert.NoError(t, err)

	// Channels should be closed (this would panic if not closed properly)
	// The background goroutine should have exited
}

func TestEventCaptureGenerateIDs(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	capture := NewEventCapture(config, rdb, logger)
	defer capture.Close()

	// Test record ID generation
	recordID1 := capture.generateRecordID()
	recordID2 := capture.generateRecordID()

	assert.NotEmpty(t, recordID1)
	assert.NotEmpty(t, recordID2)
	assert.NotEqual(t, recordID1, recordID2)
	assert.True(t, len(recordID1) > 4) // "rec_" prefix + hex
	assert.True(t, len(recordID2) > 4)

	// Test event ID generation
	eventID1 := capture.generateEventID()
	eventID2 := capture.generateEventID()

	assert.NotEmpty(t, eventID1)
	assert.NotEmpty(t, eventID2)
	assert.NotEqual(t, eventID1, eventID2)
	assert.True(t, len(eventID1) > 4) // "evt_" prefix + hex
	assert.True(t, len(eventID2) > 4)
}
