## In deployments/scripts/setup-monitoring.sh around lines 117 to 120, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066974

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:120)

```text
In deployments/scripts/setup-monitoring.sh around lines 117 to 120, the
SCRIPT_DIR variable is being computed twice (once globally and again inside a
function); remove the duplicate computation by keeping the existing top-level
SCRIPT_DIR assignment and deleting the redundant assignment within the function,
then ensure any code inside the function references the top-level SCRIPT_DIR
variable (no redefinition) so the script still resolves paths correctly.
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
