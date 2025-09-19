## In deployments/admin-api/deploy.sh around lines 101-107 you currently print a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066685

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:107)

```text
In deployments/admin-api/deploy.sh around lines 101-107 you currently print a
readiness failure and continue; change the else branch so the script exits
non-zero (e.g., echo the failure to stderr and run exit 1) so a failed readiness
check fails the run; alternatively ensure the script is running with set -e and
propagate the curl failure, but the minimal fix is to add an exit 1 in the else
path after printing the failure.
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
