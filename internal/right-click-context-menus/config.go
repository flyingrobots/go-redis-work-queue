package context_menus

import (
	"time"
)

// Config holds configuration for the context menu system
type Config struct {
	// Enable or disable the context menu system
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Animation settings
	Animation AnimationConfig `yaml:"animation" json:"animation"`

	// Menu appearance
	Appearance AppearanceConfig `yaml:"appearance" json:"appearance"`

	// Behavior settings
	Behavior BehaviorConfig `yaml:"behavior" json:"behavior"`
}

// AnimationConfig controls menu animations
type AnimationConfig struct {
	// Enable menu animations
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Duration for menu show/hide animations
	Duration time.Duration `yaml:"duration" json:"duration"`

	// Easing function for animations
	Easing string `yaml:"easing" json:"easing"`
}

// AppearanceConfig controls menu visual appearance
type AppearanceConfig struct {
	// Border style for menus
	BorderStyle string `yaml:"border_style" json:"border_style"`

	// Background color
	BackgroundColor string `yaml:"background_color" json:"background_color"`

	// Text color
	TextColor string `yaml:"text_color" json:"text_color"`

	// Selected item highlight color
	HighlightColor string `yaml:"highlight_color" json:"highlight_color"`

	// Destructive action color
	DestructiveColor string `yaml:"destructive_color" json:"destructive_color"`

	// Accelerator key color
	AcceleratorColor string `yaml:"accelerator_color" json:"accelerator_color"`

	// Shadow settings
	Shadow ShadowConfig `yaml:"shadow" json:"shadow"`

	// Minimum menu width
	MinWidth int `yaml:"min_width" json:"min_width"`

	// Maximum menu width
	MaxWidth int `yaml:"max_width" json:"max_width"`
}

// ShadowConfig controls menu shadow appearance
type ShadowConfig struct {
	// Enable shadows
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Shadow color
	Color string `yaml:"color" json:"color"`

	// Shadow offset
	OffsetX int `yaml:"offset_x" json:"offset_x"`
	OffsetY int `yaml:"offset_y" json:"offset_y"`

	// Shadow blur radius
	Blur int `yaml:"blur" json:"blur"`
}

// BehaviorConfig controls menu behavior
type BehaviorConfig struct {
	// Auto-hide menu after this duration (0 = never)
	AutoHideTimeout time.Duration `yaml:"auto_hide_timeout" json:"auto_hide_timeout"`

	// Require confirmation for destructive actions
	ConfirmDestructive bool `yaml:"confirm_destructive" json:"confirm_destructive"`

	// Close menu after action execution
	CloseAfterAction bool `yaml:"close_after_action" json:"close_after_action"`

	// Mouse settings
	Mouse MouseConfig `yaml:"mouse" json:"mouse"`

	// Keyboard settings
	Keyboard KeyboardConfig `yaml:"keyboard" json:"keyboard"`
}

// MouseConfig controls mouse interaction behavior
type MouseConfig struct {
	// Enable right-click context menus
	RightClickEnabled bool `yaml:"right_click_enabled" json:"right_click_enabled"`

	// Enable left-click menu selection
	LeftClickEnabled bool `yaml:"left_click_enabled" json:"left_click_enabled"`

	// Double-click timeout
	DoubleClickTimeout time.Duration `yaml:"double_click_timeout" json:"double_click_timeout"`

	// Hide menu on click outside
	HideOnClickOutside bool `yaml:"hide_on_click_outside" json:"hide_on_click_outside"`
}

