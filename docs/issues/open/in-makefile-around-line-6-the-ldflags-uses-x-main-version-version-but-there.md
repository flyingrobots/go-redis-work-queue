## In Makefile around line 6: the LDFLAGS uses -X main.version=$(VERSION) but there

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067141

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (Makefile:6)

```text
In Makefile around line 6: the LDFLAGS uses -X main.version=$(VERSION) but there
is no package-level var named version in any package main; either add a
package-level variable declaration like `var version string` in your app's main
package (e.g., cmd/<app>/main.go) or change the -X value to the correct
fully-qualified import path and symbol that actually exists (e.g., -X
github.com/your/module/cmd/<app>.version=$(VERSION)); update the Makefile or the
main package accordingly so the linker symbol matches an existing variable.
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
