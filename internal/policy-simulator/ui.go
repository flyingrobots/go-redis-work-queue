// Copyright 2025 James Ross
package policysimulator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PolicySimulatorUI provides an interactive interface for the policy simulator
type PolicySimulatorUI struct {
	simulator      *PolicySimulator
	width          int
	height         int
	currentTab     TabType
	policyInputs   map[string]textinput.Model
	trafficInputs  map[string]textinput.Model
	help           help.Model
	activeInput    string
	currentPolicy  *PolicyConfig
	currentPattern *TrafficPattern
	lastResult     *SimulationResult
	charts         []ChartData
	warnings       []string
	statusMessage  string
}

// TabType represents different UI tabs
type TabType int

const (
	TabPolicies TabType = iota
	TabTraffic
	TabSimulation
	TabResults
	TabCharts
)

// KeyMap defines keyboard controls for the UI
type KeyMap struct {
	Tab        key.Binding
	PrevTab    key.Binding
	Enter      key.Binding
	Simulate   key.Binding
	Reset      key.Binding
	Help       key.Binding
	Quit       key.Binding
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
}

// DefaultKeyMap returns default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous tab"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm/select"),
		),
		Simulate: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "run simulation"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset to defaults"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "decrease value"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "increase value"),
		),
	}
}

// ShortHelp returns key bindings for the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Simulate, k.Reset, k.Help, k.Quit}
}

// FullHelp returns key bindings for the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.PrevTab, k.Enter},
		{k.Up, k.Down, k.Left, k.Right},
		{k.Simulate, k.Reset, k.Help, k.Quit},
	}
}

// NewPolicySimulatorUI creates a new policy simulator UI
func NewPolicySimulatorUI(simulator *PolicySimulator) *PolicySimulatorUI {
	ui := &PolicySimulatorUI{
		simulator:     simulator,
		currentTab:    TabPolicies,
		policyInputs:  make(map[string]textinput.Model),
		trafficInputs: make(map[string]textinput.Model),
		help:          help.New(),
		currentPolicy: DefaultPolicyConfig(),
		currentPattern: DefaultTrafficPattern(),
		charts:        make([]ChartData, 0),
		warnings:      make([]string, 0),
	}

	ui.initializeInputs()
	return ui
}

// DefaultTrafficPattern returns a default traffic pattern
func DefaultTrafficPattern() *TrafficPattern {
	return &TrafficPattern{
		Name:     "Steady Load",
		Type:     TrafficConstant,
		BaseRate: 50.0, // 50 messages per second
		Duration: 5 * time.Minute,
		Variations: []TrafficVariation{
			{
				StartTime:   2 * time.Minute,
				EndTime:     3 * time.Minute,
				Multiplier:  2.0,
				Description: "2x spike for 1 minute",
			},
		},
		Probability: 1.0,
		Metadata:    make(map[string]interface{}),
	}
}

// initializeInputs sets up input fields for policy and traffic configuration
func (ui *PolicySimulatorUI) initializeInputs() {
	// Policy input fields
	policyFields := []struct {
		key         string
		placeholder string
		value       string
	}{
		{"max_retries", "Maximum retries", "3"},
		{"initial_backoff", "Initial backoff (seconds)", "1"},
		{"max_backoff", "Maximum backoff (seconds)", "30"},
		{"max_rate", "Max rate (msg/sec)", "100"},
		{"max_concurrency", "Max concurrency", "5"},
		{"queue_size", "Queue size", "1000"},
		{"processing_timeout", "Processing timeout (seconds)", "30"},
	}

	for _, field := range policyFields {
		input := textinput.New()
		input.Placeholder = field.placeholder
		input.SetValue(field.value)
		input.CharLimit = 20
		ui.policyInputs[field.key] = input
	}

	// Traffic pattern input fields
	trafficFields := []struct {
		key         string
		placeholder string
		value       string
	}{
		{"base_rate", "Base rate (msg/sec)", "50"},
		{"spike_multiplier", "Spike multiplier", "2.0"},
		{"spike_start", "Spike start (minutes)", "2"},
		{"spike_duration", "Spike duration (minutes)", "1"},
		{"simulation_duration", "Simulation duration (minutes)", "5"},
	}

	for _, field := range trafficFields {
		input := textinput.New()
		input.Placeholder = field.placeholder
		input.SetValue(field.value)
		input.CharLimit = 20
		ui.trafficInputs[field.key] = input
	}

	// Focus first input
	if len(ui.policyInputs) > 0 {
		ui.activeInput = "max_retries"
		input := ui.policyInputs[ui.activeInput]
		input.Focus()
		ui.policyInputs[ui.activeInput] = input
	}
}

