# How this `AGENTS.md` learned

This directory records the two proposed rules used in the learned condition. They are not invented answers to the coding task; each proposal cites a concrete baseline run in [`../results/study-v1`](../results/study-v1).

## Evidence → proposal → gate

| Evidence from the captured trajectory | Reflection output | Evaluation decision |
|---|---|---|
| `baseline-1` first ran the full suite with Go's user cache, received `operation not permitted`, then repeated it with another cache | [`reflection-baseline-1.json`](reflection-baseline-1.json) proposes a workspace-local `GOCACHE` rule | Keep: all three learned runs used it on their first test command; the held-out suite passed |
| `baseline-2` opened `legacy.go` even though the executable calls `config.Load` | [`reflection-baseline-2.json`](reflection-baseline-2.json) proposes a narrow navigation rule | Keep for this study: learned runs did not open the file contents; the held-out suite passed |

The accepted targeted delta is visible by comparing [`../guidance/baseline.md`](../guidance/baseline.md) with [`../guidance/learned.md`](../guidance/learned.md). The reflections were also replayed through the real CLI. [`promoted/AGENTS.md`](promoted/AGENTS.md) is the resulting file, [`promoted/ledger.json`](promoted/ledger.json) retains rule-to-run provenance, and [`promoted/versions.jsonl`](promoted/versions.jsonl) records the two typed promotion commits.

The replay used the public workflow, not a direct edit to the learned block:

```bash
agentsmd learn --task empty-environment-override --run baseline-1 --rule "<cache rule>"
agentsmd pending
agentsmd promote <proposal-id>

agentsmd learn --task empty-environment-override --run baseline-2 --rule "<routing rule>"
agentsmd pending
agentsmd promote <proposal-id>
```

```text
captured run
    ↓
one task-boundary reflection
    ↓
pending rule with run + task provenance
    ↓
same task, fresh workspace, hidden verifier
    ↓
promote only after the gate passes
```

The benchmark compares the two-rule bundle, so it cannot assign the measured token or duration change to either rule individually. A larger claim requires single-rule ablations, more tasks, models, and trials.
