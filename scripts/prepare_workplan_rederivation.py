#!/usr/bin/env python3
"""Prepare and harvest an isolated workplan generation.

The re-derivation checkout is a new local Git repository with a single root
commit, no remote, and only the new workplan generation. The agent running in
that checkout cannot inspect earlier workplan generations, source-repository
branches, or this orchestration script.

Usage:
    python scripts/prepare_workplan_rederivation.py prepare --checkout ../code-winch-v2
    # run the workplan derivation agent inside ../code-winch-v2
    python scripts/prepare_workplan_rederivation.py harvest
"""

from __future__ import annotations

import argparse
import io
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
from collections import defaultdict
from pathlib import Path
from typing import Any

WORKPLAN = Path("docs/workplan")
CURRENT = "CURRENT"
STATE_NAME = "workplan-rederive.json"
SCHEMA_SOURCE = Path("skills/workplan/tasks.schema.json")
SELF = Path("scripts/prepare_workplan_rederivation.py")
SELF_TEST = Path("tests/test_prepare_workplan_rederivation.py")
GENERATION_RE = re.compile(r"^v([1-9][0-9]*)$")
TASK_ID_RE = re.compile(r"\bP[0-9]+-[0-9]{3}\b")


class WorkplanError(RuntimeError):
    pass


def run(
    cwd: Path,
    *args: str,
    check: bool = True,
    text: bool = True,
    input_data: str | bytes | None = None,
) -> subprocess.CompletedProcess[Any]:
    result = subprocess.run(
        args,
        cwd=cwd,
        check=False,
        text=text,
        input=input_data,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    if check and result.returncode:
        stdout = result.stdout.decode() if isinstance(result.stdout, bytes) else result.stdout
        stderr = result.stderr.decode() if isinstance(result.stderr, bytes) else result.stderr
        detail = (stderr or stdout or "").strip()
        raise WorkplanError(f"{' '.join(args)} failed: {detail}")
    return result


def git(repo: Path, *args: str) -> str:
    return str(run(repo, "git", *args).stdout).strip()


def repo_root(cwd: Path) -> Path:
    return Path(git(cwd, "rev-parse", "--show-toplevel"))


def require_clean(repo: Path) -> None:
    status = git(repo, "status", "--porcelain")
    if status:
        raise WorkplanError("worktree must be clean:\n" + status)


def current_branch(repo: Path) -> str:
    branch = git(repo, "symbolic-ref", "--quiet", "--short", "HEAD")
    if not branch:
        raise WorkplanError("detached HEAD is not supported")
    return branch


def generation_number(path: Path) -> int | None:
    match = GENERATION_RE.fullmatch(path.name)
    return int(match.group(1)) if match else None


def generation_dirs(root: Path) -> list[Path]:
    return sorted(
        [p for p in root.iterdir() if p.is_dir() and generation_number(p) is not None],
        key=lambda p: generation_number(p) or 0,
    )


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as error:
        raise WorkplanError(f"cannot read {path}: {error}") from error


def save_json(path: Path, value: Any) -> None:
    path.write_text(json.dumps(value, indent=2) + "\n")


def active_generation(root: Path) -> Path:
    current = root / CURRENT
    if current.is_file():
        target = root / current.read_text().strip()
        if target.is_dir() and generation_number(target) is not None:
            return target
        raise WorkplanError(f"{current} points to an invalid generation")
    generations = generation_dirs(root)
    if not generations:
        raise WorkplanError(f"no workplan generation found under {root}")
    return generations[-1]


def archive_unversioned(root: Path) -> Path:
    if not (root / "tasks.json").is_file():
        return active_generation(root)

    destination = root / "v1"
    if destination.exists():
        raise WorkplanError(f"cannot create {destination}: it already exists")
    destination.mkdir()
    for entry in list(root.iterdir()):
        if entry == destination or entry.name == CURRENT:
            continue
        if generation_number(entry) is not None:
            continue
        shutil.move(str(entry), destination / entry.name)
    return destination


def next_generation(source: Path) -> Path:
    number = generation_number(source)
    if number is None:
        raise WorkplanError(f"not a versioned generation: {source}")
    return source.parent / f"v{number + 1}"


def completed_tasks(source: Path) -> list[dict[str, Any]]:
    tracker = load_json(source / "tasks.json")
    if not isinstance(tracker, dict) or not isinstance(tracker.get("tasks"), list):
        raise WorkplanError(f"invalid tracker: {source / 'tasks.json'}")

    tasks: list[dict[str, Any]] = []
    for raw in tracker["tasks"]:
        if raw.get("status") != "completed":
            continue
        task = dict(raw)
        task.setdefault("workplan_version", generation_number(source) or 1)
        task.setdefault("supersedes", [])
        task.setdefault("superseded_by", [])
        task.setdefault("removal_reason", None)
        task["owner"] = None
        task["blocked_reason"] = None
        tasks.append(task)

    ids = {task["id"] for task in tasks}
    for task in tasks:
        missing = [dep for dep in task.get("depends_on", []) if dep not in ids]
        if missing:
            raise WorkplanError(
                f"completed task {task['id']} depends on non-completed task(s): "
                + ", ".join(missing)
            )
    return tasks


def copy_completed_briefs(source: Path, target: Path, tasks: list[dict[str, Any]]) -> None:
    for task in tasks:
        brief = Path(str(task.get("brief", "")))
        if brief.is_absolute() or ".." in brief.parts or not brief.parts:
            raise WorkplanError(f"invalid brief path for {task.get('id')}: {brief}")
        src = source / brief
        if not src.is_file():
            raise WorkplanError(f"missing completed brief for {task['id']}: {src}")
        dst = target / brief
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)


