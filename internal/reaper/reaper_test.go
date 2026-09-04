// Copyright 2025 James Ross
package reaper

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type competingUnorderedReaperHook struct {
	once        sync.Once
	competitor  *redis.Client
	processing  string
	destination string
	err         error
}

type failOrderedDiscardOnceHook struct {
	once   sync.Once
	failed chan struct{}
}

const canonicalReaperWorkerID = "reaper-test-host-100-1000000000-abcd-0"

func (*failOrderedDiscardOnceHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *failOrderedDiscardOnceHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		args := cmd.Args()
		if len(args) > 0 && fmt.Sprint(args[len(args)-1]) == string(queue.OrderedDiscard) {
			failed := false
			h.once.Do(func() {
				failed = true
				close(h.failed)
			})
			if failed {
				return errors.New("injected ordered discard failure")
			}
		}
		return next(ctx, cmd)
	}
}

func (*failOrderedDiscardOnceHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (*competingUnorderedReaperHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *competingUnorderedReaperHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if err != nil || cmd.Name() != "lindex" || len(cmd.Args()) < 3 || fmt.Sprint(cmd.Args()[1]) != h.processing || fmt.Sprint(cmd.Args()[2]) != "-1" {
			return err
		}
		h.once.Do(func() {
			payload, popErr := h.competitor.RPop(ctx, h.processing).Result()
			if popErr != nil {
				h.err = popErr
				return
			}
			h.err = h.competitor.LPush(ctx, h.destination, payload).Err()
		})
		return err
	}
}

func (*competingUnorderedReaperHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

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
	workerID := canonicalReaperWorkerID
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
	processing := queuekeys.Format(cfg.Worker.ProcessingListPattern, canonicalReaperWorkerID)

	job := queue.NewJob("custom-key-job", "", 0, "low", "", "")
	payload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(context.Background(), processing, payload).Err(); err != nil {
		t.Fatal(err)
	}

	New(cfg, rdb, zap.NewNop()).scanOnce(context.Background())

	if got, err := rdb.LLen(context.Background(), "custom:low").Result(); err != nil || got != 1 {
		t.Fatalf("custom queue length = %d, want 1 (err=%v)", got, err)
	}
	if mr.Exists(processing) {
		t.Fatal("custom processing list was not drained")
	}
}

