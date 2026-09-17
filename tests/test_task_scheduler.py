import copy
import importlib.util
import io
import json
import os
import subprocess
import sys
import tempfile
import unittest
from contextlib import redirect_stderr
from pathlib import Path
from unittest.mock import patch


REPO_ROOT = Path(__file__).parents[1]
SPEC = importlib.util.spec_from_file_location(
    "task_scheduler", REPO_ROOT / "scripts/task_scheduler.py"
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
                    "task_url": "ignored",
                }
            }
        }
        effective = task_scheduler.effective_tracker(tracker, state)
        self.assertEqual(effective["tasks"][0]["title"], "Upstream")
        self.assertEqual(effective["tasks"][0]["status"], "in_progress")
        self.assertEqual(effective["tasks"][0]["owner"], "worker")
        self.assertNotIn("task_url", effective["tasks"][0])

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

    def test_record_completions_keeps_the_dispatched_task_url(self):
        state = {
            "schema_version": 1,
            "tasks": {"P0-001": {"status": "in_progress", "task_url": "https://example/task"}},
        }
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
        self.assertEqual(state["tasks"]["P0-001"]["task_url"], "https://example/task")
        self.assertEqual(state["tasks"]["P0-001"]["pull_request"], "https://example/pr/1")

    def test_retire_completed_keeps_the_entry_the_tracker_caught_up_with(self):
        tracker = {"tasks": [{"id": "P0-001", "status": "completed"}]}
        state = {
            "schema_version": 1,
            "tasks": {
                "P0-001": {
                    "status": "in_progress",
                    "owner": "worker",
                    "pull_request": "https://example/pr/1",
                }
            },
        }
        self.assertTrue(task_scheduler.retire_completed(tracker, state))
        entry = state["tasks"]["P0-001"]
        self.assertEqual(entry["status"], "completed")
        self.assertIsNone(entry["owner"])
        self.assertEqual(entry["pull_request"], "https://example/pr/1")
        self.assertFalse(task_scheduler.retire_completed(tracker, state))

    def test_task_url_survives_noise_on_the_dispatch_output(self):
        url = "https://chatgpt.com/codex/tasks/task_e_6a7c22f3f85c83209a1c137b454a5ed9"
        self.assertEqual(task_scheduler.task_url_from(url), url)
        self.assertEqual(task_scheduler.task_url_from(f"warning: stale login\n{url}\n"), url)
        self.assertIsNone(task_scheduler.task_url_from("Not signed in. Please run 'codex login'."))

    def test_failure_detail_keeps_the_captured_stderr(self):
        error = subprocess.CalledProcessError(1, ["codex"], stderr="environment not found\n")
        self.assertIn("environment not found", task_scheduler.failure_detail(error))
        self.assertEqual(task_scheduler.failure_detail(OSError("no codex")), "no codex")

    def test_stage_prompts_render_with_their_fields(self):
        task = {"id": "P0-001", "title": "Upstream", "brief": "phase-0/P0-001-upstream.md"}
        pull = {"pr_url": "https://example/pr/1", "head": "0123abcd"}
        stages = {
            task_scheduler.PROMPT_TEMPLATE: task,
            Path("scripts/task-refine-prompt.md"): task,
            Path("scripts/task-audit-prompt.md"): {**task, **pull},
        }
        for path, fields in stages.items():
            with self.subTest(path=path):
                template = task_scheduler.load_prompt_template(REPO_ROOT, path)
                prompt = template.substitute(fields)
                for value in fields.values():
                    self.assertIn(value, prompt)
                self.assertNotIn("$", prompt)


TASK = {"id": "P0-001", "title": "Upstream"}
TRACKER = {
    "schema_version": 1,
    "tasks": [{**TASK, "status": "pending", "owner": None, "blocked_reason": None}],
}


def git(cwd: Path, *arguments: str) -> str:
    return task_scheduler.run("git", *arguments, cwd=cwd)


