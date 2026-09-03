// Copyright 2025 James Ross
package reaper

import (
	"context"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Reaper struct {
	cfg *config.Config
	rdb *redis.Client
	log *zap.Logger
}

func New(cfg *config.Config, rdb *redis.Client, log *zap.Logger) *Reaper {
	return &Reaper{cfg: cfg, rdb: rdb, log: log}
}

func (r *Reaper) Run(ctx context.Context) {
	ticker := time.NewTicker(scanInterval(r.cfg.Worker.HeartbeatTTL))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.scanOnce(ctx)
		}
	}
}

func (r *Reaper) scanOnce(ctx context.Context) {
	// Scan all processing lists
	processingPattern := r.cfg.Worker.ProcessingListPattern
	if processingPattern == "" {
		processingPattern = queuekeys.DefaultProcessingListPattern
	}
	var cursor uint64
	for {
		keys, cur, err := r.rdb.Scan(ctx, cursor, queuekeys.ScanPattern(processingPattern), 100).Result()
		if err != nil {
			r.log.Warn("reaper scan error", obs.Err(err))
			return
		}
		cursor = cur
		for _, plist := range keys {
			workerID, ok := queuekeys.Extract(processingPattern, plist)
			if !ok || workerID == "" {
				continue
			}
			hbPattern := r.cfg.Worker.HeartbeatKeyPattern
			if hbPattern == "" {
				hbPattern = queuekeys.DefaultHeartbeatKeyPattern
			}
			hbKey := queuekeys.Format(hbPattern, workerID)
			exists, _ := r.rdb.Exists(ctx, hbKey).Result()
			if exists == 1 {
				continue
			} // worker healthy

			// Requeue all jobs from processing list. Ordered jobs remain in
			// place until the compare-owned lease recovery script wins.
			for {
				payload, err := r.rdb.LIndex(ctx, plist, -1).Result()
				if err == redis.Nil {
					break
				}
				if err != nil {
					r.log.Warn("reaper rpop error", obs.Err(err))
					break
				}
				job, err := queue.UnmarshalJob(payload)
				if err != nil {
					_ = r.rdb.RPop(ctx, plist).Err()
					continue
				}
				if job.OrderingKey != "" {
					delivery := queue.OrderedDeliveryFor(r.cfg.OrderingLayout(), job, payload)
					recovered, recoverErr := queue.RecoverOrdered(
						ctx,
						r.rdb,
						r.cfg.OrderingLayout(),
						delivery,
						plist,
						hbKey,
					)
					if recoverErr != nil {
						r.log.Error("ordered recovery failed", obs.Err(recoverErr))
						break
					}
					if !recovered {
						break
					}
					obs.ReaperRecovered.Inc()
					r.log.Warn("requeued abandoned ordered job",
						obs.String("id", job.ID),
						obs.String("to", delivery.QueueKey),
						obs.String("trace_id", job.TraceID),
						obs.String("span_id", job.SpanID),
					)
					continue
				}
				popped, err := r.rdb.RPop(ctx, plist).Result()
				if err == redis.Nil {
					break
				}
				if err != nil {
					r.log.Warn("reaper rpop error", obs.Err(err))
					break
				}
				payload = popped
				job, err = queue.UnmarshalJob(payload)
				if err != nil {
					continue
				}
				prio := job.Priority
				dest := r.cfg.Worker.Queues[prio]
				if dest == "" {
					dest = r.cfg.Worker.Queues[r.cfg.Producer.DefaultPriority]
				}
				if err := r.rdb.LPush(ctx, dest, payload).Err(); err != nil {
					r.log.Error("requeue failed", obs.Err(err))
				} else {
					obs.ReaperRecovered.Inc()
					r.log.Warn("requeued abandoned job", obs.String("id", job.ID), obs.String("to", dest), obs.String("trace_id", job.TraceID), obs.String("span_id", job.SpanID))
				}
			}
		}
		if cursor == 0 {
			break
		}
	}
}

func scanInterval(heartbeatTTL time.Duration) time.Duration {
	interval := heartbeatTTL / 4
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}
	if interval > 5*time.Second {
		return 5 * time.Second
	}
	return interval
}
