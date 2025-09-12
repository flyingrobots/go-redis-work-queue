// Copyright 2025 James Ross
package obs

import (
    "context"
    "fmt"
    "net/http"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartHTTPServer exposes /metrics, /healthz and /readyz.
// readiness is a callback that should return nil when the app is ready.
func StartHTTPServer(cfg *config.Config, readiness func(context.Context) error) *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        // Liveness: if the process is up, return 200
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })
    mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
        if readiness == nil {
            w.WriteHeader(http.StatusOK)
            _, _ = w.Write([]byte("ready"))
            return
        }
        if err := readiness(r.Context()); err != nil {
            http.Error(w, fmt.Sprintf("not ready: %v", err), http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ready"))
    })
    srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Observability.MetricsPort), Handler: mux}
    go func() { _ = srv.ListenAndServe() }()
    return srv
}
