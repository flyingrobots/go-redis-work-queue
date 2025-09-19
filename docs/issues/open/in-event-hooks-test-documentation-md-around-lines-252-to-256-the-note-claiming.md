## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 252 to 256, the note claiming

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061582

- [review_comment] 2025-09-18T16:01:20Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:256)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 252 to 256, the note claiming
“Captured with `BENCH_MEM=1`” is incorrect and the reproduce instruction is
vague; update the table's note to state the real flag `-benchmem` (e.g.,
"Captured with `-benchmem` to record allocations"), and modify the reproduce
paragraph to show concrete, reproducible commands using `go test -bench=...
-benchtime=... -benchmem > benchmarks/event-hooks/latest.txt` (or one file per
benchmark), ensuring the examples include the exact `-bench` pattern,
`-benchtime`, `-benchmem`, and redirection to persist raw output alongside the
commit.
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
