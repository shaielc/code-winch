"""Workplan orchestration owned by the control-panel API."""
from __future__ import annotations

import json
import logging
import threading
import uuid
import subprocess
from contextlib import contextmanager
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from . import prompts, state as task_state
from .integrations import codex, repository
from .integrations.process import failure_message


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
        self._jobs_lock = threading.Lock()
        self._jobs = {}
        self._resync_requested = False

    def start_sync(self, event=None, only=None):
        """Return promptly so proxy timeouts cannot interrupt the HTTP response."""
        if only is not None and not any(t["id"] == only for t in self.tracker()["tasks"]):
            raise TaskError(404, "Unknown task")
        with self._jobs_lock:
            active = next((job for job in self._jobs.values() if job["status"] == "running"), None)
            if active:
                # A scheduling pass coalesces into the running job, but a request to prepare
                # one named task cannot: folding it in would silently drop the selection.
                if only is not None:
                    raise TaskError(409, "A sync is already running; retry when it finishes")
                self._resync_requested = True
                return dict(active)
            job_id = uuid.uuid4().hex
            job = {"id": job_id, "status": "running"}
            # Bound retained results; task reservations themselves live on disk.
            if len(self._jobs) >= 32:
                self._jobs.pop(next(iter(self._jobs)))
            self._jobs[job_id] = job
        threading.Thread(target=self._run_sync, args=(job_id, event, only), daemon=True).start()
        return {"id": job_id, "status": "running"}

    def _run_sync(self, job_id, event, only=None):
        try:
            while True:
                result = {"status": "succeeded", "result": self.sync(event, only)}
                with self._jobs_lock:
                    if self._resync_requested:
                        # A merge landed mid-pass; finish as a full scheduling pass.
                        self._resync_requested = False
                        event, only = None, None
                        continue
                    self._jobs[job_id].update(result)
                    return
        except TaskError as error:
            result = {"status": "failed", "error": str(error), **error.details}
        except (OSError, subprocess.SubprocessError) as error:
            result = {"status": "failed", "error": failure_message(error)}
        except Exception:
            # Do not log a command, prompt, or exception text that might contain secrets.
            logging.getLogger("control_panel").error("Unexpected sync failure")
            result = {"status": "failed", "error": "Sync failed unexpectedly; check the tracker and panel configuration."}
        with self._jobs_lock:
            self._resync_requested = False
            self._jobs[job_id].update(result)

    def sync_status(self, job_id):
        with self._jobs_lock:
            if job_id not in self._jobs:
                raise TaskError(404, "Sync status is unavailable after a restart; retry Sync main.")
            return dict(self._jobs[job_id])

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

    def requested(self, task_id: str, effective, done, active):
        """Admit one named task, applying the rules the scheduling pass applies to a batch.

        Preparation is idempotent, so a task already holding a lease is allowed through to
        retry its branch work; it is only refused a slot it does not already occupy.
        """
        task = next((t for t in effective["tasks"] if t["id"] == task_id), None)
        if task is None:
            raise TaskError(404, "Unknown task")
        if task["status"] not in ("pending", "in_progress"):
            raise TaskError(409, f"{task_id} is {task['status']} and cannot be prepared")
        unmet = sorted(set(task["depends_on"]) - done)
        if unmet:
            raise TaskError(409, f"{task_id} waits on {', '.join(unmet)}")
        if task["status"] != "in_progress" and active >= self.capacity:
            raise TaskError(409, f"All {self.capacity} scheduler slots are in flight; "
                                 f"expire a lease before preparing {task_id}")
        return [task]

    def sync(self, event: dict[str, Any] | None = None, only: str | None = None):
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
            active = sum(t["status"] == "in_progress" for t in effective["tasks"])
            if only is not None:
                # Preparing one named task stands alone: an unrelated task's failure
                # must not fail the request the operator actually made.
                pending, selected = [], self.requested(only, effective, done, active)
            else:
                available = [t for t in effective["tasks"]
                             if t["status"] == "pending" and set(t["depends_on"]) <= done]
                selected = available[:max(0, self.capacity - active)]
                # Retry incomplete preparation before claiming more work. The reservation
                # survives process death and branch helpers safely reuse an existing branch.
                pending = [t for t in effective["tasks"] if t["status"] == "in_progress"
                           and not state["tasks"].get(t["id"], {}).get("prepared")]
            prepared = []
            for task in pending + selected:
                record = state["tasks"].setdefault(task["id"], {})
                record.update(status="in_progress", expired=False, branch=repository.task_branch(task["id"]),
                              owner="control-panel", updated_at=datetime.now(UTC).isoformat())
                task_state.write_json(self.state_file, state)
                try:
                    repository.ensure_task_branch(self.clone, "origin", "main", task["id"], base_commit=base)
                    repository.push_opening_commit(self.clone, "origin", task["id"])
                except (OSError, subprocess.SubprocessError, ValueError) as error:
                    record["prepare_error"] = failure_message(error)
                    task_state.write_json(self.state_file, state)
                    raise TaskError(502, f"Main refreshed, but {task['id']} could not be prepared. "
                                    + record["prepare_error"], prepared=prepared) from error
                record.pop("prepare_error", None)
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

    def expire_stage(self, task_id: str, stage: str):
        """Override the idempotence head grants, so the stage can be sent again unchanged.

        A moved branch tip releases a stage on its own; this releases one whose tip has not
        moved. The attempt is marked rather than deleted, because it carries the only link to
        the Codex conversation it opened. It needs no lease: a record can outlive the
        reservation that created it, and clearing one is bookkeeping rather than scheduling.
        """
        if stage not in ("refine", "implement"):
            raise TaskError(404, "Unknown stage")
        with self.locked():
            state = task_state.load_state(self.state_file)
            if not any(t["id"] == task_id for t in self.tracker()["tasks"]):
                raise TaskError(404, "Unknown task")
            attempts = state["tasks"].get(task_id, {}).get("stages", {}).get(stage, [])
            if not attempts:
                raise TaskError(409, f"{task_id} has no recorded {stage} submission to expire")
            head = repository.task_head(self.clone, task_id)
            latest = task_state.live_attempt(attempts, head)
            if latest is None:
                raise TaskError(409, f"The {stage} stage of {task_id} has no live submission at "
                                     "its current revision, so it can already be submitted")
            latest.update(status="expired", updated_at=datetime.now(UTC).isoformat())
            task_state.write_json(self.state_file, state)
            return {"expired": task_id, "stage": stage, "head": head}

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
            attempts = record.setdefault("stages", {}).setdefault(stage, [])
            # Idempotence is scoped to the branch revision: the newest attempt against this
            # one decides, and a tip that has moved on leaves nothing in the way.
            latest = task_state.live_attempt(attempts, head)
            if latest and latest["status"] == "submitted":
                return {**latest, "stage": stage, "reused": True}
            if latest:
                raise TaskError(409, "The previous submission has an uncertain outcome; "
                                f"check Codex, then expire the {stage} stage to submit again")
            prompt = template.substitute(fields) + f"\nBase your work and pull request on `{branch}`.\n"
            attempts.append({"status": "submitting", "head": head,
                             "updated_at": datetime.now(UTC).isoformat()})
            task_state.write_json(self.state_file, state)
            try:
                url = codex.submit(self.clone, self.environment, branch, prompt)
            except (OSError, ValueError, subprocess.SubprocessError) as error:
                print(error)
                # Preserve the attempt: a transport failure may happen after acceptance.
                raise TaskError(502, "Codex submission failed; check Codex, then expire the "
                                f"{stage} stage to submit again") from error
            attempts[-1].update(status="submitted", task_url=url)
            record.update(task_url=url, updated_at=datetime.now(UTC).isoformat())
            task_state.write_json(self.state_file, state)
            return {**attempts[-1], "stage": stage}
