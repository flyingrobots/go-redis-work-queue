// Copyright 2025 James Ross
package themeplayground

import (
	"encoding/json"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ThemeCategory represents the type of theme
type ThemeCategory string

const (
	CategoryStandard      ThemeCategory = "standard"
	CategoryAccessibility ThemeCategory = "accessibility"
	CategorySpecialty     ThemeCategory = "specialty"
	CategoryCustom        ThemeCategory = "custom"
)

// Color represents a color with multiple format support
type Color struct {
	Hex        string `json:"hex"`
	RGB        RGB    `json:"rgb"`
	HSL        HSL    `json:"hsl"`
	Adaptive   bool   `json:"adaptive"`   // Whether color adapts to terminal background
	Fallback   string `json:"fallback"`   // Fallback color for limited terminals
	ANSICode   int    `json:"ansi_code"`  // ANSI 256 color code
	TrueColor  bool   `json:"true_color"` // Whether this requires true color support
}

// RGB represents RGB color values
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// HSL represents HSL color values
type HSL struct {
	H float64 `json:"h"` // Hue (0-360)
	S float64 `json:"s"` // Saturation (0-100)
	L float64 `json:"l"` // Lightness (0-100)
}

// ColorPalette defines the core color scheme
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
	Outline  Color `json:"outline"`
	Shadow   Color `json:"shadow"`

	// Status colors for queue management
	StatusRunning   Color `json:"status_running"`
	StatusPending   Color `json:"status_pending"`
	StatusCompleted Color `json:"status_completed"`
	StatusFailed    Color `json:"status_failed"`
	StatusPaused    Color `json:"status_paused"`
	StatusDraining  Color `json:"status_draining"`

	// Chart and data visualization colors
	ChartPrimary   Color   `json:"chart_primary"`
	ChartSecondary Color   `json:"chart_secondary"`
	ChartSeries    []Color `json:"chart_series"`
}

// ComponentStyles defines styling for UI components
type ComponentStyles struct {
	Button     ButtonStyles     `json:"button"`
	Table      TableStyles      `json:"table"`
	Modal      ModalStyles      `json:"modal"`
	Navigation NavigationStyles `json:"navigation"`
	Form       FormStyles       `json:"form"`
	Chart      ChartStyles      `json:"chart"`
	Status     StatusStyles     `json:"status"`
}

// ButtonStyles defines button component styling
type ButtonStyles struct {
	Primary   ComponentVariant `json:"primary"`
	Secondary ComponentVariant `json:"secondary"`
	Danger    ComponentVariant `json:"danger"`
	Ghost     ComponentVariant `json:"ghost"`
	Disabled  ComponentVariant `json:"disabled"`
}

// TableStyles defines table component styling
type TableStyles struct {
	Header     ComponentVariant `json:"header"`
	Row        ComponentVariant `json:"row"`
	RowAlt     ComponentVariant `json:"row_alt"`
	RowHover   ComponentVariant `json:"row_hover"`
	RowSelected ComponentVariant `json:"row_selected"`
	Cell       ComponentVariant `json:"cell"`
	Border     BorderStyle      `json:"border"`
}

// ModalStyles defines modal component styling
type ModalStyles struct {
	Overlay    ComponentVariant `json:"overlay"`
	Container  ComponentVariant `json:"container"`
	Header     ComponentVariant `json:"header"`
	Content    ComponentVariant `json:"content"`
	Footer     ComponentVariant `json:"footer"`
	CloseButton ComponentVariant `json:"close_button"`
}

// NavigationStyles defines navigation component styling
type NavigationStyles struct {
	TabActive    ComponentVariant `json:"tab_active"`
	TabInactive  ComponentVariant `json:"tab_inactive"`
	TabHover     ComponentVariant `json:"tab_hover"`
	TabDisabled  ComponentVariant `json:"tab_disabled"`
	Breadcrumb   ComponentVariant `json:"breadcrumb"`
	Menu         ComponentVariant `json:"menu"`
	MenuItem     ComponentVariant `json:"menu_item"`
	MenuSeparator ComponentVariant `json:"menu_separator"`
}

