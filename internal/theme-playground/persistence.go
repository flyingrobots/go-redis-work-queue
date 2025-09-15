// Copyright 2025 James Ross
package themeplayground

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PersistenceManager handles theme and preference persistence
type PersistenceManager struct {
	mu           sync.RWMutex
	configDir    string
	themesDir    string
	prefsFile    string
	customThemes map[string]*Theme
	preferences  *ThemePreferences
}

// PersistenceConfig configures the persistence layer
type PersistenceConfig struct {
	ConfigDir string // Base config directory (defaults to ~/.config/go-redis-work-queue)
	ThemesDir string // Custom themes directory (defaults to ConfigDir/themes)
	PrefsFile string // Preferences file (defaults to ConfigDir/preferences.json)
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(config *PersistenceConfig) (*PersistenceManager, error) {
	if config == nil {
		config = &PersistenceConfig{}
	}

	// Set defaults
	if config.ConfigDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		config.ConfigDir = filepath.Join(homeDir, ".config", "go-redis-work-queue")
	}

	if config.ThemesDir == "" {
		config.ThemesDir = filepath.Join(config.ConfigDir, "themes")
	}

	if config.PrefsFile == "" {
		config.PrefsFile = filepath.Join(config.ConfigDir, "preferences.json")
	}

	pm := &PersistenceManager{
		configDir:    config.ConfigDir,
		themesDir:    config.ThemesDir,
		prefsFile:    config.PrefsFile,
		customThemes: make(map[string]*Theme),
	}

	// Ensure directories exist
	if err := pm.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Load existing data
	if err := pm.loadPreferences(); err != nil {
		// If preferences don't exist, create defaults
		pm.preferences = pm.createDefaultPreferences()
		if err := pm.savePreferences(); err != nil {
			return nil, fmt.Errorf("failed to save default preferences: %w", err)
		}
	}

	if err := pm.loadCustomThemes(); err != nil {
		return nil, fmt.Errorf("failed to load custom themes: %w", err)
	}

	return pm, nil
}

// ensureDirectories creates necessary directories
func (pm *PersistenceManager) ensureDirectories() error {
	dirs := []string{pm.configDir, pm.themesDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// createDefaultPreferences creates default theme preferences
func (pm *PersistenceManager) createDefaultPreferences() *ThemePreferences {
	now := time.Now()
	return &ThemePreferences{
		ActiveTheme:        ThemeDefault,
		AutoDetectTerminal: true,
		RespectNoColor:     true,
		SyncWithSystem:     false,
		AccessibilityMode:  false,
		MotionReduced:      false,
		Overrides:          make(map[string]string),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// LoadPreferences loads theme preferences from disk
func (pm *PersistenceManager) loadPreferences() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.prefsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return err // Let caller handle missing file
		}
		return fmt.Errorf("failed to read preferences file: %w", err)
	}

	var prefs ThemePreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	pm.preferences = &prefs
	return nil
}

// SavePreferences saves theme preferences to disk
func (pm *PersistenceManager) savePreferences() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.preferences == nil {
		pm.preferences = pm.createDefaultPreferences()
	}

	pm.preferences.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(pm.preferences, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	if err := os.WriteFile(pm.prefsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %w", err)
	}

	return nil
}

// savePreferencesUnsafe saves theme preferences without acquiring a lock
func (pm *PersistenceManager) savePreferencesUnsafe() error {
	if pm.preferences == nil {
		pm.preferences = pm.createDefaultPreferences()
	}

	pm.preferences.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(pm.preferences, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	if err := os.WriteFile(pm.prefsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %w", err)
	}

	return nil
}

// GetPreferences returns a copy of current preferences
func (pm *PersistenceManager) GetPreferences() *ThemePreferences {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.preferences == nil {
		return pm.createDefaultPreferences()
	}

	// Return a copy to prevent external mutation
	prefsCopy := *pm.preferences
	if pm.preferences.CustomAccent != nil {
		accentCopy := *pm.preferences.CustomAccent
		prefsCopy.CustomAccent = &accentCopy
	}

	if pm.preferences.Overrides != nil {
		prefsCopy.Overrides = make(map[string]string)
		for k, v := range pm.preferences.Overrides {
			prefsCopy.Overrides[k] = v
		}
	}

	return &prefsCopy
}

// UpdatePreferences updates and saves preferences
func (pm *PersistenceManager) UpdatePreferences(prefs *ThemePreferences) error {
	pm.mu.Lock()
	pm.preferences = prefs
	pm.mu.Unlock()

	return pm.savePreferences()
}

// LoadCustomThemes loads all custom themes from disk
func (pm *PersistenceManager) loadCustomThemes() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Clear existing custom themes
	pm.customThemes = make(map[string]*Theme)

	// Read themes directory
	entries, err := os.ReadDir(pm.themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No custom themes directory yet
		}
		return fmt.Errorf("failed to read themes directory: %w", err)
	}

	// Load each .json file as a theme
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		themePath := filepath.Join(pm.themesDir, entry.Name())
		theme, err := pm.loadThemeFromFile(themePath)
		if err != nil {
			// Log error but continue loading other themes
			fmt.Printf("Warning: failed to load theme from %s: %v\n", themePath, err)
			continue
		}

		pm.customThemes[theme.Name] = theme
	}

	return nil
}

