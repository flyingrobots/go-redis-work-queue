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
)

func main() {
	var configPath string
	var refresh time.Duration
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&configPath, "config", "config/config.yaml", "Path to YAML config")
	fs.DurationVar(&refresh, "refresh", 2*time.Second, "Refresh interval for stats")
	_ = fs.Parse(os.Args[1:])

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
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

	m := initialModel(cfg, rdb, logger, refresh)
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
