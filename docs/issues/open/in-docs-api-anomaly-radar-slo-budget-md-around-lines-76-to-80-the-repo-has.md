## In docs/api/anomaly-radar-slo-budget.md around lines 76 to 80, the repo has

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679598

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:119)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 76 to 80, the repo has
mixed go-redis client versions (github.com/go-redis/redis/v8 vs
github.com/redis/go-redis/v9); choose one version (preferably migrate all to v9
or standardize on v8), update all import paths listed in the comment to the
chosen module, update go.mod accordingly, run go mod tidy, run the full test
suite, and fix any API incompatibilities caused by the version change before
merging.
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
