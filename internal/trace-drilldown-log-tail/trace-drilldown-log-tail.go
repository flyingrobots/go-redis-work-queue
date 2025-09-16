// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceManager manages trace collection and viewing
type TraceManager struct {
	config       *TracingConfig
	redis        *redis.Client
	logger       *zap.Logger
	httpClient   *http.Client
	traces       map[string]*TraceInfo
	mu           sync.RWMutex
}

// NewTraceManager creates a new trace manager
func NewTraceManager(config *TracingConfig, redis *redis.Client, logger *zap.Logger) *TraceManager {
	if config == nil {
		config = &TracingConfig{
			Enabled:      true,
			Provider:     "jaeger",
			ServiceName:  "go-redis-work-queue",
			SamplingRate: 1.0,
		}
	}

	return &TraceManager{
		config:     config,
		redis:      redis,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		traces:     make(map[string]*TraceInfo),
	}
}

// StartTrace starts a new trace
func (tm *TraceManager) StartTrace(ctx context.Context, operationName string) (*TraceContext, context.Context) {
	if !tm.config.Enabled {
		return nil, ctx
	}

	traceID := generateTraceID()
	spanID := generateSpanID()

	traceCtx := &TraceContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: tm.shouldSample(),
		Baggage: make(map[string]string),
	}

	// Store in context
	ctx = context.WithValue(ctx, "trace", traceCtx)

	// Create trace info
	traceInfo := &TraceInfo{
		TraceID:       traceID,
		SpanID:        spanID,
		ServiceName:   tm.config.ServiceName,
		OperationName: operationName,
		StartTime:     time.Now(),
		Status:        "active",
		Tags:          make(map[string]string),
		Logs:          make([]TraceLog, 0),
	}

	tm.mu.Lock()
	tm.traces[traceID] = traceInfo
	tm.mu.Unlock()

	// Store in Redis for distributed access
	tm.storeTrace(traceInfo)

	return traceCtx, ctx
}

// EndTrace ends a trace
func (tm *TraceManager) EndTrace(ctx context.Context, status string) {
	traceCtx := tm.getTraceContext(ctx)
	if traceCtx == nil {
		return
	}

	tm.mu.Lock()
	if trace, exists := tm.traces[traceCtx.TraceID]; exists {
		trace.EndTime = time.Now()
		trace.Duration = trace.EndTime.Sub(trace.StartTime)
		trace.Status = status
	}
	tm.mu.Unlock()

	// Update in Redis
	tm.updateTrace(traceCtx.TraceID, status)
}

// AddTraceLog adds a log to the current trace
func (tm *TraceManager) AddTraceLog(ctx context.Context, level, message string, fields map[string]interface{}) {
	traceCtx := tm.getTraceContext(ctx)
	if traceCtx == nil || !traceCtx.Sampled {
		return
	}

	log := TraceLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	tm.mu.Lock()
	if trace, exists := tm.traces[traceCtx.TraceID]; exists {
		trace.Logs = append(trace.Logs, log)
	}
	tm.mu.Unlock()
}

// GetTrace retrieves trace information
func (tm *TraceManager) GetTrace(traceID string) (*TraceInfo, error) {
	// Check local cache first
	tm.mu.RLock()
	if trace, exists := tm.traces[traceID]; exists {
		tm.mu.RUnlock()
		return trace, nil
	}
	tm.mu.RUnlock()

	// Try Redis
	return tm.loadTrace(traceID)
}

// GetTraceLink generates an external link for viewing the trace
func (tm *TraceManager) GetTraceLink(traceID string) (*TraceLink, error) {
	if tm.config.URLTemplate == "" {
		return nil, fmt.Errorf("no URL template configured")
	}

	// Replace placeholders in template
	url := strings.ReplaceAll(tm.config.URLTemplate, "{trace_id}", traceID)
	url = strings.ReplaceAll(url, "{service}", tm.config.ServiceName)

	link := &TraceLink{
		Type:        tm.config.Provider,
		URL:         url,
		DisplayName: fmt.Sprintf("View in %s", tm.config.Provider),
	}

	return link, nil
}

