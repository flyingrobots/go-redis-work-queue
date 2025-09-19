## In demos/responsive-tui.tape around lines 28-29 (also apply the same fix at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792466

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:28)

```text
In demos/responsive-tui.tape around lines 28-29 (also apply the same fix at
85-86, 144-145, 231-232, 373-374): the script sets a fake terminal width via
"export COLUMNS=35" but never restores or unsets it, leaking the environment
variable to the rest of the session; change each snippet to save the prior
COLUMNS (e.g., OLD_COLUMNS="$COLUMNS"), set the test value, then after the test
restore the prior value (if non-empty) or unset COLUMNS (e.g., if [ -z
"$OLD_COLUMNS" ]; then unset COLUMNS; else export COLUMNS="$OLD_COLUMNS"; fi) so
downstream commands do not inherit the fake width.
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
