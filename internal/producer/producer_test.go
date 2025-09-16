package producer

import (
    "context"
    "testing"
    "time"

    "github.com/alicebob/miniredis/v2"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

func TestPriorityForExt(t *testing.T) {
    p := &Producer{cfg: &config.Config{Producer: config.Producer{DefaultPriority: "low", HighPriorityExts: []string{".pdf"}}}}
    if got := p.priorityForExt(".pdf"); got != "high" { t.Fatalf("expected high, got %s", got) }
    if got := p.priorityForExt(".txt"); got != "low" { t.Fatalf("expected low, got %s", got) }
}

func TestRateLimit(t *testing.T) {
    mr, _ := miniredis.Run()
    defer mr.Close()
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    cfg := &config.Config{Producer: config.Producer{RateLimitPerSec: 1, RateLimitKey: "rl"}}
    log, _ := zap.NewDevelopment()
    p := New(cfg, rdb, log)
    if err := p.rateLimit(context.Background()); err != nil { t.Fatal(err) }
    start := time.Now()
    if err := p.rateLimit(context.Background()); err != nil { t.Fatal(err) } // second call exceeds limit; will sleep ~ttl
    if time.Since(start) < 100*time.Millisecond {
        t.Fatalf("expected limiter to sleep when exceeded")
    }
}
