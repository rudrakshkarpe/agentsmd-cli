# agentsmd — Development Handoff

Everything decided so far, written for a fresh Claude session that will build
this framework. Read this top to bottom before writing code. It is the source
of truth for scope, architecture, and the decisions already made (and why).

There is a runnable Python skeleton that accompanies this document
(`agentsmd-skeleton/`). This handoff explains what it is and what to build next.

---

## 0. What this project is

A **version-controlled authoring and self-improvement tool for `AGENTS.md`**.

Three capabilities, in priority order:
1. **Author** an AGENTS.md (from a template or scratch), via a verb-driven CLI.
2. **Version-control** it like git, where each version records *why* it changed
   and what it saved.
3. **Self-improve** it: capture what a coding agent learned from its own
   mistakes in a session, and turn that into a rule in AGENTS.md so the next
   run is cheaper (fewer tokens, fewer dead paths).

The unique selling point is #3: **the layer figures out what the agent learned
from a mistake and saves tokens next time by improving AGENTS.md itself.**

It is a long-term open-source project (Apache-2.0), designed to eventually
align with the Agentic AI Foundation (AAIF), which already stewards AGENTS.md,
MCP, and goose.

---

## 1. Background context (why this exists)

- **AGENTS.md** is an open standard: a markdown file at a repo root telling any
  coding agent how to work in that repo. Released Aug 2025 out of OpenAI Codex
  tooling; adopted by 60,000+ repos and 20+ tools; donated Dec 9, 2025 to the
  AAIF under the Linux Foundation alongside Anthropic's MCP and Block's goose.
- The problem: AGENTS.md is **static**. Written once, by hand. It drifts as code
  moves, bloats as rules accumulate, points at dead paths, and never learns from
  the agent's repeated failures. Auto-generating it once doesn't fix this — it
  produces a file that *describes* the repo but doesn't encode what the agent
  gets *wrong* in it. (Two named failure modes from the ACE paper: **brevity
  bias** — detail dropped for concise summaries; **context collapse** — detail
  eroded by repeated rewriting.)
- The gap: every other layer of the agent stack has an eval loop. The
  instruction layer (AGENTS.md) has none. This tool is that loop.

Key references the receiving session should read:
- **GEPA** (arXiv 2507.19457) — reflective prompt evolution; beats RL (GRPO) by
  6–20% using up to 35x fewer rollouts; core mechanism is reflective mutation +
  Pareto selection. Has a real library: `pip install gepa`, with an
  `optimize_anything(seed_candidate, evaluator, objective)` API (released Feb
  2026) that optimizes any text artifact against any metric.
- **ACE** (arXiv 2510.04618) — evolving contexts as "playbooks" via
  generation / reflection / curation with incremental delta updates; the online,
  natural-execution-feedback framing. This is the shape of our live loop.

---

## 2. The core mechanism (how "learns and saves tokens" actually works)

An agent wastes tokens the same way twice: it explores a wrong path, backs out,
does the right thing. That exploration costs tokens. Next run, same repo, same
wrong path, same cost.

The layer writes the lesson from that wrong path as **one rule** into AGENTS.md.
Next run, the rule is already in context, so the agent skips the exploration.
**Tokens saved = the exploration that no longer happens**, minus the small
standing cost of the rule sitting in context.

This gives the honest nuance that must be stated, not hidden: a rule only pays
off if it fires more than it costs. Hence `cited` counting and `prune`. The
principle:

> **Fast to suggest, slow to trust.** Propose a rule instantly; promote it to
> the active file only on evidence (human approval, or measured improvement on
> later runs of the same task).

---

## 3. The loop (the engine)

Six stages: **run → reflect → mutate → validate → select → commit.**

- **run** — execute a task with the current AGENTS.md against an agent CLI.
- **reflect** — read the trajectory; attribute the failure. Emit one of four
  verdicts: missing rule / wrong rule / stale rule / **not-an-AGENTS.md-problem**.
  The fourth verdict is what keeps the file from bloating — the reflector must be
  allowed to propose *nothing*.
