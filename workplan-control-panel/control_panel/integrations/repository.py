"""Maintain the dedicated main checkout and task branches; read task PRs.

Commands that reach GitHub go through github_run, which is the only thing here that knows a
credential is involved; everything local stays on run().
"""

import json
import re
import subprocess
import tempfile
from collections.abc import Iterable
from pathlib import Path
from typing import Any

from .github import github_run
from .process import run

TRACKER = Path("docs/workplan/tasks.json")


class CheckoutConflict(Exception):
    """The checkout cannot safely be updated automatically."""


class RebaseConflict(Exception):
    """A branch could not be replayed without a conflict, so nothing was pushed."""


def ensure_clone(clone: Path, url: str) -> None:
    """Create the panel's dedicated main checkout, unless it is already there.

    The panel owns this checkout, so making it belongs behind the same credential as every
    later fetch rather than in the entrypoint, which holds no GitHub credential of its own.
    """
    if (clone / ".git").exists():
        return
    github_run("git", "clone", "--branch", "main", "--single-branch", url, str(clone),
               cwd=clone.parent)


def pull_main(clone: Path) -> tuple[dict[str, Any], str]:
    if run("git", "branch", "--show-current", cwd=clone) != "main":
        raise CheckoutConflict("The control-panel checkout must be on main")
    if run("git", "status", "--porcelain", '-uno', cwd=clone):
        raise CheckoutConflict("The control-panel checkout must be clean")
    github_run("git", "pull", "--ff-only", "origin", "main", cwd=clone)
    base = run("git", "rev-parse", "HEAD", cwd=clone)
    if base != run("git", "rev-parse", "origin/main", cwd=clone):
        raise CheckoutConflict("The main checkout has local commits; reconcile it before syncing")
    return json.loads((clone / TRACKER).read_text()), base


def task_head(clone: Path, task_id: str) -> str:
    branch = task_branch(task_id)
    github_run("git", "fetch", "origin",
               f"refs/heads/{branch}:refs/remotes/origin/{branch}", cwd=clone)
    return run("git", "rev-parse", f"origin/{branch}", cwd=clone)


def save_tracker(path: Path, tracker: dict[str, Any]) -> None:
    path.write_text(json.dumps(tracker, indent=2) + "\n")


def task_branch(task_id: str) -> str:
    return f"task/{task_id}"


class StageBranchExists(ValueError):
    """The remote already has the stage branch, so nothing was pushed."""


def stage_branch(stage: str, task_id: str, number: int) -> str:
    return f"{stage}/{task_id}-{number}"


def stage_branch_number(stage: str, task_id: str, branch: Any) -> int | None:
    """The N in `<stage>/<ID>-<N>`, or None for any other name.

    Matching is exact rather than by glob, so another stage's or another task's branch, or a
    name that only shares the prefix, never counts as one of this task's numbers.
    """
    if not isinstance(branch, str):
        return None
    found = re.fullmatch(rf"{re.escape(stage)}/{re.escape(task_id)}-([0-9]+)", branch)
    return int(found.group(1)) if found else None


def next_stage_number(clone: Path, remote: str, stage: str, task_id: str,
                      recorded: Iterable[Any] = ()) -> int:
    """One past the highest `<stage>/<ID>-<N>` the remote has or the task's records name.

    The remote says what exists, including a branch whose push landed without the panel
    hearing back, or whose record is gone. The records say what was used: a stage branch is
    deleted once its pull request merges, and its number must not be handed out again.
    """
    output = github_run("git", "ls-remote", "--heads", remote,
                        f"refs/heads/{stage}/{task_id}-*", cwd=clone)
    names = [line.split("\t", 1)[1].removeprefix("refs/heads/")
             for line in output.splitlines() if "\t" in line]
    numbers = (stage_branch_number(stage, task_id, name) for name in (*names, *recorded))
    return max((number for number in numbers if number is not None), default=0) + 1


def create_stage_branch(clone: Path, remote: str, branch: str, base: str) -> None:
    """Push a new stage branch at the recorded task tip without rewriting an old one."""
    ref = f"refs/heads/{branch}"
    if github_run("git", "ls-remote", "--heads", remote, ref, cwd=clone):
        raise StageBranchExists(f"stage branch {branch} already exists")
    github_run("git", "push", remote, f"{base}:{ref}", cwd=clone)


def delete_stage_branches(clone: Path, remote: str, task_id: str) -> None:
    """Delete every numbered Refine and Implement branch owned by a task."""
    output = github_run("git", "ls-remote", "--heads", remote,
                        f"refs/heads/refine/{task_id}-*",
                        f"refs/heads/implement/{task_id}-*", cwd=clone)
    refs = [line.split("\t", 1)[1] for line in output.splitlines() if "\t" in line]
    if refs:
        github_run("git", "push", "--quiet", remote, "--delete", *refs, cwd=clone)
    for ref in refs:
        run("git", "update-ref", "-d", f"refs/remotes/{remote}/{ref.removeprefix('refs/heads/')}",
            cwd=clone)


def ensure_task_branch(clone: Path, remote: str, base: str, task_id: str,
                       base_commit: str | None = None) -> None:
    """Create the task branch on the remote from the base branch unless it exists."""
    ref = f"refs/heads/{task_branch(task_id)}"
    if base_commit is None:
        github_run("git", "fetch", remote, base, cwd=clone)
    if not github_run("git", "ls-remote", "--heads", remote, ref, cwd=clone):
        source = base_commit or f"{remote}/{base}"
        github_run("git", "push", remote, f"{source}:{ref}", cwd=clone)
    github_run("git", "fetch", remote,
               f"{ref}:refs/remotes/{remote}/{task_branch(task_id)}", cwd=clone)


