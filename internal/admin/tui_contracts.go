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

// ErrInvalidDLQCursor indicates a malformed or out-of-range DLQ page cursor.
var ErrInvalidDLQCursor = errors.New("invalid dead-letter cursor")

// ErrStaleDLQCursor indicates that the DLQ changed after a page cursor was issued.
var ErrStaleDLQCursor = errors.New("stale dead-letter cursor")

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

const (
    dlqSelectionHandlePrefix = "dlq:v3:"
    dlqPageCursorPrefix = "dlq:cursor:v1:"
    dlqVersionLua = `
local function require_type(key, expected)
  local type_reply = redis.call('TYPE', key)
  local actual = type(type_reply) == 'table' and type_reply['ok'] or type_reply
  if actual ~= 'none' and actual ~= expected then
    return 'WRONGTYPE key ' .. key .. ' has type ' .. actual .. ', expected ' .. expected
  end
end

local function require_dlq_keys(list_key, generation_key)
  if list_key == generation_key then
    return 'dead-letter list and generation key must differ'
  end
  local problem = require_type(list_key, 'list')
  if problem then return problem end
  return require_type(generation_key, 'string')
end

local function generation_value(key)
  local generation = redis.call('GET', key)
  if not generation then return '0' end
  if not string.match(generation, '^%d+$') then return nil end
  if string.len(generation) > 1 and string.sub(generation, 1, 1) == '0' then return nil end
  if string.len(generation) > 19 or
      (string.len(generation) == 19 and generation > '9223372036854775807') then
    return nil
  end
  return generation
end

local function generation_has_room(generation)
  if string.len(generation) < 19 then return true end
  if string.len(generation) > 19 then return false end
  return generation < '9223372036854775807'
end

local function dlq_version(list_key, generation_key)
  local generation = generation_value(generation_key)
  if not generation then return nil end
  return generation .. ':' .. tostring(redis.call('LLEN', list_key))
end
`
)

var dlqVersionScript = redis.NewScript(dlqVersionLua + `
local problem = require_dlq_keys(KEYS[1], KEYS[2])
if problem then return redis.error_reply(problem) end
local version = dlq_version(KEYS[1], KEYS[2])
if not version then return redis.error_reply('invalid dead-letter generation') end
return version
`)

var dlqListPageScript = redis.NewScript(dlqVersionLua + `
local problem = require_dlq_keys(KEYS[1], KEYS[2])
if problem then return redis.error_reply(problem) end
local version = dlq_version(KEYS[1], KEYS[2])
if not version then return redis.error_reply('invalid dead-letter generation') end
if ARGV[3] ~= '' and ARGV[3] ~= version then
  return {'0', version}
end
local page = redis.call('LRANGE', KEYS[1], ARGV[1], ARGV[2])
local result = {'1', version}
for _, entry in ipairs(page) do
  result[#result + 1] = entry
end
return result
`)

var removeDLQEntryAtScript = redis.NewScript(dlqVersionLua + `
local problem = require_dlq_keys(KEYS[1], KEYS[2])
if problem then return redis.error_reply(problem) end
local index = tonumber(ARGV[2])
if not index or index < 0 then
  return redis.error_reply('selection index must be non-negative')
end
local current_version = dlq_version(KEYS[1], KEYS[2])
if not current_version then return redis.error_reply('invalid dead-letter generation') end
if current_version ~= ARGV[4] then
  return {'0', current_version}
end
if redis.call('LINDEX', KEYS[1], index) ~= ARGV[1] then
  return {'0', current_version}
end
if ARGV[3] == ARGV[1] then
  return redis.error_reply('selection marker must differ from the envelope')
end
local generation = generation_value(KEYS[2])
if not generation_has_room(generation) then
  return redis.error_reply('dead-letter generation exhausted')
end
redis.call('INCR', KEYS[2])
redis.call('LSET', KEYS[1], index, ARGV[3])
if redis.call('LREM', KEYS[1], 1, ARGV[3]) ~= 1 then
  return redis.error_reply('selected envelope could not be removed')
end
return {'1', dlq_version(KEYS[1], KEYS[2])}
`)

type dlqSelection struct {
    handle string
    index  int64
    raw    string
}

func makeDLQSelectionHandle(snapshot string, index int64, raw string) string {
    digest := sha256.Sum256([]byte(raw))
    return fmt.Sprintf("%s%s:%d:%x", dlqSelectionHandlePrefix, snapshot, index, digest)
}

