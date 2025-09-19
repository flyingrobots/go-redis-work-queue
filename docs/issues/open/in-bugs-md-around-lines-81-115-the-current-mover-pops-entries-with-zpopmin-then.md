## In BUGS.md around lines 81-115, the current mover pops entries with ZPOPMIN then

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061011

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:115)

```text
In BUGS.md around lines 81-115, the current mover pops entries with ZPOPMIN then
uses a pipeline to re-add future items which can lose jobs if the pipeline
fails; replace the whole pop+pipe loop with a single server-side Lua script
executed via rdb.Eval that atomically moves due members (score <= now) from the
ZSET to the LIST up to a limit, passing schedKey and queueKey as KEYS and now
and limit as ARGV, then check the returned moved count and error and remove the
old loop and pipeline logic.
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
