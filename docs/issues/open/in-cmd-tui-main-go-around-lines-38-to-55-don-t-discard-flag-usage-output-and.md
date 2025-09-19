## In cmd/tui/main.go around lines 38 to 55, don't discard flag usage output and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033071

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:55)

```text
In cmd/tui/main.go around lines 38 to 55, don't discard flag usage output and
add a --version flag: remove or stop calling fs.SetOutput(io.Discard) so
help/usage is printed to the user (use the default or os.Stdout), add a new flag
(e.g., boolVar(&showVersion, "version", false, "Show version and exit")), and
after parsing check if showVersion is set and print the program version string
to stdout then exit; keep parse error handling but ensure normal --help and
--version both produce visible output.
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
