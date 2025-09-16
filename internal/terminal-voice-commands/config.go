package voice

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigManager handles voice command configuration
type ConfigManager struct {
	configPath string
	config     *VoiceConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) (*ConfigManager, error) {
	if configPath == "" {
		// Use default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(home, ".config", "go-redis-work-queue", "voice.yaml")
	}

	cm := &ConfigManager{
		configPath: configPath,
	}

	// Load existing config or create default
	if err := cm.Load(); err != nil {
		// If config doesn't exist, create default
		cm.config = DefaultVoiceConfig()
		if err := cm.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return cm, nil
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	config := &VoiceConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.config = config
	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *VoiceConfig {
	return cm.config
}

// UpdateConfig updates the configuration
func (cm *ConfigManager) UpdateConfig(config *VoiceConfig) error {
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.config = config
	return cm.Save()
}

// validateConfig validates configuration values
func (cm *ConfigManager) validateConfig(config *VoiceConfig) error {
	if config.WakeWord == "" {
		return fmt.Errorf("wake word cannot be empty")
	}

	if config.Language == "" {
		return fmt.Errorf("language cannot be empty")
	}

	if config.ConfidenceThreshold < 0.0 || config.ConfidenceThreshold > 1.0 {
		return fmt.Errorf("confidence threshold must be between 0.0 and 1.0")
	}

	if config.ProcessingTimeout <= 0 {
		return fmt.Errorf("processing timeout must be positive")
	}

	validBackends := []string{"whisper", "google", "azure", "aws"}
	backendValid := false
	for _, backend := range validBackends {
		if config.RecognitionBackend == backend {
			backendValid = true
			break
		}
	}
	if !backendValid {
		return fmt.Errorf("invalid recognition backend: %s", config.RecognitionBackend)
	}

	return nil
}

// LoadFromJSON loads configuration from JSON file
func (cm *ConfigManager) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON config file: %w", err)
	}

	config := &VoiceConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse JSON config file: %w", err)
	}

	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.config = config
	return nil
}

// SaveToJSON saves configuration to JSON file
func (cm *ConfigManager) SaveToJSON(path string) error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// SetWakeWord updates the wake word configuration
func (cm *ConfigManager) SetWakeWord(wakeWord string) error {
	if wakeWord == "" {
		return fmt.Errorf("wake word cannot be empty")
	}

	cm.config.WakeWord = wakeWord
	return cm.Save()
}

// SetRecognitionBackend updates the recognition backend
func (cm *ConfigManager) SetRecognitionBackend(backend string) error {
	validBackends := []string{"whisper", "google", "azure", "aws"}
	for _, valid := range validBackends {
		if backend == valid {
			cm.config.RecognitionBackend = backend
			return cm.Save()
		}
	}

	return fmt.Errorf("invalid recognition backend: %s", backend)
}

// SetLanguage updates the recognition language
func (cm *ConfigManager) SetLanguage(language string) error {
	if language == "" {
		return fmt.Errorf("language cannot be empty")
	}

	cm.config.Language = language
	return cm.Save()
}

// SetConfidenceThreshold updates the confidence threshold
func (cm *ConfigManager) SetConfidenceThreshold(threshold float64) error {
	if threshold < 0.0 || threshold > 1.0 {
		return fmt.Errorf("confidence threshold must be between 0.0 and 1.0")
	}

	cm.config.ConfidenceThreshold = threshold
	return cm.Save()
}

// SetLocalOnly updates the local-only setting
func (cm *ConfigManager) SetLocalOnly(localOnly bool) error {
	cm.config.LocalOnly = localOnly
	return cm.Save()
}

// SetAudioFeedback updates the audio feedback setting
func (cm *ConfigManager) SetAudioFeedback(enabled bool) error {
	cm.config.AudioFeedback = enabled
	return cm.Save()
}

// SetProcessingTimeout updates the processing timeout
func (cm *ConfigManager) SetProcessingTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("processing timeout must be positive")
	}

	cm.config.ProcessingTimeout = timeout
	return cm.Save()
}

// ResetToDefaults resets configuration to default values
func (cm *ConfigManager) ResetToDefaults() error {
	cm.config = DefaultVoiceConfig()
	return cm.Save()
}

// ExportConfig exports configuration for backup/sharing
func (cm *ConfigManager) ExportConfig() (map[string]interface{}, error) {
	data, err := json.Marshal(cm.config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var export map[string]interface{}
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Add metadata
	export["exported_at"] = time.Now().UTC()
	export["version"] = "1.0"

	return export, nil
}

// ImportConfig imports configuration from exported data
func (cm *ConfigManager) ImportConfig(data map[string]interface{}) error {
	// Remove metadata fields
	delete(data, "exported_at")
	delete(data, "version")

	configData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal import data: %w", err)
	}

	config := &VoiceConfig{}
	if err := json.Unmarshal(configData, config); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid imported configuration: %w", err)
	}

	cm.config = config
	return cm.Save()
}

