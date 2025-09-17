package timetraveldebugger

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ReplayEngine handles timeline reconstruction and playback
type ReplayEngine struct {
	redis  *redis.Client
	logger *zap.Logger

	// Active replay sessions
	sessions map[string]*ReplaySession
	mu       sync.RWMutex
}

// ReplayState represents the current state during replay
type ReplayState struct {
	Position      TimelinePosition `json:"position"`
	JobState      *JobState        `json:"job_state"`
	SystemState   *StateSnapshot   `json:"system_state"`
	PreviousState *StateSnapshot   `json:"previous_state,omitempty"`
	Changes       []SystemChange   `json:"changes,omitempty"`
}

// NewReplayEngine creates a new replay engine
func NewReplayEngine(redis *redis.Client, logger *zap.Logger) *ReplayEngine {
	return &ReplayEngine{
		redis:    redis,
		logger:   logger,
		sessions: make(map[string]*ReplaySession),
	}
}

// LoadRecording loads a recording from storage
func (re *ReplayEngine) LoadRecording(recordID string) (*ExecutionRecord, error) {
	key := fmt.Sprintf("time_travel:recording:%s", recordID)

	data, err := re.redis.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRecordingNotFound
		}
		return nil, fmt.Errorf("failed to load recording: %w", err)
	}

	// Try to decompress if it looks like compressed data
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewRecordingError(recordID, "", "failed to decompress recording", err)
		}
		defer reader.Close()

		var decompressed bytes.Buffer
		_, err = decompressed.ReadFrom(reader)
		if err != nil {
			return nil, NewRecordingError(recordID, "", "failed to read decompressed data", err)
		}
		data = decompressed.Bytes()
	}

	var record ExecutionRecord
	err = json.Unmarshal(data, &record)
	if err != nil {
		return nil, NewRecordingError(recordID, "", "failed to unmarshal recording", err)
	}

	// Validate the recording
	if err := re.validateRecording(&record); err != nil {
		return nil, NewRecordingError(recordID, record.JobID, "recording validation failed", err)
	}

	return &record, nil
}

// LoadRecordingByJobID loads a recording by job ID
func (re *ReplayEngine) LoadRecordingByJobID(jobID string) (*ExecutionRecord, error) {
	jobKey := fmt.Sprintf("time_travel:job_index:%s", jobID)

	recordID, err := re.redis.Get(context.Background(), jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrRecordingNotFound
		}
		return nil, fmt.Errorf("failed to find recording for job %s: %w", jobID, err)
	}

	return re.LoadRecording(recordID)
}

// CreateReplaySession creates a new replay session for a recording
func (re *ReplayEngine) CreateReplaySession(recordID, userID string) (*ReplaySession, error) {
	recording, err := re.LoadRecording(recordID)
	if err != nil {
		return nil, err
	}

	sessionID := re.generateSessionID()

	session := &ReplaySession{
		ID:        sessionID,
		RecordID:  recordID,
		UserID:    userID,
		StartTime: time.Now(),
		CurrentPos: TimelinePosition{
			EventIndex:  0,
			Timestamp:   recording.StartTime,
			Description: "Recording start",
		},
		PlaybackSpeed: 1.0,
		IsPlaying:     false,
		Bookmarks:     make([]TimelinePosition, 0),
		Annotations:   make(map[string]string),
	}

	re.mu.Lock()
	re.sessions[sessionID] = session
	re.mu.Unlock()

	re.logger.Info("Created replay session",
		zap.String("session_id", sessionID),
		zap.String("record_id", recordID),
		zap.String("user_id", userID),
		zap.String("job_id", recording.JobID),
	)

	return session, nil
}

// GetReplaySession retrieves an active replay session
func (re *ReplayEngine) GetReplaySession(sessionID string) (*ReplaySession, error) {
	re.mu.RLock()
	session, exists := re.sessions[sessionID]
	re.mu.RUnlock()

	if !exists {
		return nil, ErrReplaySessionNotFound
	}

	return session, nil
}

// SeekTo seeks to a specific position in the timeline
func (re *ReplayEngine) SeekTo(sessionID string, position TimelinePosition) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	// Validate position
	if position.EventIndex < 0 || position.EventIndex >= len(recording.Events) {
		return nil, NewReplayError(sessionID, position, "event index out of bounds", ErrInvalidEventIndex)
	}

	// Update session position
	session.CurrentPos = position

	// Reconstruct state at this position
	state, err := re.reconstructStateAtPosition(recording, position.EventIndex)
	if err != nil {
		return nil, NewReplayError(sessionID, position, "failed to reconstruct state", err)
	}

	re.logger.Debug("Seeked to position",
		zap.String("session_id", sessionID),
		zap.Int("event_index", position.EventIndex),
		zap.Time("timestamp", position.Timestamp),
	)

	return state, nil
}

