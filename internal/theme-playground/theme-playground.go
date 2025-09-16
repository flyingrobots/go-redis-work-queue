// Copyright 2025 James Ross
package themeplayground

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ThemeManager implements the core theme management functionality
type ThemeManager struct {
	mu            sync.RWMutex
	registry      map[string]*Theme
	activeTheme   *Theme
	preferences   *ThemePreferences
	configDir     string
	colorUtils    *ColorUtilities
	accessibility *AccessibilityChecker
	callbacks     []func(*Theme)
}

// NewThemeManager creates a new theme manager instance
func NewThemeManager(configDir string) *ThemeManager {
	tm := &ThemeManager{
		registry:      make(map[string]*Theme),
		configDir:     configDir,
		colorUtils:    NewColorUtilities(),
		accessibility: NewAccessibilityChecker(),
		callbacks:     make([]func(*Theme), 0),
	}

	// Load built-in themes
	tm.loadBuiltinThemes()

	// Load user preferences
	tm.loadPreferences()

	// Load custom themes
	tm.loadCustomThemes()

	// Set default theme if none active
	if tm.activeTheme == nil {
		tm.SetActiveTheme(ThemeDefault)
	}

	return tm
}

// GetActiveTheme returns the currently active theme
func (tm *ThemeManager) GetActiveTheme() *Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.activeTheme
}

// SetActiveTheme sets the active theme
func (tm *ThemeManager) SetActiveTheme(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	theme, exists := tm.registry[name]
	if !exists {
		return ErrThemeNotFound.WithDetails(name)
	}

	tm.activeTheme = theme
	tm.preferences.ActiveTheme = name
	tm.preferences.UpdatedAt = time.Now()

	// Save preferences
	tm.savePreferences()

	// Notify callbacks
	for _, callback := range tm.callbacks {
		callback(theme)
	}

	return nil
}

// GetTheme retrieves a theme by name
func (tm *ThemeManager) GetTheme(name string) (*Theme, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, exists := tm.registry[name]
	if !exists {
		return nil, ErrThemeNotFound.WithDetails(name)
	}

	return theme, nil
}

// ListThemes returns all available themes
func (tm *ThemeManager) ListThemes() []Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]Theme, 0, len(tm.registry))
	for _, theme := range tm.registry {
		themes = append(themes, *theme)
	}

	// Sort by name
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name < themes[j].Name
	})

	return themes
}

// GetThemesByCategory returns themes filtered by category
func (tm *ThemeManager) GetThemesByCategory(category ThemeCategory) []Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]Theme, 0)
	for _, theme := range tm.registry {
		if theme.Category == category {
			themes = append(themes, *theme)
		}
	}

	return themes
}

