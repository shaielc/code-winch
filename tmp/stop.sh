#!/usr/bin/env bash
# Stop a run. Interim stand-in for `winch run stop`; deleted by P1-051.
#
#   ./tmp/stop.sh [run-id]
#
# Stopping records the intent and signals the harness; the run reaches its
# terminal state from the harness exit that follows, not from this call.
. "$(dirname "$0")/common.sh"

id=${1:-$(cat "$last_run_file" 2>/dev/null || true)}
require_run_id "$id"

curl -fsS -X POST "$ENDPOINT/api/v1/runs/$id/stop" "${write_auth[@]}" \
	-H "Idempotency-Key: $(idempotency_key stop)" \
	-H "If-Match: \"$(run_version "$id")\"" | jq -c '{id, state, version}'
