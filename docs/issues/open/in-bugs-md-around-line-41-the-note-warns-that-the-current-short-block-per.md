## In BUGS.md around line 41, the note warns that the current "short block per

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912392

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:74)

```text
In BUGS.md around line 41, the note warns that the current "short block per
queue in priority order" can starve low-priority work; replace that vague
instruction with a concrete weighted round‑robin algorithm using per‑priority
token buckets (example weights 8:2:1 for High:Med:Low), describe the refill
logic (reset bucket to weight when zero and decrement on each claim), require
weights configurable and validated as >0, and instruct adding per‑priority
metrics (tokens, claims, starve counters) so behavior is observable.
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