// RegisterTheme adds a new theme to the registry
func (tm *ThemeManager) RegisterTheme(theme *Theme) error {
	if theme == nil {
		return ErrThemeInvalid.WithDetails("theme is nil")
	}

	if theme.Name == "" {
		return ErrThemeInvalid.WithDetails("theme name is required")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.registry[theme.Name]; exists {
		return ErrThemeExists.WithDetails(theme.Name)
	}

	// Validate theme
	if err := tm.validateTheme(theme); err != nil {
		return err
	}

	theme.CreatedAt = time.Now()
	theme.UpdatedAt = time.Now()

	tm.registry[theme.Name] = theme

	return nil
}

// ValidateTheme validates a theme for correctness and accessibility
func (tm *ThemeManager) ValidateTheme(theme *Theme) error {
	return tm.validateTheme(theme)
}

// validateTheme performs internal theme validation
func (tm *ThemeManager) validateTheme(theme *Theme) error {
	if theme.Name == "" {
		return ErrThemeInvalid.WithDetails("name is required")
	}

	if theme.Description == "" {
		return ErrThemeInvalid.WithDetails("description is required")
	}

	// Validate color palette
	if err := tm.validateColorPalette(&theme.Palette); err != nil {
		return err
	}

	// Check accessibility if enabled
	if tm.preferences != nil && tm.preferences.AccessibilityMode {
		info, err := tm.accessibility.CheckAccessibility(theme)
		if err != nil {
			return err
		}

		theme.Accessibility = *info

		if !info.ColorBlindSafe && len(info.Warnings) > 0 {
			return ErrAccessibilityFail.WithDetails("theme fails accessibility checks")
		}
	}

	return nil
}

// validateColorPalette validates all colors in a palette
func (tm *ThemeManager) validateColorPalette(palette *ColorPalette) error {
	colors := []*Color{
		&palette.Background, &palette.Surface, &palette.Primary, &palette.Secondary, &palette.Accent,
		&palette.Success, &palette.Warning, &palette.Error, &palette.Info,
		&palette.TextPrimary, &palette.TextSecondary, &palette.TextDisabled, &palette.TextInverse,
		&palette.Border, &palette.Divider, &palette.Focus, &palette.Selected, &palette.Hover,
		&palette.StatusPending, &palette.StatusRunning, &palette.StatusCompleted,
		&palette.StatusFailed, &palette.StatusRetrying,
	}

	for _, color := range colors {
		if err := tm.validateColor(color); err != nil {
			return err
		}
	}

	return nil
}

// validateColor validates a single color
func (tm *ThemeManager) validateColor(color *Color) error {
	if color.Hex == "" {
		return ErrColorInvalid.WithDetails("hex value is required")
	}

	// Validate hex format
	hexRegex := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	if !hexRegex.MatchString(color.Hex) {
		return ErrColorInvalid.WithDetails(fmt.Sprintf("invalid hex format: %s", color.Hex))
	}

	// Populate RGB and HSL if missing
	if color.RGB.R == 0 && color.RGB.G == 0 && color.RGB.B == 0 {
		rgb, err := tm.colorUtils.HexToRGB(color.Hex)
		if err != nil {
			return err
		}
		color.RGB = *rgb
	}

	if color.HSL.H == 0 && color.HSL.S == 0 && color.HSL.L == 0 {
		hsl, err := tm.colorUtils.RGBToHSL(color.RGB)
		if err != nil {
			return err
		}
		color.HSL = *hsl
	}

	return nil
}

// GetStyleFor returns a Lip Gloss style for a component and variant
func (tm *ThemeManager) GetStyleFor(component, variant string) lipgloss.Style {
	theme := tm.GetActiveTheme()
	if theme == nil {
		return lipgloss.NewStyle()
	}

	switch component {
	case "button":
		return tm.getButtonStyle(theme, variant)
	case "table":
		return tm.getTableStyle(theme, variant)
	case "modal":
		return tm.getModalStyle(theme)
	case "input":
		return tm.getInputStyle(theme, variant)
	case "navigation":
		return tm.getNavigationStyle(theme, variant)
	case "status_card":
		return tm.getStatusCardStyle(theme)
	case "progress_bar":
		return tm.getProgressBarStyle(theme)
	case "notification":
		return tm.getNotificationStyle(theme)
	default:
		return tm.getBaseStyle(theme)
	}
}

// getButtonStyle returns button styling
func (tm *ThemeManager) getButtonStyle(theme *Theme, variant string) lipgloss.Style {
	var btnVariant ButtonVariant

	switch variant {
	case "primary":
		btnVariant = theme.Components.Button.Primary
	case "secondary":
		btnVariant = theme.Components.Button.Secondary
	case "danger":
		btnVariant = theme.Components.Button.Danger
	case "success":
		btnVariant = theme.Components.Button.Success
	case "ghost":
		btnVariant = theme.Components.Button.Ghost
	default:
		btnVariant = theme.Components.Button.Primary
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(btnVariant.Background.Hex)).
		Foreground(lipgloss.Color(btnVariant.Text.Hex)).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(btnVariant.Border.Hex)).
		Padding(0, 2)
}

// getTableStyle returns table styling
func (tm *ThemeManager) getTableStyle(theme *Theme, variant string) lipgloss.Style {
	table := theme.Components.Table

	switch variant {
	case "header":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(table.HeaderBackground.Hex)).
			Foreground(lipgloss.Color(table.HeaderText.Hex)).
			Bold(true).
			Padding(0, 1)
	case "row":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(table.RowBackground.Hex)).
			Foreground(lipgloss.Color(table.RowText.Hex)).
			Padding(0, 1)
	case "row_alt":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(table.RowBackgroundAlt.Hex)).
			Foreground(lipgloss.Color(table.RowText.Hex)).
			Padding(0, 1)
	case "selected":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(table.SelectedRow.Hex)).
			Foreground(lipgloss.Color(table.RowText.Hex)).
			Padding(0, 1)
	default:
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(table.Border.Hex))
	}
}

