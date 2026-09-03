// Copyright 2026 James Ross
package admin

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

func TestDLQRequeuePreservesOrderedIntake(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()

	job := queue.NewJob("ordered-dlq", "", 0, "low", "", "")
	job.OrderingKey = "account:42"
	raw, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(context.Background(), cfg.Worker.DeadLetterList, raw).Err(); err != nil {
		t.Fatal(err)
	}

	requeued, err := DLQRequeue(context.Background(), cfg, rdb, "", []string{job.ID}, "")
	if err != nil {
		t.Fatal(err)
	}
	if requeued != 1 {
		t.Fatalf("requeued = %d, want 1", requeued)
	}
	for alias, key := range cfg.Worker.Queues {
		if length := rdb.LLen(context.Background(), key).Val(); length != 0 {
			t.Fatalf("ordered job leaked into ordinary %s queue %q: length=%d", alias, key, length)
		}
	}

	digest := queuekeys.OrderingDigest(job.OrderingKey)
	orderedQueue := queuekeys.Format(cfg.Queue.OrderedQueuePattern, digest)
	if got := rdb.LRange(context.Background(), orderedQueue, 0, -1).Val(); len(got) != 1 || got[0] != raw {
		t.Fatalf("ordered queue = %#v, want original envelope", got)
	}
	if !rdb.SIsMember(context.Background(), cfg.Queue.OrderedActiveSet, digest).Val() {
		t.Fatal("ordered digest is missing from active set")
	}
	if got := rdb.LRange(context.Background(), cfg.Queue.OrderedReadyList, 0, -1).Val(); len(got) != 1 || got[0] != digest {
		t.Fatalf("ready ring = %#v, want digest", got)
	}
	if length := rdb.LLen(context.Background(), cfg.Worker.DeadLetterList).Val(); length != 0 {
		t.Fatalf("dead-letter length = %d, want 0", length)
	}
}