def delete_task_branch(clone: Path, remote: str, task_id: str) -> None:
    """Remove the task branch from the remote and the clone's tracking ref, if either exists.

    GitHub closes any open pull request whose head branch is deleted.
    """
    ref = f"refs/heads/{task_branch(task_id)}"
    if github_run("git", "ls-remote", "--heads", remote, ref, cwd=clone):
        github_run("git", "push", "--quiet", remote, "--delete", ref, cwd=clone)
    run("git", "update-ref", "-d", f"refs/remotes/{remote}/{task_branch(task_id)}", cwd=clone)


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
            "--detach",
            str(worktree),
            f"{remote}/{branch}",
            cwd=clone,
        )
        try:
            save_tracker(worktree / TRACKER, tracker)
            run("git", "add", str(TRACKER), cwd=worktree)
            message = f"chore: mark {task_id} in progress"
            run("git", "commit", "-m", message, cwd=worktree)
            github_run("git", "push", remote, f"HEAD:refs/heads/{branch}", cwd=worktree)
        finally:
            run("git", "worktree", "remove", "--force", str(worktree), cwd=clone)
    return True


def task_pull_request(clone: Path, task_id: str) -> dict[str, Any] | None:
    """The pull request carrying the task branch into main: an open one, else the merged one."""
    listed = json.loads(github_run(
        "gh", "pr", "list", "--head", task_branch(task_id), "--base", "main",
        "--state", "all", "--json", "url,state", "--limit", "20", cwd=clone,
    ))
    for state in ("OPEN", "MERGED"):
        found = next((pull for pull in listed if pull["state"] == state), None)
        if found:
            return found
    return None


def create_task_pull_request(clone: Path, task_id: str, title: str) -> str:
    """Open the pull request carrying the task branch into main, and return its URL.

    The body carries `Task: <ID>` like the pull requests the agents open, which is how the
    completion workflow ties a merge back to its task.

    The author GitHub records is whoever the token belongs to, which is the App: the pull
    request is opened by `<app>[bot]` so that a human remains free to review and approve it.
    """
    created = github_run("gh", "pr", "create", "--base", "main", "--head", task_branch(task_id),
                         "--title", f"{task_id}: {title}", "--body", f"Task: {task_id}", cwd=clone)
    url = created.splitlines()[-1].strip() if created else ""
    if not url.startswith("http"):
        raise ValueError("gh returned no pull request URL")
    return url


def task_pull_requests(clone: Path, task_id: str) -> list[dict[str, Any]]:
    """List the open pull requests into the task branch with the commit each points at."""
    listed = github_run(
        "gh",
        "pr",
        "list",
        "--base",
        task_branch(task_id),
        "--state",
        "open",
        "--json",
        "number,url,headRefOid,headRefName,isCrossRepository",
        cwd=clone,
    )
    return json.loads(listed)


def rebase_task_branch(clone: Path, remote: str, task_id: str,
                       pulls: list[dict[str, Any]]) -> dict[str, Any] | None:
    """Rebase the task branch onto main, and each pull request into it onto the result.

    Returns None when the branch already contains main. Every branch is replayed in a
    temporary worktree before anything is pushed, and the push is atomic and leased on the
    fetched tips, so a conflict or a branch that moved meanwhile leaves the remote untouched.
    The pull requests' head branches must live on `remote`.
    """
    branch = task_branch(task_id)
    heads = [branch] + [pull["headRefName"] for pull in pulls]
    # Forced, because a rebase from elsewhere is exactly the update this must accept.
    github_run("git", "fetch", remote, f"+refs/heads/main:refs/remotes/{remote}/main",
               *(f"+refs/heads/{head}:refs/remotes/{remote}/{head}" for head in heads), cwd=clone)
    main = run("git", "rev-parse", f"{remote}/main", cwd=clone)
    old = {head: run("git", "rev-parse", f"{remote}/{head}", cwd=clone) for head in heads}
    if run("git", "merge-base", main, old[branch], cwd=clone) == main:
        return None

    new = {}
    with tempfile.TemporaryDirectory() as scratch:
        worktree = Path(scratch) / "worktree"
        run("git", "worktree", "add", "--detach", str(worktree), old[branch], cwd=clone)
        try:
            new[branch] = replay(worktree, branch, main, old[branch])
            for head in heads[1:]:
                new[head] = replay(worktree, head, new[branch], old[head], since=old[branch])
        finally:
            run("git", "worktree", "remove", "--force", str(worktree), cwd=clone)

    github_run("git", "push", "--atomic", "--quiet", remote,
               *(f"--force-with-lease=refs/heads/{head}:{old[head]}" for head in heads),
               *(f"{new[head]}:refs/heads/{head}" for head in heads), cwd=clone)
    # The clone fetches only main by default, so the push does not move these on its own.
    for head in heads:
        run("git", "update-ref", f"refs/remotes/{remote}/{head}", new[head], cwd=clone)
    return {"main_commit": main, "head": new[branch],
            "propagated": [{"number": pull["number"], "url": pull["url"],
                            "head": new[pull["headRefName"]]} for pull in pulls]}


def replay(worktree: Path, name: str, onto: str, tip: str, since: str | None = None) -> str:
    """Replay `tip`'s commits after `since` (default: those not on `onto`) onto `onto`."""
    run("git", "checkout", "--quiet", "--detach", tip, cwd=worktree)
    try:
        run("git", "rebase", "--quiet", "--onto", onto, since or onto, cwd=worktree)
    except subprocess.CalledProcessError as error:
        try:
            run("git", "rebase", "--abort", cwd=worktree)
        except subprocess.CalledProcessError:
            pass
        raise RebaseConflict(f"{name} does not rebase cleanly; resolve it by hand") from error
    return run("git", "rev-parse", "HEAD", cwd=worktree)
