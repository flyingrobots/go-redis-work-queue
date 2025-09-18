package voice

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfigManager(t *testing.T) {
	// Create temporary config directory
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	// Test creating new config manager
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	if cm == nil {
		t.Fatal("Config manager is nil")
	}

	if cm.configPath != configPath {
		t.Errorf("Expected config path '%s', got '%s'", configPath, cm.configPath)
	}

	// Config file should be created with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	config := cm.GetConfig()
	if config == nil {
		t.Fatal("Config is nil")
	}

	// Should have default values
	if config.WakeWord != "hey queue" {
		t.Errorf("Expected default wake word 'hey queue', got '%s'", config.WakeWord)
	}
}

func TestConfigManagerLoadSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	// Create config manager
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Modify config
	originalConfig := cm.GetConfig()
	originalConfig.WakeWord = "test word"
	originalConfig.ConfidenceThreshold = 0.9

	// Save modified config
	if err := cm.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create new config manager to test loading
	cm2, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create second config manager: %v", err)
	}

	loadedConfig := cm2.GetConfig()
	if loadedConfig.WakeWord != "test word" {
		t.Errorf("Expected wake word 'test word', got '%s'", loadedConfig.WakeWord)
	}

	if loadedConfig.ConfidenceThreshold != 0.9 {
		t.Errorf("Expected confidence threshold 0.9, got %f", loadedConfig.ConfidenceThreshold)
	}
}

func TestConfigManagerUpdateConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test valid config update
	newConfig := &VoiceConfig{
		WakeWord:            "new wake word",
		RecognitionBackend:  "google",
		LocalOnly:           false,
		AudioFeedback:       false,
		Language:            "es",
		ConfidenceThreshold: 0.8,
		ProcessingTimeout:   10 * time.Second,
		NoAudioRecording:    false,
		SanitizeLogs:        false,
	}

	if err := cm.UpdateConfig(newConfig); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Verify update
	updated := cm.GetConfig()
	if updated.WakeWord != "new wake word" {
		t.Errorf("Expected wake word 'new wake word', got '%s'", updated.WakeWord)
	}

	if updated.RecognitionBackend != "google" {
		t.Errorf("Expected backend 'google', got '%s'", updated.RecognitionBackend)
	}

	// Test invalid config update
	invalidConfig := &VoiceConfig{
		WakeWord:            "", // Invalid
		RecognitionBackend:  "whisper",
		Language:            "en",
		ConfidenceThreshold: 0.7,
		ProcessingTimeout:   5 * time.Second,
	}

	if err := cm.UpdateConfig(invalidConfig); err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestConfigManagerSetters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test SetWakeWord
	if err := cm.SetWakeWord("custom wake"); err != nil {
		t.Fatalf("Failed to set wake word: %v", err)
	}

	if cm.GetConfig().WakeWord != "custom wake" {
		t.Error("Wake word not updated")
	}

	// Test SetRecognitionBackend
	if err := cm.SetRecognitionBackend("azure"); err != nil {
		t.Fatalf("Failed to set backend: %v", err)
	}

	if cm.GetConfig().RecognitionBackend != "azure" {
		t.Error("Backend not updated")
	}

	// Test SetLanguage
	if err := cm.SetLanguage("fr"); err != nil {
		t.Fatalf("Failed to set language: %v", err)
	}

	if cm.GetConfig().Language != "fr" {
		t.Error("Language not updated")
	}

	// Test SetConfidenceThreshold
	if err := cm.SetConfidenceThreshold(0.85); err != nil {
		t.Fatalf("Failed to set threshold: %v", err)
	}

	if cm.GetConfig().ConfidenceThreshold != 0.85 {
		t.Error("Confidence threshold not updated")
	}

	// Test SetLocalOnly
	if err := cm.SetLocalOnly(false); err != nil {
		t.Fatalf("Failed to set local only: %v", err)
	}

	if cm.GetConfig().LocalOnly {
		t.Error("LocalOnly not updated")
	}

	// Test SetAudioFeedback
	if err := cm.SetAudioFeedback(false); err != nil {
		t.Fatalf("Failed to set audio feedback: %v", err)
	}

	if cm.GetConfig().AudioFeedback {
		t.Error("AudioFeedback not updated")
	}

	// Test SetProcessingTimeout
	if err := cm.SetProcessingTimeout(15 * time.Second); err != nil {
		t.Fatalf("Failed to set timeout: %v", err)
	}

	if cm.GetConfig().ProcessingTimeout != 15*time.Second {
		t.Error("ProcessingTimeout not updated")
	}
}