// GetSpanSummary retrieves a summary of spans for a trace
func (tm *TraceManager) GetSpanSummary(ctx context.Context, traceID string) (*SpanSummary, error) {
	// Fetch from external tracing system if configured
	if tm.config.Endpoint != "" {
		return tm.fetchSpanSummary(ctx, traceID)
	}

	// Otherwise use local data
	trace, err := tm.GetTrace(traceID)
	if err != nil {
		return nil, err
	}

	summary := &SpanSummary{
		TraceID:    traceID,
		TotalSpans: 1, // Basic implementation
		Duration:   trace.Duration,
		Services:   []string{trace.ServiceName},
		Timeline:   make([]TimelineEvent, 0),
	}

	// Add timeline events
	summary.Timeline = append(summary.Timeline, TimelineEvent{
		Timestamp: trace.StartTime,
		SpanID:    trace.SpanID,
		Operation: trace.OperationName,
		Service:   trace.ServiceName,
		EventType: "start",
	})

	if !trace.EndTime.IsZero() {
		summary.Timeline = append(summary.Timeline, TimelineEvent{
			Timestamp: trace.EndTime,
			SpanID:    trace.SpanID,
			Operation: trace.OperationName,
			Service:   trace.ServiceName,
			Duration:  trace.Duration,
			EventType: "end",
		})
	}

	return summary, nil
}

// SearchTraces searches for traces
func (tm *TraceManager) SearchTraces(ctx context.Context, filter *LogFilter) (*TraceSearchResult, error) {
	result := &TraceSearchResult{
		Traces: make([]TraceInfo, 0),
	}

	// Search in Redis
	keys, err := tm.redis.Keys(ctx, "trace:*").Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		data, err := tm.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var trace TraceInfo
		if err := json.Unmarshal([]byte(data), &trace); err != nil {
			continue
		}

		// Apply filters
		if tm.matchesFilter(&trace, filter) {
			result.Traces = append(result.Traces, trace)
		}

		if filter.MaxResults > 0 && len(result.Traces) >= filter.MaxResults {
			result.HasMore = true
			break
		}
	}

	result.TotalCount = len(result.Traces)
	return result, nil
}

// PropagateTrace propagates trace context to downstream services
func (tm *TraceManager) PropagateTrace(ctx context.Context, headers http.Header) {
	traceCtx := tm.getTraceContext(ctx)
	if traceCtx == nil {
		return
	}

	// Add standard trace headers
	headers.Set("X-Trace-Id", traceCtx.TraceID)
	headers.Set("X-Span-Id", traceCtx.SpanID)
	headers.Set("X-Sampled", fmt.Sprintf("%t", traceCtx.Sampled))

	// Add provider-specific headers
	switch tm.config.Provider {
	case "jaeger":
		headers.Set("uber-trace-id", fmt.Sprintf("%s:%s:0:%d",
			traceCtx.TraceID, traceCtx.SpanID, boolToInt(traceCtx.Sampled)))
	case "zipkin":
		headers.Set("X-B3-TraceId", traceCtx.TraceID)
		headers.Set("X-B3-SpanId", traceCtx.SpanID)
		headers.Set("X-B3-Sampled", fmt.Sprintf("%d", boolToInt(traceCtx.Sampled)))
	case "datadog":
		headers.Set("x-datadog-trace-id", traceCtx.TraceID)
		headers.Set("x-datadog-parent-id", traceCtx.SpanID)
	}

	// Add custom propagation headers
	for _, header := range tm.config.PropagateHeaders {
		if value := ctx.Value(header); value != nil {
			headers.Set(header, fmt.Sprintf("%v", value))
		}
	}
}

// ExtractTrace extracts trace context from incoming headers
func (tm *TraceManager) ExtractTrace(headers http.Header) *TraceContext {
	var traceID, spanID string
	sampled := false

	// Try standard headers first
	if tid := headers.Get("X-Trace-Id"); tid != "" {
		traceID = tid
		spanID = headers.Get("X-Span-Id")
		sampled = headers.Get("X-Sampled") == "true"
	} else {
		// Try provider-specific headers
		switch tm.config.Provider {
		case "jaeger":
			if uber := headers.Get("uber-trace-id"); uber != "" {
				parts := strings.Split(uber, ":")
				if len(parts) >= 4 {
					traceID = parts[0]
					spanID = parts[1]
					sampled = parts[3] == "1"
				}
			}
		case "zipkin":
			traceID = headers.Get("X-B3-TraceId")
			spanID = headers.Get("X-B3-SpanId")
			sampled = headers.Get("X-B3-Sampled") == "1"
		case "datadog":
			traceID = headers.Get("x-datadog-trace-id")
			spanID = headers.Get("x-datadog-parent-id")
			sampled = true // Datadog doesn't have explicit sampling header
		}
	}

	if traceID == "" {
		return nil
	}

	return &TraceContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: sampled,
		Baggage: make(map[string]string),
	}
}

