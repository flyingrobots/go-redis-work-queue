package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color Palette - Professional Terminal Colors
var (
	// Base Colors
	colorPrimary      = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}
	colorSecondary    = lipgloss.AdaptiveColor{Light: "#6f42c1", Dark: "#d2a8ff"}
	colorSuccess      = lipgloss.AdaptiveColor{Light: "#1a7f37", Dark: "#56d364"}
	colorWarning      = lipgloss.AdaptiveColor{Light: "#bf8700", Dark: "#f9e71e"}
	colorError        = lipgloss.AdaptiveColor{Light: "#cf222e", Dark: "#f85149"}
	colorInfo         = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#79c0ff"}

	// Background Colors
	colorBgPrimary    = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"}
	colorBgSecondary  = lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#21262d"}
	colorBgAccent     = lipgloss.AdaptiveColor{Light: "#e1f5fe", Dark: "#161b22"}
	colorBgMuted      = lipgloss.AdaptiveColor{Light: "#f0f6fc", Dark: "#1c2128"}

	// Border Colors
	colorBorderPrimary   = lipgloss.AdaptiveColor{Light: "#d0d7de", Dark: "#30363d"}
	colorBorderSecondary = lipgloss.AdaptiveColor{Light: "#afb8c1", Dark: "#21262d"}
	colorBorderAccent    = lipgloss.AdaptiveColor{Light: "#58a6ff", Dark: "#58a6ff"}
	colorBorderMuted     = lipgloss.AdaptiveColor{Light: "#f0f6fc", Dark: "#1c2128"}

	// Text Colors
	colorTextPrimary   = lipgloss.AdaptiveColor{Light: "#24292f", Dark: "#f0f6fc"}
	colorTextSecondary = lipgloss.AdaptiveColor{Light: "#656d76", Dark: "#8b949e"}
	colorTextMuted     = lipgloss.AdaptiveColor{Light: "#8c959f", Dark: "#6e7681"}
	colorTextInverse   = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#21262d"}
)

// Responsive Styles - Adapt to terminal width
type StyleSet struct {
	// Headers and Titles
	AppTitle    lipgloss.Style
	SectionTitle lipgloss.Style
	SubTitle    lipgloss.Style

	// Content Containers
	Panel       lipgloss.Style
	PanelActive lipgloss.Style
	Card        lipgloss.Style
	Modal       lipgloss.Style

	// Status and State Indicators
	StatusSuccess lipgloss.Style
	StatusWarning lipgloss.Style
	StatusError   lipgloss.Style
	StatusInfo    lipgloss.Style
	StatusMuted   lipgloss.Style

	// Interactive Elements
	Button        lipgloss.Style
	ButtonActive  lipgloss.Style
	ButtonDanger  lipgloss.Style
	TabActive     lipgloss.Style
	TabInactive   lipgloss.Style

	// Layout Elements
	Separator     lipgloss.Style
	Progress      lipgloss.Style
	Loading       lipgloss.Style

	// Terminal Breakpoint Info
	Breakpoint string
	Width      int
	Height     int
}

