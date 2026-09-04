# AgentCon Japan demo plan

Talk: “Let AGENTS.md Write Itself: Self-Improving Coding Agents, No RL Required!”  
Slot: September 10, 2026, 10:40–11:05 JST, Hall C.

## Five-minute product story

1. Initialize a real repository with `agentsmd init`.
2. Run a repeatable task with no learned rule and record tokens and outcome.
3. Show the normalized trajectory and a task-boundary reflection verdict.
4. Show the proposed rule in `agentsmd pending`; emphasize that the active file has not changed.
5. Promote the reviewed rule and show `agentsmd blame` and typed `agentsmd log` provenance.
6. Repeat the same task and show `agentsmd savings` plus held-out regression status.
7. End with the same `AGENTS.md` being usable by Claude Code, Codex, Cursor, and goose.

## Delivery plan

- September 5: Go library boundaries, usable authoring/versioning core, CI, demo contract.
- September 6: Claude trajectory fixtures, normalizer, task-boundary reflector provider.
- September 7: validation gate, repeatable benchmark task, token accounting, honest before/after result.
- September 8: Codex capture path, cross-platform binaries, failure recovery, security pass.
- September 9: freeze demo behavior, record GIFs, rehearse offline fallback, prepare checksums.
- September 10: use the frozen binary and recorded trajectories; do not depend on venue networking.

## Evidence bar

Do not present “we beat hand-written and auto-generated files” until the repository contains reproducible results supporting it. Report model, task, seed count, token definition, success criteria, held-out regressions, and total evaluation cost. A no-change reflection verdict is a successful anti-bloat result, not a failure.

## Offline fallback

Keep a local binary, demo repository, trajectory fixtures, expected pending proposal, before/after snapshots, benchmark JSON, and short GIF for every live step.

