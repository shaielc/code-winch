import importlib.util
import io
import json
import shutil
import subprocess
import tarfile
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch


ROOT = Path(__file__).parents[1]
SPEC = importlib.util.spec_from_file_location(
    "prepare_workplan_rederivation", ROOT / "scripts/prepare_workplan_rederivation.py"
)
module = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(module)


class WorkplanRederivationTests(unittest.TestCase):
    def git(self, repo: Path, *args: str, check: bool = True) -> str:
        result = subprocess.run(
            ("git", *args),
            cwd=repo,
            check=check,
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
        (workplan / "phase-0/p0-001-done.md").write_text(
            """# P0-001: Done

## Objective

A real implementation exists.

## Scope

- Implement the finished behavior.

## Runtime reachability

The behavior is reachable now. A later binding was once tracked by P1-002.

## Demonstration

    $ demo
    -> expect: done

## Acceptance criteria

- [ ] The finished behavior works.

## Non-goals

- Future work was once tracked by P1-002.

## Deferrals

| Deferred | Owner |
|---|---|
| old future | P1-002 |

## Traces to

`docs/architecture.md`
"""
        )
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
        shutil.copy2(ROOT / "skills/workplan/tasks.schema.json", schema)
        (root / "skills/workplan/SKILL.md").write_text(
            "# Workplan\nNormal planning rules only.\n"
        )
        contracts = root / "skills/workplan/contracts"
        contracts.mkdir(parents=True)
        (contracts / "v1.md").write_text("# V1 contract\n")
        (contracts / "v2.md").write_text("# V2 contract\nNormal planning rules only.\n")
        (root / "skills/README.md").write_text(
            "# Skills\nUse a task ID from the active tracker.\n"
        )

        script = root / "scripts/prepare_workplan_rederivation.py"
        script.parent.mkdir(parents=True)
        shutil.copy2(ROOT / "scripts/prepare_workplan_rederivation.py", script)
        for name in (
            "test_prepare_workplan_rederivation.py",
            "test_task_scheduler.py",
            "test_stamp_task_completion.py",
        ):
            path = root / "tests" / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text("operator tooling P1-002\n")

        (root / "docs/workplan/rederivation-operator.md").write_text(
            "operator-only sanitize/harvest instructions mentioning P1-002\n"
        )
        (root / "AGENTS.md").write_text(
            "# Agent rules\n\nStable rules.\n\n## Current state\nP1-002 is next.\n"
        )
        (root / "product.txt").write_text("ordinary product source\n")

        self.git(root, "add", ".")
        self.git(root, "commit", "-m", "initial")
        return root

    def checkout_path(self, root: Path, suffix: str) -> Path:
        checkout = root.parent / f"{root.name}-{suffix}"
        self.addCleanup(lambda: shutil.rmtree(checkout, ignore_errors=True))
        return checkout

    def add_valid_task(
        self, workplan: Path, task_id: str = "P1-002", depends=None
    ):
        generation = workplan / "v2"
        tracker_path = generation / "tasks.json"
        tracker = json.loads(tracker_path.read_text())
        brief = f"phase-1/{task_id.lower()}-fresh.md"
        (generation / "phase-1").mkdir(exist_ok=True)
        (generation / brief).write_text(f"# {task_id}: fresh\n")
        task = {
            "id": task_id,
            "title": "Freshly derived work",
            "phase": 1,
            "phase_name": "Fresh",
            "status": "pending",
            "depends_on": list(depends or ["P0-001"]),
            "owner": None,
            "blocked_reason": None,
            "brief": brief,
            "workplan_version": 2,
            "supersedes": [],
            "superseded_by": [],
            "removal_reason": None,
        }
        tracker["tasks"].append(task)
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        return tracker, task

    def test_prepare_uses_clean_completed_snapshot_and_hides_orchestration(self):
        root = self.make_repo()
        checkout = self.checkout_path(root, "derive")
        state = module.prepare(root, checkout)
        self.assertEqual(state["generation"], "v2")
        self.assertTrue(
            (root / "docs/workplan/v1/phase-1/p1-002-abandoned.md").is_file()
        )
        self.assertFalse((checkout / "docs/workplan/v1").exists())
        self.assertEqual((checkout / "docs/workplan/CURRENT").read_text().strip(), "v2")
        self.assertFalse(
            (checkout / "scripts/prepare_workplan_rederivation.py").exists()
        )
        self.assertFalse((checkout / "docs/workplan/rederivation-operator.md").exists())
        self.assertFalse(
            (checkout / "docs/workplan/workplan-v2-review-rejection.md").exists()
        )

        tracker = json.loads((checkout / "docs/workplan/v2/tasks.json").read_text())
        self.assertEqual([task["id"] for task in tracker["tasks"]], ["P0-001"])
        snapshot = checkout / "docs/workplan/v2" / tracker["tasks"][0]["brief"]
        text = snapshot.read_text()
        self.assertIn("A real implementation exists", text)
        self.assertNotIn("P1-002", text)
        self.assertNotIn("Deferrals", text)
        self.assertNotIn("Non-goals", text)

        visible = "\n".join(text for _, text in module.text_files(checkout))
        self.assertNotIn("P1-002", visible)
        self.assertNotIn("SECRET-ANCHOR", visible)
        self.assertNotIn("prepare_workplan_rederivation", visible)
        self.assertNotIn(
            "harvest", (checkout / "skills/workplan/SKILL.md").read_text().lower()
        )
        self.assertNotIn(
            "sanitize",
            (checkout / "skills/workplan/contracts/v2.md").read_text().lower(),
        )
        self.assertEqual(self.git(checkout, "rev-list", "--count", "HEAD"), "1")
        self.assertEqual(self.git(checkout, "remote"), "")

    def test_absent_old_id_is_reusable_inside_v2(self):
        root = self.make_repo()
        checkout = self.checkout_path(root, "reuse")
        state = module.prepare(root, checkout)
        self.add_valid_task(checkout / "docs/workplan")
        module.validate_generation(
            checkout / "docs/workplan", "v2", state["completed"]
        )

    def test_prepare_failure_leaves_source_retryable(self):
        root = self.make_repo()
        (root / "product.txt").write_text("implementation note for P1-002\n")
        self.git(root, "add", "product.txt")
        self.git(root, "commit", "-m", "add leak")
        baseline = self.git(root, "rev-parse", "HEAD")
        checkout = self.checkout_path(root, "failure")
        with self.assertRaisesRegex(
            module.WorkplanError, "still exposes unfinished task IDs"
        ):
            module.prepare(root, checkout)
        self.assertEqual(self.git(root, "rev-parse", "HEAD"), baseline)
        self.assertTrue((root / "docs/workplan/tasks.json").is_file())
        self.assertFalse((root / "docs/workplan/CURRENT").exists())
        self.assertFalse((root / "docs/workplan/v1").exists())
        self.assertFalse(checkout.exists())

    def test_prepare_rolls_back_if_isolated_initialization_fails(self):
        root = self.make_repo()
        baseline = self.git(root, "rev-parse", "HEAD")
        checkout = self.checkout_path(root, "init-failure")
        with patch.object(
            module,
            "init_isolated_repo",
            side_effect=module.WorkplanError("boom"),
        ):
            with self.assertRaisesRegex(module.WorkplanError, "boom"):
                module.prepare(root, checkout)
        self.assertEqual(self.git(root, "rev-parse", "HEAD"), baseline)
        self.assertTrue((root / "docs/workplan/tasks.json").is_file())
        self.assertFalse(checkout.exists())

    def prepared(self):
        root = self.make_repo()
        checkout = self.checkout_path(root, "prepared")
        state = module.prepare(root, checkout)
        return root, checkout, state

    def test_validation_rejects_schema_invalid_tracker(self):
        _, checkout, state = self.prepared()
        tracker_path = checkout / "docs/workplan/v2/tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"][0].pop("workplan_version")
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "does not validate"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_duplicate_ids(self):
        _, checkout, state = self.prepared()
        tracker_path = checkout / "docs/workplan/v2/tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"].append(dict(tracker["tasks"][0]))
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "duplicate task IDs"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_missing_brief(self):
        _, checkout, state = self.prepared()
        _, task = self.add_valid_task(checkout / "docs/workplan")
        (checkout / "docs/workplan/v2" / task["brief"]).unlink()
        with self.assertRaisesRegex(module.WorkplanError, "missing brief"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_unknown_dependency(self):
        _, checkout, state = self.prepared()
        self.add_valid_task(checkout / "docs/workplan", depends=["P1-099"])
        with self.assertRaisesRegex(module.WorkplanError, "unknown dependency"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_dependency_cycle(self):
        _, checkout, state = self.prepared()
        self.add_valid_task(checkout / "docs/workplan", "P1-002", ["P1-003"])
        self.add_valid_task(checkout / "docs/workplan", "P1-003", ["P1-002"])
        with self.assertRaisesRegex(module.WorkplanError, "dependency cycle"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_invalid_replacement_reciprocity(self):
        _, checkout, state = self.prepared()
        tracker, old = self.add_valid_task(checkout / "docs/workplan", "P1-002")
        old["status"] = "superseded"
        old["superseded_by"] = ["P1-003"]
        generation = checkout / "docs/workplan/v2"
        (generation / "phase-1/p1-003-fresh.md").write_text("# fresh\n")
        tracker["tasks"].append(
            {
                "id": "P1-003",
                "title": "Replacement",
                "phase": 1,
                "phase_name": "Fresh",
                "status": "pending",
                "depends_on": ["P0-001"],
                "owner": None,
                "blocked_reason": None,
                "brief": "phase-1/p1-003-fresh.md",
                "workplan_version": 2,
                "supersedes": [],
                "superseded_by": [],
                "removal_reason": None,
            }
        )
        (generation / "tasks.json").write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "reciprocally supersede"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_completed_history_mutation(self):
        _, checkout, state = self.prepared()
        tracker_path = checkout / "docs/workplan/v2/tasks.json"
        tracker = json.loads(tracker_path.read_text())
        tracker["tasks"][0]["title"] = "Changed"
        tracker_path.write_text(json.dumps(tracker, indent=2) + "\n")
        with self.assertRaisesRegex(module.WorkplanError, "removed or mutated"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_validation_rejects_invalid_current_marker(self):
        _, checkout, state = self.prepared()
        (checkout / "docs/workplan/CURRENT").write_text("v3\n")
        with self.assertRaisesRegex(module.WorkplanError, "active-generation marker"):
            module.validate_generation(
                checkout / "docs/workplan", "v2", state["completed"]
            )

    def test_prepare_succeeds_against_repository_tree(self):
        probe = subprocess.run(
            ("git", "-C", str(ROOT), "rev-parse", "--verify", "HEAD"),
            check=False,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        if probe.returncode:
            self.skipTest("test checkout is not inside a git repository")

        source = Path(tempfile.mkdtemp())
        self.addCleanup(lambda: shutil.rmtree(source, ignore_errors=True))
        archive = subprocess.run(
            ("git", "-C", str(ROOT), "archive", "--format=tar", "HEAD"),
            check=True,
            stdout=subprocess.PIPE,
        ).stdout
        with tarfile.open(fileobj=io.BytesIO(archive), mode="r:") as tar:
            tar.extractall(source)
        self.git(source, "init", "-b", "test-source")
        self.git(source, "config", "user.name", "Test")
        self.git(source, "config", "user.email", "test@example.com")
        self.git(source, "add", ".")
        self.git(source, "commit", "-m", "repository snapshot")

        checkout = self.checkout_path(source, "real-tree")
        state = module.prepare(source, checkout)
        self.assertEqual(state["generation"], "v2")
        visible = "\n".join(text for _, text in module.text_files(checkout))
        original = json.loads((source / "docs/workplan/v1/tasks.json").read_text())
        unfinished = {
            task["id"] for task in original["tasks"] if task["status"] != "completed"
        }
        self.assertFalse(set(module.TASK_ID_RE.findall(visible)) & unfinished)

    def test_harvest_refuses_source_branch_drift(self):
        root, checkout, _ = self.prepared()
        (root / "unrelated.txt").write_text("drift\n")
        self.git(root, "add", "unrelated.txt")
        self.git(root, "commit", "-m", "drift")
        with self.assertRaisesRegex(module.WorkplanError, "source branch moved"):
            module.harvest(root)

    def test_harvest_replaces_only_v2_and_preserves_v1(self):
        root, checkout, state = self.prepared()
        readme = checkout / "docs/workplan/v2/README.md"
        readme.write_text(readme.read_text() + "\nFresh derivation marker.\n")
        self.git(checkout, "add", "docs/workplan/v2/README.md")
        self.git(checkout, "commit", "-m", "derive plan")
        module.harvest(root)
        self.assertTrue(
            (root / "docs/workplan/v1/phase-1/p1-002-abandoned.md").is_file()
        )
        self.assertIn(
            "Fresh derivation marker",
            (root / "docs/workplan/v2/README.md").read_text(),
        )
        self.assertEqual(state["source_branch"], "main")


if __name__ == "__main__":
    unittest.main()
