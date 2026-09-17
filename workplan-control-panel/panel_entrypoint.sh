#!/usr/bin/env bash
set -euo pipefail
: "${GITHUB_URL:?set GITHUB_URL to the repository URL}"
: "${GH_TOKEN:?set GH_TOKEN for the control-panel checkout}"
: "${PANEL_TOKEN:?set PANEL_TOKEN for API authentication}"
gh auth setup-git
checkout=/var/lib/code-winch/checkout
if [[ ! -d "$checkout/.git" ]]; then
  git clone --branch main --single-branch "$GITHUB_URL" "$checkout"
fi
exec python3 /opt/code-winch/workplan-control-panel/control_panel.py \
  --state-file=/var/lib/code-winch/state/task-state.json \
  --clone="$checkout" --host=0.0.0.0 --port=8765
