package reaper

import (
    "context"
    "fmt"
    "testing"

    "github.com/alicebob/miniredis/v2"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/queue"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

func TestReaperRequeuesWithoutHeartbeat(t *testing.T) {
    mr, _ := miniredis.Run()
    defer mr.Close()
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    cfg, err := config.Load("nonexistent.yaml")
    if err != nil { t.Fatal(err) }
    cfg.Redis.Addr = mr.Addr()
    log, _ := zap.NewDevelopment()
    rep := New(cfg, rdb, log)

    ctx := context.Background()
    workerID := "w1"
    plist := fmt.Sprintf(cfg.Worker.ProcessingListPattern, workerID)
    hbKey := fmt.Sprintf(cfg.Worker.HeartbeatKeyPattern, workerID)
    // Simulate dead worker: no heartbeat key
    job := queue.NewJob("id1", "/tmp/file.txt", 10, "low", "", "")
    payload, _ := job.Marshal()
    if err := rdb.LPush(ctx, plist, payload).Err(); err != nil { t.Fatal(err) }

    rep.scanOnce(ctx)

    // Expect job moved back to low priority queue
    n, _ := rdb.LLen(context.Background(), cfg.Worker.Queues["low"]).Result()
    if n != 1 {
        t.Fatalf("expected 1 job in low queue, got %d", n)
    }
    if mr.Exists(hbKey) {
        t.Fatalf("heartbeat should not exist")
    }
}
