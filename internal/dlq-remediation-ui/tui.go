package dlqremediationui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModePeek
	ViewModeConfirm
)

type TUIModel struct {
	manager     DLQManager
	logger      *slog.Logger

	viewMode    ViewMode
	entries     []DLQEntry
	patterns    []ErrorPattern
	stats       *DLQStats

	table       table.Model
	paginator   paginator.Model
	filterInput textinput.Model

	selectedIDs []string
	currentPage int
	pageSize    int
	totalCount  int

	filter      DLQFilter
	loading     bool
	error       string
	message     string

	confirmAction string
	confirmPrompt string

	peekEntry   *DLQEntry

	keyMap      KeyMap
	styles      Styles
}

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	Requeue     key.Binding
	Purge       key.Binding
	PurgeAll    key.Binding
	Peek        key.Binding
	Filter      key.Binding
	Refresh     key.Binding
	NextPage    key.Binding
	PrevPage    key.Binding
	Confirm     key.Binding
	Cancel      key.Binding
	Quit        key.Binding
}

type Styles struct {
	Base          lipgloss.Style
	Header        lipgloss.Style
	Footer        lipgloss.Style
	Selected      lipgloss.Style
	Error         lipgloss.Style
	Success       lipgloss.Style
	Warning       lipgloss.Style
	Info          lipgloss.Style
	Table         lipgloss.Style
	FilterInput   lipgloss.Style
	ConfirmDialog lipgloss.Style
	PeekDialog    lipgloss.Style
}

func NewTUIModel(manager DLQManager, logger *slog.Logger) *TUIModel {
	keyMap := KeyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:      key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		Select:    key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
		SelectAll: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "select all")),
		Requeue:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "requeue")),
		Purge:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "purge")),
		PurgeAll:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "purge all")),
		Peek:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "peek")),
		Filter:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
		Refresh:   key.NewBinding(key.WithKeys("F5"), key.WithHelp("F5", "refresh")),
		NextPage:  key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next page")),
		PrevPage:  key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "prev page")),
		Confirm:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}

	styles := Styles{
		Base:      lipgloss.NewStyle().Padding(1),
		Header:    lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		Footer:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		Info:      lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		Table:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
		FilterInput: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		ConfirmDialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1).
			Width(50),
		PeekDialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1).
			Width(80).
			Height(30),
	}

	columns := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "Queue", Width: 15},
		{Title: "Type", Width: 15},
		{Title: "Error", Width: 40},
		{Title: "Failed At", Width: 20},
		{Title: "Attempts", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	filterInput := textinput.New()
	filterInput.Placeholder = "Filter by queue, type, or error pattern..."
	filterInput.CharLimit = 100
	filterInput.Width = 50

	pag := paginator.New()
	pag.Type = paginator.Dots
	pag.PerPage = 20
	pag.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{}).Render("•")
	pag.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{}).Render("○")

	return &TUIModel{
		manager:     manager,
		logger:      logger,
		viewMode:    ViewModeList,
		table:       t,
		paginator:   pag,
		filterInput: filterInput,
		pageSize:    20,
		currentPage: 1,
		selectedIDs: make([]string, 0),
		keyMap:      keyMap,
		styles:      styles,
	}
}

func (m *TUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadEntries(),
		m.loadStats(),
	)
}

func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case entriesLoadedMsg:
		m.loading = false
		m.entries = msg.entries
		m.patterns = msg.patterns
		m.totalCount = msg.totalCount
		m.error = ""
		m.updateTable()
		return m, nil

	case statsLoadedMsg:
		m.stats = msg.stats
		return m, nil

	case operationCompletedMsg:
		m.loading = false
		m.message = msg.message
		m.viewMode = ViewModeList
		return m, tea.Batch(m.loadEntries(), m.loadStats())

	case errorMsg:
		m.loading = false
		m.error = msg.error
		return m, nil
	}

	switch m.viewMode {
	case ViewModeList:
		m.table, cmd = m.table.Update(msg)
	case ViewModePeek:
		// Handle peek mode updates
	case ViewModeConfirm:
		// Handle confirm mode updates
	}

	return m, cmd
}

func (m *TUIModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case ViewModeList:
		return m.handleListKeyPress(msg)
	case ViewModePeek:
		return m.handlePeekKeyPress(msg)
	case ViewModeConfirm:
		return m.handleConfirmKeyPress(msg)
	}
	return m, nil
}

