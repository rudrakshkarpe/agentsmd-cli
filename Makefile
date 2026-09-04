.PHONY: bootstrap format lint type test build package-check check clean

PYTHON ?= python3
VENV ?= .venv
BIN := $(VENV)/bin

bootstrap:
	$(PYTHON) -m venv $(VENV)
	$(BIN)/python -m pip install --upgrade pip
	$(BIN)/python -m pip install -e '.[dev]'

format:
	$(BIN)/ruff format .
	$(BIN)/ruff check --fix .

lint:
	$(BIN)/ruff format --check .
	$(BIN)/ruff check .

type:
	$(BIN)/mypy agentsmd

test:
	$(BIN)/python -m unittest discover -s tests -v

build:
	$(BIN)/python -m build

package-check: build
	$(BIN)/twine check dist/*
	$(BIN)/python scripts/smoke_wheel.py

check: lint type test package-check

clean:
	rm -rf build dist .mypy_cache .ruff_cache
