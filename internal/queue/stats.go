// Copyright 2026 James Ross
package queue

import (
	"context"
	"fmt"

	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

// OrderedQueueLengths returns each per-key FIFO backlog and their aggregate.
// Redis SCAN may return a key more than once, so duplicate observations are
// ignored within a snapshot.
func OrderedQueueLengths(ctx context.Context, rdb redis.Cmdable, pattern string) (map[string]int64, int64, error) {
	if _, _, ok := queuekeys.SplitPattern(pattern); !ok {
		return nil, 0, fmt.Errorf("ordered queue pattern must contain exactly one %%s placeholder")
	}

	lengths := map[string]int64{}
	seen := map[string]struct{}{}
	var total int64
	var cursor uint64
	for {
		keys, next, err := rdb.Scan(ctx, cursor, queuekeys.ScanPattern(pattern), 200).Result()
		if err != nil {
			return nil, 0, fmt.Errorf("scan ordered queues: %w", err)
		}
		for _, key := range keys {
			identifier, matches := queuekeys.Extract(pattern, key)
			if !matches || !isOrderingDigest(identifier) {
				continue
			}
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			length, err := rdb.LLen(ctx, key).Result()
			if err != nil {
				return nil, 0, fmt.Errorf("read ordered queue %q: %w", key, err)
			}
			lengths[key] = length
			total += length
		}
		cursor = next
		if cursor == 0 {
			return lengths, total, nil
		}
	}
}

func isOrderingDigest(identifier string) bool {
	if len(identifier) != 64 {
		return false
	}
	for _, char := range identifier {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
