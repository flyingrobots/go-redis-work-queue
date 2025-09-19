# Add Priority Fairness via Token Buckets

## Summary
Current worker fetch loop favors high-priority queues, starving low-priority traffic. Introduce a lightweight token bucket or time-slice across priorities to guarantee periodic low-priority processing.

## Acceptance Criteria
- Fetch loop enforces configured weights (e.g., 8:2:1) using per-priority tokens or rotating windows.
- Low-priority queues still drain under sustained high-priority load in tests.
- Configuration allows tuning ratios without code changes.

## Dependencies / Inputs
- Worker fetch loop implementation.
- Configuration management for priority weights.

## Deliverables / Outputs
- Updated fetch algorithm with fairness mechanism.
- Benchmarks/tests demonstrating balanced throughput across priorities.
