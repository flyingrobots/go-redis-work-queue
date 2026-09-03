// Copyright 2026 James Ross
package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

// OrderingLayout names the Redis structures used by non-empty ordering keys.
// QueuePattern and LeasePattern contain one %s placeholder for the SHA-256
// digest of the exact ordering key.
type OrderingLayout struct {
	ReadyList    string
	ActiveSet    string
	QueuePattern string
	LeasePattern string
}

// DefaultOrderingLayout returns the standard ordered queue key layout.
func DefaultOrderingLayout() OrderingLayout {
	return OrderingLayout{
		ReadyList:    queuekeys.DefaultOrderedReadyList,
		ActiveSet:    queuekeys.DefaultOrderedActiveSet,
		QueuePattern: queuekeys.DefaultOrderedQueuePattern,
		LeasePattern: queuekeys.DefaultOrderedLeasePattern,
	}
}

// Validate checks that the layout can safely format claimed digests.
func (l OrderingLayout) Validate() error {
	if l.ReadyList == "" {
		return errors.New("ordered ready list must be non-empty")
	}
	if l.ActiveSet == "" {
		return errors.New("ordered active set must be non-empty")
	}
	if _, _, ok := queuekeys.SplitPattern(l.QueuePattern); !ok {
		return errors.New("ordered queue pattern must contain exactly one %s placeholder")
	}
	if _, _, ok := queuekeys.SplitPattern(l.LeasePattern); !ok {
		return errors.New("ordered lease pattern must contain exactly one %s placeholder")
	}
	if l.QueuePattern == l.LeasePattern {
		return errors.New("ordered queue and lease patterns must differ")
	}
	return nil
}

// OrderedDelivery is the ownership record returned by an atomic key claim.
// Payload is already present in the worker's processing list.
type OrderedDelivery struct {
	Digest   string
	Payload  string
	QueueKey string
	LeaseKey string
}

type OrderedTransition string

const (
	OrderedComplete   OrderedTransition = "complete"
	OrderedRetry      OrderedTransition = "retry"
	OrderedDeadLetter OrderedTransition = "dead_letter"
	OrderedDiscard    OrderedTransition = "discard"
)

var enqueueOrderedScript = redis.NewScript(`
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[2], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[3], 'set')
if problem then return redis.error_reply(problem) end

redis.call('LPUSH', KEYS[1], ARGV[1])
if redis.call('SADD', KEYS[3], ARGV[2]) == 1 then
  redis.call('LPUSH', KEYS[2], ARGV[2])
end
return 1
`)

var requeueEncodedScript = redis.NewScript(`
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[2], 'list')
if problem then return redis.error_reply(problem) end
if ARGV[3] == 'ordered' then
  problem = require_type(KEYS[3], 'list')
  if problem then return redis.error_reply(problem) end
  problem = require_type(KEYS[4], 'set')
  if problem then return redis.error_reply(problem) end
end

if redis.call('LREM', KEYS[1], 1, ARGV[1]) ~= 1 then
  return 0
end
redis.call('LPUSH', KEYS[2], ARGV[1])
if ARGV[3] == 'ordered' and redis.call('SADD', KEYS[4], ARGV[2]) == 1 then
  redis.call('LPUSH', KEYS[3], ARGV[2])
end
return 1
`)

var claimOrderedScript = redis.NewScript(`
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
local digest = redis.call('LINDEX', KEYS[1], -1)
if not digest then
  return nil
end

local queue_key = ARGV[1] .. digest .. ARGV[2]
local lease_key = ARGV[3] .. digest .. ARGV[4]
local claim_keys = {KEYS[1], KEYS[2], KEYS[3], KEYS[4], queue_key, lease_key}
for i = 1, #claim_keys do
  for j = i + 1, #claim_keys do
    if claim_keys[i] == claim_keys[j] then
      return redis.error_reply('ordered claim keys must differ')
    end
  end
end

problem = require_type(KEYS[2], 'set')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[3], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[4], 'string')
if problem then return redis.error_reply(problem) end
problem = require_type(queue_key, 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(lease_key, 'string')
if problem then return redis.error_reply(problem) end

local acquired = redis.call('SET', lease_key, ARGV[5], 'NX', 'PX', ARGV[6])
if not acquired then
  return nil
end
redis.call('RPOP', KEYS[1])

local payload = redis.call('RPOP', queue_key)
if not payload then
  redis.call('DEL', lease_key)
  redis.call('SREM', KEYS[2], digest)
  return nil
end

redis.call('LPUSH', KEYS[3], payload)
redis.call('PSETEX', KEYS[4], ARGV[7], payload)
return {digest, payload}
`)

var renewOrderedLeaseScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return 1
end
return 0
`)

var transitionOrderedScript = redis.NewScript(`
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local transition = ARGV[5]
if transition ~= 'retry' and transition ~= 'complete' and transition ~= 'dead_letter' and transition ~= 'discard' then
  return redis.error_reply('unknown ordered transition')
end

local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[3], 'string')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[4], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[5], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[6], 'set')
if problem then return redis.error_reply(problem) end
if transition == 'complete' or transition == 'dead_letter' then
  problem = require_type(KEYS[7], 'list')
  if problem then return redis.error_reply(problem) end
end

if redis.call('GET', KEYS[3]) ~= ARGV[1] then
  return 0
end
if redis.call('LREM', KEYS[1], 1, ARGV[2]) ~= 1 then
  return 0
end

if transition == 'retry' then
  redis.call('RPUSH', KEYS[4], ARGV[3])
elseif transition == 'complete' or transition == 'dead_letter' then
  redis.call('LPUSH', KEYS[7], ARGV[3])
end

redis.call('DEL', KEYS[2])
redis.call('DEL', KEYS[3])
if redis.call('LLEN', KEYS[4]) > 0 then
  redis.call('SADD', KEYS[6], ARGV[4])
  redis.call('LPUSH', KEYS[5], ARGV[4])
else
  redis.call('SREM', KEYS[6], ARGV[4])
  redis.call('DEL', KEYS[4])
end
return 1
`)

var recoverOrderedScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[2]) == 1 then
  return 0
end
if redis.call('EXISTS', KEYS[3]) == 1 then
  return 0
end

local recovery_keys = {KEYS[1], KEYS[2], KEYS[3], KEYS[4], KEYS[5], KEYS[6]}
for i = 1, #recovery_keys do
  for j = i + 1, #recovery_keys do
    if recovery_keys[i] == recovery_keys[j] then
      return redis.error_reply('ordered recovery keys must differ')
    end
  end
end

local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[4], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[5], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[6], 'set')
if problem then return redis.error_reply(problem) end

if redis.call('LREM', KEYS[1], 1, ARGV[1]) ~= 1 then
  return 0
end

redis.call('RPUSH', KEYS[4], ARGV[1])
redis.call('SADD', KEYS[6], ARGV[2])
redis.call('LPUSH', KEYS[5], ARGV[2])
return 1
`)

// AppendEncoded writes a validated envelope to either the unchanged priority
// list path or the atomic per-key FIFO path.
func AppendEncoded(ctx context.Context, rdb redis.Cmdable, queueName string, job Job, encoded string, layout OrderingLayout) error {
	if job.OrderingKey == "" {
		if err := rdb.LPush(ctx, queueName, encoded).Err(); err != nil {
			return fmt.Errorf("enqueue job: %w", err)
		}
		return nil
	}
	if err := layout.Validate(); err != nil {
		return err
	}
	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey := queuekeys.Format(layout.QueuePattern, digest)
	if err := enqueueOrderedScript.Eval(ctx, rdb,
		[]string{queueKey, layout.ReadyList, layout.ActiveSet}, encoded, digest).Err(); err != nil {
		return fmt.Errorf("enqueue ordered job: %w", err)
	}
	return nil
}

// RequeueEncoded atomically removes one exact envelope from a source list and
// returns it to ordinary or ordered intake according to its ordering key.
func RequeueEncoded(ctx context.Context, rdb redis.Cmdable, source, destination string, job Job, encoded string, layout OrderingLayout) (bool, error) {
	if source == "" {
		return false, errors.New("requeue source must be non-empty")
	}
	if job.OrderingKey == "" {
		if destination == "" {
			return false, errors.New("requeue destination must be non-empty")
		}
		result, err := requeueEncodedScript.Eval(ctx, rdb,
			[]string{source, destination, destination, destination}, encoded, "", "ordinary").Int64()
		if err != nil {
			return false, fmt.Errorf("requeue job: %w", err)
		}
		return result == 1, nil
	}

	if err := layout.Validate(); err != nil {
		return false, err
	}
	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey := queuekeys.Format(layout.QueuePattern, digest)
	result, err := requeueEncodedScript.Eval(ctx, rdb,
		[]string{source, queueKey, layout.ReadyList, layout.ActiveSet}, encoded, digest, "ordered").Int64()
	if err != nil {
		return false, fmt.Errorf("requeue ordered job: %w", err)
	}
	return result == 1, nil
}

