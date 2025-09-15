// Copyright 2025 James Ross
package themeplayground

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestThemeManager_NewThemeManager(t *testing.T) {
	tm := NewThemeManager()

	if tm == nil {
		t.Fatal("NewThemeManager returned nil")
	}

	if tm.activeTheme == nil {
		t.Error("NewThemeManager should set default active theme")
	}

	if tm.activeTheme.Name != ThemeDefault {
		t.Errorf("Expected default theme %s, got %s", ThemeDefault, tm.activeTheme.Name)
	}

	if len(tm.registry) == 0 {
		t.Error("NewThemeManager should register built-in themes")
	}
}

func TestThemeManager_GetActiveTheme(t *testing.T) {
	tm := NewThemeManager()

	activeTheme := tm.GetActiveTheme()
	if activeTheme == nil {
		t.Fatal("GetActiveTheme returned nil")
	}

	if activeTheme.Name != ThemeDefault {
		t.Errorf("Expected default theme %s, got %s", ThemeDefault, activeTheme.Name)
	}
}

func TestThemeManager_SetActiveTheme(t *testing.T) {
	tm := NewThemeManager()

	// Test setting valid theme
	err := tm.SetActiveTheme(ThemeTokyoNight)
	if err != nil {
		t.Errorf("Unexpected error setting theme: %v", err)
	}

	if tm.GetActiveTheme().Name != ThemeTokyoNight {
		t.Errorf("Expected theme %s, got %s", ThemeTokyoNight, tm.GetActiveTheme().Name)
	}

	// Test setting invalid theme
	err = tm.SetActiveTheme("nonexistent-theme")
	if err == nil {
		t.Error("Expected error for nonexistent theme")
	}

	// Verify theme didn't change
	if tm.GetActiveTheme().Name != ThemeTokyoNight {
		t.Error("Theme should not have changed after error")
	}
}

func TestThemeManager_RegisterTheme(t *testing.T) {
	tm := NewThemeManager()

	// Create test theme
	testTheme := &Theme{
		Name:        "test-theme",
		Description: "Test theme",
		Category:    CategoryCustom,
		Version:     "1.0.0",
		Author:      "Test Author",
		Palette: ColorPalette{
			Background:  Color{Hex: "#000000", Name: "Black"},
			Primary:     Color{Hex: "#ffffff", Name: "White"},
			TextPrimary: Color{Hex: "#ffffff", Name: "White"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := tm.RegisterTheme(testTheme)
	if err != nil {
		t.Errorf("Unexpected error registering theme: %v", err)
	}

	// Verify theme was registered
	theme, exists := tm.GetTheme("test-theme")
	if !exists {
		t.Error("Theme was not registered")
	}

	if theme.Name != "test-theme" {
		t.Errorf("Expected theme name 'test-theme', got %s", theme.Name)
	}

	// Test registering duplicate theme
	err = tm.RegisterTheme(testTheme)
	if err == nil {
		t.Error("Expected error for duplicate theme registration")
	}
}

func TestThemeManager_GetTheme(t *testing.T) {
	tm := NewThemeManager()

	// Test getting existing theme
	theme, exists := tm.GetTheme(ThemeDefault)
	if !exists {
		t.Error("Default theme should exist")
	}

	if theme.Name != ThemeDefault {
		t.Errorf("Expected theme %s, got %s", ThemeDefault, theme.Name)
	}

	// Test getting nonexistent theme
	_, exists = tm.GetTheme("nonexistent")
	if exists {
		t.Error("Nonexistent theme should not exist")
	}
}

func TestThemeManager_ListThemes(t *testing.T) {
	tm := NewThemeManager()

	themes := tm.ListThemes()
	if len(themes) == 0 {
		t.Error("ListThemes should return built-in themes")
	}

	// Verify default theme is in list
	found := false
	for _, theme := range themes {
		if theme.Name == ThemeDefault {
			found = true
			break
		}
	}

	if !found {
		t.Error("Default theme should be in themes list")
	}
}

func TestThemeManager_GetStyleFor(t *testing.T) {
	tm := NewThemeManager()

	style := tm.GetStyleFor("text", "primary")
	if style == (lipgloss.Style{}) {
		t.Error("GetStyleFor should return non-empty style")
	}

	// Test that style has the expected color
	expectedColor := tm.GetActiveTheme().Palette.TextPrimary.Hex
	if style.GetForeground() != lipgloss.Color(expectedColor) {
		t.Errorf("Expected color %s, got %s", expectedColor, style.GetForeground())
	}
}

func TestThemeManager_OnThemeChanged(t *testing.T) {
	tm := NewThemeManager()

	callbackCalled := false
	var callbackTheme string

	tm.OnThemeChanged(func(themeName string) {
		callbackCalled = true
		callbackTheme = themeName
	})

	err := tm.SetActiveTheme(ThemeTokyoNight)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !callbackCalled {
		t.Error("Theme change callback was not called")
	}

	if callbackTheme != ThemeTokyoNight {
		t.Errorf("Expected callback theme %s, got %s", ThemeTokyoNight, callbackTheme)
	}
}

func TestColorUtilities_HexToRGB(t *testing.T) {
	cu := NewColorUtilities()

	tests := []struct {
		hex      string
		expected RGB
		hasError bool
	}{
		{"#ff0000", RGB{R: 255, G: 0, B: 0}, false},
		{"#00ff00", RGB{R: 0, G: 255, B: 0}, false},
		{"#0000ff", RGB{R: 0, G: 0, B: 255}, false},
		{"#ffffff", RGB{R: 255, G: 255, B: 255}, false},
		{"#000000", RGB{R: 0, G: 0, B: 0}, false},
		{"invalid", RGB{}, true},
		{"#gg0000", RGB{}, true},
	}

	for _, test := range tests {
		rgb, err := cu.HexToRGB(test.hex)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for hex %s", test.hex)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for hex %s: %v", test.hex, err)
			}
			if rgb != test.expected {
				t.Errorf("For hex %s, expected %+v, got %+v", test.hex, test.expected, rgb)
			}
		}
	}
}

