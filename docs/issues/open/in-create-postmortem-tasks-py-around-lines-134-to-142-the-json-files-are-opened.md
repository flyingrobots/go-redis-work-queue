## In create_postmortem_tasks.py around lines 134 to 142, the JSON files are opened

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912526

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (create_postmortem_tasks.py:142)

```text
In create_postmortem_tasks.py around lines 134 to 142, the JSON files are opened
without an explicit encoding and json.dump is left to default ASCII-escaping;
update the two open() calls to specify encoding='utf-8' and call json.dump(...,
ensure_ascii=False, indent=2) so the files are written deterministically in
UTF-8 and non-ASCII characters are not escaped.
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
