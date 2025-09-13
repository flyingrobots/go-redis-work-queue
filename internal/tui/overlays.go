package tui

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// staticStringModel is a tiny tea.Model wrapper around a fixed string view.
type staticStringModel struct{ s string }
func (s staticStringModel) Init() tea.Cmd { return nil }
func (s staticStringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s staticStringModel) View() string { return s.s }

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
    if width <= 0 { width = 80 }
    modal := box.Render(content)
    pad := 0
    if w := lipgloss.Width(modal); width > w { pad = (width - w) / 2 }
    return strings.Repeat(" ", pad) + modal
}
