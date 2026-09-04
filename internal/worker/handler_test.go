// Copyright 2026 James Ross
package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	miniredisserver "github.com/alicebob/miniredis/v2/server"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func newHandlerTestWorker(t *testing.T, count, maxRetries int, log *zap.Logger) (*Worker, *config.Config, *redis.Client) {
	t.Helper()
	w, cfg, rdb, _ := newHandlerTestWorkerWithServer(t, count, maxRetries, log)
	return w, cfg, rdb
}

func newHandlerTestWorkerWithServer(t *testing.T, count, maxRetries int, log *zap.Logger) (*Worker, *config.Config, *redis.Client, *miniredis.Miniredis) {
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
	return New(cfg, rdb, log), cfg, rdb, mr
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

func TestRunRequiresHandlerBeforeStarting(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	job := queue.NewJob("must-not-ack", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.Run(ctx); !errors.Is(err, ErrHandlerRequired) {
		t.Fatalf("worker without an application handler returned %v, want ErrHandlerRequired", err)
	}
	if queued, err := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result(); err != nil || queued != 1 {
		t.Fatalf("handler-less worker changed source queue: length=%d err=%v", queued, err)
	}
}

func TestClearingHandlerPausesConsumptionUntilReplacement(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	activeStarted := make(chan struct{})
	releaseActive := make(chan struct{})
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() { close(releaseActive) })
	}
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		if job.ID == "active" {
			close(activeStarted)
			<-releaseActive
		}
		return nil
	}))

	active := queue.NewJob("active", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], active, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}
	cancel, done := startHandlerTestWorker(w)
	t.Cleanup(func() {
		release()
		stopHandlerTestWorker(t, cancel, done)
	})
	select {
	case <-activeStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("active handler did not start")
	}

	w.Handle(nil)
	held := queue.NewJob("held", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], held, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}
	release()
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 1)
	time.Sleep(100 * time.Millisecond)

	if got, err := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result(); err != nil || got != 1 {
		t.Fatalf("queue length with cleared handler = %d (err=%v), want held job untouched", got, err)
	}
	if got, err := rdb.LLen(context.Background(), cfg.Worker.DeadLetterList).Result(); err != nil || got != 0 {
		t.Fatalf("dead-letter length with cleared handler = %d (err=%v), want 0", got, err)
	}

	replacementCalled := make(chan queue.Job, 1)
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		replacementCalled <- job
		return nil
	}))
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 2)
	select {
	case got := <-replacementCalled:
		if got.ID != held.ID {
			t.Fatalf("replacement handler received %q, want %q", got.ID, held.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("replacement handler did not receive held job")
	}
}

func TestClearingHandlerRestoresDeliveryFromBlockedDequeue(t *testing.T) {
	w, cfg, rdb, mr := newHandlerTestWorkerWithServer(t, 1, 0, nil)
	blocked := make(chan struct{})
	var blockedOnce sync.Once
	mr.Server().SetPreHook(func(_ *miniredisserver.Peer, command string, _ ...string) bool {
		if command == "BRPOPLPUSH" {
			blockedOnce.Do(func() { close(blocked) })
		}
		return false
	})

	called := make(chan queue.Job, 1)
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		called <- job
		return nil
	}))
	cancel, done := startHandlerTestWorker(w)
	t.Cleanup(func() { stopHandlerTestWorker(t, cancel, done) })
	select {
	case <-blocked:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not enter blocking dequeue")
	}

	w.Handle(nil)
	held := queue.NewJob("blocked-held", "", 0, "low", "", "")
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], held, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	if got, err := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result(); err != nil || got != 1 {
		t.Fatalf("queue length after clearing blocked dequeue = %d (err=%v), want held job restored", got, err)
	}
	processing := queuekeys.Format(cfg.Worker.ProcessingListPattern, w.baseID+"-0")
	if got, err := rdb.LLen(context.Background(), processing).Result(); err != nil || got != 0 {
		t.Fatalf("processing length after clearing blocked dequeue = %d (err=%v), want 0", got, err)
	}
	heartbeat := queuekeys.Format(cfg.Worker.HeartbeatKeyPattern, w.baseID+"-0")
	if got, err := rdb.Exists(context.Background(), heartbeat).Result(); err != nil || got != 0 {
		t.Fatalf("heartbeat exists after clearing blocked dequeue = %d (err=%v), want 0", got, err)
	}
	select {
	case job := <-called:
		t.Fatalf("cleared handler received held job %q", job.ID)
	default:
	}

	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		called <- job
		return nil
	}))
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 1)
	select {
	case got := <-called:
		if got.ID != held.ID {
			t.Fatalf("replacement handler received %q, want %q", got.ID, held.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("replacement handler did not receive restored job")
	}
}

func TestHandlerClearedBeforeOrderedClaimRestoresDelivery(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	w.Handle(nil)
	job := queue.NewJob("ordered-held", "", 0, "low", "", "")
	job.OrderingKey = "paused-account"
	if err := queue.EnqueueWithOrdering(
		context.Background(),
		rdb,
		cfg.Worker.Queues["low"],
		job,
		cfg.Queue.MaxPayloadSize,
		cfg.OrderingLayout(),
	); err != nil {
		t.Fatal(err)
	}

	const workerID = "handlerless-ordered"
	processing := queuekeys.Format(cfg.Worker.ProcessingListPattern, workerID)
	heartbeat := queuekeys.Format(cfg.Worker.HeartbeatKeyPattern, workerID)
	if _, found, err := w.claimOrdered(context.Background(), workerID, processing, heartbeat); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatal("ordered claim remained assigned without a handler")
	}

	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest)
	leaseKey := queuekeys.Format(cfg.Queue.OrderedLeasePattern, digest)
	if got, err := rdb.LLen(context.Background(), queueKey).Result(); err != nil || got != 1 {
		t.Fatalf("ordered queue length after handlerless claim = %d (err=%v), want restored job", got, err)
	}
	for _, key := range []string{processing, heartbeat, leaseKey} {
		if got, err := rdb.Exists(context.Background(), key).Result(); err != nil || got != 0 {
			t.Fatalf("handlerless ordered claim left key %q: exists=%d err=%v", key, got, err)
		}
	}
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
	if generation := rdb.Get(context.Background(), queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)).Val(); generation != "1" {
		t.Fatalf("dead-letter generation = %q, want 1", generation)
	}
}

