# Kubernetes Operator

- **Status:** BROKEN (`go build ./internal/kubernetes-operator/...` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Controller/webhook packages were copied from an early design and still reference scaffolding that was removed during the controller-runtime upgrade.
- Tests require kind/envtest setup that is not vendored; default builds fail before even linking.

## Next steps
- Re-generate controllers with the current controller-runtime version and commit envtest scaffolding.
- Ensure controllers/webhooks compile and basic reconciliation tests pass before reenabling in CI.
