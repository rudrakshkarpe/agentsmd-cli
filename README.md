# agentsmd

Version-controlled authoring and self-improvement for `AGENTS.md`.

Author the file, track every version and why it changed, and let your coding
agent's own mistakes improve it over time. Works across any CLI that reads
AGENTS.md (Claude Code, Codex, Cursor, goose).

> Skeleton / draft 0. Authoring and version-control commands run today.
> The learning loop is wired but its reflect step is a stub. See `SPEC.md`.

## Install (dev)

```
make bootstrap
.venv/bin/agentsmd --version
```

See `DEVELOPMENT.md` and `CONTRIBUTING.md` for the local workflow and CI gates.

Long-term the real distribution is a single binary (`brew`, `curl | sh`).
The skeleton is stdlib-only Python; the layout ports 1:1 to Go/cobra.

## Quick start

```
agentsmd init                    # pick a template or start from scratch
agentsmd log                     # see the version history
agentsmd commit -m "tighten setup rules"
agentsmd diff                    # last two versions
agentsmd blame                   # which run/task introduced each rule
```

## Commands

Authoring: `init`, `template list|use`, `edit`, `render`, `lint`
Version control: `log`, `diff`, `status`, `commit`, `revert`, `tag`, `blame`
The loop: `watch`, `learn`, `pending`, `promote`, `reject`, `prune`, `optimize`, `savings`

## Layout

```
AGENTS.md            # the rendered artifact agents read
.agentsmd/
  config.yaml        # template, adapters, loop settings  (the manifest)
  ledger.json        # rules with provenance + cited counts
  versions/          # snapshots + versions.jsonl index
  pending/           # proposed rules awaiting promotion
  runs/              # captured trajectories, token logs
```

## Design notes

- **Rules, not text.** AGENTS.md is rendered from a rule ledger, so it can
  dedup, prune, and carry per-rule provenance.
- **Fast to suggest, slow to trust.** A learned rule is proposed instantly but
  promoted only on evidence or human approval.
- **Capture differs per CLI; the write target is universal.** Every tool reads
  AGENTS.md, so one improvement transfers to all of them.
- **Version = snapshot + reason.** History records how the file got better,
  not just what changed.

## Relationship to GEPA

`optimize` runs an offline batch (GEPA) to prove the mechanism and produce
citable numbers. The live `learn` loop is the online, per-session front end.
They share the reflection step.

## License

Apache-2.0. (Add the full LICENSE text before first release.)
