package voice

import (
	"regexp"
	"time"
)

// VoiceManager coordinates all voice command functionality
type VoiceManager struct {
	recognizer    SpeechRecognizer
	processor     *CommandProcessor
	feedback      *AudioFeedback
	wakeDetector  *WakeWordDetector
	privacy       *PrivacyManager

	// State management
	listening     bool
	processing    bool
	lastCommand   *Command

	// Configuration
	config        *VoiceConfig

	// Channels for communication
	commandCh     chan *Command
	audioCh       chan []byte
	stateCh       chan VoiceState
}

// VoiceConfig holds voice command configuration
type VoiceConfig struct {
	WakeWord            string        `yaml:"wake_word" json:"wake_word"`
	RecognitionBackend  string        `yaml:"recognition_backend" json:"recognition_backend"`
	LocalOnly           bool          `yaml:"local_only" json:"local_only"`
	AudioFeedback       bool          `yaml:"audio_feedback" json:"audio_feedback"`
	Language            string        `yaml:"language" json:"language"`
	ConfidenceThreshold float64       `yaml:"confidence_threshold" json:"confidence_threshold"`
	ProcessingTimeout   time.Duration `yaml:"processing_timeout" json:"processing_timeout"`
	NoAudioRecording    bool          `yaml:"no_audio_recording" json:"no_audio_recording"`
	SanitizeLogs        bool          `yaml:"sanitize_logs" json:"sanitize_logs"`
}

// VoiceState represents the current state of voice processing
type VoiceState int

const (
	VoiceStateIdle VoiceState = iota
	VoiceStateListening
	VoiceStateProcessing
	VoiceStateError
)

// SpeechRecognizer interface for speech recognition backends
type SpeechRecognizer interface {
	StartListening() error
	StopListening() error
	ProcessAudio([]byte) (*Recognition, error)
	SetLanguage(string) error
	Close() error
}

// Recognition represents the result of speech recognition
type Recognition struct {
	Text        string    `json:"text"`
	Confidence  float64   `json:"confidence"`
	Timestamp   time.Time `json:"timestamp"`
	Entities    []Entity  `json:"entities"`
	Intent      Intent    `json:"intent"`
	ProcessTime time.Duration `json:"process_time"`
}

// CommandProcessor handles natural language command parsing
type CommandProcessor struct {
	patterns  []CommandPattern
	entities  *EntityExtractor
	context   *CommandContext
}

// CommandPattern defines a command recognition pattern
type CommandPattern struct {
	Pattern     *regexp.Regexp
	Intent      Intent
	Required    []EntityType
	Optional    []EntityType
	Handler     CommandHandler
	Description string
	Examples    []string
}

// Intent represents the user's intention
type Intent int

const (
	IntentUnknown Intent = iota
	IntentStatusQuery
	IntentWorkerControl
	IntentQueueManagement
	IntentNavigation
	IntentConfirmation
	IntentCancel
	IntentHelp
)

// String returns the string representation of an Intent
func (i Intent) String() string {
	switch i {
	case IntentStatusQuery:
		return "status_query"
	case IntentWorkerControl:
		return "worker_control"
	case IntentQueueManagement:
		return "queue_management"
	case IntentNavigation:
		return "navigation"
	case IntentConfirmation:
		return "confirmation"
	case IntentCancel:
		return "cancel"
	case IntentHelp:
		return "help"
	default:
		return "unknown"
	}
}

// Command represents a parsed voice command
type Command struct {
	Intent      Intent            `json:"intent"`
	Entities    []Entity          `json:"entities"`
	RawText     string            `json:"raw_text"`
	Confidence  float64           `json:"confidence"`
	Timestamp   time.Time         `json:"timestamp"`
	Context     map[string]string `json:"context"`
	Sanitized   string            `json:"sanitized"`
}

// GetEntity returns the first entity of the specified type
func (c *Command) GetEntity(entityType EntityType) *Entity {
	for _, entity := range c.Entities {
		if entity.Type == entityType {
			return &entity
		}
	}
	return nil
}

// GetAllEntities returns all entities of the specified type
func (c *Command) GetAllEntities(entityType EntityType) []Entity {
	var entities []Entity
	for _, entity := range c.Entities {
		if entity.Type == entityType {
			entities = append(entities, entity)
		}
	}
	return entities
}

// Entity represents an extracted entity from a command
type Entity struct {
	Type       EntityType `json:"type"`
	Value      string     `json:"value"`
	Start      int        `json:"start"`
	End        int        `json:"end"`
	Similarity float64    `json:"similarity"`
	Confidence float64    `json:"confidence"`
}

// EntityType represents the type of an entity
type EntityType int

const (
	EntityUnknown EntityType = iota
	EntityWorkerID
	EntityQueueName
	EntityTarget
	EntityDestination
	EntityNumber
	EntityTimeRange
	EntityAction
)

// String returns the string representation of an EntityType
func (e EntityType) String() string {
	switch e {
	case EntityWorkerID:
		return "worker_id"
	case EntityQueueName:
		return "queue_name"
	case EntityTarget:
		return "target"
	case EntityDestination:
		return "destination"
	case EntityNumber:
		return "number"
	case EntityTimeRange:
		return "time_range"
	case EntityAction:
		return "action"
	default:
		return "unknown"
	}
}

// CommandHandler defines the function signature for command handlers
type CommandHandler func(*Command, TUIController) error

