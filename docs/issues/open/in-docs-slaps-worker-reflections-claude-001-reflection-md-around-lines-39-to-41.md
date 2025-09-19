## In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 39 to 41

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033393

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:41)

```text
In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 39 to 41
the docs claim a METRICS_ENABLED toggle and per-test registry but the codebase
lacks that guard and still uses global prometheus.MustRegister calls; update
internal/obs/metrics.go (init around line ~66) and
internal/producer-backpressure/types.go (BackpressureMetrics.Register around
line ~291) to remove package-global MustRegister usage and instead accept an
injected prometheus.Registerer or gate registrations behind a configuration flag
(METRICS_ENABLED / observability.metrics.enabled). Implement one of the two
options: A) add a config-driven guard that skips global registration and allow
tests to provide prometheus.NewRegistry(), or B) refactor so no init()
registrations occur and all Register methods take a Registerer parameter (use
dependency injection), then update tests to inject per-test registries and add a
regression test to assert no duplicate-collector panics occur.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD
