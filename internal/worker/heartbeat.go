// Copyright 2026 James Ross
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
)

// maintainHeartbeat writes the initial ownership marker and renews it while a
// handler or its retry backoff is running. The returned stop function waits for
// the renewal goroutine, preventing a late SET from racing with cleanup.
func (w *Worker) maintainHeartbeat(ctx context.Context, key, payload string) func() {
	ttl := w.cfg.Worker.HeartbeatTTL
	if ttl <= 0 {
		ttl = time.Second
	}
	interval := ttl / 3
	if interval <= 0 {
		interval = time.Millisecond
	}

	heartbeatCtx, cancel := context.WithCancel(ctx)
	set := func() {
		if err := w.rdb.Set(heartbeatCtx, key, payload, ttl).Err(); err != nil && heartbeatCtx.Err() == nil {
			w.log.Warn("heartbeat refresh failed", obs.String("key", key), obs.Err(err))
		}
	}
	set()

	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				set()
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}
}
