## In docs/api/anomaly-radar-openapi.yaml around lines 80-109 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039155

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:109)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 80-109 (and similarly
404-421), several array schemas lack maxItems which can lead to unbounded
responses; add a maxItems property to every array-type schema and array query
parameter (e.g., the alerts array in AlertsResponse and validation_errors in
ValidationErrorResponse) and ensure their limits make sense for the field
(suggest setting alerts maxItems to a sane upper bound like 1000,
validation_errors to something smaller like 100, and align any snapshot list
maxItems with the max_samples cap), updating any related descriptions to reflect
the cap.
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
