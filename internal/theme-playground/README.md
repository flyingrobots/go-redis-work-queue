# Theme Playground API Documentation

The Theme Playground provides a comprehensive theming system for Go applications with live preview capabilities, accessibility validation, and persistent theme management.

## Overview

The theme playground consists of several key components:

- **Theme Manager**: Core theming system with built-in themes and custom theme support
- **Playground UI**: Interactive terminal UI for previewing and selecting themes
- **Persistence Manager**: Save/load themes and preferences across sessions
- **Color Utilities**: Color conversion and accessibility validation
- **Accessibility Checker**: WCAG compliance validation

## Quick Start

```go
package main

import (
    "github.com/flyingrobots/go-redis-work-queue/internal/theme-playground"
)

func main() {
    // Initialize theme manager
    tm := themeplayground.NewThemeManager("/path/to/config")

    // Get current theme
    activeTheme := tm.GetActiveTheme()

    // Apply styles
    style := tm.GetStyleFor("text", "primary")

    // Run interactive playground
    themeplayground.RunPlayground(tm)
}
```

## Core API

### ThemeManager

The `ThemeManager` is the central component for theme management.

#### Constructor

```go
func NewThemeManager(configDir string) *ThemeManager
```

Creates a new theme manager with the specified configuration directory.

#### Core Methods

```go
// Theme Management
func (tm *ThemeManager) GetActiveTheme() *Theme
func (tm *ThemeManager) SetActiveTheme(name string) error
func (tm *ThemeManager) GetTheme(name string) (*Theme, error)
func (tm *ThemeManager) ListThemes() []Theme
func (tm *ThemeManager) RegisterTheme(theme *Theme) error

// Style Application
func (tm *ThemeManager) GetStyleFor(component, variant string) lipgloss.Style

// Event Handling
func (tm *ThemeManager) OnThemeChange(callback func(*Theme))
```

#### Example Usage

```go
tm := themeplayground.NewThemeManager("~/.config/myapp")

// Set theme
err := tm.SetActiveTheme("tokyo-night")
if err != nil {
    log.Fatal(err)
}

// Get styled text
headerStyle := tm.GetStyleFor("text", "primary")
fmt.Print(headerStyle.Render("Welcome!"))

// Register theme change callback
tm.OnThemeChange(func(theme *Theme) {
    fmt.Printf("Theme changed to: %s\n", theme.Name)
})
```

### Built-in Themes

The following themes are available by default:

- `default`: Clean light theme
- `tokyo-night`: Popular dark theme with purple accents
- `github`: GitHub-inspired light theme
- `one-dark`: Atom One Dark theme
- `solarized-light`: Solarized light variant
- `solarized-dark`: Solarized dark variant
- `dracula`: Dracula theme
- `high-contrast`: High contrast for accessibility
- `monochrome`: Black and white theme
- `terminal-classic`: Classic terminal colors

### Theme Structure

```go
type Theme struct {
    Name            string            `json:"name"`
    Description     string            `json:"description"`
    Category        ThemeCategory     `json:"category"`
    Version         string            `json:"version"`
    Author          string            `json:"author"`
    Accessibility   AccessibilityInfo `json:"accessibility"`
    Palette         ColorPalette      `json:"palette"`
    Components      ComponentStyles   `json:"components"`
    Typography      Typography        `json:"typography"`
    Animations      AnimationConfig   `json:"animations"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
}
```

### Color Palette

```go
type ColorPalette struct {
    // Base colors
    Background Color `json:"background"`
    Surface    Color `json:"surface"`
    Primary    Color `json:"primary"`
    Secondary  Color `json:"secondary"`
    Accent     Color `json:"accent"`

    // Semantic colors
    Success Color `json:"success"`
    Warning Color `json:"warning"`
    Error   Color `json:"error"`
    Info    Color `json:"info"`

    // Text colors
    TextPrimary   Color `json:"text_primary"`
    TextSecondary Color `json:"text_secondary"`
    TextDisabled  Color `json:"text_disabled"`
    TextInverse   Color `json:"text_inverse"`

    // Status colors for queue states
    StatusPending   Color `json:"status_pending"`
    StatusRunning   Color `json:"status_running"`
    StatusCompleted Color `json:"status_completed"`
    StatusFailed    Color `json:"status_failed"`
    StatusRetrying  Color `json:"status_retrying"`
}
```

## Style Components

The theme system provides pre-configured styles for common UI components:

### Available Components and Variants

- **text**: `primary`, `secondary`, `tertiary`, `inverse`
- **button**: `primary`, `secondary`, `danger`, `success`, `ghost`
- **table**: `default`, `header`, `row`, `selected`
- **modal**: `default`
- **input**: `default`, `focus`, `error`
- **navigation**: `default`, `active`, `hover`
- **status**: `pending`, `running`, `completed`, `failed`, `error`
- **header**: `default`
- **surface**: `default`

### Style Usage Examples

```go
// Basic text styles
primaryText := tm.GetStyleFor("text", "primary")
secondaryText := tm.GetStyleFor("text", "secondary")

