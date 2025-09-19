## In demos/responsive-tui.tape around lines 30 to 73, the script emits an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792467

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:72)

```text
In demos/responsive-tui.tape around lines 30 to 73, the script emits an
excessive sequence of individual "Type"/"Enter" steps to produce a static block;
compress these into a single paste/heredoc operation (e.g., one cat << 'EOF' ...
EOF paste) so the entire block is inserted in one step. Replace the repeated
Type/Enter lines with a single paste action that contains the full ASCII UI,
ensure correct quoting/escaping so no extra interpolation occurs, and remove the
redundant keystroke steps so the tape uses a single bulk-paste operation
supported by VHS.
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
