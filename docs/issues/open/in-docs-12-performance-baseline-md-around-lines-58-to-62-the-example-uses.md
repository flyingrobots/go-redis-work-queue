## In docs/12_performance_baseline.md around lines 58 to 62, the example uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072364

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/12_performance_baseline.md:62)

```text
In docs/12_performance_baseline.md around lines 58 to 62, the example uses
--bench-rate=1000 which conflates jobs/sec vs jobs/min and invalidates the "≥1k
jobs/min" claim; change the example to use a per‑second rate that matches the
target (e.g., --bench-rate=20 to target ≈1.2k/min), explicitly state the unit in
the flag description near the example ("--bench-rate is jobs/second"), and
update the Expected Results text to reflect the corrected rate/throughput
numbers.
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
