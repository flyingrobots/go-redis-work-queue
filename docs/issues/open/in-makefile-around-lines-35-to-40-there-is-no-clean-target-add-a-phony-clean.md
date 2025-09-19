## In Makefile around lines 35 to 40, there is no clean target; add a PHONY clean

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575225

- [review_comment] 2025-09-16T03:18:49Z by coderabbitai[bot] (Makefile:53)

```text
In Makefile around lines 35 to 40, there is no clean target; add a PHONY clean
target that removes common build artifacts and temporary files (e.g., build/,
dist/, *.o, *.pyc, .cache, node_modules/ or other project-specific outputs) and
update the .PHONY declaration to include clean so make clean always runs;
implement the clean rule to use safe rm -rf on those paths and keep it minimal
and project-appropriate.
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
