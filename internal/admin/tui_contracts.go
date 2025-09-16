package admin

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/redis/go-redis/v9"
)

// ErrNotImplemented indicates a contract that has not yet been implemented.
var ErrNotImplemented = errors.New("not implemented")

// DLQItem represents a dead‑letter entry suitable for TUI listing and actions.
// Implementations should populate ID and Queue from payload/metadata when possible.
type DLQItem struct {
    ID        string    `json:"id"`
    Queue     string    `json:"queue"`
    Payload   []byte    `json:"payload"`
    Reason    string    `json:"reason,omitempty"`
    Attempts  int       `json:"attempts,omitempty"`
    FirstSeen time.Time `json:"first_seen,omitempty"`
    LastSeen  time.Time `json:"last_seen,omitempty"`
}

// DLQService defines the contract for listing and acting on DLQ items.
type DLQService interface {
    DLQList(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, cursor string, limit int) ([]DLQItem, string, error)
    DLQRequeue(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, ids []string, destQueue string) (int, error)
    DLQPurge(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, ids []string) (int, error)
}

// DLQList returns a page of DLQ items along with an opaque cursor for the next page.
// The cursor semantics are implementation‑defined and should be treated as opaque by callers.
func DLQList(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, cursor string, limit int) ([]DLQItem, string, error) {
    if cfg.Worker.DeadLetterList == "" {
        return nil, "", errors.New("dead letter list not configured")
    }
    if limit <= 0 || limit > 500 {
        limit = 100
    }
    // Cursor is a simple decimal offset into the list
    var offset int64
    if cursor != "" {
        var parsed int64
        _, err := fmt.Sscan(cursor, &parsed)
        if err == nil && parsed >= 0 {
            offset = parsed
        }
    }
    // Compute stop index and fetch
    start := offset
    stop := offset + int64(limit) - 1
    items, err := rdb.LRange(ctx, cfg.Worker.DeadLetterList, start, stop).Result()
    if err != nil {
        return nil, "", err
    }
    out := make([]DLQItem, 0, len(items))
    for _, raw := range items {
        var meta struct {
            ID           string `json:"id"`
            Reason       string `json:"error"`
            Attempts     int    `json:"retries"`
            CreationTime string `json:"creation_time"`
        }
        _ = json.Unmarshal([]byte(raw), &meta)
        it := DLQItem{
            ID:       meta.ID,
            Queue:    "", // unknown from payload; left blank
            Payload:  []byte(raw),
            Reason:   meta.Reason,
            Attempts: meta.Attempts,
        }
        if t, err := time.Parse(time.RFC3339Nano, meta.CreationTime); err == nil {
            it.FirstSeen = t
            it.LastSeen = t
        }
        out = append(out, it)
    }
    // Determine next cursor
    if len(items) < limit {
        return out, "", nil
    }
    next := fmt.Sprintf("%d", offset+int64(len(items)))
    return out, next, nil
}

// DLQRequeue moves the specified DLQ item IDs back to a destination queue.
// If destQueue is empty, the original queue (if available) should be used.
func DLQRequeue(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, ids []string, destQueue string) (int, error) {
    if cfg.Worker.DeadLetterList == "" {
        return 0, errors.New("dead letter list not configured")
    }
    if len(ids) == 0 {
        return 0, nil
    }
    // Resolve destination queue; default to high priority
    if destQueue == "" {
        if q, ok := cfg.Worker.Queues["high"]; ok && q != "" {
            destQueue = q
        } else {
            // fallback to low or DLQ (no-op)
            destQueue = cfg.Worker.Queues["low"]
        }
    }
    // Build a set for quick lookup
    idset := map[string]struct{}{}
    for _, id := range ids {
        if id != "" {
            idset[id] = struct{}{}
        }
    }
    // Iterate DLQ in chunks to find matching items
    const chunk = 500
    requeued := 0
    var start int64
    for {
        batch, err := rdb.LRange(ctx, cfg.Worker.DeadLetterList, start, start+chunk-1).Result()
        if err != nil {
            return requeued, err
        }
        if len(batch) == 0 {
            break
        }
        for _, raw := range batch {
            var meta struct{ ID string `json:"id"` }
            if err := json.Unmarshal([]byte(raw), &meta); err != nil {
                continue
            }
            if _, ok := idset[meta.ID]; !ok {
                continue
            }
            // Remove one matching occurrence and push to destination
            if _, err := rdb.LRem(ctx, cfg.Worker.DeadLetterList, 1, raw).Result(); err != nil {
                return requeued, err
            }
            if err := rdb.LPush(ctx, destQueue, raw).Err(); err != nil {
                return requeued, err
            }
            requeued++
        }
        if len(batch) < chunk {
            break
        }
        start += chunk
    }
    return requeued, nil
}

