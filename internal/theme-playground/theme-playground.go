// Copyright 2025 James Ross
package themeplayground

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

// ThemeManager manages theme registry, application, and persistence
type ThemeManager struct {
	mu              sync.RWMutex
	registry        *ThemeRegistry
	current         *Theme
	persistence     *ThemePersistence
	accessibility   *AccessibilityChecker
	cache           *ThemeCache
	validator       *ThemeValidator
	configDir       string
	logger          *zap.Logger
	preferences     *UserPreferences
	styleCache      map[string]*LipGlossStyle
	observers       []ThemeObserver
	metrics         *ThemeMetrics
}

// ThemeObserver receives notifications when themes change
type ThemeObserver interface {
	OnThemeChanged(oldTheme, newTheme *Theme)
	OnThemeApplied(theme *Theme)
	OnThemeError(err error)
}

// NewThemeManager creates a new theme manager instance
func NewThemeManager(configDir string, logger *zap.Logger) (*ThemeManager, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	tm := &ThemeManager{
		configDir:  configDir,
		logger:     logger,
		styleCache: make(map[string]*LipGlossStyle),
		observers:  make([]ThemeObserver, 0),
		metrics:    NewThemeMetrics(),
	}

	// Initialize components
	var err error
	tm.registry, err = NewThemeRegistry(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create theme registry: %w", err)
	}

	tm.persistence, err = NewThemePersistence(configDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create theme persistence: %w", err)
	}

	tm.accessibility = NewAccessibilityChecker(logger)
	tm.cache = NewThemeCache(logger)
	tm.validator = NewThemeValidator(logger)

	// Load user preferences
	tm.preferences, err = tm.persistence.LoadPreferences()
	if err != nil {
		logger.Warn("Failed to load user preferences, using defaults", zap.Error(err))
		tm.preferences = tm.getDefaultPreferences()
	}

	// Load built-in themes
	if err := tm.loadBuiltinThemes(); err != nil {
		return nil, fmt.Errorf("failed to load built-in themes: %w", err)
	}

	// Load custom themes
	if err := tm.loadCustomThemes(); err != nil {
		logger.Warn("Failed to load custom themes", zap.Error(err))
	}

	// Apply active theme
	if tm.preferences.ActiveTheme != "" {
		if err := tm.ApplyTheme(tm.preferences.ActiveTheme); err != nil {
			logger.Warn("Failed to apply active theme, using default",
				zap.String("theme", tm.preferences.ActiveTheme),
				zap.Error(err))
			tm.ApplyTheme("default")
		}
	} else {
		tm.ApplyTheme("default")
	}

	return tm, nil
}

// GetCurrentTheme returns the currently active theme
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

// GetAvailableThemes returns all available themes
func (tm *ThemeManager) GetAvailableThemes() map[string]*Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make(map[string]*Theme)

	// Add built-in themes
	for name, theme := range tm.registry.Builtin {
		themes[name] = theme
	}

	// Add custom themes
	for name, theme := range tm.registry.Custom {
		themes[name] = theme
	}

	return themes
}

// GetThemesByCategory returns themes filtered by category
func (tm *ThemeManager) GetThemesByCategory(category ThemeCategory) []*Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var themes []*Theme

	for _, theme := range tm.registry.Builtin {
		if theme.Category == category {
			themes = append(themes, theme)
		}
	}

	for _, theme := range tm.registry.Custom {
		if theme.Category == category {
			themes = append(themes, theme)
		}
	}

	// Sort by name
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name < themes[j].Name
	})

	return themes
}

