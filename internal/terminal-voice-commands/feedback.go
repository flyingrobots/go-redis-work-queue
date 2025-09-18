package voice

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

// NewAudioFeedback creates a new audio feedback system
func NewAudioFeedback(enabled bool) (*AudioFeedback, error) {
	tts, err := NewTextToSpeech()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TTS: %w", err)
	}

	return &AudioFeedback{
		enabled: enabled,
		voice: Voice{
			Name:     "default",
			Gender:   "neutral",
			Language: "en-US",
			Speed:    1.0,
			Pitch:    1.0,
		},
		volume: 0.7,
		tts:    tts,
	}, nil
}

// SpeakResponse converts text to speech and plays it
func (a *AudioFeedback) SpeakResponse(response string) error {
	if !a.enabled {
		return nil
	}

	// Sanitize response text for TTS
	cleanText := a.sanitizeForTTS(response)

	// Generate speech audio
	audio, err := a.tts.Synthesize(cleanText, a.voice)
	if err != nil {
		return fmt.Errorf("TTS synthesis failed: %w", err)
	}

	// Play audio with volume adjustment
	return a.playAudio(audio, a.volume)
}

// PlayConfirmationSound plays a confirmation beep
func (a *AudioFeedback) PlayConfirmationSound() error {
	if !a.enabled {
		return nil
	}

	log.Printf("Playing confirmation sound")
	// In a real implementation, this would play an actual audio file
	return nil
}

// PlayErrorSound plays an error sound
func (a *AudioFeedback) PlayErrorSound() error {
	if !a.enabled {
		return nil
	}

	log.Printf("Playing error sound")
	// In a real implementation, this would play an actual audio file
	return nil
}

// PlayNotificationSound plays a notification sound
func (a *AudioFeedback) PlayNotificationSound() error {
	if !a.enabled {
		return nil
	}

	log.Printf("Playing notification sound")
	return nil
}

// SetEnabled enables or disables audio feedback
func (a *AudioFeedback) SetEnabled(enabled bool) {
	a.enabled = enabled
	log.Printf("Audio feedback %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// SetVolume sets the audio volume (0.0 to 1.0)
func (a *AudioFeedback) SetVolume(volume float64) error {
	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	a.volume = volume
	return a.tts.SetVolume(volume)
}

// SetVoice sets the TTS voice configuration
func (a *AudioFeedback) SetVoice(voice Voice) error {
	a.voice = voice
	log.Printf("Voice set to: %s (%s, %s)", voice.Name, voice.Gender, voice.Language)
	return nil
}

// Close closes the audio feedback system
func (a *AudioFeedback) Close() error {
	return a.tts.Close()
}

// sanitizeForTTS cleans text for better TTS output
func (a *AudioFeedback) sanitizeForTTS(text string) string {
	// Remove or replace characters that cause TTS issues
	clean := strings.ReplaceAll(text, "\n", " ")
	clean = strings.ReplaceAll(clean, "\t", " ")

	// Replace numbers with words for better pronunciation
	replacements := map[string]string{
		"1":  "one",
		"2":  "two",
		"3":  "three",
		"4":  "four",
		"5":  "five",
		"6":  "six",
		"7":  "seven",
		"8":  "eight",
		"9":  "nine",
		"10": "ten",
	}

	for number, word := range replacements {
		clean = strings.ReplaceAll(clean, number, word)
	}

	// Replace technical terms with more pronounceable versions
	techReplacements := map[string]string{
		"DLQ":   "dead letter queue",
		"API":   "A P I",
		"HTTP":  "H T T P",
		"JSON":  "JSON",
		"Redis": "Redis",
		"TUI":   "text user interface",
	}

	for tech, replacement := range techReplacements {
		clean = strings.ReplaceAll(clean, tech, replacement)
	}

	return clean
}

// playAudio plays audio data with specified volume
func (a *AudioFeedback) playAudio(audio []byte, volume float64) error {
	// In a real implementation, this would use an audio library
	// to play the synthesized audio data
	log.Printf("Playing audio: %d bytes at volume %.2f", len(audio), volume)
	return nil
}

// MockTextToSpeech implements TextToSpeech for development/testing
type MockTextToSpeech struct {
	enabled bool
	volume  float64
	mutex   sync.RWMutex
}

// NewTextToSpeech creates a new text-to-speech engine
func NewTextToSpeech() (TextToSpeech, error) {
	return &MockTextToSpeech{
		enabled: true,
		volume:  0.7,
	}, nil
}

// Synthesize converts text to speech audio
func (m *MockTextToSpeech) Synthesize(text string, voice Voice) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.enabled {
		return nil, fmt.Errorf("TTS not enabled")
	}

	// Simulate TTS processing time
	time.Sleep(time.Duration(len(text)) * 10 * time.Millisecond)

	// Generate mock audio data
	audioData := make([]byte, len(text)*100) // Mock audio data
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	log.Printf("TTS synthesized: '%s' -> %d bytes audio", text, len(audioData))
	return audioData, nil
}

// SetVolume sets the TTS volume
func (m *MockTextToSpeech) SetVolume(volume float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	m.volume = volume
	log.Printf("TTS volume set to: %.2f", volume)
	return nil
}

// Close closes the TTS engine
func (m *MockTextToSpeech) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.enabled = false
	log.Printf("TTS engine closed")
	return nil
}

