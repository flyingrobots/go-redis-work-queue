## In docs/api/anomaly-radar-openapi.yaml around lines 433-446, numeric fields lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039176

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:446)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 433-446, numeric fields lack
bounds; add validation constraints: set error_rate to minimum: 0 and maximum: 1;
set error_count to minimum: 0 (integer); set p50_latency_ms, p90_latency_ms,
p95_latency_ms, p99_latency_ms to minimum: 0 (number); and apply equivalent
min/max rules for any other percentile or threshold fields elsewhere (percentile
probabilities use [0,1], counts use >=0, latencies/thresholds use >=0 or
appropriate upper limits).
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