func TestDeadLetterAppendFailureLeavesJobInProcessing(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	w.Handle(Handler(func(_ context.Context, _ queue.Job) error {
		return errors.New("terminal handler failure")
	}))
	ctx := context.Background()
	workerID := w.baseID + "-0"
	processing := queuekeys.Format(cfg.Worker.ProcessingListPattern, workerID)
	heartbeat := queuekeys.Format(cfg.Worker.HeartbeatKeyPattern, workerID)

	job := queue.NewJob("preserve-on-dlq-error", "", 0, "low", "", "")
	payload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, processing, payload).Err(); err != nil {
		t.Fatal(err)
	}
	generationKey := queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)
	if err := rdb.Set(ctx, generationKey, "invalid", 0).Err(); err != nil {
		t.Fatal(err)
	}

	if completed := w.processJob(ctx, workerID, cfg.Worker.Queues["low"], processing, heartbeat, payload); completed {
		t.Fatal("terminal handler failure reported successful completion")
	}
	if got := rdb.LRange(ctx, processing, 0, -1).Val(); !slices.Equal(got, []string{payload}) {
		t.Fatalf("processing after failed DLQ append = %#v, want original envelope", got)
	}
	if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); len(got) != 0 {
		t.Fatalf("DLQ after rejected append = %#v, want empty", got)
	}
	if got := rdb.Get(ctx, generationKey).Val(); got != "invalid" {
		t.Fatalf("DLQ generation after rejected append = %q, want invalid value preserved", got)
	}
	if exists := rdb.Exists(ctx, heartbeat).Val(); exists != 1 {
		t.Fatalf("heartbeat after failed DLQ append exists = %d, want natural-expiry handoff", exists)
	}
}

