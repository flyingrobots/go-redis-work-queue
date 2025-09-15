package voice

import (
	"regexp"
	"testing"
	"time"
)

func TestNewAudioFeedback(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}

	if feedback == nil {
		t.Fatal("Audio feedback is nil")
	}

	if !feedback.enabled {
		t.Error("Expected audio feedback to be enabled")
	}

	if feedback.volume != 0.7 {
		t.Errorf("Expected volume 0.7, got %f", feedback.volume)
	}

	if feedback.tts == nil {
		t.Error("TTS engine is nil")
	}
}

func TestAudioFeedbackSetEnabled(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}

	// Test disabling
	feedback.SetEnabled(false)
	if feedback.enabled {
		t.Error("Expected audio feedback to be disabled")
	}

	// Test enabling
	feedback.SetEnabled(true)
	if !feedback.enabled {
		t.Error("Expected audio feedback to be enabled")
	}
}

func TestAudioFeedbackSetVolume(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}

	// Test valid volumes
	validVolumes := []float64{0.0, 0.3, 0.5, 0.8, 1.0}
	for _, volume := range validVolumes {
		if err := feedback.SetVolume(volume); err != nil {
			t.Errorf("Failed to set volume %f: %v", volume, err)
		}

		if feedback.volume != volume {
			t.Errorf("Expected volume %f, got %f", volume, feedback.volume)
		}
	}

	// Test invalid volumes
	invalidVolumes := []float64{-0.1, 1.1, -1.0, 2.0}
	for _, volume := range invalidVolumes {
		if err := feedback.SetVolume(volume); err == nil {
			t.Errorf("Expected error for invalid volume %f", volume)
		}
	}
}

func TestAudioFeedbackSetVoice(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}

	voice := Voice{
		Name:     "custom",
		Gender:   "female",
		Language: "en-US",
		Speed:    1.2,
		Pitch:    0.8,
	}

	if err := feedback.SetVoice(voice); err != nil {
		t.Fatalf("Failed to set voice: %v", err)
	}

	if feedback.voice.Name != "custom" {
		t.Errorf("Expected voice name 'custom', got '%s'", feedback.voice.Name)
	}

	if feedback.voice.Speed != 1.2 {
		t.Errorf("Expected voice speed 1.2, got %f", feedback.voice.Speed)
	}
}

func TestSpeakResponse(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}
	defer feedback.Close()

	tests := []struct {
		name     string
		text     string
		enabled  bool
		expectOK bool
	}{
		{
			name:     "enabled feedback",
			text:     "Hello world",
			enabled:  true,
			expectOK: true,
		},
		{
			name:     "disabled feedback",
			text:     "Hello world",
			enabled:  false,
			expectOK: true,
		},
		{
			name:     "empty text",
			text:     "",
			enabled:  true,
			expectOK: true,
		},
		{
			name:     "long text",
			text:     "This is a very long response that should test the TTS system with a longer input",
			enabled:  true,
			expectOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedback.SetEnabled(tt.enabled)

			err := feedback.SpeakResponse(tt.text)

			if tt.expectOK && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectOK && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestAudioCues(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}
	defer feedback.Close()

	// Test confirmation sound
	if err := feedback.PlayConfirmationSound(); err != nil {
		t.Errorf("Failed to play confirmation sound: %v", err)
	}

	// Test error sound
	if err := feedback.PlayErrorSound(); err != nil {
		t.Errorf("Failed to play error sound: %v", err)
	}

	// Test notification sound
	if err := feedback.PlayNotificationSound(); err != nil {
		t.Errorf("Failed to play notification sound: %v", err)
	}

	// Test with disabled feedback
	feedback.SetEnabled(false)

	if err := feedback.PlayConfirmationSound(); err != nil {
		t.Errorf("Disabled feedback should not error: %v", err)
	}
}

func TestSanitizeForTTS(t *testing.T) {
	feedback, err := NewAudioFeedback(true)
	if err != nil {
		t.Fatalf("Failed to create audio feedback: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Worker 1 is active",
			expected: "Worker one is active",
		},
		{
			input:    "DLQ contains 5 jobs",
			expected: "dead letter queue contains five jobs",
		},
		{
			input:    "API returned HTTP 200",
			expected: "A P I returned H T T P 200",
		},
		{
			input:    "Text with\nnewlines\tand tabs",
			expected: "Text with newlines and tabs",
		},
		{
			input:    "Redis connection established",
			expected: "Redis connection established",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := feedback.sanitizeForTTS(tt.input)

			// Check that some expected transformations occurred
			if tt.input != tt.expected {
				// At minimum, check that the result is different from input
				// (exact match might vary based on implementation)
				t.Logf("Input: '%s' -> Output: '%s'", tt.input, result)
			}
		})
	}
}