- **mutate** — one targeted delta, never a whole-file rewrite (rewrites cause
  context collapse). Each rule carries provenance.
- **validate** — an eval gate: net gain on the task set, zero regressions on a
  held-out set, N seeds (not one) to absorb non-determinism, token cost within
  budget. Fail any → discard.
- **select** — Pareto pool, not greedy. A candidate strong on build tasks and one
  strong on test tasks both survive and merge, instead of one killing the other.
- **commit** — output is a **pull request, not a file**: a small diff with a
  linked eval result that a human reviews.

Two shapes of this loop, and they compose:
- **Offline (GEPA):** batch, needs a task set + rollout budget; run once to prove
  the mechanism and produce citable numbers. Use `gepa.optimize_anything` with
  AGENTS.md as the seed candidate and **token cost inside the objective** so the
  USP is literally what's being optimized. Cost warning: each metric call is one
  full agent run; their example uses `max_metric_calls=100`. For demos use 15–25.
- **Online (this tool's live layer):** ACE-shaped, one reflect step per real
  task at the task boundary, no curated task set. This is what "drop it on the
  repo you already have" means. Reuse GEPA's reflection component per-session.

---

## 4. The one design decision everything hangs on

**Do not reflect mid-task. Reflect at the task boundary.** The session that just
failed is contaminated as its own evaluator, and streaming reflection is
expensive. Fire once when a task/session ends.

**Capture is per-CLI. The write target (AGENTS.md) is universal**, because every
CLI reads it. That asymmetry IS the architecture: four different capture paths,
one shared everything-else.

---

## 5. Capture surface per CLI (verified, current as of Sep 2026)

| CLI | Mechanism | How to capture |
|---|---|---|
| **Claude Code** | Rich lifecycle hooks (PreToolUse, PostToolUse, Stop, SubagentStop, SessionEnd). Every event passes JSON on stdin with `session_id` and `transcript_path`. | Register a `Stop` hook → read the transcript JSONL at `transcript_path` → normalize. Event-driven, low latency. |
| **Codex CLI** | Reads AGENTS.md. Writes transcripts as JSONL (`~/.codex/transcripts/*.jsonl`). | Tail/watch the transcript dir. |
| **Cursor** | Reads AGENTS.md. Writes transcripts as JSONL (`~/.cursor/*/transcript.jsonl`). | Tail/watch the transcript dir. |
| **goose** | **No hook/lifecycle system.** Extensions are MCP servers. Sessions persist in SQLite (`~/.config/goose/sessions/*.db`) and can be exported (`goose session export --format json`). | Read the session store / export after the fact. Ship the write side as an MCP extension. |

Two unified capture strategies: **hook where one exists** (Claude Code), a
**transcript-tailing daemon everywhere else** (staleness detection to finalize
idle sessions).

**Prior art to be aware of:** the *CEMS Observer Daemon* already does multi-tool
session watching (Claude Code, Cursor, Codex, goose) via per-tool adapters +
signal hooks + staleness detection. Our differentiator: it ships transcripts to
a cloud server for memory storage; **we stay local and write back into AGENTS.md
itself**, with a token-cost objective, GEPA-style reflection, and rule
provenance/pruning. Don't reinvent the watcher; do differentiate the payoff.

---

## 6. The data model (the vendor-neutral core — this is the real project)

The CLI is one implementation. These three schemas are normative. Publish them
as `SPEC.md`; they are the piece that could eventually go to the AAIF.

### 6.1 Trajectory — normalized output of one session (adapter produces this)
```json
{
  "session_id": "string",
  "tool": "claude | codex | cursor | goose",
  "task": "optional task id",
  "steps": [{ "role": "assistant|tool", "summary": "string" }],
  "tool_calls": [{ "name": "string", "args": {}, "result": "string" }],
  "files_touched": [{ "path": "string", "diff": "unified diff" }],
  "commands": [{ "argv": ["..."], "exit_code": 0 }],
  "test_results": { "passed": 0, "failed": 0, "errors": 0 },
  "tokens": { "in": 0, "out": 0, "cached": 0 },
  "wall_time_s": 0.0,
  "final_diff": "unified diff of the artifact under review"
}
```
The adapter's only job is to emit this shape. Trace normalization is the
unglamorous bulk of the engineering — every CLI logs differently.

### 6.2 Ledger — source of truth; AGENTS.md is rendered from it
```json
{
  "rules": [{
    "id": "r000",
    "text": "imperative, repo-specific, one line",
    "origin": { "run": "session_id", "task": "string", "version": "v0003" },
    "cited": 0,
    "born": "YYYY-MM-DD"
  }],
  "runs": { "<task_id>": [tokens_run1, tokens_run2] }
}
```
**Why store rules, not raw text:** it's the single decision that makes deltas
into list ops, gives provenance a stable home (rule id, not a line number that
moves), makes hit-counting possible, and turns Pareto merging into set union.
A plain text file can do none of this.

### 6.3 Version — one entry per snapshot
```json
{
  "id": "v0000",
  "parent": "v-prev or null",
  "ts": "ISO-8601",
  "reason": "learned | manual | template",
  "message": "string",
  "meta": { "task": "011", "token_delta": -6860, "eval": "12/12 held-out pass" }
}
```
The `reason` + `meta` is the point: history shows **how** the file got better,
which a raw diff log cannot.

---

## 7. CLI command surface

Three families. Authoring borrows KitOps verb discipline; version control
borrows git semantics; the loop is the USP.

**Authoring**
- `agentsmd init` — create AGENTS.md + `.agentsmd/`; interactive template-or-scratch
- `agentsmd template list | use <name>`
- `agentsmd edit` — open in `$EDITOR`
- `agentsmd render` — rebuild AGENTS.md from the ledger
- `agentsmd lint` — duplicate rules, dead rules, bloat

**Version control**
- `agentsmd log` — history with reason + token delta per version
- `agentsmd diff [a] [b]`
- `agentsmd status` — pending rules + uncommitted edits
- `agentsmd commit -m`
- `agentsmd revert <v>`
- `agentsmd tag <v> <name>`
- `agentsmd blame` — per-rule provenance (which run/task introduced each rule).
  This is git-blame mapped onto rules; it's a signature feature.

**The loop**
- `agentsmd watch` — start the capture daemon
- `agentsmd learn [--adapter claude]` — reflect on last session → propose to `pending/`
- `agentsmd pending`
- `agentsmd promote <id>` / `agentsmd reject <id>` — the human gate
- `agentsmd prune` — retire rules that never fire
- `agentsmd optimize` — offline GEPA batch
- `agentsmd savings <task>` — token-savings report (the USP number)

---

## 8. Directory & code layout

Runtime layout in a user's repo (mirrors git: artifact at root, machinery hidden):
```
repo/
  AGENTS.md            # the rendered artifact agents read
  .agentsmd/
    config.yaml        # template, adapters, loop settings (the "manifest")
    ledger.json        # rules + provenance + cited + per-task token runs
    versions/          # v0000.md ... + versions.jsonl index + tags.json
    pending/           # proposed rules awaiting promotion
    runs/              # captured trajectories, token logs, gate results
```

Codebase layout (skeleton is Python; port to Go/cobra when surface freezes):
```
agentsmd/
  cmd/agentsmd/       # CLI entrypoint            (skeleton: agentsmd/cli.py)
  core/               # ledger, render, dedup     (agentsmd/core.py)
  vc/                 # version control layer     (agentsmd/vc.py)
  capture/            # adapters + daemon         (agentsmd/adapters/)
  reflect/            # reflection + GEPA bridge   (agentsmd/loop.py::_reflect stub)
  gate/               # dedup, validation, promote (agentsmd/loop.py + core.find_duplicate)
  registry/           # push/pull, template hub   (not in skeleton)
  schema/             # the normative schemas     (SPEC.md)
  templates/          # built-in AGENTS.md starts  (templates/*.md)
```

---

## 9. What the skeleton already does vs what to build

**Runs today (tested):** `init` (template/scratch), `template list/use`, `commit`,
`log` (with typed reasons), `diff`, `revert`, `tag`, `blame`, `render`, `lint`,
and the full `promote` path (pending → dedup → ledger → render → version with
reason=`learned`).

**Stubbed with clear TODO markers:**
- `agentsmd/loop.py::_reflect` — returns None. Replace with a real LLM/GEPA
  reflection call that takes a normalized trajectory and returns one rule (or
  None for the not-an-AGENTS.md-problem verdict).
- `agentsmd/adapters/claude.py::latest_trajectory` — parse the JSONL at
  `transcript_path` into the trajectory schema.
- `watch` (daemon) and `optimize` (GEPA bridge) — printed placeholders.

**Highest-value next step (Phase 2):** make `_reflect` and the Claude adapter
real so `learn → promote → savings` produces a live token-drop from an actual
Claude Code session. That single path is the demo and the USP.

---

## 10. Build decisions already made (do not silently reverse)

1. **Rules-as-ledger, render the file.** Not: edit AGENTS.md text directly.
2. **Version = snapshot + typed reason + meta.** Skeleton uses a self-contained
   snapshot store (not git) because the reason metadata is the value and it's
   dependency-free. Production *may* back it with libgit2; keep the model.
3. **Fast to suggest, slow to trust.** `pending/` → `promote` gate is mandatory;
   never auto-write learned rules straight into the active file.
4. **Reflect at task boundary, not mid-task.** In-session validation is
   contaminated.
5. **One thin adapter per CLI; everything downstream written once.** Adding a CLI
   = implementing the `Adapter` interface, nothing else.
6. **Deltas, never rewrites.** Whole-file rewrites reintroduce context collapse.
7. **The reflector may propose nothing.** The anti-bloat verdict is a feature.
8. **GEPA offline for proof, live loop for adoption.** Both, not either.
9. **Apache-2.0**, spec-first, spec is normative and CLI is one implementation.
10. **Python skeleton, Go/cobra for the shipped single binary** (brew / curl|sh).
    Layout is 1:1 so the port is mechanical.

---

## 11. Open items (decisions still needed)

- **Project name.** `agentsmd` is fine as the binary, weak as a project name.
  Needed before repo, brew formula, and talk copy are finalized.
- **Git-backed vs self-contained version store** for production (skeleton is
  self-contained; wrapping real git is a legitimate alternative).
- **Reflection model** for `_reflect`: direct LLM call vs GEPA's reflection
  component. Start with whichever is faster to get the live demo working.
- **Real numbers** for the offline results (task success, steps, token cost,
  regressions, rule utilization) — pending a real benchmark run.

---

## 12. Distribution (long-term)

There is **no universal plugin format** across these CLIs. The single binary is
the product; per-CLI glue is thin:
- Binary: Homebrew tap, `curl -fsSL … | sh`, plus cargo/npm.
- Claude Code: one-line `Stop` hook in `.claude/settings.json`, optionally a
  Claude Code plugin.
- goose: an MCP extension for the write/query side.
- Codex / Cursor: nothing installed in the tool; point the daemon at their
  transcript dirs.
- Universal zero-install fallback: an AGENTS.md "self-improvement protocol"
  block (prose the agent follows to edit AGENTS.md itself). Works day one in any
  tool; the binary just makes it reliable and measurable. The tool that improves
  AGENTS.md ships partly *as* AGENTS.md — it bootstraps itself.

---

## 13. Talk context (for tone/priorities only)

This framework is the subject of a 25-minute talk at AgentCon Japan (Sep 10,
2026, Hall C) titled "Let AGENTS.md Write Itself: Self-Improving Coding Agents,
No RL Required" (Rudraksh Karpe, Simplismart; Satyam Soni, NitroStack). The talk
demos the loop via GIFs (no live terminal). Priorities that follow from this:
the token-savings path must be real and recordable, and the honest limitations
(eval gaming, overfitting, non-determinism, cost, attribution error) are stated,
not hidden.
