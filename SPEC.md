# agentsmd spec (draft 0)

The CLI is one implementation. These three formats are the actual project.
Anyone can implement a capture adapter, a reflector, or a UI against them.
Keep this file implementation-independent.

## 1. Trajectory

Normalized output of one agent session, produced by a capture adapter.
Every CLI logs differently; the adapter's only job is to emit this shape.

```json
{
  "session_id": "string",
  "tool": "claude | codex | cursor | goose | klaatcode",
  "task": "string, optional task id",
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

## 2. Ledger

The source of truth. AGENTS.md is rendered from it.

```json
{
  "rules": [{
    "id": "r000",
    "text": "imperative, repo-specific, one line",
    "origin": { "run": "session_id", "task": "string", "version": "v0003" },
    "cited": 0,
    "born": "YYYY-MM-DD"
  }],
  "runs": { "<task_id>": [tokens_run1, tokens_run2, "..."] }
}
```

- `origin` powers `blame`. `cited` powers `prune` and the savings story.
- Storing rules (not raw text) is what lets the file dedup, prune, and carry
  provenance. A plain text file can do none of that.

## 3. Version

One entry per snapshot. The `reason` is the point: it records HOW the file
improved, which a raw diff history cannot.

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

## Promotion policy (fast to suggest, slow to trust)

1. `learn` proposes a rule to `pending/` immediately.
2. A rule is promoted to the ledger only on evidence: a human accepts it,
   or subsequent runs of the same task show a token/step improvement.
3. Cheap guards at write time: dedup, contradiction check, repo-specificity.
