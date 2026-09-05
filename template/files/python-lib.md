# AGENTS.md

## Environment

- Use the repository virtual environment for every Python command.
- Prefer the lockfile-backed tool already present (`uv`, Poetry, or pip); do not mix them.
- Install the project in editable mode before running tests when the project supports it.

## Verification

- Run focused tests first, then the full suite.
- Run the configured formatter, linter, and type checker before handoff.
- Test the lowest supported Python version when compatibility-sensitive code changes.

## Library conventions

- Preserve public import paths, type annotations, and documented exceptions.
- Treat changes to exported symbols as API changes and update docs or release notes when required.
- Avoid adding runtime dependencies for behavior available in the supported standard library.

## Lessons

- Record only repository-specific rules backed by observed failures.