def pushes(spy) -> list[tuple[str, ...]]:
    return [call.args for call in spy.call_args_list if call.args[:2] == ("git", "push")]


class TaskBranchTests(unittest.TestCase):
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.root = Path(scratch.name)
        isolated = patch.dict(
            os.environ, {"GIT_CONFIG_GLOBAL": os.devnull, "GIT_CONFIG_NOSYSTEM": "1"}
        )
        isolated.start()
        self.addCleanup(isolated.stop)

        self.origin = self.root / "origin.git"
        git(self.root, "init", "--quiet", "--bare", "--initial-branch=main", str(self.origin))
        seed = self.clone("seed")
        (seed / task_scheduler.TRACKER).parent.mkdir(parents=True)
        # Not save_tracker's layout, so only a tracker written by save_tracker matches it.
        (seed / task_scheduler.TRACKER).write_text(json.dumps(TRACKER, indent=4))
        git(seed, "add", ".")
        git(seed, "commit", "--quiet", "-m", "seed")
        git(seed, "push", "--quiet", "origin", "HEAD:main")

    def clone(self, name: str) -> Path:
        path = self.root / name
        git(self.root, "clone", "--quiet", str(self.origin), str(path))
        git(path, "config", "user.name", "Scheduler Test")
        git(path, "config", "user.email", "scheduler@example.com")
        return path

    def start(self, clone: Path) -> bool:
        task_scheduler.ensure_task_branch(clone, "origin", "main", "P0-001")
        return task_scheduler.push_opening_commit(clone, "origin", "P0-001")

    def commits_ahead(self) -> str:
        return git(self.origin, "rev-list", "--count", "main..task/P0-001")

    def test_repeated_calls_create_the_branch_and_opening_commit_once(self):
        clone = self.clone("worker")
        with patch.object(task_scheduler, "run", wraps=task_scheduler.run) as spy:
            self.assertTrue(self.start(clone))
            self.assertEqual(len(pushes(spy)), 2)
            spy.reset_mock()
            self.assertFalse(self.start(clone))
            self.assertEqual(pushes(spy), [])
        self.assertEqual(self.commits_ahead(), "1")
        self.assertEqual(len(git(clone, "worktree", "list").splitlines()), 1)

    def test_a_second_clone_skips_the_opening_commit_already_on_origin(self):
        self.start(self.clone("first"))
        tip = git(self.origin, "rev-parse", "task/P0-001")
        with patch.object(task_scheduler, "run", wraps=task_scheduler.run) as spy:
            self.assertFalse(self.start(self.clone("second")))
        self.assertEqual(pushes(spy), [])
        self.assertEqual(git(self.origin, "rev-parse", "task/P0-001"), tip)

    def test_the_committed_tracker_is_save_trackers_output_byte_for_byte(self):
        self.start(self.clone("worker"))
        tracker = copy.deepcopy(TRACKER)
        tracker["tasks"][0]["status"] = "in_progress"
        expected = self.root / "expected.json"
        task_scheduler.save_tracker(expected, tracker)
        committed = subprocess.run(
            ("git", "show", f"task/P0-001:{task_scheduler.TRACKER}"),
            cwd=self.origin,
            check=True,
            capture_output=True,
        ).stdout
        self.assertEqual(committed, expected.read_bytes())

    def test_a_failed_push_still_removes_the_worktree(self):
        clone = self.clone("worker")
        task_scheduler.ensure_task_branch(clone, "origin", "main", "P0-001")
        real_run = task_scheduler.run

        def rejecting_push(*command, cwd, capture=True):
            if command[:2] == ("git", "push"):
                raise subprocess.CalledProcessError(1, command, stderr="rejected")
            return real_run(*command, cwd=cwd, capture=capture)

        with patch.object(task_scheduler, "run", side_effect=rejecting_push):
            with self.assertRaises(subprocess.CalledProcessError):
                task_scheduler.push_opening_commit(clone, "origin", "P0-001")
        self.assertEqual(len(git(clone, "worktree", "list").splitlines()), 1)
        self.assertEqual(self.commits_ahead(), "0")

    def test_a_stale_clone_reads_and_branches_from_the_current_origin_main(self):
        stale = self.clone("stale")
        (stale / task_scheduler.TRACKER).write_text("{}")
        other = self.clone("other")
        tracker = copy.deepcopy(TRACKER)
        tracker["tasks"][0]["title"] = "Renamed upstream"
        task_scheduler.save_tracker(other / task_scheduler.TRACKER, tracker)
        git(other, "commit", "--quiet", "-am", "rename")
        git(other, "push", "--quiet", "origin", "HEAD:main")

        self.assertEqual(task_scheduler.load_tracker(stale, "origin", "main"), tracker)
        self.start(stale)
        self.assertEqual(
            git(self.origin, "rev-parse", "task/P0-001^"), git(self.origin, "rev-parse", "main")
        )


