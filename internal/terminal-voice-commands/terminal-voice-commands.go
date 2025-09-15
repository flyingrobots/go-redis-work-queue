package voice

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// NewVoiceManager creates a new voice command manager
func NewVoiceManager(ctx context.Context, config *VoiceConfig) (*VoiceManager, error) {
	if config == nil {
		config = DefaultVoiceConfig()
	}

	// Initialize privacy manager
	privacy := &PrivacyManager{
		localOnly:    config.LocalOnly,
		recordAudio:  !config.NoAudioRecording,
		logCommands:  !config.SanitizeLogs,
		cloudConsent: false,
	}

	// Initialize data sanitizer
	sanitizer, err := NewDataSanitizer()
	if err != nil {
		return nil, fmt.Errorf("failed to create data sanitizer: %w", err)
	}
	privacy.sanitizer = sanitizer

	// Initialize recognition backend
	var recognizer SpeechRecognizer
	if config.LocalOnly || config.RecognitionBackend == "whisper" {
		recognizer, err = NewWhisperRecognizer(config.Language)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Whisper recognizer: %w", err)
		}
	} else {
		recognizer, err = NewCloudRecognizer(config.RecognitionBackend, config.Language)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cloud recognizer: %w", err)
		}
	}

	// Initialize command processor
	processor, err := NewCommandProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize command processor: %w", err)
	}

	// Initialize audio feedback
	feedback, err := NewAudioFeedback(config.AudioFeedback)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio feedback: %w", err)
	}

	// Initialize wake word detector
	wakeDetector, err := NewWakeWordDetector(config.WakeWord)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wake word detector: %w", err)
	}

	vm := &VoiceManager{
		recognizer:   recognizer,
		processor:    processor,
		feedback:     feedback,
		wakeDetector: wakeDetector,
		privacy:      privacy,
		config:       config,
		commandCh:    make(chan *Command, 10),
		audioCh:      make(chan []byte, 100),
		stateCh:      make(chan VoiceState, 10),
	}

	return vm, nil
}

// DefaultVoiceConfig returns default configuration
func DefaultVoiceConfig() *VoiceConfig {
	return &VoiceConfig{
		WakeWord:            "hey queue",
		RecognitionBackend:  "whisper",
		LocalOnly:           true,
		AudioFeedback:       true,
		Language:            "en",
		ConfidenceThreshold: 0.7,
		ProcessingTimeout:   5 * time.Second,
		NoAudioRecording:    true,
		SanitizeLogs:        true,
	}
}

// Start begins voice command processing
func (v *VoiceManager) Start() error {
	// Start recognition backend
	if err := v.recognizer.StartListening(); err != nil {
		return fmt.Errorf("failed to start recognizer: %w", err)
	}

	// Start processing goroutines
	go v.processAudio()
	go v.handleCommands()

	log.Printf("Voice command manager started with backend: %s", v.config.RecognitionBackend)
	return nil
}

// Stop stops voice command processing
func (v *VoiceManager) Stop() error {
	// Stop recognition
	if err := v.recognizer.StopListening(); err != nil {
		log.Printf("Error stopping recognizer: %v", err)
	}

	// Close channels
	close(v.commandCh)
	close(v.audioCh)
	close(v.stateCh)

	// Close components
	if err := v.recognizer.Close(); err != nil {
		log.Printf("Error closing recognizer: %v", err)
	}

	if err := v.feedback.Close(); err != nil {
		log.Printf("Error closing audio feedback: %v", err)
	}

	return nil
}

// ProcessCommand processes a voice command from TUI
func (v *VoiceManager) ProcessCommand(command string, tui TUIController) (*CommandResponse, error) {
	// Create command from text input
	cmd := &Command{
		RawText:   command,
		Timestamp: time.Now(),
		Context:   make(map[string]string),
	}

	// Sanitize command if required
	if v.config.SanitizeLogs {
		cmd.Sanitized = v.privacy.sanitizer.SanitizeCommand(command)
	}

	// Parse command
	if err := v.processor.ParseCommand(cmd); err != nil {
		return &CommandResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse command: %v", err),
		}, err
	}

	// Execute command
	return v.executeCommand(cmd, tui)
}

// ToggleListening toggles voice listening on/off
func (v *VoiceManager) ToggleListening() error {
	v.listening = !v.listening

	if v.listening {
		v.stateCh <- VoiceStateListening
		return v.feedback.PlayConfirmationSound()
	} else {
		v.stateCh <- VoiceStateIdle
		return v.feedback.PlayConfirmationSound()
	}
}

