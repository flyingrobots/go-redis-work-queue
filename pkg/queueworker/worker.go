// Copyright 2026 James Ross
// Package queueworker exposes the supported application-worker runtime.
package queueworker

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	internalconfig "github.com/flyingrobots/go-redis-work-queue/internal/config"
	internalreaper "github.com/flyingrobots/go-redis-work-queue/internal/reaper"
	internalworker "github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	// ErrHandlerRequired is returned before a worker can consume jobs without
	// application logic.
	ErrHandlerRequired = internalworker.ErrHandlerRequired
	// ErrBenchJobFailed identifies the explicit legacy benchmark failure case.
	ErrBenchJobFailed = internalworker.ErrBenchJobFailed
)

// Handler executes one durable job. Multiple worker goroutines may call it
// concurrently, while jobs with the same non-empty ordering key are serialized.
type Handler func(ctx context.Context, job queueclient.Job) error

// BackoffConfig controls retry delays after Handler returns an error.
type BackoffConfig struct {
	Base time.Duration
	Max  time.Duration
}

// CircuitBreakerConfig controls consumption pauses after handler failures.
type CircuitBreakerConfig struct {
	FailureThreshold float64
	Window           time.Duration
	CooldownPeriod   time.Duration
	MinSamples       int
}

// Config defines worker concurrency, delivery state, and Redis queue keys.
// Start with DefaultConfig and then apply application-specific overrides.
type Config struct {
	Count             int
	HeartbeatTTL      time.Duration
	MaxRetries        int
	Backoff           BackoffConfig
	Priorities        []string
	QueueConfig       queueclient.Config
	BRPopLPushTimeout time.Duration
	BreakerPause      time.Duration
	CircuitBreaker    CircuitBreakerConfig
}

// DefaultConfig returns an independent copy of the standard worker layout.
func DefaultConfig() Config {
	cfg := internalconfig.Default()
	queues := make(map[string]string, len(cfg.Worker.Queues))
	for name, key := range cfg.Worker.Queues {
		queues[name] = key
	}
	return Config{
		Count:        cfg.Worker.Count,
		HeartbeatTTL: cfg.Worker.HeartbeatTTL,
		MaxRetries:   cfg.Worker.MaxRetries,
		Backoff: BackoffConfig{
			Base: cfg.Worker.Backoff.Base,
			Max:  cfg.Worker.Backoff.Max,
		},
		Priorities: append([]string(nil), cfg.Worker.Priorities...),
		QueueConfig: queueclient.Config{
			Queues:                queues,
			DefaultPriority:       cfg.Producer.DefaultPriority,
			ProcessingListPattern: cfg.Worker.ProcessingListPattern,
			HeartbeatKeyPattern:   cfg.Worker.HeartbeatKeyPattern,
			CompletedList:         cfg.Worker.CompletedList,
			DeadLetterList:        cfg.Worker.DeadLetterList,
			MaxPayloadSize:        cfg.Queue.MaxPayloadSize,
			OrderedReadyList:      cfg.Queue.OrderedReadyList,
			OrderedActiveSet:      cfg.Queue.OrderedActiveSet,
			OrderedQueuePattern:   cfg.Queue.OrderedQueuePattern,
			OrderedLeasePattern:   cfg.Queue.OrderedLeasePattern,
		},
		BRPopLPushTimeout: cfg.Worker.BRPopLPushTimeout,
		BreakerPause:      cfg.Worker.BreakerPause,
		CircuitBreaker: CircuitBreakerConfig{
			FailureThreshold: cfg.CircuitBreaker.FailureThreshold,
			Window:           cfg.CircuitBreaker.Window,
			CooldownPeriod:   cfg.CircuitBreaker.CooldownPeriod,
			MinSamples:       cfg.CircuitBreaker.MinSamples,
		},
	}
}

// Worker runs application handlers over the repository's crash-safe Redis
// intake, retry, dead-letter, heartbeat, reaper, and per-key FIFO protocol.
type Worker struct {
	inner  *internalworker.Worker
	reaper *internalreaper.Reaper
	rdb    *redis.Client
	owned  bool
}

// New constructs a worker that owns its Redis connection. A nil Handler is
// always rejected; pass BenchHandler explicitly for benchmark-only workloads.
func New(redisOpts *redis.Options, cfg Config, handler Handler, logger *zap.Logger) (*Worker, error) {
	if redisOpts == nil {
		return nil, errors.New("redis options must not be nil")
	}
	internalCfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, ErrHandlerRequired
	}
	rdb := redis.NewClient(redisOpts)
	return newWithClient(rdb, internalCfg, handler, logger, true), nil
}

// NewWithClient constructs a worker that borrows an existing Redis client.
// Close is a no-op for a borrowed client.
func NewWithClient(rdb *redis.Client, cfg Config, handler Handler, logger *zap.Logger) (*Worker, error) {
	if rdb == nil {
		return nil, errors.New("redis client must not be nil")
	}
	internalCfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, ErrHandlerRequired
	}
	return newWithClient(rdb, internalCfg, handler, logger, false), nil
}

