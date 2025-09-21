# requestidlint

Static analyzer that keeps the Admin API’s HTTP handlers consistent about
emitting the `X-Request-ID` header. The linter walks Go files under
`internal/admin-api` (skipping tests) and reports any handler that:

- calls `http.Error` directly, or
- writes a status code ≥ 400 via `WriteHeader` outside the approved
  helpers (`writeError`, `writeJSON`).

The goal is to force all error responses through the shared helpers so the
request ID pipeline never regresses.

## Running the analyzer

```
go run ./tools/requestidlint/cmd/requestidlint ./...
```

The command accepts the same package patterns as `go test`. When run at the
module root it exits 0 if everything is compliant, otherwise it prints
diagnostics with file/line references and a non-zero exit code.

## Tests

Fixture-based tests live beside the analyzer. Run them with:

```
go test ./tools/requestidlint
```

The fixtures under `testdata/` model both valid and invalid handler
implementations so the analyzer stays accurate as the helpers evolve.
