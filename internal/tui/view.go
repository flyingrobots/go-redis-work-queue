package tui

import (
    "encoding/json"
    "fmt"
    "sort"
    "strings"
    "time"

    overlay "github.com/rmhubbert/bubbletea-overlay"
    "github.com/charmbracelet/lipgloss"
    asciigraph "github.com/guptarohit/asciigraph"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

func (m model) View() string {
    header := lipgloss.NewStyle().Bold(true).Render("Job Queue TUI — Redis " + m.cfg.Redis.Addr)
    sub := fmt.Sprintf("Focus: %s  |  Heartbeats: %d  |  Processing lists: %d", focusName(m.focus), m.lastStats.Heartbeats, len(m.lastStats.ProcessingLists))
    if m.errText != "" { sub += "  |  Error: " + m.errText }
    if m.loading { sub += "  " + m.spinner.View() }

    fb := renderFilterBar(m)
    left := m.tbl.View()
    if fb != "" { left = fb + "\n" + left }
    left = m.boxBody.Render(m.boxTitle.Render("Queues") + "\n" + left)

    m.vpCharts.SetContent(renderCharts(m))
    right := m.boxBody.Render(m.boxTitle.Render("Charts") + "\n" + m.vpCharts.View())

    info := summarizeKeys(m.lastKeys)
    if len(m.lastPeek.Items) > 0 { info += "\n\n" + renderPeek(m.lastPeek) }
    if m.benchCount.Focused() || m.benchRate.Focused() || m.benchPriority.Focused() || m.benchTimeout.Focused() || m.lastBench.Count > 0 {
        info += "\n\n" + renderBenchForm(m)
        if m.lastBench.Count > 0 { info += "\n" + renderBenchResult(m.lastBench) }
    }
    if m.pbActive && m.pbTotal > 0 { info += "\n\nBench Progress:\n" + m.pb.View() }
    m.vpInfo.SetContent(info)
    bottom := m.boxBody.Render(m.boxTitle.Render("Info") + "\n" + m.vpInfo.View())

    gap := lipgloss.NewStyle().Width(2).Render(" ")
    topRow := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
    body := topRow + "\n" + bottom

    base := header + "\n" + sub + "\n\n" + body
    if m.confirmOpen {
        baseDim := lipgloss.NewStyle().Faint(true).Render(base)
        back := staticStringModel{baseDim}
        fore := staticStringModel{renderConfirmModal(m)}
        ov := overlay.New(fore, back, overlay.Center, overlay.Center, 0, 0)
        return ov.View()
    }
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
    parts := []string{fmt.Sprintf("processing_lists=%d", k.ProcessingLists), fmt.Sprintf("processing_items=%d", k.ProcessingItems), fmt.Sprintf("heartbeats=%d", k.Heartbeats)}
    if k.RateLimitKey != "" { rl := "rate_limit_key=" + k.RateLimitKey; if k.RateLimitTTL != "" { rl += " ttl=" + k.RateLimitTTL }; parts = append(parts, rl) }
    return strings.Join(parts, "  |  ")
}

func renderKeys(k admin.KeysStats) string {
    keys := make([]string, 0, len(k.QueueLengths)); for name := range k.QueueLengths { keys = append(keys, name) }
    sort.Strings(keys)
    b := &strings.Builder{}
    fmt.Fprintf(b, "Queue Lengths:\n")
    for _, name := range keys { fmt.Fprintf(b, "  %-40s %8d\n", name, k.QueueLengths[name]) }
    fmt.Fprintf(b, "\nProcessing lists: %d\nProcessing items: %d\nHeartbeats: %d\n", k.ProcessingLists, k.ProcessingItems, k.Heartbeats)
    if k.RateLimitKey != "" { fmt.Fprintf(b, "Rate limit key: %s  TTL: %s\n", k.RateLimitKey, k.RateLimitTTL) }
    return b.String()
}

func renderPeek(p admin.PeekResult) string {
    b := &strings.Builder{}
    fmt.Fprintf(b, "Peek: %s\n", p.Queue)
    if len(p.Items) == 0 { fmt.Fprintf(b, "(no items)\n"); return b.String() }
    for i := len(p.Items) - 1; i >= 0; i-- {
        it := p.Items[i]
        var v map[string]any
        if json.Unmarshal([]byte(it), &v) == nil { pp, _ := json.MarshalIndent(v, "", "  "); fmt.Fprintf(b, "[%d]\n%s\n\n", i, string(pp)) } else { fmt.Fprintf(b, "[%d] %s\n", i, it) }
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
    if b.Count == 0 { return "" }
    return fmt.Sprintf("Bench: count=%d  duration=%s  thr=%.1f/s  p50=%s  p95=%s", b.Count, b.Duration.Truncate(time.Millisecond), b.Throughput, b.P50.Truncate(time.Millisecond), b.P95.Truncate(time.Millisecond))
}

func helpBar() string {
    return strings.Join([]string{"q:quit", "tab/shift+tab:focus panel", "r:refresh", "j/k:down/up", "wheel/mouse: scroll/select", "enter/p:peek", "b:bench form", "f:filter (queues)", "D:purge DLQ (y/n)", "A:purge ALL (y/n)"}, "  ")
}

func focusName(f focusArea) string {
    switch f { case focusQueues: return "Queues"; case focusCharts: return "Charts"; case focusInfo: return "Info" }
    return "?"
}

func renderCharts(m model) string {
    width := m.width; if width <= 10 { width = 80 }; plotW := width - 2; if plotW < 10 { plotW = 10 }
    h := 8
    makePlot := func(title string, data []float64) string {
        if len(data) == 0 { return fmt.Sprintf("%s\n(no data yet)", title) }
        g := asciigraph.Plot(data, asciigraph.Height(h), asciigraph.Width(plotW), asciigraph.Caption(title)); return g
    }
    parts := []string{}
    parts = append(parts, makePlot("High Priority", m.series["high"]))
    parts = append(parts, makePlot("Low Priority", m.series["low"]))
    parts = append(parts, makePlot("Completed", m.series["completed"]))
    parts = append(parts, makePlot("Dead Letter", m.series["dead_letter"]))
    return strings.Join(parts, "\n\n")
}

func renderFilterBar(m model) string {
    if m.filterActive { return "Filter: " + m.filter.View() + "  (esc to clear)" }
    if v := strings.TrimSpace(m.filter.Value()); v != "" { return "Filter: " + m.filter.View() + "  (press f to edit, esc to clear)" }
    return "Press 'f' to filter queues"
}
