## In deployments/scripts/deploy-staging.sh around lines 52 to 64, the file defines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066855

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:64)

```text
In deployments/scripts/deploy-staging.sh around lines 52 to 64, the file defines
duplicate log_info/log_warn/log_error helpers; remove these local definitions
and instead source the shared logging lib (deployments/scripts/lib/logging.sh)
before any logging is used. Add a check that the logging.sh file exists and
source it (or exit with an error if missing) so the script fails fast when the
shared helper is unavailable; do not reimplement the functions locally to avoid
drift.
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
