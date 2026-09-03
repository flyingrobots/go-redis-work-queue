# Worker Runtime

The worker combines Redis reliable-list intake with an application handler. It
atomically moves a job from a configured priority queue to a per-worker
processing list, maintains a heartbeat while that job is active, and then
completes, retries, or dead-letters it based on the handler result.

## Registering a handler

Code inside this module can install a handler before or during `Run`:

```go
wrk := worker.New(cfg, redisClient, logger)
wrk.Handle(func(ctx context.Context, job queue.Job) error {
    return process(ctx, job.ID, job.Payload, job.PayloadSchema)
})
```

The same handler may be called concurrently by `worker.count` goroutines. It
must synchronize shared state and return when `ctx` is canceled. Calling
`Handle(nil)` restores `BenchHandler`.

## Result semantics

- `nil`: append the original job envelope to the completed list and remove it
  from the processing list.
- Error: apply exponential backoff, increment the serialized retry count, and
  requeue. After `worker.max_retries`, move the job to the dead-letter list.
  The setting counts retries after the initial call, so total calls are
  `max_retries + 1`.
- Panic: recover inside the worker, log the panic and stack, and treat it as an
  ordinary handler error. The worker goroutine continues.
- Worker context cancellation: stop renewing the heartbeat and leave the
  original job in the processing list. The reaper owns redelivery after the
  final heartbeat expires; cancellation does not consume a retry.

`BenchHandler` is the default when no application handler is registered. It is
the only code that interprets the legacy `FileSize` delay and the
`"fail"`-in-`FilePath` demo convention. It never inspects application payload
bytes.

## Heartbeats and delivery guarantee

The worker sets a heartbeat before invoking the handler and refreshes it every
third of `worker.heartbeat_ttl`, including during retry backoff. Normal
completion/retry cleanup first stops the refresher, then removes the heartbeat,
so a late refresh cannot race with cleanup.

Delivery is at least once, not exactly once. If Redis cannot observe a renewal,
the reaper may redeliver a job while the original handler is still finishing.
Handlers should therefore use the stable job ID as an idempotency key when side
effects require deduplication. The completed list records executions and may
contain more than one entry for a job ID after redelivery.

## Verification

```bash
go test ./internal/worker ./internal/reaper -race -count=5
E2E_REDIS_ADDR=127.0.0.1:6379 \
  go test -tags e2e_tests ./test/e2e \
  -run '^TestE2E_WorkerCompletesJobWithRealRedis$' -race -count=5 -v
```

The focused suites cover exact job delivery across two workers, retry/backoff
and DLQ counts, panic recovery with stack logging, shutdown retention, default
benchmark behavior, and protection from reaper theft during a long handler.
The tagged smoke proves payload delivery and heartbeat renewal against Redis.
