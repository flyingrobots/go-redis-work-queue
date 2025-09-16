// Copyright 2025 James Ross
// Package adminapi provides a secure HTTP API for admin operations on Redis work queues.
// It includes authentication, rate limiting, audit logging, and confirmation requirements
// for destructive operations.
package adminapi

import (
	"context"
	"fmt"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Run starts the admin API server
func Run(ctx context.Context, cfg *Config, appCfg *config.Config, rdb *redis.Client, logger *zap.Logger) error {
	server, err := NewServer(cfg, appCfg, rdb, logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down admin API server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}