"""Built-in AGENTS.md starting points. Ship as plain .md files under
templates/ so the community can PR new ones without touching code."""
import pathlib

TEMPLATE_DIR = pathlib.Path(__file__).parent.parent / "templates"


def available():
    if not TEMPLATE_DIR.is_dir():
        return []
    return sorted(p.stem for p in TEMPLATE_DIR.glob("*.md"))


def load(name):
    p = TEMPLATE_DIR / f"{name}.md"
    if not p.exists():
        raise SystemExit(f"unknown template: {name}. available: {', '.join(available())}")
    return p.read_text()
