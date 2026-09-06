# KlaatCode interactive lifecycle evidence

Captured 2026-09-06 UTC (2026-09-07 in Asia/Kolkata), macOS arm64.
KlaatCode source: `062c4ac89a5a6de38f87cb402f3ba63b959e73e8` (package 2.5.0).
Agentsmd: PR #18 with the permission/Doctor review fixes.

This is an actual interactive TUI lifecycle check, driven through a PTY, **not a
model task or a cross-harness benchmark**. No real API credentials were used.
The API base URL was `http://127.0.0.1:1`, with a dummy credential; update checks
and telemetry were disabled. The TUI's supported `!` shell input changed the
fixture, and `/exit` closed the session normally. The reflector is a deterministic
test stub, so its proposal is pipeline evidence, not a measured improvement in
rule quality.

## Observations

- Exactly one `session_start` and one `session_end` hook payload, with matching
  session IDs and project roots.
- One normalized trajectory, with a positive duration and the expected
  `answer.txt` diff (`before` to `after`).
- One completed reflection queue job, one successful evaluator result, and a
  pending proposal. Automatic promotion was disabled.
- Empty model steps and tool calls are expected: these are absent from the
  lifecycle payload. There is no model metadata or provider outcome. The
  normalized `outcome: success` and test counts come from the configured fixture
  evaluator, not from KlaatCode. Schema-default zero token counts mean unavailable
  usage here, not measured zero consumption.

The JSON/JSONL files preserve payloads and generated records, with two portability
substitutions: the absolute project path is `$FIXTURE` and the local Python
executable is `python3`. Session IDs, timestamps, Git IDs, diffs, and evaluation
results are unchanged. Terminal frames are not needed for the capture contract
and are not included.

## Reproduce

1. Build agentsmd (`go build -o /absolute/path/agentsmd ./cmd/agentsmd`). Check out
   the pinned KlaatCode commit and build with `bun run build` after installing
   dependencies. At this commit `npm ci` fails because its lockfile lacks the
   two tree-sitter dependencies; this run used
   `npm install --ignore-scripts --package-lock=false --no-audit --no-fund`
   without modifying the upstream lockfile. Thus the source is pinned, but the
   full transitive dependency installation is not reproducibly locked.
2. Create a fresh Git fixture with tracked `answer.txt` containing `before\n`;
   initialize agentsmd and run `agentsmd connect klaatcode`.
3. Configure a reflector that reads the trajectory from stdin and prints:

   ```json
   {"verdict":"missing_rule","rule":"Check answer.txt with the fixture evaluator before finishing.","confidence":0.95}
   ```

   Configure an evaluator that asserts `answer.txt` contains `after\n` and
   prints `fixture passed`. Leave automatic promotion off.
4. For evidence capture only, wrap each installed lifecycle command with a
   program that reads stdin, appends it as one line to `raw-hooks.jsonl`, then
   passes the identical stdin to `agentsmd hook klaatcode`. Ignore `.agentsmd/`,
   `.klaatai/`, and `raw-hooks.jsonl` in Git. Commit the fixture baseline.
5. Launch the pinned build in a real terminal:

   ```sh
   KLAATAI_API_KEY=lifecycle-test-no-credentials KLAATAI_TELEMETRY=0 \
     bun /absolute/path/klaatcode/dist/klaatai.js \
     --no-update-check --base-url http://127.0.0.1:1
   ```

   Enter `!printf "after\n" > answer.txt`, then `/exit`. Wait for the detached
   worker to finish before inspecting `.agentsmd/runs`, `queue`, `evaluations`,
   and `pending`.

## Outstanding review gates

The requested same-task benchmark across KlaatCode and the other harnesses has
not been run. This fixture does not establish agent task success, cross-harness
regressions, token efficiency, or learned-rule quality. Goose is not installed
in this environment; the available `agent` executable is a different CLI, not a
verified Cursor harness. Those environments and a shared model-task protocol
still need to be provisioned for the integration batch.

Upstream GitHub Actions run `33988545477` reports `action_required`; a maintainer
must approve the fork workflow before the required repository matrix can run.
Local `make check` passes, but is not a substitute for that matrix.
