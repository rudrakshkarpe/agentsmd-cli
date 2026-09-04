# Development setup

```sh
make bootstrap
go run ./cmd/agentsmd --version
```

The production implementation is Go, with the original dependency-free Python skeleton retained as a compatibility reference. The immediate milestone is Phase 2: normalize a real Claude Code transcript, reflect once at the session boundary, create a pending rule or no-op verdict, promote an approved rule, and demonstrate measurable token savings.

Project context and previous decisions are preserved in `docs/reference/HANDOFF.md`. The imported README is in `docs/reference/ORIGINAL-README.md`; the repository's active overview remains `README.md`.

## Useful checks

```sh
make format        # format and autofix imports
make go-check      # format, vet, race tests, and Go binary
make py-check      # Python compatibility and package gates
make check         # all local gates
```

CI applies the same gates on every pull request and push to `main`. It tests Python 3.9 and 3.13, covers Linux, macOS, and Windows, and builds and installs the wheel in a clean environment.

## Language plan

The production CLI and reusable library are written in Go with Cobra. The Python executable specification stays in the repository until compatibility fixtures prove that the Go implementation preserves the intended behavior. Users will receive a single binary through Homebrew and direct downloads.
