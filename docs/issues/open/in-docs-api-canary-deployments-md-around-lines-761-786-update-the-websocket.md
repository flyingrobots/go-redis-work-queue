## In docs/api/canary-deployments.md around lines 761–786, update the WebSocket

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912812

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:786)

```text
In docs/api/canary-deployments.md around lines 761–786, update the WebSocket
section to use wss by default and add an Authentication subsection: replace the
ws:// URL with wss://, state that clients should send Authorization: Bearer
<token> during the WebSocket handshake as the preferred method, document support
for an optional ?token=<...> query param only if the server enables it, and
include a brief wscat example showing how to connect with an Authorization
header (e.g., wscat -c wss://... -H "Authorization: Bearer <token>") so readers
know how to pass the bearer token.
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
