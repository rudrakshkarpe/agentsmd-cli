# AGENTS.md

## Benchmark protocol

- Record the task identifier, seed, model, token counts, wall time, and final result.
- Keep evaluation tasks separate from held-out regression tasks.
- Compare candidates across multiple seeds before promotion.
- Reject candidates that improve the target task while regressing held-out tasks.

## Reproducibility

- Pin task data, harness versions, prompts, and evaluation commands.
- Store raw outcomes separately from summaries; never edit results after a run.
- Keep training/development examples isolated from held-out tasks.

## Promotion gate

- Define the baseline, acceptance threshold, cost ceiling, and regression budget before running candidates.
- Promote only when the same evaluation procedure shows a repeatable gain.
- Record the exact evidence and rejected alternatives for every accepted rule.

## Lessons

- Add one targeted, attributable rule per accepted learning event.
