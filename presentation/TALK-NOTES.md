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
5. The accepted rule is rendered as a small, traceable addition without
   rewriting the maintainer-owned instructions.

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
sed -n '1,120p' AGENTS.md

# 2. Show a captured session and its evidence.
agentsmd sessions show baseline-1

# 3. Show the proposed rule before it changes the file.
agentsmd pending

# 4. Promote the reviewed proposal.
agentsmd promote p0001

# 5. Show how the promoted rule points back to the session that produced it.
agentsmd blame
```

## Demo recovery

- Keep a compiled `agentsmd` binary on the presentation laptop.
- Keep the checked-in example trajectory and promoted ledger available locally.
- Use the recorded CLI animations if the live command stalls.
- Use `agentsmd doctor --repair` if an interrupted reflection worker leaves a
  stale processing job.
- Do not modify the checked-in example during rehearsal. Use a disposable demo
  repository for live commands.
