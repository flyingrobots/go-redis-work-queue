## In README.md around lines 83-86 (and also lines 47-49 and 53-56), the examples

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856281

- [review_comment] 2025-09-16T23:20:25Z by coderabbitai[bot] (README.md:92)

```text
In README.md around lines 83-86 (and also lines 47-49 and 53-56), the examples
run/build Go commands without ensuring modules are fetched, which causes "module
not found" errors for new users; update the instructions to run "go mod
download" before any "go run" or "make build" examples (or explicitly note that
the Makefile performs module fetching) so users fetch dependencies first and
avoid first-run failures.
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
