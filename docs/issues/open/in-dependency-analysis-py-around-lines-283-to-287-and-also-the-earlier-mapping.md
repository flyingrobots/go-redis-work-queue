## In dependency_analysis.py around lines 283 to 287 (and also the earlier mapping

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033126

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (dependency_analysis.py:287)

```text
In dependency_analysis.py around lines 283 to 287 (and also the earlier mapping
at ~line 147), there's a duplicate "storage-backends" key causing collisions;
keep the feature entry at line 285 and rename the infra mapping at line 147 to a
distinct name (e.g. storage_backends_runtime), then update all references/usages
throughout the codebase to the new name so the DAG keys are unique; verify
imports/strings and run tests/linters before merging.
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