func newWithClient(rdb *redis.Client, cfg *internalconfig.Config, handler Handler, logger *zap.Logger, owned bool) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}
	inner := internalworker.New(cfg, rdb, logger).Handle(internalworker.Handler(handler))
	return &Worker{inner: inner, reaper: internalreaper.New(cfg, rdb, logger), rdb: rdb, owned: owned}
}

// Run consumes jobs until ctx is canceled. The same Worker must not be run
// more than once concurrently.
func (w *Worker) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reaperDone := make(chan struct{})
	go func() {
		defer close(reaperDone)
		w.reaper.Run(runCtx)
	}()
	err := w.inner.Run(runCtx)
	cancel()
	<-reaperDone
	return err
}

// Close releases a Redis connection created by New.
func (w *Worker) Close() error {
	if w == nil || !w.owned || w.rdb == nil {
		return nil
	}
	return w.rdb.Close()
}

// BenchHandler is the explicit legacy benchmark handler. It never interprets
// application payload bytes.
func BenchHandler(ctx context.Context, job queueclient.Job) error {
	return internalworker.BenchHandler(ctx, job)
}

func normalizeConfig(cfg Config) (*internalconfig.Config, error) {
	if reflect.DeepEqual(cfg, Config{}) {
		cfg = DefaultConfig()
	}
	if cfg.MaxRetries < 0 {
		return nil, fmt.Errorf("max retries must be >= 0")
	}
	if cfg.Backoff.Base <= 0 || cfg.Backoff.Max < cfg.Backoff.Base {
		return nil, fmt.Errorf("retry backoff must have 0 < base <= max")
	}
	if cfg.BreakerPause <= 0 {
		return nil, fmt.Errorf("breaker pause must be > 0")
	}
	if cfg.CircuitBreaker.FailureThreshold <= 0 || cfg.CircuitBreaker.FailureThreshold > 1 {
		return nil, fmt.Errorf("circuit breaker failure threshold must be within (0, 1]")
	}
	if cfg.CircuitBreaker.Window <= 0 || cfg.CircuitBreaker.CooldownPeriod <= 0 || cfg.CircuitBreaker.MinSamples < 1 {
		return nil, fmt.Errorf("circuit breaker window, cooldown, and minimum samples must be positive")
	}

	queueCfg, err := queueclient.NormalizeConfig(cfg.QueueConfig)
	if err != nil {
		return nil, fmt.Errorf("queue config: %w", err)
	}
	internalCfg := internalconfig.Default()
	internalCfg.Worker.Count = cfg.Count
	internalCfg.Worker.HeartbeatTTL = cfg.HeartbeatTTL
	internalCfg.Worker.MaxRetries = cfg.MaxRetries
	internalCfg.Worker.Backoff.Base = cfg.Backoff.Base
	internalCfg.Worker.Backoff.Max = cfg.Backoff.Max
	internalCfg.Worker.Priorities = append([]string(nil), cfg.Priorities...)
	internalCfg.Worker.Queues = queueCfg.Queues
	internalCfg.Worker.ProcessingListPattern = queueCfg.ProcessingListPattern
	internalCfg.Worker.HeartbeatKeyPattern = queueCfg.HeartbeatKeyPattern
	internalCfg.Worker.CompletedList = queueCfg.CompletedList
	internalCfg.Worker.DeadLetterList = queueCfg.DeadLetterList
	internalCfg.Worker.BRPopLPushTimeout = cfg.BRPopLPushTimeout
	internalCfg.Worker.BreakerPause = cfg.BreakerPause
	internalCfg.Producer.DefaultPriority = queueCfg.DefaultPriority
	internalCfg.Queue.MaxPayloadSize = queueCfg.MaxPayloadSize
	internalCfg.Queue.OrderedReadyList = queueCfg.OrderedReadyList
	internalCfg.Queue.OrderedActiveSet = queueCfg.OrderedActiveSet
	internalCfg.Queue.OrderedQueuePattern = queueCfg.OrderedQueuePattern
	internalCfg.Queue.OrderedLeasePattern = queueCfg.OrderedLeasePattern
	internalCfg.CircuitBreaker.FailureThreshold = cfg.CircuitBreaker.FailureThreshold
	internalCfg.CircuitBreaker.Window = cfg.CircuitBreaker.Window
	internalCfg.CircuitBreaker.CooldownPeriod = cfg.CircuitBreaker.CooldownPeriod
	internalCfg.CircuitBreaker.MinSamples = cfg.CircuitBreaker.MinSamples
	if err := internalconfig.Validate(internalCfg); err != nil {
		return nil, fmt.Errorf("worker config: %w", err)
	}
	return internalCfg, nil
}
