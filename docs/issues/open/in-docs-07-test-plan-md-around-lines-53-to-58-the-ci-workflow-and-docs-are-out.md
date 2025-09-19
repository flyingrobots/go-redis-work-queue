## In docs/07_test_plan.md around lines 53 to 58, the CI/workflow and docs are out

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039120

- [review_comment] 2025-09-18T15:56:35Z by coderabbitai[bot] (docs/07_test_plan.md:58)

```text
In docs/07_test_plan.md around lines 53 to 58, the CI/workflow and docs are out
of sync: workflows don't set GOMAXPROCS or BENCHMARK_SEED and the runner is not
pinned to the documented ubuntu-22.04/8vCPU instance; update the CI benchmarking
job to export BENCHMARK_SEED (and log it with results) and set GOMAXPROCS=8 (via
job env or export before running benchmarks), and either pin runs-on to the
documented ubuntu-22.04 runner type (or the exact instance type) in
.github/workflows/* where benchmarks run or change this doc line to match the
actual runner used; also ensure the synthetic producer accepts a seed parameter
and that the workflows record the chosen seed in the test artifacts/logs.
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
