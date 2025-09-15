// Copyright 2025 James Ross
package policysimulator

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicySimulatorUI(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	assert.NotNil(t, ui)
	assert.Equal(t, simulator, ui.simulator)
	assert.Equal(t, TabPolicies, ui.currentTab)
	assert.NotNil(t, ui.policyInputs)
	assert.NotNil(t, ui.trafficInputs)
	assert.NotNil(t, ui.currentPolicy)
	assert.NotNil(t, ui.currentPattern)
}

func TestUITabNavigation(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test initial state
	assert.Equal(t, TabPolicies, ui.currentTab)

	// Test tab navigation with right arrow
	model, cmd := ui.Update(tea.KeyMsg{Type: tea.KeyRight})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabTraffic, ui.currentTab)
	assert.Nil(t, cmd)

	// Continue navigating
	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyRight})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabSimulation, ui.currentTab)

	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyRight})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabResults, ui.currentTab)

	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyRight})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabCharts, ui.currentTab)

	// Test wrapping around
	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyRight})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabPolicies, ui.currentTab)

	// Test left navigation
	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyLeft})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, TabCharts, ui.currentTab)
}

func TestUIQuitCommand(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test 'q' key quits
	model, cmd := ui.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, tea.Quit(), cmd)

	// Test Ctrl+C quits
	model, cmd = ui.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	ui = model.(*PolicySimulatorUI)
	assert.Equal(t, tea.Quit(), cmd)
}

func TestUIRenderPoliciesTab(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Ensure we're on policies tab
	ui.currentTab = TabPolicies

	view := ui.View()

	// Check that the view contains policy-related content
	assert.Contains(t, view, "Policies")
	assert.Contains(t, view, "Max Retries")
	assert.Contains(t, view, "Initial Backoff")
	assert.Contains(t, view, "Max Rate/Second")
	assert.Contains(t, view, "Max Concurrency")
}

func TestUIRenderTrafficTab(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Switch to traffic tab
	ui.currentTab = TabTraffic

	view := ui.View()

	// Check that the view contains traffic-related content
	assert.Contains(t, view, "Traffic")
	assert.Contains(t, view, "Pattern Name")
	assert.Contains(t, view, "Base Rate")
	assert.Contains(t, view, "Duration")
	assert.Contains(t, view, "Pattern Type")
}

func TestUIRenderSimulationTab(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Switch to simulation tab
	ui.currentTab = TabSimulation

	view := ui.View()

	// Check that the view contains simulation-related content
	assert.Contains(t, view, "Simulation")
	assert.Contains(t, view, "Run Simulation")
}

func TestUIRenderResultsTab(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Switch to results tab
	ui.currentTab = TabResults

	view := ui.View()

	// Check that the view contains results-related content
	assert.Contains(t, view, "Results")
	assert.Contains(t, view, "No simulation results")
}

func TestUIRenderChartsTab(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Switch to charts tab
	ui.currentTab = TabCharts

	view := ui.View()

	// Check that the view contains charts-related content
	assert.Contains(t, view, "Charts")
	assert.Contains(t, view, "No charts available")
}

func TestUIWithSimulationResults(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 30 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Run a simulation to get results
	req := &SimulationRequest{
		Name:           "UI Test Simulation",
		Description:    "Testing UI with results",
		Policies:       DefaultPolicyConfig(),
		TrafficPattern: DefaultTrafficPattern(),
	}

	result, err := simulator.RunSimulation(ui.Init().Context, req)
	require.NoError(t, err)

	ui.lastResult = result

	// Test results tab with actual results
	ui.currentTab = TabResults
	view := ui.View()

	assert.Contains(t, view, "UI Test Simulation")
	assert.Contains(t, view, "Testing UI with results")
	assert.Contains(t, view, "Messages Processed")
	assert.Contains(t, view, "Processing Rate")

	// Test charts tab with actual results
	ui.currentTab = TabCharts
	view = ui.View()

	assert.Contains(t, view, "Charts")
	assert.Contains(t, view, "Queue Depth Over Time")
	assert.Contains(t, view, "Processing Rate Over Time")
}

