package timetraveldebugger

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TimeTravelDebugger is the main service that combines event capture and replay
type TimeTravelDebugger struct {
	config *CaptureConfig
	redis  *redis.Client
	logger *zap.Logger

	// Core components
	capture *EventCapture
	replay  *ReplayEngine

	// State
	mu      sync.RWMutex
	enabled bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewTimeTravelDebugger creates a new time travel debugger instance
func NewTimeTravelDebugger(config *CaptureConfig, redis *redis.Client, logger *zap.Logger) *TimeTravelDebugger {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ttd := &TimeTravelDebugger{
		config:  config,
		redis:   redis,
		logger:  logger,
		enabled: config.Enabled,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize components
	ttd.capture = NewEventCapture(config, redis, logger.Named("capture"))
	ttd.replay = NewReplayEngine(redis, logger.Named("replay"))

	return ttd
}

// Start initializes and starts the time travel debugger
func (ttd *TimeTravelDebugger) Start() error {
	ttd.mu.Lock()
	defer ttd.mu.Unlock()

	if err := ttd.config.Validate(); err != nil {
		return err
	}

	ttd.logger.Info("Starting Time Travel Debugger",
		zap.Bool("enabled", ttd.config.Enabled),
		zap.Float64("sampling_rate", ttd.config.SamplingRate),
		zap.Bool("force_on_failure", ttd.config.ForceOnFailure),
		zap.Int("max_events", ttd.config.MaxEvents),
	)

	return nil
}

// Stop shuts down the time travel debugger
func (ttd *TimeTravelDebugger) Stop() error {
	ttd.mu.Lock()
	defer ttd.mu.Unlock()

	ttd.logger.Info("Stopping Time Travel Debugger")

	ttd.cancel()

	if ttd.capture != nil {
		if err := ttd.capture.Close(); err != nil {
			ttd.logger.Error("Error closing event capture", zap.Error(err))
		}
	}

	return nil
}

// IsEnabled returns whether the debugger is currently enabled
func (ttd *TimeTravelDebugger) IsEnabled() bool {
	ttd.mu.RLock()
	defer ttd.mu.RUnlock()
	return ttd.enabled && ttd.config.Enabled
}

// SetEnabled enables or disables the debugger
func (ttd *TimeTravelDebugger) SetEnabled(enabled bool) {
	ttd.mu.Lock()
	defer ttd.mu.Unlock()
	ttd.enabled = enabled
	ttd.logger.Info("Time Travel Debugger enabled state changed", zap.Bool("enabled", enabled))
}

// Event Capture Interface

// StartRecording begins recording events for a job
func (ttd *TimeTravelDebugger) StartRecording(jobID string, reason string, importance int) error {
	if !ttd.IsEnabled() {
		return ErrCaptureDisabled
	}
	return ttd.capture.StartRecording(jobID, reason, importance)
}

// CaptureJobEvent records a job-related event
func (ttd *TimeTravelDebugger) CaptureJobEvent(jobID string, eventType EventType, jobState *JobState, context map[string]interface{}) error {
	if !ttd.IsEnabled() {
		return ErrCaptureDisabled
	}
	return ttd.capture.CaptureEvent(jobID, eventType, jobState, nil, context)
}

// CaptureSystemEvent records a system-level event with state changes
func (ttd *TimeTravelDebugger) CaptureSystemEvent(jobID string, eventType EventType, jobState *JobState, systemChanges []SystemChange, context map[string]interface{}) error {
	if !ttd.IsEnabled() {
		return ErrCaptureDisabled
	}
	return ttd.capture.CaptureEvent(jobID, eventType, jobState, systemChanges, context)
}

// FinishRecording completes and stores a job's recording
func (ttd *TimeTravelDebugger) FinishRecording(jobID string, finalJobState *JobState) error {
	if !ttd.IsEnabled() {
		return ErrCaptureDisabled
	}
	return ttd.capture.FinishRecording(jobID, finalJobState)
}

// GetActiveRecordings returns a list of currently active recordings
func (ttd *TimeTravelDebugger) GetActiveRecordings() []string {
	if !ttd.IsEnabled() {
		return []string{}
	}
	return ttd.capture.GetActiveRecordings()
}

// Replay Interface

// LoadRecording loads a recording by ID
func (ttd *TimeTravelDebugger) LoadRecording(recordID string) (*ExecutionRecord, error) {
	return ttd.replay.LoadRecording(recordID)
}

// LoadRecordingByJobID loads a recording by job ID
func (ttd *TimeTravelDebugger) LoadRecordingByJobID(jobID string) (*ExecutionRecord, error) {
	return ttd.replay.LoadRecordingByJobID(jobID)
}

// CreateReplaySession creates a new replay session
func (ttd *TimeTravelDebugger) CreateReplaySession(recordID, userID string) (*ReplaySession, error) {
	return ttd.replay.CreateReplaySession(recordID, userID)
}

// GetReplaySession retrieves an active replay session
func (ttd *TimeTravelDebugger) GetReplaySession(sessionID string) (*ReplaySession, error) {
	return ttd.replay.GetReplaySession(sessionID)
}

// SeekTo seeks to a specific position in the timeline
func (ttd *TimeTravelDebugger) SeekTo(sessionID string, position TimelinePosition) (*ReplayState, error) {
	return ttd.replay.SeekTo(sessionID, position)
}

// SeekToTimestamp seeks to a specific timestamp
func (ttd *TimeTravelDebugger) SeekToTimestamp(sessionID string, timestamp time.Time) (*ReplayState, error) {
	return ttd.replay.SeekToTimestamp(sessionID, timestamp)
}

// StepForward moves one event forward
func (ttd *TimeTravelDebugger) StepForward(sessionID string) (*ReplayState, error) {
	return ttd.replay.StepForward(sessionID)
}

// StepBackward moves one event backward
func (ttd *TimeTravelDebugger) StepBackward(sessionID string) (*ReplayState, error) {
	return ttd.replay.StepBackward(sessionID)
}

// JumpToNextError jumps to the next error event
func (ttd *TimeTravelDebugger) JumpToNextError(sessionID string) (*ReplayState, error) {
	return ttd.replay.JumpToNextError(sessionID)
}

// JumpToNextRetry jumps to the next retry event
func (ttd *TimeTravelDebugger) JumpToNextRetry(sessionID string) (*ReplayState, error) {
	return ttd.replay.JumpToNextRetry(sessionID)
}

// AddBookmark adds a bookmark at the current position
func (ttd *TimeTravelDebugger) AddBookmark(sessionID, description string) error {
	return ttd.replay.AddBookmark(sessionID, description)
}

// GetTimeline returns the complete timeline for a recording
func (ttd *TimeTravelDebugger) GetTimeline(recordID string) ([]TimelinePosition, error) {
	return ttd.replay.GetTimeline(recordID)
}

// CloseSession closes a replay session
func (ttd *TimeTravelDebugger) CloseSession(sessionID string) error {
	return ttd.replay.CloseSession(sessionID)
}

// Management Interface

// GetRecentRecordings returns recently created recordings
func (ttd *TimeTravelDebugger) GetRecentRecordings(limit int) ([]RecordMetadata, error) {
	return ttd.replay.GetRecentRecordings(ttd.ctx, limit)
}

// SearchRecordings searches for recordings matching criteria
func (ttd *TimeTravelDebugger) SearchRecordings(pattern string, limit int) ([]RecordMetadata, error) {
	return ttd.replay.SearchRecordings(ttd.ctx, pattern, limit)
}

// GetConfig returns the current configuration
func (ttd *TimeTravelDebugger) GetConfig() *CaptureConfig {
	ttd.mu.RLock()
	defer ttd.mu.RUnlock()
	return ttd.config
}

// UpdateConfig updates the configuration
func (ttd *TimeTravelDebugger) UpdateConfig(config *CaptureConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	ttd.mu.Lock()
	defer ttd.mu.Unlock()

	ttd.config = config
	ttd.enabled = config.Enabled

	ttd.logger.Info("Time Travel Debugger configuration updated",
		zap.Bool("enabled", config.Enabled),
		zap.Float64("sampling_rate", config.SamplingRate),
	)

	return nil
}

// Helper methods for job middleware integration

// ShouldRecord determines if a job should be recorded based on configuration and job characteristics
func (ttd *TimeTravelDebugger) ShouldRecord(jobID string, jobStatus string, hasRetries bool, isFailure bool) bool {
	if !ttd.IsEnabled() {
		return false
	}
	return ttd.config.ShouldCapture(jobStatus, hasRetries, isFailure)
}

// AutoStartRecording automatically starts recording for jobs that meet criteria
func (ttd *TimeTravelDebugger) AutoStartRecording(jobID string, jobStatus string, hasRetries bool, isFailure bool) error {
	if !ttd.ShouldRecord(jobID, jobStatus, hasRetries, isFailure) {
		return nil
	}

	reason := "auto"
	importance := 5 // medium importance

	if isFailure {
		reason = "failure"
		importance = 7
	} else if hasRetries {
		reason = "retry"
		importance = 6
	}

	return ttd.StartRecording(jobID, reason, importance)
}

// Health check method
func (ttd *TimeTravelDebugger) Health() map[string]interface{} {
	ttd.mu.RLock()
	defer ttd.mu.RUnlock()

	health := map[string]interface{}{
		"enabled":           ttd.enabled,
		"config_enabled":    ttd.config.Enabled,
		"sampling_rate":     ttd.config.SamplingRate,
		"active_recordings": len(ttd.GetActiveRecordings()),
	}

	return health
}