func (m *TUIModel) handleListKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keyMap.Refresh):
		m.loading = true
		m.error = ""
		m.message = ""
		return m, tea.Batch(m.loadEntries(), m.loadStats())

	case key.Matches(msg, m.keyMap.Select):
		if len(m.entries) > 0 {
			selectedRow := m.table.Cursor()
			if selectedRow < len(m.entries) {
				entryID := m.entries[selectedRow].ID
				if m.isSelected(entryID) {
					m.removeSelection(entryID)
				} else {
					m.selectedIDs = append(m.selectedIDs, entryID)
				}
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.SelectAll):
		if len(m.selectedIDs) == len(m.entries) {
			m.selectedIDs = make([]string, 0)
		} else {
			m.selectedIDs = make([]string, len(m.entries))
			for i, entry := range m.entries {
				m.selectedIDs[i] = entry.ID
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Peek):
		if len(m.entries) > 0 {
			selectedRow := m.table.Cursor()
			if selectedRow < len(m.entries) {
				m.viewMode = ViewModePeek
				m.peekEntry = &m.entries[selectedRow]
			}
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Requeue):
		if len(m.selectedIDs) > 0 {
			m.confirmAction = "requeue"
			m.confirmPrompt = fmt.Sprintf("Requeue %d selected entries?", len(m.selectedIDs))
			m.viewMode = ViewModeConfirm
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Purge):
		if len(m.selectedIDs) > 0 {
			m.confirmAction = "purge"
			m.confirmPrompt = fmt.Sprintf("Purge %d selected entries? This cannot be undone.", len(m.selectedIDs))
			m.viewMode = ViewModeConfirm
		}
		return m, nil

	case key.Matches(msg, m.keyMap.PurgeAll):
		m.confirmAction = "purge_all"
		m.confirmPrompt = "Purge ALL entries matching current filter? This cannot be undone."
		m.viewMode = ViewModeConfirm
		return m, nil

	case key.Matches(msg, m.keyMap.NextPage):
		if m.currentPage*m.pageSize < m.totalCount {
			m.currentPage++
			m.loading = true
			return m, m.loadEntries()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.PrevPage):
		if m.currentPage > 1 {
			m.currentPage--
			m.loading = true
			return m, m.loadEntries()
		}
		return m, nil
	}

	return m, nil
}

func (m *TUIModel) handlePeekKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Cancel):
		m.viewMode = ViewModeList
		m.peekEntry = nil
		return m, nil
	}
	return m, nil
}

func (m *TUIModel) handleConfirmKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Confirm):
		m.loading = true
		m.viewMode = ViewModeList

		switch m.confirmAction {
		case "requeue":
			return m, m.requeueSelected()
		case "purge":
			return m, m.purgeSelected()
		case "purge_all":
			return m, m.purgeAll()
		}
		return m, nil

	case key.Matches(msg, m.keyMap.Cancel):
		m.viewMode = ViewModeList
		m.confirmAction = ""
		m.confirmPrompt = ""
		return m, nil
	}
	return m, nil
}

func (m *TUIModel) View() string {
	switch m.viewMode {
	case ViewModeList:
		return m.viewList()
	case ViewModePeek:
		return m.viewPeek()
	case ViewModeConfirm:
		return m.viewConfirm()
	}
	return ""
}

func (m *TUIModel) viewList() string {
	var sections []string

	header := m.styles.Header.Render("DLQ Remediation")
	sections = append(sections, header)

	if m.stats != nil {
		statsLine := fmt.Sprintf("Total: %d entries", m.stats.TotalEntries)
		if len(m.stats.ByQueue) > 0 {
			statsLine += " | Queues: "
			for queue, count := range m.stats.ByQueue {
				statsLine += fmt.Sprintf("%s(%d) ", queue, count)
			}
		}
		sections = append(sections, m.styles.Info.Render(statsLine))
	}

	if m.loading {
		sections = append(sections, m.styles.Info.Render("Loading..."))
	}

	if m.error != "" {
		sections = append(sections, m.styles.Error.Render("Error: "+m.error))
	}

	if m.message != "" {
		sections = append(sections, m.styles.Success.Render(m.message))
	}

	tableView := m.styles.Table.Render(m.table.View())
	sections = append(sections, tableView)

	pageInfo := fmt.Sprintf("Page %d of %d | %d entries | %d selected",
		m.currentPage, (m.totalCount+m.pageSize-1)/m.pageSize, m.totalCount, len(m.selectedIDs))
	sections = append(sections, m.styles.Footer.Render(pageInfo))

	help := "space: select | r: requeue | d: purge | D: purge all | p: peek | F5: refresh | q: quit"
	sections = append(sections, m.styles.Footer.Render(help))

	return m.styles.Base.Render(strings.Join(sections, "\n\n"))
}

