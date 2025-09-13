package tui

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

func (m model) refreshCmd() tea.Cmd {
    return func() tea.Msg {
        s, err := admin.Stats(m.ctx, m.cfg, m.rdb)
        if err != nil { return statsMsg{err: err} }
        return statsMsg{s: s}
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

func (m model) benchPollCmd() tea.Cmd {
    return func() tea.Msg {
        n, _ := m.rdb.LLen(m.ctx, m.cfg.Worker.CompletedList).Result()
        return benchProgMsg{done: n}
    }
}
