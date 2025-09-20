// Copyright 2025 James Ross
package fixtures

import (
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingTestConfig provides standard test configurations for tracing tests
type TracingTestConfig struct {
	Name     string
	Config   *config.Config
	Expected ExpectedResults
}

type ExpectedResults struct {
	SpanCount         int
	TraceIDsGenerated int
	AttributesPerSpan int
	SamplingRate      float64
}

// GetTracingTestConfigs returns a set of test configurations for different scenarios
func GetTracingTestConfigs() []TracingTestConfig {
	return []TracingTestConfig{
		{
			Name: "always_sampling",
			Config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled:          true,
						Endpoint:         "http://localhost:4318/v1/traces",
						Environment:      "test",
						SamplingStrategy: "always",
						SamplingRate:     1.0,
					},
				},
			},
			Expected: ExpectedResults{
				SpanCount:         10,
				TraceIDsGenerated: 10,
				AttributesPerSpan: 5,
				SamplingRate:      1.0,
			},
		},
		{
			Name: "never_sampling",
			Config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled:          true,
						Endpoint:         "http://localhost:4318/v1/traces",
						Environment:      "test",
						SamplingStrategy: "never",
						SamplingRate:     0.0,
					},
				},
			},
			Expected: ExpectedResults{
				SpanCount:         0,
				TraceIDsGenerated: 0,
				AttributesPerSpan: 0,
				SamplingRate:      0.0,
			},
		},
		{
			Name: "probabilistic_sampling",
			Config: &config.Config{
				Observability: config.ObservabilityConfig{
					Tracing: config.TracingConfig{
						Enabled:          true,
						Endpoint:         "http://localhost:4318/v1/traces",
						Environment:      "test",
						SamplingStrategy: "probabilistic",
						SamplingRate:     0.5,
					},
				},
			},
			Expected: ExpectedResults{
				SpanCount:         5, // Approximate with 50% sampling
				TraceIDsGenerated: 5,
				AttributesPerSpan: 5,
				SamplingRate:      0.5,
			},
		},
	}
}

// TestJob provides sample job data for testing
type TestJob struct {
	Job      queue.Job
	Expected JobExpectedResults
}

type JobExpectedResults struct {
	ShouldCreateSpan   bool
	ExpectedAttributes []attribute.KeyValue
	ExpectedEvents     []string
}

// GetTestJobs returns a variety of test jobs for different scenarios
func GetTestJobs() []TestJob {
	now := time.Now().Format(time.RFC3339)

	return []TestJob{
		{
			Job: queue.Job{
				ID:           "test-job-1",
				FilePath:     "/test/data/file1.txt",
				FileSize:     1024,
				Priority:     "high",
				Retries:      0,
				CreationTime: now,
				TraceID:      "4bf92f3577b34da6a3ce929d0e0e4736",
				SpanID:       "00f067aa0ba902b7",
			},
			Expected: JobExpectedResults{
				ShouldCreateSpan: true,
				ExpectedAttributes: []attribute.KeyValue{
					attribute.String("job.id", "test-job-1"),
					attribute.String("job.filepath", "/test/data/file1.txt"),
					attribute.Int64("job.filesize", 1024),
					attribute.String("job.priority", "high"),
					attribute.Int("job.retries", 0),
				},
				ExpectedEvents: []string{"job.processing.started", "job.processing.completed"},
			},
		},
		{
			Job: queue.Job{
				ID:           "test-job-2",
				FilePath:     "/test/data/large-file.bin",
				FileSize:     1073741824, // 1GB
				Priority:     "normal",
				Retries:      2,
				CreationTime: now,
			},
			Expected: JobExpectedResults{
				ShouldCreateSpan: true,
				ExpectedAttributes: []attribute.KeyValue{
					attribute.String("job.id", "test-job-2"),
					attribute.String("job.filepath", "/test/data/large-file.bin"),
					attribute.Int64("job.filesize", 1073741824),
					attribute.String("job.priority", "normal"),
					attribute.Int("job.retries", 2),
				},
				ExpectedEvents: []string{"job.processing.started"},
			},
		},
		{
			Job: queue.Job{
				ID:           "test-job-error",
				FilePath:     "/invalid/path/file.txt",
				FileSize:     512,
				Priority:     "low",
				Retries:      3,
				CreationTime: now,
			},
			Expected: JobExpectedResults{
				ShouldCreateSpan: true,
				ExpectedAttributes: []attribute.KeyValue{
					attribute.String("job.id", "test-job-error"),
					attribute.String("job.filepath", "/invalid/path/file.txt"),
					attribute.Int64("job.filesize", 512),
					attribute.String("job.priority", "low"),
					attribute.Int("job.retries", 3),
				},
				ExpectedEvents: []string{"job.processing.started", "job.processing.failed"},
			},
		},
	}
}

