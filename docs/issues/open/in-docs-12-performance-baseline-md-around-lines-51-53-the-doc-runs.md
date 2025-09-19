## In docs/12_performance_baseline.md around lines 51-53, the doc runs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067078

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/12_performance_baseline.md:53)

```text
In docs/12_performance_baseline.md around lines 51-53, the doc runs
./bin/job-queue-system without explaining how that binary is produced; add a
prerequisite build step immediately before "3) In one shell, run the worker"
that instructs readers to build the binary (for example: run make build or run
go build ./cmd/job-queue-system -o ./bin/job-queue-system) and mention the
resulting path ./bin/job-queue-system so users don’t have to guess.
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
