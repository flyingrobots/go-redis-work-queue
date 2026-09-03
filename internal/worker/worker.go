// Copyright 2025 James Ross
package worker

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/breaker"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Worker struct {
	cfg    *config.Config
	rdb    *redis.Client
	log    *zap.Logger
	cb     *breaker.CircuitBreaker
	baseID string

	handlerMu sync.RWMutex
	handler   Handler
}

type delivery struct {
	payload     string
	sourceQueue string
	ordered     *queue.OrderedDelivery
}

func New(cfg *config.Config, rdb *redis.Client, log *zap.Logger) *Worker {
	if log == nil {
		log = zap.NewNop()
	}
	cb := breaker.New(cfg.CircuitBreaker.Window, cfg.CircuitBreaker.CooldownPeriod, cfg.CircuitBreaker.FailureThreshold, cfg.CircuitBreaker.MinSamples)
	host, _ := os.Hostname()
	pid := os.Getpid()
	now := time.Now().UnixNano()
	randSfx := fmt.Sprintf("%04x", time.Now().UnixNano()&0xffff)
	base := fmt.Sprintf("%s-%d-%d-%s", host, pid, now, randSfx)
	return &Worker{cfg: cfg, rdb: rdb, log: log, cb: cb, baseID: base}
}

func (w *Worker) Run(ctx context.Context) error {
	if w.selectedHandler() == nil {
		return ErrHandlerRequired
	}

	var wg sync.WaitGroup
	for i := 0; i < w.cfg.Worker.Count; i++ {
		wg.Add(1)
		id := fmt.Sprintf("%s-%d", w.baseID, i)
		go func(workerID string) {
			defer wg.Done()
			obs.WorkerActive.Inc()
			defer obs.WorkerActive.Dec()
			w.runOne(ctx, workerID)
		}(id)
	}

	// periodically update breaker state metric
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				switch w.cb.State() {
				case breaker.Closed:
					obs.CircuitBreakerState.Set(0)
				case breaker.HalfOpen:
					obs.CircuitBreakerState.Set(1)
				case breaker.Open:
					obs.CircuitBreakerState.Set(2)
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

func (w *Worker) runOne(ctx context.Context, workerID string) {
	procList := queuekeys.Format(w.cfg.Worker.ProcessingListPattern, workerID)
	hbKey := queuekeys.Format(w.cfg.Worker.HeartbeatKeyPattern, workerID)
	preferUnordered := false
	unorderedBurst := 0

	for ctx.Err() == nil {
		if !w.cb.Allow() {
			time.Sleep(w.cfg.Worker.BreakerPause)
			continue
		}

		var next delivery
		found := false
		if preferUnordered && unorderedBurst < 32 {
			var err error
			next, found, err = w.dequeueUnordered(ctx, workerID, procList, hbKey, false)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				w.log.Warn("RPOPLPUSH error", obs.Err(err))
				time.Sleep(50 * time.Millisecond)
				continue
			}
			if found {
				unorderedBurst++
			}
		}

		if !found {
			claimed, ok, err := w.claimOrdered(ctx, workerID, procList, hbKey)
			unorderedBurst = 0
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				w.log.Warn("ordered claim error", obs.Err(err))
				time.Sleep(50 * time.Millisecond)
				continue
			}
			if ok {
				next = claimed
				found = true
				preferUnordered = true
				unorderedBurst = 31
			}
		}

		if !found {
			var err error
			next, found, err = w.dequeueUnordered(ctx, workerID, procList, hbKey, true)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				w.log.Warn("BRPOPLPUSH error", obs.Err(err))
				time.Sleep(50 * time.Millisecond)
				continue
			}
			if found {
				preferUnordered = true
				if next.ordered != nil {
					unorderedBurst = 31
				} else {
					unorderedBurst = 1
				}
			}
		}
		if !found {
			continue
		}

		obs.JobsConsumed.Inc()

		// measure state transition around Record() to count trips
		start := time.Now()
		// process job
		ok := w.processDelivery(ctx, workerID, procList, hbKey, next)
		obs.JobProcessingDuration.Observe(time.Since(start).Seconds())
		prev := w.cb.State()
		w.cb.Record(ok)
		curr := w.cb.State()
		if prev != curr && curr == breaker.Open {
			obs.CircuitBreakerTrips.Inc()
		}
	}
}

