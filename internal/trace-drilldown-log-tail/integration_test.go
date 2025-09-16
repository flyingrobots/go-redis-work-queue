// Copyright 2025 James Ross
//go:build integration
// +build integration

package tracedrilldownlogtail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/distributed-tracing-integration"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestIntegration_TraceToLogCorrelation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	// Setup trace manager
	tracingConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "integration-test",
		SamplingRate: 1.0,
		URLTemplate:  "http://localhost:16686/trace/{trace_id}",
	}
	traceManager := NewTraceManager(tracingConfig, rdb, logger)

	// Setup log tailer
	loggingConfig := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}
	logTailer := NewLogTailer(loggingConfig, rdb, logger)
	defer logTailer.Shutdown()

	ctx := context.Background()

	// Start a trace
	traceCtx, newCtx := traceManager.StartTrace(ctx, "integration-test-operation")
	require.NotNil(t, traceCtx)

	// Add some logs to the trace
	traceManager.AddTraceLog(newCtx, "info", "Starting integration test", map[string]interface{}{
		"test_id": "integration-001",
	})

	// Write corresponding log entries
	logEntries := []*LogEntry{
		{
			Level:   "info",
			Message: "Processing job",
			Source:  "worker",
			JobID:   "job-123",
			TraceID: traceCtx.TraceID,
			SpanID:  traceCtx.SpanID,
		},
		{
			Level:   "debug",
			Message: "Job processing details",
			Source:  "worker",
			JobID:   "job-123",
			TraceID: traceCtx.TraceID,
			SpanID:  traceCtx.SpanID,
		},
		{
			Level:   "info",
			Message: "Job completed successfully",
			Source:  "worker",
			JobID:   "job-123",
			TraceID: traceCtx.TraceID,
			SpanID:  traceCtx.SpanID,
		},
	}

	for _, entry := range logEntries {
		err := logTailer.WriteLog(entry)
		require.NoError(t, err)
	}

	// End the trace
	time.Sleep(10 * time.Millisecond)
	traceManager.EndTrace(newCtx, "completed")

	// Test correlation: search logs by trace ID
	filter := &LogFilter{
		TraceIDs:   []string{traceCtx.TraceID},
		MaxResults: 10,
	}

	logResult, err := logTailer.SearchLogs(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 3, len(logResult.Logs))
	assert.Equal(t, 3, logResult.TotalCount)

	// Verify all logs have the correct trace ID
	for _, log := range logResult.Logs {
		assert.Equal(t, traceCtx.TraceID, log.TraceID)
		assert.Equal(t, traceCtx.SpanID, log.SpanID)
		assert.Equal(t, "job-123", log.JobID)
	}

	// Test trace retrieval
	trace, err := traceManager.GetTrace(traceCtx.TraceID)
	require.NoError(t, err)
	assert.Equal(t, "completed", trace.Status)
	assert.Equal(t, "integration-test-operation", trace.OperationName)
	assert.Len(t, trace.Logs, 1) // One log added via AddTraceLog
}

func TestIntegration_LogTailWithBackpressure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	config := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 1 * time.Hour,
	}

	logTailer := NewLogTailer(config, rdb, logger)
	defer logTailer.Shutdown()

	// Start a tail session with low rate limits
	tailConfig := &TailConfig{
		Follow:            true,
		BufferSize:        10,
		MaxLinesPerSecond: 5,
		BackpressureLimit: 8,
		FlushInterval:     50 * time.Millisecond,
	}

	session, eventCh, err := logTailer.StartTail(tailConfig)
	require.NoError(t, err)

	// Write logs rapidly to trigger backpressure
	go func() {
		for i := 0; i < 20; i++ {
			entry := &LogEntry{
				Level:   "info",
				Message: fmt.Sprintf("Test message %d", i),
				Source:  "backpressure-test",
				JobID:   fmt.Sprintf("job-%d", i),
			}
			_ = logTailer.WriteLog(entry)
			time.Sleep(5 * time.Millisecond) // Fast generation
		}
	}()

	// Collect events
	var events []LogStreamEvent
	timeout := time.After(2 * time.Second)
	done := false

	for !done {
		select {
		case event := <-eventCh:
			events = append(events, event)
			if event.Type == "backpressure" {
				// Backpressure was triggered
				done = true
			}
		case <-timeout:
			done = true
		}
	}

	// Stop the session
	err = logTailer.StopTail(session.ID)
	assert.NoError(t, err)

	// Verify backpressure was triggered
	hasBackpressureEvent := false
	for _, event := range events {
		if event.Type == "backpressure" {
			hasBackpressureEvent = true
			break
		}
	}

	assert.True(t, hasBackpressureEvent, "Expected backpressure event to be triggered")
}

