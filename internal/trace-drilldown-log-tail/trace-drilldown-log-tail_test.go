// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestTraceManager_StartTrace(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "test-service",
		SamplingRate: 1.0,
		URLTemplate:  "http://localhost:16686/trace/{trace_id}",
	}

	tm := NewTraceManager(config, rdb, logger)

	ctx := context.Background()
	traceCtx, newCtx := tm.StartTrace(ctx, "test-operation")

	assert.NotNil(t, traceCtx)
	assert.NotEmpty(t, traceCtx.TraceID)
	assert.NotEmpty(t, traceCtx.SpanID)
	assert.True(t, traceCtx.Sampled)
	assert.NotNil(t, newCtx)

	// Verify trace is stored
	trace, err := tm.GetTrace(traceCtx.TraceID)
	assert.NoError(t, err)
	assert.Equal(t, traceCtx.TraceID, trace.TraceID)
	assert.Equal(t, traceCtx.SpanID, trace.SpanID)
	assert.Equal(t, "test-operation", trace.OperationName)
	assert.Equal(t, "test-service", trace.ServiceName)
	assert.Equal(t, "active", trace.Status)
}

func TestTraceManager_EndTrace(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "test-service",
		SamplingRate: 1.0,
	}

	tm := NewTraceManager(config, rdb, logger)

	ctx := context.Background()
	traceCtx, newCtx := tm.StartTrace(ctx, "test-operation")

	// End the trace
	time.Sleep(10 * time.Millisecond) // Small delay to ensure duration > 0
	tm.EndTrace(newCtx, "completed")

	// Verify trace is updated
	trace, err := tm.GetTrace(traceCtx.TraceID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", trace.Status)
	assert.False(t, trace.EndTime.IsZero())
	assert.Greater(t, trace.Duration, time.Duration(0))
}

func TestTraceManager_AddTraceLog(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "test-service",
		SamplingRate: 1.0,
	}

	tm := NewTraceManager(config, rdb, logger)

	ctx := context.Background()
	_, newCtx := tm.StartTrace(ctx, "test-operation")

	// Add a log
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	tm.AddTraceLog(newCtx, "info", "Test log message", fields)

	// Verify log is added
	traceCtx := tm.getTraceContext(newCtx)
	trace, err := tm.GetTrace(traceCtx.TraceID)
	assert.NoError(t, err)
	assert.Len(t, trace.Logs, 1)
	assert.Equal(t, "info", trace.Logs[0].Level)
	assert.Equal(t, "Test log message", trace.Logs[0].Message)
	assert.Equal(t, fields, trace.Logs[0].Fields)
}

func TestTraceManager_GetTraceLink(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:     true,
		Provider:    "jaeger",
		ServiceName: "test-service",
		URLTemplate: "http://localhost:16686/trace/{trace_id}",
	}

	tm := NewTraceManager(config, rdb, logger)

	traceID := "test-trace-id-123"
	link, err := tm.GetTraceLink(traceID)

	assert.NoError(t, err)
	assert.Equal(t, "jaeger", link.Type)
	assert.Equal(t, "http://localhost:16686/trace/test-trace-id-123", link.URL)
	assert.Equal(t, "View in jaeger", link.DisplayName)
}

func TestTraceManager_PropagateTrace(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:     true,
		Provider:    "jaeger",
		ServiceName: "test-service",
	}

	tm := NewTraceManager(config, rdb, logger)

	ctx := context.Background()
	traceCtx, newCtx := tm.StartTrace(ctx, "test-operation")

	// Test Jaeger propagation
	headers := make(map[string][]string)
	tm.PropagateTrace(newCtx, headers)

	assert.Equal(t, traceCtx.TraceID, headers["X-Trace-Id"][0])
	assert.Equal(t, traceCtx.SpanID, headers["X-Span-Id"][0])
	assert.Equal(t, "true", headers["X-Sampled"][0])
	assert.Contains(t, headers["uber-trace-id"][0], traceCtx.TraceID)
	assert.Contains(t, headers["uber-trace-id"][0], traceCtx.SpanID)
}

