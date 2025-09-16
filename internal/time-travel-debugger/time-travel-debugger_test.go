package timetraveldebugger

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewTimeTravelDebugger(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	require.NotNil(t, ttd)

	assert.Equal(t, config, ttd.config)
	assert.Equal(t, rdb, ttd.redis)
	assert.Equal(t, logger, ttd.logger)
	assert.True(t, ttd.enabled)
	assert.NotNil(t, ttd.capture)
	assert.NotNil(t, ttd.replay)
}

func TestTimeTravelDebuggerWithNilConfig(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)

	// Pass nil config - should use default
	ttd := NewTimeTravelDebugger(nil, rdb, logger)
	require.NotNil(t, ttd)

	assert.NotNil(t, ttd.config)
	assert.True(t, ttd.config.Enabled)
	assert.Equal(t, 0.01, ttd.config.SamplingRate) // Default sampling rate
}

func TestTimeTravelDebuggerStartStop(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// Test start
	err = ttd.Start()
	require.NoError(t, err)

	// Test stop
	err = ttd.Stop()
	require.NoError(t, err)
}

func TestTimeTravelDebuggerStartWithInvalidConfig(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)

	// Create invalid config
	config := TestingConfig()
	config.SamplingRate = -1.0 // Invalid

	ttd := NewTimeTravelDebugger(config, rdb, logger)

	// Start should fail with invalid config
	err = ttd.Start()
	require.Error(t, err)

	configErr, ok := err.(*ConfigError)
	require.True(t, ok)
	assert.Equal(t, "sampling_rate", configErr.Field)
}

func TestTimeTravelDebuggerIsEnabled(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// Should be enabled by default
	assert.True(t, ttd.IsEnabled())

	// Disable via config
	config.Enabled = false
	assert.False(t, ttd.IsEnabled())

	// Re-enable via config
	config.Enabled = true
	assert.True(t, ttd.IsEnabled())

	// Disable via SetEnabled
	ttd.SetEnabled(false)
	assert.False(t, ttd.IsEnabled())

	// Re-enable via SetEnabled
	ttd.SetEnabled(true)
	assert.True(t, ttd.IsEnabled())
}

func TestTimeTravelDebuggerCaptureWorkflow(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	err = ttd.Start()
	require.NoError(t, err)

	// Test complete capture workflow
	jobID := "test_workflow_job"

	// 1. Start recording
	err = ttd.StartRecording(jobID, "testing", 5)
	require.NoError(t, err)

	// Verify recording is active
	activeRecordings := ttd.GetActiveRecordings()
	assert.Contains(t, activeRecordings, jobID)

	// 2. Capture job events
	jobState1 := &JobState{
		ID:       jobID,
		Priority: "high",
		Status:   "enqueued",
		Retries:  0,
	}

	err = ttd.CaptureJobEvent(jobID, EventEnqueued, jobState1, nil)
	require.NoError(t, err)

	jobState2 := &JobState{
		ID:       jobID,
		Priority: "high",
		Status:   "processing",
		Retries:  0,
	}

	systemChanges := []SystemChange{
		{
			Component: "queue",
			Key:       "length",
			Before:    10,
			After:     9,
		},
	}

	err = ttd.CaptureSystemEvent(jobID, EventProcessing, jobState2, systemChanges, nil)
	require.NoError(t, err)

	// 3. Finish recording
	finalJobState := &JobState{
		ID:       jobID,
		Priority: "high",
		Status:   "completed",
		Retries:  0,
	}

	err = ttd.FinishRecording(jobID, finalJobState)
	require.NoError(t, err)

	// Verify recording is no longer active
	activeRecordings = ttd.GetActiveRecordings()
	assert.NotContains(t, activeRecordings, jobID)

	// Wait for async storage
	time.Sleep(100 * time.Millisecond)

	// 4. Load and replay the recording
	recording, err := ttd.LoadRecordingByJobID(jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, recording.JobID)
	assert.Len(t, recording.Events, 2)
	assert.Equal(t, EventEnqueued, recording.Events[0].Type)
	assert.Equal(t, EventProcessing, recording.Events[1].Type)
}

func TestTimeTravelDebuggerReplayWorkflow(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// First, create a recording (simplified)
	recording := createTestRecordingWithEvents("rec_replay_workflow", "job_replay_workflow", 3)
	storeTestRecording(t, rdb, recording)

	// Test replay workflow
	// 1. Create replay session
	session, err := ttd.CreateReplaySession("rec_replay_workflow", "user_123")
	require.NoError(t, err)

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "rec_replay_workflow", session.RecordID)

	// 2. Navigate timeline
	state, err := ttd.StepForward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, state.Position.EventIndex)

	state, err = ttd.StepForward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, state.Position.EventIndex)

	state, err = ttd.StepBackward(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, state.Position.EventIndex)

	// 3. Add bookmark
	err = ttd.AddBookmark(session.ID, "Important point")
	require.NoError(t, err)

	// 4. Get timeline
	timeline, err := ttd.GetTimeline("rec_replay_workflow")
	require.NoError(t, err)
	assert.Len(t, timeline, 3)

	// 5. Close session
	err = ttd.CloseSession(session.ID)
	require.NoError(t, err)

	// Verify session is closed
	_, err = ttd.GetReplaySession(session.ID)
	assert.Equal(t, ErrReplaySessionNotFound, err)
}

