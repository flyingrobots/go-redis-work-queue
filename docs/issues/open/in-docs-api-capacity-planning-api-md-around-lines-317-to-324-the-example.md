## In docs/api/capacity-planning-api.md around lines 317 to 324, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039292

- [review_comment] 2025-09-18T15:56:38Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:324)

```text
In docs/api/capacity-planning-api.md around lines 317 to 324, the example
instructs importing an internal package which will fail for external consumers;
either move the package out of internal into pkg/ with a stable public import
path and update the example import to that new module path (and update
go.mod/tests/CI references accordingly), or explicitly mark this code block as
internal-only and remove/replace the import with the public, exported API you
want external users to consume; update the documentation to show the correct
public path and a short note about internal-only visibility.
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
