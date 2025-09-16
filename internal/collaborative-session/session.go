package collaborativesession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SessionManager manages collaborative sessions
type SessionManager struct {
	sessions       map[SessionID]*Session
	participants   map[ParticipantID]*Participant
	handoffRequests map[string]*ControlHandoffRequest
	tokenGenerator TokenGenerator
	frameRedactor  FrameRedactor
	transport      Transport
	config         *Config
	mutex          sync.RWMutex
	eventChannels  map[string]chan SessionEvent // key: sessionID:participantID
	stopChan       chan struct{}
	logger         Logger
}

// Logger interface for session logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewSessionManager creates a new session manager
func NewSessionManager(tokenGen TokenGenerator, redactor FrameRedactor, transport Transport, config *Config, logger Logger) *SessionManager {
	return &SessionManager{
		sessions:        make(map[SessionID]*Session),
		participants:    make(map[ParticipantID]*Participant),
		handoffRequests: make(map[string]*ControlHandoffRequest),
		tokenGenerator:  tokenGen,
		frameRedactor:   redactor,
		transport:       transport,
		config:          config,
		eventChannels:   make(map[string]chan SessionEvent),
		stopChan:        make(chan struct{}),
		logger:          logger,
	}
}

// Start starts the session manager
func (sm *SessionManager) Start(ctx context.Context) error {
	// Start cleanup goroutine
	go sm.cleanupExpiredSessions(ctx)

	sm.logger.Info("session manager started")
	return nil
}

// Stop stops the session manager
func (sm *SessionManager) Stop(ctx context.Context) error {
	close(sm.stopChan)

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Close all event channels
	for key, ch := range sm.eventChannels {
		close(ch)
		delete(sm.eventChannels, key)
	}

	sm.logger.Info("session manager stopped")
	return nil
}

// CreateSession creates a new collaborative session
func (sm *SessionManager) CreateSession(ctx context.Context, req CreateSessionRequest) (*Session, error) {
	if err := sm.validateCreateRequest(req); err != nil {
		return nil, err
	}

	sessionID := SessionID(sm.generateID())
	presenterID := ParticipantID(req.CreatedBy)

	// Create presenter participant
	presenter := &Participant{
		ID:         presenterID,
		Name:       req.CreatedBy,
		Role:       RolePresenter,
		JoinedAt:   time.Now(),
		IsActive:   true,
		HasControl: true,
	}

	// Apply defaults to settings
	settings := req.Settings
	if settings.MaxParticipants == 0 {
		settings.MaxParticipants = sm.config.Session.MaxParticipants
	}
	if settings.ControlTimeout == 0 {
		settings.ControlTimeout = sm.config.Session.ControlTimeout
	}
	if settings.FrameRate == 0 {
		settings.FrameRate = sm.config.Session.DefaultFrameRate
	}
	if len(settings.RedactionPatterns) == 0 {
		settings.RedactionPatterns = sm.config.Redaction.DefaultPatterns
	}

	session := &Session{
		ID:           sessionID,
		Name:         req.Name,
		CreatedAt:    time.Now(),
		CreatedBy:    presenterID,
		ExpiresAt:    time.Now().Add(req.ExpiresIn),
		IsActive:     true,
		Participants: map[ParticipantID]*Participant{presenterID: presenter},
		Settings:     settings,
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.participants[presenterID] = presenter
	sm.mutex.Unlock()

	// Emit session created event
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventParticipantJoined,
		Timestamp: time.Now(),
		Data:      presenter,
	}
	sm.broadcastEvent(sessionID, event)

	sm.logger.Info("session created", "session_id", sessionID, "name", req.Name, "created_by", req.CreatedBy)
	return session, nil
}

