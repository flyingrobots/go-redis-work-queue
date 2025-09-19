## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 265, the perf tables

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061546

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:265)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 265, the perf tables
lack traceability metadata; update each table row (or add a single preface block
above the Unit and Integration test tables) to include the git commit SHA used,
the exact random test seed (if any), and exact tool/runtime versions (Go
version, OS, CPU/host, Redis/NATS/docker image tags, TLS/other flags). For each
row either append columns or add parenthetical metadata that lists: commit:
<full SHA>, seed: <value or "n/a">, tooling: Go <x.y.z>, OS <name + version>,
CPU <model>, Redis <version+source>, NATS <version+source>, Docker <version if
used>, and a path to the persisted raw output file (e.g.,
benchmarks/event-hooks/latest.txt or artifacts/...). Ensure format is consistent
across all rows and include the exact string values (not ranges or approximate
names).
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
