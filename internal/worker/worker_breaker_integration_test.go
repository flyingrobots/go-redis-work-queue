//go:build worker_tests
// +build worker_tests

// Copyright 2025 James Ross
package worker

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Test that repeated failures trip the breaker, and while Open the worker
// does not drain the queue until cooldown elapses.
func TestWorkerBreakerTripsAndPausesConsumption(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	cfg, _ := config.Load("nonexistent.yaml")
	cfg.Redis.Addr = mr.Addr()
	cfg.Worker.Count = 1
	// Make retries immediate and short
	cfg.Worker.Backoff.Base = 1 * time.Millisecond
	cfg.Worker.Backoff.Max = 2 * time.Millisecond
	cfg.Worker.BRPopLPushTimeout = 5 * time.Millisecond
	// Breaker tuned for quick transition
	cfg.CircuitBreaker.Window = 20 * time.Millisecond
	cfg.CircuitBreaker.CooldownPeriod = 100 * time.Millisecond
	cfg.CircuitBreaker.FailureThreshold = 0.5
	cfg.CircuitBreaker.MinSamples = 1
	cfg.Worker.BreakerPause = 5 * time.Millisecond

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Enqueue failing jobs (filename contains "fail")
	for i := 0; i < 5; i++ {
		j := queue.NewJob(
			"id-fail-",
			"/tmp/fail-test.txt", // contains "fail" to force failure
			1,
			"low",
			"",
			"",
		)
		payload, _ := j.Marshal()
		if err := rdb.LPush(context.Background(), cfg.Worker.Queues["low"], payload).Err(); err != nil {
			t.Fatalf("lpush: %v", err)
		}
	}

	log, _ := zap.NewDevelopment()
	w := New(cfg, rdb, log)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); _ = w.Run(ctx) }()

	// Wait up to 2s for breaker to open
	deadline := time.Now().Add(2 * time.Second)
	opened := false
	for time.Now().Before(deadline) {
		if w.cb.State() == 2 { // Open
			opened = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !opened {
		cancel()
		<-done
		t.Fatalf("breaker did not open under failures")
	}

	// While breaker is Open (cooldown 100ms), queue length should not decrease
	n1, _ := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result()
	time.Sleep(50 * time.Millisecond) // less than cooldown
	n2, _ := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result()
	if n2 < n1 {
		cancel()
		<-done
		t.Fatalf("queue drained during breaker open: before=%d after=%d", n1, n2)
	}

	cancel()
	<-done
}
