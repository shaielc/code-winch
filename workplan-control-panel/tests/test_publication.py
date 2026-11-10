"""Publishing a finished Codex stage as one commit and one pull request into task/<ID>.

Codex and gh are replaced at the panel's integration seam; git runs for real against a local
bare origin. What this does not prove: the real `codex cloud list`/`diff` output shapes, and
that GitHub accepts the push and `gh pr create --fill` with the App's credential.
"""

import json
import subprocess
import unittest
from collections.abc import Callable
from pathlib import Path
from unittest.mock import patch

from control_panel import state as task_state
from control_panel.integrations import codex, process, repository
from control_panel.task_scheduler import TaskError, TaskScheduler
from control_panel.ui import render
from helpers import GitRepositoryFixture, TRACKER, git

BRIEF = "docs/workplan/phase-0/first.md"
CODEX_URL = "https://chatgpt.com/codex/cloud/tasks/task_e_1"
PULL = {"number": 7, "url": "https://example/pr/7"}


class PublicationTests(GitRepositoryFixture, unittest.TestCase):
    def setUp(self):
        super().setUp()
        seed = self.root / "seed"
        tracker = json.loads(json.dumps(TRACKER))
        tracker["tasks"][0].update(depends_on=[], brief="phase-0/first.md")
        repository.save_tracker(seed / repository.TRACKER, tracker)
        (seed / "README.md").write_text("readme\n")
        # The brief exists at base, as it does once derived, so Refine edits a tracked file.
        (seed / BRIEF).parent.mkdir(parents=True, exist_ok=True)
        (seed / BRIEF).write_text("derived brief\n")
        git(seed, "add", "-A")
        git(seed, "commit", "-m", "task graph")
        git(seed, "push", "origin", "main")
        self.panel = TaskScheduler(self.clone("panel"), self.root / "state.json",
                                   self.root / "tracker.json", "test-env", 1,
                                   "https://api.example/fire", "routine-token")
        for name in ("task_pull_request", "stage_pull_request"):
            patcher = patch.object(repository, name, return_value=None)
            patcher.start()
            self.addCleanup(patcher.stop)
        self.panel.sync()

    # Fixtures

    def submit(self, stage: str) -> dict:
        with patch.object(codex, "submit", return_value=CODEX_URL):
            return self.panel.stage("P0-001", stage)

    def attempt(self, stage: str) -> dict:
        return self.panel.snapshot()["state"]["tasks"]["P0-001"]["stages"][stage][-1]

    def make_diff(self, change: Callable[[Path], None]) -> str:
        """A unified diff produced by git from the stage's base, as Codex would return it."""
        work = self.clone(f"agent-{len(list(self.root.iterdir()))}")
        git(work, "checkout", "--quiet", "--detach", git(self.origin, "rev-parse", "task/P0-001"))
        change(work)
        git(work, "add", "-A")
        return process.run("git", "diff", "--cached", cwd=work, strip=False)

    def brief_diff(self) -> str:
        return self.make_diff(lambda work: (work / BRIEF).write_text("refined brief\n"))

    def tips(self) -> dict[str, str]:
        refs = git(self.origin, "for-each-ref", "--format=%(refname:short) %(objectname)",
                   "refs/heads/refine", "refs/heads/implement")
        return dict(line.split() for line in refs.splitlines())

    def finished(self, patch_text: str, status: str = "ready"):
        return (patch.object(codex, "find_task", return_value={"id": "task_e_1", "status": status}),
                patch.object(codex, "diff", return_value=patch_text))

    def publish(self, patch_text: str, status: str = "ready", pull=PULL):
        find, diff = self.finished(patch_text, status)
        with find, diff, patch.object(repository, "create_stage_pull_request",
                                      return_value=pull) as create:
            self.panel.publish_finished()
        return create

    def assert_unpublished(self, stage: str, status: str, reason: str, create) -> None:
        attempt = self.attempt(stage)
        self.assertEqual(attempt["status"], status)
        self.assertIn(reason, attempt["reason"])
        self.assertNotIn("published_tip", attempt)
        self.assertEqual(self.tips()[f"{stage}/P0-001-1"], attempt["base"])
        create.assert_not_called()

    # Outcomes that publish nothing

    def test_a_pending_task_is_left_running_and_untouched(self):
        self.submit("implement")
        create = self.publish(self.brief_diff(), status="pending")
        self.assertEqual(self.attempt("implement")["status"], "submitted")
        create.assert_not_called()

    def test_a_codex_error_is_recorded_and_pushes_nothing(self):
        self.submit("implement")
        self.assert_unpublished("implement", "failed", "Codex reported an error",
                                self.publish(self.brief_diff(), status="error"))

    def test_an_empty_diff_is_recorded_as_no_changes(self):
        self.submit("implement")
        self.assert_unpublished("implement", "no-changes", "no changes",
                                self.publish("\n"))

    def test_a_diff_that_does_not_apply_pushes_nothing(self):
        self.submit("implement")
        broken = self.make_diff(lambda work: (work / "README.md").write_text("changed\n"))
        broken = broken.replace("-readme", "-something else")
        self.assert_unpublished("implement", "failed", "did not apply",
                                self.publish(broken))
        self.assertEqual(len(git(self.panel.clone, "worktree", "list").splitlines()), 1)

    def test_refine_outside_its_brief_pushes_nothing_for_new_deleted_or_renamed_files(self):
        # Each edits the brief as well, so a check that saw only tracked edits would pass it.
        def new_file(work):
            (work / BRIEF).write_text("refined brief\n")
            (work / "evil.sh").write_text("echo owned\n")

        def deleted(work):
            (work / BRIEF).write_text("refined brief\n")
            (work / "README.md").unlink()

        def renamed(work):
            (work / BRIEF).write_text("refined brief\n")
            (work / "README.md").rename(work / "docs/workplan/README.md")

        def only_new(work):
            (work / "evil.sh").write_text("echo owned\n")

        def tracker_edit(work):
            # Pointing the tracker at another brief does not widen what Refine may change.
            tracker = json.loads((work / repository.TRACKER).read_text())
            tracker["tasks"][0]["brief"] = "../../evil.sh"
            (work / repository.TRACKER).write_text(json.dumps(tracker))
            (work / "evil.sh").write_text("echo owned\n")

        for change in (new_file, deleted, renamed, only_new, tracker_edit):
            # Before the fix, new_file reached origin: `git diff --name-only` omits untracked files.
            with self.subTest(change.__name__):
                text = self.make_diff(change)
                self.assertTrue(text.strip())
                with self.assertRaises(repository.PublishConflict) as refused:
                    repository.publish_stage(self.panel.clone, "P0-001", "Upstream", "refine",
                                             "refine/P0-001-unused",
                                             git(self.origin, "rev-parse", "task/P0-001"), text)
                self.assertIn("outside its task brief", str(refused.exception))
        self.assertNotIn("refine/P0-001-unused", self.tips())
        self.assertEqual(len(git(self.panel.clone, "worktree", "list").splitlines()), 1)

        self.submit("refine")
        self.assert_unpublished("refine", "failed", "outside its task brief",
                                self.publish(self.make_diff(new_file)))

    def test_a_rejected_push_is_recorded_and_leaves_the_branch_as_it_was(self):
        self.submit("implement")
        other = self.clone("other")
        git(other, "checkout", "--quiet", "-b", "moved", "origin/implement/P0-001-1")
        (other / "other.txt").write_text("someone else\n")
        git(other, "add", "-A")
        git(other, "commit", "--quiet", "-m", "moved meanwhile")
        git(other, "push", "--quiet", "origin", "HEAD:refs/heads/implement/P0-001-1")
        moved = git(other, "rev-parse", "HEAD")

        create = self.publish(self.brief_diff())
        attempt = self.attempt("implement")
        self.assertEqual(attempt["status"], "failed")
        self.assertNotIn("published_tip", attempt)
        self.assertEqual(self.tips()["implement/P0-001-1"], moved)
        create.assert_not_called()

    # Publishing

    def test_refine_publishes_one_panel_commit_and_a_pull_request(self):
        self.submit("refine")
        base = self.attempt("refine")["base"]
        create = self.publish(self.brief_diff())
        attempt = self.attempt("refine")
        self.assertEqual((attempt["status"], attempt["pull_request_url"]), ("published", PULL["url"]))
        tip = self.tips()["refine/P0-001-1"]
        self.assertEqual(attempt["published_tip"], tip)
        self.assertEqual(git(self.origin, "rev-list", "--count", f"{base}..{tip}"), "1")
        self.assertEqual(git(self.origin, "log", "-1", "--format=%s%n%b", tip).strip(),
                         "P0-001: Refine Upstream\nTask: P0-001")
        self.assertEqual(git(self.origin, "diff", "--name-only", base, tip), BRIEF)
        create.assert_called_once_with(self.panel.clone, "P0-001", "refine/P0-001-1")

        # The next poll and an explicit retry find it published and commit nothing more.
        create = self.publish(self.brief_diff())
        create.assert_not_called()
        find, diff = self.finished(self.brief_diff())
        with find as found, diff, patch.object(repository, "create_stage_pull_request",
                                               return_value=PULL):
            self.assertEqual(self.panel.retry_publish("P0-001", "refine")["status"], "published")
        found.assert_not_called()
        self.assertEqual(self.tips()["refine/P0-001-1"], tip)

    def test_a_gh_failure_after_the_push_retries_without_another_commit(self):
        self.submit("implement")
        failing = subprocess.CalledProcessError(1, ["gh", "pr", "create"], stderr="HTTP 502")
        find, diff = self.finished(self.brief_diff())
        with find, diff, patch.object(repository, "create_stage_pull_request",
                                      side_effect=failing):
            self.panel.publish_finished()
        attempt = self.attempt("implement")
        tip = self.tips()["implement/P0-001-1"]
        self.assertEqual((attempt["status"], attempt["published_tip"]), ("failed", tip))

        find, diff = self.finished(self.brief_diff())
        with find as found, diff as diffed, \
                patch.object(repository, "publish_stage") as publish, \
                patch.object(repository, "create_stage_pull_request", return_value=PULL) as create:
            for _ in range(2):
                self.assertEqual(self.panel.retry_publish("P0-001", "implement")["status"],
                                 "published")
        publish.assert_not_called()
        found.assert_not_called()
        diffed.assert_not_called()
        self.assertEqual(create.call_count, 2)
        self.assertEqual(self.tips()["implement/P0-001-1"], tip)
        self.assertEqual(self.attempt("implement")["pull_request_number"], 7)

    def test_creating_the_stage_pull_request_reuses_an_open_one(self):
        with patch.object(repository, "stage_pull_request", return_value=PULL), \
                patch.object(repository, "github_run") as gh:
            self.assertEqual(repository.create_stage_pull_request(
                self.panel.clone, "P0-001", "refine/P0-001-1"), PULL)
        gh.assert_not_called()

        with patch.object(repository, "stage_pull_request", side_effect=[None, PULL]), \
                patch.object(repository, "github_run", return_value=PULL["url"]) as gh:
            repository.create_stage_pull_request(self.panel.clone, "P0-001", "refine/P0-001-1")
        self.assertEqual(gh.call_args.args, ("gh", "pr", "create", "--base", "task/P0-001",
                                             "--head", "refine/P0-001-1", "--fill"))

    def test_retry_refuses_a_stage_with_no_attempt(self):
        with self.assertRaises(TaskError) as refused:
            self.panel.retry_publish("P0-001", "implement")
        self.assertEqual(refused.exception.status, 409)

    # Isolation

    def test_no_subprocess_runs_under_the_state_lock_and_codex_never_sees_gh_token(self):
        self.submit("implement")
        real = subprocess.run
        calls = []
        codex_output = {
            "list": json.dumps({"tasks": [{"id": "task_e_1", "status": "ready"}]}),
            "diff": self.brief_diff(),
        }

        def guarded(command, *args, **kwargs):
            """Record whether the state lock is held, then run git or answer for codex."""
            try:
                task_state.acquire_lock(self.panel.state_file).close()
                held = False
            except RuntimeError:
                held = True
            calls.append((tuple(command), held, kwargs.get("env") or {}))
            if command[0] == "codex":
                return subprocess.CompletedProcess(command, 0, codex_output[command[2]], "")
            return real(command, *args, **kwargs)

        with patch.dict("os.environ", {"GH_TOKEN": "ambient-secret"}), \
                patch.object(process.subprocess, "run", side_effect=guarded), \
                patch.object(repository, "create_stage_pull_request", return_value=PULL):
            self.panel.publish_finished()
        self.assertEqual(self.attempt("implement")["status"], "published")
        self.assertTrue(any(command[0] == "codex" for command, _, _ in calls))
        self.assertTrue(any(command[:2] == ("git", "push") for command, _, _ in calls))
        self.assertEqual([command for command, held, _ in calls if held], [])
        for command, _, env in calls:
            if command[0] == "codex":
                self.assertNotIn("GH_TOKEN", env)

    # Display

    def test_the_stage_cell_shows_each_publication_state(self):
        self.submit("implement")
        cases = {
            "submitted": ({}, ">running</small>"),
            "published": ({"pull_request_url": PULL["url"], "reason": "Published"},
                          f'<a href="{PULL["url"]}" target="_blank" rel="noopener">published</a>'),
            "failed": ({"reason": "The Codex diff did not apply cleanly"},
                       ">failed: The Codex diff did not apply cleanly</small>"),
            "no-changes": ({"reason": "Codex finished with no changes"},
                           ">no-changes: Codex finished with no changes</small>"),
        }
        for status, (extra, expected) in cases.items():
            with self.subTest(status):
                with self.panel.locked():
                    state = task_state.load_state(self.panel.state_file)
                    attempt = state["tasks"]["P0-001"]["stages"]["implement"][-1]
                    for key in ("pull_request_url", "reason"):
                        attempt.pop(key, None)
                    attempt.update(extra, status=status)
                    task_state.write_json(self.panel.state_file, state)
                page = render(self.panel.tracker(), self.panel.snapshot()["state"], "", False)
                self.assertIn(expected, page)


if __name__ == "__main__":
    unittest.main()
