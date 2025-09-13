// Copyright 2025 James Ross
package tui

import (
    "fmt"
    "strings"
    "time"
    "sort"

    bubprog "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/lithammer/fuzzysearch/fuzzy"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

// Simple, pragmatic TUI for observing and administering the queue system.

// types and messages moved to model.go

// initialModel moved to init.go

func (m model) Init() tea.Cmd {
	// Start with an immediate refresh and ticking
	return tea.Batch(m.refreshCmd(), tea.Every(m.refreshEvery, func(time.Time) tea.Msg { return tick{} }), spinner.Tick)
}

// command helpers moved to commands.go

// TODO: enqueue helper will be added below

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// When a confirmation modal is open, only handle confirm/cancel keys
		if m.confirmOpen {
			switch msg.String() {
			case "y", "enter":
				// Run the pending action
				if m.confirmAction == "quit" {
					m.confirmOpen = false
					m.cancel()
					return m, tea.Quit
				}
				switch m.confirmAction {
				case "purge-dlq":
					m.loading = true
					m.errText = ""
					m.confirmOpen = false
					cmds = append(cmds, func() tea.Msg {
						err := admin.PurgeDLQ(m.ctx, m.cfg, m.rdb)
						if err != nil {
							return statsMsg{err: err}
						}
						return statsMsg{}
					}, spinner.Tick, m.refreshCmd(), m.fetchKeysCmd())
				case "purge-all":
					m.loading = true
					m.errText = ""
					m.confirmOpen = false
					cmds = append(cmds, func() tea.Msg {
						_, err := admin.PurgeAll(m.ctx, m.cfg, m.rdb)
						if err != nil {
							return statsMsg{err: err}
						}
						return statsMsg{}
					}, spinner.Tick, m.refreshCmd(), m.fetchKeysCmd())
				}
			case "n", "esc":
				m.confirmOpen = false
			case "q", "ctrl+c":
				// allow quitting from modal too
				m.cancel()
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			// Ask for confirmation to quit
			m.confirmOpen = true
			m.confirmAction = "quit"
			return m, nil
		case "tab":
			// Cycle focus across panels
			m.focus = (m.focus + 1) % 3
			return m, nil
		case "shift+tab":
			if m.focus == 0 {
				m.focus = 2
			} else {
				m.focus--
			}
			return m, nil
		case "r":
			return m, tea.Batch(m.refreshCmd(), m.fetchKeysCmd())
		case "h", "?":
			m.help2.SetIsActive(!m.help2.Active)
			if m.help2.Active {
				m.help2.GotoTop()
			}
			return m, nil
		case "f", "/":
			if m.focus == focusQueues {
				m.filterActive = true
				m.filter.Focus()
			}
		case "p":
			if len(m.peekTargets) > 0 {
				i := m.tbl.Cursor()
				if i >= 0 && i < len(m.peekTargets) {
					m.loading = true
					m.errText = ""
					cmds = append(cmds, m.doPeekCmd(m.peekTargets[i], 10), spinner.Tick)
				}
			}
		case "b":
			// open bench form in inline info viewport
			// focus remains unchanged; enter runs
			// Focus first input
			m.benchCount.Focus()
		case "enter":
			// If bench inputs focused, run bench
			if m.benchCount.Focused() || m.benchRate.Focused() || m.benchPriority.Focused() || m.benchTimeout.Focused() {
				// Parse inputs and run
				count := atoiDefault(m.benchCount.Value(), 1000)
				rate := atoiDefault(m.benchRate.Value(), 500)
				prio := strings.TrimSpace(m.benchPriority.Value())
				if prio == "" {
					prio = m.cfg.Producer.DefaultPriority
				}
				to := time.Duration(atoiDefault(m.benchTimeout.Value(), 60)) * time.Second
				m.loading = true
				m.errText = ""
				m.pbActive = true
				m.pbTotal = count
				cmds = append(cmds, m.doBenchCmd(prio, count, rate, to), spinner.Tick)
				// start progress polling
				cmds = append(cmds, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return benchPollTick{} }))
			}
		case "esc":
			// Clear filter/modal or blur bench inputs
			if m.confirmOpen {
				m.confirmOpen = false
			} else if m.filterActive {
				m.filterActive = false
				m.filter.SetValue("")
				m.applyFilterAndSetRows()
			}
		case "D":
			// Open confirmation modal for DLQ purge
			m.confirmOpen = true
			m.confirmAction = "purge-dlq"
		case "A":
			// Open confirmation modal for ALL purge
			m.confirmOpen = true
			m.confirmAction = "purge-all"
		}

		// Navigate bench inputs
		if m.benchCount.Focused() || m.benchRate.Focused() || m.benchPriority.Focused() || m.benchTimeout.Focused() {
			switch msg.String() {
			case "tab", "shift+tab":
				cycleBenchFocus(&m)
			}
			var c tea.Cmd
			m.benchCount, c = m.benchCount.Update(msg)
			cmds = append(cmds, c)
			m.benchRate, c = m.benchRate.Update(msg)
			cmds = append(cmds, c)
			m.benchPriority, c = m.benchPriority.Update(msg)
			cmds = append(cmds, c)
			m.benchTimeout, c = m.benchTimeout.Update(msg)
			cmds = append(cmds, c)
		}

		// Update filter input when active
		if m.filterActive {
			var c tea.Cmd
			m.filter, c = m.filter.Update(msg)
			cmds = append(cmds, c)
			m.applyFilterAndSetRows()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Layout calculations
		headerLines := 3 // header, sub, blank
		if m.filterActive || strings.TrimSpace(m.filter.Value()) != "" {
			headerLines++
		}
        footerLines := 1 + 1 // statusbar(1 line) and a blank
		availH := m.height - headerLines - footerLines
		if availH < 6 {
			availH = 6
		}
		// Split into top row (queues + charts) and bottom row (info)
		bottomH := availH / 3
		topH := availH - bottomH
		// Set widths
		leftW := m.width / 2
		if leftW < 30 {
			leftW = 30
		}
		rightW := m.width - leftW - 3
		if rightW < 20 {
			rightW = 20
		}
		// Apply sizes
		m.tbl.SetWidth(leftW - 4)
		m.tbl.SetHeight(topH - 3)
		m.vpCharts.Width = rightW - 2
		m.vpCharts.Height = topH - 2
		m.vpInfo.Width = m.width - 4
		m.vpInfo.Height = bottomH - 2

		// Resize status bar and help
		m.sb.SetSize(m.width)
		hw := m.width - 10
		if hw < 40 {
			hw = m.width - 2
		}
		hh := m.height - 6
		if hh < 6 {
			hh = m.height - 2
		}
		m.help2.SetSize(hw, hh)
		if m.width > 0 {
			m.pb.Width = m.width - 20
		}
	case tea.MouseMsg:
		if !m.confirmOpen {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				if msg.Action == tea.MouseActionPress {
					switch m.focus {
					case focusQueues:
						m.tbl.MoveUp(1)
					case focusCharts:
						m.vpCharts.LineUp(1)
					case focusInfo:
						m.vpInfo.LineUp(1)
					}
				}
			case tea.MouseButtonWheelDown:
				if msg.Action == tea.MouseActionPress {
					switch m.focus {
					case focusQueues:
						m.tbl.MoveDown(1)
					case focusCharts:
						m.vpCharts.LineDown(1)
					case focusInfo:
						m.vpInfo.LineDown(1)
					}
				}
			case tea.MouseButtonLeft:
				if msg.Action == tea.MouseActionPress {
					// Left click peeks current queue
					if len(m.peekTargets) > 0 {
						i := m.tbl.Cursor()
						if i >= 0 && i < len(m.peekTargets) {
							m.loading = true
							m.errText = ""
							cmds = append(cmds, m.doPeekCmd(m.peekTargets[i], 10), spinner.Tick)
						}
					}
				}
			}
		}
	case tick:
		cmds = append(cmds, m.refreshCmd(), m.fetchKeysCmd(), tea.Every(m.refreshEvery, func(time.Time) tea.Msg { return tick{} }))
	case statsMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
		} else {
			m.lastStats = msg.s
			m.errText = ""
			// Update series with latest counts per known queues
			m.addSample("high", m.cfg.Worker.Queues["high"], msg.s)
			m.addSample("low", m.cfg.Worker.Queues["low"], msg.s)
			m.addSample("completed", m.cfg.Worker.CompletedList, msg.s)
			m.addSample("dead_letter", m.cfg.Worker.DeadLetterList, msg.s)
			// Update table with queues only (friendly names)
			rows := []table.Row{}
			m.peekTargets = m.peekTargets[:0]
			// Show configured queues and special lists for reliable ordering
			ordered := make([]string, 0, len(m.cfg.Worker.Queues)+2)
			// priorities first
			for _, p := range m.cfg.Worker.Priorities {
				key := m.cfg.Worker.Queues[p]
				display := fmt.Sprintf("%s (%s)", p, key)
				ordered = append(ordered, display)
			}
			ordered = append(ordered, fmt.Sprintf("completed (%s)", m.cfg.Worker.CompletedList))
			ordered = append(ordered, fmt.Sprintf("dead_letter (%s)", m.cfg.Worker.DeadLetterList))

			for _, display := range ordered {
				cnt := msg.s.Queues[display]
				rows = append(rows, table.Row{display, fmt.Sprintf("%d", cnt)})
				// extract the key inside parentheses for peek target
				if idx := strings.LastIndex(display, "("); idx != -1 && strings.HasSuffix(display, ")") {
					m.peekTargets = append(m.peekTargets, display[idx+1:len(display)-1])
				} else {
					m.peekTargets = append(m.peekTargets, display)
				}
			}
			m.allRows = rows
			m.allTargets = append([]string(nil), m.peekTargets...)
			m.applyFilterAndSetRows()
			if m.tbl.Cursor() >= len(rows) && len(rows) > 0 {
				m.tbl.SetCursor(len(rows) - 1)
			}
		}
		m.loading = false
	case keysMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
		} else {
			m.lastKeys = msg.k
			m.errText = ""
		}
	case peekMsg:
		m.loading = false
		if msg.err != nil {
			m.errText = msg.err.Error()
		} else {
			m.lastPeek = msg.p
		}
	case benchMsg:
		m.loading = false
		if msg.err != nil {
			m.errText = msg.err.Error()
		} else {
			m.lastBench = msg.b
			m.pbActive = false
			if cmd := m.pb.SetPercent(1.0); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case benchPollTick:
		if m.pbActive {
			cmds = append(cmds, m.benchPollCmd(), tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return benchPollTick{} }))
		}
	case benchProgMsg:
		if m.pbTotal > 0 {
			percent := float64(msg.done) / float64(m.pbTotal)
			if percent > 1 {
				percent = 1
			}
			if cmd := m.pb.SetPercent(percent); cmd != nil {
				cmds = append(cmds, cmd)
			}
			if percent >= 1 {
				m.pbActive = false
			}
		}
	}

	// Spinner update when loading
	if m.loading {
		var c tea.Cmd
		m.spinner, c = m.spinner.Update(msg)
		cmds = append(cmds, c)
	}
	// Progress animate update
	{
		md, c := m.pb.Update(msg)
		if pm, ok := md.(bubprog.Model); ok {
			m.pb = pm
		}
		cmds = append(cmds, c)
	}
	// Always update table on dashboard
	{
		var c tea.Cmd
		m.tbl, c = m.tbl.Update(msg)
		cmds = append(cmds, c)
	}

	return m, tea.Batch(cmds...)
}