// Init implements tea.Model
func (ui *PolicySimulatorUI) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (ui *PolicySimulatorUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.width = msg.Width
		ui.height = msg.Height
		ui.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap().Quit):
			return ui, tea.Quit

		case key.Matches(msg, DefaultKeyMap().Help):
			ui.help.ShowAll = !ui.help.ShowAll

		case key.Matches(msg, DefaultKeyMap().Tab):
			ui.nextTab()

		case key.Matches(msg, DefaultKeyMap().PrevTab):
			ui.prevTab()

		case key.Matches(msg, DefaultKeyMap().Simulate):
			if ui.currentTab == TabSimulation {
				return ui, ui.runSimulation()
			}

		case key.Matches(msg, DefaultKeyMap().Reset):
			ui.resetToDefaults()

		case key.Matches(msg, DefaultKeyMap().Enter):
			if ui.currentTab == TabPolicies || ui.currentTab == TabTraffic {
				ui.nextInput()
			}

		case key.Matches(msg, DefaultKeyMap().Up):
			if ui.currentTab == TabPolicies || ui.currentTab == TabTraffic {
				ui.prevInput()
			}

		case key.Matches(msg, DefaultKeyMap().Down):
			if ui.currentTab == TabPolicies || ui.currentTab == TabTraffic {
				ui.nextInput()
			}

		default:
			// Handle input updates
			if ui.currentTab == TabPolicies && ui.activeInput != "" {
				if input, exists := ui.policyInputs[ui.activeInput]; exists {
					input, cmd = input.Update(msg)
					ui.policyInputs[ui.activeInput] = input
					cmds = append(cmds, cmd)
				}
			} else if ui.currentTab == TabTraffic && ui.activeInput != "" {
				if input, exists := ui.trafficInputs[ui.activeInput]; exists {
					input, cmd = input.Update(msg)
					ui.trafficInputs[ui.activeInput] = input
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return ui, tea.Batch(cmds...)
}

// View implements tea.Model
func (ui *PolicySimulatorUI) View() string {
	var sections []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1, 2)

	header := headerStyle.Render("🎛️  Policy Simulator")
	sections = append(sections, header)

	// Tab navigation
	tabs := ui.renderTabs()
	sections = append(sections, tabs)

	// Content based on current tab
	content := ui.renderTabContent()
	sections = append(sections, content)

	// Status message
	if ui.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Padding(0, 2)
		sections = append(sections, statusStyle.Render(ui.statusMessage))
	}

	// Help
	helpView := ui.help.View(DefaultKeyMap())
	if helpView != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 2)
		sections = append(sections, helpStyle.Render(helpView))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTabs creates the tab navigation bar
