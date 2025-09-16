// Copyright 2025 James Ross
package themeplayground

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ThemeIntegration provides integration with the main application
type ThemeIntegration struct {
	themeManager *ThemeManager
	styleCache   map[string]lipgloss.Style
	cacheMu      sync.RWMutex
	callbacks    []ThemeChangeCallback
	callbacksMu  sync.RWMutex
}

// ThemeChangeCallback is called when the theme changes
type ThemeChangeCallback func(oldTheme, newTheme *Theme)

// NewThemeIntegration creates a new theme integration instance
func NewThemeIntegration(configDir string) (*ThemeIntegration, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, ErrConfigDirectoryNotFound.WithDetails(err.Error())
	}

	themeManager := NewThemeManager(configDir)

	integration := &ThemeIntegration{
		themeManager: themeManager,
		styleCache:   make(map[string]lipgloss.Style),
		callbacks:    make([]ThemeChangeCallback, 0),
	}

	// Register for theme change notifications
	themeManager.OnThemeChange(integration.onThemeChanged)

	// Detect terminal capabilities and adjust theme if needed
	integration.detectAndAdjustTheme()

	return integration, nil
}

// GetThemeManager returns the underlying theme manager
func (ti *ThemeIntegration) GetThemeManager() *ThemeManager {
	return ti.themeManager
}

// GetStyle returns a cached or newly created style for a component
func (ti *ThemeIntegration) GetStyle(component, variant string) lipgloss.Style {
	cacheKey := fmt.Sprintf("%s:%s", component, variant)

	ti.cacheMu.RLock()
	if style, exists := ti.styleCache[cacheKey]; exists {
		ti.cacheMu.RUnlock()
		return style
	}
	ti.cacheMu.RUnlock()

	// Create new style
	style := ti.themeManager.GetStyleFor(component, variant)

	ti.cacheMu.Lock()
	ti.styleCache[cacheKey] = style
	ti.cacheMu.Unlock()

	return style
}

// ClearStyleCache clears the style cache
func (ti *ThemeIntegration) ClearStyleCache() {
	ti.cacheMu.Lock()
	defer ti.cacheMu.Unlock()
	ti.styleCache = make(map[string]lipgloss.Style)
}

// OnThemeChange registers a callback for theme changes
func (ti *ThemeIntegration) OnThemeChange(callback ThemeChangeCallback) {
	ti.callbacksMu.Lock()
	defer ti.callbacksMu.Unlock()
	ti.callbacks = append(ti.callbacks, callback)
}

// onThemeChanged handles theme change events
func (ti *ThemeIntegration) onThemeChanged(newTheme *Theme) {
	// Clear style cache when theme changes
	ti.ClearStyleCache()

	// Notify registered callbacks
	ti.callbacksMu.RLock()
	callbacks := make([]ThemeChangeCallback, len(ti.callbacks))
	copy(callbacks, ti.callbacks)
	ti.callbacksMu.RUnlock()

	for _, callback := range callbacks {
		callback(nil, newTheme) // We don't track old theme here
	}
}

// detectAndAdjustTheme detects terminal capabilities and adjusts theme accordingly
func (ti *ThemeIntegration) detectAndAdjustTheme() {
	prefs := ti.themeManager.preferences
	if prefs == nil || !prefs.AutoDetectTerminal {
		return
	}

	// Check for NO_COLOR environment variable
	if prefs.RespectNoColor && os.Getenv("NO_COLOR") != "" {
		ti.themeManager.SetActiveTheme(ThemeMonochrome)
		return
	}

	// Check terminal color support
	colorTerm := os.Getenv("COLORTERM")
	term := os.Getenv("TERM")

	// Prefer high contrast for accessibility if requested
	if prefs.AccessibilityMode {
		ti.themeManager.SetActiveTheme(ThemeHighContrast)
		return
	}

	// Detect dark/light mode from terminal or system
	if ti.isDarkMode() {
		// Use a dark theme as default
		if ti.supportsFullColors(colorTerm, term) {
			ti.themeManager.SetActiveTheme(ThemeTokyoNight)
		} else {
			ti.themeManager.SetActiveTheme(ThemeOneDark)
		}
	} else {
		// Use a light theme as default
		if ti.supportsFullColors(colorTerm, term) {
			ti.themeManager.SetActiveTheme(ThemeGitHub)
		} else {
			ti.themeManager.SetActiveTheme(ThemeDefault)
		}
	}
}

// isDarkMode detects if the terminal/system is in dark mode
func (ti *ThemeIntegration) isDarkMode() bool {
	// Check various environment variables and heuristics

	// Check TERM_PROGRAM for known terminals
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		// Could query iTerm2 for theme info, but this is complex
		return true // Default to dark for iTerm
	case "vscode":
		// VS Code terminal
		return true // Usually dark
	}

	// Check for explicit dark mode indicators
	if strings.Contains(strings.ToLower(os.Getenv("TERM")), "dark") {
		return true
	}

	// Default heuristic: assume dark mode for modern terminals
	colorTerm := os.Getenv("COLORTERM")
	return colorTerm == "truecolor" || colorTerm == "24bit"
}

// supportsFullColors checks if the terminal supports full color palette
func (ti *ThemeIntegration) supportsFullColors(colorTerm, term string) bool {
	// Check for truecolor support
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return true
	}

	// Check TERM for 256 color support
	if strings.Contains(term, "256") || strings.Contains(term, "xterm") {
		return true
	}

	return false
}

// StyleHelper provides convenient methods for common styling operations
type StyleHelper struct {
	integration *ThemeIntegration
}

