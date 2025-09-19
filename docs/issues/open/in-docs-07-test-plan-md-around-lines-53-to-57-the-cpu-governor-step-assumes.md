## In docs/07_test_plan.md around lines 53 to 57, the CPU governor step assumes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061220

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (docs/07_test_plan.md:57)

```text
In docs/07_test_plan.md around lines 53 to 57, the CPU governor step assumes
cpupower is present and runnable with sudo; on stock GitHub runners cpupower may
not exist and the step will fail — change the instructions to make this
best‑effort by first checking for cpupower (command -v cpupower) and only
attempting sudo cpupower frequency-set -g performance when available, allowing
the command to fail silently (e.g., || true), and apply the same guarded
approach for restoring the governor on exit so the job won’t fail if cpupower is
absent or non‑runnable.
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
