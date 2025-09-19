## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 268 to 273, the resource metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856271

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:273)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 268 to 273, the resource metrics
(15GB RAM, 78% CPU, load avg >20, multiple 4.5-hour rate-limiting pauses, 10
parallel developers) lack provenance; update those lines to either (A) attach
how and when each metric was measured (tool/command, metric source, exact
timestamps or time ranges) and include links or references to the raw
logs/monitoring screenshots/dashboards and any aggregation queries used, or (B)
remove or convert the numbers to qualitative statements if provenance cannot be
provided; ensure each retained metric has a clear source line (e.g., "measured
via Prometheus node exporter, 2025-09-10 02:00–06:30 UTC, see Grafana dashboard
link") so reviewers can verify.
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
