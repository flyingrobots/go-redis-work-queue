# Reaper: SCAN-Based Recovery vs Worker Registry

Decision

- Keep SCAN-based discovery of processing lists for v0.4.0-alpha, instead of maintaining a registry of active workers or relying on keyspace notifications.

Rationale

- Simplicity and robustness: SCAN requires no extra moving parts or configuration and tolerates sudden worker exits.
- Predictable load: bounded SCAN page sizes and periodic cadence maintain manageable overhead.

Tradeoffs

- SCAN is O(keys); at very large worker fleets, registry/notifications can reduce overhead.

Revisit Criteria

- If reaper CPU or Redis time spent on SCAN becomes material (observed via profiling/metrics) under expected fleet sizes.

Future Work

- Add optional worker registry with TTL stored in Redis; reaper would iterate registry members and target per-worker keys directly.
- Consider Redis keyspace notifications where operationally acceptable.
