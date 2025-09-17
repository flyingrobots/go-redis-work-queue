# Anomaly Radar API Contract Decisions (Draft)

_Last updated: 2025-09-16_

This note captures the outstanding decisions for the Anomaly Radar + SLO Budget API so we can resolve the remaining CodeRabbit chunk_008 feedback.

## Auth & Authorization

- **Current state**: All endpoints expect the Admin API bearer token. Enforcement is handled by shared middleware but scopes/roles are not documented.
- **Decision needed**:
  - Define canonical scopes (proposal: `slo_reader`, `slo_admin`).
  - Map each endpoint to required scope.
  - Decide whether health is public or still requires auth.
- **Next actions**:
  - Draft scope matrix in this document.
  - Update middleware to enforce scope checks once agreed.
  - Reflect changes in `docs/api/anomaly-radar-slo-budget.md` and OpenAPI security requirements.

## Error Envelope

- **Current state**: Documentation references `ErrorResponse` generically; handler implementations are inconsistent.
- **Decision needed**:
  - Confirm the standard error payload we will emit across Admin + Anomaly APIs (`error`, `code`, `details`, `request_id`).
  - Ensure validation errors return 422 with detailed field errors.
- **Next actions**:
  - Audit handlers, add shared helper for 4xx/5xx responses.
  - Update docs and OpenAPI responses to match final shape.

## Pagination Strategy

- **Current state**: `/metrics` endpoint returns bounded arrays without pagination; large windows could grow unbounded.
- **Options**:
  - Cursor-based pagination (`next_cursor` + `max_samples` hard limit).
  - Time-based pagination (pass `start`/`end` plus `page_token`).
- **Decision needed**: Choose approach + limits (proposal: cursor-based with default 1000 samples, hard cap 5000).
- **Next actions**:
  - Prototype pagination in handler (respecting Redis storage constraints).
  - Update doc / OpenAPI with new params and response fields.

## Timeline & Ownership

- Primary owner: Observability squad (@alice)
- Target milestone: align with “Publish Anomaly Radar OpenAPI spec + client CI” task.

Once these decisions are implemented we can revisit the chunk_008 review rejections and mark them resolved.

## Scope Matrix (Proposed)

| Endpoint | Method | Required Scope |
|----------|--------|----------------|
| `/api/v1/anomaly-radar/status` | GET | `slo_reader` |
| `/api/v1/anomaly-radar/metrics` | GET | `slo_reader` |
| `/api/v1/anomaly-radar/alerts` | GET | `slo_reader` |
| `/api/v1/anomaly-radar/slo-budget` | GET | `slo_reader` |
| `/api/v1/anomaly-radar/config` | GET | `slo_reader` |
| `/api/v1/anomaly-radar/config` | PUT | `slo_admin` |
| `/api/v1/anomaly-radar/start` | POST | `slo_admin` |
| `/api/v1/anomaly-radar/stop` | POST | `slo_admin` |

Health endpoint remains authenticated (no anonymous access) to avoid leaking status in restricted environments.

## Error Envelope Decision

- Adopt a shared helper that emits:
  ```json
  {
    "error": "human readable message",
    "code": "MACHINE_CODE",
    "details": {
      "field": "message"
    },
    "request_id": "uuid"
  }
  ```
- Validation failures return HTTP 422 with `validation_errors` array (field/message pairs).

## Pagination Decision

- Implement cursor-based pagination:
  - Request params: `window`, `max_samples` (default 1000, max 5000), `next_cursor` (opaque).
  - Response includes `next_cursor` (null when complete) and `count`.
- Update handler to page through stored snapshots and surface cursor tokens.
