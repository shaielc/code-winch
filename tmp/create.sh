#!/usr/bin/env bash
# Create a run. Interim stand-in for `winch run create`; deleted by P1-051.
#
#   ./tmp/create.sh [workspace-path] [harness-profile] [sandbox-profile]
#
# Prints the run ID and records it in tmp/.last-run for the other scripts.
. "$(dirname "$0")/common.sh"

workspace=${1:-/workspace}
harness=${2:-fake}
sandbox=${3:-local}

body=$(jq -nc --arg w "$workspace" --arg h "$harness" --arg s "$sandbox" \
	'{workspacePath:$w, harnessProfile:$h, sandboxProfile:$s}')

run=$(curl -fsS -X POST "$ENDPOINT/api/v1/runs" "${write_auth[@]}" \
	-H "Idempotency-Key: $(idempotency_key create)" -d "$body")

printf '%s' "$(jq -r .id <<<"$run")" > "$last_run_file"
jq -c '{id, state, version, harnessProfile, sandboxProfile}' <<<"$run"
