# Forecasting & Recommendations

- **Status:** BROKEN (`go test ./internal/forecasting` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Holt-Winters tests expect the older seasonal window sizes; the new defaults demand more observations and return "insufficient observations".
- Recommendation engine assumes anomaly windows exist and panics on empty slices.

## Next steps
- Align the test fixtures with the new smoothing parameters or adapt the code to handle small samples.
- Harden anomaly handling before re-running the suite.
