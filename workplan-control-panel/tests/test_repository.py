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

    def spies(self):
        """Watch both runners, since a push is authenticated and the commit before it is not."""
        return (patch.object(repository, "run", wraps=repository.run),
                patch.object(repository, "github_run", wraps=repository.github_run))

    def test_repeated_calls_create_the_branch_and_opening_commit_once(self):
        clone = self.clone("worker")
        local, authenticated = self.spies()
        with local as plain, authenticated as authorized:
            self.assertTrue(self.start(clone))
            self.assertEqual(len(pushes(plain, authorized)), 2)
            # Every push is a GitHub command, so none of them runs without a credential.
            self.assertEqual(pushes(plain), [])
            plain.reset_mock()
            authorized.reset_mock()
            self.assertFalse(self.start(clone))
            self.assertEqual(pushes(plain, authorized), [])
        self.assertEqual(self.commits_ahead(), "1")
        self.assertEqual(len(git(clone, "worktree", "list").splitlines()), 1)

    def test_a_second_clone_skips_the_opening_commit_already_on_origin(self):
        self.start(self.clone("first"))
        tip = git(self.origin, "rev-parse", "task/P0-001")
        local, authenticated = self.spies()
        with local as plain, authenticated as authorized:
            self.assertFalse(self.start(self.clone("second")))
        self.assertEqual(pushes(plain, authorized), [])
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
        real_run = repository.github_run

        def rejecting_push(*command, cwd, capture=True):
            if command[:2] == ("git", "push"):
                raise subprocess.CalledProcessError(1, command, stderr="rejected")
            return real_run(*command, cwd=cwd, capture=capture)

        with patch.object(repository, "github_run", side_effect=rejecting_push):
            with self.assertRaises(subprocess.CalledProcessError):
                repository.push_opening_commit(clone, "origin", "P0-001")
        self.assertEqual(len(git(clone, "worktree", "list").splitlines()), 1)
        self.assertEqual(self.commits_ahead(), "0")

    def test_the_panels_checkout_is_cloned_once_and_then_reused(self):
        checkout = self.root / "panel-checkout"
        repository.ensure_clone(checkout, str(self.origin))
        self.assertEqual(git(checkout, "branch", "--show-current"), "main")
        self.assertTrue((checkout / repository.TRACKER).exists())
        with patch.object(repository, "github_run") as runner:
            repository.ensure_clone(checkout, str(self.origin))
        runner.assert_not_called()

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
        with patch.object(repository, "github_run", return_value=json.dumps(pulls)) as run:
            from pathlib import Path
            clone = Path("/checkout")
            self.assertEqual(repository.task_pull_requests(clone, "P0-001"), pulls)
            run.assert_called_once_with(
                "gh", "pr", "list", "--base", "task/P0-001", "--state", "open",
                "--json", "number,url,headRefOid", cwd=clone)

    def test_creating_the_task_pull_request_names_the_branch_and_returns_its_url(self):
        clone = Path("/checkout")
        url = "https://github.com/owner/repo/pull/12"
        with patch.object(repository, "github_run", return_value="Creating pull request\n" + url) as run:
            self.assertEqual(repository.create_task_pull_request(clone, "P0-001", "First"), url)
        args = run.call_args.args
        self.assertEqual(args[:3], ("gh", "pr", "create"))
        self.assertIn("task/P0-001", args)
        self.assertIn("P0-001: First", args)
        self.assertIn("Task: P0-001", args)
        with patch.object(repository, "github_run", return_value="no url here"):
            with self.assertRaises(ValueError):
                repository.create_task_pull_request(clone, "P0-001", "First")

    def test_the_task_pull_request_into_main_prefers_open_then_merged(self):
        listed = [{"url": "https://example/pr/1", "state": "CLOSED"},
                  {"url": "https://example/pr/2", "state": "MERGED"},
                  {"url": "https://example/pr/3", "state": "OPEN"}]
        clone = Path("/checkout")
        with patch.object(repository, "github_run", return_value=json.dumps(listed)) as run:
            self.assertEqual(repository.task_pull_request(clone, "P0-001")["url"],
                             "https://example/pr/3")
            self.assertEqual(run.call_args.args[3:7], ("--head", "task/P0-001", "--base", "main"))
        with patch.object(repository, "github_run", return_value=json.dumps(listed[:2])):
            self.assertEqual(repository.task_pull_request(clone, "P0-001")["state"], "MERGED")
        # A closed pull request that never merged does not count as one.
        with patch.object(repository, "github_run", return_value=json.dumps(listed[:1])):
            self.assertIsNone(repository.task_pull_request(clone, "P0-001"))
