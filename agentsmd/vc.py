"""Version control for AGENTS.md.

We do NOT reinvent git's object store, and we don't force raw git either.
Each version is a snapshot of AGENTS.md PLUS the reason it exists:
  - learned   a promoted rule (records run, task, token delta)
  - manual    a hand edit
  - template  a base template change

That 'reason' metadata is the whole point: it lets `log` show HOW the file
got better, which `git log` alone cannot. Production can back this with
libgit2 for durability; the model stays the same.
"""

import difflib
import json
import time

INDEX = "versions.jsonl"


def _index_path(proj):
    return proj.versions_dir / INDEX


def _read_index(proj):
    p = _index_path(proj)
    if not p.exists():
        return []
    return [json.loads(line) for line in p.read_text().splitlines() if line.strip()]


def commit(proj, message, reason="manual", meta=None):
    """Snapshot the current AGENTS.md as a new version."""
    idx = _read_index(proj)
    vid = f"v{len(idx):04d}"
    content = proj.artifact.read_text() if proj.artifact.exists() else ""
    (proj.versions_dir / f"{vid}.md").write_text(content)
    entry = {
        "id": vid,
        "parent": idx[-1]["id"] if idx else None,
        "ts": time.strftime("%Y-%m-%dT%H:%M:%S"),
        "reason": reason,  # learned | manual | template
        "message": message,
        "meta": meta or {},  # e.g. {"task": "011", "token_delta": -6860}
    }
    with open(_index_path(proj), "a") as f:
        f.write(json.dumps(entry) + "\n")
    return entry


def log(proj):
    return _read_index(proj)


def _content(proj, vid):
    return (proj.versions_dir / f"{vid}.md").read_text()


def diff(proj, a, b):
    ta, tb = _content(proj, a).splitlines(), _content(proj, b).splitlines()
    return "\n".join(difflib.unified_diff(ta, tb, fromfile=a, tofile=b, lineterm=""))


def revert(proj, vid):
    proj.artifact.write_text(_content(proj, vid))
    return commit(proj, f"revert to {vid}", reason="manual", meta={"reverted_to": vid})


def tag(proj, vid, name):
    tags = proj.versions_dir / "tags.json"
    d = json.loads(tags.read_text()) if tags.exists() else {}
    d[name] = vid
    tags.write_text(json.dumps(d, indent=2))
