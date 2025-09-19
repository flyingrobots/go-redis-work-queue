## .github/workflows/ci.yml around lines 54 to 62: the CI job uses Bash-specific

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724332

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/ci.yml:62)

```text
.github/workflows/ci.yml around lines 54 to 62: the CI job uses Bash-specific
brace expansion for the loop and lacks timestamps; make the loop POSIX-shell
safe (e.g., use seq or a while counter) so it works under sh/other runners, and
prefix/append each test run with timestamped log lines (use date) to aid
debugging and measure duration; keep set -euo pipefail and ensure any non-zero
test causes workflow failure.
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