// ClaimOrdered atomically leases the oldest ready key and moves its oldest job
// into the existing per-worker processing list. An empty ready ring returns
// ok=false without error.
func ClaimOrdered(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, processingList, heartbeatKey, owner string, ttl time.Duration) (delivery OrderedDelivery, ok bool, err error) {
	if err := layout.Validate(); err != nil {
		return delivery, false, err
	}
	queuePrefix, queueSuffix, _ := queuekeys.SplitPattern(layout.QueuePattern)
	leasePrefix, leaseSuffix, _ := queuekeys.SplitPattern(layout.LeasePattern)
	ttlMillis := durationMillis(ttl)
	result, err := claimOrderedScript.Eval(ctx, rdb,
		[]string{layout.ReadyList, layout.ActiveSet, processingList, heartbeatKey},
		queuePrefix, queueSuffix, leasePrefix, leaseSuffix, owner, ttlMillis, ttlMillis,
	).Result()
	if err == redis.Nil {
		return delivery, false, nil
	}
	if err != nil {
		return delivery, false, fmt.Errorf("claim ordered job: %w", err)
	}
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return delivery, false, fmt.Errorf("claim ordered job: unexpected Redis response %T", result)
	}
	digest, ok := values[0].(string)
	if !ok || digest == "" {
		return delivery, false, fmt.Errorf("claim ordered job: invalid digest response")
	}
	payload, ok := values[1].(string)
	if !ok || payload == "" {
		return delivery, false, fmt.Errorf("claim ordered job: invalid payload response")
	}
	return OrderedDelivery{
		Digest:   digest,
		Payload:  payload,
		QueueKey: queuekeys.Format(layout.QueuePattern, digest),
		LeaseKey: queuekeys.Format(layout.LeasePattern, digest),
	}, true, nil
}

// RenewOrderedLease extends a lease only for its current owner.
func RenewOrderedLease(ctx context.Context, rdb redis.Cmdable, leaseKey, owner string, ttl time.Duration) (bool, error) {
	result, err := renewOrderedLeaseScript.Eval(ctx, rdb, []string{leaseKey}, owner, durationMillis(ttl)).Int64()
	if err != nil {
		return false, fmt.Errorf("renew ordered lease: %w", err)
	}
	return result == 1, nil
}

// TransitionOrdered acknowledges an ordered processing envelope only when the
// caller still owns its lease. RetryPayload is the envelope written to the
// destination or returned to the head of the per-key FIFO.
func TransitionOrdered(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, delivery OrderedDelivery, processingList, heartbeatKey, owner, destination, retryPayload string, transition OrderedTransition) (bool, error) {
	if err := layout.Validate(); err != nil {
		return false, err
	}
	if transition != OrderedComplete && transition != OrderedRetry && transition != OrderedDeadLetter && transition != OrderedDiscard {
		return false, fmt.Errorf("unknown ordered transition %q", transition)
	}
	if (transition == OrderedComplete || transition == OrderedDeadLetter) && destination == "" {
		return false, fmt.Errorf("ordered %s destination must be non-empty", transition)
	}
	if destination == "" {
		// Lua KEYS entries cannot be empty. Discard and retry never access this
		// key, so a stable existing control key is a safe placeholder.
		destination = layout.ActiveSet
	}
	result, err := transitionOrderedScript.Eval(ctx, rdb, []string{
		processingList,
		heartbeatKey,
		delivery.LeaseKey,
		delivery.QueueKey,
		layout.ReadyList,
		layout.ActiveSet,
		destination,
	}, owner, delivery.Payload, retryPayload, delivery.Digest, string(transition)).Int64()
	if err != nil {
		return false, fmt.Errorf("transition ordered job: %w", err)
	}
	return result == 1, nil
}

// RecoverOrdered returns one abandoned processing envelope to the front of its
// per-key FIFO after both the worker heartbeat and lease have expired. LREM in
// the script makes concurrent reapers and late completions single-winner.
func RecoverOrdered(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, delivery OrderedDelivery, processingList, heartbeatKey string) (bool, error) {
	if err := layout.Validate(); err != nil {
		return false, err
	}
	result, err := recoverOrderedScript.Eval(ctx, rdb, []string{
		processingList,
		heartbeatKey,
		delivery.LeaseKey,
		delivery.QueueKey,
		layout.ReadyList,
		layout.ActiveSet,
	}, delivery.Payload, delivery.Digest).Int64()
	if err != nil {
		return false, fmt.Errorf("recover ordered job: %w", err)
	}
	return result == 1, nil
}

// OrderedDeliveryFor reconstructs recovery metadata from the durable job.
func OrderedDeliveryFor(layout OrderingLayout, job Job, payload string) OrderedDelivery {
	digest := queuekeys.OrderingDigest(job.OrderingKey)
	return OrderedDelivery{
		Digest:   digest,
		Payload:  payload,
		QueueKey: queuekeys.Format(layout.QueuePattern, digest),
		LeaseKey: queuekeys.Format(layout.LeasePattern, digest),
	}
}

func durationMillis(ttl time.Duration) int64 {
	millis := ttl.Milliseconds()
	if millis < 1 {
		return 1
	}
	return millis
}