// FormStyles defines form component styling
type FormStyles struct {
	Input        ComponentVariant `json:"input"`
	InputFocus   ComponentVariant `json:"input_focus"`
	InputError   ComponentVariant `json:"input_error"`
	InputDisabled ComponentVariant `json:"input_disabled"`
	Label        ComponentVariant `json:"label"`
	Placeholder  ComponentVariant `json:"placeholder"`
	HelpText     ComponentVariant `json:"help_text"`
	ErrorText    ComponentVariant `json:"error_text"`
	Checkbox     ComponentVariant `json:"checkbox"`
	Radio        ComponentVariant `json:"radio"`
}

// ChartStyles defines chart component styling
type ChartStyles struct {
	Background ComponentVariant `json:"background"`
	Grid       ComponentVariant `json:"grid"`
	Axis       ComponentVariant `json:"axis"`
	Label      ComponentVariant `json:"label"`
	Legend     ComponentVariant `json:"legend"`
	Tooltip    ComponentVariant `json:"tooltip"`
	SeriesColors []Color        `json:"series_colors"`
}

// StatusStyles defines status indicator styling
type StatusStyles struct {
	Badge      ComponentVariant `json:"badge"`
	Indicator  ComponentVariant `json:"indicator"`
	Progress   ComponentVariant `json:"progress"`
	ProgressBar ComponentVariant `json:"progress_bar"`
	Health     HealthStyles     `json:"health"`
}

// HealthStyles defines health status styling
type HealthStyles struct {
	Healthy   ComponentVariant `json:"healthy"`
	Warning   ComponentVariant `json:"warning"`
	Critical  ComponentVariant `json:"critical"`
	Unknown   ComponentVariant `json:"unknown"`
	Degraded  ComponentVariant `json:"degraded"`
}

// ComponentVariant represents styling for a component variant
type ComponentVariant struct {
	Foreground   Color       `json:"foreground"`
	Background   Color       `json:"background"`
	Border       BorderStyle `json:"border"`
	Padding      Spacing     `json:"padding"`
	Margin       Spacing     `json:"margin"`
	BorderRadius int         `json:"border_radius"`
	FontWeight   string      `json:"font_weight"`
	FontStyle    string      `json:"font_style"`
	TextAlign    string      `json:"text_align"`
	Opacity      float64     `json:"opacity"`
}

// BorderStyle defines border styling
type BorderStyle struct {
	Width int         `json:"width"`
	Style string      `json:"style"` // solid, dashed, dotted
	Color Color       `json:"color"`
	Sides BorderSides `json:"sides"`
}

// BorderSides defines which sides have borders
type BorderSides struct {
	Top    bool `json:"top"`
	Right  bool `json:"right"`
	Bottom bool `json:"bottom"`
	Left   bool `json:"left"`
}

// Spacing defines padding/margin values
type Spacing struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

// Typography defines text styling
type Typography struct {
	FontFamily   string            `json:"font_family"`
	FontSize     int               `json:"font_size"`
	LineHeight   float64           `json:"line_height"`
	LetterSpacing float64          `json:"letter_spacing"`
	Weights      map[string]string `json:"weights"`
	Variants     TypographyVariants `json:"variants"`
}

// TypographyVariants defines different text variants
type TypographyVariants struct {
	H1          TextVariant `json:"h1"`
	H2          TextVariant `json:"h2"`
	H3          TextVariant `json:"h3"`
	H4          TextVariant `json:"h4"`
	Body        TextVariant `json:"body"`
	BodySmall   TextVariant `json:"body_small"`
	Caption     TextVariant `json:"caption"`
	Label       TextVariant `json:"label"`
	Code        TextVariant `json:"code"`
	Monospace   TextVariant `json:"monospace"`
}

// TextVariant defines styling for text elements
type TextVariant struct {
	FontSize      int     `json:"font_size"`
	FontWeight    string  `json:"font_weight"`
	LineHeight    float64 `json:"line_height"`
	LetterSpacing float64 `json:"letter_spacing"`
	TextTransform string  `json:"text_transform"`
	Color         Color   `json:"color"`
}

