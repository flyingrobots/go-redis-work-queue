## In docs/api/canary-deployments.md around lines 100 to 116 (and also apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912759

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:118)

```text
In docs/api/canary-deployments.md around lines 100 to 116 (and also apply same
changes to 118-130, 507-553, 573-588), the request example uses inconsistent
field names and duration formats that drift from the canonical Deployment.config
schema; rename max_duration/min_duration to
max_canary_duration/min_canary_duration (or vice-versa to match
Deployment.config exactly), normalize all duration values to the canonical
format (e.g., "5m0s" rather than "5m"), and update the Parameters sections
mentioned to use the same field names and canonical duration format so clients
see a single consistent public schema across the document.
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
