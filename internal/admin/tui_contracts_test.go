// Copyright 2026 James Ross
package admin

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

var errUnboundedDLQSnapshot = errors.New("DLQ operation attempted an unbounded whole-list snapshot")

type rejectUnboundedDLQSnapshotHook struct{}

func (*rejectUnboundedDLQSnapshotHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (*rejectUnboundedDLQSnapshotHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		for _, arg := range cmd.Args() {
			script, ok := arg.(string)
			if ok && strings.Contains(script, "redis.call('LRANGE', key, 0, -1)") {
				return errUnboundedDLQSnapshot
			}
		}
		return next(ctx, cmd)
	}
}

func (*rejectUnboundedDLQSnapshotHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestDLQOperationsUseBoundedSnapshotScripts(t *testing.T) {
	for _, action := range []string{"list", "purge", "requeue"} {
		t.Run(action, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			cfg := config.Default()
			job := queue.NewJob("bounded-dlq", "", 0, "low", "", "")
			raw, err := job.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := rdb.RPush(context.Background(), cfg.Worker.DeadLetterList, raw).Err(); err != nil {
				t.Fatal(err)
			}

			var handles []string
			if action != "list" {
				items, _, err := DLQList(context.Background(), cfg, rdb, "", "", 1)
				if err != nil {
					t.Fatal(err)
				}
				handles = []string{items[0].Handle}
			}
			rdb.AddHook(&rejectUnboundedDLQSnapshotHook{})

			switch action {
			case "list":
				_, _, err = DLQList(context.Background(), cfg, rdb, "", "", 1)
			case "purge":
				_, err = DLQPurge(context.Background(), cfg, rdb, "", handles)
			case "requeue":
				_, err = DLQRequeue(context.Background(), cfg, rdb, "", handles, "bounded:retry")
			}
			if err != nil {
				t.Fatal(err)
			}
			if action != "list" {
				generation, err := rdb.Get(context.Background(), queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)).Result()
				if err != nil || generation != "1" {
					t.Fatalf("%s generation = %q (err=%v), want 1", action, generation, err)
				}
			}
		})
	}
}

func TestDLQGenerationInvalidatesSameLengthIdenticalReplacement(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cfg := config.Default()
	ctx := context.Background()
	job := queue.NewJob("identical-aba", "", 0, "low", "", "")
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
	if purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[1].Handle}); err != nil || purged != 1 {
		t.Fatalf("first purge = %d (err=%v), want 1", purged, err)
	}
	if err := queue.AppendDeadLetter(ctx, rdb, cfg.Worker.DeadLetterList, raw); err != nil {
		t.Fatal(err)
	}
	if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, []string{raw, raw}) {
		t.Fatalf("reconstructed DLQ = %#v, want byte-identical two-entry state", got)
	}

	if purged, err := DLQPurge(ctx, cfg, rdb, "", []string{items[0].Handle}); err != nil || purged != 0 {
		t.Fatalf("stale pre-replacement purge = %d (err=%v), want 0", purged, err)
	}
	if generation := rdb.Get(ctx, queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)).Val(); generation != "2" {
		t.Fatalf("DLQ generation = %q, want 2", generation)
	}
}

func TestDLQListRejectsInvalidGeneration(t *testing.T) {
	for _, generation := range []string{"01", "not-a-number", "9223372036854775808"} {
		t.Run(generation, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			cfg := config.Default()
			ctx := context.Background()
			if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, `{"id":"preserved"}`).Err(); err != nil {
				t.Fatal(err)
			}
			if err := rdb.Set(ctx, queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList), generation, 0).Err(); err != nil {
				t.Fatal(err)
			}

			if _, _, err := DLQList(ctx, cfg, rdb, "", "", 10); err == nil {
				t.Fatalf("DLQList accepted invalid generation %q", generation)
			}
			if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, []string{`{"id":"preserved"}`}) {
				t.Fatalf("dead-letter queue after rejected generation = %#v", got)
			}
		})
	}
}

func TestDLQMutationsFailClosedAtExhaustedGeneration(t *testing.T) {
	const exhausted = "9223372036854775807"
	for _, action := range []string{"selected purge", "selected requeue", "append", "full purge"} {
		t.Run(action, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			cfg := config.Default()
			ctx := context.Background()
			raw := `{"id":"preserved"}`
			if err := rdb.RPush(ctx, cfg.Worker.DeadLetterList, raw).Err(); err != nil {
				t.Fatal(err)
			}
			generationKey := queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList)
			if err := rdb.Set(ctx, generationKey, exhausted, 0).Err(); err != nil {
				t.Fatal(err)
			}
			items, _, err := DLQList(ctx, cfg, rdb, "", "", 10)
			if err != nil {
				t.Fatal(err)
			}

			switch action {
			case "selected purge":
				_, err = DLQPurge(ctx, cfg, rdb, "", []string{items[0].Handle})
			case "selected requeue":
				_, err = DLQRequeue(ctx, cfg, rdb, "", []string{items[0].Handle}, "retry:queue")
			case "append":
				err = queue.AppendDeadLetter(ctx, rdb, cfg.Worker.DeadLetterList, `{"id":"new"}`)
			case "full purge":
				err = PurgeDLQ(ctx, cfg, rdb)
			}
			if err == nil || !strings.Contains(err.Error(), "generation exhausted") {
				t.Fatalf("%s error = %v, want generation exhausted", action, err)
			}
			if got := rdb.LRange(ctx, cfg.Worker.DeadLetterList, 0, -1).Val(); !slices.Equal(got, []string{raw}) {
				t.Fatalf("dead-letter queue after %s = %#v, want preserved entry", action, got)
			}
			if got := rdb.Get(ctx, generationKey).Val(); got != exhausted {
				t.Fatalf("generation after %s = %q, want %s", action, got, exhausted)
			}
			if got := rdb.LLen(ctx, "retry:queue").Val(); got != 0 {
				t.Fatalf("retry queue after %s has %d entries, want 0", action, got)
			}
		})
	}
}

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

func TestDLQStaleSelectionHandleDoesNotRetargetIdenticalEntryAfterLPush(t *testing.T) {
	for _, action := range []string{"requeue", "purge"} {
		t.Run(action, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			cfg := config.Default()
			ctx := context.Background()

			job := queue.NewJob("byte-identical", "", 0, "low", "", "")
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
			if err := rdb.LPush(ctx, cfg.Worker.DeadLetterList, raw).Err(); err != nil {
				t.Fatal(err)
			}

			var changed int
			switch action {
			case "requeue":
				changed, err = DLQRequeue(ctx, cfg, rdb, "", []string{items[1].Handle}, "retry:queue")
			case "purge":
				changed, err = DLQPurge(ctx, cfg, rdb, "", []string{items[1].Handle})
			}
			if err != nil {
				t.Fatal(err)
			}
			if changed != 0 {
				t.Fatalf("%s changed %d entries, want stale selection to be a no-op", action, changed)
			}
			if got := rdb.LLen(ctx, cfg.Worker.DeadLetterList).Val(); got != 3 {
				t.Fatalf("dead-letter length = %d, want all three identical entries preserved", got)
			}
			if got := rdb.LLen(ctx, "retry:queue").Val(); got != 0 {
				t.Fatalf("retry queue length = %d, want 0", got)
			}
		})
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
