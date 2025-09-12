package obs

import (
    "context"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// MaybeInitTracing optionally initializes a global tracer provider.
func MaybeInitTracing(cfg *config.Config) (*sdktrace.TracerProvider, error) {
    if !cfg.Observability.Tracing.Enabled || cfg.Observability.Tracing.Endpoint == "" {
        return nil, nil
    }
    exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(
        otlptracehttp.WithEndpoint(cfg.Observability.Tracing.Endpoint),
        otlptracehttp.WithInsecure(),
    ))
    if err != nil {
        return nil, err
    }
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("go-redis-work-queue"),
        )),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}

