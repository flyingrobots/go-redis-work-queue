// Copyright 2025 James Ross
package obs

import (
    "fmt"
    "net/http"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/prometheus/client_golang/prometheus"
    promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    JobsProduced = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_produced_total",
        Help: "Total number of jobs produced",
    })
    JobsConsumed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_consumed_total",
        Help: "Total number of jobs consumed by workers",
    })
    JobsCompleted = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_completed_total",
        Help: "Total number of successfully completed jobs",
    })
    JobsFailed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_failed_total",
        Help: "Total number of failed jobs",
    })
    JobsRetried = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_retried_total",
        Help: "Total number of job retries",
    })
    JobsDeadLetter = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jobs_dead_letter_total",
        Help: "Total number of jobs moved to dead letter queue",
    })
    JobProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "job_processing_duration_seconds",
        Help:    "Histogram of job processing durations",
        Buckets: prometheus.DefBuckets,
    })
    QueueLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "queue_length",
        Help: "Current length of Redis queues",
    }, []string{"queue"})
    CircuitBreakerState = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "circuit_breaker_state",
        Help: "0 Closed, 1 HalfOpen, 2 Open",
    })
    CircuitBreakerTrips = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "circuit_breaker_trips_total",
        Help: "Count of times the circuit breaker transitioned to Open",
    })
    ReaperRecovered = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "reaper_recovered_total",
        Help: "Total number of jobs recovered by the reaper from processing lists",
    })
    WorkerActive = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "worker_active",
        Help: "Number of active worker goroutines",
    })
)

func init() {
    prometheus.MustRegister(JobsProduced, JobsConsumed, JobsCompleted, JobsFailed, JobsRetried, JobsDeadLetter, JobProcessingDuration, QueueLength, CircuitBreakerState, CircuitBreakerTrips, ReaperRecovered, WorkerActive)
}

// StartMetricsServer exposes /metrics and returns a server for controlled shutdown.
// StartMetricsServer is retained for compatibility but consider using StartHTTPServer
// which also registers health endpoints.
func StartMetricsServer(cfg *config.Config) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Observability.MetricsPort), Handler: mux}
    go func() { _ = srv.ListenAndServe() }()
    return srv
}
