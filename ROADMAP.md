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

## Phase 4 — offline GEPA bridge
- `optimize` -> gepa.optimize_anything, token cost in the objective.
- Published before/after results.

## Phase 5 — sharing
- `push` / `pull` an AGENTS.md pack (OCI, KitOps-style).
- Community template hub.

## Governance
- Apache-2.0 (matches AGENTS.md, MCP, goose under the AAIF).
- Stable adapter + reflector interfaces so others extend without forking.
- Semantic versioning. Spec is normative; CLI is one implementation.
- Long-term: propose the spec to the Agentic AI Foundation.
