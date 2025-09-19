## In claude_worker.py around lines 124-127, the except block for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856170

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:134)

```text
In claude_worker.py around lines 124-127, the except block for
json.JSONDecodeError and OSError currently just prints and returns False leaving
the task file in place; instead, create (if missing) a failed-tasks directory
next to my_dir and atomically move or write a failure payload there that records
the original task file name, the error message/stack, timestamp, and
(optionally) the original file contents; ensure you capture the exception as
err, build a JSON payload with those fields, write it to failed-tasks using a
deterministic filename (e.g. originalname + ".failed.json" or a UUID), remove or
rename the original task file so it is no longer left in my_dir, and then return
False after the move/write completes.
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
