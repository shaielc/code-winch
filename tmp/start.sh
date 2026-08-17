#!/usr/bin/env bash
# Start a run. Interim stand-in for `winch run start`; deleted by P1-051.
#
#   ./tmp/start.sh [run-id]
#
# The response reports the run after its launch, so a `running` state here means
# a harness process exists and its first events are already recorded.
. "$(dirname "$0")/common.sh"

id=${1:-$(cat "$last_run_file" 2>/dev/null || true)}
require_run_id "$id"

curl -fsS -X POST "$ENDPOINT/api/v1/runs/$id/start" "${write_auth[@]}" \
	-H "Idempotency-Key: $(idempotency_key start)" \
	-H "If-Match: \"$(run_version "$id")\"" |
	jq -c '{id, state, version, lastSequence}'
