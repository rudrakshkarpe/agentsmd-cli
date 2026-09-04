# AGENTS.md

## Setup
- Run all sweeps from the repo root, never from subdirectories.
- Config lives in `sweep_config.yaml`; edit via the provided make target, not by hand.

## Output contract
- Every run must validate against the JSON schema before writing a summary.
- Derived metrics (RTF, QPS) are computed, never hand-entered.

## Self-improvement protocol
- After each task, if you hit a wrong path, propose ONE rule that would have avoided it.
- Do not duplicate an existing lesson. Prune lessons that never fire.

## Lessons
<!-- managed by agentsmd; agents read these first. rules accumulate below. -->
