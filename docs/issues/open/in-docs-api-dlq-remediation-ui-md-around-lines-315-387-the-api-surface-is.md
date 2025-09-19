## In docs/api/dlq-remediation-ui.md around lines 315-387 the API surface is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856260

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:387)

```text
In docs/api/dlq-remediation-ui.md around lines 315-387 the API surface is
described with ad‑hoc Markdown tables rather than a formal OpenAPI contract;
create and publish an openapi.yaml (e.g., docs/api/openapi.yaml) that defines
all schemas shown (DLQEntry, ErrorDetails, JobMetadata, AttemptRecord,
ErrorPattern, BulkOperationResult, OperationError) and endpoints, include
explicit enums for fields like ErrorPattern.severity (low, medium, high,
critical) and any role enums, and add request/response schemas; then wire
server-side validation middleware (e.g., ajv/express-openapi-validator or
framework equivalent) to enforce the contract for incoming requests and outgoing
responses, update docs to reference the openapi.yaml, and commit both the
openapi.yaml and the validation integration so the Markdown tables become a
generated/derived view from the authoritative OpenAPI file.
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
