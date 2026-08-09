#!/usr/bin/env bash

set -euo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
tracker="$repo_root/docs/workplan/tasks.json"

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required to list available tasks" >&2
  exit 1
fi

jq '. as $tracker | [
  $tracker.tasks[]
  | select(.status == "pending")
  | select(.depends_on | all(. as $dependency
      | any($tracker.tasks[]; .id == $dependency and .status == "completed")))
  | {id, title, phase, owner, brief}
]' "$tracker"
