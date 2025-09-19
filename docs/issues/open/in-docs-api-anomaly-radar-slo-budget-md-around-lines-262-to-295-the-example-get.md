## In docs/api/anomaly-radar-slo-budget.md around lines 262 to 295 the example GET

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072450

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:295)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 262 to 295 the example GET
/config response includes extra fields "summary" and "is_valid" that are not
present in the OpenAPI GetConfigResponse; align the contract by either removing
"summary" and "is_valid" from this example in the docs or add those fields to
the OpenAPI GetConfigResponse schema and update the implementation to return
them (update schema, regenerate clients if any, and ensure server handler sets
these fields) so the documentation, schema, and code are consistent.
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
