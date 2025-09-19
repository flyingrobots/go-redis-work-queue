## In config/config.example.yaml around lines 72 to 75, the example DSN includes a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856191

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (config/config.example.yaml:75)

```text
In config/config.example.yaml around lines 72 to 75, the example DSN includes a
fake but parseable credential which can trigger scanners and encourage bad
habits; replace the DSN value with a non-parsable placeholder (e.g. an empty
string or clearly non-credential placeholder like "<DSN_HERE>") and remove any
fake username/password, and add a commented example environment variable entry
(OUTBOX_DSN) showing how to supply the DSN via env with a note marking it as
secret (e.g. "# OUTBOX_DSN (secret):
postgresql://user:password@host:port/db?sslmode=... — DO NOT COMMIT real
credentials"); ensure the file contains no parseable fake secrets.
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
