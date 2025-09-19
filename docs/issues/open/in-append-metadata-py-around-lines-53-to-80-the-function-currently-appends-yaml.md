## In append_metadata.py around lines 53 to 80, the function currently appends YAML

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060966

- [review_comment] 2025-09-18T16:01:13Z by coderabbitai[bot] (append_metadata.py:80)

```text
In append_metadata.py around lines 53 to 80, the function currently appends YAML
front matter to the end of the file which breaks tooling; change the logic to
detect existing front matter at the top (use content.lstrip().startswith("---")
or content.startswith("---")) and, when no front matter exists, write the file
with yaml_metadata + "\n\n" + content.lstrip("\n") instead of appending at EOF;
keep the same read/try/except structure and update the checks and write call
accordingly so metadata is prepended with a blank line separator.
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