// loadThemeFromFile loads a single theme from a JSON file
func (pm *PersistenceManager) loadThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("failed to unmarshal theme: %w", err)
	}

	// Validate theme
	if err := validateTheme(&theme); err != nil {
		return nil, fmt.Errorf("theme validation failed: %w", err)
	}

	return &theme, nil
}

// SaveCustomTheme saves a custom theme to disk
func (pm *PersistenceManager) SaveCustomTheme(theme *Theme) error {
	if theme == nil {
		return ErrThemeInvalid.WithDetails("theme cannot be nil")
	}

	// Validate theme
	if err := validateTheme(theme); err != nil {
		return ErrThemeInvalid.WithDetails(err.Error())
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update timestamps
	now := time.Now()
	if theme.CreatedAt.IsZero() {
		theme.CreatedAt = now
	}
	theme.UpdatedAt = now

	// Save to disk
	filename := sanitizeFilename(theme.Name) + ".json"
	themePath := filepath.Join(pm.themesDir, filename)

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Update in-memory cache
	pm.customThemes[theme.Name] = theme

	return nil
}

// saveCustomThemeUnsafe saves a custom theme without acquiring a lock
// This method should only be called when the caller already holds the lock
func (pm *PersistenceManager) saveCustomThemeUnsafe(theme *Theme) error {
	// Update timestamps
	now := time.Now()
	if theme.CreatedAt.IsZero() {
		theme.CreatedAt = now
	}
	theme.UpdatedAt = now

	// Save to disk
	filename := sanitizeFilename(theme.Name) + ".json"
	themePath := filepath.Join(pm.themesDir, filename)

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Update in-memory cache
	pm.customThemes[theme.Name] = theme

	return nil
}

// DeleteCustomTheme removes a custom theme from disk and memory
func (pm *PersistenceManager) DeleteCustomTheme(themeName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if theme exists
	if _, exists := pm.customThemes[themeName]; !exists {
		return ErrThemeNotFound.WithDetails(themeName)
	}

	// Remove from disk
	filename := sanitizeFilename(themeName) + ".json"
	themePath := filepath.Join(pm.themesDir, filename)

	if err := os.Remove(themePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete theme file: %w", err)
	}

	// Remove from memory
	delete(pm.customThemes, themeName)

	return nil
}

// GetCustomThemes returns a copy of all custom themes
func (pm *PersistenceManager) GetCustomThemes() map[string]*Theme {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	themes := make(map[string]*Theme)
	for name, theme := range pm.customThemes {
		themeCopy := *theme
		themes[name] = &themeCopy
	}

	return themes
}

// GetCustomTheme returns a specific custom theme
func (pm *PersistenceManager) GetCustomTheme(themeName string) (*Theme, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	theme, exists := pm.customThemes[themeName]
	if !exists {
		return nil, false
	}

	themeCopy := *theme
	return &themeCopy, true
}

// ExportTheme exports a theme to a specified file path
func (pm *PersistenceManager) ExportTheme(themeName string, filePath string) error {
	pm.mu.RLock()
	theme, exists := pm.customThemes[themeName]
	pm.mu.RUnlock()

	if !exists {
		return ErrThemeNotFound.WithDetails(themeName)
	}

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme for export: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write exported theme: %w", err)
	}

	return nil
}

