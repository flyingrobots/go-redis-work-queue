// Copyright 2026 James Ross
package queue

import (
	"context"
	"fmt"

	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

// PreparedEnqueue is a locally validated job ready for a batch append.
type PreparedEnqueue struct {
	QueueName string
	Job       Job
	Encoded   string
}

var enqueueBatchScript = redis.NewScript(`
local key_count = tonumber(ARGV[1])
for i = 1, key_count do
  local expected = ARGV[i + 1]
  local type_reply = redis.call('TYPE', KEYS[i])
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return redis.error_reply('WRONGTYPE key ' .. KEYS[i] .. ' has type ' .. actual .. ', expected ' .. expected)
  end
end

local position = key_count + 2
local item_count = tonumber(ARGV[position])
position = position + 1
for i = 1, item_count do
  local kind = ARGV[position]
  position = position + 1
  if kind == 'ordinary' then
    local queue_index = tonumber(ARGV[position])
    local payload = ARGV[position + 1]
    position = position + 2
    redis.call('LPUSH', KEYS[queue_index], payload)
  else
    local queue_index = tonumber(ARGV[position])
    local ready_index = tonumber(ARGV[position + 1])
    local active_index = tonumber(ARGV[position + 2])
    local payload = ARGV[position + 3]
    local digest = ARGV[position + 4]
    position = position + 5
    redis.call('LPUSH', KEYS[queue_index], payload)
    if redis.call('SADD', KEYS[active_index], digest) == 1 then
      redis.call('LPUSH', KEYS[ready_index], digest)
    end
  end
end
return item_count
`)

// AppendEncodedBatch prevalidates every destination key's Redis type, then
// appends the accepted jobs in one atomic script. This prevents a wrong-type
// error for a later item from leaving earlier items partially enqueued.
func AppendEncodedBatch(ctx context.Context, rdb redis.Cmdable, items []PreparedEnqueue, layout OrderingLayout) error {
	if len(items) == 0 {
		return nil
	}

	type operation struct {
		kind        string
		queueIndex  int
		readyIndex  int
		activeIndex int
		payload     string
		digest      string
	}

	keys := make([]string, 0, len(items)+2)
	expectedTypes := make([]string, 0, len(items)+2)
	keyIndexes := make(map[string]int, len(items)+2)
	addKey := func(key, expectedType string) (int, error) {
		if index, ok := keyIndexes[key]; ok {
			if expectedTypes[index-1] != expectedType {
				return 0, fmt.Errorf("Redis key %q is required as both %s and %s", key, expectedTypes[index-1], expectedType)
			}
			return index, nil
		}
		keys = append(keys, key)
		expectedTypes = append(expectedTypes, expectedType)
		index := len(keys)
		keyIndexes[key] = index
		return index, nil
	}

	hasOrdered := false
	for _, item := range items {
		if item.Job.OrderingKey != "" {
			hasOrdered = true
			break
		}
	}
	if hasOrdered {
		if err := layout.Validate(); err != nil {
			return err
		}
	}

	operations := make([]operation, 0, len(items))
	for _, item := range items {
		if item.Job.OrderingKey == "" {
			queueIndex, err := addKey(item.QueueName, "list")
			if err != nil {
				return err
			}
			operations = append(operations, operation{kind: "ordinary", queueIndex: queueIndex, payload: item.Encoded})
			continue
		}

		digest := queuekeys.OrderingDigest(item.Job.OrderingKey)
		queueIndex, err := addKey(queuekeys.Format(layout.QueuePattern, digest), "list")
		if err != nil {
			return err
		}
		readyIndex, err := addKey(layout.ReadyList, "list")
		if err != nil {
			return err
		}
		activeIndex, err := addKey(layout.ActiveSet, "set")
		if err != nil {
			return err
		}
		operations = append(operations, operation{
			kind:        "ordered",
			queueIndex:  queueIndex,
			readyIndex:  readyIndex,
			activeIndex: activeIndex,
			payload:     item.Encoded,
			digest:      digest,
		})
	}

	args := make([]interface{}, 0, 2+len(expectedTypes)+len(operations)*6)
	args = append(args, len(keys))
	for _, expectedType := range expectedTypes {
		args = append(args, expectedType)
	}
	args = append(args, len(operations))
	for _, operation := range operations {
		args = append(args, operation.kind, operation.queueIndex)
		if operation.kind == "ordinary" {
			args = append(args, operation.payload)
			continue
		}
		args = append(args, operation.readyIndex, operation.activeIndex, operation.payload, operation.digest)
	}

	if err := enqueueBatchScript.Eval(ctx, rdb, keys, args...).Err(); err != nil {
		return fmt.Errorf("enqueue batch: %w", err)
	}
	return nil
}
