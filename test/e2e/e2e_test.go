// Copyright 2025 James Ross
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestE2E_WorkerCompletesJobWithRealRedis(t *testing.T) {
	addr := os.Getenv("E2E_REDIS_ADDR")
	if addr == "" {
		t.Skip("E2E_REDIS_ADDR not set; skipping e2e test")
	}
	cfg, _ := config.Load("nonexistent.yaml")
	cfg.Redis.Addr = addr
	cfg.Worker.Count = 1
	cfg.Worker.Backoff.Base = 1 * time.Millisecond
	cfg.Worker.Backoff.Max = 2 * time.Millisecond

	// Connect to real Redis and flush DB
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flushdb: %v", err)
	}

	// Enqueue a job in low priority queue
	j := queue.NewJob("e2e-id", "/tmp/e2e-ok.txt", 10, "low", "", "")
	payload, _ := j.Marshal()
	if err := rdb.LPush(context.Background(), cfg.Worker.Queues["low"], payload).Err(); err != nil {
		t.Fatalf("lpush: %v", err)
	}

	// Run worker
	log, _ := zap.NewDevelopment()
	wrk := worker.New(cfg, redisclient.New(cfg), log)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); _ = wrk.Run(ctx) }()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		n, _ := rdb.LLen(context.Background(), cfg.Worker.CompletedList).Result()
		if n == 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	<-done

	if n, _ := rdb.LLen(context.Background(), cfg.Worker.CompletedList).Result(); n != 1 {
		t.Fatalf("expected completed 1, got %d", n)
	}
}