func (w *Worker) claimOrdered(ctx context.Context, workerID, procList, hbKey string) (delivery, bool, error) {
	claimed, ok, err := queue.ClaimOrdered(
		ctx,
		w.rdb,
		w.cfg.OrderingLayout(),
		procList,
		hbKey,
		workerID,
		w.cfg.Worker.HeartbeatTTL,
	)
	if err != nil || !ok {
		return delivery{}, false, err
	}
	return delivery{payload: claimed.Payload, sourceQueue: claimed.QueueKey, ordered: &claimed}, true, nil
}

func (w *Worker) dequeueUnordered(ctx context.Context, workerID, procList, hbKey string, block bool) (delivery, bool, error) {
	for _, priority := range w.cfg.Worker.Priorities {
		key := w.cfg.Worker.Queues[priority]
		if key == "" {
			continue
		}

		deqCtx, deqSpan := obs.StartDequeueSpan(ctx, key)
		var (
			payload string
			err     error
		)
		if block {
			payload, err = w.rdb.BRPopLPush(deqCtx, key, procList, w.cfg.Worker.BRPopLPushTimeout).Result()
		} else {
			payload, err = w.rdb.RPopLPush(deqCtx, key, procList).Result()
		}
		if err == redis.Nil {
			deqSpan.End()
			if block {
				claimed, ok, claimErr := w.claimOrdered(ctx, workerID, procList, hbKey)
				if claimErr != nil || ok {
					return claimed, ok, claimErr
				}
			}
			continue
		}
		if err != nil {
			obs.RecordError(deqCtx, err)
			deqSpan.End()
			return delivery{}, false, err
		}

		obs.SetSpanSuccess(deqCtx)
		obs.AddEvent(deqCtx, "job_dequeued", obs.KeyValue("queue", key))
		deqSpan.End()
		return delivery{payload: payload, sourceQueue: key}, true, nil
	}
	return delivery{}, false, nil
}

func (w *Worker) processJob(ctx context.Context, workerID, srcQueue, procList, hbKey, payload string) bool {
	return w.processDelivery(ctx, workerID, procList, hbKey, delivery{payload: payload, sourceQueue: srcQueue})
}

