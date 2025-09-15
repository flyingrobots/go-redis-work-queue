# Rate Limiter: Fixed Window vs Token Bucket

Decision

- Keep fixed-window limiter (INCR + 1s EXPIRE + TTL-based sleep with jitter) for v0.4.0-alpha.

Rationale

- Simplicity and predictable behavior for batch-like producers.
- TTL-based sleep avoids busy waiting; jitter reduces thundering herd.
- Meets initial SLOs with fewer moving parts; easier to operate/debug.

Tradeoffs

- Boundary bursts possible vs sliding window/token bucket.
- At very high RPS, token bucket better smooths flow.

Revisit Criteria

- If sustained RPS saturation shows boundary spikes causing Redis latency or worker oscillations.
- If customer use-cases require fine-grained smoothing or multiple producers coordinating tightly.

Future Work

- Evaluate token bucket in Redis using LUA script to atomically debit tokens per producer key.
