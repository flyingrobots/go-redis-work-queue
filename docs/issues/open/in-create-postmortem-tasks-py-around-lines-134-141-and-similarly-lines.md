## In create_postmortem_tasks.py around lines 134-141 (and similarly lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072286

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:141)

```text
In create_postmortem_tasks.py around lines 134-141 (and similarly lines
139-141), the JSON files are opened without specifying encoding and json.dump
defaults to ASCII-escaping non-ASCII chars; update both file writes to open(...,
'w', encoding='utf-8') and call json.dump with ensure_ascii=False and
deterministic options (e.g., sort_keys=True, keep indent) so output is UTF-8,
non-ASCII characters are preserved, and file contents are deterministic.
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
