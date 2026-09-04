"""Locates the .agentsmd project dir and reads/writes config. The layout
mirrors git: AGENTS.md is the artifact at repo root, .agentsmd/ is machinery."""

import json
import pathlib

DIR = ".agentsmd"
ARTIFACT = "AGENTS.md"


def find_root(start="."):
    """Walk up from start looking for a .agentsmd dir. Returns Path or None."""
    p = pathlib.Path(start).resolve()
    for cand in [p, *p.parents]:
        if (cand / DIR).is_dir():
            return cand
    return None


class Project:
    def __init__(self, root):
        self.root = pathlib.Path(root)
        self.dir = self.root / DIR
        self.artifact = self.root / ARTIFACT
        self.ledger_path = self.dir / "ledger.json"
        self.config_path = self.dir / "config.yaml"
        self.versions_dir = self.dir / "versions"
        self.pending_dir = self.dir / "pending"
        self.runs_dir = self.dir / "runs"

    @classmethod
    def require(cls):
        root = find_root()
        if not root:
            raise SystemExit("not an agentsmd project (no .agentsmd dir). run: agentsmd init")
        return cls(root)

    def scaffold(self):
        for d in (self.dir, self.versions_dir, self.pending_dir, self.runs_dir):
            d.mkdir(parents=True, exist_ok=True)

    def load_ledger(self):
        if self.ledger_path.exists():
            return json.loads(self.ledger_path.read_text())
        return {"rules": [], "runs": {}}

    def save_ledger(self, db):
        self.ledger_path.write_text(json.dumps(db, indent=2))
