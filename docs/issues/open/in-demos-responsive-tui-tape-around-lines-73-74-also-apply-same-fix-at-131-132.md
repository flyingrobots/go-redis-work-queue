## In demos/responsive-tui.tape around lines 73-74 (also apply same fix at 131-132,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061143

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (demos/responsive-tui.tape:74)

```text
In demos/responsive-tui.tape around lines 73-74 (also apply same fix at 131-132,
217-218, 311-313): the test sets a fake COLUMNS environment variable but does
not restore the original value, leaking the fake into downstream steps; modify
each section to save the original value (e.g., prev="$COLUMNS" or detect unset),
set the fake COLUMNS for the test, and then after the section restore it by
exporting COLUMNS="$prev" if prev was set or by unsetting COLUMNS if prev was
originally unset so downstream steps see the original environment.
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
