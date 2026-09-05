# Config-precedence study

This is a small, reproducible coding-agent study—not a universal benchmark score. It demonstrates the complete agentsmd claim on a concrete Go bug: capture a baseline trajectory, derive targeted repository instructions, run fresh sessions with and without those instructions, and grade both conditions with tests the agent cannot see while editing.

## The task

`envmerge` loads a region from `config.json`, then applies `APP_REGION`. Its loader mistakenly treats an explicitly empty environment variable as if it were unset. The agent must preserve unset behavior, support an explicit empty override, add regression coverage, and keep the CLI unchanged.

Each trial receives:

- the same fixture and task prompt;
- a fresh Git repository and isolated workspace;
- either the [static baseline](guidance/baseline.md) or [learned guidance](guidance/learned.md);
- the same agent and model configuration;
- a held-out test copied in only after the agent exits.

The deterministic verifier is `go test ./...`. The expected minimal fix is also stored as [`oracle.patch`](oracle.patch) for audit; it is never exposed to the agent.

## What the loop learned

The baseline traces exposed two sources of waste: Go tried to use a cache outside the sandbox, and one run inspected a compatibility-only implementation outside the executable call path. The [learning record](learning/README.md) connects each proposed rule to its source trajectory, preserves the actual promoted ledger and `AGENTS.md`, and explains the validation decision.

## Recorded result

The checked-in [`study-v1/report.md`](results/study-v1/report.md) contains six real Codex runs: three per condition with `gpt-5.6-luna` at low reasoning effort.

| Condition | Held-out pass rate | Median reported tokens | Median commands | Median duration |
|---|---:|---:|---:|---:|
| Static baseline | 3/3 | 90,724 | 5 | 27.9 s |
| Learned rules | 3/3 | 54,274 | 3 | 23.1 s |
| Change | unchanged | −40.2% | −40.0% | −17.0% |

“Reported tokens” means input plus output tokens from the agent's JSONL usage event; cached input is retained separately in `report.json`. The result shows that the learned bundle preserved success and used fewer resources on this task. With only one task and three trials per condition, it does not establish a general improvement across repositories or agents.

## Reproduce it

Build the current CLI, then run the study from the repository root:

```bash
go build -o ./bin/agentsmd ./cmd/agentsmd

./bin/agentsmd benchmark \
  --spec benchmarks/config-precedence/spec.json \
  --trials 3 \
  --output benchmarks/config-precedence/results/my-run
```

The default runner invokes a clean, non-interactive Codex session with JSONL output, a workspace-write sandbox, ignored user configuration, and the model recorded in the spec. A different compatible agent can be supplied with `--agent-command`; it must accept the task on standard input and emit Codex-compatible JSONL usage and command events if token and command metrics are required.

Each run retains `agent.jsonl`, a normalized `trajectory.json`, verifier output, and the final workspace. Reports are checkpointed after every trial so interrupted studies retain completed evidence.

## Methodology references

- [Terminal-Bench](https://github.com/harbor-framework/terminal-bench) motivates the task/environment/verifier structure for terminal agents.
- [SWE-bench](https://github.com/SWE-bench/SWE-bench) motivates repository-level issues graded by executable tests.
- [GEPA](https://arxiv.org/abs/2507.19457) motivates reflective instruction evolution from execution feedback.
- [ACE](https://arxiv.org/abs/2510.04618) motivates incremental, evidence-linked updates instead of repeatedly rewriting the whole instruction file.

The local case is intentionally much smaller than those suites. Its purpose is to make the mechanics and evidence inspectable before scaling to a multi-task held-out benchmark.
