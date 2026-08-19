import importlib.util
import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SPEC = importlib.util.spec_from_file_location(
    "prepare_workplan_rederivation",
    Path(__file__).parents[1] / "scripts/prepare_workplan_rederivation.py",
)
module = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(module)


class WorkplanRederivationTests(unittest.TestCase):
    def git(self, repo: Path, *args: str) -> str:
        result = subprocess.run(
            ("git", *args),
            cwd=repo,
            check=True,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        return result.stdout.strip()

    def make_repo(self) -> Path:
        root = Path(tempfile.mkdtemp())
        self.addCleanup(lambda: __import__("shutil").rmtree(root, ignore_errors=True))
        self.git(root, "init", "-b", "main")
        self.git(root, "config", "user.name", "Test")
        self.git(root, "config", "user.email", "test@example.com")

        workplan = root / "docs/workplan"
        (workplan / "phase-0").mkdir(parents=True)
        (workplan / "phase-1").mkdir(parents=True)
        (workplan / "phase-0/p0-001-done.md").write_text("# P0-001 done\n")
        (workplan / "phase-1/p1-002-future.md").write_text("# P1-002 future\n")
        (workplan / "README.md").write_text("old future plan mentioning P1-002\n")
        (workplan / "tasks.schema.json").write_text("{}\n")
        tracker = {
            "schema_version": 1,
            "status_values": ["pending", "in_progress", "blocked", "completed"],
            "tasks": [
                {
                    "id": "P0-001",
                    "title": "Done",
                    "phase": 0,
                    "phase_name": "Foundation",
                    "status": "completed",
                    "depends_on": [],
                    "owner": None,
                    "blocked_reason": None,
                    "brief": "phase-0/p0-001-done.md",
                },
                {
                    "id": "P1-002",
                    "title": "Future",
                    "phase": 1,
                    "phase_name": "Future",
                    "status": "pending",
                    "depends_on": ["P0-001"],
                    "owner": None,
                    "blocked_reason": None,
                    "brief": "phase-1/p1-002-future.md",
                },
            ],
        }
        (workplan / "tasks.json").write_text(json.dumps(tracker, indent=2) + "\n")
        (root / "README.md").write_text("repo\n")
        self.git(root, "add", ".")
        self.git(root, "commit", "-m", "initial")
        return root

    def test_prepare_archives_full_v1_and_exposes_completed_only_v2(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")

        self.assertEqual(state["generation"], "v2")
        self.assertEqual(self.git(root, "branch", "--show-current"), "derive-v2")

        workplan = root / "docs/workplan"
        self.assertFalse((workplan / "v1").exists())
        self.assertTrue((workplan / "v2/phase-0/p0-001-done.md").is_file())
        self.assertFalse((workplan / "v2/phase-1/p1-002-future.md").exists())
        self.assertEqual((workplan / "CURRENT").read_text().strip(), "v2")

        tracker = json.loads((workplan / "v2/tasks.json").read_text())
        self.assertEqual([task["id"] for task in tracker["tasks"]], ["P0-001"])
        self.assertNotIn("P1-002", (workplan / "v2/README.md").read_text())

        self.git(root, "switch", "main")
        self.assertTrue((workplan / "v1/phase-1/p1-002-future.md").is_file())
        archived = json.loads((workplan / "v1/tasks.json").read_text())
        self.assertEqual(len(archived["tasks"]), 2)
        self.assertTrue((workplan / "v2/tasks.json").is_file())

    def test_new_generation_can_reuse_an_abandoned_id(self):
        root = self.make_repo()
        module.prepare(root, Path("docs/workplan"), "derive-v2")
        tracker_path = root / "docs/workplan/v2/tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"].append(
            {
                "id": "P1-002",
                "title": "Freshly derived work",
                "phase": 1,
                "phase_name": "Fresh",
                "status": "pending",
                "depends_on": ["P0-001"],
                "owner": None,
                "blocked_reason": None,
                "brief": "phase-1/p1-002-fresh.md",
            }
        )
        (root / "docs/workplan/v2/phase-1").mkdir(exist_ok=True)
        (root / "docs/workplan/v2/phase-1/p1-002-fresh.md").write_text("# fresh\n")
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        self.git(root, "add", "docs/workplan/v2")
        self.git(root, "commit", "-m", "rederive")

        module.harvest(root)
        self.assertEqual(self.git(root, "branch", "--show-current"), "main")
        self.assertTrue((root / "docs/workplan/v1/phase-1/p1-002-future.md").is_file())
        self.assertTrue((root / "docs/workplan/v2/phase-1/p1-002-fresh.md").is_file())
        harvested = json.loads((root / "docs/workplan/v2/tasks.json").read_text())
        fresh = next(task for task in harvested["tasks"] if task["id"] == "P1-002")
        self.assertEqual(fresh["title"], "Freshly derived work")

    def test_prepare_rejects_completed_task_that_depends_on_pending_task(self):
        root = self.make_repo()
        tracker_path = root / "docs/workplan/tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"][0]["depends_on"] = ["P1-002"]
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        self.git(root, "add", str(tracker_path.relative_to(root)))
        self.git(root, "commit", "-m", "make invalid frontier")

        with self.assertRaises(module.WorkplanError):
            module.prepare(root, Path("docs/workplan"), "derive-v2")


if __name__ == "__main__":
    unittest.main()
