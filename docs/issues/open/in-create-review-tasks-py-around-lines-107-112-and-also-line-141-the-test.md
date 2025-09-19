## In create_review_tasks.py around lines 107-112 (and also line 141), the test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038835

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (create_review_tasks.py:112)

```text
In create_review_tasks.py around lines 107-112 (and also line 141), the test
coverage threshold is inconsistent (90% in one place vs 80% elsewhere); pick a
single canonical threshold (e.g., 90%) and update every occurrence in this file
to match it so docs and checks agree — search for any "80%" or "90%" coverage
strings or numeric threshold variables in the file and replace them with the
chosen value, and ensure any related comments/messages reflect the same
threshold.
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
