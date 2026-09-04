"""The self-improvement loop.

Design: FAST TO SUGGEST, SLOW TO TRUST.
  learn    reflect on the last captured session, propose ONE delta -> pending/
  promote  human (or evidence) moves a pending rule into the active ledger
  savings  report token delta across runs of the same task

The reflect step here is a stub. In production it calls an LLM (or GEPA's
reflection component) with the normalized trajectory. Kept separate so the
capture, reflection, and gate layers can evolve independently.
"""
import json
import time


def learn(proj, trajectory):
    """Reflect on a trajectory, write a proposed rule to pending/.
    STUB: replace _reflect with a real LLM/GEPA call."""
    proposal = _reflect(trajectory)
    if not proposal:
        return None
    pid = f"p{int(time.time())}"
    (proj.pending_dir / f"{pid}.json").write_text(json.dumps({
        "id": pid,
        "text": proposal["text"],
        "origin": proposal.get("origin", {}),
        "proposed": time.strftime("%Y-%m-%dT%H:%M:%S"),
    }, indent=2))
    return pid


def _reflect(trajectory):
    """STUB. Return {"text": rule, "origin": {...}} or None.
    Real version: attribute the failure, emit one targeted rule, or None
    when the failure is not AGENTS.md-attributable (the anti-bloat verdict)."""
    return None


def pending(proj):
    return [json.loads(p.read_text()) for p in sorted(proj.pending_dir.glob("*.json"))]


def promote(proj, pid, core, vc):
    """Move a pending rule into the active ledger, render, and version it."""
    src = proj.pending_dir / f"{pid}.json"
    if not src.exists():
        raise SystemExit(f"no pending rule {pid}")
    prop = json.loads(src.read_text())
    db = proj.load_ledger()
    rule, dup = core.add_rule(db, prop["text"], origin=prop.get("origin"))
    if dup:
        src.unlink()
        return None, dup
    proj.save_ledger(db)
    proj.artifact.write_text(core.render(db))
    src.unlink()
    vc.commit(proj, f"learned: {rule['text']}", reason="learned", meta=prop.get("origin", {}))
    return rule, None


def savings(proj, task):
    db = proj.load_ledger()
    runs = db.get("runs", {}).get(task, [])
    if len(runs) < 2:
        return None
    first, last = runs[0], runs[-1]
    return {"task": task, "first": first, "last": last,
            "pct": 100 * (first - last) / first, "runs": len(runs)}
