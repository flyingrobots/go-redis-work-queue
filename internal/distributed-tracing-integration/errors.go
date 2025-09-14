// Copyright 2025 James Ross
package distributedtracing

import "errors"

var (
	// ErrTracingNotEnabled indicates tracing is not enabled in configuration.
	ErrTracingNotEnabled = errors.New("tracing is not enabled")

	// ErrInvalidTraceContext indicates the trace context could not be parsed.
	ErrInvalidTraceContext = errors.New("invalid trace context")

	// ErrInvalidSamplingRate indicates the sampling rate is outside valid range.
	ErrInvalidSamplingRate = errors.New("sampling rate must be between 0.0 and 1.0")

	// ErrInvalidSamplingStrategy indicates an unknown sampling strategy.
	ErrInvalidSamplingStrategy = errors.New("invalid sampling strategy")

	// ErrInvalidPropagationFormat indicates an unknown propagation format.
	ErrInvalidPropagationFormat = errors.New("invalid propagation format")

	// ErrProviderNotInitialized indicates the tracer provider is not initialized.
	ErrProviderNotInitialized = errors.New("tracer provider not initialized")

	// ErrSpanNotFound indicates no span was found in the context.
	ErrSpanNotFound = errors.New("span not found in context")
)