package admin

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "math"
    "sort"
    "strings"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/go-redis/redis/v8"
)

type StatsResult struct {
    Queues map[string]int64 `json:"queues"`
    ProcessingLists map[string]int64 `json:"processing_lists"`
    Heartbeats int64 `json:"heartbeats"`
}

func Stats(ctx context.Context, cfg *config.Config, rdb *redis.Client) (StatsResult, error) {
    res := StatsResult{Queues: map[string]int64{}, ProcessingLists: map[string]int64{}}
    // Count standard queues
    qset := map[string]string{}
    for p, q := range cfg.Worker.Queues { qset[p] = q }
    qset["completed"] = cfg.Worker.CompletedList
    qset["dead_letter"] = cfg.Worker.DeadLetterList
    for name, key := range qset {
        n, err := rdb.LLen(ctx, key).Result()
        if err != nil { return res, err }
        res.Queues[name+"("+key+")"] = n
    }
    // Scan processing lists
    var cursor uint64
    for {
        keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:worker:*:processing", 200).Result()
        if err != nil { return res, err }
        cursor = cur
        for _, k := range keys {
            n, _ := rdb.LLen(ctx, k).Result()
            res.ProcessingLists[k] = n
        }
        if cursor == 0 { break }
    }
    // Heartbeats
    var hbc int64
    cursor = 0
    for {
        keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:processing:worker:*", 500).Result()
        if err != nil { return res, err }
        cursor = cur
        hbc += int64(len(keys))
        if cursor == 0 { break }
    }
    res.Heartbeats = hbc
    return res, nil
}

type PeekResult struct {
    Queue string   `json:"queue"`
    Items []string `json:"items"`
}

func Peek(ctx context.Context, cfg *config.Config, rdb *redis.Client, queueAlias string, n int64) (PeekResult, error) {
    qkey, err := resolveQueue(cfg, queueAlias)
    if err != nil { return PeekResult{}, err }
    if n <= 0 { n = 10 }
    // Items to be consumed next are at the right end; take last N
    items, err := rdb.LRange(ctx, qkey, -n, -1).Result()
    if err != nil { return PeekResult{}, err }
    return PeekResult{Queue: qkey, Items: items}, nil
}

func PurgeDLQ(ctx context.Context, cfg *config.Config, rdb *redis.Client) error {
    if cfg.Worker.DeadLetterList == "" { return errors.New("dead letter list not configured") }
    return rdb.Del(ctx, cfg.Worker.DeadLetterList).Err()
}

func resolveQueue(cfg *config.Config, alias string) (string, error) {
    a := strings.ToLower(alias)
    if a == "completed" { return cfg.Worker.CompletedList, nil }
    if a == "dead_letter" || a == "dlq" { return cfg.Worker.DeadLetterList, nil }
    if q, ok := cfg.Worker.Queues[a]; ok { return q, nil }
    // Otherwise, assume full key
    if strings.HasPrefix(alias, "jobqueue:") { return alias, nil }
    // Suggest options
    keys := make([]string, 0, len(cfg.Worker.Queues))
    for k := range cfg.Worker.Queues { keys = append(keys, k) }
    sort.Strings(keys)
    b, _ := json.Marshal(keys)
    return "", fmt.Errorf("unknown queue alias %q; known: %s, completed, dead_letter or full key starting with jobqueue:", alias, string(b))
}

type BenchResult struct {
    Count      int           `json:"count"`
    Duration   time.Duration `json:"duration"`
    Throughput float64       `json:"throughput_jobs_per_sec"`
    P50        time.Duration `json:"p50_latency"`
    P95        time.Duration `json:"p95_latency"`
}

// Bench enqueues count jobs to the chosen queue and waits for completion
// (observing the completed list) up to timeout. It computes simple latency
// stats using job creation_time vs. measurement time.
func Bench(ctx context.Context, cfg *config.Config, rdb *redis.Client, priority string, count int, rate int, timeout time.Duration) (BenchResult, error) {
    res := BenchResult{Count: count}
    if count <= 0 { return res, fmt.Errorf("count must be > 0") }
    if rate <= 0 { rate = 100 }
    qkey, err := resolveQueue(cfg, priority)
    if err != nil { return res, err }
    // Clear completed
    _ = rdb.Del(ctx, cfg.Worker.CompletedList).Err()

    // Enqueue
    ticker := time.NewTicker(time.Second / time.Duration(rate))
    defer ticker.Stop()
    start := time.Now()
    for i := 0; i < count; i++ {
        select { case <-ctx.Done(): return res, ctx.Err(); case <-ticker.C: }
        payload := fmt.Sprintf(`{"id":"bench-%d","filepath":"/bench/%d","filesize":1,"priority":"%s","retries":0,"creation_time":"%s","trace_id":"","span_id":""}`,
            i, i, priority, time.Now().UTC().Format(time.RFC3339Nano))
        if err := rdb.LPush(ctx, qkey, payload).Err(); err != nil { return res, err }
    }

    // Wait for completion
    doneBy := time.Now().Add(timeout)
    for time.Now().Before(doneBy) {
        n, _ := rdb.LLen(ctx, cfg.Worker.CompletedList).Result()
        if int(n) >= count { break }
        time.Sleep(50 * time.Millisecond)
    }
    res.Duration = time.Since(start)
    if res.Duration > 0 { res.Throughput = float64(count) / res.Duration.Seconds() }

    // Fetch and compute latencies
    items, _ := rdb.LRange(ctx, cfg.Worker.CompletedList, 0, -1).Result()
    lats := make([]float64, 0, len(items))
    now := time.Now()
    for _, it := range items {
        var j struct{ CreationTime string `json:"creation_time"` }
        if err := json.Unmarshal([]byte(it), &j); err == nil {
            if t, err2 := time.Parse(time.RFC3339Nano, j.CreationTime); err2 == nil {
                lats = append(lats, now.Sub(t).Seconds())
            }
        }
    }
    if len(lats) > 0 {
        sort.Float64s(lats)
        res.P50 = time.Duration(lats[int(math.Round(0.50*float64(len(lats)-1)))]*float64(time.Second))
        res.P95 = time.Duration(lats[int(math.Round(0.95*float64(len(lats)-1)))]*float64(time.Second))
    }
    return res, nil
}