func TestConfigManagerInvalidSetters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test invalid wake word
	if err := cm.SetWakeWord(""); err == nil {
		t.Error("Expected error for empty wake word")
	}

	// Test invalid backend
	if err := cm.SetRecognitionBackend("invalid"); err == nil {
		t.Error("Expected error for invalid backend")
	}

	// Test invalid language
	if err := cm.SetLanguage(""); err == nil {
		t.Error("Expected error for empty language")
	}

	// Test invalid confidence threshold
	if err := cm.SetConfidenceThreshold(-0.1); err == nil {
		t.Error("Expected error for negative confidence threshold")
	}

	if err := cm.SetConfidenceThreshold(1.1); err == nil {
		t.Error("Expected error for confidence threshold > 1.0")
	}

	// Test invalid timeout
	if err := cm.SetProcessingTimeout(-1 * time.Second); err == nil {
		t.Error("Expected error for negative timeout")
	}
}

func TestConfigManagerJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")
	jsonPath := filepath.Join(tempDir, "voice.json")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Modify config
	cm.SetWakeWord("json test")
	cm.SetConfidenceThreshold(0.75)

	// Save to JSON
	if err := cm.SaveToJSON(jsonPath); err != nil {
		t.Fatalf("Failed to save JSON: %v", err)
	}

	// Verify JSON file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("JSON file was not created")
	}

	// Load from JSON
	if err := cm.LoadFromJSON(jsonPath); err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	// Verify loaded values
	config := cm.GetConfig()
	if config.WakeWord != "json test" {
		t.Errorf("Expected wake word 'json test', got '%s'", config.WakeWord)
	}

	if config.ConfidenceThreshold != 0.75 {
		t.Errorf("Expected confidence threshold 0.75, got %f", config.ConfidenceThreshold)
	}
}

func TestConfigManagerPresets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test listing presets
	presets := cm.ListPresets()
	if len(presets) == 0 {
		t.Error("No presets available")
	}

	expectedPresets := []string{"default", "high_accuracy", "privacy_focused", "performance", "accessibility"}
	for _, expected := range expectedPresets {
		found := false
		for _, preset := range presets {
			if preset == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected preset '%s' not found", expected)
		}
	}

	// Test applying preset
	if err := cm.ApplyPreset("privacy_focused"); err != nil {
		t.Fatalf("Failed to apply preset: %v", err)
	}

	config := cm.GetConfig()
	if !config.LocalOnly {
		t.Error("Privacy focused preset should have LocalOnly=true")
	}

	// Test invalid preset
	if err := cm.ApplyPreset("nonexistent"); err == nil {
		t.Error("Expected error for nonexistent preset")
	}
}

func TestConfigManagerComparison(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Modify current config
	cm.SetWakeWord("modified wake")
	cm.SetLocalOnly(false)

	// Compare with default preset
	differences, err := cm.CompareWithPreset("default")
	if err != nil {
		t.Fatalf("Failed to compare with preset: %v", err)
	}

	if len(differences) == 0 {
		t.Error("Expected differences but got none")
	}

	// Should have difference in wake word
	if _, exists := differences["wake_word"]; !exists {
		t.Error("Expected wake_word difference")
	}

	// Should have difference in local_only
	if _, exists := differences["local_only"]; !exists {
		t.Error("Expected local_only difference")
	}

	// Test comparison with invalid preset
	_, err = cm.CompareWithPreset("nonexistent")
	if err == nil {
		t.Error("Expected error for invalid preset comparison")
	}
}

func TestConfigManagerExportImport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Modify config
	cm.SetWakeWord("export test")
	cm.SetConfidenceThreshold(0.65)
	cm.SetLocalOnly(false)

	// Export config
	exported, err := cm.ExportConfig()
	if err != nil {
		t.Fatalf("Failed to export config: %v", err)
	}

	if exported == nil {
		t.Fatal("Exported config is nil")
	}

	// Check metadata
	if _, exists := exported["exported_at"]; !exists {
		t.Error("Expected exported_at metadata")
	}

	if _, exists := exported["version"]; !exists {
		t.Error("Expected version metadata")
	}

	// Modify current config
	cm.SetWakeWord("changed")

	// Import exported config
	if err := cm.ImportConfig(exported); err != nil {
		t.Fatalf("Failed to import config: %v", err)
	}

	// Verify imported values
	config := cm.GetConfig()
	if config.WakeWord != "export test" {
		t.Errorf("Expected wake word 'export test', got '%s'", config.WakeWord)
	}

	if config.ConfidenceThreshold != 0.65 {
		t.Errorf("Expected confidence threshold 0.65, got %f", config.ConfidenceThreshold)
	}
}

