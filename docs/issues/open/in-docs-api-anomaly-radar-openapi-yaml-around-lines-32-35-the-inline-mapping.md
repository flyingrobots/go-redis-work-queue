## In docs/api/anomaly-radar-openapi.yaml around lines 32-35, the inline mapping

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072387

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:35)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 32-35, the inline mapping
style like "{ $ref: '#/components/responses/Unauthorized' }" violates yamllint;
replace each inline curly-brace map with block-style YAML (use a named key
mapping, e.g. set the response code to a block mapping with $ref on its own
line) and apply the same conversion to the other reported ranges (51-53, 74-78,
116-119, 135-137, 153-156, 177-179, 200-203, 225-228) so all inline "{ $ref: ...
}" occurrences are converted to block mappings.
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
