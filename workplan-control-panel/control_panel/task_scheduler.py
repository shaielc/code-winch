"""Workplan orchestration owned by the control-panel API."""
from __future__ import annotations

import json
import subprocess
from contextlib import contextmanager
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from . import prompts, state as task_state
from .integrations import codex, repository


class TaskError(Exception):
    def __init__(self, status: int, message: str, **details: Any):
        super().__init__(message)
        self.status = status
        self.details = details


class TaskScheduler:
    def __init__(self, clone: Path, state_file: Path,
                 tracker_path: Path, environment: str, capacity: int = 3):
        if capacity < 1:
            raise ValueError("capacity must be positive")
        self.clone = clone
        self.state_file, self.tracker_path = state_file, tracker_path
        self.environment, self.capacity = environment, capacity

    @contextmanager
    def locked(self):
        try:
            lock = task_state.acquire_lock(self.state_file)
        except RuntimeError as error:
            raise TaskError(409, "Another control-panel operation is running; retry shortly") from error
        try:
            yield
        finally:
            lock.close()

    def tracker(self):
        if not self.tracker_path.exists():
            return {"tasks": []}
        return json.loads(self.tracker_path.read_text())

    def snapshot(self):
        return {"tracker": self.tracker(), "state": task_state.load_state(self.state_file)}

    def sync(self, event: dict[str, Any] | None = None):
        if event is not None:
            pull = event.get("pull_request")
            if (event.get("action") != "closed" or not isinstance(pull, dict)
                    or not pull.get("merged_at") or not isinstance(pull.get("base"), dict)
                    or pull["base"].get("ref") != "main"):
                return {"ignored": True, "reason": "Only merges into main trigger scheduling"}
        with self.locked():
            try:
                tracker, base = repository.pull_main(self.clone)
            except repository.CheckoutConflict as error:
                raise TaskError(409, str(error)) from error
            state = task_state.load_state(self.state_file)
            task_state.retire_completed(tracker, state)
            task_state.write_json(self.tracker_path, tracker)
            task_state.write_json(self.state_file, state)
            effective = task_state.effective_tracker(tracker, state)
            done = {t["id"] for t in tracker["tasks"] if t["status"] == "completed"}
            available = [t for t in effective["tasks"]
                         if t["status"] == "pending" and set(t["depends_on"]) <= done]
            active = sum(t["status"] == "in_progress" for t in effective["tasks"])
            selected = available[:max(0, self.capacity - active)]
            # Retry incomplete preparation before claiming more work. The reservation
            # survives process death and branch helpers safely reuse an existing branch.
            pending = [t for t in tracker["tasks"] if t["id"] in state["tasks"]
                       and state["tasks"][t["id"]]["status"] == "in_progress"
                       and not state["tasks"][t["id"]].get("expired")
                       and not state["tasks"][t["id"]].get("prepared")]
            prepared = []
            for task in pending + selected:
                record = state["tasks"].setdefault(task["id"], {})
                record.update(status="in_progress", expired=False, branch=repository.task_branch(task["id"]),
                              owner="control-panel", updated_at=datetime.now(UTC).isoformat())
                task_state.write_json(self.state_file, state)
                repository.ensure_task_branch(self.clone, "origin", "main", task["id"], base_commit=base)
                repository.push_opening_commit(self.clone, "origin", task["id"])
                record["prepared"] = True
                task_state.write_json(self.state_file, state)
                prepared.append(task["id"])
            return {"prepared": prepared, "main_commit": base}

    def task(self, task_id: str, state):
        task = next((t for t in self.tracker()["tasks"] if t["id"] == task_id), None)
        if task is None:
            raise TaskError(404, "Unknown task")
        record = state["tasks"].get(task_id, {})
        if task["status"] == "completed" or record.get("expired") or record.get("status") != "in_progress" or not record.get("prepared"):
            raise TaskError(409, "Sync and prepare an available task before choosing a stage")
        return task, record

    def expire(self, task_id: str):
        """Release a local reservation, retaining its conversation history."""
        with self.locked():
            state = task_state.load_state(self.state_file)
            task = next((t for t in self.tracker()["tasks"] if t["id"] == task_id), None)
            if task is None:
                raise TaskError(404, "Unknown task")
            record = state["tasks"].get(task_id)
            if not record or task["status"] == "completed" or record.get("status") == "completed":
                raise TaskError(409, "This task has no active local reservation to expire")
            record.update(expired=True, owner=None, updated_at=datetime.now(UTC).isoformat())
            task_state.write_json(self.state_file, state)
            return {"expired": task_id}

    def stage(self, task_id: str, stage: str, pr_number: int | None = None):
        if stage not in ("refine", "implement", "audit"):
            raise TaskError(404, "Unknown stage")
        with self.locked():
            state = task_state.load_state(self.state_file)
            task, record = self.task(task_id, state)
            template = prompts.load(stage)
            fields = {key: task[key] for key in ("id", "title", "brief")}
            branch = repository.task_branch(task_id)
            if stage == "audit":
                pulls = repository.task_pull_requests(self.clone, task_id)
                candidates = [p for p in pulls if pr_number is None or p["number"] == pr_number]
                if len(candidates) != 1:
                    raise TaskError(409, "Select one open implementation pull request into the task branch", pull_requests=pulls)
                pull = candidates[0]
                fields.update(pr_url=pull["url"], head=pull["headRefOid"])
                return {"prompt": template.substitute(fields), "pull_request": pull["url"], "head": pull["headRefOid"]}
            if not self.environment:
                raise TaskError(503, "Set CODEX_ENV_ID on the control panel")
            head = repository.task_head(self.clone, task_id)
            stages = record.setdefault("stages", {})
            key = f"{stage}:{head}"
            previous = stages.get(key)
            if previous:
                if previous["status"] == "submitted":
                    return {**previous, "reused": True}
                raise TaskError(409, "The previous submission has an uncertain outcome; check Codex before clearing its stage record")
            prompt = template.substitute(fields) + f"\nBase your work and pull request on `{branch}`.\n"
            stages[key] = {"status": "submitting", "stage": stage, "head": head,
                           "updated_at": datetime.now(UTC).isoformat()}
            task_state.write_json(self.state_file, state)
            try:
                url = codex.submit(self.clone, self.environment, branch, prompt)
            except (OSError, ValueError, subprocess.SubprocessError) as error:
                # Preserve the reservation: a transport failure may happen after acceptance.
                raise TaskError(502, "Codex submission failed; check Codex before clearing its stage record") from error
            stages[key].update(status="submitted", task_url=url)
            record.update(task_url=url, updated_at=datetime.now(UTC).isoformat())
            task_state.write_json(self.state_file, state)
            return stages[key]
