## deployments/scripts/lib/logging.sh lines 4-6: the guard pattern that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033228

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/lib/logging.sh:6)

```text
deployments/scripts/lib/logging.sh lines 4-6: the guard pattern that
returns/exists when the script is already sourced triggers shellcheck SC2317;
annotate it to avoid accidental future changes. Add a ShellCheck directive
immediately above the guard (e.g. a comment disabling SC2317) so the intentional
return/exit is documented and not flagged, and include a brief comment
explaining why the guard is needed; do not change the guard logic itself.
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
