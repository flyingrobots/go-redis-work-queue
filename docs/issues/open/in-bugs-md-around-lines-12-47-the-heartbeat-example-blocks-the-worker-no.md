## In BUGS.md around lines 12–47: the heartbeat example blocks the worker (no

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033007

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:47)

```text
In BUGS.md around lines 12–47: the heartbeat example blocks the worker (no
goroutine), renews without verifying ownership, and relies on a non‑standard
jitter package; replace it with: perform an initial SET NX with TTL (PX) to
claim the heartbeat, start a background goroutine that does jittered time.Sleep
using the standard library (math/rand + time) and on each tick runs an atomic
Lua script which checks the key value == workerID and only then extends the TTL
(so only owner can renew), handle transient errors with retry logic and
non‑blocking warning logs, cancel the goroutine via context cancel before final
LREM/DEL so it exits cleanly, and document the Lua script and standard library
jitter approach for portability.
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
