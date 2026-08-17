#!/usr/bin/env bash
# Shared configuration for the interim API scripts. Deleted by P1-051.
#
# Sourced, never run. Each capability has its own script beside this one, in the
# shape `winch run <command>` will take, so P1-051 replaces them one for one.
set -euo pipefail

ENDPOINT=${WINCH_ENDPOINT:-http://localhost:8080}
TOKEN=${WINCH_TOKEN:-local-development-session-token-0000000000}
CSRF=${WINCH_CSRF_TOKEN:-local-development-csrf-token-000000000000}
ORIGIN=${WINCH_ALLOWED_ORIGIN:-http://localhost:8080}

# Reads carry the session token; mutations additionally carry the CSRF token and
# the origin the daemon was configured to accept.
read_auth=(-H "Authorization: Bearer $TOKEN")
write_auth=("${read_auth[@]}" -H "X-CSRF-Token: $CSRF" -H "Origin: $ORIGIN" -H "Content-Type: application/json")

# Every mutation needs a client-generated idempotency key. Reusing one with a
# different request is an idempotency conflict, not a second operation.
idempotency_key() { printf '%s-%s-%s' "${1:-command}" "$$" "$RANDOM"; }

# The ETag moves whenever the run changes, and recording an event changes it, so
# a conditional write reads the current version immediately before sending.
run_version() {
	curl -fsS "$ENDPOINT/api/v1/runs/$1" "${read_auth[@]}" | jq -r .version
}

require_run_id() {
	if [ -z "${1:-}" ]; then
		echo "usage: $(basename "$0") <run-id> ${2:-}" >&2
		exit 2
	fi
}

# The last created run, so the scripts can be chained without copying IDs.
last_run_file="$(dirname "${BASH_SOURCE[0]}")/.last-run"
