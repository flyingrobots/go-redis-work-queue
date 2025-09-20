# Kubernetes Operator

- **Status:** BUILDS (`go build ./internal/kubernetes-operator/...` passes; controllers/webhooks still scaffolding)
- **Last checked:** 2025-09-18

## Notes
- Controllers, webhooks, and the Admin API client compile against controller-runtime v0.22; runtime logic remains stubbed.
- Envtest/kind integration is still missing; only `go build` is guaranteed right now.

## Next steps
- Flesh out Admin API interactions and add envtest scaffolding before reenabling the operator in CI.
- Revisit webhook validation/defaulting once persistence APIs are available.
