## In docs/api/canary-deployments.md around lines 110-113 (also apply fixes at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912771

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:115)

```text
In docs/api/canary-deployments.md around lines 110-113 (also apply fixes at
512-517, 529-537, 543-552), the duration values use mixed formats like "2h" and
"5m" rather than canonical Go time.Duration strings; normalize all duration
fields (e.g., max_duration, min_duration, metrics_window, and any other duration
keys) to full Go canonical form such as "2h0m0s" and "5m0s" consistently across
the file, updating the example JSON/YAML values and any explanatory text so
every duration uses the same canonical format.
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
