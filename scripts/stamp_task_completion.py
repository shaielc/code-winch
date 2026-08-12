#!/usr/bin/env python3
"""Stamp a pull request's task as completed in the workplan tracker."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path
from typing import Any

from task_scheduler import failure_detail, run, task_id_for_pr

TRACKER = Path("docs/workplan/tasks.json")
BOT_NAME = "github-actions[bot]"
BOT_EMAIL = "41898282+github-actions[bot]@users.noreply.github.com"


def load_tracker(path: Path) -> dict[str, Any]:
    tracker = json.loads(path.read_text())
    if tracker.get("schema_version") != 1:
        raise ValueError(f"unsupported tracker schema in {path}")
    return tracker


def save_tracker(path: Path, tracker: dict[str, Any]) -> None:
    path.write_text(json.dumps(tracker, indent=2) + "\n")


def contains_base(repo_root: Path, remote: str, branch: str) -> bool:
    run("git", "fetch", "--quiet", remote, branch, cwd=repo_root)
    result = subprocess.run(
        ("git", "merge-base", "--is-ancestor", f"{remote}/{branch}", "HEAD"),
        cwd=repo_root,
        check=False,
    )
    return result.returncode == 0


def merge_base_branch(repo_root: Path, remote: str, branch: str) -> bool:
    """Merge the base branch in, leaving the branch untouched when it conflicts."""
    result = subprocess.run(
        ("git", "merge", "--no-edit", f"{remote}/{branch}"),
        cwd=repo_root,
        check=False,
    )
    if result.returncode:
        subprocess.run(("git", "merge", "--abort"), cwd=repo_root, check=False)
    return result.returncode == 0


def is_fork(pr: dict[str, Any], payload: dict[str, Any]) -> bool:
    head = pr.get("head", {}).get("repo") or {}
    base = pr.get("base", {}).get("repo") or {}
    return bool(head.get("full_name")) and head["full_name"] != base.get("full_name")


def skipped(pr: dict[str, Any], label: str) -> bool:
    return any(item.get("name") == label for item in pr.get("labels", []))


def configure_identity(repo_root: Path) -> None:
    run("git", "config", "user.name", BOT_NAME, cwd=repo_root)
    run("git", "config", "user.email", BOT_EMAIL, cwd=repo_root)


def push_head(repo_root: Path, remote: str, head_ref: str) -> None:
    run("git", "push", remote, f"HEAD:refs/heads/{head_ref}", cwd=repo_root)


def push_stamp(repo_root: Path, remote: str, head_ref: str, task_id: str) -> None:
    configure_identity(repo_root)
    run("git", "add", str(TRACKER), cwd=repo_root)
    run("git", "commit", "-m", f"chore: mark {task_id} completed", cwd=repo_root)
    push_head(repo_root, remote, head_ref)


def catch_up_with_base(
    repo_root: Path,
    args: argparse.Namespace,
    pr: dict[str, Any],
    payload: dict[str, Any],
    task_id: str,
) -> int:
    """Fold the base branch into a stale branch so the next approval can stamp it."""
    base = f"{args.remote}/{args.branch}"
    if is_fork(pr, payload):
        print(
            f"error: this branch does not contain {base} and lives in a fork, so it "
            f"cannot be updated from here. Merge {args.branch} into it and push again.",
            file=sys.stderr,
        )
        return 1

    configure_identity(repo_root)
    if not merge_base_branch(repo_root, args.remote, args.branch):
        print(
            f"error: {base} conflicts with this branch. Resolve the conflicts locally "
            "and push again so the tracker cannot revert another task's status.",
            file=sys.stderr,
        )
        return 1

    try:
        push_head(repo_root, args.remote, pr["head"]["ref"])
    except subprocess.CalledProcessError as error:
        print(f"error: could not push the {base} merge: {failure_detail(error)}", file=sys.stderr)
        return 1

    print(f"merged {base}; approve again to stamp {task_id} completed")
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--event-file", type=Path, required=True)
    parser.add_argument("--remote", default="origin")
    parser.add_argument("--branch", default="main")
    parser.add_argument("--skip-label", default="no-task")
    parser.add_argument("--check-only", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo_root = Path(run("git", "rev-parse", "--show-toplevel", cwd=Path.cwd()))
    tracker_path = repo_root / TRACKER

    payload = json.loads(args.event_file.read_text())
    pr = payload.get("pull_request")
    if not isinstance(pr, dict):
        print("event carries no pull request; nothing to do")
        return 0

    if skipped(pr, args.skip_label):
        print(f"pull request is labelled {args.skip_label}; skipping the task gate")
        return 0

    tracker = load_tracker(tracker_path)
    known_ids = {task["id"] for task in tracker["tasks"]}
    task_id = task_id_for_pr(pr, known_ids)
    if not task_id:
        print(
            "error: this pull request must reference exactly one known task ID in its "
            "title, body, or branch name. Add `Task: P0-000` to the body, or apply the "
            f"`{args.skip_label}` label if it does not implement a workplan task.",
            file=sys.stderr,
        )
        return 1

    if args.check_only:
        print(f"{task_id} resolved; it will be stamped completed once approved")
        return 0

    task = next(item for item in tracker["tasks"] if item["id"] == task_id)
    if task["status"] == "completed":
        print(f"{task_id} is already completed in the tracker")
        return 0

    if not contains_base(repo_root, args.remote, args.branch):
        return catch_up_with_base(repo_root, args, pr, payload, task_id)

    task["status"] = "completed"
    task["owner"] = None
    task["blocked_reason"] = None

    if is_fork(pr, payload):
        print(
            f"error: {task_id} needs a tracker update but the branch lives in a fork, so "
            "it cannot be pushed. Set the status manually in docs/workplan/tasks.json.",
            file=sys.stderr,
        )
        return 1

    save_tracker(tracker_path, tracker)
    push_stamp(repo_root, args.remote, pr["head"]["ref"], task_id)
    print(f"stamped {task_id} completed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
