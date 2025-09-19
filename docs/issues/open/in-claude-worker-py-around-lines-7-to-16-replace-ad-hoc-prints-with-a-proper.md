## In claude_worker.py around lines 7 to 16, replace ad-hoc prints with a proper

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856164

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:16)

```text
In claude_worker.py around lines 7 to 16, replace ad-hoc prints with a proper
Python logger: import the logging module, configure basic logging (level and
format) or load config, and create a module-level logger via
logging.getLogger(__name__); then replace all print(...) calls across the file
with appropriate logger methods (logger.debug/info/warning/error/critical)
according to message severity. Ensure logger configuration happens once at
process startup (not inside functions) and avoid printing sensitive data; keep
fallback to stdout only for local dev if needed.
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
