## In create_review_tasks.py around lines 160 to 162, the file is opened without an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856204

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:164)

```text
In create_review_tasks.py around lines 160 to 162, the file is opened without an
explicit encoding which can cause platform-dependent issues; update the open
call to specify UTF-8 and preserve Unicode by using: open(filename, "w",
encoding="utf-8"), and pass ensure_ascii=False to json.dump (e.g.,
json.dump(task, f, indent=2, ensure_ascii=False)) so non-ASCII characters are
written correctly.
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
