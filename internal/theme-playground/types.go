// Copyright 2025 James Ross
package themeplayground

import (
	"time"
)

// Theme represents a complete visual theme for the application
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

// ThemeCategory defines the type/purpose of a theme
type ThemeCategory string

const (
	CategoryStandard      ThemeCategory = "standard"
	CategoryAccessibility ThemeCategory = "accessibility"
	CategorySpecialty     ThemeCategory = "specialty"
	CategoryCustom        ThemeCategory = "custom"
)

// Color represents a theme color with multiple representations
type Color struct {
	Hex         string  `json:"hex"`         // #ff6b9d
	RGB         RGB     `json:"rgb"`         // {255, 107, 157}
	HSL         HSL     `json:"hsl"`         // {338, 100%, 71%}
	Name        string  `json:"name"`        // "Bright Pink"
	Description string  `json:"description"` // "Primary accent color"
}

// RGB represents RGB color values
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// HSL represents HSL color values
type HSL struct {
	H uint16 `json:"h"` // 0-360
	S uint8  `json:"s"` // 0-100
	L uint8  `json:"l"` // 0-100
}

// ColorPalette defines all colors used in a theme
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

	// Border and divider colors
	Border   Color `json:"border"`
	Divider  Color `json:"divider"`
	Focus    Color `json:"focus"`
	Selected Color `json:"selected"`
	Hover    Color `json:"hover"`

	// Status colors for queue states
	StatusPending   Color `json:"status_pending"`
	StatusRunning   Color `json:"status_running"`
	StatusCompleted Color `json:"status_completed"`
	StatusFailed    Color `json:"status_failed"`
	StatusRetrying  Color `json:"status_retrying"`
}

// ComponentStyles defines styling for specific UI components
type ComponentStyles struct {
	Button       ButtonStyles       `json:"button"`
	Table        TableStyles        `json:"table"`
	Modal        ModalStyles        `json:"modal"`
	Input        InputStyles        `json:"input"`
	Navigation   NavigationStyles   `json:"navigation"`
	StatusCard   StatusCardStyles   `json:"status_card"`
	ProgressBar  ProgressBarStyles  `json:"progress_bar"`
	Chart        ChartStyles        `json:"chart"`
	Notification NotificationStyles `json:"notification"`
}

// ButtonStyles defines button appearance
type ButtonStyles struct {
	Primary   ButtonVariant `json:"primary"`
	Secondary ButtonVariant `json:"secondary"`
	Danger    ButtonVariant `json:"danger"`
	Success   ButtonVariant `json:"success"`
	Ghost     ButtonVariant `json:"ghost"`
}

// ButtonVariant defines a specific button style variant
type ButtonVariant struct {
	Background      Color  `json:"background"`
	Text            Color  `json:"text"`
	Border          Color  `json:"border"`
	BackgroundHover Color  `json:"background_hover"`
	TextHover       Color  `json:"text_hover"`
	BorderHover     Color  `json:"border_hover"`
	BorderRadius    string `json:"border_radius"`
	Padding         string `json:"padding"`
}

// TableStyles defines table appearance
type TableStyles struct {
	HeaderBackground Color  `json:"header_background"`
	HeaderText       Color  `json:"header_text"`
	RowBackground    Color  `json:"row_background"`
	RowBackgroundAlt Color  `json:"row_background_alt"`
	RowText          Color  `json:"row_text"`
	Border           Color  `json:"border"`
	SelectedRow      Color  `json:"selected_row"`
	HoverRow         Color  `json:"hover_row"`
	CellPadding      string `json:"cell_padding"`
}

// ModalStyles defines modal dialog appearance
type ModalStyles struct {
	Background    Color  `json:"background"`
	Overlay       Color  `json:"overlay"`
	Border        Color  `json:"border"`
	Shadow        string `json:"shadow"`
	BorderRadius  string `json:"border_radius"`
	MaxWidth      string `json:"max_width"`
	Padding       string `json:"padding"`
}

// InputStyles defines form input appearance
type InputStyles struct {
	Background      Color  `json:"background"`
	Text            Color  `json:"text"`
	Border          Color  `json:"border"`
	BorderFocus     Color  `json:"border_focus"`
	BorderError     Color  `json:"border_error"`
	Placeholder     Color  `json:"placeholder"`
	BorderRadius    string `json:"border_radius"`
	Padding         string `json:"padding"`
}

// NavigationStyles defines navigation appearance
type NavigationStyles struct {
	Background      Color  `json:"background"`
	Text            Color  `json:"text"`
	TextActive      Color  `json:"text_active"`
	TextHover       Color  `json:"text_hover"`
	Border          Color  `json:"border"`
	ActiveIndicator Color  `json:"active_indicator"`
	Padding         string `json:"padding"`
}

// StatusCardStyles defines status card appearance
type StatusCardStyles struct {
	Background   Color  `json:"background"`
	Border       Color  `json:"border"`
	Title        Color  `json:"title"`
	Value        Color  `json:"value"`
	Description  Color  `json:"description"`
	BorderRadius string `json:"border_radius"`
	Padding      string `json:"padding"`
}

