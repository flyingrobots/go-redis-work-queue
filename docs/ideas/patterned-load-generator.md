# Patterned Load Generator

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Medium | Bench/Load | Existing bench plumbing | Overloading env, noisy graphs | ~200–350 | Medium (per tick O(1)) | 3 (Fib) | Medium |

## Executive Summary
Extend the bench tool to support sine/burst/ramp patterns, with guardrails and live visualization.

## Motivation
Validate behavior under realistic traffic and create great demos.

## Tech Plan
- Implement pattern generators; controls for duration/amplitude; guardrails (max rate/total).
- Overlay target vs actual enqueue rate on charts; profile persistence optional.

## User Stories + Acceptance Criteria
- As a tester, I can run predefined patterns and see accurate live charts.
- Acceptance:
  - [ ] Sine, burst, ramp patterns; cancel/stop supported.
  - [ ] Guardrails prevent runaway load.
  - [ ] Saved profiles can be reloaded.

## Definition of Done
Patterns implemented with charts; docs include examples and cautions.

## Test Plan
- Unit: pattern math; guardrails.
- Manual: visualize and compare patterns; cancellation behavior.

## Task List
- [ ] Implement sine/burst/ramp
- [ ] Add controls + guardrails
- [ ] Chart overlay target vs actual
- [ ] Save/load profiles

