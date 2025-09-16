package voice

import (
	"context"
	"testing"
	"time"
)

func TestNewVoiceManager(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	if vm == nil {
		t.Fatal("Voice manager is nil")
	}

	if vm.config.WakeWord != "hey queue" {
		t.Errorf("Expected wake word 'hey queue', got '%s'", vm.config.WakeWord)
	}

	if vm.config.RecognitionBackend != "whisper" {
		t.Errorf("Expected backend 'whisper', got '%s'", vm.config.RecognitionBackend)
	}
}

func TestVoiceManagerStartStop(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	// Test start
	if err := vm.Start(); err != nil {
		t.Fatalf("Failed to start voice manager: %v", err)
	}

	// Test state
	state := vm.GetState()
	if state != VoiceStateIdle {
		t.Errorf("Expected state idle, got %v", state)
	}

	// Test stop
	if err := vm.Stop(); err != nil {
		t.Fatalf("Failed to stop voice manager: %v", err)
	}
}

func TestToggleListening(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	if err := vm.Start(); err != nil {
		t.Fatalf("Failed to start voice manager: %v", err)
	}
	defer vm.Stop()

	// Initially not listening
	if vm.listening {
		t.Error("Expected listening to be false initially")
	}

	// Toggle on
	if err := vm.ToggleListening(); err != nil {
		t.Fatalf("Failed to toggle listening: %v", err)
	}

	if !vm.listening {
		t.Error("Expected listening to be true after toggle")
	}

	// Toggle off
	if err := vm.ToggleListening(); err != nil {
		t.Fatalf("Failed to toggle listening: %v", err)
	}

	if vm.listening {
		t.Error("Expected listening to be false after second toggle")
	}
}

func TestProcessCommand(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	tui := &MockTUIController{}

	tests := []struct {
		name           string
		command        string
		expectedIntent Intent
		expectedError  bool
	}{
		{
			name:           "status query",
			command:        "show queue status",
			expectedIntent: IntentStatusQuery,
			expectedError:  false,
		},
		{
			name:           "worker control",
			command:        "drain worker 3",
			expectedIntent: IntentWorkerControl,
			expectedError:  false,
		},
		{
			name:           "navigation",
			command:        "go to workers",
			expectedIntent: IntentNavigation,
			expectedError:  false,
		},
		{
			name:           "confirmation",
			command:        "yes",
			expectedIntent: IntentConfirmation,
			expectedError:  false,
		},
		{
			name:           "help",
			command:        "help",
			expectedIntent: IntentHelp,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := vm.ProcessCommand(tt.command, tui)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectedError && response != nil {
				if !response.Success {
					t.Errorf("Expected successful response, got: %s", response.Message)
				}
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	metrics := vm.GetMetrics()
	if metrics == nil {
		t.Fatal("Metrics is nil")
	}

	if metrics.CommandsByIntent == nil {
		t.Error("CommandsByIntent map is nil")
	}

	if metrics.ErrorsByType == nil {
		t.Error("ErrorsByType map is nil")
	}
}

// MockTUIController for testing
type MockTUIController struct {
	queueStatus   *QueueStatus
	workerStatus  []WorkerStatus
	dlqCount      int
	lastNavTab    string
	lastMessage   string
	workerActions map[string]string
}

func (m *MockTUIController) GetQueueStatus() *QueueStatus {
	if m.queueStatus == nil {
		return &QueueStatus{
			High:   10,
			Normal: 25,
			Low:    5,
			Total:  40,
		}
	}
	return m.queueStatus
}

func (m *MockTUIController) GetWorkerStatus() []WorkerStatus {
	if m.workerStatus == nil {
		return []WorkerStatus{
			{ID: "1", Status: "active", Queue: "high", Jobs: 3},
			{ID: "2", Status: "active", Queue: "normal", Jobs: 5},
			{ID: "3", Status: "paused", Queue: "low", Jobs: 0},
		}
	}
	return m.workerStatus
}

func (m *MockTUIController) GetDLQCount() int {
	return m.dlqCount
}

func (m *MockTUIController) NavigateToTab(tab string) error {
	m.lastNavTab = tab
	return nil
}

func (m *MockTUIController) DrainWorker(id string) error {
	if m.workerActions == nil {
		m.workerActions = make(map[string]string)
	}
	m.workerActions[id] = "drain"
	return nil
}

func (m *MockTUIController) PauseWorker(id string) error {
	if m.workerActions == nil {
		m.workerActions = make(map[string]string)
	}
	m.workerActions[id] = "pause"
	return nil
}

func (m *MockTUIController) ResumeWorker(id string) error {
	if m.workerActions == nil {
		m.workerActions = make(map[string]string)
	}
	m.workerActions[id] = "resume"
	return nil
}

func (m *MockTUIController) RequeueFailedJobs() error {
	return nil
}

func (m *MockTUIController) ClearCompletedJobs() error {
	return nil
}

func (m *MockTUIController) ShowMessage(message string) error {
	m.lastMessage = message
	return nil
}

func TestDefaultVoiceConfig(t *testing.T) {
	config := DefaultVoiceConfig()

	if config.WakeWord != "hey queue" {
		t.Errorf("Expected wake word 'hey queue', got '%s'", config.WakeWord)
	}

	if config.RecognitionBackend != "whisper" {
		t.Errorf("Expected backend 'whisper', got '%s'", config.RecognitionBackend)
	}

	if !config.LocalOnly {
		t.Error("Expected LocalOnly to be true")
	}

	if !config.AudioFeedback {
		t.Error("Expected AudioFeedback to be true")
	}

	if config.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", config.Language)
	}

	if config.ConfidenceThreshold != 0.7 {
		t.Errorf("Expected confidence threshold 0.7, got %f", config.ConfidenceThreshold)
	}

	if config.ProcessingTimeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", config.ProcessingTimeout)
	}
}

func TestHandleStatusQuery(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	tui := &MockTUIController{}

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "queue status",
			command:  "show queue status",
			expected: "Total: 40 jobs",
		},
		{
			name:     "worker status",
			command:  "show workers",
			expected: "2 of 3 workers are active",
		},
		{
			name:     "dlq status",
			command:  "show dlq",
			expected: "Dead letter queue contains 0 failed jobs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := vm.ProcessCommand(tt.command, tui)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response == nil {
				t.Fatal("Response is nil")
			}

			if !response.Success {
				t.Errorf("Expected successful response, got: %s", response.Message)
			}

			// Check if expected text is in response
			if tt.expected != "" && !containsText(response.Message, tt.expected) {
				t.Errorf("Expected response to contain '%s', got '%s'", tt.expected, response.Message)
			}
		})
	}
}

