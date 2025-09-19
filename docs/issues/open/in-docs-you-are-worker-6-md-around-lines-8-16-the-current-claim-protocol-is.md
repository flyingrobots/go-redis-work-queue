## In docs/YOU ARE WORKER 6.md around lines 8–16, the current claim protocol is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033401

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.md:16)

```text
In docs/YOU ARE WORKER 6.md around lines 8–16, the current claim protocol is
racy across filesystems; replace the mv-based approach with an atomic claim
procedure that stages the file on the target filesystem and performs an atomic
rename or uses an O_CREAT|O_EXCL lock to fail if another worker already claimed
it. Specifically: create a temp file in the worker directory (so it lives on the
same FS as the destination), copy the source into that temp, attempt an atomic
exclusive claim (e.g., create/link a lockname using O_CREAT|O_EXCL or ln to fail
if lock exists), on success rename the temp to the final target atomically,
remove the lock and then remove the original source only after verifying the
rename succeeded, and on any failure leave the source untouched and log the
error.
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
