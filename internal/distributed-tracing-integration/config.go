// Copyright 2025 James Ross
package distributed_tracing_integration

import "strings"

// TracingUIConfig represents configuration for tracing UI integration
type TracingUIConfig struct {
	// JaegerBaseURL is the base URL for Jaeger UI (e.g., "http://localhost:16686")
	JaegerBaseURL string `mapstructure:"jaeger_base_url"`

	// ZipkinBaseURL is the base URL for Zipkin UI (e.g., "http://localhost:9411")
	ZipkinBaseURL string `mapstructure:"zipkin_base_url"`

	// CustomTraceURL is a custom URL template for trace viewing
	// Use {traceID} as placeholder, e.g., "https://my-tracing.com/trace/{traceID}"
	CustomTraceURL string `mapstructure:"custom_trace_url"`

	// EnableCopyActions enables copy trace ID to clipboard actions
	EnableCopyActions bool `mapstructure:"enable_copy_actions"`

	// EnableOpenActions enables open trace in browser actions
	EnableOpenActions bool `mapstructure:"enable_open_actions"`

	// DefaultTraceViewer sets the default trace viewer ("jaeger", "zipkin", "custom")
	DefaultTraceViewer string `mapstructure:"default_trace_viewer"`
}

// DefaultTracingUIConfig returns default configuration for tracing UI
func DefaultTracingUIConfig() TracingUIConfig {
	return TracingUIConfig{
		JaegerBaseURL:      "http://localhost:16686",
		ZipkinBaseURL:      "http://localhost:9411",
		CustomTraceURL:     "",
		EnableCopyActions:  true,
		EnableOpenActions:  true,
		DefaultTraceViewer: "jaeger",
	}
}

// GetTraceURL returns the trace URL for a given trace ID
func (c *TracingUIConfig) GetTraceURL(traceID string) string {
	if traceID == "" {
		return ""
	}

	switch c.DefaultTraceViewer {
	case "jaeger":
		if c.JaegerBaseURL != "" {
			return c.JaegerBaseURL + "/trace/" + traceID
		}
	case "zipkin":
		if c.ZipkinBaseURL != "" {
			return c.ZipkinBaseURL + "/zipkin/traces/" + traceID
		}
	case "custom":
		if c.CustomTraceURL != "" {
			return replaceTraceID(c.CustomTraceURL, traceID)
		}
	}

	// Fallback to Jaeger
	if c.JaegerBaseURL != "" {
		return c.JaegerBaseURL + "/trace/" + traceID
	}

	return ""
}

// replaceTraceID replaces {traceID} placeholder in URL template
func replaceTraceID(template, traceID string) string {
	result := strings.ReplaceAll(template, "{traceID}", traceID)
	result = strings.ReplaceAll(result, "{trace_id}", traceID)
	result = strings.ReplaceAll(result, "{{traceID}}", traceID)
	result = strings.ReplaceAll(result, "{{trace_id}}", traceID)
	return result
}