// Button styles
primaryButton := tm.GetStyleFor("button", "primary")
dangerButton := tm.GetStyleFor("button", "danger")

// Status indicators
successStatus := tm.GetStyleFor("status", "completed")
errorStatus := tm.GetStyleFor("status", "failed")

// Apply styles
fmt.Print(primaryText.Render("Important text"))
fmt.Print(primaryButton.Render(" Save "))
fmt.Print(successStatus.Render(" Complete"))
```

## Interactive Playground

### Running the Playground

```go
func RunPlayground(themeManager *ThemeManager) error
```

Launches an interactive terminal UI for theme selection and preview.

#### Keyboard Controls

- `‘/“` or `k/j`: Navigate theme list
- `Enter`: Apply selected theme
- `p`: Preview theme (temporary)
- `a`: Apply previewed theme (permanent)
- `r`: Reset to default theme
- `?`: Toggle help
- `q`: Quit

### Theme Selector Component

```go
func ThemeSelector(themeManager *ThemeManager, compact bool) string
```

Creates a theme selection UI component for embedding in other interfaces.

### Quick Theme Toggle

```go
type QuickThemeToggle struct {
    // ...
}

func NewQuickThemeToggle(themeManager *ThemeManager) *QuickThemeToggle
func (qt *QuickThemeToggle) HandleKey(key string) bool
```

Provides quick keyboard shortcuts for theme switching (1-9, 0 for different themes).

## Persistence

### PersistenceManager

Handles saving and loading themes and preferences.

```go
func NewPersistenceManager(config *PersistenceConfig) (*PersistenceManager, error)

// Theme management
func (pm *PersistenceManager) SaveCustomTheme(theme *Theme) error
func (pm *PersistenceManager) GetCustomThemes() map[string]*Theme
func (pm *PersistenceManager) DeleteCustomTheme(themeName string) error

// Import/Export
func (pm *PersistenceManager) ExportTheme(themeName string, filePath string) error
func (pm *PersistenceManager) ImportTheme(filePath string) (*Theme, error)

// Backup/Restore
func (pm *PersistenceManager) BackupThemes(backupPath string) error
func (pm *PersistenceManager) RestoreThemes(backupPath string, overwrite bool) error

// Preferences
func (pm *PersistenceManager) GetPreferences() *ThemePreferences
func (pm *PersistenceManager) UpdatePreferences(prefs *ThemePreferences) error
```

### Theme Preferences

```go
type ThemePreferences struct {
    ActiveTheme        string            `json:"active_theme"`
    AutoDetectTerminal bool              `json:"auto_detect_terminal"`
    RespectNoColor     bool              `json:"respect_no_color"`
    SyncWithSystem     bool              `json:"sync_with_system"`
    AccessibilityMode  bool              `json:"accessibility_mode"`
    MotionReduced      bool              `json:"motion_reduced"`
    CustomAccent       *Color            `json:"custom_accent,omitempty"`
    Overrides          map[string]string `json:"overrides"`
}
```

## Color Utilities

### ColorUtilities

Provides color conversion and manipulation functions.

```go
func NewColorUtilities() *ColorUtilities

// Color conversion
func (cu *ColorUtilities) HexToRGB(hex string) (*RGB, error)
func (cu *ColorUtilities) RGBToHex(rgb RGB) string
func (cu *ColorUtilities) RGBToHSL(rgb RGB) (*HSL, error)

// Accessibility
func (cu *ColorUtilities) ContrastRatio(color1, color2 Color) (float64, error)
```

### Usage Examples

```go
cu := themeplayground.NewColorUtilities()

// Convert colors
rgb, err := cu.HexToRGB("#ff6b9d")
hex := cu.RGBToHex(RGB{R: 255, G: 107, B: 157})

// Check contrast
color1 := Color{Hex: "#000000"}
color2 := Color{Hex: "#ffffff"}
ratio, err := cu.ContrastRatio(color1, color2)
// ratio H 21 (maximum contrast)
```

## Accessibility

### AccessibilityChecker

Validates themes for WCAG compliance.

```go
func NewAccessibilityChecker() *AccessibilityChecker
func (ac *AccessibilityChecker) CheckAccessibility(theme *Theme) (*AccessibilityInfo, error)
```

### AccessibilityInfo

```go
type AccessibilityInfo struct {
    ContrastRatio         float64         `json:"contrast_ratio"`
    WCAGLevel            string          `json:"wcag_level"` // AA, AAA
    ColorBlindSafe       bool            `json:"color_blind_safe"`
    MotionSafe           bool            `json:"motion_safe"`
    HighContrast         bool            `json:"high_contrast"`
    LowVisionFriendly    bool            `json:"low_vision_friendly"`
    Warnings             []string        `json:"warnings"`
    Recommendations      []string        `json:"recommendations"`
    ContrastCheckResults []ContrastCheck `json:"contrast_checks"`
}
```

## Custom Themes

### Creating Custom Themes

```go
customTheme := &themeplayground.Theme{
    Name:        "my-custom-theme",
    Description: "My custom theme",
    Category:    themeplayground.CategoryCustom,
    Version:     "1.0.0",
    Author:      "Your Name",
    Palette: themeplayground.ColorPalette{
        Background:  themeplayground.Color{Hex: "#1a1b26", Name: "Background"},
        Primary:     themeplayground.Color{Hex: "#7aa2f7", Name: "Primary Blue"},
        TextPrimary: themeplayground.Color{Hex: "#c0caf5", Name: "Text"},
        // ... define all required colors
    },
    // ... other fields
}