// SeekToTimestamp seeks to a specific timestamp in the timeline
func (re *ReplayEngine) SeekToTimestamp(sessionID string, timestamp time.Time) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	// Find the event closest to the timestamp
	eventIndex := re.findEventIndexByTimestamp(recording, timestamp)
	if eventIndex < 0 {
		return nil, NewReplayError(sessionID, session.CurrentPos, "no events found for timestamp", ErrInvalidTimestamp)
	}

	position := TimelinePosition{
		EventIndex:  eventIndex,
		Timestamp:   recording.Events[eventIndex].Timestamp,
		Description: fmt.Sprintf("Event: %s", recording.Events[eventIndex].Type),
	}

	return re.SeekTo(sessionID, position)
}

// StepForward moves one event forward in the timeline
func (re *ReplayEngine) StepForward(sessionID string) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	nextIndex := session.CurrentPos.EventIndex + 1
	if nextIndex >= len(recording.Events) {
		return nil, NewReplayError(sessionID, session.CurrentPos, "already at end of timeline", ErrTimelinePosition)
	}

	nextEvent := recording.Events[nextIndex]
	position := TimelinePosition{
		EventIndex:  nextIndex,
		Timestamp:   nextEvent.Timestamp,
		Description: fmt.Sprintf("Event: %s", nextEvent.Type),
	}

	return re.SeekTo(sessionID, position)
}

// StepBackward moves one event backward in the timeline
func (re *ReplayEngine) StepBackward(sessionID string) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentPos.EventIndex <= 0 {
		return nil, NewReplayError(sessionID, session.CurrentPos, "already at beginning of timeline", ErrTimelinePosition)
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	prevIndex := session.CurrentPos.EventIndex - 1
	prevEvent := recording.Events[prevIndex]
	position := TimelinePosition{
		EventIndex:  prevIndex,
		Timestamp:   prevEvent.Timestamp,
		Description: fmt.Sprintf("Event: %s", prevEvent.Type),
	}

	return re.SeekTo(sessionID, position)
}

// JumpToNextError finds and jumps to the next error event
func (re *ReplayEngine) JumpToNextError(sessionID string) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	// Find next error starting from current position
	for i := session.CurrentPos.EventIndex + 1; i < len(recording.Events); i++ {
		event := recording.Events[i]
		if event.Type == EventFailed || event.Type == EventDLQ {
			position := TimelinePosition{
				EventIndex:  i,
				Timestamp:   event.Timestamp,
				Description: fmt.Sprintf("Error: %s", event.Type),
			}
			return re.SeekTo(sessionID, position)
		}
	}

	return nil, NewReplayError(sessionID, session.CurrentPos, "no error events found after current position", ErrTimelinePosition)
}

// JumpToNextRetry finds and jumps to the next retry event
func (re *ReplayEngine) JumpToNextRetry(sessionID string) (*ReplayState, error) {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return nil, err
	}

	recording, err := re.LoadRecording(session.RecordID)
	if err != nil {
		return nil, err
	}

	// Find next retry starting from current position
	for i := session.CurrentPos.EventIndex + 1; i < len(recording.Events); i++ {
		event := recording.Events[i]
		if event.Type == EventRetrying {
			position := TimelinePosition{
				EventIndex:  i,
				Timestamp:   event.Timestamp,
				Description: fmt.Sprintf("Retry: %s", event.Type),
			}
			return re.SeekTo(sessionID, position)
		}
	}

	return nil, NewReplayError(sessionID, session.CurrentPos, "no retry events found after current position", ErrTimelinePosition)
}

// AddBookmark adds a bookmark at the current position
func (re *ReplayEngine) AddBookmark(sessionID, description string) error {
	session, err := re.GetReplaySession(sessionID)
	if err != nil {
		return err
	}

	bookmark := session.CurrentPos
	bookmark.Description = description
	bookmark.IsBookmark = true

	session.Bookmarks = append(session.Bookmarks, bookmark)

	re.logger.Debug("Added bookmark",
		zap.String("session_id", sessionID),
		zap.String("description", description),
		zap.Int("event_index", bookmark.EventIndex),
	)

	return nil
}

// GetTimeline returns the complete timeline for a recording
func (re *ReplayEngine) GetTimeline(recordID string) ([]TimelinePosition, error) {
	recording, err := re.LoadRecording(recordID)
	if err != nil {
		return nil, err
	}

	timeline := make([]TimelinePosition, len(recording.Events))
	for i, event := range recording.Events {
		timeline[i] = TimelinePosition{
			EventIndex:  i,
			Timestamp:   event.Timestamp,
			Description: fmt.Sprintf("%s", event.Type),
		}
	}

	return timeline, nil
}