// GetStyleSet returns an appropriate style set for the terminal size
func GetStyleSet(width, height int) StyleSet {
	var breakpoint string

	// Determine breakpoint
	switch {
	case width <= 40:
		breakpoint = "mobile"
	case width <= 80:
		breakpoint = "tablet"
	case width <= 120:
		breakpoint = "desktop"
	default:
		breakpoint = "ultrawide"
	}

	// Base styles that work everywhere
	base := StyleSet{
		Breakpoint: breakpoint,
		Width:      width,
		Height:     height,

		// Headers adapt to space
		AppTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(colorBgSecondary).
			Padding(0, 1).
			Margin(0, 0, 1, 0),

		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTextPrimary).
			Padding(0, 1).
			Background(colorBgAccent).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(colorBorderAccent),

		SubTitle: lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Italic(true),

		// Status indicators with consistent colors
		StatusSuccess: lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true),

		StatusWarning: lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true),

		StatusError: lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true),

		StatusInfo: lipgloss.NewStyle().
			Foreground(colorInfo).
			Bold(true),

		StatusMuted: lipgloss.NewStyle().
			Foreground(colorTextMuted),

		// Interactive elements
		Button: lipgloss.NewStyle().
			Foreground(colorTextInverse).
			Background(colorPrimary).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true),

		ButtonActive: lipgloss.NewStyle().
			Foreground(colorTextInverse).
			Background(colorSuccess).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorSuccess),

		ButtonDanger: lipgloss.NewStyle().
			Foreground(colorTextInverse).
			Background(colorError).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true),

		// Loading and progress
		Progress: lipgloss.NewStyle().
			Foreground(colorPrimary).
			Background(colorBgMuted),

		Loading: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true),

		Separator: lipgloss.NewStyle().
			Foreground(colorBorderMuted).
			Background(colorBgMuted),
	}

	// Adapt styles based on breakpoint
	switch breakpoint {
	case "mobile":
		// Mobile: Minimal chrome, larger touch targets
		base.Panel = lipgloss.NewStyle().
			Background(colorBgSecondary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderPrimary).
			Padding(1).
			Margin(1, 0)

		base.PanelActive = base.Panel.Copy().
			BorderForeground(colorBorderAccent).
			Background(colorBgAccent)

		base.Card = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderSecondary).
			Padding(1).
			Margin(0, 0, 1, 0).
			Width(width - 4) // Full width minus margins

		base.TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(colorBgAccent).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(colorBorderAccent)

		base.TabInactive = lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Background(colorBgMuted).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(colorBorderMuted)

	case "tablet":
		// Tablet: Two-column layouts, moderate spacing
		base.Panel = lipgloss.NewStyle().
			Background(colorBgSecondary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderPrimary).
			Padding(1, 2).
			Margin(0, 1)

		base.PanelActive = base.Panel.Copy().
			BorderForeground(colorBorderAccent).
			Background(colorBgAccent).
			Border(lipgloss.DoubleBorder())

		base.Card = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorderSecondary).
			Padding(1, 2).
			Margin(0, 1, 1, 0).
			Width((width / 2) - 3) // Half width for two columns

		base.TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(colorBgAccent).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorBorderAccent)

		base.TabInactive = lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Background(colorBgMuted).
			Padding(0, 2)

	case "desktop":
		// Desktop: Multi-panel, rich visuals
		base.Panel = lipgloss.NewStyle().
			Background(colorBgSecondary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderPrimary).
			Padding(1, 2).
			Margin(0, 1)

		base.PanelActive = base.Panel.Copy().
			BorderForeground(colorBorderAccent).
			Background(colorBgAccent).
			Border(lipgloss.ThickBorder())

		base.Card = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorderSecondary).
			Padding(2).
			Margin(1)

		base.Modal = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorBorderAccent).
			Padding(2, 4).
			Margin(2)

		base.TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(colorBgAccent).
			Padding(0, 3).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorBorderAccent)

		base.TabInactive = lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Background(colorBgMuted).
			Padding(0, 3).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorBorderMuted)

	case "ultrawide":
		// Ultrawide: Maximum feature density
		base.Panel = lipgloss.NewStyle().
			Background(colorBgSecondary).
			Border(lipgloss.ThickBorder()).
			BorderForeground(colorBorderPrimary).
			Padding(2, 3).
			Margin(1)

		base.PanelActive = base.Panel.Copy().
			BorderForeground(colorBorderAccent).
			Background(colorBgAccent).
			Border(lipgloss.DoubleBorder())

		base.Card = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorderSecondary).
			Padding(2, 3).
			Margin(1, 2)

		base.Modal = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorBorderAccent).
			Padding(3, 6).
			Margin(3).
			Width(width / 2) // Centered modal

		base.AppTitle = base.AppTitle.Copy().
			Padding(1, 3).
			Margin(0, 0, 2, 0).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(colorBorderAccent)

		base.TabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Background(colorBgAccent).
			Padding(1, 4).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(colorBorderAccent)

		base.TabInactive = lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Background(colorBgMuted).
			Padding(1, 4).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(colorBorderMuted)
	}

	return base
}

// Utility functions for creating consistent visual elements

// CreateHeader creates a professional header with branding
func (s StyleSet) CreateHeader(title, subtitle string) string {
	headerContent := s.AppTitle.Render("🔄 " + title)
	if subtitle != "" {
		headerContent += "\n" + s.SubTitle.Render(subtitle)
	}
	return headerContent
}

// CreateStatusBadge creates colored status indicators
func (s StyleSet) CreateStatusBadge(text, status string) string {
	var style lipgloss.Style
	var icon string

	switch status {
	case "success", "active", "healthy":
		style = s.StatusSuccess
		icon = "✓"
	case "warning", "degraded", "slow":
		style = s.StatusWarning
		icon = "⚠"
	case "error", "failed", "down":
		style = s.StatusError
		icon = "✗"
	case "info", "pending", "loading":
		style = s.StatusInfo
		icon = "ℹ"
	default:
		style = s.StatusMuted
		icon = "○"
	}

	return style.Render(icon + " " + text)
}

// CreateProgressBar creates visual progress indicators
func (s StyleSet) CreateProgressBar(current, max int, width int) string {
	if max == 0 {
		return s.Progress.Width(width).Render("")
	}

	percentage := float64(current) / float64(max)
	filled := int(percentage * float64(width))

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return s.Progress.Render(bar)
}

// CreateInfoCard creates consistent information cards
func (s StyleSet) CreateInfoCard(title, content string) string {
	cardTitle := s.SectionTitle.Render(title)
	cardContent := s.Card.Render(cardTitle + "\n" + content)
	return cardContent
}

// CreateButtonBar creates consistent button layouts
func (s StyleSet) CreateButtonBar(buttons []string, active int) string {
	var renderedButtons []string

	for i, button := range buttons {
		if i == active {
			renderedButtons = append(renderedButtons, s.ButtonActive.Render(button))
		} else {
			renderedButtons = append(renderedButtons, s.Button.Render(button))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, renderedButtons...)
}

// CreateMetricDisplay creates consistent metric displays
func (s StyleSet) CreateMetricDisplay(label, value, trend string) string {
	valueStyle := s.StatusInfo
	trendIcon := ""

	switch trend {
	case "up":
		trendIcon = "↗"
		valueStyle = s.StatusSuccess
	case "down":
		trendIcon = "↘"
		valueStyle = s.StatusError
	case "flat":
		trendIcon = "→"
		valueStyle = s.StatusMuted
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		s.StatusMuted.Render(label+":"),
		" ",
		valueStyle.Render(value),
		" ",
		s.StatusMuted.Render(trendIcon),
	)
}

// CreateLoadingSpinner creates animated loading indicators
func (s StyleSet) CreateLoadingSpinner(text string) string {
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏" // Braille spinner
	// In real implementation, you'd cycle through these characters
	return s.Loading.Render("⠋ " + text)
}