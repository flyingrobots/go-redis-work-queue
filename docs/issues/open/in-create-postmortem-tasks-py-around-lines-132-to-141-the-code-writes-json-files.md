## In create_postmortem_tasks.py around lines 132 to 141 the code writes JSON files

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072263

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:141)

```text
In create_postmortem_tasks.py around lines 132 to 141 the code writes JSON files
directly which can leave corrupted or partial files on crash; change the writes
to atomically replace the target: write to a temporary file in the same
directory (e.g., same filename + a .tmp suffix or use
tempfile.NamedTemporaryFile(dir=... , delete=False)), flush and fsync the file
descriptor to ensure data is on disk, close it, then call os.replace(temp_path,
final_path) to atomically move it into place; apply this for both the per-task
loop and the coordinator task write.
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