// getModalStyle returns modal styling
func (tm *ThemeManager) getModalStyle(theme *Theme) lipgloss.Style {
	modal := theme.Components.Modal

	return lipgloss.NewStyle().
		Background(lipgloss.Color(modal.Background.Hex)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(modal.Border.Hex)).
		Padding(2, 4)
}

// getInputStyle returns input styling
func (tm *ThemeManager) getInputStyle(theme *Theme, variant string) lipgloss.Style {
	input := theme.Components.Input

	style := lipgloss.NewStyle().
		Background(lipgloss.Color(input.Background.Hex)).
		Foreground(lipgloss.Color(input.Text.Hex)).
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)

	switch variant {
	case "focus":
		style = style.BorderForeground(lipgloss.Color(input.BorderFocus.Hex))
	case "error":
		style = style.BorderForeground(lipgloss.Color(input.BorderError.Hex))
	default:
		style = style.BorderForeground(lipgloss.Color(input.Border.Hex))
	}

	return style
}

// getNavigationStyle returns navigation styling
func (tm *ThemeManager) getNavigationStyle(theme *Theme, variant string) lipgloss.Style {
	nav := theme.Components.Navigation

	switch variant {
	case "active":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(nav.Background.Hex)).
			Foreground(lipgloss.Color(nav.TextActive.Hex)).
			Bold(true).
			Padding(0, 2)
	case "hover":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(nav.Background.Hex)).
			Foreground(lipgloss.Color(nav.TextHover.Hex)).
			Padding(0, 2)
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(nav.Background.Hex)).
			Foreground(lipgloss.Color(nav.Text.Hex)).
			Padding(0, 2)
	}
}

// getStatusCardStyle returns status card styling
func (tm *ThemeManager) getStatusCardStyle(theme *Theme) lipgloss.Style {
	card := theme.Components.StatusCard

	return lipgloss.NewStyle().
		Background(lipgloss.Color(card.Background.Hex)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(card.Border.Hex)).
		Padding(1, 2)
}

// getProgressBarStyle returns progress bar styling
func (tm *ThemeManager) getProgressBarStyle(theme *Theme) lipgloss.Style {
	bar := theme.Components.ProgressBar

	return lipgloss.NewStyle().
		Background(lipgloss.Color(bar.Background.Hex)).
		Foreground(lipgloss.Color(bar.Fill.Hex))
}

// getNotificationStyle returns notification styling
func (tm *ThemeManager) getNotificationStyle(theme *Theme) lipgloss.Style {
	notif := theme.Components.Notification

	return lipgloss.NewStyle().
		Background(lipgloss.Color(notif.Background.Hex)).
		Foreground(lipgloss.Color(notif.Text.Hex)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(notif.Border.Hex)).
		Padding(1, 2)
}

// getBaseStyle returns base styling
func (tm *ThemeManager) getBaseStyle(theme *Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Background.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex))
}

// OnThemeChange registers a callback for theme changes
func (tm *ThemeManager) OnThemeChange(callback func(*Theme)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.callbacks = append(tm.callbacks, callback)
}

// SaveTheme saves a theme to disk
func (tm *ThemeManager) SaveTheme(theme *Theme) error {
	if err := tm.validateTheme(theme); err != nil {
		return err
	}

	themesDir := filepath.Join(tm.configDir, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return ErrPersistenceFail.WithDetails(err.Error())
	}

	filename := filepath.Join(themesDir, theme.Name+".json")
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return ErrPersistenceFail.WithDetails(err.Error())
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return ErrPersistenceFail.WithDetails(err.Error())
	}

	// Register theme if not already registered
	tm.mu.Lock()
	tm.registry[theme.Name] = theme
	tm.mu.Unlock()

	return nil
}

// loadBuiltinThemes loads the built-in theme collection
func (tm *ThemeManager) loadBuiltinThemes() {
	themes := []*Theme{
		tm.createDefaultTheme(),
		tm.createTokyoNightTheme(),
		tm.createGitHubTheme(),
		tm.createOneDarkTheme(),
		tm.createHighContrastTheme(),
		tm.createMonochromeTheme(),
	}

	for _, theme := range themes {
		tm.registry[theme.Name] = theme
	}
}

