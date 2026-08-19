#!/usr/bin/env python3
"""Prepare and harvest a versioned workplan re-derivation.

`prepare` archives the current generation, seeds the next generation with every
inherited task, and creates a clean branch that contains only the new generation.
Completed history remains immutable legacy work. Every inherited unfinished task
is preserved so V2 must explicitly rewrite, supersede, or remove it.

`harvest` validates those dispositions and copies only the re-derived generation
back to the archive branch. It never merges the clean branch.
"""

from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any

WORKPLAN = Path("docs/workplan")
CURRENT = "CURRENT"
STATE_NAME = "workplan-rederive.json"
COMPLETED = "completed"
SUPERSEDED = "superseded"
REMOVED = "removed"
EXECUTABLE = {"pending", "in_progress", "blocked"}
GENERATION_RE = re.compile(r"^v([1-9][0-9]*)$")
V2_SCHEMA = Path("skills/workplan/tasks.schema.json")


class WorkplanError(RuntimeError):
    pass


def run_git(repo_root: Path, *args: str, capture: bool = True) -> str:
    result = subprocess.run(
        ("git", *args), cwd=repo_root, check=False, text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
    )
    if result.returncode:
        detail = (result.stderr or result.stdout or "").strip()
        raise WorkplanError(f"git {' '.join(args)} failed: {detail}")
    return (result.stdout or "").strip()


def repo_root_from(cwd: Path) -> Path:
    return Path(run_git(cwd, "rev-parse", "--show-toplevel"))


def require_clean_worktree(repo_root: Path) -> None:
    status = run_git(repo_root, "status", "--porcelain")
    if status:
        raise WorkplanError("worktree must be clean before changing workplan generations:\n" + status)


def current_branch(repo_root: Path) -> str:
    branch = run_git(repo_root, "symbolic-ref", "--quiet", "--short", "HEAD")
    if not branch:
        raise WorkplanError("cannot prepare a workplan generation from detached HEAD")
    return branch


def generation_number(path: Path) -> int | None:
    match = GENERATION_RE.fullmatch(path.name)
    return int(match.group(1)) if match else None


def generation_dirs(workplan_root: Path) -> list[Path]:
    return sorted(
        [p for p in workplan_root.iterdir() if p.is_dir() and generation_number(p) is not None],
        key=lambda p: generation_number(p) or 0,
    )


def active_generation(workplan_root: Path) -> Path:
    current_file = workplan_root / CURRENT
    if current_file.exists():
        name = current_file.read_text().strip()
        target = workplan_root / name
        if generation_number(target) is None or not target.is_dir():
            raise WorkplanError(f"{current_file} points at invalid generation {name!r}")
        return target
    generations = generation_dirs(workplan_root)
    if not generations:
        raise WorkplanError(f"no versioned workplan generation exists in {workplan_root}")
    return generations[-1]


def load_tracker(path: Path) -> dict[str, Any]:
    try:
        tracker = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as error:
        raise WorkplanError(f"cannot read tracker {path}: {error}") from error
    if not isinstance(tracker.get("tasks"), list):
        raise WorkplanError(f"tracker {path} has no task list")
    return tracker


def save_tracker(path: Path, tracker: dict[str, Any]) -> None:
    path.write_text(json.dumps(tracker, indent=2) + "\n")


def safe_relative_brief(brief: str) -> Path:
    path = Path(brief)
    if path.is_absolute() or ".." in path.parts:
        raise WorkplanError(f"task brief must stay inside its generation: {brief}")
    return path


def archive_unversioned(workplan_root: Path) -> Path:
    tracker = workplan_root / "tasks.json"
    if not tracker.exists():
        return active_generation(workplan_root)
    destination = workplan_root / "v1"
    if destination.exists():
        raise WorkplanError(f"cannot archive unversioned workplan: {destination} already exists")
    entries = [p for p in workplan_root.iterdir() if generation_number(p) is None and p.name != CURRENT]
    destination.mkdir()
    for entry in entries:
        shutil.move(str(entry), destination / entry.name)
    return destination


def copy_briefs(source: Path, target: Path, tasks: list[dict[str, Any]]) -> None:
    copied: set[Path] = set()
    for task in tasks:
        brief_value = task.get("brief")
        if not isinstance(brief_value, str) or not brief_value:
            raise WorkplanError(f"task {task.get('id')} has no brief")
        relative = safe_relative_brief(brief_value)
        if relative in copied:
            continue
        source_brief = source / relative
        if not source_brief.is_file():
            raise WorkplanError(f"task {task.get('id')} points at missing brief {source_brief}")
        target_brief = target / relative
        target_brief.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source_brief, target_brief)
        copied.add(relative)