// TraceContext provides sample trace contexts for testing
type TraceContext struct {
	TraceID    string
	SpanID     string
	TraceFlags byte
	Baggage    map[string]string
}

// GetTraceContexts returns various trace contexts for testing propagation
func GetTraceContexts() []TraceContext {
	return []TraceContext{
		{
			TraceID:    "4bf92f3577b34da6a3ce929d0e0e4736",
			SpanID:     "00f067aa0ba902b7",
			TraceFlags: 1, // Sampled
			Baggage:    map[string]string{"tenant": "test", "user": "alice"},
		},
		{
			TraceID:    "12345678901234567890123456789012",
			SpanID:     "1234567890123456",
			TraceFlags: 0, // Not sampled
			Baggage:    map[string]string{},
		},
		{
			TraceID:    "abcdefabcdefabcdefabcdefabcdefab",
			SpanID:     "abcdefabcdefabcd",
			TraceFlags: 1, // Sampled
			Baggage:    map[string]string{"environment": "staging", "version": "1.2.3"},
		},
	}
}

// MockSpanData represents expected span data for verification
type MockSpanData struct {
	Name       string
	Attributes map[string]interface{}
	Events     []MockEvent
	Status     MockSpanStatus
	Kind       trace.SpanKind
	Duration   time.Duration
}

type MockEvent struct {
	Name       string
	Attributes map[string]interface{}
	Timestamp  time.Time
}

type MockSpanStatus struct {
	Code    int
	Message string
}

// GetExpectedSpans returns expected span data for various operations
func GetExpectedSpans() []MockSpanData {
	return []MockSpanData{
		{
			Name: "queue.enqueue",
			Attributes: map[string]interface{}{
				"queue.name":      "test-queue",
				"queue.priority":  "high",
				"queue.operation": "enqueue",
			},
			Events: []MockEvent{
				{
					Name:       "job.enqueued",
					Attributes: map[string]interface{}{"job.id": "test-job-1"},
				},
			},
			Status: MockSpanStatus{Code: 1, Message: "success"}, // OK
			Kind:   trace.SpanKindProducer,
		},
		{
			Name: "queue.dequeue",
			Attributes: map[string]interface{}{
				"queue.name":      "test-queue",
				"queue.operation": "dequeue",
			},
			Events: []MockEvent{
				{
					Name:       "job.dequeued",
					Attributes: map[string]interface{}{"job.id": "test-job-1"},
				},
			},
			Status: MockSpanStatus{Code: 1, Message: "success"}, // OK
			Kind:   trace.SpanKindConsumer,
		},
		{
			Name: "job.process",
			Attributes: map[string]interface{}{
				"job.id":            "test-job-1",
				"job.filepath":      "/test/data/file1.txt",
				"job.filesize":      int64(1024),
				"job.priority":      "high",
				"job.retries":       0,
				"job.creation_time": "2025-01-14T10:00:00Z",
				"queue.type":        "worker",
			},
			Events: []MockEvent{
				{Name: "job.processing.started"},
				{Name: "job.processing.completed"},
			},
			Status: MockSpanStatus{Code: 1, Message: "success"}, // OK
			Kind:   trace.SpanKindInternal,
		},
	}
}

// TestScenario represents a complete test scenario with multiple operations
type TestScenario struct {
	Name        string
	Description string
	Jobs        []TestJob
	Operations  []string // "enqueue", "dequeue", "process"
	Config      *config.Config
	Expected    ScenarioExpected
}

type ScenarioExpected struct {
	TotalSpans      int
	UniqueTraces    int
	LinkedSpans     int // Spans that should be linked by trace ID
	ErrorSpans      int
	SuccessfulSpans int
}

