//go:build tui_experimental
// +build tui_experimental

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// EnhancedView creates a polished, responsive view using the new styling system
func (m model) EnhancedView() string {
	// Get responsive styles for current terminal size
	styles := GetStyleSet(m.width, m.height)

	// Create the main application header with branding
	headerTitle := "Redis Work Queue Monitor"
	headerSubtitle := fmt.Sprintf("Connected to %s • %s mode",
		m.cfg.Redis.Addr,
		strings.Title(styles.Breakpoint))

	header := styles.CreateHeader(headerTitle, headerSubtitle)

	// Create responsive status bar with connection and system info
	statusBar := m.buildEnhancedStatusBar(styles)

	// Create enhanced tab navigation
	tabBar := m.buildEnhancedTabBar(styles)

	// Main content area adapts to breakpoint
	var content string
	switch styles.Breakpoint {
	case "mobile":
		content = m.buildMobileLayout(styles)
	case "tablet":
		content = m.buildTabletLayout(styles)
	case "desktop":
		content = m.buildDesktopLayout(styles)
	case "ultrawide":
		content = m.buildUltrawideLayout(styles)
	}

	// Error handling with styled alerts
	if m.errText != "" {
		errorAlert := styles.Panel.Copy().
			BorderForeground(lipgloss.Color("#f85149")).
			Background(lipgloss.AdaptiveColor{Light: "#ffebe9", Dark: "#2d1b1e"}).
			Render(styles.StatusError.Render("Error: " + m.errText))
		content = errorAlert + "\n" + content
	}

	// Loading overlay
	if m.loading {
		loadingOverlay := m.buildLoadingOverlay(styles)
		return loadingOverlay
	}

	// Combine all elements
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		statusBar,
		tabBar,
		content,
		m.buildEnhancedFooter(styles),
	)
}

