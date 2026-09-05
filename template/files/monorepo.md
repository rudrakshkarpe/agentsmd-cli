# AGENTS.md

## Repository map

- Record workspace roots, package ownership, shared libraries, applications, and generated-code boundaries.
- Read the nearest package-level AGENTS.md or README before editing.

## Monorepo workflow

- Identify the owning package and read its local guidance before editing.
- Use workspace-level dependency commands; do not create nested lockfiles.
- Validate affected packages first, then run repository-wide checks.
- Keep cross-package API and generated-file changes in the same review.

## Commands and validation

- Use package filters for development and focused tests; reserve full-workspace checks for final validation.
- Validate dependants when a shared contract, schema, or public package API changes.
- Never create a nested lockfile or add a second workspace tool.

## Handoff

- List affected packages, cross-package effects, generated artifacts, checks run, and migration risk.
