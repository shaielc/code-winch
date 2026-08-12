import argparse
import importlib.util
import io
import subprocess
import sys
import unittest
from contextlib import redirect_stderr, redirect_stdout
from pathlib import Path
from unittest.mock import patch

SCRIPTS = Path(__file__).parents[1] / "scripts"
sys.path.insert(0, str(SCRIPTS))

SPEC = importlib.util.spec_from_file_location(
    "stamp_task_completion", SCRIPTS / "stamp_task_completion.py"
)
stamp = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(stamp)

ARGS = argparse.Namespace(remote="origin", branch="main")
REPO = Path("/repo")


def pull_request(head_repo: str = "owner/repo") -> dict:
    return {
        "head": {"ref": "work", "repo": {"full_name": head_repo}},
        "base": {"repo": {"full_name": "owner/repo"}},
    }


class CatchUpWithBaseTests(unittest.TestCase):
    def test_clean_merge_is_pushed_but_left_unstamped(self):
        with (
            patch.object(stamp, "configure_identity"),
            patch.object(stamp, "merge_base_branch", return_value=True),
            patch.object(stamp, "push_head") as push,
            redirect_stdout(io.StringIO()) as out,
        ):
            code = stamp.catch_up_with_base(REPO, ARGS, pull_request(), {}, "P0-001")
        self.assertEqual(code, 0)
        push.assert_called_once_with(REPO, "origin", "work")
        self.assertIn("approve again to stamp P0-001", out.getvalue())

    def test_conflicting_merge_is_reported_and_never_pushed(self):
        with (
            patch.object(stamp, "configure_identity"),
            patch.object(stamp, "merge_base_branch", return_value=False),
            patch.object(stamp, "push_head") as push,
            redirect_stderr(io.StringIO()) as err,
        ):
            code = stamp.catch_up_with_base(REPO, ARGS, pull_request(), {}, "P0-001")
        self.assertEqual(code, 1)
        push.assert_not_called()
        self.assertIn("conflicts with this branch", err.getvalue())

    def test_fork_branches_are_refused_before_the_merge(self):
        with (
            patch.object(stamp, "merge_base_branch") as merge,
            patch.object(stamp, "push_head") as push,
            redirect_stderr(io.StringIO()) as err,
        ):
            code = stamp.catch_up_with_base(
                REPO, ARGS, pull_request("contributor/repo"), {}, "P0-001"
            )
        self.assertEqual(code, 1)
        merge.assert_not_called()
        push.assert_not_called()
        self.assertIn("lives in a fork", err.getvalue())

    def test_a_rejected_push_fails_loudly(self):
        error = subprocess.CalledProcessError(1, ["git"], stderr="non-fast-forward\n")
        with (
            patch.object(stamp, "configure_identity"),
            patch.object(stamp, "merge_base_branch", return_value=True),
            patch.object(stamp, "push_head", side_effect=error),
            redirect_stderr(io.StringIO()) as err,
        ):
            code = stamp.catch_up_with_base(REPO, ARGS, pull_request(), {}, "P0-001")
        self.assertEqual(code, 1)
        self.assertIn("non-fast-forward", err.getvalue())


class MergeBaseBranchTests(unittest.TestCase):
    def test_a_failed_merge_leaves_no_half_merged_tree(self):
        results = [subprocess.CompletedProcess([], 1), subprocess.CompletedProcess([], 0)]
        with patch.object(stamp.subprocess, "run", side_effect=results) as run:
            self.assertFalse(stamp.merge_base_branch(REPO, "origin", "main"))
        self.assertEqual(run.call_args_list[0].args[0][:2], ("git", "merge"))
        self.assertEqual(run.call_args_list[1].args[0], ("git", "merge", "--abort"))

    def test_a_clean_merge_is_kept(self):
        with patch.object(stamp.subprocess, "run", return_value=subprocess.CompletedProcess([], 0)) as run:
            self.assertTrue(stamp.merge_base_branch(REPO, "origin", "main"))
        run.assert_called_once()


if __name__ == "__main__":
    unittest.main()
