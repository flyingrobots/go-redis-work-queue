## In BUGS.md around lines 65-66, the note currently lists BLMOVE as a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912425

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:126)

```text
In BUGS.md around lines 65-66, the note currently lists BLMOVE as a
"nice-to-have"; change the documentation and implementation guidance to make
BLMOVE the default for Redis ≥6.2 with a runtime fallback to BRPOPLPUSH when a
feature-probe or capability check fails. Update the text to instruct: perform a
Redis version or command-probing check at startup or before use; if BLMOVE is
available, use it by default; if the probe indicates absence or fails,
automatically fall back to BRPOPLPUSH and log a clear warning that the legacy
command is being used.
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
