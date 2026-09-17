import copy
import json
import subprocess
import unittest
from pathlib import Path
from unittest.mock import patch

from control_panel.integrations import repository
from helpers import GitRepositoryFixture, git, pushes, TRACKER


class TaskBranchTests(GitRepositoryFixture, unittest.TestCase):
    def start(self, clone: Path) -> bool:
        repository.ensure_task_branch(clone, "origin", "main", "P0-001")
        return repository.push_opening_commit(clone, "origin", "P0-001")

    def test_repeated_calls_create_the_branch_and_opening_commit_once(self):
        clone = self.clone("worker")
        with patch.object(repository, "run", wraps=repository.run) as spy:
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
        with patch.object(repository, "run", wraps=repository.run) as spy:
            self.assertFalse(self.start(self.clone("second")))
        self.assertEqual(pushes(spy), [])
        self.assertEqual(git(self.origin, "rev-parse", "task/P0-001"), tip)

    def test_the_committed_tracker_is_save_trackers_output_byte_for_byte(self):
        self.start(self.clone("worker"))
        tracker = copy.deepcopy(TRACKER)
        tracker["tasks"][0]["status"] = "in_progress"
        expected = self.root / "expected.json"
        repository.save_tracker(expected, tracker)
        committed = subprocess.run(
            ("git", "show", f"task/P0-001:{repository.TRACKER}"),
            cwd=self.origin,
            check=True,
            capture_output=True,
        ).stdout
        self.assertEqual(committed, expected.read_bytes())

    def test_a_failed_push_still_removes_the_worktree(self):
        clone = self.clone("worker")
        repository.ensure_task_branch(clone, "origin", "main", "P0-001")
        real_run = repository.run

        def rejecting_push(*command, cwd, capture=True):
            if command[:2] == ("git", "push"):
                raise subprocess.CalledProcessError(1, command, stderr="rejected")
            return real_run(*command, cwd=cwd, capture=capture)

        with patch.object(repository, "run", side_effect=rejecting_push):
            with self.assertRaises(subprocess.CalledProcessError):
                repository.push_opening_commit(clone, "origin", "P0-001")
        self.assertEqual(len(git(clone, "worktree", "list").splitlines()), 1)
        self.assertEqual(self.commits_ahead(), "0")

    def test_a_stale_clone_reads_and_branches_from_the_current_origin_main(self):
        stale = self.clone("stale")
        other = self.clone("other")
        tracker = copy.deepcopy(TRACKER)
        tracker["tasks"][0]["title"] = "Renamed upstream"
        repository.save_tracker(other / repository.TRACKER, tracker)
        git(other, "commit", "--quiet", "-am", "rename")
        git(other, "push", "--quiet", "origin", "HEAD:main")

        self.assertEqual(repository.pull_main(stale)[0], tracker)
        self.start(stale)
        self.assertEqual(
            git(self.origin, "rev-parse", "task/P0-001^"), git(self.origin, "rev-parse", "main")
        )


class PullRequestTests(unittest.TestCase):
    def test_open_pull_requests_into_task_branch_include_head(self):
        pulls = [{"number": 9, "url": "https://example/pr/9", "headRefOid": "0123abcd"}]
        with patch.object(repository, "run", return_value=json.dumps(pulls)) as run:
            from pathlib import Path
            clone = Path("/checkout")
            self.assertEqual(repository.task_pull_requests(clone, "P0-001"), pulls)
            run.assert_called_once_with(
                "gh", "pr", "list", "--base", "task/P0-001", "--state", "open",
                "--json", "number,url,headRefOid", cwd=clone)
