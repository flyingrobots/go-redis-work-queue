// Copyright 2025 James Ross
package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("WORKER_COUNT")
	cfg, err := Load("nonexistent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Worker.Count != 16 {
		t.Fatalf("expected default worker count 16, got %d", cfg.Worker.Count)
	}
	if cfg.Redis.Addr == "" {
		t.Fatalf("expected default redis addr")
	}
}

func TestValidateFails(t *testing.T) {
	cfg := defaultConfig()
	cfg.Worker.Count = 0
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected error for worker.count < 1")
	}
	cfg = defaultConfig()
	cfg.Worker.HeartbeatTTL = 3 * 1e9 // 3s
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected error for heartbeat ttl < 5s")
	}
	cfg = defaultConfig()
	cfg.Worker.BRPopLPushTimeout = cfg.Worker.HeartbeatTTL
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected error for brpoplpush_timeout > heartbeat_ttl/2")
	}
}