// TUIController interface for TUI interactions
type TUIController interface {
	GetQueueStatus() *QueueStatus
	GetWorkerStatus() []WorkerStatus
	GetDLQCount() int
	NavigateToTab(string) error
	DrainWorker(string) error
	PauseWorker(string) error
	ResumeWorker(string) error
	RequeueFailedJobs() error
	ClearCompletedJobs() error
	ShowMessage(string) error
}

// QueueStatus represents queue status information
type QueueStatus struct {
	High   int `json:"high"`
	Normal int `json:"normal"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

// WorkerStatus represents worker status information
type WorkerStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Queue  string `json:"queue"`
	Jobs   int    `json:"jobs"`
}

// EntityExtractor extracts entities from recognized text
type EntityExtractor struct {
	queueNames  []string
	workerIDs   []string
	patterns    map[EntityType]*regexp.Regexp
}

// CommandContext maintains command execution context
type CommandContext struct {
	lastCommand    *Command
	currentView    string
	selectedWorker string
	selectedQueue  string
	confirmPending bool
}

// AudioFeedback handles text-to-speech and audio cues
type AudioFeedback struct {
	enabled   bool
	voice     Voice
	volume    float64
	tts       TextToSpeech
}

// Voice represents TTS voice configuration
type Voice struct {
	Name     string  `json:"name"`
	Gender   string  `json:"gender"`
	Language string  `json:"language"`
	Speed    float64 `json:"speed"`
	Pitch    float64 `json:"pitch"`
}

// TextToSpeech interface for TTS backends
type TextToSpeech interface {
	Synthesize(text string, voice Voice) ([]byte, error)
	SetVolume(float64) error
	Close() error
}

// WakeWordDetector detects wake words in audio streams
type WakeWordDetector struct {
	model      WakeWordModel
	wakeWords  []string
	threshold  float64
	buffer     *RingBuffer
	enabled    bool
}

// WakeWordModel interface for wake word detection models
type WakeWordModel interface {
	Predict([]float32) ([]float64, error)
	LoadModel(string) error
	Close() error
}

// RingBuffer implements a circular audio buffer
type RingBuffer struct {
	data    []byte
	size    int
	start   int
	end     int
	full    bool
}

// PrivacyManager handles privacy and security aspects
type PrivacyManager struct {
	localOnly     bool
	recordAudio   bool
	logCommands   bool
	sanitizer     *DataSanitizer
	cloudConsent  bool
}

// DataSanitizer removes sensitive information from commands
type DataSanitizer struct {
	sensitivePatterns []SensitivePattern
}

// SensitivePattern defines patterns for sensitive data detection
type SensitivePattern struct {
	Pattern     *regexp.Regexp
	Replacement string
	Description string
}

// VoiceMetrics tracks voice command performance and usage
type VoiceMetrics struct {
	CommandsProcessed   int64                 `json:"commands_processed"`
	RecognitionFailures int64                 `json:"recognition_failures"`
	AverageLatency      time.Duration         `json:"average_latency"`
	AverageAccuracy     float64               `json:"average_accuracy"`
	WakeWordTriggers    int64                 `json:"wake_word_triggers"`
	CommandsByIntent    map[Intent]int64      `json:"commands_by_intent"`
	ErrorsByType        map[string]int64      `json:"errors_by_type"`
	LastUpdated         time.Time             `json:"last_updated"`
}

// PerformanceMonitor tracks recognition performance
type PerformanceMonitor struct {
	recognitionTimes []time.Duration
	processingTimes  []time.Duration
	accuracyScores   []float64
	maxSamples       int
}

// VoiceError represents voice command errors
type VoiceError struct {
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Context string    `json:"context"`
}

// Error types
const (
	ErrorTypeRecognition   = "recognition"
	ErrorTypeProcessing    = "processing"
	ErrorTypeAudioDevice   = "audio_device"
	ErrorTypeNetwork       = "network"
	ErrorTypeModel         = "model"
	ErrorTypePermission    = "permission"
	ErrorTypeConfiguration = "configuration"
)

// WhisperRecognizer implements SpeechRecognizer using Whisper.cpp
type WhisperRecognizer struct {
	model       WhisperModel
	samples     []float32
	sampleRate  int
	language    string
	enabled     bool
}

// WhisperModel interface for Whisper.cpp integration
type WhisperModel interface {
	Process(samples []float32, params WhisperParams) (*WhisperResult, error)
	LoadModel(path string) error
	Close() error
}

// WhisperParams configuration for Whisper processing
type WhisperParams struct {
	Language      string
	Translate     bool
	NoContext     bool
	SingleSegment bool
	Temperature   float32
}

// WhisperResult represents Whisper recognition output
type WhisperResult struct {
	Text        string
	Probability float64
	Duration    time.Duration
}

// CloudRecognizer implements SpeechRecognizer using cloud services
type CloudRecognizer struct {
	provider    string
	apiKey      string
	endpoint    string
	language    string
	sampleRate  int
}

// CommandResponse represents the response to a voice command
type CommandResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Data       map[string]interface{} `json:"data,omitempty"`
	AudioCue   string            `json:"audio_cue,omitempty"`
	NextAction string            `json:"next_action,omitempty"`
}

// ValidationResult represents command validation results
type ValidationResult struct {
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors"`
	Warnings  []string `json:"warnings"`
	Confidence float64 `json:"confidence"`
}