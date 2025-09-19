## In docs/PRD.md around lines 154–156, the phrase "scan" implies using Redis KEYS

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067133

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (docs/PRD.md:156)

```text
In docs/PRD.md around lines 154–156, the phrase "scan" implies using Redis KEYS
which can block; update the text to mandate using Redis SCAN with a MATCH
pattern and a COUNT parameter and describe a cursor-based loop with safe bounds
(e.g., iteration limits) and incremental backoff between iterations to avoid
overwhelming Redis; specify using MATCH for heartbeat key pattern, set a
reasonable COUNT value, resume from the returned cursor until zero, and include
guidance to back off (sleep) when SCAN returns many keys or after a full pass to
prevent blocking and thrashing.
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
