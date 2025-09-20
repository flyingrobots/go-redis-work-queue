//go:build worker_tests
// +build worker_tests

// Copyright 2025 James Ross
package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func setupWorkerTest(t *testing.T) (*Worker, *config.Config, *redis.Client, func()) {
	t.Helper()
	mr, _ := miniredis.Run()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg, _ := config.Load("nonexistent.yaml")
	cfg.Redis.Addr = mr.Addr()
	cfg.Worker.Backoff.Base = 1 * time.Millisecond
	cfg.Worker.Backoff.Max = 2 * time.Millisecond
	cfg.Worker.MaxRetries = 1
	log, _ := zap.NewDevelopment()
	w := New(cfg, rdb, log)
	cleanup := func() { mr.Close() }
	return w, cfg, rdb, cleanup
}

func TestProcessJobSuccess(t *testing.T) {
	w, cfg, rdb, cleanup := setupWorkerTest(t)
	defer cleanup()
	workerID := "w1"
	procList := fmt.Sprintf(cfg.Worker.ProcessingListPattern, workerID)
	hbKey := fmt.Sprintf(cfg.Worker.HeartbeatKeyPattern, workerID)
	job := queue.NewJob("id1", "/tmp/ok.txt", 10, "low", "", "")
	payload, _ := job.Marshal()
	ctx := context.Background()
	ok := w.processJob(ctx, workerID, cfg.Worker.Queues["low"], procList, hbKey, payload)
	if !ok {
		t.Fatalf("expected success")
	}
	if n, _ := rdb.LLen(ctx, cfg.Worker.CompletedList).Result(); n != 1 {
		t.Fatalf("expected completed 1, got %d", n)
	}
}

func TestProcessJobRetryThenDLQ(t *testing.T) {
	w, cfg, rdb, cleanup := setupWorkerTest(t)
	defer cleanup()
	workerID := "w1"
	procList := fmt.Sprintf(cfg.Worker.ProcessingListPattern, workerID)
	hbKey := fmt.Sprintf(cfg.Worker.HeartbeatKeyPattern, workerID)
	// filename contains "fail" to trigger failure
	job := queue.NewJob("id1", "/tmp/fail.txt", 10, "low", "", "")
	payload, _ := job.Marshal()
	ctx := context.Background()
	ok := w.processJob(ctx, workerID, cfg.Worker.Queues["low"], procList, hbKey, payload)
	if ok {
		t.Fatalf("expected failure")
	}
	// After one retry allowed, it should have been requeued to low
	if n, _ := rdb.LLen(ctx, cfg.Worker.Queues["low"]).Result(); n != 1 {
		t.Fatalf("expected requeued 1, got %d", n)
	}
	// Process again to exceed retries -> DLQ
	payload2, _ := rdb.LPop(ctx, cfg.Worker.Queues["low"]).Result()
	_ = rdb.LPush(ctx, procList, payload2).Err()
	ok2 := w.processJob(ctx, workerID, cfg.Worker.Queues["low"], procList, hbKey, payload2)
	if ok2 {
		t.Fatalf("expected failure to DLQ")
	}
	if n, _ := rdb.LLen(ctx, cfg.Worker.DeadLetterList).Result(); n != 1 {
		t.Fatalf("expected DLQ 1, got %d", n)
	}
}