// NewWakeWordDetector creates a new wake word detector
func NewWakeWordDetector(wakeWord string) (*WakeWordDetector, error) {
	wakeWords := strings.Split(strings.ToLower(wakeWord), " ")

	model := &MockWakeWordModel{}
	if err := model.LoadModel("models/wake-word.tflite"); err != nil {
		return nil, fmt.Errorf("failed to load wake word model: %w", err)
	}

	buffer := NewRingBuffer(16000 * 2) // 2 seconds at 16kHz

	return &WakeWordDetector{
		model:     model,
		wakeWords: wakeWords,
		threshold: 0.8,
		buffer:    buffer,
		enabled:   true,
	}, nil
}

// DetectWakeWord analyzes audio for wake word patterns
func (w *WakeWordDetector) DetectWakeWord(audio []byte) (bool, string, error) {
	if !w.enabled {
		return false, "", nil
	}

	// Add audio to circular buffer
	w.buffer.Write(audio)

	// Extract features for wake word detection
	window := w.buffer.GetWindow(1600) // 100ms window
	features := w.extractMFCC(window)

	// Run wake word model
	prediction, err := w.model.Predict(features)
	if err != nil {
		return false, "", fmt.Errorf("wake word prediction failed: %w", err)
	}

	// Check if any wake word exceeded threshold
	for i, score := range prediction {
		if score > w.threshold && i < len(w.wakeWords) {
			wakeWord := strings.Join(w.wakeWords, " ")
			log.Printf("Wake word detected: '%s' (confidence: %.2f)", wakeWord, score)
			return true, wakeWord, nil
		}
	}

	return false, "", nil
}

// SetThreshold sets the wake word detection threshold
func (w *WakeWordDetector) SetThreshold(threshold float64) {
	w.threshold = threshold
	log.Printf("Wake word threshold set to: %.2f", threshold)
}

// SetEnabled enables or disables wake word detection
func (w *WakeWordDetector) SetEnabled(enabled bool) {
	w.enabled = enabled
	log.Printf("Wake word detection %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// extractMFCC extracts MFCC features from audio data
func (w *WakeWordDetector) extractMFCC(audio []byte) []float32 {
	// Convert bytes to float32 samples
	samples := make([]float32, len(audio)/2)
	for i := 0; i < len(audio); i += 2 {
		sample := int16(audio[i]) | int16(audio[i+1])<<8
		samples[i/2] = float32(sample) / 32768.0
	}

	// Mock MFCC extraction (real implementation would use DSP library)
	features := make([]float32, 13) // 13 MFCC coefficients
	for i := range features {
		features[i] = 0.0
		for j := 0; j < len(samples); j += len(samples) / 13 {
			if j+i < len(samples) {
				features[i] += samples[j+i]
			}
		}
		features[i] /= float32(len(samples) / 13)
	}

	return features
}

// MockWakeWordModel implements WakeWordModel for testing
type MockWakeWordModel struct {
	modelPath string
	loaded    bool
	mutex     sync.RWMutex
}

// LoadModel loads the wake word detection model
func (m *MockWakeWordModel) LoadModel(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.modelPath = path
	m.loaded = true

	log.Printf("Mock wake word model loaded from: %s", path)
	return nil
}

// Predict predicts wake word presence from MFCC features
func (m *MockWakeWordModel) Predict(features []float32) ([]float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.loaded {
		return nil, fmt.Errorf("model not loaded")
	}

	// Calculate simple energy-based prediction
	energy := 0.0
	for _, feature := range features {
		energy += float64(feature * feature)
	}
	energy = energy / float64(len(features))

	// Mock prediction scores for different wake words
	predictions := []float64{
		energy * 1.2, // "hey queue"
		energy * 0.8, // other phrases
		energy * 0.5, // background noise
	}

	// Clamp predictions to [0.0, 1.0]
	for i := range predictions {
		if predictions[i] > 1.0 {
			predictions[i] = 1.0
		}
		if predictions[i] < 0.0 {
			predictions[i] = 0.0
		}
	}

	return predictions, nil
}

// Close closes the wake word model
func (m *MockWakeWordModel) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.loaded = false
	log.Printf("Mock wake word model closed")
	return nil
}

// NewRingBuffer creates a new ring buffer with specified size
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]byte, size),
		size: size,
	}
}

