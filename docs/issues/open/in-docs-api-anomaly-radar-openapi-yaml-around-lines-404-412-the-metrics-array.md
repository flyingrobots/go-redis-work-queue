## In docs/api/anomaly-radar-openapi.yaml around lines 404-412, the 'metrics' array

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072433

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:412)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 404-412, the 'metrics' array
(and other arrays) lack a maxItems constraint; update the schemas to add a
sensible maxItems value consistent with your API pagination/default limits
(e.g., default page size or a documented upper bound) to 'metrics' and any other
unbounded arrays in this file (notably the arrays at 450-458 and 120-137), and
propagate similar maxItems limits to any nested/ referenced array schemas (or
add global constants if you use reusable components) so OpenAPI validators no
longer flag unbounded arrays.
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
