package timetraveldebugger

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUIModel represents the time travel debugger TUI state
type TUIModel struct {
	engine *ReplayEngine

	// Current state
	activeSession *ReplaySession
	currentState  *ReplayState
	recordings    []RecordMetadata

	// UI components
	recordingList  list.Model
	timeline       progress.Model
	stateViewer    viewport.Model
	diffViewer     viewport.Model
	searchInput    textinput.Model
	help           help.Model

	// Layout state
	width       int
	height      int
	focused     focusedPanel
	showHelp    bool
	showSearch  bool

	// Playback state
	isPlaying     bool
	playbackSpeed float64

	// Error handling
	errorMsg string

	// Styles
	styles TUIStyles
}

type focusedPanel int

const (
	focusRecordings focusedPanel = iota
	focusTimeline
	focusStateViewer
	focusDiffViewer
	focusSearch
)

// TUIStyles contains styling for the TUI
type TUIStyles struct {
	Panel       lipgloss.Style
	PanelTitle  lipgloss.Style
	Timeline    lipgloss.Style
	Event       lipgloss.Style
	EventActive lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style
	Diff        lipgloss.Style
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
}

// NewTUIModel creates a new TUI model for the time travel debugger
func NewTUIModel(engine *ReplayEngine) *TUIModel {
	// Create recording list
	recordingList := list.New([]list.Item{}, recordingDelegate{}, 0, 0)
	recordingList.Title = "Recordings"
	recordingList.SetShowStatusBar(false)
	recordingList.SetFilteringEnabled(true)

	// Create timeline progress bar
	timeline := progress.New(progress.WithDefaultGradient())

	// Create viewports
	stateViewer := viewport.New(0, 0)
	stateViewer.YPosition = 0

	diffViewer := viewport.New(0, 0)
	diffViewer.YPosition = 0

	// Create search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search recordings..."
	searchInput.CharLimit = 50

	// Create help
	helpModel := help.New()

	return &TUIModel{
		engine:        engine,
		recordingList: recordingList,
		timeline:      timeline,
		stateViewer:   stateViewer,
		diffViewer:    diffViewer,
		searchInput:   searchInput,
		help:          helpModel,
		focused:       focusRecordings,
		playbackSpeed: 1.0,
		styles:        defaultTUIStyles(),
	}
}

// Init initializes the TUI model
func (m *TUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadRecentRecordings(),
		textinput.Blink,
	)
}

// Update handles TUI updates
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case recordingsLoadedMsg:
		m.recordings = msg.recordings
		items := make([]list.Item, len(msg.recordings))
		for i, r := range msg.recordings {
			items[i] = recordingItem(r)
		}
		m.recordingList.SetItems(items)

	case sessionCreatedMsg:
		m.activeSession = msg.session
		return m, m.loadTimeline()

	case timelineLoadedMsg:
		m.timeline = progress.New(progress.WithDefaultGradient())
		return m, m.seekToPosition(0)

	case stateReconstructedMsg:
		m.currentState = msg.state
		m.updateStateViewer()
		m.updateDiffViewer()
		m.updateTimelinePosition()

	case errorMsg:
		m.errorMsg = string(msg)
	}

	// Update focused component
	switch m.focused {
	case focusRecordings:
		m.recordingList, cmd = m.recordingList.Update(msg)
		cmds = append(cmds, cmd)

	case focusStateViewer:
		m.stateViewer, cmd = m.stateViewer.Update(msg)
		cmds = append(cmds, cmd)

	case focusDiffViewer:
		m.diffViewer, cmd = m.diffViewer.Update(msg)
		cmds = append(cmds, cmd)

	case focusSearch:
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *TUIModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var sections []string

	// Header with title and controls
	header := m.renderHeader()
	sections = append(sections, header)

	// Main content area
	if m.activeSession == nil {
		// Show recording selection
		content := m.renderRecordingSelection()
		sections = append(sections, content)
	} else {
		// Show replay interface
		content := m.renderReplayInterface()
		sections = append(sections, content)
	}

	// Footer with help
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Key bindings
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Select   key.Binding
	Back     key.Binding
	Play     key.Binding
	Step     key.Binding
	Jump     key.Binding
	Bookmark key.Binding
	Search   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Play, k.Step, k.Jump, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Back, k.Play, k.Step},
		{k.Jump, k.Bookmark, k.Search, k.Help},
		{k.Quit},
	}
}

