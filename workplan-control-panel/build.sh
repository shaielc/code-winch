#!/bin/bash
# Build the control-panel image the same way up.sh starts it.
#
# Rerun this after changing panel code or prompts; the image copies them in.
# Extra arguments are forwarded to `docker compose build`.

runner_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

compose_files=(-f "$runner_dir/compose.yml")
if [ -f "$runner_dir/compose.override.yml" ]; then
    compose_files+=(-f "$runner_dir/compose.override.yml")
fi

docker compose \
    --env-file "$runner_dir/.env" \
    "${compose_files[@]}" \
    build "$@"
