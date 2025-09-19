## In docs/api/admin-api.md around lines 331 to 354, the Go example references a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974289

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:354)

```text
In docs/api/admin-api.md around lines 331 to 354, the Go example references a
non‑existent StatsResponse type which will break copy‑paste; either add a
concise StatsResponse struct definition showing the expected fields (e.g.,
counts/timestamps/metrics returned by /api/v1/stats) and mention it above the
sample, or change the example to decode into a raw map[string]interface{} and
update the function signature/return type accordingly; also ensure the sample
handles the http.NewRequest error and returns the json.Decoder.Decode error
instead of ignoring it.
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
