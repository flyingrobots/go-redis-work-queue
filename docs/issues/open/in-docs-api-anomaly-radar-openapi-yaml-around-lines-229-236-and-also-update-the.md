## In docs/api/anomaly-radar-openapi.yaml around lines 229-236 (and also update the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072412

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:236)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 229-236 (and also update the
corresponding block at lines 17-21), the OpenAPI components/responses section
currently omits the percentiles endpoint definition referenced in the Markdown
docs; add a complete path entry for GET /api/v1/anomaly-radar/percentiles
including its operationId, parameters, security, responses (200 with schema for
the percentiles payload, and relevant 4xx/5xx responses), and any referenced
component schemas, or alternatively add a reusable response/schema under
components and reference it from the path so the OpenAPI contract matches the
documented endpoint.
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
