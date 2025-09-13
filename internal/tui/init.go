package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-redis/redis/v8"
	tchelp "github.com/mistakenelf/teacup/help"
	"github.com/mistakenelf/teacup/statusbar"
	"go.uber.org/zap"

	bubprog "github.com/charmbracelet/bubbles/progress"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
)

func initialModel(cfg *config.Config, rdb *redis.Client, logger *zap.Logger, refreshEvery time.Duration) model {
	ctx, cancel := context.WithCancel(context.Background())

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	columns := []table.Column{{Title: "Queue", Width: 40}, {Title: "Count", Width: 10}}
	t := table.New(table.WithColumns(columns), table.WithFocused(true))
	t.KeyMap.LineUp.SetKeys("k", "up")
	t.KeyMap.LineDown.SetKeys("j", "down")
	t.KeyMap.PageDown.SetKeys("ctrl+f")
	t.KeyMap.PageUp.SetKeys("ctrl+b")
	t.SetStyles(table.Styles{Header: lipgloss.NewStyle().Bold(true), Selected: lipgloss.NewStyle().Bold(true)})

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

	fi := textinput.New()
	fi.Placeholder = "filter"
	fi.CharLimit = 64

	boxTitle := lipgloss.NewStyle().Bold(true)
	boxBody := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	sb := statusbar.New(
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#000000", Light: "#ffffff"}, Background: lipgloss.AdaptiveColor{Dark: "#ffaa00", Light: "#111111"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}, Background: lipgloss.AdaptiveColor{Dark: "#333333", Light: "#eeeeee"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}, Background: lipgloss.AdaptiveColor{Dark: "#444444", Light: "#dddddd"}},
		statusbar.ColorConfig{Foreground: lipgloss.AdaptiveColor{Dark: "#000000", Light: "#ffffff"}, Background: lipgloss.AdaptiveColor{Dark: "#55dd55", Light: "#226622"}},
	)

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
		tableTopY:     3,
		series:        map[string][]float64{"high": {}, "low": {}, "completed": {}, "dead_letter": {}},
		seriesMax:     180,
		filter:        fi,
		vpCharts:      viewport.New(0, 10),
		vpInfo:        viewport.New(0, 10),
		boxTitle:      boxTitle,
		boxBody:       boxBody,
		sb:            sb,
		help2:         help2,
        pb:            bubprog.New(bubprog.WithDefaultGradient()),
        activeTab:     tabJobs,
    }
}
