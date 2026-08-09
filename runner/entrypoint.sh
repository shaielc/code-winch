#!/usr/bin/env bash

set -euo pipefail

cd /home/runner

config_dir=/var/lib/code-winch/runner-config
mkdir -p "$config_dir"

for file in .runner .credentials .credentials_rsaparams; do
  if [[ ! -f "$file" && -f "$config_dir/$file" ]]; then
    cp "$config_dir/$file" "$file"
  fi
done

if [[ ! -f .runner ]]; then
  : "${GITHUB_URL:?set GITHUB_URL to the repository URL}"
  : "${RUNNER_TOKEN:?set RUNNER_TOKEN to a current registration token}"

  ./config.sh \
    --unattended \
    --url "$GITHUB_URL" \
    --token "$RUNNER_TOKEN" \
    --name "${RUNNER_NAME:-code-winch-local}" \
    --labels "${RUNNER_LABELS:-code-winch,task-scheduler}" \
    --work /home/runner/_work \
    --replace

  for file in .runner .credentials .credentials_rsaparams; do
    if [[ -f "$file" ]]; then
      cp "$file" "$config_dir/$file"
    fi
  done
fi

unset RUNNER_TOKEN
exec ./run.sh
