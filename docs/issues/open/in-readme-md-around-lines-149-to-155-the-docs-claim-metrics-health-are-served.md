## In README.md around lines 149 to 155, the docs claim metrics/health are served

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569852

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:155)

```text
In README.md around lines 149 to 155, the docs claim metrics/health are served
on port 9090 which conflicts with Prometheus' default; update the README to
change the default metrics/health port to a non-conflicting port (e.g., 9091 or
2112) and clearly document the potential clash with local Prometheus (include a
note explaining how to override the port or how to avoid collision), ensuring
both the endpoint URLs and any startup/config examples reflect the new default.
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
