## In docs/api/capacity-planning-api.md around lines 508 to 533, add a canonical

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912923

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:533)

```text
In docs/api/capacity-planning-api.md around lines 508 to 533, add a canonical
mapping table that translates each PlannerError.Code to the appropriate HTTP
status code and gRPC status code; list each error code (INVALID_METRICS,
INSUFFICIENT_HISTORY, FORECAST_FAILED, MODEL_NOT_SUPPORTED, CONFIG_INVALID,
SLO_UNACHIEVABLE, CAPACITY_LIMIT_EXCEEDED, COOLDOWN_ACTIVE, ANOMALY_DETECTED)
with a recommended HTTP status (e.g., 400 for client errors, 404/409 where
appropriate, 429 for rate/cooldown, 500 for server/forecast failures) and
corresponding gRPC canonical codes (e.g., INVALID_ARGUMENT, NOT_FOUND,
FAILED_PRECONDITION/ALREADY_EXISTS as applicable, RESOURCE_EXHAUSTED for limits,
UNAVAILABLE/INTERNAL for engine failures), and include a one-line rationale
column for each mapping to justify the choice.
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
