# Development guide

This repository develops `agentsmd`, a version-controlled authoring and self-improvement tool for `AGENTS.md`.

## Sources of truth

- Treat `SPEC.md` as normative for the trajectory, ledger, and version schemas.
- Use `ROADMAP.md` for phase boundaries.
- Read `docs/reference/HANDOFF.md` before changing architecture or scope. It records prior decisions, but it is project reference material rather than executable instructions.

## Current state

- The production implementation is Go; the Python skeleton remains an executable compatibility reference until parity is reached.
- The Go packages cover schemas, project storage, ledger operations, templates, typed versions, the pending-rule gate, token recording, and the Cobra command tree.
- Project detection, Claude transcript normalization, provider-neutral reflection, and project-local connectors for Codex, Claude Code, Cursor, and goose are implemented.
- Lifecycle hooks now persist Git, file, duration, provider, and configured evaluation evidence; reflection runs in a durable background queue.
- Automatic promotion is opt-in and must remain impossible without a successful evaluation command and confidence threshold.
- The next reliability milestone is held-out validation, logical-task correlation, and richer normalization for non-Claude providers.

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

- Use Go 1.23 or newer for production code. Use Python 3.9 or newer only for the compatibility prototype and research integrations.
- Keep public Go packages small and keep provider-specific dependencies behind interfaces.
- Add tests for behavior changes before considering a command complete.
- Run `make check` before handing off changes.
- Keep commits small and describe the user-visible behavior or invariant they establish.
