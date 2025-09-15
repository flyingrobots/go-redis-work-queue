package jsonpayloadstudio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// DefaultConfig returns the default configuration for JSON Payload Studio
func DefaultConfig() *StudioConfig {
	return &StudioConfig{
		// Editor settings
		EditorTheme:     "dark",
		SyntaxHighlight: true,
		LineNumbers:     true,
		AutoFormat:      false,
		BracketMatching: true,
		AutoComplete:    true,
		ValidateOnType:  true,

		// Template settings
		TemplatesPath:     "config/templates",
		AutoLoadTemplates: true,
		TemplateDirs:      []string{"config/templates", "templates"},

		// Schema settings
		SchemasPath:      "config/schemas",
		DefaultSchema:    "",
		StrictValidation: false,

		// Safety settings
		MaxPayloadSize:  10 * 1024 * 1024, // 10MB
		MaxFieldCount:   10000,
		MaxNestingDepth: 50,
		StripSecrets:    true,
		SecretPatterns: []string{
			"password",
			"passwd",
			"pwd",
			"secret",
			"token",
			"api_key",
			"apikey",
			"auth",
			"credential",
			"private_key",
			"privatekey",
			"Bearer\\s+[\\w-]+",
		},
		RequireConfirm: true,

		// UI settings
		ShowPreview:      true,
		PreviewLines:     20,
		HistorySize:      100,
		AutoSave:         true,
		AutoSaveInterval: 30 * time.Second,
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*StudioConfig, error) {
	config := DefaultConfig()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func (c *StudioConfig) SaveConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *StudioConfig) Validate() error {
	if c.MaxPayloadSize <= 0 {
		return fmt.Errorf("max_payload_size must be positive")
	}

	if c.MaxFieldCount <= 0 {
		return fmt.Errorf("max_field_count must be positive")
	}

	if c.MaxNestingDepth <= 0 {
		return fmt.Errorf("max_nesting_depth must be positive")
	}

	if c.PreviewLines <= 0 {
		c.PreviewLines = 20
	}

	if c.HistorySize < 0 {
		c.HistorySize = 0
	}

	if c.AutoSaveInterval < 0 {
		c.AutoSaveInterval = 0
		c.AutoSave = false
	}

	return nil
}

// ConfigBuilder provides a fluent interface for building configurations
type ConfigBuilder struct {
	config *StudioConfig
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// WithEditorTheme sets the editor theme
func (b *ConfigBuilder) WithEditorTheme(theme string) *ConfigBuilder {
	b.config.EditorTheme = theme
	return b
}

// WithSyntaxHighlight enables or disables syntax highlighting
func (b *ConfigBuilder) WithSyntaxHighlight(enabled bool) *ConfigBuilder {
	b.config.SyntaxHighlight = enabled
	return b
}

// WithLineNumbers enables or disables line numbers
func (b *ConfigBuilder) WithLineNumbers(enabled bool) *ConfigBuilder {
	b.config.LineNumbers = enabled
	return b
}

// WithAutoFormat enables or disables auto-formatting
func (b *ConfigBuilder) WithAutoFormat(enabled bool) *ConfigBuilder {
	b.config.AutoFormat = enabled
	return b
}

// WithAutoComplete enables or disables auto-completion
func (b *ConfigBuilder) WithAutoComplete(enabled bool) *ConfigBuilder {
	b.config.AutoComplete = enabled
	return b
}

// WithValidateOnType enables or disables validation while typing
func (b *ConfigBuilder) WithValidateOnType(enabled bool) *ConfigBuilder {
	b.config.ValidateOnType = enabled
	return b
}

// WithTemplatesPath sets the templates directory path
func (b *ConfigBuilder) WithTemplatesPath(path string) *ConfigBuilder {
	b.config.TemplatesPath = path
	return b
}

// WithTemplateDirs sets additional template directories
func (b *ConfigBuilder) WithTemplateDirs(dirs []string) *ConfigBuilder {
	b.config.TemplateDirs = dirs
	return b
}

// WithSchemasPath sets the schemas directory path
func (b *ConfigBuilder) WithSchemasPath(path string) *ConfigBuilder {
	b.config.SchemasPath = path
	return b
}

// WithDefaultSchema sets the default schema to use
func (b *ConfigBuilder) WithDefaultSchema(schema string) *ConfigBuilder {
	b.config.DefaultSchema = schema
	return b
}

// WithStrictValidation enables or disables strict schema validation
func (b *ConfigBuilder) WithStrictValidation(enabled bool) *ConfigBuilder {
	b.config.StrictValidation = enabled
	return b
}

// WithMaxPayloadSize sets the maximum payload size in bytes
func (b *ConfigBuilder) WithMaxPayloadSize(size int) *ConfigBuilder {
	b.config.MaxPayloadSize = size
	return b
}

// WithMaxFieldCount sets the maximum number of fields
func (b *ConfigBuilder) WithMaxFieldCount(count int) *ConfigBuilder {
	b.config.MaxFieldCount = count
	return b
}

// WithMaxNestingDepth sets the maximum nesting depth
func (b *ConfigBuilder) WithMaxNestingDepth(depth int) *ConfigBuilder {
	b.config.MaxNestingDepth = depth
	return b
}

// WithSecretStripping enables or disables secret stripping
func (b *ConfigBuilder) WithSecretStripping(enabled bool) *ConfigBuilder {
	b.config.StripSecrets = enabled
	return b
}

// WithSecretPatterns sets the patterns to identify secrets
func (b *ConfigBuilder) WithSecretPatterns(patterns []string) *ConfigBuilder {
	b.config.SecretPatterns = patterns
	return b
}

// WithRequireConfirm enables or disables confirmation before enqueue
func (b *ConfigBuilder) WithRequireConfirm(enabled bool) *ConfigBuilder {
	b.config.RequireConfirm = enabled
	return b
}

// WithShowPreview enables or disables preview panel
func (b *ConfigBuilder) WithShowPreview(enabled bool) *ConfigBuilder {
	b.config.ShowPreview = enabled
	return b
}

// WithPreviewLines sets the number of preview lines
func (b *ConfigBuilder) WithPreviewLines(lines int) *ConfigBuilder {
	b.config.PreviewLines = lines
	return b
}

// WithHistorySize sets the history size
func (b *ConfigBuilder) WithHistorySize(size int) *ConfigBuilder {
	b.config.HistorySize = size
	return b
}

// WithAutoSave enables or disables auto-save
func (b *ConfigBuilder) WithAutoSave(enabled bool) *ConfigBuilder {
	b.config.AutoSave = enabled
	return b
}

// WithAutoSaveInterval sets the auto-save interval
func (b *ConfigBuilder) WithAutoSaveInterval(interval time.Duration) *ConfigBuilder {
	b.config.AutoSaveInterval = interval
	return b
}

// Build builds and validates the configuration
func (b *ConfigBuilder) Build() (*StudioConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// ConfigWatcher watches for configuration file changes
type ConfigWatcher struct {
	path      string
	config    *StudioConfig
	onChange  func(*StudioConfig)
	stopChan  chan bool
	interval  time.Duration
	lastMod   time.Time
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(path string, onChange func(*StudioConfig)) *ConfigWatcher {
	return &ConfigWatcher{
		path:     path,
		onChange: onChange,
		stopChan: make(chan bool),
		interval: 5 * time.Second,
	}
}

// Start starts watching for configuration changes
func (w *ConfigWatcher) Start() error {
	// Load initial config
	config, err := LoadConfig(w.path)
	if err != nil {
		return err
	}
	w.config = config

	// Get initial modification time
	info, err := os.Stat(w.path)
	if err == nil {
		w.lastMod = info.ModTime()
	}

	// Start watching
	go w.watch()

	return nil
}

// Stop stops watching for configuration changes
func (w *ConfigWatcher) Stop() {
	close(w.stopChan)
}

// GetConfig returns the current configuration
func (w *ConfigWatcher) GetConfig() *StudioConfig {
	return w.config
}

func (w *ConfigWatcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkForChanges()
		case <-w.stopChan:
			return
		}
	}
}

func (w *ConfigWatcher) checkForChanges() {
	info, err := os.Stat(w.path)
	if err != nil {
		return
	}

	if info.ModTime().After(w.lastMod) {
		w.lastMod = info.ModTime()

		config, err := LoadConfig(w.path)
		if err != nil {
			return
		}

		w.config = config
		if w.onChange != nil {
			w.onChange(config)
		}
	}
}

// EnvironmentConfig loads configuration with environment variable overrides
type EnvironmentConfig struct {
	base *StudioConfig
}

// NewEnvironmentConfig creates a new environment-aware configuration
func NewEnvironmentConfig(base *StudioConfig) *EnvironmentConfig {
	return &EnvironmentConfig{
		base: base,
	}
}

// Load loads configuration with environment variable overrides
func (ec *EnvironmentConfig) Load() *StudioConfig {
	config := ec.base

	// Override with environment variables
	if theme := os.Getenv("JSON_STUDIO_THEME"); theme != "" {
		config.EditorTheme = theme
	}

	if maxSize := os.Getenv("JSON_STUDIO_MAX_PAYLOAD_SIZE"); maxSize != "" {
		var size int
		fmt.Sscanf(maxSize, "%d", &size)
		if size > 0 {
			config.MaxPayloadSize = size
		}
	}

	if maxFields := os.Getenv("JSON_STUDIO_MAX_FIELDS"); maxFields != "" {
		var count int
		fmt.Sscanf(maxFields, "%d", &count)
		if count > 0 {
			config.MaxFieldCount = count
		}
	}

	if maxDepth := os.Getenv("JSON_STUDIO_MAX_DEPTH"); maxDepth != "" {
		var depth int
		fmt.Sscanf(maxDepth, "%d", &depth)
		if depth > 0 {
			config.MaxNestingDepth = depth
		}
	}

	if stripSecrets := os.Getenv("JSON_STUDIO_STRIP_SECRETS"); stripSecrets != "" {
		config.StripSecrets = stripSecrets == "true" || stripSecrets == "1"
	}

	if requireConfirm := os.Getenv("JSON_STUDIO_REQUIRE_CONFIRM"); requireConfirm != "" {
		config.RequireConfirm = requireConfirm == "true" || requireConfirm == "1"
	}

	if autoSave := os.Getenv("JSON_STUDIO_AUTO_SAVE"); autoSave != "" {
		config.AutoSave = autoSave == "true" || autoSave == "1"
	}

	return config
}