// ImportTheme imports a theme from a specified file path
func (pm *PersistenceManager) ImportTheme(filePath string) (*Theme, error) {
	theme, err := pm.loadThemeFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import theme: %w", err)
	}

	// Check if theme already exists
	pm.mu.RLock()
	_, exists := pm.customThemes[theme.Name]
	pm.mu.RUnlock()

	if exists {
		return nil, ErrThemeExists.WithDetails(theme.Name)
	}

	// Save as custom theme
	if err := pm.SaveCustomTheme(theme); err != nil {
		return nil, fmt.Errorf("failed to save imported theme: %w", err)
	}

	return theme, nil
}

// BackupThemes creates a backup of all custom themes
func (pm *PersistenceManager) BackupThemes(backupPath string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	backup := struct {
		Timestamp time.Time           `json:"timestamp"`
		Themes    map[string]*Theme   `json:"themes"`
		Preferences *ThemePreferences `json:"preferences"`
	}{
		Timestamp:   time.Now(),
		Themes:      pm.customThemes,
		Preferences: pm.preferences,
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// RestoreThemes restores themes from a backup file
func (pm *PersistenceManager) RestoreThemes(backupPath string, overwrite bool) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var backup struct {
		Timestamp   time.Time           `json:"timestamp"`
		Themes      map[string]*Theme   `json:"themes"`
		Preferences *ThemePreferences   `json:"preferences"`
	}

	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Restore themes
	for name, theme := range backup.Themes {
		if !overwrite {
			if _, exists := pm.customThemes[name]; exists {
				continue // Skip existing themes
			}
		}

		// Save theme directly without additional locking since we already hold the lock
		if err := pm.saveCustomThemeUnsafe(theme); err != nil {
			return fmt.Errorf("failed to restore theme %s: %w", name, err)
		}
	}

	// Restore preferences if provided
	if backup.Preferences != nil && overwrite {
		pm.preferences = backup.Preferences
		if err := pm.savePreferencesUnsafe(); err != nil {
			return fmt.Errorf("failed to restore preferences: %w", err)
		}
	}

	return nil
}

// validateTheme performs basic validation on a theme
func validateTheme(theme *Theme) error {
	if theme.Name == "" {
		return fmt.Errorf("theme name cannot be empty")
	}

	if theme.Version == "" {
		theme.Version = "1.0.0"
	}

	// Validate color palette
	if err := validateColorPalette(&theme.Palette); err != nil {
		return fmt.Errorf("invalid color palette: %w", err)
	}

	return nil
}

// validateColorPalette validates a color palette
func validateColorPalette(palette *ColorPalette) error {
	// Check required colors have valid hex values
	requiredColors := map[string]Color{
		"background":    palette.Background,
		"primary":       palette.Primary,
		"text_primary":  palette.TextPrimary,
	}

	for name, color := range requiredColors {
		if color.Hex == "" {
			return fmt.Errorf("required color %s has empty hex value", name)
		}

		if !isValidHexColor(color.Hex) {
			return fmt.Errorf("invalid hex color for %s: %s", name, color.Hex)
		}
	}

	return nil
}

// isValidHexColor checks if a string is a valid hex color
func isValidHexColor(hex string) bool {
	if len(hex) != 7 || hex[0] != '#' {
		return false
	}

	for i := 1; i < 7; i++ {
		c := hex[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// sanitizeFilename sanitizes a filename by removing invalid characters
func sanitizeFilename(name string) string {
	// Replace invalid filename characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Limit length
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}