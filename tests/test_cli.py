import contextlib
import io
import os
import tempfile
import unittest
from pathlib import Path

from agentsmd import cli, templates


class CliTests(unittest.TestCase):
    def test_templates_are_available(self):
        self.assertIn("minimal", templates.available())

    def test_init_creates_a_versioned_project(self):
        with tempfile.TemporaryDirectory() as directory:
            previous = os.getcwd()
            output = io.StringIO()
            try:
                os.chdir(directory)
                with contextlib.redirect_stdout(output):
                    cli.main(["init", "--template", "minimal"])
                    cli.main(["log"])
            finally:
                os.chdir(previous)

            self.assertTrue(Path(directory, "AGENTS.md").exists())
            self.assertFalse(Path(directory, ".agentsmd", "ledger.json").exists())
            self.assertIn("v0000", output.getvalue())


if __name__ == "__main__":
    unittest.main()
