## In cmd/admin-api/main.go around lines 69 to 71, remove the call to logger.Fatal

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061071

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:71)

```text
In cmd/admin-api/main.go around lines 69 to 71, remove the call to logger.Fatal
which calls os.Exit and skips defers; instead log the error (e.g.
logger.Error/with context) and propagate a non-zero exit path so deferred
cleanup runs—either return the error from main and let os.Exit be called after
deferred cleanup or set an exitCode variable and call os.Exit(exitCode) only
after all defers have run; ensure Redis Close and logger.Sync defer calls remain
untouched.
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
