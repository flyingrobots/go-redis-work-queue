## In docs/api/canary-deployments.md around lines 304-316 (and also adjust

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039243

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:316)

```text
In docs/api/canary-deployments.md around lines 304-316 (and also adjust
occurrences at 608-610 and 626), the metrics snapshot percent fields are
ambiguous; ensure all percent fields use a 0–100 percentage scale (not
fractions) and add a single clarifying sentence to the Metrics Snapshot section
stating "All percent fields (error_percent, success_percent, etc.) are expressed
on a 0–100 scale (e.g., 0.96 means 0.96%)." Update the example values and any
nearby percent descriptions to match that convention and verify the other
referenced lines (608-610, 626) use the same wording and numeric format.
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
