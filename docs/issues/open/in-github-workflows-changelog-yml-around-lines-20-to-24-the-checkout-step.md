## In .github/workflows/changelog.yml around lines 20 to 24, the checkout step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724321

- [review_comment] 2025-09-16T21:42:53Z by coderabbitai[bot] (.github/workflows/changelog.yml:24)

```text
In .github/workflows/changelog.yml around lines 20 to 24, the checkout step
always targets the repository default_branch for tag-triggered runs which can
accidentally write to the default branch; either scope the job to only run on
tag events or make the intent explicit by using the tag ref when the event is a
tag. Fix by adding a workflow-level trigger or job-level condition to only run
on tag events (or distinguish tag vs non-tag runs), and set the checkout ref to
the actual tag ref (or to an explicit variable) when handling a tag so changelog
pushes go to the intended branch only.
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
