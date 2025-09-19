## In docs/api/advanced-rate-limiting-api.md around lines 374 to 381, the note that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575084

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:383)

```text
In docs/api/advanced-rate-limiting-api.md around lines 374 to 381, the note that
each consume is a "single Redis round‑trip via Lua" lacks durability and Redis
Cluster details; add a "Redis Details" subsection that (1) instructs preloading
the Lua script (SCRIPT LOAD) and using EVALSHA with a safe fallback to EVAL on
NOSCRIPT, (2) documents key slotting requirements for Redis Cluster and
recommends a hash‑tag naming convention (example pattern like
{rl}:{scope}:bucket) so all keys share the same slot, and (3) specifies
operational guidance for handling transient Redis errors: timeouts, exponential
backoff and retry on NOSCRIPT, handling READONLY errors during failover/replica
writes, and suggested retry limits and logging for observability.
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
