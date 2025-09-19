## In BUGS.md around lines 14-21, the current code treats SetArgs as returning

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060989

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:21)

```text
In BUGS.md around lines 14-21, the current code treats SetArgs as returning
(bool, error) which is incorrect; change the logic to either use SetNX
(rdb.SetNX(ctx, hbKey, workerID, cfg.Worker.HeartbeatTTL)) which returns (bool,
error) and check the bool to detect existing heartbeat, or if you must use
SetArgs, call Result() on the StatusCmd and handle redis.Nil as the "already
exists" case, treat any non-nil error as a failure, and ensure a non-"OK" result
is also handled as an error.
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
