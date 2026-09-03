//go:build e2e_tests
// +build e2e_tests

// Gated because end-to-end scenarios require explicit services; un-gate individual files when default CI provisions their dependencies.

// Copyright 2025 James Ross
package e2e

import (
	"bytes"
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
	cfg.Worker.Priorities = []string{"low"}
	cfg.Worker.Backoff.Base = 1 * time.Millisecond
	cfg.Worker.Backoff.Max = 2 * time.Millisecond
	cfg.Worker.BRPopLPushTimeout = time.Second
	cfg.Worker.HeartbeatTTL = 150 * time.Millisecond

	// Connect to real Redis and flush DB
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flushdb: %v", err)
	}

	// Enqueue a job in low priority queue
	j := queue.NewJob("e2e-id", "/tmp/e2e-ok.txt", 10, "low", "", "")
	j.Payload = []byte("real Redis payload: 世界\x00\xff")
	j.PayloadSchema = "e2e.v1"
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], j, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// Run worker
	log, _ := zap.NewDevelopment()
	wrk := worker.New(cfg, redisclient.New(cfg), log)
	started := make(chan queue.Job, 1)
	release := make(chan struct{})
	wrk.Handle(worker.Handler(func(ctx context.Context, job queue.Job) error {
		started <- job
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	done := make(chan struct{})
	go func() { defer close(done); _ = wrk.Run(ctx) }()

	var handled queue.Job
	select {
	case handled = <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not start")
	}
	if handled.ID != j.ID || !bytes.Equal(handled.Payload, j.Payload) || handled.PayloadSchema != j.PayloadSchema {
		t.Fatalf("handler received changed job: %#v", handled)
	}

	// The handler outlives the original heartbeat, so the key must be renewed.
	time.Sleep(3 * cfg.Worker.HeartbeatTTL)
	heartbeats, err := rdb.Keys(context.Background(), "jobqueue:processing:worker:*").Result()
	if err != nil || len(heartbeats) != 1 {
		t.Fatalf("expected renewed heartbeat for live handler, got %v (err=%v)", heartbeats, err)
	}
	close(release)
	released = true

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		completed, _ := rdb.LLen(context.Background(), cfg.Worker.CompletedList).Result()
		processing, _ := rdb.Keys(context.Background(), "jobqueue:worker:*:processing").Result()
		heartbeats, _ := rdb.Keys(context.Background(), "jobqueue:processing:worker:*").Result()
		if completed == 1 && len(processing) == 0 && len(heartbeats) == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	<-done

	if n, _ := rdb.LLen(context.Background(), cfg.Worker.CompletedList).Result(); n != 1 {
		t.Fatalf("expected completed 1, got %d", n)
	}
	if processing, _ := rdb.Keys(context.Background(), "jobqueue:worker:*:processing").Result(); len(processing) != 0 {
		t.Fatalf("expected processing cleanup, got %v", processing)
	}
	if heartbeats, _ := rdb.Keys(context.Background(), "jobqueue:processing:worker:*").Result(); len(heartbeats) != 0 {
		t.Fatalf("expected heartbeat cleanup, got %v", heartbeats)
	}
}