def render_readme(generation: str, baseline: str, tasks: list[dict[str, Any]]) -> str:
    phases: dict[tuple[int, str], list[dict[str, Any]]] = defaultdict(list)
    for task in tasks:
        phases[(int(task["phase"]), str(task["phase_name"]))].append(task)

    lines = [
        f"# Implementation workplan — {generation}",
        "",
        f"Repository baseline: `{baseline}`.",
        "",
        "Task IDs are scoped to this workplan generation.",
        "",
        "## Completed implementation history",
        "",
    ]
    if not tasks:
        lines.extend(["No completed implementation tasks are recorded.", ""])
        return "\n".join(lines)

    for (phase, name), phase_tasks in sorted(phases.items()):
        lines.extend(
            [
                f"### Phase {phase} — {name}",
                "",
                "| ID | Task | Brief |",
                "|---|---|---|",
            ]
        )
        for task in phase_tasks:
            brief = task["brief"]
            lines.append(f"| {task['id']} | {task['title']} | [{brief}]({brief}) |")
        lines.append("")
    return "\n".join(lines)


def seed_generation(repo: Path, source: Path, target: Path, baseline: str) -> list[dict[str, Any]]:
    if target.exists():
        raise WorkplanError(f"target generation already exists: {target}")
    tasks = completed_tasks(source)
    target.mkdir(parents=True)
    copy_completed_briefs(source, target, tasks)

    schema = repo / SCHEMA_SOURCE
    if schema.is_file():
        shutil.copy2(schema, target / "tasks.schema.json")

    tracker = {
        "schema_version": 2,
        "status_values": [
            "pending",
            "in_progress",
            "blocked",
            "completed",
            "superseded",
            "removed",
        ],
        "tasks": tasks,
    }
    save_json(target / "tasks.json", tracker)
    (target / "README.md").write_text(render_readme(target.name, baseline, tasks))
    return tasks


def commit_workplan(repo: Path, root: Path, message: str) -> str:
    relative = root.relative_to(repo)
    git(repo, "add", "-A", str(relative))
    if not git(repo, "diff", "--cached", "--name-only"):
        raise WorkplanError("workplan preparation produced no changes")
    git(repo, "commit", "-m", message)
    return git(repo, "rev-parse", "HEAD")


