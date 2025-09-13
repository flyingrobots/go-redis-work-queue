// Copyright 2025 James Ross
package tui

import (
    "context"
    "encoding/json"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/help"
    bubprog "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    asciigraph "github.com/guptarohit/asciigraph"
    "github.com/lithammer/fuzzysearch/fuzzy"
    tchelp "github.com/mistakenelf/teacup/help"
    "github.com/mistakenelf/teacup/statusbar"
    overlay "github.com/rmhubbert/bubbletea-overlay"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

// Simple, pragmatic TUI for observing and administering the queue system.

// types and messages moved to model.go

func initialModel(cfg *config.Config, rdb *redis.Client, logger *zap.Logger, refreshEvery time.Duration) model {
	ctx, cancel := context.WithCancel(context.Background())

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Setup table
	columns := []table.Column{{Title: "Queue", Width: 40}, {Title: "Count", Width: 10}}
	t := table.New(table.WithColumns(columns), table.WithFocused(true))
	t.KeyMap.LineUp.SetKeys("k", "up")
	t.KeyMap.LineDown.SetKeys("j", "down")
	t.KeyMap.PageDown.SetKeys("ctrl+f")
	t.KeyMap.PageUp.SetKeys("ctrl+b")
	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Bold(true),
		Selected: lipgloss.NewStyle().Bold(true),
	})

	// Bench inputs defaults
	bi := textinput.New()
	bi.Placeholder = "count"
	bi.SetValue("1000")
	br := textinput.New()
	br.Placeholder = "rate"
	br.SetValue("500")
	bp := textinput.New()
	bp.Placeholder = "priority"
	bp.SetValue(cfg.Producer.DefaultPriority)
	bt := textinput.New()
	bt.Placeholder = "timeout (s)"
	bt.SetValue("60")

	// Filter input
	fi := textinput.New()
	fi.Placeholder = "filter"
	fi.CharLimit = 64

	// Panels styles
	boxTitle := lipgloss.NewStyle().Bold(true)
	boxBody := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	// Status bar colors and model
	sb := statusbar.New(
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#000000", Light: "#ffffff"}, Background: lipgloss.AdaptiveColor{Dark: "#ffaa00", Light: "#111111"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}, Background: lipgloss.AdaptiveColor{Dark: "#333333", Light: "#eeeeee"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}, Background: lipgloss.AdaptiveColor{Dark: "#444444", Light: "#dddddd"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#000000", Light: "#ffffff"}, Background: lipgloss.AdaptiveColor{Dark: "#55dd55", Light: "#226622"}},
	)

	// Teacup help entries
	entries := []tchelp.Entry{
		{Key: "q", Description: "Quit"},
		{Key: "tab/shift+tab", Description: "Focus next/prev panel"},
		{Key: "j/k, wheel", Description: "Scroll selected panel"},
		{Key: "f or /", Description: "Filter queues (fuzzy)"},
		{Key: "p", Description: "Peek selected queue"},
		{Key: "b", Description: "Bench form (enter to run)"},
		{Key: "D / A", Description: "Purge DLQ / ALL (y/n)"},
		{Key: "h/?", Description: "Toggle help"},
	}
	help2 := tchelp.New(false, false, "Help",
		tchelp.TitleColor{Background: lipgloss.AdaptiveColor{Dark: "#444444", Light: "#DDDDDD"}, Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}},
		lipgloss.AdaptiveColor{Dark: "#888888", Light: "#222222"}, entries)

	return model{
		ctx:           ctx,
		cancel:        cancel,
		cfg:           cfg,
		rdb:           rdb,
		logger:        logger,
		focus:         focusQueues,
		help:          help.New(),
		spinner:       sp,
		tbl:           t,
		benchCount:    bi,
		benchRate:     br,
		benchPriority: bp,
		benchTimeout:  bt,
		refreshEvery:  refreshEvery,
		tableTopY:     3, // header + sub + blank line
		series:        map[string][]float64{"high": {}, "low": {}, "completed": {}, "dead_letter": {}},
		seriesMax:     180, // keep last N points
		filter:        fi,
		vpCharts:      viewport.New(0, 10),
		vpInfo:        viewport.New(0, 10),
		boxTitle:      boxTitle,
		boxBody:       boxBody,
		sb:            sb,
		help2:         help2,
		pb:            bubprog.New(bubprog.WithDefaultGradient()),
	}
}

func (m model) Init() tea.Cmd {
	// Start with an immediate refresh and ticking
	return tea.Batch(m.refreshCmd(), tea.Every(m.refreshEvery, func(time.Time) tea.Msg { return tick{} }), spinner.Tick)
}

func (m model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		// Fetch stats and keys sequentially (fast ops)
		s, err := admin.Stats(m.ctx, m.cfg, m.rdb)
		if err != nil {
			return statsMsg{err: err}
		}
		return statsMsg{s: s, err: nil}
	}
}