func (m *TUIModel) viewPeek() string {
	if m.peekEntry == nil {
		return "No entry to peek"
	}

	var content strings.Builder

	content.WriteString(m.styles.Header.Render("Entry Details"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("ID: %s\n", m.peekEntry.ID))
	content.WriteString(fmt.Sprintf("Job ID: %s\n", m.peekEntry.JobID))
	content.WriteString(fmt.Sprintf("Queue: %s\n", m.peekEntry.Queue))
	content.WriteString(fmt.Sprintf("Type: %s\n", m.peekEntry.Type))
	content.WriteString(fmt.Sprintf("Failed At: %s\n", m.peekEntry.FailedAt.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("Attempts: %d\n", len(m.peekEntry.Attempts)))
	content.WriteString("\n")

	content.WriteString("Error:\n")
	content.WriteString(fmt.Sprintf("  Message: %s\n", m.peekEntry.Error.Message))
	content.WriteString(fmt.Sprintf("  Code: %s\n", m.peekEntry.Error.Code))
	content.WriteString("\n")

	content.WriteString("Payload:\n")
	var prettyPayload interface{}
	if err := json.Unmarshal(m.peekEntry.Payload, &prettyPayload); err == nil {
		if payloadBytes, err := json.MarshalIndent(prettyPayload, "  ", "  "); err == nil {
			content.WriteString("  " + string(payloadBytes))
		}
	} else {
		content.WriteString("  " + string(m.peekEntry.Payload))
	}

	content.WriteString("\n\nPress ESC to return")

	return m.styles.PeekDialog.Render(content.String())
}

func (m *TUIModel) viewConfirm() string {
	content := m.confirmPrompt + "\n\nPress ENTER to confirm, ESC to cancel"
	return m.styles.ConfirmDialog.Render(content)
}

func (m *TUIModel) updateTable() {
	rows := make([]table.Row, len(m.entries))
	for i, entry := range m.entries {
		selected := ""
		if m.isSelected(entry.ID) {
			selected = "✓ "
		}

		rows[i] = table.Row{
			selected + entry.ID[:min(18, len(entry.ID))],
			entry.Queue,
			entry.Type,
			truncateString(entry.Error.Message, 38),
			entry.FailedAt.Format("2006-01-02 15:04"),
			strconv.Itoa(len(entry.Attempts)),
		}
	}
	m.table.SetRows(rows)
}

func (m *TUIModel) loadEntries() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		pagination := PaginationRequest{
			Page:      m.currentPage,
			PageSize:  m.pageSize,
			SortBy:    "failed_at",
			SortOrder: SortOrderDesc,
		}

		response, err := m.manager.ListEntries(ctx, m.filter, pagination)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		return entriesLoadedMsg{
			entries:    response.Entries,
			patterns:   response.Patterns,
			totalCount: response.TotalCount,
		}
	}
}

func (m *TUIModel) loadStats() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		stats, err := m.manager.GetStats(ctx)
		if err != nil {
			m.logger.Warn("Failed to load stats", "error", err)
			return nil
		}

		return statsLoadedMsg{stats: stats}
	}
}

func (m *TUIModel) requeueSelected() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		result, err := m.manager.RequeueEntries(ctx, m.selectedIDs)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		message := fmt.Sprintf("Requeued %d entries successfully", len(result.Successful))
		if len(result.Failed) > 0 {
			message += fmt.Sprintf(" (%d failed)", len(result.Failed))
		}

		return operationCompletedMsg{message: message}
	}
}

func (m *TUIModel) purgeSelected() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		result, err := m.manager.PurgeEntries(ctx, m.selectedIDs)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		message := fmt.Sprintf("Purged %d entries successfully", len(result.Successful))
		if len(result.Failed) > 0 {
			message += fmt.Sprintf(" (%d failed)", len(result.Failed))
		}

		return operationCompletedMsg{message: message}
	}
}

func (m *TUIModel) purgeAll() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		result, err := m.manager.PurgeAll(ctx, m.filter)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		message := fmt.Sprintf("Purged %d entries successfully", len(result.Successful))
		if len(result.Failed) > 0 {
			message += fmt.Sprintf(" (%d failed)", len(result.Failed))
		}

		return operationCompletedMsg{message: message}
	}
}

func (m *TUIModel) isSelected(id string) bool {
	for _, selectedID := range m.selectedIDs {
		if selectedID == id {
			return true
		}
	}
	return false
}

func (m *TUIModel) removeSelection(id string) {
	for i, selectedID := range m.selectedIDs {
		if selectedID == id {
			m.selectedIDs = append(m.selectedIDs[:i], m.selectedIDs[i+1:]...)
			break
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type entriesLoadedMsg struct {
	entries    []DLQEntry
	patterns   []ErrorPattern
	totalCount int
}

type statsLoadedMsg struct {
	stats *DLQStats
}

type operationCompletedMsg struct {
	message string
}

type errorMsg struct {
	error string
}