func TestIntegration_HTTPHandlers(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	// Setup components
	tracingConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "http-test",
		SamplingRate: 1.0,
		URLTemplate:  "http://localhost:16686/trace/{trace_id}",
	}
	traceManager := NewTraceManager(tracingConfig, rdb, logger)

	loggingConfig := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}
	logTailer := NewLogTailer(loggingConfig, rdb, logger)
	defer logTailer.Shutdown()

	// Create handlers
	handlers := NewHTTPHandlers(traceManager, logTailer, logger)

	// Setup router
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	// Create test data
	ctx := context.Background()
	traceCtx, newCtx := traceManager.StartTrace(ctx, "http-test-operation")
	traceManager.EndTrace(newCtx, "completed")

	// Test get trace endpoint
	t.Run("GetTrace", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/trace-drilldown/traces/%s", traceCtx.TraceID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), traceCtx.TraceID)
		assert.Contains(t, rr.Body.String(), "http-test-operation")
	})

	// Test get trace links endpoint
	t.Run("GetTraceLinks", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/trace-drilldown/traces/%s/links", traceCtx.TraceID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), traceCtx.TraceID)
		assert.Contains(t, rr.Body.String(), "localhost:16686")
	})

	// Test open trace endpoint
	t.Run("OpenTrace", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/trace-drilldown/traces/%s/open", traceCtx.TraceID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), traceCtx.TraceID)
		assert.Contains(t, rr.Body.String(), "url")
	})

	// Test search traces endpoint
	t.Run("SearchTraces", func(t *testing.T) {
		body := strings.NewReader(`{"search_text": "http-test", "max_results": 10}`)
		req := httptest.NewRequest("POST", "/api/v1/trace-drilldown/traces/search", body)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "traces")
	})

	// Test log stats endpoint
	t.Run("GetLogStats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/trace-drilldown/logs/stats", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "total_lines")
	})
}

func TestIntegration_EnhancedAdmin(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	// Setup tracing integration
	tracingConfig := distributed_tracing_integration.DefaultTracingUIConfig()
	tracingIntegration := distributed_tracing_integration.New(tracingConfig)

	// Setup trace manager
	traceManagerConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "enhanced-admin-test",
		SamplingRate: 1.0,
		URLTemplate:  "http://localhost:16686/trace/{trace_id}",
	}
	traceManager := NewTraceManager(traceManagerConfig, rdb, logger)

	// Create enhanced admin
	enhancedAdmin := NewEnhancedAdmin(tracingIntegration, traceManager, logger)

	// Create test job data with trace
	jobData := `{
		"id": "test-job-123",
		"filepath": "/test/file.txt",
		"filesize": 1024,
		"priority": "high",
		"retries": 0,
		"creation_time": "2023-01-01T12:00:00Z",
		"trace_id": "test-trace-456",
		"span_id": "test-span-789"
	}`

	// Test parsing job with trace
	jobInfo, err := enhancedAdmin.parseJobWithTrace(jobData)
	require.NoError(t, err)
	assert.Equal(t, "test-job-123", jobInfo.JobID)
	assert.Equal(t, "/test/file.txt", jobInfo.FilePath)
	assert.Equal(t, "high", jobInfo.Priority)
	assert.Equal(t, "test-trace-456", jobInfo.TraceID)
	assert.Equal(t, "test-span-789", jobInfo.SpanID)

	// Test getting trace actions
	actions, err := enhancedAdmin.GetJobTraceActions(jobData)
	require.NoError(t, err)
	assert.NotEmpty(t, actions)

	// Should have at least view and copy actions
	hasViewAction := false
	hasCopyAction := false
	for _, action := range actions {
		if action.Type == "view" {
			hasViewAction = true
		}
		if action.Type == "copy" {
			hasCopyAction = true
		}
	}
	assert.True(t, hasViewAction)
	assert.True(t, hasCopyAction)

	// Test opening trace
	result, err := enhancedAdmin.OpenJobTrace(jobData)
	require.NoError(t, err)
	assert.Equal(t, "test-job-123", result.JobID)
	assert.Equal(t, "test-trace-456", result.TraceID)
	assert.Equal(t, "open", result.Action)
	assert.NotEmpty(t, result.URL)
	assert.True(t, result.Success)
}

