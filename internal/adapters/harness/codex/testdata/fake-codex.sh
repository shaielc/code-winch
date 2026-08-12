#!/bin/sh
set -eu
[ "$1" = exec ]
[ "$2" = --json ]
[ "${3:-}" = - ]
IFS= read -r prompt
[ "$prompt" = "fixture prompt" ]
cat "$(dirname "$0")/transcript.jsonl"