var keys = keyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "back")),
	Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "forward")),
	Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Play:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "play/pause")),
	Step:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "step")),
	Jump:     key.NewBinding(key.WithKeys("e", "r"), key.WithHelp("e/r", "jump error/retry")),
	Bookmark: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "bookmark")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// Handle key presses
func (m *TUIModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, keys.Search):
		if !m.showSearch {
			m.showSearch = true
			m.focused = focusSearch
			return m, m.searchInput.Focus()
		}

	case key.Matches(msg, keys.Back):
		if m.showSearch {
			m.showSearch = false
			m.focused = focusRecordings
			return m, nil
		}
		if m.activeSession != nil {
			m.activeSession = nil
			m.currentState = nil
			return m, m.loadRecentRecordings()
		}

	case key.Matches(msg, keys.Select):
		if m.focused == focusRecordings && m.activeSession == nil {
			if item, ok := m.recordingList.SelectedItem().(recordingItem); ok {
				return m, m.createSession(string(item))
			}
		}

	case key.Matches(msg, keys.Play):
		if m.activeSession != nil {
			return m, m.togglePlayback()
		}

	case key.Matches(msg, keys.Left):
		if m.activeSession != nil {
			return m, m.stepBackward()
		}

	case key.Matches(msg, keys.Right):
		if m.activeSession != nil {
			return m, m.stepForward()
		}

	case key.Matches(msg, keys.Step):
		if m.activeSession != nil {
			return m, m.stepForward()
		}

	case key.Matches(msg, keys.Jump):
		if m.activeSession != nil {
			switch msg.String() {
			case "e":
				return m, m.jumpToNextError()
			case "r":
				return m, m.jumpToNextRetry()
			}
		}

	case key.Matches(msg, keys.Bookmark):
		if m.activeSession != nil {
			return m, m.addBookmark()
		}
	}

	return m, nil
}

// Commands
type (
	recordingsLoadedMsg  struct{ recordings []RecordMetadata }
	sessionCreatedMsg    struct{ session *ReplaySession }
	timelineLoadedMsg    struct{}
	stateReconstructedMsg struct{ state *ReplayState }
	errorMsg             string
)

func (m *TUIModel) loadRecentRecordings() tea.Cmd {
	return func() tea.Msg {
		recordings, err := m.engine.GetRecentRecordings(nil, 50)
		if err != nil {
			return errorMsg(err.Error())
		}
		return recordingsLoadedMsg{recordings: recordings}
	}
}

func (m *TUIModel) createSession(recordID string) tea.Cmd {
	return func() tea.Msg {
		session, err := m.engine.CreateReplaySession(recordID, "tui-user")
		if err != nil {
			return errorMsg(err.Error())
		}
		return sessionCreatedMsg{session: session}
	}
}

func (m *TUIModel) loadTimeline() tea.Cmd {
	return func() tea.Msg {
		return timelineLoadedMsg{}
	}
}

func (m *TUIModel) seekToPosition(index int) tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		position := TimelinePosition{
			EventIndex: index,
		}

		state, err := m.engine.SeekTo(m.activeSession.ID, position)
		if err != nil {
			return errorMsg(err.Error())
		}
		return stateReconstructedMsg{state: state}
	}
}

func (m *TUIModel) stepForward() tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		state, err := m.engine.StepForward(m.activeSession.ID)
		if err != nil {
			return errorMsg(err.Error())
		}
		return stateReconstructedMsg{state: state}
	}
}

func (m *TUIModel) stepBackward() tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		state, err := m.engine.StepBackward(m.activeSession.ID)
		if err != nil {
			return errorMsg(err.Error())
		}
		return stateReconstructedMsg{state: state}
	}
}