func TestIntegration_EndToEnd_TraceFlow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	// This test simulates the full flow:
	// 1. Job created with trace
	// 2. Worker processes job (generating logs)
	// 3. Admin views job with trace actions
	// 4. User searches logs by trace ID

	// Setup all components
	tracingConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "e2e-test",
		SamplingRate: 1.0,
		URLTemplate:  "http://localhost:16686/trace/{trace_id}",
	}
	traceManager := NewTraceManager(tracingConfig, rdb, logger)

	loggingConfig := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
	}
	logTailer := NewLogTailer(loggingConfig, rdb, logger)
	defer logTailer.Shutdown()

	ctx := context.Background()

	// Step 1: Create a job with trace context
	jobCtx, jobWithTrace := traceManager.StartTrace(ctx, "job.process")
	jobData := fmt.Sprintf(`{
		"id": "e2e-job-123",
		"filepath": "/data/file.txt",
		"filesize": 2048,
		"priority": "high",
		"retries": 0,
		"creation_time": "%s",
		"trace_id": "%s",
		"span_id": "%s"
	}`, time.Now().UTC().Format(time.RFC3339Nano), jobCtx.TraceID, jobCtx.SpanID)

	// Step 2: Simulate job processing with logs
	logEntries := []*LogEntry{
		{
			Level:   "info",
			Message: "Job started",
			Source:  "worker",
			JobID:   "e2e-job-123",
			TraceID: jobCtx.TraceID,
			SpanID:  jobCtx.SpanID,
		},
		{
			Level:   "debug",
			Message: "Reading file /data/file.txt",
			Source:  "worker",
			JobID:   "e2e-job-123",
			TraceID: jobCtx.TraceID,
			SpanID:  jobCtx.SpanID,
		},
		{
			Level:   "info",
			Message: "File processed successfully",
			Source:  "worker",
			JobID:   "e2e-job-123",
			TraceID: jobCtx.TraceID,
			SpanID:  jobCtx.SpanID,
		},
	}

	for _, entry := range logEntries {
		err := logTailer.WriteLog(entry)
		require.NoError(t, err)
	}

	// Add trace logs
	traceManager.AddTraceLog(jobWithTrace, "info", "Processing started", map[string]interface{}{
		"file_size": 2048,
	})
	traceManager.AddTraceLog(jobWithTrace, "info", "Processing completed", map[string]interface{}{
		"duration_ms": 150,
	})

	// Step 3: End the trace
	time.Sleep(10 * time.Millisecond)
	traceManager.EndTrace(jobWithTrace, "success")

	// Step 4: Admin views the job (simulated)
	tracingIntegration := distributed_tracing_integration.NewWithDefaults()
	enhancedAdmin := NewEnhancedAdmin(tracingIntegration, traceManager, logger)

	// Parse job and get trace actions
	jobInfo, err := enhancedAdmin.parseJobWithTrace(jobData)
	require.NoError(t, err)
	assert.Equal(t, jobCtx.TraceID, jobInfo.TraceID)

	actions, err := enhancedAdmin.GetJobTraceActions(jobData)
	require.NoError(t, err)
	assert.NotEmpty(t, actions)

	// Step 5: Search logs by trace ID
	filter := &LogFilter{
		TraceIDs:   []string{jobCtx.TraceID},
		MaxResults: 10,
	}

	logResult, err := logTailer.SearchLogs(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 3, logResult.TotalCount)
	assert.Len(t, logResult.Logs, 3)

	// Verify log correlation
	for _, log := range logResult.Logs {
		assert.Equal(t, jobCtx.TraceID, log.TraceID)
		assert.Equal(t, "e2e-job-123", log.JobID)
	}

	// Step 6: Get trace summary
	summary, err := traceManager.GetSpanSummary(ctx, jobCtx.TraceID)
	require.NoError(t, err)
	assert.Equal(t, jobCtx.TraceID, summary.TraceID)
	assert.Greater(t, summary.Duration, time.Duration(0))

	// Step 7: Verify trace retrieval
	trace, err := traceManager.GetTrace(jobCtx.TraceID)
	require.NoError(t, err)
	assert.Equal(t, "success", trace.Status)
	assert.Equal(t, "job.process", trace.OperationName)
	assert.Len(t, trace.Logs, 2) // Two logs added via AddTraceLog

	t.Logf("End-to-end test completed successfully")
	t.Logf("Trace ID: %s", jobCtx.TraceID)
	t.Logf("Found %d correlated log entries", len(logResult.Logs))
	t.Logf("Trace duration: %v", trace.Duration)
}

