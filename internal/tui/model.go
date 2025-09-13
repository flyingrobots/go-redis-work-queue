package tui

import (
    "context"
    "time"

    "github.com/charmbracelet/bubbles/help"
    bubprog "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/lipgloss"
    tchelp "github.com/mistakenelf/teacup/help"
    "github.com/mistakenelf/teacup/statusbar"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
)

// focusable panels on the dashboard
type focusArea int

const (
    focusQueues focusArea = iota
    focusCharts
    focusInfo
)

// messages
type (
    statsMsg struct{ s admin.StatsResult; err error }
    keysMsg  struct{ k admin.KeysStats; err error }
    peekMsg  struct{ p admin.PeekResult; err error }
    benchMsg struct{ b admin.BenchResult; err error }
    enqueueMsg struct{ n int; key string; err error }
    tick       struct{}
    benchPollTick struct{}
    benchProgMsg  struct{ done int64 }
)

type model struct {
    ctx    context.Context
    cancel context.CancelFunc

    cfg    *config.Config
    rdb    *redis.Client
    logger *zap.Logger

    width  int
    height int

    focus   focusArea
    help    help.Model
    spinner spinner.Model
    loading bool
    errText string

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
    tableTopY int

    // time series for charts
    series    map[string][]float64
    seriesMax int

    // confirmation modal state
    confirmOpen   bool
    confirmAction string

    // Filter state for queues view
    filter       textinput.Model
    filterActive bool
    allRows      []table.Row
    allTargets   []string

    // Dashboard viewports
    vpCharts viewport.Model
    vpInfo   viewport.Model

    // Styles
    boxTitle lipgloss.Style
    boxBody  lipgloss.Style

    // teacup components
    sb    statusbar.Model
    help2 tchelp.Model

    // Progress for bench
    pb       bubprog.Model
    pbActive bool
    pbTotal  int
}

