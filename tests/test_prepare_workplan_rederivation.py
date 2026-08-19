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
        (workplan / "phase-1/p1-002-abandoned.md").write_text(
            "# P1-002 abandoned future plan\nSECRET-ANCHOR\n"
        )
        (workplan / "README.md").write_text("old plan names P1-002 and SECRET-ANCHOR\n")
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
                    "id": "P1-002", "title": "Abandoned Future", "phase": 1,
                    "phase_name": "Future", "status": "pending",
                    "depends_on": ["P0-001"], "owner": None, "blocked_reason": None,
                    "brief": "phase-1/p1-002-abandoned.md",
                },
            ],
        }
        (workplan / "tasks.json").write_text(json.dumps(tracker, indent=2) + "\n")
        self.git(root, "add", ".")
        self.git(root, "commit", "-m", "initial")
        return root

    def test_prepare_is_clean_slate_completed_history_only(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        self.assertEqual(state["generation"], "v2")
        self.assertEqual(self.git(root, "branch", "--show-current"), "derive-v2")

        generation = root / "docs/workplan/v2"
        tracker = json.loads((generation / "tasks.json").read_text())
        self.assertEqual([task["id"] for task in tracker["tasks"]], ["P0-001"])
        self.assertFalse((generation / "phase-1/p1-002-abandoned.md").exists())

        visible = "\n".join(path.read_text() for path in generation.rglob("*") if path.is_file())
        self.assertNotIn("P1-002", visible)
        self.assertNotIn("Abandoned Future", visible)
        self.assertNotIn("SECRET-ANCHOR", visible)

        self.git(root, "switch", "main")
        self.assertTrue((root / "docs/workplan/v1/phase-1/p1-002-abandoned.md").is_file())

    def test_abandoned_id_can_be_reused_in_new_generation(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"].append({
            "id": "P1-002", "title": "Freshly derived work", "phase": 1,
            "phase_name": "Fresh", "status": "pending", "depends_on": ["P0-001"],
            "owner": None, "blocked_reason": None, "brief": "phase-1/p1-002-fresh.md",
            "workplan_version": 2, "supersedes": [], "superseded_by": [], "removal_reason": None,
        })
        (generation / "phase-1").mkdir(exist_ok=True)
        (generation / "phase-1/p1-002-fresh.md").write_text("# fresh\n")
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        module.validate_generation(generation, state["completed_history"])

    def test_validation_rejects_missing_brief(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"].append({
            "id": "P1-003", "title": "Broken", "phase": 1, "phase_name": "Fresh",
            "status": "pending", "depends_on": ["P0-001"], "owner": None,
            "blocked_reason": None, "brief": "phase-1/missing.md", "workplan_version": 2,
            "supersedes": [], "superseded_by": [], "removal_reason": None,
        })
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "missing brief"):
            module.validate_generation(generation, state["completed_history"])

    def test_validation_rejects_cycle(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        (generation / "phase-1").mkdir(exist_ok=True)
        for task_id, dependency in (("P1-003", "P1-004"), ("P1-004", "P1-003")):
            brief = f"phase-1/{task_id.lower()}.md"
            (generation / brief).write_text(f"# {task_id}\n")
            tracker["tasks"].append({
                "id": task_id, "title": task_id, "phase": 1, "phase_name": "Fresh",
                "status": "pending", "depends_on": [dependency], "owner": None,
                "blocked_reason": None, "brief": brief, "workplan_version": 2,
                "supersedes": [], "superseded_by": [], "removal_reason": None,
            })
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "dependency cycle"):
            module.validate_generation(generation, state["completed_history"])

    def test_validation_rejects_mutated_completed_history(self):
        root = self.make_repo()
        state = module.prepare(root, Path("docs/workplan"), "derive-v2")
        generation = root / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"][0]["title"] = "Rewritten history"
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "immutable"):
            module.validate_generation(generation, state["completed_history"])


if __name__ == "__main__":
    unittest.main()
