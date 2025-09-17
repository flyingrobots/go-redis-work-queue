# API Overview

The project exposes several HTTP APIs for operations, observability, and automation. Each API follows the same versioning policy:

## Versioning & Deprecation Policy

- **Supported versions**: Current APIs are served under `/api/v1`. Minor and patch releases remain backward compatible.
- **Breaking changes**: Require design review, a published migration guide, and a new major version when compatibility cannot be maintained.
- **Deprecation window**: Deprecated endpoints or fields include `Deprecation`/`Sunset` headers and remain available for at least 90 days before removal.
- **Change tracking**: Release notes and the per-API changelog documents capture all published changes.
- **Client guidance**: Consumers should pin to versioned routes and monitor release notes; migration examples are included alongside the relevant type definitions (e.g., SLOConfig, BurnRateThresholds, Alert, MetricSnapshot).

Refer to individual API documents (e.g., `admin-api.md`, `anomaly-radar-slo-budget.md`) for endpoint-specific details.
