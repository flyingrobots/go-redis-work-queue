// Copyright 2025 James Ross
package themeplayground

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPersistenceManager(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	config := &PersistenceConfig{
		ConfigDir: tempDir,
	}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	if pm == nil {
		t.Fatal("NewPersistenceManager returned nil")
	}

	// Verify directories were created
	if _, err := os.Stat(pm.configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	if _, err := os.Stat(pm.themesDir); os.IsNotExist(err) {
		t.Error("Themes directory was not created")
	}

	// Verify default preferences were created
	prefs := pm.GetPreferences()
	if prefs == nil {
		t.Error("Default preferences were not created")
	}

	if prefs.ActiveTheme != ThemeDefault {
		t.Errorf("Expected default theme %s, got %s", ThemeDefault, prefs.ActiveTheme)
	}
}

func TestPersistenceManager_SaveLoadPreferences(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Modify preferences
	prefs := pm.GetPreferences()
	prefs.ActiveTheme = ThemeTokyoNight
	prefs.AutoDetectTerminal = false
	prefs.Overrides = map[string]string{"test": "value"}

	// Save preferences
	err = pm.UpdatePreferences(prefs)
	if err != nil {
		t.Errorf("Failed to save preferences: %v", err)
	}

	// Create new persistence manager to test loading
	pm2, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create second persistence manager: %v", err)
	}

	// Verify preferences were loaded correctly
	loadedPrefs := pm2.GetPreferences()
	if loadedPrefs.ActiveTheme != ThemeTokyoNight {
		t.Errorf("Expected theme %s, got %s", ThemeTokyoNight, loadedPrefs.ActiveTheme)
	}

	if loadedPrefs.AutoDetectTerminal {
		t.Error("AutoDetectTerminal should be false")
	}

	if loadedPrefs.Overrides["test"] != "value" {
		t.Error("Override value was not preserved")
	}
}

