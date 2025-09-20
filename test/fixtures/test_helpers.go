// Copyright 2025 James Ross
package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TestTracerProvider creates a tracer provider for testing
func TestTracerProvider() *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp
}

// MockOTLPCollector simulates an OTLP trace collector
type MockOTLPCollector struct {
	server      *httptest.Server
	spans       []map[string]interface{}
	requests    [][]byte
	mutex       sync.RWMutex
	t           *testing.T
	CallCount   int
	LastReqTime time.Time
}

// NewMockOTLPCollector creates a new mock OTLP collector
func NewMockOTLPCollector(t *testing.T) *MockOTLPCollector {
	collector := &MockOTLPCollector{
		spans:    make([]map[string]interface{}, 0),
		requests: make([][]byte, 0),
		t:        t,
	}

	collector.server = httptest.NewServer(http.HandlerFunc(collector.handleRequest))
	return collector
}

func (m *MockOTLPCollector) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CallCount++
	m.LastReqTime = time.Now()

	if r.URL.Path == "/v1/traces" && r.Method == "POST" {
		// Read and store raw request
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		m.requests = append(m.requests, body)

		// Try to parse as JSON
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			m.spans = append(m.spans, payload)
			if m.t != nil {
				m.t.Logf("OTLP Collector received span batch: %d total batches", len(m.spans))
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}

// URL returns the collector's endpoint URL
func (m *MockOTLPCollector) URL() string {
	return m.server.URL
}

// Close shuts down the mock collector
func (m *MockOTLPCollector) Close() {
	m.server.Close()
}

// GetSpans returns collected spans
func (m *MockOTLPCollector) GetSpans() []map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	spans := make([]map[string]interface{}, len(m.spans))
	copy(spans, m.spans)
	return spans
}

// GetRawRequests returns raw HTTP request bodies
func (m *MockOTLPCollector) GetRawRequests() [][]byte {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	requests := make([][]byte, len(m.requests))
	for i, req := range m.requests {
		requests[i] = make([]byte, len(req))
		copy(requests[i], req)
	}
	return requests
}

