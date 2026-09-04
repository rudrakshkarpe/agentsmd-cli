"""Capture adapters, one per CLI.

claude  hooks (Stop/SessionEnd) + transcript JSONL       [implemented: stub]
codex   watch ~/.codex/transcripts/*.jsonl               [planned]
cursor  watch ~/.cursor/*/transcript.jsonl               [planned]
goose   no hooks; export session / read SQLite store      [planned]
"""

from .claude import ClaudeAdapter

REGISTRY = {
    "claude": ClaudeAdapter,
    # "codex": CodexAdapter,
    # "cursor": CursorAdapter,
    # "goose": GooseAdapter,
}


def get(name):
    if name not in REGISTRY:
        raise SystemExit(f"unknown adapter: {name}. available: {', '.join(REGISTRY)}")
    return REGISTRY[name]()
