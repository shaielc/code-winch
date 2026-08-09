import importlib.util
import unittest
from pathlib import Path


SPEC = importlib.util.spec_from_file_location(
    "task_daemon", Path(__file__).parents[1] / "scripts/task_daemon.py"
)
task_daemon = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(task_daemon)


class TaskDaemonTests(unittest.TestCase):
    def test_task_id_for_pr_accepts_one_known_id(self):
        pull = {"title": "Implement P0-001", "body": "Task: P0-001", "head": {"ref": "work"}}
        self.assertEqual(task_daemon.task_id_for_pr(pull, {"P0-001"}), "P0-001")

    def test_task_id_for_pr_rejects_ambiguous_ids(self):
        pull = {"title": "P0-001 and P0-002", "body": "", "head": {"ref": "work"}}
        self.assertIsNone(task_daemon.task_id_for_pr(pull, {"P0-001", "P0-002"}))

    def test_effective_tracker_applies_only_local_lifecycle_fields(self):
        tracker = {
            "tasks": [
                {
                    "id": "P0-001",
                    "title": "Upstream",
                    "status": "pending",
                    "owner": None,
                    "blocked_reason": None,
                }
            ]
        }
        state = {
            "tasks": {
                "P0-001": {
                    "status": "in_progress",
                    "owner": "worker",
                    "blocked_reason": None,
                    "cloud_output": "ignored",
                }
            }
        }
        effective = task_daemon.effective_tracker(tracker, state)
        self.assertEqual(effective["tasks"][0]["title"], "Upstream")
        self.assertEqual(effective["tasks"][0]["status"], "in_progress")
        self.assertEqual(effective["tasks"][0]["owner"], "worker")
        self.assertNotIn("cloud_output", effective["tasks"][0])

    def test_record_completions_is_idempotent(self):
        state = {"schema_version": 1, "tasks": {}}
        pulls = [
            {
                "title": "P0-001",
                "body": "",
                "head": {"ref": "work"},
                "html_url": "https://example/pr/1",
                "merged_at": "2026-01-01T00:00:00Z",
            }
        ]
        self.assertTrue(task_daemon.record_completions(state, pulls, {"P0-001"}))
        self.assertFalse(task_daemon.record_completions(state, pulls, {"P0-001"}))
        self.assertEqual(state["tasks"]["P0-001"]["status"], "completed")


if __name__ == "__main__":
    unittest.main()