func TestPersistenceManager_SaveLoadCustomTheme(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Create test theme
	testTheme := &Theme{
		Name:        "test-custom-theme",
		Description: "A test custom theme",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Author:      "Test Author",
		Palette: ColorPalette{
			Background:  Color{Hex: "#1a1b26", Name: "Background"},
			Primary:     Color{Hex: "#7aa2f7", Name: "Primary"},
			TextPrimary: Color{Hex: "#c0caf5", Name: "Text Primary"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save theme
	err = pm.SaveCustomTheme(testTheme)
	if err != nil {
		t.Errorf("Failed to save custom theme: %v", err)
	}

	// Verify theme file was created
	filename := "test-custom-theme.json"
	themePath := filepath.Join(pm.themesDir, filename)
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		t.Error("Theme file was not created")
	}

	// Load theme
	loadedTheme, exists := pm.GetCustomTheme("test-custom-theme")
	if !exists {
		t.Error("Custom theme was not found")
	}

	if loadedTheme.Name != testTheme.Name {
		t.Errorf("Expected theme name %s, got %s", testTheme.Name, loadedTheme.Name)
	}

	if loadedTheme.Palette.Background.Hex != testTheme.Palette.Background.Hex {
		t.Error("Theme palette was not preserved")
	}

	// Test loading after restart
	pm2, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create second persistence manager: %v", err)
	}

	themes := pm2.GetCustomThemes()
	if len(themes) != 1 {
		t.Errorf("Expected 1 custom theme, got %d", len(themes))
	}

	if themes["test-custom-theme"] == nil {
		t.Error("Custom theme was not loaded on restart")
	}
}

func TestPersistenceManager_DeleteCustomTheme(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Create and save test theme
	testTheme := &Theme{
		Name:        "deletable-theme",
		Description: "A theme to delete",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	err = pm.SaveCustomTheme(testTheme)
	if err != nil {
		t.Errorf("Failed to save theme: %v", err)
	}

	// Verify theme exists
	_, exists := pm.GetCustomTheme("deletable-theme")
	if !exists {
		t.Error("Theme should exist before deletion")
	}

	// Delete theme
	err = pm.DeleteCustomTheme("deletable-theme")
	if err != nil {
		t.Errorf("Failed to delete theme: %v", err)
	}

	// Verify theme was deleted
	_, exists = pm.GetCustomTheme("deletable-theme")
	if exists {
		t.Error("Theme should not exist after deletion")
	}

	// Verify file was deleted
	filename := "deletable-theme.json"
	themePath := filepath.Join(pm.themesDir, filename)
	if _, err := os.Stat(themePath); !os.IsNotExist(err) {
		t.Error("Theme file should be deleted")
	}

	// Test deleting nonexistent theme
	err = pm.DeleteCustomTheme("nonexistent-theme")
	if err == nil {
		t.Error("Expected error when deleting nonexistent theme")
	}
}

func TestPersistenceManager_ExportImportTheme(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Create test theme
	testTheme := &Theme{
		Name:        "exportable-theme",
		Description: "A theme for export/import testing",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Author:      "Test Author",
		Palette: ColorPalette{
			Background:  Color{Hex: "#2e3440", Name: "Background"},
			Primary:     Color{Hex: "#5e81ac", Name: "Primary"},
			TextPrimary: Color{Hex: "#d8dee9", Name: "Text Primary"},
		},
	}

	err = pm.SaveCustomTheme(testTheme)
	if err != nil {
		t.Errorf("Failed to save theme: %v", err)
	}

	// Export theme
	exportPath := filepath.Join(tempDir, "exported-theme.json")
	err = pm.ExportTheme("exportable-theme", exportPath)
	if err != nil {
		t.Errorf("Failed to export theme: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Delete original theme
	err = pm.DeleteCustomTheme("exportable-theme")
	if err != nil {
		t.Errorf("Failed to delete original theme: %v", err)
	}

	// Import theme
	importedTheme, err := pm.ImportTheme(exportPath)
	if err != nil {
		t.Errorf("Failed to import theme: %v", err)
	}

	if importedTheme.Name != testTheme.Name {
		t.Errorf("Expected imported theme name %s, got %s", testTheme.Name, importedTheme.Name)
	}

	// Verify theme is now available
	_, exists := pm.GetCustomTheme("exportable-theme")
	if !exists {
		t.Error("Imported theme should be available")
	}

	// Test importing duplicate theme
	_, err = pm.ImportTheme(exportPath)
	if err == nil {
		t.Error("Expected error when importing duplicate theme")
	}
}

func TestPersistenceManager_BackupRestore(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Create test themes
	theme1 := &Theme{
		Name:        "backup-theme-1",
		Description: "First backup theme",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	theme2 := &Theme{
		Name:        "backup-theme-2",
		Description: "Second backup theme",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#ffffff"},
			Primary:     Color{Hex: "#000000"},
			TextPrimary: Color{Hex: "#000000"},
		},
	}

	err = pm.SaveCustomTheme(theme1)
	if err != nil {
		t.Errorf("Failed to save theme1: %v", err)
	}

	err = pm.SaveCustomTheme(theme2)
	if err != nil {
		t.Errorf("Failed to save theme2: %v", err)
	}

	// Modify preferences
	prefs := pm.GetPreferences()
	prefs.ActiveTheme = "backup-theme-1"
	prefs.AutoDetectTerminal = false
	err = pm.UpdatePreferences(prefs)
	if err != nil {
		t.Errorf("Failed to update preferences: %v", err)
	}

	// Create backup
	backupPath := filepath.Join(tempDir, "backup.json")
	err = pm.BackupThemes(backupPath)
	if err != nil {
		t.Errorf("Failed to create backup: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Delete themes and reset preferences
	err = pm.DeleteCustomTheme("backup-theme-1")
	if err != nil {
		t.Errorf("Failed to delete theme1: %v", err)
	}

	err = pm.DeleteCustomTheme("backup-theme-2")
	if err != nil {
		t.Errorf("Failed to delete theme2: %v", err)
	}

	defaultPrefs := pm.createDefaultPreferences()
	err = pm.UpdatePreferences(defaultPrefs)
	if err != nil {
		t.Errorf("Failed to reset preferences: %v", err)
	}

	// Verify themes are gone
	themes := pm.GetCustomThemes()
	if len(themes) != 0 {
		t.Errorf("Expected 0 themes, got %d", len(themes))
	}

	// Restore from backup
	err = pm.RestoreThemes(backupPath, true)
	if err != nil {
		t.Errorf("Failed to restore backup: %v", err)
	}

	// Verify themes were restored
	themes = pm.GetCustomThemes()
	if len(themes) != 2 {
		t.Errorf("Expected 2 restored themes, got %d", len(themes))
	}

	if themes["backup-theme-1"] == nil {
		t.Error("backup-theme-1 was not restored")
	}

	if themes["backup-theme-2"] == nil {
		t.Error("backup-theme-2 was not restored")
	}

	// Verify preferences were restored
	restoredPrefs := pm.GetPreferences()
	if restoredPrefs.ActiveTheme != "backup-theme-1" {
		t.Errorf("Expected active theme backup-theme-1, got %s", restoredPrefs.ActiveTheme)
	}

	if restoredPrefs.AutoDetectTerminal {
		t.Error("AutoDetectTerminal should be false after restore")
	}
}

func TestValidateTheme(t *testing.T) {
	// Test valid theme
	validTheme := &Theme{
		Name:    "valid-theme",
		Version: "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	err := validateTheme(validTheme)
	if err != nil {
		t.Errorf("Valid theme should pass validation: %v", err)
	}

	// Test theme with empty name
	invalidTheme := &Theme{
		Name:    "",
		Version: "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	err = validateTheme(invalidTheme)
	if err == nil {
		t.Error("Theme with empty name should fail validation")
	}

	// Test theme with invalid color
	invalidColorTheme := &Theme{
		Name:    "invalid-color-theme",
		Version: "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "invalid-hex"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	err = validateTheme(invalidColorTheme)
	if err == nil {
		t.Error("Theme with invalid color should fail validation")
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		hex     string
		isValid bool
	}{
		{"#000000", true},
		{"#ffffff", true},
		{"#ff0000", true},
		{"#ABCDEF", true},
		{"#123456", true},
		{"000000", false},   // Missing #
		{"#00000", false},   // Too short
		{"#0000000", false}, // Too long
		{"#gggggg", false},  // Invalid characters
		{"", false},         // Empty
		{"#", false},        // Just #
	}

	for _, test := range tests {
		result := isValidHexColor(test.hex)
		if result != test.isValid {
			t.Errorf("For hex %s, expected %t, got %t", test.hex, test.isValid, result)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple-name", "simple-name"},
		{"name with spaces", "name_with_spaces"},
		{"name/with\\slashes", "name_with_slashes"},
		{"name:with*special?chars", "name_with_special_chars"},
		{"name\"with<quotes>and|pipes", "name_with_quotes_and_pipes"},
	}

	for _, test := range tests {
		result := sanitizeFilename(test.input)
		if result != test.expected {
			t.Errorf("For input %s, expected %s, got %s", test.input, test.expected, result)
		}
	}

	// Test length limit
	longName := make([]byte, 150)
	for i := range longName {
		longName[i] = 'a'
	}
	result := sanitizeFilename(string(longName))
	if len(result) > 100 {
		t.Errorf("Sanitized filename should be limited to 100 characters, got %d", len(result))
	}
}

func TestPersistenceManager_InvalidOperations(t *testing.T) {
	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Test saving nil theme
	err = pm.SaveCustomTheme(nil)
	if err == nil {
		t.Error("Expected error when saving nil theme")
	}

	// Test exporting nonexistent theme
	err = pm.ExportTheme("nonexistent", "/tmp/test.json")
	if err == nil {
		t.Error("Expected error when exporting nonexistent theme")
	}

	// Test importing invalid JSON file
	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidJSONPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	_, err = pm.ImportTheme(invalidJSONPath)
	if err == nil {
		t.Error("Expected error when importing invalid JSON")
	}

	// Test restoring from nonexistent backup
	err = pm.RestoreThemes("/nonexistent/backup.json", true)
	if err == nil {
		t.Error("Expected error when restoring from nonexistent file")
	}
}

func TestPersistenceManager_FilePermissions(t *testing.T) {
	// Skip on Windows as file permissions work differently
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping file permission test on Windows")
	}

	tempDir := t.TempDir()
	config := &PersistenceConfig{ConfigDir: tempDir}

	pm, err := NewPersistenceManager(config)
	if err != nil {
		t.Fatalf("Failed to create persistence manager: %v", err)
	}

	// Create theme and verify file permissions
	testTheme := &Theme{
		Name:    "permission-test",
		Version: "1.0.0",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000"},
			Primary:     Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#ffffff"},
		},
	}

	err = pm.SaveCustomTheme(testTheme)
	if err != nil {
		t.Errorf("Failed to save theme: %v", err)
	}

	// Check file permissions
	themePath := filepath.Join(pm.themesDir, "permission-test.json")
	info, err := os.Stat(themePath)
	if err != nil {
		t.Errorf("Failed to stat theme file: %v", err)
	}

	expectedMode := os.FileMode(0644)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected file mode %v, got %v", expectedMode, info.Mode().Perm())
	}

	// Check preferences file permissions
	prefsInfo, err := os.Stat(pm.prefsFile)
	if err != nil {
		t.Errorf("Failed to stat preferences file: %v", err)
	}

	if prefsInfo.Mode().Perm() != expectedMode {
		t.Errorf("Expected preferences file mode %v, got %v", expectedMode, prefsInfo.Mode().Perm())
	}
}