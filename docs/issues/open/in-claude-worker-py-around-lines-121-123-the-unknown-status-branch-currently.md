## In claude_worker.py around lines 121-123, the unknown-status branch currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856166

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:130)

```text
In claude_worker.py around lines 121-123, the unknown-status branch currently
only logs and returns, leaving the task file orphaned in my_dir; modify this
branch to atomically move the task's file from my_dir into the help queue
directory (e.g., help_dir or my_dir/help) and write or attach contextual
metadata (task_id, status value, timestamp, and any error/trace info) so humans
can triage; handle and log any filesystem errors and ensure the function returns
a value indicating the task was requeued to help.
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
