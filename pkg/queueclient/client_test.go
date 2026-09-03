// Copyright 2026 James Ross
package queueclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
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

func TestNormalizeConfigRejectsIdenticalOrderedQueueAndLeasePatterns(t *testing.T) {
	cfg := testClientConfig()
	cfg.OrderedQueuePattern = "custom:ordered:%s"
	cfg.OrderedLeasePattern = cfg.OrderedQueuePattern
	if _, err := queueclient.NormalizeConfig(cfg); err == nil {
		t.Fatal("expected identical ordered queue and lease patterns to fail")
	}
}

func TestNormalizeConfigRejectsIdenticalOrderedReadyAndActiveKeys(t *testing.T) {
	cfg := testClientConfig()
	cfg.OrderedReadyList = "custom:ordered:control"
	cfg.OrderedActiveSet = cfg.OrderedReadyList
	if _, err := queueclient.NormalizeConfig(cfg); err == nil {
		t.Fatal("expected identical ordered ready and active keys to fail")
	}
}

func TestNormalizeConfigRejectsReservedPriorityAliases(t *testing.T) {
	for _, alias := range []string{"completed", "Completed", "dead_letter", "DEAD_LETTER", "dlq", "DLQ"} {
		t.Run(alias, func(t *testing.T) {
			cfg := testClientConfig()
			cfg.Queues[alias] = "jobqueue:test:reserved"
			if _, err := queueclient.NormalizeConfig(cfg); err == nil {
				t.Fatalf("expected reserved priority alias %q to fail", alias)
			}
		})
	}
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

func TestEnqueueBatchDoesNotPartiallyWriteOnRedisTypeError(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)
	mr.Set(cfg.Queues["high"], "not-a-list")

	jobs := []queueclient.Job{
		{Priority: "low", Payload: []byte("would-be-written-first")},
		{Priority: "high", Payload: []byte("wrong-type")},
	}
	err := client.EnqueueBatch(context.Background(), jobs)
	var connectionErr *queueclient.ConnectionError
	if !errors.As(err, &connectionErr) {
		t.Fatalf("expected ConnectionError, got %T: %v", err, err)
	}
	if mr.Exists(cfg.Queues["low"]) {
		items, listErr := mr.List(cfg.Queues["low"])
		t.Fatalf("failed batch partially changed Redis: %v (err=%v)", items, listErr)
	}
	for i, job := range jobs {
		if job.ID != "" || job.CreationTime != "" {
			t.Errorf("failed batch mutated job %d: %#v", i, job)
		}
	}
}

func TestEnqueueBatchPrevalidatesOrderedControlTypes(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	cfg.OrderedReadyList = "custom:ordered:ready"
	cfg.OrderedActiveSet = "custom:ordered:active"
	cfg.OrderedQueuePattern = "custom:ordered:queue:%s"
	cfg.OrderedLeasePattern = "custom:ordered:lease:%s"
	client := newTestClient(t, mr, cfg)
	mr.Set(cfg.OrderedActiveSet, "not-a-set")

	jobs := []queueclient.Job{
		{Priority: "low", Payload: []byte("ordinary-first")},
		{Priority: "high", OrderingKey: "account:42", Payload: []byte("ordered-second")},
	}
	if err := client.EnqueueBatch(context.Background(), jobs); err == nil {
		t.Fatal("expected wrong-type batch error")
	}
	if mr.Exists(cfg.Queues["low"]) {
		items, listErr := mr.List(cfg.Queues["low"])
		t.Fatalf("ordered control error partially changed Redis: %v (err=%v)", items, listErr)
	}
}

func TestEnqueueBatchCopiesGeneratedMetadataAfterSuccess(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)
	jobs := []queueclient.Job{
		{Priority: "low", Payload: []byte("ordinary")},
		{Priority: "high", OrderingKey: "account:42", Payload: []byte("ordered")},
	}

	if err := client.EnqueueBatch(context.Background(), jobs); err != nil {
		t.Fatal(err)
	}
	for i, job := range jobs {
		if job.ID == "" || job.CreationTime == "" {
			t.Errorf("successful batch did not populate job %d: %#v", i, job)
		}
	}
}

