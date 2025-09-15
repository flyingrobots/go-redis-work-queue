// Copyright 2025 James Ross
package themeplayground

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PlaygroundModel represents the theme playground UI state
type PlaygroundModel struct {
	themeManager    *ThemeManager
	selectedTheme   string
	previewMode     bool
	showHelp        bool
	table          table.Model
	help           help.Model
	width          int
	height         int
	ready          bool
	err            error
}

// PlaygroundKeyMap defines keyboard controls for the playground
type PlaygroundKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	Preview     key.Binding
	Apply       key.Binding
	Reset       key.Binding
	ToggleHelp  key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() PlaygroundKeyMap {
	return PlaygroundKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("‘/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("“/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select theme"),
		),
		Preview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "preview theme"),
		),
		Apply: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "apply theme"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset to default"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns key bindings to show in the mini help view
func (k PlaygroundKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Preview, k.Apply, k.ToggleHelp, k.Quit}
}

// FullHelp returns key bindings to show in the full help view
func (k PlaygroundKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Preview, k.Apply, k.Reset},
		{k.ToggleHelp, k.Quit},
	}
}

// NewPlaygroundModel creates a new playground model
func NewPlaygroundModel(themeManager *ThemeManager) *PlaygroundModel {
	columns := []table.Column{
		{Title: "Theme", Width: 20},
		{Title: "Category", Width: 15},
		{Title: "Description", Width: 40},
		{Title: "WCAG", Width: 8},
		{Title: "Status", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	m := &PlaygroundModel{
		themeManager:  themeManager,
		selectedTheme: themeManager.GetActiveTheme().Name,
		table:        t,
		help:         help.New(),
	}

	m.updateTable()
	return m
}

// updateTable refreshes the table with current theme data
func (m *PlaygroundModel) updateTable() {
	themes := m.themeManager.ListThemes()
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name < themes[j].Name
	})

	rows := make([]table.Row, len(themes))
	for i, theme := range themes {
		wcagLevel := theme.Accessibility.WCAGLevel
		if wcagLevel == "" {
			wcagLevel = "N/A"
		}

		status := "Available"
		if theme.Name == m.themeManager.GetActiveTheme().Name {
			status = "Active"
		} else if theme.Name == m.selectedTheme && m.previewMode {
			status = "Preview"
		}

		rows[i] = table.Row{
			theme.Name,
			string(theme.Category),
			theme.Description,
			wcagLevel,
			status,
		}
	}

	m.table.SetRows(rows)
}

// Init implements tea.Model
func (m *PlaygroundModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *PlaygroundModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update table dimensions
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 8)

		// Update help dimensions
		m.help.Width = msg.Width

	case tea.KeyMsg:
		if m.table.Focused() {
			switch {
			case key.Matches(msg, DefaultKeyMap().Quit):
				return m, tea.Quit

			case key.Matches(msg, DefaultKeyMap().ToggleHelp):
				m.showHelp = !m.showHelp

			case key.Matches(msg, DefaultKeyMap().Preview):
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					themeName := selectedRow[0]
					m.selectedTheme = themeName
					m.previewMode = true
					m.updateTable()

					// Apply theme temporarily for preview
					originalTheme := m.themeManager.GetActiveTheme().Name
					if err := m.themeManager.SetActiveTheme(themeName); err == nil {
						// Schedule revert after a short delay or user action
						return m, func() tea.Msg {
							return previewThemeMsg{themeName, originalTheme}
						}
					}
				}

			case key.Matches(msg, DefaultKeyMap().Apply):
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					themeName := selectedRow[0]
					if err := m.themeManager.SetActiveTheme(themeName); err != nil {
						m.err = err
					} else {
						m.selectedTheme = themeName
						m.previewMode = false
						m.updateTable()
					}
				}

			case key.Matches(msg, DefaultKeyMap().Reset):
				if err := m.themeManager.SetActiveTheme(ThemeDefault); err != nil {
					m.err = err
				} else {
					m.selectedTheme = ThemeDefault
					m.previewMode = false
					m.updateTable()
				}

			case key.Matches(msg, DefaultKeyMap().Enter):
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					themeName := selectedRow[0]
					if err := m.themeManager.SetActiveTheme(themeName); err != nil {
						m.err = err
					} else {
						m.selectedTheme = themeName
						m.previewMode = false
						m.updateTable()
					}
				}
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// previewThemeMsg represents a theme preview message
type previewThemeMsg struct {
	ThemeName    string
	OriginalTheme string
}

