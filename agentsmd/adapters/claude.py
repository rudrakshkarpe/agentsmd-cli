"""Claude Code adapter.

Claude Code fires lifecycle hooks (Stop, SessionEnd, PostToolUse) and every
event carries `transcript_path` and `session_id` on stdin as JSON. So the
capture path is: register a Stop hook -> read the transcript JSONL at the
given path -> normalize.

Install (one line in .claude/settings.json):
  "hooks": { "Stop": [ { "hooks": [ {
      "type": "command",
      "command": "agentsmd _hook claude"
  } ] } ] }
"""
import json
import sys

from .base import Adapter


class ClaudeAdapter(Adapter):
    name = "claude"

    def capabilities(self):
        return {"hooks": True, "transcript": "jsonl", "events": ["Stop", "SessionEnd", "PostToolUse"]}

    def read_hook_event(self):
        """Called by `agentsmd _hook claude`. Reads the hook JSON from stdin."""
        return json.load(sys.stdin)

    def latest_trajectory(self, event=None):
        """STUB. Parse the JSONL at event['transcript_path'] into the
        normalized trajectory schema: steps, tool_calls, files_touched,
        commands+exit_codes, test_results, tokens, wall_time, final_diff."""
        # transcript_path = event["transcript_path"]
        # ... parse JSONL, normalize ...
        return None
