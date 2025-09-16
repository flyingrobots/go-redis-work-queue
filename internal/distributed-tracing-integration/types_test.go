// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func TestSpanAttributes_ToAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attrs    *SpanAttributes
		expected []attribute.KeyValue
	}{
		{
			name:     "empty attributes",
			attrs:    &SpanAttributes{},
			expected: []attribute.KeyValue{},
		},
		{
			name: "basic job attributes",
			attrs: &SpanAttributes{
				QueueName:      "high",
				QueueOperation: "enqueue",
				JobID:          "job-123",
				JobFilePath:    "/test/file.txt",
				JobFileSize:    1024,
				JobPriority:    "high",
				JobRetries:     2,
				WorkerID:       "worker-1",
			},
			expected: []attribute.KeyValue{
				attribute.String("queue.name", "high"),
				attribute.String("queue.operation", "enqueue"),
				attribute.String("job.id", "job-123"),
				attribute.String("job.filepath", "/test/file.txt"),
				attribute.Int64("job.filesize", 1024),
				attribute.String("job.priority", "high"),
				attribute.Int("job.retries", 2),
				attribute.String("worker.id", "worker-1"),
			},
		},
		{
			name: "with processing time",
			attrs: &SpanAttributes{
				ProcessingTime: 150 * time.Millisecond,
			},
			expected: []attribute.KeyValue{
				attribute.Int64("processing.time_ms", 150),
			},
		},
		{
			name: "with custom attributes",
			attrs: &SpanAttributes{
				JobID: "job-456",
				Custom: map[string]interface{}{
					"custom_string": "value",
					"custom_int":    42,
					"custom_int64":  int64(999),
					"custom_float":  3.14,
					"custom_bool":   true,
				},
			},
			expected: []attribute.KeyValue{
				attribute.String("job.id", "job-456"),
				attribute.String("custom_string", "value"),
				attribute.Int("custom_int", 42),
				attribute.Int64("custom_int64", 999),
				attribute.Float64("custom_float", 3.14),
				attribute.Bool("custom_bool", true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attrs.ToAttributes()

			// Create maps for comparison since order might differ
			expectedMap := make(map[string]interface{})
			resultMap := make(map[string]interface{})

			for _, kv := range tt.expected {
				expectedMap[string(kv.Key)] = kv.Value.AsInterface()
			}
			for _, kv := range result {
				resultMap[string(kv.Key)] = kv.Value.AsInterface()
			}

			if len(expectedMap) != len(resultMap) {
				t.Errorf("expected %d attributes, got %d", len(expectedMap), len(resultMap))
				return
			}

			for key, expectedValue := range expectedMap {
				if resultValue, ok := resultMap[key]; !ok {
					t.Errorf("missing attribute %s", key)
				} else if resultValue != expectedValue {
					t.Errorf("attribute %s: expected %v, got %v", key, expectedValue, resultValue)
				}
			}
		})
	}
}

func TestTraceableJob_Structure(t *testing.T) {
	job := &TraceableJob{
		ID:           "test-job",
		FilePath:     "/test/path",
		FileSize:     2048,
		Priority:     "high",
		Retries:      1,
		CreationTime: "2023-01-01T00:00:00Z",
		TraceID:      "abc123def456",
		SpanID:       "fed654cba321",
		TraceInfo: TraceInfo{
			TraceID: "abc123def456",
			SpanID:  "fed654cba321",
			Sampled: true,
		},
	}

	if job.ID != "test-job" {
		t.Errorf("expected ID test-job, got %s", job.ID)
	}

	if job.TraceInfo.TraceID != "abc123def456" {
		t.Errorf("expected TraceID abc123def456, got %s", job.TraceInfo.TraceID)
	}

	if !job.TraceInfo.Sampled {
		t.Error("expected trace to be sampled")
	}
}