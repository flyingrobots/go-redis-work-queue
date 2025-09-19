## In cmd/admin-api/main.go around line 59, the deferred call defer logger.Sync()

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061036

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:59)

```text
In cmd/admin-api/main.go around line 59, the deferred call defer logger.Sync()
ignores its returned error; replace it with a deferred closure that captures and
checks the error (e.g. defer func(){ if err := logger.Sync(); err != nil {
fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err) } }()), so Sync errors
are not silently swallowed and are written to stderr for visibility.
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
