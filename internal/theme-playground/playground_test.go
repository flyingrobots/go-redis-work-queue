// Copyright 2025 James Ross
package themeplayground

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPlaygroundModel(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	if model == nil {
		t.Fatal("NewPlaygroundModel returned nil")
	}

	if model.themeManager != tm {
		t.Error("ThemeManager not set correctly")
	}

	if model.selectedTheme != tm.GetActiveTheme().Name {
		t.Error("Selected theme should match active theme initially")
	}

	if model.previewMode {
		t.Error("Preview mode should be false initially")
	}

	if model.table.Focused() != true {
		t.Error("Table should be focused by default")
	}
}

func TestPlaygroundModel_Init(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestPlaygroundModel_UpdateWindowSize(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(*PlaygroundModel)
	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}

	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}

	if !m.ready {
		t.Error("Model should be ready after window size update")
	}
}

func TestPlaygroundModel_KeyHandling(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	// Test help toggle
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel2, _ := model.Update(keyMsg)

	m := updatedModel2.(*PlaygroundModel)
	if !m.showHelp {
		t.Error("Help should be shown after pressing '?'")
	}

	// Toggle help again
	updatedModel3, _ := m.Update(keyMsg)
	m = updatedModel3.(*PlaygroundModel)
	if m.showHelp {
		t.Error("Help should be hidden after pressing '?' again")
	}
}

func TestPlaygroundModel_View(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Test view when not ready
	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Error("View should show loading message when not ready")
	}

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	// Test view when ready
	view = model.View()
	if strings.Contains(view, "Loading") {
		t.Error("View should not show loading message when ready")
	}

	if !strings.Contains(view, "Theme Playground") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "Active:") {
		t.Error("View should show active theme info")
	}
}

func TestPlaygroundModel_PreviewMode(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	m := model

	// Enable preview mode manually for testing
	m.previewMode = true
	m.selectedTheme = ThemeTokyoNight

	view := m.View()
	if !strings.Contains(view, "Preview Mode") {
		t.Error("View should show preview mode indicator")
	}

	if !strings.Contains(view, "Press 'a' to apply") {
		t.Error("View should show apply instruction in preview mode")
	}
}

func TestPlaygroundModel_ThemePreview(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	m := model

	// Test theme preview rendering
	preview := m.renderThemePreview(ThemeTokyoNight)
	if preview == "" {
		t.Error("Theme preview should not be empty")
	}

	if !strings.Contains(preview, "Preview:") {
		t.Error("Preview should contain preview header")
	}

	if !strings.Contains(preview, "WCAG:") {
		t.Error("Preview should contain accessibility info")
	}

	// Test preview for nonexistent theme
	preview = m.renderThemePreview("nonexistent-theme")
	if preview != "" {
		t.Error("Preview for nonexistent theme should be empty")
	}
}

func TestThemeSelector(t *testing.T) {
	tm := NewThemeManager(t.TempDir())

	// Test compact selector
	compactSelector := ThemeSelector(tm, true)
	if compactSelector == "" {
		t.Error("Compact selector should not be empty")
	}

	if !strings.Contains(compactSelector, "*") {
		t.Error("Compact selector should show active theme indicator")
	}

	// Test full selector
	fullSelector := ThemeSelector(tm, false)
	if fullSelector == "" {
		t.Error("Full selector should not be empty")
	}

	if !strings.Contains(fullSelector, "(") {
		t.Error("Full selector should show theme descriptions")
	}

	// Verify active theme is marked
	activeTheme := tm.GetActiveTheme().Name
	if !strings.Contains(fullSelector, activeTheme) {
		t.Error("Selector should contain active theme name")
	}
}

func TestQuickThemeToggle(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	toggle := NewQuickThemeToggle(tm)

	if toggle == nil {
		t.Fatal("NewQuickThemeToggle returned nil")
	}

	if toggle.themeManager != tm {
		t.Error("ThemeManager not set correctly")
	}

	if len(toggle.keyMap) == 0 {
		t.Error("Key map should not be empty")
	}

	// Test handling valid key
	originalTheme := tm.GetActiveTheme().Name
	handled := toggle.HandleKey("2") // Should map to Tokyo Night
	if !handled {
		t.Error("Valid key should be handled")
	}

	newTheme := tm.GetActiveTheme().Name
	if newTheme == originalTheme {
		t.Error("Theme should have changed")
	}

	// Test handling invalid key
	handled = toggle.HandleKey("x")
	if handled {
		t.Error("Invalid key should not be handled")
	}
}

func TestQuickThemeToggle_GetKeyHelp(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	toggle := NewQuickThemeToggle(tm)

	help := toggle.GetKeyHelp()
	if help == "" {
		t.Error("Key help should not be empty")
	}

	if !strings.Contains(help, ":") {
		t.Error("Help should contain key mappings")
	}

	if !strings.Contains(help, "|") {
		t.Error("Help should contain separators")
	}
}

