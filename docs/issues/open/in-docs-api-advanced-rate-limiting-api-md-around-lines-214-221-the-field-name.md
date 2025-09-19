## In docs/api/advanced-rate-limiting-api.md around lines 214–221 the field name

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912577

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:221)

```text
In docs/api/advanced-rate-limiting-api.md around lines 214–221 the field name
DryRunWouldAllow is ambiguous; update the documentation to explicitly state that
this boolean means "true if the request would have been allowed when
DryRun=false (i.e., in non-dry-run mode)". Alternatively, if you prefer a
clearer identifier, rename the field in the code and docs to AllowedIfNotDryRun
(or WouldHaveBeenAllowedIfNotDryRun), update all references and API consumers,
and ensure the doc comment matches the new name and semantics.
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