GH_STUB = """
import json, os, sys

with open(os.environ["GH_STUB_CALLS"], "a") as calls:
    calls.write(json.dumps(sys.argv[1:]) + "\\n")
if sys.argv[1:3] == ["pr", "list"]:
    print(open(os.environ["GH_STUB_LIST"]).read())
elif sys.argv[1:3] == ["pr", "create"]:
    print("https://github.com/owner/repo/pull/8")
"""


class PullRequestTests(unittest.TestCase):
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.root = Path(scratch.name)
        stub = self.root / "gh"
        stub.write_text(f"#!{sys.executable}{GH_STUB}")
        stub.chmod(0o755)
        self.calls_file = self.root / "calls.jsonl"
        self.listed = self.root / "listed.json"
        environment = patch.dict(
            os.environ,
            {
                "PATH": f"{self.root}{os.pathsep}{os.environ['PATH']}",
                "GH_STUB_CALLS": str(self.calls_file),
                "GH_STUB_LIST": str(self.listed),
            },
        )
        environment.start()
        self.addCleanup(environment.stop)

    def calls(self) -> list[list[str]]:
        return [json.loads(line) for line in self.calls_file.read_text().splitlines()]

    def test_a_draft_is_created_when_no_pull_request_is_open(self):
        self.listed.write_text("[]")
        url = task_scheduler.ensure_draft_pull_request(self.root, "main", TASK)
        self.assertEqual(url, "https://github.com/owner/repo/pull/8")
        self.assertEqual(
            self.calls(),
            [
                ["pr", "list", "--head", "task/P0-001", "--base", "main"]
                + ["--state", "open", "--json", "number,url"],
                ["pr", "create", "--draft", "--base", "main", "--head", "task/P0-001"]
                + ["--title", "P0-001: Upstream", "--body", "Task: P0-001"],
            ],
        )

    def test_an_open_pull_request_is_reused(self):
        self.listed.write_text(
            json.dumps([{"number": 7, "url": "https://github.com/owner/repo/pull/7"}])
        )
        url = task_scheduler.ensure_draft_pull_request(self.root, "main", TASK)
        self.assertEqual(url, "https://github.com/owner/repo/pull/7")
        self.assertEqual([call[:2] for call in self.calls()], [["pr", "list"]])

    def test_open_pull_requests_into_the_task_branch_carry_their_head(self):
        pulls = [{"number": 9, "url": "https://example/pr/9", "headRefOid": "0123abcd"}]
        self.listed.write_text(json.dumps(pulls))
        self.assertEqual(task_scheduler.task_pull_requests(self.root, "P0-001"), pulls)
        self.assertEqual(
            self.calls(),
            [
                ["pr", "list", "--base", "task/P0-001"]
                + ["--state", "open", "--json", "number,url,headRefOid"]
            ],
        )


if __name__ == "__main__":
    unittest.main()