func parseDLQSelectionHandle(handle string) (string, int64, bool) {
    encoded, ok := strings.CutPrefix(handle, dlqSelectionHandlePrefix)
    if !ok {
        return "", 0, false
    }
    generation, encoded, ok := strings.Cut(encoded, ":")
    if !ok {
        return "", 0, false
    }
    length, encoded, ok := strings.Cut(encoded, ":")
    if !ok {
        return "", 0, false
    }
    indexText, digest, ok := strings.Cut(encoded, ":")
    if !ok || len(digest) != sha256.Size*2 {
        return "", 0, false
    }
    index, err := strconv.ParseInt(indexText, 10, 64)
    if err != nil || index < 0 {
        return "", 0, false
    }
    snapshot := generation + ":" + length
    if !validDLQVersion(snapshot) {
        return "", 0, false
    }
    return snapshot, index, true
}

func validDLQVersion(version string) bool {
    generationText, lengthText, ok := strings.Cut(version, ":")
    if !ok {
        return false
    }
    generation, err := strconv.ParseInt(generationText, 10, 64)
    if err != nil || generation < 0 || strconv.FormatInt(generation, 10) != generationText {
        return false
    }
    length, err := strconv.ParseInt(lengthText, 10, 64)
    return err == nil && length >= 0 && strconv.FormatInt(length, 10) == lengthText
}

func makeDLQPageCursor(snapshot string, offset int64) string {
    return fmt.Sprintf("%s%s:%d", dlqPageCursorPrefix, snapshot, offset)
}

func parseDLQPageCursor(cursor string) (string, int64, error) {
    encoded, ok := strings.CutPrefix(cursor, dlqPageCursorPrefix)
    if !ok {
        return "", 0, fmt.Errorf("%w: %q", ErrInvalidDLQCursor, cursor)
    }
    generation, encoded, ok := strings.Cut(encoded, ":")
    if !ok {
        return "", 0, fmt.Errorf("%w: %q", ErrInvalidDLQCursor, cursor)
    }
    lengthText, offsetText, ok := strings.Cut(encoded, ":")
    snapshot := generation + ":" + lengthText
    if !ok || !validDLQVersion(snapshot) {
        return "", 0, fmt.Errorf("%w: %q", ErrInvalidDLQCursor, cursor)
    }
    length, _ := strconv.ParseInt(lengthText, 10, 64)
    offset, err := strconv.ParseInt(offsetText, 10, 64)
    if err != nil || offset < 0 || offset > length || strconv.FormatInt(offset, 10) != offsetText {
        return "", 0, fmt.Errorf("%w: %q", ErrInvalidDLQCursor, cursor)
    }
    return snapshot, offset, nil
}

func resolveDLQSelections(ctx context.Context, rdb *redis.Client, list string, handles []string) ([]dlqSelection, string, error) {
    snapshotResult, err := dlqVersionScript.Run(
        ctx,
        rdb,
        []string{list, queuekeys.DLQGenerationKey(list)},
    ).Result()
    if err != nil {
        return nil, "", err
    }
    snapshot, ok := snapshotResult.(string)
    if !ok || !validDLQVersion(snapshot) {
        return nil, "", fmt.Errorf("unexpected DLQ snapshot response %T", snapshotResult)
    }
    selections := make([]dlqSelection, 0, len(handles))
    seen := make(map[string]struct{}, len(handles))
    for _, handle := range handles {
        if _, duplicate := seen[handle]; duplicate {
            continue
        }
        seen[handle] = struct{}{}
        handleSnapshot, index, ok := parseDLQSelectionHandle(handle)
        if !ok || handleSnapshot != snapshot {
            continue
        }
        raw, err := rdb.LIndex(ctx, list, index).Result()
        if errors.Is(err, redis.Nil) {
            continue
        }
        if err != nil {
            return nil, "", err
        }
        if makeDLQSelectionHandle(snapshot, index, raw) != handle {
            continue
        }
        selections = append(selections, dlqSelection{handle: handle, index: index, raw: raw})
    }
    sort.Slice(selections, func(i, j int) bool {
        return selections[i].index > selections[j].index
    })
    return selections, snapshot, nil
}

func dlqSelectionMarker(handle string) string {
    return "\x00grq-dlq-selection:" + handle
}

