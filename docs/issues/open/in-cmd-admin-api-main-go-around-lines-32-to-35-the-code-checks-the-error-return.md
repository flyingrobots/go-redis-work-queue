## In cmd/admin-api/main.go around lines 32 to 35, the code checks the error return

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061022

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (cmd/admin-api/main.go:35)

```text
In cmd/admin-api/main.go around lines 32 to 35, the code checks the error return
from fs.Parse even though the FlagSet was created with flag.ExitOnError so Parse
will never return — remove the dead if-block and simply call
fs.Parse(os.Args[1:]) (or assign its result to _ if you prefer) without handling
an error; ensure no other logic depends on that removed branch and keep the
program flow unchanged.
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
