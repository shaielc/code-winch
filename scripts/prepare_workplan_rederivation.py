#!/usr/bin/env python3
"""Prepare and harvest a clean-slate, versioned workplan re-derivation.

`prepare` archives the current workplan generation, creates the next generation
from completed implementation history only, commits that history on the current
branch, then creates a derivation branch containing only the new generation.

`harvest` validates the complete newly derived generation and copies only that
generation back to the archive branch. It never merges the derivation branch.
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
GENERATION_RE = re.compile(r"^v([1-9][0-9]*)$")
V2_SCHEMA = Path("skills/workplan/tasks.schema.json")
STATUS_VALUES = ["pending", "in_progress", "blocked", "completed", "superseded", "removed"]


class WorkplanError(RuntimeError):
    pass


def run_git(repo_root: Path, *args: str, capture: bool = True) -> str:
    result = subprocess.run(
        ("git", *args),
        cwd=repo_root,
        check=False,
        text=True,
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
        raise WorkplanError(
            "worktree must be clean before changing workplan generations:\n" + status
        )


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
        [
            path
            for path in workplan_root.iterdir()
            if path.is_dir() and generation_number(path) is not None
        ],
        key=lambda path: generation_number(path) or 0,
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


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as error:
        raise WorkplanError(f"cannot read JSON {path}: {error}") from error


def load_tracker(path: Path) -> dict[str, Any]:
    tracker = load_json(path)
    if not isinstance(tracker, dict) or not isinstance(tracker.get("tasks"), list):
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
        raise WorkplanError(
            f"cannot archive unversioned workplan: {destination} already exists"
        )

    entries = [
        path
        for path in workplan_root.iterdir()
        if generation_number(path) is None and path.name != CURRENT
    ]
    destination.mkdir()
    for entry in entries:
        shutil.move(str(entry), destination / entry.name)
    return destination


def normalize_completed_task(task: dict[str, Any]) -> dict[str, Any]:
    result = dict(task)
    result.setdefault("workplan_version", 1)
    result.setdefault("supersedes", [])
    result.setdefault("superseded_by", [])
    result.setdefault("removal_reason", None)
    result["owner"] = None
    result["blocked_reason"] = None
    return result


def completed_history(tasks: list[dict[str, Any]]) -> list[dict[str, Any]]:
    completed = [
        normalize_completed_task(task)
        for task in tasks
        if task.get("status") == COMPLETED
    ]
    completed_ids = {task["id"] for task in completed}
    for task in completed:
        for dependency in task.get("depends_on", []):
            if dependency not in completed_ids:
                raise WorkplanError(
                    f"completed task {task.get('id')} depends on non-completed task "
                    f"{dependency}; cannot seed truthful completed history"
                )
    return completed


def copy_briefs(source: Path, target: Path, tasks: list[dict[str, Any]]) -> None:
    copied: set[Path] = set()
    for task in tasks:
        brief = task.get("brief")
        if not isinstance(brief, str) or not brief:
            raise WorkplanError(f"task {task.get('id')} has no brief")
        relative = safe_relative_brief(brief)
        if relative in copied:
            continue
        source_brief = source / relative
        if not source_brief.is_file():
            raise WorkplanError(
                f"task {task.get('id')} points at missing brief {source_brief}"
            )
        target_brief = target / relative
        target_brief.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source_brief, target_brief)
        copied.add(relative)


def render_readme(
    generation: str, tasks: list[dict[str, Any]], baseline_sha: str
) -> str:
    by_phase: dict[tuple[int, str], list[dict[str, Any]]] = defaultdict(list)
    for task in tasks:
        by_phase[(int(task["phase"]), str(task["phase_name"]))].append(task)

    lines = [
        f"# Implementation workplan — {generation}",
        "",
        f"This clean-slate generation starts from repository baseline `{baseline_sha}`.",
        "Only completed implementation history was carried forward. Unfinished tasks,",
        "briefs, IDs, titles, and dependency structure from the previous generation are",
        "not inputs to this derivation.",
        "",
        "Derive all remaining work from HEAD, the design baseline, completed history, and",
        "the current workplan contract.",
        "",
        "## Completed implementation history",
        "",
    ]

    if not tasks:
        lines.extend(["No completed tasks are carried into this generation.", ""])
        return "\n".join(lines)

    for (phase, phase_name), phase_tasks in sorted(by_phase.items()):
        lines.extend(
            [
                f"### Phase {phase} — {phase_name}",
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


def copy_v2_schema(repo_root: Path, target: Path) -> None:
    schema = repo_root / V2_SCHEMA
    if not schema.is_file():
        raise WorkplanError(f"no V2 tracker schema found at {schema}")
    shutil.copy2(schema, target / "tasks.schema.json")


def seed_next_generation(
    repo_root: Path, source: Path, target: Path, baseline_sha: str
) -> list[dict[str, Any]]:
    if target.exists():
        raise WorkplanError(f"target generation already exists: {target}")

    source_tracker = load_tracker(source / "tasks.json")
    completed = completed_history([dict(task) for task in source_tracker["tasks"]])

    target.mkdir(parents=True)
    copy_briefs(source, target, completed)
    copy_v2_schema(repo_root, target)
    tracker = {
        "schema_version": 2,
        "status_values": STATUS_VALUES,
        "tasks": completed,
    }
    save_tracker(target / "tasks.json", tracker)
    (target / "README.md").write_text(
        render_readme(target.name, completed, baseline_sha)
    )
    return completed


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
    return (
        subprocess.run(
            ("git", "show-ref", "--verify", "--quiet", f"refs/heads/{branch}"),
            cwd=repo_root,
            check=False,
        ).returncode
        == 0
    )


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
        raise WorkplanError(
            f"no re-derivation state found at {path}; run prepare first"
        )
    state = load_json(path)
    if not isinstance(state, dict):
        raise WorkplanError(f"invalid re-derivation state at {path}")
    return state


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


def _json_type_matches(value: Any, expected: str) -> bool:
    if expected == "object":
        return isinstance(value, dict)
    if expected == "array":
        return isinstance(value, list)
    if expected == "string":
        return isinstance(value, str)
    if expected == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if expected == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    if expected == "boolean":
        return isinstance(value, bool)
    if expected == "null":
        return value is None
    return True


def _resolve_ref(root_schema: dict[str, Any], ref: str) -> dict[str, Any]:
    if not ref.startswith("#/"):
        raise WorkplanError(f"unsupported JSON Schema reference {ref!r}")
    node: Any = root_schema
    for part in ref[2:].split("/"):
        node = node[part.replace("~1", "/").replace("~0", "~")]
    if not isinstance(node, dict):
        raise WorkplanError(f"JSON Schema reference {ref!r} is not an object")
    return node


def _schema_errors(
    value: Any,
    schema: dict[str, Any],
    root_schema: dict[str, Any],
    path: str,
) -> list[str]:
    if "$ref" in schema:
        return _schema_errors(value, _resolve_ref(root_schema, schema["$ref"]), root_schema, path)

    errors: list[str] = []
    expected = schema.get("type")
    if expected is not None:
        options = expected if isinstance(expected, list) else [expected]
        if not any(_json_type_matches(value, option) for option in options):
            return [f"{path}: expected type {expected!r}"]

    if "const" in schema and value != schema["const"]:
        errors.append(f"{path}: must equal {schema['const']!r}")
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{path}: {value!r} is not an allowed value")

    if isinstance(value, str):
        if "minLength" in schema and len(value) < schema["minLength"]:
            errors.append(f"{path}: string is too short")
        if "pattern" in schema and re.fullmatch(schema["pattern"], value) is None:
            errors.append(f"{path}: does not match {schema['pattern']!r}")

    if isinstance(value, int) and not isinstance(value, bool):
        if "minimum" in schema and value < schema["minimum"]:
            errors.append(f"{path}: below minimum {schema['minimum']}")
        if "maximum" in schema and value > schema["maximum"]:
            errors.append(f"{path}: above maximum {schema['maximum']}")

    if isinstance(value, list):
        if "minItems" in schema and len(value) < schema["minItems"]:
            errors.append(f"{path}: needs at least {schema['minItems']} item(s)")
        if "maxItems" in schema and len(value) > schema["maxItems"]:
            errors.append(f"{path}: has too many items")
        if schema.get("uniqueItems"):
            rendered = [json.dumps(item, sort_keys=True) for item in value]
            if len(set(rendered)) != len(rendered):
                errors.append(f"{path}: items must be unique")
        if isinstance(schema.get("items"), dict):
            for index, item in enumerate(value):
                errors.extend(
                    _schema_errors(
                        item, schema["items"], root_schema, f"{path}[{index}]"
                    )
                )

    if isinstance(value, dict):
        required = schema.get("required", [])
        for key in required:
            if key not in value:
                errors.append(f"{path}: missing required property {key!r}")
        properties = schema.get("properties", {})
        if schema.get("additionalProperties") is False:
            for key in value:
                if key not in properties:
                    errors.append(f"{path}: unexpected property {key!r}")
        for key, child_schema in properties.items():
            if key in value:
                errors.extend(
                    _schema_errors(
                        value[key], child_schema, root_schema, f"{path}.{key}"
                    )
                )

    for child in schema.get("allOf", []):
        errors.extend(_schema_errors(value, child, root_schema, path))

    if "if" in schema:
        condition_errors = _schema_errors(value, schema["if"], root_schema, path)
        branch = schema.get("then") if not condition_errors else schema.get("else")
        if isinstance(branch, dict):
            errors.extend(_schema_errors(value, branch, root_schema, path))

    return errors


def validate_against_generation_schema(generation: Path, tracker: dict[str, Any]) -> None:
    schema_path = generation / "tasks.schema.json"
    schema = load_json(schema_path)
    if not isinstance(schema, dict):
        raise WorkplanError(f"generation schema is not an object: {schema_path}")
    errors = _schema_errors(tracker, schema, schema, "$")
    if errors:
        raise WorkplanError(
            "tracker does not validate against generation JSON Schema:\n- "
            + "\n- ".join(errors)
        )


def validate_completed_history(
    tracker: dict[str, Any], expected_completed: list[dict[str, Any]]
) -> None:
    expected = {task["id"]: task for task in expected_completed}
    current = {task.get("id"): task for task in tracker["tasks"]}
    errors: list[str] = []
    for task_id, original in expected.items():
        if task_id not in current:
            errors.append(f"{task_id}: completed historical task was removed")
        elif current[task_id] != original:
            errors.append(f"{task_id}: completed historical task was mutated")
    if errors:
        raise WorkplanError(
            "completed implementation history is immutable:\n- "
            + "\n- ".join(errors)
        )


def validate_dependency_graph(tasks: list[dict[str, Any]]) -> None:
    by_id = {task["id"]: task for task in tasks}
    errors: list[str] = []
    for task in tasks:
        for dependency in task.get("depends_on", []):
            if dependency not in by_id:
                errors.append(f"{task['id']}: unknown dependency {dependency}")
    if errors:
        raise WorkplanError("invalid dependency graph:\n- " + "\n- ".join(errors))

    visiting: set[str] = set()
    visited: set[str] = set()

    def visit(task_id: str, stack: list[str]) -> None:
        if task_id in visited:
            return
        if task_id in visiting:
            start = stack.index(task_id)
            cycle = stack[start:] + [task_id]
            raise WorkplanError("dependency cycle: " + " -> ".join(cycle))
        visiting.add(task_id)
        stack.append(task_id)
        for dependency in by_id[task_id].get("depends_on", []):
            visit(dependency, stack)
        stack.pop()
        visiting.remove(task_id)
        visited.add(task_id)

    for task_id in by_id:
        visit(task_id, [])


def validate_replacement_relationships(tasks: list[dict[str, Any]]) -> None:
    by_id = {task["id"]: task for task in tasks}
    errors: list[str] = []
    for task in tasks:
        task_id = task["id"]
        status = task["status"]
        supersedes = task.get("supersedes", [])
        superseded_by = task.get("superseded_by", [])

        if status == "superseded":
            for replacement in superseded_by:
                target = by_id.get(replacement)
                if target is None:
                    errors.append(f"{task_id}: replacement {replacement} does not exist")
                elif task_id not in target.get("supersedes", []):
                    errors.append(
                        f"{task_id}: replacement {replacement} does not point back via supersedes"
                    )
        elif superseded_by:
            errors.append(f"{task_id}: only superseded tasks may set superseded_by")

        for old_id in supersedes:
            old = by_id.get(old_id)
            if old is None:
                errors.append(f"{task_id}: supersedes unknown task {old_id}")
            elif old.get("status") != "superseded":
                errors.append(f"{task_id}: supersedes {old_id}, which is not superseded")
            elif task_id not in old.get("superseded_by", []):
                errors.append(
                    f"{task_id}: supersedes {old_id}, but {old_id} does not point back"
                )

    if errors:
        raise WorkplanError(
            "invalid terminal/replacement relationships:\n- "
            + "\n- ".join(errors)
        )


def validate_briefs(generation: Path, tasks: list[dict[str, Any]]) -> None:
    errors: list[str] = []
    for task in tasks:
        brief = task.get("brief")
        if not isinstance(brief, str) or not brief:
            errors.append(f"{task.get('id')}: missing brief path")
            continue
        relative = safe_relative_brief(brief)
        if not (generation / relative).is_file():
            errors.append(f"{task.get('id')}: missing brief file {brief}")
    if errors:
        raise WorkplanError("invalid task briefs:\n- " + "\n- ".join(errors))


def validate_generation(
    generation: Path, expected_completed: list[dict[str, Any]]
) -> None:
    tracker = load_tracker(generation / "tasks.json")
    validate_against_generation_schema(generation, tracker)

    ids = [task.get("id") for task in tracker["tasks"]]
    if len(set(ids)) != len(ids):
        raise WorkplanError("V2 tracker contains duplicate task IDs")

    validate_briefs(generation, tracker["tasks"])
    validate_dependency_graph(tracker["tasks"])
    validate_replacement_relationships(tracker["tasks"])
    validate_completed_history(tracker, expected_completed)


def prepare(
    repo_root: Path,
    workplan_relative: Path,
    clean_branch_name: str | None,
) -> dict[str, Any]:
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
            f"generation changed during preparation: expected {generation_name}, "
            f"got {actual_generation_name}"
        )

    target = workplan_root / generation_name
    completed = seed_next_generation(repo_root, source, target, baseline_sha)
    (workplan_root / CURRENT).write_text(generation_name + "\n")
    archive_commit = commit_path(
        repo_root,
        workplan_root,
        f"chore: archive and seed workplan {generation_name}",
    )

    run_git(repo_root, "switch", "-c", clean_branch)
    sanitize_generations(workplan_root, target)
    clean_commit = commit_path(
        repo_root,
        workplan_root,
        f"chore: sanitize workplan for {generation_name} re-derivation",
    )

    state = {
        "archive_branch": archive_branch,
        "archive_commit": archive_commit,
        "baseline_sha": baseline_sha,
        "clean_branch": clean_branch,
        "clean_commit": clean_commit,
        "generation": generation_name,
        "workplan_root": workplan_relative.as_posix(),
        "completed_history": completed,
    }
    write_state(repo_root, state)

    print(
        f"prepared {generation_name}: {len(completed)} completed tasks carried forward\n"
        f"archive branch: {archive_branch} ({archive_commit})\n"
        f"clean branch:   {clean_branch} ({clean_commit})\n"
        f"active plan:    {workplan_relative / generation_name}"
    )
    return state


def restore_generation(
    repo_root: Path, source_ref: str, generation_path: Path
) -> None:
    relative = generation_path.relative_to(repo_root)
    run_git(
        repo_root,
        "restore",
        "--source",
        source_ref,
        "--staged",
        "--worktree",
        "--",
        str(relative),
    )


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

    clean_head = run_git(repo_root, "rev-parse", clean_branch)
    current = current_branch(repo_root)
    if current != clean_branch:
        run_git(repo_root, "switch", clean_branch)
    generation_path = repo_root / workplan_relative / generation
    validate_generation(
        generation_path, list(state.get("completed_history", []))
    )

    run_git(repo_root, "switch", archive_branch)
    generation_path = repo_root / workplan_relative / generation
    restore_generation(repo_root, clean_head, generation_path)

    relative = generation_path.relative_to(repo_root)
    changed = run_git(
        repo_root, "diff", "--cached", "--name-only", "--", str(relative)
    )
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
        "--workplan-root",
        type=Path,
        default=WORKPLAN,
        help=f"workplan directory relative to repository root (default: {WORKPLAN})",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)
    prepare_parser = subparsers.add_parser(
        "prepare", help="archive current generation and create the clean branch"
    )
    prepare_parser.add_argument(
        "--clean-branch",
        help="local branch for clean-slate re-derivation (default: workplan-rederive-vN)",
    )
    subparsers.add_parser(
        "harvest",
        help="validate and copy only the re-derived generation back to the archive branch",
    )
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