func TestReaperPreservesUnrelatedBroadPatternMatches(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	cfg.Worker.ProcessingListPattern = "tenant:%s"
	cfg.Worker.HeartbeatKeyPattern = "heartbeat:%s"
	cfg.Worker.Queues["low"] = "managed:low"
	ctx := context.Background()

	job := queue.NewJob("unrelated-job-shaped-value", "", 0, "low", "", "")
	payload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{payload, "unrelated-malformed-value"}
	if err := rdb.RPush(ctx, "tenant:settings", want).Err(); err != nil {
		t.Fatal(err)
	}

	New(cfg, rdb, zap.NewNop()).scanOnce(ctx)

	if got := rdb.LRange(ctx, "tenant:settings", 0, -1).Val(); !slices.Equal(got, want) {
		t.Fatalf("unrelated broad-pattern list = %#v, want preserved %#v", got, want)
	}
	if got := rdb.LRange(ctx, cfg.Worker.Queues["low"], 0, -1).Val(); len(got) != 0 {
		t.Fatalf("managed queue received unrelated data: %#v", got)
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

func TestReaperAdvancesOrderedKeyAfterMalformedDiscardFailure(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()

	const orderingKey = "account:malformed-discard"
	digest := queuekeys.OrderingDigest(orderingKey)
	queueKey := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest)
	malformed := `{"id":`
	later := queue.NewJob("after-malformed", "", 0, "low", "", "")
	later.OrderingKey = orderingKey
	laterRaw, err := later.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.RPush(ctx, queueKey, laterRaw, malformed).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.SAdd(ctx, cfg.Queue.OrderedActiveSet, digest).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, cfg.Queue.OrderedReadyList, digest).Err(); err != nil {
		t.Fatal(err)
	}

	hook := &failOrderedDiscardOnceHook{failed: make(chan struct{})}
	rdb.AddHook(hook)
	w := worker.New(cfg, rdb, zap.NewNop())
	w.Handle(worker.Handler(func(context.Context, queue.Job) error { return nil }))
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- w.Run(workerCtx) }()
	select {
	case <-hook.failed:
		cancel()
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("worker did not attempt to discard malformed ordered job")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("worker stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
	processingKeys := rdb.Keys(ctx, queuekeys.ScanPattern(cfg.Worker.ProcessingListPattern)).Val()
	if len(processingKeys) != 1 {
		t.Fatalf("processing keys after failed discard = %#v, want one", processingKeys)
	}
	claimsKey := queuekeys.OrderedClaimsKey(cfg.Queue.OrderedActiveSet)
	if got, err := rdb.HGet(ctx, claimsKey, processingKeys[0]).Result(); err != nil || got != digest {
		t.Fatalf("durable claim digest = %q (err=%v), want %q", got, err, digest)
	}

	mr.FastForward(cfg.Worker.HeartbeatTTL + time.Millisecond)
	New(cfg, rdb, zap.NewNop()).scanOnce(ctx)
	if got, err := rdb.HLen(ctx, claimsKey).Result(); err != nil || got != 0 {
		t.Fatalf("ordered claim registry size after recovery = %d (err=%v), want 0", got, err)
	}

	delivery, ok, err := queue.ClaimOrdered(
		ctx,
		rdb,
		cfg.OrderingLayout(),
		"test:recovered:processing",
		"test:recovered:heartbeat",
		"recovered-worker",
		time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("next same-key job remained stranded after malformed discard failure")
	}
	claimed, err := queue.UnmarshalJob(delivery.Payload)
	if err != nil {
		t.Fatalf("claimed next job is invalid: %v", err)
	}
	if claimed.ID != later.ID {
		t.Fatalf("claimed job = %q, want %q", claimed.ID, later.ID)
	}
}

func TestReaperDoesNotMisrouteOrderedTailExposedByConcurrentUnorderedRecovery(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	competitor := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
		_ = competitor.Close()
	})
	cfg := config.Default()
	ctx := context.Background()
	workerID := canonicalReaperWorkerID
	processing := fmt.Sprintf(cfg.Worker.ProcessingListPattern, workerID)

	unordered := queue.NewJob("unordered-tail", "", 0, "low", "", "")
	unorderedRaw, err := unordered.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ordered := queue.NewJob("ordered-behind-tail", "", 0, "low", "", "")
	ordered.OrderingKey = "account:ordered"
	orderedRaw, err := ordered.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.RPush(ctx, processing, orderedRaw, unorderedRaw).Err(); err != nil {
		t.Fatal(err)
	}

	hook := &competingUnorderedReaperHook{
		competitor:  competitor,
		processing:  processing,
		destination: cfg.Worker.Queues["low"],
	}
	rdb.AddHook(hook)
	New(cfg, rdb, zap.NewNop()).scanOnce(ctx)
	if hook.err != nil {
		t.Fatalf("competing recovery failed: %v", hook.err)
	}
	if got := rdb.LRange(ctx, cfg.Worker.Queues["low"], 0, -1).Val(); !slices.Equal(got, []string{unorderedRaw}) {
		t.Fatalf("ordinary queue = %#v, want only concurrently recovered unordered job", got)
	}
	digest := queuekeys.OrderingDigest(ordered.OrderingKey)
	orderedQueue := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest)
	if got := rdb.LRange(ctx, orderedQueue, 0, -1).Val(); !slices.Equal(got, []string{orderedRaw}) {
		t.Fatalf("ordered queue = %#v, want exposed ordered job recovered through FIFO", got)
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