// WaitForSpans waits for at least expectedCount span batches or timeout
func (m *MockOTLPCollector) WaitForSpans(expectedCount int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m.mutex.RLock()
		count := len(m.spans)
		m.mutex.RUnlock()

		if count >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TraceValidator helps validate trace data
type TraceValidator struct {
	t             *testing.T
	expectedSpans map[string]MockSpanData
}

// NewTraceValidator creates a new trace validator
func NewTraceValidator(t *testing.T) *TraceValidator {
	return &TraceValidator{
		t:             t,
		expectedSpans: make(map[string]MockSpanData),
	}
}

// AddExpectedSpan adds a span expectation
func (tv *TraceValidator) AddExpectedSpan(name string, data MockSpanData) {
	tv.expectedSpans[name] = data
}

// ValidateSpans validates collected spans against expectations
func (tv *TraceValidator) ValidateSpans(spans []map[string]interface{}) {
	if len(spans) == 0 {
		tv.t.Error("No spans collected")
		return
	}

	totalSpanCount := 0
	spansByName := make(map[string]int)

	// Parse collected spans
	for _, spanBatch := range spans {
		tv.parseSpanBatch(spanBatch, &totalSpanCount, spansByName)
	}

	tv.t.Logf("Collected %d total spans", totalSpanCount)

	// Validate against expectations
	for expectedName := range tv.expectedSpans {
		count := spansByName[expectedName]
		if count == 0 {
			tv.t.Errorf("Expected span '%s' not found", expectedName)
		} else {
			tv.t.Logf("Found %d spans of type '%s'", count, expectedName)
		}
	}
}

func (tv *TraceValidator) parseSpanBatch(spanBatch map[string]interface{}, totalCount *int, spansByName map[string]int) {
	if resourceSpans, ok := spanBatch["resourceSpans"].([]interface{}); ok {
		for _, rs := range resourceSpans {
			if rsMap, ok := rs.(map[string]interface{}); ok {
				if instrumentationLibrarySpans, ok := rsMap["instrumentationLibrarySpans"].([]interface{}); ok {
					for _, ils := range instrumentationLibrarySpans {
						if ilsMap, ok := ils.(map[string]interface{}); ok {
							if spans, ok := ilsMap["spans"].([]interface{}); ok {
								*totalCount += len(spans)
								for _, span := range spans {
									if spanMap, ok := span.(map[string]interface{}); ok {
										if name, ok := spanMap["name"].(string); ok {
											spansByName[name]++
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// RedisTestHelper provides utilities for Redis testing
type RedisTestHelper struct {
	Client *redis.Client
	Prefix string
	t      *testing.T
}

// NewRedisTestHelper creates a Redis test helper
func NewRedisTestHelper(t *testing.T, testName string) *RedisTestHelper {
	prefix := fmt.Sprintf("test:%s:", testName)

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not available for test: %v", err)
	}

	return &RedisTestHelper{
		Client: client,
		Prefix: prefix,
		t:      t,
	}
}

// Cleanup removes test keys from Redis
func (rh *RedisTestHelper) Cleanup() {
	pattern := rh.Prefix + "*"
	keys, err := rh.Client.Keys(context.Background(), pattern).Result()
	if err != nil {
		rh.t.Logf("Warning: failed to get keys for cleanup: %v", err)
		return
	}

	if len(keys) > 0 {
		if err := rh.Client.Del(context.Background(), keys...).Err(); err != nil {
			rh.t.Logf("Warning: failed to clean up keys: %v", err)
		} else {
			rh.t.Logf("Cleaned up %d Redis keys", len(keys))
		}
	}
}

// Close closes the Redis connection
func (rh *RedisTestHelper) Close() {
	rh.Cleanup()
	rh.Client.Close()
}

// GetTestConfig returns a Redis work queue config for testing
func (rh *RedisTestHelper) GetTestConfig() *config.Config {
	return &config.Config{
		Redis: config.Redis{
			Addr: "localhost:6379",
			DB:   0,
		},
		Producer: config.Producer{
			RateLimitKey: rh.Prefix + "rate-limit",
		},
		Worker: config.Worker{
			Queues: map[string]string{
				"high": rh.Prefix + "queue",
				"low":  rh.Prefix + "queue",
			},
			CompletedList:         rh.Prefix + "completed",
			DeadLetterList:        rh.Prefix + "dlq",
			ProcessingListPattern: rh.Prefix + "worker:%s:processing",
			HeartbeatKeyPattern:   rh.Prefix + "heartbeat:%s",
			HeartbeatTTL:          5 * time.Second,
			BRPopLPushTimeout:     time.Second,
			MaxRetries:            2,
			Backoff: config.Backoff{
				Base: 50 * time.Millisecond,
				Max:  2 * time.Second,
			},
		},
	}
}

// TracingTestSuite provides a complete test suite setup
type TracingTestSuite struct {
	T              *testing.T
	TracerProvider *sdktrace.TracerProvider
	OTLPCollector  *MockOTLPCollector
	RedisHelper    *RedisTestHelper
	Config         *config.Config
	TraceValidator *TraceValidator
}

// NewTracingTestSuite creates a complete test suite
func NewTracingTestSuite(t *testing.T, testName string) *TracingTestSuite {
	suite := &TracingTestSuite{
		T:              t,
		TracerProvider: TestTracerProvider(),
		OTLPCollector:  NewMockOTLPCollector(t),
		RedisHelper:    NewRedisTestHelper(t, testName),
		TraceValidator: NewTraceValidator(t),
	}

	// Configure tracing with mock collector
	baseConfig := suite.RedisHelper.GetTestConfig()
	baseConfig.Observability = config.ObservabilityConfig{
		Tracing: config.TracingConfig{
			Enabled:          true,
			Endpoint:         suite.OTLPCollector.URL() + "/v1/traces",
			Environment:      "test",
			SamplingStrategy: "always",
			SamplingRate:     1.0,
		},
	}
	suite.Config = baseConfig

	return suite
}

// Cleanup cleans up all test resources
func (ts *TracingTestSuite) Cleanup() {
	if ts.TracerProvider != nil {
		ts.TracerProvider.Shutdown(context.Background())
	}
	if ts.OTLPCollector != nil {
		ts.OTLPCollector.Close()
	}
	if ts.RedisHelper != nil {
		ts.RedisHelper.Close()
	}
}

// InitializeTracing initializes tracing for the test
func (ts *TracingTestSuite) InitializeTracing() (*sdktrace.TracerProvider, error) {
	return obs.MaybeInitTracing(ts.Config)
}

// WaitForSpansAndValidate waits for spans and validates them
func (ts *TracingTestSuite) WaitForSpansAndValidate(expectedCount int, timeout time.Duration) {
	if !ts.OTLPCollector.WaitForSpans(expectedCount, timeout) {
		ts.T.Errorf("Timeout waiting for %d span batches", expectedCount)
		return
	}

	spans := ts.OTLPCollector.GetSpans()
	ts.TraceValidator.ValidateSpans(spans)
}

// SpanMatcher provides utilities for matching spans
type SpanMatcher struct {
	t *testing.T
}

// NewSpanMatcher creates a new span matcher
func NewSpanMatcher(t *testing.T) *SpanMatcher {
	return &SpanMatcher{t: t}
}

// MatchSpanByName finds spans with the given name
func (sm *SpanMatcher) MatchSpanByName(spans []map[string]interface{}, name string) []map[string]interface{} {
	matches := make([]map[string]interface{}, 0)

	for _, spanBatch := range spans {
		matches = append(matches, sm.findSpansInBatch(spanBatch, func(span map[string]interface{}) bool {
			spanName, ok := span["name"].(string)
			return ok && spanName == name
		})...)
	}

	return matches
}

func (sm *SpanMatcher) findSpansInBatch(spanBatch map[string]interface{}, matcher func(map[string]interface{}) bool) []map[string]interface{} {
	matches := make([]map[string]interface{}, 0)

	if resourceSpans, ok := spanBatch["resourceSpans"].([]interface{}); ok {
		for _, rs := range resourceSpans {
			if rsMap, ok := rs.(map[string]interface{}); ok {
				if instrumentationLibrarySpans, ok := rsMap["instrumentationLibrarySpans"].([]interface{}); ok {
					for _, ils := range instrumentationLibrarySpans {
						if ilsMap, ok := ils.(map[string]interface{}); ok {
							if spans, ok := ilsMap["spans"].([]interface{}); ok {
								for _, span := range spans {
									if spanMap, ok := span.(map[string]interface{}); ok && matcher(spanMap) {
										matches = append(matches, spanMap)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return matches
}

// CreateTestSpanContext creates a test span context with known IDs
func CreateTestSpanContext(traceID, spanID string) trace.SpanContext {
	tid, _ := trace.TraceIDFromHex(traceID)
	sid, _ := trace.SpanIDFromHex(spanID)

	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
}

// TestMetrics provides utilities for measuring test performance
type TestMetrics struct {
	StartTime      time.Time
	EndTime        time.Time
	SpansCreated   int
	SpansExported  int
	MemoryUsage    int64
	OperationCount int
}

// NewTestMetrics creates a new test metrics collector
func NewTestMetrics() *TestMetrics {
	return &TestMetrics{
		StartTime: time.Now(),
	}
}

// Finish marks the end of measurement
func (tm *TestMetrics) Finish() {
	tm.EndTime = time.Now()
}

// Duration returns the total test duration
func (tm *TestMetrics) Duration() time.Duration {
	return tm.EndTime.Sub(tm.StartTime)
}

// ThroughputPerSecond returns operations per second
func (tm *TestMetrics) ThroughputPerSecond() float64 {
	duration := tm.Duration()
	if duration == 0 {
		return 0
	}
	return float64(tm.OperationCount) / duration.Seconds()
}

// LogResults logs the test metrics
func (tm *TestMetrics) LogResults(t *testing.T, testName string) {
	t.Logf("%s Results:", testName)
	t.Logf("  Duration: %v", tm.Duration())
	t.Logf("  Spans Created: %d", tm.SpansCreated)
	t.Logf("  Spans Exported: %d", tm.SpansExported)
	t.Logf("  Operations: %d", tm.OperationCount)
	t.Logf("  Throughput: %.2f ops/sec", tm.ThroughputPerSecond())
}
