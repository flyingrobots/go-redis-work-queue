# ROADMAP — Queue Core Completion

Six work items that turn this repo from a queue **benchmark** into a queue **product**: a general-purpose, importable, Redis-backed work queue with real payloads, pluggable workers, and per-key ordering. Written 2026-09-02 from a source-level audit; every claim below was verified against the code at commit `442a3b11`, with file:line citations so you don't have to re-derive the findings.

**Why these six.** The crash-safety core was already built — reliable-queue pattern (`BRPOPLPUSH` to a per-worker processing list, heartbeat keys, reaper requeue: `internal/worker/worker.go:88-118`, `internal/reaper/reaper.go:40`), retries/backoff/DLQ (`internal/worker/worker.go:243-290`), circuit breaker, priorities, admin API (`internal/admin-api/server.go:97-122`), TUI, and Prometheus metrics. These six now-complete items supplied the connective tissue that lets applications use it: a payload, a handler, enqueue surfaces, per-key ordering, and tests that run honestly.

**Ground rules for every item** (this is a flyingrobots repo):
- Failing test first. Never weaken a test to pass it.
- Conventional commits. Update README and any touched docs in the same change.
- Do not touch the 29 unwired `internal/` feature modules (`visual-dag-builder`, `time-travel-debugger`, etc.) — they are out of scope for all six items. Scope discipline is a requirement, not a suggestion.
- Go 1.25, go-redis v9, stdlib `testing` + testify + miniredis are the established harness. Use them.

---

## Item 0 — Fix the idempotency config lie

**Status:** Complete on 2026-09-02. The runtime now rejects unsupported config
keys, and the example plus feature ledger no longer advertise exactly-once
wiring that does not exist.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. The example config advertises a feature that is not wired up, which will mislead any user who reads it.
>
> **Context (verified):** `config/config.example.yaml:47-49` sets `exactly_once.idempotency.enabled: true`. But `internal/config/config.go:204-207` has the four `SetDefault` calls for idempotency **commented out**, and the package import at `internal/config/config.go:11` is **commented out**. Grepping `exactly` across `internal/worker/`, `internal/producer/`, and `cmd/` returns zero hits — no execution path consumes the setting. The only consumer is `internal/admin-api/exactly_once_handler.go`, which reports stats for a subsystem nothing feeds. Note also that TWO parallel idempotency implementations exist: `internal/exactly_once/` and `internal/exactly-once-patterns/`.
>
> **Requirements:**
> 1. Decide with the repo owner (or default to the smaller change): either (a) remove the `exactly_once` block from `config.example.yaml` and mark the feature "not wired" in `docs/features-ledger.md`, or (b) actually wire `internal/exactly_once.CheckAndReserve` (`internal/exactly_once/idempotency.go:76`) into the worker's job-intake path behind the config flag. Option (a) is honest and cheap; option (b) is real work — do not do (b) casually, and if you do it, pick ONE of the two implementations and record why in the commit.
> 2. Either way: the example config, the config loader, and the features ledger must end up telling the same story.
>
> **Acceptance criteria:**
> - No config key exists in `config.example.yaml` that the loader ignores. (Test this generically if cheap: load the example config, assert every top-level key is consumed or explicitly listed as inert.)
> - `docs/features-ledger.md` reflects the actual wiring state.
> - If (b): a job enqueued twice with the same idempotency key is processed once, and the admin stats endpoint reflects the dedupe.
>
> **Test plan:**
> - *Golden:* load `config.example.yaml` through `internal/config`; assert the exactly_once state matches the documented behavior (absent for (a); enabled and consumed for (b)).
> - *Known failures (write these RED first):* for (b), the double-enqueue test must FAIL on current main (both copies process) before the wiring lands; for (a), a "no dead config keys" test must FAIL on current main because `exactly_once.*` is dead.
> - *Edges:* config file with the key absent entirely; key present but `enabled: false`; for (b), Redis restart between the two enqueues (idempotency key TTL semantics — document what is guaranteed).

---

## Item 1 — Give `Job` a payload

