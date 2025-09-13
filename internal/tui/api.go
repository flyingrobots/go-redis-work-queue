package tui

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
)

// New constructs the TUI model.
func New(cfg *config.Config, rdb *redis.Client, logger *zap.Logger, refreshEvery time.Duration) tea.Model {
    return initialModel(cfg, rdb, logger, refreshEvery)
}

