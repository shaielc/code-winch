#!/usr/bin/env python3
"""Shared task/git helpers and a thin client for the control-panel API."""

from __future__ import annotations

import argparse
import copy
import fcntl
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path
from string import Template
from typing import Any

TASK_ID = re.compile(r"(?<![A-Z0-9])P\d+-\d{3}(?![A-Z0-9])", re.IGNORECASE)
TASK_URL = re.compile(r"https?://\S+/codex/(?:cloud/)?tasks/\S+")
PROMPT_TEMPLATE = Path("scripts/task-implementation-prompt.md")
TRACKER = Path("docs/workplan/tasks.json")


def run(*command: str, cwd: Path, capture: bool = True) -> str:
    result = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
        timeout=300,
    )
    return result.stdout.strip() if capture else ""


def failure_detail(error: Exception) -> str:
    """Recover the diagnostics that check=True drops from the exception text."""
    stderr = getattr(error, "stderr", None)
    return f"{error}: {stderr.strip()}" if stderr else str(error)


def task_url_from(output: str) -> str | None:
    match = TASK_URL.search(output)
    return match.group(0) if match else None


def default_state_file(repo_root: Path) -> Path:
    git_dir = run("git", "rev-parse", "--git-common-dir", cwd=repo_root)
    path = Path(git_dir)
    if not path.is_absolute():
        path = repo_root / path
    return path.resolve() / "code-winch" / "task-scheduler.json"


def load_state(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"schema_version": 1, "tasks": {}}
    state = json.loads(path.read_text())
    if state.get("schema_version") != 1 or not isinstance(state.get("tasks"), dict):
        raise ValueError(f"unsupported scheduler state in {path}")
    return state


def write_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n")
    temporary.replace(path)


def save_state(path: Path, state: dict[str, Any]) -> None:
    write_json(path, state)


def acquire_lock(state_file: Path) -> Any:
    lock_file = state_file.with_suffix(".lock")
    lock_file.parent.mkdir(parents=True, exist_ok=True)
    handle = lock_file.open("w")
    try:
        fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        handle.close()
        raise RuntimeError(f"another task scheduler holds {lock_file}") from None
    handle.write(str(os.getpid()))
    handle.flush()
    return handle


def load_tracker(repo_root: Path, remote: str, branch: str) -> dict[str, Any]:
    run("git", "fetch", "--quiet", remote, branch, cwd=repo_root)
    contents = run("git", "show", f"{remote}/{branch}:{TRACKER}", cwd=repo_root)
    return json.loads(contents)


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
    task = next(item for item in tracker["tasks"] if item["id"] == task_id)
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


def ensure_draft_pull_request(clone: Path, base: str, task: dict[str, Any]) -> str:
    """Return the task branch's open pull request URL, opening a draft when there is none."""
    branch = task_branch(task["id"])
    listed = run(
        "gh",
        "pr",
        "list",
        "--head",
        branch,
        "--base",
        base,
        "--state",
        "open",
        "--json",
        "number,url",
        cwd=clone,
    )
    pulls = json.loads(listed)
    if pulls:
        return pulls[0]["url"]
    return run(
        "gh",
        "pr",
        "create",
        "--draft",
        "--base",
        base,
        "--head",
        branch,
        "--title",
        f"{task['id']}: {task['title']}",
        "--body",
        f"Task: {task['id']}",
        cwd=clone,
    )


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


def effective_tracker(tracker: dict[str, Any], state: dict[str, Any]) -> dict[str, Any]:
    effective = copy.deepcopy(tracker)
    overrides = state["tasks"]
    for task in effective["tasks"]:
        local = overrides.get(task["id"])
        if local and local["status"] != "completed" and task["status"] != "completed":
            task["status"] = local["status"]
            task["owner"] = local.get("owner")
            task["blocked_reason"] = local.get("blocked_reason")
    return effective


def retire_completed(tracker: dict[str, Any], state: dict[str, Any]) -> bool:
    """Mark the entries the tracker caught up with, keeping the record they carry.

    A retired entry stops overriding anything, because the overlay skips a task the tracker
    already calls completed, but it still holds the pull request and the Codex task the work
    went through — the only trace of how the task finished once the tracker agrees.
    """
    completed = {task["id"] for task in tracker["tasks"] if task["status"] == "completed"}
    retired = [
        task_id
        for task_id, entry in state["tasks"].items()
        if task_id in completed and entry["status"] != "completed"
    ]
    for task_id in retired:
        state["tasks"][task_id]["status"] = "completed"
        state["tasks"][task_id]["owner"] = None
    return bool(retired)


def merged_pull_request(payload: dict[str, Any]) -> dict[str, Any] | None:
    if payload.get("action") != "closed":
        return None
    pull = payload.get("pull_request")
    return pull if isinstance(pull, dict) and pull.get("merged_at") else None


def task_id_for_pr(pr: dict[str, Any], known_ids: set[str]) -> str | None:
    text = "\n".join(
        str(value or "")
        for value in (pr.get("title"), pr.get("body"), pr.get("head", {}).get("ref"))
    )
    matches = {match.upper() for match in TASK_ID.findall(text)} & known_ids
    return next(iter(matches)) if len(matches) == 1 else None


def record_completions(
    state: dict[str, Any], pulls: list[dict[str, Any]], known_ids: set[str]
) -> bool:
    changed = False
    for pr in pulls:
        task_id = task_id_for_pr(pr, known_ids)
        if not task_id:
            continue
        current = state["tasks"].get(task_id, {})
        if current.get("status") == "completed":
            continue
        state["tasks"][task_id] = {
            **current,
            "status": "completed",
            "owner": None,
            "blocked_reason": None,
            "pull_request": pr["html_url"],
            "updated_at": pr["merged_at"],
        }
        changed = True
    return changed


def available_tasks(repo_root: Path, tracker: dict[str, Any]) -> list[dict[str, Any]]:
    with tempfile.NamedTemporaryFile(mode="w", suffix=".json") as temporary:
        json.dump(tracker, temporary)
        temporary.flush()
        output = run(
            str(repo_root / "scripts/list-available-tasks.sh"),
            temporary.name,
            cwd=repo_root,
        )
    return json.loads(output)


def load_prompt_template(repo_root: Path, path: Path) -> Template:
    resolved = path if path.is_absolute() else repo_root / path
    return Template(resolved.read_text())


def prompt_for(template: Template, task: dict[str, Any]) -> str:
    return template.substitute(id=task["id"], title=task["title"], brief=task["brief"])


def main() -> int:
    """Compatibility command: scheduling lives exclusively in the control-panel API."""
    from urllib.request import Request, urlopen

    parser = argparse.ArgumentParser(description="Trigger the control-panel API")
    parser.add_argument("--url", default=os.environ.get("CONTROL_PANEL_URL", "http://localhost:8765"))
    parser.add_argument("--event-file", type=Path)
    args = parser.parse_args()
    token = os.environ.get("PANEL_TOKEN")
    if not token:
        parser.error("set PANEL_TOKEN")
    path = "/api/events/merge" if args.event_file else "/api/sync"
    body = args.event_file.read_bytes() if args.event_file else b"{}"
    request = Request(args.url.rstrip("/") + path, data=body, headers={
        "Authorization": "Bearer " + token, "Content-Type": "application/json",
    })
    with urlopen(request, timeout=840) as response:
        print(response.read().decode())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
