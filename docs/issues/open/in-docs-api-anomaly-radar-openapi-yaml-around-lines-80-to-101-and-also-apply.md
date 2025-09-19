## In docs/api/anomaly-radar-openapi.yaml around lines 80 to 101 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072403

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:101)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 80 to 101 (and also apply
the same change at lines 341-345), the query parameters "window" and
"max_samples" only document defaults in prose; update their parameter schemas to
include explicit default values: add default: "24h" under the window schema
(type: string) and default: 1000 under the max_samples schema (type: integer),
ensuring the OpenAPI spec reflects the defaults directly in the schema for both
occurrences.
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
