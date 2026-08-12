import importlib.util
import io
import sys
import unittest
from contextlib import redirect_stderr
from pathlib import Path
from unittest.mock import patch


SPEC = importlib.util.spec_from_file_location(
    "task_scheduler", Path(__file__).parents[1] / "scripts/task_scheduler.py"
)
task_scheduler = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(task_scheduler)


class TaskSchedulerTests(unittest.TestCase):
    def test_merged_pull_request_only_accepts_merged_close_event(self):
        pull = {"merged_at": "2026-01-01T00:00:00Z"}
        payload = {"action": "closed", "pull_request": pull}
        self.assertIs(task_scheduler.merged_pull_request(payload), pull)
        self.assertIsNone(
            task_scheduler.merged_pull_request(
                {"action": "opened", "pull_request": pull}
            )
        )

    def test_task_id_for_pr_accepts_one_known_id(self):
        pull = {"title": "Implement P0-001", "body": "Task: P0-001", "head": {"ref": "work"}}
        self.assertEqual(task_scheduler.task_id_for_pr(pull, {"P0-001"}), "P0-001")

    def test_task_id_for_pr_rejects_ambiguous_ids(self):
        pull = {"title": "P0-001 and P0-002", "body": "", "head": {"ref": "work"}}
        self.assertIsNone(task_scheduler.task_id_for_pr(pull, {"P0-001", "P0-002"}))

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
        effective = task_scheduler.effective_tracker(tracker, state)
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
        self.assertTrue(task_scheduler.record_completions(state, pulls, {"P0-001"}))
        self.assertFalse(task_scheduler.record_completions(state, pulls, {"P0-001"}))
        self.assertEqual(state["tasks"]["P0-001"]["status"], "completed")

    def test_repo_root_requires_an_explicit_state_file(self):
        argv = ["task_scheduler.py", "--env", "environment", "--repo-root", "/opt/code-winch"]
        with patch.object(sys, "argv", argv), redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit):
                task_scheduler.parse_args()
        with patch.object(sys, "argv", [*argv, "--state-file", "/var/lib/state.json"]):
            self.assertEqual(task_scheduler.parse_args().repo_root, Path("/opt/code-winch"))


if __name__ == "__main__":
    unittest.main()
