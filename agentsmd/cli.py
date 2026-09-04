"""agentsmd CLI entrypoint. Three command families:
authoring     init, template, edit, render, lint
version ctrl  log, diff, status, commit, revert, tag, blame
the loop      watch, learn, pending, promote, reject, prune, optimize, savings
"""

import argparse
import os
import subprocess

from . import __version__, adapters, core, loop, templates, vc
from .project import ARTIFACT, Project, find_root


# ---------- authoring ----------
def cmd_init(args):
    root = find_root() or os.getcwd()
    if find_root():
        print("already an agentsmd project at", root)
    proj = Project(root)
    proj.scaffold()
    if proj.artifact.exists() and not args.force:
        print(f"{ARTIFACT} already exists (use --force to overwrite)")
    else:
        if args.template:
            content = templates.load(args.template)
        elif args.scratch:
            content = "# AGENTS.md\n\n## Lessons\n"
        else:
            content = _interactive_init()
        proj.artifact.write_text(content)
        print(f"created {ARTIFACT}")
    vc.commit(
        proj,
        "init",
        reason="template" if args.template else "manual",
        meta={"template": args.template} if args.template else {},
    )
    print("initialized .agentsmd/  (run: agentsmd log)")


def _interactive_init():
    opts = templates.available()
    print("Start from a template, or scratch?\n")
    for i, t in enumerate(opts, 1):
        print(f"  {i}. {t}")
    print(f"  {len(opts) + 1}. scratch (empty)")
    choice = input("\nchoice> ").strip()
    try:
        n = int(choice)
        if 1 <= n <= len(opts):
            return templates.load(opts[n - 1])
    except ValueError:
        pass
    return "# AGENTS.md\n\n## Lessons\n"


def cmd_template(args):
    if args.action == "list":
        for t in templates.available():
            print(t)
    elif args.action == "use":
        proj = Project.require()
        proj.artifact.write_text(templates.load(args.name))
        vc.commit(proj, f"template: {args.name}", reason="template", meta={"template": args.name})
        print(f"applied template {args.name}")


def cmd_edit(args):
    proj = Project.require()
    editor = os.environ.get("EDITOR", "vi")
    subprocess.call([editor, str(proj.artifact)])


def cmd_render(args):
    proj = Project.require()
    db = proj.load_ledger()
    proj.artifact.write_text(core.render(db))
    print(f"rendered {len(db['rules'])} rules into {ARTIFACT}")


def cmd_lint(args):
    proj = Project.require()
    db = proj.load_ledger()
    issues = 0
    seen: list[dict] = []
    for r in db["rules"]:
        for s in seen:
            if core.overlap(r["text"], s["text"]) >= core.DEDUP_THRESHOLD:
                print(f"duplicate: [{r['id']}] ~ [{s['id']}]")
                issues += 1
        seen.append(r)
        if r["cited"] == 0:
            print(f"never fired: [{r['id']}] {r['text']}")
            issues += 1
    print(f"{issues} issue(s)" if issues else "clean")


# ---------- version control ----------
def cmd_log(args):
    proj = Project.require()
    for e in vc.log(proj):
        delta = e["meta"].get("token_delta")
        tag = f"  [{e['reason']}]"
        extra = f"  tokens {delta:+d}" if isinstance(delta, int) else ""
        print(f"{e['id']}  {e['ts']}{tag}  {e['message']}{extra}")


def cmd_diff(args):
    proj = Project.require()
    hist = vc.log(proj)
    if len(hist) < 2 and not (args.a and args.b):
        print("need two versions to diff")
        return
    a = args.a or hist[-2]["id"]
    b = args.b or hist[-1]["id"]
    print(vc.diff(proj, a, b))


def cmd_status(args):
    proj = Project.require()
    pend = loop.pending(proj)
    print(f"pending rules: {len(pend)}")
    for p in pend:
        print(f"  {p['id']}  {p['text']}")


def cmd_commit(args):
    proj = Project.require()
    e = vc.commit(proj, args.message, reason="manual")
    print(f"committed {e['id']}")


def cmd_revert(args):
    proj = Project.require()
    e = vc.revert(proj, args.version)
    print(f"reverted to {args.version} as {e['id']}")


def cmd_tag(args):
    proj = Project.require()
    vc.tag(proj, args.version, args.name)
    print(f"tagged {args.version} as {args.name}")


def cmd_blame(args):
    """Per-rule provenance: which run/task/version introduced each rule."""
    proj = Project.require()
    for r in proj.load_ledger()["rules"]:
        o = r.get("origin", {})
        src = f"run {o.get('run', '?')}, task {o.get('task', '?')}" if o else "(no provenance)"
        print(f"[{r['id']}]  {src:<28}  {r['text']}")