// NewStyleHelper creates a new style helper
func NewStyleHelper(integration *ThemeIntegration) *StyleHelper {
	return &StyleHelper{
		integration: integration,
	}
}

// Button creates a styled button
func (sh *StyleHelper) Button(text, variant string) string {
	style := sh.integration.GetStyle("button", variant)
	return style.Render(text)
}

// Table creates table styling helpers
func (sh *StyleHelper) Table() *TableStyleHelper {
	return &TableStyleHelper{styleHelper: sh}
}

// Status creates status styling helpers
func (sh *StyleHelper) Status() *StatusStyleHelper {
	return &StatusStyleHelper{styleHelper: sh}
}

// TableStyleHelper provides table-specific styling
type TableStyleHelper struct {
	styleHelper *StyleHelper
}

// Header renders a table header cell
func (tsh *TableStyleHelper) Header(text string) string {
	style := tsh.styleHelper.integration.GetStyle("table", "header")
	return style.Render(text)
}

// Row renders a table row cell
func (tsh *TableStyleHelper) Row(text string, isAlternate bool) string {
	variant := "row"
	if isAlternate {
		variant = "row_alt"
	}
	style := tsh.styleHelper.integration.GetStyle("table", variant)
	return style.Render(text)
}

// SelectedRow renders a selected table row cell
func (tsh *TableStyleHelper) SelectedRow(text string) string {
	style := tsh.styleHelper.integration.GetStyle("table", "selected")
	return style.Render(text)
}

// StatusStyleHelper provides status-specific styling
type StatusStyleHelper struct {
	styleHelper *StyleHelper
}

// StatusBadge renders a status badge with appropriate color
func (ssh *StatusStyleHelper) StatusBadge(status string) string {
	theme := ssh.styleHelper.integration.themeManager.GetActiveTheme()
	if theme == nil {
		return status
	}

	var color Color
	switch strings.ToLower(status) {
	case "pending":
		color = theme.Palette.StatusPending
	case "running":
		color = theme.Palette.StatusRunning
	case "completed", "success":
		color = theme.Palette.StatusCompleted
	case "failed", "error":
		color = theme.Palette.StatusFailed
	case "retrying":
		color = theme.Palette.StatusRetrying
	default:
		color = theme.Palette.TextSecondary
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color.Hex)).
		Bold(true)

	return style.Render(status)
}

// ProgressBar renders a progress bar
func (ssh *StatusStyleHelper) ProgressBar(current, total int, width int) string {
	if total == 0 || width <= 0 {
		return ""
	}

	theme := ssh.styleHelper.integration.themeManager.GetActiveTheme()
	if theme == nil {
		return fmt.Sprintf("%d/%d", current, total)
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	fillStyle := lipgloss.NewStyle().Background(lipgloss.Color(theme.Palette.Primary.Hex))
	emptyStyle := lipgloss.NewStyle().Background(lipgloss.Color(theme.Palette.Border.Hex))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fillStyle.Render(bar[:filled]) + emptyStyle.Render(bar[filled:])
}

// ThemePlayground provides the main playground interface
type ThemePlayground struct {
	integration *ThemeIntegration
	handler     *PlaygroundHandler
	styleHelper *StyleHelper
	isActive    bool
	mu          sync.RWMutex
}

// NewThemePlayground creates a new theme playground instance
func NewThemePlayground(configDir string) (*ThemePlayground, error) {
	integration, err := NewThemeIntegration(configDir)
	if err != nil {
		return nil, err
	}

	playground := &ThemePlayground{
		integration: integration,
		handler:     NewPlaygroundHandler(integration.GetThemeManager()),
		styleHelper: NewStyleHelper(integration),
		isActive:    true,
	}

	return playground, nil
}

// Start starts the theme playground
func (tp *ThemePlayground) Start(ctx context.Context) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.isActive {
		return nil // Already running
	}

	tp.isActive = true

	// Start background tasks if needed
	go tp.backgroundTasks(ctx)

	return nil
}

// Stop stops the theme playground
func (tp *ThemePlayground) Stop() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.isActive = false
	return nil
}

// IsActive returns whether the playground is active
func (tp *ThemePlayground) IsActive() bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.isActive
}

// GetIntegration returns the theme integration
func (tp *ThemePlayground) GetIntegration() *ThemeIntegration {
	return tp.integration
}

// GetHandler returns the HTTP handler
func (tp *ThemePlayground) GetHandler() *PlaygroundHandler {
	return tp.handler
}

// GetStyleHelper returns the style helper
func (tp *ThemePlayground) GetStyleHelper() *StyleHelper {
	return tp.styleHelper
}

// backgroundTasks runs background maintenance tasks
func (tp *ThemePlayground) backgroundTasks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !tp.IsActive() {
				return
			}

			// Perform maintenance tasks
			tp.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance
func (tp *ThemePlayground) performMaintenance() {
	// Clear old cache entries
	tp.integration.ClearStyleCache()

	// Check for configuration updates
	// (Could reload themes from disk if they've changed)
}

// GetConfigDirectory returns the configuration directory path
func (tp *ThemePlayground) GetConfigDirectory() string {
	return tp.integration.themeManager.configDir
}

// GetThemesDirectory returns the themes directory path
func (tp *ThemePlayground) GetThemesDirectory() string {
	return filepath.Join(tp.integration.themeManager.configDir, "themes")
}

// GetPreferencesFile returns the preferences file path
func (tp *ThemePlayground) GetPreferencesFile() string {
	return filepath.Join(tp.integration.themeManager.configDir, "theme_preferences.json")
}