package collaborativesession

import (
	"context"
	"time"
)

// SessionID uniquely identifies a collaborative session
type SessionID string

// ParticipantID uniquely identifies a session participant
type ParticipantID string

// SessionToken represents authentication token for session access
type SessionToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	SessionID SessionID `json:"session_id"`
	CanControl bool     `json:"can_control"`
}

// Participant represents a session participant
type Participant struct {
	ID       ParticipantID `json:"id"`
	Name     string        `json:"name"`
	Role     ParticipantRole `json:"role"`
	JoinedAt time.Time     `json:"joined_at"`
	IsActive bool          `json:"is_active"`
	HasControl bool        `json:"has_control"`
}

// ParticipantRole defines the role of a participant
type ParticipantRole string

const (
	RolePresenter ParticipantRole = "presenter"
	RoleObserver  ParticipantRole = "observer"
)

// Session represents a collaborative session
type Session struct {
	ID           SessionID              `json:"id"`
	Name         string                 `json:"name"`
	CreatedAt    time.Time              `json:"created_at"`
	CreatedBy    ParticipantID          `json:"created_by"`
	ExpiresAt    time.Time              `json:"expires_at"`
	IsActive     bool                   `json:"is_active"`
	Participants map[ParticipantID]*Participant `json:"participants"`
	CurrentFrame *Frame                 `json:"current_frame,omitempty"`
	Settings     SessionSettings        `json:"settings"`
}

// SessionSettings configures session behavior
type SessionSettings struct {
	MaxParticipants   int           `json:"max_participants"`
	AllowControlHandoff bool        `json:"allow_control_handoff"`
	ControlTimeout    time.Duration `json:"control_timeout"`
	FrameRate         int           `json:"frame_rate"` // FPS for frame updates
	RedactionPatterns []string      `json:"redaction_patterns"`
	RequireApproval   bool          `json:"require_approval"`
}

// Frame represents a terminal frame state
type Frame struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Content   [][]Cell  `json:"content"`
	CursorX   int       `json:"cursor_x"`
	CursorY   int       `json:"cursor_y"`
	Title     string    `json:"title,omitempty"`
}

// Cell represents a single terminal cell
type Cell struct {
	Char       rune      `json:"char"`
	Foreground Color     `json:"foreground"`
	Background Color     `json:"background"`
	Style      CellStyle `json:"style"`
}

// Color represents terminal colors
type Color struct {
	R, G, B uint8 `json:"r,g,b"`
	IsDefault bool `json:"is_default"`
}

// CellStyle represents text styling
type CellStyle struct {
	Bold      bool `json:"bold"`
	Italic    bool `json:"italic"`
	Underline bool `json:"underline"`
	Strikethrough bool `json:"strikethrough"`
	Blink     bool `json:"blink"`
}

// FrameDelta represents changes between two frames
type FrameDelta struct {
	FromFrameID string        `json:"from_frame_id"`
	ToFrameID   string        `json:"to_frame_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Changes     []CellChange  `json:"changes"`
	CursorX     int           `json:"cursor_x"`
	CursorY     int           `json:"cursor_y"`
	SizeChanged bool          `json:"size_changed"`
	Width       int           `json:"width,omitempty"`
	Height      int           `json:"height,omitempty"`
}

// CellChange represents a change to a specific cell
type CellChange struct {
	X    int  `json:"x"`
	Y    int  `json:"y"`
	Cell Cell `json:"cell"`
}

// InputEvent represents user input
type InputEvent struct {
	ID           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	ParticipantID ParticipantID `json:"participant_id"`
	Type         InputType     `json:"type"`
	Data         interface{}   `json:"data"`
}

// InputType defines the type of input event
type InputType string

const (
	InputTypeKey    InputType = "key"
	InputTypeMouse  InputType = "mouse"
	InputTypeResize InputType = "resize"
)

// KeyEvent represents a keyboard input
type KeyEvent struct {
	Key       string `json:"key"`
	Modifiers []string `json:"modifiers"`
}

// MouseEvent represents a mouse input
type MouseEvent struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button"`
	Action string `json:"action"` // press, release, move
}

