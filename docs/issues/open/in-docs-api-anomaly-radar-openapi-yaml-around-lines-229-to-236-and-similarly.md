## In docs/api/anomaly-radar-openapi.yaml around lines 229 to 236 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039164

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:236)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 229 to 236 (and similarly
254-266), replace any inline short/brace map usage like "{ key: value }" with
expanded standard YAML block mappings: put each key on its own line under the
parent with proper indentation and no braces or extra spaces inside braces; do
this for the securitySchemes/description and for all response blocks that
currently use inline brace maps so they conform to yamllint rules about spacing
and mapping style.
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