// GetVoicePresets returns predefined voice configuration presets
func GetVoicePresets() map[string]*VoiceConfig {
	return map[string]*VoiceConfig{
		"default": {
			WakeWord:            "hey queue",
			RecognitionBackend:  "whisper",
			LocalOnly:           true,
			AudioFeedback:       true,
			Language:            "en",
			ConfidenceThreshold: 0.7,
			ProcessingTimeout:   5 * time.Second,
			NoAudioRecording:    true,
			SanitizeLogs:        true,
		},
		"high_accuracy": {
			WakeWord:            "hey queue",
			RecognitionBackend:  "google",
			LocalOnly:           false,
			AudioFeedback:       true,
			Language:            "en",
			ConfidenceThreshold: 0.8,
			ProcessingTimeout:   10 * time.Second,
			NoAudioRecording:    false,
			SanitizeLogs:        true,
		},
		"privacy_focused": {
			WakeWord:            "hey queue",
			RecognitionBackend:  "whisper",
			LocalOnly:           true,
			AudioFeedback:       false,
			Language:            "en",
			ConfidenceThreshold: 0.6,
			ProcessingTimeout:   3 * time.Second,
			NoAudioRecording:    true,
			SanitizeLogs:        true,
		},
		"performance": {
			WakeWord:            "queue",
			RecognitionBackend:  "whisper",
			LocalOnly:           true,
			AudioFeedback:       false,
			Language:            "en",
			ConfidenceThreshold: 0.5,
			ProcessingTimeout:   2 * time.Second,
			NoAudioRecording:    true,
			SanitizeLogs:        false,
		},
		"accessibility": {
			WakeWord:            "hey queue assistant",
			RecognitionBackend:  "whisper",
			LocalOnly:           true,
			AudioFeedback:       true,
			Language:            "en",
			ConfidenceThreshold: 0.6,
			ProcessingTimeout:   8 * time.Second,
			NoAudioRecording:    true,
			SanitizeLogs:        true,
		},
	}
}

// ApplyPreset applies a predefined configuration preset
func (cm *ConfigManager) ApplyPreset(presetName string) error {
	presets := GetVoicePresets()
	preset, exists := presets[presetName]
	if !exists {
		return fmt.Errorf("unknown preset: %s", presetName)
	}

	// Create a copy of the preset
	configData, err := json.Marshal(preset)
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	config := &VoiceConfig{}
	if err := json.Unmarshal(configData, config); err != nil {
		return fmt.Errorf("failed to unmarshal preset: %w", err)
	}

	cm.config = config
	return cm.Save()
}

// ListPresets returns available configuration presets
func (cm *ConfigManager) ListPresets() []string {
	presets := GetVoicePresets()
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	return names
}

// CompareWithPreset compares current config with a preset
func (cm *ConfigManager) CompareWithPreset(presetName string) (map[string]interface{}, error) {
	presets := GetVoicePresets()
	preset, exists := presets[presetName]
	if !exists {
		return nil, fmt.Errorf("unknown preset: %s", presetName)
	}

	differences := make(map[string]interface{})

	if cm.config.WakeWord != preset.WakeWord {
		differences["wake_word"] = map[string]string{
			"current": cm.config.WakeWord,
			"preset":  preset.WakeWord,
		}
	}

	if cm.config.RecognitionBackend != preset.RecognitionBackend {
		differences["recognition_backend"] = map[string]string{
			"current": cm.config.RecognitionBackend,
			"preset":  preset.RecognitionBackend,
		}
	}

	if cm.config.LocalOnly != preset.LocalOnly {
		differences["local_only"] = map[string]bool{
			"current": cm.config.LocalOnly,
			"preset":  preset.LocalOnly,
		}
	}

	if cm.config.AudioFeedback != preset.AudioFeedback {
		differences["audio_feedback"] = map[string]bool{
			"current": cm.config.AudioFeedback,
			"preset":  preset.AudioFeedback,
		}
	}

	if cm.config.Language != preset.Language {
		differences["language"] = map[string]string{
			"current": cm.config.Language,
			"preset":  preset.Language,
		}
	}

	if cm.config.ConfidenceThreshold != preset.ConfidenceThreshold {
		differences["confidence_threshold"] = map[string]float64{
			"current": cm.config.ConfidenceThreshold,
			"preset":  preset.ConfidenceThreshold,
		}
	}

	return differences, nil
}

// ValidateConfiguration performs comprehensive configuration validation
func ValidateConfiguration(config *VoiceConfig) []string {
	var issues []string

	if config.WakeWord == "" {
		issues = append(issues, "Wake word cannot be empty")
	}

	if len(config.WakeWord) > 50 {
		issues = append(issues, "Wake word is too long (max 50 characters)")
	}

	if config.Language == "" {
		issues = append(issues, "Language cannot be empty")
	}

	validLanguages := []string{"en", "es", "fr", "de", "ja", "zh", "ko"}
	languageValid := false
	for _, lang := range validLanguages {
		if config.Language == lang {
			languageValid = true
			break
		}
	}
	if !languageValid {
		issues = append(issues, fmt.Sprintf("Unsupported language: %s", config.Language))
	}

	if config.ConfidenceThreshold < 0.0 || config.ConfidenceThreshold > 1.0 {
		issues = append(issues, "Confidence threshold must be between 0.0 and 1.0")
	}

	if config.ConfidenceThreshold < 0.3 {
		issues = append(issues, "Warning: Very low confidence threshold may cause false positives")
	}

	if config.ProcessingTimeout <= 0 {
		issues = append(issues, "Processing timeout must be positive")
	}

	if config.ProcessingTimeout > 30*time.Second {
		issues = append(issues, "Warning: Very long processing timeout may impact responsiveness")
	}

	validBackends := []string{"whisper", "google", "azure", "aws"}
	backendValid := false
	for _, backend := range validBackends {
		if config.RecognitionBackend == backend {
			backendValid = true
			break
		}
	}
	if !backendValid {
		issues = append(issues, fmt.Sprintf("Invalid recognition backend: %s", config.RecognitionBackend))
	}

	if !config.LocalOnly && config.RecognitionBackend == "whisper" {
		issues = append(issues, "Warning: Whisper backend is local-only but LocalOnly is false")
	}

	return issues
}