**Status:** Complete on 2026-09-02. Jobs now carry byte-exact opaque payloads
and optional caller-owned schemas; a typed, configurable 1 MiB guard rejects
oversized payloads before Redis is modified, while legacy job JSON remains
compatible.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. The `Job` type cannot carry work.
>
> **Context (verified):** `internal/queue/job.go:8-17` defines `Job{ID, FilePath, FileSize, Priority, Retries, CreationTime, TraceID, SpanID}`. There is no body/args/headers field of any kind — a job is a filepath and a filesize, because the repo was built as a load benchmark. Producers construct jobs only by walking a directory (`internal/producer/producer.go:33-43`).
>
> **Requirements:**
> 1. Add `Payload json.RawMessage` (or `[]byte` with documented encoding) and `PayloadSchema string` (a caller-owned version/type discriminator) to `Job`.
> 2. Preserve backward compatibility: jobs serialized by the current code (no payload field) must still deserialize — old fields keep their JSON names, new fields are `omitempty`.
> 3. Add a documented size guard: max payload size configurable (default e.g. 1 MiB), enforced at enqueue with a typed error — Redis lists are not blob storage; the limit and the "store a reference, not the bytes" guidance go in the README.
> 4. `FilePath`/`FileSize` stay (existing bench/demo flows keep working) but the doc comment must say they are legacy bench fields, not the payload.
>
> **Acceptance criteria:**
> - A job round-trips enqueue→Redis→dequeue with payload and schema intact, byte-for-byte.
> - A pre-change serialized job (fixture captured from current main) deserializes without error, payload empty.
> - Oversized payload at enqueue returns the typed error and enqueues nothing (queue length unchanged).
>
> **Test plan:**
> - *Golden:* round-trip a payload containing UTF-8 text, embedded JSON, and non-ASCII bytes; assert byte equality. Fixture-based backward-compat test with a JSON literal copied from current main's serialization.
> - *Known failures (RED first):* round-trip test written against current `Job` must fail to compile / fail on the missing field before the change.
> - *Edges:* empty payload; payload exactly at the size limit; payload one byte over; `PayloadSchema` empty (allowed? decide and test); JSON payload containing the string `"fail"` (must NOT trip the legacy failure oracle — see Item 2); priority values outside high/low.

---

## Item 2 — Real worker: a handler interface

