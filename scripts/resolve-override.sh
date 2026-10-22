#!/bin/bash
# Print the absolute path of a gitignored, site-specific file such as a Compose
# override, or nothing when no copy exists.
#
# Usage: resolve-override.sh <repo-relative path>
#
# The first existing copy wins, searched in this order:
#   1. this worktree
#   2. the main worktree, so linked worktrees share the operator's overrides
#   3. ${CODE_WINCH_CONFIG_DIR:-/var/code-winch}, mirroring the repo layout
set -euo pipefail

rel=${1:?usage: resolve-override.sh <repo-relative path>}
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

candidates=()
if worktree=$(git -C "$script_dir" rev-parse --show-toplevel 2>/dev/null); then
    candidates+=("$worktree")
    # The first entry of `git worktree list` is always the main worktree.
    main=$(git -C "$script_dir" worktree list --porcelain | sed -n '1s/^worktree //p')
    if [ -n "$main" ] && [ "$main" != "$worktree" ]; then
        candidates+=("$main")
    fi
else
    candidates+=("$(dirname -- "$script_dir")")
fi
candidates+=("${CODE_WINCH_CONFIG_DIR:-/var/code-winch}")

for dir in "${candidates[@]}"; do
    if [ -f "$dir/$rel" ]; then
        printf '%s\n' "$dir/$rel"
        exit 0
    fi
done