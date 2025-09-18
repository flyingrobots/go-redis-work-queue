package voice

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWhisperRecognizer(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create Whisper recognizer: %v", err)
	}

	if recognizer == nil {
		t.Fatal("Recognizer is nil")
	}

	if recognizer.language != "en" {
		t.Errorf("Expected language 'en', got '%s'", recognizer.language)
	}

	if recognizer.sampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", recognizer.sampleRate)
	}

	if recognizer.enabled {
		t.Error("Expected recognizer to be disabled until StartListening is called")
	}
}

func TestWhisperRecognizerStartStop(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	// Test start
	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	if !recognizer.enabled {
		t.Error("Expected recognizer to be enabled after start")
	}

	// Test stop
	if err := recognizer.StopListening(); err != nil {
		t.Fatalf("Failed to stop listening: %v", err)
	}

	if recognizer.enabled {
		t.Error("Expected recognizer to be disabled after stop")
	}
}

func TestWhisperProcessAudio(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start recognizer: %v", err)
	}
	defer recognizer.Close()

	// Generate test audio data (16-bit PCM, 1 second at 16kHz)
	audioData := make([]byte, 16000*2)
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	recognition, err := recognizer.ProcessAudio(audioData)
	if err != nil {
		t.Fatalf("Failed to process audio: %v", err)
	}

	if recognition == nil {
		t.Fatal("Recognition result is nil")
	}

	if recognition.Text == "" {
		t.Error("Recognition text is empty")
	}

	if recognition.Confidence < 0.0 || recognition.Confidence > 1.0 {
		t.Errorf("Invalid confidence value: %f", recognition.Confidence)
	}

	if recognition.ProcessTime <= 0 {
		t.Error("Process time should be positive")
	}

	if recognition.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestWhisperSetLanguage(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	languages := []string{"en", "es", "fr", "de", "ja"}

	for _, lang := range languages {
		if err := recognizer.SetLanguage(lang); err != nil {
			t.Errorf("Failed to set language '%s': %v", lang, err)
		}

		if recognizer.language != lang {
			t.Errorf("Expected language '%s', got '%s'", lang, recognizer.language)
		}
	}
}

func TestWhisperConvertToFloat32(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	// Test audio data (16-bit PCM samples)
	audioData := []byte{
		0x00, 0x00, // 0
		0x00, 0x80, // -32768 (min)
		0xFF, 0x7F, // 32767 (max)
		0x00, 0x40, // 16384 (half max)
	}

	samples := recognizer.convertToFloat32(audioData)

	expectedSamples := []float32{0.0, -1.0, 0.999969482, 0.5}
	tolerance := float32(0.001)

	if len(samples) != len(expectedSamples) {
		t.Fatalf("Expected %d samples, got %d", len(expectedSamples), len(samples))
	}

	for i, expected := range expectedSamples {
		if abs32(samples[i]-expected) > tolerance {
			t.Errorf("Sample %d: expected %f, got %f", i, expected, samples[i])
		}
	}
}

func TestMockWhisperModel(t *testing.T) {
	model := &MockWhisperModel{}

	// Test loading
	if err := model.LoadModel("test-model.bin"); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	if !model.loaded {
		t.Error("Model should be loaded")
	}

	// Test processing
	samples := make([]float32, 1000)
	for i := range samples {
		samples[i] = float32(i) / 1000.0
	}

	params := WhisperParams{
		Language:      "en",
		Translate:     false,
		NoContext:     true,
		SingleSegment: true,
	}

	result, err := model.Process(samples, params)
	if err != nil {
		t.Fatalf("Failed to process audio: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Text == "" {
		t.Error("Result text is empty")
	}

	if result.Probability < 0.0 || result.Probability > 1.0 {
		t.Errorf("Invalid probability: %f", result.Probability)
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	// Test closing
	if err := model.Close(); err != nil {
		t.Fatalf("Failed to close model: %v", err)
	}

	if model.loaded {
		t.Error("Model should not be loaded after close")
	}
}

func TestNewCloudRecognizer(t *testing.T) {
	providers := []string{"google", "azure"}

	for _, provider := range providers {
		recognizer, err := NewCloudRecognizer(provider, "en")
		if err != nil {
			t.Fatalf("Failed to create %s recognizer: %v", provider, err)
		}

		if recognizer.provider != provider {
			t.Errorf("Expected provider '%s', got '%s'", provider, recognizer.provider)
		}

		if recognizer.language != "en" {
			t.Errorf("Expected language 'en', got '%s'", recognizer.language)
		}

		if recognizer.sampleRate != 16000 {
			t.Errorf("Expected sample rate 16000, got %d", recognizer.sampleRate)
		}
	}

	// Test unsupported provider
	_, err := NewCloudRecognizer("unsupported", "en")
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}
}

func TestCloudRecognizerStartStop(t *testing.T) {
	recognizer, err := NewCloudRecognizer("google", "en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	// Test start
	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Test stop
	if err := recognizer.StopListening(); err != nil {
		t.Fatalf("Failed to stop listening: %v", err)
	}

	// Test close
	if err := recognizer.Close(); err != nil {
		t.Fatalf("Failed to close recognizer: %v", err)
	}
}

func TestCloudProcessAudio(t *testing.T) {
	recognizer, err := NewCloudRecognizer("google", "en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start recognizer: %v", err)
	}
	defer recognizer.Close()

	// Generate test audio data
	audioData := make([]byte, 8000) // 0.5 seconds at 16kHz
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	start := time.Now()
	recognition, err := recognizer.ProcessAudio(audioData)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to process audio: %v", err)
	}

	if recognition == nil {
		t.Fatal("Recognition result is nil")
	}

	if recognition.Text == "" {
		t.Error("Recognition text is empty")
	}

	if recognition.Confidence < 0.0 || recognition.Confidence > 1.0 {
		t.Errorf("Invalid confidence value: %f", recognition.Confidence)
	}

	// Cloud recognition should take some time (mocked delay)
	if duration < 100*time.Millisecond {
		t.Error("Cloud recognition should take at least 100ms")
	}
}

