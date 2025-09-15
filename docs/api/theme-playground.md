# Theme Playground API Documentation

The Theme Playground provides a comprehensive theming system for the Go Redis Work Queue application, including built-in themes, custom theme creation, accessibility validation, and live theme switching.

## Overview

The Theme Playground system consists of several key components:

- **ThemeManager**: Core theme management and registry
- **ThemeIntegration**: Integration with the main application
- **PlaygroundHandler**: HTTP API endpoints for theme operations
- **AccessibilityChecker**: WCAG compliance validation
- **ColorUtilities**: Color manipulation and contrast calculations

## Core Types

### Theme

Represents a complete visual theme for the application.

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

### ColorPalette

Defines all colors used in a theme, including semantic and status colors.

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

### Color

Represents a color with multiple format representations.

```go
type Color struct {
    Hex         string `json:"hex"`         // #ff6b9d
    RGB         RGB    `json:"rgb"`         // {255, 107, 157}
    HSL         HSL    `json:"hsl"`         // {338, 100%, 71%}
    Name        string `json:"name"`        // "Bright Pink"
    Description string `json:"description"` // "Primary accent color"
}
```

## ThemeManager API

### Creating a ThemeManager

```go
func NewThemeManager(configDir string) *ThemeManager
```

Creates a new theme manager with the specified configuration directory. The manager automatically:
- Loads built-in themes
- Loads user preferences from disk
- Loads custom themes from the themes directory
- Sets a default active theme

### Core Methods

#### GetActiveTheme
```go
func (tm *ThemeManager) GetActiveTheme() *Theme
```
Returns the currently active theme. Thread-safe.

#### SetActiveTheme
```go
func (tm *ThemeManager) SetActiveTheme(name string) error
```
Sets the active theme by name. Updates preferences and notifies callbacks.

#### GetTheme
```go
func (tm *ThemeManager) GetTheme(name string) (*Theme, error)
```
Retrieves a specific theme by name.

#### ListThemes
```go
func (tm *ThemeManager) ListThemes() []Theme
```
Returns all available themes, sorted by name.

#### RegisterTheme
```go
func (tm *ThemeManager) RegisterTheme(theme *Theme) error
```
Registers a new theme in the registry. Validates the theme before registration.

#### ValidateTheme
```go
func (tm *ThemeManager) ValidateTheme(theme *Theme) error
```
Validates a theme for correctness and accessibility compliance.

#### SaveTheme
```go
func (tm *ThemeManager) SaveTheme(theme *Theme) error
```
Saves a theme to disk as a JSON file in the themes directory.

#### GetStyleFor
```go
func (tm *ThemeManager) GetStyleFor(component, variant string) lipgloss.Style
```
Returns a Lip Gloss style for a specific component and variant.

**Supported Components:**
- `button` (variants: primary, secondary, danger, success, ghost)
- `table` (variants: header, row, row_alt, selected)
- `modal`
- `input` (variants: default, focus, error)
- `navigation` (variants: default, active, hover)
- `status_card`
- `progress_bar`
- `notification`

#### OnThemeChange
```go
func (tm *ThemeManager) OnThemeChange(callback func(*Theme))
```
Registers a callback function that is called when the active theme changes.

## Built-in Themes

The system includes several built-in themes:

- **default**: Clean light theme with professional styling
- **tokyo-night**: Dark theme inspired by Tokyo's neon-lit nights
- **github**: Clean light theme matching GitHub's interface
- **one-dark**: Popular dark theme from Atom editor
- **high-contrast**: High contrast theme for accessibility
- **monochrome**: Pure black and white theme for minimal terminals

## Theme Integration API

### Creating Integration

```go
func NewThemeIntegration(configDir string) (*ThemeIntegration, error)
```

Creates a new theme integration instance that:
- Sets up a theme manager
- Detects terminal capabilities
- Provides style caching
- Handles theme change notifications

### Style Helper

The integration provides a style helper for common styling operations:

```go
func NewStyleHelper(integration *ThemeIntegration) *StyleHelper
```

#### Button Styling
```go
func (sh *StyleHelper) Button(text, variant string) string
```

#### Table Styling
```go
func (sh *StyleHelper) Table() *TableStyleHelper
```

Table helper methods:
- `Header(text string) string`
- `Row(text string, isAlternate bool) string`
- `SelectedRow(text string) string`