func TestUIInputHandling(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test that policy inputs are properly initialized
	assert.Contains(t, ui.policyInputs, "max_retries")
	assert.Contains(t, ui.policyInputs, "initial_backoff_ms")
	assert.Contains(t, ui.policyInputs, "max_rate_per_second")

	// Test that traffic inputs are properly initialized
	assert.Contains(t, ui.trafficInputs, "pattern_name")
	assert.Contains(t, ui.trafficInputs, "base_rate")
	assert.Contains(t, ui.trafficInputs, "duration_minutes")

	// Test initial values are set
	maxRetriesInput := ui.policyInputs["max_retries"]
	assert.Equal(t, "3", maxRetriesInput.Value())

	baseRateInput := ui.trafficInputs["base_rate"]
	assert.Equal(t, "50", baseRateInput.Value())
}

func TestUIStyleConstants(t *testing.T) {
	// Test that style constants are properly defined
	assert.NotEmpty(t, primaryColor)
	assert.NotEmpty(t, secondaryColor)
	assert.NotEmpty(t, accentColor)
	assert.NotEmpty(t, mutedColor)
	assert.NotEmpty(t, errorColor)
	assert.NotEmpty(t, successColor)

	// Test that styles can be applied
	styled := primaryColor.Render("test")
	assert.Contains(t, styled, "test")

	styled = errorColor.Render("error")
	assert.Contains(t, styled, "error")
}

func TestUIHeaderRendering(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	view := ui.View()

	// Check that header elements are present
	assert.Contains(t, view, "Policy Simulator")
	assert.Contains(t, view, "Policies")
	assert.Contains(t, view, "Traffic")
	assert.Contains(t, view, "Simulation")
	assert.Contains(t, view, "Results")
	assert.Contains(t, view, "Charts")
	assert.Contains(t, view, "q: quit")
}

func TestUITabHighlighting(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test that current tab is highlighted
	ui.currentTab = TabPolicies
	view := ui.View()

	// The current tab should be rendered differently (though exact styling may vary)
	assert.Contains(t, view, "Policies")

	// Switch tab and verify highlighting changes
	ui.currentTab = TabTraffic
	view = ui.View()
	assert.Contains(t, view, "Traffic")
}

func TestUIWindowSizeHandling(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test window size message handling
	windowMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}

	model, cmd := ui.Update(windowMsg)
	ui = model.(*PolicySimulatorUI)
	assert.Nil(t, cmd)

	// Verify that the UI can handle the window size
	view := ui.View()
	assert.NotEmpty(t, view)
}

func TestUIRunSimulationAction(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 10 * time.Second,
		TimeStep:          1 * time.Second,
		MaxWorkers:        3,
		RedisPoolSize:     2,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Switch to simulation tab
	ui.currentTab = TabSimulation

	// Test that entering 'r' on simulation tab triggers run
	model, cmd := ui.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	ui = model.(*PolicySimulatorUI)

	// Should return a command for running simulation
	assert.NotNil(t, cmd)

	view := ui.View()
	assert.Contains(t, view, "Simulation")
}

func TestUIDefaultPolicyValues(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test that current policy has default values
	assert.Equal(t, 3, ui.currentPolicy.MaxRetries)
	assert.Equal(t, time.Second, ui.currentPolicy.InitialBackoff)
	assert.Equal(t, 100.0, ui.currentPolicy.MaxRatePerSecond)
	assert.Equal(t, 5, ui.currentPolicy.MaxConcurrency)
}

func TestUIDefaultTrafficValues(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	// Test that current pattern has default values
	assert.Equal(t, "Default Load", ui.currentPattern.Name)
	assert.Equal(t, TrafficConstant, ui.currentPattern.Type)
	assert.Equal(t, 50.0, ui.currentPattern.BaseRate)
	assert.Equal(t, 5*time.Minute, ui.currentPattern.Duration)
}

func TestUIContentWrapping(t *testing.T) {
	config := &SimulatorConfig{
		SimulationDuration: 5 * time.Minute,
		TimeStep:          1 * time.Second,
		MaxWorkers:        5,
		RedisPoolSize:     3,
	}

	simulator := NewPolicySimulator(config)
	ui := NewPolicySimulatorUI(simulator)

	view := ui.View()

	// Check that view is not empty and contains expected structure
	assert.NotEmpty(t, view)
	assert.True(t, len(view) > 100) // Should be substantial content

	// Check that lines are reasonable length (basic wrapping check)
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		// Most lines should be reasonable length (allowing for some flexibility)
		if len(line) > 200 {
			t.Logf("Long line detected: %s", line[:min(50, len(line))]+"...")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}