func (w *Worker) processDelivery(ctx context.Context, workerID, procList, hbKey string, next delivery) bool {
	payload := next.payload
	srcQueue := next.sourceQueue
	job, err := queue.UnmarshalJob(payload)
	if err != nil {
		w.log.Error("invalid job payload", obs.Err(err))
		if next.ordered != nil {
			_, _ = queue.TransitionOrdered(ctx, w.rdb, w.cfg.OrderingLayout(), *next.ordered, procList, hbKey, workerID, "", "", queue.OrderedDiscard)
		} else {
			// remove from processing to avoid poison pill loop
			_ = w.rdb.LRem(ctx, procList, 1, payload).Err()
			_ = w.rdb.Del(ctx, hbKey).Err()
		}
		return false
	}
	// Start span with job's TraceID/SpanID when available
	ctx, span := obs.ContextWithJobSpan(ctx, job)
	defer span.End()

	// Add worker and queue attributes
	obs.AddSpanAttributes(ctx,
		obs.KeyValue("worker.id", workerID),
		obs.KeyValue("queue.source", srcQueue),
		obs.KeyValue("processing.list", procList),
	)

	// Add processing started event
	obs.AddEvent(ctx, "job.processing.started",
		obs.KeyValue("job.id", job.ID),
		obs.KeyValue("worker.id", workerID),
	)

	if ctx.Err() != nil {
		w.log.Info("job interrupted before handler; leaving in processing for reaper",
			obs.String("id", job.ID),
			obs.String("worker_id", workerID),
		)
		return false
	}

	leaseKey := ""
	if next.ordered != nil {
		leaseKey = next.ordered.LeaseKey
	}
	handlerCtx, stopHeartbeat := w.maintainHeartbeat(ctx, hbKey, payload, leaseKey, workerID)
	defer stopHeartbeat()
	if handlerCtx.Err() != nil {
		w.log.Info("job interrupted before handler; leaving in processing for reaper",
			obs.String("id", job.ID),
			obs.String("worker_id", workerID),
		)
		return false
	}
	processingStart := time.Now()
	handlerErr := w.invokeHandler(handlerCtx, job)
	processingDuration := time.Since(processingStart)
	obs.AddSpanAttributes(ctx, obs.KeyValue("processing.duration_ms", processingDuration.Milliseconds()))

	if handlerCtx.Err() != nil {
		obs.AddEvent(ctx, "job.processing.interrupted",
			obs.KeyValue("job.id", job.ID),
			obs.KeyValue("reason", handlerCtx.Err().Error()),
		)
		w.log.Info("job interrupted; leaving in processing for reaper",
			obs.String("id", job.ID),
			obs.String("worker_id", workerID),
		)
		return false
	}

	if handlerErr == nil {
		stopHeartbeat()
		if next.ordered != nil {
			transitioned, transitionErr := queue.TransitionOrdered(
				ctx,
				w.rdb,
				w.cfg.OrderingLayout(),
				*next.ordered,
				procList,
				hbKey,
				workerID,
				w.cfg.Worker.CompletedList,
				payload,
				queue.OrderedComplete,
			)
			if transitionErr != nil {
				w.log.Error("ordered completion failed", obs.Err(transitionErr))
				obs.RecordError(ctx, transitionErr)
				return false
			}
			if !transitioned {
				w.log.Warn("ordered completion lost lease; leaving job for reaper",
					obs.String("id", job.ID),
					obs.String("worker_id", workerID),
				)
				return false
			}
		}
		// Mark span as successful
		obs.SetSpanSuccess(ctx)
		obs.AddEvent(ctx, "job.processing.completed",
			obs.KeyValue("job.id", job.ID),
			obs.KeyValue("duration_ms", processingDuration.Milliseconds()),
		)

		if next.ordered == nil {
			// complete
			if err := w.rdb.LPush(ctx, w.cfg.Worker.CompletedList, payload).Err(); err != nil {
				w.log.Error("LPUSH completed failed", obs.Err(err))
				obs.RecordError(ctx, err)
			}
			if err := w.rdb.LRem(ctx, procList, 1, payload).Err(); err != nil {
				w.log.Error("LREM processing failed", obs.Err(err))
			}
			if err := w.rdb.Del(ctx, hbKey).Err(); err != nil {
				w.log.Error("DEL heartbeat failed", obs.Err(err))
			}
		}
		obs.JobsCompleted.Inc()
		w.log.Info("job completed", obs.String("id", job.ID), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
		return true
	}

	// failure path with retry
	obs.JobsFailed.Inc()

	// Record failure in span
	failureReason := handlerErr.Error()
	obs.RecordError(ctx, handlerErr)
	obs.AddEvent(ctx, "job.processing.failed",
		obs.KeyValue("job.id", job.ID),
		obs.KeyValue("reason", failureReason),
		obs.KeyValue("retries", job.Retries),
	)

	job.Retries++
	// backoff
	bo := backoff(job.Retries, w.cfg.Worker.Backoff.Base, w.cfg.Worker.Backoff.Max)
	timer := time.NewTimer(bo)
	defer timer.Stop()
	select {
	case <-handlerCtx.Done():
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		w.log.Info("retry interrupted; leaving job in processing for reaper",
			obs.String("id", job.ID),
			obs.String("worker_id", workerID),
		)
		return false
	case <-timer.C:
	}

	if job.Retries <= w.cfg.Worker.MaxRetries {
		stopHeartbeat()
		payload2, marshalErr := job.Marshal()
		if marshalErr != nil {
			w.log.Error("marshal retry failed", obs.Err(marshalErr))
			return false
		}
		if next.ordered != nil {
			transitioned, transitionErr := queue.TransitionOrdered(
				ctx,
				w.rdb,
				w.cfg.OrderingLayout(),
				*next.ordered,
				procList,
				hbKey,
				workerID,
				"",
				payload2,
				queue.OrderedRetry,
			)
			if transitionErr != nil {
				w.log.Error("ordered retry failed", obs.Err(transitionErr))
				obs.RecordError(ctx, transitionErr)
				return false
			}
			if !transitioned {
				w.log.Warn("ordered retry lost lease; leaving job for reaper",
					obs.String("id", job.ID),
					obs.String("worker_id", workerID),
				)
				return false
			}
		} else {
			if err := w.rdb.LPush(ctx, srcQueue, payload2).Err(); err != nil {
				w.log.Error("LPUSH retry failed", obs.Err(err))
				obs.RecordError(ctx, err)
			}
			if err := w.rdb.LRem(ctx, procList, 1, payload).Err(); err != nil {
				w.log.Error("LREM processing failed", obs.Err(err))
			}
			if err := w.rdb.Del(ctx, hbKey).Err(); err != nil {
				w.log.Error("DEL heartbeat failed", obs.Err(err))
			}
		}
		obs.JobsRetried.Inc()
		obs.AddEvent(ctx, "job.retrying",
			obs.KeyValue("job.id", job.ID),
			obs.KeyValue("retry_count", job.Retries),
			obs.KeyValue("backoff_ms", bo.Milliseconds()),
		)
		w.log.Warn("job retried", obs.String("id", job.ID), obs.Int("retries", job.Retries), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
		return false
	}

	// dead letter
	stopHeartbeat()
	obs.AddEvent(ctx, "job.dead_lettered",
		obs.KeyValue("job.id", job.ID),
		obs.KeyValue("max_retries_exceeded", true),
	)

	if next.ordered != nil {
		transitioned, transitionErr := queue.TransitionOrdered(
			ctx,
			w.rdb,
			w.cfg.OrderingLayout(),
			*next.ordered,
			procList,
			hbKey,
			workerID,
			w.cfg.Worker.DeadLetterList,
			payload,
			queue.OrderedDeadLetter,
		)
		if transitionErr != nil {
			w.log.Error("ordered dead-letter failed", obs.Err(transitionErr))
			obs.RecordError(ctx, transitionErr)
			return false
		}
		if !transitioned {
			w.log.Warn("ordered dead-letter lost lease; leaving job for reaper",
				obs.String("id", job.ID),
				obs.String("worker_id", workerID),
			)
			return false
		}
	} else {
		if err := w.rdb.LPush(ctx, w.cfg.Worker.DeadLetterList, payload).Err(); err != nil {
			w.log.Error("LPUSH DLQ failed", obs.Err(err))
			obs.RecordError(ctx, err)
		}
		if err := w.rdb.LRem(ctx, procList, 1, payload).Err(); err != nil {
			w.log.Error("LREM processing failed", obs.Err(err))
		}
		if err := w.rdb.Del(ctx, hbKey).Err(); err != nil {
			w.log.Error("DEL heartbeat failed", obs.Err(err))
		}
	}
	obs.JobsDeadLetter.Inc()
	w.log.Error("job dead-lettered", obs.String("id", job.ID), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
	return false
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func backoff(retries int, base, max time.Duration) time.Duration {
	d := time.Duration(1<<uint(retries-1)) * base
	if d > max {
		return max
	}
	if d < 0 {
		return max
	}
	return d
}