// ApplyTheme applies a theme by name
func (tm *ThemeManager) ApplyTheme(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	startTime := time.Now()
	defer func() {
		tm.metrics.RecordThemeApplication(name, time.Since(startTime))
	}()

	// Get theme from registry
	theme, err := tm.getTheme(name)
	if err != nil {
		tm.notifyObserversError(err)
		return fmt.Errorf("theme not found: %w", err)
	}

	// Validate theme if accessibility mode is enabled
	if tm.preferences.AccessibilityMode {
		if err := tm.validateAccessibility(theme); err != nil {
			tm.notifyObserversError(err)
			return fmt.Errorf("theme fails accessibility standards: %w", err)
		}
	}

	// Store previous theme for observers
	oldTheme := tm.current

	// Apply theme
	tm.current = theme
	tm.clearStyleCache()

	// Update preferences
	tm.preferences.ActiveTheme = name
	tm.addToHistory(name)

	// Persist preferences
	if err := tm.persistence.SavePreferences(tm.preferences); err != nil {
		tm.logger.Warn("Failed to save theme preferences", zap.Error(err))
	}

	// Notify observers
	if oldTheme != nil {
		tm.notifyObserversChanged(oldTheme, theme)
	}
	tm.notifyObserversApplied(theme)

	tm.logger.Info("Theme applied successfully",
		zap.String("theme", name),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// PreviewTheme temporarily applies a theme for preview (doesn't persist)
func (tm *ThemeManager) PreviewTheme(name string) (*Theme, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, err := tm.getTheme(name)
	if err != nil {
		return nil, fmt.Errorf("theme not found: %w", err)
	}

	// Validate theme if accessibility mode is enabled
	if tm.preferences.AccessibilityMode {
		if err := tm.validateAccessibility(theme); err != nil {
			return nil, fmt.Errorf("theme fails accessibility standards: %w", err)
		}
	}

	return theme, nil
}

// CreateCustomTheme creates a new custom theme
func (tm *ThemeManager) CreateCustomTheme(theme *Theme) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate theme
	result := tm.validator.ValidateTheme(theme)
	if !result.Valid {
		return fmt.Errorf("theme validation failed: %v", result.Errors)
	}

	// Set metadata
	theme.Category = CategoryCustom
	theme.CreatedAt = time.Now()
	theme.UpdatedAt = time.Now()

	// Add to registry
	tm.registry.Custom[theme.Name] = theme

	// Save to file
	if err := tm.persistence.SaveCustomTheme(theme); err != nil {
		return fmt.Errorf("failed to save custom theme: %w", err)
	}

	tm.logger.Info("Custom theme created",
		zap.String("name", theme.Name),
		zap.String("category", string(theme.Category)))

	return nil
}

// DeleteCustomTheme removes a custom theme
func (tm *ThemeManager) DeleteCustomTheme(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if theme exists and is custom
	theme, exists := tm.registry.Custom[name]
	if !exists {
		return fmt.Errorf("custom theme not found: %s", name)
	}

	if theme.Category != CategoryCustom {
		return fmt.Errorf("cannot delete built-in theme: %s", name)
	}

	// Remove from registry
	delete(tm.registry.Custom, name)

	// Delete from file system
	if err := tm.persistence.DeleteCustomTheme(name); err != nil {
		tm.logger.Warn("Failed to delete theme file", zap.String("theme", name), zap.Error(err))
	}

	// If this was the active theme, switch to default
	if tm.current != nil && tm.current.Name == name {
		tm.ApplyTheme("default")
	}

	tm.logger.Info("Custom theme deleted", zap.String("name", name))
	return nil
}

// ExportTheme exports a theme to a file
func (tm *ThemeManager) ExportTheme(name, filePath string) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, err := tm.getTheme(name)
	if err != nil {
		return fmt.Errorf("theme not found: %w", err)
	}

	return tm.persistence.ExportTheme(theme, filePath)
}

// ImportTheme imports a theme from a file
func (tm *ThemeManager) ImportTheme(filePath string) (*Theme, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	theme, err := tm.persistence.ImportTheme(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import theme: %w", err)
	}

	// Validate theme
	result := tm.validator.ValidateTheme(theme)
	if !result.Valid {
		return nil, fmt.Errorf("imported theme validation failed: %v", result.Errors)
	}

	// Set as custom theme
	theme.Category = CategoryCustom
	theme.CreatedAt = time.Now()
	theme.UpdatedAt = time.Now()

	// Add to registry
	tm.registry.Custom[theme.Name] = theme

	tm.logger.Info("Theme imported successfully",
		zap.String("name", theme.Name),
		zap.String("file", filePath))

	return theme, nil
}