// Register the theme
err := tm.RegisterTheme(customTheme)
```

### Theme Validation

All themes are automatically validated for:

- Required fields (name, description)
- Valid hex color codes for all palette colors
- Accessibility compliance (if accessibility mode is enabled)

## Error Handling

The theme system provides structured error handling:

```go
type ThemeError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// Predefined errors
var (
    ErrThemeNotFound      = NewThemeError("THEME_NOT_FOUND", "theme not found")
    ErrThemeInvalid       = NewThemeError("THEME_INVALID", "theme validation failed")
    ErrThemeExists        = NewThemeError("THEME_EXISTS", "theme already exists")
    ErrColorInvalid       = NewThemeError("COLOR_INVALID", "invalid color format")
    ErrAccessibilityFail  = NewThemeError("ACCESSIBILITY_FAIL", "accessibility check failed")
    ErrPersistenceFail    = NewThemeError("PERSISTENCE_FAIL", "failed to save/load theme")
    ErrPlaygroundInactive = NewThemeError("PLAYGROUND_INACTIVE", "playground not running")
)
```

## Integration Examples

### With CLI Applications

```go
// Add theme support to your CLI
tm := themeplayground.NewThemeManager("~/.config/myapp")

// Style command output
successStyle := tm.GetStyleFor("status", "completed")
errorStyle := tm.GetStyleFor("status", "failed")

fmt.Print(successStyle.Render(" Operation completed"))
fmt.Print(errorStyle.Render(" Operation failed"))
```

### With TUI Applications

```go
// Integrate with Bubble Tea applications
type model struct {
    themeManager *themeplayground.ThemeManager
    // ... other fields
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "t":
            // Open theme selector
            return m, tea.ExecProcess(themeplayground.RunPlayground(m.themeManager), nil)
        }
    }
    return m, nil
}

func (m model) View() string {
    headerStyle := m.themeManager.GetStyleFor("text", "primary")
    return headerStyle.Render("My Application")
}
```

### With Web Applications

```go
// Generate CSS from themes
theme := tm.GetActiveTheme()
css := fmt.Sprintf(`
:root {
    --color-background: %s;
    --color-primary: %s;
    --color-text: %s;
}
`, theme.Palette.Background.Hex, theme.Palette.Primary.Hex, theme.Palette.TextPrimary.Hex)
```

## Configuration

### Default File Locations

- **Config Directory**: `~/.config/go-redis-work-queue/`
- **Custom Themes**: `~/.config/go-redis-work-queue/themes/`
- **Preferences**: `~/.config/go-redis-work-queue/preferences.json`

### Environment Variables

- `NO_COLOR`: Respects the NO_COLOR environment variable
- `FORCE_COLOR`: Forces color output even in non-TTY environments

## Performance Considerations

- Theme switching is instant (no file I/O during runtime)
- Built-in themes are loaded once at startup
- Custom themes are cached in memory
- Style generation is lazy and cached
- Thread-safe operations with minimal locking

## Testing

The theme system includes comprehensive test coverage:

- Unit tests for all components
- Integration tests for theme switching
- Accessibility validation tests
- Performance benchmarks
- Concurrent access tests

Run tests with:

```bash
go test -v -cover ./internal/theme-playground
```

## Troubleshooting

### Common Issues

1. **Theme not found**: Ensure theme name is spelled correctly and theme is registered
2. **Colors not displaying**: Check terminal color support and NO_COLOR environment variable
3. **Persistence errors**: Verify write permissions for config directory
4. **Accessibility warnings**: Use accessibility checker to validate theme contrast ratios

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
// Enable verbose logging for theme operations
tm.SetDebugMode(true)
```

### Reset to Defaults

```go
// Reset to default theme
err := tm.SetActiveTheme("default")

// Clear all custom themes and preferences
pm.RestoreDefaults()
```

## Contributing

To contribute new built-in themes:

1. Follow the existing theme creation patterns
2. Ensure WCAG AA compliance minimum
3. Include comprehensive color palette
4. Add theme to built-in theme constants
5. Update documentation

## License

Copyright 2025 James Ross. All rights reserved.