#### Status Styling
```go
func (sh *StyleHelper) Status() *StatusStyleHelper
```

Status helper methods:
- `StatusBadge(status string) string`
- `ProgressBar(current, total int, width int) string`

## HTTP API Endpoints

The PlaygroundHandler provides REST API endpoints for theme operations:

### GET /api/themes
Returns all available themes with active theme information.

**Response:**
```json
{
  "themes": [
    {
      "name": "default",
      "description": "Default light theme",
      "category": "standard",
      // ... theme details
    }
  ],
  "active": "default"
}
```

### GET /api/themes/{name}
Returns a specific theme by name.

### POST /api/themes/active
Sets the active theme.

**Request Body:**
```json
{
  "theme": "tokyo-night"
}
```

### GET /api/themes/{name}/preview
Returns theme preview information.

**Response:**
```json
{
  "name": "tokyo-night",
  "description": "Dark theme inspired by Tokyo's neon-lit nights",
  "palette": {
    "background": "#1a1b26",
    "primary": "#7aa2f7",
    // ... color mappings
  },
  "accessibility": {
    "contrast_ratio": 15.2,
    "wcag_level": "AAA",
    "color_blind_safe": true
  }
}
```

### POST /api/themes/validate
Validates a theme configuration.

**Request Body:**
```json
{
  "name": "custom-theme",
  "description": "My custom theme",
  "palette": {
    // ... color definitions
  }
}
```

### POST /api/themes/save
Saves a custom theme.

### GET /api/themes/{name}/export
Exports a theme as a downloadable JSON file.

### POST /api/themes/import
Imports a theme from an uploaded JSON file.

### GET /api/preferences
Returns user theme preferences.

### POST /api/preferences
Updates user theme preferences.

### GET /api/config-path
Returns configuration directory paths.

## Color Utilities API

### ColorUtilities

Provides color manipulation and validation utilities.

```go
func NewColorUtilities() *ColorUtilities
```

#### Color Conversion
```go
func (cu *ColorUtilities) HexToRGB(hex string) (*RGB, error)
func (cu *ColorUtilities) RGBToHex(rgb RGB) string
func (cu *ColorUtilities) RGBToHSL(rgb RGB) (*HSL, error)
```

#### Contrast Calculation
```go
func (cu *ColorUtilities) ContrastRatio(color1, color2 Color) (float64, error)
```

Calculates the WCAG contrast ratio between two colors.

## Accessibility Checker API

### AccessibilityChecker

Validates themes for accessibility compliance.

```go
func NewAccessibilityChecker() *AccessibilityChecker
```

#### CheckAccessibility
```go
func (ac *AccessibilityChecker) CheckAccessibility(theme *Theme) (*AccessibilityInfo, error)
```

Performs comprehensive accessibility validation and returns:

```go
type AccessibilityInfo struct {
    ContrastRatio         float64         `json:"contrast_ratio"`
    WCAGLevel            string          `json:"wcag_level"`     // AA, AAA, Fail
    ColorBlindSafe       bool            `json:"color_blind_safe"`
    MotionSafe           bool            `json:"motion_safe"`
    HighContrast         bool            `json:"high_contrast"`
    LowVisionFriendly    bool            `json:"low_vision_friendly"`
    Warnings             []string        `json:"warnings"`
    Recommendations      []string        `json:"recommendations"`
    ContrastCheckResults []ContrastCheck `json:"contrast_checks"`
}
```

## Theme Configuration

### Configuration Directory Structure

```
$XDG_CONFIG_HOME/go-redis-wq/
├── theme_preferences.json      # User preferences
└── themes/                     # Custom themes directory
    ├── custom-theme-1.json
    ├── custom-theme-2.json
    └── backups/               # Theme backups
        └── custom-theme-1_20231201_143022.json.bak
```

### Preferences File Format

```json
{
  "active_theme": "tokyo-night",
  "auto_detect_terminal": true,
  "respect_no_color": true,
  "sync_with_system": false,
  "accessibility_mode": false,
  "motion_reduced": false,
  "custom_accent": {
    "hex": "#0066cc",
    "name": "Custom Blue"
  },
  "overrides": {
    "primary": "#custom-color"
  },
  "created_at": "2023-12-01T14:30:22Z",
  "updated_at": "2023-12-01T14:30:22Z"
}
```

## Environment Variables

