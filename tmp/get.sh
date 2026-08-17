#!/usr/bin/env bash
# Read a run. Interim stand-in for `winch run get`; deleted by P1-051.
#
#   ./tmp/get.sh [run-id]
. "$(dirname "$0")/common.sh"

id=${1:-$(cat "$last_run_file" 2>/dev/null || true)}
require_run_id "$id"

curl -fsS "$ENDPOINT/api/v1/runs/$id" "${read_auth[@]}" | jq .