// GetState returns current voice manager state
func (v *VoiceManager) GetState() VoiceState {
	if v.processing {
		return VoiceStateProcessing
	}
	if v.listening {
		return VoiceStateListening
	}
	return VoiceStateIdle
}

// GetMetrics returns voice command metrics
func (v *VoiceManager) GetMetrics() *VoiceMetrics {
	// This would be implemented with actual metrics collection
	return &VoiceMetrics{
		CommandsProcessed:   0,
		RecognitionFailures: 0,
		AverageLatency:      0,
		AverageAccuracy:     0,
		WakeWordTriggers:    0,
		CommandsByIntent:    make(map[Intent]int64),
		ErrorsByType:        make(map[string]int64),
		LastUpdated:         time.Now(),
	}
}

// processAudio handles audio processing pipeline
func (v *VoiceManager) processAudio() {
	buffer := make([]byte, 0, 16000*2) // 1 second buffer at 16kHz

	for audioData := range v.audioCh {
		buffer = append(buffer, audioData...)

		// Process in chunks
		if len(buffer) >= 16000*2 { // 1 second of audio
			v.processAudioChunk(buffer)
			buffer = buffer[16000:] // Keep some overlap
		}
	}
}

// processAudioChunk processes a chunk of audio data
func (v *VoiceManager) processAudioChunk(audio []byte) {
	// First check for wake word if not already listening
	if !v.listening {
		detected, wakeWord, err := v.wakeDetector.DetectWakeWord(audio)
		if err != nil {
			log.Printf("Wake word detection error: %v", err)
			return
		}

		if detected {
			v.startListening(wakeWord)
			return
		}
	}

	// If listening, process for commands
	if v.listening {
		v.processing = true
		v.stateCh <- VoiceStateProcessing

		recognition, err := v.recognizer.ProcessAudio(audio)
		if err != nil {
			log.Printf("Recognition error: %v", err)
			v.processing = false
			v.stateCh <- VoiceStateError
			return
		}

		// Only process if confidence is high enough
		if recognition.Confidence >= v.config.ConfidenceThreshold {
			v.processRecognition(recognition)
		}

		v.processing = false
		v.stateCh <- VoiceStateListening
	}
}

// startListening activates voice listening mode
func (v *VoiceManager) startListening(wakeWord string) {
	v.listening = true
	v.stateCh <- VoiceStateListening

	log.Printf("Wake word detected: %s", wakeWord)

	// Provide audio feedback
	if err := v.feedback.PlayConfirmationSound(); err != nil {
		log.Printf("Error playing confirmation sound: %v", err)
	}
}

// processRecognition processes a speech recognition result
func (v *VoiceManager) processRecognition(recognition *Recognition) {
	cmd := &Command{
		RawText:    recognition.Text,
		Confidence: recognition.Confidence,
		Timestamp:  recognition.Timestamp,
		Context:    make(map[string]string),
	}

	// Sanitize if required
	if v.config.SanitizeLogs {
		cmd.Sanitized = v.privacy.sanitizer.SanitizeCommand(recognition.Text)
	}

	// Parse and queue command
	if err := v.processor.ParseCommand(cmd); err != nil {
		log.Printf("Failed to parse command '%s': %v", recognition.Text, err)
		v.feedback.PlayErrorSound()
		return
	}

	// Send to command handler
	select {
	case v.commandCh <- cmd:
		log.Printf("Command queued: %s (intent: %s, confidence: %.2f)",
			cmd.RawText, cmd.Intent.String(), cmd.Confidence)
	default:
		log.Printf("Command queue full, dropping command: %s", cmd.RawText)
		v.feedback.PlayErrorSound()
	}
}

// handleCommands processes queued commands
func (v *VoiceManager) handleCommands() {
	for cmd := range v.commandCh {
		// This would integrate with the TUI controller
		log.Printf("Processing command: %s (intent: %s)", cmd.RawText, cmd.Intent.String())

		// Store as last command for context
		v.lastCommand = cmd
	}
}

