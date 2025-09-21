# Terminal Voice Commands API Reference

> [!WARNING]
> **Archived Feature** — The `internal/terminal-voice-commands` package was removed from the repository on 2025-09-20 during repository cleanup. This document is retained for historical reference only.

## Overview

The Terminal Voice Commands feature provides hands-free queue management through natural language voice commands. This API documentation covers the Go package interfaces, configuration options, and integration patterns for implementing voice control in the terminal user interface.

## Core Architecture

### VoiceManager

The main entry point for voice command functionality.

```go
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

    // Communication channels
    commandCh     chan *Command
    audioCh       chan []byte
    stateCh       chan VoiceState
}
```

#### Constructor

```go
func NewVoiceManager(ctx context.Context, config *VoiceConfig) (*VoiceManager, error)
```

Creates a new voice manager with the specified configuration. Initializes all components including speech recognition, command processing, audio feedback, and privacy management.

**Parameters:**
- `ctx`: Context for lifecycle management
- `config`: Voice configuration settings (nil uses defaults)

**Returns:**
- `*VoiceManager`: Configured voice manager instance
- `error`: Any initialization errors

**Example:**
```go
config := DefaultVoiceConfig()
config.LocalOnly = true
config.AudioFeedback = false

vm, err := NewVoiceManager(context.Background(), config)
if err != nil {
    log.Fatalf("Failed to create voice manager: %v", err)
}
```

#### Methods

##### Start
```go
func (v *VoiceManager) Start() error
```

Starts voice command processing, including audio capture and command handling goroutines.

##### Stop
```go
func (v *VoiceManager) Stop() error
```

Stops voice command processing and releases all resources.

##### ProcessCommand
```go
func (v *VoiceManager) ProcessCommand(command string, tui TUIController) (*CommandResponse, error)
```

Processes a text command (for testing or direct input).

**Parameters:**
- `command`: Raw command text
- `tui`: TUI controller for executing commands

**Returns:**
- `*CommandResponse`: Execution result with success status and response data
- `error`: Processing error

##### ToggleListening
```go
func (v *VoiceManager) ToggleListening() error
```

Toggles voice listening mode on/off.

##### GetState
```go
func (v *VoiceManager) GetState() VoiceState
```

Returns current voice manager state (idle, listening, processing, error).

##### GetMetrics
```go
func (v *VoiceManager) GetMetrics() *VoiceMetrics
```

Returns performance and usage metrics.

### Configuration

#### VoiceConfig
```go
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
```

#### Default Configuration
```go
func DefaultVoiceConfig() *VoiceConfig
```

Returns default voice configuration:
- Wake word: "hey queue"
- Recognition backend: "whisper" (local)
- Local only: true
- Audio feedback: true
- Language: "en"
- Confidence threshold: 0.7
- Processing timeout: 5 seconds
- No audio recording: true
- Sanitize logs: true

#### Configuration Management
```go
type ConfigManager struct {
    configPath string
    config     *VoiceConfig
}

func NewConfigManager(configPath string) (*ConfigManager, error)
```

Manages voice configuration persistence and validation.

**Key Methods:**
- `Load() error`: Load configuration from file
- `Save() error`: Save configuration to file
- `ApplyPreset(name string) error`: Apply predefined configuration preset
- `SetWakeWord(word string) error`: Update wake word
- `SetConfidenceThreshold(threshold float64) error`: Update confidence threshold

### Speech Recognition

#### SpeechRecognizer Interface
```go
type SpeechRecognizer interface {
    StartListening() error
    StopListening() error
    ProcessAudio([]byte) (*Recognition, error)
    SetLanguage(string) error
    Close() error
}
```

#### Whisper Recognition (Local)
```go
func NewWhisperRecognizer(language string) (*WhisperRecognizer, error)
```

Creates a local Whisper-based speech recognizer for privacy-conscious environments.

#### Cloud Recognition
```go
func NewCloudRecognizer(provider, language string) (*CloudRecognizer, error)
```