// buildEnhancedStatusBar creates a professional status bar
func (m model) buildEnhancedStatusBar(styles StyleSet) string {
	elements := []string{}

	// Connection status
	if m.lastStats.Heartbeats > 0 {
		elements = append(elements, styles.CreateStatusBadge("Connected", "success"))
	} else {
		elements = append(elements, styles.CreateStatusBadge("Disconnected", "error"))
	}

	// Queue health summary
	totalQueues := len(m.lastStats.ProcessingLists)
	if totalQueues > 0 {
		elements = append(elements, styles.CreateMetricDisplay("Queues", fmt.Sprintf("%d", totalQueues), "flat"))
	}

	// Worker status
	if m.lastStats.Heartbeats > 0 {
		elements = append(elements, styles.CreateMetricDisplay("Workers", fmt.Sprintf("%d", m.lastStats.Heartbeats), "up"))
	}

	// DLQ alerts
	dlqCount := m.lastStats.DeadLetterQueueJobs
	if dlqCount > 0 {
		elements = append(elements, styles.CreateStatusBadge(fmt.Sprintf("%d DLQ", dlqCount), "warning"))
	}

	// Time since last update
	lastUpdate := time.Since(time.Now()).Truncate(time.Second) // Placeholder
	elements = append(elements, styles.StatusMuted.Render("Updated: just now"))

	statusContent := lipgloss.JoinHorizontal(lipgloss.Left, elements...)

	return styles.Panel.Copy().
		Background(lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#1c2128"}).
		Padding(0, 1).
		Margin(0).
		Render(statusContent)
}

// buildEnhancedTabBar creates responsive tab navigation
func (m model) buildEnhancedTabBar(styles StyleSet) string {
	tabs := []struct {
		id    tabID
		label string
		icon  string
		color string
	}{
		{tabJobs, "Jobs", "⚡", "#58a6ff"},
		{tabWorkers, "Workers", "👥", "#56d364"},
		{tabDLQ, "DLQ", "💀", "#f85149"},
		{tabTimeTravel, "Debug", "⏰", "#d2a8ff"},
		{tabEventHooks, "Events", "🔗", "#f9e71e"},
		{tabSettings, "Settings", "⚙️", "#8b949e"},
	}

	var tabElements []string

	for _, tab := range tabs {
		var tabStyle lipgloss.Style
		var content string

		if styles.Breakpoint == "mobile" {
			// Mobile: Icon only
			content = tab.icon
		} else {
			// Tablet+: Icon + Label
			content = tab.icon + " " + tab.label
		}

		if m.activeTab == tab.id {
			tabStyle = styles.TabActive.Copy().BorderForeground(lipgloss.Color(tab.color))
		} else {
			tabStyle = styles.TabInactive
		}

		tabElements = append(tabElements, tabStyle.Render(content))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabElements...)

	// Add help hint for larger screens
	if styles.Width > 60 {
		helpHint := styles.StatusMuted.Render(" • Press [?] for help • [Ctrl+P] Command Palette")
		tabBar = lipgloss.JoinHorizontal(lipgloss.Left, tabBar, helpHint)
	}

	return tabBar
}

// buildMobileLayout creates a card-based mobile layout
func (m model) buildMobileLayout(styles StyleSet) string {
	cards := []string{}

	switch m.activeTab {
	case tabJobs:
		// Queue overview cards
		for queueName, count := range m.getQueueSummary() {
			cardTitle := fmt.Sprintf("📋 %s", queueName)
			cardContent := fmt.Sprintf("Jobs: %d\n%s",
				count,
				styles.CreateProgressBar(int(count), 1000, 15))
			cards = append(cards, styles.CreateInfoCard(cardTitle, cardContent))
		}

		// Worker health card
		if m.lastStats.Heartbeats > 0 {
			workerCard := fmt.Sprintf("Active: %d workers\nHealthy: %s",
				m.lastStats.Heartbeats,
				styles.CreateStatusBadge("All systems go", "success"))
			cards = append(cards, styles.CreateInfoCard("👥 Workers", workerCard))
		}

	case tabDLQ:
		// DLQ alert card
		if m.lastStats.DeadLetterQueueJobs > 0 {
			dlqCard := fmt.Sprintf("Failed Jobs: %d\nRequires attention",
				m.lastStats.DeadLetterQueueJobs)
			cards = append(cards, styles.CreateInfoCard("💀 Dead Letter Queue", dlqCard))
		}
	}

	// Voice command hint for mobile
	voiceHint := styles.StatusInfo.Render("🎤 Say \"Hey Queue\" for voice commands")
	cards = append(cards, styles.Card.Render(voiceHint))

	return lipgloss.JoinVertical(lipgloss.Left, cards...)
}

// buildTabletLayout creates a two-column tablet layout
func (m model) buildTabletLayout(styles StyleSet) string {
	var leftColumn, rightColumn []string

	switch m.activeTab {
	case tabJobs:
		// Left: Queue table
		queueTable := m.buildEnhancedQueueTable(styles)
		leftColumn = append(leftColumn, queueTable)

		// Right: Charts and metrics
		chartPanel := m.buildChartsPanel(styles)
		rightColumn = append(rightColumn, chartPanel)

		metricsPanel := m.buildMetricsPanel(styles)
		rightColumn = append(rightColumn, metricsPanel)
	}

	left := lipgloss.JoinVertical(lipgloss.Left, leftColumn...)
	right := lipgloss.JoinVertical(lipgloss.Left, rightColumn...)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// buildDesktopLayout creates a multi-panel desktop layout
func (m model) buildDesktopLayout(styles StyleSet) string {
	switch m.activeTab {
	case tabJobs:
		return m.buildJobsDashboard(styles)
	case tabWorkers:
		return m.buildWorkerDashboard(styles)
	case tabDLQ:
		return m.buildDLQDashboard(styles)
	default:
		return m.buildGenericDashboard(styles)
	}
}

// buildUltrawideLayout creates a mission control ultrawide layout
func (m model) buildUltrawideLayout(styles StyleSet) string {
	// Ultrawide gets side-by-side multi-view
	mainPanel := m.buildDesktopLayout(styles)

	// Add sidebar with quick actions and monitoring
	sidebar := m.buildSidebar(styles)

	return lipgloss.JoinHorizontal(lipgloss.Top, mainPanel, sidebar)
}

// buildJobsDashboard creates the enhanced jobs dashboard
func (m model) buildJobsDashboard(styles StyleSet) string {
	// Top row: Queue table + Quick actions
	queuePanel := styles.CreateInfoCard("📋 Active Queues", m.buildEnhancedQueueTable(styles))

	actionButtons := styles.CreateButtonBar([]string{"Pause All", "Resume All", "Drain Safely"}, -1)
	actionsPanel := styles.Panel.Render("Quick Actions\n" + actionButtons)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, queuePanel, actionsPanel)

	// Middle row: Charts + Worker status
	chartsPanel := styles.CreateInfoCard("📊 Queue Metrics", m.buildChartsPanel(styles))
	workersPanel := styles.CreateInfoCard("👥 Worker Fleet", m.buildWorkerSummary(styles))

	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, chartsPanel, workersPanel)

	// Bottom row: Recent activity + System health
	activityPanel := styles.CreateInfoCard("📝 Recent Activity", m.buildActivityLog(styles))
	healthPanel := styles.CreateInfoCard("❤️ System Health", m.buildHealthSummary(styles))

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, activityPanel, healthPanel)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, middleRow, bottomRow)
}

