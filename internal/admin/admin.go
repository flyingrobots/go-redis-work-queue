package admin

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strings"

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

