## In docs/PRD.md around lines 134-136, the current recommendation of using

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575173

- [review_comment] 2025-09-16T03:18:48Z by coderabbitai[bot] (docs/PRD.md:139)

```text
In docs/PRD.md around lines 134-136, the current recommendation of using
BRPOPLPUSH with a 1s per-queue timeout is latency-hostile for low-priority jobs;
update the doc to describe two configurable modes: (1) a low-latency mode that
reduces per-queue timeout to a much smaller value (e.g., 50-200ms) and explains
the increased CPU/redis load tradeoff, and (2) an atomic-priority mode that uses
a Lua script to probe priority queues and atomically RPOPLPUSH a job into
processing in one call (or a batched probe) and documents its complexity and
guarantees; add config knobs (mode name and timeout) to the spec and a short
paragraph comparing tradeoffs, recommended defaults, and when to choose each
mode.
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
