# AgentCon Japan talk notes

## Readiness assessment

The deck has enough material for the 25-minute slot. The planned content uses
roughly 17 minutes, followed by four minutes for questions. The remaining time
covers transitions, pauses on the community posts, and one short live command.

The strongest story is narrow and auditable:

1. A maintainer starts with a useful static `AGENTS.md`.
2. A coding-agent trajectory reveals an avoidable failure or detour.
3. One task-boundary reflection proposes a targeted rule with provenance.
4. The proposal remains pending until a human or configured evaluation gate
   accepts it.
5. Fresh held-out trials test whether the updated guidance preserves success
   while reducing waste.

Avoid claiming that the current study proves a general improvement across
repositories. It covers one task, one model, and three trials per condition.

## Story beats before the terminal demo

- Slide 3 recreates the introduction as six controlled beats. Advance one beat
  per sentence or use the play and scrub controls.
- Slide 4 reveals one community post per click. Describe these as perspectives,
  not experimental evidence.
- Slide 5 unfolds the August 2025 release, December 2025 AAIF contribution,
  and the dated adoption snapshot.
- Slide 6 shows the manual synchronization gap.
- Slide 9 introduces the agentsmd contribution without implying model training.
- Slide 10 follows one session record through reflection, evaluation, the
  ledger, and a targeted render. Use the visible controls or advance one
  causal beat at a time.

## Five-minute terminal demo

Run from the `agentsmd-cli` repository root.

Slides 12 and 14 contain looping terminal recordings. Press `R` while either
slide is active to restart its GIF from the first frame. Slide 13 exposes the
same lifecycle as speaker-controlled workflow fragments.

```bash
# 1. Show the static starting instructions.
sed -n '1,120p' benchmarks/config-precedence/guidance/baseline.md

# 2. Show the two evidence-linked reflection verdicts.
cat benchmarks/config-precedence/learning/reflection-baseline-1.json
cat benchmarks/config-precedence/learning/reflection-baseline-2.json

# 3. Show the exact targeted change to AGENTS.md.
diff -u \
  benchmarks/config-precedence/guidance/baseline.md \
  benchmarks/config-precedence/guidance/learned.md

# 4. Show the real promoted ledger and typed version history.
cat benchmarks/config-precedence/learning/promoted/ledger.json
cat benchmarks/config-precedence/learning/promoted/versions.jsonl

# 5. Show the checked-in result without relying on venue networking.
cat benchmarks/config-precedence/results/study-v1/report.md
```

If the network is reliable and at least two minutes remain, run one fresh
trial into a new directory:

```bash
agentsmd benchmark \
  --spec benchmarks/config-precedence/spec.json \
  --trials 1 \
  --output benchmarks/config-precedence/results/demo-live
```

## Benchmark language

Use this wording on stage:

> On this configuration task, all six trials passed the same hidden tests. The
> two learned rules reduced median reported tokens from 90,724 to 54,274,
> commands from five to three, and wall time from 27.9 to 23.1 seconds. This is
> a mechanics demonstration, not a universal benchmark claim.

The multi-task suite and single-rule ablation runner are implemented, but no
new multi-task model results have been checked in yet.

## Demo recovery

- Keep a compiled `agentsmd` binary on the presentation laptop.
- Keep the entire `study-v1` directory available locally.
- Use the recorded report if the fresh model call stalls.
- Use `agentsmd doctor --repair` if an interrupted reflection worker leaves a
  stale processing job.
- Do not modify the checked-in study during rehearsal. Write live output to
  `demo-live` or another ignored directory.
