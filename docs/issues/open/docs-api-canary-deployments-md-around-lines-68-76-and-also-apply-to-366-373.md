## docs/api/canary-deployments.md around lines 68-76 (and also apply to 366-373,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039200

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:76)

```text
docs/api/canary-deployments.md around lines 68-76 (and also apply to 366-373,
375-380, 783-789): the example JSON uses IDs like "canary_<uuid>" while the spec
later mandates ULIDs; choose one consistent ID scheme and update both examples
and the spec. Either (A) change examples to plain ULIDs (remove the "canary_"
prefix) and update any example values/show regex to match ULID format, or (B)
update the spec to state "prefix + ULID", adjust the descriptive text and
provide a regex that matches the "canary_" prefix followed by a ULID, then
update all listed example occurrences to follow that chosen pattern so examples
and regexes are consistent.
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
