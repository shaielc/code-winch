#!/usr/bin/env bash
# Refresh the Codex CLI login held in the runner's codex-auth volume.
#
# The stored credential expires, so rerun this whenever the scheduler workflow
# starts failing with a Codex authentication error. The runner service does not
# need to be stopped: this starts a throwaway container that shares the same
# codex-auth volume, so the refreshed credential is visible to the running
# runner immediately.
#
# Extra arguments are forwarded to `codex login`.

set -euo pipefail

runner_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ ! -f "$runner_dir/.env" ]]; then
  echo "runner/.env is missing; copy runner/.env.example and set GITHUB_URL" >&2
  exit 1
fi

NETWORK_MODE=host docker compose \
  --env-file "$runner_dir/.env" \
  -f "$runner_dir/compose.yml" \
  run  --rm --entrypoint codex runner "$@"
