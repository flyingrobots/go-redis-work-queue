// Copyright 2025 James Ross
package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if cfg.Queue.MaxPayloadSize != 1<<20 {
		t.Fatalf("expected default max payload size 1 MiB, got %d", cfg.Queue.MaxPayloadSize)
	}
	if err := cfg.OrderingLayout().Validate(); err != nil {
		t.Fatalf("default ordering layout is invalid: %v", err)
	}
}

func TestLoadRejectsUnsupportedExactlyOnceConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte("exactly_once:\n  idempotency:\n    enabled: false\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unsupported exactly_once config to be rejected")
	}
	if !strings.Contains(err.Error(), "exactly_once") {
		t.Fatalf("expected error to identify exactly_once, got %q", err)
	}
}

func TestExampleConfigContainsOnlySupportedKeys(t *testing.T) {
	path := filepath.Join("..", "..", "config", "config.example.yaml")
	if _, err := Load(path); err != nil {
		t.Fatalf("example config contains an unsupported key: %v", err)
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
	cfg = defaultConfig()
	cfg.Queue.MaxPayloadSize = 0
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected error for queue.max_payload_size <= 0")
	}
	cfg = defaultConfig()
	cfg.Worker.ProcessingListPattern = "jobqueue:processing-without-placeholder"
	if err := Validate(cfg); err == nil {
		t.Fatal("expected processing pattern without a placeholder to fail")
	}
	cfg = defaultConfig()
	cfg.Worker.HeartbeatKeyPattern = "jobqueue:%s:heartbeat:%s"
	if err := Validate(cfg); err == nil {
		t.Fatal("expected heartbeat pattern with two placeholders to fail")
	}
	cfg = defaultConfig()
	cfg.Queue.OrderedQueuePattern = "jobqueue:ordered:without-placeholder"
	if err := Validate(cfg); err == nil {
		t.Fatal("expected ordered queue pattern without a placeholder to fail")
	}
	cfg = defaultConfig()
	cfg.Queue.OrderedLeasePattern = "jobqueue:%s:ordered:%s"
	if err := Validate(cfg); err == nil {
		t.Fatal("expected ordered lease pattern with two placeholders to fail")
	}
	cfg = defaultConfig()
	cfg.Queue.OrderedLeasePattern = cfg.Queue.OrderedQueuePattern
	if err := Validate(cfg); err == nil {
		t.Fatal("expected identical ordered queue and lease patterns to fail")
	}
}
