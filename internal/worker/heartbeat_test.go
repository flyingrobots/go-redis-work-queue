// Copyright 2026 James Ross
package worker

import (
	"context"
	"strings"
	"testing"
	"time"
)

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
