# Event Hooks Test Status and Plan

This document records the Event Hooks coverage that exists in the repository
and the work required before the feature can claim integration, security, or
performance coverage.

## Current Status

As of 2026-09-03, `internal/event-hooks` builds and has seven tests in two
files. The package-level race and coverage run passes with 11.6% statement
coverage:

```bash
go test -race -cover ./internal/event-hooks -count=1
```

That percentage is an observed snapshot, not a coverage target or a
production-readiness claim. The package
[README](../../internal/event-hooks/README.md) still classifies Event Hooks as
scaffolding that requires manager, replay, and transport work.

List the exact live tests without executing them:

```bash
go test ./internal/event-hooks -list '^Test'
```

## Coverage That Exists

### Event types and filters

[types_test.go](../../internal/event-hooks/types_test.go) covers:

- event-type, queue, wildcard-queue, and minimum-priority matching in
  `TestEventFilter_Matches`;
- the values returned by `DefaultRetryPolicy`; and
- concurrent locking around `WebhookSubscription.FailureCount`.

It does not cover wildcard event patterns, subscription validation, bulk
filter selection, or backoff execution.

### Webhook delivery and registry

[webhook_test.go](../../internal/event-hooks/webhook_test.go) covers:

- one successful HTTP delivery through an `httptest` server, including the
  presence of signature and event headers;
- classification of an unreachable endpoint as a retryable
  `DeliveryError`;
- a rate-limit smoke path; and
- add, get, list, and remove operations on `WebhookDeliverer`.

The rate-limit test references a public `httpbin.org` endpoint and permits
multiple outcomes. It is not deterministic integration evidence and should be
replaced with a local server.

## Runnable Commands

Run only the Event Hooks package:

```bash
go test ./internal/event-hooks -count=1
go test -race ./internal/event-hooks -count=1
go test -cover ./internal/event-hooks -count=1
```

The default repository suite also includes these tests:

```bash
go test ./... -race -count=1
```

There is currently no Event Hooks build tag, standalone module, acceptance
script, benchmark suite, fuzz corpus, or service-backed integration command.

## Coverage That Does Not Exist

The following files were removed and have not been relocated:

- `test/integration/webhook_harness_test.go`;
- `test/integration/nats_transport_test.go`;
- `test/integration/dlh_replay_test.go`; and
- `test/fixtures/webhook_test_data.go`.

There are also no live Event Hooks files named `webhook_signature_test.go`,
`event_filter_test.go`, or `security_test.go`.

Consequently, selectors such as `TestWebhookHarness_`, `TestNATSTransport_`,
`TestDLH_`, and `TestSignatureService_` do not identify current tests.
The prior claims about 45+ tests, 150+ cases, security scenario percentages,
integration coverage percentages, benchmark percentiles, and archived
benchmark artifacts were not reproducible from the repository and have been
retired.

## Known Test Gaps

Before Event Hooks can be considered production-ready, tests still need to
cover:

- exact HMAC-SHA256 signature bytes, malformed signatures, and payload
  tampering;
- retry scheduling, maximum attempts, response classification, timeouts, and
  cancellation;
- payload inclusion, field filtering, redaction, size limits, and custom
  headers;
- deterministic rate limiting without public network access;
- concurrent delivery and subscriber lifecycle behavior;
- Redis-backed subscription creation, update, deletion, and reload;
- manager start/stop behavior and API handler authorization/error responses;
- NATS and JetStream publishing, subjects, headers, reconnects, and failures;
- dead-letter-hook storage, listing, replay, idempotency, and concurrency; and
- bounded-load performance and allocation measurements.

## Reconstruction Roadmap

### 1. Make webhook unit tests deterministic

- Replace the `httpbin.org` dependency with `httptest.Server`.
- Inject or expose the HTTP client, clock, and retry scheduler where necessary.
- Assert the request body, signature bytes, headers, redaction, and response
  classification.
- Run the package repeatedly under the race detector.

Acceptance command:

```bash
go test -race ./internal/event-hooks -count=20
```

### 2. Add local webhook integration coverage

- Exercise Event Bus to subscriber to HTTP receiver as one flow.
- Cover success, retryable and terminal responses, timeout, cancellation, and
  concurrent delivery.
- Keep the harness local and deterministic; do not depend on public services.

The integration selector and command should be documented only after its test
file exists.

### 3. Cover persistence, manager, and API behavior

- Use `miniredis` for subscription persistence and reload scenarios.
- Exercise manager lifecycle and route handlers through `httptest`.
- Verify cleanup when registration or subscription fails.
- Add authorization and stable error-envelope assertions when the routes are
  connected to the Admin API boundary.

### 4. Restore NATS integration deliberately

- Add a service-backed test target with an explicit build tag.
- Pin and document the NATS/JetStream version.
- Cover publish, subject generation, headers, reconnects, and unavailable
  service behavior.
- Make the CI job and local startup command part of the same change as the
  suite.

### 5. Implement and test dead-letter-hook replay

- Define the durable storage and replay contract first.
- Cover single and batch replay, missing entries, retry state, idempotency,
  concurrency, and cancellation.
- Do not claim DLH coverage while only HTTP queue dead-letter behavior exists.

### 6. Add security and performance gates

- Add tamper and redaction tests using fixed vectors.
- Add fuzzing only with a checked-in seed corpus and a documented bounded
  command.
- Add benchmarks only after the behavior is stable.
- Record environment, workload, raw output path, and repeat count with every
  published performance number.

## Definition of Done

Event Hooks testing is complete only when:

- all documented test files and selectors exist;
- unit tests are deterministic and require no public network;
- service-backed suites declare their dependencies and build tags;
- unit and integration suites pass under the race detector;
- security claims are backed by executable fixed-vector tests;
- performance claims link to checked-in raw evidence; and
- this document and the package README describe the same implementation state.