// Helper methods

func (tm *TraceManager) getTraceContext(ctx context.Context) *TraceContext {
	if trace, ok := ctx.Value("trace").(*TraceContext); ok {
		return trace
	}
	return nil
}

func (tm *TraceManager) shouldSample() bool {
	// Simple sampling based on rate
	return time.Now().UnixNano()%100 < int64(tm.config.SamplingRate*100)
}

func (tm *TraceManager) storeTrace(trace *TraceInfo) {
	ctx := context.Background()
	data, _ := json.Marshal(trace)
	key := fmt.Sprintf("trace:%s", trace.TraceID)
	tm.redis.Set(ctx, key, string(data), 24*time.Hour)
}

func (tm *TraceManager) updateTrace(traceID, status string) {
	ctx := context.Background()
	key := fmt.Sprintf("trace:%s", traceID)

	data, err := tm.redis.Get(ctx, key).Result()
	if err != nil {
		return
	}

	var trace TraceInfo
	if err := json.Unmarshal([]byte(data), &trace); err != nil {
		return
	}

	trace.Status = status
	trace.EndTime = time.Now()
	trace.Duration = trace.EndTime.Sub(trace.StartTime)

	updatedData, _ := json.Marshal(trace)
	tm.redis.Set(ctx, key, string(updatedData), 24*time.Hour)
}

func (tm *TraceManager) loadTrace(traceID string) (*TraceInfo, error) {
	ctx := context.Background()
	key := fmt.Sprintf("trace:%s", traceID)

	data, err := tm.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("trace not found: %w", err)
	}

	var trace TraceInfo
	if err := json.Unmarshal([]byte(data), &trace); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trace: %w", err)
	}

	return &trace, nil
}

