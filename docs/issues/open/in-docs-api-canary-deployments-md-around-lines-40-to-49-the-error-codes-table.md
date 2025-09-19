## In docs/api/canary-deployments.md around lines 40 to 49, the error codes table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912707

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:51)

```text
In docs/api/canary-deployments.md around lines 40 to 49, the error codes table
is missing authentication/authorization entries; add two rows to the table: one
for the 401 case (e.g., code `UNAUTHENTICATED` or `UNAUTHORIZED` with
description like "Authentication required" and HTTP Status `401`) and one for
the 403 case (e.g., code `FORBIDDEN` with description like "Insufficient
permissions" and HTTP Status `403`), ensuring they follow the same table
formatting as the existing rows.
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