// AnimationConfig defines animation settings
type AnimationConfig struct {
	Enabled         bool              `json:"enabled"`
	Duration        time.Duration     `json:"duration"`
	Easing          string            `json:"easing"`
	ReducedMotion   bool              `json:"reduced_motion"`
	Transitions     TransitionConfig  `json:"transitions"`
	Hover           HoverConfig       `json:"hover"`
	Focus           FocusConfig       `json:"focus"`
}

// TransitionConfig defines transition animations
type TransitionConfig struct {
	ColorChange    time.Duration `json:"color_change"`
	BackgroundChange time.Duration `json:"background_change"`
	BorderChange   time.Duration `json:"border_change"`
	OpacityChange  time.Duration `json:"opacity_change"`
	Transform      time.Duration `json:"transform"`
}

// HoverConfig defines hover animations
type HoverConfig struct {
	Enabled    bool          `json:"enabled"`
	Duration   time.Duration `json:"duration"`
	Scale      float64       `json:"scale"`
	Brightness float64       `json:"brightness"`
	Opacity    float64       `json:"opacity"`
}

// FocusConfig defines focus animations
type FocusConfig struct {
	Enabled     bool          `json:"enabled"`
	Duration    time.Duration `json:"duration"`
	OutlineWidth int          `json:"outline_width"`
	OutlineColor Color        `json:"outline_color"`
	Shadow      ShadowConfig  `json:"shadow"`
}

// ShadowConfig defines shadow effects
type ShadowConfig struct {
	Enabled bool  `json:"enabled"`
	X       int   `json:"x"`
	Y       int   `json:"y"`
	Blur    int   `json:"blur"`
	Spread  int   `json:"spread"`
	Color   Color `json:"color"`
}

// AccessibilityInfo contains accessibility validation data
type AccessibilityInfo struct {
	ContrastRatio    float64             `json:"contrast_ratio"`
	WCAGLevel        string              `json:"wcag_level"` // AA, AAA
	ColorBlindSafe   bool                `json:"color_blind_safe"`
	MotionSafe       bool                `json:"motion_safe"`
	HighContrast     bool                `json:"high_contrast"`
	ReducedMotion    bool                `json:"reduced_motion"`
	FocusVisible     bool                `json:"focus_visible"`
	Warnings         []AccessibilityWarning `json:"warnings"`
	ContrastChecks   []ContrastCheck     `json:"contrast_checks"`
}

// AccessibilityWarning represents an accessibility concern
type AccessibilityWarning struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"` // low, medium, high, critical
	Message     string  `json:"message"`
	Component   string  `json:"component"`
	Suggestion  string  `json:"suggestion"`
	WCAGRef     string  `json:"wcag_ref"`
	ContrastRatio float64 `json:"contrast_ratio,omitempty"`
}

// ContrastCheck represents a contrast ratio validation
type ContrastCheck struct {
	Foreground    string  `json:"foreground"`
	Background    string  `json:"background"`
	Ratio         float64 `json:"ratio"`
	Level         string  `json:"level"` // AAA, AA, A, FAIL
	Component     string  `json:"component"`
	Context       string  `json:"context"`
	Required      float64 `json:"required"`
	Passed        bool    `json:"passed"`
}

