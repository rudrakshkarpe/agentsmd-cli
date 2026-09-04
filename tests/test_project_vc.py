import tempfile
import unittest
from pathlib import Path

from agentsmd import vc
from agentsmd.project import Project, find_root


class ProjectVersionControlTests(unittest.TestCase):
    def test_scaffold_commit_diff_and_revert(self):
        with tempfile.TemporaryDirectory() as directory:
            project = Project(directory)
            project.scaffold()
            project.artifact.write_text("first\n")
            first = vc.commit(project, "initial")
            project.artifact.write_text("second\n")
            second = vc.commit(project, "change")

            self.assertIn("-first", vc.diff(project, first["id"], second["id"]))
            reverted = vc.revert(project, first["id"])
            self.assertEqual(project.artifact.read_text(), "first\n")
            self.assertEqual(reverted["meta"]["reverted_to"], first["id"])

    def test_find_root_walks_up_from_a_child(self):
        with tempfile.TemporaryDirectory() as directory:
            project = Project(directory)
            project.scaffold()
            child = Path(directory) / "src" / "nested"
            child.mkdir(parents=True)

            self.assertEqual(find_root(child), Path(directory).resolve())


if __name__ == "__main__":
    unittest.main()
