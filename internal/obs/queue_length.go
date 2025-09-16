// Copyright 2025 James Ross
package obs

import (
	"context"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// StartQueueLengthUpdater samples queue lengths and updates a gauge.
func StartQueueLengthUpdater(ctx context.Context, cfg *config.Config, rdb *redis.Client, log *zap.Logger) {
	interval := 2 * time.Second
	if cfg.Observability.QueueSampleInterval > 0 {
		interval = cfg.Observability.QueueSampleInterval
	}
	// Build set of queues to poll
	qset := map[string]struct{}{}
	for _, q := range cfg.Worker.Queues {
		qset[q] = struct{}{}
	}
	qset[cfg.Worker.CompletedList] = struct{}{}
	qset[cfg.Worker.DeadLetterList] = struct{}{}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for q := range qset {
					n, err := rdb.LLen(ctx, q).Result()
					if err != nil {
						log.Debug("queue length poll error", String("queue", q), Err(err))
						continue
					}
					QueueLength.WithLabelValues(q).Set(float64(n))
				}
			}
		}
	}()
}
