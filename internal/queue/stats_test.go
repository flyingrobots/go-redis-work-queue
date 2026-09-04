// Copyright 2026 James Ross
package queue

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

func TestOrderedQueueLengthsEscapesLiteralRedisGlobCharacters(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()

	pattern := `tenant[1]:ordered:%s`
	literalKey := queuekeys.Format(pattern, queuekeys.OrderingDigest("literal"))
	if err := rdb.LPush(ctx, literalKey, "one", "two").Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, "tenant1:ordered:lookalike", "wrong").Err(); err != nil {
		t.Fatal(err)
	}

	lengths, total, err := OrderedQueueLengths(ctx, rdb, pattern)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || lengths[literalKey] != 2 || len(lengths) != 1 {
		t.Fatalf("ordered lengths = %#v, total %d; want only literal key", lengths, total)
	}
}

func TestOrderedQueueLengthsIgnoresBroadPatternLookalikes(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()

	pattern := `custom:%s`
	digest := queuekeys.OrderingDigest("account:42")
	queueKey := queuekeys.Format(pattern, digest)
	if err := rdb.LPush(ctx, queueKey, "one", "two").Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(ctx, "custom:lease:"+digest, "worker-a", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.LPush(ctx, "custom:ready", digest).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.SAdd(ctx, "custom:active", digest).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(ctx, "custom:"+strings.Repeat("g", 64), "not-a-digest", 0).Err(); err != nil {
		t.Fatal(err)
	}

	lengths, total, err := OrderedQueueLengths(ctx, rdb, pattern)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || lengths[queueKey] != 2 || len(lengths) != 1 {
		t.Fatalf("ordered lengths = %#v, total %d; want only digest queue", lengths, total)
	}
}