func TestDefaultKeyMap(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test that all required bindings exist
	if len(keyMap.Up.Keys()) == 0 {
		t.Error("Up key binding should not be empty")
	}

	if len(keyMap.Down.Keys()) == 0 {
		t.Error("Down key binding should not be empty")
	}

	if len(keyMap.Enter.Keys()) == 0 {
		t.Error("Enter key binding should not be empty")
	}

	if len(keyMap.Quit.Keys()) == 0 {
		t.Error("Quit key binding should not be empty")
	}

	// Test help methods
	shortHelp := keyMap.ShortHelp()
	if len(shortHelp) == 0 {
		t.Error("Short help should not be empty")
	}

	fullHelp := keyMap.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("Full help should not be empty")
	}

	// Verify help structure
	for _, group := range fullHelp {
		if len(group) == 0 {
			t.Error("Help group should not be empty")
		}
	}
}

func TestPlaygroundModel_UpdateTable(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Get initial row count
	initialRows := len(model.table.Rows())

	// Register a new theme
	testTheme := &Theme{
		Name:        "test-update-theme",
		Description: "Test theme for update",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Palette: ColorPalette{
			Background:       Color{Hex: "#000000", Name: "Black"},
			Surface:          Color{Hex: "#111111", Name: "Dark Gray"},
			Primary:          Color{Hex: "#ffffff", Name: "White"},
			Secondary:        Color{Hex: "#cccccc", Name: "Gray"},
			Accent:           Color{Hex: "#0066cc", Name: "Blue"},
			Success:          Color{Hex: "#00ff00", Name: "Green"},
			Warning:          Color{Hex: "#ffff00", Name: "Yellow"},
			Error:            Color{Hex: "#ff0000", Name: "Red"},
			Info:             Color{Hex: "#00ccff", Name: "Light Blue"},
			TextPrimary:      Color{Hex: "#ffffff", Name: "White"},
			TextSecondary:    Color{Hex: "#cccccc", Name: "Gray"},
			TextDisabled:     Color{Hex: "#666666", Name: "Dark Gray"},
			TextInverse:      Color{Hex: "#000000", Name: "Black"},
			Border:           Color{Hex: "#333333", Name: "Border"},
			Divider:          Color{Hex: "#222222", Name: "Divider"},
			Focus:            Color{Hex: "#0066cc", Name: "Focus"},
			Selected:         Color{Hex: "#004499", Name: "Selected"},
			Hover:            Color{Hex: "#003366", Name: "Hover"},
			StatusPending:    Color{Hex: "#ffaa00", Name: "Pending"},
			StatusRunning:    Color{Hex: "#0066cc", Name: "Running"},
			StatusCompleted:  Color{Hex: "#00aa00", Name: "Completed"},
			StatusFailed:     Color{Hex: "#cc0000", Name: "Failed"},
			StatusRetrying:   Color{Hex: "#ff6600", Name: "Retrying"},
		},
	}

	err := tm.RegisterTheme(testTheme)
	if err != nil {
		t.Errorf("Failed to register test theme: %v", err)
	}

	// Update table
	model.updateTable()

	// Verify row count increased
	newRows := len(model.table.Rows())
	if newRows != initialRows+1 {
		t.Errorf("Expected %d rows, got %d", initialRows+1, newRows)
	}

	// Verify new theme appears in table
	found := false
	for _, row := range model.table.Rows() {
		if len(row) > 0 && row[0] == "test-update-theme" {
			found = true
			break
		}
	}

	if !found {
		t.Error("New theme should appear in table")
	}
}

func TestPlaygroundModel_ErrorHandling(t *testing.T) {
	tm := NewThemeManager(t.TempDir())
	model := NewPlaygroundModel(tm)

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	m := model

	// Simulate error by trying to set invalid theme
	m.selectedTheme = "nonexistent-theme"

	// Try to apply non-existent theme (this should set an error)
	originalTheme := m.themeManager.GetActiveTheme().Name

	// The error would be set in a real key handling scenario
	m.err = ErrThemeNotFound.WithDetails("nonexistent-theme")

	view := m.View()
	if !strings.Contains(view, "Error:") {
		t.Error("View should show error message")
	}

	// Verify theme didn't change
	if m.themeManager.GetActiveTheme().Name != originalTheme {
		t.Error("Theme should not have changed when error occurred")
	}
}

func TestPreviewThemeMsg(t *testing.T) {
	msg := previewThemeMsg{
		ThemeName:     "test-theme",
		OriginalTheme: "original-theme",
	}

	if msg.ThemeName != "test-theme" {
		t.Error("ThemeName not set correctly")
	}

	if msg.OriginalTheme != "original-theme" {
		t.Error("OriginalTheme not set correctly")
	}
}

func BenchmarkPlaygroundModel_View(b *testing.B) {
	tm := NewThemeManager(b.TempDir())
	model := NewPlaygroundModel(tm)

	// Set window size to make model ready
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(*PlaygroundModel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkThemeSelector(b *testing.B) {
	tm := NewThemeManager(b.TempDir())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ThemeSelector(tm, false)
	}
}

func BenchmarkQuickThemeToggle_HandleKey(b *testing.B) {
	tm := NewThemeManager(b.TempDir())
	toggle := NewQuickThemeToggle(tm)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := string(rune('1' + (i % 9)))
		_ = toggle.HandleKey(key)
	}
}