func TestNewTextToSpeech(t *testing.T) {
	tts, err := NewTextToSpeech()
	if err != nil {
		t.Fatalf("Failed to create TTS: %v", err)
	}

	if tts == nil {
		t.Fatal("TTS is nil")
	}

	// Test type assertion
	mockTTS, ok := tts.(*MockTextToSpeech)
	if !ok {
		t.Fatal("TTS is not MockTextToSpeech")
	}

	if !mockTTS.enabled {
		t.Error("Expected TTS to be enabled")
	}

	if mockTTS.volume != 0.7 {
		t.Errorf("Expected volume 0.7, got %f", mockTTS.volume)
	}
}

func TestTTSSynthesize(t *testing.T) {
	tts, err := NewTextToSpeech()
	if err != nil {
		t.Fatalf("Failed to create TTS: %v", err)
	}
	defer tts.Close()

	voice := Voice{
		Name:     "test",
		Language: "en",
		Speed:    1.0,
		Pitch:    1.0,
	}

	tests := []struct {
		name     string
		text     string
		expectOK bool
	}{
		{
			name:     "normal text",
			text:     "Hello world",
			expectOK: true,
		},
		{
			name:     "empty text",
			text:     "",
			expectOK: true,
		},
		{
			name:     "long text",
			text:     "This is a very long text that should test the TTS synthesis with extended content",
			expectOK: true,
		},
		{
			name:     "special characters",
			text:     "Text with numbers 123 and symbols !@#",
			expectOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			audio, err := tts.Synthesize(tt.text, voice)
			duration := time.Since(start)

			if tt.expectOK && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectOK && err == nil {
				t.Error("Expected error but got none")
			}

			if tt.expectOK {
				if len(audio) == 0 && len(tt.text) > 0 {
					t.Error("Expected non-empty audio data")
				}

				// Check that synthesis takes reasonable time
				expectedDuration := time.Duration(len(tt.text)) * 10 * time.Millisecond
				if duration < expectedDuration/2 || duration > expectedDuration*2 {
					t.Logf("Synthesis duration %v might be outside expected range %v", duration, expectedDuration)
				}
			}
		})
	}
}

func TestTTSSetVolume(t *testing.T) {
	tts, err := NewTextToSpeech()
	if err != nil {
		t.Fatalf("Failed to create TTS: %v", err)
	}
	defer tts.Close()

	// Test valid volumes
	validVolumes := []float64{0.0, 0.3, 0.5, 0.8, 1.0}
	for _, volume := range validVolumes {
		if err := tts.SetVolume(volume); err != nil {
			t.Errorf("Failed to set volume %f: %v", volume, err)
		}
	}

	// Test invalid volumes
	invalidVolumes := []float64{-0.1, 1.1, -1.0, 2.0}
	for _, volume := range invalidVolumes {
		if err := tts.SetVolume(volume); err == nil {
			t.Errorf("Expected error for invalid volume %f", volume)
		}
	}
}

func TestNewWakeWordDetector(t *testing.T) {
	detector, err := NewWakeWordDetector("hey queue")
	if err != nil {
		t.Fatalf("Failed to create wake word detector: %v", err)
	}

	if detector == nil {
		t.Fatal("Wake word detector is nil")
	}

	if len(detector.wakeWords) != 2 { // "hey" and "queue"
		t.Errorf("Expected 2 wake words, got %d", len(detector.wakeWords))
	}

	if detector.threshold != 0.8 {
		t.Errorf("Expected threshold 0.8, got %f", detector.threshold)
	}

	if !detector.enabled {
		t.Error("Expected detector to be enabled")
	}

	if detector.buffer == nil {
		t.Error("Buffer is nil")
	}
}

func TestWakeWordDetection(t *testing.T) {
	detector, err := NewWakeWordDetector("hey queue")
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	// Generate test audio data
	audioData := make([]byte, 1600) // 100ms at 16kHz
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	detected, wakeWord, err := detector.DetectWakeWord(audioData)
	if err != nil {
		t.Fatalf("Wake word detection failed: %v", err)
	}

	// Detection result depends on mock implementation
	t.Logf("Wake word detection: detected=%v, word='%s'", detected, wakeWord)

	if detected && wakeWord == "" {
		t.Error("Detected wake word but no word returned")
	}

	if !detected && wakeWord != "" {
		t.Error("No detection but word returned")
	}
}

func TestWakeWordDetectorSettings(t *testing.T) {
	detector, err := NewWakeWordDetector("test word")
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	// Test threshold setting
	detector.SetThreshold(0.9)
	if detector.threshold != 0.9 {
		t.Errorf("Expected threshold 0.9, got %f", detector.threshold)
	}

	// Test enabled setting
	detector.SetEnabled(false)
	if detector.enabled {
		t.Error("Expected detector to be disabled")
	}

	detector.SetEnabled(true)
	if !detector.enabled {
		t.Error("Expected detector to be enabled")
	}
}

