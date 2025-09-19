## In docs/api/advanced-rate-limiting-api.md around lines 223 to 236, the Status

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912602

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:236)

```text
In docs/api/advanced-rate-limiting-api.md around lines 223 to 236, the Status
struct mixes scope-level and priority-level fields (Priority/Weight) creating
ambiguity about the contract for GetStatus(ctx, scope). Either remove Priority
and Weight from Status and introduce a separate FairnessStatus (or
PriorityStatus) type and update examples to call GetFairnessStatus/GetStatus as
appropriate, or document that Status represents a (scope, priority) tuple by
renaming the type to StatusForPriority and updating method signatures/examples
to accept/return a priority-scoped status; pick one approach and make the
corresponding API doc changes consistently (type name, field list, example
calls, and method description).
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
