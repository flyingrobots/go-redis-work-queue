## In BUGS.md around lines 55 to 69 (and also line 71) the worker registry

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033018

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:69)

```text
In BUGS.md around lines 55 to 69 (and also line 71) the worker registry
currently uses a plain SADD which never expires and lets dead worker IDs
accumulate; change the design so membership is self‑healing by either (a)
switching to a ZSET storing lastSeen timestamps and updating the member score on
each heartbeat so the reaper can ZREMRANGEBYSCORE (or ZRANGEBYSCORE to find
stale IDs) and remove entries older than a timeout, or (b) creating a per‑worker
key with a short TTL that the worker refreshes on heartbeat and having the
reaper only consider workers with an existing key (and remove any orphaned SADD
entries if you keep the set). Update reaper logic to use ZRANGEBYSCORE or check
TTLs instead of scanning the whole set so zombies are pruned automatically.
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