// GetStyle returns a theme-aware style for a component
func (tm *ThemeManager) GetStyle(component, variant string) *LipGlossStyle {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.current == nil {
		return &LipGlossStyle{Style: lipgloss.NewStyle()}
	}

	cacheKey := fmt.Sprintf("%s:%s:%s", tm.current.Name, component, variant)

	if style, exists := tm.styleCache[cacheKey]; exists {
		return style
	}

	// Build style from theme
	style := tm.buildComponentStyle(component, variant)
	tm.styleCache[cacheKey] = style

	return style
}

// StyleBuilder returns a new style builder
func (tm *ThemeManager) StyleBuilder() *StyleBuilder {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return &StyleBuilder{
		theme:     tm.current,
		modifiers: make([]StyleModifier, 0),
		context:   make(map[string]interface{}),
	}
}

// AddObserver adds a theme change observer
func (tm *ThemeManager) AddObserver(observer ThemeObserver) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.observers = append(tm.observers, observer)
}

// RemoveObserver removes a theme change observer
func (tm *ThemeManager) RemoveObserver(observer ThemeObserver) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, obs := range tm.observers {
		if obs == observer {
			tm.observers = append(tm.observers[:i], tm.observers[i+1:]...)
			break
		}
	}
}

// GetPreferences returns current user preferences
func (tm *ThemeManager) GetPreferences() *UserPreferences {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.preferences
}

// UpdatePreferences updates user preferences
func (tm *ThemeManager) UpdatePreferences(prefs *UserPreferences) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.preferences = prefs

	// Apply accessibility settings
	if prefs.AccessibilityMode && tm.current != nil {
		if err := tm.validateAccessibility(tm.current); err != nil {
			// Switch to high contrast theme if current theme fails
			if err := tm.ApplyTheme("high-contrast-dark"); err != nil {
				tm.logger.Warn("Failed to apply high contrast theme", zap.Error(err))
			}
		}
	}

	return tm.persistence.SavePreferences(prefs)
}

// ValidateTheme validates a theme for errors and accessibility
func (tm *ThemeManager) ValidateTheme(theme *Theme) *ValidationResult {
	return tm.validator.ValidateTheme(theme)
}

// GetThemeMetrics returns theme usage metrics
func (tm *ThemeManager) GetThemeMetrics() *ThemeMetrics {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.metrics
}

// Cleanup performs cleanup operations
func (tm *ThemeManager) Cleanup() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Save final preferences
	if err := tm.persistence.SavePreferences(tm.preferences); err != nil {
		tm.logger.Warn("Failed to save final preferences", zap.Error(err))
	}

	// Clear caches
	tm.clearStyleCache()
	tm.cache.Clear()

	tm.logger.Info("Theme manager cleanup completed")
	return nil
}

// Private methods

func (tm *ThemeManager) getTheme(name string) (*Theme, error) {
	// Check built-in themes first
	if theme, exists := tm.registry.Builtin[name]; exists {
		return theme, nil
	}

	// Check custom themes
	if theme, exists := tm.registry.Custom[name]; exists {
		return theme, nil
	}

	return nil, fmt.Errorf("theme not found: %s", name)
}

func (tm *ThemeManager) validateAccessibility(theme *Theme) error {
	report := tm.accessibility.ValidateTheme(theme)

	if !report.MeetsStandards(tm.preferences.AccessibilityMode) {
		var issues []string
		for _, warning := range theme.Accessibility.Warnings {
			if warning.Severity == "high" || warning.Severity == "critical" {
				issues = append(issues, warning.Message)
			}
		}
		return fmt.Errorf("accessibility validation failed: %s", strings.Join(issues, "; "))
	}

	return nil
}

func (tm *ThemeManager) loadBuiltinThemes() error {
	themes := []*Theme{
		tm.createDefaultTheme(),
		tm.createTokyoNightTheme(),
		tm.createGitHubTheme(),
		tm.createOneDarkTheme(),
		tm.createHighContrastDarkTheme(),
		tm.createHighContrastLightTheme(),
		tm.createMonochromeTheme(),
		tm.createSolarizedDarkTheme(),
		tm.createSolarizedLightTheme(),
		tm.createDraculaTheme(),
	}

	for _, theme := range themes {
		tm.registry.Builtin[theme.Name] = theme
	}

	tm.logger.Info("Loaded built-in themes", zap.Int("count", len(themes)))
	return nil
}

