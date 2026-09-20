"""Persist task reservations and reconcile them with the main tracker."""

import copy
import fcntl
import json
import os
from pathlib import Path
from typing import Any


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


def effective_tracker(tracker: dict[str, Any], state: dict[str, Any]) -> dict[str, Any]:
    effective = copy.deepcopy(tracker)
    overrides = state["tasks"]
    for task in effective["tasks"]:
        local = overrides.get(task["id"])
        if local and not local.get("expired") and local["status"] != "completed" and task["status"] != "completed":
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
