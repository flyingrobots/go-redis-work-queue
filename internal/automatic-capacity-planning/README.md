# Automatic Capacity Planning

- **Status:** BROKEN (`go test ./internal/automatic-capacity-planning` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- The forecasting harness still expects the v1 EWMA/linear timestamp model. The new code returns absolute wall-clock times and needs historical seeding, so tests raise "insufficient history" and timestamp mismatch assertions.
- Simulator and recommendation tests panic because the queueing/recommendation pipeline was only partially ported (e.g. anomaly recommendation assumes non-empty violation slices).

## Next steps
- Rework `forecaster.Predict` tests (or adapter) to seed history for the new algorithms.
- Finish the simulator/recommendation migration and update expectations.
- Once fixed, re-enable the package in the default test matrix.