func TestEnqueueBatchUsesConfiguredOrderedFIFO(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	cfg.OrderedReadyList = "custom:ready"
	cfg.OrderedActiveSet = "custom:active"
	cfg.OrderedQueuePattern = "custom:queue:%s"
	cfg.OrderedLeasePattern = "custom:lease:%s"
	client := newTestClient(t, mr, cfg)

	jobs := []queueclient.Job{
		{ID: "first", Priority: "low", OrderingKey: "same"},
		{ID: "second", Priority: "high", OrderingKey: "same"},
		{ID: "third", Priority: "low", OrderingKey: "same"},
	}
	if err := client.EnqueueBatch(context.Background(), jobs); err != nil {
		t.Fatal(err)
	}
	queueKey := fmt.Sprintf(cfg.OrderedQueuePattern, queuekeys.OrderingDigest("same"))
	items, err := mr.List(queueKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("ordered batch length = %d, want 3", len(items))
	}
	claimIndex := 0
	for _, encoded := range slices.Backward(items) {
		var job queueclient.Job
		if err := json.Unmarshal([]byte(encoded), &job); err != nil {
			t.Fatal(err)
		}
		if want := jobs[claimIndex].ID; job.ID != want {
			t.Fatalf("claim order index %d = %q, want %q", claimIndex, job.ID, want)
		}
		claimIndex++
	}
	if got, err := mr.List(cfg.OrderedReadyList); err != nil || len(got) != 1 {
		t.Fatalf("ready tokens = %v (err=%v), want one", got, err)
	}
	if got, err := mr.SMembers(cfg.OrderedActiveSet); err != nil || len(got) != 1 {
		t.Fatalf("active members = %v (err=%v), want one", got, err)
	}
	if mr.Exists(cfg.Queues["high"]) || mr.Exists(cfg.Queues["low"]) {
		t.Fatal("ordered batch touched ordinary priority lists")
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

func TestEnqueueResetsCallerControlledRetries(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)

	if _, err := client.Enqueue(context.Background(), queueclient.Job{
		ID:       "retry-single",
		Priority: "low",
		Retries:  math.MaxInt,
	}); err != nil {
		t.Fatal(err)
	}
	items, err := mr.List(cfg.Queues["low"])
	if err != nil || len(items) != 1 {
		t.Fatalf("stored jobs = %#v (err=%v)", items, err)
	}
	var stored queueclient.Job
	if err := json.Unmarshal([]byte(items[0]), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Retries != 0 {
		t.Fatalf("stored retries = %d, want 0", stored.Retries)
	}
}

func TestEnqueueBatchResetsCallerControlledRetries(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	client := newTestClient(t, mr, cfg)
	jobs := []queueclient.Job{{
		ID:       "retry-batch",
		Priority: "high",
		Retries:  math.MaxInt,
	}}

	if err := client.EnqueueBatch(context.Background(), jobs); err != nil {
		t.Fatal(err)
	}
	if jobs[0].Retries != 0 {
		t.Fatalf("returned retries = %d, want 0", jobs[0].Retries)
	}
	items, err := mr.List(cfg.Queues["high"])
	if err != nil || len(items) != 1 {
		t.Fatalf("stored jobs = %#v (err=%v)", items, err)
	}
	var stored queueclient.Job
	if err := json.Unmarshal([]byte(items[0]), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Retries != 0 {
		t.Fatalf("stored retries = %d, want 0", stored.Retries)
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
	for i := 0; i < 3; i++ {
		if _, err := client.Enqueue(context.Background(), queueclient.Job{
			ID:          fmt.Sprintf("ordered-job-%d", i),
			Priority:    "low",
			OrderingKey: "account:42",
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
	if stats.OrderedPending != 3 {
		t.Fatalf("ordered pending count = %d, want 3", stats.OrderedPending)
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

func TestPeekPreservesExactCaseSensitiveQueueAliases(t *testing.T) {
	mr := miniredis.RunT(t)
	cfg := testClientConfig()
	cfg.Queues = map[string]string{
		"Urgent": "jobqueue:test:urgent-title-case",
		"urgent": "jobqueue:test:urgent-lower-case",
	}
	cfg.DefaultPriority = "Urgent"
	client := newTestClient(t, mr, cfg)

	mr.Lpush(cfg.Queues["Urgent"], "title-case-job")
	mr.Lpush(cfg.Queues["urgent"], "lower-case-job")

	for alias, wantQueue := range cfg.Queues {
		peek, err := client.Peek(context.Background(), alias, 1)
		if err != nil {
			t.Fatalf("peek %q: %v", alias, err)
		}
		if peek.Queue != wantQueue {
			t.Errorf("peek %q resolved queue %q, want %q", alias, peek.Queue, wantQueue)
		}
	}
}