func TestLogTailer_WriteLog(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}

	lt := NewLogTailer(config, rdb, logger)
	defer lt.Shutdown()

	entry := &LogEntry{
		Level:     "info",
		Message:   "Test log message",
		Source:    "test",
		JobID:     "job-123",
		WorkerID:  "worker-456",
		QueueName: "test-queue",
		TraceID:   "trace-789",
		SpanID:    "span-abc",
		Fields: map[string]interface{}{
			"custom": "field",
		},
	}

	err := lt.WriteLog(entry)
	assert.NoError(t, err)
	assert.False(t, entry.Timestamp.IsZero())

	// Verify log is stored in Redis
	key := "logs:" + time.Now().Format("2006-01-02")
	count, err := rdb.ZCard(context.Background(), key).Result()
	assert.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestLogTailer_SearchLogs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}

	lt := NewLogTailer(config, rdb, logger)
	defer lt.Shutdown()

	// Write test logs
	entries := []*LogEntry{
		{
			Level:    "info",
			Message:  "Info message",
			Source:   "test",
			TraceID:  "trace-123",
		},
		{
			Level:    "error",
			Message:  "Error message",
			Source:   "test",
			TraceID:  "trace-456",
		},
		{
			Level:    "debug",
			Message:  "Debug message",
			Source:   "other",
			TraceID:  "trace-123",
		},
	}

	for _, entry := range entries {
		err := lt.WriteLog(entry)
		require.NoError(t, err)
	}

	ctx := context.Background()

	// Test search by level
	filter := &LogFilter{
		Levels:     []string{"error"},
		MaxResults: 10,
	}
	result, err := lt.SearchLogs(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Logs))
	assert.Equal(t, "error", result.Logs[0].Level)
	assert.Equal(t, "Error message", result.Logs[0].Message)

	// Test search by trace ID
	filter = &LogFilter{
		TraceIDs:   []string{"trace-123"},
		MaxResults: 10,
	}
	result, err = lt.SearchLogs(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Logs))

	// Test search by source
	filter = &LogFilter{
		Sources:    []string{"other"},
		MaxResults: 10,
	}
	result, err = lt.SearchLogs(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Logs))
	assert.Equal(t, "other", result.Logs[0].Source)
}

func TestLogTailer_StartTail(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}

	lt := NewLogTailer(config, rdb, logger)
	defer lt.Shutdown()

	tailConfig := &TailConfig{
		Follow:            true,
		BufferSize:        100,
		MaxLinesPerSecond: 10,
		FlushInterval:     100 * time.Millisecond,
	}

	session, eventCh, err := lt.StartTail(tailConfig)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotNil(t, eventCh)
	assert.NotEmpty(t, session.ID)
	assert.True(t, session.Connected)

	// Stop the tail session
	err = lt.StopTail(session.ID)
	assert.NoError(t, err)
	assert.False(t, session.Connected)
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(2.0) // 2 tokens per second

	// Should allow first two requests immediately
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())

	// Third request should be denied
	assert.False(t, limiter.Allow())

	// Wait and try again
	time.Sleep(600 * time.Millisecond) // Should refill one token
	assert.True(t, limiter.Allow())
	assert.False(t, limiter.Allow())
}

func TestTraceSearch(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "test-service",
		SamplingRate: 1.0,
	}

	tm := NewTraceManager(config, rdb, logger)

	ctx := context.Background()

	// Create some test traces
	traces := []struct {
		operation string
		status    string
	}{
		{"user.login", "completed"},
		{"user.logout", "completed"},
		{"payment.process", "error"},
	}

	for _, tr := range traces {
		traceCtx, newCtx := tm.StartTrace(ctx, tr.operation)
		time.Sleep(10 * time.Millisecond)
		tm.EndTrace(newCtx, tr.status)

		// Store trace context for later reference
		_ = traceCtx
	}

	// Search for traces
	filter := &LogFilter{
		SearchText: "user",
		MaxResults: 10,
	}

	result, err := tm.SearchTraces(ctx, filter)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Traces), 2) // Should find user.login and user.logout
}

func TestMatchesLogFilter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled: true,
	}

	lt := NewLogTailer(config, rdb, logger)

	entry := &LogEntry{
		Level:     "error",
		Message:   "Test error message",
		Source:    "test-source",
		JobID:     "job-123",
		WorkerID:  "worker-456",
		QueueName: "test-queue",
		TraceID:   "trace-789",
		Fields: map[string]interface{}{
			"custom": "value",
		},
	}

	// Test level filter
	filter := &LogFilter{Levels: []string{"error"}}
	assert.True(t, lt.matchesLogFilter(entry, filter))

	filter = &LogFilter{Levels: []string{"info"}}
	assert.False(t, lt.matchesLogFilter(entry, filter))

	// Test source filter
	filter = &LogFilter{Sources: []string{"test-source"}}
	assert.True(t, lt.matchesLogFilter(entry, filter))

	filter = &LogFilter{Sources: []string{"other-source"}}
	assert.False(t, lt.matchesLogFilter(entry, filter))

	// Test job ID filter
	filter = &LogFilter{JobIDs: []string{"job-123"}}
	assert.True(t, lt.matchesLogFilter(entry, filter))

	filter = &LogFilter{JobIDs: []string{"job-456"}}
	assert.False(t, lt.matchesLogFilter(entry, filter))

	// Test search text filter
	filter = &LogFilter{SearchText: "error"}
	assert.True(t, lt.matchesLogFilter(entry, filter))

	filter = &LogFilter{SearchText: "success"}
	assert.False(t, lt.matchesLogFilter(entry, filter))

	// Test multiple filters (AND logic)
	filter = &LogFilter{
		Levels:     []string{"error"},
		Sources:    []string{"test-source"},
		SearchText: "Test",
	}
	assert.True(t, lt.matchesLogFilter(entry, filter))

	filter = &LogFilter{
		Levels:     []string{"info"}, // This should fail
		Sources:    []string{"test-source"},
		SearchText: "Test",
	}
	assert.False(t, lt.matchesLogFilter(entry, filter))
}

