## In deployments/scripts/setup-monitoring.sh around line 11, the script blindly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039038

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:11)

```text
In deployments/scripts/setup-monitoring.sh around line 11, the script blindly
sources "${SCRIPT_DIR}/lib/logging.sh" which may not exist; add a guard that
checks the file is present and readable before sourcing, and if missing print a
clear error to stderr and exit non‑zero (fail fast). Use a conditional to test
-r or -f on "${SCRIPT_DIR}/lib/logging.sh" and only source it when the check
passes; otherwise echo a descriptive error to >&2 and exit 1 so ShellCheck
SC1091 is addressed.
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