Creates a cloud-based speech recognizer. Supported providers:
- "google": Google Speech-to-Text
- "azure": Azure Speech Services

#### Recognition Result
```go
type Recognition struct {
    Text        string    `json:"text"`
    Confidence  float64   `json:"confidence"`
    Timestamp   time.Time `json:"timestamp"`
    Entities    []Entity  `json:"entities"`
    Intent      Intent    `json:"intent"`
    ProcessTime time.Duration `json:"process_time"`
}
```

### Command Processing

#### CommandProcessor
```go
type CommandProcessor struct {
    patterns  []CommandPattern
    entities  *EntityExtractor
    context   *CommandContext
}

func NewCommandProcessor() (*CommandProcessor, error)
```

Handles natural language command parsing and intent recognition.

#### Intent Types
```go
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
```

#### Command Structure
```go
type Command struct {
    Intent      Intent            `json:"intent"`
    Entities    []Entity          `json:"entities"`
    RawText     string            `json:"raw_text"`
    Confidence  float64           `json:"confidence"`
    Timestamp   time.Time         `json:"timestamp"`
    Context     map[string]string `json:"context"`
    Sanitized   string            `json:"sanitized"`
}
```

#### Entity Types
```go
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
```

### TUI Integration

#### TUIController Interface
```go
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
```

Implement this interface in your TUI to enable voice command integration.

#### Status Structures
```go
type QueueStatus struct {
    High   int `json:"high"`
    Normal int `json:"normal"`
    Low    int `json:"low"`
    Total  int `json:"total"`
}

type WorkerStatus struct {
    ID     string `json:"id"`
    Status string `json:"status"`
    Queue  string `json:"queue"`
    Jobs   int    `json:"jobs"`
}
```

### Audio Feedback

#### AudioFeedback
```go
type AudioFeedback struct {
    enabled   bool
    voice     Voice
    volume    float64
    tts       TextToSpeech
}

func NewAudioFeedback(enabled bool) (*AudioFeedback, error)
```

Provides text-to-speech responses and audio cues.

#### Methods
- `SpeakResponse(text string) error`: Convert text to speech
- `PlayConfirmationSound() error`: Play confirmation beep
- `PlayErrorSound() error`: Play error sound
- `SetVolume(volume float64) error`: Set audio volume (0.0-1.0)
- `SetVoice(voice Voice) error`: Configure TTS voice

#### Voice Configuration
```go
type Voice struct {
    Name     string  `json:"name"`
    Gender   string  `json:"gender"`
    Language string  `json:"language"`
    Speed    float64 `json:"speed"`
    Pitch    float64 `json:"pitch"`
}
```

### Wake Word Detection

#### WakeWordDetector
```go
type WakeWordDetector struct {
    model      WakeWordModel
    wakeWords  []string
    threshold  float64
    buffer     *RingBuffer
    enabled    bool
}

func NewWakeWordDetector(wakeWord string) (*WakeWordDetector, error)
```

Detects configurable wake words in audio streams.

#### Methods
- `DetectWakeWord(audio []byte) (bool, string, error)`: Analyze audio for wake words
- `SetThreshold(threshold float64)`: Set detection sensitivity
- `SetEnabled(enabled bool)`: Enable/disable detection

### Privacy and Security

#### PrivacyManager
```go
type PrivacyManager struct {
    localOnly     bool
    recordAudio   bool
    logCommands   bool
    sanitizer     *DataSanitizer
    cloudConsent  bool
}
```

Manages privacy settings and data protection.

#### DataSanitizer
```go
type DataSanitizer struct {
    sensitivePatterns []SensitivePattern
}

func NewDataSanitizer() (*DataSanitizer, error)
```

Removes sensitive information from voice commands before logging.

**Built-in Patterns:**
- Email addresses → [EMAIL]
- IP addresses → [IP]
- Credit card numbers → [CARD]
- Passwords/tokens → [CREDENTIAL]
- Social Security Numbers → [SSN]

