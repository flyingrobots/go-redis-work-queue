package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// staticStringModel is a tiny tea.Model wrapper around a fixed string view.
type staticStringModel struct{ s string }

func (s staticStringModel) Init() tea.Cmd                           { return nil }
func (s staticStringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s staticStringModel) View() string                            { return s.s }

func renderConfirmModal(m model) string {
	title := "Confirm"
	msg := ""
	switch m.confirmAction {
	case "purge-dlq":
		msg = "Purge dead letter queue?"
	case "purge-all":
		msg = "Purge ALL managed keys?"
	default:
		msg = m.confirmAction
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		msg,
		"[y] Yes   [n] No",
	)

	width := m.width
	if width <= 0 {
		width = 80
	}
	modal := box.Render(content)
	pad := 0
	if w := lipgloss.Width(modal); width > w {
		pad = (width - w) / 2
	}
	return strings.Repeat(" ", pad) + modal
}

// renderOverlayScreen builds a full-screen dimmed scrim and centers the confirm modal.
func renderOverlayScreen(m model) string {
    width := m.width
    height := m.height
    if width <= 0 { width = 80 }
    if height <= 0 { height = 24 }

    scrimCell := lipgloss.NewStyle().Background(lipgloss.Color("236")).Faint(true).Render(" ")
    line := strings.Repeat(scrimCell, width)
    lines := make([]string, height)
    for i := 0; i < height; i++ { lines[i] = line }

    modal := renderConfirmModal(m)
    modalLines := strings.Split(modal, "\n")
    modalH := len(modalLines)
    modalW := 0
    for _, ml := range modalLines {
        if w := lipgloss.Width(ml); w > modalW { modalW = w }
    }
    top := (height - modalH) / 2
    left := (width - modalW) / 2
    if top < 0 { top = 0 }
    if left < 0 { left = 0 }
    for i := 0; i < modalH && (top+i) < height; i++ {
        ml := modalLines[i]
        lp := left
        rp := width - (left + lipgloss.Width(ml))
        if lp < 0 { lp = 0 }
        if rp < 0 { rp = 0 }
        leftPad := strings.Repeat(scrimCell, lp)
        rightPad := strings.Repeat(scrimCell, rp)
        lines[top+i] = leftPad + ml + rightPad
    }
    return strings.Join(lines, "\n")
}

// renderHelpOverlay dims the background and centers the help view.
func renderHelpOverlay(m model, _ string) string {
    width := m.width
    height := m.height
    if width <= 0 { width = 80 }
    if height <= 0 { height = 24 }

    scrimCell := lipgloss.NewStyle().Background(lipgloss.Color("236")).Faint(true).Render(" ")
    line := strings.Repeat(scrimCell, width)
    lines := make([]string, height)
    for i := 0; i < height; i++ { lines[i] = line }

    hv := m.help2.View()
    hvLines := strings.Split(hv, "\n")
    hH := len(hvLines)
    hW := 0
    for _, l := range hvLines {
        if w := lipgloss.Width(l); w > hW { hW = w }
    }
    top := (height - hH) / 2
    left := (width - hW) / 2
    if top < 0 { top = 0 }
    if left < 0 { left = 0 }
    for i := 0; i < hH && (top+i) < height; i++ {
        ml := hvLines[i]
        lp := left
        rp := width - (left + lipgloss.Width(ml))
        if lp < 0 { lp = 0 }
        if rp < 0 { rp = 0 }
        leftPad := strings.Repeat(scrimCell, lp)
        rightPad := strings.Repeat(scrimCell, rp)
        lines[top+i] = leftPad + ml + rightPad
    }
    return strings.Join(lines, "\n")
}
