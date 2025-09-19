## In docs/api/canary-deployments.md around lines 90-92 (and also adjust at lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039224

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:92)

```text
In docs/api/canary-deployments.md around lines 90-92 (and also adjust at lines
118 and 744-756), the response envelope is inconsistent: it returns
"deployments":[...], "count": n while the rest of the docs use data/pagination.
Update the examples to a consistent envelope that uses "data" with a
"pagination" object (e.g. wrap lists under data.<resource> and replace top-level
"count" with data.pagination { total, limit, offset/page }), and apply the same
change to Events and Workers list response examples so all list endpoints use
the same data/pagination structure.
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
