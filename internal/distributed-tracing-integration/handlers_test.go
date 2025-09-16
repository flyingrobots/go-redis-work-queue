// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"strings"
	"testing"
)

func TestParseJobWithTrace(t *testing.T) {
	tests := []struct {
		name        string
		jobJSON     string
		expected    *TraceableJob
		expectError bool
	}{
		{
			name: "valid job with trace",
			jobJSON: `{
				"id": "job-123",
				"filepath": "/test/file.txt",
				"filesize": 1024,
				"priority": "high",
				"retries": 1,
				"creation_time": "2023-01-01T00:00:00Z",
				"trace_id": "abc123def456",
				"span_id": "fed654cba321"
			}`,
			expected: &TraceableJob{
				ID:           "job-123",
				FilePath:     "/test/file.txt",
				FileSize:     1024,
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
			},
			expectError: false,
		},
		{
			name: "job without trace info",
			jobJSON: `{
				"id": "job-456",
				"filepath": "/test/file2.txt",
				"filesize": 2048,
				"priority": "low",
				"retries": 0,
				"creation_time": "2023-01-01T00:00:00Z",
				"trace_id": "",
				"span_id": ""
			}`,
			expected: &TraceableJob{
				ID:           "job-456",
				FilePath:     "/test/file2.txt",
				FileSize:     2048,
				Priority:     "low",
				Retries:      0,
				CreationTime: "2023-01-01T00:00:00Z",
				TraceID:      "",
				SpanID:       "",
				TraceInfo: TraceInfo{
					TraceID: "",
					SpanID:  "",
					Sampled: false,
				},
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jobJSON:     `{"invalid": json}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJobWithTrace(tt.jobJSON)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result but got nil")
				return
			}

			// Compare fields
			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %s, got %s", tt.expected.ID, result.ID)
			}
			if result.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath: expected %s, got %s", tt.expected.FilePath, result.FilePath)
			}
			if result.TraceID != tt.expected.TraceID {
				t.Errorf("TraceID: expected %s, got %s", tt.expected.TraceID, result.TraceID)
			}
			if result.TraceInfo.Sampled != tt.expected.TraceInfo.Sampled {
				t.Errorf("TraceInfo.Sampled: expected %t, got %t", tt.expected.TraceInfo.Sampled, result.TraceInfo.Sampled)
			}
		})
	}
}

func TestGenerateTraceActions(t *testing.T) {
	config := TracingUIConfig{
		JaegerBaseURL:      "http://localhost:16686",
		ZipkinBaseURL:      "http://localhost:9411",
		EnableCopyActions:  true,
		EnableOpenActions:  true,
		DefaultTraceViewer: "jaeger",
	}

	tests := []struct {
		name     string
		traceID  string
		expected int // expected number of actions
	}{
		{
			name:     "valid trace ID",
			traceID:  "abc123def456",
			expected: 3, // copy, open, view
		},
		{
			name:     "empty trace ID",
			traceID:  "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := GenerateTraceActions(tt.traceID, config)

			if len(actions) != tt.expected {
				t.Errorf("expected %d actions, got %d", tt.expected, len(actions))
				return
			}

			if tt.traceID != "" {
				// Check that we have the expected action types
				actionTypes := make(map[string]bool)
				for _, action := range actions {
					actionTypes[action.Type] = true
				}

				expectedTypes := []string{"copy", "open", "view"}
				for _, expectedType := range expectedTypes {
					if !actionTypes[expectedType] {
						t.Errorf("expected action type %s not found", expectedType)
					}
				}
			}
		})
	}
}

func TestFormatTraceForDisplay(t *testing.T) {
	config := TracingUIConfig{
		JaegerBaseURL:      "http://localhost:16686",
		DefaultTraceViewer: "jaeger",
	}

	tests := []struct {
		name     string
		job      *TraceableJob
		expected string
	}{
		{
			name: "job with trace info",
			job: &TraceableJob{
				TraceID: "abc123def456",
				SpanID:  "fed654cba321",
			},
			expected: "Trace ID: abc123def456\nSpan ID: fed654cba321\nTrace URL: http://localhost:16686/trace/abc123def456",
		},
		{
			name: "job without trace info",
			job: &TraceableJob{
				TraceID: "",
				SpanID:  "",
			},
			expected: "No trace information available",
		},
		{
			name: "job with only trace ID",
			job: &TraceableJob{
				TraceID: "onlytraceabc123",
				SpanID:  "",
			},
			expected: "Trace ID: onlytraceabc123\nTrace URL: http://localhost:16686/trace/onlytraceabc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTraceForDisplay(tt.job, config)

			// Normalize line endings for comparison
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")

			if result != expected {
				t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
			}
		})
	}
}

func TestExtractTraceInfoFromJobs(t *testing.T) {
	jobJSONs := []string{
		`{
			"id": "job-1",
			"filepath": "/test/file1.txt",
			"filesize": 1024,
			"priority": "high",
			"retries": 0,
			"creation_time": "2023-01-01T00:00:00Z",
			"trace_id": "trace1",
			"span_id": "span1"
		}`,
		`invalid json`,
		`{
			"id": "job-2",
			"filepath": "/test/file2.txt",
			"filesize": 2048,
			"priority": "low",
			"retries": 1,
			"creation_time": "2023-01-01T00:00:00Z",
			"trace_id": "trace2",
			"span_id": "span2"
		}`,
	}

	jobs, err := ExtractTraceInfoFromJobs(jobJSONs)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Should have 2 valid jobs (skipping the invalid JSON)
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
		return
	}

	if jobs[0].ID != "job-1" {
		t.Errorf("first job ID: expected job-1, got %s", jobs[0].ID)
	}

	if jobs[1].ID != "job-2" {
		t.Errorf("second job ID: expected job-2, got %s", jobs[1].ID)
	}
}