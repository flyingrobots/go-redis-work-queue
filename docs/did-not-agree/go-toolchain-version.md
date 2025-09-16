# Go Toolchain Version: 1.25.x

Decision

- Use Go 1.25.x across CI and release workflows; go.mod set to 1.25.0.

Rationale

- Addresses govulncheck-reported stdlib CVEs affecting earlier versions.
- CI and local builds succeed with 1.25.x in this environment; evidence (`go_info`) confirms runtime version.

Tradeoffs

- Requires builders to have Go 1.25 installed; older toolchains will not match CI.

Revisit Criteria

- If compatibility issues arise for consumers on earlier Go versions.

Future Work

- Consider matrix-testing across supported Go versions once stability goals are met.
