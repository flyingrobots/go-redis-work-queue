// Copyright 2025 James Ross
package distributedtracing

import (
	"time"
)

// TracingConfig holds configuration for distributed tracing.
type TracingConfig struct {
	// Enabled determines if tracing is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Endpoint is the OTLP endpoint for trace export
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Environment label for the service
	Environment string `json:"environment" yaml:"environment"`

	// SamplingStrategy: "always", "never", "probabilistic", "adaptive"
	SamplingStrategy string `json:"sampling_strategy" yaml:"sampling_strategy"`

	// SamplingRate for probabilistic sampling (0.0 to 1.0)
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"`

	// BatchTimeout for batching trace exports
	BatchTimeout time.Duration `json:"batch_timeout" yaml:"batch_timeout"`

	// MaxExportBatchSize is the maximum number of spans to export in a batch
	MaxExportBatchSize int `json:"max_export_batch_size" yaml:"max_export_batch_size"`

	// Headers to include in OTLP exports (e.g., for authentication)
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Insecure disables TLS for the exporter
	Insecure bool `json:"insecure" yaml:"insecure"`

	// PropagationFormat: "w3c", "b3", "jaeger"
	PropagationFormat string `json:"propagation_format" yaml:"propagation_format"`

	// AttributeAllowlist defines which attributes to include in spans
	AttributeAllowlist []string `json:"attribute_allowlist" yaml:"attribute_allowlist"`

	// RedactSensitive removes sensitive data from spans
	RedactSensitive bool `json:"redact_sensitive" yaml:"redact_sensitive"`

	// EnableMetricExemplars attaches trace IDs to metrics
	EnableMetricExemplars bool `json:"enable_metric_exemplars" yaml:"enable_metric_exemplars"`
}

// DefaultTracingConfig returns default tracing configuration.
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:               false,
		Endpoint:              "localhost:4317",
		Environment:           "development",
		SamplingStrategy:      "probabilistic",
		SamplingRate:          0.1,
		BatchTimeout:          5 * time.Second,
		MaxExportBatchSize:    512,
		Headers:               make(map[string]string),
		Insecure:              true,
		PropagationFormat:     "w3c",
		AttributeAllowlist:    []string{},
		RedactSensitive:       true,
		EnableMetricExemplars: true,
	}
}

// Validate checks if the configuration is valid.
func (c *TracingConfig) Validate() error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return ErrInvalidSamplingRate
	}

	if c.SamplingStrategy != "always" &&
		c.SamplingStrategy != "never" &&
		c.SamplingStrategy != "probabilistic" &&
		c.SamplingStrategy != "adaptive" {
		return ErrInvalidSamplingStrategy
	}

	if c.PropagationFormat != "w3c" &&
		c.PropagationFormat != "b3" &&
		c.PropagationFormat != "jaeger" {
		return ErrInvalidPropagationFormat
	}

	return nil
}