## In deployments/scripts/setup-monitoring.sh around line 11, ShellCheck is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066944

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:11)

```text
In deployments/scripts/setup-monitoring.sh around line 11, ShellCheck is
complaining about sourcing a file (SC1091); add an explicit source hint comment
immediately above the source line to satisfy CI. Place a ShellCheck source
directive that points to the referenced file (for example: # shellcheck
source=./lib/logging.sh) on the line above the existing source
"${SCRIPT_DIR}/lib/logging.sh" so the linter recognizes the target and the
warning is silenced.
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
