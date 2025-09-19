## In cmd/tui/main.go around lines 109 to 112, the Redis Ping is using the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033107

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:112)

```text
In cmd/tui/main.go around lines 109 to 112, the Redis Ping is using the
background context and can hang on dead networks; wrap the ping call in a short
cancellable context (e.g., context.WithTimeout(ctx, 2*time.Second)), defer
cancel(), then call rdb.Ping(timeoutCtx).Result() and handle the error as before
(including exiting on failure); ensure you import time if not already imported.
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
