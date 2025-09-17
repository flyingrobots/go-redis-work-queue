// Copyright 2025 James Ross
package reaper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
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
	ticker := time.NewTicker(5 * time.Second)
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
	var cursor uint64
	for {
		keys, cur, err := r.rdb.Scan(ctx, cursor, "jobqueue:worker:*:processing", 100).Result()
		if err != nil {
			r.log.Warn("reaper scan error", obs.Err(err))
			return
		}
		cursor = cur
		for _, plist := range keys {
			// derive worker id
			// jobqueue:worker:<ID>:processing
			parts := strings.Split(plist, ":")
			if len(parts) < 4 {
				continue
			}
			workerID := parts[2]
			hbKey := fmt.Sprintf(r.cfg.Worker.HeartbeatKeyPattern, workerID)
			exists, _ := r.rdb.Exists(ctx, hbKey).Result()
			if exists == 1 {
				continue
			} // worker healthy

			// Requeue all jobs from processing list
			for {
				payload, err := r.rdb.RPop(ctx, plist).Result()
				if err == redis.Nil {
					break
				}
				if err != nil {
					r.log.Warn("reaper rpop error", obs.Err(err))
					break
				}
				job, err := queue.UnmarshalJob(payload)
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
