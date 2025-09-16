package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "time"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/obs"
    "github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
    itui "github.com/flyingrobots/go-redis-work-queue/internal/tui"
)

func main() {
    var configPath string
    var refresh time.Duration
    var redisURL string
    var cluster string
    var namespace string
    var readOnly bool
    var metricsAddr string
    var logLevel string
    var theme string
    var fps int
    var noMouse bool

    fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    fs.StringVar(&configPath, "config", "config/config.yaml", "Path to YAML config")
    fs.DurationVar(&refresh, "refresh", 2*time.Second, "Refresh interval for stats")
    fs.StringVar(&redisURL, "redis-url", "", "Quick connect Redis URL (redis://[:pass@]host:port/db)")
    fs.StringVar(&cluster, "cluster", "", "Named cluster from config")
    fs.StringVar(&namespace, "namespace", "", "Key namespace/prefix")
    fs.BoolVar(&readOnly, "read-only", false, "Force read-only mode (guardrails on)")
    fs.StringVar(&metricsAddr, "metrics-addr", ":9090", "Prometheus metrics address")
    fs.StringVar(&logLevel, "log-level", "info", "Log level: debug,info,warn,error")
    fs.StringVar(&theme, "theme", "auto", "Theme: auto,dark,light,high-contrast")
    fs.IntVar(&fps, "fps", 60, "FPS cap for rendering")
    fs.BoolVar(&noMouse, "no-mouse", false, "Disable mouse handling")
    _ = fs.Parse(os.Args[1:])

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

    // Allow overriding log level via flag
    if logLevel != "" {
        cfg.Observability.LogLevel = logLevel
    }
    logger, err := obs.NewLogger(cfg.Observability.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	rdb := redisclient.New(cfg)
	defer rdb.Close()
    if _, err := rdb.Ping(context.Background()).Result(); err != nil {
        fmt.Fprintf(os.Stderr, "redis ping failed: %v\n", err)
    }

    // TODO: wire redisURL/cluster/namespace/readOnly/metricsAddr/theme/fps into TUI once supported
    m := itui.New(cfg, rdb, logger, refresh)
    opts := []tea.ProgramOption{tea.WithAltScreen()}
    if !noMouse {
        opts = append(opts, tea.WithMouseAllMotion())
    }
    if _, err := tea.NewProgram(m, opts...).Run(); err != nil {
        fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
        os.Exit(1)
    }
}
