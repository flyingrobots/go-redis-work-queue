## In docs/PRD.md around lines 130-131, the rate-limiter note is inaccurate:

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575159

- [review_comment] 2025-09-16T03:18:47Z by coderabbitai[bot] (docs/PRD.md:134)

```text
In docs/PRD.md around lines 130-131, the rate-limiter note is inaccurate:
calling INCR + EX=1s implements a fixed-window counter that allows bursts at
window boundaries; either amend the text to explicitly state this fixed-window
behavior and its bursty edge-case, or change the described implementation to a
Lua-based token-bucket (which is already referenced elsewhere in this PR) and
link to that section. Update the doc to clearly state which approach is used,
describe its observable behavior (bursting vs. smooth token refill), and, if
switching to the Lua token-bucket, point readers to the existing token-bucket
snippet/section in this PR for implementation details.
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
