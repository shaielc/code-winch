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

    def test_qualified_pr_identity_must_match_generation(self):
        pull = {"title": "Implement v2/P1-050", "body": "Task: v2/P1-050", "head": {"ref": "work"}}
        self.assertEqual(task_scheduler.task_identity_for_pr(pull, {"P1-050"}, "v2"), "P1-050")
        self.assertIsNone(task_scheduler.task_identity_for_pr(pull, {"P1-050"}, "v3"))
        bare = {"title": "Implement P1-050", "body": "Task: P1-050", "head": {"ref": "work"}}
        self.assertIsNone(task_scheduler.task_identity_for_pr(bare, {"P1-050"}, "v2"))

    def test_stale_legacy_state_cannot_override_reused_v2_id(self):
        tracker = {
            "_workplan_dir": "docs/workplan/v2", "_generation": "v2",
            "tasks": [{"id": "P1-050", "status": "pending", "owner": None, "blocked_reason": None}],
        }
        state = {
            "schema_version": 1,
            "tasks": {
                "P1-050": {"status": "in_progress", "owner": "old-worker", "blocked_reason": None},
                "v1/P1-050": {"status": "completed", "owner": None, "blocked_reason": None},
            },
        }
        effective = task_scheduler.effective_tracker(tracker, state)
        self.assertEqual(effective["tasks"][0]["status"], "pending")
        self.assertIsNone(effective["tasks"][0]["owner"])

    def test_v2_state_key_does_not_leak_into_v3(self):
        state = {"schema_version": 1, "tasks": {"v2/P1-050": {"status": "in_progress", "owner": "worker", "blocked_reason": None}}}
        v2 = {"_workplan_dir": "docs/workplan/v2", "_generation": "v2", "tasks": [{"id": "P1-050", "status": "pending", "owner": None, "blocked_reason": None}]}
        v3 = {"_workplan_dir": "docs/workplan/v3", "_generation": "v3", "tasks": [{"id": "P1-050", "status": "pending", "owner": None, "blocked_reason": None}]}
        self.assertEqual(task_scheduler.effective_tracker(v2, state)["tasks"][0]["status"], "in_progress")
        self.assertEqual(task_scheduler.effective_tracker(v3, state)["tasks"][0]["status"], "pending")

    def test_record_completion_persists_qualified_identity(self):
        state = {"schema_version": 1, "tasks": {}}
        pull = {
            "title": "v2/P1-050", "body": "Task: v2/P1-050", "head": {"ref": "work"},
            "html_url": "https://example/pr/1", "merged_at": "2026-01-01T00:00:00Z",
        }
        self.assertTrue(task_scheduler.record_completions(state, [pull], {"P1-050"}, "v2"))
        self.assertIn("v2/P1-050", state["tasks"])
        self.assertNotIn("P1-050", state["tasks"])
        self.assertEqual(state["tasks"]["v2/P1-050"]["generation"], "v2")

    def test_remote_workplan_dir_uses_current_generation(self):
        with patch.object(task_scheduler, "git_show_optional", return_value="v2"):
            self.assertEqual(task_scheduler.remote_workplan_dir(Path("/repo"), "origin", "main"), "docs/workplan/v2")

    def test_prompt_uses_qualified_task_identity(self):
        template = Template("Read $workplan_dir/$brief for $task_key")
        task = {"id": "P1-050", "title": "Runs", "brief": "phase-1/p1-050.md"}
        prompt = task_scheduler.prompt_for(template, task, "docs/workplan/v2", "v2")
        self.assertEqual(prompt, "Read docs/workplan/v2/phase-1/p1-050.md for v2/P1-050")

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