// DLQPurge deletes the specified DLQ item IDs.
func DLQPurge(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, ids []string) (int, error) {
    if cfg.Worker.DeadLetterList == "" {
        return 0, errors.New("dead letter list not configured")
    }
    if len(ids) == 0 {
        return 0, nil
    }
    idset := map[string]struct{}{}
    for _, id := range ids {
        if id != "" {
            idset[id] = struct{}{}
        }
    }
    purged := 0
    const chunk = 500
    var start int64
    for {
        batch, err := rdb.LRange(ctx, cfg.Worker.DeadLetterList, start, start+chunk-1).Result()
        if err != nil {
            return purged, err
        }
        if len(batch) == 0 {
            break
        }
        for _, raw := range batch {
            var meta struct{ ID string `json:"id"` }
            if err := json.Unmarshal([]byte(raw), &meta); err != nil {
                continue
            }
            if _, ok := idset[meta.ID]; !ok {
                continue
            }
            if _, err := rdb.LRem(ctx, cfg.Worker.DeadLetterList, 1, raw).Result(); err != nil {
                return purged, err
            }
            purged++
        }
        if len(batch) < chunk {
            break
        }
        start += chunk
    }
    return purged, nil
}

// WorkerInfo summarizes a worker’s status for the TUI Workers tab.
type WorkerInfo struct {
    ID            string     `json:"id"`
    LastHeartbeat time.Time  `json:"last_heartbeat"`
    Queue         string     `json:"queue,omitempty"`
    JobID         string     `json:"job_id,omitempty"`
    StartedAt     *time.Time `json:"started_at,omitempty"`
    Version       string     `json:"version,omitempty"`
    Host          string     `json:"host,omitempty"`
}

// WorkerService defines the contract for querying worker status.
type WorkerService interface {
    Workers(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string) ([]WorkerInfo, error)
}

// Workers lists currently known workers in the given namespace.
func Workers(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string) ([]WorkerInfo, error) {
    // Discover workers from heartbeat and processing keys
    hbPattern := "jobqueue:processing:worker:*"
    plPattern := "jobqueue:worker:*:processing"

    workerMap := map[string]*WorkerInfo{}

    // Heartbeats: presence implies online; we don’t have timestamps stored yet, so set to now
    var cursor uint64
    for {
        keys, cur, err := rdb.Scan(ctx, cursor, hbPattern, 500).Result()
        if err != nil {
            return nil, err
        }
        cursor = cur
        for _, k := range keys {
            id := k[strings.LastIndex(k, ":")+1:]
            wi := workerMap[id]
            if wi == nil {
                wi = &WorkerInfo{ID: id}
                workerMap[id] = wi
            }
            wi.LastHeartbeat = time.Now()
        }
        if cursor == 0 {
            break
        }
    }

    // Processing lists: derive worker IDs and attempt to read active job
    cursor = 0
    for {
        keys, cur, err := rdb.Scan(ctx, cursor, plPattern, 500).Result()
        if err != nil {
            return nil, err
        }
        cursor = cur
        for _, k := range keys {
            // Format: jobqueue:worker:<id>:processing
            parts := strings.Split(k, ":")
            id := ""
            if len(parts) >= 3 {
                id = parts[2]
            }
            if id == "" {
                continue
            }
            wi := workerMap[id]
            if wi == nil {
                wi = &WorkerInfo{ID: id}
                workerMap[id] = wi
            }
            // Peek last item as the most recent
            raw, err := rdb.LIndex(ctx, k, -1).Result()
            if err == nil && raw != "" {
                var meta struct{ ID string `json:"id"` }
                _ = json.Unmarshal([]byte(raw), &meta)
                wi.JobID = meta.ID
            }
        }
        if cursor == 0 {
            break
        }
    }

    out := make([]WorkerInfo, 0, len(workerMap))
    for _, wi := range workerMap {
        out = append(out, *wi)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
    return out, nil
}

// JobEvent is a timeline event for a job used by the Time Travel debugger.
type JobEvent struct {
    TS   time.Time         `json:"ts"`
    Type string            `json:"type"`
    Data map[string]any    `json:"data,omitempty"`
}

// TimelineService defines the contract for job timeline retrieval and streaming.
type TimelineService interface {
    JobTimeline(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace, jobID string, start, end *time.Time, limit int) ([]JobEvent, error)
    SubscribeJob(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace, jobID string) (<-chan JobEvent, func(), error)
}

// JobTimeline returns a bounded slice of events for a job ID, optionally filtered by time.
func JobTimeline(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace, jobID string, start, end *time.Time, limit int) ([]JobEvent, error) {
    return nil, ErrNotImplemented
}

// SubscribeJob opens a live event stream for a job; returns a channel and a cancel func.
func SubscribeJob(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace, jobID string) (<-chan JobEvent, func(), error) {
    return nil, func() {}, ErrNotImplemented
}
