## In cmd/admin-api/main.go around lines 84 to 89, the current error check uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061081

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:89)

```text
In cmd/admin-api/main.go around lines 84 to 89, the current error check uses
os.IsNotExist which does not detect Viper's missing-config error; replace that
check to detect viper.ConfigFileNotFoundError instead (e.g. use a type
assertion: if _, ok := err.(viper.ConfigFileNotFoundError); ok { return cfg, nil
} ) and return other errors as before; add the viper import if not already
present (and remove or keep os import only if still used elsewhere).
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
