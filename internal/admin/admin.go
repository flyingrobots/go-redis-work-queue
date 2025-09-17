// Copyright 2025 James Ross
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
	"github.com/flyingrobots/go-redis-work-queue/internal/distributed-tracing-integration"
	"github.com/redis/go-redis/v9"
)

type StatsResult struct {
	Queues          map[string]int64 `json:"queues"`
	ProcessingLists map[string]int64 `json:"processing_lists"`
	Heartbeats      int64            `json:"heartbeats"`
}

func Stats(ctx context.Context, cfg *config.Config, rdb *redis.Client) (StatsResult, error) {
	res := StatsResult{Queues: map[string]int64{}, ProcessingLists: map[string]int64{}}
	// Count standard queues
	qset := map[string]string{}
	for p, q := range cfg.Worker.Queues {
		qset[p] = q
	}
	qset["completed"] = cfg.Worker.CompletedList
	qset["dead_letter"] = cfg.Worker.DeadLetterList
	for name, key := range qset {
		n, err := rdb.LLen(ctx, key).Result()
		if err != nil {
			return res, err
		}
		res.Queues[name+"("+key+")"] = n
	}
	// Scan processing lists
	var cursor uint64
	for {
		keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:worker:*:processing", 200).Result()
		if err != nil {
			return res, err
		}
		cursor = cur
		for _, k := range keys {
			n, _ := rdb.LLen(ctx, k).Result()
			res.ProcessingLists[k] = n
		}
		if cursor == 0 {
			break
		}
	}
	// Heartbeats
	var hbc int64
	cursor = 0
	for {
		keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:processing:worker:*", 500).Result()
		if err != nil {
			return res, err
		}
		cursor = cur
		hbc += int64(len(keys))
		if cursor == 0 {
			break
		}
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
	if err != nil {
		return PeekResult{}, err
	}
	if n <= 0 {
		n = 10
	}
	// Items to be consumed next are at the right end; take last N
	items, err := rdb.LRange(ctx, qkey, -n, -1).Result()
	if err != nil {
		return PeekResult{}, err
	}
	return PeekResult{Queue: qkey, Items: items}, nil
}

func PurgeDLQ(ctx context.Context, cfg *config.Config, rdb *redis.Client) error {
	if cfg.Worker.DeadLetterList == "" {
		return errors.New("dead letter list not configured")
	}
	return rdb.Del(ctx, cfg.Worker.DeadLetterList).Err()
}

func resolveQueue(cfg *config.Config, alias string) (string, error) {
	a := strings.ToLower(alias)
	if a == "completed" {
		return cfg.Worker.CompletedList, nil
	}
	if a == "dead_letter" || a == "dlq" {
		return cfg.Worker.DeadLetterList, nil
	}
	if q, ok := cfg.Worker.Queues[a]; ok {
		return q, nil
	}
	// Otherwise, assume full key
	if strings.HasPrefix(alias, "jobqueue:") {
		return alias, nil
	}
	// Suggest options
	keys := make([]string, 0, len(cfg.Worker.Queues))
	for k := range cfg.Worker.Queues {
		keys = append(keys, k)
	}
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
func Bench(ctx context.Context, cfg *config.Config, rdb *redis.Client, priority string, count int, rate int, payloadSize int, timeout time.Duration) (BenchResult, error) {
	res := BenchResult{Count: count}
	if count <= 0 {
		return res, fmt.Errorf("count must be > 0")
	}
	if rate <= 0 {
		rate = 100
	}
	if payloadSize <= 0 {
		payloadSize = 1024
	}
	qkey, err := resolveQueue(cfg, priority)
	if err != nil {
		return res, err
	}
	// Clear completed
	_ = rdb.Del(ctx, cfg.Worker.CompletedList).Err()

	// Enqueue
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()
	start := time.Now()
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-ticker.C:
		}
		payload := fmt.Sprintf(`{"id":"bench-%d","filepath":"/bench/%d","filesize":%d,"priority":"%s","retries":0,"creation_time":"%s","trace_id":"","span_id":""}`,
			i, i, payloadSize, priority, time.Now().UTC().Format(time.RFC3339Nano))
		if err := rdb.LPush(ctx, qkey, payload).Err(); err != nil {
			return res, err
		}
	}

	// Wait for completion
	doneBy := time.Now().Add(timeout)
	for time.Now().Before(doneBy) {
		n, _ := rdb.LLen(ctx, cfg.Worker.CompletedList).Result()
		if int(n) >= count {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	res.Duration = time.Since(start)
	if res.Duration > 0 {
		res.Throughput = float64(count) / res.Duration.Seconds()
	}

	// Fetch and compute latencies
	items, _ := rdb.LRange(ctx, cfg.Worker.CompletedList, 0, -1).Result()
	lats := make([]float64, 0, len(items))
	now := time.Now()
	for _, it := range items {
		var j struct {
			CreationTime string `json:"creation_time"`
		}
		if err := json.Unmarshal([]byte(it), &j); err == nil {
			if t, err2 := time.Parse(time.RFC3339Nano, j.CreationTime); err2 == nil {
				lats = append(lats, now.Sub(t).Seconds())
			}
		}
	}
	if len(lats) > 0 {
		sort.Float64s(lats)
		res.P50 = time.Duration(lats[int(math.Round(0.50*float64(len(lats)-1)))] * float64(time.Second))
		res.P95 = time.Duration(lats[int(math.Round(0.95*float64(len(lats)-1)))] * float64(time.Second))
	}
	return res, nil
}

// KeysStats summarizes managed Redis keys and queue lengths.
type KeysStats struct {
	QueueLengths    map[string]int64 `json:"queue_lengths"`
	ProcessingLists int64            `json:"processing_lists"`
	ProcessingItems int64            `json:"processing_items"`
	Heartbeats      int64            `json:"heartbeats"`
	RateLimitKey    string           `json:"rate_limit_key"`
	RateLimitTTL    string           `json:"rate_limit_ttl,omitempty"`
}

// StatsKeys scans for managed keys and returns counts and lengths.
func StatsKeys(ctx context.Context, cfg *config.Config, rdb *redis.Client) (KeysStats, error) {
	out := KeysStats{QueueLengths: map[string]int64{}}
	// Known queues
	qset := map[string]string{
		"high":        cfg.Worker.Queues["high"],
		"low":         cfg.Worker.Queues["low"],
		"completed":   cfg.Worker.CompletedList,
		"dead_letter": cfg.Worker.DeadLetterList,
	}
	for name, key := range qset {
		if key == "" {
			continue
		}
		n, err := rdb.LLen(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return out, err
		}
		out.QueueLengths[name+"("+key+")"] = n
	}
	// Processing lists
	var cursor uint64
	for {
		keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:worker:*:processing", 500).Result()
		if err != nil {
			return out, err
		}
		cursor = cur
		out.ProcessingLists += int64(len(keys))
		for _, k := range keys {
			n, _ := rdb.LLen(ctx, k).Result()
			out.ProcessingItems += n
		}
		if cursor == 0 {
			break
		}
	}
	// Heartbeats
	cursor = 0
	for {
		keys, cur, err := rdb.Scan(ctx, cursor, "jobqueue:processing:worker:*", 1000).Result()
		if err != nil {
			return out, err
		}
		cursor = cur
		out.Heartbeats += int64(len(keys))
		if cursor == 0 {
			break
		}
	}
	// Rate limiter
	if cfg.Producer.RateLimitKey != "" {
		out.RateLimitKey = cfg.Producer.RateLimitKey
		if ttl, err := rdb.TTL(ctx, cfg.Producer.RateLimitKey).Result(); err == nil && ttl > 0 {
			out.RateLimitTTL = ttl.String()
		}
	}
	return out, nil
}

// PurgeAll deletes common test keys used by this system, including
// priority queues, completed/dead_letter, rate limiter key, and
// per-worker processing lists and heartbeats. Returns number of keys deleted.
func PurgeAll(ctx context.Context, cfg *config.Config, rdb *redis.Client) (int64, error) {
	var deleted int64
	// Explicit keys
	keys := []string{
		cfg.Worker.Queues["high"], cfg.Worker.Queues["low"],
		cfg.Worker.CompletedList, cfg.Worker.DeadLetterList,
	}
	if cfg.Producer.RateLimitKey != "" {
		keys = append(keys, cfg.Producer.RateLimitKey)
	}
	// Dedup
	uniq := map[string]struct{}{}
	ek := make([]string, 0, len(keys))
	for _, k := range keys {
		if k == "" {
			continue
		}
		if _, ok := uniq[k]; ok {
			continue
		}
		uniq[k] = struct{}{}
		ek = append(ek, k)
	}
	if len(ek) > 0 {
		n, err := rdb.Del(ctx, ek...).Result()
		if err != nil {
			return deleted, err
		}
		deleted += n
	}
	// Patterns: processing lists and heartbeats
	patterns := []string{
		"jobqueue:worker:*:processing",
		"jobqueue:processing:worker:*",
	}
	for _, pat := range patterns {
		var cursor uint64
		for {
			keys, cur, err := rdb.Scan(ctx, cursor, pat, 500).Result()
			if err != nil {
				return deleted, err
			}
			cursor = cur
			if len(keys) > 0 {
				n, err := rdb.Del(ctx, keys...).Result()
				if err != nil {
					return deleted, err
				}
				deleted += n
			}
			if cursor == 0 {
				break
			}
		}
	}
	return deleted, nil
}

// PeekWithTracing enhances the standard Peek function with tracing information
type PeekWithTracingResult struct {
	PeekResult
	TraceJobs    []*distributed_tracing_integration.TraceableJob          `json:"trace_jobs,omitempty"`
	TraceActions map[string][]distributed_tracing_integration.TraceAction `json:"trace_actions,omitempty"`
	TraceInfo    []string                                                 `json:"trace_info,omitempty"`
}

func PeekWithTracing(ctx context.Context, cfg *config.Config, rdb *redis.Client, queueAlias string, n int64) (PeekWithTracingResult, error) {
	// Get basic peek result
	basicResult, err := Peek(ctx, cfg, rdb, queueAlias, n)
	if err != nil {
		return PeekWithTracingResult{}, err
	}

	result := PeekWithTracingResult{
		PeekResult:   basicResult,
		TraceJobs:    make([]*distributed_tracing_integration.TraceableJob, 0),
		TraceActions: make(map[string][]distributed_tracing_integration.TraceAction),
		TraceInfo:    make([]string, 0),
	}

	// Initialize tracing integration with defaults
	tracing := distributed_tracing_integration.NewWithDefaults()

	// Parse jobs and extract trace information
	for _, item := range basicResult.Items {
		job, err := distributed_tracing_integration.ParseJobWithTrace(item)
		if err != nil {
			// Skip items that can't be parsed, but continue processing
			continue
		}

		result.TraceJobs = append(result.TraceJobs, job)

		// Generate trace actions for jobs with trace information
		if job.TraceID != "" {
			actions := distributed_tracing_integration.GenerateTraceActions(job.TraceID, tracing.GetConfig())
			result.TraceActions[job.ID] = actions

			// Add formatted trace info
			traceInfo := distributed_tracing_integration.FormatTraceForDisplay(job, tracing.GetConfig())
			result.TraceInfo = append(result.TraceInfo, fmt.Sprintf("Job %s: %s", job.ID, traceInfo))
		} else {
			result.TraceInfo = append(result.TraceInfo, fmt.Sprintf("Job %s: No trace information", job.ID))
		}
	}

	return result, nil
}

// InfoWithTracing provides detailed job information including trace data
type JobInfoResult struct {
	JobID        string                                        `json:"job_id"`
	Queue        string                                        `json:"queue"`
	Job          *distributed_tracing_integration.TraceableJob `json:"job,omitempty"`
	TraceActions []distributed_tracing_integration.TraceAction `json:"trace_actions,omitempty"`
	TraceURL     string                                        `json:"trace_url,omitempty"`
	RawJSON      string                                        `json:"raw_json"`
}

func InfoWithTracing(ctx context.Context, cfg *config.Config, rdb *redis.Client, queueAlias string, jobIndex int) (JobInfoResult, error) {
	// Get items from queue
	peekResult, err := Peek(ctx, cfg, rdb, queueAlias, int64(jobIndex+1))
	if err != nil {
		return JobInfoResult{}, err
	}

	if jobIndex >= len(peekResult.Items) {
		return JobInfoResult{}, fmt.Errorf("job index %d out of range (queue has %d items)", jobIndex, len(peekResult.Items))
	}

	jobJSON := peekResult.Items[jobIndex]

	result := JobInfoResult{
		Queue:   peekResult.Queue,
		RawJSON: jobJSON,
	}

	// Parse job with trace information
	job, err := distributed_tracing_integration.ParseJobWithTrace(jobJSON)
	if err != nil {
		return result, nil // Return basic info even if trace parsing fails
	}

	result.JobID = job.ID
	result.Job = job

	// Generate trace actions and URL if trace info is available
	if job.TraceID != "" {
		tracing := distributed_tracing_integration.NewWithDefaults()
		result.TraceActions = distributed_tracing_integration.GenerateTraceActions(job.TraceID, tracing.GetConfig())
		result.TraceURL = tracing.GetTraceURL(job.TraceID)
	}

	return result, nil
}

// GetTraceActions returns available actions for a specific trace ID
func GetTraceActions(traceID string) []distributed_tracing_integration.TraceAction {
	if traceID == "" {
		return []distributed_tracing_integration.TraceAction{}
	}

	tracing := distributed_tracing_integration.NewWithDefaults()
	return distributed_tracing_integration.GenerateTraceActions(traceID, tracing.GetConfig())
}
