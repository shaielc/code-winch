import importlib.util
import json
import shutil
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
            ("git", *args), cwd=repo, check=True, text=True,
            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
        )
        return result.stdout.strip()

    def make_repo(self) -> Path:
        root = Path(tempfile.mkdtemp())
        self.addCleanup(lambda: shutil.rmtree(root, ignore_errors=True))
        self.git(root, "init", "-b", "main")
        self.git(root, "config", "user.name", "Test")
        self.git(root, "config", "user.email", "test@example.com")

        schema = root / "skills/workplan/tasks.schema.json"
        schema.parent.mkdir(parents=True)
        shutil.copy2(Path(__file__).parents[1] / "skills/workplan/tasks.schema.json", schema)

        workplan = root / "docs/workplan"
        (workplan / "phase-0").mkdir(parents=True)
        (workplan / "phase-1").mkdir(parents=True)
        (workplan / "phase-0/p0-001-done.md").write_text("# done\n")
        (workplan / "phase-1/p1-002-future.md").write_text("# future\n")
        (workplan / "README.md").write_text("legacy\n")
        (workplan / "tasks.schema.json").write_text("{}\n")
        tracker = {
            "schema_version": 1,
            "status_values": ["pending", "in_progress", "blocked", "completed"],
            "tasks": [
                {
                    "id": "P0-001", "title": "Done", "phase": 0,
                    "phase_name": "Foundation", "status": "completed",
                    "depends_on": [], "owner": None, "blocked_reason": None,
                    "brief": "phase-0/p0-001-done.md",
                },
                {
                    "id": "P1-002", "title": "Future", "phase": 1,
                    "phase_name": "Future", "status": "pending",
                    "depends_on": ["P0-001"], "owner": None, "blocked_reason": None,
                    "brief": "phase-1/p1-002-future.md",
                },
            ],
        }
        (workplan / "tasks.json").write_text(json.dumps(tracker, indent=2) + "\n")
        self.git(root, "add", ".")
        self.git(root, "commit", "-m", "initial")
        return root

    def test_prepare_preserves_inherited_unfinished_task(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        self.assertEqual(state["inherited_unfinished"], ["P1-002"])
        tracker = json.loads((root / "docs/workplan/v2/tasks.json").read_text())
        self.assertEqual(tracker["schema_version"], 2)
        self.assertEqual([task["id"] for task in tracker["tasks"]], ["P0-001", "P1-002"])
        self.assertTrue((root / "docs/workplan/v2/phase-1/p1-002-future.md").is_file())
        self.assertEqual(tracker["tasks"][0]["workplan_version"], 1)
        self.assertEqual(tracker["tasks"][1]["workplan_version"], 1)

    def test_validation_requires_rewrite_or_terminal_disposition(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        with self.assertRaisesRegex(module.WorkplanError, "no V2 disposition"):
            module.validate_v2_rederivation(generation, state["inherited_unfinished"])

        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        future = next(task for task in tracker["tasks"] if task["id"] == "P1-002")
        future["workplan_version"] = 2
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        module.validate_v2_rederivation(generation, state["inherited_unfinished"])

    def test_validation_accepts_supersession_with_existing_replacement(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        old = next(task for task in tracker["tasks"] if task["id"] == "P1-002")
        old["status"] = "superseded"
        old["superseded_by"] = ["P1-003"]
        tracker["tasks"].append(
            {
                "id": "P1-003", "title": "Replacement", "phase": 1,
                "phase_name": "Future", "status": "pending",
                "depends_on": ["P0-001"], "owner": None, "blocked_reason": None,
                "brief": "phase-1/p1-003-replacement.md", "workplan_version": 2,
                "supersedes": ["P1-002"], "superseded_by": [], "removal_reason": None,
            }
        )
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        module.validate_v2_rederivation(generation, state["inherited_unfinished"])

    def test_validation_requires_removed_reason(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        old = next(task for task in tracker["tasks"] if task["id"] == "P1-002")
        old["status"] = "removed"
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "removal_reason"):
            module.validate_v2_rederivation(generation, state["inherited_unfinished"])


if __name__ == "__main__":
    unittest.main()
