// Copyright 2025 James Ross
package redisclient

import (
	"runtime"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/redis/go-redis/v9"
)

// Options translates application configuration into go-redis options.
func Options(cfg *config.Config) *redis.Options {
	poolSize := cfg.Redis.PoolSizeMultiplier * runtime.NumCPU()
	if poolSize <= 0 {
		poolSize = 10 * runtime.NumCPU()
	}
	return &redis.Options{
		Addr:            cfg.Redis.Addr,
		Username:        cfg.Redis.Username,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        poolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		MaxRetries:      cfg.Redis.MaxRetries,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// New returns a configured go-redis v9 client with pooling and retries.
func New(cfg *config.Config) *redis.Client {
	return redis.NewClient(Options(cfg))
}
