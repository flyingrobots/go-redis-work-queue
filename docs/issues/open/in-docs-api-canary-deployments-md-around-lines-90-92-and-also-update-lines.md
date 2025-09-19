## In docs/api/canary-deployments.md around lines 90-92 (and also update lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039224

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:92)

```text
In docs/api/canary-deployments.md around lines 90-92 (and also update lines
128-131 and 156-164), the spec lacks numeric validation for percentage updates;
add a clear constraint that percentage values must be numeric between 0 and 100
inclusive and may include decimals up to 2 decimal places, and specify that
requests with values outside this range or invalid formats must return HTTP 400;
update the JSON schema/examples and plain-language description in those sections
to state the allowed range, decimal precision, and the 400 error response for
invalid input.
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
