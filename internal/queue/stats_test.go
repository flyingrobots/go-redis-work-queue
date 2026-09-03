// Copyright 2026 James Ross
package queue

import (
	"context"
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
	literalKey := queuekeys.Format(pattern, "literal")
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
