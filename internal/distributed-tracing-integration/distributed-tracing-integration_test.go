//go:build tracing_tests
// +build tracing_tests

// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
)

func TestNew(t *testing.T) {
	config := TracingUIConfig{
		JaegerBaseURL:      "http://localhost:16686",
		DefaultTraceViewer: "jaeger",
	}

	integration := New(config)

	if integration == nil {
		t.Error("expected integration to be created")
		return
	}

	if integration.config.JaegerBaseURL != "http://localhost:16686" {
		t.Errorf("expected Jaeger URL http://localhost:16686, got %s", integration.config.JaegerBaseURL)
	}
}

func TestNewWithDefaults(t *testing.T) {
	integration := NewWithDefaults()

	if integration == nil {
		t.Error("expected integration to be created")
		return
	}

	if integration.config.DefaultTraceViewer != "jaeger" {
		t.Errorf("expected default viewer jaeger, got %s", integration.config.DefaultTraceViewer)
	}
}

func TestTracingIntegration_EnhancePeekWithTracing(t *testing.T) {
	integration := NewWithDefaults()

	jobJSONs := []string{
		`{
			"id": "job-123",
			"filepath": "/test/file.txt",
			"filesize": 1024,
			"priority": "high",
			"retries": 0,
			"creation_time": "2023-01-01T00:00:00Z",
			"trace_id": "abc123def456",
			"span_id": "fed654cba321"
		}`,
		`{
			"id": "job-456",
			"filepath": "/test/file2.txt",
			"filesize": 2048,
			"priority": "low",
			"retries": 1,
			"creation_time": "2023-01-01T00:00:00Z",
			"trace_id": "",
			"span_id": ""
		}`,
	}

	result, err := integration.EnhancePeekWithTracing("test-queue", jobJSONs)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result.Queue != "test-queue" {
		t.Errorf("expected queue test-queue, got %s", result.Queue)
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}

	if len(result.TraceJobs) != 2 {
		t.Errorf("expected 2 trace jobs, got %d", len(result.TraceJobs))
	}

	// Check that trace actions are generated for the job with trace info
	if len(result.TraceActions) != 1 {
		t.Errorf("expected 1 job with trace actions, got %d", len(result.TraceActions))
	}

	if _, exists := result.TraceActions["job-123"]; !exists {
		t.Error("expected trace actions for job-123")
	}
}

func TestTracingIntegration_ValidateTracingSetup(t *testing.T) {
	integration := NewWithDefaults()

	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Observability: config.Observability{
					Tracing: config.Tracing{
						Enabled:  true,
						Endpoint: "http://localhost:4317",
					},
				},
			},
			expectError: false,
		},
		{
			name: "tracing disabled",
			cfg: &config.Config{
				Observability: config.Observability{
					Tracing: config.Tracing{
						Enabled: false,
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing endpoint",
			cfg: &config.Config{
				Observability: config.Observability{
					Tracing: config.Tracing{
						Enabled:  true,
						Endpoint: "",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := integration.ValidateTracingSetup(tt.cfg)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTracingIntegration_FormatJobsForTUIDisplay(t *testing.T) {
	integration := NewWithDefaults()

	jobs := []*TraceableJob{
		{
			ID:       "job-123",
			FilePath: "/test/file1.txt",
			FileSize: 1024,
			Priority: "high",
			Retries:  0,
			TraceID:  "abc123def456789",
		},
		{
			ID:       "job-456",
			FilePath: "/test/file2.txt",
			FileSize: 2048,
			Priority: "low",
			Retries:  1,
			TraceID:  "", // No trace ID
		},
	}

	formatted := integration.FormatJobsForTUIDisplay(jobs)

	if len(formatted) != 2 {
		t.Errorf("expected 2 formatted jobs, got %d", len(formatted))
		return
	}

	// Check first job (with trace)
	if !strings.Contains(formatted[0], "job-123") {
		t.Error("first formatted job should contain job-123")
	}
	if !strings.Contains(formatted[0], "[Trace: abc123de]") {
		t.Error("first formatted job should contain truncated trace ID")
	}

	// Check second job (no trace)
	if !strings.Contains(formatted[1], "job-456") {
		t.Error("second formatted job should contain job-456")
	}
	if strings.Contains(formatted[1], "[Trace:") {
		t.Error("second formatted job should not contain trace information")
	}
}