// loadCustomThemes loads user-created themes from disk
func (tm *ThemeManager) loadCustomThemes() {
	themesDir := filepath.Join(tm.configDir, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return
	}

	files, err := filepath.Glob(filepath.Join(themesDir, "*.json"))
	if err != nil {
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var theme Theme
		if err := json.Unmarshal(data, &theme); err != nil {
			continue
		}

		tm.registry[theme.Name] = &theme
	}
}

// loadPreferences loads user preferences from disk
func (tm *ThemeManager) loadPreferences() {
	prefsFile := filepath.Join(tm.configDir, "theme_preferences.json")
	data, err := os.ReadFile(prefsFile)
	if err != nil {
		// Create default preferences
		tm.preferences = &ThemePreferences{
			ActiveTheme:        ThemeDefault,
			AutoDetectTerminal: true,
			RespectNoColor:     true,
			SyncWithSystem:     false,
			AccessibilityMode:  false,
			MotionReduced:      false,
			Overrides:          make(map[string]string),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		return
	}

	var prefs ThemePreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		tm.preferences = &ThemePreferences{
			ActiveTheme:        ThemeDefault,
			AutoDetectTerminal: true,
			RespectNoColor:     true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		return
	}

	tm.preferences = &prefs
}

// savePreferences saves user preferences to disk
func (tm *ThemeManager) savePreferences() {
	if tm.preferences == nil {
		return
	}

	if err := os.MkdirAll(tm.configDir, 0755); err != nil {
		return
	}

	prefsFile := filepath.Join(tm.configDir, "theme_preferences.json")
	data, err := json.MarshalIndent(tm.preferences, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(prefsFile, data, 0644)
}

// ColorUtilities provides color manipulation and validation utilities
type ColorUtilities struct{}

// NewColorUtilities creates a new color utilities instance
func NewColorUtilities() *ColorUtilities {
	return &ColorUtilities{}
}

// HexToRGB converts hex color to RGB
func (cu *ColorUtilities) HexToRGB(hex string) (*RGB, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return nil, ErrColorInvalid.WithDetails("invalid hex format")
	}

	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return nil, ErrColorInvalid.WithDetails("invalid red component")
	}

	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return nil, ErrColorInvalid.WithDetails("invalid green component")
	}

	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return nil, ErrColorInvalid.WithDetails("invalid blue component")
	}

	return &RGB{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, nil
}

// RGBToHex converts RGB to hex color
func (cu *ColorUtilities) RGBToHex(rgb RGB) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

// RGBToHSL converts RGB to HSL
func (cu *ColorUtilities) RGBToHSL(rgb RGB) (*HSL, error) {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))

	h := 0.0
	s := 0.0
	l := (max + min) / 2.0

	if max != min {
		delta := max - min

		if l > 0.5 {
			s = delta / (2.0 - max - min)
		} else {
			s = delta / (max + min)
		}

		switch max {
		case r:
			h = (g-b)/delta + (func() float64 {
				if g < b {
					return 6.0
				}
				return 0.0
			})()
		case g:
			h = (b-r)/delta + 2.0
		case b:
			h = (r-g)/delta + 4.0
		}

		h /= 6.0
	}

	return &HSL{
		H: uint16(h * 360),
		S: uint8(s * 100),
		L: uint8(l * 100),
	}, nil
}

// ContrastRatio calculates the contrast ratio between two colors
func (cu *ColorUtilities) ContrastRatio(color1, color2 Color) (float64, error) {
	rgb1, err := cu.HexToRGB(color1.Hex)
	if err != nil {
		return 0, err
	}

	rgb2, err := cu.HexToRGB(color2.Hex)
	if err != nil {
		return 0, err
	}

	l1 := cu.relativeLuminance(*rgb1)
	l2 := cu.relativeLuminance(*rgb2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05), nil
}

// relativeLuminance calculates the relative luminance of an RGB color
func (cu *ColorUtilities) relativeLuminance(rgb RGB) float64 {
	rsRGB := float64(rgb.R) / 255.0
	gsRGB := float64(rgb.G) / 255.0
	bsRGB := float64(rgb.B) / 255.0

	r := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}

	return 0.2126*r(rsRGB) + 0.7152*r(gsRGB) + 0.0722*r(bsRGB)
}

// AccessibilityChecker validates themes for accessibility compliance
type AccessibilityChecker struct {
	colorUtils *ColorUtilities
}

