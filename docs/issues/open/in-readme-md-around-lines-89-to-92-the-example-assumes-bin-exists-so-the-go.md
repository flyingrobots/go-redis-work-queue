## In README.md around lines 89 to 92, the example assumes bin/ exists so the go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856283

- [review_comment] 2025-09-16T23:20:25Z by coderabbitai[bot] (README.md:98)

```text
In README.md around lines 89 to 92, the example assumes bin/ exists so the go
build -o bin/tui command will fail if it doesn't; update the instructions to
create the directory first (e.g., run mkdir -p bin) or use a build step that
ensures the output directory exists before running go build, then run the binary
as shown.
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