func (m *TUIModel) jumpToNextError() tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		state, err := m.engine.JumpToNextError(m.activeSession.ID)
		if err != nil {
			return errorMsg(err.Error())
		}
		return stateReconstructedMsg{state: state}
	}
}

func (m *TUIModel) jumpToNextRetry() tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		state, err := m.engine.JumpToNextRetry(m.activeSession.ID)
		if err != nil {
			return errorMsg(err.Error())
		}
		return stateReconstructedMsg{state: state}
	}
}

func (m *TUIModel) addBookmark() tea.Cmd {
	return func() tea.Msg {
		if m.activeSession == nil {
			return errorMsg("no active session")
		}

		err := m.engine.AddBookmark(m.activeSession.ID, "User bookmark")
		if err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m *TUIModel) togglePlayback() tea.Cmd {
	return func() tea.Msg {
		m.isPlaying = !m.isPlaying
		// TODO: Implement auto-playback
		return nil
	}
}

// Rendering methods

func (m *TUIModel) renderHeader() string {
	title := m.styles.PanelTitle.Render("⏰ Time Travel Debugger")

	var controls strings.Builder
	if m.activeSession != nil {
		playSymbol := "▶"
		if m.isPlaying {
			playSymbol = "⏸"
		}
		controls.WriteString(fmt.Sprintf("%s %.1fx", playSymbol, m.playbackSpeed))

		if m.currentState != nil {
			controls.WriteString(fmt.Sprintf(" | Event %d", m.currentState.Position.EventIndex))
		}
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | "),
		controls.String(),
	)

	return m.styles.Panel.Width(m.width).Render(header)
}

func (m *TUIModel) renderRecordingSelection() string {
	if m.showSearch {
		search := lipgloss.JoinVertical(lipgloss.Left,
			"Search recordings:",
			m.searchInput.View(),
		)
		recordings := m.recordingList.View()
		return lipgloss.JoinVertical(lipgloss.Left, search, recordings)
	}

	return m.recordingList.View()
}

func (m *TUIModel) renderReplayInterface() string {
	// Three-panel layout: timeline (top), state (bottom-left), diff (bottom-right)
	timelineHeight := 3
	contentHeight := m.height - 6 // header + footer + timeline

	// Timeline at top
	timelineView := m.renderTimeline()

	// Split bottom area
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth

	// State viewer (left)
	stateTitle := "State"
	if m.currentState != nil && m.currentState.JobState != nil {
		stateTitle = fmt.Sprintf("State (Job: %s)", m.currentState.JobState.ID)
	}
	statePanel := m.styles.Panel.
		Width(leftWidth).
		Height(contentHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			m.styles.PanelTitle.Render(stateTitle),
			m.stateViewer.View(),
		))

	// Diff viewer (right)
	diffPanel := m.styles.Panel.
		Width(rightWidth).
		Height(contentHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			m.styles.PanelTitle.Render("Changes"),
			m.diffViewer.View(),
		))

	content := lipgloss.JoinHorizontal(lipgloss.Top, statePanel, diffPanel)

	return lipgloss.JoinVertical(lipgloss.Left, timelineView, content)
}

func (m *TUIModel) renderTimeline() string {
	if m.currentState == nil {
		return m.styles.Timeline.Render("No timeline data")
	}

	// Simple progress bar for now
	progress := 0.0
	if m.activeSession != nil {
		// This would need the total event count from the recording
		progress = 0.5 // placeholder
	}

	return m.styles.Timeline.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("Timeline - %s", m.currentState.Position.Description),
			m.timeline.ViewAs(progress),
		),
	)
}

func (m *TUIModel) renderFooter() string {
	if m.showHelp {
		return m.help.View(keys)
	}

	if m.errorMsg != "" {
		return m.styles.Error.Render(fmt.Sprintf("Error: %s", m.errorMsg))
	}

	return m.styles.Info.Render("Press ? for help, q to quit")
}