func (tm *TraceManager) fetchSpanSummary(ctx context.Context, traceID string) (*SpanSummary, error) {
	if tm.config.Endpoint == "" {
		return nil, fmt.Errorf("no tracing endpoint configured")
	}

	// Build request URL
	endpoint, _ := url.Parse(tm.config.Endpoint)
	endpoint.Path = fmt.Sprintf("/api/traces/%s", traceID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add auth if configured
	if tm.config.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tm.config.AuthToken))
	}

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch trace: %s", resp.Status)
	}

	// Parse response (simplified - actual format depends on provider)
	var summary SpanSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (tm *TraceManager) matchesFilter(trace *TraceInfo, filter *LogFilter) bool {
	if filter == nil {
		return true
	}

	// Check time range
	if !filter.StartTime.IsZero() && trace.StartTime.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && trace.StartTime.After(filter.EndTime) {
		return false
	}

	// Check trace IDs
	if len(filter.TraceIDs) > 0 {
		found := false
		for _, id := range filter.TraceIDs {
			if trace.TraceID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check search text
	if filter.SearchText != "" {
		searchLower := strings.ToLower(filter.SearchText)
		if !strings.Contains(strings.ToLower(trace.OperationName), searchLower) &&
			!strings.Contains(strings.ToLower(trace.ServiceName), searchLower) {
			return false
		}
	}

	return true
}

// LogTailer handles log tailing with backpressure protection
type LogTailer struct {
	config    *LoggingConfig
	redis     *redis.Client
	logger    *zap.Logger
	sessions  map[string]*TailSession
	mu        sync.RWMutex
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewLogTailer creates a new log tailer
func NewLogTailer(config *LoggingConfig, redis *redis.Client, logger *zap.Logger) *LogTailer {
	if config == nil {
		config = &LoggingConfig{
			Enabled:         true,
			RetentionPeriod: 7 * 24 * time.Hour,
			MaxStorageSize:  1024 * 1024 * 1024, // 1GB
		}
	}

	lt := &LogTailer{
		config:   config,
		redis:    redis,
		logger:   logger,
		sessions: make(map[string]*TailSession),
		stopCh:   make(chan struct{}),
	}

	// Start cleanup routine
	lt.wg.Add(1)
	go lt.cleanupLoop()

	return lt
}

// StartTail starts a new tailing session
func (lt *LogTailer) StartTail(config *TailConfig) (*TailSession, <-chan LogStreamEvent, error) {
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.MaxLinesPerSecond == 0 {
		config.MaxLinesPerSecond = 100
	}
	if config.BackpressureLimit == 0 {
		config.BackpressureLimit = 5000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 100 * time.Millisecond
	}

	session := &TailSession{
		ID:             uuid.New().String(),
		Config:         *config,
		StartedAt:      time.Now(),
		Connected:      true,
		LastActivity:   time.Now(),
		BackpressureStatus: BackpressureStatus{
			MaxRate: config.MaxLinesPerSecond,
		},
	}

	// Create event channel
	eventCh := make(chan LogStreamEvent, config.BufferSize)

	lt.mu.Lock()
	lt.sessions[session.ID] = session
	lt.mu.Unlock()

	// Start tailing goroutine
	lt.wg.Add(1)
	go lt.tailLoop(session, eventCh)

	return session, eventCh, nil
}

// StopTail stops a tailing session
func (lt *LogTailer) StopTail(sessionID string) error {
	lt.mu.Lock()
	session, exists := lt.sessions[sessionID]
	if exists {
		session.Connected = false
		delete(lt.sessions, sessionID)
	}
	lt.mu.Unlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// WriteLog writes a log entry
func (lt *LogTailer) WriteLog(entry *LogEntry) error {
	if !lt.config.Enabled {
		return nil
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Store in Redis
	ctx := context.Background()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Use sorted set for time-based queries
	key := fmt.Sprintf("logs:%s", time.Now().Format("2006-01-02"))
	score := float64(entry.Timestamp.UnixNano())

	if err := lt.redis.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: string(data),
	}).Err(); err != nil {
		return err
	}

	// Set expiration
	lt.redis.Expire(ctx, key, lt.config.RetentionPeriod)

	// Index by various fields for filtering
	lt.indexLog(entry)

	return nil
}

// SearchLogs searches for logs
func (lt *LogTailer) SearchLogs(ctx context.Context, filter *LogFilter) (*LogSearchResult, error) {
	result := &LogSearchResult{
		Logs:  make([]LogEntry, 0),
		Stats: &LogStats{
			LevelBreakdown: make(map[string]int64),
		},
	}

	// Determine date range
	startDate := filter.StartTime
	if startDate.IsZero() {
		startDate = time.Now().Add(-24 * time.Hour)
	}

	endDate := filter.EndTime
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Search each day's logs
	for date := startDate; date.Before(endDate); date = date.Add(24 * time.Hour) {
		key := fmt.Sprintf("logs:%s", date.Format("2006-01-02"))

		// Get logs within time range
		min := fmt.Sprintf("%d", startDate.UnixNano())
		max := fmt.Sprintf("%d", endDate.UnixNano())

		logs, err := lt.redis.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: min,
			Max: max,
		}).Result()

		if err != nil {
			continue
		}

		for _, logData := range logs {
			var entry LogEntry
			if err := json.Unmarshal([]byte(logData), &entry); err != nil {
				continue
			}

			// Apply filters
			if lt.matchesLogFilter(&entry, filter) {
				result.Logs = append(result.Logs, entry)

				// Update stats
				result.Stats.TotalLines++
				result.Stats.LevelBreakdown[entry.Level]++

				if filter.MaxResults > 0 && len(result.Logs) >= filter.MaxResults {
					result.HasMore = true
					goto done
				}
			}
		}
	}

done:
	result.TotalCount = len(result.Logs)
	return result, nil
}

// GetLogStats returns log statistics
func (lt *LogTailer) GetLogStats(ctx context.Context) (*LogStats, error) {
	stats := &LogStats{
		LevelBreakdown: make(map[string]int64),
	}

	// Get all log keys
	keys, err := lt.redis.Keys(ctx, "logs:*").Result()
	if err != nil {
		return nil, err
	}

	uniqueTraces := make(map[string]bool)
	uniqueJobs := make(map[string]bool)
	uniqueWorkers := make(map[string]bool)

	for _, key := range keys {
		// Get all logs for this key
		logs, err := lt.redis.ZRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}

		for _, logData := range logs {
			var entry LogEntry
			if err := json.Unmarshal([]byte(logData), &entry); err != nil {
				continue
			}

			stats.TotalLines++
			stats.LevelBreakdown[entry.Level]++

			if entry.Level == "error" {
				stats.ErrorCount++
			} else if entry.Level == "warning" || entry.Level == "warn" {
				stats.WarningCount++
			}

			if entry.TraceID != "" {
				uniqueTraces[entry.TraceID] = true
			}
			if entry.JobID != "" {
				uniqueJobs[entry.JobID] = true
			}
			if entry.WorkerID != "" {
				uniqueWorkers[entry.WorkerID] = true
			}

			if stats.OldestEntry.IsZero() || entry.Timestamp.Before(stats.OldestEntry) {
				stats.OldestEntry = entry.Timestamp
			}
			if entry.Timestamp.After(stats.NewestEntry) {
				stats.NewestEntry = entry.Timestamp
			}
		}
	}

	stats.UniqueTraces = len(uniqueTraces)
	stats.UniqueJobs = len(uniqueJobs)
	stats.UniqueWorkers = len(uniqueWorkers)

	// Calculate rate
	if !stats.OldestEntry.IsZero() && !stats.NewestEntry.IsZero() {
		duration := stats.NewestEntry.Sub(stats.OldestEntry).Seconds()
		if duration > 0 {
			stats.LinesPerSecond = float64(stats.TotalLines) / duration
		}
	}

	return stats, nil
}

