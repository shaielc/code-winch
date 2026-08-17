#!/usr/bin/env bash
# Send input to a running harness. Interim stand-in for `winch run input`;
# deleted by P1-051.
#
#   ./tmp/input.sh 'echo hello'            # text, to the last created run
#   ./tmp/input.sh 'echo hello' <run-id>
#   ./tmp/input.sh --resize 40 120 [run-id]
#
# The fake harness understands `help`, `echo <text>`, `fail`, and `exit`; any
# other line is echoed back.
. "$(dirname "$0")/common.sh"

if [ "${1:-}" = "--resize" ]; then
	rows=${2:?usage: input.sh --resize <rows> <columns> [run-id]}
	columns=${3:?usage: input.sh --resize <rows> <columns> [run-id]}
	id=${4:-$(cat "$last_run_file" 2>/dev/null || true)}
	body=$(jq -nc --argjson r "$rows" --argjson c "$columns" '{kind:"resize", rows:$r, columns:$c}')
else
	text=${1:-echo hello}
	id=${2:-$(cat "$last_run_file" 2>/dev/null || true)}
	body=$(jq -nc --arg t "$text" '{kind:"text", text:$t}')
fi
require_run_id "$id" "<text>"

curl -fsS -X POST "$ENDPOINT/api/v1/runs/$id/input" "${write_auth[@]}" \
	-H "Idempotency-Key: $(idempotency_key input)" \
	-H "If-Match: \"$(run_version "$id")\"" -d "$body" | jq -c .