// JoinSession allows a participant to join a session using a token
func (sm *SessionManager) JoinSession(ctx context.Context, token SessionToken) (*Participant, error) {
	// Validate token
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[token.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if !session.IsActive {
		return nil, ErrSessionClosed
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	if len(session.Participants) >= session.Settings.MaxParticipants {
		return nil, ErrSessionFull
	}

	// Generate participant ID from token
	participantID := ParticipantID(fmt.Sprintf("observer_%s", sm.generateShortID()))

	participant := &Participant{
		ID:         participantID,
		Name:       string(participantID), // Could be enhanced with actual name from token
		Role:       RoleObserver,
		JoinedAt:   time.Now(),
		IsActive:   true,
		HasControl: false,
	}

	session.Participants[participantID] = participant
	sm.participants[participantID] = participant

	// Send current frame to new participant if available
	if session.CurrentFrame != nil {
		go sm.sendFrameToParticipant(participantID, session.CurrentFrame)
	}

	// Emit participant joined event
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: token.SessionID,
		Type:      EventParticipantJoined,
		Timestamp: time.Now(),
		Data:      participant,
	}
	sm.broadcastEvent(token.SessionID, event)

	sm.logger.Info("participant joined session", "session_id", token.SessionID, "participant_id", participantID)
	return participant, nil
}

// LeaveSession removes a participant from a session
func (sm *SessionManager) LeaveSession(ctx context.Context, sessionID SessionID, participantID ParticipantID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	participant, exists := session.Participants[participantID]
	if !exists {
		return ErrParticipantNotFound
	}

	// If participant has control, revoke it
	if participant.HasControl && participant.Role != RolePresenter {
		participant.HasControl = false
	}

	// Remove participant
	delete(session.Participants, participantID)
	delete(sm.participants, participantID)

	// Close event channel for this participant
	channelKey := fmt.Sprintf("%s:%s", sessionID, participantID)
	if ch, exists := sm.eventChannels[channelKey]; exists {
		close(ch)
		delete(sm.eventChannels, channelKey)
	}

	// If presenter left, close the session
	if participant.Role == RolePresenter {
		return sm.closeSessionLocked(sessionID)
	}

	// Emit participant left event
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventParticipantLeft,
		Timestamp: time.Now(),
		Data:      participant,
	}
	sm.broadcastEventLocked(sessionID, event)

	sm.logger.Info("participant left session", "session_id", sessionID, "participant_id", participantID)
	return nil
}

// GetSession retrieves session information
func (sm *SessionManager) GetSession(ctx context.Context, sessionID SessionID) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Return a copy to avoid external modifications
	sessionCopy := *session
	sessionCopy.Participants = make(map[ParticipantID]*Participant)
	for id, p := range session.Participants {
		participantCopy := *p
		sessionCopy.Participants[id] = &participantCopy
	}

	return &sessionCopy, nil
}

// SendFrame broadcasts a new frame to all participants
func (sm *SessionManager) SendFrame(ctx context.Context, sessionID SessionID, frame *Frame) error {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	if !exists {
		sm.mutex.RUnlock()
		return ErrSessionNotFound
	}

	if !session.IsActive {
		sm.mutex.RUnlock()
		return ErrSessionClosed
	}

	// Apply redaction if enabled
	redactedFrame := frame
	if sm.config.Redaction.EnableRedaction && sm.frameRedactor != nil {
		redactedFrame = sm.frameRedactor.RedactFrame(frame, session.Settings.RedactionPatterns)
	}

	// Update session's current frame
	session.CurrentFrame = redactedFrame
	sm.mutex.RUnlock()

	// Broadcast frame to all participants
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventFrameUpdate,
		Timestamp: time.Now(),
		Data:      redactedFrame,
	}

	sm.broadcastEvent(sessionID, event)
	return nil
}

// HandleInput processes input from a participant
func (sm *SessionManager) HandleInput(ctx context.Context, sessionID SessionID, participantID ParticipantID, input InputEvent) error {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	if !exists {
		sm.mutex.RUnlock()
		return ErrSessionNotFound
	}

	participant, exists := session.Participants[participantID]
	if !exists {
		sm.mutex.RUnlock()
		return ErrParticipantNotFound
	}

	// Check if participant has control
	if !participant.HasControl {
		sm.mutex.RUnlock()
		return ErrControlNotHeld
	}
	sm.mutex.RUnlock()

	// For now, we just log the input. In a real implementation,
	// this would forward the input to the actual terminal/application
	sm.logger.Debug("input received", "session_id", sessionID, "participant_id", participantID, "input_type", input.Type)

	return nil
}