// executeCommand executes a parsed command
func (v *VoiceManager) executeCommand(cmd *Command, tui TUIController) (*CommandResponse, error) {
	switch cmd.Intent {
	case IntentStatusQuery:
		return v.handleStatusQuery(cmd, tui)
	case IntentWorkerControl:
		return v.handleWorkerControl(cmd, tui)
	case IntentQueueManagement:
		return v.handleQueueManagement(cmd, tui)
	case IntentNavigation:
		return v.handleNavigation(cmd, tui)
	case IntentConfirmation:
		return v.handleConfirmation(cmd, tui)
	case IntentCancel:
		return v.handleCancel(cmd, tui)
	case IntentHelp:
		return v.handleHelp(cmd, tui)
	default:
		return &CommandResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown command intent: %s", cmd.Intent.String()),
		}, fmt.Errorf("unknown command intent: %v", cmd.Intent)
	}
}

// handleStatusQuery handles status query commands
func (v *VoiceManager) handleStatusQuery(cmd *Command, tui TUIController) (*CommandResponse, error) {
	target := cmd.GetEntity(EntityTarget)
	if target == nil {
		return &CommandResponse{
			Success: false,
			Message: "Status query requires a target",
		}, fmt.Errorf("status query requires target")
	}

	switch strings.ToLower(target.Value) {
	case "queue", "queues":
		status := tui.GetQueueStatus()
		message := fmt.Sprintf("High priority: %d jobs, Normal: %d jobs, Low: %d jobs, Total: %d jobs",
			status.High, status.Normal, status.Low, status.Total)

		v.feedback.SpeakResponse(message)

		return &CommandResponse{
			Success: true,
			Message: message,
			Data: map[string]interface{}{
				"high":   status.High,
				"normal": status.Normal,
				"low":    status.Low,
				"total":  status.Total,
			},
			AudioCue: "status_update",
		}, nil

	case "workers":
		workers := tui.GetWorkerStatus()
		activeCount := 0
		for _, worker := range workers {
			if worker.Status == "active" {
				activeCount++
			}
		}

		message := fmt.Sprintf("%d of %d workers are active", activeCount, len(workers))
		v.feedback.SpeakResponse(message)

		return &CommandResponse{
			Success: true,
			Message: message,
			Data: map[string]interface{}{
				"active_workers": activeCount,
				"total_workers":  len(workers),
				"workers":        workers,
			},
			AudioCue: "worker_status",
		}, nil

	case "dlq", "dead letter queue":
		dlqCount := tui.GetDLQCount()
		message := fmt.Sprintf("Dead letter queue contains %d failed jobs", dlqCount)

		// Navigate to DLQ tab
		if err := tui.NavigateToTab("DLQ"); err != nil {
			log.Printf("Failed to navigate to DLQ tab: %v", err)
		}

		v.feedback.SpeakResponse(message)

		return &CommandResponse{
			Success: true,
			Message: message,
			Data: map[string]interface{}{
				"dlq_count": dlqCount,
			},
			AudioCue:   "dlq_status",
			NextAction: "navigate_dlq",
		}, nil

	default:
		message := fmt.Sprintf("Unknown status target: %s", target.Value)
		return &CommandResponse{
			Success: false,
			Message: message,
		}, fmt.Errorf("unknown status target: %s", target.Value)
	}
}

// handleWorkerControl handles worker control commands
func (v *VoiceManager) handleWorkerControl(cmd *Command, tui TUIController) (*CommandResponse, error) {
	workerEntity := cmd.GetEntity(EntityWorkerID)
	actionEntity := cmd.GetEntity(EntityAction)

	if workerEntity == nil {
		return &CommandResponse{
			Success: false,
			Message: "Worker control requires worker ID",
		}, fmt.Errorf("worker control requires worker ID")
	}

	workerID := workerEntity.Value
	action := "drain" // default action
	if actionEntity != nil {
		action = strings.ToLower(actionEntity.Value)
	}

	var err error
	var message string

	switch action {
	case "drain", "stop":
		err = tui.DrainWorker(workerID)
		message = fmt.Sprintf("Worker %s is now draining", workerID)
	case "pause":
		err = tui.PauseWorker(workerID)
		message = fmt.Sprintf("Worker %s has been paused", workerID)
	case "resume", "start":
		err = tui.ResumeWorker(workerID)
		message = fmt.Sprintf("Worker %s has been resumed", workerID)
	default:
		return &CommandResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown worker action: %s", action),
		}, fmt.Errorf("unknown worker action: %s", action)
	}

	if err != nil {
		message = fmt.Sprintf("Failed to %s worker %s: %v", action, workerID, err)
		v.feedback.PlayErrorSound()
		return &CommandResponse{
			Success: false,
			Message: message,
		}, err
	}

	v.feedback.SpeakResponse(message)

	return &CommandResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"worker_id": workerID,
			"action":    action,
		},
		AudioCue: "worker_action_success",
	}, nil
}