// ResizeEvent represents a terminal resize
type ResizeEvent struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ControlHandoffRequest represents a request to transfer control
type ControlHandoffRequest struct {
	ID            string        `json:"id"`
	FromParticipant ParticipantID `json:"from_participant"`
	ToParticipant ParticipantID `json:"to_participant"`
	RequestedAt   time.Time     `json:"requested_at"`
	ExpiresAt     time.Time     `json:"expires_at"`
	Message       string        `json:"message,omitempty"`
}

// ControlHandoffResponse represents response to handoff request
type ControlHandoffResponse struct {
	RequestID string    `json:"request_id"`
	Accepted  bool      `json:"accepted"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// SessionEvent represents events that occur in a session
type SessionEvent struct {
	ID        string      `json:"id"`
	SessionID SessionID   `json:"session_id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventType defines types of session events
type EventType string

const (
	EventParticipantJoined EventType = "participant_joined"
	EventParticipantLeft   EventType = "participant_left"
	EventControlHandoff    EventType = "control_handoff"
	EventControlRevoked    EventType = "control_revoked"
	EventFrameUpdate       EventType = "frame_update"
	EventSessionClosed     EventType = "session_closed"
)

// SessionServer defines the interface for the collaborative session server
type SessionServer interface {
	// CreateSession creates a new collaborative session
	CreateSession(ctx context.Context, req CreateSessionRequest) (*Session, error)

	// JoinSession allows a participant to join a session
	JoinSession(ctx context.Context, token SessionToken) (*Participant, error)

	// LeaveSession removes a participant from a session
	LeaveSession(ctx context.Context, sessionID SessionID, participantID ParticipantID) error

	// GetSession retrieves session information
	GetSession(ctx context.Context, sessionID SessionID) (*Session, error)

	// SendFrame broadcasts a new frame to all participants
	SendFrame(ctx context.Context, sessionID SessionID, frame *Frame) error

	// HandleInput processes input from a participant
	HandleInput(ctx context.Context, sessionID SessionID, participantID ParticipantID, input InputEvent) error

	// RequestControlHandoff initiates control transfer
	RequestControlHandoff(ctx context.Context, req ControlHandoffRequest) error

	// RespondToHandoff responds to a control handoff request
	RespondToHandoff(ctx context.Context, resp ControlHandoffResponse) error

	// RevokeControl removes control from a participant
	RevokeControl(ctx context.Context, sessionID SessionID, participantID ParticipantID) error

	// CloseSession terminates a session
	CloseSession(ctx context.Context, sessionID SessionID) error

	// Subscribe returns a channel for session events
	Subscribe(ctx context.Context, sessionID SessionID, participantID ParticipantID) (<-chan SessionEvent, error)
}

// CreateSessionRequest represents a request to create a new session
type CreateSessionRequest struct {
	Name      string          `json:"name"`
	CreatedBy string          `json:"created_by"`
	ExpiresIn time.Duration   `json:"expires_in"`
	Settings  SessionSettings `json:"settings"`
}

// TokenGenerator generates session tokens
type TokenGenerator interface {
	GenerateToken(sessionID SessionID, participantID ParticipantID, canControl bool, expiresIn time.Duration) (*SessionToken, error)
	ValidateToken(token string) (*SessionToken, error)
}

// FrameRedactor handles redaction of sensitive content
type FrameRedactor interface {
	RedactFrame(frame *Frame, patterns []string) *Frame
	RedactDelta(delta *FrameDelta, patterns []string) *FrameDelta
}

// Transport handles network communication
type Transport interface {
	Start(ctx context.Context, addr string) error
	Stop(ctx context.Context) error
	Broadcast(sessionID SessionID, event SessionEvent) error
	Send(participantID ParticipantID, event SessionEvent) error
}

// DefaultSessionSettings returns default session configuration
func DefaultSessionSettings() SessionSettings {
	return SessionSettings{
		MaxParticipants:     10,
		AllowControlHandoff: true,
		ControlTimeout:      5 * time.Minute,
		FrameRate:          30,
		RedactionPatterns:  []string{
			`password\s*[:=]\s*\S+`,
			`token\s*[:=]\s*\S+`,
			`key\s*[:=]\s*\S+`,
			`secret\s*[:=]\s*\S+`,
		},
		RequireApproval: true,
	}
}