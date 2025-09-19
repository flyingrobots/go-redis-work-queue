## In demos/responsive-tui.tape around line 9, the TypingSpeed is set to 80ms which

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792455

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:9)

```text
In demos/responsive-tui.tape around line 9, the TypingSpeed is set to 80ms which
overly slows the demo; lower it to a more reasonable value (e.g., 10–25ms) to
reduce runtime. Edit the tape file to change "Set TypingSpeed 80ms" to a faster
value (pick one consistent with other demos) so the demo runs snappily while
preserving readability.
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
