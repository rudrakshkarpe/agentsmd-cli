# Development setup

```sh
make bootstrap
.venv/bin/agentsmd --version
```

The current implementation is a dependency-free Python skeleton. The immediate milestone is Phase 2: normalize a real Claude Code transcript, reflect once at the session boundary, create a pending rule or no-op verdict, promote an approved rule, and demonstrate measurable token savings.

Project context and previous decisions are preserved in `docs/reference/HANDOFF.md`. The imported README is in `docs/reference/ORIGINAL-README.md`; the repository's active overview remains `README.md`.

## Useful checks

```sh
make format        # format and autofix imports
make test          # standard-library unit tests
make check         # exact local quality, test, and package gates
```

CI applies the same gates on every pull request and push to `main`. It tests Python 3.9 and 3.13, covers Linux, macOS, and Windows, and builds and installs the wheel in a clean environment.

## Language plan

The executable specification is being developed in Python while the CLI surface, schemas, and learning loop stabilize. The production distribution will be ported to Go with Cobra once that surface is frozen, giving users a single binary for Homebrew and direct downloads. Avoid coupling the normative schemas to Python so the port remains mechanical.