// buildEnhancedQueueTable creates a styled queue table
func (m model) buildEnhancedQueueTable(styles StyleSet) string {
	if len(m.allRows) == 0 {
		return styles.StatusMuted.Render("No queues detected")
	}

	var tableRows []string

	// Header
	headerRow := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.StatusInfo.Render("Queue Name".ljust(20)),
		styles.StatusInfo.Render("Jobs".rjust(8)),
		styles.StatusInfo.Render("Rate".rjust(8)),
		styles.StatusInfo.Render("Status".ljust(12)),
	)
	tableRows = append(tableRows, headerRow)

	// Separator
	separator := strings.Repeat("─", styles.Width-4)
	tableRows = append(tableRows, styles.Separator.Render(separator))

	// Data rows
	for i, row := range m.allRows {
		var rowStyle lipgloss.Style
		if i == m.tbl.Cursor() {
			rowStyle = styles.StatusInfo // Highlighted row
		} else {
			rowStyle = lipgloss.NewStyle() // Default row
		}

		// Determine status based on job count
		status := "idle"
		count := parseCount(row[1]) // Assuming row[1] is count
		if count > 100 {
			status = "busy"
		} else if count > 0 {
			status = "active"
		}

		statusBadge := styles.CreateStatusBadge(status, status)

		dataRow := lipgloss.JoinHorizontal(lipgloss.Left,
			rowStyle.Render(truncate(row[0], 20)), // Queue name
			rowStyle.Render(row[1].rjust(8)),      // Job count
			rowStyle.Render("45/min".rjust(8)),    // Rate (placeholder)
			statusBadge,                           // Status
		)
		tableRows = append(tableRows, dataRow)
	}

	return strings.Join(tableRows, "\n")
}

// buildLoadingOverlay creates an animated loading screen
func (m model) buildLoadingOverlay(styles StyleSet) string {
	spinner := m.spinner.View()

	loadingContent := lipgloss.JoinVertical(lipgloss.Center,
		styles.Loading.Render(spinner+" Connecting to Redis..."),
		"",
		styles.StatusMuted.Render("• Detecting queue configuration"),
		styles.StatusMuted.Render("• Scanning for active workers"),
		styles.StatusMuted.Render("• Loading dashboard components"),
		"",
		styles.CreateProgressBar(75, 100, 30),
	)

	// Center the loading screen
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.Modal.Render(loadingContent))
}

// buildEnhancedFooter creates a contextual footer
func (m model) buildEnhancedFooter(styles StyleSet) string {
	var hints []string

	// Context-sensitive hints
	switch m.activeTab {
	case tabJobs:
		hints = []string{"[Enter] Peek Queue", "[Space] Pause/Resume", "[/] Filter"}
	case tabWorkers:
		hints = []string{"[D] Drain Worker", "[R] Restart", "[S] Scale Fleet"}
	case tabDLQ:
		hints = []string{"[R] Requeue All", "[P] Peek Failed Job", "[C] Clear DLQ"}
	}

	// Universal hints
	hints = append(hints, "[Tab] Switch Panel", "[?] Help", "[Q] Quit")

	hintText := strings.Join(hints, " • ")

	return styles.Panel.Copy().
		Background(lipgloss.AdaptiveColor{Light: "#f0f6fc", Dark: "#0d1117"}).
		Padding(0, 1).
		Margin(0).
		Render(styles.StatusMuted.Render(hintText))
}

// Helper functions

func (m model) getQueueSummary() map[string]int64 {
	// Mock implementation - replace with actual queue data
	return map[string]int64{
		"prod-jobs":       1234,
		"async-emails":    789,
		"data-processing": 45,
	}
}

func parseCount(s string) int {
	// Helper to parse count from string
	return 0 // Placeholder
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// String helper methods (would typically be in a utils package)
func (s string) ljust(width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func (s string) rjust(width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return strings.Repeat(" ", width-len(s)) + s
}