func TestColorUtilities_RGBToHex(t *testing.T) {
	cu := NewColorUtilities()

	tests := []struct {
		rgb      RGB
		expected string
	}{
		{RGB{R: 255, G: 0, B: 0}, "#ff0000"},
		{RGB{R: 0, G: 255, B: 0}, "#00ff00"},
		{RGB{R: 0, G: 0, B: 255}, "#0000ff"},
		{RGB{R: 255, G: 255, B: 255}, "#ffffff"},
		{RGB{R: 0, G: 0, B: 0}, "#000000"},
	}

	for _, test := range tests {
		hex := cu.RGBToHex(test.rgb)
		if hex != test.expected {
			t.Errorf("For RGB %+v, expected %s, got %s", test.rgb, test.expected, hex)
		}
	}
}

func TestColorUtilities_CalculateContrastRatio(t *testing.T) {
	cu := NewColorUtilities()

	// Test black on white (maximum contrast)
	ratio := cu.CalculateContrastRatio("#000000", "#ffffff")
	if ratio < 20.9 || ratio > 21.1 { // Allow small floating point variance
		t.Errorf("Expected contrast ratio ~21, got %f", ratio)
	}

	// Test white on white (minimum contrast)
	ratio = cu.CalculateContrastRatio("#ffffff", "#ffffff")
	if ratio < 0.9 || ratio > 1.1 {
		t.Errorf("Expected contrast ratio ~1, got %f", ratio)
	}

	// Test that ratio is symmetric
	ratio1 := cu.CalculateContrastRatio("#ff0000", "#00ff00")
	ratio2 := cu.CalculateContrastRatio("#00ff00", "#ff0000")
	if ratio1 != ratio2 {
		t.Errorf("Contrast ratio should be symmetric: %f vs %f", ratio1, ratio2)
	}
}

func TestAccessibilityChecker_CheckContrast(t *testing.T) {
	ac := NewAccessibilityChecker()

	// Test high contrast (should pass AA and AAA)
	check := ac.CheckContrast("#000000", "#ffffff", "test-component")
	if !check.AACompliant {
		t.Error("Black on white should be AA compliant")
	}
	if !check.AAACompliant {
		t.Error("Black on white should be AAA compliant")
	}

	// Test low contrast (should fail both)
	check = ac.CheckContrast("#888888", "#999999", "test-component")
	if check.AACompliant {
		t.Error("Low contrast should not be AA compliant")
	}
	if check.AAACompliant {
		t.Error("Low contrast should not be AAA compliant")
	}
}

