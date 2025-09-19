## In docs/api/advanced-rate-limiting-api.md around lines 74–84, the example shows

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912542

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:84)

```text
In docs/api/advanced-rate-limiting-api.md around lines 74–84, the example shows
sleeping once then returning an error; change it to demonstrate a capped retry
loop with backoff and proper cancellation: introduce a package-level
ErrRateLimited for callers to check, then replace the single sleep/return with a
loop that attempts rl.Consume up to a maxRetries, uses result.RetryAfter (or an
exponential backoff capped to a maxDelay) between attempts, respects ctx
cancellation/deadline, and returns ErrRateLimited if retries are exhausted (or
the context is done) so callers can handle rate-limit errors explicitly.
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
