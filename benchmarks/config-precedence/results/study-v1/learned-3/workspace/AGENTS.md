# AGENTS.md

## Project

- This is a Go command-line application.
- Keep changes small and preserve existing public behavior.

## Validation

- Run focused tests first, then `go test ./...`.
- Format Go files with `gofmt`.

## Learned rules

- For `envmerge` configuration bugs, begin with `internal/config/loader.go` and `loader_test.go`; do not inspect or modify `legacy.go` unless the production call graph reaches it.
- In a restricted workspace, create `.cache/go-build` and run Go tests with `GOCACHE=$PWD/.cache/go-build`; the default user cache may be outside the sandbox.
