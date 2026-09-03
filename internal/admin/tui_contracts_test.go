// Copyright 2026 James Ross
package admin

import (
	"context"
	"slices"
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

	items, _, err := DLQList(context.Background(), cfg, rdb, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Handle == "" {
		t.Fatalf("listed items = %#v, want one item with a selection handle", items)
	}

	requeued, err := DLQRequeue(context.Background(), cfg, rdb, "", []string{items[0].Handle}, "")
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

func TestDLQSelectionHandlesDistinguishDuplicateJobIDs(t *testing.T) {
	for _, action := range []string{"requeue", "purge"} {
		t.Run(action, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			cfg := config.Default()
			ctx := context.Background()

			first := queue.NewJob("shared-id", "", 0, "low", "", "")
			first.Payload = []byte("first delivery")
			second := queue.NewJob("shared-id", "", 0, "low", "", "")
			second.Payload = []byte("second delivery")
			firstRaw, err := first.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			secondRaw, err := second.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, firstRaw, secondRaw).Err(); err != nil {
				t.Fatal(err)
			}

			items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
			if err != nil {
				t.Fatal(err)
			}
			if len(items) != 2 {
				t.Fatalf("listed %d items, want 2", len(items))
			}
			if items[0].ID != "shared-id" || items[1].ID != "shared-id" {
				t.Fatalf("listed IDs = %q, %q, want shared-id", items[0].ID, items[1].ID)
			}
			if items[0].Handle == "" || items[1].Handle == "" || items[0].Handle == items[1].Handle {
				t.Fatalf("selection handles = %q, %q, want distinct non-empty handles", items[0].Handle, items[1].Handle)
			}

			switch action {
			case "requeue":
				moved, err := DLQRequeue(ctx, cfg, rdb, "", []string{items[1].Handle}, "retry:queue")
				if err != nil {
					t.Fatal(err)
				}
				if moved != 1 {
					t.Fatalf("requeued = %d, want 1", moved)
				}
				if got := rdb.LRange(ctx, "retry:queue", 0, -1).Val(); !slices.Equal(got, []string{secondRaw}) {
					t.Fatalf("retry queue = %#v, want selected second delivery", got)
				}
			case "purge":
				purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[1].Handle})
				if err != nil {
					t.Fatal(err)
				}
				if purged != 1 {
					t.Fatalf("purged = %d, want 1", purged)
				}
			}

			if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, []string{firstRaw}) {
				t.Fatalf("dead-letter queue = %#v, want unselected first delivery", got)
			}
		})
	}
}

func TestDLQSelectionHandlesDistinguishIdenticalEnvelopes(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()

	job := queue.NewJob("shared-id", "", 0, "low", "", "")
	raw, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, raw, raw).Err(); err != nil {
		t.Fatal(err)
	}

	items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Handle == "" || items[1].Handle == "" || items[0].Handle == items[1].Handle {
		t.Fatalf("listed items = %#v, want distinct handles for identical envelopes", items)
	}

	purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[1].Handle})
	if err != nil {
		t.Fatal(err)
	}
	if purged != 1 || rdb.LLen(ctx, cfg.Worker.DeadLetterList).Val() != 1 {
		t.Fatalf("purged=%d remaining=%d, want 1 and 1", purged, rdb.LLen(ctx, cfg.Worker.DeadLetterList).Val())
	}
}

func TestDLQBulkSelectionRemovesDescendingSnapshotIndices(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()
	raw := []string{
		`{"id":"first"}`,
		`{"id":"middle"}`,
		`{"id":"last"}`,
	}
	if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, raw[0], raw[1], raw[2]).Err(); err != nil {
		t.Fatal(err)
	}

	items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[0].Handle, items[2].Handle})
	if err != nil {
		t.Fatal(err)
	}
	if purged != 2 {
		t.Fatalf("purged = %d, want 2", purged)
	}
	if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, raw[1:2]) {
		t.Fatalf("dead-letter queue = %#v, want only middle entry", got)
	}
}

func TestDLQStaleSelectionHandleDoesNotTargetShiftedEntry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()
	if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, `{"id":"selected"}`).Err(); err != nil {
		t.Fatal(err)
	}
	items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, cfg.Worker.DeadLetterList, `{"id":"new"}`).Err(); err != nil {
		t.Fatal(err)
	}

	purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[0].Handle})
	if err != nil {
		t.Fatal(err)
	}
	if purged != 0 {
		t.Fatalf("purged = %d, want stale selection to be a no-op", purged)
	}
	if got := rdb.LLen(ctx, cfg.Worker.DeadLetterList).Val(); got != 2 {
		t.Fatalf("dead-letter length = %d, want both entries preserved", got)
	}
}

func TestDLQRequeueWrongTypePreservesSelectedEntry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()
	raw := `{"id":"selected"}`
	if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, raw).Err(); err != nil {
		t.Fatal(err)
	}
	items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(ctx, "wrong-type-destination", "not a list", 0).Err(); err != nil {
		t.Fatal(err)
	}

	if moved, err := DLQRequeue(ctx, cfg, rdb, "", []string{items[0].Handle}, "wrong-type-destination"); err == nil || moved != 0 {
		t.Fatalf("requeue = (%d, %v), want (0, wrong-type error)", moved, err)
	}
	if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, []string{raw}) {
		t.Fatalf("dead-letter queue after failed requeue = %#v, want selected entry preserved", got)
	}
}
