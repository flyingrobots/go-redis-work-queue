// Copyright 2025 James Ross
package reaper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestReaperRequeuesWithoutHeartbeat(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Redis.Addr = mr.Addr()
	log, _ := zap.NewDevelopment()
	rep := New(cfg, rdb, log)

	ctx := context.Background()
	workerID := "w1"
	plist := fmt.Sprintf(cfg.Worker.ProcessingListPattern, workerID)
	hbKey := fmt.Sprintf(cfg.Worker.HeartbeatKeyPattern, workerID)
	// Simulate dead worker: no heartbeat key
	job := queue.NewJob("id1", "/tmp/file.txt", 10, "low", "", "")
	payload, _ := job.Marshal()
	if err := rdb.LPush(ctx, plist, payload).Err(); err != nil {
		t.Fatal(err)
	}

	rep.scanOnce(ctx)

	// Expect job moved back to low priority queue
	n, _ := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result()
	if n != 1 {
		t.Fatalf("expected 1 job in low queue, got %d", n)
	}
	if mr.Exists(hbKey) {
		t.Fatalf("heartbeat should not exist")
	}
}

func TestReaperUsesConfiguredProcessingPattern(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Worker.ProcessingListPattern = "custom:{worker:%s}:active"
	cfg.Worker.HeartbeatKeyPattern = "custom:{heartbeat:%s}"
	cfg.Worker.Queues["low"] = "custom:low"

	job := queue.NewJob("custom-key-job", "", 0, "low", "", "")
	payload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(context.Background(), "custom:{worker:alpha:one}:active", payload).Err(); err != nil {
		t.Fatal(err)
	}

	New(cfg, rdb, zap.NewNop()).scanOnce(context.Background())

	if got, err := rdb.LLen(context.Background(), "custom:low").Result(); err != nil || got != 1 {
		t.Fatalf("custom queue length = %d, want 1 (err=%v)", got, err)
	}
	if mr.Exists("custom:{worker:alpha:one}:active") {
		t.Fatal("custom processing list was not drained")
	}
}

func TestRemoveMalformedTailDoesNotPopReplacement(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	processing := "jobqueue:test:processing"

	job := queue.NewJob("valid-job", "", 0, "low", "", "")
	validPayload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	invalidPayload := []byte("not-a-job-envelope")
	if err := rdb.RPush(ctx, processing, validPayload, invalidPayload).Err(); err != nil {
		t.Fatal(err)
	}

	inspected, err := rdb.LIndex(ctx, processing, -1).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rdb.RPop(ctx, processing).Result(); err != nil {
		t.Fatalf("competing reaper did not remove malformed tail: %v", err)
	}

	removed, err := removeTailIfEqual(ctx, rdb, processing, inspected)
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Fatal("removed replacement job after inspected malformed tail changed")
	}
	remaining, err := rdb.RPop(ctx, processing).Bytes()
	if err != nil {
		t.Fatalf("replacement job was lost: %v", err)
	}
	if string(remaining) != string(validPayload) {
		t.Fatalf("remaining payload = %q, want %q", remaining, validPayload)
	}
}

func TestReaperDoesNotStealJobFromLiveHandler(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Redis.Addr = mr.Addr()
	cfg.Worker.Count = 1
	cfg.Worker.Priorities = []string{"low"}
	cfg.Worker.HeartbeatTTL = 90 * time.Millisecond
	cfg.Worker.BRPopLPushTimeout = time.Second
	log := zap.NewNop()

	started := make(chan struct{})
	release := make(chan struct{})
	wrk := worker.New(cfg, rdb, log)
	wrk.Handle(worker.Handler(func(ctx context.Context, _ queue.Job) error {
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}))

	job := queue.NewJob("long-running", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- wrk.Run(ctx) }()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not start")
	}
	heartbeats, err := rdb.Keys(context.Background(), "jobqueue:processing:worker:*").Result()
	if err != nil || len(heartbeats) != 1 {
		t.Fatalf("expected one heartbeat, got %v (err=%v)", heartbeats, err)
	}

	// Move close to the original expiry, then wait for the ticker to renew it.
	mr.FastForward(80 * time.Millisecond)
	deadline := time.Now().Add(time.Second)
	for mr.TTL(heartbeats[0]) < 80*time.Millisecond && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if ttl := mr.TTL(heartbeats[0]); ttl < 80*time.Millisecond {
		t.Fatalf("heartbeat was not renewed: ttl=%v", ttl)
	}

	// Cumulatively exceed the initial TTL. A one-shot heartbeat would now be gone.
	mr.FastForward(80 * time.Millisecond)
	New(cfg, rdb, log).scanOnce(ctx)
	if queued, err := rdb.LLen(ctx, cfg.Worker.Queues["low"]).Result(); err != nil || queued != 0 {
		t.Fatalf("reaper stole live job: source length=%d (err=%v)", queued, err)
	}
	processing, err := rdb.Keys(ctx, "jobqueue:worker:*:processing").Result()
	if err != nil || len(processing) != 1 {
		t.Fatalf("expected one processing list, got %v (err=%v)", processing, err)
	}
	if got, err := rdb.LLen(ctx, processing[0]).Result(); err != nil || got != 1 {
		t.Fatalf("expected live job to remain in processing, got %d (err=%v)", got, err)
	}

	close(release)
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		completed, _ := rdb.LLen(context.Background(), cfg.Worker.CompletedList).Result()
		processing, _ := rdb.Keys(context.Background(), "jobqueue:worker:*:processing").Result()
		heartbeats, _ := rdb.Keys(context.Background(), "jobqueue:processing:worker:*").Result()
		if completed == 1 && len(processing) == 0 && len(heartbeats) == 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("long-running handler did not complete after release")
}
