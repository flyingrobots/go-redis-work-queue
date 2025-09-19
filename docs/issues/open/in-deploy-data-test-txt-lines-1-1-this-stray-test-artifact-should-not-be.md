## In deploy/data/test.txt (lines 1-1) this stray test artifact should not be

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814679

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/data/test.txt:1)

```text
In deploy/data/test.txt (lines 1-1) this stray test artifact should not be
shipped; either delete the file from deploy/ or relocate it to a proper test
fixture path such as producer/testdata/input.txt (preferred for Go tooling), and
if you relocate it add a short README.md next to it explaining its purpose and
format so CI/images don’t pick up deploy/ artifacts by accident.
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