def export_head(repo: Path, destination: Path) -> None:
    if destination.exists():
        if any(destination.iterdir()):
            raise WorkplanError(f"checkout destination is not empty: {destination}")
    else:
        destination.mkdir(parents=True)

    archive = run(repo, "git", "archive", "--format=tar", "HEAD", text=False).stdout
    assert isinstance(archive, bytes)
    with tarfile.open(fileobj=io.BytesIO(archive), mode="r:") as tar:
        destination_resolved = destination.resolve()
        for member in tar.getmembers():
            extracted = (destination / member.name).resolve()
            if destination_resolved not in extracted.parents and extracted != destination_resolved:
                raise WorkplanError(f"unsafe archive member: {member.name}")
        tar.extractall(destination)


def strip_agent_current_state(checkout: Path) -> None:
    path = checkout / "AGENTS.md"
    if not path.is_file():
        return
    text = path.read_text()
    marker = "\n## Current state\n"
    if marker in text:
        text = text.split(marker, 1)[0].rstrip() + "\n"
        path.write_text(text)


def sanitize_checkout(checkout: Path, generation: str) -> None:
    workplan = checkout / WORKPLAN
    for candidate in generation_dirs(workplan):
        if candidate.name != generation:
            shutil.rmtree(candidate)
    (workplan / CURRENT).write_text(generation + "\n")

    for path in (checkout / SELF, checkout / SELF_TEST):
        if path.exists():
            path.unlink()
    strip_agent_current_state(checkout)


def unfinished_ids(source_generation: Path) -> set[str]:
    tracker = load_json(source_generation / "tasks.json")
    return {
        str(task["id"])
        for task in tracker.get("tasks", [])
        if task.get("status") != "completed"
    }


def text_files(root: Path):
    ignored = {".git", "node_modules", "dist", "vendor"}
    for current, dirs, files in os.walk(root):
        dirs[:] = [name for name in dirs if name not in ignored]
        base = Path(current)
        for name in files:
            path = base / name
            try:
                data = path.read_bytes()
            except OSError:
                continue
            if b"\x00" in data:
                continue
            try:
                yield path, data.decode("utf-8")
            except UnicodeDecodeError:
                continue


def reject_unfinished_id_leaks(checkout: Path, forbidden: set[str]) -> None:
    leaks: dict[str, list[str]] = defaultdict(list)
    if not forbidden:
        return
    for path, text in text_files(checkout):
        found = sorted(set(TASK_ID_RE.findall(text)) & forbidden)
        if found:
            leaks[str(path.relative_to(checkout))].extend(found)
    if leaks:
        detail = "\n".join(
            f"- {path}: {', '.join(ids)}" for path, ids in sorted(leaks.items())
        )
        raise WorkplanError(
            "sanitized checkout still exposes unfinished task IDs:\n" + detail
        )


def init_isolated_repo(checkout: Path, branch: str) -> str:
    run(checkout, "git", "init", "-b", branch)
    run(checkout, "git", "config", "user.name", "Workplan Rederivation")
    run(checkout, "git", "config", "user.email", "workplan@local.invalid")
    run(checkout, "git", "add", "-A")
    run(checkout, "git", "commit", "-m", f"baseline for {branch}")
    return git(checkout, "rev-parse", "HEAD")


def state_path(repo: Path) -> Path:
    raw = git(repo, "rev-parse", "--git-path", STATE_NAME)
    path = Path(raw)
    return path if path.is_absolute() else repo / path


def write_state(repo: Path, state: dict[str, Any]) -> None:
    path = state_path(repo)
    path.parent.mkdir(parents=True, exist_ok=True)
    save_json(path, state)


def read_state(repo: Path) -> dict[str, Any]:
    path = state_path(repo)
    if not path.is_file():
        raise WorkplanError(f"no re-derivation state found at {path}")
    state = load_json(path)
    if not isinstance(state, dict):
        raise WorkplanError(f"invalid state file: {path}")
    return state