func TestRetryEnqueueFailureLeavesJobInProcessing(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 1, nil)
	w.Handle(Handler(func(_ context.Context, _ queue.Job) error {
		return errors.New("retryable handler failure")
	}))
	ctx := context.Background()
	workerID := w.baseID + "-0"
	processing := queuekeys.Format(cfg.Worker.ProcessingListPattern, workerID)
	heartbeat := queuekeys.Format(cfg.Worker.HeartbeatKeyPattern, workerID)
	source := cfg.Worker.Queues["low"]

	job := queue.NewJob("preserve-on-retry-error", "", 0, "low", "", "")
	payload, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, processing, payload).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(ctx, source, "wrong-type-source", 0).Err(); err != nil {
		t.Fatal(err)
	}

	if completed := w.processJob(ctx, workerID, source, processing, heartbeat, payload); completed {
		t.Fatal("retryable handler failure reported successful completion")
	}
	if got := rdb.LRange(ctx, processing, 0, -1).Val(); !slices.Equal(got, []string{payload}) {
		t.Fatalf("processing after failed retry enqueue = %#v, want original envelope", got)
	}
	if got := rdb.Get(ctx, source).Val(); got != "wrong-type-source" {
		t.Fatalf("source after failed retry enqueue = %q, want wrong-type value preserved", got)
	}
	if exists := rdb.Exists(ctx, heartbeat).Val(); exists != 1 {
		t.Fatalf("heartbeat after failed retry enqueue exists = %d, want natural-expiry handoff", exists)
	}
}

func TestHandlerPayloadMutationDoesNotChangeRetryEnvelope(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 1, nil)
	wantPayload := []byte{0x00, 0x01, 0x7f, 0x80, 0xfe, 0xff}
	var mu sync.Mutex
	var attempts [][]byte
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		mu.Lock()
		attempts = append(attempts, bytes.Clone(job.Payload))
		attempt := len(attempts)
		mu.Unlock()

		for i := range job.Payload {
			job.Payload[i] ^= 0xff
		}
		if attempt == 1 {
			return errors.New("retry without persisting handler mutation")
		}
		return nil
	}))

	job := queue.NewJob("mutates-payload", "", 0, "low", "", "")
	job.Payload = bytes.Clone(wantPayload)
	if err := queue.Enqueue(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize); err != nil {
		t.Fatal(err)
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 1)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	mu.Lock()
	gotAttempts := append([][]byte(nil), attempts...)
	mu.Unlock()
	if len(gotAttempts) != 2 {
		t.Fatalf("handler attempts = %d, want initial attempt plus one retry", len(gotAttempts))
	}
	for i, got := range gotAttempts {
		if !bytes.Equal(got, wantPayload) {
			t.Fatalf("attempt %d payload = %v, want original %v", i+1, got, wantPayload)
		}
	}

	encoded, err := rdb.LIndex(context.Background(), cfg.Worker.CompletedList, 0).Result()
	if err != nil {
		t.Fatal(err)
	}
	completed, err := queue.UnmarshalJob(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Retries != 1 {
		t.Fatalf("completed retries = %d, want 1", completed.Retries)
	}
	if !bytes.Equal(completed.Payload, wantPayload) {
		t.Fatalf("completed payload = %v, want original %v", completed.Payload, wantPayload)
	}
}

func TestOrderedHandlerRetryRunsBeforeLaterSameKeyJob(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 2, 1, nil)
	var mu sync.Mutex
	calls := make([]string, 0, 3)
	firstAttempts := 0
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		mu.Lock()
		calls = append(calls, job.ID)
		if job.ID == "ordered-first" {
			firstAttempts++
			if firstAttempts == 1 {
				mu.Unlock()
				return errors.New("retry first")
			}
		}
		mu.Unlock()
		return nil
	}))

	for _, id := range []string{"ordered-first", "ordered-later"} {
		job := queue.NewJob(id, "", 0, "low", "", "")
		job.OrderingKey = "same-key"
		if err := queue.EnqueueWithOrdering(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize, cfg.OrderingLayout()); err != nil {
			t.Fatal(err)
		}
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 2)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	mu.Lock()
	got := append([]string(nil), calls...)
	mu.Unlock()
	want := []string{"ordered-first", "ordered-first", "ordered-later"}
	if !slices.Equal(got, want) {
		t.Fatalf("ordered retry calls = %v, want %v", got, want)
	}
}

