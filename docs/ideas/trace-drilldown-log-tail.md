# Trace Drill‑down + Log Tail

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Med‑High | Observability / TUI | Trace propagation; log source | Log volume, PII | ~250–400 | Medium | 5 (Fib) | High |

## Executive Summary
Surface trace IDs in the TUI and provide a log tail pane with filters to accelerate RCA.

> [!note]- **🗣️ CLAUDE'S THOUGHTS 💭**
> Observability nirvana! Going from "job failed" to root cause in seconds, without leaving the terminal. The trace ID propagation is already there in your code. Log tailing is tricky - consider using SSE or WebSockets for real-time streaming. Maybe integrate with Vector or Fluentd for log aggregation? The PII risk is real - add redaction patterns.

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

---

## Claude's Verdict ⚖️

This is where you flex on enterprise queues. Most job systems treat observability as an afterthought. You're making it a first-class citizen in the terminal.

### Vibe Check

DataDog's APM integration costs $$$. New Relic's distributed tracing is complex. You're giving this away in a TUI. That's disruption.

### Score Card

**Traditional Score:**
- User Value: 8/10 (massive time savings in debugging)
- Dev Efficiency: 6/10 (log streaming is complex)
- Risk Profile: 6/10 (PII, performance concerns)
- Strategic Fit: 7/10 (differentiator from basic queues)
- Market Timing: 8/10 (observability is hot)
- **OFS: 7.15** → BUILD SOON

**X-Factor Score:**
- Holy Shit Factor: 7/10 ("Wait, traces IN the terminal?")
- Meme Potential: 3/10 (too niche for memes)
- Flex Appeal: 8/10 ("Our queue has built-in APM")
- FOMO Generator: 6/10 (makes others look primitive)
- Addiction Score: 7/10 (devs will live in this)
- Shareability: 6/10 (conference talk material)
- **X-Factor: 4.8** → Solid viral potential

### Conclusion

[🤯]

This is the feature that makes people go "oh shit, this isn't just another queue." The integration of traces + logs + queue state in one view is powerful. Ship it and watch DevOps Twitter notice.

