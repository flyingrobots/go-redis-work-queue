// Copyright 2026 James Ross
package queueworker

import (
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestNewWithClientRequiresApplicationHandler(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	worker, err := NewWithClient(rdb, DefaultConfig(), nil, nil)
	if worker != nil {
		t.Fatal("constructor returned a worker without a handler")
	}
	if !errors.Is(err, ErrHandlerRequired) {
		t.Fatalf("constructor error = %v, want ErrHandlerRequired", err)
	}
}

func TestDefaultConfigReturnsIndependentQueueMaps(t *testing.T) {
	first := DefaultConfig()
	second := DefaultConfig()
	first.QueueConfig.Queues["low"] = "changed"
	if second.QueueConfig.Queues["low"] == "changed" {
		t.Fatal("DefaultConfig shared its queue map")
	}
}
