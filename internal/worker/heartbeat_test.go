// Copyright 2026 James Ross
package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type failHeartbeatRefreshOnceHook struct {
	key      string
	setCount atomic.Int32
}

func (*failHeartbeatRefreshOnceHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *failHeartbeatRefreshOnceHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		args := cmd.Args()
		if cmd.Name() == "set" && len(args) > 1 && fmt.Sprint(args[1]) == h.key && h.setCount.Add(1) == 2 {
			return errors.New("injected heartbeat refresh failure")
		}
		return next(ctx, cmd)
	}
}

func (*failHeartbeatRefreshOnceHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestMaintainHeartbeatStoresOwnershipMarkerInsteadOfPayload(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	cfg.Worker.HeartbeatTTL = time.Second
	const owner = "worker-owner"
	key := "test:heartbeat:worker-owner"
	payload := strings.Repeat("x", 1<<20)

	ctx, cancel := context.WithCancel(context.Background())
	_, stop := w.maintainHeartbeat(ctx, key, "", owner)
	t.Cleanup(func() {
		cancel()
		stop()
	})
	value, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if value == payload {
		t.Fatalf("heartbeat duplicated the %d-byte job envelope", len(payload))
	}
	if value != owner {
		t.Fatalf("heartbeat value = %q, want ownership marker %q", value, owner)
	}
}

func TestMaintainHeartbeatCancelsHandlerAfterRefreshFailure(t *testing.T) {
	w, cfg, rdb := newHandlerTestWorker(t, 1, 0, nil)
	cfg.Worker.HeartbeatTTL = 30 * time.Millisecond
	const owner = "worker-owner"
	const key = "test:heartbeat:refresh-failure"
	hook := &failHeartbeatRefreshOnceHook{key: key}
	rdb.AddHook(hook)

	ctx, cancel := context.WithCancel(context.Background())
	handlerCtx, stop := w.maintainHeartbeat(ctx, key, "", owner)
	t.Cleanup(func() {
		cancel()
		stop()
	})
	select {
	case <-handlerCtx.Done():
		if !errors.Is(handlerCtx.Err(), context.Canceled) {
			t.Fatalf("handler context error = %v, want context.Canceled", handlerCtx.Err())
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handler remained active after heartbeat refresh failed")
	}
}
