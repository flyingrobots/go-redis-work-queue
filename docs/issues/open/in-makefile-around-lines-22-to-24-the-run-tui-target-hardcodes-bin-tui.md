## In Makefile around lines 22 to 24, the run-tui target hardcodes "./bin/tui"

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067151

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (Makefile:24)

```text
In Makefile around lines 22 to 24, the run-tui target hardcodes "./bin/tui"
instead of using the project variable; change the invocation to use
$(BIN_DIR)/tui (or the existing BIN_DIR variable) so the binary directory is
referenced consistently, and ensure any other references in this target (e.g.,
path to config) follow the same variable usage if applicable.
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