func (lt *LogTailer) tailLoop(session *TailSession, eventCh chan LogStreamEvent) {
	defer lt.wg.Done()
	defer close(eventCh)

	ticker := time.NewTicker(session.Config.FlushInterval)
	defer ticker.Stop()

	rateLimiter := NewRateLimiter(float64(session.Config.MaxLinesPerSecond))
	buffer := make([]LogEntry, 0, session.Config.BufferSize)

	var linesProcessed int64
	var droppedLines int64

	ctx := context.Background()

	for session.Connected {
		select {
		case <-lt.stopCh:
			return
		case <-ticker.C:
			// Fetch new logs
			logs, err := lt.fetchNewLogs(ctx, session, &buffer)
			if err != nil {
				eventCh <- LogStreamEvent{
					Type:      "error",
					Timestamp: time.Now(),
					Data:      err.Error(),
				}
				continue
			}

			// Apply backpressure
			if len(buffer) > session.Config.BackpressureLimit {
				if !session.BackpressureStatus.Active {
					session.BackpressureStatus.Active = true
					session.BackpressureStatus.LastActivated = time.Now()

					eventCh <- LogStreamEvent{
						Type:      "backpressure",
						Timestamp: time.Now(),
						Data:      session.BackpressureStatus,
					}
				}

				// Drop oldest logs
				dropCount := len(buffer) - session.Config.BufferSize
				buffer = buffer[dropCount:]
				atomic.AddInt64(&droppedLines, int64(dropCount))
			} else if session.BackpressureStatus.Active && len(buffer) < session.Config.BufferSize/2 {
				// Deactivate backpressure
				session.BackpressureStatus.Active = false
			}

			// Send logs with rate limiting
			for _, log := range logs {
				if !rateLimiter.Allow() {
					atomic.AddInt64(&droppedLines, 1)
					continue
				}

				select {
				case eventCh <- LogStreamEvent{
					Type:      "log",
					Timestamp: time.Now(),
					Data:      log,
				}:
					atomic.AddInt64(&linesProcessed, 1)
				default:
					// Channel full, drop log
					atomic.AddInt64(&droppedLines, 1)
				}
			}

			// Update session stats
			lt.mu.Lock()
			session.LinesProcessed = atomic.LoadInt64(&linesProcessed)
			session.BackpressureStatus.DroppedLines = atomic.LoadInt64(&droppedLines)
			session.BackpressureStatus.BufferUsage = float64(len(buffer)) / float64(session.Config.BufferSize) * 100
			session.LastActivity = time.Now()
			lt.mu.Unlock()

			// Send status update
			if time.Since(session.LastActivity) > 5*time.Second {
				eventCh <- LogStreamEvent{
					Type:      "status",
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"lines_processed": linesProcessed,
						"dropped_lines":   droppedLines,
						"buffer_usage":    session.BackpressureStatus.BufferUsage,
					},
				}
			}
		}
	}
}