// View implements tea.Model
func (m *PlaygroundModel) View() string {
	if !m.ready {
		return "Loading theme playground..."
	}

	var sections []string

	// Header
	headerStyle := m.themeManager.GetStyleFor("header", "default").
		Bold(true).
		Padding(1, 2).
		Margin(1, 0)

	header := headerStyle.Render("<¨ Theme Playground")
	sections = append(sections, header)

	// Current theme info
	activeTheme := m.themeManager.GetActiveTheme()
	infoStyle := m.themeManager.GetStyleFor("text", "secondary").
		Padding(0, 2)

	info := infoStyle.Render(fmt.Sprintf("Active: %s (%s)",
		activeTheme.Name, activeTheme.Description))
	sections = append(sections, info)

	// Preview indicator
	if m.previewMode {
		previewStyle := m.themeManager.GetStyleFor("status", "running").
			Bold(true).
			Padding(0, 2)
		preview := previewStyle.Render("=A  Preview Mode - Press 'a' to apply or select another theme")
		sections = append(sections, preview)
	}

	// Theme table
	tableStyle := m.themeManager.GetStyleFor("table", "default").
		Padding(1, 2)

	sections = append(sections, tableStyle.Render(m.table.View()))

	// Error display
	if m.err != nil {
		errorStyle := m.themeManager.GetStyleFor("status", "error").
			Bold(true).
			Padding(0, 2)
		sections = append(sections, errorStyle.Render("Error: "+m.err.Error()))
	}

	// Help section
	helpStyle := m.themeManager.GetStyleFor("text", "tertiary").
		Padding(1, 2)

	if m.showHelp {
		sections = append(sections, helpStyle.Render(m.help.View(DefaultKeyMap())))
	} else {
		sections = append(sections, helpStyle.Render(m.help.ShortHelpView(DefaultKeyMap().ShortHelp())))
	}

	// Theme preview section
	if m.previewMode || len(m.table.SelectedRow()) > 0 {
		themeName := m.selectedTheme
		if len(m.table.SelectedRow()) > 0 && !m.previewMode {
			themeName = m.table.SelectedRow()[0]
		}

		preview := m.renderThemePreview(themeName)
		sections = append(sections, preview)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderThemePreview creates a visual preview of a theme
func (m *PlaygroundModel) renderThemePreview(themeName string) string {
	theme, exists := m.themeManager.GetTheme(themeName)
	if !exists {
		return ""
	}

	var preview strings.Builder

	// Preview header
	previewHeaderStyle := m.themeManager.GetStyleFor("text", "primary").
		Bold(true).
		Padding(1, 2)

	preview.WriteString(previewHeaderStyle.Render(fmt.Sprintf("Preview: %s", theme.Name)))
	preview.WriteString("\n")

	// Color palette preview
	paletteStyle := m.themeManager.GetStyleFor("surface", "default").
		Padding(1, 2).
		Margin(0, 2)

	var colorSamples []string

	// Primary colors
	primaryStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Primary.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex)).
		Padding(0, 2)
	colorSamples = append(colorSamples, primaryStyle.Render("Primary"))

	// Secondary colors
	secondaryStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Secondary.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextSecondary.Hex)).
		Padding(0, 2)
	colorSamples = append(colorSamples, secondaryStyle.Render("Secondary"))

	// Status colors
	successStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Success.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex)).
		Padding(0, 2)
	colorSamples = append(colorSamples, successStyle.Render("Success"))

	warningStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Warning.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex)).
		Padding(0, 2)
	colorSamples = append(colorSamples, warningStyle.Render("Warning"))

	errorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Palette.Error.Hex)).
		Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex)).
		Padding(0, 2)
	colorSamples = append(colorSamples, errorStyle.Render("Error"))

	palette := paletteStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, colorSamples...))
	preview.WriteString(palette)
	preview.WriteString("\n")

	// Accessibility info
	accessibilityStyle := m.themeManager.GetStyleFor("text", "tertiary").
		Padding(0, 2)

	accessInfo := fmt.Sprintf("WCAG: %s | Contrast: %.1f | Color Blind Safe: %t",
		theme.Accessibility.WCAGLevel,
		theme.Accessibility.ContrastRatio,
		theme.Accessibility.ColorBlindSafe)

	preview.WriteString(accessibilityStyle.Render(accessInfo))

	return preview.String()
}

// RunPlayground starts the theme playground TUI
func RunPlayground(themeManager *ThemeManager) error {
	model := NewPlaygroundModel(themeManager)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()

	return err
}

// ThemeSelector creates a simple theme selection UI component
func ThemeSelector(themeManager *ThemeManager, compact bool) string {
	themes := themeManager.ListThemes()
	activeTheme := themeManager.GetActiveTheme().Name

	var options []string

	for _, theme := range themes {
		indicator := "Ë"
		if theme.Name == activeTheme {
			indicator = "Ï"
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Palette.TextPrimary.Hex))
		if theme.Name == activeTheme {
			style = style.Bold(true)
		}

		option := fmt.Sprintf("%s %s", indicator, theme.Name)
		if !compact {
			option += fmt.Sprintf(" (%s)", theme.Description)
		}

		options = append(options, style.Render(option))
	}

	if compact {
		return lipgloss.JoinHorizontal(lipgloss.Left, options...)
	}

	return lipgloss.JoinVertical(lipgloss.Left, options...)
}

// QuickThemeToggle provides quick keyboard shortcuts for theme switching
type QuickThemeToggle struct {
	themeManager *ThemeManager
	keyMap       map[string]string // key -> theme name
}

// NewQuickThemeToggle creates a new quick toggle with default key mappings
func NewQuickThemeToggle(themeManager *ThemeManager) *QuickThemeToggle {
	return &QuickThemeToggle{
		themeManager: themeManager,
		keyMap: map[string]string{
			"1": ThemeDefault,
			"2": ThemeTokyoNight,
			"3": ThemeGitHub,
			"4": ThemeOneDark,
			"5": ThemeSolarizedLight,
			"6": ThemeSolarizedDark,
			"7": ThemeDracula,
			"8": ThemeHighContrast,
			"9": ThemeMonochrome,
			"0": ThemeTerminalClassic,
		},
	}
}

// HandleKey processes a key press and switches theme if mapped
func (qt *QuickThemeToggle) HandleKey(key string) bool {
	if themeName, exists := qt.keyMap[key]; exists {
		if err := qt.themeManager.SetActiveTheme(themeName); err == nil {
			return true
		}
	}
	return false
}

// GetKeyHelp returns help text for quick keys
func (qt *QuickThemeToggle) GetKeyHelp() string {
	var help []string
	for key, themeName := range qt.keyMap {
		help = append(help, fmt.Sprintf("%s: %s", key, themeName))
	}
	return strings.Join(help, " | ")
}