func TestBackpressureManager(t *testing.T) {
	config := DefaultConfig{
		MaxLinesPerSecond: 5,
	}

	bm := NewBackpressureManager(config)

	// Should allow up to rate limit
	for i := 0; i < 5; i++ {
		assert.False(t, bm.ShouldDrop(), "Request %d should be allowed", i)
	}

	// Should start dropping after rate limit
	assert.True(t, bm.ShouldDrop())
	bm.RecordDrop()

	// Check status
	status := bm.GetStatus()
	assert.True(t, status.Active)
	assert.Equal(t, int64(1), status.DroppedLines)
	assert.Equal(t, 5, status.CurrentRate)
	assert.Equal(t, 5, status.MaxRate)

	// Wait for rate to reset
	time.Sleep(1100 * time.Millisecond)
	assert.False(t, bm.ShouldDrop())

	status = bm.GetStatus()
	assert.False(t, status.Active)
}

func TestTraceInfo_JSON(t *testing.T) {
	traceInfo := &TraceInfo{
		TraceID:       "trace-123",
		SpanID:        "span-456",
		ServiceName:   "test-service",
		OperationName: "test-operation",
		StartTime:     time.Now(),
		Status:        "active",
		Tags: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(traceInfo)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaled TraceInfo
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, traceInfo.TraceID, unmarshaled.TraceID)
	assert.Equal(t, traceInfo.SpanID, unmarshaled.SpanID)
	assert.Equal(t, traceInfo.ServiceName, unmarshaled.ServiceName)
	assert.Equal(t, traceInfo.OperationName, unmarshaled.OperationName)
	assert.Equal(t, traceInfo.Status, unmarshaled.Status)
	assert.Equal(t, traceInfo.Tags, unmarshaled.Tags)
}

func TestLogEntry_JSON(t *testing.T) {
	logEntry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Test message",
		Source:    "test",
		JobID:     "job-123",
		WorkerID:  "worker-456",
		QueueName: "test-queue",
		TraceID:   "trace-789",
		SpanID:    "span-abc",
		Fields: map[string]interface{}{
			"custom": "field",
			"number": 42,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(logEntry)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaled LogEntry
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, logEntry.Level, unmarshaled.Level)
	assert.Equal(t, logEntry.Message, unmarshaled.Message)
	assert.Equal(t, logEntry.Source, unmarshaled.Source)
	assert.Equal(t, logEntry.JobID, unmarshaled.JobID)
	assert.Equal(t, logEntry.WorkerID, unmarshaled.WorkerID)
	assert.Equal(t, logEntry.QueueName, unmarshaled.QueueName)
	assert.Equal(t, logEntry.TraceID, unmarshaled.TraceID)
	assert.Equal(t, logEntry.SpanID, unmarshaled.SpanID)
	assert.Equal(t, logEntry.Fields, unmarshaled.Fields)
}

// Helper function to create a test Redis client
func createTestRedisClient(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available for testing")
	}

	// Clean up any existing test data
	keys, _ := rdb.Keys(ctx, "trace:*").Result()
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}

	keys, _ = rdb.Keys(ctx, "logs:*").Result()
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}

	return rdb
}

// Benchmark tests
func BenchmarkTraceManager_StartTrace(b *testing.B) {
	logger := zaptest.NewLogger(b)
	rdb := createTestRedisClient(b.(*testing.T))
	defer rdb.Close()

	config := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "bench-service",
		SamplingRate: 1.0,
	}

	tm := NewTraceManager(config, rdb, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tm.StartTrace(ctx, "bench-operation")
	}
}

func BenchmarkLogTailer_WriteLog(b *testing.B) {
	logger := zaptest.NewLogger(b)
	rdb := createTestRedisClient(b.(*testing.T))
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}

	lt := NewLogTailer(config, rdb, logger)
	defer lt.Shutdown()

	entry := &LogEntry{
		Level:     "info",
		Message:   "Benchmark log message",
		Source:    "bench",
		TraceID:   "bench-trace",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lt.WriteLog(entry)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000.0) // High rate for benchmarking

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}