func (ui *PolicySimulatorUI) renderTabs() string {
	tabs := []string{"Policies", "Traffic", "Simulation", "Results", "Charts"}
	var renderedTabs []string

	for i, tab := range tabs {
		style := lipgloss.NewStyle().Padding(0, 2)
		if TabType(i) == ui.currentTab {
			style = style.Background(lipgloss.Color("12")).Foreground(lipgloss.Color("0"))
		}
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	tabStyle := lipgloss.NewStyle().Padding(1, 2)
	return tabStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
}

// renderTabContent renders content based on the current tab
func (ui *PolicySimulatorUI) renderTabContent() string {
	switch ui.currentTab {
	case TabPolicies:
		return ui.renderPoliciesTab()
	case TabTraffic:
		return ui.renderTrafficTab()
	case TabSimulation:
		return ui.renderSimulationTab()
	case TabResults:
		return ui.renderResultsTab()
	case TabCharts:
		return ui.renderChartsTab()
	default:
		return "Unknown tab"
	}
}

// renderPoliciesTab renders the policies configuration tab
func (ui *PolicySimulatorUI) renderPoliciesTab() string {
	var sections []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(1, 2)

	sections = append(sections, titleStyle.Render("Policy Configuration"))

	contentStyle := lipgloss.NewStyle().Padding(0, 2)

	// Render input fields
	fields := []string{"max_retries", "initial_backoff", "max_backoff", "max_rate", "max_concurrency", "queue_size", "processing_timeout"}

	for _, field := range fields {
		if input, exists := ui.policyInputs[field]; exists {
			labelStyle := lipgloss.NewStyle().Width(25)
			inputStyle := lipgloss.NewStyle().Width(20)

			label := labelStyle.Render(strings.ReplaceAll(field, "_", " ") + ":")
			inputView := inputStyle.Render(input.View())

			focused := ""
			if field == ui.activeInput {
				focused = " ← "
			}

			fieldView := lipgloss.JoinHorizontal(lipgloss.Left, label, inputView, focused)
			sections = append(sections, contentStyle.Render(fieldView))
		}
	}

	// Current policy summary
	if ui.currentPolicy != nil {
		ui.updatePolicyFromInputs()
		summary := ui.renderPolicySummary()
		sections = append(sections, summary)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTrafficTab renders the traffic pattern configuration tab
func (ui *PolicySimulatorUI) renderTrafficTab() string {
	var sections []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(1, 2)

	sections = append(sections, titleStyle.Render("Traffic Pattern Configuration"))

	contentStyle := lipgloss.NewStyle().Padding(0, 2)

	// Render input fields
	fields := []string{"base_rate", "spike_multiplier", "spike_start", "spike_duration", "simulation_duration"}

	for _, field := range fields {
		if input, exists := ui.trafficInputs[field]; exists {
			labelStyle := lipgloss.NewStyle().Width(25)
			inputStyle := lipgloss.NewStyle().Width(20)

			label := labelStyle.Render(strings.ReplaceAll(field, "_", " ") + ":")
			inputView := inputStyle.Render(input.View())

			focused := ""
			if field == ui.activeInput {
				focused = " ← "
			}

			fieldView := lipgloss.JoinHorizontal(lipgloss.Left, label, inputView, focused)
			sections = append(sections, contentStyle.Render(fieldView))
		}
	}

	// Traffic pattern preview
	if ui.currentPattern != nil {
		ui.updateTrafficFromInputs()
		preview := ui.renderTrafficPreview()
		sections = append(sections, preview)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderSimulationTab renders the simulation execution tab
func (ui *PolicySimulatorUI) renderSimulationTab() string {
	var sections []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(1, 2)

	sections = append(sections, titleStyle.Render("Run Simulation"))

	contentStyle := lipgloss.NewStyle().Padding(0, 2)

	// Simulation overview
	overview := "Ready to simulate policy impact with current configuration.\n\n"
	overview += "This will predict:\n"
	overview += "• Queue depth and wait times\n"
	overview += "• Throughput and utilization\n"
	overview += "• Failure and retry rates\n"
	overview += "• Resource usage estimates\n\n"
	overview += "Press 's' to start simulation"

	sections = append(sections, contentStyle.Render(overview))

	// Configuration summary
	summary := ui.renderConfigSummary()
	sections = append(sections, summary)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderResultsTab renders the simulation results tab
func (ui *PolicySimulatorUI) renderResultsTab() string {
	var sections []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(1, 2)

	sections = append(sections, titleStyle.Render("Simulation Results"))

	if ui.lastResult == nil {
		contentStyle := lipgloss.NewStyle().Padding(0, 2)
		sections = append(sections, contentStyle.Render("No simulation results available. Run a simulation first."))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Render results
	results := ui.renderSimulationResults()
	sections = append(sections, results)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderChartsTab renders the charts visualization tab
func (ui *PolicySimulatorUI) renderChartsTab() string {
	var sections []string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(1, 2)

	sections = append(sections, titleStyle.Render("Performance Charts"))

	if len(ui.charts) == 0 {
		contentStyle := lipgloss.NewStyle().Padding(0, 2)
		sections = append(sections, contentStyle.Render("No chart data available. Run a simulation first."))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Render charts (ASCII representations)
	charts := ui.renderCharts()
	sections = append(sections, charts)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Helper methods for navigation and updates
func (ui *PolicySimulatorUI) nextTab() {
	ui.currentTab = TabType((int(ui.currentTab) + 1) % 5)
	ui.updateActiveInput()
}

func (ui *PolicySimulatorUI) prevTab() {
	ui.currentTab = TabType((int(ui.currentTab) + 4) % 5)
	ui.updateActiveInput()
}

func (ui *PolicySimulatorUI) updateActiveInput() {
	// Clear focus from all inputs
	for key, input := range ui.policyInputs {
		input.Blur()
		ui.policyInputs[key] = input
	}
	for key, input := range ui.trafficInputs {
		input.Blur()
		ui.trafficInputs[key] = input
	}

	// Focus appropriate input based on current tab
	if ui.currentTab == TabPolicies {
		ui.activeInput = "max_retries"
		if input, exists := ui.policyInputs[ui.activeInput]; exists {
			input.Focus()
			ui.policyInputs[ui.activeInput] = input
		}
	} else if ui.currentTab == TabTraffic {
		ui.activeInput = "base_rate"
		if input, exists := ui.trafficInputs[ui.activeInput]; exists {
			input.Focus()
			ui.trafficInputs[ui.activeInput] = input
		}
	} else {
		ui.activeInput = ""
	}
}

func (ui *PolicySimulatorUI) nextInput() {
	if ui.currentTab == TabPolicies {
		fields := []string{"max_retries", "initial_backoff", "max_backoff", "max_rate", "max_concurrency", "queue_size", "processing_timeout"}
		ui.switchInput(fields, ui.policyInputs, 1)
	} else if ui.currentTab == TabTraffic {
		fields := []string{"base_rate", "spike_multiplier", "spike_start", "spike_duration", "simulation_duration"}
		ui.switchInput(fields, ui.trafficInputs, 1)
	}
}

func (ui *PolicySimulatorUI) prevInput() {
	if ui.currentTab == TabPolicies {
		fields := []string{"max_retries", "initial_backoff", "max_backoff", "max_rate", "max_concurrency", "queue_size", "processing_timeout"}
		ui.switchInput(fields, ui.policyInputs, -1)
	} else if ui.currentTab == TabTraffic {
		fields := []string{"base_rate", "spike_multiplier", "spike_start", "spike_duration", "simulation_duration"}
		ui.switchInput(fields, ui.trafficInputs, -1)
	}
}

func (ui *PolicySimulatorUI) switchInput(fields []string, inputs map[string]textinput.Model, direction int) {
	currentIndex := 0
	for i, field := range fields {
		if field == ui.activeInput {
			currentIndex = i
			break
		}
	}

	// Clear current focus
	if input, exists := inputs[ui.activeInput]; exists {
		input.Blur()
		inputs[ui.activeInput] = input
	}

	// Calculate next index
	nextIndex := (currentIndex + direction + len(fields)) % len(fields)
	ui.activeInput = fields[nextIndex]

	// Focus new input
	if input, exists := inputs[ui.activeInput]; exists {
		input.Focus()
		inputs[ui.activeInput] = input
	}
}

func (ui *PolicySimulatorUI) resetToDefaults() {
	ui.currentPolicy = DefaultPolicyConfig()
	ui.currentPattern = DefaultTrafficPattern()
	ui.initializeInputs()
	ui.statusMessage = "Reset to default configuration"
}

// runSimulation executes a simulation with current configuration
func (ui *PolicySimulatorUI) runSimulation() tea.Cmd {
	return func() tea.Msg {
		ui.updatePolicyFromInputs()
		ui.updateTrafficFromInputs()

		request := &SimulationRequest{
			Name:           "UI Simulation",
			Description:    "Simulation run from interactive UI",
			Policies:       ui.currentPolicy,
			TrafficPattern: ui.currentPattern,
		}

		result, err := ui.simulator.RunSimulation(nil, request)
		if err != nil {
			return simulationErrorMsg{err}
		}

		// Wait for simulation to complete (simplified for demo)
		time.Sleep(2 * time.Second)

		// Get updated result
		completedResult, err := ui.simulator.GetSimulation(result.ID)
		if err != nil {
			return simulationErrorMsg{err}
		}

		return simulationCompleteMsg{completedResult}
	}
}

// Message types for simulation events
type simulationCompleteMsg struct {
	result *SimulationResult
}

type simulationErrorMsg struct {
	err error
}

// Update policy configuration from input values
func (ui *PolicySimulatorUI) updatePolicyFromInputs() {
	if ui.currentPolicy == nil {
		ui.currentPolicy = DefaultPolicyConfig()
	}

	// Parse input values and update policy
	if value := ui.policyInputs["max_retries"].Value(); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			ui.currentPolicy.MaxRetries = intVal
		}
	}

	// Add other field parsing...
	// (Simplified for brevity)
}

// Update traffic pattern from input values
func (ui *PolicySimulatorUI) updateTrafficFromInputs() {
	if ui.currentPattern == nil {
		ui.currentPattern = DefaultTrafficPattern()
	}

	// Parse input values and update pattern
	if value := ui.trafficInputs["base_rate"].Value(); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			ui.currentPattern.BaseRate = floatVal
		}
	}

	// Add other field parsing...
	// (Simplified for brevity)
}

// Render methods for different sections
func (ui *PolicySimulatorUI) renderPolicySummary() string {
	// Implementation for policy summary
	return "Policy Summary: " + fmt.Sprintf("Max Retries: %d, Max Concurrency: %d",
		ui.currentPolicy.MaxRetries, ui.currentPolicy.MaxConcurrency)
}

func (ui *PolicySimulatorUI) renderTrafficPreview() string {
	// Implementation for traffic preview
	return fmt.Sprintf("Traffic Preview: %.1f msg/sec base rate", ui.currentPattern.BaseRate)
}

func (ui *PolicySimulatorUI) renderConfigSummary() string {
	// Implementation for configuration summary
	return "Configuration ready for simulation"
}

func (ui *PolicySimulatorUI) renderSimulationResults() string {
	if ui.lastResult == nil || ui.lastResult.Metrics == nil {
		return "No results available"
	}

	// Render detailed results
	metrics := ui.lastResult.Metrics
	results := fmt.Sprintf(`
Simulation Results:
• Average Queue Depth: %.1f
• Maximum Queue Depth: %d
• Average Wait Time: %.1f ms
• Processing Rate: %.1f msg/sec
• Utilization: %.1f%%
• Failure Rate: %.1f%%
`,
		metrics.AvgQueueDepth,
		metrics.MaxQueueDepth,
		metrics.AvgWaitTime,
		metrics.ProcessingRate,
		metrics.Utilization,
		metrics.FailureRate)

	return results
}

func (ui *PolicySimulatorUI) renderCharts() string {
	// Simple ASCII chart representation
	return "Charts: [Queue Depth] [Throughput] [Latency] (ASCII visualization)"
}

// RunPolicySimulatorUI starts the interactive policy simulator UI
func RunPolicySimulatorUI(simulator *PolicySimulator) error {
	ui := NewPolicySimulatorUI(simulator)

	program := tea.NewProgram(ui, tea.WithAltScreen())
	_, err := program.Run()

	return err
}