package admin

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/queue"
    "github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
    "github.com/redis/go-redis/v9"
)

// ErrNotImplemented indicates a contract that has not yet been implemented.
var ErrNotImplemented = errors.New("not implemented")

// DLQItem represents a dead‑letter entry suitable for TUI listing and actions.
// Implementations should populate ID and Queue from payload/metadata when possible.
type DLQItem struct {
    Handle    string    `json:"handle"`
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
    DLQRequeue(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, handles []string, destQueue string) (int, error)
    DLQPurge(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, handles []string) (int, error)
}

const dlqSelectionHandlePrefix = "dlq:v1:"

var removeDLQEntryAtScript = redis.NewScript(`
local type_reply = redis.call('TYPE', KEYS[1])
local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
if actual ~= 'none' and actual ~= 'list' then
  return redis.error_reply('WRONGTYPE key ' .. KEYS[1] .. ' has type ' .. actual .. ', expected list')
end
local index = tonumber(ARGV[2])
if not index or index < 0 then
  return redis.error_reply('selection index must be non-negative')
end
if redis.call('LINDEX', KEYS[1], index) ~= ARGV[1] then
  return 0
end
if ARGV[3] == ARGV[1] then
  return redis.error_reply('selection marker must differ from the envelope')
end
redis.call('LSET', KEYS[1], index, ARGV[3])
if redis.call('LREM', KEYS[1], 1, ARGV[3]) ~= 1 then
  return redis.error_reply('selected envelope could not be removed')
end
return 1
`)

type dlqSelection struct {
    handle string
    index  int64
    raw    string
}

func makeDLQSelectionHandle(index int64, raw string) string {
    digest := sha256.Sum256([]byte(raw))
    return fmt.Sprintf("%s%d:%x", dlqSelectionHandlePrefix, index, digest)
}

func parseDLQSelectionIndex(handle string) (int64, bool) {
    encoded, ok := strings.CutPrefix(handle, dlqSelectionHandlePrefix)
    if !ok {
        return 0, false
    }
    indexText, digest, ok := strings.Cut(encoded, ":")
    if !ok || len(digest) != sha256.Size*2 {
        return 0, false
    }
    index, err := strconv.ParseInt(indexText, 10, 64)
    if err != nil || index < 0 {
        return 0, false
    }
    return index, true
}

func resolveDLQSelections(ctx context.Context, rdb *redis.Client, list string, handles []string) ([]dlqSelection, error) {
    selections := make([]dlqSelection, 0, len(handles))
    seen := make(map[string]struct{}, len(handles))
    for _, handle := range handles {
        if _, duplicate := seen[handle]; duplicate {
            continue
        }
        seen[handle] = struct{}{}
        index, ok := parseDLQSelectionIndex(handle)
        if !ok {
            continue
        }
        raw, err := rdb.LIndex(ctx, list, index).Result()
        if errors.Is(err, redis.Nil) {
            continue
        }
        if err != nil {
            return nil, err
        }
        if makeDLQSelectionHandle(index, raw) != handle {
            continue
        }
        selections = append(selections, dlqSelection{handle: handle, index: index, raw: raw})
    }
    sort.Slice(selections, func(i, j int) bool {
        return selections[i].index > selections[j].index
    })
    return selections, nil
}

func dlqSelectionMarker(handle string) string {
    return "\x00grq-dlq-selection:" + handle
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
    for itemIndex, raw := range items {
        var meta struct {
            ID           string `json:"id"`
            Reason       string `json:"error"`
            Attempts     int    `json:"retries"`
            CreationTime string `json:"creation_time"`
        }
        _ = json.Unmarshal([]byte(raw), &meta)
        it := DLQItem{
            Handle:   makeDLQSelectionHandle(offset+int64(itemIndex), raw),
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

// DLQRequeue moves the specified DLQ selection handles to a destination queue.
// If destQueue is empty, the original queue (if available) should be used.
func DLQRequeue(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, handles []string, destQueue string) (int, error) {
    if cfg.Worker.DeadLetterList == "" {
        return 0, errors.New("dead letter list not configured")
    }
    if len(handles) == 0 {
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
    selections, err := resolveDLQSelections(ctx, rdb, cfg.Worker.DeadLetterList, handles)
    if err != nil {
        return 0, err
    }
    requeued := 0
    for _, selection := range selections {
        job, err := queue.UnmarshalJob(selection.raw)
        if err != nil {
            continue
        }
        moved, err := queue.RequeueEncodedAt(
            ctx,
            rdb,
            cfg.Worker.DeadLetterList,
            destQueue,
            selection.index,
            job,
            selection.raw,
            dlqSelectionMarker(selection.handle),
            cfg.OrderingLayout(),
        )
        if err != nil {
            return requeued, err
        }
        if moved {
            requeued++
        }
    }
    return requeued, nil
}

// DLQPurge deletes the specified DLQ selection handles.
func DLQPurge(ctx context.Context, cfg *config.Config, rdb *redis.Client, namespace string, handles []string) (int, error) {
    if cfg.Worker.DeadLetterList == "" {
        return 0, errors.New("dead letter list not configured")
    }
    if len(handles) == 0 {
        return 0, nil
    }
    selections, err := resolveDLQSelections(ctx, rdb, cfg.Worker.DeadLetterList, handles)
    if err != nil {
        return 0, err
    }
    purged := 0
    for _, selection := range selections {
        removed, err := removeDLQEntryAtScript.Eval(
            ctx,
            rdb,
            []string{cfg.Worker.DeadLetterList},
            selection.raw,
            selection.index,
            dlqSelectionMarker(selection.handle),
        ).Int64()
        if err != nil {
            return purged, err
        }
        if removed == 1 {
            purged++
        }
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
    hbFormat := heartbeatKeyPattern(cfg)
    plFormat := processingKeyPattern(cfg)
    hbPattern := queuekeys.ScanPattern(hbFormat)
    plPattern := queuekeys.ScanPattern(plFormat)

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
            id, ok := queuekeys.Extract(hbFormat, k)
            if !ok || id == "" {
                continue
            }
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
            id, ok := queuekeys.Extract(plFormat, k)
            if !ok || id == "" {
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