**Status:** Complete on 2026-09-02. Workers now invoke a concurrent-safe
application handler, preserve the explicit benchmark handler as the default,
recover panics with stacks, leave canceled work for the reaper, and renew
heartbeats throughout long calls and retry backoff.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. The worker does not execute work — it simulates it.
>
> **Context (verified):** `internal/worker/worker.go:172-202`: processing is a cancellable `time.Sleep` proportional to `FileSize`, and success is `!canceled && !strings.Contains(strings.ToLower(job.FilePath), "fail")`. There is no handler registration anywhere (grepped `RegisterHandler|HandlerFunc|type Handler|Handle(ctx` across `internal/` and `pkg/`). Everything AROUND the middle is real and must be preserved: `BRPopLPush` intake with per-worker processing list (`worker.go:88-118`), heartbeat, retries/backoff/DLQ (`worker.go:243-290`), circuit breaker (`worker.go:80`), reaper requeue (`internal/reaper/reaper.go:40`).
>
> **Requirements:**
> 1. Define `type Handler func(ctx context.Context, job queue.Job) error` and give the worker a way to receive one (constructor arg or `Worker.Handle(h)`); optionally dispatch by `PayloadSchema` (`Worker.HandleSchema(schema string, h Handler)`) with a default handler fallback.
> 2. Handler result semantics, exactly: `nil` → success path (existing completion flow); `error` → existing retry/backoff flow, DLQ after `MaxRetries`; handler panic → recovered, treated as error, stack logged; ctx canceled (shutdown) → job returns to its processing list for the reaper, not DLQ.
> 3. The sleep-simulator becomes the explicit default/bench handler (preserving current bench behavior when no handler is registered), and the `"fail"`-substring oracle lives ONLY inside that bench handler.
> 4. Handler execution must respect the existing heartbeat: long handlers keep the heartbeat alive (extend it on a ticker) so the reaper doesn't steal in-flight jobs.
>
> **Acceptance criteria:**
> - A registered handler receives the exact enqueued job (ID and payload) and its return value drives completion/retry/DLQ exactly as specified above.
> - Bench mode (`--role=all` demo flow) behaves identically to current main when no handler is registered.
> - A job whose handler runs longer than the heartbeat TTL is NOT requeued by the reaper while the handler is alive.
>
> **Test plan (miniredis for unit, real Redis for the e2e tag):**
> - *Golden:* handler increments a counter and records the payload; enqueue 3 jobs; assert 3 invocations, payloads match, queue drained, completed count = 3.
> - *Known failures (RED first):* handler returning an error must land the job in DLQ after exactly `MaxRetries` attempts with backoff calls observed — write it against current main to prove there's no seam, then build the seam.
> - *Edges:* handler panics (job retries, worker survives); handler blocks past heartbeat TTL (reaper must NOT steal); ctx cancellation mid-handler (job requeued, not lost, not DLQ'd); two workers, five jobs (each processed exactly once — assert by unique-ID ledger); handler returning nil after the job was already reaper-requeued (double-completion must not corrupt counts — document the at-least-once semantics honestly).

---

## Item 3 — An enqueue surface: `pkg/client` + CLI

**Status:** Complete on 2026-09-02. The supported `pkg/queueclient` package now
aliases the durable core job type, validates single and transactional batch
enqueue operations, reports typed priority/payload/connection errors, and
mirrors Stats/Peek. The binary and Admin API use the same guard and shared key
layout. Duplicate explicit IDs remain separate at-least-once deliveries; empty
stdin is a valid payload. Item 4 now enforces every non-empty `OrderingKey`.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. Nothing outside this repo can enqueue a job.
>
> **Context (verified):** all queue logic lives under `internal/` (unimportable by other modules — the only exported packages are `pkg/anomaly-radar-slo-budget` and `pkg/chaos-harness`, neither the queue). The producer only enqueues by walking a directory (`internal/producer/producer.go:33-43`); `--role=admin` has `Stats|Peek|PurgeDLQ|StatsKeys|PurgeAll|Bench` (`internal/admin/admin.go`) but no enqueue; the admin HTTP API (`internal/admin-api/server.go:97-122`) has no enqueue endpoint.
>
> **Requirements:**
> 1. Create `pkg/queueclient` (exported): `New(redisOpts, cfg) (*Client, error)`, `Enqueue(ctx, Job) (id string, err error)`, `EnqueueBatch(ctx, []Job) error`, and read-only `Stats(ctx)` / `Peek(ctx, queue, n)` mirroring `internal/admin`. Move or alias the `Job` type so it is importable (`pkg/queueclient.Job` re-exporting the internal type is acceptable; a clean move is better — your judgment, recorded in the commit).
> 2. Same Redis key layout as the worker reads — one source of truth for key names (extract the key-name constants into the shared package; grep for every hardcoded key string and consolidate).
> 3. CLI: `--role=producer` gains an enqueue mode (or a new `enqueue` subcommand): reads payload from a file or stdin, takes `--schema`, `--priority`, `--ordering-key` (forward-compatible with Item 4), prints the job ID.
> 4. Admin HTTP API gains `POST /api/v1/enqueue` with the same size guard as Item 1, documented in the OpenAPI file the server already serves.
>
> **Acceptance criteria:**
> - An external Go module (test with a `go mod` fixture under `test/`, or a `go run` example in `examples/`) imports `pkg/queueclient`, enqueues, and a worker with an Item-2 handler processes it.
> - `echo '{"x":1}' | job-queue-system enqueue --schema demo.v1 --priority high` → job ID printed, job visible in `Peek`, processed by a running worker.
> - The HTTP enqueue returns 201 + job ID; oversized payload returns 4xx with the typed error's message; queue state unchanged on rejection.
>
> **Test plan:**
> - *Golden:* end-to-end: client enqueue → worker handler receives payload → completed count increments; CLI enqueue → `admin Stats` shows it.
> - *Known failures (RED first):* the external-import test fails on current main with the `internal/` compile error — capture that error in the test's docstring as the reason the package exists.
> - *Edges:* enqueue with Redis down (typed connection error, no partial state); batch where one job is oversized (define: whole batch rejected — document); duplicate explicit job IDs (define the semantics: last-wins vs reject — document and test); CLI with empty stdin; priority string outside high/low.

---

## Item 4 — Per-key FIFO (the new primitive)

**Status:** Complete on 2026-09-02. Non-empty ordering keys use hashed
per-key lists, an O(1) round-robin ready ring, compare-owned expiring leases,
heartbeat renewal, and atomic completion/retry/reaper transitions. Same-key
jobs are serial across all workers; different keys remain parallel. The
mechanism, crash boundary, fairness bound, and measurements are recorded in
`design/per-key-fifo.md`.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. This is the design-bearing item; read the context fully before writing code, and write a short design note (`design/per-key-fifo.md`) BEFORE the implementation commit.
>
> **Context (verified):** there is no per-key ordering and no lock/lease primitive anywhere: `SetNX` appears zero times in non-test code; the only Lua lives in `internal/advanced-rate-limiting/rate_limiter.go` and `internal/exactly_once/idempotency.go`. With N workers doing `BRPopLPush` off a shared list, two jobs touching the same resource are processed concurrently by design. The motivating consumer: a multi-agent workcell where jobs are file writes and writes to the SAME path must apply in enqueue order, while writes to different paths parallelize.
>
> **Requirements:**
> 1. Add `OrderingKey string` to `Job` (empty = today's unordered behavior, zero regression).
> 2. Guarantee: jobs sharing a non-empty `OrderingKey` are processed in enqueue order, at most one in flight at a time, across all workers. Jobs with different keys interleave freely.
> 3. Choose and defend ONE mechanism in the design note (options to weigh, not prescriptions): (a) key-sharded lists — `LPUSH queue:{orderingkey}` + a ready-set of keys workers claim via `SetNX` lease with TTL + heartbeat, releasing on completion; (b) a Lua-scripted claim that pops the next job only if its key is unleased; (c) Redis Streams consumer groups per key-hash partition. Whatever you choose: the crash story must go through the EXISTING reaper pattern (a worker dying mid-job must not wedge its key forever — lease TTL + reaper release).
> 4. Fairness: a hot key must not starve other keys (round-robin the ready-key claim, or document the starvation bound).
> 5. The bench path (empty key) must show no measurable regression (use the existing bench harness; record numbers in the design note).
>
> **Acceptance criteria:**
> - 100 jobs, one key, 8 workers: processed strictly in enqueue order (assert by sequence numbers in payloads), exactly-one-in-flight observed via instrumentation.
> - 100 jobs across 10 keys, 8 workers: per-key order holds; total wall-clock shows real parallelism across keys (assert > 1 key in flight concurrently at least once).
> - Kill a worker holding a key mid-job (test hook): the key unwedges within the lease TTL; the interrupted job is redelivered BEFORE any later job on that key (order preserved through the crash, given at-least-once redelivery).
>
> **Test plan (this item needs real Redis for the concurrency tests — tag them, but ALSO wire them into CI with the tag, see Item 5):**
> - *Golden:* the two acceptance scenarios above, plus empty-key jobs behaving exactly as current main (fixture comparison of Redis key layout).
> - *Known failures (RED first):* the 100-jobs-one-key-8-workers order test MUST fail on current main (interleaving observed) — this failure is the reason the feature exists; keep its RED output in the PR description.
> - *Edges:* ordering key with braces/colons/unicode (Redis key-safety — hash or escape, test both); lease TTL shorter than handler duration (heartbeat extension must hold the lease); the same key enqueued at both priorities (define: ordering wins over priority within a key — document); reaper and a live worker racing for an expired lease (exactly one wins — Lua or WATCH, prove with a stress loop); 10k distinct keys (ready-set scan cost — measure, put the number in the design note).

---

## Item 5 — Make the tests tell the truth

**Status:** Complete on 2026-09-02. The worker, queue, producer, reaper, and
breaker suites run by default and pass under the race detector; CI enables and
proves the Redis e2e smoke test; a 25-package minimum guards against silent
re-gating; and every remaining opt-in test names its gate and exit condition.

> **PROMPT**
>
> You are working in `go-redis-work-queue`. The test suite is majority-disabled and one CI step is vacuous.
>
> **Context (verified):** 931 test funcs across 120 files, but 71 of 120 files are behind opt-in `//go:build` tags — only 42% of tests compile by default, and the core worker's tests are among the gated. The gating was explicit triage: commits `573b78c7`, `6160c9e3`, `84b994ac`, and `be57f7ed` ("docs: track real-green follow-ups for gated suites"). The CI "E2E determinism (5x)" step runs `go test ./test/e2e -run TestE2E_WorkerCompletesJobWithRealRedis` WITHOUT `-tags e2e_tests`, while every file in `test/e2e/` requires that tag (`test/e2e/e2e_test.go:1`) — it compiles to zero tests and "passes" five times, proving nothing. The repo's own `test/README.md:20` documents the correct invocation. One default-run failure exists: `internal/theme-playground` `TestPersistenceManager_FilePermissions` expects `0644`, gets `0640` (umask-sensitive).
>
> **Requirements:**
> 1. Fix the vacuous CI step: add the tag so the e2e test actually runs against the CI Redis service (it exists in `ci.yml` already). If it then fails, THAT IS THE POINT — fix the code or the test honestly, never by re-gating.
> 2. Un-gate, in this order of importance: the core worker suite, queue/producer, reaper, breaker. For each un-gated file: make it pass against miniredis or the CI Redis honestly. Suites for the 29 unwired feature modules stay gated — add a one-line header comment to each gated file naming WHY it's gated and what un-gates it (feature unwired).
> 3. Fix the umask-sensitive permissions test (assert `& 0o600` semantics or set umask in the test — pick the honest oracle).
> 4. CI must fail if the default `go test ./...` run's package count drops (a canary against silent re-gating): record the expected "packages with tests" count in a script assertion.
>
> **Acceptance criteria:**
> - `go test ./...` (no tags) runs the worker/queue/producer/reaper/breaker suites and passes.
> - CI's e2e step demonstrably executes ≥1 test (assert on `go test -v` output in the workflow, e.g. grep for `--- PASS: TestE2E_`).
> - `go test ./... -count=1` exits 0 on a clean checkout on Linux and macOS (the umask fix).
>
> **Test plan:**
> - *Golden:* CI green with the e2e step provably non-empty; local default run green.
> - *Known failures (RED first, and this item is MADE of them):* before fixing, run the e2e step WITH the tag locally and record what actually breaks — that list is the work. Ditto each un-gated core suite.
> - *Edges:* CI without the Redis service reachable (the tagged suite must fail loudly, not skip silently — no `t.Skip` on connection error for the CI path); parallel `-race` runs of the newly un-gated worker tests (they were gated for a reason — find it).

---

## Suggested sequence and sizing

| Order | Item | Depends on | Rough size |
|---|---|---|---|
| 1 | Item 0 (config lie) | — | hours |
| 2 | Item 5 (tests tell the truth) | — | 1–2 days (unknowns live here) |
| 3 | Item 1 (payload) | — | half day |
| 4 | Item 2 (handler) | Item 1 | 1 day |
| 5 | Item 3 (client + CLI) | Items 1–2 | 1 day |
| 6 | Item 4 (per-key FIFO) | Items 1–3, design note first | 2–4 days |

Items 0/5/1 can start in parallel. Item 4 is the only one with real design risk; its design note is a deliverable on its own.

## First consumer (context for design decisions)

A 4-agent local workcell wants a write-job queue for a git repo: jobs = {target filepath, content, provenance}, per-filepath FIFO mandatory, a single downstream committer consumes completions. When Item 4's design decisions need a tiebreaker, decide in favor of this shape: modest throughput, strict per-key order, crash visibility over raw speed.