// View moved to view.go

// moved to view.go: summarizeKeys

// moved to view.go: renderKeys

// moved to view.go: renderPeek

// moved to view.go: renderBenchForm

// moved to view.go: renderBenchResult

// moved to view.go: helpBar

// moved to view.go: focusName

func cycleBenchFocus(m *model) {
	if m.benchCount.Focused() {
		m.benchCount.Blur()
		m.benchRate.Focus()
		return
	}
	if m.benchRate.Focused() {
		m.benchRate.Blur()
		m.benchPriority.Focus()
		return
	}
	if m.benchPriority.Focused() {
		m.benchPriority.Blur()
		m.benchTimeout.Focus()
		return
	}
	m.benchTimeout.Blur()
	m.benchCount.Focus()
}

func atoiDefault(s string, def int) int {
	var v int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
	if err != nil {
		return def
	}
	return v
}

// renderConfirmModal moved to overlays.go

// renderHelpOverlay2 (legacy duplicate) dims the base and centers the teacup help view.
func renderHelpOverlay2(m model, base string) string {
	dim := lipgloss.NewStyle().Faint(true).Render(base)
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	scrimCell := lipgloss.NewStyle().Background(lipgloss.Color("236")).Faint(true).Render(" ")
	line := strings.Repeat(scrimCell, width)
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		lines[i] = line
	}

	hv := m.help2.View()
	hvLines := strings.Split(hv, "\n")
	hH := len(hvLines)
	hW := 0
	for _, l := range hvLines {
		if w := lipgloss.Width(l); w > hW {
			hW = w
		}
	}
	top := (height - hH) / 2
	left := (width - hW) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}
	for i := 0; i < hH && (top+i) < height; i++ {
		ml := hvLines[i]
		lp := left
		rp := width - (left + lipgloss.Width(ml))
		if lp < 0 {
			lp = 0
		}
		if rp < 0 {
			rp = 0
		}
		leftPad := strings.Repeat(scrimCell, lp)
		rightPad := strings.Repeat(scrimCell, rp)
		lines[top+i] = leftPad + ml + rightPad
	}
	return dim + "\n" + strings.Join(lines, "\n")
}

