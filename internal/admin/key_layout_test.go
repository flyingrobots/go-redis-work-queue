// Copyright 2026 James Ross
package admin

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
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
	if deleted != 3 || len(mr.Keys()) != 0 {
		t.Fatalf("purge deleted %d keys; remaining=%v", deleted, mr.Keys())
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

	for _, key := range []string{
		cfg.Queue.OrderedReadyList,
		cfg.Queue.OrderedActiveSet,
		"jobqueue:ordered:queue:digest-a",
		"jobqueue:ordered:lease:digest-a",
	} {
		mr.Set(key, "value")
	}
	deleted, err := PurgeAll(context.Background(), cfg, rdb)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 4 || len(mr.Keys()) != 0 {
		t.Fatalf("purge deleted %d ordered keys; remaining=%v", deleted, mr.Keys())
	}
}