func TestHandleWorkerControl(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	tui := &MockTUIController{}

	tests := []struct {
		name           string
		command        string
		expectedAction string
		workerID       string
	}{
		{
			name:           "drain worker",
			command:        "drain worker 1",
			expectedAction: "drain",
			workerID:       "1",
		},
		{
			name:           "pause worker",
			command:        "pause worker 2",
			expectedAction: "pause",
			workerID:       "2",
		},
		{
			name:           "resume worker",
			command:        "resume worker 3",
			expectedAction: "resume",
			workerID:       "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := vm.ProcessCommand(tt.command, tui)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response == nil {
				t.Fatal("Response is nil")
			}

			if !response.Success {
				t.Errorf("Expected successful response, got: %s", response.Message)
			}

			// Check if action was recorded
			if tui.workerActions == nil {
				t.Fatal("Worker actions not recorded")
			}

			action, exists := tui.workerActions[tt.workerID]
			if !exists {
				t.Errorf("No action recorded for worker %s", tt.workerID)
			}

			if action != tt.expectedAction {
				t.Errorf("Expected action '%s', got '%s'", tt.expectedAction, action)
			}
		})
	}
}

func TestHandleNavigation(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	tui := &MockTUIController{}

	tests := []struct {
		name        string
		command     string
		expectedTab string
	}{
		{
			name:        "navigate to workers",
			command:     "go to workers",
			expectedTab: "Workers",
		},
		{
			name:        "navigate to dlq",
			command:     "show dlq",
			expectedTab: "DLQ",
		},
		{
			name:        "navigate to stats",
			command:     "go to stats",
			expectedTab: "Stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := vm.ProcessCommand(tt.command, tui)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response == nil {
				t.Fatal("Response is nil")
			}

			if !response.Success {
				t.Errorf("Expected successful response, got: %s", response.Message)
			}

			if tui.lastNavTab != tt.expectedTab {
				t.Errorf("Expected navigation to '%s', got '%s'", tt.expectedTab, tui.lastNavTab)
			}
		})
	}
}

func TestVoiceStateTransitions(t *testing.T) {
	ctx := context.Background()
	config := DefaultVoiceConfig()

	vm, err := NewVoiceManager(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create voice manager: %v", err)
	}

	// Initial state should be idle
	state := vm.GetState()
	if state != VoiceStateIdle {
		t.Errorf("Expected initial state idle, got %v", state)
	}

	// Start listening
	if err := vm.ToggleListening(); err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	state = vm.GetState()
	if state != VoiceStateListening {
		t.Errorf("Expected listening state, got %v", state)
	}

	// Stop listening
	if err := vm.ToggleListening(); err != nil {
		t.Fatalf("Failed to stop listening: %v", err)
	}

	state = vm.GetState()
	if state != VoiceStateIdle {
		t.Errorf("Expected idle state after stopping, got %v", state)
	}
}

// Helper function to check if text contains expected substring
func containsText(text, expected string) bool {
	return len(text) >= len(expected) &&
		   (text == expected ||
		    stringContains(text, expected))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}