// Write adds data to the ring buffer
func (r *RingBuffer) Write(data []byte) {
	for _, b := range data {
		r.data[r.end] = b
		r.end = (r.end + 1) % r.size

		if r.full {
			r.start = (r.start + 1) % r.size
		}

		if r.end == r.start {
			r.full = true
		}
	}
}

// GetWindow returns a window of data from the buffer
func (r *RingBuffer) GetWindow(size int) []byte {
	if size > r.size {
		size = r.size
	}

	window := make([]byte, size)

	if r.full {
		// Buffer is full, read from current position backwards
		pos := r.end
		for i := size - 1; i >= 0; i-- {
			pos = (pos - 1 + r.size) % r.size
			window[i] = r.data[pos]
		}
	} else {
		// Buffer not full, read available data
		available := r.end - r.start
		if available < 0 {
			available += r.size
		}
		if size > available {
			size = available
		}

		for i := 0; i < size; i++ {
			pos := (r.end - size + i + r.size) % r.size
			window[i] = r.data[pos]
		}
	}

	return window[:size]
}

// Len returns the current amount of data in the buffer
func (r *RingBuffer) Len() int {
	if r.full {
		return r.size
	}

	if r.end >= r.start {
		return r.end - r.start
	}

	return r.size - r.start + r.end
}

// Clear empties the ring buffer
func (r *RingBuffer) Clear() {
	r.start = 0
	r.end = 0
	r.full = false
}

// NewDataSanitizer creates a new data sanitizer for privacy protection
func NewDataSanitizer() (*DataSanitizer, error) {
	patterns := []SensitivePattern{
		{
			Pattern:     regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			Replacement: "[EMAIL]",
			Description: "Email addresses",
		},
		{
			Pattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			Replacement: "[SSN]",
			Description: "Social Security Numbers",
		},
		{
			Pattern:     regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
			Replacement: "[CARD]",
			Description: "Credit card numbers",
		},
		{
			Pattern:     regexp.MustCompile(`\b(?:password|pass|pwd|token|key|secret)\s*[:=]\s*\S+`),
			Replacement: "[CREDENTIAL]",
			Description: "Passwords and tokens",
		},
		{
			Pattern:     regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			Replacement: "[IP]",
			Description: "IP addresses",
		},
	}

	return &DataSanitizer{
		sensitivePatterns: patterns,
	}, nil
}

// SanitizeCommand removes sensitive information from command text
func (d *DataSanitizer) SanitizeCommand(text string) string {
	sanitized := text

	for _, pattern := range d.sensitivePatterns {
		sanitized = pattern.Pattern.ReplaceAllString(sanitized, pattern.Replacement)
	}

	return sanitized
}

// AddPattern adds a new sensitive data pattern
func (d *DataSanitizer) AddPattern(pattern *regexp.Regexp, replacement, description string) {
	d.sensitivePatterns = append(d.sensitivePatterns, SensitivePattern{
		Pattern:     pattern,
		Replacement: replacement,
		Description: description,
	})
}

// GetPatterns returns all configured sensitive patterns
func (d *DataSanitizer) GetPatterns() []SensitivePattern {
	return d.sensitivePatterns
}
