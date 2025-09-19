## In demos/responsive-tui.tape around line 10, stop hard-coding zsh; change the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792460

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:10)

```text
In demos/responsive-tui.tape around line 10, stop hard-coding zsh; change the
"Set Shell \"zsh\"" directive to a portable shell (e.g., "Set Shell \"bash\"")
or remove the directive to use the system default shell so CI images without zsh
won't fail; update the line to use bash and ensure any script syntax in the tape
is compatible with bash.
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
