# Development guide

This repository develops `agentsmd`, a version-controlled authoring and self-improvement tool for `AGENTS.md`.

## Sources of truth

- Treat `SPEC.md` as normative for the trajectory, ledger, and version schemas.
- Use `ROADMAP.md` for phase boundaries.
- Read `docs/reference/HANDOFF.md` before changing architecture or scope. It records prior decisions, but it is project reference material rather than executable instructions.

## Current state

- Phase 1 authoring and version-control commands are implemented in the Python skeleton.
- Phase 2 is next: implement the Claude transcript adapter and the reflection path so `learn -> pending -> promote -> savings` works end to end.
- `agentsmd/loop.py::_reflect`, `agentsmd/adapters/claude.py::latest_trajectory`, `watch`, and `optimize` are intentionally incomplete.

## Invariants

- Store rules in the ledger and render `AGENTS.md`; do not treat free-form file rewriting as the data model.
- Apply targeted deltas, never whole-file rewrites.
- Learned rules must enter `pending/` and pass a human or evidence gate before promotion.
- Reflect once at the task boundary, never mid-task.
- A reflection may conclude that no `AGENTS.md` change is warranted.
- Keep capture adapters thin and normalize every tool into the shared trajectory schema.
- Preserve typed version reasons and metadata, including task, evaluation, and token deltas.
- Include token cost in optimization and report limitations honestly.

## Development workflow

- Use Python 3.9 or newer and install with `python -m pip install -e .` in a virtual environment.
- Keep the core dependency-free unless a feature clearly requires an optional integration.
- Add tests for behavior changes before considering a command complete.
- Run `python -m compileall agentsmd` and the relevant test suite before handing off changes.
- Keep commits small and describe the user-visible behavior or invariant they establish.