// KeyboardConfig controls keyboard interaction behavior
type KeyboardConfig struct {
	// Enable 'm' key to open context menu
	MKeyEnabled bool `yaml:"m_key_enabled" json:"m_key_enabled"`

	// Enable arrow key navigation
	ArrowNavigation bool `yaml:"arrow_navigation" json:"arrow_navigation"`

	// Enable vim-style navigation (hjkl)
	VimNavigation bool `yaml:"vim_navigation" json:"vim_navigation"`

	// Enable accelerator keys
	AcceleratorKeys bool `yaml:"accelerator_keys" json:"accelerator_keys"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Animation: AnimationConfig{
			Enabled:  true,
			Duration: 150 * time.Millisecond,
			Easing:   "ease-out",
		},
		Appearance: AppearanceConfig{
			BorderStyle:      "rounded",
			BackgroundColor:  "235",
			TextColor:        "250",
			HighlightColor:   "62",
			DestructiveColor: "196",
			AcceleratorColor: "244",
			Shadow: ShadowConfig{
				Enabled: true,
				Color:   "0",
				OffsetX: 1,
				OffsetY: 1,
				Blur:    2,
			},
			MinWidth: 20,
			MaxWidth: 60,
		},
		Behavior: BehaviorConfig{
			AutoHideTimeout:    0, // Never auto-hide
			ConfirmDestructive: true,
			CloseAfterAction:   true,
			Mouse: MouseConfig{
				RightClickEnabled:   true,
				LeftClickEnabled:    true,
				DoubleClickTimeout:  500 * time.Millisecond,
				HideOnClickOutside:  true,
			},
			Keyboard: KeyboardConfig{
				MKeyEnabled:     true,
				ArrowNavigation: true,
				VimNavigation:   true,
				AcceleratorKeys: true,
			},
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Appearance.MinWidth < 10 {
		return &ContextMenuError{
			Message: "minimum width must be at least 10",
			Code:    "INVALID_MIN_WIDTH",
		}
	}

	if c.Appearance.MaxWidth < c.Appearance.MinWidth {
		return &ContextMenuError{
			Message: "maximum width must be greater than minimum width",
			Code:    "INVALID_MAX_WIDTH",
		}
	}

	if c.Animation.Duration < 0 {
		return &ContextMenuError{
			Message: "animation duration cannot be negative",
			Code:    "INVALID_DURATION",
		}
	}

	if c.Behavior.AutoHideTimeout < 0 {
		return &ContextMenuError{
			Message: "auto-hide timeout cannot be negative",
			Code:    "INVALID_TIMEOUT",
		}
	}

	return nil
}

// Apply applies the configuration to a context menu system
func (c *Config) Apply(cms *ContextMenuSystem) error {
	if err := c.Validate(); err != nil {
		return err
	}

	// Enable/disable the system
	cms.SetEnabled(c.Enabled)

	// Additional configuration application would go here
	// This is a placeholder for future configuration options

	return nil
}

// Merge merges another configuration into this one
func (c *Config) Merge(other Config) {
	if other.Enabled != c.Enabled {
		c.Enabled = other.Enabled
	}

	// Merge animation config
	if other.Animation.Enabled != c.Animation.Enabled {
		c.Animation.Enabled = other.Animation.Enabled
	}
	if other.Animation.Duration != 0 {
		c.Animation.Duration = other.Animation.Duration
	}
	if other.Animation.Easing != "" {
		c.Animation.Easing = other.Animation.Easing
	}

	// Merge appearance config
	if other.Appearance.BorderStyle != "" {
		c.Appearance.BorderStyle = other.Appearance.BorderStyle
	}
	if other.Appearance.BackgroundColor != "" {
		c.Appearance.BackgroundColor = other.Appearance.BackgroundColor
	}
	if other.Appearance.TextColor != "" {
		c.Appearance.TextColor = other.Appearance.TextColor
	}
	if other.Appearance.HighlightColor != "" {
		c.Appearance.HighlightColor = other.Appearance.HighlightColor
	}
	if other.Appearance.DestructiveColor != "" {
		c.Appearance.DestructiveColor = other.Appearance.DestructiveColor
	}
	if other.Appearance.AcceleratorColor != "" {
		c.Appearance.AcceleratorColor = other.Appearance.AcceleratorColor
	}
	if other.Appearance.MinWidth > 0 {
		c.Appearance.MinWidth = other.Appearance.MinWidth
	}
	if other.Appearance.MaxWidth > 0 {
		c.Appearance.MaxWidth = other.Appearance.MaxWidth
	}

	// Merge shadow config
	if other.Appearance.Shadow.Enabled != c.Appearance.Shadow.Enabled {
		c.Appearance.Shadow.Enabled = other.Appearance.Shadow.Enabled
	}
	if other.Appearance.Shadow.Color != "" {
		c.Appearance.Shadow.Color = other.Appearance.Shadow.Color
	}
	if other.Appearance.Shadow.OffsetX != 0 {
		c.Appearance.Shadow.OffsetX = other.Appearance.Shadow.OffsetX
	}
	if other.Appearance.Shadow.OffsetY != 0 {
		c.Appearance.Shadow.OffsetY = other.Appearance.Shadow.OffsetY
	}
	if other.Appearance.Shadow.Blur > 0 {
		c.Appearance.Shadow.Blur = other.Appearance.Shadow.Blur
	}

	// Merge behavior config
	if other.Behavior.AutoHideTimeout != 0 {
		c.Behavior.AutoHideTimeout = other.Behavior.AutoHideTimeout
	}
	if other.Behavior.ConfirmDestructive != c.Behavior.ConfirmDestructive {
		c.Behavior.ConfirmDestructive = other.Behavior.ConfirmDestructive
	}
	if other.Behavior.CloseAfterAction != c.Behavior.CloseAfterAction {
		c.Behavior.CloseAfterAction = other.Behavior.CloseAfterAction
	}

	// Merge mouse config
	if other.Behavior.Mouse.RightClickEnabled != c.Behavior.Mouse.RightClickEnabled {
		c.Behavior.Mouse.RightClickEnabled = other.Behavior.Mouse.RightClickEnabled
	}
	if other.Behavior.Mouse.LeftClickEnabled != c.Behavior.Mouse.LeftClickEnabled {
		c.Behavior.Mouse.LeftClickEnabled = other.Behavior.Mouse.LeftClickEnabled
	}
	if other.Behavior.Mouse.DoubleClickTimeout != 0 {
		c.Behavior.Mouse.DoubleClickTimeout = other.Behavior.Mouse.DoubleClickTimeout
	}
	if other.Behavior.Mouse.HideOnClickOutside != c.Behavior.Mouse.HideOnClickOutside {
		c.Behavior.Mouse.HideOnClickOutside = other.Behavior.Mouse.HideOnClickOutside
	}

	// Merge keyboard config
	if other.Behavior.Keyboard.MKeyEnabled != c.Behavior.Keyboard.MKeyEnabled {
		c.Behavior.Keyboard.MKeyEnabled = other.Behavior.Keyboard.MKeyEnabled
	}
	if other.Behavior.Keyboard.ArrowNavigation != c.Behavior.Keyboard.ArrowNavigation {
		c.Behavior.Keyboard.ArrowNavigation = other.Behavior.Keyboard.ArrowNavigation
	}
	if other.Behavior.Keyboard.VimNavigation != c.Behavior.Keyboard.VimNavigation {
		c.Behavior.Keyboard.VimNavigation = other.Behavior.Keyboard.VimNavigation
	}
	if other.Behavior.Keyboard.AcceleratorKeys != c.Behavior.Keyboard.AcceleratorKeys {
		c.Behavior.Keyboard.AcceleratorKeys = other.Behavior.Keyboard.AcceleratorKeys
	}
}