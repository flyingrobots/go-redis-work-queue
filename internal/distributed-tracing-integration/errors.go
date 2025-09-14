// Copyright 2025 James Ross
package distributed_tracing_integration

import "errors"

var (
	// ErrTraceIDNotFound indicates that a trace ID was not found in the context or job
	ErrTraceIDNotFound = errors.New("trace ID not found")

	// ErrInvalidTraceID indicates that a trace ID format is invalid
	ErrInvalidTraceID = errors.New("invalid trace ID format")

	// ErrInvalidSpanID indicates that a span ID format is invalid
	ErrInvalidSpanID = errors.New("invalid span ID format")

	// ErrTracingDisabled indicates that tracing is disabled
	ErrTracingDisabled = errors.New("tracing is disabled")

	// ErrInvalidJobFormat indicates that a job has an invalid format for trace extraction
	ErrInvalidJobFormat = errors.New("invalid job format for trace extraction")

	// ErrTracingUINotConfigured indicates that tracing UI integration is not configured
	ErrTracingUINotConfigured = errors.New("tracing UI not configured")
)