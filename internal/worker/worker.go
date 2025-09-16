package worker

import (
    "context"
    "fmt"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/breaker"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/obs"
    "github.com/flyingrobots/go-redis-work-queue/internal/queue"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type Worker struct {
    cfg *config.Config
    rdb *redis.Client
    log *zap.Logger
    cb  *breaker.CircuitBreaker
}

func New(cfg *config.Config, rdb *redis.Client, log *zap.Logger) *Worker {
    cb := breaker.New(cfg.CircuitBreaker.Window, cfg.CircuitBreaker.CooldownPeriod, cfg.CircuitBreaker.FailureThreshold, cfg.CircuitBreaker.MinSamples)
    return &Worker{cfg: cfg, rdb: rdb, log: log, cb: cb}
}

func (w *Worker) Run(ctx context.Context) error {
    var wg sync.WaitGroup
    host, _ := os.Hostname()
    pid := os.Getpid()
    for i := 0; i < w.cfg.Worker.Count; i++ {
        wg.Add(1)
        id := fmt.Sprintf("%s-%d-%d", host, pid, i)
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
    procList := fmt.Sprintf(w.cfg.Worker.ProcessingListPattern, workerID)
    hbKey := fmt.Sprintf(w.cfg.Worker.HeartbeatKeyPattern, workerID)

    for ctx.Err() == nil {
        if !w.cb.Allow() {
            time.Sleep(100 * time.Millisecond)
            continue
        }

        // fetch by priority using BRPOPLPUSH with short timeout
        var payload string
        var srcQueue string
        for _, p := range w.cfg.Worker.Priorities {
            key := w.cfg.Worker.Queues[p]
            if key == "" { continue }
            v, err := w.rdb.BRPopLPush(ctx, key, procList, w.cfg.Worker.BRPopLPushTimeout).Result()
            if err == redis.Nil {
                continue
            }
            if err != nil {
                if ctx.Err() != nil { return }
                w.log.Warn("BRPOPLPUSH error", obs.Err(err))
                time.Sleep(50 * time.Millisecond)
                continue
            }
            payload = v
            srcQueue = key
            break
        }
        if payload == "" {
            continue // timeout across all priorities
        }

        obs.JobsConsumed.Inc()
        // heartbeat set
        _ = w.rdb.Set(ctx, hbKey, payload, w.cfg.Worker.HeartbeatTTL).Err()

        start := time.Now()
        // process job
        ok := w.processJob(ctx, workerID, srcQueue, procList, hbKey, payload)
        obs.JobProcessingDuration.Observe(time.Since(start).Seconds())
        w.cb.Record(ok)
    }
}

func (w *Worker) processJob(ctx context.Context, workerID, srcQueue, procList, hbKey, payload string) bool {
    job, err := queue.UnmarshalJob(payload)
    if err != nil {
        w.log.Error("invalid job payload", obs.Err(err))
        // remove from processing to avoid poison pill loop
        _ = w.rdb.LRem(ctx, procList, 1, payload).Err()
        _ = w.rdb.Del(ctx, hbKey).Err()
        return false
    }
    // Start span with job's TraceID/SpanID when available
    ctx, span := obs.ContextWithJobSpan(ctx, job)
    defer span.End()

    // Simulated processing: sleep based on filesize
    dur := time.Duration(min64(job.FileSize/1024, 1000)) * time.Millisecond
    canceled := false
    select {
    case <-ctx.Done():
        canceled = true
    case <-time.After(dur):
    }

    // For demonstration, consider processing success unless canceled or filename contains "fail"
    success := !canceled && !strings.Contains(strings.ToLower(job.FilePath), "fail")

    if success {
        // complete
        _ = w.rdb.LPush(ctx, w.cfg.Worker.CompletedList, payload).Err()
        _ = w.rdb.LRem(ctx, procList, 1, payload).Err()
        _ = w.rdb.Del(ctx, hbKey).Err()
        obs.JobsCompleted.Inc()
        w.log.Info("job completed", obs.String("id", job.ID), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
        return true
    }

    // failure path with retry
    obs.JobsFailed.Inc()
    job.Retries++
    // backoff
    bo := backoff(job.Retries, w.cfg.Worker.Backoff.Base, w.cfg.Worker.Backoff.Max)
    select {
    case <-ctx.Done():
    case <-time.After(bo):
    }

    if job.Retries <= w.cfg.Worker.MaxRetries {
        obs.JobsRetried.Inc()
        payload2, _ := job.Marshal()
        _ = w.rdb.LPush(ctx, srcQueue, payload2).Err()
        _ = w.rdb.LRem(ctx, procList, 1, payload).Err()
        _ = w.rdb.Del(ctx, hbKey).Err()
        w.log.Warn("job retried", obs.String("id", job.ID), obs.Int("retries", job.Retries), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
        return false
    }

    // dead letter
    _ = w.rdb.LPush(ctx, w.cfg.Worker.DeadLetterList, payload).Err()
    _ = w.rdb.LRem(ctx, procList, 1, payload).Err()
    _ = w.rdb.Del(ctx, hbKey).Err()
    obs.JobsDeadLetter.Inc()
    w.log.Error("job dead-lettered", obs.String("id", job.ID), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID), obs.String("worker_id", workerID))
    return false
}

func min64(a, b int64) int64 { if a < b { return a }; return b }

func backoff(retries int, base, max time.Duration) time.Duration {
    d := time.Duration(1<<uint(retries-1)) * base
    if d > max { return max }
    if d < 0 { return max }
    return d
}
