# Commit roadmap

This is a plan for genuine, reviewable commits—not manufactured history. Commit dates must reflect when work is actually completed. Each commit should build, keep tests green, and carry one coherent behavior, test, document, or release concern.

| Commits | Theme | Intended slices |
|---|---|---|
| 001–008 | Product and specification | terminology, invariants, schemas, compatibility policy, ADRs, threat model, benchmark contract, roadmap |
| 009–018 | Go foundation | module, Cobra root, public packages, error model, atomic writes, embedded assets, test helpers, build metadata, docs, parity fixture |
| 019–030 | Project storage | discovery, scaffold, config, ledger load/save, migrations, locking, corruption recovery, paths, permissions, backups, conformance tests |
| 031–042 | Authoring | init, scratch, template list/use, edit, render, lint, import, export, managed blocks, status, authoring integration tests |
| 043–054 | Version control | commit, log, diff, revert, tags, blame, typed reasons, metadata, snapshot integrity, history repair, concurrency, VC integration tests |
| 055–066 | Capture | adapter contract, fixture corpus, Claude hook, Claude normalization, Codex discovery, Codex normalization, Cursor discovery, Cursor normalization, goose export, goose normalization, watcher, staleness finalization |
| 067–078 | Reflection and gate | verdict schema, provider contract, command provider, OpenAI provider, GEPA bridge, prompt contract, no-op verdict, proposal persistence, dedup, validation gate, promotion transaction, rejection audit |
| 079–090 | Evaluation and savings | run recording, token accounting, task sets, seed control, held-out sets, regression rules, Pareto selection, merge, utilization, savings reports, JSON output, benchmark fixtures |
| 091–100 | CLI experience | structured output, colors, quiet mode, completions, help examples, diagnostics, config command, migration command, exit codes, shell integration |
| 101–110 | Distribution and security | cross-builds, checksums, SBOM, signatures, provenance, Homebrew formula, install script, release notes, vulnerability scan, release dry run |
| 111–120 | AgentCon proof | demo repository, baseline trace, learned trace, measured comparison, limitation slide data, failure fallback, GIF recorder, talk script, rehearsal checklist, frozen demo release |

## Commit rules

- Never backdate, pad, or split a change solely to inflate the count.
- A code commit includes its closest tests; a separate test commit is appropriate only when it captures a reproducer before a fix.
- Use conventional prefixes: `feat`, `fix`, `test`, `docs`, `refactor`, `build`, `ci`, `perf`, `security`.
- Keep `main` green. Develop on short-lived branches and merge reviewed milestones.
- Do not push without explicit approval.