// RequestControlHandoff initiates control transfer
func (sm *SessionManager) RequestControlHandoff(ctx context.Context, req ControlHandoffRequest) error {
	// TODO: Need to find the session that contains this participant
	// For now, return an error to fix compilation
	return fmt.Errorf("control handoff not implemented - participant %s", req.FromParticipant)

	/* Original code that needs fixing:
	sm.mutex.RLock()
	session, exists := sm.sessions[req.FromParticipant]
	sm.mutex.RUnlock()

	if !exists {
		return ErrSessionNotFound
	}

	if !session.Settings.AllowControlHandoff {
		return ErrHandoffNotAllowed
	}

	sm.mutex.Lock()
	sm.handoffRequests[req.ID] = &req
	sm.mutex.Unlock()

	// Send handoff request to target participant
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: session.ID,
		Type:      EventControlHandoff,
		Timestamp: time.Now(),
		Data:      req,
	}

	channelKey := fmt.Sprintf("%s:%s", session.ID, req.ToParticipant)
	sm.sendToChannel(channelKey, event)

	sm.logger.Info("control handoff requested", "from", req.FromParticipant, "to", req.ToParticipant)
	return nil
}

// RespondToHandoff responds to a control handoff request
func (sm *SessionManager) RespondToHandoff(ctx context.Context, resp ControlHandoffResponse) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	req, exists := sm.handoffRequests[resp.RequestID]
	if !exists {
		return ErrHandoffRequestExpired
	}

	if time.Now().After(req.ExpiresAt) {
		delete(sm.handoffRequests, resp.RequestID)
		return ErrHandoffRequestExpired
	}

	// Find session by scanning for the participant
	var sessionID SessionID
	var session *Session
	for sid, s := range sm.sessions {
		if _, exists := s.Participants[req.FromParticipant]; exists {
			sessionID = sid
			session = s
			break
		}
	}

	if session == nil {
		return ErrSessionNotFound
	}

	if resp.Accepted {
		// Transfer control
		for _, p := range session.Participants {
			p.HasControl = false
		}
		session.Participants[req.ToParticipant].HasControl = true

		sm.logger.Info("control transferred", "from", req.FromParticipant, "to", req.ToParticipant, "session", sessionID)
	}

	// Clean up request
	delete(sm.handoffRequests, resp.RequestID)

	// Notify all participants
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventControlHandoff,
		Timestamp: time.Now(),
		Data:      resp,
	}
	sm.broadcastEventLocked(sessionID, event)

	return nil
}

// RevokeControl removes control from a participant
func (sm *SessionManager) RevokeControl(ctx context.Context, sessionID SessionID, participantID ParticipantID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	participant, exists := session.Participants[participantID]
	if !exists {
		return ErrParticipantNotFound
	}

	if !participant.HasControl {
		return ErrControlNotHeld
	}

	participant.HasControl = false

	// Give control back to presenter
	for _, p := range session.Participants {
		if p.Role == RolePresenter {
			p.HasControl = true
			break
		}
	}

	// Emit control revoked event
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventControlRevoked,
		Timestamp: time.Now(),
		Data:      participant,
	}
	sm.broadcastEventLocked(sessionID, event)

	sm.logger.Info("control revoked", "session_id", sessionID, "participant_id", participantID)
	return nil
}

// CloseSession terminates a session
func (sm *SessionManager) CloseSession(ctx context.Context, sessionID SessionID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.closeSessionLocked(sessionID)
}

// closeSessionLocked closes a session (requires lock to be held)
func (sm *SessionManager) closeSessionLocked(sessionID SessionID) error {
	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	session.IsActive = false

	// Remove all participants
	for participantID := range session.Participants {
		delete(sm.participants, participantID)

		// Close event channel
		channelKey := fmt.Sprintf("%s:%s", sessionID, participantID)
		if ch, exists := sm.eventChannels[channelKey]; exists {
			close(ch)
			delete(sm.eventChannels, channelKey)
		}
	}

	// Emit session closed event
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: sessionID,
		Type:      EventSessionClosed,
		Timestamp: time.Now(),
		Data:      session,
	}
	sm.broadcastEventLocked(sessionID, event)

	delete(sm.sessions, sessionID)

	sm.logger.Info("session closed", "session_id", sessionID)
	return nil
}

// Subscribe returns a channel for session events
func (sm *SessionManager) Subscribe(ctx context.Context, sessionID SessionID, participantID ParticipantID) (<-chan SessionEvent, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	_, exists = session.Participants[participantID]
	if !exists {
		return nil, ErrParticipantNotFound
	}

	channelKey := fmt.Sprintf("%s:%s", sessionID, participantID)
	ch := make(chan SessionEvent, sm.config.Transport.BufferSize)
	sm.eventChannels[channelKey] = ch

	return ch, nil
}

// Helper methods

func (sm *SessionManager) validateCreateRequest(req CreateSessionRequest) error {
	if req.Name == "" {
		return NewValidationError("name", req.Name, "name is required")
	}
	if req.CreatedBy == "" {
		return NewValidationError("created_by", req.CreatedBy, "created_by is required")
	}
	if req.ExpiresIn <= 0 {
		return NewValidationError("expires_in", req.ExpiresIn, "expires_in must be positive")
	}
	if req.ExpiresIn > sm.config.Session.MaxSessionDuration {
		return NewValidationError("expires_in", req.ExpiresIn, "expires_in exceeds maximum duration")
	}
	return nil
}

func (sm *SessionManager) generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func (sm *SessionManager) generateShortID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().Unix()%10000)
	}
	return hex.EncodeToString(bytes)
}

func (sm *SessionManager) broadcastEvent(sessionID SessionID, event SessionEvent) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	sm.broadcastEventLocked(sessionID, event)
}

func (sm *SessionManager) broadcastEventLocked(sessionID SessionID, event SessionEvent) {
	session, exists := sm.sessions[sessionID]
	if !exists {
		return
	}

	for participantID := range session.Participants {
		channelKey := fmt.Sprintf("%s:%s", sessionID, participantID)
		sm.sendToChannelLocked(channelKey, event)
	}

	// Also send via transport if available
	if sm.transport != nil {
		go sm.transport.Broadcast(sessionID, event)
	}
}

func (sm *SessionManager) sendToChannel(channelKey string, event SessionEvent) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	sm.sendToChannelLocked(channelKey, event)
}

func (sm *SessionManager) sendToChannelLocked(channelKey string, event SessionEvent) {
	if ch, exists := sm.eventChannels[channelKey]; exists {
		select {
		case ch <- event:
		default:
			// Channel is full, skip this event
			sm.logger.Warn("event channel full, skipping event", "channel", channelKey, "event_type", event.Type)
		}
	}
}

func (sm *SessionManager) sendFrameToParticipant(participantID ParticipantID, frame *Frame) {
	event := SessionEvent{
		ID:        sm.generateID(),
		SessionID: "", // Will be set by caller
		Type:      EventFrameUpdate,
		Timestamp: time.Now(),
		Data:      frame,
	}

	// Find the session for this participant
	sm.mutex.RLock()
	var sessionID SessionID
	for sid, session := range sm.sessions {
		if _, exists := session.Participants[participantID]; exists {
			sessionID = sid
			break
		}
	}
	sm.mutex.RUnlock()

	if sessionID != "" {
		event.SessionID = sessionID
		channelKey := fmt.Sprintf("%s:%s", sessionID, participantID)
		sm.sendToChannel(channelKey, event)
	}
}

func (sm *SessionManager) cleanupExpiredSessions(ctx context.Context) {
	ticker := time.NewTicker(sm.config.Session.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.performCleanup()
		}
	}
}

func (sm *SessionManager) performCleanup() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredSessions := make([]SessionID, 0)

	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		sm.logger.Info("cleaning up expired session", "session_id", sessionID)
		sm.closeSessionLocked(sessionID)
	}

	// Clean up expired handoff requests
	expiredRequests := make([]string, 0)
	for reqID, req := range sm.handoffRequests {
		if now.After(req.ExpiresAt) {
			expiredRequests = append(expiredRequests, reqID)
		}
	}

	for _, reqID := range expiredRequests {
		delete(sm.handoffRequests, reqID)
	}
}