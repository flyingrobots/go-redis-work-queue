// Copyright 2026 James Ross
// Package queueclient provides the supported, repository-external enqueue and
// inspection surface for the Redis work queue.
package queueclient

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	internalqueue "github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Job is the durable queue envelope. It aliases the core type so public
// clients and workers exchange exactly the same JSON representation.
type Job = internalqueue.Job

// PayloadTooLargeError reports a payload rejected before Redis is modified.
type PayloadTooLargeError = internalqueue.PayloadTooLargeError

const DefaultMaxPayloadSize = internalqueue.DefaultMaxPayloadSize

var (
	// ErrPayloadTooLarge identifies payload size guard failures.
	ErrPayloadTooLarge = internalqueue.ErrPayloadTooLarge
	// ErrConnection identifies Redis command failures.
	ErrConnection = errors.New("queue Redis connection failed")
	// ErrUnknownPriority identifies jobs whose priority is not configured.
	ErrUnknownPriority = errors.New("unknown queue priority")
	// ErrUnknownQueue identifies Peek aliases that are not configured.
	ErrUnknownQueue = errors.New("unknown queue")
)

// ConnectionError adds the failed operation while retaining the underlying
// go-redis or network error for errors.Is/errors.As inspection.
type ConnectionError struct {
	Operation string
	Err       error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("redis %s: %v", e.Operation, e.Err)
}

func (e *ConnectionError) Unwrap() error { return e.Err }

func (e *ConnectionError) Is(target error) bool { return target == ErrConnection }

// UnknownPriorityError lists the configured alternatives to an invalid job
// priority. An empty job priority selects Config.DefaultPriority instead.
type UnknownPriorityError struct {
	Priority string
	Known    []string
}

func (e *UnknownPriorityError) Error() string {
	return fmt.Sprintf("unknown priority %q; configured priorities: %s", e.Priority, strings.Join(e.Known, ", "))
}

func (e *UnknownPriorityError) Unwrap() error { return ErrUnknownPriority }

// UnknownQueueError reports an invalid Peek queue alias.
type UnknownQueueError struct {
	Queue string
	Known []string
}

func (e *UnknownQueueError) Error() string {
	return fmt.Sprintf("unknown queue %q; known queues: %s", e.Queue, strings.Join(e.Known, ", "))
}

func (e *UnknownQueueError) Unwrap() error { return ErrUnknownQueue }

// Config names the Redis lists shared with workers. A zero Config uses the
// standard queue layout. Queues may define priority names beyond high and low.
type Config struct {
	Queues                map[string]string
	DefaultPriority       string
	ProcessingListPattern string
	HeartbeatKeyPattern   string
	CompletedList         string
	DeadLetterList        string
	MaxPayloadSize        int
	OrderedReadyList      string
	OrderedActiveSet      string
	OrderedQueuePattern   string
	OrderedLeasePattern   string
}

// DefaultConfig returns an independent copy of the standard queue layout.
func DefaultConfig() Config {
	return Config{
		Queues: map[string]string{
			"high": queuekeys.DefaultHighPriorityQueue,
			"low":  queuekeys.DefaultLowPriorityQueue,
		},
		DefaultPriority:       "low",
		ProcessingListPattern: queuekeys.DefaultProcessingListPattern,
		HeartbeatKeyPattern:   queuekeys.DefaultHeartbeatKeyPattern,
		CompletedList:         queuekeys.DefaultCompletedList,
		DeadLetterList:        queuekeys.DefaultDeadLetterList,
		MaxPayloadSize:        DefaultMaxPayloadSize,
		OrderedReadyList:      queuekeys.DefaultOrderedReadyList,
		OrderedActiveSet:      queuekeys.DefaultOrderedActiveSet,
		OrderedQueuePattern:   queuekeys.DefaultOrderedQueuePattern,
		OrderedLeasePattern:   queuekeys.DefaultOrderedLeasePattern,
	}
}

