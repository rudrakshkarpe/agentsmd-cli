# AGENTS.md

## Project

- This is a Go command-line application.
- Keep changes small and preserve existing public behavior.

## Validation

- Run focused tests first, then `go test ./...`.
- Format Go files with `gofmt`.

## Learned rules

- For retry-delay bugs, begin with `internal/retry/policy.go`; `legacy_seconds.go` is compatibility-only unless the production call graph reaches it.
- In a restricted workspace, run Go tests with `GOCACHE=$PWD/.cache/go-build` after creating that directory.
