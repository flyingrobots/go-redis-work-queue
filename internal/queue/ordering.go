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
	if l.ReadyList == l.ActiveSet {
		return errors.New("ordered ready list and active set must differ")
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

// keysForDigest resolves every ordered role that is specific to one ordering
// key and rejects collisions with the shared controls before Redis is touched.
func (l OrderingLayout) keysForDigest(digest string) (queueKey, leaseKey string, err error) {
	if err := l.Validate(); err != nil {
		return "", "", err
	}
	queueKey = queuekeys.Format(l.QueuePattern, digest)
	leaseKey = queuekeys.Format(l.LeasePattern, digest)
	roles := []struct {
		name string
		key  string
	}{
		{name: "ready list", key: l.ReadyList},
		{name: "active set", key: l.ActiveSet},
		{name: "claim registry", key: l.claimsKey()},
		{name: "per-key queue", key: queueKey},
		{name: "per-key lease", key: leaseKey},
	}
	seen := make(map[string]string, len(roles))
	for _, role := range roles {
		if previous, ok := seen[role.key]; ok {
			return "", "", fmt.Errorf("ordered %s must differ from %s (%q)", role.name, previous, role.key)
		}
		seen[role.key] = role.name
	}
	return queueKey, leaseKey, nil
}

func (l OrderingLayout) claimsKey() string {
	return queuekeys.OrderedClaimsKey(l.ActiveSet)
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

const listVersionLua = `
local function generation_value(key)
  local generation = redis.call('GET', key)
  if not generation then return '0' end
  if not string.match(generation, '^%d+$') then return nil end
  if string.len(generation) > 1 and string.sub(generation, 1, 1) == '0' then return nil end
  if string.len(generation) > 19 or
      (string.len(generation) == 19 and generation > '9223372036854775807') then
    return nil
  end
  return generation
end

local function generation_has_room(generation)
  if string.len(generation) < 19 then return true end
  if string.len(generation) > 19 then return false end
  return generation < '9223372036854775807'
end

local function list_version(list_key, generation_key)
  local generation = generation_value(generation_key)
  if not generation then return nil end
  return generation .. ':' .. tostring(redis.call('LLEN', list_key))
end
`

var requeueEncodedAtScript = redis.NewScript(listVersionLua + `
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
for i = 1, 4 do
  if KEYS[5] == KEYS[i] then
    return redis.error_reply('dead-letter generation key must differ from queue state keys')
  end
end
problem = require_type(KEYS[5], 'string')
if problem then return redis.error_reply(problem) end

local index = tonumber(ARGV[4])
if not index or index < 0 then
  return redis.error_reply('selection index must be non-negative')
end
local current_version = list_version(KEYS[1], KEYS[5])
if not current_version then
  return redis.error_reply('invalid dead-letter generation')
end
if current_version ~= ARGV[6] then
  return {'0', current_version}
end
if redis.call('LINDEX', KEYS[1], index) ~= ARGV[1] then
  return {'0', current_version}
end
if ARGV[5] == ARGV[1] then
  return redis.error_reply('selection marker must differ from the envelope')
end
local generation = generation_value(KEYS[5])
if not generation_has_room(generation) then
  return redis.error_reply('dead-letter generation exhausted')
end
redis.call('INCR', KEYS[5])

redis.call('LSET', KEYS[1], index, ARGV[5])
if redis.call('LREM', KEYS[1], 1, ARGV[5]) ~= 1 then
  return redis.error_reply('selected envelope could not be removed')
end
redis.call('LPUSH', KEYS[2], ARGV[1])
if ARGV[3] == 'ordered' and redis.call('SADD', KEYS[4], ARGV[2]) == 1 then
  redis.call('LPUSH', KEYS[3], ARGV[2])
end
return {'1', list_version(KEYS[1], KEYS[5])}
`)

var appendDeadLetterScript = redis.NewScript(listVersionLua + `
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

if KEYS[1] == KEYS[2] then
  return redis.error_reply('dead-letter list and generation key must differ')
end
local problem = require_type(KEYS[1], 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(KEYS[2], 'string')
if problem then return redis.error_reply(problem) end
local generation = generation_value(KEYS[2])
if not generation then return redis.error_reply('invalid dead-letter generation') end
if not generation_has_room(generation) then
  return redis.error_reply('dead-letter generation exhausted')
end
redis.call('INCR', KEYS[2])
redis.call('LPUSH', KEYS[1], ARGV[1])
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
if string.len(digest) ~= 64 or not string.match(digest, '^[0-9a-f]+$') then
  return redis.error_reply('invalid ordered digest in ready ring')
end

local queue_key = ARGV[1] .. digest .. ARGV[2]
local lease_key = ARGV[3] .. digest .. ARGV[4]
local claim_keys = {KEYS[1], KEYS[2], KEYS[3], KEYS[4], KEYS[5], queue_key, lease_key}
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
problem = require_type(KEYS[5], 'hash')
if problem then return redis.error_reply(problem) end
problem = require_type(queue_key, 'list')
if problem then return redis.error_reply(problem) end
problem = require_type(lease_key, 'string')
if problem then return redis.error_reply(problem) end
if redis.call('HEXISTS', KEYS[5], KEYS[3]) == 1 then
  return redis.error_reply('processing list already has ordered claim metadata')
end
if redis.call('LLEN', KEYS[3]) ~= 0 then
  return nil
end

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
redis.call('PSETEX', KEYS[4], ARGV[7], ARGV[5])
redis.call('HSET', KEYS[5], KEYS[3], digest)
return {digest, payload}
`)

var renewOrderedLeaseScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return 1
end
return 0
`)

var transitionOrderedScript = redis.NewScript(listVersionLua + `
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
for i = 1, 8 do
  if KEYS[9] == KEYS[i] then
    return redis.error_reply('ordered claim registry must differ from queue state keys')
  end
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
problem = require_type(KEYS[9], 'hash')
if problem then return redis.error_reply(problem) end
if transition == 'complete' or transition == 'dead_letter' then
  problem = require_type(KEYS[7], 'list')
  if problem then return redis.error_reply(problem) end
end
if transition == 'dead_letter' then
  for i = 1, 7 do
    if KEYS[8] == KEYS[i] then
      return redis.error_reply('dead-letter generation key must differ from queue state keys')
    end
  end
  problem = require_type(KEYS[8], 'string')
  if problem then return redis.error_reply(problem) end
  local generation = generation_value(KEYS[8])
  if not generation then return redis.error_reply('invalid dead-letter generation') end
  if not generation_has_room(generation) then
    return redis.error_reply('dead-letter generation exhausted')
  end
end

local claim_digest = redis.call('HGET', KEYS[9], KEYS[1])
if claim_digest and claim_digest ~= ARGV[4] then
  return redis.error_reply('ordered claim metadata digest mismatch')
end
if redis.call('GET', KEYS[3]) ~= ARGV[1] then
  return 0
end
if redis.call('LREM', KEYS[1], 1, ARGV[2]) ~= 1 then
  return 0
end

if transition == 'retry' then
  redis.call('RPUSH', KEYS[4], ARGV[3])
elseif transition == 'dead_letter' then
  redis.call('INCR', KEYS[8])
  redis.call('LPUSH', KEYS[7], ARGV[3])
elseif transition == 'complete' then
  redis.call('LPUSH', KEYS[7], ARGV[3])
end

redis.call('DEL', KEYS[2])
redis.call('DEL', KEYS[3])
redis.call('HDEL', KEYS[9], KEYS[1])
if redis.call('LLEN', KEYS[4]) > 0 then
  redis.call('SADD', KEYS[6], ARGV[4])
  redis.call('LPUSH', KEYS[5], ARGV[4])
else
  redis.call('SREM', KEYS[6], ARGV[4])
  redis.call('DEL', KEYS[4])
end
return 1
`)

var settleAbandonedOrderedScript = redis.NewScript(`
local mode = ARGV[3]
if mode ~= 'recover' and mode ~= 'discard' then
  return redis.error_reply('unknown abandoned ordered settlement')
end
if redis.call('EXISTS', KEYS[2]) == 1 then
  return 0
end
if redis.call('EXISTS', KEYS[3]) == 1 then
  return 0
end

local recovery_keys = {KEYS[1], KEYS[2], KEYS[3], KEYS[4], KEYS[5], KEYS[6], KEYS[7]}
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
problem = require_type(KEYS[7], 'hash')
if problem then return redis.error_reply(problem) end

local claim_digest = redis.call('HGET', KEYS[7], KEYS[1])
if mode == 'discard' and not claim_digest then
  return 0
end
if claim_digest and claim_digest ~= ARGV[2] then
  return redis.error_reply('ordered claim metadata digest mismatch')
end

if redis.call('LREM', KEYS[1], 1, ARGV[1]) ~= 1 then
  return 0
end

redis.call('HDEL', KEYS[7], KEYS[1])
if mode == 'recover' then
  redis.call('RPUSH', KEYS[4], ARGV[1])
end
if redis.call('LLEN', KEYS[4]) > 0 then
  redis.call('SADD', KEYS[6], ARGV[2])
  redis.call('LPUSH', KEYS[5], ARGV[2])
else
  redis.call('SREM', KEYS[6], ARGV[2])
  redis.call('DEL', KEYS[4])
end
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
	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey, _, err := layout.keysForDigest(digest)
	if err != nil {
		return err
	}
	if err := enqueueOrderedScript.Eval(ctx, rdb,
		[]string{queueKey, layout.ReadyList, layout.ActiveSet}, encoded, digest).Err(); err != nil {
		return fmt.Errorf("enqueue ordered job: %w", err)
	}
	return nil
}

// AppendDeadLetter atomically appends an envelope and advances the bounded
// mutation generation used by administrative selection handles.
func AppendDeadLetter(ctx context.Context, rdb redis.Cmdable, deadLetterList, encoded string) error {
	if deadLetterList == "" {
		return errors.New("dead-letter list must be non-empty")
	}
	if err := appendDeadLetterScript.Eval(
		ctx,
		rdb,
		[]string{deadLetterList, queuekeys.DLQGenerationKey(deadLetterList)},
		encoded,
	).Err(); err != nil {
		return fmt.Errorf("append dead-letter job: %w", err)
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

	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey, _, err := layout.keysForDigest(digest)
	if err != nil {
		return false, err
	}
	result, err := requeueEncodedScript.Eval(ctx, rdb,
		[]string{source, queueKey, layout.ReadyList, layout.ActiveSet}, encoded, digest, "ordered").Int64()
	if err != nil {
		return false, fmt.Errorf("requeue ordered job: %w", err)
	}
	return result == 1, nil
}

// RequeueEncodedAt atomically removes the exact envelope at a non-negative
// source-list index and returns it to ordinary or ordered intake. A changed
// source snapshot, index, or envelope is treated as a stale selection and
// returns moved=false with the current snapshot.
func RequeueEncodedAt(ctx context.Context, rdb redis.Cmdable, source, destination string, index int64, job Job, encoded, marker, expectedSnapshot string, layout OrderingLayout) (bool, string, error) {
	if source == "" {
		return false, "", errors.New("requeue source must be non-empty")
	}
	if index < 0 {
		return false, "", errors.New("requeue source index must be non-negative")
	}
	if marker == "" || marker == encoded {
		return false, "", errors.New("requeue selection marker must be non-empty and differ from the envelope")
	}
	if expectedSnapshot == "" {
		return false, "", errors.New("requeue source snapshot must be non-empty")
	}
	if job.OrderingKey == "" {
		if destination == "" {
			return false, "", errors.New("requeue destination must be non-empty")
		}
		result, err := requeueEncodedAtScript.Eval(ctx, rdb,
			[]string{source, destination, destination, destination, queuekeys.DLQGenerationKey(source)}, encoded, "", "ordinary", index, marker, expectedSnapshot).Result()
		if err != nil {
			return false, "", fmt.Errorf("requeue selected job: %w", err)
		}
		return parseSelectedRequeueResult(result)
	}

	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey, _, err := layout.keysForDigest(digest)
	if err != nil {
		return false, "", err
	}
	result, err := requeueEncodedAtScript.Eval(ctx, rdb,
		[]string{source, queueKey, layout.ReadyList, layout.ActiveSet, queuekeys.DLQGenerationKey(source)}, encoded, digest, "ordered", index, marker, expectedSnapshot).Result()
	if err != nil {
		return false, "", fmt.Errorf("requeue selected ordered job: %w", err)
	}
	return parseSelectedRequeueResult(result)
}

func parseSelectedRequeueResult(result interface{}) (bool, string, error) {
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return false, "", fmt.Errorf("unexpected selected requeue response %T", result)
	}
	moved, ok := values[0].(string)
	if !ok || (moved != "0" && moved != "1") {
		return false, "", fmt.Errorf("unexpected selected requeue status %v", values[0])
	}
	snapshot, ok := values[1].(string)
	if !ok || snapshot == "" {
		return false, "", fmt.Errorf("unexpected selected requeue snapshot %T", values[1])
	}
	return moved == "1", snapshot, nil
}

// ClaimOrdered atomically leases the oldest ready key and moves its oldest job
// into the existing per-worker processing list. An empty ready ring or occupied
// processing list returns ok=false without error.
func ClaimOrdered(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, processingList, heartbeatKey, owner string, ttl time.Duration) (delivery OrderedDelivery, ok bool, err error) {
	if err := layout.Validate(); err != nil {
		return delivery, false, err
	}
	queuePrefix, queueSuffix, _ := queuekeys.SplitPattern(layout.QueuePattern)
	leasePrefix, leaseSuffix, _ := queuekeys.SplitPattern(layout.LeasePattern)
	ttlMillis := durationMillis(ttl)
	result, err := claimOrderedScript.Eval(ctx, rdb,
		[]string{layout.ReadyList, layout.ActiveSet, processingList, heartbeatKey, layout.claimsKey()},
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
	generationKey := destination
	if transition == OrderedDeadLetter {
		generationKey = queuekeys.DLQGenerationKey(destination)
	}
	result, err := transitionOrderedScript.Eval(ctx, rdb, []string{
		processingList,
		heartbeatKey,
		delivery.LeaseKey,
		delivery.QueueKey,
		layout.ReadyList,
		layout.ActiveSet,
		destination,
		generationKey,
		layout.claimsKey(),
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
	result, err := settleAbandonedOrderedScript.Eval(ctx, rdb, []string{
		processingList,
		heartbeatKey,
		delivery.LeaseKey,
		delivery.QueueKey,
		layout.ReadyList,
		layout.ActiveSet,
		layout.claimsKey(),
	}, delivery.Payload, delivery.Digest, "recover").Int64()
	if err != nil {
		return false, fmt.Errorf("recover ordered job: %w", err)
	}
	return result == 1, nil
}

// OrderedDeliveryFromClaim reconstructs recovery metadata for an in-flight
// ordered envelope even when the envelope itself is malformed. ClaimOrdered
// records the digest by processing list before returning the delivery.
func OrderedDeliveryFromClaim(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, processingList, payload string) (OrderedDelivery, bool, error) {
	if err := layout.Validate(); err != nil {
		return OrderedDelivery{}, false, err
	}
	digest, err := rdb.HGet(ctx, layout.claimsKey(), processingList).Result()
	if err == redis.Nil {
		return OrderedDelivery{}, false, nil
	}
	if err != nil {
		return OrderedDelivery{}, false, fmt.Errorf("read ordered claim metadata: %w", err)
	}
	if !queuekeys.IsOrderingDigest(digest) {
		return OrderedDelivery{}, false, fmt.Errorf("invalid ordered claim digest %q", digest)
	}
	return OrderedDelivery{
		Digest:   digest,
		Payload:  payload,
		QueueKey: queuekeys.Format(layout.QueuePattern, digest),
		LeaseKey: queuekeys.Format(layout.LeasePattern, digest),
	}, true, nil
}

// DiscardAbandonedOrdered removes a malformed ordered envelope after its
// worker heartbeat and lease have expired, then advances the same-key FIFO.
// The durable claim registry binds the opaque payload to its ordering digest.
func DiscardAbandonedOrdered(ctx context.Context, rdb redis.Cmdable, layout OrderingLayout, delivery OrderedDelivery, processingList, heartbeatKey string) (bool, error) {
	if err := layout.Validate(); err != nil {
		return false, err
	}
	result, err := settleAbandonedOrderedScript.Eval(ctx, rdb, []string{
		processingList,
		heartbeatKey,
		delivery.LeaseKey,
		delivery.QueueKey,
		layout.ReadyList,
		layout.ActiveSet,
		layout.claimsKey(),
	}, delivery.Payload, delivery.Digest, "discard").Int64()
	if err != nil {
		return false, fmt.Errorf("discard abandoned ordered job: %w", err)
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
