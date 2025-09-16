package timetraveldebugger

import (
	"errors"
	"fmt"
)

// Common errors for the time travel debugger
var (
	ErrRecordingNotFound     = errors.New("recording not found")
	ErrInvalidEventIndex     = errors.New("invalid event index")
	ErrCorruptedRecording    = errors.New("recording data is corrupted")
	ErrReplaySessionNotFound = errors.New("replay session not found")
	ErrInvalidTimestamp      = errors.New("invalid timestamp")
	ErrCaptureDisabled       = errors.New("event capture is disabled")
	ErrStorageUnavailable    = errors.New("storage backend is unavailable")
	ErrInvalidExportFormat   = errors.New("invalid export format")
	ErrInsufficientData      = errors.New("insufficient data for replay")
	ErrTimelinePosition      = errors.New("invalid timeline position")
	ErrCompressionFailed     = errors.New("failed to compress recording data")
	ErrDecompressionFailed   = errors.New("failed to decompress recording data")
)

// RecordingError represents an error specific to recording operations
type RecordingError struct {
	RecordID string
	JobID    string
	Message  string
	Cause    error
}

func (e RecordingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("recording error for job %s (record %s): %s: %v",
			e.JobID, e.RecordID, e.Message, e.Cause)
	}
	return fmt.Sprintf("recording error for job %s (record %s): %s",
		e.JobID, e.RecordID, e.Message)
}

func (e RecordingError) Unwrap() error {
	return e.Cause
}

// ReplayError represents an error specific to replay operations
type ReplayError struct {
	SessionID string
	Position  TimelinePosition
	Message   string
	Cause     error
}

func (e ReplayError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("replay error in session %s at position %d: %s: %v",
			e.SessionID, e.Position.EventIndex, e.Message, e.Cause)
	}
	return fmt.Sprintf("replay error in session %s at position %d: %s",
		e.SessionID, e.Position.EventIndex, e.Message)
}

func (e ReplayError) Unwrap() error {
	return e.Cause
}

// CaptureError represents an error during event capture
type CaptureError struct {
	JobID     string
	EventType EventType
	Message   string
	Cause     error
}

func (e CaptureError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("capture error for job %s event %s: %s: %v",
			e.JobID, e.EventType, e.Message, e.Cause)
	}
	return fmt.Sprintf("capture error for job %s event %s: %s",
		e.JobID, e.EventType, e.Message)
}

func (e CaptureError) Unwrap() error {
	return e.Cause
}

// ExportError represents an error during export operations
type ExportError struct {
	RecordID string
	Format   ExportFormat
	Message  string
	Cause    error
}

func (e ExportError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("export error for record %s format %s: %s: %v",
			e.RecordID, e.Format, e.Message, e.Cause)
	}
	return fmt.Sprintf("export error for record %s format %s: %s",
		e.RecordID, e.Format, e.Message)
}

func (e ExportError) Unwrap() error {
	return e.Cause
}

// NewRecordingError creates a new RecordingError
func NewRecordingError(recordID, jobID, message string, cause error) *RecordingError {
	return &RecordingError{
		RecordID: recordID,
		JobID:    jobID,
		Message:  message,
		Cause:    cause,
	}
}

// NewReplayError creates a new ReplayError
func NewReplayError(sessionID string, position TimelinePosition, message string, cause error) *ReplayError {
	return &ReplayError{
		SessionID: sessionID,
		Position:  position,
		Message:   message,
		Cause:     cause,
	}
}

// NewCaptureError creates a new CaptureError
func NewCaptureError(jobID string, eventType EventType, message string, cause error) *CaptureError {
	return &CaptureError{
		JobID:     jobID,
		EventType: eventType,
		Message:   message,
		Cause:     cause,
	}
}

// NewExportError creates a new ExportError
func NewExportError(recordID string, format ExportFormat, message string, cause error) *ExportError {
	return &ExportError{
		RecordID: recordID,
		Format:   format,
		Message:  message,
		Cause:    cause,
	}
}