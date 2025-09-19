## In docs/YOU ARE WORKER 6.rb around line 28, the mv command may mis-handle

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814732

- [review_comment] 2025-09-16T22:46:58Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:28)

```text
In docs/YOU ARE WORKER 6.rb around line 28, the mv command may mis-handle
filenames that begin with a dash; update the command to include the
end-of-options marker so it becomes mv -n --
"slaps-coordination/open-tasks/P1.T001.json" "slaps-coordination/claude-001/" to
ensure paths starting with “-” are treated as operands rather than options.
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
