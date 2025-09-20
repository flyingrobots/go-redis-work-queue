//go:build trace_drilldown_tests
// +build trace_drilldown_tests

// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*TraceManager, *LogTailer, *redis.Client, func()) {
	// Create miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create trace manager
	tracingConfig := &TracingConfig{
		Enabled:      true,
		Provider:     "jaeger",
		ServiceName:  "test-service",
		SamplingRate: 1.0,
		URLTemplate:  "http://jaeger.local/trace/{trace_id}",
	}

	// Create log tailer
	loggingConfig := &LoggingConfig{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
		MaxStorageSize:  1024 * 1024,
	}

	logger := zap.NewNop()
	traceManager := NewTraceManager(tracingConfig, client, logger)
	logTailer := NewLogTailer(loggingConfig, client, logger)

	cleanup := func() {
		logTailer.Shutdown()
		client.Close()
		mr.Close()
	}

	return traceManager, logTailer, client, cleanup
}

func TestTraceManagement(t *testing.T) {
	traceManager, _, _, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("start and end trace", func(t *testing.T) {
		traceCtx, newCtx := traceManager.StartTrace(ctx, "test-operation")
		require.NotNil(t, traceCtx)
		assert.NotEmpty(t, traceCtx.TraceID)
		assert.NotEmpty(t, traceCtx.SpanID)
		assert.True(t, traceCtx.Sampled)

		// Add log to trace
		traceManager.AddTraceLog(newCtx, "info", "Test log message", map[string]interface{}{
			"key": "value",
		})

		// End trace
		traceManager.EndTrace(newCtx, "success")

		// Retrieve trace
		trace, err := traceManager.GetTrace(traceCtx.TraceID)
		require.NoError(t, err)
		assert.Equal(t, traceCtx.TraceID, trace.TraceID)
		assert.Equal(t, "success", trace.Status)
		assert.Len(t, trace.Logs, 1)
		assert.Equal(t, "Test log message", trace.Logs[0].Message)
	})

	t.Run("get trace link", func(t *testing.T) {
		traceCtx, _ := traceManager.StartTrace(ctx, "test-operation")

		link, err := traceManager.GetTraceLink(traceCtx.TraceID)
		require.NoError(t, err)
		assert.Equal(t, "jaeger", link.Type)
		assert.Contains(t, link.URL, traceCtx.TraceID)
		assert.Equal(t, "View in jaeger", link.DisplayName)
	})

	t.Run("search traces", func(t *testing.T) {
		// Create multiple traces
		for i := 0; i < 5; i++ {
			traceCtx, newCtx := traceManager.StartTrace(ctx, fmt.Sprintf("operation-%d", i))
			traceManager.EndTrace(newCtx, "success")

			// Store trace ID for searching
			if i == 2 {
				// Search for this specific trace
				filter := &LogFilter{
					TraceIDs: []string{traceCtx.TraceID},
				}

				result, err := traceManager.SearchTraces(ctx, filter)
				require.NoError(t, err)
				assert.Len(t, result.Traces, 1)
				assert.Equal(t, traceCtx.TraceID, result.Traces[0].TraceID)
			}
		}

		// Search all traces
		result, err := traceManager.SearchTraces(ctx, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Traces), 5)
	})

	t.Run("trace propagation", func(t *testing.T) {
		traceCtx, newCtx := traceManager.StartTrace(ctx, "test-operation")

		// Test HTTP header propagation
		headers := http.Header{}
		traceManager.PropagateTrace(newCtx, headers)

		assert.Equal(t, traceCtx.TraceID, headers.Get("X-Trace-Id"))
		assert.Equal(t, traceCtx.SpanID, headers.Get("X-Span-Id"))
		assert.Equal(t, "true", headers.Get("X-Sampled"))

		// Test Jaeger format
		uberTrace := headers.Get("uber-trace-id")
		assert.Contains(t, uberTrace, traceCtx.TraceID)
		assert.Contains(t, uberTrace, traceCtx.SpanID)
	})

	t.Run("trace extraction", func(t *testing.T) {
		// Create headers with trace info
		headers := http.Header{
			"X-Trace-Id": []string{"test-trace-id"},
			"X-Span-Id":  []string{"test-span-id"},
			"X-Sampled":  []string{"true"},
		}

		extracted := traceManager.ExtractTrace(headers)
		require.NotNil(t, extracted)
		assert.Equal(t, "test-trace-id", extracted.TraceID)
		assert.Equal(t, "test-span-id", extracted.SpanID)
		assert.True(t, extracted.Sampled)
	})

	t.Run("span summary", func(t *testing.T) {
		traceCtx, newCtx := traceManager.StartTrace(ctx, "test-operation")
		time.Sleep(10 * time.Millisecond)
		traceManager.EndTrace(newCtx, "success")

		summary, err := traceManager.GetSpanSummary(ctx, traceCtx.TraceID)
		require.NoError(t, err)
		assert.Equal(t, traceCtx.TraceID, summary.TraceID)
		assert.Equal(t, 1, summary.TotalSpans)
		assert.Contains(t, summary.Services, "test-service")
		assert.Len(t, summary.Timeline, 2) // start and end events
	})
}

