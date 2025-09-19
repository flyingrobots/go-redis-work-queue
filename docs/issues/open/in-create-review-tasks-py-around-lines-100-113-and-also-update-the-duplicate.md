## In create_review_tasks.py around lines 100-113 (and also update the duplicate

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061102

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (create_review_tasks.py:113)

```text
In create_review_tasks.py around lines 100-113 (and also update the duplicate
entries around lines 138-143), the coverage threshold is inconsistent between
90% in the DoD and 80% in the task instructions; standardize both to 90%. Update
the task definition entries so any mention of coverage uses "90%+" (or a numeric
90) and remove or replace the 80% references to ensure both the criteria and
instructions are aligned to the 90% threshold.
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