func TestCloudSetLanguage(t *testing.T) {
	recognizer, err := NewCloudRecognizer("azure", "en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	languages := []string{"en", "es", "fr", "de", "ja"}

	for _, lang := range languages {
		if err := recognizer.SetLanguage(lang); err != nil {
			t.Errorf("Failed to set language '%s': %v", lang, err)
		}

		if recognizer.language != lang {
			t.Errorf("Expected language '%s', got '%s'", lang, recognizer.language)
		}
	}
}

func TestMockCloudRecognition(t *testing.T) {
	recognizer, err := NewCloudRecognizer("google", "en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	// Test with different audio lengths
	testCases := []struct {
		name       string
		audioSize  int
		expectText bool
	}{
		{"empty audio", 0, false},
		{"very short audio", 500, false},
		{"normal audio", 5000, true},
		{"long audio", 20000, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			audioData := make([]byte, tc.audioSize)
			for i := range audioData {
				audioData[i] = byte(i % 256)
			}

			text := recognizer.mockCloudRecognition(audioData)

			if tc.expectText && text == "" {
				t.Error("Expected non-empty text")
			}

			if !tc.expectText && text != "" {
				t.Error("Expected empty text")
			}
		})
	}
}

func TestCloudConfidenceCalculation(t *testing.T) {
	recognizer, err := NewCloudRecognizer("google", "en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	testCases := []struct {
		name      string
		audioSize int
		text      string
		minConf   float64
		maxConf   float64
	}{
		{"empty text", 1000, "", 0.0, 0.0},
		{"short audio", 1000, "hello", 0.85, 0.95},
		{"long audio", 50000, "hello world", 0.95, 1.0},
		{"very long audio", 100000, "test command", 1.0, 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			audioData := make([]byte, tc.audioSize)

			confidence := recognizer.calculateCloudConfidence(audioData, tc.text)

			if confidence < tc.minConf || confidence > tc.maxConf {
				t.Errorf("Expected confidence between %f and %f, got %f",
					tc.minConf, tc.maxConf, confidence)
			}
		})
	}
}

// Helper function for float32 absolute difference
func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func TestRecognizerPerformance(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start recognizer: %v", err)
	}
	defer recognizer.Close()

	// Test performance with different audio lengths
	testSizes := []int{8000, 16000, 32000, 48000} // 0.5s, 1s, 2s, 3s

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("audio_%dms", size/32), func(t *testing.T) {
			audioData := make([]byte, size)
			for i := range audioData {
				audioData[i] = byte(i % 256)
			}

			start := time.Now()
			recognition, err := recognizer.ProcessAudio(audioData)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Failed to process audio: %v", err)
			}

			// Check that processing time is reasonable (< 1 second for mock)
			if duration > time.Second {
				t.Errorf("Processing took too long: %v", duration)
			}

			if recognition.ProcessTime <= 0 {
				t.Error("Process time should be recorded")
			}

			t.Logf("Audio size: %d bytes, Processing time: %v, Recognition: '%s'",
				size, duration, recognition.Text)
		})
	}
}

func TestRecognizerErrorHandling(t *testing.T) {
	recognizer, err := NewWhisperRecognizer("en")
	if err != nil {
		t.Fatalf("Failed to create recognizer: %v", err)
	}

	// Test processing without starting
	audioData := make([]byte, 1000)
	_, err = recognizer.ProcessAudio(audioData)
	if err == nil {
		t.Error("Expected error when processing audio without starting")
	}

	// Test with started recognizer
	if err := recognizer.StartListening(); err != nil {
		t.Fatalf("Failed to start recognizer: %v", err)
	}

	// Test with empty audio
	_, err = recognizer.ProcessAudio([]byte{})
	if err != nil {
		t.Logf("Processing empty audio returned error (expected): %v", err)
	}

	// Test with very small audio
	smallAudio := []byte{0x00, 0x01}
	_, err = recognizer.ProcessAudio(smallAudio)
	if err != nil {
		t.Logf("Processing very small audio returned error: %v", err)
	}

	recognizer.Close()
}