# ---------- the loop ----------
def cmd_learn(args):
    proj = Project.require()
    adapter = adapters.get(args.adapter)
    traj = adapter.latest_trajectory()
    if traj is None:
        print("no trajectory captured yet (adapter stub). see agentsmd/loop.py::_reflect")
        return
    pid = loop.learn(proj, traj)
    print(f"proposed {pid}  (review: agentsmd pending)" if pid else "no lesson to add")


def cmd_pending(args):
    proj = Project.require()
    for p in loop.pending(proj):
        print(f"{p['id']}  {p['text']}")


def cmd_promote(args):
    proj = Project.require()
    rule, dup = loop.promote(proj, args.id, core, vc)
    if dup:
        print(f"skipped, already covered by [{dup['id']}]")
    else:
        print(f"promoted -> [{rule['id']}] {rule['text']}")


def cmd_reject(args):
    proj = Project.require()
    (proj.pending_dir / f"{args.id}.json").unlink(missing_ok=True)
    print(f"rejected {args.id}")


def cmd_prune(args):
    proj = Project.require()
    db = proj.load_ledger()
    tasks_seen = sum(len(v) for v in db.get("runs", {}).values())
    keep: list[dict] = []
    dropped: list[dict] = []
    for r in db["rules"]:
        (dropped if r["cited"] == 0 and tasks_seen >= 20 else keep).append(r)
    db["rules"] = keep
    proj.save_ledger(db)
    proj.artifact.write_text(core.render(db))
    print("pruned: " + (", ".join(x["id"] for x in dropped) if dropped else "nothing"))


def cmd_savings(args):
    proj = Project.require()
    s = loop.savings(proj, args.task)
    if not s:
        print(f"task {args.task}: need >= 2 runs")
    else:
        print(
            f"task {s['task']}: {s['first']} -> {s['last']} tokens "
            f"({s['pct']:+.0f}% by run #{s['runs']})"
        )


def cmd_optimize(args):
    print("offline GEPA batch: see reflect/ (bridge to gepa.optimize_anything). not in skeleton.")


def cmd_watch(args):
    print("capture daemon: tails transcripts / receives hooks. not in skeleton.")


# ---------- parser ----------
def build_parser():
    p = argparse.ArgumentParser(prog="agentsmd", description=__doc__)
    p.add_argument("--version", action="version", version=f"agentsmd {__version__}")
    sub = p.add_subparsers(dest="cmd", required=True)

    a = sub.add_parser("init", help="create AGENTS.md and .agentsmd/")
    a.add_argument("--template")
    a.add_argument("--scratch", action="store_true")
    a.add_argument("--force", action="store_true")
    a.set_defaults(fn=cmd_init)

    t = sub.add_parser("template", help="list or use templates")
    t.add_argument("action", choices=["list", "use"])
    t.add_argument("name", nargs="?")
    t.set_defaults(fn=cmd_template)

    for name, fn, help_ in [
        ("edit", cmd_edit, "open AGENTS.md in $EDITOR"),
        ("render", cmd_render, "rebuild AGENTS.md from the ledger"),
        ("lint", cmd_lint, "check for duplicate / dead rules"),
        ("status", cmd_status, "show pending rules"),
        ("blame", cmd_blame, "per-rule provenance"),
        ("pending", cmd_pending, "list proposed rules"),
        ("prune", cmd_prune, "retire rules that never fire"),
        ("optimize", cmd_optimize, "offline GEPA batch run"),
        ("watch", cmd_watch, "start the capture daemon"),
    ]:
        s = sub.add_parser(name, help=help_)
        s.set_defaults(fn=fn)

    lg = sub.add_parser("log", help="version history")
    lg.set_defaults(fn=cmd_log)

    d = sub.add_parser("diff", help="diff two versions")
    d.add_argument("a", nargs="?")
    d.add_argument("b", nargs="?")
    d.set_defaults(fn=cmd_diff)

    c = sub.add_parser("commit", help="snapshot AGENTS.md")
    c.add_argument("-m", "--message", required=True)
    c.set_defaults(fn=cmd_commit)

    r = sub.add_parser("revert", help="roll back to a version")
    r.add_argument("version")
    r.set_defaults(fn=cmd_revert)

    tg = sub.add_parser("tag", help="name a version")
    tg.add_argument("version")
    tg.add_argument("name")
    tg.set_defaults(fn=cmd_tag)

    ln = sub.add_parser("learn", help="reflect on last session, propose a rule")
    ln.add_argument("--adapter", default="claude")
    ln.set_defaults(fn=cmd_learn)

    pr = sub.add_parser("promote", help="accept a pending rule")
    pr.add_argument("id")
    pr.set_defaults(fn=cmd_promote)

    rj = sub.add_parser("reject", help="discard a pending rule")
    rj.add_argument("id")
    rj.set_defaults(fn=cmd_reject)

    sv = sub.add_parser("savings", help="token savings for a task")
    sv.add_argument("task")
    sv.set_defaults(fn=cmd_savings)
    return p


def main(argv=None):
    args = build_parser().parse_args(argv)
    args.fn(args)


if __name__ == "__main__":
    main()