The theme system respects the following environment variables:

- `NO_COLOR`: When set, forces monochrome theme
- `COLORTERM`: Used for terminal capability detection
- `TERM`: Used for color support detection
- `TERM_PROGRAM`: Used for terminal-specific optimizations

## Error Handling

The theme system uses custom error types for better error handling:

### ThemeError
```go
type ThemeError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### Common Error Codes

- `THEME_NOT_FOUND`: Theme does not exist
- `THEME_INVALID`: Theme validation failed
- `THEME_EXISTS`: Theme already exists
- `COLOR_INVALID`: Invalid color format
- `ACCESSIBILITY_FAIL`: Accessibility check failed
- `PERSISTENCE_FAIL`: Failed to save/load theme

## Usage Examples

### Basic Theme Management

```go
// Create theme manager
tm := NewThemeManager("/path/to/config")

// List available themes
themes := tm.ListThemes()
fmt.Printf("Available themes: %d\n", len(themes))

// Switch to dark theme
err := tm.SetActiveTheme("tokyo-night")
if err != nil {
    log.Printf("Failed to set theme: %v", err)
}

// Get current theme colors
activeTheme := tm.GetActiveTheme()
bgColor := activeTheme.Palette.Background.Hex
textColor := activeTheme.Palette.TextPrimary.Hex
```

### Creating Custom Themes

```go
// Create custom theme
customTheme := &Theme{
    Name:        "my-theme",
    Description: "My custom theme",
    Category:    CategoryCustom,
    Version:     "1.0.0",
    Author:      "Me",
    Palette: ColorPalette{
        Background:  Color{Hex: "#1e1e1e", Name: "Dark Gray"},
        TextPrimary: Color{Hex: "#ffffff", Name: "White"},
        Primary:     Color{Hex: "#007acc", Name: "Blue"},
        // ... other colors
    },
}

// Validate and register
err := tm.ValidateTheme(customTheme)
if err != nil {
    log.Printf("Theme validation failed: %v", err)
    return
}

err = tm.RegisterTheme(customTheme)
if err != nil {
    log.Printf("Failed to register theme: %v", err)
    return
}

// Save to disk
err = tm.SaveTheme(customTheme)
if err != nil {
    log.Printf("Failed to save theme: %v", err)
}
```

### Using Styles

```go
// Create integration
integration, err := NewThemeIntegration("/path/to/config")
if err != nil {
    log.Fatal(err)
}

// Get style helper
styleHelper := NewStyleHelper(integration)

// Style buttons
primaryBtn := styleHelper.Button("Submit", "primary")
secondaryBtn := styleHelper.Button("Cancel", "secondary")

// Style table
table := styleHelper.Table()
header := table.Header("Name")
row := table.Row("Value", false)
selectedRow := table.SelectedRow("Selected Value")

// Style status
status := styleHelper.Status()
badge := status.StatusBadge("completed")
progress := status.ProgressBar(75, 100, 20)
```

### Accessibility Checking

```go
// Check theme accessibility
ac := NewAccessibilityChecker()
info, err := ac.CheckAccessibility(customTheme)
if err != nil {
    log.Printf("Accessibility check failed: %v", err)
    return
}

fmt.Printf("WCAG Level: %s\n", info.WCAGLevel)
fmt.Printf("Contrast Ratio: %.2f:1\n", info.ContrastRatio)
fmt.Printf("Color Blind Safe: %t\n", info.ColorBlindSafe)

if len(info.Warnings) > 0 {
    fmt.Println("Warnings:")
    for _, warning := range info.Warnings {
        fmt.Printf("  - %s\n", warning)
    }
}
```

## Performance Considerations

- Styles are cached automatically by the integration layer
- Theme switching triggers cache invalidation
- Color calculations are optimized for repeated use
- File I/O is minimized through intelligent caching
- Concurrent access is protected with read-write locks

## Thread Safety

All public APIs are thread-safe:
- ThemeManager uses read-write mutexes for safe concurrent access
- Integration layer handles concurrent style requests
- Callback notifications are executed safely

## Migration Guide

When upgrading or migrating theme configurations:

1. Backup existing configuration directory
2. Update theme file format if needed (automatic migration supported)
3. Validate custom themes against new schema
4. Update integration code for new style helpers if applicable

For more detailed examples and advanced usage, see the test files in the theme-playground package.