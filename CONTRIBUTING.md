# Contributing

## Workflow

1. Branch from `main` with a focused change.
2. Run `make bootstrap` once, then `gofmt -w` on changed Go files and `make format` for Python changes.
3. Add or update tests for behavior changes.
4. Run `make check` before opening a pull request.
5. Keep pull requests small and explain the task, evidence, and affected invariant.

CI runs quality checks, the test suite on Python 3.9 and 3.13 across Linux, macOS, and Windows, and a clean wheel build/install smoke test.

## Architecture changes

`SPEC.md` is normative. Consult `docs/reference/HANDOFF.md` before changing the rules-as-ledger model, task-boundary reflection, pending-rule gate, adapter boundary, or typed version metadata. Record intentional changes rather than silently reversing an established decision.

## Release direction

The production CLI and reusable packages are written in Go with Cobra and ship as a single binary. Python remains as a temporary compatibility implementation and for research integrations such as GEPA; users will not need Python to run the final CLI.
