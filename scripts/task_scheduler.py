#!/usr/bin/env python3
"""Process GitHub merge events and dispatch available tasks to Codex Cloud."""

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
import time
from datetime import UTC, datetime
from pathlib import Path
from string import Template
from typing import Any

TASK_ID = re.compile(r"(?<![A-Z0-9])P\d+-\d{3}(?![A-Z0-9])", re.IGNORECASE)
TASK_URL = re.compile(r"https?://\S+/codex/tasks/\S+")
PROMPT_TEMPLATE = Path("scripts/task-prompt.md")


def run(*command: str, cwd: Path, capture: bool = True) -> str:
    result = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
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
    contents = run(
        "git", "show", f"{remote}/{branch}:docs/workplan/tasks.json", cwd=repo_root
    )
    return json.loads(contents)


def effective_tracker(tracker: dict[str, Any], state: dict[str, Any]) -> dict[str, Any]:
    effective = copy.deepcopy(tracker)
    overrides = state["tasks"]
    for task in effective["tasks"]:
        local = overrides.get(task["id"])
        if local and task["status"] != "completed":
            task["status"] = local["status"]
            task["owner"] = local.get("owner")
            task["blocked_reason"] = local.get("blocked_reason")
    return effective


def prune_completed(tracker: dict[str, Any], state: dict[str, Any]) -> bool:
    completed = {task["id"] for task in tracker["tasks"] if task["status"] == "completed"}
    stale = [task_id for task_id in state["tasks"] if task_id in completed]
    for task_id in stale:
        del state["tasks"][task_id]
    return bool(stale)


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


def dispatch(
    repo_root: Path,
    state_file: Path,
    state: dict[str, Any],
    tasks: list[dict[str, Any]],
    template: Template,
    environment: str,
    capacity: int,
) -> None:
    active = sum(1 for value in state["tasks"].values() if value["status"] == "in_progress")
    for task in tasks[: max(0, capacity - active)]:
        task_id = task["id"]
        owner = f"codex-cloud:{task_id}:{int(time.time())}"
        state["tasks"][task_id] = {
            "status": "in_progress",
            "owner": owner,
            "blocked_reason": None,
            "updated_at": datetime.now(UTC).isoformat(),
        }
        save_state(state_file, state)
        try:
            output = run(
                "codex",
                "cloud",
                "exec",
                "--env",
                environment,
                prompt_for(template, task),
                cwd=repo_root,
            )
        except (OSError, subprocess.CalledProcessError) as error:
            detail = failure_detail(error)
            state["tasks"][task_id] = {
                "status": "pending",
                "owner": None,
                "blocked_reason": None,
                "launch_error": detail,
                "updated_at": datetime.now(UTC).isoformat(),
            }
            save_state(state_file, state)
            print(f"failed to dispatch {task_id}: {detail}", file=sys.stderr)
            continue
        state["tasks"][task_id]["task_url"] = task_url_from(output) or output
        save_state(state_file, state)
        print(f"dispatched {task_id}: {output}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--env", required=True, help="Codex Cloud environment ID")
    parser.add_argument("--remote", default="origin")
    parser.add_argument("--branch", default="main")
    parser.add_argument(
        "--repo-root",
        type=Path,
        help="directory holding scripts/; defaults to the enclosing git checkout",
    )
    parser.add_argument("--state-file", type=Path)
    parser.add_argument(
        "--tracker-file",
        type=Path,
        help="read the tracker from this file instead of the remote default branch",
    )
    parser.add_argument(
        "--tracker-snapshot",
        type=Path,
        help="write the tracker this run scheduled from to this path",
    )
    parser.add_argument("--prompt-template", type=Path, default=PROMPT_TEMPLATE)
    parser.add_argument("--task", help="dispatch only this task ID when it is available")
    parser.add_argument("--max-concurrent", type=int, default=3)
    parser.add_argument(
        "--event-file",
        type=Path,
        help="GitHub Actions event file; omit for a manual run",
    )
    arguments = parser.parse_args()
    if arguments.repo_root and not arguments.state_file:
        parser.error("--state-file is required when --repo-root is not a git checkout")
    return arguments


def schedule(
    repo_root: Path,
    args: argparse.Namespace,
    state_file: Path,
    pulls: list[dict[str, Any]],
) -> None:
    tracker = (
        json.loads(args.tracker_file.read_text())
        if args.tracker_file
        else load_tracker(repo_root, args.remote, args.branch)
    )
    if args.tracker_snapshot:
        write_json(args.tracker_snapshot, tracker)
    state = load_state(state_file)
    known_ids = {task["id"] for task in tracker["tasks"]}
    changed = record_completions(state, pulls, known_ids)
    if prune_completed(tracker, state) or changed:
        save_state(state_file, state)
    template = load_prompt_template(repo_root, args.prompt_template)
    available = available_tasks(repo_root, effective_tracker(tracker, state))
    if args.task:
        requested = args.task.upper()
        available = [task for task in available if task["id"] == requested]
        if not available:
            print(f"{requested} is not currently available; nothing to dispatch")
            return
    dispatch(
        repo_root, state_file, state, available, template, args.env, args.max_concurrent
    )


def main() -> int:
    args = parse_args()
    repo_root = (
        args.repo_root or Path(run("git", "rev-parse", "--show-toplevel", cwd=Path.cwd()))
    ).resolve()
    state_file = (args.state_file or default_state_file(repo_root)).resolve()
    lock_handle = acquire_lock(state_file)
    try:
        payload = json.loads(args.event_file.read_text()) if args.event_file else {}
        pull = merged_pull_request(payload)
        if "pull_request" in payload and not pull:
            print("event is not a merged pull request; nothing to do")
            return 0
        schedule(repo_root, args, state_file, [pull] if pull else [])
        return 0
    finally:
        lock_handle.close()


if __name__ == "__main__":
    raise SystemExit(main())