func TestAccessibilityChecker_ValidateTheme(t *testing.T) {
	ac := NewAccessibilityChecker()

	// Create test theme with good contrast
	theme := &Theme{
		Name: "high-contrast-test",
		Palette: ColorPalette{
			Background:  Color{Hex: "#ffffff"},
			TextPrimary: Color{Hex: "#000000"},
			Primary:     Color{Hex: "#0066cc"},
			Secondary:   Color{Hex: "#666666"},
		},
	}

	info := ac.ValidateTheme(theme)
	if info.ContrastRatio < 10 {
		t.Errorf("Expected high contrast ratio, got %f", info.ContrastRatio)
	}

	if info.WCAGLevel != "AAA" {
		t.Errorf("Expected AAA compliance, got %s", info.WCAGLevel)
	}

	// Create test theme with poor contrast
	poorTheme := &Theme{
		Name: "poor-contrast-test",
		Palette: ColorPalette{
			Background:  Color{Hex: "#888888"},
			TextPrimary: Color{Hex: "#999999"},
			Primary:     Color{Hex: "#aaaaaa"},
			Secondary:   Color{Hex: "#bbbbbb"},
		},
	}

	info = ac.ValidateTheme(poorTheme)
	if len(info.Warnings) == 0 {
		t.Error("Poor contrast theme should have warnings")
	}
}

func TestThemeError(t *testing.T) {
	err := NewThemeError("TEST_CODE", "test message")
	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code TEST_CODE, got %s", err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", err.Message)
	}

	if err.Error() != "test message" {
		t.Errorf("Expected error string 'test message', got %s", err.Error())
	}

	// Test with details
	errWithDetails := err.WithDetails("additional details")
	expected := "test message: additional details"
	if errWithDetails.Error() != expected {
		t.Errorf("Expected error string '%s', got %s", expected, errWithDetails.Error())
	}
}

func TestBuiltInThemes(t *testing.T) {
	builtInThemes := []string{
		ThemeDefault,
		ThemeTokyoNight,
		ThemeGitHub,
		ThemeOneDark,
		ThemeSolarizedLight,
		ThemeSolarizedDark,
		ThemeDracula,
		ThemeHighContrast,
		ThemeMonochrome,
		ThemeTerminalClassic,
	}

	for _, themeName := range builtInThemes {
		t.Run(themeName, func(t *testing.T) {
			tm := NewThemeManager()
			theme, exists := tm.GetTheme(themeName)
			if !exists {
				t.Errorf("Built-in theme %s should exist", themeName)
				return
			}

			if theme.Name != themeName {
				t.Errorf("Expected theme name %s, got %s", themeName, theme.Name)
			}

			if theme.Palette.Background.Hex == "" {
				t.Error("Theme should have background color")
			}

			if theme.Palette.TextPrimary.Hex == "" {
				t.Error("Theme should have primary text color")
			}

			// Test that theme can be set as active
			err := tm.SetActiveTheme(themeName)
			if err != nil {
				t.Errorf("Failed to set theme %s as active: %v", themeName, err)
			}
		})
	}
}

func TestThemeManager_Concurrent(t *testing.T) {
	tm := NewThemeManager()

	// Test concurrent read operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_ = tm.GetActiveTheme()
				_ = tm.ListThemes()
				_, _ = tm.GetTheme(ThemeDefault)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent write operations
	themes := []string{ThemeDefault, ThemeTokyoNight, ThemeGitHub, ThemeOneDark}
	for i := 0; i < len(themes); i++ {
		go func(themeName string) {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				_ = tm.SetActiveTheme(themeName)
			}
		}(themes[i])
	}

	// Wait for all write operations to complete
	for i := 0; i < len(themes); i++ {
		<-done
	}

	// Verify manager is still in valid state
	activeTheme := tm.GetActiveTheme()
	if activeTheme == nil {
		t.Error("Active theme should not be nil after concurrent operations")
	}
}

func BenchmarkThemeManager_GetActiveTheme(b *testing.B) {
	tm := NewThemeManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tm.GetActiveTheme()
	}
}

func BenchmarkThemeManager_SetActiveTheme(b *testing.B) {
	tm := NewThemeManager()
	themes := []string{ThemeDefault, ThemeTokyoNight, ThemeGitHub, ThemeOneDark}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		theme := themes[i%len(themes)]
		_ = tm.SetActiveTheme(theme)
	}
}

func BenchmarkColorUtilities_CalculateContrastRatio(b *testing.B) {
	cu := NewColorUtilities()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cu.CalculateContrastRatio("#000000", "#ffffff")
	}
}