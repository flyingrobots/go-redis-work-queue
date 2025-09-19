//go:build tui_experimental
// +build tui_experimental

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DemoTransformation shows before/after of styling improvements
func DemoTransformation(width, height int) string {
	styles := GetStyleSet(width, height)

	// BEFORE: Basic text output (like your screenshot)
	before := `? toggle help • q quit

Duck
Duck
Duck
Honk -> Goose
Duck

> |`

	// AFTER: Professional styled version
	header := styles.CreateHeader("Redis Work Queue Monitor", "Production Environment • Desktop Mode")

	statusBar := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.CreateStatusBadge("Connected", "success"),
		" ",
		styles.CreateMetricDisplay("Queues", "5", "up"),
		" ",
		styles.CreateMetricDisplay("Workers", "12", "up"),
		" ",
		styles.CreateStatusBadge("3 DLQ", "warning"),
	)

	// Transform the queue list with proper styling
	queueData := [][]string{
		{"prod-jobs", "1,234", "active"},
		{"async-emails", "789", "active"},
		{"data-processing", "45", "idle"},
		{"notifications", "156", "busy"},
		{"cleanup-tasks", "23", "idle"},
	}

	var queueRows []string

	// Table header
	headerRow := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.StatusInfo.Render("Queue Name".ljust(25)),
		styles.StatusInfo.Render("Jobs".rjust(8)),
		styles.StatusInfo.Render("Status".ljust(15)),
	)
	queueRows = append(queueRows, headerRow)

	// Separator line
	separator := strings.Repeat("─", 50)
	queueRows = append(queueRows, styles.Separator.Render(separator))

	// Data rows
	for i, row := range queueData {
		var nameStyle lipgloss.Style
		if i == 0 { // Highlight first row (selected)
			nameStyle = styles.StatusInfo.Bold(true)
		} else {
			nameStyle = lipgloss.NewStyle()
		}

		statusBadge := styles.CreateStatusBadge(row[2], row[2])

		dataRow := lipgloss.JoinHorizontal(lipgloss.Left,
			nameStyle.Render(row[0].ljust(25)),
			nameStyle.Render(row[1].rjust(8)),
			" ",
			statusBadge,
		)
		queueRows = append(queueRows, dataRow)
	}

	queueTable := strings.Join(queueRows, "\n")

	// Create panels
	queuePanel := styles.CreateInfoCard("📋 Active Queues", queueTable)

	// Quick actions
	actionButtons := styles.CreateButtonBar([]string{"Pause", "Resume", "Peek", "Drain"}, 0)
	actionsPanel := styles.Panel.Render("Quick Actions\n" + actionButtons)

	// Charts panel (simplified)
	chartContent := `Processing Rate (jobs/minute)
     ▄▆█▆▄▃▂▁▂▃▄▆█▆▄  Current: 245/min
     ▁▂▃▄▆█▆▄▃▂▁▂▃▄▆  Peak: 487/min`
	chartsPanel := styles.CreateInfoCard("📊 Metrics", chartContent)

	// Worker status
	workerContent := `Fleet Status: ✓ All Healthy
Active Workers: 12/15
CPU Usage: ▓▓▓▓▓░░░░░ 52%
Memory: ▓▓▓▓▓▓▓░░░ 68%`
	workerPanel := styles.CreateInfoCard("👥 Worker Fleet", workerContent)

	// Footer with help
	footer := styles.buildEnhancedFooter(styles)

	// Arrange layout based on terminal size
	var after string
	if width > 100 {
		// Desktop layout - 2x2 grid
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, queuePanel, actionsPanel)
		bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, chartsPanel, workerPanel)
		content := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

		after = lipgloss.JoinVertical(lipgloss.Left,
			header,
			statusBar,
			content,
			footer,
		)
	} else {
		// Mobile/Tablet - stacked layout
		after = lipgloss.JoinVertical(lipgloss.Left,
			header,
			statusBar,
			queuePanel,
			chartsPanel,
			footer,
		)
	}

	// Create side-by-side comparison
	beforePanel := styles.Panel.Copy().
		BorderForeground(lipgloss.Color("#f85149")).
		Width(width/2 - 2).
		Render("BEFORE - Basic Text Output\n\n" + before)

	afterPanel := styles.Panel.Copy().
		BorderForeground(lipgloss.Color("#56d364")).
		Width(width/2 - 2).
		Render("AFTER - Professional LipGloss Styling\n\n" + after)

	comparison := lipgloss.JoinHorizontal(lipgloss.Top, beforePanel, afterPanel)

	title := styles.AppTitle.Copy().
		Width(width).
		Align(lipgloss.Center).
		Render("🎨 LipGloss Transformation Demo")

	return lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		comparison,
	)
}

// Helper method extensions
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

// buildEnhancedFooter creates a contextual footer (simplified version for demo)
func (s StyleSet) buildEnhancedFooter(styles StyleSet) string {
	hints := []string{"[Enter] Peek Queue", "[Space] Pause/Resume", "[Tab] Switch Panel", "[?] Help", "[Q] Quit"}
	hintText := strings.Join(hints, " • ")

	return styles.Panel.Copy().
		Background(lipgloss.AdaptiveColor{Light: "#f0f6fc", Dark: "#0d1117"}).
		Padding(0, 1).
		Margin(0).
		Render(styles.StatusMuted.Render(hintText))
}
