# Anomaly Radar + SLO Budget

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Medium | Observability | Metrics exposure/sampling | Threshold tuning, noise | ~200–350 | Medium (per tick O(1)) | 5 (Fib) | Med‑High |

## Executive Summary
A compact widget showing backlog growth, error rate, and p95 with SLO budget and burn alerts.

> [!note]- **🗣️ CLAUDE'S THOUGHTS 💭**
> SRE candy! This is Google SRE book meets terminal aesthetics. Error budgets are how mature teams think about reliability. The burn rate calculation is key - alert on acceleration, not just threshold breaches. Consider adding a "time until budget exhausted" countdown for extra drama. Maybe integrate with PagerDuty when budget burns too fast?

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

---

## Claude's Verdict ⚖️

This transforms your queue from a tool into an SRE platform. Error budgets are how Google runs production. Having this in a terminal UI is chef's kiss.

### Vibe Check

DataDog SLO tracking costs $$$. PagerDuty's error budgets are enterprise-only. You're giving this away in a TUI. That's disruption with a capital D.

### Score Card

**Traditional Score:**
- User Value: 7/10 (critical for SRE teams)
- Dev Efficiency: 7/10 (2 weeks, math is known)
- Risk Profile: 7/10 (threshold tuning needed)
- Strategic Fit: 7/10 (positions as enterprise-ready)
- Market Timing: 7/10 (SRE practices are mainstream)
- **OFS: 7.0** → BUILD SOON

**X-Factor Score:**
- Holy Shit Factor: 5/10 ("Error budgets in my terminal!")
- Meme Potential: 6/10 (screenshot budget burns)
- Flex Appeal: 8/10 ("We track error budgets")
- FOMO Generator: 6/10 (makes others look amateur)
- Addiction Score: 8/10 (checked constantly)
- Shareability: 6/10 (SRE conference talks)
- **X-Factor: 5.1** → Solid viral potential

### Conclusion

[🌶️]

This is spicy ops tooling. Not revolutionary, but excellently executed. The combination of SLO tracking + terminal UI + real-time updates hits different. Ship this and watch SRE teams take notice.