func (lt *LogTailer) fetchNewLogs(ctx context.Context, session *TailSession, buffer *[]LogEntry) ([]LogEntry, error) {
	// Get current date key
	key := fmt.Sprintf("logs:%s", time.Now().Format("2006-01-02"))

	// Get logs since last fetch
	min := fmt.Sprintf("%d", session.LastActivity.UnixNano())
	max := "+inf"

	logs, err := lt.redis.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()

	if err != nil {
		return nil, err
	}

	result := make([]LogEntry, 0, len(logs))
	for _, logData := range logs {
		var entry LogEntry
		if err := json.Unmarshal([]byte(logData), &entry); err != nil {
			continue
		}

		// Apply filter
		if lt.matchesLogFilter(&entry, session.Config.Filter) {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (lt *LogTailer) matchesLogFilter(entry *LogEntry, filter *LogFilter) bool {
	if filter == nil {
		return true
	}

	// Check levels
	if len(filter.Levels) > 0 {
		found := false
		for _, level := range filter.Levels {
			if strings.EqualFold(entry.Level, level) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check sources
	if len(filter.Sources) > 0 && !contains(filter.Sources, entry.Source) {
		return false
	}

	// Check job IDs
	if len(filter.JobIDs) > 0 && !contains(filter.JobIDs, entry.JobID) {
		return false
	}

	// Check worker IDs
	if len(filter.WorkerIDs) > 0 && !contains(filter.WorkerIDs, entry.WorkerID) {
		return false
	}

	// Check queue names
	if len(filter.QueueNames) > 0 && !contains(filter.QueueNames, entry.QueueName) {
		return false
	}

	// Check trace IDs
	if len(filter.TraceIDs) > 0 && !contains(filter.TraceIDs, entry.TraceID) {
		return false
	}

	// Check search text
	if filter.SearchText != "" {
		searchLower := strings.ToLower(filter.SearchText)
		if !strings.Contains(strings.ToLower(entry.Message), searchLower) {
			// Check fields
			found := false
			for _, v := range entry.Fields {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", v)), searchLower) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

func (lt *LogTailer) indexLog(entry *LogEntry) {
	ctx := context.Background()

	// Index by trace ID
	if entry.TraceID != "" {
		lt.redis.SAdd(ctx, fmt.Sprintf("log:trace:%s", entry.TraceID), entry.Timestamp.UnixNano())
	}

	// Index by job ID
	if entry.JobID != "" {
		lt.redis.SAdd(ctx, fmt.Sprintf("log:job:%s", entry.JobID), entry.Timestamp.UnixNano())
	}

	// Index by worker ID
	if entry.WorkerID != "" {
		lt.redis.SAdd(ctx, fmt.Sprintf("log:worker:%s", entry.WorkerID), entry.Timestamp.UnixNano())
	}
}

func (lt *LogTailer) cleanupLoop() {
	defer lt.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-lt.stopCh:
			return
		case <-ticker.C:
			lt.cleanup()
		}
	}
}

func (lt *LogTailer) cleanup() {
	ctx := context.Background()

	// Remove old logs
	cutoff := time.Now().Add(-lt.config.RetentionPeriod)

	// Get all log keys
	keys, err := lt.redis.Keys(ctx, "logs:*").Result()
	if err != nil {
		return
	}

	for _, key := range keys {
		// Parse date from key
		parts := strings.Split(key, ":")
		if len(parts) < 2 {
			continue
		}

		date, err := time.Parse("2006-01-02", parts[1])
		if err != nil {
			continue
		}

		// Delete if too old
		if date.Before(cutoff) {
			lt.redis.Del(ctx, key)
		}
	}

	// Clean up disconnected sessions
	lt.mu.Lock()
	for id, session := range lt.sessions {
		if time.Since(session.LastActivity) > 5*time.Minute {
			delete(lt.sessions, id)
		}
	}
	lt.mu.Unlock()
}

// Shutdown gracefully shuts down the log tailer
func (lt *LogTailer) Shutdown() {
	close(lt.stopCh)
	lt.wg.Wait()
}

// Helper functions

func generateTraceID() string {
	return uuid.New().String()
}

func generateSpanID() string {
	return uuid.New().String()[:8]
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RateLimiter provides rate limiting for log processing
type RateLimiter struct {
	rate       float64
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		tokens:     rate,
		lastUpdate: time.Now(),
	}
}

// Allow checks if an operation is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens = min(rl.rate, rl.tokens+rl.rate*elapsed)
	rl.lastUpdate = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}