// Theme represents a complete theme definition
type Theme struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Author          string            `json:"author"`
	License         string            `json:"license"`
	Category        ThemeCategory     `json:"category"`
	Tags            []string          `json:"tags"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Accessibility   AccessibilityInfo `json:"accessibility"`
	Palette         ColorPalette      `json:"palette"`
	Components      ComponentStyles   `json:"components"`
	Typography      Typography        `json:"typography"`
	Animations      AnimationConfig   `json:"animations"`
	Metadata        ThemeMetadata     `json:"metadata"`
	Preview         ThemePreview      `json:"preview"`
}

// ThemeMetadata contains additional theme information
type ThemeMetadata struct {
	BasedOn         string            `json:"based_on,omitempty"`
	Terminal        TerminalSupport   `json:"terminal"`
	Environment     []string          `json:"environment"`
	Screenshots     []string          `json:"screenshots"`
	Documentation   string            `json:"documentation"`
	Repository      string            `json:"repository"`
	Homepage        string            `json:"homepage"`
	Keywords        []string          `json:"keywords"`
	Compatibility   CompatibilityInfo `json:"compatibility"`
}

// TerminalSupport defines terminal compatibility
type TerminalSupport struct {
	TrueColor      bool     `json:"true_color"`
	Color256       bool     `json:"color_256"`
	Color16        bool     `json:"color_16"`
	Monochrome     bool     `json:"monochrome"`
	Tested         []string `json:"tested"`
	Recommended    []string `json:"recommended"`
	NotSupported   []string `json:"not_supported"`
}

// CompatibilityInfo defines version compatibility
type CompatibilityInfo struct {
	MinVersion     string   `json:"min_version"`
	MaxVersion     string   `json:"max_version,omitempty"`
	BreakingChanges []string `json:"breaking_changes,omitempty"`
	Deprecated     bool     `json:"deprecated"`
	Replacement    string   `json:"replacement,omitempty"`
}

// ThemePreview contains preview data for theme selection
type ThemePreview struct {
	Colors        []Color         `json:"colors"`
	Components    []ComponentDemo `json:"components"`
	Screenshot    string          `json:"screenshot"`
	Thumbnail     string          `json:"thumbnail"`
	Examples      []PreviewExample `json:"examples"`
}

// ComponentDemo represents a component preview
type ComponentDemo struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Props       map[string]interface{} `json:"props"`
	Content     string            `json:"content"`
	Style       lipgloss.Style    `json:"-"`
	Screenshot  string            `json:"screenshot"`
}

// PreviewExample shows theme in context
type PreviewExample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Context     string `json:"context"`
	Screenshot  string `json:"screenshot"`
}

// ThemeRegistry manages available themes
type ThemeRegistry struct {
	Builtin     map[string]*Theme         `json:"builtin"`
	Custom      map[string]*Theme         `json:"custom"`
	Collections map[string]*ThemeCollection `json:"collections"`
	Index       ThemeIndex               `json:"index"`
}

// ThemeCollection groups related themes
type ThemeCollection struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Themes      []string  `json:"themes"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	Featured    bool      `json:"featured"`
}

// ThemeIndex provides search and discovery
type ThemeIndex struct {
	Categories map[ThemeCategory][]string `json:"categories"`
	Tags       map[string][]string        `json:"tags"`
	Authors    map[string][]string        `json:"authors"`
	Popular    []string                   `json:"popular"`
	Recent     []string                   `json:"recent"`
	Featured   []string                   `json:"featured"`
}

// UserPreferences stores user theme preferences
type UserPreferences struct {
	ActiveTheme        string            `json:"active_theme"`
	AutoDetectTerminal bool              `json:"auto_detect_terminal"`
	RespectNoColor     bool              `json:"respect_no_color"`
	SyncWithSystem     bool              `json:"sync_with_system"`
	AccessibilityMode  bool              `json:"accessibility_mode"`
	ReducedMotion      bool              `json:"reduced_motion"`
	HighContrast       bool              `json:"high_contrast"`
	CustomThemesDir    string            `json:"custom_themes_dir"`
	Favorites          []string          `json:"favorites"`
	History            []ThemeHistoryEntry `json:"history"`
	Customizations     map[string]interface{} `json:"customizations"`
}

// ThemeHistoryEntry tracks theme usage
type ThemeHistoryEntry struct {
	Theme       string        `json:"theme"`
	AppliedAt   time.Time     `json:"applied_at"`
	Duration    time.Duration `json:"duration"`
	Context     string        `json:"context"`
	Automatic   bool          `json:"automatic"`
}

// ThemeConfig contains theme system configuration
type ThemeConfig struct {
	Version        string          `json:"version"`
	UserPreferences UserPreferences `json:"user_preferences"`
	CustomThemes   []*Theme        `json:"custom_themes"`
	Collections    []ThemeCollection `json:"collections"`
	Cache          ThemeCacheConfig `json:"cache"`
	Accessibility  AccessibilityConfig `json:"accessibility"`
}

