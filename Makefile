.PHONY: bootstrap format lint type test build package-check py-check go-format go-vet go-test go-build go-check release-dry-run check clean

PYTHON ?= python3
VENV ?= .venv
BIN := $(VENV)/bin
GO_FILES := $(shell find cmd cli capture detect integration learning ledger project reflect schema template version -name '*.go' -type f)

bootstrap:
	$(PYTHON) -m venv $(VENV)
	$(BIN)/python -m pip install --upgrade pip
	$(BIN)/python -m pip install -e '.[dev]'
	go mod download

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
	$(BIN)/twine check dist/*.whl dist/*.tar.gz
	$(BIN)/python scripts/smoke_wheel.py

py-check: lint type test package-check

go-format:
	@test -z "$$(gofmt -l $(GO_FILES))" || (gofmt -l $(GO_FILES) && exit 1)

go-vet:
	go vet ./...

go-test:
	go test -race ./...

go-build:
	go build -trimpath -o dist/agentsmd ./cmd/agentsmd

go-check: go-format go-vet go-test go-build

release-dry-run:
	./scripts/package-release.sh v0.0.0-test dist/release

check: go-check py-check

clean:
	rm -rf build dist .mypy_cache .ruff_cache
