## In docs/api/dlq-remediation-pipeline.md around lines 197 to 252, the matcher

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569775

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:266)

```text
In docs/api/dlq-remediation-pipeline.md around lines 197 to 252, the matcher
block uses free-form strings (e.g., "retry_count": "< 3", "job_type":
"business_hours") without a formal grammar or schema; add a clear BNF or JSON
Schema for matcher fields, enumerating allowed keys/types (error_pattern as
regex, job_type enum or pattern, retry_count as structured comparator object
with operator and integer, time windows as structured objects like {start, end}
or named set references), update the example to use the structured form, and
document validation/error responses (HTTP 4xx with specific field and error
messages) so callers can validate and avoid undefined behavior.
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