## Command Grammar

### Supported Commands

#### Status Queries
- "show queue status"
- "what are the workers doing"
- "how many jobs in high priority"
- "show me the DLQ"

#### Worker Control
- "drain worker 3"
- "pause worker 1"
- "resume all workers"
- "stop the third worker"

#### Queue Management
- "requeue failed jobs"
- "clear completed jobs"
- "pause the queue"

#### Navigation
- "go to workers"
- "show charts"
- "navigate to settings"

#### Confirmations
- "yes" / "confirm" / "proceed"
- "no" / "cancel" / "abort"

#### Help
- "help"
- "what can you do"
- "show commands"

## Integration Example

### Basic Setup

```go
package main

import (
    "context"
    "log"

    voice "github.com/flyingrobots/go-redis-work-queue/internal/terminal-voice-commands"
)

type MyTUI struct {
    // TUI implementation
}

func (t *MyTUI) GetQueueStatus() *voice.QueueStatus {
    // Return current queue status
    return &voice.QueueStatus{
        High:   10,
        Normal: 25,
        Low:    5,
        Total:  40,
    }
}

func (t *MyTUI) DrainWorker(id string) error {
    // Implement worker draining
    log.Printf("Draining worker %s", id)
    return nil
}

// Implement other TUIController methods...

func main() {
    // Create voice manager
    config := voice.DefaultVoiceConfig()
    config.LocalOnly = true
    config.AudioFeedback = true

    vm, err := voice.NewVoiceManager(context.Background(), config)
    if err != nil {
        log.Fatalf("Failed to create voice manager: %v", err)
    }

    // Start voice processing
    if err := vm.Start(); err != nil {
        log.Fatalf("Failed to start voice manager: %v", err)
    }
    defer vm.Stop()

    // Create TUI
    tui := &MyTUI{}

    // Process voice commands
    for {
        // In a real implementation, you would integrate this with your TUI event loop
        // Commands are processed automatically via audio input

        // For testing, you can process text commands directly:
        response, err := vm.ProcessCommand("show queue status", tui)
        if err != nil {
            log.Printf("Command failed: %v", err)
            continue
        }

        log.Printf("Command result: %s", response.Message)
    }
}
```

### Configuration Presets

```go
// High accuracy setup (cloud-based)
config := voice.GetVoicePresets()["high_accuracy"]

// Privacy-focused setup (local-only)
config := voice.GetVoicePresets()["privacy_focused"]

// Performance setup (minimal latency)
config := voice.GetVoicePresets()["performance"]

// Accessibility setup (enhanced feedback)
config := voice.GetVoicePresets()["accessibility"]
```

### Custom Wake Words

```go
config := voice.DefaultVoiceConfig()
config.WakeWord = "custom activation phrase"

detector, err := voice.NewWakeWordDetector(config.WakeWord)
```

### Command Validation

```go
processor, _ := voice.NewCommandProcessor()

cmd := &voice.Command{
    Intent:     voice.IntentWorkerControl,
    Confidence: 0.8,
    Entities: []voice.Entity{
        {Type: voice.EntityWorkerID, Value: "1"},
    },
}

result := processor.ValidateCommand(cmd)
if !result.Valid {
    log.Printf("Validation errors: %v", result.Errors)
}
```

## Performance Considerations

### Recognition Latency
- **Local (Whisper)**: 200-500ms
- **Cloud**: 100-300ms + network latency
- **Wake word detection**: <50ms

### Memory Usage
- **Whisper base model**: ~150MB
- **Audio buffers**: 2-4MB
- **Command history**: ~10KB per command

### Optimization Tips

1. **Use appropriate model size**: Base model for speed, small model for accuracy
2. **Adjust confidence threshold**: Lower for responsive, higher for accuracy
3. **Enable local-only mode**: Eliminates network latency and privacy concerns
4. **Configure audio feedback**: Disable for performance-critical applications