func (tm *ThemeManager) loadCustomThemes() error {
	customDir := filepath.Join(tm.configDir, "themes")
	if _, err := os.Stat(customDir); os.IsNotExist(err) {
		return nil // No custom themes directory
	}

	files, err := filepath.Glob(filepath.Join(customDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list custom themes: %w", err)
	}

	for _, file := range files {
		theme, err := tm.persistence.LoadCustomTheme(filepath.Base(file))
		if err != nil {
			tm.logger.Warn("Failed to load custom theme",
				zap.String("file", file),
				zap.Error(err))
			continue
		}

		tm.registry.Custom[theme.Name] = theme
	}

	tm.logger.Info("Loaded custom themes", zap.Int("count", len(tm.registry.Custom)))
	return nil
}

func (tm *ThemeManager) buildComponentStyle(component, variant string) *LipGlossStyle {
	if tm.current == nil {
		return &LipGlossStyle{Style: lipgloss.NewStyle()}
	}

	style := lipgloss.NewStyle()
	var compVariant ComponentVariant

	// Get component variant from theme
	switch component {
	case "button":
		switch variant {
		case "primary":
			compVariant = tm.current.Components.Button.Primary
		case "secondary":
			compVariant = tm.current.Components.Button.Secondary
		case "danger":
			compVariant = tm.current.Components.Button.Danger
		case "ghost":
			compVariant = tm.current.Components.Button.Ghost
		case "disabled":
			compVariant = tm.current.Components.Button.Disabled
		}
	case "table":
		switch variant {
		case "header":
			compVariant = tm.current.Components.Table.Header
		case "row":
			compVariant = tm.current.Components.Table.Row
		case "row_alt":
			compVariant = tm.current.Components.Table.RowAlt
		case "row_hover":
			compVariant = tm.current.Components.Table.RowHover
		case "cell":
			compVariant = tm.current.Components.Table.Cell
		}
	// Add more components as needed
	default:
		// Return basic style with theme colors
		style = style.
			Foreground(tm.current.Palette.TextPrimary.ToLipgloss()).
			Background(tm.current.Palette.Background.ToLipgloss())
	}

	// Apply component variant styling
	if compVariant.Foreground.Hex != "" {
		style = style.Foreground(compVariant.Foreground.ToLipgloss())
	}
	if compVariant.Background.Hex != "" {
		style = style.Background(compVariant.Background.ToLipgloss())
	}

	// Apply border if defined
	if compVariant.Border.Width > 0 {
		style = style.Border(lipgloss.NormalBorder(),
			compVariant.Border.Sides.Top,
			compVariant.Border.Sides.Right,
			compVariant.Border.Sides.Bottom,
			compVariant.Border.Sides.Left).
			BorderForeground(compVariant.Border.Color.ToLipgloss())
	}

	// Apply padding
	if compVariant.Padding.Top > 0 || compVariant.Padding.Right > 0 ||
		compVariant.Padding.Bottom > 0 || compVariant.Padding.Left > 0 {
		style = style.Padding(
			compVariant.Padding.Top,
			compVariant.Padding.Right,
			compVariant.Padding.Bottom,
			compVariant.Padding.Left)
	}

	// Apply margin
	if compVariant.Margin.Top > 0 || compVariant.Margin.Right > 0 ||
		compVariant.Margin.Bottom > 0 || compVariant.Margin.Left > 0 {
		style = style.Margin(
			compVariant.Margin.Top,
			compVariant.Margin.Right,
			compVariant.Margin.Bottom,
			compVariant.Margin.Left)
	}

	return &LipGlossStyle{
		Style:     style,
		theme:     tm.current,
		component: component,
		variant:   variant,
		context:   make(map[string]interface{}),
	}
}

func (tm *ThemeManager) clearStyleCache() {
	tm.styleCache = make(map[string]*LipGlossStyle)
}

func (tm *ThemeManager) addToHistory(themeName string) {
	entry := ThemeHistoryEntry{
		Theme:     themeName,
		AppliedAt: time.Now(),
		Context:   "manual",
		Automatic: false,
	}

	// Update duration of previous entry
	if len(tm.preferences.History) > 0 {
		last := &tm.preferences.History[len(tm.preferences.History)-1]
		last.Duration = entry.AppliedAt.Sub(last.AppliedAt)
	}

	tm.preferences.History = append(tm.preferences.History, entry)

	// Keep only last 100 entries
	if len(tm.preferences.History) > 100 {
		tm.preferences.History = tm.preferences.History[len(tm.preferences.History)-100:]
	}
}

func (tm *ThemeManager) getDefaultPreferences() *UserPreferences {
	return &UserPreferences{
		ActiveTheme:        "default",
		AutoDetectTerminal: true,
		RespectNoColor:     true,
		SyncWithSystem:     false,
		AccessibilityMode:  false,
		ReducedMotion:      false,
		HighContrast:       false,
		CustomThemesDir:    filepath.Join(tm.configDir, "themes"),
		Favorites:          []string{},
		History:            []ThemeHistoryEntry{},
		Customizations:     make(map[string]interface{}),
	}
}

func (tm *ThemeManager) notifyObserversChanged(oldTheme, newTheme *Theme) {
	for _, observer := range tm.observers {
		go observer.OnThemeChanged(oldTheme, newTheme)
	}
}

func (tm *ThemeManager) notifyObserversApplied(theme *Theme) {
	for _, observer := range tm.observers {
		go observer.OnThemeApplied(theme)
	}
}

func (tm *ThemeManager) notifyObserversError(err error) {
	for _, observer := range tm.observers {
		go observer.OnThemeError(err)
	}
}

// StyleBuilder methods

// Component sets the component type
func (sb *StyleBuilder) Component(component string) *StyleBuilder {
	sb.component = component
	return sb
}

// Variant sets the component variant
func (sb *StyleBuilder) Variant(variant string) *StyleBuilder {
	sb.variant = variant
	return sb
}

// With adds a style modifier
func (sb *StyleBuilder) With(modifier StyleModifier) *StyleBuilder {
	sb.modifiers = append(sb.modifiers, modifier)
	return sb
}

// Context sets context data
func (sb *StyleBuilder) Context(key string, value interface{}) *StyleBuilder {
	sb.context[key] = value
	return sb
}

// Build creates the final styled component
func (sb *StyleBuilder) Build() *LipGlossStyle {
	style := &LipGlossStyle{
		Style:     lipgloss.NewStyle(),
		theme:     sb.theme,
		component: sb.component,
		variant:   sb.variant,
		context:   sb.context,
	}

	// Apply modifiers
	for _, modifier := range sb.modifiers {
		style = modifier(style)
	}

	return style
}

// Common style modifiers

// WithColor sets foreground color
func WithColor(color Color) StyleModifier {
	return func(style *LipGlossStyle) *LipGlossStyle {
		style.Style = style.Style.Foreground(color.ToLipgloss())
		return style
	}
}

// WithBackground sets background color
func WithBackground(color Color) StyleModifier {
	return func(style *LipGlossStyle) *LipGlossStyle {
		style.Style = style.Style.Background(color.ToLipgloss())
		return style
	}
}

// WithBorder adds border styling
func WithBorder(borderStyle BorderStyle) StyleModifier {
	return func(style *LipGlossStyle) *LipGlossStyle {
		style.Style = style.Style.Border(lipgloss.NormalBorder(),
			borderStyle.Sides.Top,
			borderStyle.Sides.Right,
			borderStyle.Sides.Bottom,
			borderStyle.Sides.Left).
			BorderForeground(borderStyle.Color.ToLipgloss())
		return style
	}
}

// WithPadding adds padding
func WithPadding(padding Spacing) StyleModifier {
	return func(style *LipGlossStyle) *LipGlossStyle {
		style.Style = style.Style.Padding(
			padding.Top, padding.Right, padding.Bottom, padding.Left)
		return style
	}
}