// staticStringModel is a tiny tea.Model wrapper around a fixed string view.
// staticStringModel moved to overlays.go

// bench progress ticking
// moved to model.go: benchPollTick, benchProgMsg

// benchPollCmd moved to commands.go

// renderOverlayScreen builds a full-screen dimmed scrim and draws the modal
// centered on top of it. This replaces the regular view while the modal is open
// for strong contrast.
func renderOverlayScreen(m model) string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Scrim style: faint + subtle background.
	scrimCell := lipgloss.NewStyle().Background(lipgloss.Color("236")).Faint(true).Render(" ")
	line := strings.Repeat(scrimCell, width)
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		lines[i] = line
	}

	// Modal content and dimensions.
	modal := renderConfirmModal(m)
	modalLines := strings.Split(modal, "\n")
	modalH := len(modalLines)
	modalW := 0
	for _, ml := range modalLines {
		if w := lipgloss.Width(ml); w > modalW {
			modalW = w
		}
	}

	// Center modal within the scrim.
	top := (height - modalH) / 2
	left := (width - modalW) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}

	// Overlay modal lines into the scrim.
	for i := 0; i < modalH && (top+i) < height; i++ {
		ml := modalLines[i]
		// Base line broken into three parts: left padding, modal, right padding.
		lp := left
		rp := width - (left + lipgloss.Width(ml))
		if lp < 0 {
			lp = 0
		}
		if rp < 0 {
			rp = 0
		}
		leftPad := strings.Repeat(scrimCell, lp)
		rightPad := strings.Repeat(scrimCell, rp)
		lines[top+i] = leftPad + ml + rightPad
	}

	return strings.Join(lines, "\n")
}