## Error Handling

### Common Error Types

```go
const (
    ErrorTypeRecognition   = "recognition"
    ErrorTypeProcessing    = "processing"
    ErrorTypeAudioDevice   = "audio_device"
    ErrorTypeNetwork       = "network"
    ErrorTypeModel         = "model"
    ErrorTypePermission    = "permission"
    ErrorTypeConfiguration = "configuration"
)
```

### Recovery Strategies

```go
func handleVoiceError(err error) {
    switch {
    case isAudioDeviceError(err):
        // Disable voice mode, continue with keyboard
        log.Printf("Audio device unavailable, switching to keyboard mode")
    case isNetworkError(err):
        // Fall back to local recognition
        log.Printf("Network error, switching to local recognition")
    case isModelError(err):
        // Reload model or use alternative
        log.Printf("Model error, attempting to reload")
    default:
        // Request user to repeat command
        log.Printf("Recognition error, please repeat command")
    }
}
```

## Metrics and Monitoring

### VoiceMetrics
```go
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
```

### Prometheus Integration

```go
// Example metrics collection
func (v *VoiceManager) recordMetrics() {
    recognitionLatencyHistogram.Observe(v.lastProcessingTime.Seconds())
    recognitionAccuracyGauge.Set(v.lastConfidence)
    commandsProcessedCounter.Inc()

    if v.lastError != nil {
        errorType := classifyError(v.lastError)
        voiceErrorsCounter.WithLabelValues(errorType).Inc()
    }
}
```

## Security Considerations

### Privacy Protection
- **Local processing**: Whisper.cpp runs entirely offline
- **Data sanitization**: Automatic removal of sensitive information
- **No recording**: Audio is processed in real-time, not stored
- **Explicit consent**: Cloud services require user opt-in

### Access Control
- **Command validation**: Prevent dangerous operations
- **Context awareness**: Limit commands based on current view
- **Confirmation prompts**: Require explicit confirmation for destructive actions

### Audit Logging
```go
// Sanitized command logging
log.Printf("Voice command executed: intent=%s, success=%v, user=%s",
    cmd.Intent, response.Success, getCurrentUser())
```

## Troubleshooting

### Common Issues

1. **Recognition not working**
   - Check microphone permissions
   - Verify audio device availability
   - Test with different confidence thresholds

2. **Wake word not detected**
   - Adjust detection threshold
   - Check background noise levels
   - Verify wake word pronunciation

3. **Commands not understood**
   - Use supported command grammar
   - Check language settings
   - Increase confidence threshold

4. **High latency**
   - Switch to local recognition
   - Use smaller Whisper model
   - Optimize audio buffer sizes

### Debug Commands

```go
// Enable debug logging
config.LogLevel = "debug"

// Test recognition directly
recognition, err := recognizer.ProcessAudio(audioData)
log.Printf("Recognition: %+v", recognition)

// Validate command parsing
cmd := &Command{RawText: "test command"}
processor.ParseCommand(cmd)
log.Printf("Parsed command: %+v", cmd)
```

## API Reference Summary

| Component | Key Functions | Purpose |
|-----------|---------------|---------|
| VoiceManager | NewVoiceManager, Start, Stop, ProcessCommand | Main voice control interface |
| ConfigManager | NewConfigManager, ApplyPreset, Load, Save | Configuration management |
| SpeechRecognizer | ProcessAudio, SetLanguage | Speech-to-text conversion |
| CommandProcessor | ParseCommand, ValidateCommand | Natural language understanding |
| AudioFeedback | SpeakResponse, PlayConfirmationSound | Text-to-speech and audio cues |
| WakeWordDetector | DetectWakeWord, SetThreshold | Wake word detection |
| PrivacyManager | SanitizeCommand | Privacy protection |

This API provides comprehensive voice command functionality while maintaining privacy, security, and performance requirements for terminal-based queue management applications.
