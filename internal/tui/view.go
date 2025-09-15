package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	flexbox "github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/lipgloss"
	asciigraph "github.com/guptarohit/asciigraph"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

func (m model) View() string {
	// Tab bar
	tabBar, _ := m.buildTabBar()

	header := lipgloss.NewStyle().Bold(true).Render("Job Queue TUI — Redis " + m.cfg.Redis.Addr)
	sub := fmt.Sprintf("Focus: %s  |  Heartbeats: %d  |  Processing lists: %d", focusName(m.focus), m.lastStats.Heartbeats, len(m.lastStats.ProcessingLists))
	if m.errText != "" {
		sub += "  |  Error: " + m.errText
	}
	if m.loading {
		sub += "  " + m.spinner.View()
	}

	// Panel border color per tab
	panelColor := "#7aa2f7" // default for Job Queue
	switch m.activeTab {
	case tabWorkers:
		panelColor = "#9ece6a"
	case tabDLQ:
		panelColor = "#f7768e"
	case tabEventHooks:
		panelColor = "#e0af68"
	case tabSettings:
		panelColor = "#bb9af7"
	}
	panel := m.boxBody.Copy().BorderForeground(lipgloss.Color(panelColor))

	var body string
	switch m.activeTab {
	case tabJobs:
		// Flex layout with borders at the cell level to avoid double borders/overflow.
		bodyW, bodyH := m.bodyDims()
		fbBox := flexbox.New(bodyW, bodyH)
		stack := bodyW < 120 // stack panels vertically on narrow terminals
		gutter := flexbox.NewCell(0, 2).SetMinWidth(2).SetContent("")
		// Animated ratios for wide mode: 0.0 => 1:1, 1.0 => 1:2 (Charts wider)
		base := 100
		lrx := base
		rrx := base + int(float64(base)*m.expPos)
		if rrx < 1 {
			rrx = 1
		}
		if lrx < 1 {
			lrx = 1
		}

		var rowTop, rowMid, rowBottom *flexbox.Row
		if !stack {
			// Wide: Queues | gutter | Charts on top, Info on bottom
			cLeft := flexbox.NewCell(lrx, 2).SetStyle(panel)
			cRight := flexbox.NewCell(rrx, 2).SetStyle(panel)
			cBottom := flexbox.NewCell(1, 1).SetStyle(panel)
			rowTop = fbBox.NewRow().AddCells(cLeft, gutter, cRight)
			rowBottom = fbBox.NewRow().AddCells(cBottom)
			fbBox.SetRows([]*flexbox.Row{rowTop, rowBottom})
		} else {
			// Narrow: stack Queues, Charts, then Info (2,2,1 ratios)
			cQueues := flexbox.NewCell(1, 2).SetStyle(panel)
			cCharts := flexbox.NewCell(1, 2).SetStyle(panel)
			cInfo := flexbox.NewCell(1, 1).SetStyle(panel)
			rowTop = fbBox.NewRow().AddCells(cQueues)
			rowMid = fbBox.NewRow().AddCells(cCharts)
			rowBottom = fbBox.NewRow().AddCells(cInfo)
			fbBox.SetRows([]*flexbox.Row{rowTop, rowMid, rowBottom})
		}

		// Size pass: compute inner dimensions for contents
		fbBox.ForceRecalculate()
		var lc, rc, bc *flexbox.Cell
		if !stack {
			lc = fbBox.GetRowCellCopy(0, 0)
			rc = fbBox.GetRowCellCopy(0, 2)
			bc = fbBox.GetRowCellCopy(1, 0)
		} else {
			lc = fbBox.GetRowCellCopy(0, 0)
			rc = fbBox.GetRowCellCopy(1, 0)
			bc = fbBox.GetRowCellCopy(2, 0)
		}
		// Panel overhead: 2 border + 2 padding (h), 2 border (v)
		innerLeftW, innerLeftH := lc.GetWidth()-4, lc.GetHeight()-2
		innerRightW, innerRightH := rc.GetWidth()-4, rc.GetHeight()-2
		innerBottomW, innerBottomH := bc.GetWidth()-4, bc.GetHeight()-2
		if innerLeftW < 1 {
			innerLeftW = 1
		}
		if innerRightW < 1 {
			innerRightW = 1
		}
		if innerBottomW < 1 {
			innerBottomW = 1
		}
		if innerLeftH < 1 {
			innerLeftH = 1
		}
		if innerRightH < 1 {
			innerRightH = 1
		}
		if innerBottomH < 1 {
			innerBottomH = 1
		}

		// Left (Queues)
		m.tbl.SetWidth(innerLeftW)
		// Height = inner cell height minus title and optional filter line
		filterLines := 0
		if strings.TrimSpace(m.filter.Value()) != "" || m.filterActive {
			filterLines = 1
		}
		tblH := innerLeftH - 1 - filterLines
		if tblH < 3 {
			tblH = 3
		}
		m.tbl.SetHeight(tblH)
		leftBody := m.tbl.View()
		if fb := renderFilterBar(m); fb != "" {
			leftBody = fb + "\n" + leftBody
		}
		leftContent := m.boxTitle.Render("Queues") + "\n" + leftBody

		// Right (Charts): render with cell-based width
		chartsStr := renderChartsWidth(m, innerRightW)
		rightContent := m.boxTitle.Render("Charts") + "\n" + chartsStr

		// Bottom (Info): viewport sized to inner dimensions
		m.vpInfo.Width = innerBottomW
		m.vpInfo.Height = innerBottomH - 1 // minus title line
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
		bottomContent := m.boxTitle.Render("Info") + "\n" + m.vpInfo.View()

		// Set contents after sizing so borders line up with cell widths
		rowTop.GetCell(0).SetContent(leftContent)
		if !stack {
			rowTop.GetCell(2).SetContent(rightContent)
		} else {
			rowMid.GetCell(0).SetContent(rightContent)
		}
		rowBottom.GetCell(0).SetContent(bottomContent)

		body = fbBox.Render()

	case tabWorkers:
		// Simple summary placeholder
		workersInfo := []string{
			fmt.Sprintf("Workers: heartbeats=%d", m.lastStats.Heartbeats),
			fmt.Sprintf("Processing lists: %d", len(m.lastStats.ProcessingLists)),
			"(Placeholder) Future: live workers view with heartbeats and active jobs",
		}
		content := strings.Join(workersInfo, "\n")
		bodyW, bodyH := m.bodyDims()
		fbBox := flexbox.New(bodyW, bodyH)
		single := fbBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(panel).SetContent(m.boxTitle.Render("Workers") + "\n" + content),
		)
		fbBox.SetRows([]*flexbox.Row{single})
		body = fbBox.Render()

	case tabDLQ:
		// DLQ summary placeholder
		dlqDisplay := fmt.Sprintf("dead_letter (%s)", m.cfg.Worker.DeadLetterList)
		dlqCount := m.lastStats.Queues[dlqDisplay]
		lines := []string{
			fmt.Sprintf("Dead Letter Queue: %s", m.cfg.Worker.DeadLetterList),
			fmt.Sprintf("Count: %d", dlqCount),
			"(Placeholder) Future: DLQ list with actions (peek/purge/requeue)",
		}
		bodyW, bodyH := m.bodyDims()
		fbBox := flexbox.New(bodyW, bodyH)
		single := fbBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(panel).SetContent(m.boxTitle.Render("Dead Letter Queue") + "\n" + strings.Join(lines, "\n")),
		)
		fbBox.SetRows([]*flexbox.Row{single})
		body = fbBox.Render()

	case tabEventHooks:
		// Event Hooks management view
		lines := []string{
			"Event Hooks - Real-time Job Event Notifications",
			"",
			"📡 Webhook Subscriptions: 0 active",
			"📊 Event Bus: Running | Events: 0 | Subscribers: 0",
			"🔄 Dead Letter Hooks: 0 failed deliveries",
			"",
			"Available Events:",
			"  • job_enqueued, job_started, job_succeeded",
			"  • job_failed, job_dlq, job_retried",
			"",
			"Management via Admin API:",
			"  POST /api/v1/event-hooks/webhooks - Create subscription",
			"  GET  /api/v1/event-hooks/health - View status",
		}
		bodyW, bodyH := m.bodyDims()
		fbBox := flexbox.New(bodyW, bodyH)
		single := fbBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(panel).SetContent(m.boxTitle.Render("Event Hooks") + "\n" + strings.Join(lines, "\n")),
		)
		fbBox.SetRows([]*flexbox.Row{single})
		body = fbBox.Render()

	case tabSettings:
		// Subset of key config values
		lines := []string{
			fmt.Sprintf("Redis: %s", m.cfg.Redis.Addr),
			fmt.Sprintf("Queues: high=%s low=%s", m.cfg.Worker.Queues["high"], m.cfg.Worker.Queues["low"]),
			fmt.Sprintf("Completed: %s", m.cfg.Worker.CompletedList),
			fmt.Sprintf("Dead Letter: %s", m.cfg.Worker.DeadLetterList),
			fmt.Sprintf("Default Priority: %s", m.cfg.Producer.DefaultPriority),
		}
		bodyW, bodyH := m.bodyDims()
		fbBox := flexbox.New(bodyW, bodyH)
		single := fbBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(panel).SetContent(m.boxTitle.Render("Settings") + "\n" + strings.Join(lines, "\n")),
		)
		fbBox.SetRows([]*flexbox.Row{single})
		body = fbBox.Render()
	}

	base := tabBar + "\n" + header + "\n" + sub + "\n\n" + body
	if m.confirmOpen {
		// Use a full-screen scrim overlay that centers the modal and preserves header/body
		return renderOverlayScreen(m)
	}
	now := time.Now().Format("15:04:05")
	m.sb.SetContent("Redis "+m.cfg.Redis.Addr, "focus:"+focusName(m.focus), m.spinner.View(), now)
	out := base + "\n" + m.sb.View()
	if m.help2.Active {
		// Dim with scrim and center the help content
		out = renderHelpOverlay(m, "")
	}
	return out
}

