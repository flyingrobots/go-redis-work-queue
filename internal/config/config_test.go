package config

import (
    "os"
    "testing"
)

func TestLoadDefaults(t *testing.T) {
    os.Unsetenv("WORKER_COUNT")
    cfg, err := Load("nonexistent.yaml")
    if err != nil { t.Fatal(err) }
    if cfg.Worker.Count != 16 { t.Fatalf("expected default worker count 16, got %d", cfg.Worker.Count) }
    if cfg.Redis.Addr == "" { t.Fatalf("expected default redis addr") }
}

