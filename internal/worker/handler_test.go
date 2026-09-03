// Copyright 2026 James Ross
package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func newHandlerTestWorker(t *testing.T, count, maxRetries int, log *zap.Logger) (*Worker, *config.Config, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Redis.Addr = mr.Addr()
	cfg.Worker.Count = count
	cfg.Worker.MaxRetries = maxRetries
	cfg.Worker.Priorities = []string{"low"}
	cfg.Worker.Backoff.Base = 15 * time.Millisecond
	cfg.Worker.Backoff.Max = 15 * time.Millisecond
	cfg.Worker.BRPopLPushTimeout = time.Second
	if log == nil {
		log = zap.NewNop()
	}
	return New(cfg, rdb, log), cfg, rdb
}

func startHandlerTestWorker(w *Worker) (context.CancelFunc, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	return cancel, done
}

func stopHandlerTestWorker(t *testing.T, cancel context.CancelFunc, done <-chan error) {
	t.Helper()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("worker stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
}

func waitForListLength(t *testing.T, rdb *redis.Client, key string, want int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := rdb.LLen(context.Background(), key).Result()
		if err == nil && got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	got, err := rdb.LLen(context.Background(), key).Result()
	t.Fatalf("list %q did not reach length %d: got %d (err=%v)", key, want, got, err)
}

func waitForNoProcessingJobs(t *testing.T, rdb *redis.Client) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		keys, err := rdb.Keys(context.Background(), "jobqueue:worker:*:processing").Result()
		if err == nil {
			var total int64
			for _, key := range keys {
				count, countErr := rdb.LLen(context.Background(), key).Result()
				if countErr != nil {
					err = countErr
					break
				}
				total += count
			}
			if err == nil && total == 0 {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("processing lists did not drain")
}

func TestRegisteredHandlerProcessesEachJobOnce(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 2, 1, nil)
	received := make(chan queue.Job, 5)
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		received <- job
		return nil
	}))

	want := make(map[string][]byte, 5)
	for i := 0; i < 5; i++ {
		job := queue.NewJob(fmt.Sprintf("job-%d", i), "", 0, "low", "", "")
		job.Payload = []byte(fmt.Sprintf(`{"sequence":%d,"text":"世界"}`, i))
		job.PayloadSchema = "handler-test.v1"
		want[job.ID] = job.Payload
		if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
			t.Fatal(err)
		}
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 5)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	seen := make(map[string]int, len(want))
	for i := 0; i < len(want); i++ {
		select {
		case job := <-received:
			seen[job.ID]++
			if !bytes.Equal(job.Payload, want[job.ID]) {
				t.Fatalf("handler payload changed for %s: want %q, got %q", job.ID, want[job.ID], job.Payload)
			}
			if job.PayloadSchema != "handler-test.v1" {
				t.Fatalf("handler schema changed for %s: %q", job.ID, job.PayloadSchema)
			}
		case <-time.After(time.Second):
			t.Fatal("handler invocation missing")
		}
	}
	for id := range want {
		if seen[id] != 1 {
			t.Fatalf("expected job %s exactly once, got %d", id, seen[id])
		}
	}
	if queued, err := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result(); err != nil || queued != 0 {
		t.Fatalf("expected source queue drained, got %d (err=%v)", queued, err)
	}
}

func TestHandlerErrorRetriesThenDeadLetters(t *testing.T) {
	const maxRetries = 2
	w, cfg, rdb := newHandlerTestWorker(t, 1, maxRetries, nil)
	wantErr := errors.New("handler rejected job")
	var mu sync.Mutex
	var attempts []time.Time
	w.Handle(Handler(func(_ context.Context, _ queue.Job) error {
		mu.Lock()
		attempts = append(attempts, time.Now())
		mu.Unlock()
		return wantErr
	}))

	job := queue.NewJob("always-fails", "", 0, "low", "", "")
	job.Payload = []byte(`{"work":"reject"}`)
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.DeadLetterList, 1)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	mu.Lock()
	gotAttempts := append([]time.Time(nil), attempts...)
	mu.Unlock()
	if len(gotAttempts) != maxRetries+1 {
		t.Fatalf("expected initial attempt plus %d retries, got %d calls", maxRetries, len(gotAttempts))
	}
	for i := 1; i < len(gotAttempts); i++ {
		if delay := gotAttempts[i].Sub(gotAttempts[i-1]); delay < 10*time.Millisecond {
			t.Fatalf("retry %d skipped configured backoff: %v", i, delay)
		}
	}

	encoded, err := rdb.LIndex(context.Background(), cfg.Worker.DeadLetterList, 0).Result()
	if err != nil {
		t.Fatal(err)
	}
	dead, err := queue.UnmarshalJob(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if dead.ID != job.ID || dead.Retries != maxRetries {
		t.Fatalf("unexpected dead-letter job: %#v", dead)
	}
}

func TestHandlerPanicIsRecoveredAndWorkerContinues(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, zap.New(core))
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		if job.ID == "panic" {
			panic("boom")
		}
		return nil
	}))

	for _, id := range []string{"panic", "after-panic"} {
		job := queue.NewJob(id, "", 0, "low", "", "")
		if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
			t.Fatal(err)
		}
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.DeadLetterList, 1)
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 1)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	panicLogs := logs.FilterMessage("job handler panicked").All()
	if len(panicLogs) != 1 {
		t.Fatalf("expected one recovered-panic log, got %d", len(panicLogs))
	}
	stack, _ := panicLogs[0].ContextMap()["stack"].(string)
	if !strings.Contains(stack, "goroutine") {
		t.Fatalf("expected panic log to contain a stack, got %q", stack)
	}
}

func TestCancellationLeavesJobForReaper(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 1, nil)
	started := make(chan struct{})
	w.Handle(Handler(func(ctx context.Context, _ queue.Job) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}))

	job := queue.NewJob("cancel-me", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}

	cancel, done := startHandlerTestWorker(w)
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not start")
	}
	stopHandlerTestWorker(t, cancel, done)

	processing := fmt.Sprintf(cfg.Worker.ProcessingListPattern, w.baseID+"-0")
	if got, err := rdb.LLen(context.Background(), processing).Result(); err != nil || got != 1 {
		t.Fatalf("expected canceled job to remain in processing list, got %d (err=%v)", got, err)
	}
	for _, key := range []string{cfg.Worker.Queues["low"], cfg.Worker.CompletedList, cfg.Worker.DeadLetterList} {
		if got, err := rdb.LLen(context.Background(), key).Result(); err != nil || got != 0 {
			t.Fatalf("expected list %q empty after cancellation, got %d (err=%v)", key, got, err)
		}
	}
	heartbeat := fmt.Sprintf(cfg.Worker.HeartbeatKeyPattern, w.baseID+"-0")
	if exists, err := rdb.Exists(context.Background(), heartbeat).Result(); err != nil || exists != 1 {
		t.Fatalf("expected final heartbeat to expire naturally for reaper handoff, exists=%d (err=%v)", exists, err)
	}
}
