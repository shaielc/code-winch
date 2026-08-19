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

        workplan = root / "docs/workplan"
        (workplan / "phase-0").mkdir(parents=True)
        (workplan / "phase-1").mkdir(parents=True)
        (workplan / "phase-0/p0-001-done.md").write_text("# done\n")
        (workplan / "phase-1/p1-002-abandoned.md").write_text(
            "# abandoned\nSECRET-ANCHOR\n"
        )
        (workplan / "README.md").write_text("future task P1-002 SECRET-ANCHOR\n")
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
                    "title": "Abandoned future",
                    "phase": 1,
                    "phase_name": "Future",
                    "status": "pending",
                    "depends_on": ["P0-001"],
                    "owner": None,
                    "blocked_reason": None,
                    "brief": "phase-1/p1-002-abandoned.md",
                },
            ],
        }
        (workplan / "tasks.json").write_text(json.dumps(tracker, indent=2) + "\n")

        schema = root / "skills/workplan/tasks.schema.json"
        schema.parent.mkdir(parents=True)
        schema.write_text("{}\n")

        script = root / "scripts/prepare_workplan_rederivation.py"
        script.parent.mkdir(parents=True)
        shutil.copy2(Path(__file__).parents[1] / "scripts/prepare_workplan_rederivation.py", script)
        test_copy = root / "tests/test_prepare_workplan_rederivation.py"
        test_copy.parent.mkdir(parents=True)
        test_copy.write_text("sanitizer test fixture\n")

        (root / "AGENTS.md").write_text(
            "# Agent rules\n\nStable rules.\n\n## Current state\nP1-002 is next.\n"
        )
        (root / "product.txt").write_text("ordinary product source\n")

        self.git(root, "add", ".")
        self.git(root, "commit", "-m", "initial")
        return root

    def test_prepare_creates_history_free_checkout(self):
        root = self.make_repo()
        checkout = root.parent / f"{root.name}-derive"
        self.addCleanup(lambda: shutil.rmtree(checkout, ignore_errors=True))

        state = module.prepare(root, checkout)
        self.assertEqual(state["generation"], "v2")

        # The source repository retains the complete v1 generation.
        self.assertTrue((root / "docs/workplan/v1/phase-1/p1-002-abandoned.md").is_file())
        self.assertTrue((root / "docs/workplan/v2/tasks.json").is_file())

        # The isolated checkout contains only v2 and no orchestration machinery.
        self.assertFalse((checkout / "docs/workplan/v1").exists())
        self.assertTrue((checkout / "docs/workplan/v2").is_dir())
        self.assertFalse((checkout / "scripts/prepare_workplan_rederivation.py").exists())
        self.assertFalse((checkout / "tests/test_prepare_workplan_rederivation.py").exists())

        tracker = json.loads((checkout / "docs/workplan/v2/tasks.json").read_text())
        self.assertEqual([task["id"] for task in tracker["tasks"]], ["P0-001"])

        visible = "\n".join(
            text for _, text in module.text_files(checkout)
        )
        self.assertNotIn("P1-002", visible)
        self.assertNotIn("SECRET-ANCHOR", visible)
        self.assertNotIn("## Current state", (checkout / "AGENTS.md").read_text())

        # It is a new repository with one root commit and no remote to fetch history from.
        self.assertEqual(self.git(checkout, "rev-list", "--count", "HEAD"), "1")
        self.assertEqual(self.git(checkout, "remote"), "")
        self.assertEqual(self.git(checkout, "branch", "--show-current"), "workplan-v2")

    def test_absent_old_id_is_reusable_inside_v2(self):
        root = self.make_repo()
        checkout = root.parent / f"{root.name}-reuse"
        self.addCleanup(lambda: shutil.rmtree(checkout, ignore_errors=True))
        state = module.prepare(root, checkout)

        generation = checkout / "docs/workplan/v2"
        tracker_path = generation / "tasks.json"
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
                "workplan_version": 2,
                "supersedes": [],
                "superseded_by": [],
                "removal_reason": None,
            }
        )
        (generation / "phase-1").mkdir(exist_ok=True)
        (generation / "phase-1/p1-002-fresh.md").write_text("# fresh\n")
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")

        module.validate_generation(generation, state["completed"])

    def test_prepare_rejects_unfinished_id_leak_outside_workplan(self):
        root = self.make_repo()
        (root / "product.txt").write_text("implementation note for P1-002\n")
        self.git(root, "add", "product.txt")
        self.git(root, "commit", "-m", "add leak")
        checkout = root.parent / f"{root.name}-leak"
        self.addCleanup(lambda: shutil.rmtree(checkout, ignore_errors=True))

        with self.assertRaisesRegex(module.WorkplanError, "still exposes unfinished task IDs"):
            module.prepare(root, checkout)

    def test_harvest_replaces_only_v2_and_preserves_v1(self):
        root = self.make_repo()
        checkout = root.parent / f"{root.name}-harvest"
        self.addCleanup(lambda: shutil.rmtree(checkout, ignore_errors=True))
        state = module.prepare(root, checkout)

        readme = checkout / "docs/workplan/v2/README.md"
        readme.write_text(readme.read_text() + "\nFresh derivation marker.\n")
        self.git(checkout, "add", "docs/workplan/v2/README.md")
        self.git(checkout, "commit", "-m", "derive plan")

        module.harvest(root)
        self.assertTrue((root / "docs/workplan/v1/phase-1/p1-002-abandoned.md").is_file())
        self.assertIn(
            "Fresh derivation marker",
            (root / "docs/workplan/v2/README.md").read_text(),
        )
        self.assertEqual(state["source_branch"], "main")


if __name__ == "__main__":
    unittest.main()
