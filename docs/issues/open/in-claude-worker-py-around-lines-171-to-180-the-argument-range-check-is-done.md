## In claude_worker.py around lines 171 to 180, the argument range check is done

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856177

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:187)

```text
In claude_worker.py around lines 171 to 180, the argument range check is done
manually; replace it by letting argparse validate the range by adding
choices=range(1, 11) to parser.add_argument("--id", ...) and remove the
subsequent if args.id < 1 or args.id > 10: ... block; update the help text if
desired to reflect the enforced range.
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