// ProgressBarStyles defines progress bar appearance
type ProgressBarStyles struct {
	Background Color  `json:"background"`
	Fill       Color  `json:"fill"`
	Border     Color  `json:"border"`
	Height     string `json:"height"`
	Radius     string `json:"radius"`
}

// ChartStyles defines chart visualization appearance
type ChartStyles struct {
	Background  Color   `json:"background"`
	Grid        Color   `json:"grid"`
	Axis        Color   `json:"axis"`
	Text        Color   `json:"text"`
	DataColors  []Color `json:"data_colors"`
	LineWidth   string  `json:"line_width"`
	PointRadius string  `json:"point_radius"`
}

// NotificationStyles defines notification appearance
type NotificationStyles struct {
	Background   Color  `json:"background"`
	Text         Color  `json:"text"`
	Border       Color  `json:"border"`
	CloseButton  Color  `json:"close_button"`
	BorderRadius string `json:"border_radius"`
	Padding      string `json:"padding"`
}

// Typography defines text styling options
type Typography struct {
	FontFamily      string `json:"font_family"`
	FontSize        string `json:"font_size"`
	LineHeight      string `json:"line_height"`
	LetterSpacing   string `json:"letter_spacing"`
	FontWeight      string `json:"font_weight"`
	MonospaceFamily string `json:"monospace_family"`
}

// AnimationConfig defines animation and transition settings
type AnimationConfig struct {
	Enabled          bool   `json:"enabled"`
	Duration         string `json:"duration"`
	Easing           string `json:"easing"`
	ReducedMotion    bool   `json:"reduced_motion"`
	ThemeTransition  string `json:"theme_transition"`
	HoverTransition  string `json:"hover_transition"`
	FadeTransition   string `json:"fade_transition"`
}

// AccessibilityInfo contains accessibility metadata and validation
type AccessibilityInfo struct {
	ContrastRatio         float64  `json:"contrast_ratio"`
	WCAGLevel            string   `json:"wcag_level"` // AA, AAA
	ColorBlindSafe       bool     `json:"color_blind_safe"`
	MotionSafe           bool     `json:"motion_safe"`
	HighContrast         bool     `json:"high_contrast"`
	LowVisionFriendly    bool     `json:"low_vision_friendly"`
	Warnings             []string `json:"warnings"`
	Recommendations      []string `json:"recommendations"`
	ContrastCheckResults []ContrastCheck `json:"contrast_checks"`
}

// ContrastCheck represents a single contrast validation result
type ContrastCheck struct {
	ForegroundColor string  `json:"foreground_color"`
	BackgroundColor string  `json:"background_color"`
	ComponentName   string  `json:"component_name"`
	Ratio           float64 `json:"ratio"`
	AACompliant     bool    `json:"aa_compliant"`
	AAACompliant    bool    `json:"aaa_compliant"`
	Critical        bool    `json:"critical"`
}

// ThemePreferences stores user theme preferences
type ThemePreferences struct {
	ActiveTheme        string            `json:"active_theme"`
	AutoDetectTerminal bool              `json:"auto_detect_terminal"`
	RespectNoColor     bool              `json:"respect_no_color"`
	SyncWithSystem     bool              `json:"sync_with_system"`
	AccessibilityMode  bool              `json:"accessibility_mode"`
	MotionReduced      bool              `json:"motion_reduced"`
	CustomAccent       *Color            `json:"custom_accent,omitempty"`
	Overrides          map[string]string `json:"overrides"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// ThemeError represents theme-related errors
type ThemeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *ThemeError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewThemeError creates a new theme error
func NewThemeError(code, message string) *ThemeError {
	return &ThemeError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to a theme error
func (e *ThemeError) WithDetails(details string) *ThemeError {
	return &ThemeError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// Error constants
var (
	ErrThemeNotFound      = NewThemeError("THEME_NOT_FOUND", "theme not found")
	ErrThemeInvalid       = NewThemeError("THEME_INVALID", "theme validation failed")
	ErrThemeExists        = NewThemeError("THEME_EXISTS", "theme already exists")
	ErrColorInvalid       = NewThemeError("COLOR_INVALID", "invalid color format")
	ErrAccessibilityFail  = NewThemeError("ACCESSIBILITY_FAIL", "accessibility check failed")
	ErrPersistenceFail    = NewThemeError("PERSISTENCE_FAIL", "failed to save/load theme")
	ErrPlaygroundInactive = NewThemeError("PLAYGROUND_INACTIVE", "playground not running")
)

// Built-in theme constants
const (
	ThemeDefault        = "default"
	ThemeTokyoNight     = "tokyo-night"
	ThemeGitHub         = "github"
	ThemeOneDark        = "one-dark"
	ThemeSolarizedLight = "solarized-light"
	ThemeSolarizedDark  = "solarized-dark"
	ThemeDracula        = "dracula"
	ThemeHighContrast   = "high-contrast"
	ThemeMonochrome     = "monochrome"
	ThemeTerminalClassic = "terminal-classic"
)