def validate_completed_dependencies(tasks: list[dict[str, Any]]) -> None:
    status = {task.get("id"): task.get("status") for task in tasks}
    for task in tasks:
        if task.get("status") != COMPLETED:
            continue
        for dependency in task.get("depends_on", []):
            if status.get(dependency) != COMPLETED:
                raise WorkplanError(
                    f"completed task {task.get('id')} depends on non-completed task {dependency}; "
                    "cannot seed truthful history"
                )


def normalize_v2_task(task: dict[str, Any]) -> dict[str, Any]:
    result = dict(task)
    result.setdefault("workplan_version", 1)
    result.setdefault("supersedes", [])
    result.setdefault("superseded_by", [])
    result.setdefault("removal_reason", None)
    return result


def seed_v2_tasks(tasks: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], list[str]]:
    seeded: list[dict[str, Any]] = []
    inherited_unfinished: list[str] = []
    for original in tasks:
        task = normalize_v2_task(original)
        if task.get("status") == COMPLETED:
            task["workplan_version"] = int(task.get("workplan_version") or 1)
        else:
            inherited_unfinished.append(str(task["id"]))
            # A non-completed inherited task has not yet received a V2 planning
            # disposition. Keep its old contract version until the planner rewrites it.
            task["workplan_version"] = int(task.get("workplan_version") or 1)
            task["owner"] = None
            if task["status"] == "in_progress":
                task["status"] = "pending"
            if task["status"] != "blocked":
                task["blocked_reason"] = None
        seeded.append(task)
    return seeded, inherited_unfinished


def render_readme(generation: str, tasks: list[dict[str, Any]], baseline_sha: str) -> str:
    by_phase: dict[tuple[int, str], list[dict[str, Any]]] = defaultdict(list)
    for task in tasks:
        by_phase[(int(task["phase"]), str(task["phase_name"]))].append(task)
    lines = [
        f"# Implementation workplan — {generation}", "",
        f"This generation was seeded from repository baseline `{baseline_sha}`.",
        "It is a V2 re-derivation workspace. Completed task history is carried forward",
        "without retroactive V2 brief requirements. Every inherited unfinished task is",
        "present and must be rewritten, superseded, or removed before harvest.", "",
        "## Inherited task inventory", "",
        "| ID | Task | Status | Workplan version | Brief |", "|---|---|---|---:|---|",
    ]
    for (_, _), phase_tasks in sorted(by_phase.items()):
        for task in phase_tasks:
            brief = task["brief"]
            lines.append(
                f"| {task['id']} | {task['title']} | {task['status']} | "
                f"{task.get('workplan_version', 1)} | [{brief}]({brief}) |"
            )
    lines.extend([
        "", "## Re-derivation report", "",
        "Record failing HEAD invariants and repair owners, every inherited unfinished",
        "task disposition, dependency-reason changes, critical paths and average widths,",
        "write collisions, contract collisions, and any legacy task relevant to a current",
        "invariant gap.", "",
    ])
    return "\n".join(lines)


def copy_v2_schema(repo_root: Path, source: Path, target: Path) -> None:
    schema = repo_root / V2_SCHEMA
    if schema.is_file():
        shutil.copy2(schema, target / "tasks.schema.json")
        return
    legacy = source / "tasks.schema.json"
    if legacy.is_file():
        shutil.copy2(legacy, target / "tasks.schema.json")
        return
    raise WorkplanError(f"no V2 tracker schema found at {schema}")


def seed_next_generation(repo_root: Path, source: Path, target: Path, baseline_sha: str) -> tuple[list[dict[str, Any]], list[str]]:
    if target.exists():
        raise WorkplanError(f"target generation already exists: {target}")
    source_tracker = load_tracker(source / "tasks.json")
    source_tasks = [dict(task) for task in source_tracker["tasks"]]
    validate_completed_dependencies(source_tasks)
    tasks, inherited_unfinished = seed_v2_tasks(source_tasks)
    target.mkdir(parents=True)
    copy_briefs(source, target, tasks)
    copy_v2_schema(repo_root, source, target)
    tracker = {
        "schema_version": 2,
        "status_values": ["pending", "in_progress", "blocked", "completed", "superseded", "removed"],
        "tasks": tasks,
    }
    save_tracker(target / "tasks.json", tracker)
    (target / "README.md").write_text(render_readme(target.name, tasks, baseline_sha))
    return tasks, inherited_unfinished


def sanitize_generations(workplan_root: Path, keep: Path) -> None:
    keep = keep.resolve()
    for generation in generation_dirs(workplan_root):
        if generation.resolve() != keep:
            shutil.rmtree(generation)


