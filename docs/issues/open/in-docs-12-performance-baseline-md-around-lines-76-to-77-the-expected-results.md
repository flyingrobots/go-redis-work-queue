## In docs/12_performance_baseline.md around lines 76 to 77, the expected-results

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072373

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/12_performance_baseline.md:77)

```text
In docs/12_performance_baseline.md around lines 76 to 77, the expected-results
sentence currently references bench-rate=1000 and a throughput/latency target
that contradicts the example bench command; update the text so the targets align
with the example using --bench-rate=20 (for example state the expected
throughput and p95 latency appropriate for bench-rate=20 on a 4 vCPU node),
replacing the numbers "bench-count=2000, bench-rate=1000 should achieve ≥1k
jobs/min throughput, with p95 latency < 2s" with phrasing that references
--bench-rate=20 and gives realistic throughput/latency targets for that rate.
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
