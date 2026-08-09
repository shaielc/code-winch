#!/usr/bin/env python3
"""Poll merged GitHub PRs and dispatch newly available tasks to Codex Cloud."""

from __future__ import annotations

import argparse
import fcntl
import json
import os
import re
import subprocess
import sys
import tempfile
import time
import urllib.request
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

TASK_ID = re.compile(r"(?<![A-Z0-9])P\d+-\d{3}(?![A-Z0-9])", re.IGNORECASE)


def run(*command: str, cwd: Path, capture: bool = True) -> str:
    result = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.STDOUT if capture else None,
    )
    return result.stdout.strip() if capture else ""


def repository_name(repo_root: Path, remote: str) -> str:
    url = run("git", "remote", "get-url", remote, cwd=repo_root)
    match = re.search(r"github\.com[/:]([^/]+/[^/]+?)(?:\.git)?$", url)
    if not match:
        raise ValueError(f"cannot derive a GitHub repository from remote URL: {url}")
    return match.group(1)


def default_state_file(repo_root: Path) -> Path:
    git_dir = run("git", "rev-parse", "--git-common-dir", cwd=repo_root)
    path = Path(git_dir)
    if not path.is_absolute():
        path = repo_root / path
    return path.resolve() / "code-winch" / "task-daemon.json"


def load_state(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"schema_version": 1, "tasks": {}}
    state = json.loads(path.read_text())
    if state.get("schema_version") != 1 or not isinstance(state.get("tasks"), dict):
        raise ValueError(f"unsupported daemon state in {path}")
    return state


def save_state(path: Path, state: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n")
    temporary.replace(path)


def acquire_lock(state_file: Path) -> Any:
    lock_file = state_file.with_suffix(".lock")
    lock_file.parent.mkdir(parents=True, exist_ok=True)
    handle = lock_file.open("w")
    try:
        fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        handle.close()
        raise RuntimeError(f"another task daemon holds {lock_file}") from None
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
    effective = json.loads(json.dumps(tracker))
    overrides = state["tasks"]
    for task in effective["tasks"]:
        local = overrides.get(task["id"])
        if local:
            task["status"] = local["status"]
            task["owner"] = local.get("owner")
            task["blocked_reason"] = local.get("blocked_reason")
    return effective


def github_request(url: str, token: str | None) -> Any:
    headers = {"Accept": "application/vnd.github+json", "User-Agent": "code-winch"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    request = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def merged_pull_requests(repository: str, token: str | None, limit: int) -> list[dict[str, Any]]:
    url = (
        f"https://api.github.com/repos/{repository}/pulls"
        f"?state=closed&sort=updated&direction=desc&per_page={limit}"
    )
    return [pr for pr in github_request(url, token) if pr.get("merged_at")]


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


def prompt_for(task: dict[str, Any]) -> str:
    return f"""Implement {task['id']}: {task['title']}.

Read docs/workplan/{task['brief']} and all applicable AGENTS.md files. Run the
required verification, commit the changes, and open a pull request. Include
`Task: {task['id']}` in the pull request body so the local scheduler can detect
completion. Do not edit task status fields in docs/workplan/tasks.json.
"""


def dispatch(
    repo_root: Path,
    state_file: Path,
    state: dict[str, Any],
    tasks: list[dict[str, Any]],
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
                "codex", "cloud", "exec", "--env", environment, prompt_for(task), cwd=repo_root
            )
        except (OSError, subprocess.CalledProcessError) as error:
            state["tasks"][task_id] = {
                "status": "pending",
                "owner": None,
                "blocked_reason": None,
                "launch_error": str(error),
                "updated_at": datetime.now(UTC).isoformat(),
            }
            save_state(state_file, state)
            print(f"failed to dispatch {task_id}: {error}", file=sys.stderr)
            continue
        state["tasks"][task_id]["cloud_output"] = output
        save_state(state_file, state)
        print(f"dispatched {task_id}: {output}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--env", required=True, help="Codex Cloud environment ID")
    parser.add_argument("--repo", help="GitHub owner/repository (derived from origin by default)")
    parser.add_argument("--remote", default="origin")
    parser.add_argument("--branch", default="main")
    parser.add_argument("--state-file", type=Path)
    parser.add_argument("--poll-interval", type=float, default=30)
    parser.add_argument("--max-concurrent", type=int, default=3)
    parser.add_argument("--pr-limit", type=int, default=100)
    parser.add_argument("--once", action="store_true", help="run one scheduling cycle")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo_root = Path(run("git", "rev-parse", "--show-toplevel", cwd=Path.cwd()))
    repository = args.repo or repository_name(repo_root, args.remote)
    state_file = (args.state_file or default_state_file(repo_root)).resolve()
    lock_handle = acquire_lock(state_file)
    token = os.environ.get("GITHUB_TOKEN")

    while True:
        try:
            tracker = load_tracker(repo_root, args.remote, args.branch)
            state = load_state(state_file)
            known_ids = {task["id"] for task in tracker["tasks"]}
            pulls = merged_pull_requests(repository, token, args.pr_limit)
            if record_completions(state, pulls, known_ids):
                save_state(state_file, state)
            available = available_tasks(repo_root, effective_tracker(tracker, state))
            dispatch(
                repo_root, state_file, state, available, args.env, args.max_concurrent
            )
        except Exception as error:
            print(f"scheduler cycle failed: {error}", file=sys.stderr)
            if args.once:
                return 1
        if args.once:
            lock_handle.close()
            return 0
        time.sleep(args.poll_interval)


if __name__ == "__main__":
    raise SystemExit(main())
