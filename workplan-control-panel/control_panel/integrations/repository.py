"""Maintain the dedicated main checkout and task branches; read task PRs."""

import json
import tempfile
from pathlib import Path
from typing import Any

from .process import run

TRACKER = Path("docs/workplan/tasks.json")


class CheckoutConflict(Exception):
    """The checkout cannot safely be updated automatically."""


def pull_main(clone: Path) -> tuple[dict[str, Any], str]:
    if run("git", "branch", "--show-current", cwd=clone) != "main":
        raise CheckoutConflict("The control-panel checkout must be on main")
    if run("git", "status", "--porcelain", cwd=clone):
        raise CheckoutConflict("The control-panel checkout must be clean")
    run("git", "pull", "--ff-only", "origin", "main", cwd=clone)
    base = run("git", "rev-parse", "HEAD", cwd=clone)
    if base != run("git", "rev-parse", "origin/main", cwd=clone):
        raise CheckoutConflict("The main checkout has local commits; reconcile it before syncing")
    return json.loads((clone / TRACKER).read_text()), base


def task_head(clone: Path, task_id: str) -> str:
    branch = task_branch(task_id)
    run("git", "fetch", "--quiet", "origin",
        f"refs/heads/{branch}:refs/remotes/origin/{branch}", cwd=clone)
    return run("git", "rev-parse", f"origin/{branch}", cwd=clone)


def save_tracker(path: Path, tracker: dict[str, Any]) -> None:
    path.write_text(json.dumps(tracker, indent=2) + "\n")


def task_branch(task_id: str) -> str:
    return f"task/{task_id}"


def ensure_task_branch(clone: Path, remote: str, base: str, task_id: str,
                       base_commit: str | None = None) -> None:
    """Create the task branch on the remote from the base branch unless it exists."""
    ref = f"refs/heads/{task_branch(task_id)}"
    if base_commit is None:
        run("git", "fetch", "--quiet", remote, base, cwd=clone)
    if not run("git", "ls-remote", "--heads", remote, ref, cwd=clone):
        source = base_commit or f"{remote}/{base}"
        run("git", "push", "--quiet", remote, f"{source}:{ref}", cwd=clone)
    run("git", "fetch", "--quiet", remote,
        f"{ref}:refs/remotes/{remote}/{task_branch(task_id)}", cwd=clone)


def push_opening_commit(clone: Path, remote: str, task_id: str) -> bool:
    """Mark the task in progress on its branch once, and say whether this call did.

    The commit is built in a temporary worktree so the clone's own checkout is never
    touched. It is also what puts the branch ahead of its base, which GitHub requires
    before it will open a pull request from it.
    """
    branch = task_branch(task_id)
    tracker = json.loads(run("git", "show", f"{remote}/{branch}:{TRACKER}", cwd=clone))
    task = next((item for item in tracker["tasks"] if item["id"] == task_id), None)
    if task is None:
        raise ValueError("Task branch does not contain its tracker entry")
    if task["status"] == "in_progress":
        return False

    task["status"] = "in_progress"
    with tempfile.TemporaryDirectory() as scratch:
        worktree = Path(scratch) / "worktree"
        run(
            "git",
            "worktree",
            "add",
            "--quiet",
            "--detach",
            str(worktree),
            f"{remote}/{branch}",
            cwd=clone,
        )
        try:
            save_tracker(worktree / TRACKER, tracker)
            run("git", "add", str(TRACKER), cwd=worktree)
            message = f"chore: mark {task_id} in progress"
            run("git", "commit", "--quiet", "-m", message, cwd=worktree)
            run("git", "push", "--quiet", remote, f"HEAD:refs/heads/{branch}", cwd=worktree)
        finally:
            run("git", "worktree", "remove", "--force", str(worktree), cwd=clone)
    return True


def task_pull_requests(clone: Path, task_id: str) -> list[dict[str, Any]]:
    """List the open pull requests into the task branch with the commit each points at."""
    listed = run(
        "gh",
        "pr",
        "list",
        "--base",
        task_branch(task_id),
        "--state",
        "open",
        "--json",
        "number,url,headRefOid",
        cwd=clone,
    )
    return json.loads(listed)
