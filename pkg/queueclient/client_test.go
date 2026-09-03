// Copyright 2026 James Ross
package queueclient_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func testClientConfig() queueclient.Config {
	return queueclient.Config{
		Queues: map[string]string{
			"high": "jobqueue:test:high",
			"low":  "jobqueue:test:low",
		},
		DefaultPriority:       "low",
		ProcessingListPattern: "jobqueue:test:worker:%s:processing",
		HeartbeatKeyPattern:   "jobqueue:test:heartbeat:%s",
		CompletedList:         "jobqueue:test:completed",
		DeadLetterList:        "jobqueue:test:dead-letter",
		MaxPayloadSize:        32,
	}
}

func newTestClient(t *testing.T, mr *miniredis.Miniredis, cfg queueclient.Config) *queueclient.Client {
	t.Helper()
	client, err := queueclient.New(&redis.Options{Addr: mr.Addr()}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close queue client: %v", err)
		}
	})
	return client
}

func TestClientEnqueueFeedsWorkerHandler(t *testing.T) {
	mr := miniredis.RunT(t)
	clientCfg := testClientConfig()
	client := newTestClient(t, mr, clientCfg)

	workerRedis := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = workerRedis.Close() })
	workerCfg, err := config.Load("testdata/does-not-exist.yaml")
	if err != nil {
		t.Fatal(err)
	}
	workerCfg.Worker.Count = 1
	workerCfg.Worker.Priorities = []string{"high", "low"}
	workerCfg.Worker.Queues = clientCfg.Queues
	workerCfg.Worker.ProcessingListPattern = clientCfg.ProcessingListPattern
	workerCfg.Worker.HeartbeatKeyPattern = clientCfg.HeartbeatKeyPattern
	workerCfg.Worker.CompletedList = clientCfg.CompletedList
	workerCfg.Worker.DeadLetterList = clientCfg.DeadLetterList
	workerCfg.Worker.BRPopLPushTimeout = time.Second
	workerCfg.Queue.MaxPayloadSize = clientCfg.MaxPayloadSize

	received := make(chan queueclient.Job, 1)
	wrk := worker.New(workerCfg, workerRedis, zap.NewNop()).Handle(func(_ context.Context, job queueclient.Job) error {
		received <- job
		return nil
	})
	workerCtx, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan error, 1)
	go func() { workerDone <- wrk.Run(workerCtx) }()
	t.Cleanup(func() {
		stopWorker()
		select {
		case err := <-workerDone:
			if err != nil {
				t.Errorf("worker stopped: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("worker did not stop")
		}
	})

	wantPayload := []byte(`{"x":1,"message":"hello"}`)
	id, err := client.Enqueue(context.Background(), queueclient.Job{
		Payload:       wantPayload,
		PayloadSchema: "demo.v1",
		Priority:      "high",
		OrderingKey:   "account:42",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("enqueue returned an empty job ID")
	}

	select {
	case got := <-received:
		if got.ID != id {
			t.Fatalf("handler got ID %q, want %q", got.ID, id)
		}
		if !bytes.Equal(got.Payload, wantPayload) {
			t.Fatalf("handler payload changed: got %q, want %q", got.Payload, wantPayload)
		}
		if got.PayloadSchema != "demo.v1" || got.OrderingKey != "account:42" {
			t.Fatalf("handler metadata changed: %#v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker handler did not receive the client-enqueued job")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		count, countErr := workerRedis.LLen(context.Background(), clientCfg.CompletedList).Result()
		if countErr == nil && count == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("completed count did not reach one")
}

func TestEnqueueBatchRejectsAllBeforeRedisWhenOnePayloadIsOversized(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	cfg.MaxPayloadSize = 4
	client := newTestClient(t, mr, cfg)

	jobs := []queueclient.Job{
		{ID: "fits", Priority: "low", Payload: []byte("1234")},
		{ID: "too-large", Priority: "low", Payload: []byte("12345")},
	}
	err := client.EnqueueBatch(context.Background(), jobs)
	var sizeErr *queueclient.PayloadTooLargeError
	if !errors.As(err, &sizeErr) {
		t.Fatalf("expected PayloadTooLargeError, got %v", err)
	}
	if sizeErr.Size != 5 || sizeErr.Limit != 4 {
		t.Fatalf("unexpected payload error: %#v", sizeErr)
	}
	if mr.Exists(cfg.Queues["low"]) {
		got, listErr := mr.List(cfg.Queues["low"])
		t.Fatalf("batch rejection changed Redis: %v (err=%v)", got, listErr)
	}
}

func TestDuplicateExplicitIDsAreSeparateDeliveries(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)

	for _, payload := range [][]byte{[]byte("first"), []byte("second")} {
		id, err := client.Enqueue(context.Background(), queueclient.Job{
			ID:       "caller-owned-id",
			Priority: "low",
			Payload:  payload,
		})
		if err != nil {
			t.Fatal(err)
		}
		if id != "caller-owned-id" {
			t.Fatalf("explicit ID changed to %q", id)
		}
	}
	items, err := mr.List(cfg.Queues["low"])
	if err != nil {
		t.Fatal(err)
	}
	if got := len(items); got != 2 {
		t.Fatalf("duplicate IDs should remain two deliveries, got %d", got)
	}
}

func TestInvalidPriorityDoesNotModifyRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)

	_, err := client.Enqueue(context.Background(), queueclient.Job{Priority: "urgent"})
	var priorityErr *queueclient.UnknownPriorityError
	if !errors.As(err, &priorityErr) {
		t.Fatalf("expected UnknownPriorityError, got %v", err)
	}
	if got := len(mr.Keys()); got != 0 {
		t.Fatalf("invalid priority created %d Redis keys", got)
	}
}

func TestRedisFailureReturnsTypedConnectionError(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client, err := queueclient.New(&redis.Options{
		Addr:         mr.Addr(),
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
		MaxRetries:   -1,
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	mr.Close()

	_, err = client.Enqueue(context.Background(), queueclient.Job{Priority: "low"})
	var connectionErr *queueclient.ConnectionError
	if !errors.As(err, &connectionErr) {
		t.Fatalf("expected ConnectionError, got %T: %v", err, err)
	}
	if !errors.Is(err, queueclient.ErrConnection) {
		t.Fatalf("expected errors.Is(err, ErrConnection), got %v", err)
	}
}

func TestStatsAndPeekMirrorCoreQueueLayout(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)

	for i := 0; i < 2; i++ {
		if _, err := client.Enqueue(context.Background(), queueclient.Job{
			ID:       fmt.Sprintf("job-%d", i),
			Priority: "high",
		}); err != nil {
			t.Fatal(err)
		}
	}
	mr.Set(fmt.Sprintf(cfg.HeartbeatKeyPattern, "worker-a"), "1")
	mr.Lpush(fmt.Sprintf(cfg.ProcessingListPattern, "worker-a"), "in-flight")

	stats, err := client.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := stats.Queues["high("+cfg.Queues["high"]+")"]; got != 2 {
		t.Fatalf("high queue count = %d, want 2", got)
	}
	if stats.Heartbeats != 1 || stats.ProcessingLists[fmt.Sprintf(cfg.ProcessingListPattern, "worker-a")] != 1 {
		t.Fatalf("unexpected worker stats: %#v", stats)
	}

	peek, err := client.Peek(context.Background(), "high", 1)
	if err != nil {
		t.Fatal(err)
	}
	if peek.Queue != cfg.Queues["high"] || len(peek.Items) != 1 {
		t.Fatalf("unexpected peek: %#v", peek)
	}
}