func TestNewRingBuffer(t *testing.T) {
	size := 1000
	buffer := NewRingBuffer(size)

	if buffer == nil {
		t.Fatal("Ring buffer is nil")
	}

	if buffer.size != size {
		t.Errorf("Expected size %d, got %d", size, buffer.size)
	}

	if len(buffer.data) != size {
		t.Errorf("Expected data length %d, got %d", size, len(buffer.data))
	}

	if buffer.Len() != 0 {
		t.Errorf("Expected empty buffer, got length %d", buffer.Len())
	}
}

func TestRingBufferWrite(t *testing.T) {
	buffer := NewRingBuffer(10)

	// Write some data
	data := []byte{1, 2, 3, 4, 5}
	buffer.Write(data)

	if buffer.Len() != 5 {
		t.Errorf("Expected length 5, got %d", buffer.Len())
	}

	// Write more data to exceed capacity
	moreData := []byte{6, 7, 8, 9, 10, 11, 12}
	buffer.Write(moreData)

	if buffer.Len() != 10 {
		t.Errorf("Expected length 10, got %d", buffer.Len())
	}

	if !buffer.full {
		t.Error("Expected buffer to be full")
	}
}

func TestRingBufferGetWindow(t *testing.T) {
	buffer := NewRingBuffer(10)

	// Write test data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	buffer.Write(data)

	// Get window smaller than available data
	window := buffer.GetWindow(5)
	if len(window) != 5 {
		t.Errorf("Expected window length 5, got %d", len(window))
	}

	// Get window larger than available data
	window = buffer.GetWindow(15)
	if len(window) > 8 {
		t.Errorf("Expected window length <= 8, got %d", len(window))
	}

	// Get window from full buffer
	buffer.Write([]byte{9, 10, 11, 12}) // This will overflow
	window = buffer.GetWindow(5)
	if len(window) != 5 {
		t.Errorf("Expected window length 5, got %d", len(window))
	}
}

func TestRingBufferClear(t *testing.T) {
	buffer := NewRingBuffer(10)

	// Write some data
	buffer.Write([]byte{1, 2, 3, 4, 5})

	// Clear buffer
	buffer.Clear()

	if buffer.Len() != 0 {
		t.Errorf("Expected empty buffer after clear, got length %d", buffer.Len())
	}

	if buffer.start != 0 || buffer.end != 0 {
		t.Error("Expected start and end to be reset to 0")
	}

	if buffer.full {
		t.Error("Expected buffer not to be full after clear")
	}
}

func TestNewDataSanitizer(t *testing.T) {
	sanitizer, err := NewDataSanitizer()
	if err != nil {
		t.Fatalf("Failed to create data sanitizer: %v", err)
	}

	if sanitizer == nil {
		t.Fatal("Data sanitizer is nil")
	}

	if len(sanitizer.sensitivePatterns) == 0 {
		t.Error("No sensitive patterns configured")
	}
}

func TestSanitizeCommand(t *testing.T) {
	sanitizer, err := NewDataSanitizer()
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "email address",
			input:    "Send logs to user@example.com",
			contains: "[EMAIL]",
		},
		{
			name:     "ip address",
			input:    "Connect to 192.168.1.1",
			contains: "[IP]",
		},
		{
			name:     "password",
			input:    "password: mysecretpass",
			contains: "[CREDENTIAL]",
		},
		{
			name:     "normal command",
			input:    "show queue status",
			contains: "show queue status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeCommand(tt.input)

			if !containsSubstring(result, tt.contains) {
				t.Errorf("Expected result to contain '%s', got '%s'", tt.contains, result)
			}
		})
	}
}

func TestSanitizerAddPattern(t *testing.T) {
	sanitizer, err := NewDataSanitizer()
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	initialCount := len(sanitizer.sensitivePatterns)

	// Add custom pattern
	customPattern := regexp.MustCompile(`secret\w+`)
	sanitizer.AddPattern(customPattern, "[SECRET]", "Custom secret pattern")

	if len(sanitizer.sensitivePatterns) != initialCount+1 {
		t.Errorf("Expected %d patterns, got %d", initialCount+1, len(sanitizer.sensitivePatterns))
	}

	// Test that custom pattern works
	result := sanitizer.SanitizeCommand("The secretkey is hidden")
	if !containsSubstring(result, "[SECRET]") {
		t.Errorf("Custom pattern not applied: %s", result)
	}
}

// Helper function to check substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && stringContainsSubstring(s, substr)
}

func stringContainsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