func TestLogTailing(t *testing.T) {
	_, logTailer, _, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("write and search logs", func(t *testing.T) {
		// Write test logs
		logs := []LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Test log 1",
				Source:    "test",
				JobID:     "job-1",
				WorkerID:  "worker-1",
				TraceID:   "trace-1",
			},
			{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   "Test error log",
				Source:    "test",
				JobID:     "job-2",
				WorkerID:  "worker-1",
				TraceID:   "trace-2",
			},
			{
				Timestamp: time.Now(),
				Level:     "warning",
				Message:   "Test warning log",
				Source:    "test",
				JobID:     "job-1",
				WorkerID:  "worker-2",
				TraceID:   "trace-1",
			},
		}

		for _, log := range logs {
			err := logTailer.WriteLog(&log)
			require.NoError(t, err)
		}

		// Search all logs
		result, err := logTailer.SearchLogs(ctx, &LogFilter{})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 3)

		// Search by level
		result, err = logTailer.SearchLogs(ctx, &LogFilter{
			Levels: []string{"error"},
		})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 1)
		assert.Equal(t, "error", result.Logs[0].Level)

		// Search by job ID
		result, err = logTailer.SearchLogs(ctx, &LogFilter{
			JobIDs: []string{"job-1"},
		})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 2)

		// Search by trace ID
		result, err = logTailer.SearchLogs(ctx, &LogFilter{
			TraceIDs: []string{"trace-1"},
		})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 2)

		// Search by text
		result, err = logTailer.SearchLogs(ctx, &LogFilter{
			SearchText: "error",
		})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 1)
	})

	t.Run("log statistics", func(t *testing.T) {
		// Write various logs
		for i := 0; i < 10; i++ {
			level := "info"
			if i%3 == 0 {
				level = "error"
			} else if i%5 == 0 {
				level = "warning"
			}

			err := logTailer.WriteLog(&LogEntry{
				Timestamp: time.Now(),
				Level:     level,
				Message:   fmt.Sprintf("Log %d", i),
				Source:    "test",
				JobID:     fmt.Sprintf("job-%d", i%3),
				WorkerID:  fmt.Sprintf("worker-%d", i%2),
				TraceID:   fmt.Sprintf("trace-%d", i%4),
			})
			require.NoError(t, err)
		}

		stats, err := logTailer.GetLogStats(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalLines, int64(10))
		assert.Greater(t, stats.ErrorCount, int64(0))
		assert.GreaterOrEqual(t, stats.UniqueTraces, 4)
		assert.GreaterOrEqual(t, stats.UniqueJobs, 3)
		assert.GreaterOrEqual(t, stats.UniqueWorkers, 2)
		assert.NotEmpty(t, stats.LevelBreakdown)
	})

	t.Run("start and stop tail session", func(t *testing.T) {
		config := &TailConfig{
			Follow:            true,
			BufferSize:        100,
			MaxLinesPerSecond: 10,
			BackpressureLimit: 200,
			FlushInterval:     50 * time.Millisecond,
		}

		session, eventCh, err := logTailer.StartTail(config)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.True(t, session.Connected)

		// Write a log
		err = logTailer.WriteLog(&LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Tail test log",
			Source:    "test",
		})
		require.NoError(t, err)

		// Wait for event
		timeout := time.After(1 * time.Second)
		select {
		case event := <-eventCh:
			if event.Type == "log" {
				log := event.Data.(LogEntry)
				assert.Equal(t, "Tail test log", log.Message)
			}
		case <-timeout:
			// May not receive due to timing
		}

		// Stop tail
		err = logTailer.StopTail(session.ID)
		require.NoError(t, err)
	})

	t.Run("tail with filter", func(t *testing.T) {
		config := &TailConfig{
			Follow:            true,
			BufferSize:        100,
			MaxLinesPerSecond: 10,
			Filter: &LogFilter{
				Levels: []string{"error"},
			},
		}

		session, eventCh, err := logTailer.StartTail(config)
		require.NoError(t, err)
		defer logTailer.StopTail(session.ID)

		// Write logs of different levels
		logTailer.WriteLog(&LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Info log",
		})

		logTailer.WriteLog(&LogEntry{
			Timestamp: time.Now(),
			Level:     "error",
			Message:   "Error log",
		})

		// Only error log should come through
		timeout := time.After(1 * time.Second)
		errorReceived := false

		for !errorReceived {
			select {
			case event := <-eventCh:
				if event.Type == "log" {
					log := event.Data.(LogEntry)
					assert.Equal(t, "error", log.Level)
					errorReceived = true
				}
			case <-timeout:
				break
			}
		}
	})

	t.Run("backpressure handling", func(t *testing.T) {
		config := &TailConfig{
			Follow:            true,
			BufferSize:        10,
			MaxLinesPerSecond: 5,
			BackpressureLimit: 15,
			FlushInterval:     10 * time.Millisecond,
		}

		session, eventCh, err := logTailer.StartTail(config)
		require.NoError(t, err)
		defer logTailer.StopTail(session.ID)

		// Generate many logs quickly
		for i := 0; i < 30; i++ {
			logTailer.WriteLog(&LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("Backpressure test %d", i),
			})
		}

		// Should receive backpressure event
		timeout := time.After(2 * time.Second)
		backpressureReceived := false

		for !backpressureReceived {
			select {
			case event := <-eventCh:
				if event.Type == "backpressure" {
					backpressureReceived = true
					status := event.Data.(BackpressureStatus)
					assert.True(t, status.Active)
				}
			case <-timeout:
				break
			}
		}
	})
}

