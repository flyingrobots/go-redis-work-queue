# Project Charter

- Last updated: 2025-09-12

## Executive Summary
Deliver a production-ready, Go-based Redis work queue with strong reliability and observability by 2025-11-07.

## Table of Contents
- [Goals and Objectives](#goals-and-objectives)
- [Stakeholders and RACI](#stakeholders-and-raci)
- [Success Metrics](#success-metrics)

## Goals and Objectives
- GA v1.0.0 by 2025-11-07.
- Reliable processing: retries, DLQ, reaper, graceful shutdown.
- Operational visibility: metrics, health, tracing.
- Performance: 1k jobs/min per 4 vCPU with p95 < 2s for small files.

## Stakeholders and RACI

| Role | Name | Responsibilities | R | A | C | I |
|------|------|------------------|---|---|---|---|
| Sponsor | James | Vision, funding, acceptance |  | X |  | X |
| Tech Lead | Maintainer (you) | Architecture, implementation, reviews | X |  | X | X |
| SRE | Platform Team | CI/CD, monitoring, SLOs |  |  | X | X |
| Security | AppSec Team | Vulnerability scanning, policy |  |  | X | X |
| Users | Operators | Deploy, run, provide feedback |  |  |  | X |

Legend: R=Responsible, A=Accountable, C=Consulted, I=Informed

## Success Metrics
- Availability: Ready > 99.9% over rolling 30 days.
- Reliability: Zero lost jobs; DLQ rate < 0.5% of consumed jobs.
- Performance: p95 processing duration < 2s for small files.
- Observability: 100% of operations emit metrics; tracing coverage ≥ 80% of job processing.
- Quality: Unit coverage ≥ 80% for core packages; CI green on main.

