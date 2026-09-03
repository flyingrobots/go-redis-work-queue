# Worker Runtime

The worker combines Redis reliable-list intake with an application handler. It
atomically moves a job from a configured priority queue to a per-worker
processing list, maintains a heartbeat while that job is active, and then
completes, retries, or dead-letters it based on the handler result. Jobs with a
non-empty `OrderingKey` use a hashed per-key FIFO and compare-owned lease, so
only one handler for that key runs at a time across all workers.

## Registering a handler

Applications outside this module use the supported `pkg/queueworker` package:

```go
cfg := queueworker.DefaultConfig()
wrk, err := queueworker.New(redisOptions, cfg, handler, logger)
if err != nil {
    return err
}
defer wrk.Close()
return wrk.Run(ctx)
```

Internal runtime code can install a handler before or during `Run`:

```go
wrk := worker.New(cfg, redisClient, logger)
wrk.Handle(func(ctx context.Context, job queue.Job) error {
    return process(ctx, job.ID, job.Payload, job.PayloadSchema)
})
```

The same handler may be called concurrently by `worker.count` goroutines for
unordered jobs or different ordering keys. Calls sharing one non-empty key are
serialized in Redis acceptance order, even when their priorities differ. The
handler must synchronize any other shared state and return when `ctx` is
canceled. Calling `Handle(nil)` clears the handler; it never selects benchmark
behavior implicitly.

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

A worker without a handler returns `ErrHandlerRequired` before consuming Redis
jobs. `BenchHandler` is the only code that interprets the legacy `FileSize`
delay and the `"fail"`-in-`FilePath` demo convention. It never inspects
application payload bytes and must be selected explicitly.

## Heartbeats and delivery guarantee

The worker sets a heartbeat before invoking the handler and refreshes it every
third of `worker.heartbeat_ttl`, including during retry backoff. Ordered jobs
also extend a lease only when its value still matches the worker ID. Normal
completion/retry cleanup first stops the refresher, then uses a Lua transition
that verifies lease ownership, so a late worker cannot acknowledge a job that
the reaper recovered.

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
  -run '^(TestE2E_WorkerCompletesJobWithRealRedis|TestE2E_PerKeyFIFO)' \
  -race -count=1 -v
```

The focused suites cover exact job delivery across two workers, retry/backoff
and DLQ counts, panic recovery with stack logging, shutdown retention, default
benchmark behavior, and protection from reaper theft during a long handler.
The tagged suite proves payload delivery, heartbeat renewal, strict same-key
order, cross-key concurrency, fair scheduling, hashed key safety, and ordered
crash recovery against Redis. See `design/per-key-fifo.md` for the protocol.
