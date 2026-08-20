#!/usr/bin/env python3
"""Prepare and harvest an isolated workplan generation.

This is operator tooling. `prepare` constructs and validates an isolated planning
checkout before mutating the source repository. `harvest` validates the complete
derived generation and refuses source-branch drift.
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
import tempfile
from collections import defaultdict
from pathlib import Path
from typing import Any, Iterable

WORKPLAN = Path("docs/workplan")
CURRENT = "CURRENT"
STATE_NAME = "workplan-rederive.json"
SCHEMA_SOURCE = Path("skills/workplan/tasks.schema.json")
GENERATION_RE = re.compile(r"^v([1-9][0-9]*)$")
TASK_ID_RE = re.compile(r"\bP[0-9]+-[0-9]{3}\b")
OPERATOR_ONLY_PATHS = (
    Path("scripts/prepare_workplan_rederivation.py"),
    Path("tests/test_prepare_workplan_rederivation.py"),
    Path("tests/test_task_scheduler.py"),
    Path("tests/test_stamp_task_completion.py"),
    Path("docs/workplan/rederivation-operator.md"),
    Path("docs/workplan/workplan-v2-review-rejection.md"),
    Path("docs/workplan/workplan-v2-ideas.md"),
)
SNAPSHOT_SECTIONS = {
    "Objective",
    "Scope",
    "Runtime reachability",
    "Demonstration",
    "Verification",
    "Acceptance criteria",
    "Traces to",
}
TERMINAL_PLANNING = {"superseded", "removed"}
EXECUTION_STATUSES = {"pending", "in_progress", "blocked", "completed"}


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
    path.parent.mkdir(parents=True, exist_ok=True)
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
        if entry == destination or entry.name == CURRENT or generation_number(entry) is not None:
            continue
        shutil.move(str(entry), destination / entry.name)
    return destination


def next_generation(source: Path) -> Path:
    number = generation_number(source)
    if number is None:
        raise WorkplanError(f"not a versioned generation: {source}")
    return source.parent / f"v{number + 1}"


def safe_brief_path(task: dict[str, Any]) -> Path:
    brief = Path(str(task.get("brief", "")))
    if brief.is_absolute() or ".." in brief.parts or not brief.parts:
        raise WorkplanError(f"invalid brief path for {task.get('id')}: {brief}")
    return brief


def unfinished_ids(source: Path) -> set[str]:
    tracker = load_json(source / "tasks.json")
    if not isinstance(tracker, dict) or not isinstance(tracker.get("tasks"), list):
        raise WorkplanError(f"invalid tracker: {source / 'tasks.json'}")
    return {str(task["id"]) for task in tracker["tasks"] if task.get("status") != "completed"}


def section_map(text: str) -> dict[str, list[str]]:
    sections: dict[str, list[str]] = defaultdict(list)
    current: str | None = None
    for line in text.splitlines():
        if line.startswith("## "):
            current = line[3:].strip()
            continue
        if current is not None:
            sections[current].append(line)
    return sections


def remove_forbidden_lines(lines: Iterable[str], forbidden: set[str]) -> list[str]:
    result: list[str] = []
    for line in lines:
        if set(TASK_ID_RE.findall(line)) & forbidden:
            continue
        result.append(line.rstrip())
    while result and not result[0]:
        result.pop(0)
    while result and not result[-1]:
        result.pop()
    return result


def render_completed_snapshot(
    task: dict[str, Any], source_text: str, forbidden: set[str]
) -> str:
    sections = section_map(source_text)
    lines = [
        f"# {task['id']}: {task['title']} — completed implementation record",
        "",
        f"**Phase:** {task['phase']} — {task['phase_name']}",
        "**Status:** completed",
        f"**Workplan version:** {task.get('workplan_version', 1)}",
        "",
        "This record contains implementation facts for completed work.",
        "",
    ]
    for heading in (
        "Objective",
        "Scope",
        "Runtime reachability",
        "Demonstration",
        "Verification",
        "Acceptance criteria",
        "Traces to",
    ):
        if heading not in SNAPSHOT_SECTIONS:
            continue
        content = remove_forbidden_lines(sections.get(heading, []), forbidden)
        if not content:
            continue
        lines.extend([f"## {heading}", "", *content, ""])
    rendered = "\n".join(lines).rstrip() + "\n"
    leaked = set(TASK_ID_RE.findall(rendered)) & forbidden
    if leaked:
        raise WorkplanError(
            f"completed snapshot for {task['id']} exposes unfinished task IDs: "
            + ", ".join(sorted(leaked))
        )
    return rendered


def completed_history(
    source: Path, target: Path, forbidden: set[str]
) -> list[dict[str, Any]]:
    tracker = load_json(source / "tasks.json")
    if not isinstance(tracker, dict) or not isinstance(tracker.get("tasks"), list):
        raise WorkplanError(f"invalid tracker: {source / 'tasks.json'}")
    completed_raw = [
        dict(task) for task in tracker["tasks"] if task.get("status") == "completed"
    ]
    completed_ids = {str(task["id"]) for task in completed_raw}
    result: list[dict[str, Any]] = []
    for raw in completed_raw:
        missing = [dep for dep in raw.get("depends_on", []) if dep not in completed_ids]
        if missing:
            raise WorkplanError(
                f"completed task {raw['id']} depends on non-completed task(s): "
                + ", ".join(missing)
            )
        source_brief = source / safe_brief_path(raw)
        if not source_brief.is_file():
            raise WorkplanError(f"missing completed brief for {raw['id']}: {source_brief}")
        phase = int(raw["phase"])
        snapshot = Path(
            f"phase-{phase}/{str(raw['id']).lower()}-completed-history.md"
        )
        task = {
            "id": raw["id"],
            "title": raw["title"],
            "phase": phase,
            "phase_name": raw["phase_name"],
            "status": "completed",
            "depends_on": list(raw.get("depends_on", [])),
            "owner": None,
            "blocked_reason": None,
            "brief": snapshot.as_posix(),
            "workplan_version": int(
                raw.get("workplan_version") or generation_number(source) or 1
            ),
            "supersedes": [],
            "superseded_by": [],
            "removal_reason": None,
        }
        destination = target / snapshot
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(
            render_completed_snapshot(task, source_brief.read_text(), forbidden)
        )
        result.append(task)
    return result


def render_readme(
    generation: str, baseline: str, tasks: list[dict[str, Any]]
) -> str:
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
        "## Completed implementation facts",
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
                "| ID | Task | Record |",
                "|---|---|---|",
            ]
        )
        for task in phase_tasks:
            brief = task["brief"]
            lines.append(f"| {task['id']} | {task['title']} | [{brief}]({brief}) |")
        lines.append("")
    return "\n".join(lines)


def seed_generation(
    repo: Path,
    source: Path,
    target: Path,
    baseline: str,
    forbidden: set[str],
) -> list[dict[str, Any]]:
    if target.exists():
        raise WorkplanError(f"target generation already exists: {target}")
    target.mkdir(parents=True)
    completed = completed_history(source, target, forbidden)
    schema = repo / SCHEMA_SOURCE
    if not schema.is_file():
        raise WorkplanError(f"missing V2 tracker schema: {schema}")
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
        "tasks": completed,
    }
    save_json(target / "tasks.json", tracker)
    (target / "README.md").write_text(
        render_readme(target.name, baseline, completed)
    )
    return completed


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
            if (
                destination_resolved not in extracted.parents
                and extracted != destination_resolved
            ):
                raise WorkplanError(f"unsafe archive member: {member.name}")
        tar.extractall(destination)


def strip_agent_current_state(checkout: Path) -> None:
    path = checkout / "AGENTS.md"
    if not path.is_file():
        return
    text = path.read_text()
    marker = "\n## Current state\n"
    if marker in text:
        path.write_text(text.split(marker, 1)[0].rstrip() + "\n")


def sanitize_checkout(checkout: Path, staged_workplan: Path, generation: str) -> None:
    workplan = checkout / WORKPLAN
    if workplan.exists():
        shutil.rmtree(workplan)
    shutil.copytree(staged_workplan, workplan)
    for entry in list(workplan.iterdir()):
        if entry.name in {generation, CURRENT}:
            continue
        if entry.is_dir():
            shutil.rmtree(entry)
        else:
            entry.unlink()
    (workplan / CURRENT).write_text(generation + "\n")
    for relative in OPERATOR_ONLY_PATHS:
        path = checkout / relative
        if path.is_dir():
            shutil.rmtree(path)
        elif path.exists():
            path.unlink()
    strip_agent_current_state(checkout)


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
            "isolated checkout still exposes unfinished task IDs:\n" + detail
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


def _schema_type_matches(value: Any, expected: str) -> bool:
    if expected == "object":
        return isinstance(value, dict)
    if expected == "array":
        return isinstance(value, list)
    if expected == "string":
        return isinstance(value, str)
    if expected == "null":
        return value is None
    if expected == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if expected == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    if expected == "boolean":
        return isinstance(value, bool)
    return False


def _resolve_ref(root: dict[str, Any], ref: str) -> dict[str, Any]:
    if not ref.startswith("#/"):
        raise WorkplanError(f"unsupported schema reference {ref!r}")
    value: Any = root
    for raw in ref[2:].split("/"):
        key = raw.replace("~1", "/").replace("~0", "~")
        value = value[key]
    if not isinstance(value, dict):
        raise WorkplanError(f"schema reference {ref!r} does not resolve to an object")
    return value


def _schema_errors(
    value: Any, schema: dict[str, Any], root: dict[str, Any], path: str = "$"
) -> list[str]:
    errors: list[str] = []
    if "$ref" in schema:
        return _schema_errors(value, _resolve_ref(root, str(schema["$ref"])), root, path)
    if "const" in schema and value != schema["const"]:
        errors.append(f"{path}: expected constant {schema['const']!r}")
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{path}: value {value!r} is not in enum")
    if "type" in schema:
        expected = schema["type"] if isinstance(schema["type"], list) else [schema["type"]]
        if not any(_schema_type_matches(value, str(item)) for item in expected):
            errors.append(f"{path}: expected type {' or '.join(map(str, expected))}")
            return errors
    if isinstance(value, dict):
        required = schema.get("required", [])
        for key in required:
            if key not in value:
                errors.append(f"{path}: missing required property {key}")
        properties = schema.get("properties", {})
        if isinstance(properties, dict):
            for key, child_schema in properties.items():
                if key in value and isinstance(child_schema, dict):
                    errors.extend(
                        _schema_errors(value[key], child_schema, root, f"{path}.{key}")
                    )
            if schema.get("additionalProperties") is False:
                unknown = sorted(set(value) - set(properties))
                for key in unknown:
                    errors.append(f"{path}: additional property {key} is not allowed")
    if isinstance(value, list):
        if "minItems" in schema and len(value) < int(schema["minItems"]):
            errors.append(f"{path}: expected at least {schema['minItems']} items")
        if "maxItems" in schema and len(value) > int(schema["maxItems"]):
            errors.append(f"{path}: expected at most {schema['maxItems']} items")
        if schema.get("uniqueItems"):
            seen: set[str] = set()
            for item in value:
                encoded = json.dumps(item, sort_keys=True)
                if encoded in seen:
                    errors.append(f"{path}: items must be unique")
                    break
                seen.add(encoded)
        item_schema = schema.get("items")
        if isinstance(item_schema, dict):
            for index, item in enumerate(value):
                errors.extend(
                    _schema_errors(item, item_schema, root, f"{path}[{index}]")
                )
    if isinstance(value, str):
        if "minLength" in schema and len(value) < int(schema["minLength"]):
            errors.append(f"{path}: string is shorter than {schema['minLength']}")
        if "pattern" in schema and not re.search(str(schema["pattern"]), value):
            errors.append(
                f"{path}: value {value!r} does not match {schema['pattern']!r}"
            )
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        if "minimum" in schema and value < schema["minimum"]:
            errors.append(f"{path}: value is below minimum {schema['minimum']}")
        if "maximum" in schema and value > schema["maximum"]:
            errors.append(f"{path}: value is above maximum {schema['maximum']}")
    for child in schema.get("allOf", []):
        if isinstance(child, dict):
            errors.extend(_schema_errors(value, child, root, path))
    if "if" in schema and isinstance(schema["if"], dict):
        condition_holds = not _schema_errors(value, schema["if"], root, path)
        branch = schema.get("then") if condition_holds else schema.get("else")
        if isinstance(branch, dict):
            errors.extend(_schema_errors(value, branch, root, path))
    return errors


def validate_schema(instance: Any, schema: dict[str, Any]) -> None:
    errors = _schema_errors(instance, schema, schema)
    if errors:
        raise WorkplanError(
            "tasks.json does not validate against tasks.schema.json:\n- "
            + "\n- ".join(errors)
        )


def dependency_cycle(by_id: dict[str, dict[str, Any]]) -> list[str] | None:
    visiting: set[str] = set()
    visited: set[str] = set()
    stack: list[str] = []

    def visit(task_id: str) -> list[str] | None:
        if task_id in visiting:
            start = stack.index(task_id)
            return stack[start:] + [task_id]
        if task_id in visited:
            return None
        visiting.add(task_id)
        stack.append(task_id)
        for dependency in by_id[task_id].get("depends_on", []):
            cycle = visit(dependency)
            if cycle:
                return cycle
        stack.pop()
        visiting.remove(task_id)
        visited.add(task_id)
        return None

    for task_id in by_id:
        cycle = visit(task_id)
        if cycle:
            return cycle
    return None


def validate_generation(
    workplan_root: Path, generation: str, completed: list[dict[str, Any]]
) -> None:
    current = workplan_root / CURRENT
    if not current.is_file() or current.read_text().strip() != generation:
        raise WorkplanError(f"active-generation marker must name {generation}")
    path = workplan_root / generation
    if not path.is_dir():
        raise WorkplanError(f"derived generation does not exist: {path}")
    schema = load_json(path / "tasks.schema.json")
    tracker = load_json(path / "tasks.json")
    if not isinstance(schema, dict) or not isinstance(tracker, dict):
        raise WorkplanError("derived schema or tracker is invalid")
    validate_schema(tracker, schema)

    tasks = tracker["tasks"]
    ids = [task["id"] for task in tasks]
    if len(ids) != len(set(ids)):
        raise WorkplanError("derived tracker contains duplicate task IDs")
    by_id = {task["id"]: task for task in tasks}
    completed_by_id = {task["id"]: task for task in completed}

    for original in completed:
        if by_id.get(original["id"]) != original:
            raise WorkplanError(
                f"completed implementation record {original['id']} was removed or mutated"
            )
    unexpected_completed = sorted(
        task["id"]
        for task in tasks
        if task["status"] == "completed" and task["id"] not in completed_by_id
    )
    if unexpected_completed:
        raise WorkplanError(
            "derived plan introduced unverified completed task(s): "
            + ", ".join(unexpected_completed)
        )

    for task in tasks:
        task_id = task["id"]
        brief = safe_brief_path(task)
        if not (path / brief).is_file():
            raise WorkplanError(f"task {task_id} references missing brief {brief}")
        if task["status"] != "completed" and task.get("workplan_version") != 2:
            raise WorkplanError(f"task {task_id} must use workplan_version 2")
        for dependency in task.get("depends_on", []):
            if dependency not in by_id:
                raise WorkplanError(f"task {task_id} has unknown dependency {dependency}")
            if (
                task["status"] in EXECUTION_STATUSES - {"completed"}
                and by_id[dependency]["status"] in TERMINAL_PLANNING
            ):
                raise WorkplanError(
                    f"task {task_id} depends on terminal planning record {dependency}"
                )
        if task_id in task.get("depends_on", []):
            raise WorkplanError(f"task {task_id} depends on itself")

        if task["status"] == "superseded":
            for replacement in task.get("superseded_by", []):
                if replacement not in by_id:
                    raise WorkplanError(
                        f"task {task_id} names unknown replacement {replacement}"
                    )
                if task_id not in by_id[replacement].get("supersedes", []):
                    raise WorkplanError(
                        f"replacement {replacement} does not reciprocally supersede {task_id}"
                    )
        for old in task.get("supersedes", []):
            if old not in by_id:
                raise WorkplanError(f"task {task_id} supersedes unknown task {old}")
            if (
                by_id[old]["status"] != "superseded"
                or task_id not in by_id[old].get("superseded_by", [])
            ):
                raise WorkplanError(
                    f"task {task_id} has invalid supersedes relationship to {old}"
                )

    cycle = dependency_cycle(by_id)
    if cycle:
        raise WorkplanError("dependency cycle: " + " -> ".join(cycle))


def replace_tree(source: Path, destination: Path) -> None:
    if destination.exists():
        shutil.rmtree(destination)
    shutil.copytree(source, destination)


def rollback_source(repo: Path, baseline: str) -> None:
    run(repo, "git", "reset", "--hard", baseline, check=False)
    run(repo, "git", "clean", "-fd", "--", str(WORKPLAN), check=False)


def prepare(repo: Path, checkout: Path) -> dict[str, Any]:
    require_clean(repo)
    source_branch = current_branch(repo)
    workplan = repo / WORKPLAN
    baseline = git(repo, "rev-parse", "HEAD")
    checkout = checkout.resolve()
    if checkout.exists() and any(checkout.iterdir()):
        raise WorkplanError(f"checkout destination is not empty: {checkout}")

    with tempfile.TemporaryDirectory(prefix="workplan-prepare-") as temporary:
        staged_workplan = Path(temporary) / "workplan"
        shutil.copytree(workplan, staged_workplan)
        source_generation = archive_unversioned(staged_workplan)
        target_generation = next_generation(source_generation)
        forbidden = unfinished_ids(source_generation)
        completed = seed_generation(
            repo, source_generation, target_generation, baseline, forbidden
        )
        (staged_workplan / CURRENT).write_text(target_generation.name + "\n")

        try:
            export_head(repo, checkout)
            sanitize_checkout(checkout, staged_workplan, target_generation.name)
            reject_unfinished_id_leaks(checkout, forbidden)
            isolated_commit = init_isolated_repo(
                checkout, f"workplan-{target_generation.name}"
            )
            replace_tree(staged_workplan, workplan)
            archive_commit = commit_workplan(
                repo,
                workplan,
                f"chore: archive and seed workplan {target_generation.name}",
            )
            state = {
                "source_branch": source_branch,
                "archive_commit": archive_commit,
                "baseline": baseline,
                "checkout": str(checkout),
                "generation": target_generation.name,
                "completed": completed,
            }
            write_state(repo, state)
        except Exception:
            rollback_source(repo, baseline)
            shutil.rmtree(checkout, ignore_errors=True)
            raise

    print(
        f"prepared {state['generation']}\n"
        f"source branch:     {source_branch} ({archive_commit})\n"
        f"isolated checkout: {checkout} ({isolated_commit})"
    )
    return state


def harvest(repo: Path) -> dict[str, Any]:
    require_clean(repo)
    state = read_state(repo)
    source_branch = str(state["source_branch"])
    if current_branch(repo) != source_branch:
        git(repo, "switch", source_branch)
        require_clean(repo)
    current_head = git(repo, "rev-parse", "HEAD")
    if current_head != state["archive_commit"]:
        raise WorkplanError(
            "source branch moved after prepare; re-derive or explicitly revalidate against the new HEAD"
        )

    checkout = Path(state["checkout"])
    if not checkout.is_dir():
        raise WorkplanError(f"isolated checkout no longer exists: {checkout}")
    require_clean(checkout)
    generation = str(state["generation"])
    validate_generation(checkout / WORKPLAN, generation, list(state["completed"]))

    destination = repo / WORKPLAN / generation
    replace_tree(checkout / WORKPLAN / generation, destination)
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
