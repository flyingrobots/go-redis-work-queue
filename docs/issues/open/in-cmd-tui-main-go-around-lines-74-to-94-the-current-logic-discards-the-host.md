## In cmd/tui/main.go around lines 74 to 94, the current logic discards the host

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033095

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:94)

```text
In cmd/tui/main.go around lines 74 to 94, the current logic discards the host
portion of metrics-addr and only stores the port, which breaks host-bound
addresses like 127.0.0.1:9090; instead preserve the full host:port when
metricsAddr contains a host, and only fall back to parsing a bare port if
metricsAddr has no colon. Set cfg.Observability.MetricsAddress (or the existing
config field) to the full metricsAddr when given, and only parse and set
MetricsPort if you need the numeric port separately; also update the
observability server startup to prefer using the full host:port value from the
config if present.
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
