## In demos/responsive-tui.tape around lines 12 to 16, the demo relies on the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792463

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:16)

```text
In demos/responsive-tui.tape around lines 12 to 16, the demo relies on the
host's locale so emoji and box-drawing characters can render incorrectly;
explicitly set the UTF-8 locale at the top of the tape or before printing UI
content (for example export LANG and LC_ALL to an en_US.UTF-8 or similar UTF-8
locale, or invoke a locale-safe wrapper) and add a short runtime check/fallback
that warns and exits if UTF-8 is not available so the emojis/box drawing render
consistently.
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