func summarizeKeys(k admin.KeysStats) string {
	parts := []string{fmt.Sprintf("processing_lists=%d", k.ProcessingLists), fmt.Sprintf("processing_items=%d", k.ProcessingItems), fmt.Sprintf("heartbeats=%d", k.Heartbeats)}
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
	fmt.Fprintf(b, "\nProcessing lists: %d\nProcessing items: %d\nHeartbeats: %d\n", k.ProcessingLists, k.ProcessingItems, k.Heartbeats)
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
	for i := len(p.Items) - 1; i >= 0; i-- {
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
	return fmt.Sprintf("Bench: count=%d  duration=%s  thr=%.1f/s  p50=%s  p95=%s", b.Count, b.Duration.Truncate(time.Millisecond), b.Throughput, b.P50.Truncate(time.Millisecond), b.P95.Truncate(time.Millisecond))
}

func helpBar() string {
	return strings.Join([]string{"q:quit", "tab/shift+tab:focus panel", "r:refresh", "j/k:down/up", "wheel/mouse: scroll/select", "enter/p:peek", "b:bench form", "f:filter (queues)", "D:purge DLQ (y/n)", "A:purge ALL (y/n)"}, "  ")
}

func focusName(f focusArea) string {
	switch f {
	case focusQueues:
		return "Queues"
	case focusCharts:
		return "Charts"
	case focusInfo:
		return "Info"
	}
	return "?"
}

func renderChartsWidth(m model, plotW int) string {
	if plotW < 10 {
		plotW = 10
	}
	h := 8
	makePlot := func(title string, data []float64) string {
		if len(data) == 0 {
			return fmt.Sprintf("%s\n(no data yet)", title)
		}
		g := asciigraph.Plot(data, asciigraph.Height(h), asciigraph.Width(plotW), asciigraph.Caption(title))
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