def validate_generation(path: Path, completed: list[dict[str, Any]]) -> None:
    tracker = load_json(path / "tasks.json")
    if not isinstance(tracker, dict) or not isinstance(tracker.get("tasks"), list):
        raise WorkplanError("derived tracker is invalid")
    ids = [task.get("id") for task in tracker["tasks"]]
    if len(ids) != len(set(ids)):
        raise WorkplanError("derived tracker contains duplicate task IDs")

    by_id = {task["id"]: task for task in tracker["tasks"]}
    for original in completed:
        current = by_id.get(original["id"])
        if current != original:
            raise WorkplanError(
                f"completed historical task {original['id']} was removed or mutated"
            )

    for task in tracker["tasks"]:
        brief = Path(str(task.get("brief", "")))
        if brief.is_absolute() or ".." in brief.parts or not (path / brief).is_file():
            raise WorkplanError(f"task {task.get('id')} has invalid brief {brief}")
        for dep in task.get("depends_on", []):
            if dep not in by_id:
                raise WorkplanError(f"task {task['id']} has unknown dependency {dep}")


def prepare(repo: Path, checkout: Path) -> dict[str, Any]:
    require_clean(repo)
    source_branch = current_branch(repo)
    workplan = repo / WORKPLAN
    baseline = git(repo, "rev-parse", "HEAD")
    source_generation = archive_unversioned(workplan)
    target_generation = next_generation(source_generation)
    completed = seed_generation(repo, source_generation, target_generation, baseline)
    (workplan / CURRENT).write_text(target_generation.name + "\n")
    archive_commit = commit_workplan(
        repo, workplan, f"chore: archive and seed workplan {target_generation.name}"
    )

    forbidden = unfinished_ids(source_generation)
    export_head(repo, checkout)
    sanitize_checkout(checkout, target_generation.name)
    reject_unfinished_id_leaks(checkout, forbidden)
    isolated_commit = init_isolated_repo(checkout, f"workplan-{target_generation.name}")

    state = {
        "source_branch": source_branch,
        "archive_commit": archive_commit,
        "baseline": baseline,
        "checkout": str(checkout.resolve()),
        "generation": target_generation.name,
        "completed": completed,
    }
    write_state(repo, state)
    print(
        f"prepared {target_generation.name}\n"
        f"history retained on: {source_branch} ({archive_commit})\n"
        f"isolated checkout:   {checkout.resolve()} ({isolated_commit})"
    )
    return state


def harvest(repo: Path) -> dict[str, Any]:
    require_clean(repo)
    state = read_state(repo)
    checkout = Path(state["checkout"])
    if not checkout.is_dir():
        raise WorkplanError(f"isolated checkout no longer exists: {checkout}")
    require_clean(checkout)

    generation = str(state["generation"])
    derived = checkout / WORKPLAN / generation
    validate_generation(derived, list(state["completed"]))

    source_branch = str(state["source_branch"])
    if current_branch(repo) != source_branch:
        git(repo, "switch", source_branch)
        require_clean(repo)

    destination = repo / WORKPLAN / generation
    if destination.exists():
        shutil.rmtree(destination)
    shutil.copytree(derived, destination)
    relative = destination.relative_to(repo)
    git(repo, "add", "-A", str(relative))
    if not git(repo, "diff", "--cached", "--name-only"):
        print(f"{generation} has no changes to harvest")
        return state
    git(repo, "commit", "-m", f"docs: rederive workplan {generation}")
    state["harvest_commit"] = git(repo, "rev-parse", "HEAD")
    write_state(repo, state)
    print(f"harvested only {relative} into {source_branch}")
    return state


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    prepare_parser = sub.add_parser("prepare")
    prepare_parser.add_argument("--checkout", type=Path, required=True)
    sub.add_parser("harvest")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        repo = repo_root(Path.cwd())
        if args.command == "prepare":
            prepare(repo, args.checkout)
        else:
            harvest(repo)
    except WorkplanError as error:
        print(f"error: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
