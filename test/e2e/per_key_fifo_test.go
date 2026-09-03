//go:build e2e_tests
// +build e2e_tests

// Gated because per-key concurrency and crash recovery require a real Redis server.

// Copyright 2026 James Ross
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/internal/reaper"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestE2E_PerKeyFIFOOneKey(t *testing.T) {
	addr := os.Getenv("E2E_REDIS_ADDR")
	if addr == "" {
		t.Skip("E2E_REDIS_ADDR not set; skipping e2e test")
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flushdb: %v", err)
	}

	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Redis.Addr = addr
	cfg.Worker.Count = 8
	cfg.Worker.Priorities = []string{"low"}
	cfg.Worker.BRPopLPushTimeout = time.Second

	client, err := queueclient.NewWithClient(rdb, queueclient.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	for sequence := 0; sequence < 100; sequence++ {
		payload, marshalErr := json.Marshal(struct {
			Sequence int `json:"sequence"`
		}{Sequence: sequence})
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		_, enqueueErr := client.Enqueue(context.Background(), queueclient.Job{
			ID:          fmt.Sprintf("one-key-%03d", sequence),
			Payload:     payload,
			OrderingKey: "repo/path/to/shared-file.go",
			Priority:    "low",
		})
		if enqueueErr != nil {
			t.Fatal(enqueueErr)
		}
	}

	var mu sync.Mutex
	inFlight := 0
	maxInFlight := 0
	completed := make([]int, 0, 100)
	w := worker.New(cfg, rdb, zap.NewNop())
	w.Handle(worker.Handler(func(_ context.Context, job queue.Job) error {
		var body struct {
			Sequence int `json:"sequence"`
		}
		if unmarshalErr := json.Unmarshal(job.Payload, &body); unmarshalErr != nil {
			return unmarshalErr
		}

		mu.Lock()
		inFlight++
		if inFlight > maxInFlight {
			maxInFlight = inFlight
		}
		mu.Unlock()

		// On the unordered implementation this reliably lets later jobs finish
		// before sequence zero. The ordered implementation never overlaps them.
		if body.Sequence == 0 {
			time.Sleep(150 * time.Millisecond)
		} else {
			time.Sleep(time.Millisecond)
		}

		mu.Lock()
		completed = append(completed, body.Sequence)
		inFlight--
		mu.Unlock()
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	waitForRedisListLength(t, rdb, cfg.Worker.CompletedList, 100, 10*time.Second)
	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("worker run: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop")
	}

	want := make([]int, 100)
	for i := range want {
		want[i] = i
	}
	mu.Lock()
	gotOrder := slices.Clone(completed)
	gotMax := maxInFlight
	mu.Unlock()
	if gotMax != 1 || !slices.Equal(gotOrder, want) {
		prefix := min(12, len(gotOrder))
		t.Fatalf("per-key FIFO violated: max_in_flight=%d completed_prefix=%v", gotMax, gotOrder[:prefix])
	}
}

func TestE2E_PerKeyFIFOMultipleKeysRunConcurrently(t *testing.T) {
	rdb, addr := newPerKeyRedis(t)
	cfg := newPerKeyConfig(t, addr, 8)
	client := newPerKeyClient(t, rdb, cfg)

	type body struct {
		Key      int `json:"key"`
		Sequence int `json:"sequence"`
	}
	for sequence := 0; sequence < 10; sequence++ {
		for key := 0; key < 10; key++ {
			payload, err := json.Marshal(body{Key: key, Sequence: sequence})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Enqueue(context.Background(), queueclient.Job{
				ID:          fmt.Sprintf("multi-%02d-%02d", key, sequence),
				Payload:     payload,
				OrderingKey: fmt.Sprintf("path-%02d", key),
				Priority:    "low",
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	var mu sync.Mutex
	perKeyInFlight := make(map[int]int, 10)
	perKeyMax := make(map[int]int, 10)
	completed := make(map[int][]int, 10)
	totalInFlight := 0
	maxTotalInFlight := 0
	w := worker.New(cfg, rdb, zap.NewNop()).Handle(func(_ context.Context, job queue.Job) error {
		var value body
		if err := json.Unmarshal(job.Payload, &value); err != nil {
			return err
		}
		mu.Lock()
		perKeyInFlight[value.Key]++
		if perKeyInFlight[value.Key] > perKeyMax[value.Key] {
			perKeyMax[value.Key] = perKeyInFlight[value.Key]
		}
		totalInFlight++
		if totalInFlight > maxTotalInFlight {
			maxTotalInFlight = totalInFlight
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		completed[value.Key] = append(completed[value.Key], value.Sequence)
		perKeyInFlight[value.Key]--
		totalInFlight--
		mu.Unlock()
		return nil
	})
	cancel, done := startPerKeyWorker(w)
	waitForRedisListLength(t, rdb, cfg.Worker.CompletedList, 100, 10*time.Second)
	stopPerKeyWorker(t, cancel, done)

	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	mu.Lock()
	defer mu.Unlock()
	if maxTotalInFlight <= 1 {
		t.Fatalf("different keys never ran concurrently: max_in_flight=%d", maxTotalInFlight)
	}
	for key := 0; key < 10; key++ {
		if perKeyMax[key] != 1 {
			t.Errorf("key %d max in flight = %d, want 1", key, perKeyMax[key])
		}
		if !slices.Equal(completed[key], want) {
			t.Errorf("key %d completion order = %v, want %v", key, completed[key], want)
		}
	}
}

func TestE2E_PerKeyFIFORecoveryRedeliversInterruptedJobFirst(t *testing.T) {
	rdb, addr := newPerKeyRedis(t)
	cfg := newPerKeyConfig(t, addr, 1)
	const leaseTTL = 300 * time.Millisecond
	cfg.Worker.HeartbeatTTL = leaseTTL
	client := newPerKeyClient(t, rdb, cfg)

	for sequence := 0; sequence < 3; sequence++ {
		payload, err := json.Marshal(struct {
			Sequence int `json:"sequence"`
		}{Sequence: sequence})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Enqueue(context.Background(), queueclient.Job{
			ID:          fmt.Sprintf("crash-%d", sequence),
			Payload:     payload,
			OrderingKey: "shared-crash-key",
			Priority:    "low",
		}); err != nil {
			t.Fatal(err)
		}
	}

	started := make(chan struct{})
	firstWorker := worker.New(cfg, rdb, zap.NewNop()).Handle(func(ctx context.Context, job queue.Job) error {
		if job.ID != "crash-0" {
			return fmt.Errorf("later same-key job ran before crash recovery: %s", job.ID)
		}
		close(started)
		<-ctx.Done()
		return ctx.Err()
	})
	cancelFirst, firstDone := startPerKeyWorker(firstWorker)
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("first worker did not claim the interrupted job")
	}
	recoveryStart := time.Now()
	stopPerKeyWorker(t, cancelFirst, firstDone)

	reaperCtx, stopReaper := context.WithCancel(context.Background())
	reaperDone := make(chan struct{})
	go func() {
		defer close(reaperDone)
		reaper.New(cfg, rdb, zap.NewNop()).Run(reaperCtx)
	}()
	defer func() {
		stopReaper()
		<-reaperDone
	}()

	waitForRedisListLength(t, rdb, cfg.Queue.OrderedReadyList, 1, 2*leaseTTL+200*time.Millisecond)
	if elapsed := time.Since(recoveryStart); elapsed > 2*leaseTTL+200*time.Millisecond {
		t.Fatalf("ordered key did not un-wedge near lease expiry: %v", elapsed)
	}

	var mu sync.Mutex
	redelivered := make([]int, 0, 3)
	secondWorker := worker.New(cfg, rdb, zap.NewNop()).Handle(func(_ context.Context, job queue.Job) error {
		var value struct {
			Sequence int `json:"sequence"`
		}
		if err := json.Unmarshal(job.Payload, &value); err != nil {
			return err
		}
		mu.Lock()
		redelivered = append(redelivered, value.Sequence)
		mu.Unlock()
		return nil
	})
	cancelSecond, secondDone := startPerKeyWorker(secondWorker)
	waitForRedisListLength(t, rdb, cfg.Worker.CompletedList, 3, 5*time.Second)
	stopPerKeyWorker(t, cancelSecond, secondDone)

	mu.Lock()
	got := slices.Clone(redelivered)
	mu.Unlock()
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("post-crash order = %v, want [0 1 2]", got)
	}
}

func TestE2E_PerKeyFIFOLeaseHeartbeatOutlivesTTL(t *testing.T) {
	rdb, addr := newPerKeyRedis(t)
	cfg := newPerKeyConfig(t, addr, 2)
	const leaseTTL = 180 * time.Millisecond
	cfg.Worker.HeartbeatTTL = leaseTTL
	client := newPerKeyClient(t, rdb, cfg)
	if _, err := client.Enqueue(context.Background(), queueclient.Job{
		ID:          "long-ordered-handler",
		OrderingKey: "long-handler-key",
		Priority:    "low",
	}); err != nil {
		t.Fatal(err)
	}

	reaperCtx, stopReaper := context.WithCancel(context.Background())
	reaperDone := make(chan struct{})
	go func() {
		defer close(reaperDone)
		reaper.New(cfg, rdb, zap.NewNop()).Run(reaperCtx)
	}()

	var calls atomic.Int32
	w := worker.New(cfg, rdb, zap.NewNop()).Handle(func(_ context.Context, _ queue.Job) error {
		calls.Add(1)
		time.Sleep(4 * leaseTTL)
		return nil
	})
	cancel, done := startPerKeyWorker(w)
	waitForRedisListLength(t, rdb, cfg.Worker.CompletedList, 1, 5*time.Second)
	stopPerKeyWorker(t, cancel, done)
	stopReaper()
	<-reaperDone

	if got := calls.Load(); got != 1 {
		t.Fatalf("long ordered handler calls = %d, want 1", got)
	}
	if got, err := rdb.SCard(context.Background(), cfg.Queue.OrderedActiveSet).Result(); err != nil || got != 0 {
		t.Fatalf("active ordered keys = %d (err=%v)", got, err)
	}
}

func TestE2E_PerKeyFIFOKeySafetyPriorityAndFairness(t *testing.T) {
	rdb, addr := newPerKeyRedis(t)
	cfg := newPerKeyConfig(t, addr, 1)
	cfg.Worker.Priorities = []string{"high", "low"}
	client := newPerKeyClient(t, rdb, cfg)
	const hotKey = "repo:{main}:目录/thing:β.go"

	for sequence := 0; sequence < 10; sequence++ {
		priority := "low"
		if sequence%2 == 0 {
			priority = "high"
		}
		if _, err := client.Enqueue(context.Background(), queueclient.Job{
			ID:          fmt.Sprintf("hot-%d", sequence),
			OrderingKey: hotKey,
			Priority:    priority,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := client.Enqueue(context.Background(), queueclient.Job{
		ID:          "cold-0",
		OrderingKey: "cold-key",
		Priority:    "low",
	}); err != nil {
		t.Fatal(err)
	}

	orderedKeys, err := rdb.Keys(context.Background(), "jobqueue:ordered:queue:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(orderedKeys) != 2 {
		t.Fatalf("ordered queue keys = %v, want two digests", orderedKeys)
	}
	wantHotRedisKey := queuekeys.Format(cfg.Queue.OrderedQueuePattern, queuekeys.OrderingDigest(hotKey))
	if !slices.Contains(orderedKeys, wantHotRedisKey) {
		t.Fatalf("special key digest %q missing from %v", wantHotRedisKey, orderedKeys)
	}
	for _, key := range orderedKeys {
		if strings.Contains(key, "{") || strings.Contains(key, "目录") || strings.Contains(key, "β") {
			t.Fatalf("raw ordering key leaked into Redis key %q", key)
		}
	}

	var mu sync.Mutex
	completed := make([]string, 0, 11)
	w := worker.New(cfg, rdb, zap.NewNop()).Handle(func(_ context.Context, job queue.Job) error {
		mu.Lock()
		completed = append(completed, job.ID)
		mu.Unlock()
		return nil
	})
	cancel, done := startPerKeyWorker(w)
	waitForRedisListLength(t, rdb, cfg.Worker.CompletedList, 11, 5*time.Second)
	stopPerKeyWorker(t, cancel, done)

	mu.Lock()
	got := slices.Clone(completed)
	mu.Unlock()
	want := []string{"hot-0", "cold-0", "hot-1", "hot-2", "hot-3", "hot-4", "hot-5", "hot-6", "hot-7", "hot-8", "hot-9"}
	if !slices.Equal(got, want) {
		t.Fatalf("round-robin/priority order = %v, want %v", got, want)
	}
}

func TestE2E_PerKeyFIFOCompletionRecoveryRaceHasOneWinner(t *testing.T) {
	rdb, _ := newPerKeyRedis(t)
	ctx := context.Background()
	for iteration := 0; iteration < 100; iteration++ {
		prefix := fmt.Sprintf("race:%03d", iteration)
		layout := queue.OrderingLayout{
			ReadyList:    prefix + ":ready",
			ActiveSet:    prefix + ":active",
			QueuePattern: prefix + ":queue:%s",
			LeasePattern: prefix + ":lease:%s",
		}
		job := queue.Job{
			ID:           prefix,
			OrderingKey:  prefix,
			Priority:     "low",
			CreationTime: time.Now().UTC().Format(time.RFC3339Nano),
		}
		payload, err := job.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		delivery := queue.OrderedDeliveryFor(layout, job, payload)
		processing := prefix + ":processing"
		heartbeat := prefix + ":heartbeat"
		completed := prefix + ":completed"
		owner := prefix + ":worker"
		if err := rdb.LPush(ctx, processing, payload).Err(); err != nil {
			t.Fatal(err)
		}
		if err := rdb.SAdd(ctx, layout.ActiveSet, delivery.Digest).Err(); err != nil {
			t.Fatal(err)
		}
		if err := rdb.Set(ctx, delivery.LeaseKey, owner, 2*time.Millisecond).Err(); err != nil {
			t.Fatal(err)
		}
		if iteration%2 == 1 {
			time.Sleep(3 * time.Millisecond)
		}

		start := make(chan struct{})
		var transitionWon, recoveryWon bool
		var transitionErr, recoveryErr error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			transitionWon, transitionErr = queue.TransitionOrdered(ctx, rdb, layout, delivery, processing, heartbeat, owner, completed, payload, queue.OrderedComplete)
		}()
		go func() {
			defer wg.Done()
			<-start
			recoveryWon, recoveryErr = queue.RecoverOrdered(ctx, rdb, layout, delivery, processing, heartbeat)
		}()
		close(start)
		wg.Wait()
		if transitionErr != nil || recoveryErr != nil {
			t.Fatalf("iteration %d errors: transition=%v recovery=%v", iteration, transitionErr, recoveryErr)
		}
		if !transitionWon && !recoveryWon {
			// Both first attempts can legitimately straddle the expiration:
			// recovery sees the old lease, then completion sees it expired.
			time.Sleep(3 * time.Millisecond)
			recoveryWon, recoveryErr = queue.RecoverOrdered(ctx, rdb, layout, delivery, processing, heartbeat)
			if recoveryErr != nil {
				t.Fatalf("iteration %d retry recovery: %v", iteration, recoveryErr)
			}
		}
		winners := 0
		if transitionWon {
			winners++
		}
		if recoveryWon {
			winners++
		}
		if winners != 1 {
			t.Fatalf("iteration %d winners = %d (transition=%v recovery=%v)", iteration, winners, transitionWon, recoveryWon)
		}
		completedCount, _ := rdb.LLen(ctx, completed).Result()
		pendingCount, _ := rdb.LLen(ctx, delivery.QueueKey).Result()
		processingCount, _ := rdb.LLen(ctx, processing).Result()
		if completedCount+pendingCount != 1 || processingCount != 0 {
			t.Fatalf("iteration %d terminal counts: completed=%d pending=%d processing=%d", iteration, completedCount, pendingCount, processingCount)
		}
	}
}

func TestE2E_PerKeyFIFOTenThousandReadyKeysIsConstantTime(t *testing.T) {
	rdb, addr := newPerKeyRedis(t)
	cfg := newPerKeyConfig(t, addr, 1)
	client := newPerKeyClient(t, rdb, cfg)
	ctx := context.Background()
	if err := rdb.ConfigResetStat(ctx).Err(); err != nil {
		t.Fatalf("reset Redis stats: %v", err)
	}

	started := time.Now()
	for i := 0; i < 10_000; i++ {
		if _, err := client.Enqueue(ctx, queueclient.Job{
			ID:          fmt.Sprintf("cardinality-%05d", i),
			OrderingKey: fmt.Sprintf("distinct-key-%05d", i),
			Priority:    "low",
		}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}
	enqueueDuration := time.Since(started)
	active, err := rdb.SCard(ctx, cfg.Queue.OrderedActiveSet).Result()
	if err != nil || active != 10_000 {
		t.Fatalf("active cardinality = %d, want 10000 (err=%v)", active, err)
	}
	ready, err := rdb.LLen(ctx, cfg.Queue.OrderedReadyList).Result()
	if err != nil || ready != 10_000 {
		t.Fatalf("ready length = %d, want 10000 (err=%v)", ready, err)
	}
	commandStats, err := rdb.Info(ctx, "commandstats").Result()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(commandStats, "cmdstat_scan:") {
		t.Fatalf("ordered enqueue issued SCAN:\n%s", commandStats)
	}

	claimStart := time.Now()
	delivery, ok, err := queue.ClaimOrdered(ctx, rdb, cfg.OrderingLayout(), "cardinality:processing", "cardinality:heartbeat", "cardinality-worker", time.Second)
	claimDuration := time.Since(claimStart)
	if err != nil || !ok {
		t.Fatalf("claim from 10000 ready keys = (%v, %v)", ok, err)
	}
	if claimDuration > 250*time.Millisecond {
		t.Fatalf("claim from 10000 ready keys took %v", claimDuration)
	}
	transitioned, err := queue.TransitionOrdered(ctx, rdb, cfg.OrderingLayout(), delivery, "cardinality:processing", "cardinality:heartbeat", "cardinality-worker", "", "", queue.OrderedDiscard)
	if err != nil || !transitioned {
		t.Fatalf("discard measured claim = (%v, %v)", transitioned, err)
	}
	t.Logf("10k distinct ordered enqueues=%v; O(1) oldest-key claim=%v", enqueueDuration, claimDuration)
}

func newPerKeyRedis(t *testing.T) (*redis.Client, string) {
	t.Helper()
	addr := os.Getenv("E2E_REDIS_ADDR")
	if addr == "" {
		t.Skip("E2E_REDIS_ADDR not set; skipping e2e test")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flushdb: %v", err)
	}
	return rdb, addr
}

func newPerKeyConfig(t *testing.T, addr string, workers int) *config.Config {
	t.Helper()
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Redis.Addr = addr
	cfg.Worker.Count = workers
	cfg.Worker.Priorities = []string{"low"}
	cfg.Worker.BRPopLPushTimeout = time.Second
	cfg.Worker.Backoff.Base = time.Millisecond
	cfg.Worker.Backoff.Max = 2 * time.Millisecond
	return cfg
}

func newPerKeyClient(t *testing.T, rdb *redis.Client, cfg *config.Config) *queueclient.Client {
	t.Helper()
	client, err := queueclient.NewWithClient(rdb, queueclient.Config{
		Queues:                cfg.Worker.Queues,
		DefaultPriority:       cfg.Producer.DefaultPriority,
		ProcessingListPattern: cfg.Worker.ProcessingListPattern,
		HeartbeatKeyPattern:   cfg.Worker.HeartbeatKeyPattern,
		CompletedList:         cfg.Worker.CompletedList,
		DeadLetterList:        cfg.Worker.DeadLetterList,
		MaxPayloadSize:        cfg.Queue.MaxPayloadSize,
		OrderedReadyList:      cfg.Queue.OrderedReadyList,
		OrderedActiveSet:      cfg.Queue.OrderedActiveSet,
		OrderedQueuePattern:   cfg.Queue.OrderedQueuePattern,
		OrderedLeasePattern:   cfg.Queue.OrderedLeasePattern,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func startPerKeyWorker(w *worker.Worker) (context.CancelFunc, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	return cancel, done
}

func stopPerKeyWorker(t *testing.T, cancel context.CancelFunc, done <-chan error) {
	t.Helper()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("worker run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop")
	}
}

func waitForRedisListLength(t *testing.T, rdb *redis.Client, key string, want int64, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		got, err := rdb.LLen(context.Background(), key).Result()
		if err == nil && got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, err := rdb.LLen(context.Background(), key).Result()
	t.Fatalf("list %q did not reach %d: got %d (err=%v)", key, want, got, err)
}
