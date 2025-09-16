// Copyright 2025 James Ross
package themeplayground

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the theme playground configuration
type Config struct {
	ThemeDir         string                 `json:"theme_dir"`
	PreferencesFile  string                 `json:"preferences_file"`
	CacheEnabled     bool                   `json:"cache_enabled"`
	CacheTTL         time.Duration          `json:"cache_ttl"`
	AutoDetect       bool                   `json:"auto_detect"`
	AccessibilityMode bool                  `json:"accessibility_mode"`
	BackupEnabled    bool                   `json:"backup_enabled"`
	BackupRetention  int                    `json:"backup_retention"`
	ValidationLevel  string                 `json:"validation_level"`
	DefaultTheme     string                 `json:"default_theme"`
	Features         map[string]bool        `json:"features"`
	Limits           ConfigLimits           `json:"limits"`
	Integration      IntegrationConfig      `json:"integration"`
}

// ConfigLimits defines operational limits
type ConfigLimits struct {
	MaxCustomThemes     int `json:"max_custom_themes"`
	MaxThemeSize        int `json:"max_theme_size"`
	MaxBackupFiles      int `json:"max_backup_files"`
	MaxPreviewDuration  int `json:"max_preview_duration"`
	MaxCacheSize        int `json:"max_cache_size"`
	ValidationTimeout   int `json:"validation_timeout"`
}

// IntegrationConfig defines external integration settings
type IntegrationConfig struct {
	VSCodeSync       bool   `json:"vscode_sync"`
	TerminalDetect   bool   `json:"terminal_detect"`
	SystemThemeSync  bool   `json:"system_theme_sync"`
	CloudSync        bool   `json:"cloud_sync"`
	TeamSharing      bool   `json:"team_sharing"`
	WebhookURL       string `json:"webhook_url"`
	NotifyChanges    bool   `json:"notify_changes"`
}

// DefaultConfig returns the default configuration
func DefaultConfig(configDir string) *Config {
	return &Config{
		ThemeDir:         filepath.Join(configDir, "themes"),
		PreferencesFile:  filepath.Join(configDir, "theme_preferences.json"),
		CacheEnabled:     true,
		CacheTTL:         time.Hour * 24,
		AutoDetect:       true,
		AccessibilityMode: false,
		BackupEnabled:    true,
		BackupRetention:  10,
		ValidationLevel:  "strict",
		DefaultTheme:     ThemeDefault,
		Features: map[string]bool{
			"live_preview":     true,
			"custom_themes":    true,
			"import_export":    true,
			"accessibility":    true,
			"animations":       true,
			"color_picker":     true,
			"theme_builder":    true,
			"contrast_checker": true,
		},
		Limits: ConfigLimits{
			MaxCustomThemes:    50,
			MaxThemeSize:       1024 * 1024, // 1MB
			MaxBackupFiles:     20,
			MaxPreviewDuration: 30, // seconds
			MaxCacheSize:       10 * 1024 * 1024, // 10MB
			ValidationTimeout:  5, // seconds
		},
		Integration: IntegrationConfig{
			VSCodeSync:      false,
			TerminalDetect:  true,
			SystemThemeSync: false,
			CloudSync:       false,
			TeamSharing:     false,
			NotifyChanges:   false,
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(filepath.Dir(configPath)), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate and set defaults for missing fields
	defaultConfig := DefaultConfig(filepath.Dir(configPath))
	if config.ThemeDir == "" {
		config.ThemeDir = defaultConfig.ThemeDir
	}
	if config.PreferencesFile == "" {
		config.PreferencesFile = defaultConfig.PreferencesFile
	}
	if config.DefaultTheme == "" {
		config.DefaultTheme = defaultConfig.DefaultTheme
	}
	if config.Features == nil {
		config.Features = defaultConfig.Features
	}
	if config.ValidationLevel == "" {
		config.ValidationLevel = defaultConfig.ValidationLevel
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(configPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsFeatureEnabled checks if a feature is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	if c.Features == nil {
		return false
	}
	enabled, exists := c.Features[feature]
	return exists && enabled
}

// SetFeature enables or disables a feature
func (c *Config) SetFeature(feature string, enabled bool) {
	if c.Features == nil {
		c.Features = make(map[string]bool)
	}
	c.Features[feature] = enabled
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ThemeDir == "" {
		return fmt.Errorf("theme_dir is required")
	}

	if c.PreferencesFile == "" {
		return fmt.Errorf("preferences_file is required")
	}

	if c.DefaultTheme == "" {
		return fmt.Errorf("default_theme is required")
	}

	if c.Limits.MaxCustomThemes < 1 {
		return fmt.Errorf("max_custom_themes must be at least 1")
	}

	if c.Limits.MaxThemeSize < 1024 {
		return fmt.Errorf("max_theme_size must be at least 1024 bytes")
	}

	if c.ValidationLevel != "none" && c.ValidationLevel != "basic" && c.ValidationLevel != "strict" {
		return fmt.Errorf("validation_level must be one of: none, basic, strict")
	}

	return nil
}

// EnsureDirectories creates necessary directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.ThemeDir,
		filepath.Dir(c.PreferencesFile),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetThemeFile returns the full path for a theme file
func (c *Config) GetThemeFile(themeName string) string {
	return filepath.Join(c.ThemeDir, themeName+".json")
}

// GetBackupFile returns the full path for a theme backup file
func (c *Config) GetBackupFile(themeName string, timestamp time.Time) string {
	backupName := fmt.Sprintf("%s_%s.json.bak", themeName, timestamp.Format("20060102_150405"))
	return filepath.Join(c.ThemeDir, "backups", backupName)
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := &Config{
		ThemeDir:         c.ThemeDir,
		PreferencesFile:  c.PreferencesFile,
		CacheEnabled:     c.CacheEnabled,
		CacheTTL:         c.CacheTTL,
		AutoDetect:       c.AutoDetect,
		AccessibilityMode: c.AccessibilityMode,
		BackupEnabled:    c.BackupEnabled,
		BackupRetention:  c.BackupRetention,
		ValidationLevel:  c.ValidationLevel,
		DefaultTheme:     c.DefaultTheme,
		Features:         make(map[string]bool),
		Limits:           c.Limits,
		Integration:      c.Integration,
	}

	// Deep copy features map
	for k, v := range c.Features {
		clone.Features[k] = v
	}

	return clone
}

// MarshalJSON implements custom JSON marshaling for Duration fields
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal(&struct {
		CacheTTL string `json:"cache_ttl"`
		*Alias
	}{
		CacheTTL: c.CacheTTL.String(),
		Alias:    (*Alias)(&c),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Duration fields
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		CacheTTL string `json:"cache_ttl"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.CacheTTL != "" {
		duration, err := time.ParseDuration(aux.CacheTTL)
		if err != nil {
			return fmt.Errorf("invalid cache_ttl format: %w", err)
		}
		c.CacheTTL = duration
	}

	return nil
}