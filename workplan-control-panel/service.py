"""Workplan orchestration owned by the control-panel API."""
from __future__ import annotations

import json
import subprocess
import sys
from contextlib import contextmanager
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
import task_scheduler as scheduler


class Problem(Exception):
    def __init__(self, status: int, message: str, **details: Any):
        super().__init__(message)
        self.status = status
        self.details = details


class ControlPanel:
    def __init__(self, repo_root: Path, clone: Path, state_file: Path,
                 tracker_path: Path, environment: str, capacity: int = 3):
        if capacity < 1:
            raise ValueError("capacity must be positive")
        self.repo_root, self.clone = repo_root, clone
        self.state_file, self.tracker_path = state_file, tracker_path
        self.environment, self.capacity = environment, capacity

    @contextmanager
    def locked(self):
        try:
            lock = scheduler.acquire_lock(self.state_file)
        except RuntimeError as error:
            raise Problem(409, "Another control-panel operation is running; retry shortly") from error
        try:
            yield
        finally:
            lock.close()

    def tracker(self):
        if not self.tracker_path.exists():
            return {"tasks": []}
        return json.loads(self.tracker_path.read_text())

    def snapshot(self):
        return {"tracker": self.tracker(), "state": scheduler.load_state(self.state_file)}

    def sync(self, event: dict[str, Any] | None = None):
        if event is not None:
            pull = scheduler.merged_pull_request(event)
            if not pull or pull.get("base", {}).get("ref") != "main":
                return {"ignored": True, "reason": "Only merges into main trigger scheduling"}
        with self.locked():
            # A dedicated main checkout: never reset/discard an operator's changes.
            if scheduler.run("git", "branch", "--show-current", cwd=self.clone) != "main":
                raise Problem(409, "The control-panel checkout must be on main")
            if scheduler.run("git", "status", "--porcelain", cwd=self.clone):
                raise Problem(409, "The control-panel checkout must be clean")
            scheduler.run("git", "pull", "--ff-only", "origin", "main", cwd=self.clone)
            base = scheduler.run("git", "rev-parse", "HEAD", cwd=self.clone)
            remote = scheduler.run("git", "rev-parse", "origin/main", cwd=self.clone)
            if base != remote:
                raise Problem(409, "The main checkout has local commits; reconcile it before syncing")
            tracker = json.loads((self.clone / scheduler.TRACKER).read_text())
            state = scheduler.load_state(self.state_file)
            scheduler.retire_completed(tracker, state)
            scheduler.write_json(self.tracker_path, tracker)
            scheduler.save_state(self.state_file, state)
            effective = scheduler.effective_tracker(tracker, state)
            done = {t["id"] for t in tracker["tasks"] if t["status"] == "completed"}
            available = [t for t in effective["tasks"]
                         if t["status"] == "pending" and set(t["depends_on"]) <= done]
            active = sum(t["status"] == "in_progress" for t in effective["tasks"])
            selected = available[:max(0, self.capacity - active)]
            # Retry incomplete preparation before claiming more work. The reservation
            # survives process death and branch helpers safely reuse an existing branch.
            pending = [t for t in tracker["tasks"] if t["id"] in state["tasks"]
                       and state["tasks"][t["id"]]["status"] == "in_progress"
                       and not state["tasks"][t["id"]].get("prepared")]
            prepared = []
            for task in pending + selected:
                record = state["tasks"].setdefault(task["id"], {})
                record.update(status="in_progress", branch=scheduler.task_branch(task["id"]),
                              owner="control-panel", updated_at=datetime.now(UTC).isoformat())
                scheduler.save_state(self.state_file, state)
                scheduler.ensure_task_branch(self.clone, "origin", "main", task["id"], base_commit=base)
                scheduler.push_opening_commit(self.clone, "origin", task["id"])
                record["prepared"] = True
                scheduler.save_state(self.state_file, state)
                prepared.append(task["id"])
            return {"prepared": prepared, "main_commit": base}

    def task(self, task_id: str, state):
        task = next((t for t in self.tracker()["tasks"] if t["id"] == task_id), None)
        if task is None:
            raise Problem(404, "Unknown task")
        record = state["tasks"].get(task_id, {})
        if task["status"] == "completed" or record.get("status") != "in_progress" or not record.get("prepared"):
            raise Problem(409, "Sync and prepare an available task before choosing a stage")
        return task, record

    def stage(self, task_id: str, stage: str, pr_number: int | None = None):
        if stage not in ("refine", "implement", "audit"):
            raise Problem(404, "Unknown stage")
        with self.locked():
            state = scheduler.load_state(self.state_file)
            task, record = self.task(task_id, state)
            template_name = "implementation" if stage == "implement" else stage
            template = scheduler.load_prompt_template(self.repo_root, Path(f"scripts/task-{template_name}-prompt.md"))
            fields = {key: task[key] for key in ("id", "title", "brief")}
            branch = scheduler.task_branch(task_id)
            if stage == "audit":
                pulls = scheduler.task_pull_requests(self.clone, task_id)
                candidates = [p for p in pulls if pr_number is None or p["number"] == pr_number]
                if len(candidates) != 1:
                    raise Problem(409, "Select one open implementation pull request into the task branch", pull_requests=pulls)
                pull = candidates[0]
                fields.update(pr_url=pull["url"], head=pull["headRefOid"])
                return {"prompt": template.substitute(fields), "pull_request": pull["url"], "head": pull["headRefOid"]}
            if not self.environment:
                raise Problem(503, "Set CODEX_ENV_ID on the control panel")
            scheduler.run("git", "fetch", "--quiet", "origin",
                          f"refs/heads/{branch}:refs/remotes/origin/{branch}", cwd=self.clone)
            head = scheduler.run("git", "rev-parse", f"origin/{branch}", cwd=self.clone)
            stages = record.setdefault("stages", {})
            key = f"{stage}:{head}"
            previous = stages.get(key)
            if previous:
                if previous["status"] == "submitted":
                    return {**previous, "reused": True}
                raise Problem(409, "The previous submission has an uncertain outcome; check Codex before clearing its stage record")
            prompt = template.substitute(fields) + f"\nBase your work and pull request on `{branch}`.\n"
            stages[key] = {"status": "submitting", "stage": stage, "head": head,
                           "updated_at": datetime.now(UTC).isoformat()}
            scheduler.save_state(self.state_file, state)
            try:
                output = scheduler.run("codex", "cloud", "exec", "--env", self.environment,
                                       "--branch", branch, prompt, cwd=self.clone)
                url = scheduler.task_url_from(output)
                if not url:
                    raise Problem(502, "Codex returned no task URL; check Codex before retrying")
            except (OSError, subprocess.SubprocessError) as error:
                # Preserve the reservation: a transport failure may happen after acceptance.
                raise Problem(502, "Codex submission failed; check Codex before clearing its stage record") from error
            stages[key].update(status="submitted", task_url=url)
            record.update(task_url=url, updated_at=datetime.now(UTC).isoformat())
            scheduler.save_state(self.state_file, state)
            return stages[key]
