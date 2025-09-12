// Copyright 2025 James Ross
package redisclient

import (
    "runtime"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/go-redis/redis/v8"
)

// New returns a configured go-redis v8 client with pooling and retries.
func New(cfg *config.Config) *redis.Client {
    poolSize := cfg.Redis.PoolSizeMultiplier * runtime.NumCPU()
    if poolSize <= 0 {
        poolSize = 10 * runtime.NumCPU()
    }
    return redis.NewClient(&redis.Options{
        Addr:         cfg.Redis.Addr,
        Username:     cfg.Redis.Username,
        Password:     cfg.Redis.Password,
        DB:           cfg.Redis.DB,
        PoolSize:     poolSize,
        MinIdleConns: cfg.Redis.MinIdleConns,
        DialTimeout:  cfg.Redis.DialTimeout,
        ReadTimeout:  cfg.Redis.ReadTimeout,
        WriteTimeout: cfg.Redis.WriteTimeout,
        MaxRetries:   cfg.Redis.MaxRetries,
        IdleTimeout:  5 * time.Minute,
    })
}
