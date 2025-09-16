// Copyright 2025 James Ross
package distributed_tracing_integration

import (
	"testing"
)

func TestTracingUIConfig_GetTraceURL(t *testing.T) {
	tests := []struct {
		name     string
		config   TracingUIConfig
		traceID  string
		expected string
	}{
		{
			name: "jaeger default viewer",
			config: TracingUIConfig{
				JaegerBaseURL:      "http://localhost:16686",
				DefaultTraceViewer: "jaeger",
			},
			traceID:  "abc123def456",
			expected: "http://localhost:16686/trace/abc123def456",
		},
		{
			name: "zipkin viewer",
			config: TracingUIConfig{
				ZipkinBaseURL:      "http://localhost:9411",
				DefaultTraceViewer: "zipkin",
			},
			traceID:  "def456abc123",
			expected: "http://localhost:9411/zipkin/traces/def456abc123",
		},
		{
			name: "custom viewer with template",
			config: TracingUIConfig{
				CustomTraceURL:     "https://my-tracing.com/trace/{traceID}",
				DefaultTraceViewer: "custom",
			},
			traceID:  "custom123",
			expected: "https://my-tracing.com/trace/custom123",
		},
		{
			name: "custom viewer with different template formats",
			config: TracingUIConfig{
				CustomTraceURL:     "https://example.com/traces/{{trace_id}}",
				DefaultTraceViewer: "custom",
			},
			traceID:  "example456",
			expected: "https://example.com/traces/example456",
		},
		{
			name: "empty trace ID",
			config: TracingUIConfig{
				JaegerBaseURL:      "http://localhost:16686",
				DefaultTraceViewer: "jaeger",
			},
			traceID:  "",
			expected: "",
		},
		{
			name: "fallback to jaeger when default viewer not available",
			config: TracingUIConfig{
				JaegerBaseURL:      "http://localhost:16686",
				DefaultTraceViewer: "nonexistent",
			},
			traceID:  "fallback123",
			expected: "http://localhost:16686/trace/fallback123",
		},
		{
			name: "no viewers configured",
			config: TracingUIConfig{
				DefaultTraceViewer: "jaeger",
			},
			traceID:  "none123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTraceURL(tt.traceID)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultTracingUIConfig(t *testing.T) {
	config := DefaultTracingUIConfig()

	if config.JaegerBaseURL != "http://localhost:16686" {
		t.Errorf("expected default Jaeger URL http://localhost:16686, got %s", config.JaegerBaseURL)
	}

	if config.ZipkinBaseURL != "http://localhost:9411" {
		t.Errorf("expected default Zipkin URL http://localhost:9411, got %s", config.ZipkinBaseURL)
	}

	if config.DefaultTraceViewer != "jaeger" {
		t.Errorf("expected default viewer jaeger, got %s", config.DefaultTraceViewer)
	}

	if !config.EnableCopyActions {
		t.Error("expected copy actions to be enabled by default")
	}

	if !config.EnableOpenActions {
		t.Error("expected open actions to be enabled by default")
	}
}

func TestReplaceTraceID(t *testing.T) {
	tests := []struct {
		name     string
		template string
		traceID  string
		expected string
	}{
		{
			name:     "basic replacement",
			template: "https://example.com/{traceID}",
			traceID:  "abc123",
			expected: "https://example.com/abc123",
		},
		{
			name:     "underscore format",
			template: "https://example.com/{trace_id}",
			traceID:  "def456",
			expected: "https://example.com/def456",
		},
		{
			name:     "double brace format",
			template: "https://example.com/{{traceID}}",
			traceID:  "ghi789",
			expected: "https://example.com/ghi789",
		},
		{
			name:     "multiple placeholders",
			template: "https://example.com/{traceID}/spans/{trace_id}",
			traceID:  "multi123",
			expected: "https://example.com/multi123/spans/multi123",
		},
		{
			name:     "no placeholders",
			template: "https://example.com/static",
			traceID:  "unused",
			expected: "https://example.com/static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceTraceID(tt.template, tt.traceID)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}