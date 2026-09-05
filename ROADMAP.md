# Roadmap

Phased, not dated. Each phase is independently useful and shippable.

## Phase 0 — the spec
- Publish `SPEC.md`: trajectory, ledger, version formats.
- Implementation-independent. This is the piece that could go to the AAIF.

## Phase 1 — authoring + versioning
- `init`, `template`, `edit`, `render`, `lint`.
- `log`, `diff`, `status`, `commit`, `revert`, `tag`, `blame`.
- No LLM required. Useful and adoptable on day one.

## Phase 2 — the loop on one adapter
- Claude Code capture via `Stop` hook + `transcript_path`.
- `learn`, `pending`, `promote`, `reject`, `savings`.
- The token-savings USP, demoable.

## Phase 3 — multi-adapter capture + daemon
- Project-local SessionEnd connectors: Codex, Claude Code, Cursor, and goose. Complete.
- Rich provider-specific transcript normalization beyond Claude Code.
- `watch` daemon with staleness detection.

## Phase 4 — cross-CLI session tracking
- [x] Give every run a provider-qualified ID.
- [ ] Correlate related runs under a stable logical task ID.
- [x] Capture session-start and session-end lifecycle events without duplicating completed reflection jobs.
- [x] Record timestamps, working tree revisions, changed files, final diffs, evaluation outcomes, wall time, model, provider, and available tokens.
- [x] Provide local `sessions` and `sessions show` inspection.
- [ ] Add task-level `progress` comparisons.
- [x] Queue reflection outside latency-sensitive hooks with durable job results.
- [ ] Add explicit recovery for interrupted/stale worker locks.
- [x] Require an evaluation command and confidence threshold for opt-in automatic promotion.
- [ ] Compare equivalent tasks before and after a promoted rule using success rate, regressions, tokens, and completion time.
- [ ] Keep raw transcripts local by default and support configurable redaction before any external reflector receives a trajectory.

Definition of done: a Codex, Claude Code, Cursor, or goose task can be traced from capture to proposal, promotion, and a later measured outcome without manual bookkeeping. Promotion remains gated by a human or evaluation policy.

## Phase 5 — offline GEPA bridge
- `optimize` -> gepa.optimize_anything, token cost in the objective.
- Published before/after results.

## Phase 6 — sharing
- `push` / `pull` an AGENTS.md pack (OCI, KitOps-style).
- Community template hub.

## Governance
- Apache-2.0 (matches AGENTS.md, MCP, goose under the AAIF).
- Stable adapter + reflector interfaces so others extend without forking.
- Semantic versioning. Spec is normative; CLI is one implementation.
- Long-term: propose the spec to the Agentic AI Foundation.
