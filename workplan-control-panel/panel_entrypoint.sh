#!/usr/bin/env bash
set -euo pipefail
: "${GITHUB_URL:?set GITHUB_URL to the repository URL}"
: "${PANEL_TOKEN:?set PANEL_TOKEN for API authentication}"
cd /opt/workplan-control-panel
exec python3 -m control_panel \
  --state-file=/var/lib/code-winch/state/task-state.json \
  --clone=/var/lib/code-winch/checkout --host=0.0.0.0 --port=8765
