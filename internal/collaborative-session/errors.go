package collaborativesession

import (
	"errors"
	"fmt"
)

// Common errors for collaborative sessions
var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionExpired        = errors.New("session has expired")
	ErrSessionClosed         = errors.New("session is closed")
	ErrSessionFull           = errors.New("session has reached maximum participants")
	ErrParticipantNotFound   = errors.New("participant not found")
	ErrInvalidToken          = errors.New("invalid session token")
	ErrTokenExpired          = errors.New("session token has expired")
	ErrUnauthorized          = errors.New("unauthorized action")
	ErrControlNotHeld        = errors.New("participant does not have control")
	ErrControlAlreadyHeld    = errors.New("control is already held by another participant")
	ErrHandoffNotAllowed     = errors.New("control handoff is not allowed")
	ErrHandoffRequestExpired = errors.New("handoff request has expired")
	ErrInvalidInput          = errors.New("invalid input event")
	ErrTransportError        = errors.New("transport error")
	ErrRedactionFailed       = errors.New("frame redaction failed")
)

// SessionError wraps session-specific errors with context
type SessionError struct {
	SessionID SessionID
	Operation string
	Err       error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session %s: %s: %v", e.SessionID, e.Operation, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

// NewSessionError creates a new session error
func NewSessionError(sessionID SessionID, operation string, err error) error {
	return &SessionError{
		SessionID: sessionID,
		Operation: operation,
		Err:       err,
	}
}

// ParticipantError wraps participant-specific errors with context
type ParticipantError struct {
	SessionID     SessionID
	ParticipantID ParticipantID
	Operation     string
	Err           error
}

func (e *ParticipantError) Error() string {
	return fmt.Sprintf("session %s participant %s: %s: %v", e.SessionID, e.ParticipantID, e.Operation, e.Err)
}

func (e *ParticipantError) Unwrap() error {
	return e.Err
}

// NewParticipantError creates a new participant error
func NewParticipantError(sessionID SessionID, participantID ParticipantID, operation string, err error) error {
	return &ParticipantError{
		SessionID:     sessionID,
		ParticipantID: participantID,
		Operation:     operation,
		Err:           err,
	}
}

// TokenError wraps token-related errors
type TokenError struct {
	Token     string
	Operation string
	Err       error
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("token error: %s: %v", e.Operation, e.Err)
}

func (e *TokenError) Unwrap() error {
	return e.Err
}

// NewTokenError creates a new token error
func NewTokenError(token, operation string, err error) error {
	return &TokenError{
		Token:     token,
		Operation: operation,
		Err:       err,
	}
}

// ValidationError represents validation failures
type ValidationError struct {
	Field   string
	Value   interface{}
	Reason  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field %s with value %v: %s", e.Field, e.Value, e.Reason)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, reason string) error {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// TransportError wraps transport-related errors
type TransportError struct {
	Operation string
	Address   string
	Err       error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("transport error: %s on %s: %v", e.Operation, e.Address, e.Err)
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

// NewTransportError creates a new transport error
func NewTransportError(operation, address string, err error) error {
	return &TransportError{
		Operation: operation,
		Address:   address,
		Err:       err,
	}
}

// IsSessionNotFound checks if error is session not found
func IsSessionNotFound(err error) bool {
	return errors.Is(err, ErrSessionNotFound)
}

// IsSessionExpired checks if error is session expired
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired)
}

// IsTokenError checks if error is token-related
func IsTokenError(err error) bool {
	return errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired)
}

// IsUnauthorized checks if error is authorization-related
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsControlError checks if error is control-related
func IsControlError(err error) bool {
	return errors.Is(err, ErrControlNotHeld) ||
		   errors.Is(err, ErrControlAlreadyHeld) ||
		   errors.Is(err, ErrHandoffNotAllowed) ||
		   errors.Is(err, ErrHandoffRequestExpired)
}