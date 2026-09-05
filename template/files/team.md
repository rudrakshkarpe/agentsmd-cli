# AGENTS.md

## Development environment

- Use the toolchain and package manager already pinned by the repository.
- Document the exact setup, development, test, lint, type-check, and build commands here.

## Team workflow

- Keep changes scoped to one reviewable outcome.
- Preserve existing conventions and explain public API changes.
- Run the narrowest relevant checks before the full suite.
- Never commit secrets, credentials, or generated local state.

## Testing and review

- Add tests beside changed behavior, including failure paths and edge cases.
- Keep dependency and lockfile changes in the same review and explain why they are needed.
- In the handoff, include behavior changed, commands run, risks, and follow-up work.

## Learned rules

- Promote a rule only when a real session supplies evidence that it prevents recurrence.
