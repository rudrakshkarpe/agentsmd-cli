"""Capture adapter interface. One per CLI. Capture differs per tool;
everything downstream (normalize output -> reflect -> gate -> render)
is written once. Adding a CLI = implementing this, nothing else."""
from abc import ABC, abstractmethod


class Adapter(ABC):
    name = "base"

    @abstractmethod
    def capabilities(self):
        """Return dict, e.g. {"hooks": True, "transcript": "jsonl"}."""

    @abstractmethod
    def latest_trajectory(self):
        """Read the most recent session and return the normalized
        trajectory schema (see schema/trajectory.schema.json)."""
