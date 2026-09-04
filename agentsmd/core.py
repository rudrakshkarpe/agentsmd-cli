"""Rule ledger and AGENTS.md rendering.

Rules are stored structured (id, text, provenance, cited count) and the
AGENTS.md file is RENDERED from them. This is what makes provenance,
dedup, hit-counting and rule-level blame possible - a plain text file
can't carry any of that.
"""

import re
import time

DEDUP_THRESHOLD = 0.6  # word-overlap above this = treated as the same rule


def _words(s):
    return set(re.findall(r"[a-z0-9]+", s.lower()))


def overlap(a, b):
    wa, wb = _words(a), _words(b)
    if not wa or not wb:
        return 0.0
    return len(wa & wb) / len(wa | wb)


def find_duplicate(db, text):
    for r in db["rules"]:
        if overlap(r["text"], text) >= DEDUP_THRESHOLD:
            return r
    return None


def add_rule(db, text, origin=None):
    dup = find_duplicate(db, text)
    if dup:
        return None, dup
    rid = f"r{len(db['rules']):03d}"
    rule = {
        "id": rid,
        "text": text.strip(),
        "origin": origin or {},  # {run, task, version} -> powers `blame`
        "cited": 0,
        "born": time.strftime("%Y-%m-%d"),
    }
    db["rules"].append(rule)
    return rule, None


def render(db, header="# AGENTS.md\n"):
    """Return the full AGENTS.md text rendered from the ledger."""
    lines = [
        header.rstrip(),
        "",
        "## Lessons",
        "<!-- managed by agentsmd; agents read these first -->",
        "",
    ]
    for r in db["rules"]:
        lines.append(f"- [{r['id']}] {r['text']}  (cited: {r['cited']})")
    return "\n".join(lines) + "\n"
