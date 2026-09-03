# Per-key FIFO

## Decision

Use one Redis list per hashed ordering key, an O(1) FIFO ring of ready key
hashes, a set that deduplicates ready-or-leased keys, and one expiring lease per
in-flight key. Small Lua scripts make enqueue, claim, completion, retry, and
recovery indivisible.

This is the key-sharded-list option from the roadmap. It fits the motivating
workcell: throughput is modest, the number of independent file paths can be
large, and a crash must remain visible in the existing processing-list/reaper
protocol. It also leaves the existing empty-key enqueue storage path and
reliable-list transition unchanged.

Streams were rejected because a consumer group per key would create unbounded
group administration, while fixed stream partitions would permit hash
collisions to serialize unrelated keys. Scanning a shared priority list for an
unleased key was rejected because claim cost grows with blocked jobs and a hot
leased key can hide work behind it.

## Redis layout

The default keys are:

| Purpose | Key |
| --- | --- |
| Ready-key FIFO | `jobqueue:ordered:ready` |
| Ready-or-leased membership | `jobqueue:ordered:active` |
| Pending jobs for a key | `jobqueue:ordered:queue:<sha256>` |
| Lease for a key | `jobqueue:ordered:lease:<sha256>` |
| In-flight job | Existing `jobqueue:worker:<worker>:processing` |
| Worker liveness | Existing `jobqueue:processing:worker:<worker>` |

`<sha256>` is the lowercase hexadecimal SHA-256 digest of the exact UTF-8
ordering-key bytes. The original key remains in the job envelope, but never
becomes part of a Redis key. Braces, colons, Unicode, long paths, and binary-ish
UTF-8 byte sequences therefore cannot alter Redis glob or cluster-tag syntax.
The four ordered key names are configurable alongside the existing queue key
layout; both per-key patterns contain exactly one `%s` digest placeholder.

The active set means “this digest has either one ready token or one leased
job.” An enqueue only adds a ready token when `SADD active digest` returns one.
That invariant prevents a hot key from occupying the ready ring more than
once.

## State transitions

Redis serializes each Lua script, so “enqueue order” means the order in which
Redis accepts the enqueue scripts.

1. **Unordered enqueue:** if `OrderingKey == ""`, perform the existing
   `LPUSH <priority-list> <job>` operation. No ordered key is read or written.
2. **Ordered enqueue:** `LPUSH` the job onto its per-key list. If the digest was
   absent from the active set, add it and `LPUSH` one token onto the ready ring.
   New work enters on the left; workers consume the oldest work from the right.
3. **Claim:** atomically `RPOP` the oldest ready digest, acquire its lease with
   `SET NX PX`, `RPOP` its oldest job, `LPUSH` that exact envelope onto the
   existing per-worker processing list, and establish the existing worker
   heartbeat. The lease value is the globally unique worker ID. No handler can
   start without winning this transition.
4. **Heartbeat:** while the handler or retry backoff runs, refresh the existing
   worker heartbeat and extend the ordering lease only when its value still
   equals the worker ID. A stale worker can never extend a successor's lease.
5. **Success or dead letter:** only the lease owner may atomically remove the
   envelope from the processing list and append it to the destination. The
   script deletes the heartbeat and lease, then either puts the digest at the
   left of the ready ring when more same-key jobs remain, or removes the digest
   from the active set.
6. **Retry:** keep the lease alive through backoff. Then atomically remove the
   old processing envelope, `RPUSH` the updated retry envelope onto the per-key
   list, release ownership, and put the digest back on the ready ring. `RPUSH`
   makes the retry the next job consumed, ahead of later same-key work.
7. **Cancellation or process death:** stop refreshing ownership and leave the
   exact envelope in the existing processing list. Nothing acknowledges it.
8. **Recovery:** after both worker heartbeat and ordering lease expire, the
   existing reaper atomically removes the abandoned envelope, `RPUSH`es it onto
   the per-key list, and restores one ready token. A live completion and reaper
   recovery race on lease ownership plus `LREM == 1`; exactly one transition
   can win.

