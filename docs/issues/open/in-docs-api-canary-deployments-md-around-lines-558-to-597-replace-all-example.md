## In docs/api/canary-deployments.md around lines 558 to 597, replace all example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061288

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/canary-deployments.md:597)

```text
In docs/api/canary-deployments.md around lines 558 to 597, replace all example
deployment IDs that use legacy placeholders like "canary_..." with valid
26-character ULIDs (Crockford base32, uppercase) that match the regex
^[0-9A-HJKMNP-TV-Z]{26}$; update the specific occurrences mentioned at lines 70,
369, 376, and 783 as well as any other examples in the document, ensuring each
ID is 26 uppercase base32 characters and adjust any accompanying validation text
or examples to use those ULIDs.
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
