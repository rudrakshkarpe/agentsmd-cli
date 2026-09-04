import json
import tempfile
import unittest

from agentsmd import core, loop, vc
from agentsmd.project import Project


class LoopTests(unittest.TestCase):
    def test_promote_moves_pending_rule_to_ledger_and_versions_it(self):
        with tempfile.TemporaryDirectory() as directory:
            project = Project(directory)
            project.scaffold()
            project.artifact.write_text("# AGENTS.md\n")
            proposal = {
                "id": "p1",
                "text": "Read the schema before changing stored data.",
                "origin": {"run": "session-1", "task": "schema"},
            }
            (project.pending_dir / "p1.json").write_text(json.dumps(proposal))

            rule, duplicate = loop.promote(project, "p1", core, vc)

            self.assertIsNone(duplicate)
            self.assertEqual(rule["id"], "r000")
            self.assertFalse((project.pending_dir / "p1.json").exists())
            self.assertIn("Read the schema", project.artifact.read_text())
            self.assertEqual(vc.log(project)[0]["reason"], "learned")

    def test_savings_requires_two_runs_and_reports_change(self):
        with tempfile.TemporaryDirectory() as directory:
            project = Project(directory)
            project.scaffold()
            project.save_ledger({"rules": [], "runs": {"task-1": [1000, 700]}})

            result = loop.savings(project, "task-1")

            self.assertEqual(result["pct"], 30)
            self.assertEqual(result["runs"], 2)
            self.assertIsNone(loop.savings(project, "unknown"))


if __name__ == "__main__":
    unittest.main()
