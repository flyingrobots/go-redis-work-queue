package producer

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/bmatcuk/doublestar/v4"
    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/obs"
    "github.com/flyingrobots/go-redis-work-queue/internal/queue"
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

type Producer struct {
    cfg *config.Config
    rdb *redis.Client
    log *zap.Logger
}

func New(cfg *config.Config, rdb *redis.Client, log *zap.Logger) *Producer {
    return &Producer{cfg: cfg, rdb: rdb, log: log}
}

func (p *Producer) Run(ctx context.Context) error {
    root := p.cfg.Producer.ScanDir
    include := p.cfg.Producer.IncludeGlobs
    exclude := p.cfg.Producer.ExcludeGlobs

    var files []string
    err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        rel, _ := filepath.Rel(root, path)
        // include check
        incMatch := len(include) == 0
        for _, g := range include { if ok, _ := doublestar.PathMatch(g, rel); ok { incMatch = true; break } }
        if !incMatch { return nil }
        for _, g := range exclude { if ok, _ := doublestar.PathMatch(g, rel); ok { return nil } }
        files = append(files, path)
        return nil
    })
    if err != nil {
        return err
    }

    for _, f := range files {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        if err := p.rateLimit(ctx); err != nil { return err }
        fi, err := os.Stat(f)
        if err != nil { continue }
        prio := p.priorityForExt(filepath.Ext(f))
        id := randID()
        traceID, spanID := randTraceAndSpan()
        j := queue.NewJob(id, f, fi.Size(), prio, traceID, spanID)
        payload, _ := j.Marshal()
        key := p.cfg.Worker.Queues[prio]
        if key == "" { key = p.cfg.Worker.Queues[p.cfg.Producer.DefaultPriority] }
        if err := p.rdb.LPush(ctx, key, payload).Err(); err != nil {
            return err
        }
        obs.JobsProduced.Inc()
        p.log.Info("enqueued job", obs.String("id", j.ID), obs.String("queue", key))
    }
    return nil
}

func (p *Producer) priorityForExt(ext string) string {
    ext = strings.ToLower(ext)
    for _, e := range p.cfg.Producer.HighPriorityExts {
        if strings.ToLower(e) == ext {
            return "high"
        }
    }
    return p.cfg.Producer.DefaultPriority
}

func (p *Producer) rateLimit(ctx context.Context) error {
    if p.cfg.Producer.RateLimitPerSec <= 0 { return nil }
    key := p.cfg.Producer.RateLimitKey
    // Fixed-window limiter
    n, err := p.rdb.Incr(ctx, key).Result()
    if err != nil { return err }
    if n == 1 {
        _ = p.rdb.Expire(ctx, key, time.Second).Err()
    }
    if int(n) > p.cfg.Producer.RateLimitPerSec {
        // Sleep until window likely reset
        time.Sleep(250 * time.Millisecond)
    }
    return nil
}

func randID() string {
    var b [16]byte
    _, _ = rand.Read(b[:])
    return hex.EncodeToString(b[:])
}

func randTraceAndSpan() (string, string) {
    var tb [16]byte
    var sb [8]byte
    _, _ = rand.Read(tb[:])
    _, _ = rand.Read(sb[:])
    return hex.EncodeToString(tb[:]), hex.EncodeToString(sb[:])
}
