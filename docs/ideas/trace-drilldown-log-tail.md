# Trace Drill‑down + Log Tail

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Med‑High | Observability / TUI | Trace propagation; log source | Log volume, PII | ~250–400 | Medium | 5 (Fib) | High |

## Executive Summary
Surface trace IDs in the TUI and provide a log tail pane with filters to accelerate RCA.

## Motivation
Tighten the feedback loop from failing jobs to actionable traces/logs.

## Tech Plan
- Ensure trace IDs captured in payload/metadata; configurable tracing base URL.
- Add “Open Trace” action (external link or inline spans summary).
- Implement lightweight log tailer with rate cap and filters by job/worker.

## User Stories + Acceptance Criteria
- As an SRE, I can open a job’s trace from the TUI.
- As a developer, I can tail logs filtered by job or worker.
- Acceptance:
  - [ ] Trace IDs visible in Peek/Info; action to open.
  - [ ] Log tail pane with follow mode, filters, and backpressure protection.
  - [ ] Configurable endpoints for tracing and logs.

## Definition of Done
Trace link + log tail pane shipped; docs include setup; basic perf validation under load.

## Test Plan
- Unit: parsing/extraction of IDs; log throttling logic.
- Manual: links open correct trace; tailing with filters; behavior under bursty logs.

## Task List
- [ ] Capture/propagate trace IDs
- [ ] Add Open Trace action
- [ ] Implement log tail pane with filters
- [ ] Docs and examples