// ThemeCacheConfig defines caching behavior
type ThemeCacheConfig struct {
	Enabled      bool          `json:"enabled"`
	TTL          time.Duration `json:"ttl"`
	MaxSize      int           `json:"max_size"`
	PreloadThemes []string     `json:"preload_themes"`
	Compression  bool          `json:"compression"`
}

// AccessibilityConfig defines accessibility settings
type AccessibilityConfig struct {
	Enabled            bool    `json:"enabled"`
	EnforceContrast    bool    `json:"enforce_contrast"`
	MinContrastRatio   float64 `json:"min_contrast_ratio"`
	HighContrastMode   bool    `json:"high_contrast_mode"`
	ReducedMotionMode  bool    `json:"reduced_motion_mode"`
	FocusIndicators    bool    `json:"focus_indicators"`
	ColorBlindnessMode string  `json:"color_blindness_mode"`
	VoiceAnnouncements bool    `json:"voice_announcements"`
}

// LipGlossStyle wraps lipgloss.Style with theme awareness
type LipGlossStyle struct {
	lipgloss.Style
	theme     *Theme
	component string
	variant   string
	context   map[string]interface{}
}

// StyleBuilder helps build theme-aware styles
type StyleBuilder struct {
	theme     *Theme
	component string
	variant   string
	modifiers []StyleModifier
	context   map[string]interface{}
}

// StyleModifier represents a style modification function
type StyleModifier func(*LipGlossStyle) *LipGlossStyle

// ValidationResult contains theme validation results
type ValidationResult struct {
	Valid       bool                    `json:"valid"`
	Errors      []ValidationError       `json:"errors"`
	Warnings    []ValidationWarning     `json:"warnings"`
	Suggestions []ValidationSuggestion  `json:"suggestions"`
	Score       ValidationScore         `json:"score"`
}

// ValidationError represents a theme validation error
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Context string `json:"context"`
}

// ValidationWarning represents a theme validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Impact  string `json:"impact"`
}

// ValidationSuggestion represents a theme improvement suggestion
type ValidationSuggestion struct {
	Field       string `json:"field"`
	Current     string `json:"current"`
	Suggested   string `json:"suggested"`
	Reason      string `json:"reason"`
	Improvement string `json:"improvement"`
}

// ValidationScore contains theme quality metrics
type ValidationScore struct {
	Overall        float64 `json:"overall"`
	Accessibility  float64 `json:"accessibility"`
	Consistency    float64 `json:"consistency"`
	Performance    float64 `json:"performance"`
	Completeness   float64 `json:"completeness"`
	Usability      float64 `json:"usability"`
}

// Implementing the Color interface for lipgloss compatibility
func (c Color) String() string {
	return c.Hex
}

// ToLipgloss converts Color to lipgloss.Color
func (c Color) ToLipgloss() lipgloss.Color {
	if c.TrueColor {
		return lipgloss.Color(c.Hex)
	}
	if c.ANSICode > 0 {
		return lipgloss.Color(string(rune(c.ANSICode)))
	}
	return lipgloss.Color(c.Fallback)
}

// MarshalJSON implements custom JSON marshaling for Duration fields
func (ac AnimationConfig) MarshalJSON() ([]byte, error) {
	type Alias AnimationConfig
	return json.Marshal(&struct {
		Duration string `json:"duration"`
		*Alias
	}{
		Duration: ac.Duration.String(),
		Alias:    (*Alias)(&ac),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Duration fields
func (ac *AnimationConfig) UnmarshalJSON(data []byte) error {
	type Alias AnimationConfig
	aux := &struct {
		Duration string `json:"duration"`
		*Alias
	}{
		Alias: (*Alias)(ac),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Duration != "" {
		duration, err := time.ParseDuration(aux.Duration)
		if err != nil {
			return err
		}
		ac.Duration = duration
	}
	return nil
}