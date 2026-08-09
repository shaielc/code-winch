#!/usr/bin/env bash

set -euo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
tracker="${1:-$repo_root/docs/workplan/tasks.json}"

if [[ $# -gt 1 ]]; then
  echo "usage: ${0##*/} [tracker.json]" >&2
  exit 2
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required to list available tasks" >&2
  exit 1
fi

if [[ ! -f "$tracker" ]]; then
  echo "error: task tracker not found: $tracker" >&2
  exit 1
fi

jq '. as $tracker | [
  $tracker.tasks[]
  | select(.status == "pending")
  | select(.depends_on | all(. as $dependency
      | any($tracker.tasks[]; .id == $dependency and .status == "completed")))
  | {id, title, phase, owner, brief}
]' "$tracker"
