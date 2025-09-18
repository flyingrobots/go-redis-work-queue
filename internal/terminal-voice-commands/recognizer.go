package voice

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// NewWhisperRecognizer creates a new Whisper-based speech recognizer
func NewWhisperRecognizer(language string) (*WhisperRecognizer, error) {
	model := &MockWhisperModel{}
	if err := model.LoadModel("models/whisper-base.bin"); err != nil {
		return nil, fmt.Errorf("failed to load Whisper model: %w", err)
	}

	return &WhisperRecognizer{
		model:      model,
		language:   language,
		sampleRate: 16000,
		enabled:    false,
	}, nil
}

// StartListening begins audio capture for recognition
func (w *WhisperRecognizer) StartListening() error {
	w.enabled = true
	log.Printf("Whisper recognizer started (language: %s)", w.language)
	return nil
}

// StopListening stops audio capture
func (w *WhisperRecognizer) StopListening() error {
	w.enabled = false
	log.Printf("Whisper recognizer stopped")
	return nil
}

// ProcessAudio processes audio data and returns recognition result
func (w *WhisperRecognizer) ProcessAudio(audio []byte) (*Recognition, error) {
	if !w.enabled {
		return nil, fmt.Errorf("recognizer not enabled")
	}

	start := time.Now()

	// Convert audio bytes to float32 samples
	samples := w.convertToFloat32(audio)

	// Process with Whisper model
	params := WhisperParams{
		Language:      w.language,
		Translate:     false,
		NoContext:     true,
		SingleSegment: true,
		Temperature:   0.0,
	}

	result, err := w.model.Process(samples, params)
	if err != nil {
		return nil, fmt.Errorf("Whisper processing failed: %w", err)
	}

	processTime := time.Since(start)

	return &Recognition{
		Text:        result.Text,
		Confidence:  result.Probability,
		Timestamp:   time.Now(),
		ProcessTime: processTime,
	}, nil
}

// SetLanguage sets the recognition language
func (w *WhisperRecognizer) SetLanguage(language string) error {
	w.language = language
	log.Printf("Whisper language set to: %s", language)
	return nil
}

// Close releases resources
func (w *WhisperRecognizer) Close() error {
	w.enabled = false
	return w.model.Close()
}

// convertToFloat32 converts byte audio data to float32 samples
func (w *WhisperRecognizer) convertToFloat32(audio []byte) []float32 {
	// Convert 16-bit PCM to float32
	samples := make([]float32, len(audio)/2)
	for i := 0; i < len(audio); i += 2 {
		// Convert little-endian 16-bit to float32 [-1.0, 1.0]
		sample := int16(audio[i]) | int16(audio[i+1])<<8
		samples[i/2] = float32(sample) / 32768.0
	}
	return samples
}

// MockWhisperModel implements WhisperModel for testing/development
type MockWhisperModel struct {
	modelPath string
	mutex     sync.RWMutex
	loaded    bool
}

// LoadModel loads the Whisper model
func (m *MockWhisperModel) LoadModel(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.modelPath = path
	m.loaded = true

	log.Printf("Mock Whisper model loaded from: %s", path)
	return nil
}

// Process processes audio samples and returns recognition result
func (m *MockWhisperModel) Process(samples []float32, params WhisperParams) (*WhisperResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.loaded {
		return nil, fmt.Errorf("model not loaded")
	}

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Mock recognition results based on audio characteristics
	text := m.mockRecognition(samples)
	confidence := m.calculateConfidence(samples, text)

	return &WhisperResult{
		Text:        text,
		Probability: confidence,
		Duration:    time.Duration(len(samples)) * time.Second / 16000,
	}, nil
}

// Close releases model resources
func (m *MockWhisperModel) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.loaded = false
	log.Printf("Mock Whisper model closed")
	return nil
}

// mockRecognition generates mock recognition text based on audio patterns
func (m *MockWhisperModel) mockRecognition(samples []float32) string {
	// Calculate basic audio characteristics
	energy := m.calculateEnergy(samples)
	duration := float64(len(samples)) / 16000.0

	// Generate mock commands based on audio characteristics
	if energy < 0.1 {
		return "" // Too quiet
	}

	if duration < 0.5 {
		return "hey queue" // Short utterance = wake word
	}

	if duration < 2.0 {
		// Short commands
		commands := []string{
			"show queue status",
			"show workers",
			"drain worker 1",
			"pause worker 2",
			"resume worker 3",
			"go to dlq",
			"help",
			"cancel",
			"yes",
			"confirm",
		}
		return commands[int(energy*100)%len(commands)]
	}

	// Longer commands
	commands := []string{
		"show me the high priority queue status",
		"drain worker 3 and pause worker 1",
		"requeue all failed jobs from the last hour",
		"show me the dead letter queue",
		"navigate to the workers tab",
		"clear all completed jobs older than one day",
		"how many jobs are in the processing queue",
		"what is the status of worker 2",
	}
	return commands[int(energy*1000)%len(commands)]
}

