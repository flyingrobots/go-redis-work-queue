// Copyright 2025 James Ross
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "sort"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    asciigraph "github.com/guptarohit/asciigraph"
    "github.com/lithammer/fuzzysearch/fuzzy"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/obs"
    "github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

// Simple, pragmatic TUI for observing and administering the queue system.

type viewMode int

const (
    modeQueues viewMode = iota
    modeKeys
    modePeek
    modeBench
    modeCharts
)

type statsMsg struct {
    s   admin.StatsResult
    err error
}

type keysMsg struct {
    k   admin.KeysStats
    err error
}

type peekMsg struct {
    p   admin.PeekResult
    err error
}

type benchMsg struct {
    b   admin.BenchResult
    err error
}

type tick struct{}

type model struct {
    ctx    context.Context
    cancel context.CancelFunc

    cfg    *config.Config
    rdb    *redis.Client
    logger *zap.Logger

    width  int
    height int

    mode      viewMode
    help      help.Model
    spinner   spinner.Model
    loading   bool
    errText   string

    // Queues table (name, count)
    tbl table.Model
    // For mapping selection to queue alias or key
    peekTargets []string

    // Cached data
    lastStats admin.StatsResult
    lastKeys  admin.KeysStats
    lastPeek  admin.PeekResult
    lastBench admin.BenchResult

    // Bench prompt inputs
    benchCount    textinput.Model
    benchRate     textinput.Model
    benchPriority textinput.Model
    benchTimeout  textinput.Model

    refreshEvery time.Duration

    // layout helpers
    tableTopY int // number of lines before the table starts

    // time series for charts
    series    map[string][]float64
    seriesMax int

    // confirmation modal state
    confirmOpen   bool
    confirmAction string // "purge-dlq" | "purge-all"

    // Filter state for queues view
    filter       textinput.Model
    filterActive bool
    allRows      []table.Row
    allTargets   []string
}

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
        Header: lipgloss.NewStyle().Bold(true),
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

    return model{
        ctx:          ctx,
        cancel:       cancel,
        cfg:          cfg,
        rdb:          rdb,
        logger:       logger,
        mode:         modeQueues,
        help:         help.New(),
        spinner:      sp,
        tbl:          t,
        benchCount:   bi,
        benchRate:    br,
        benchPriority: bp,
        benchTimeout: bt,
        refreshEvery: refreshEvery,
        tableTopY:    3, // header + sub + blank line
        series:       map[string][]float64{"high": {}, "low": {}, "completed": {}, "dead_letter": {}},
        seriesMax:    180, // keep last N points
        filter:       fi,
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // When a confirmation modal is open, only handle confirm/cancel keys
        if m.confirmOpen {
            switch msg.String() {
            case "y", "enter":
                // Run the pending action
                switch m.confirmAction {
                case "purge-dlq":
                    m.loading = true
                    m.errText = ""
                    m.confirmOpen = false
                    cmds = append(cmds, func() tea.Msg {
                        err := admin.PurgeDLQ(m.ctx, m.cfg, m.rdb)
                        if err != nil { return statsMsg{err: err} }
                        return statsMsg{}
                    }, spinner.Tick, m.refreshCmd(), m.fetchKeysCmd())
                case "purge-all":
                    m.loading = true
                    m.errText = ""
                    m.confirmOpen = false
                    cmds = append(cmds, func() tea.Msg {
                        _, err := admin.PurgeAll(m.ctx, m.cfg, m.rdb)
                        if err != nil { return statsMsg{err: err} }
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
            m.cancel()
            return m, tea.Quit
        case "tab":
            // Toggle between queues and keys views
            if m.mode == modeQueues { m.mode = modeKeys } else { m.mode = modeQueues }
            return m, tea.Batch(m.refreshCmd(), m.fetchKeysCmd())
        case "r":
            return m, tea.Batch(m.refreshCmd(), m.fetchKeysCmd())
        case "f", "/":
            if m.mode == modeQueues {
                m.filterActive = true
                m.filter.Focus()
            }
        case "c":
            m.mode = modeCharts
        case "p":
            if m.mode == modeQueues && len(m.peekTargets) > 0 {
                i := m.tbl.Cursor()
                if i >= 0 && i < len(m.peekTargets) {
                    m.loading = true
                    m.errText = ""
                    m.mode = modePeek
                    cmds = append(cmds, m.doPeekCmd(m.peekTargets[i], 10), spinner.Tick)
                }
            }
        case "b":
            m.mode = modeBench
            // Focus first input
            m.benchCount.Focus()
        case "enter":
            if m.mode == modeBench {
                // Parse inputs and run
                count := atoiDefault(m.benchCount.Value(), 1000)
                rate := atoiDefault(m.benchRate.Value(), 500)
                prio := strings.TrimSpace(m.benchPriority.Value())
                if prio == "" { prio = m.cfg.Producer.DefaultPriority }
                to := time.Duration(atoiDefault(m.benchTimeout.Value(), 60)) * time.Second
                m.loading = true
                m.errText = ""
                cmds = append(cmds, m.doBenchCmd(prio, count, rate, to), spinner.Tick)
            }
        case "esc":
            // Return to queues view or clear filter/modal
            if m.confirmOpen {
                m.confirmOpen = false
            } else if m.mode == modeBench {
                m.mode = modeQueues
            } else if m.mode == modeQueues && m.filterActive {
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
        if m.mode == modeBench {
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
        if m.mode == modeQueues && m.filterActive {
            var c tea.Cmd
            m.filter, c = m.filter.Update(msg)
            cmds = append(cmds, c)
            m.applyFilterAndSetRows()
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Fit table to window
        if m.width > 0 { m.tbl.SetWidth(m.width) }
        // Leave a bit of space for footer/help
        if m.height > 6 { m.tbl.SetHeight(m.height - 6) }
    case tea.MouseMsg:
        // Basic mouse support for queues view: wheel scroll and click to select/peek
        if m.mode == modeQueues && !m.confirmOpen {
            top := m.tableTopY
            if m.filterActive || strings.TrimSpace(m.filter.Value()) != "" { top++ }
            switch msg.Button {
            case tea.MouseButtonWheelUp:
                if msg.Action == tea.MouseActionPress { m.tbl.MoveUp(1) }
            case tea.MouseButtonWheelDown:
                if msg.Action == tea.MouseActionPress { m.tbl.MoveDown(1) }
            case tea.MouseButtonNone:
                if msg.Action == tea.MouseActionMotion {
                    // Hover highlight: move cursor under pointer
                    rowWithin := msg.Y - (top + 1)
                    if rowWithin >= 0 && rowWithin < m.tbl.Height() {
                        start := clamp(m.tbl.Cursor()-m.tbl.Height(), 0, m.tbl.Cursor())
                        idx := start + rowWithin
                        if idx >= 0 && idx < len(m.tbl.Rows()) {
                            m.tbl.SetCursor(idx)
                        }
                    }
                }
            case tea.MouseButtonLeft:
                if msg.Action == tea.MouseActionPress {
                    // Attempt to map Y position to a visible row
                    // table header is 1 line; rows follow
                    rowWithin := msg.Y - (top + 1)
                    if rowWithin >= 0 && rowWithin < m.tbl.Height() {
                        // Compute starting row index using table logic
                        start := clamp(m.tbl.Cursor()-m.tbl.Height(), 0, m.tbl.Cursor())
                        idx := start + rowWithin
                        if idx >= 0 && idx < len(m.tbl.Rows()) {
                            m.tbl.SetCursor(idx)
                        }
                    }
                }
            case tea.MouseButtonRight:
                if msg.Action == tea.MouseActionPress {
                    // Right-click: peek selected
                    if len(m.peekTargets) > 0 {
                        i := m.tbl.Cursor()
                        if i >= 0 && i < len(m.peekTargets) {
                            m.loading = true
                            m.errText = ""
                            m.mode = modePeek
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
            if m.tbl.Cursor() >= len(rows) && len(rows) > 0 { m.tbl.SetCursor(len(rows) - 1) }
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
        }
    }

    // Spinner update when loading
    if m.loading {
        var c tea.Cmd
        m.spinner, c = m.spinner.Update(msg)
        cmds = append(cmds, c)
    }
    // Table update when in queues view
    if m.mode == modeQueues {
        var c tea.Cmd
        m.tbl, c = m.tbl.Update(msg)
        cmds = append(cmds, c)
    }

    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    // Header
    header := lipgloss.NewStyle().Bold(true).Render("Job Queue TUI — Redis " + m.cfg.Redis.Addr)
    sub := fmt.Sprintf("Mode: %s  |  Heartbeats: %d  |  Processing lists: %d",
        modeName(m.mode), m.lastStats.Heartbeats, len(m.lastStats.ProcessingLists))
    if m.errText != "" {
        sub += "  |  Error: " + m.errText
    }
    if m.loading {
        sub += "  " + m.spinner.View()
    }

    body := ""
    switch m.mode {
    case modeQueues:
        fb := renderFilterBar(m)
        if fb != "" { m.tableTopY = 4 } else { m.tableTopY = 3 }
        if fb != "" { body = fb + "\n" + m.tbl.View() } else { body = m.tbl.View() }
        // Footer summary from keys
        body += "\n" + summarizeKeys(m.lastKeys)
        body += "\n" + helpBar()
    case modeKeys:
        body = renderKeys(m.lastKeys)
        body += "\n" + helpBar()
    case modePeek:
        body = renderPeek(m.lastPeek)
        body += "\n" + helpBar()
    case modeBench:
        body = renderBenchForm(m)
        if (m.lastBench.Count > 0 && !m.loading) || m.errText != "" {
            body += "\n" + renderBenchResult(m.lastBench)
        }
        body += "\n" + helpBar()
    case modeCharts:
        body = renderCharts(m)
        body += "\n" + helpBar()
    }

    // Compose base view
    base := header + "\n" + sub + "\n\n" + body

    // Overlay confirmation modal if needed, with dimmed background
    if m.confirmOpen {
        dim := lipgloss.NewStyle().Faint(true).Render(base)
        return dim + "\n" + renderConfirmModal(m)
    }
    return base
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
        if k.RateLimitTTL != "" { rl += " ttl=" + k.RateLimitTTL }
        parts = append(parts, rl)
    }
    return strings.Join(parts, "  |  ")
}

func renderKeys(k admin.KeysStats) string {
    // Deterministic order
    keys := make([]string, 0, len(k.QueueLengths))
    for name := range k.QueueLengths { keys = append(keys, name) }
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
    if b.Count == 0 { return "" }
    return fmt.Sprintf("Bench: count=%d  duration=%s  thr=%.1f/s  p50=%s  p95=%s",
        b.Count, b.Duration.Truncate(time.Millisecond), b.Throughput, b.P50.Truncate(time.Millisecond), b.P95.Truncate(time.Millisecond))
}

func helpBar() string {
    return strings.Join([]string{
        "q:quit",
        "tab:switch view",
        "r:refresh",
        "j/k:down/up",
        "wheel/mouse: scroll/select",
        "right-click: peek",
        "p:peek",
        "b:bench",
        "c:charts",
        "f:filter",
        "D:purge DLQ (y/n)",
        "A:purge ALL (y/n)",
    }, "  ")
}

func modeName(m viewMode) string {
    switch m {
    case modeQueues:
        return "Queues"
    case modeKeys:
        return "Keys"
    case modePeek:
        return "Peek"
    case modeBench:
        return "Bench"
    default:
        return "?"
    }
}

func cycleBenchFocus(m *model) {
    if m.benchCount.Focused() {
        m.benchCount.Blur(); m.benchRate.Focus(); return
    }
    if m.benchRate.Focused() {
        m.benchRate.Blur(); m.benchPriority.Focus(); return
    }
    if m.benchPriority.Focused() {
        m.benchPriority.Blur(); m.benchTimeout.Focus(); return
    }
    m.benchTimeout.Blur(); m.benchCount.Focus()
}

func atoiDefault(s string, def int) int {
    var v int
    _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
    if err != nil { return def }
    return v
}

func main() {
    var configPath string
    var refresh time.Duration
    fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    fs.StringVar(&configPath, "config", "config/config.yaml", "Path to YAML config")
    fs.DurationVar(&refresh, "refresh", 2*time.Second, "Refresh interval for stats")
    _ = fs.Parse(os.Args[1:])

    // Load configuration
    cfg, err := config.Load(configPath)
    if err != nil { fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err); os.Exit(1) }

    // Logger
    logger, err := obs.NewLogger(cfg.Observability.LogLevel)
    if err != nil { fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err); os.Exit(1) }
    defer logger.Sync()

    // Redis
    rdb := redisclient.New(cfg)
    defer rdb.Close()
    if _, err := rdb.Ping(context.Background()).Result(); err != nil {
        fmt.Fprintf(os.Stderr, "redis ping failed: %v\n", err)
    }

    m := initialModel(cfg, rdb, logger, refresh)
    if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
        fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
        os.Exit(1)
    }
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
    if width <= 0 { width = 80 }
    modal := box.Render(content)
    // Center horizontally by padding spaces
    pad := 0
    if w := lipgloss.Width(modal); width > w { pad = (width - w) / 2 }
    return strings.Repeat(" ", pad) + modal
}

// addSample appends a value to a named series using StatsResult map.
func (m *model) addSample(alias, key string, s admin.StatsResult) {
    if alias == "" || key == "" { return }
    display := fmt.Sprintf("%s (%s)", alias, key)
    val := s.Queues[display]
    arr := m.series[alias]
    arr = append(arr, float64(val))
    if len(arr) > m.seriesMax { arr = arr[len(arr)-m.seriesMax:] }
    m.series[alias] = arr
}

func renderCharts(m model) string {
    width := m.width
    if width <= 10 { width = 80 }
    plotW := width - 2
    if plotW < 10 { plotW = 10 }
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
    if m.mode != modeQueues { return "" }
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
    for i, r := range m.allRows { labels[i] = r[0] }
    ranks := fuzzy.RankFindNormalizedFold(q, labels)
    fuzzy.SortRanks(ranks)
    rows := make([]table.Row, 0, len(ranks))
    targets := make([]string, 0, len(ranks))
    for _, rk := range ranks {
        rows = append(rows, m.allRows[rk.OriginalIndex])
        targets = append(targets, m.allTargets[rk.OriginalIndex])
    }
    m.tbl.SetRows(rows)
    m.peekTargets = targets
}
