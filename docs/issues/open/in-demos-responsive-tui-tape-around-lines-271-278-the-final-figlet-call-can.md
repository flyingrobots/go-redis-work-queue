## In demos/responsive-tui.tape around lines 271-278 the final figlet call can

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569614

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/responsive-tui.tape:282)

```text
In demos/responsive-tui.tape around lines 271-278 the final figlet call can
throw when figlet isn’t installed; wrap the figlet invocation in a safe guard so
the demo won’t crash on clean systems by detecting availability (e.g., try/catch
around require/spawn or check for the binary) and providing a graceful fallback
(render plain text or a simpler ASCII header) when figlet isn’t present,
ensuring the tape continues without error.
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
