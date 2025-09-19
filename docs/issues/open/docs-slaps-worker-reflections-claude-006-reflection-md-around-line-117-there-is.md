## docs/SLAPS/worker-reflections/claude-006-reflection.md around line 117: there is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814725

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-006-reflection.md:117)

```text
docs/SLAPS/worker-reflections/claude-006-reflection.md around line 117: there is
a stray internal '---' separator that can be mis-parsed as YAML front‑matter;
replace that internal '---' with '***' (or an explicit <hr/> or remove it) so
only the file header remains as YAML front matter, save and commit the change;
optionally grep the docs/SLAPS tree for other files containing internal '---'
separators and apply the same replacement to avoid front‑matter parsing issues.
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