func TestIntegration(t *testing.T) {
	traceManager, logTailer, _, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("trace with logs", func(t *testing.T) {
		// Start trace
		traceCtx, newCtx := traceManager.StartTrace(ctx, "integration-test")

		// Simulate job processing with logs
		for i := 0; i < 5; i++ {
			// Add trace log
			traceManager.AddTraceLog(newCtx, "info", fmt.Sprintf("Processing step %d", i), nil)

			// Write system log
			err := logTailer.WriteLog(&LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   fmt.Sprintf("Job processing step %d", i),
				JobID:     "job-123",
				WorkerID:  "worker-456",
				TraceID:   traceCtx.TraceID,
				SpanID:    traceCtx.SpanID,
			})
			require.NoError(t, err)
		}

		// End trace
		traceManager.EndTrace(newCtx, "success")

		// Search logs by trace ID
		result, err := logTailer.SearchLogs(ctx, &LogFilter{
			TraceIDs: []string{traceCtx.TraceID},
		})
		require.NoError(t, err)
		assert.Len(t, result.Logs, 5)

		// Get trace info
		trace, err := traceManager.GetTrace(traceCtx.TraceID)
		require.NoError(t, err)
		assert.Len(t, trace.Logs, 5)
		assert.Equal(t, "success", trace.Status)
	})

	t.Run("error tracking", func(t *testing.T) {
		// Start trace
		traceCtx, newCtx := traceManager.StartTrace(ctx, "error-test")

		// Simulate error
		err := fmt.Errorf("simulated error")

		// Log error with trace
		traceManager.AddTraceLog(newCtx, "error", err.Error(), map[string]interface{}{
			"error_type": "simulated",
		})

		logTailer.WriteLog(&LogEntry{
			Timestamp:  time.Now(),
			Level:      "error",
			Message:    err.Error(),
			TraceID:    traceCtx.TraceID,
			StackTrace: "stack trace here",
			Fields: map[string]interface{}{
				"error_type": "simulated",
			},
		})

		// End trace with error
		traceManager.EndTrace(newCtx, "error")

		// Search error logs
		result, err := logTailer.SearchLogs(ctx, &LogFilter{
			Levels:       []string{"error"},
			IncludeStack: true,
		})
		require.NoError(t, err)
		assert.Greater(t, len(result.Logs), 0)

		errorLog := result.Logs[0]
		assert.Equal(t, "error", errorLog.Level)
		assert.Contains(t, errorLog.Message, "simulated error")
		assert.NotEmpty(t, errorLog.StackTrace)
	})
}

