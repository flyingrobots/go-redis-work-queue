# Metrics Scope: What We Deferred

Decision
- Added essential metrics (job counters, duration histogram, queue length, breaker state, trips, reaper recovered, worker_active). Defer additional metrics like worker restart count and Redis pool internals for now.

Rationale
- Worker restarts are process-level events better captured by orchestrator; tracking inside the binary can be misleading without a supervisor.
- Redis pool internals vary by client/runtime; better to surface via existing client/exporter when needed.

Tradeoffs
- Less granular visibility into certain failure modes without external instrumentation.

Revisit Criteria
- If operators need in-binary restart counts for specific environments without orchestration.
- If visibility gaps are identified during soak/chaos tests.

Future Work
- Integrate with process metrics (e.g., kube-state-metrics) and Redis exporter for pool stats.