func (m model) fetchKeysCmd() tea.Cmd {
	return func() tea.Msg {
		k, err := admin.StatsKeys(m.ctx, m.cfg, m.rdb)
		return keysMsg{k: k, err: err}
	}
}

func (m model) doPeekCmd(target string, n int) tea.Cmd {
	return func() tea.Msg {
		p, err := admin.Peek(m.ctx, m.cfg, m.rdb, target, int64(n))
		return peekMsg{p: p, err: err}
	}
}

func (m model) doBenchCmd(priority string, count, rate int, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		b, err := admin.Bench(m.ctx, m.cfg, m.rdb, priority, count, rate, timeout)
		return benchMsg{b: b, err: err}
	}
}

func (m model) doEnqueueCmd(queueKey string, count int) tea.Cmd {
	return func() tea.Msg {
		return enqueueMsg{n: 0, key: queueKey, err: nil}
	}
}

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
		footerLines := statusbar.Height + 1 // statusbar and a blank
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

func (m model) View() string {
	// Header
	header := lipgloss.NewStyle().Bold(true).Render("Job Queue TUI — Redis " + m.cfg.Redis.Addr)
	sub := fmt.Sprintf("Focus: %s  |  Heartbeats: %d  |  Processing lists: %d",
		focusName(m.focus), m.lastStats.Heartbeats, len(m.lastStats.ProcessingLists))
	if m.errText != "" {
		sub += "  |  Error: " + m.errText
	}
	if m.loading {
		sub += "  " + m.spinner.View()
	}

	// Panels content
	fb := renderFilterBar(m)
	left := m.tbl.View()
	if fb != "" {
		left = fb + "\n" + left
	}
	left = m.boxBody.Render(m.boxTitle.Render("Queues") + "\n" + left)

	// Charts panel content
	m.vpCharts.SetContent(renderCharts(m))
	right := m.boxBody.Render(m.boxTitle.Render("Charts") + "\n" + m.vpCharts.View())

	// Info panel: keys summary + optional peek + optional bench form/result
	info := summarizeKeys(m.lastKeys)
	if len(m.lastPeek.Items) > 0 {
		info += "\n\n" + renderPeek(m.lastPeek)
	}
	if m.benchCount.Focused() || m.benchRate.Focused() || m.benchPriority.Focused() || m.benchTimeout.Focused() || m.lastBench.Count > 0 {
		info += "\n\n" + renderBenchForm(m)
		if m.lastBench.Count > 0 {
			info += "\n" + renderBenchResult(m.lastBench)
		}
	}
	if m.pbActive && m.pbTotal > 0 {
		info += "\n\nBench Progress:\n" + m.pb.View()
	}
	m.vpInfo.SetContent(info)
	bottom := m.boxBody.Render(m.boxTitle.Render("Info") + "\n" + m.vpInfo.View())

	// Side-by-side top row
	gap := lipgloss.NewStyle().Width(2).Render(" ")
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	body := topRow + "\n" + bottom

	// Compose base view
	base := header + "\n" + sub + "\n\n" + body

	if m.confirmOpen {
		// Dim base and overlay centered confirm modal using overlay lib
		baseDim := lipgloss.NewStyle().Faint(true).Render(base)
		back := staticStringModel{baseDim}
		fore := staticStringModel{renderConfirmModal(m)}
		ov := overlay.New(fore, back, overlay.Center, overlay.Center, 0, 0)
		return ov.View()
	}
	// Append status bar and optional help overlay
	now := time.Now().Format("15:04:05")
	m.sb.SetContent("Redis "+m.cfg.Redis.Addr, "focus:"+focusName(m.focus), m.spinner.View(), now)
	out := base + "\n" + m.sb.View()
	if m.help2.Active {
		baseDim := lipgloss.NewStyle().Faint(true).Render(out)
		back := staticStringModel{baseDim}
		fore := staticStringModel{m.help2.View()}
		ov := overlay.New(fore, back, overlay.Center, overlay.Center, 0, 0)
		out = ov.View()
	}
	return out
}

func summarizeKeys(k admin.KeysStats) string {
	// Show totals and rate limiter info
	parts := []string{
		fmt.Sprintf("processing_lists=%d", k.ProcessingLists),
		fmt.Sprintf("processing_items=%d", k.ProcessingItems),
		fmt.Sprintf("heartbeats=%d", k.Heartbeats),
	}
	if k.RateLimitKey != "" {
		rl := "rate_limit_key=" + k.RateLimitKey
		if k.RateLimitTTL != "" {
			rl += " ttl=" + k.RateLimitTTL
		}
		parts = append(parts, rl)
	}
	return strings.Join(parts, "  |  ")
}