func parseDLQMutationResult(result interface{}) (bool, string, error) {
    values, ok := result.([]interface{})
    if !ok || len(values) != 2 {
        return false, "", fmt.Errorf("unexpected DLQ mutation response %T", result)
    }
    changed, ok := values[0].(string)
    if !ok || (changed != "0" && changed != "1") {
        return false, "", fmt.Errorf("unexpected DLQ mutation status %v", values[0])
    }
    snapshot, ok := values[1].(string)
    if !ok || !validDLQVersion(snapshot) {
        return false, "", fmt.Errorf("unexpected DLQ mutation snapshot %T", values[1])
    }
    return changed == "1", snapshot, nil
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
    var expectedSnapshot string
    var offset int64
    if cursor != "" {
        var err error
        expectedSnapshot, offset, err = parseDLQPageCursor(cursor)
        if err != nil {
            return nil, "", err
        }
    }
    start := offset
    stop := offset + int64(limit) - 1
    if stop < start {
        return nil, "", fmt.Errorf("%w: offset overflow", ErrInvalidDLQCursor)
    }
    pageResult, err := dlqListPageScript.Run(
        ctx,
        rdb,
        []string{
            cfg.Worker.DeadLetterList,
            queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList),
        },
        start,
        stop,
        expectedSnapshot,
    ).Result()
    if err != nil {
        return nil, "", err
    }
    pageValues, ok := pageResult.([]interface{})
    if !ok || len(pageValues) < 2 {
        return nil, "", fmt.Errorf("unexpected DLQ page response %T", pageResult)
    }
    status, ok := pageValues[0].(string)
    if !ok || (status != "0" && status != "1") {
        return nil, "", fmt.Errorf("unexpected DLQ page status %v", pageValues[0])
    }
    snapshot, ok := pageValues[1].(string)
    if !ok || !validDLQVersion(snapshot) {
        return nil, "", fmt.Errorf("unexpected DLQ snapshot response %T", pageValues[1])
    }
    if status == "0" {
        return nil, "", fmt.Errorf("%w: expected %s, current %s", ErrStaleDLQCursor, expectedSnapshot, snapshot)
    }
    items := make([]string, 0, len(pageValues)-2)
    for _, value := range pageValues[2:] {
        raw, ok := value.(string)
        if !ok {
            return nil, "", fmt.Errorf("unexpected DLQ entry response %T", value)
        }
        items = append(items, raw)
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
            Handle:   makeDLQSelectionHandle(snapshot, offset+int64(itemIndex), raw),
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
    _, lengthText, _ := strings.Cut(snapshot, ":")
    snapshotLength, _ := strconv.ParseInt(lengthText, 10, 64)
    nextOffset := offset + int64(len(items))
    if nextOffset < offset {
        return nil, "", fmt.Errorf("unexpected DLQ page offset overflow")
    }
    if nextOffset >= snapshotLength {
        return out, "", nil
    }
    return out, makeDLQPageCursor(snapshot, nextOffset), nil
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
    selections, snapshot, err := resolveDLQSelections(ctx, rdb, cfg.Worker.DeadLetterList, handles)
    if err != nil {
        return 0, err
    }
    requeued := 0
    for _, selection := range selections {
        job, err := queue.UnmarshalJob(selection.raw)
        if err != nil {
            continue
        }
        destination := destQueue
        if job.OrderingKey == "" && destination == "" {
            destination = cfg.Worker.Queues[job.Priority]
        }
        moved, nextSnapshot, err := queue.RequeueEncodedAt(
            ctx,
            rdb,
            cfg.Worker.DeadLetterList,
            destination,
            selection.index,
            job,
            selection.raw,
            dlqSelectionMarker(selection.handle),
            snapshot,
            cfg.OrderingLayout(),
        )
        if err != nil {
            return requeued, err
        }
        if !moved {
            break
        }
        snapshot = nextSnapshot
        requeued++
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
    selections, snapshot, err := resolveDLQSelections(ctx, rdb, cfg.Worker.DeadLetterList, handles)
    if err != nil {
        return 0, err
    }
    purged := 0
    for _, selection := range selections {
        result, err := removeDLQEntryAtScript.Eval(
            ctx,
            rdb,
            []string{
                cfg.Worker.DeadLetterList,
                queuekeys.DLQGenerationKey(cfg.Worker.DeadLetterList),
            },
            selection.raw,
            selection.index,
            dlqSelectionMarker(selection.handle),
            snapshot,
        ).Result()
        if err != nil {
            return purged, err
        }
        removed, nextSnapshot, err := parseDLQMutationResult(result)
        if err != nil {
            return purged, err
        }
        if !removed {
            break
        }
        snapshot = nextSnapshot
        purged++
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