def next_generation(source: Path) -> tuple[int, str]:
    number = generation_number(source)
    if number is None:
        raise WorkplanError(f"source is not a versioned generation: {source}")
    number += 1
    return number, f"v{number}"


def branch_exists(repo_root: Path, branch: str) -> bool:
    return subprocess.run(
        ("git", "show-ref", "--verify", "--quiet", f"refs/heads/{branch}"),
        cwd=repo_root, check=False,
    ).returncode == 0


def state_path(repo_root: Path) -> Path:
    raw = run_git(repo_root, "rev-parse", "--git-path", STATE_NAME)
    path = Path(raw)
    return path if path.is_absolute() else repo_root / path


def write_state(repo_root: Path, state: dict[str, Any]) -> None:
    path = state_path(repo_root)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(state, indent=2) + "\n")


def read_state(repo_root: Path) -> dict[str, Any]:
    path = state_path(repo_root)
    if not path.is_file():
        raise WorkplanError(f"no re-derivation state found at {path}; run prepare first")
    try:
        return json.loads(path.read_text())
    except json.JSONDecodeError as error:
        raise WorkplanError(f"invalid re-derivation state at {path}: {error}") from error


def commit_path(repo_root: Path, path: Path, message: str) -> str:
    relative = path.relative_to(repo_root)
    run_git(repo_root, "add", "-A", str(relative))
    if not run_git(repo_root, "diff", "--cached", "--name-only"):
        raise WorkplanError(f"nothing changed under {relative}; refusing empty commit")
    run_git(repo_root, "commit", "-m", message)
    return run_git(repo_root, "rev-parse", "HEAD")


def planned_generation(workplan_root: Path) -> str:
    if (workplan_root / "tasks.json").exists():
        return "v2"
    _, name = next_generation(active_generation(workplan_root))
    return name


def validate_v2_rederivation(generation: Path, inherited_unfinished: list[str]) -> None:
    tracker = load_tracker(generation / "tasks.json")
    if tracker.get("schema_version") != 2:
        raise WorkplanError(f"{generation / 'tasks.json'} must use schema_version 2")
    tasks = tracker["tasks"]
    by_id = {task.get("id"): task for task in tasks}
    if len(by_id) != len(tasks):
        raise WorkplanError("V2 tracker contains duplicate task IDs")

    errors: list[str] = []
    allowed = {"pending", "in_progress", "blocked", "completed", SUPERSEDED, REMOVED}
    for task in tasks:
        task_id = str(task.get("id"))
        status = task.get("status")
        if status not in allowed:
            errors.append(f"{task_id}: unsupported status {status!r}")
        if status in EXECUTABLE and task.get("workplan_version") != 2:
            errors.append(f"{task_id}: executable task must be rewritten to workplan_version 2")
        if status == SUPERSEDED:
            replacements = task.get("superseded_by") or []
            if not replacements:
                errors.append(f"{task_id}: superseded task must name superseded_by")
            for replacement in replacements:
                if replacement not in by_id:
                    errors.append(f"{task_id}: replacement {replacement} does not exist")
        if status == REMOVED and not str(task.get("removal_reason") or "").strip():
            errors.append(f"{task_id}: removed task must record removal_reason")
        if status not in {SUPERSEDED, REMOVED} and task.get("superseded_by"):
            errors.append(f"{task_id}: only superseded tasks may set superseded_by")
        if status != REMOVED and task.get("removal_reason") is not None:
            errors.append(f"{task_id}: only removed tasks may set removal_reason")
        for dependency in task.get("depends_on", []):
            if dependency not in by_id:
                errors.append(f"{task_id}: unknown dependency {dependency}")
            elif status in EXECUTABLE and by_id[dependency].get("status") in {SUPERSEDED, REMOVED}:
                errors.append(f"{task_id}: executable dependency {dependency} is terminal planning history")
        for old in task.get("supersedes", []):
            if old not in by_id:
                errors.append(f"{task_id}: supersedes unknown task {old}")

    for task_id in inherited_unfinished:
        task = by_id.get(task_id)
        if task is None:
            errors.append(f"{task_id}: inherited unfinished task disappeared from tracker")
            continue
        status = task.get("status")
        rewritten = status in EXECUTABLE and task.get("workplan_version") == 2
        dispositioned = status in {COMPLETED, SUPERSEDED, REMOVED}
        if not (rewritten or dispositioned):
            errors.append(f"{task_id}: inherited unfinished task has no V2 disposition")

    if errors:
        raise WorkplanError("V2 re-derivation is incomplete:\n- " + "\n- ".join(errors))


