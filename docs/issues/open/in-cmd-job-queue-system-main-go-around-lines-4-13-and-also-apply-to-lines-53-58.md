## In cmd/job-queue-system/main.go around lines 4-13 and also apply to lines 53-58,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912467

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (cmd/job-queue-system/main.go:13)

```text
In cmd/job-queue-system/main.go around lines 4-13 and also apply to lines 53-58,
the flag values are not normalized which lets values like "Admin" vs "admin"
cause bugs; add the "strings" import and after flag.Parse() trim spaces and
convert relevant flag variables to lowercase (e.g., flagVar =
strings.ToLower(strings.TrimSpace(flagVar))) for each flag that affects
behavior, ensuring all flag usages thereafter use the normalized variables.
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
