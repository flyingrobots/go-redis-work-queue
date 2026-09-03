#!/bin/sh
set -eu

# Core and public-client coverage are part of the default suite as of Items 5 and 3.
expected_minimum=25

packages="$(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)"
actual="$(printf '%s\n' "$packages" | awk 'NF {count++} END {print count+0}')"

if [ "$actual" -lt "$expected_minimum" ]; then
	printf 'default test package count dropped: expected at least %s, got %s\n' \
		"$expected_minimum" "$actual" >&2
	printf '%s\n' "$packages" | awk 'NF' >&2
	exit 1
fi

printf 'default test packages: %s (minimum %s)\n' "$actual" "$expected_minimum"