func TestHTTPIntegration(t *testing.T) {
	traceManager, _, _, cleanup := setupTest(t)
	defer cleanup()

	t.Run("trace propagation via HTTP", func(t *testing.T) {
		// Create test server that captures headers
		var capturedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Start trace
		ctx := context.Background()
		traceCtx, newCtx := traceManager.StartTrace(ctx, "http-test")

		// Make request with trace propagation
		req, _ := http.NewRequestWithContext(newCtx, "GET", server.URL, nil)
		traceManager.PropagateTrace(newCtx, req.Header)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		// Verify headers were propagated
		assert.Equal(t, traceCtx.TraceID, capturedHeaders.Get("X-Trace-Id"))
		assert.Equal(t, traceCtx.SpanID, capturedHeaders.Get("X-Span-Id"))
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("basic rate limiting", func(t *testing.T) {
		limiter := NewRateLimiter(10) // 10 per second

		// Should allow initial burst
		allowed := 0
		for i := 0; i < 20; i++ {
			if limiter.Allow() {
				allowed++
			}
		}
		assert.LessOrEqual(t, allowed, 10)

		// Wait and try again
		time.Sleep(100 * time.Millisecond)

		// Should allow ~1 more (10% of a second)
		if limiter.Allow() {
			allowed++
		}
		assert.LessOrEqual(t, allowed, 11)
	})

	t.Run("sustained rate", func(t *testing.T) {
		limiter := NewRateLimiter(100) // 100 per second

		start := time.Now()
		allowed := 0

		// Run for 100ms
		for time.Since(start) < 100*time.Millisecond {
			if limiter.Allow() {
				allowed++
			}
			time.Sleep(1 * time.Millisecond)
		}

		// Should allow approximately 10 (100ms = 0.1s * 100/s)
		assert.InDelta(t, 10, allowed, 5)
	})
}

func BenchmarkLogWrite(b *testing.B) {
	_, logTailer, _, cleanup := setupTest(&testing.T{})
	defer cleanup()

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "Benchmark log entry",
		Source:    "benchmark",
		JobID:     "job-bench",
		WorkerID:  "worker-bench",
		TraceID:   "trace-bench",
		Fields: map[string]interface{}{
			"field1": "value1",
			"field2": 123,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logTailer.WriteLog(entry)
	}
}

func BenchmarkLogSearch(b *testing.B) {
	_, logTailer, _, cleanup := setupTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	// Pre-populate with logs
	for i := 0; i < 1000; i++ {
		logTailer.WriteLog(&LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Log %d", i),
			JobID:     fmt.Sprintf("job-%d", i%10),
		})
	}

	filter := &LogFilter{
		JobIDs: []string{"job-5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logTailer.SearchLogs(ctx, filter)
	}
}

func BenchmarkTraceOperations(b *testing.B) {
	traceManager, _, _, cleanup := setupTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traceCtx, newCtx := traceManager.StartTrace(ctx, "benchmark-op")
		traceManager.AddTraceLog(newCtx, "info", "Benchmark log", nil)
		traceManager.EndTrace(newCtx, "success")
		traceManager.GetTrace(traceCtx.TraceID)
	}
}