// addSample appends a value to a named series using StatsResult map.
func (m *model) addSample(alias, key string, s admin.StatsResult) {
	if alias == "" || key == "" {
		return
	}
	display := fmt.Sprintf("%s (%s)", alias, key)
	val := s.Queues[display]
	arr := m.series[alias]
	arr = append(arr, float64(val))
	if len(arr) > m.seriesMax {
		arr = arr[len(arr)-m.seriesMax:]
	}
	m.series[alias] = arr
}

// moved to view.go: renderCharts

// moved to view.go: renderFilterBar

func (m *model) applyFilterAndSetRows() {
	q := strings.TrimSpace(m.filter.Value())
	if q == "" {
		m.tbl.SetRows(m.allRows)
		m.peekTargets = append([]string(nil), m.allTargets...)
		return
	}
	labels := make([]string, len(m.allRows))
	for i, r := range m.allRows {
		labels[i] = r[0]
	}
	ranks := fuzzy.RankFindNormalizedFold(q, labels)
	sort.Sort(ranks)
	rows := make([]table.Row, 0, len(ranks))
	targets := make([]string, 0, len(ranks))
	for _, rk := range ranks {
		rows = append(rows, m.allRows[rk.OriginalIndex])
		targets = append(targets, m.allTargets[rk.OriginalIndex])
	}
	m.tbl.SetRows(rows)
	m.peekTargets = targets
}

// clamp limits v to the inclusive range [low, high].
func clamp(v, low, high int) int {
	if high < low {
		return low
	}
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