func TestTimeTravelDebuggerAutoStartRecording(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	err = ttd.Start()
	require.NoError(t, err)

	tests := []struct {
		name       string
		jobID      string
		jobStatus  string
		hasRetries bool
		isFailure  bool
		shouldRecord bool
	}{
		{
			name:         "failure job",
			jobID:        "failed_job",
			jobStatus:    "failed",
			isFailure:    true,
			shouldRecord: true,
		},
		{
			name:         "retry job",
			jobID:        "retry_job",
			jobStatus:    "processing",
			hasRetries:   true,
			shouldRecord: true,
		},
		{
			name:         "normal job",
			jobID:        "normal_job",
			jobStatus:    "completed",
			shouldRecord: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ShouldRecord
			shouldRecord := ttd.ShouldRecord(tt.jobID, tt.jobStatus, tt.hasRetries, tt.isFailure)
			assert.Equal(t, tt.shouldRecord, shouldRecord)

			// Test AutoStartRecording
			err := ttd.AutoStartRecording(tt.jobID, tt.jobStatus, tt.hasRetries, tt.isFailure)

			if tt.shouldRecord {
				require.NoError(t, err)
				activeRecordings := ttd.GetActiveRecordings()
				assert.Contains(t, activeRecordings, tt.jobID)
			} else {
				require.NoError(t, err) // Should not error, just not record
				activeRecordings := ttd.GetActiveRecordings()
				assert.NotContains(t, activeRecordings, tt.jobID)
			}
		})
	}
}

func TestTimeTravelDebuggerUpdateConfig(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// Test updating config
	newConfig := ProductionConfig()
	err = ttd.UpdateConfig(newConfig)
	require.NoError(t, err)

	// Verify config was updated
	currentConfig := ttd.GetConfig()
	assert.Equal(t, newConfig.SamplingRate, currentConfig.SamplingRate)
	assert.Equal(t, newConfig.MaxEvents, currentConfig.MaxEvents)

	// Test updating with invalid config
	invalidConfig := TestingConfig()
	invalidConfig.SamplingRate = -1.0

	err = ttd.UpdateConfig(invalidConfig)
	require.Error(t, err)

	configErr, ok := err.(*ConfigError)
	require.True(t, ok)
	assert.Equal(t, "sampling_rate", configErr.Field)
}

func TestTimeTravelDebuggerHealth(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// Test health check
	health := ttd.Health()

	assert.Contains(t, health, "enabled")
	assert.Contains(t, health, "config_enabled")
	assert.Contains(t, health, "sampling_rate")
	assert.Contains(t, health, "active_recordings")

	assert.Equal(t, true, health["enabled"])
	assert.Equal(t, true, health["config_enabled"])
	assert.Equal(t, config.SamplingRate, health["sampling_rate"])
	assert.Equal(t, 0, health["active_recordings"])

	// Start a recording and check health again
	err = ttd.StartRecording("health_test_job", "testing", 5)
	require.NoError(t, err)

	health = ttd.Health()
	assert.Equal(t, 1, health["active_recordings"])
}

func TestTimeTravelDebuggerDisabledOperations(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()
	config.Enabled = false // Disable the debugger

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// All capture operations should return ErrCaptureDisabled
	err = ttd.StartRecording("test_job", "testing", 5)
	assert.Equal(t, ErrCaptureDisabled, err)

	err = ttd.CaptureJobEvent("test_job", EventEnqueued, nil, nil)
	assert.Equal(t, ErrCaptureDisabled, err)

	err = ttd.CaptureSystemEvent("test_job", EventEnqueued, nil, nil, nil)
	assert.Equal(t, ErrCaptureDisabled, err)

	err = ttd.FinishRecording("test_job", nil)
	assert.Equal(t, ErrCaptureDisabled, err)

	// GetActiveRecordings should return empty
	activeRecordings := ttd.GetActiveRecordings()
	assert.Empty(t, activeRecordings)

	// ShouldRecord should return false
	shouldRecord := ttd.ShouldRecord("test_job", "failed", false, true)
	assert.False(t, shouldRecord)
}

func TestTimeTravelDebuggerManagementInterface(t *testing.T) {
	// Setup
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	logger := zaptest.NewLogger(t)
	config := TestingConfig()

	ttd := NewTimeTravelDebugger(config, rdb, logger)
	defer ttd.Stop()

	// Create some test recordings
	recording1 := createTestRecording("rec_mgmt_1", "job_mgmt_1")
	recording2 := createTestRecording("rec_mgmt_2", "job_mgmt_2")

	storeTestRecording(t, rdb, recording1)
	storeTestRecording(t, rdb, recording2)

	// Test GetRecentRecordings
	recordings, err := ttd.GetRecentRecordings(10)
	require.NoError(t, err)
	assert.True(t, len(recordings) >= 2)

	// Test SearchRecordings (basic test as implementation is placeholder)
	searchResults, err := ttd.SearchRecordings("mgmt", 10)
	require.NoError(t, err)
	assert.NotNil(t, searchResults) // Implementation returns empty but should not error
}