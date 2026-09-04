# Development setup

```sh
python3 -m venv .venv
.venv/bin/python -m pip install -e .
.venv/bin/agentsmd --version
```

The current implementation is a dependency-free Python skeleton. The immediate milestone is Phase 2: normalize a real Claude Code transcript, reflect once at the session boundary, create a pending rule or no-op verdict, promote an approved rule, and demonstrate measurable token savings.

Project context and previous decisions are preserved in `docs/reference/HANDOFF.md`. The imported README is in `docs/reference/ORIGINAL-README.md`; the repository's active overview remains `README.md`.

## Useful checks

```sh
.venv/bin/python -m compileall agentsmd
.venv/bin/agentsmd --help
```

