package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabZone defines a clickable region for a tab on the first row
type tabZone struct {
	id    tabID
	start int // inclusive x
	end   int // exclusive x
}

func (m model) buildTabBar() (string, []tabZone) {
	// Labels in order
	items := []struct {
		id    tabID
		label string
		color string
	}{
		{tabJobs, "Job Queue", "#7aa2f7"},
		{tabWorkers, "Workers", "#9ece6a"},
		{tabDLQ, "Dead Letter", "#f7768e"},
		{tabSettings, "Settings", "#bb9af7"},
	}

	// Styles
	// Compact, borderless tabs to conserve width
	base := lipgloss.NewStyle().Padding(0, 1)
	inactive := base.Foreground(lipgloss.AdaptiveColor{Dark: "#bbbbbb", Light: "#333333"})

	b := &strings.Builder{}
	zones := make([]tabZone, 0, len(items))
	x := 0
	// left margin
	leftPad := " "
	b.WriteString(leftPad)
	x += lipgloss.Width(leftPad)

	for i, it := range items {
		st := inactive
		if it.id == m.activeTab {
			st = base.Bold(true).Foreground(lipgloss.Color(it.color)).Underline(true)
		}
		seg := st.Render(it.label)
		b.WriteString(seg)
		zones = append(zones, tabZone{id: it.id, start: x, end: x + lipgloss.Width(seg)})
		x += lipgloss.Width(seg)
		if i != len(items)-1 {
			sep := " "
			b.WriteString(sep)
			x += lipgloss.Width(sep)
		}
	}
	return b.String(), zones
}
