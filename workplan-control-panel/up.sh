#!/bin/bash

runner_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

compose_files=(-f "$runner_dir/compose.yml")
env_file=$(bash "$runner_dir/../scripts/resolve-override.sh" workplan-control-panel/.env)
env_file=${env_file:-$runner_dir/.env}

override=$(bash "$runner_dir/../scripts/resolve-override.sh" workplan-control-panel/compose.override.yml)
if [ -n "$override" ]; then
    compose_files+=(-f "$override")
fi

docker compose \
    --env-file "$env_file" \
    "${compose_files[@]}" \
    up -d