The ordering lease TTL is the configured worker heartbeat TTL. The refresher
runs every TTL/3. The reaper polls no slower than the lesser of five seconds and
TTL/4, so crash recovery occurs after lease expiry plus at most one reaper
interval. This is an at-least-once queue: a handler whose process dies after an
external side effect but before its Redis completion transition will run again.
Handlers must therefore be idempotent.

## Ordering, priority, and fairness

Priority selects the ordinary queue only for jobs without an ordering key. All
jobs for one non-empty key share one FIFO regardless of their stored priority;
ordering wins over priority within that key. Different key hashes may be leased
by different workers concurrently.

The ready list is a round-robin ring. A completed hot key is put on the left,
behind every older ready token, while claims come from the right. With `R`
other ready keys, at most `R` successful key claims occur before the hot key is
eligible again. After an ordered delivery, a worker gives ordinary priority
work one non-blocking claim opportunity before returning to the ordered ring.
During continuously ordinary traffic, it probes the ordered ring after at most
32 ordinary deliveries. Neither path can therefore starve the other.

No operation scans the active set or the per-key keyspace. Enqueue, claim,
completion, retry, and recovery are O(1); the 10,000-key measurement records
command latency and cardinality rather than hiding a key scan.

## Verification contract

Real-Redis tagged tests cover:

- 100 jobs on one key with eight workers: strict sequence and maximum one
  handler in flight;
- 100 jobs on ten keys with eight workers: strict sequence per key and observed
  cross-key concurrency;
- a killed/canceled owner: expiry, reaper recovery, and interrupted-job-first
  redelivery;
- braces, colons, and Unicode in the key; high/low jobs sharing a key; a handler
  longer than the lease TTL; and a repeated completion-vs-reaper race;
- 10,000 distinct ready keys without any `SCAN`; and
- an empty-key Redis-layout fixture identical to the pre-feature layout.

The tagged suite runs in CI against Redis 7. Focused miniredis tests cover the
scripts' deterministic state transitions where miniredis supports the used
commands.

## Performance record

Baseline captured from commit `87dabf61` on 2026-09-02 using the existing admin
bench harness, Redis on loopback, eight workers, 1,000 empty-key high-priority
jobs per run, 1,000 requested jobs/s, and a one-byte legacy bench payload:

| Run | Duration | Throughput |
| ---: | ---: | ---: |
| 1 | 1.051126 s | 951.36 jobs/s |
| 2 | 1.051547 s | 950.98 jobs/s |
| 3 | 1.051516 s | 951.01 jobs/s |
| 4 | 1.051594 s | 950.94 jobs/s |
| 5 | 1.051513 s | 951.01 jobs/s |

The baseline median is **1.051516 seconds / 951.01 jobs/s**. The same five-run
command will be repeated after implementation. “No measurable regression”
means the post-change median stays within 3% of baseline; this tolerance is
wider than the observed 0.05% baseline spread but narrow enough to expose an
extra Redis round trip on the empty-key path.

Post-change results from the same machine and command:

| Run | Duration | Throughput |
| ---: | ---: | ---: |
| 1 | 1.051556 s | 950.97 jobs/s |
| 2 | 1.051534 s | 950.99 jobs/s |
| 3 | 1.051630 s | 950.90 jobs/s |
| 4 | 1.051534 s | 950.99 jobs/s |
| 5 | 1.050661 s | 951.78 jobs/s |

The post-change median is **1.051534 seconds / 950.99 jobs/s**: throughput
changed by -0.002% and duration by +0.002%, far below the 3% regression
threshold and the practical resolution of this harness.

The real-Redis 10,000-distinct-key test enqueued all keys in **813.045 ms**
and claimed the oldest ready key in **203.917 µs**. Redis command statistics
showed no `SCAN`; the active-set cardinality and ready-list length were both
exactly 10,000 before the claim. The measured claim is therefore consistent
with the protocol's O(1) claim cost rather than a hidden ready-set scan.
