## In docs/api/canary-deployments.md around lines 146 to 162, the request body for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912780

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:164)

```text
In docs/api/canary-deployments.md around lines 146 to 162, the request body for
the PUT /deployments/{id}/percentage endpoint does not define the numeric type,
valid bounds or validation/rounding behavior; update the docs to state that
"percentage" is a required number between 0 and 100 (inclusive), specify whether
integers and decimals are accepted (e.g., allow decimals to one or two decimal
places and preserve float precision), define handling for edge cases (reject
NaN, +Inf, -Inf; cap or reject >100 or <0 according to API policy — prefer
rejecting out-of-range values), and document validation/rounding rules (e.g.,
server validates and returns 400 for invalid values, or rounds to two decimal
places if automatic rounding is applied). Also add the error response shape and
status code for validation failures (e.g., 400 with JSON body containing
error_code, message, and field errors array detailing the "percentage" issue) so
callers know expected validation behavior.
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