func TestOrderedDeadLetterReleasesLaterSameKeyJob(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 2, 0, nil)
	var mu sync.Mutex
	calls := make([]string, 0, 2)
	w.Handle(Handler(func(_ context.Context, job queue.Job) error {
		mu.Lock()
		calls = append(calls, job.ID)
		mu.Unlock()
		if job.ID == "ordered-dead" {
			return errors.New("dead letter this job")
		}
		return nil
	}))

	for _, id := range []string{"ordered-dead", "ordered-after-dead"} {
		job := queue.NewJob(id, "", 0, "low", "", "")
		job.OrderingKey = "same-dead-letter-key"
		if err := queue.EnqueueWithOrdering(context.Background(), rdb, cfg.Worker.Queues["low"], job, cfg.Queue.MaxPayloadSize, cfg.OrderingLayout()); err != nil {
			t.Fatal(err)
		}
	}

	cancel, done := startHandlerTestWorker(w)
	waitForListLength(t, rdb, cfg.Worker.DeadLetterList, 1)
	waitForListLength(t, rdb, cfg.Worker.CompletedList, 1)
	waitForNoProcessingJobs(t, rdb)
	stopHandlerTestWorker(t, cancel, done)

	mu.Lock()
	got := append([]string(nil), calls...)
	mu.Unlock()
	want := []string{"ordered-dead", "ordered-after-dead"}
	if !slices.Equal(got, want) {
		t.Fatalf("ordered dead-letter calls = %v, want %v", got, want)
	}
	if generation := rdb.Get(context.Background(), queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)).Val(); generation != "1" {
		t.Fatalf("ordered dead-letter generation = %q, want 1", generation)
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

func TestOrderedLeaseLossCancelsRunningHandler(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	cfg.Worker.HeartbeatTTL = 90 * time.Millisecond
	started := make(chan struct{})
	canceled := make(chan error, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseHandler := func() {
		releaseOnce.Do(func() { close(release) })
	}
	defer releaseHandler()

	w.Handle(Handler(func(ctx context.Context, _ queue.Job) error {
		close(started)
		select {
		case <-ctx.Done():
			canceled <- ctx.Err()
			return ctx.Err()
		case <-release:
			return nil
		}
	}))

	job := queue.NewJob("lose-ordered-lease", "", 0, "low", "", "")
	job.OrderingKey = "lease-loss-key"
	if err := queue.EnqueueWithOrdering(
		context.Background(),
		rdb,
		cfg.Worker.Queues["low"],
		job,
		cfg.Queue.MaxPayloadSize,
		cfg.OrderingLayout(),
	); err != nil {
		t.Fatal(err)
	}

	cancelWorker, done := startHandlerTestWorker(w)
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		stopHandlerTestWorker(t, cancelWorker, done)
		t.Fatal("ordered handler did not start")
	}

	digest := queuekeys.OrderingDigest(job.OrderingKey)
	leaseKey := queuekeys.Format(cfg.OrderingLayout().LeasePattern, digest)
	if owner, err := rdb.Get(context.Background(), leaseKey).Result(); err != nil || owner == "" {
		releaseHandler()
		stopHandlerTestWorker(t, cancelWorker, done)
		t.Fatalf("ordered lease was not acquired: owner=%q err=%v", owner, err)
	}
	if err := rdb.Set(context.Background(), leaseKey, "replacement-worker", time.Second).Err(); err != nil {
		releaseHandler()
		stopHandlerTestWorker(t, cancelWorker, done)
		t.Fatal(err)
	}

	select {
	case err := <-canceled:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("handler cancellation error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		releaseHandler()
		stopHandlerTestWorker(t, cancelWorker, done)
		t.Fatal("ordered handler continued after losing its lease")
	}

	stopHandlerTestWorker(t, cancelWorker, done)
	processing := fmt.Sprintf(cfg.Worker.ProcessingListPattern, w.baseID+"-0")
	if got, err := rdb.LLen(context.Background(), processing).Result(); err != nil || got != 1 {
		t.Fatalf("lease-lost delivery should remain in processing: length=%d err=%v", got, err)
	}
	for _, key := range []string{cfg.Worker.CompletedList, cfg.Worker.DeadLetterList} {
		if got, err := rdb.LLen(context.Background(), key).Result(); err != nil || got != 0 {
			t.Fatalf("lease-lost delivery reached %q: length=%d err=%v", key, got, err)
		}
	}
}