func TestIntegration_PerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := zaptest.NewLogger(t)
	rdb := setupTestRedis(t)
	defer rdb.Close()

	// Setup components
	tracingConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "perf-test",
		SamplingRate: 0.1, // Sample only 10% for performance
	}
	traceManager := NewTraceManager(tracingConfig, rdb, logger)

	loggingConfig := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 1 * time.Hour,
	}
	logTailer := NewLogTailer(loggingConfig, rdb, logger)
	defer logTailer.Shutdown()

	ctx := context.Background()
	numOperations := 1000

	start := time.Now()

	// Generate load
	for i := 0; i < numOperations; i++ {
		// Start trace
		traceCtx, newCtx := traceManager.StartTrace(ctx, fmt.Sprintf("perf-operation-%d", i))

		// Write logs
		entry := &LogEntry{
			Level:   "info",
			Message: fmt.Sprintf("Performance test message %d", i),
			Source:  "perf-test",
			JobID:   fmt.Sprintf("perf-job-%d", i),
			TraceID: traceCtx.TraceID,
		}
		_ = logTailer.WriteLog(entry)

		// End trace
		traceManager.EndTrace(newCtx, "completed")

		// Add some variation
		if i%100 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	duration := time.Since(start)
	throughput := float64(numOperations) / duration.Seconds()

	t.Logf("Performance test completed:")
	t.Logf("Operations: %d", numOperations)
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f ops/sec", throughput)

	// Verify we can still search effectively
	filter := &LogFilter{
		Sources:    []string{"perf-test"},
		MaxResults: 100,
	}

	searchStart := time.Now()
	result, err := logTailer.SearchLogs(ctx, filter)
	searchDuration := time.Since(searchStart)

	require.NoError(t, err)
	assert.Greater(t, len(result.Logs), 90) // Should find most logs
	assert.Less(t, searchDuration, 1*time.Second) // Search should be fast

	t.Logf("Search found %d logs in %v", len(result.Logs), searchDuration)

	// Performance assertions
	assert.Greater(t, throughput, 100.0, "Should handle at least 100 ops/sec")
	assert.Less(t, searchDuration, 500*time.Millisecond, "Search should complete in < 500ms")
}

// Helper functions

func setupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available for integration testing")
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

	keys, _ = rdb.Keys(ctx, "log:*").Result()
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}

	return rdb
}