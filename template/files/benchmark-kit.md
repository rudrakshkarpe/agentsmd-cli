# AGENTS.md

## Benchmark protocol

- Record the task identifier, seed, model, token counts, wall time, and final result.
- Keep evaluation tasks separate from held-out regression tasks.
- Compare candidates across multiple seeds before promotion.
- Reject candidates that improve the target task while regressing held-out tasks.

## Lessons

- Add one targeted, attributable rule per accepted learning event.

