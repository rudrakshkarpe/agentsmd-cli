# AGENTS.md

## Setup
- Install: `pip install -e ".[dev]"`
- Test: `pytest -q`
- Lint: `ruff check .`

## Conventions
- Type-hint all public functions.
- No network calls in unit tests.
- Keep public API changes out of unrelated PRs.

## Build & test protocol
- Run the full test suite before proposing a diff.
- If a test is flaky, quarantine it, do not delete it.

## Lessons
<!-- managed by agentsmd; agents read these first. rules accumulate below. -->