func renderKeys(k admin.KeysStats) string {
	// Deterministic order
	keys := make([]string, 0, len(k.QueueLengths))
	for name := range k.QueueLengths {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	b := &strings.Builder{}
	fmt.Fprintf(b, "Queue Lengths:\n")
	for _, name := range keys {
		fmt.Fprintf(b, "  %-40s %8d\n", name, k.QueueLengths[name])
	}
	fmt.Fprintf(b, "\nProcessing lists: %d\nProcessing items: %d\nHeartbeats: %d\n",
		k.ProcessingLists, k.ProcessingItems, k.Heartbeats)
	if k.RateLimitKey != "" {
		fmt.Fprintf(b, "Rate limit key: %s  TTL: %s\n", k.RateLimitKey, k.RateLimitTTL)
	}
	return b.String()
}

func renderPeek(p admin.PeekResult) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Peek: %s\n", p.Queue)
	if len(p.Items) == 0 {
		fmt.Fprintf(b, "(no items)\n")
		return b.String()
	}
	// Show items prettified if JSON
	for i := len(p.Items) - 1; i >= 0; i-- { // show newest at bottom visually
		it := p.Items[i]
		var v map[string]any
		if json.Unmarshal([]byte(it), &v) == nil {
			pp, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintf(b, "[%d]\n%s\n\n", i, string(pp))
		} else {
			fmt.Fprintf(b, "[%d] %s\n", i, it)
		}
	}
	return b.String()
}

func renderBenchForm(m model) string {
	// Simple inline form
	return strings.Join([]string{
		"Bench (enter to run, esc to back):",
		fmt.Sprintf("  Count:    %s", m.benchCount.View()),
		fmt.Sprintf("  Rate/s:   %s", m.benchRate.View()),
		fmt.Sprintf("  Priority: %s", m.benchPriority.View()),
		fmt.Sprintf("  Timeout:  %s seconds", m.benchTimeout.View()),
	}, "\n")
}

func renderBenchResult(b admin.BenchResult) string {
	if b.Count == 0 {
		return ""
	}
	return fmt.Sprintf("Bench: count=%d  duration=%s  thr=%.1f/s  p50=%s  p95=%s",
		b.Count, b.Duration.Truncate(time.Millisecond), b.Throughput, b.P50.Truncate(time.Millisecond), b.P95.Truncate(time.Millisecond))
}

func helpBar() string {
	return strings.Join([]string{
		"q:quit",
		"tab/shift+tab:focus panel",
		"r:refresh",
		"j/k:down/up",
		"wheel/mouse: scroll/select",
		"enter/p:peek",
		"b:bench form",
		"f:filter (queues)",
		"D:purge DLQ (y/n)",
		"A:purge ALL (y/n)",
	}, "  ")
}

func focusName(f focusArea) string {
	switch f {
	case focusQueues:
		return "Queues"
	case focusCharts:
		return "Charts"
	case focusInfo:
		return "Info"
	default:
		return "?"
	}
}

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

	// Centered width
	width := m.width
	if width <= 0 {
		width = 80
	}
	modal := box.Render(content)
	// Center horizontally by padding spaces
	pad := 0
	if w := lipgloss.Width(modal); width > w {
		pad = (width - w) / 2
	}
	return strings.Repeat(" ", pad) + modal
}

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
type staticStringModel struct{ s string }

func (s staticStringModel) Init() tea.Cmd                           { return nil }
func (s staticStringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s staticStringModel) View() string                            { return s.s }

// bench progress ticking
// moved to model.go: benchPollTick, benchProgMsg

func (m model) benchPollCmd() tea.Cmd {
	return func() tea.Msg {
		n, _ := m.rdb.LLen(m.ctx, m.cfg.Worker.CompletedList).Result()
		return benchProgMsg{done: n}
	}
}

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

func renderCharts(m model) string {
	width := m.width
	if width <= 10 {
		width = 80
	}
	plotW := width - 2
	if plotW < 10 {
		plotW = 10
	}
	// Height per chart
	h := 8
	makePlot := func(title string, data []float64) string {
		if len(data) == 0 {
			return fmt.Sprintf("%s\n(no data yet)", title)
		}
		g := asciigraph.Plot(data,
			asciigraph.Height(h),
			asciigraph.Width(plotW),
			asciigraph.Caption(title),
		)
		return g
	}
	parts := []string{}
	parts = append(parts, makePlot("High Priority", m.series["high"]))
	parts = append(parts, makePlot("Low Priority", m.series["low"]))
	parts = append(parts, makePlot("Completed", m.series["completed"]))
	parts = append(parts, makePlot("Dead Letter", m.series["dead_letter"]))
	return strings.Join(parts, "\n\n")
}

func renderFilterBar(m model) string {
	if m.filterActive {
		return "Filter: " + m.filter.View() + "  (esc to clear)"
	}
	if v := strings.TrimSpace(m.filter.Value()); v != "" {
		return "Filter: " + m.filter.View() + "  (press f to edit, esc to clear)"
	}
	return "Press 'f' to filter queues"
}

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
