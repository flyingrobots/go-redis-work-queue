## In docs/api/dlq-remediation-pipeline.md around lines 313-319 (and also apply the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856256

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:324)

```text
In docs/api/dlq-remediation-pipeline.md around lines 313-319 (and also apply the
same changes at 355-358 and 760-787), the matcher grammar sections mix the
structured JSON schema with free-form string examples; update the document to
consistently present only the structured matcher schema (error_pattern,
job_type, retry_count, optional time_window) and remove all ad-hoc/free-form
string examples in the "Update Rule" and "Patterns" sections, and add a short
note stating that free-form strings are deprecated and will be rejected by
validation so clients must use the structured schema.
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
