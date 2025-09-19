## In AGENTS.md around lines 566 to 594, headings inside the blockquote/admonition

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974220

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (AGENTS.md:594)

```text
In AGENTS.md around lines 566 to 594, headings inside the blockquote/admonition
make fragile anchors because GitHub renders anchors inconsistently; move the
heading(s) out of the blockquote or add plain headings immediately after the
admonition so stable anchors exist (e.g., keep the admonition content but
duplicate the heading as a non-blockquote line right after, or convert the
blockquoted heading into a normal heading and keep the admonition content below)
to ensure TOC links and anchor targets resolve reliably.
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