// Client enqueues and inspects jobs using the same Redis list protocol as the
// worker. A Client created with New owns its Redis connection and should be
// closed. A Client created with NewWithClient borrows the supplied connection.
type Client struct {
	rdb   *redis.Client
	cfg   Config
	owned bool
}

// New constructs a queue client. It validates local configuration but does not
// contact Redis; connection failures are reported as *ConnectionError by the
// first operation.
func New(redisOpts *redis.Options, cfg Config) (*Client, error) {
	if redisOpts == nil {
		return nil, errors.New("redis options must not be nil")
	}
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{rdb: redis.NewClient(redisOpts), cfg: normalized, owned: true}, nil
}

// NewWithClient constructs a queue client that borrows an existing go-redis
// client. Close is a no-op for borrowed clients.
func NewWithClient(rdb *redis.Client, cfg Config) (*Client, error) {
	if rdb == nil {
		return nil, errors.New("redis client must not be nil")
	}
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{rdb: rdb, cfg: normalized}, nil
}

// Close releases a Redis connection owned by this client.
func (c *Client) Close() error {
	if c == nil || !c.owned || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// Enqueue validates and appends one job, returning its durable ID. Empty IDs
// and creation times are generated. Empty priorities select DefaultPriority.
// Retries is worker-owned bookkeeping and is reset to zero for a new delivery.
// Explicit duplicate IDs are accepted as separate at-least-once deliveries;
// callers that need deduplication must enforce it at the handler boundary.
func (c *Client) Enqueue(ctx context.Context, job Job) (string, error) {
	prepared, queueName, encoded, err := c.prepare(job)
	if err != nil {
		return "", err
	}
	if err := internalqueue.AppendEncoded(ctx, c.rdb, queueName, prepared, encoded, c.orderingLayout()); err != nil {
		return "", connectionError("enqueue", err)
	}
	return prepared.ID, nil
}

// EnqueueBatch validates and encodes every job before running one Redis
// script. If any job is locally invalid or any destination has the wrong Redis
// type, none are written. Redis executes the accepted batch atomically. On
// success, generated IDs/timestamps are copied back into the caller-provided
// slice and worker-owned retry counters are reset. Explicit duplicate IDs
// remain separate deliveries.
func (c *Client) EnqueueBatch(ctx context.Context, jobs []Job) error {
	if len(jobs) == 0 {
		return nil
	}
	entries := make([]internalqueue.PreparedEnqueue, len(jobs))
	for i, job := range jobs {
		prepared, queueName, encoded, err := c.prepare(job)
		if err != nil {
			return fmt.Errorf("job %d: %w", i, err)
		}
		entries[i] = internalqueue.PreparedEnqueue{Job: prepared, QueueName: queueName, Encoded: encoded}
	}

	if err := internalqueue.AppendEncodedBatch(ctx, c.rdb, entries, c.orderingLayout()); err != nil {
		return connectionError("enqueue batch", err)
	}
	for i := range entries {
		jobs[i] = entries[i].Job
	}
	return nil
}

// StatsResult mirrors the core admin queue and worker counts.
type StatsResult struct {
	Queues          map[string]int64 `json:"queues"`
	OrderedPending  int64            `json:"ordered_pending"`
	ProcessingLists map[string]int64 `json:"processing_lists"`
	Heartbeats      int64            `json:"heartbeats"`
}

// Stats returns queue lengths, in-flight list lengths, and heartbeat count.
func (c *Client) Stats(ctx context.Context) (StatsResult, error) {
	result := StatsResult{
		Queues:          make(map[string]int64, len(c.cfg.Queues)+2),
		ProcessingLists: map[string]int64{},
	}
	queues := make(map[string]string, len(c.cfg.Queues)+2)
	for name, key := range c.cfg.Queues {
		queues[name] = key
	}
	queues["completed"] = c.cfg.CompletedList
	queues["dead_letter"] = c.cfg.DeadLetterList
	for name, key := range queues {
		count, err := c.rdb.LLen(ctx, key).Result()
		if err != nil {
			return result, connectionError("read queue statistics", err)
		}
		result.Queues[name+"("+key+")"] = count
	}
	_, orderedPending, err := internalqueue.OrderedQueueLengths(ctx, c.rdb, c.cfg.OrderedQueuePattern)
	if err != nil {
		return result, connectionError("read ordered queue statistics", err)
	}
	result.OrderedPending = orderedPending

	var cursor uint64
	processingPattern := queuekeys.ScanPattern(c.cfg.ProcessingListPattern)
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, processingPattern, 200).Result()
		if err != nil {
			return result, connectionError("scan processing lists", err)
		}
		for _, key := range keys {
			count, err := c.rdb.LLen(ctx, key).Result()
			if err != nil {
				return result, connectionError("read processing list", err)
			}
			result.ProcessingLists[key] = count
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	cursor = 0
	heartbeatPattern := queuekeys.ScanPattern(c.cfg.HeartbeatKeyPattern)
	heartbeatKeys := map[string]struct{}{}
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, heartbeatPattern, 500).Result()
		if err != nil {
			return result, connectionError("scan heartbeats", err)
		}
		for _, key := range keys {
			heartbeatKeys[key] = struct{}{}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	result.Heartbeats = int64(len(heartbeatKeys))
	return result, nil
}

// PeekResult contains jobs without removing them from Redis.
type PeekResult struct {
	Queue string   `json:"queue"`
	Items []string `json:"items"`
}

// Peek mirrors the admin peek operation. Non-positive n defaults to ten.
func (c *Client) Peek(ctx context.Context, queueAlias string, n int64) (PeekResult, error) {
	queueName, err := c.resolveQueue(queueAlias)
	if err != nil {
		return PeekResult{}, err
	}
	if n <= 0 {
		n = 10
	}
	items, err := c.rdb.LRange(ctx, queueName, -n, -1).Result()
	if err != nil {
		return PeekResult{}, connectionError("peek queue", err)
	}
	return PeekResult{Queue: queueName, Items: items}, nil
}

func (c *Client) prepare(job Job) (Job, string, string, error) {
	job.Retries = 0
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	if job.CreationTime == "" {
		job.CreationTime = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if job.Priority == "" {
		job.Priority = c.cfg.DefaultPriority
	}
	queueName, ok := c.cfg.Queues[job.Priority]
	if !ok {
		return Job{}, "", "", &UnknownPriorityError{Priority: job.Priority, Known: sortedKeys(c.cfg.Queues)}
	}
	encoded, err := internalqueue.EncodeForEnqueue(job, c.cfg.MaxPayloadSize)
	if err != nil {
		return Job{}, "", "", err
	}
	return job, queueName, encoded, nil
}

func (c *Client) orderingLayout() internalqueue.OrderingLayout {
	return internalqueue.OrderingLayout{
		ReadyList:    c.cfg.OrderedReadyList,
		ActiveSet:    c.cfg.OrderedActiveSet,
		QueuePattern: c.cfg.OrderedQueuePattern,
		LeasePattern: c.cfg.OrderedLeasePattern,
	}
}

func (c *Client) resolveQueue(alias string) (string, error) {
	if queueName, ok := c.cfg.Queues[alias]; ok {
		return queueName, nil
	}

	normalized := strings.ToLower(alias)
	if normalized == "completed" {
		return c.cfg.CompletedList, nil
	}
	if normalized == "dead_letter" || normalized == "dlq" {
		return c.cfg.DeadLetterList, nil
	}
	if queueName, ok := c.cfg.Queues[normalized]; ok {
		return queueName, nil
	}
	for _, queueName := range c.cfg.Queues {
		if alias == queueName {
			return queueName, nil
		}
	}
	if alias == c.cfg.CompletedList || alias == c.cfg.DeadLetterList || strings.HasPrefix(alias, queuekeys.Namespace) {
		return alias, nil
	}
	known := sortedKeys(c.cfg.Queues)
	known = append(known, "completed", "dead_letter")
	return "", &UnknownQueueError{Queue: alias, Known: known}
}

// NormalizeConfig applies safe defaults, copies caller-owned queue maps, and
// validates the Redis key layout without contacting Redis.
func NormalizeConfig(cfg Config) (Config, error) {
	defaults := DefaultConfig()
	if len(cfg.Queues) == 0 {
		cfg.Queues = defaults.Queues
	} else {
		queues := make(map[string]string, len(cfg.Queues))
		for name, key := range cfg.Queues {
			name = strings.TrimSpace(name)
			key = strings.TrimSpace(key)
			if name == "" || key == "" {
				return Config{}, errors.New("queue priority names and Redis keys must not be empty")
			}
			if queuekeys.IsReservedQueueAlias(name) {
				return Config{}, fmt.Errorf("queue priority %q is reserved", name)
			}
			queues[name] = key
		}
		cfg.Queues = queues
	}
	if cfg.DefaultPriority == "" {
		if _, ok := cfg.Queues[defaults.DefaultPriority]; ok {
			cfg.DefaultPriority = defaults.DefaultPriority
		} else {
			cfg.DefaultPriority = sortedKeys(cfg.Queues)[0]
		}
	}
	if _, ok := cfg.Queues[cfg.DefaultPriority]; !ok {
		return Config{}, fmt.Errorf("default priority %q has no configured queue", cfg.DefaultPriority)
	}
	if cfg.ProcessingListPattern == "" {
		cfg.ProcessingListPattern = defaults.ProcessingListPattern
	}
	if cfg.HeartbeatKeyPattern == "" {
		cfg.HeartbeatKeyPattern = defaults.HeartbeatKeyPattern
	}
	if strings.Count(cfg.ProcessingListPattern, "%s") != 1 {
		return Config{}, errors.New("processing list pattern must contain exactly one %s placeholder")
	}
	if strings.Count(cfg.HeartbeatKeyPattern, "%s") != 1 {
		return Config{}, errors.New("heartbeat key pattern must contain exactly one %s placeholder")
	}
	if cfg.CompletedList == "" {
		cfg.CompletedList = defaults.CompletedList
	}
	if cfg.DeadLetterList == "" {
		cfg.DeadLetterList = defaults.DeadLetterList
	}
	if cfg.MaxPayloadSize <= 0 {
		cfg.MaxPayloadSize = defaults.MaxPayloadSize
	}
	if cfg.OrderedReadyList == "" {
		cfg.OrderedReadyList = defaults.OrderedReadyList
	}
	if cfg.OrderedActiveSet == "" {
		cfg.OrderedActiveSet = defaults.OrderedActiveSet
	}
	if cfg.OrderedReadyList == cfg.OrderedActiveSet {
		return Config{}, errors.New("ordered ready list and active set must differ")
	}
	if cfg.OrderedQueuePattern == "" {
		cfg.OrderedQueuePattern = defaults.OrderedQueuePattern
	}
	if cfg.OrderedLeasePattern == "" {
		cfg.OrderedLeasePattern = defaults.OrderedLeasePattern
	}
	if strings.Count(cfg.OrderedQueuePattern, "%s") != 1 {
		return Config{}, errors.New("ordered queue pattern must contain exactly one %s placeholder")
	}
	if strings.Count(cfg.OrderedLeasePattern, "%s") != 1 {
		return Config{}, errors.New("ordered lease pattern must contain exactly one %s placeholder")
	}
	if cfg.OrderedQueuePattern == cfg.OrderedLeasePattern {
		return Config{}, errors.New("ordered queue and lease patterns must differ")
	}
	return cfg, nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func connectionError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &ConnectionError{Operation: operation, Err: err}
}
