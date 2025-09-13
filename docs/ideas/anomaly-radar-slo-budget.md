# Anomaly Radar + SLO Budget

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Medium | Observability | Metrics exposure/sampling | Threshold tuning, noise | ~200–350 | Medium (per tick O(1)) | 5 (Fib) | Med‑High |

## Executive Summary
A compact widget showing backlog growth, error rate, and p95 with SLO budget and burn alerts.

## Motivation
Provide immediate health signals and guide operational action.

## Tech Plan
- Compute rolling rates and percentiles with light sampling; thresholds for colorization.
- Configurable SLO target and window; simple burn rate calculation.

## User Stories + Acceptance Criteria
- As an SRE, I can see whether we’re inside SLO and how fast we’re burning budget.
- Acceptance:
  - [ ] Backlog growth, failure rate, and p95 displayed with thresholds.
  - [ ] SLO config and budget burn shown; alert when burning too fast.
  - [ ] Lightweight CPU/memory footprint.

## Definition of Done
Widget integrated; configs documented; behavior validated under synthetic load.

## Test Plan
- Unit: rolling window calcs; thresholding logic.
- Manual: verify color transitions and alerting conditions.

## Task List
- [ ] Implement rolling metrics
- [ ] Add SLO config + budget calc
- [ ] Integrate widget + thresholds
- [ ] Document usage and tuning

