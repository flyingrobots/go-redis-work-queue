# Architecture

- Last updated: 2025-09-12

## Executive Summary

Single Go binary operates as producer, worker, or both. Redis provides prioritized lists, per-worker processing lists, and heartbeats. A reaper rescues orphaned jobs. Circuit breaker protects Redis during failure spikes. Observability includes Prometheus metrics and optional OTEL tracing.

## Table of Contents

- [System Diagram](#system-diagram)
- [Components](#components)
- [Data Flows](#data-flows)
- [Technology Stack](#technology-stack)
- [Scaling Strategy](#scaling-strategy)
- [Performance Targets](#performance-targets)

## System Diagram

```mermaid
flowchart LR
  subgraph Proc[Producer]
    FS[Filesystem Scanner]-->SER[JSON Serialize]
    SER-->QSel[Select Priority Queue]
    QSel-->LP[LPUSH]
  end

  subgraph Redis[(Redis)]
    HQ[high_priority]
    LQ[low_priority]
    PL[(worker:<id>:processing)]
    HB[(processing:worker:<id>)]
    DLQ[dead_letter]
    CMP[completed]
  end

  subgraph Work[Worker Pool]
    PRIO{Priority Loop}
    PRIO-->BRP[BRPOPLPUSH]
    BRP-->HBSet[SET EX Heartbeat]
    HBSet-->Exec[Process Job]
    Exec-- success -->Done[LPUSH completed + LREM proc + DEL HB]
    Exec-- fail -->Retry[Increment Retry, Backoff]
    Retry-- requeue -->LP2[LPUSH priority queue]
    Retry-- DLQ -->DL[LPUSH dead_letter]
  end

  Reap[Reaper]-->Scan[SCAN processing lists]
  Scan-->Rescue[RPOP processing -> LPUSH orig queue]

  Proc-->Redis
  Work-->Redis
  Reap-->Redis
```

## Components

- Producer: scans directories, prioritizes files, rate-limits enqueue.
- Worker: prioritized fetch via short-timeout BRPOPLPUSH per queue; heartbeat and cleanup.
- Reaper: rescues jobs from processing lists when heartbeats expire.
- Circuit Breaker: sliding window with cooldown; pauses fetch on high failure rate.
- Observability: metrics server, structured logging, optional OTEL tracing.

## Data Flows

1) Produce: file -> Job JSON -> LPUSH to `high` or `low`.
2) Consume: BRPOPLPUSH from `high` then `low` (short timeout) -> processing list.
3) Heartbeat: SET `processing:worker:<id>` with EX=ttl, value=payload.
4) Complete: LPUSH `completed`; LREM processing; DEL heartbeat.
5) Fail: increment retries; backoff; LPUSH back or DLQ after threshold.
6) Reap: if heartbeat missing, RPOP processing list items back to originating priority queue.

## Technology Stack

- Language: Go 1.21+
- Redis Client: go-redis v8
- Metrics: Prometheus client_golang
- Logging: zap (JSON)
- Tracing: OpenTelemetry (OTLP HTTP exporter)
- CI/CD: GitHub Actions
- Container: Distroless base image

## Scaling Strategy

- Horizontal scale workers across nodes; each worker uses `PoolSize=10*NumCPU`.
- Tune `MinIdleConns`, timeouts, and backoff per environment.
- Shard queues by prefix if needed: e.g., `jobqueue:high:0..N` with consistent hashing.

## Performance Targets

- Throughput: ≥ 1k jobs/min per 4 vCPU worker node (small files <1MB)
- Latency: p95 < 2s for small files end-to-end under normal load
- Recovery: < 10s to drain orphaned processing lists after crash of a single worker node
