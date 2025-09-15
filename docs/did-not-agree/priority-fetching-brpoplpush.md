# Priority Fetching and BRPOPLPUSH Semantics

Decision

- Use per-queue BRPOPLPUSH with short timeout to emulate multi-queue priority rather than a single command returning queue name and value.

Rationale

- Redis does not provide multi-source BRPOPLPUSH. Looping priorities with a short timeout preserves atomic move semantics per queue and delivers predictable prioritization.
- go-redis returns only the value for BRPopLPush. We record the source queue implicitly by the loop order and use the known `srcQueue` when processing.

Tradeoffs

- Lower-priority jobs may incur up to the per-queue timeout in latency when higher-priority queues are empty.
- We do not rely on returned queue name; this is documented and tested.

Revisit Criteria

- If sub-second latency for low priority becomes unacceptable or we need multi-queue fairness beyond simple priority preference.

Future Work

- Explore a Lua-assisted sweep to pick the first non-empty queue without waiting the full timeout per queue in sequence.