// CloseSession closes and cleans up a replay session
func (re *ReplayEngine) CloseSession(sessionID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	session, exists := re.sessions[sessionID]
	if !exists {
		return ErrReplaySessionNotFound
	}

	delete(re.sessions, sessionID)

	re.logger.Info("Closed replay session",
		zap.String("session_id", sessionID),
		zap.String("record_id", session.RecordID),
		zap.Duration("duration", time.Since(session.StartTime)),
	)

	return nil
}

// Helper methods

func (re *ReplayEngine) reconstructStateAtPosition(recording *ExecutionRecord, eventIndex int) (*ReplayState, error) {
	if eventIndex < 0 || eventIndex >= len(recording.Events) {
		return nil, ErrInvalidEventIndex
	}

	currentEvent := recording.Events[eventIndex]

	// Get the job state after this event
	var jobState *JobState
	if currentEvent.StateChange.JobStateAfter != nil {
		jobState = currentEvent.StateChange.JobStateAfter
	}

	// Find the nearest snapshot before or at this position
	var systemState *StateSnapshot
	for _, snapshot := range recording.Snapshots {
		if !snapshot.Timestamp.After(currentEvent.Timestamp) {
			if systemState == nil || snapshot.Timestamp.After(systemState.Timestamp) {
				snap := snapshot // copy to avoid pointer issues
				systemState = &snap
			}
		}
	}

	// Get previous state for comparison
	var previousState *StateSnapshot
	if eventIndex > 0 {
		prevEvent := recording.Events[eventIndex-1]
		for _, snapshot := range recording.Snapshots {
			if !snapshot.Timestamp.After(prevEvent.Timestamp) {
				if previousState == nil || snapshot.Timestamp.After(previousState.Timestamp) {
					snap := snapshot // copy to avoid pointer issues
					previousState = &snap
				}
			}
		}
	}

	return &ReplayState{
		Position: TimelinePosition{
			EventIndex:  eventIndex,
			Timestamp:   currentEvent.Timestamp,
			Description: fmt.Sprintf("Event: %s", currentEvent.Type),
		},
		JobState:      jobState,
		SystemState:   systemState,
		PreviousState: previousState,
		Changes:       currentEvent.StateChange.SystemChanges,
	}, nil
}

func (re *ReplayEngine) findEventIndexByTimestamp(recording *ExecutionRecord, timestamp time.Time) int {
	if len(recording.Events) == 0 {
		return -1
	}

	// Binary search for closest timestamp
	left, right := 0, len(recording.Events)-1
	closest := 0
	minDiff := time.Duration(1<<63 - 1) // max duration

	for left <= right {
		mid := (left + right) / 2
		diff := timestamp.Sub(recording.Events[mid].Timestamp)
		if diff < 0 {
			diff = -diff
		}

		if diff < minDiff {
			minDiff = diff
			closest = mid
		}

		if recording.Events[mid].Timestamp.Before(timestamp) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return closest
}

func (re *ReplayEngine) validateRecording(record *ExecutionRecord) error {
	if record.JobID == "" {
		return fmt.Errorf("missing job ID")
	}

	if len(record.Events) == 0 {
		return ErrInsufficientData
	}

	// Validate events are sorted by timestamp
	for i := 1; i < len(record.Events); i++ {
		if record.Events[i].Timestamp.Before(record.Events[i-1].Timestamp) {
			return fmt.Errorf("events not in chronological order at index %d", i)
		}
	}

	return nil
}

func (re *ReplayEngine) generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// SearchRecordings searches for recordings matching criteria
func (re *ReplayEngine) SearchRecordings(ctx context.Context, pattern string, limit int) ([]RecordMetadata, error) {
	// This would implement a search across stored recordings
	// For now, return empty results
	return []RecordMetadata{}, nil
}

// GetRecentRecordings returns recently created recordings
func (re *ReplayEngine) GetRecentRecordings(ctx context.Context, limit int) ([]RecordMetadata, error) {
	// Scan for recent recording keys
	keys, err := re.redis.Keys(ctx, "time_travel:recording:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan for recordings: %w", err)
	}

	if len(keys) == 0 {
		return []RecordMetadata{}, nil
	}

	// Sort keys and limit
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}

	var recordings []RecordMetadata
	for _, key := range keys {
		data, err := re.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var record ExecutionRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		recordings = append(recordings, record.Metadata)
	}

	return recordings, nil
}