// NewAccessibilityChecker creates a new accessibility checker
func NewAccessibilityChecker() *AccessibilityChecker {
	return &AccessibilityChecker{
		colorUtils: NewColorUtilities(),
	}
}

// CheckAccessibility performs comprehensive accessibility validation
func (ac *AccessibilityChecker) CheckAccessibility(theme *Theme) (*AccessibilityInfo, error) {
	info := &AccessibilityInfo{
		ColorBlindSafe:       true,
		MotionSafe:           !theme.Animations.Enabled || theme.Animations.ReducedMotion,
		HighContrast:         false,
		LowVisionFriendly:    true,
		Warnings:             make([]string, 0),
		Recommendations:      make([]string, 0),
		ContrastCheckResults: make([]ContrastCheck, 0),
	}

	// Check contrast ratios for critical color combinations
	criticalChecks := []struct {
		fg   Color
		bg   Color
		name string
	}{
		{theme.Palette.TextPrimary, theme.Palette.Background, "primary_text_background"},
		{theme.Palette.TextSecondary, theme.Palette.Background, "secondary_text_background"},
		{theme.Components.Button.Primary.Text, theme.Components.Button.Primary.Background, "primary_button"},
		{theme.Components.Table.HeaderText, theme.Components.Table.HeaderBackground, "table_header"},
		{theme.Components.Input.Text, theme.Components.Input.Background, "input_field"},
	}

	minRatio := 21.0 // Track minimum ratio
	for _, check := range criticalChecks {
		ratio, err := ac.colorUtils.ContrastRatio(check.fg, check.bg)
		if err != nil {
			continue
		}

		if ratio < minRatio {
			minRatio = ratio
		}

		contrastCheck := ContrastCheck{
			ForegroundColor: check.fg.Hex,
			BackgroundColor: check.bg.Hex,
			ComponentName:   check.name,
			Ratio:           ratio,
			AACompliant:     ratio >= 4.5,
			AAACompliant:    ratio >= 7.0,
			Critical:        true,
		}

		info.ContrastCheckResults = append(info.ContrastCheckResults, contrastCheck)

		if ratio < 4.5 {
			info.Warnings = append(info.Warnings, fmt.Sprintf("Low contrast in %s: %.2f:1", check.name, ratio))
			info.LowVisionFriendly = false
		}

		if ratio < 3.0 {
			info.ColorBlindSafe = false
		}
	}

	info.ContrastRatio = minRatio

	// Determine WCAG level
	if minRatio >= 7.0 {
		info.WCAGLevel = "AAA"
		info.HighContrast = true
	} else if minRatio >= 4.5 {
		info.WCAGLevel = "AA"
	} else {
		info.WCAGLevel = "Fail"
		info.Recommendations = append(info.Recommendations, "Increase contrast ratios to meet WCAG AA standards")
	}

	// Check for high contrast theme
	if strings.Contains(strings.ToLower(theme.Name), "high") ||
		strings.Contains(strings.ToLower(theme.Name), "contrast") {
		info.HighContrast = true
	}

	return info, nil
}

// Built-in theme creation functions

