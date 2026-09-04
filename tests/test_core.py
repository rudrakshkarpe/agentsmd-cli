import unittest

from agentsmd import core


class CoreTests(unittest.TestCase):
    def test_add_rule_records_provenance_and_renders(self):
        ledger = {"rules": [], "runs": {}}

        rule, duplicate = core.add_rule(
            ledger,
            "Run the focused tests before the full suite.",
            origin={"run": "session-1", "task": "tests"},
        )

        self.assertIsNone(duplicate)
        self.assertEqual(rule["id"], "r000")
        self.assertEqual(rule["origin"]["run"], "session-1")
        self.assertIn("[r000] Run the focused tests", core.render(ledger))

    def test_add_rule_rejects_a_near_duplicate(self):
        ledger = {"rules": [], "runs": {}}
        first, _ = core.add_rule(ledger, "Run focused tests before the full test suite")
        second, duplicate = core.add_rule(ledger, "Run the focused test suite before full tests")

        self.assertIsNotNone(first)
        self.assertIsNone(second)
        self.assertEqual(duplicate["id"], "r000")


if __name__ == "__main__":
    unittest.main()