// GetTestScenarios returns comprehensive test scenarios
func GetTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:        "complete_job_lifecycle",
			Description: "Tests complete job lifecycle from enqueue through processing",
			Jobs:        GetTestJobs()[:1], // Use first job
			Operations:  []string{"enqueue", "dequeue", "process"},
			Config:      GetTracingTestConfigs()[0].Config, // Always sampling
			Expected: ScenarioExpected{
				TotalSpans:      3, // enqueue + dequeue + process
				UniqueTraces:    1, // Should be linked
				LinkedSpans:     3,
				ErrorSpans:      0,
				SuccessfulSpans: 3,
			},
		},
		{
			Name:        "batch_processing",
			Description: "Tests multiple jobs processed in parallel",
			Jobs:        GetTestJobs()[:2], // Use first two jobs
			Operations:  []string{"enqueue", "process"},
			Config:      GetTracingTestConfigs()[0].Config,
			Expected: ScenarioExpected{
				TotalSpans:      4, // 2 enqueue + 2 process
				UniqueTraces:    2,
				LinkedSpans:     4,
				ErrorSpans:      0,
				SuccessfulSpans: 4,
			},
		},
		{
			Name:        "error_handling",
			Description: "Tests tracing behavior with job processing errors",
			Jobs:        []TestJob{GetTestJobs()[2]}, // Error job
			Operations:  []string{"enqueue", "process"},
			Config:      GetTracingTestConfigs()[0].Config,
			Expected: ScenarioExpected{
				TotalSpans:      2, // enqueue + process
				UniqueTraces:    1,
				LinkedSpans:     2,
				ErrorSpans:      1, // process span should have error
				SuccessfulSpans: 1, // enqueue should succeed
			},
		},
		{
			Name:        "sampling_effects",
			Description: "Tests how different sampling rates affect span collection",
			Jobs:        GetTestJobs()[:1],
			Operations:  []string{"enqueue", "process"},
			Config:      GetTracingTestConfigs()[2].Config, // Probabilistic sampling
			Expected: ScenarioExpected{
				TotalSpans:      1, // Approximate due to sampling
				UniqueTraces:    1,
				LinkedSpans:     1,
				ErrorSpans:      0,
				SuccessfulSpans: 1,
			},
		},
	}
}

// RedisTestConfig provides Redis connection settings for tests
type RedisTestConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

// GetRedisTestConfig returns Redis configuration for testing
func GetRedisTestConfig(testName string) RedisTestConfig {
	return RedisTestConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Prefix:   "test:" + testName + ":",
	}
}

// CleanupHelper provides utilities for test cleanup
type CleanupHelper struct {
	TestName string
	Prefix   string
}

// NewCleanupHelper creates a new cleanup helper
func NewCleanupHelper(testName string) *CleanupHelper {
	return &CleanupHelper{
		TestName: testName,
		Prefix:   "test:" + testName + ":",
	}
}

// GetKeysToCleanup returns Redis keys that should be cleaned up after test
func (ch *CleanupHelper) GetKeysToCleanup() []string {
	return []string{
		ch.Prefix + "queue",
		ch.Prefix + "completed",
		ch.Prefix + "dlq",
		ch.Prefix + "processing",
		ch.Prefix + "heartbeat",
	}
}

// MockOTLPServer provides configuration for mock OTLP collectors
type MockOTLPServer struct {
	Port     int
	Endpoint string
	TLS      bool
}

// GetMockOTLPConfigs returns different OTLP server configurations for testing
func GetMockOTLPConfigs() []MockOTLPServer {
	return []MockOTLPServer{
		{
			Port:     14318,
			Endpoint: "http://localhost:14318/v1/traces",
			TLS:      false,
		},
		{
			Port:     14319,
			Endpoint: "https://localhost:14319/v1/traces",
			TLS:      true,
		},
	}
}

// PerformanceTestData provides data for performance testing
type PerformanceTestData struct {
	JobCount          int
	ConcurrentWorkers int
	SpanCount         int
	ExpectedLatency   time.Duration
}

// GetPerformanceTestData returns performance test scenarios
func GetPerformanceTestData() []PerformanceTestData {
	return []PerformanceTestData{
		{
			JobCount:          100,
			ConcurrentWorkers: 1,
			SpanCount:         300, // enqueue + dequeue + process per job
			ExpectedLatency:   100 * time.Millisecond,
		},
		{
			JobCount:          1000,
			ConcurrentWorkers: 5,
			SpanCount:         3000,
			ExpectedLatency:   500 * time.Millisecond,
		},
		{
			JobCount:          10000,
			ConcurrentWorkers: 10,
			SpanCount:         30000,
			ExpectedLatency:   2 * time.Second,
		},
	}
}