func (tm *ThemeManager) createDefaultTheme() *Theme {
	return &Theme{
		Name:        ThemeDefault,
		Description: "Default light theme with clean, professional styling",
		Category:    CategoryStandard,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#ffffff", Name: "White"},
			Surface:         Color{Hex: "#f8f9fa", Name: "Light Gray"},
			Primary:         Color{Hex: "#007bff", Name: "Blue"},
			Secondary:       Color{Hex: "#6c757d", Name: "Gray"},
			Accent:          Color{Hex: "#28a745", Name: "Green"},
			Success:         Color{Hex: "#28a745", Name: "Success Green"},
			Warning:         Color{Hex: "#ffc107", Name: "Warning Yellow"},
			Error:           Color{Hex: "#dc3545", Name: "Error Red"},
			Info:            Color{Hex: "#17a2b8", Name: "Info Blue"},
			TextPrimary:     Color{Hex: "#212529", Name: "Dark Gray"},
			TextSecondary:   Color{Hex: "#6c757d", Name: "Medium Gray"},
			TextDisabled:    Color{Hex: "#adb5bd", Name: "Light Gray"},
			TextInverse:     Color{Hex: "#ffffff", Name: "White"},
			Border:          Color{Hex: "#dee2e6", Name: "Border Gray"},
			Divider:         Color{Hex: "#e9ecef", Name: "Divider Gray"},
			Focus:           Color{Hex: "#80bdff", Name: "Focus Blue"},
			Selected:        Color{Hex: "#e7f3ff", Name: "Selected Blue"},
			Hover:           Color{Hex: "#f1f3f5", Name: "Hover Gray"},
			StatusPending:   Color{Hex: "#6c757d", Name: "Pending Gray"},
			StatusRunning:   Color{Hex: "#007bff", Name: "Running Blue"},
			StatusCompleted: Color{Hex: "#28a745", Name: "Completed Green"},
			StatusFailed:    Color{Hex: "#dc3545", Name: "Failed Red"},
			StatusRetrying:  Color{Hex: "#ffc107", Name: "Retrying Yellow"},
		},
		Components: ComponentStyles{
			Button: ButtonStyles{
				Primary: ButtonVariant{
					Background:      Color{Hex: "#007bff"},
					Text:            Color{Hex: "#ffffff"},
					Border:          Color{Hex: "#007bff"},
					BackgroundHover: Color{Hex: "#0056b3"},
					TextHover:       Color{Hex: "#ffffff"},
					BorderHover:     Color{Hex: "#0056b3"},
				},
				Secondary: ButtonVariant{
					Background:      Color{Hex: "#6c757d"},
					Text:            Color{Hex: "#ffffff"},
					Border:          Color{Hex: "#6c757d"},
					BackgroundHover: Color{Hex: "#545b62"},
					TextHover:       Color{Hex: "#ffffff"},
					BorderHover:     Color{Hex: "#545b62"},
				},
			},
			Table: TableStyles{
				HeaderBackground: Color{Hex: "#f8f9fa"},
				HeaderText:       Color{Hex: "#212529"},
				RowBackground:    Color{Hex: "#ffffff"},
				RowBackgroundAlt: Color{Hex: "#f8f9fa"},
				RowText:          Color{Hex: "#212529"},
				Border:           Color{Hex: "#dee2e6"},
				SelectedRow:      Color{Hex: "#e7f3ff"},
				HoverRow:         Color{Hex: "#f1f3f5"},
			},
		},
		Typography: Typography{
			FontFamily:      "system-ui, -apple-system, sans-serif",
			FontSize:        "14px",
			LineHeight:      "1.5",
			LetterSpacing:   "0",
			FontWeight:      "400",
			MonospaceFamily: "SFMono-Regular, Consolas, monospace",
		},
		Animations: AnimationConfig{
			Enabled:         true,
			Duration:        "200ms",
			Easing:          "ease-in-out",
			ReducedMotion:   false,
			ThemeTransition: "all 200ms ease-in-out",
			HoverTransition: "all 150ms ease-in-out",
			FadeTransition:  "opacity 200ms ease-in-out",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (tm *ThemeManager) createTokyoNightTheme() *Theme {
	return &Theme{
		Name:        ThemeTokyoNight,
		Description: "Dark theme inspired by Tokyo's neon-lit nights",
		Category:    CategoryStandard,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#1a1b26", Name: "Tokyo Night"},
			Surface:         Color{Hex: "#24283b", Name: "Dark Surface"},
			Primary:         Color{Hex: "#7aa2f7", Name: "Blue"},
			Secondary:       Color{Hex: "#565f89", Name: "Gray Blue"},
			Accent:          Color{Hex: "#bb9af7", Name: "Purple"},
			Success:         Color{Hex: "#9ece6a", Name: "Green"},
			Warning:         Color{Hex: "#e0af68", Name: "Yellow"},
			Error:           Color{Hex: "#f7768e", Name: "Red"},
			Info:            Color{Hex: "#7dcfff", Name: "Cyan"},
			TextPrimary:     Color{Hex: "#c0caf5", Name: "Light Blue"},
			TextSecondary:   Color{Hex: "#9aa5ce", Name: "Medium Blue"},
			TextDisabled:    Color{Hex: "#565f89", Name: "Dark Blue"},
			TextInverse:     Color{Hex: "#1a1b26", Name: "Dark"},
			Border:          Color{Hex: "#414868", Name: "Border Blue"},
			Divider:         Color{Hex: "#32344a", Name: "Divider"},
			Focus:           Color{Hex: "#7aa2f7", Name: "Focus Blue"},
			Selected:        Color{Hex: "#364a82", Name: "Selected"},
			Hover:           Color{Hex: "#2d3149", Name: "Hover"},
			StatusPending:   Color{Hex: "#565f89", Name: "Pending"},
			StatusRunning:   Color{Hex: "#7aa2f7", Name: "Running"},
			StatusCompleted: Color{Hex: "#9ece6a", Name: "Completed"},
			StatusFailed:    Color{Hex: "#f7768e", Name: "Failed"},
			StatusRetrying:  Color{Hex: "#e0af68", Name: "Retrying"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (tm *ThemeManager) createGitHubTheme() *Theme {
	return &Theme{
		Name:        ThemeGitHub,
		Description: "Clean light theme matching GitHub's interface",
		Category:    CategoryStandard,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#ffffff", Name: "White"},
			Surface:         Color{Hex: "#f6f8fa", Name: "Light Gray"},
			Primary:         Color{Hex: "#0366d6", Name: "GitHub Blue"},
			Secondary:       Color{Hex: "#586069", Name: "Gray"},
			Accent:          Color{Hex: "#28a745", Name: "Green"},
			Success:         Color{Hex: "#28a745", Name: "Success"},
			Warning:         Color{Hex: "#ffd33d", Name: "Warning"},
			Error:           Color{Hex: "#d73a49", Name: "Error"},
			Info:            Color{Hex: "#0366d6", Name: "Info"},
			TextPrimary:     Color{Hex: "#24292e", Name: "Dark"},
			TextSecondary:   Color{Hex: "#586069", Name: "Gray"},
			TextDisabled:    Color{Hex: "#959da5", Name: "Light Gray"},
			TextInverse:     Color{Hex: "#ffffff", Name: "White"},
			Border:          Color{Hex: "#e1e4e8", Name: "Border"},
			Divider:         Color{Hex: "#eaecef", Name: "Divider"},
			Focus:           Color{Hex: "#79b8ff", Name: "Focus"},
			Selected:        Color{Hex: "#f1f8ff", Name: "Selected"},
			Hover:           Color{Hex: "#f6f8fa", Name: "Hover"},
			StatusPending:   Color{Hex: "#959da5", Name: "Pending"},
			StatusRunning:   Color{Hex: "#0366d6", Name: "Running"},
			StatusCompleted: Color{Hex: "#28a745", Name: "Completed"},
			StatusFailed:    Color{Hex: "#d73a49", Name: "Failed"},
			StatusRetrying:  Color{Hex: "#ffd33d", Name: "Retrying"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (tm *ThemeManager) createOneDarkTheme() *Theme {
	return &Theme{
		Name:        ThemeOneDark,
		Description: "Popular dark theme from Atom editor",
		Category:    CategoryStandard,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#282c34", Name: "Dark Gray"},
			Surface:         Color{Hex: "#21252b", Name: "Darker Gray"},
			Primary:         Color{Hex: "#61afef", Name: "Blue"},
			Secondary:       Color{Hex: "#5c6370", Name: "Gray"},
			Accent:          Color{Hex: "#c678dd", Name: "Purple"},
			Success:         Color{Hex: "#98c379", Name: "Green"},
			Warning:         Color{Hex: "#e5c07b", Name: "Yellow"},
			Error:           Color{Hex: "#e06c75", Name: "Red"},
			Info:            Color{Hex: "#56b6c2", Name: "Cyan"},
			TextPrimary:     Color{Hex: "#abb2bf", Name: "Light Gray"},
			TextSecondary:   Color{Hex: "#828997", Name: "Medium Gray"},
			TextDisabled:    Color{Hex: "#5c6370", Name: "Dark Gray"},
			TextInverse:     Color{Hex: "#282c34", Name: "Dark"},
			Border:          Color{Hex: "#3e4451", Name: "Border"},
			Divider:         Color{Hex: "#353b45", Name: "Divider"},
			Focus:           Color{Hex: "#61afef", Name: "Focus"},
			Selected:        Color{Hex: "#3e4451", Name: "Selected"},
			Hover:           Color{Hex: "#2c313c", Name: "Hover"},
			StatusPending:   Color{Hex: "#5c6370", Name: "Pending"},
			StatusRunning:   Color{Hex: "#61afef", Name: "Running"},
			StatusCompleted: Color{Hex: "#98c379", Name: "Completed"},
			StatusFailed:    Color{Hex: "#e06c75", Name: "Failed"},
			StatusRetrying:  Color{Hex: "#e5c07b", Name: "Retrying"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (tm *ThemeManager) createHighContrastTheme() *Theme {
	return &Theme{
		Name:        ThemeHighContrast,
		Description: "High contrast theme for accessibility",
		Category:    CategoryAccessibility,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#000000", Name: "Black"},
			Surface:         Color{Hex: "#1a1a1a", Name: "Dark Gray"},
			Primary:         Color{Hex: "#ffffff", Name: "White"},
			Secondary:       Color{Hex: "#cccccc", Name: "Light Gray"},
			Accent:          Color{Hex: "#ffff00", Name: "Yellow"},
			Success:         Color{Hex: "#00ff00", Name: "Bright Green"},
			Warning:         Color{Hex: "#ffff00", Name: "Bright Yellow"},
			Error:           Color{Hex: "#ff0000", Name: "Bright Red"},
			Info:            Color{Hex: "#00ffff", Name: "Bright Cyan"},
			TextPrimary:     Color{Hex: "#ffffff", Name: "White"},
			TextSecondary:   Color{Hex: "#cccccc", Name: "Light Gray"},
			TextDisabled:    Color{Hex: "#666666", Name: "Gray"},
			TextInverse:     Color{Hex: "#000000", Name: "Black"},
			Border:          Color{Hex: "#ffffff", Name: "White"},
			Divider:         Color{Hex: "#666666", Name: "Gray"},
			Focus:           Color{Hex: "#ffff00", Name: "Yellow"},
			Selected:        Color{Hex: "#333333", Name: "Dark Gray"},
			Hover:           Color{Hex: "#333333", Name: "Dark Gray"},
			StatusPending:   Color{Hex: "#cccccc", Name: "Light Gray"},
			StatusRunning:   Color{Hex: "#00ffff", Name: "Cyan"},
			StatusCompleted: Color{Hex: "#00ff00", Name: "Green"},
			StatusFailed:    Color{Hex: "#ff0000", Name: "Red"},
			StatusRetrying:  Color{Hex: "#ffff00", Name: "Yellow"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (tm *ThemeManager) createMonochromeTheme() *Theme {
	return &Theme{
		Name:        ThemeMonochrome,
		Description: "Pure black and white theme for minimal terminals",
		Category:    CategoryAccessibility,
		Version:     "1.0.0",
		Author:      "System",
		Palette: ColorPalette{
			Background:      Color{Hex: "#ffffff", Name: "White"},
			Surface:         Color{Hex: "#f0f0f0", Name: "Light Gray"},
			Primary:         Color{Hex: "#000000", Name: "Black"},
			Secondary:       Color{Hex: "#666666", Name: "Gray"},
			Accent:          Color{Hex: "#333333", Name: "Dark Gray"},
			Success:         Color{Hex: "#000000", Name: "Black"},
			Warning:         Color{Hex: "#666666", Name: "Gray"},
			Error:           Color{Hex: "#000000", Name: "Black"},
			Info:            Color{Hex: "#333333", Name: "Dark Gray"},
			TextPrimary:     Color{Hex: "#000000", Name: "Black"},
			TextSecondary:   Color{Hex: "#666666", Name: "Gray"},
			TextDisabled:    Color{Hex: "#cccccc", Name: "Light Gray"},
			TextInverse:     Color{Hex: "#ffffff", Name: "White"},
			Border:          Color{Hex: "#000000", Name: "Black"},
			Divider:         Color{Hex: "#cccccc", Name: "Light Gray"},
			Focus:           Color{Hex: "#000000", Name: "Black"},
			Selected:        Color{Hex: "#e0e0e0", Name: "Very Light Gray"},
			Hover:           Color{Hex: "#f0f0f0", Name: "Light Gray"},
			StatusPending:   Color{Hex: "#666666", Name: "Gray"},
			StatusRunning:   Color{Hex: "#333333", Name: "Dark Gray"},
			StatusCompleted: Color{Hex: "#000000", Name: "Black"},
			StatusFailed:    Color{Hex: "#000000", Name: "Black"},
			StatusRetrying:  Color{Hex: "#666666", Name: "Gray"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}