<div align="center">

# agentsmd

### Let `AGENTS.md` learn from the work your coding agent just did.

Author, version, measure, and improve repository instructions from real agent trajectories—without fine-tuning or framework lock-in.

[![CI](https://img.shields.io/github/actions/workflow/status/rudrakshkarpe/agentsmd-cli/ci.yml?branch=main&style=for-the-badge&logo=github&label=CI)](https://github.com/rudrakshkarpe/agentsmd-cli/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/rudrakshkarpe/agentsmd-cli?style=for-the-badge&logo=go)](go.mod)
[![License](https://img.shields.io/github/license/rudrakshkarpe/agentsmd-cli?style=for-the-badge)](LICENSE)

[Install](#install) · [Quick start](#quick-start) · [How it works](#how-it-works) · [CLI](#cli) · [Go library](#go-library) · [Roadmap](#roadmap)

</div>

> **Active development.** Project detection, cross-CLI lifecycle capture, automatic reflection queues, command-based evaluation gates, opt-in automatic promotion, and macOS/Linux releases work today. Held-out evaluation and offline optimization are still being built for the AgentCon Japan demo.

## Why agentsmd?

`AGENTS.md` tells coding agents how to work in a repository. But the file is usually static: it drifts as the code changes, accumulates stale rules, and cannot learn when an agent repeatedly explores the same wrong path.

agentsmd adds an evidence loop around that file:

```text
run a task → capture its trajectory → reflect once at the task boundary
           → propose one rule → review and validate → promote → measure
```

The active file never changes merely because a model suggested something. Every learned rule starts in a pending queue, carries its source run and task, and becomes active only after review or evaluation.

## CLI showcase

The write target is universal: all four tools read `AGENTS.md`. Capture is implemented separately because every CLI records sessions differently.

<table>
<tr>
<td align="center" width="25%">
<a href="https://claude.com/product/claude-code"><img src="https://github.com/anthropics.png?size=120" alt="Claude Code" width="48" height="48" /></a><br/>
<strong>Claude Code</strong><br/>
<sub>Session hook + JSONL normalization</sub>
</td>
<td align="center" width="25%">
<a href="https://github.com/openai/codex"><img src="https://github.com/openai.png?size=120" alt="Codex CLI" width="48" height="48" /></a><br/>
<strong>Codex CLI</strong><br/>
<sub>Project SessionEnd hook</sub>
</td>
<td align="center" width="25%">
<a href="https://cursor.com"><picture><source media="(prefers-color-scheme: dark)" srcset="https://svgl.app/library/cursor_dark.svg"><img src="https://svgl.app/library/cursor_light.svg" alt="Cursor" width="48" height="48" /></picture></a><br/>
<strong>Cursor</strong><br/>
<sub>Project sessionEnd hook</sub>
</td>
<td align="center" width="25%">
<a href="https://github.com/block/goose"><img src="https://github.com/block.png?size=120" alt="goose" width="48" height="48" /></a><br/>
<strong>goose</strong><br/>
<sub>Project hook plugin</sub>
</td>
</tr>
</table>

## Install

### Shell installer (macOS and Linux)

```bash
curl -fsSL https://rudrakshkarpe.com/install.sh | sh
```

The installer selects Apple Silicon, Intel macOS, or Linux automatically, verifies the release archive's SHA-256 checksum, and installs to `~/.local/bin`. Choose another directory when needed:

```bash
curl -fsSL https://rudrakshkarpe.com/install.sh | \
  AGENTSMD_INSTALL_DIR=/usr/local/bin sh
```

Once installed from the shell release, update in place with the same checksum verification:

```bash
agentsmd update
```

Use `agentsmd update --check` to check without installing. `agentsmd upgrade` is an alias. Package-manager symlinks are deliberately not overwritten; update those installations through their package manager.

If GitHub's release CDN is slow on your route and Go is already installed, bypass the binary download:

```bash
GOBIN="$HOME/.local/bin" go install github.com/rudrakshkarpe/agentsmd-cli/cmd/agentsmd@latest
```

### Go install

Requirements: Go 1.23 or newer.

```bash
go install github.com/rudrakshkarpe/agentsmd-cli/cmd/agentsmd@latest
```

### Build from source

```bash
git clone https://github.com/rudrakshkarpe/agentsmd-cli.git
cd agentsmd-cli
go build -trimpath -o ./bin/agentsmd ./cmd/agentsmd
./bin/agentsmd --version
```

## Quick start

Let agentsmd inspect the repository and create a useful baseline:

```bash
cd your-project
agentsmd init
agentsmd doctor
```

Connect the coding tools you actually use. Each command installs a project-local end-of-session hook:

```bash
agentsmd connect codex
agentsmd connect claude
agentsmd connect cursor
agentsmd connect goose
```

Configure a reflector and the command that must pass before a proposal is eligible for automatic promotion:

```bash
agentsmd automate \
  --reflect-command "./scripts/reflect-agentsmd" \
  --evaluate-command "go test ./..."
```

This automatically reflects after capture but leaves successful proposals pending for review. Automatic promotion is an explicit additional policy:

```bash
agentsmd automate --auto-promote --min-confidence 0.90
```

`--auto-promote` is rejected unless both reflection and evaluation commands are configured. The evaluation command receives `AGENTSMD_PROPOSAL_ID`, `AGENTSMD_RULE`, and `AGENTSMD_RUN_ID` in its environment.

Captured sessions land in `.agentsmd/runs/`. Inspect their provider, outcome, duration, changed files, and test summary with:

```bash
agentsmd sessions
agentsmd sessions show <run-id>
```

A reflector—or a person—can propose one targeted lesson, which stays pending until reviewed:

```bash
agentsmd learn \
  --task parser-regression \
  --run session-baseline \
  --rule "Run the focused parser fixture before the full suite."

agentsmd pending
agentsmd promote <proposal-id>
```

## How it works

### 1. Rules are structured data

Rules live in `.agentsmd/ledger.json`, with stable IDs, provenance, citation counts, and per-task token runs. Only the marker-delimited learned-rules block is generated; project setup, commands, and hand-written guidance are preserved.

### 2. Versions explain why a change exists

Every snapshot includes a typed reason (`manual`, `template`, or `learned`) and metadata such as the originating task, session, evaluation, and token delta.

### 3. Reflection happens at the task boundary

The reflector receives one normalized trajectory and returns exactly one of four verdicts:

- `missing_rule`
- `wrong_rule`
- `stale_rule`
- `not_an_agentsmd_problem`

The fourth verdict is deliberate: sometimes the correct improvement is no new instruction.

### 4. Suggestions are not trusted automatically

Learned changes enter `.agentsmd/pending/`. Promotion is a separate operation so humans or evaluation gates can reject weak, duplicated, overfitted, or costly rules.

## Automatic reflection

agentsmd can delegate reflection to any executable that reads a normalized trajectory as JSON on standard input and writes a verdict as JSON on standard output:

```bash
agentsmd learn \
  --trajectory .agentsmd/runs/session-123.json \
  --reflect-command "./my-reflector"
```

Expected output shape:

```json
{
  "verdict": "missing_rule",
  "rule": "Run the focused parser fixture before the full suite.",
  "confidence": 0.91,
  "origin": { "run": "session-123", "task": "parser-regression" },
  "rationale": "The session spent time debugging the wrong parser path."
}
```

This contract keeps model providers outside the core. Direct providers and a GEPA bridge can implement the same interface.

## CLI connections

`agentsmd connect` configures the supported tools without replacing their existing settings:

- Codex: `.codex/hooks.json`
- Claude Code: `.claude/settings.local.json`
- Cursor: `.cursor/hooks.json`
- goose: `.agents/plugins/agentsmd/`

All connectors capture start/end lifecycle events into a provider-neutral trajectory. The end event is persisted immediately, while Git evidence, transcript parsing, reflection, and evaluation run in a detached local worker. Runs record available start/end times, wall duration, Git revisions, worktree status, changed files, final diff, provider status, evaluation command, test outcome, model, and tokens. Claude Code additionally normalizes its JSONL transcript into assistant steps, tool calls, shell commands, and token usage. Codex transcript internals are intentionally not parsed because their documented format is unstable.

## CLI

| Goal | Command |
|---|---|
| Check for or install the latest release | `agentsmd update --check`, `agentsmd update` |
| Detect the project and create `AGENTS.md` | `agentsmd init` |
| Browse or apply reusable baselines | `agentsmd templates`, `agentsmd templates use NAME` |
| Connect a coding tool | `agentsmd connect codex\|claude\|cursor\|goose` |
| Configure automatic reflection and gating | `agentsmd automate` |
| Diagnose the local setup | `agentsmd doctor` |
| Inspect measured sessions | `agentsmd sessions`, `agentsmd sessions show RUN` |
| Review the improvement queue | `agentsmd pending`, `agentsmd promote ID`, `agentsmd reject ID` |
| Propose a targeted rule | `agentsmd learn ...` |

Legacy authoring, versioning, and measurement commands remain compatible but are hidden from the primary help while their UX is consolidated.

## Repository layout

```text
AGENTS.md                 rendered instructions read by coding agents
.agentsmd/
  config.yaml             project configuration
  connections.json        portable records of configured CLI hooks
  automation.json         reflection, evaluation, and promotion policy
  ledger.json             rules, provenance, citations, token runs
  versions/               typed snapshots, index, and tags
  pending/                proposed rules awaiting review
  runs/                   normalized trajectories and measurements
  sessions/               lifecycle baselines
  inbox/                  durable raw hook events
  queue/                  idempotent background reflection jobs
  evaluations/            gate results and command output
```

## Go library

The CLI is one consumer of reusable packages:

| Package | Responsibility |
|---|---|
| `schema` | Vendor-neutral trajectory, ledger, rule, proposal, and version types |
| `project` | Discovery, scaffolding, atomic persistence, and repository paths |
| `ledger` | Rule identity, deduplication, rendering, and linting |
| `version` | Typed snapshots, history, diff, tags, and revert |
| `capture` | Adapter interface for coding-agent session formats |
| `capture/claude` | Claude Code JSONL normalization |
| `detect` | Conservative stack and command discovery |
| `integration` | Project-local lifecycle hooks for supported coding tools |
| `session` | Lifecycle baselines plus Git, duration, file, and status evidence |
| `automation` | Durable reflection jobs, evaluation records, and gated promotion |
| `reflect` | Reflection verdict contract and command provider |
| `learning` | Propose, review, promote, reject, prune, and measure workflow |
| `cli` | Embeddable Cobra command tree |

`SPEC.md` is normative; Go packages should remain compatible with its schemas.

## Safety and trust model

- Local-first storage; no transcript upload is required by the core.
- Learned rules never bypass the pending-review gate.
- Automatic promotion is disabled by default and cannot be enabled without an evaluation command.
- Whole-file reflective rewrites are avoided; learning produces targeted deltas.
- Version and proposal identifiers reject path traversal.
- CI tests Go 1.23 and 1.25 on Linux, macOS, and Windows and runs the race detector.
- Token savings are reported only from recorded runs of the same task.

## Roadmap

- [x] Go library and Cobra CLI
- [x] Authoring, typed versions, blame, and rule ledger
- [x] Pending-rule review and deterministic demo path
- [x] Claude Code trajectory normalization
- [x] Provider-neutral task-boundary reflector
- [ ] Validation gate with held-out regression checks
- [x] Codex, Claude Code, Cursor, and goose lifecycle connectors
- [ ] Rich transcript normalization beyond Claude Code
- [ ] Logical-task identity across related sessions
- [x] Provider-qualified session identity and automatic Git/duration evidence capture
- [x] Configurable command/test outcome capture through the evaluation gate
- [x] Local session listing and complete run inspection
- [x] Idempotent background reflection queue
- [x] Opt-in automatic promotion behind evaluation and confidence gates
- [ ] Logical-task correlation and before/after progress comparisons
- [ ] Watch daemon with session staleness detection
- [ ] Offline GEPA optimization bridge
- [ ] Reproducible benchmark report and token-savings evidence
- [x] Checksummed macOS and Linux release archives with a shell installer
- [ ] Signed releases and Homebrew installation

See the [development roadmap](ROADMAP.md), [120-slice commit plan](docs/COMMIT-ROADMAP.md), and [AgentCon demo plan](docs/AGENTCON-DEMO.md).

## AgentCon Japan

agentsmd is being developed for **“Let AGENTS.md Write Itself: Self-Improving Coding Agents, No RL Required!”**, presented by Rudraksh Karpe and Satyam Soni at AgentCon Japan on September 10, 2026.

The demo will use recorded trajectories and reproducible task runs. Claims about task success, regressions, or token savings will be published only with the model, task, seeds, success criteria, and evaluation cost needed to reproduce them.

## Development

```bash
make bootstrap
make check
```

Read [DEVELOPMENT.md](DEVELOPMENT.md) and [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. `main` is protected and changes land through green CI.

## License

Apache License 2.0.
