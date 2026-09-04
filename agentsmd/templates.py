"""Built-in AGENTS.md starting points. Ship as plain .md files under
templates/ so the community can PR new ones without touching code."""

import pathlib
import sys


def _template_dir():
    source_dir = pathlib.Path(__file__).parent.parent / "templates"
    if source_dir.is_dir():
        return source_dir
    return pathlib.Path(sys.prefix) / "share" / "agentsmd" / "templates"


def available():
    template_dir = _template_dir()
    if not template_dir.is_dir():
        return []
    return sorted(p.stem for p in template_dir.glob("*.md"))


def load(name):
    p = _template_dir() / f"{name}.md"
    if not p.exists():
        raise SystemExit(f"unknown template: {name}. available: {', '.join(available())}")
    return p.read_text()
