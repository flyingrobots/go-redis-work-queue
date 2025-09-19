## In cmd/admin-api/main.go around line 67 (and similarly for lines 98-112), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061044

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:67)

```text
In cmd/admin-api/main.go around line 67 (and similarly for lines 98-112), the
signal handler uses a hard-coded 5s timeout and never calls signal.Stop, which
can undercut cfg.ShutdownTimeout and leave the signal channel wired; change the
handler to use cfg.ShutdownTimeout (or derive timeout from ctx/cancel) when
creating the shutdown context, call signal.Notify on a channel and ensure you
call signal.Stop(ch) (preferably via defer) once the handler exits, and replace
the fixed time.After(5*time.Second) logic with a context-with-timeout using
cfg.ShutdownTimeout so the server shutdown honors configured timeout and the
notifier is cleaned up.
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