func TestConfigManagerResetToDefaults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "voice-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "voice.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Modify config
	cm.SetWakeWord("modified")
	cm.SetConfidenceThreshold(0.9)
	cm.SetLocalOnly(false)

	// Reset to defaults
	if err := cm.ResetToDefaults(); err != nil {
		t.Fatalf("Failed to reset to defaults: %v", err)
	}

	// Verify default values
	config := cm.GetConfig()
	defaults := DefaultVoiceConfig()

	if config.WakeWord != defaults.WakeWord {
		t.Errorf("Expected default wake word '%s', got '%s'", defaults.WakeWord, config.WakeWord)
	}

	if config.ConfidenceThreshold != defaults.ConfidenceThreshold {
		t.Errorf("Expected default confidence threshold %f, got %f", defaults.ConfidenceThreshold, config.ConfidenceThreshold)
	}

	if config.LocalOnly != defaults.LocalOnly {
		t.Errorf("Expected default LocalOnly %v, got %v", defaults.LocalOnly, config.LocalOnly)
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    *VoiceConfig
		expectErr bool
	}{
		{
			name:      "valid config",
			config:    DefaultVoiceConfig(),
			expectErr: false,
		},
		{
			name: "empty wake word",
			config: &VoiceConfig{
				WakeWord:            "",
				RecognitionBackend:  "whisper",
				Language:            "en",
				ConfidenceThreshold: 0.7,
				ProcessingTimeout:   5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid confidence threshold",
			config: &VoiceConfig{
				WakeWord:            "test",
				RecognitionBackend:  "whisper",
				Language:            "en",
				ConfidenceThreshold: -0.1,
				ProcessingTimeout:   5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid backend",
			config: &VoiceConfig{
				WakeWord:            "test",
				RecognitionBackend:  "invalid",
				Language:            "en",
				ConfidenceThreshold: 0.7,
				ProcessingTimeout:   5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "negative timeout",
			config: &VoiceConfig{
				WakeWord:            "test",
				RecognitionBackend:  "whisper",
				Language:            "en",
				ConfidenceThreshold: 0.7,
				ProcessingTimeout:   -1 * time.Second,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateConfiguration(tt.config)

			hasErrors := len(issues) > 0
			if hasErrors != tt.expectErr {
				t.Errorf("Expected hasErrors=%v, got hasErrors=%v (issues: %v)",
					tt.expectErr, hasErrors, issues)
			}
		})
	}
}

func TestGetVoicePresets(t *testing.T) {
	presets := GetVoicePresets()

	if len(presets) == 0 {
		t.Error("No voice presets available")
	}

	expectedPresets := []string{"default", "high_accuracy", "privacy_focused", "performance", "accessibility"}
	for _, expected := range expectedPresets {
		if _, exists := presets[expected]; !exists {
			t.Errorf("Expected preset '%s' not found", expected)
		}
	}

	// Validate each preset
	for name, preset := range presets {
		issues := ValidateConfiguration(preset)
		if len(issues) > 0 {
			t.Errorf("Preset '%s' has validation issues: %v", name, issues)
		}
	}
}

func TestConfigManagerEdgeCases(t *testing.T) {
	// Test with empty config path (should use default)
	cm, err := NewConfigManager("")
	if err != nil {
		t.Logf("Default config creation failed (expected in test env): %v", err)
		// This might fail in test environment without home directory
		return
	}

	if cm.GetConfigPath() == "" {
		t.Error("Expected non-empty config path")
	}

	// Test with invalid JSON import
	invalidJSON := map[string]interface{}{
		"confidence_threshold": "invalid", // Should be float64
	}

	if err := cm.ImportConfig(invalidJSON); err == nil {
		t.Error("Expected error for invalid JSON import")
	}
}
