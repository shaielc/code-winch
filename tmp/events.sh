#!/usr/bin/env bash
# Read a run's persisted event history. Interim stand-in for
# `winch run watch --after-sequence` over storage; deleted by P1-051.
#
#   ./tmp/events.sh [run-id] [after-sequence]
#
# This is the durable history, which stays authoritative: the live stream in
# stream.py is a notification of the same records, not a separate source.
. "$(dirname "$0")/common.sh"

id=${1:-$(cat "$last_run_file" 2>/dev/null || true)}
after=${2:-0}
require_run_id "$id" "[after-sequence]"

curl -fsS "$ENDPOINT/api/v1/runs/$id/events?after_sequence=$after" "${read_auth[@]}" |
	jq -r '.events[] | "\(.sequence)  \(.kind)  \(.payload | tostring | .[0:96])"'
