## In deployments/scripts/health-check-rbac.sh around lines 354 to 380, the parsed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066924

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:380)

```text
In deployments/scripts/health-check-rbac.sh around lines 354 to 380, the parsed
TIMEOUT value isn't validated and non-integer input will cause arithmetic
failures later; after assigning TIMEOUT in parse_args (or immediately after
parse_args returns) validate it with a simple integer check (e.g. regex like
^[0-9]+$) and ensure it's positive, and if the check fails print a clear error
message including the invalid value and exit 1 so the script fails fast on
garbage input.
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
