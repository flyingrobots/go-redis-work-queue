package tui

import (
	"fmt"
	"strings"
	"time"

	bubprog "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), tea.Every(m.refreshEvery, func(time.Time) tea.Msg { return tick{} }), spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmOpen {
			switch msg.String() {
			case "y", "enter":
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
						if err := admin.PurgeDLQ(m.ctx, m.cfg, m.rdb); err != nil {
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
				m.cancel()
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.confirmOpen = true
			m.confirmAction = "quit"
			return m, nil
		case "tab":
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
			m.benchCount.Focus()
		case "enter":
			if m.benchCount.Focused() || m.benchRate.Focused() || m.benchPriority.Focused() || m.benchTimeout.Focused() {
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
				cmds = append(cmds, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return benchPollTick{} }))
			}
		case "esc":
			if m.confirmOpen {
				m.confirmOpen = false
			} else if m.filterActive {
				m.filterActive = false
				m.filter.SetValue("")
				m.applyFilterAndSetRows()
			}
		case "D":
			m.confirmOpen = true
			m.confirmAction = "purge-dlq"
		case "A":
			m.confirmOpen = true
			m.confirmAction = "purge-all"
		}
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
		if m.filterActive {
			var c tea.Cmd
			m.filter, c = m.filter.Update(msg)
			cmds = append(cmds, c)
			m.applyFilterAndSetRows()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerLines := 3
		if m.filterActive || strings.TrimSpace(m.filter.Value()) != "" {
			headerLines++
		}
		footerLines := 2
		availH := m.height - headerLines - footerLines
		if availH < 6 {
			availH = 6
		}
		bottomH := availH / 3
		topH := availH - bottomH
		leftW := m.width / 2
		if leftW < 30 {
			leftW = 30
		}
		rightW := m.width - leftW - 3
		if rightW < 20 {
			rightW = 20
		}
		m.tbl.SetWidth(leftW - 4)
		m.tbl.SetHeight(topH - 3)
		m.vpCharts.Width = rightW - 2
		m.vpCharts.Height = topH - 2
		m.vpInfo.Width = m.width - 4
		m.vpInfo.Height = bottomH - 2
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
			m.addSample("high", m.cfg.Worker.Queues["high"], msg.s)
			m.addSample("low", m.cfg.Worker.Queues["low"], msg.s)
			m.addSample("completed", m.cfg.Worker.CompletedList, msg.s)
			m.addSample("dead_letter", m.cfg.Worker.DeadLetterList, msg.s)
			rows := []table.Row{}
			m.peekTargets = m.peekTargets[:0]
			ordered := make([]string, 0, len(m.cfg.Worker.Queues)+2)
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

	if m.loading {
		var c tea.Cmd
		m.spinner, c = m.spinner.Update(msg)
		cmds = append(cmds, c)
	}
	{
		md, c := m.pb.Update(msg)
		if pm, ok := md.(bubprog.Model); ok {
			m.pb = pm
		}
		cmds = append(cmds, c)
	}
	{
		var c tea.Cmd
		m.tbl, c = m.tbl.Update(msg)
		cmds = append(cmds, c)
	}

	return m, tea.Batch(cmds...)
}
