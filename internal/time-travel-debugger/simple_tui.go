package timetraveldebugger

import (
	"fmt"
	"strings"
	"time"
)

// SimpleTUI provides a simple text-based interface for time travel debugging
type SimpleTUI struct {
	engine        *ReplayEngine
	activeSession *ReplaySession
	currentState  *ReplayState
}

// NewSimpleTUI creates a new simple TUI for time travel debugging
func NewSimpleTUI(engine *ReplayEngine) *SimpleTUI {
	return &SimpleTUI{
		engine: engine,
	}
}

// RenderRecordingsList returns a formatted list of available recordings
func (tui *SimpleTUI) RenderRecordingsList(recordings []RecordMetadata) string {
	if len(recordings) == 0 {
		return "No recordings available.\n"
	}

	var output strings.Builder
	output.WriteString("Available Recordings:\n")
	output.WriteString("====================\n")

	for i, record := range recordings {
		output.WriteString(fmt.Sprintf("%d. %s (%s)\n",
			i+1,
			record.RecordID[:8],
			record.CreatedAt.Format("Jan 2 15:04")))
		output.WriteString(fmt.Sprintf("   Reason: %s, Events: %d\n",
			record.Reason,
			record.EventCount))
		output.WriteString("\n")
	}

	return output.String()
}

// StartSession creates a new replay session
func (tui *SimpleTUI) StartSession(recordID, userID string) error {
	session, err := tui.engine.CreateReplaySession(recordID, userID)
	if err != nil {
		return err
	}

	tui.activeSession = session

	// Seek to the beginning
	return tui.SeekToPosition(0)
}

// SeekToPosition seeks to a specific event index
func (tui *SimpleTUI) SeekToPosition(eventIndex int) error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	position := TimelinePosition{
		EventIndex: eventIndex,
	}

	state, err := tui.engine.SeekTo(tui.activeSession.ID, position)
	if err != nil {
		return err
	}

	tui.currentState = state
	return nil
}

// StepForward moves one event forward
func (tui *SimpleTUI) StepForward() error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	state, err := tui.engine.StepForward(tui.activeSession.ID)
	if err != nil {
		return err
	}

	tui.currentState = state
	return nil
}

// StepBackward moves one event backward
func (tui *SimpleTUI) StepBackward() error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	state, err := tui.engine.StepBackward(tui.activeSession.ID)
	if err != nil {
		return err
	}

	tui.currentState = state
	return nil
}

// RenderCurrentState returns a formatted view of the current state
func (tui *SimpleTUI) RenderCurrentState() string {
	if tui.currentState == nil {
		return "No state available.\n"
	}

	var output strings.Builder

	// Header
	output.WriteString("Time Travel Debugger - Current State\n")
	output.WriteString("===================================\n\n")

	// Position information
	output.WriteString(fmt.Sprintf("Position: Event %d\n",
		tui.currentState.Position.EventIndex))
	output.WriteString(fmt.Sprintf("Timestamp: %s\n",
		tui.currentState.Position.Timestamp.Format(time.RFC3339)))
	output.WriteString(fmt.Sprintf("Description: %s\n\n",
		tui.currentState.Position.Description))

	// Job state
	if tui.currentState.JobState != nil {
		output.WriteString("Job State:\n")
		output.WriteString(fmt.Sprintf("  ID: %s\n", tui.currentState.JobState.ID))
		output.WriteString(fmt.Sprintf("  Status: %s\n", tui.currentState.JobState.Status))
		output.WriteString(fmt.Sprintf("  Priority: %s\n", tui.currentState.JobState.Priority))
		output.WriteString(fmt.Sprintf("  Retries: %d/%d\n",
			tui.currentState.JobState.Retries,
			tui.currentState.JobState.MaxRetries))

		if tui.currentState.JobState.ErrorMessage != "" {
			output.WriteString(fmt.Sprintf("  Error: %s\n", tui.currentState.JobState.ErrorMessage))
		}
		output.WriteString("\n")
	}

	// System changes
	if len(tui.currentState.Changes) > 0 {
		output.WriteString("System Changes:\n")
		for _, change := range tui.currentState.Changes {
			output.WriteString(fmt.Sprintf("  %s.%s: %v -> %v\n",
				change.Component,
				change.Key,
				change.Before,
				change.After))
		}
		output.WriteString("\n")
	}

	// System metrics
	if tui.currentState.SystemState != nil && tui.currentState.SystemState.SystemMetrics != nil {
		metrics := tui.currentState.SystemState.SystemMetrics
		output.WriteString("System Metrics:\n")
		output.WriteString(fmt.Sprintf("  Queue Length: %d\n", metrics.QueueLength))
		output.WriteString(fmt.Sprintf("  Worker Count: %d\n", metrics.WorkerCount))
		output.WriteString(fmt.Sprintf("  Memory Usage: %.1f MB\n", metrics.MemoryUsageMB))
		output.WriteString(fmt.Sprintf("  CPU Usage: %.1f%%\n", metrics.CPUUsagePercent))
		output.WriteString("\n")
	}

	// Controls
	output.WriteString("Controls:\n")
	output.WriteString("  h - Step backward\n")
	output.WriteString("  l - Step forward\n")
	output.WriteString("  j - Jump to next error\n")
	output.WriteString("  r - Jump to next retry\n")
	output.WriteString("  m - Add bookmark\n")
	output.WriteString("  q - Quit\n")

	return output.String()
}

