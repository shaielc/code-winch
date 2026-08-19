import importlib.util
import io
import subprocess
import sys
import unittest
from contextlib import redirect_stderr
from pathlib import Path
from string import Template
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
        self.assertIs(task_scheduler.merged_pull_request({"action": "closed", "pull_request": pull}), pull)
        self.assertIsNone(task_scheduler.merged_pull_request({"action": "opened", "pull_request": pull}))

    def test_task_id_for_pr_accepts_one_known_id(self):
        pull = {"title": "Implement P0-001", "body": "Task: P0-001", "head": {"ref": "work"}}
        self.assertEqual(task_scheduler.task_id_for_pr(pull, {"P0-001"}), "P0-001")

    def test_effective_tracker_never_revives_terminal_planning_history(self):
        tracker = {
            "tasks": [
                {"id": "P0-001", "status": "superseded", "owner": None, "blocked_reason": None},
                {"id": "P0-002", "status": "removed", "owner": None, "blocked_reason": None},
            ]
        }
        state = {
            "tasks": {
                "P0-001": {"status": "in_progress", "owner": "worker", "blocked_reason": None},
                "P0-002": {"status": "pending", "owner": None, "blocked_reason": None},
            }
        }
        effective = task_scheduler.effective_tracker(tracker, state)
        self.assertEqual([task["status"] for task in effective["tasks"]], ["superseded", "removed"])

    def test_retire_completed_retires_all_terminal_tracker_states(self):
        tracker = {"tasks": [{"id": "P0-001", "status": "superseded"}, {"id": "P0-002", "status": "removed"}]}
        state = {
            "schema_version": 1,
            "tasks": {
                "P0-001": {"status": "in_progress", "owner": "a"},
                "P0-002": {"status": "blocked", "owner": "b"},
            },
        }
        self.assertTrue(task_scheduler.retire_completed(tracker, state))
        self.assertEqual(state["tasks"]["P0-001"]["status"], "superseded")
        self.assertEqual(state["tasks"]["P0-002"]["status"], "removed")
        self.assertIsNone(state["tasks"]["P0-001"]["owner"])
        self.assertIsNone(state["tasks"]["P0-002"]["owner"])

    def test_remote_workplan_dir_uses_current_generation(self):
        with patch.object(task_scheduler, "git_show_optional", return_value="v2"):
            self.assertEqual(task_scheduler.remote_workplan_dir(Path("/repo"), "origin", "main"), "docs/workplan/v2")

    def test_remote_workplan_dir_falls_back_to_legacy_root(self):
        with patch.object(task_scheduler, "git_show_optional", return_value=None):
            self.assertEqual(task_scheduler.remote_workplan_dir(Path("/repo"), "origin", "main"), "docs/workplan")

    def test_prompt_uses_active_workplan_directory(self):
        template = Template("Read $workplan_dir/$brief for $id")
        task = {"id": "P1-050", "title": "Runs", "brief": "phase-1/p1-050.md"}
        prompt = task_scheduler.prompt_for(template, task, "docs/workplan/v2")
        self.assertEqual(prompt, "Read docs/workplan/v2/phase-1/p1-050.md for P1-050")

    def test_task_url_survives_noise_on_dispatch_output(self):
        url = "https://chatgpt.com/codex/tasks/task_e_123"
        self.assertEqual(task_scheduler.task_url_from(f"warning\n{url}\n"), url)
        self.assertIsNone(task_scheduler.task_url_from("not signed in"))

    def test_failure_detail_keeps_captured_stderr(self):
        error = subprocess.CalledProcessError(1, ["codex"], stderr="environment not found\n")
        self.assertIn("environment not found", task_scheduler.failure_detail(error))

    def test_repo_root_requires_explicit_state_file(self):
        argv = ["task_scheduler.py", "--env", "environment", "--repo-root", "/opt/code-winch"]
        with patch.object(sys, "argv", argv), redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit):
                task_scheduler.parse_args()
        with patch.object(sys, "argv", [*argv, "--state-file", "/var/lib/state.json"]):
            self.assertEqual(task_scheduler.parse_args().repo_root, Path("/opt/code-winch"))


if __name__ == "__main__":
    unittest.main()
