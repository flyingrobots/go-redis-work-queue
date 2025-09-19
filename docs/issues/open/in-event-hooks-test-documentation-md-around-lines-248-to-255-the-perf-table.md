## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 248 to 255, the perf table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033421

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:255)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 248 to 255, the perf table
incorrectly states Go 1.22.5 while repo go.mod files declare Go 1.25/1.25.0;
update the table to list Go 1.25 (or re-run the benchmarks under Go 1.22.5 and
replace the benchmark numbers if you prefer to keep 1.22.5), and add a note
about the exact Go toolchain used (including patch version) and where raw
outputs live; also ensure CI/workflows pin the Go version used for benchmarking
(update .github workflows to use go-version: 1.25.x) so results are
reproducible.
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