def prepare(repo_root: Path, workplan_relative: Path, clean_branch_name: str | None) -> dict[str, Any]:
    require_clean_worktree(repo_root)
    archive_branch = current_branch(repo_root)
    workplan_root = repo_root / workplan_relative
    if not workplan_root.is_dir():
        raise WorkplanError(f"workplan directory does not exist: {workplan_root}")
    generation_name = planned_generation(workplan_root)
    clean_branch = clean_branch_name or f"workplan-rederive-{generation_name}"
    if branch_exists(repo_root, clean_branch):
        raise WorkplanError(f"clean branch already exists: {clean_branch}")
    baseline_sha = run_git(repo_root, "rev-parse", "HEAD")
    source = archive_unversioned(workplan_root)
    _, actual_generation_name = next_generation(source)
    if actual_generation_name != generation_name:
        raise WorkplanError(
            f"generation changed during preparation: expected {generation_name}, got {actual_generation_name}"
        )
    target = workplan_root / generation_name
    tasks, inherited_unfinished = seed_next_generation(repo_root, source, target, baseline_sha)
    (workplan_root / CURRENT).write_text(generation_name + "\n")
    archive_commit = commit_path(repo_root, workplan_root, f"chore: archive and seed workplan {generation_name}")
    run_git(repo_root, "switch", "-c", clean_branch)
    sanitize_generations(workplan_root, target)
    clean_commit = commit_path(repo_root, workplan_root, f"chore: sanitize workplan for {generation_name} re-derivation")
    state = {
        "archive_branch": archive_branch,
        "archive_commit": archive_commit,
        "baseline_sha": baseline_sha,
        "clean_branch": clean_branch,
        "clean_commit": clean_commit,
        "generation": generation_name,
        "workplan_root": workplan_relative.as_posix(),
        "inherited_unfinished": inherited_unfinished,
    }
    write_state(repo_root, state)
    print(
        f"prepared {generation_name}: {len(tasks)} inherited tasks preserved; "
        f"{len(inherited_unfinished)} require disposition\n"
        f"archive branch: {archive_branch} ({archive_commit})\n"
        f"clean branch:   {clean_branch} ({clean_commit})\n"
        f"active plan:    {workplan_relative / generation_name}"
    )
    return state


def restore_generation(repo_root: Path, source_ref: str, generation_path: Path) -> None:
    relative = generation_path.relative_to(repo_root)
    run_git(repo_root, "restore", "--source", source_ref, "--staged", "--worktree", "--", str(relative))


def harvest(repo_root: Path) -> dict[str, Any]:
    require_clean_worktree(repo_root)
    state = read_state(repo_root)
    archive_branch = str(state["archive_branch"])
    clean_branch = str(state["clean_branch"])
    generation = str(state["generation"])
    workplan_relative = Path(str(state["workplan_root"]))
    generation_path = repo_root / workplan_relative / generation
    if not branch_exists(repo_root, archive_branch):
        raise WorkplanError(f"archive branch no longer exists: {archive_branch}")
    if not branch_exists(repo_root, clean_branch):
        raise WorkplanError(f"clean branch no longer exists: {clean_branch}")
    validate_v2_rederivation(generation_path, list(state.get("inherited_unfinished", [])))
    clean_head = run_git(repo_root, "rev-parse", clean_branch)
    run_git(repo_root, "switch", archive_branch)
    restore_generation(repo_root, clean_head, generation_path)
    relative = generation_path.relative_to(repo_root)
    changed = run_git(repo_root, "diff", "--cached", "--name-only", "--", str(relative))
    if not changed:
        print(f"{generation} has no changes to harvest")
        return state
    run_git(repo_root, "commit", "-m", f"docs: rederive workplan {generation}")
    harvest_commit = run_git(repo_root, "rev-parse", "HEAD")
    state["harvest_commit"] = harvest_commit
    state["clean_head"] = clean_head
    write_state(repo_root, state)
    print(
        f"harvested only {workplan_relative / generation} from {clean_branch}\n"
        f"archive branch: {archive_branch} ({harvest_commit})"
    )
    return state


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--workplan-root", type=Path, default=WORKPLAN,
        help=f"workplan directory relative to repository root (default: {WORKPLAN})",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)
    prepare_parser = subparsers.add_parser("prepare", help="archive current generation and create the clean branch")
    prepare_parser.add_argument("--clean-branch", help="local branch for clean re-derivation (default: workplan-rederive-vN)")
    subparsers.add_parser("harvest", help="validate and copy the re-derived generation back to the archive branch")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        repo_root = repo_root_from(Path.cwd())
        if args.command == "prepare":
            prepare(repo_root, args.workplan_root, args.clean_branch)
        else:
            harvest(repo_root)
    except WorkplanError as error:
        print(f"error: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