// Update layout when window size changes
func (m *TUIModel) updateLayout() {
	listHeight := m.height - 4 // header + footer
	m.recordingList.SetSize(m.width, listHeight)

	contentHeight := m.height - 8 // header + footer + timeline + margins
	contentWidth := m.width/2 - 2 // split screen with margins

	m.stateViewer.Width = contentWidth
	m.stateViewer.Height = contentHeight

	m.diffViewer.Width = contentWidth
	m.diffViewer.Height = contentHeight
}

func (m *TUIModel) updateStateViewer() {
	if m.currentState == nil || m.currentState.JobState == nil {
		m.stateViewer.SetContent("No state data available")
		return
	}

	var content strings.Builder

	// Job state
	content.WriteString("Job State:\n")
	content.WriteString(fmt.Sprintf("  ID: %s\n", m.currentState.JobState.ID))
	content.WriteString(fmt.Sprintf("  Status: %s\n", m.currentState.JobState.Status))
	content.WriteString(fmt.Sprintf("  Retries: %d/%d\n", m.currentState.JobState.Retries, m.currentState.JobState.MaxRetries))
	content.WriteString(fmt.Sprintf("  Priority: %s\n", m.currentState.JobState.Priority))

	if m.currentState.JobState.ErrorMessage != "" {
		content.WriteString(fmt.Sprintf("  Error: %s\n", m.currentState.JobState.ErrorMessage))
	}

	// System state
	if m.currentState.SystemState != nil && m.currentState.SystemState.SystemMetrics != nil {
		content.WriteString("\nSystem Metrics:\n")
		metrics := m.currentState.SystemState.SystemMetrics
		content.WriteString(fmt.Sprintf("  Queue Length: %d\n", metrics.QueueLength))
		content.WriteString(fmt.Sprintf("  Worker Count: %d\n", metrics.WorkerCount))
		content.WriteString(fmt.Sprintf("  Memory: %.1f MB\n", metrics.MemoryUsageMB))
		content.WriteString(fmt.Sprintf("  CPU: %.1f%%\n", metrics.CPUUsagePercent))
	}

	m.stateViewer.SetContent(content.String())
}

func (m *TUIModel) updateDiffViewer() {
	if m.currentState == nil || len(m.currentState.Changes) == 0 {
		m.diffViewer.SetContent("No changes to display")
		return
	}

	var content strings.Builder
	content.WriteString("State Changes:\n\n")

	for _, change := range m.currentState.Changes {
		content.WriteString(fmt.Sprintf("%s.%s:\n", change.Component, change.Key))
		content.WriteString(fmt.Sprintf("  Before: %v\n", change.Before))
		content.WriteString(fmt.Sprintf("  After:  %v\n", change.After))
		content.WriteString("\n")
	}

	m.diffViewer.SetContent(content.String())
}

func (m *TUIModel) updateTimelinePosition() {
	// Update timeline progress based on current position
	// This would calculate the actual progress through the recording
}

// Recording list item
type recordingItem RecordMetadata

func (r recordingItem) FilterValue() string {
	return r.RecordID
}

func (r recordingItem) Title() string {
	return fmt.Sprintf("Record %s", r.RecordID[:8])
}

func (r recordingItem) Description() string {
	return fmt.Sprintf("%s - %s (%d events)", r.Reason, r.CreatedAt.Format("Jan 2 15:04"), r.EventCount)
}

// Recording list delegate
type recordingDelegate struct{}

func (d recordingDelegate) Height() int                               { return 2 }
func (d recordingDelegate) Spacing() int                              { return 1 }
func (d recordingDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d recordingDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	r, ok := listItem.(recordingItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, r.Title())
	desc := r.Description()

	if index == m.Index() {
		str = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(str)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", str, desc)
}

// Default styles
func defaultTUIStyles() TUIStyles {
	return TUIStyles{
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
		PanelTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true),
		Timeline: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
		Event: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		EventActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		Diff: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		DiffAdd: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		DiffRemove: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
	}
}

