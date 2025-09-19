## In README.md around lines 166 to 185, the provided docker run example starts the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569857

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:185)

```text
In README.md around lines 166 to 185, the provided docker run example starts the
image with --role=all but does not mount the required configuration or provide
environment overrides, which will cause startup failures; update the example to
show a bind‑mount of the host config (e.g., -v
./config/config.yaml:/app/config/config.yaml) and/or pass an env file or
explicit ENV mappings (e.g., --env-file env.list or -e KEY=VALUE) so the
container has required config values, and add a short note telling users to
ensure their local config and env file exist before running.
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
