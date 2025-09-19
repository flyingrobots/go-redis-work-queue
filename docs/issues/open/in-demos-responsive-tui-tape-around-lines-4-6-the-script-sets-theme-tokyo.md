## In demos/responsive-tui.tape around lines 4-6, the script sets Theme "Tokyo

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792445

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:6)

```text
In demos/responsive-tui.tape around lines 4-6, the script sets Theme "Tokyo
Night" and FontFamily "Fira Code" which are non-deterministic across runners;
update to either remove these assumptions or add deterministic fallbacks: either
drop the Theme/FontFamily lines, or change FontFamily to a comma-separated
fallback (e.g., "Fira Code, monospace") and ensure the test
environment/container includes the Fira Code font (or bundle the font into the
test image) and pin the theme resource so rendering is reproducible across
runners.
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
