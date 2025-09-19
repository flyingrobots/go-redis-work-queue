## In CHANGELOG.md around line 20, the entry "Smarter rate limiting that sleeps

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792422

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (CHANGELOG.md:20)

```text
In CHANGELOG.md around line 20, the entry "Smarter rate limiting that sleeps
using TTL and jitter for fairness ([#PR?])" is marketing-y and vague; replace it
with a terse, precise description naming the algorithm and behavior such as
"Fixed-window rate limiter with per-key TTL and randomized jitter for backoff
([#PR?])" or the actual algorithm used (e.g., "Token bucket with per-key TTL and
randomized jitter"), keeping it short and factual.
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