// calculateEnergy calculates RMS energy of audio samples
func (m *MockWhisperModel) calculateEnergy(samples []float32) float64 {
	var sum float64
	for _, sample := range samples {
		sum += float64(sample * sample)
	}
	return sum / float64(len(samples))
}

// calculateConfidence calculates recognition confidence based on audio quality
func (m *MockWhisperModel) calculateConfidence(samples []float32, text string) float64 {
	energy := m.calculateEnergy(samples)

	// Base confidence on audio energy and text length
	baseConfidence := 0.6
	energyBonus := energy * 0.3
	lengthPenalty := 0.0

	if len(text) == 0 {
		return 0.0
	}

	if len(text) > 50 {
		lengthPenalty = 0.1
	}

	confidence := baseConfidence + energyBonus - lengthPenalty

	// Clamp to [0.0, 1.0]
	if confidence < 0.0 {
		confidence = 0.0
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// NewCloudRecognizer creates a new cloud-based speech recognizer
func NewCloudRecognizer(provider, language string) (*CloudRecognizer, error) {
	switch provider {
	case "google":
		return &CloudRecognizer{
			provider:   "google",
			endpoint:   "https://speech.googleapis.com/v1/speech:recognize",
			language:   language,
			sampleRate: 16000,
		}, nil
	case "azure":
		return &CloudRecognizer{
			provider:   "azure",
			endpoint:   "https://eastus.stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1",
			language:   language,
			sampleRate: 16000,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// StartListening begins audio capture for cloud recognition
func (c *CloudRecognizer) StartListening() error {
	log.Printf("Cloud recognizer started (provider: %s, language: %s)", c.provider, c.language)
	return nil
}

// StopListening stops audio capture
func (c *CloudRecognizer) StopListening() error {
	log.Printf("Cloud recognizer stopped")
	return nil
}

// ProcessAudio processes audio data using cloud recognition service
func (c *CloudRecognizer) ProcessAudio(audio []byte) (*Recognition, error) {
	start := time.Now()

	// Simulate cloud API call
	time.Sleep(200 * time.Millisecond)

	// Mock cloud recognition
	text := c.mockCloudRecognition(audio)
	confidence := c.calculateCloudConfidence(audio, text)

	processTime := time.Since(start)

	return &Recognition{
		Text:        text,
		Confidence:  confidence,
		Timestamp:   time.Now(),
		ProcessTime: processTime,
	}, nil
}

// SetLanguage sets the recognition language
func (c *CloudRecognizer) SetLanguage(language string) error {
	c.language = language
	log.Printf("Cloud recognizer language set to: %s", language)
	return nil
}

// Close releases resources
func (c *CloudRecognizer) Close() error {
	log.Printf("Cloud recognizer closed")
	return nil
}

// mockCloudRecognition simulates cloud speech recognition
func (c *CloudRecognizer) mockCloudRecognition(audio []byte) string {
	// Simulate higher accuracy cloud recognition
	if len(audio) < 1000 {
		return ""
	}

	commands := []string{
		"hey queue show me the queue status",
		"drain worker number three",
		"what is the current worker status",
		"navigate to the dead letter queue",
		"requeue all failed jobs",
		"pause worker one and two",
		"show me statistics",
		"clear completed jobs",
		"help with voice commands",
		"cancel current operation",
	}

	// Use audio length as seed for consistent results
	index := len(audio) % len(commands)
	return commands[index]
}

// calculateCloudConfidence calculates confidence for cloud recognition
func (c *CloudRecognizer) calculateCloudConfidence(audio []byte, text string) float64 {
	// Cloud services typically have higher confidence
	baseConfidence := 0.85

	if len(text) == 0 {
		return 0.0
	}

	// Adjust based on audio quality indicators
	qualityScore := float64(len(audio)) / 10000.0
	if qualityScore > 1.0 {
		qualityScore = 1.0
	}

	confidence := baseConfidence + (qualityScore * 0.15)

	// Clamp to [0.0, 1.0]
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
