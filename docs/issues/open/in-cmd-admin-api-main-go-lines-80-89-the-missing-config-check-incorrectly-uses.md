## In cmd/admin-api/main.go lines 80-89, the missing-config check incorrectly uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033042

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/admin-api/main.go:89)

```text
In cmd/admin-api/main.go lines 80-89, the missing-config check incorrectly uses
os.IsNotExist on the error from v.ReadInConfig; Viper returns
viper.ConfigFileNotFoundError instead. Replace the os.IsNotExist check with a
type check for viper.ConfigFileNotFoundError (e.g., via errors.As) and only
return the default cfg, nil in that case; for all other errors from
ReadInConfig, return the error as-is.
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