// handleQueueManagement handles queue management commands
func (v *VoiceManager) handleQueueManagement(cmd *Command, tui TUIController) (*CommandResponse, error) {
	actionEntity := cmd.GetEntity(EntityAction)
	if actionEntity == nil {
		return &CommandResponse{
			Success: false,
			Message: "Queue management requires an action",
		}, fmt.Errorf("queue management requires action")
	}

	action := strings.ToLower(actionEntity.Value)
	var err error
	var message string

	switch action {
	case "requeue", "retry":
		err = tui.RequeueFailedJobs()
		message = "Requeuing all failed jobs"
	case "clear", "cleanup":
		err = tui.ClearCompletedJobs()
		message = "Clearing all completed jobs"
	default:
		return &CommandResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown queue action: %s", action),
		}, fmt.Errorf("unknown queue action: %s", action)
	}

	if err != nil {
		message = fmt.Sprintf("Failed to %s: %v", action, err)
		v.feedback.PlayErrorSound()
		return &CommandResponse{
			Success: false,
			Message: message,
		}, err
	}

	v.feedback.SpeakResponse(message)

	return &CommandResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"action": action,
		},
		AudioCue: "queue_action_success",
	}, nil
}

// handleNavigation handles navigation commands
func (v *VoiceManager) handleNavigation(cmd *Command, tui TUIController) (*CommandResponse, error) {
	destination := cmd.GetEntity(EntityDestination)
	if destination == nil {
		return &CommandResponse{
			Success: false,
			Message: "Navigation requires a destination",
		}, fmt.Errorf("navigation requires destination")
	}

	dest := strings.ToLower(destination.Value)

	// Map voice destinations to tab names
	tabMap := map[string]string{
		"queue":       "Queue",
		"queues":      "Queue",
		"workers":     "Workers",
		"worker":      "Workers",
		"dlq":         "DLQ",
		"dead letter": "DLQ",
		"stats":       "Stats",
		"statistics":  "Stats",
		"charts":      "Charts",
		"graph":       "Charts",
		"logs":        "Logs",
		"log":         "Logs",
		"config":      "Config",
		"settings":    "Config",
	}

	tabName, exists := tabMap[dest]
	if !exists {
		return &CommandResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown destination: %s", dest),
		}, fmt.Errorf("unknown destination: %s", dest)
	}

	if err := tui.NavigateToTab(tabName); err != nil {
		message := fmt.Sprintf("Failed to navigate to %s: %v", tabName, err)
		v.feedback.PlayErrorSound()
		return &CommandResponse{
			Success: false,
			Message: message,
		}, err
	}

	message := fmt.Sprintf("Navigated to %s tab", tabName)
	v.feedback.SpeakResponse(message)

	return &CommandResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"destination": tabName,
		},
		AudioCue: "navigation_success",
	}, nil
}

// handleConfirmation handles confirmation commands
func (v *VoiceManager) handleConfirmation(cmd *Command, tui TUIController) (*CommandResponse, error) {
	message := "Command confirmed"
	v.feedback.SpeakResponse(message)

	return &CommandResponse{
		Success: true,
		Message: message,
		AudioCue: "confirmation",
	}, nil
}

// handleCancel handles cancel commands
func (v *VoiceManager) handleCancel(cmd *Command, tui TUIController) (*CommandResponse, error) {
	message := "Command cancelled"
	v.feedback.SpeakResponse(message)

	return &CommandResponse{
		Success: true,
		Message: message,
		AudioCue: "cancellation",
	}, nil
}

// handleHelp handles help commands
func (v *VoiceManager) handleHelp(cmd *Command, tui TUIController) (*CommandResponse, error) {
	helpText := "Available commands: show queue status, show workers, drain worker, requeue failed jobs, go to DLQ, help"

	if err := tui.ShowMessage(helpText); err != nil {
		log.Printf("Failed to show help message: %v", err)
	}

	v.feedback.SpeakResponse("Help information displayed")

	return &CommandResponse{
		Success: true,
		Message: helpText,
		Data: map[string]interface{}{
			"help_text": helpText,
		},
		AudioCue: "help",
	}, nil
}