// RenderTimeline returns a simple timeline view
func (tui *SimpleTUI) RenderTimeline(timeline []TimelinePosition) string {
	if len(timeline) == 0 {
		return "No timeline data available.\n"
	}

	var output strings.Builder
	output.WriteString("Timeline:\n")
	output.WriteString("=========\n")

	currentIndex := -1
	if tui.currentState != nil {
		currentIndex = tui.currentState.Position.EventIndex
	}

	for i, position := range timeline {
		marker := "  "
		if i == currentIndex {
			marker = "> "
		}

		output.WriteString(fmt.Sprintf("%s%d. %s [%s]\n",
			marker,
			i,
			position.Description,
			position.Timestamp.Format("15:04:05")))
	}

	return output.String()
}

// AddBookmark adds a bookmark at the current position
func (tui *SimpleTUI) AddBookmark(description string) error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	return tui.engine.AddBookmark(tui.activeSession.ID, description)
}

// JumpToNextError jumps to the next error event
func (tui *SimpleTUI) JumpToNextError() error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	state, err := tui.engine.JumpToNextError(tui.activeSession.ID)
	if err != nil {
		return err
	}

	tui.currentState = state
	return nil
}

// JumpToNextRetry jumps to the next retry event
func (tui *SimpleTUI) JumpToNextRetry() error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	state, err := tui.engine.JumpToNextRetry(tui.activeSession.ID)
	if err != nil {
		return err
	}

	tui.currentState = state
	return nil
}

// CloseSession closes the current replay session
func (tui *SimpleTUI) CloseSession() error {
	if tui.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	err := tui.engine.CloseSession(tui.activeSession.ID)
	if err != nil {
		return err
	}

	tui.activeSession = nil
	tui.currentState = nil
	return nil
}

// GetSessionInfo returns information about the current session
func (tui *SimpleTUI) GetSessionInfo() string {
	if tui.activeSession == nil {
		return "No active session.\n"
	}

	var output strings.Builder
	output.WriteString("Session Information:\n")
	output.WriteString("===================\n")
	output.WriteString(fmt.Sprintf("Session ID: %s\n", tui.activeSession.ID))
	output.WriteString(fmt.Sprintf("Record ID: %s\n", tui.activeSession.RecordID))
	output.WriteString(fmt.Sprintf("User ID: %s\n", tui.activeSession.UserID))
	output.WriteString(fmt.Sprintf("Started: %s\n",
		tui.activeSession.StartTime.Format(time.RFC3339)))
	output.WriteString(fmt.Sprintf("Playback Speed: %.1fx\n", tui.activeSession.PlaybackSpeed))
	output.WriteString(fmt.Sprintf("Bookmarks: %d\n", len(tui.activeSession.Bookmarks)))

	return output.String()
}