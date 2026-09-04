// Copyright 2026 James Ross
package admin

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

func TestAdminUsesConfiguredWorkerKeyPatterns(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"low":    "custom:low",
				"urgent": "custom:urgent",
			},
			ProcessingListPattern: "custom:{worker:%s}:active",
			HeartbeatKeyPattern:   "custom:{heartbeat:%s}",
			CompletedList:         "custom:completed",
			DeadLetterList:        "custom:dead-letter",
		},
	}

	mr.Lpush("custom:{worker:alpha:one}:active", `{"id":"in-flight"}`)
	mr.Set("custom:{heartbeat:alpha:one}", "1")
	mr.Lpush("custom:urgent", `{"id":"urgent"}`)

	stats, err := Stats(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if stats.ProcessingLists["custom:{worker:alpha:one}:active"] != 1 || stats.Heartbeats != 1 {
		t.Fatalf("configured patterns were not used: %#v", stats)
	}
	workers, err := Workers(context.Background(), cfg, rdb, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(workers) != 1 || workers[0].ID != "alpha:one" || workers[0].JobID != "in-flight" {
		t.Fatalf("configured pattern extraction failed: %#v", workers)
	}

	deleted, err := PurgeAll(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	generationKey := queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)
	if deleted != 3 || len(mr.Keys()) != 1 || !mr.Exists(generationKey) {
		t.Fatalf("purge deleted %d keys; remaining=%v", deleted, mr.Keys())
	}
	if generation, err := mr.Get(generationKey); err != nil || generation != "1" {
		t.Fatalf("DLQ generation after purge = %q (err=%v), want 1", generation, err)
	}
}

func TestPurgeAllRemovesOrderedQueueState(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}

	digest := queuekeys.OrderingDigest("account:42")
	for _, key := range []string{
		cfg.Queue.OrderedReadyList,
		cfg.Queue.OrderedActiveSet,
		queuekeys.OrderedClaimsKey(cfg.Queue.OrderedActiveSet),
		queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest),
		queuekeys.Format(cfg.Queue.OrderedLeasePattern, digest),
	} {
		mr.Set(key, "value")
	}
	deleted, err := PurgeAll(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	generationKey := queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)
	if deleted != 5 || len(mr.Keys()) != 1 || !mr.Exists(generationKey) {
		t.Fatalf("purge deleted %d ordered keys; remaining=%v", deleted, mr.Keys())
	}
	if generation, err := mr.Get(generationKey); err != nil || generation != "1" {
		t.Fatalf("DLQ generation after purge = %q (err=%v), want 1", generation, err)
	}
}

func TestPurgeAllPreservesNonDigestKeysMatchedByOrderedPatterns(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	cfg.Queue.OrderedQueuePattern = "tenant:%s"
	cfg.Queue.OrderedLeasePattern = "tenant:lease:%s"
	digest := queuekeys.OrderingDigest("account:42")
	queueKey := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest)
	leaseKey := queuekeys.Format(cfg.Queue.OrderedLeasePattern, digest)
	const unrelated = "tenant:settings"

	for _, key := range []string{queueKey, leaseKey, unrelated} {
		mr.Set(key, "value")
	}
	deleted, err := PurgeAll(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Fatalf("purge deleted %d keys, want only the generated queue and lease", deleted)
	}
	if !mr.Exists(unrelated) {
		t.Fatal("purge deleted an unrelated key matched by the broad ordered pattern")
	}
	if mr.Exists(queueKey) || mr.Exists(leaseKey) {
		t.Fatalf("generated ordered state remains: %v", mr.Keys())
	}
}

func TestPurgeAllWithoutConfiguredDLQDoesNotCreateMetadata(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := &config.Config{
		Worker: config.Worker{
			Queues:                map[string]string{"low": "custom:low"},
			ProcessingListPattern: "custom:worker:%s",
			HeartbeatKeyPattern:   "custom:heartbeat:%s",
		},
	}
	mr.Lpush("custom:low", `{"id":"queued"}`)

	deleted, err := PurgeAll(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 || len(mr.Keys()) != 0 {
		t.Fatalf("purge deleted %d keys; remaining=%v, want no metadata side effect", deleted, mr.Keys())
	}
}

func TestPurgeAllInvalidDLQGenerationPreservesOtherQueues(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()
	queueKey := cfg.Worker.Queues["high"]
	if err := rdb.LPush(ctx, queueKey, `{"id":"queued"}`).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.HSet(ctx, queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList), "invalid", "type").Err(); err != nil {
		t.Fatal(err)
	}

	if _, err := PurgeAll(ctx, cfg, rdb); err == nil {
		t.Fatal("expected invalid DLQ generation metadata to reject PurgeAll")
	}
	if got := rdb.LLen(ctx, queueKey).Val(); got != 1 {
		t.Fatalf("queue length after rejected purge = %d, want 1", got)
	}
}

func TestStatsIncludeOrderedBacklog(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg, err := config.Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Queue.OrderedQueuePattern = "custom:ordered:%s:pending"
	digestA := queuekeys.OrderingDigest("account:a")
	digestB := queuekeys.OrderingDigest("account:b")
	queueA := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digestA)
	queueB := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digestB)

	mr.Lpush(queueA, "one")
	mr.Lpush(queueA, "two")
	mr.Lpush(queueB, "three")

	stats, err := Stats(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if stats.OrderedPending != 3 {
		t.Fatalf("ordered pending count = %d, want 3", stats.OrderedPending)
	}

	keys, err := StatsKeys(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if keys.OrderedPending != 3 {
		t.Fatalf("ordered key pending count = %d, want 3", keys.OrderedPending)
	}
	if keys.QueueLengths["ordered("+queueA+")"] != 2 ||
		keys.QueueLengths["ordered("+queueB+")"] != 1 {
		t.Fatalf("ordered queue lengths missing: %#v", keys.QueueLengths)
	}
}
