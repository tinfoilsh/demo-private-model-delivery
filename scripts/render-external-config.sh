#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work_dir=${WORK_DIR:-"$repo_root/.private"}
key_file=${KEY_FILE:-"$work_dir/private-model-key.b64"}
output=${OUTPUT:-"$work_dir/external-config.yml"}

[[ -f "$key_file" ]]
key=$(tr -d '\n' < "$key_file")
[[ $(printf '%s' "$key" | base64 -d | wc -c) -eq 64 ]]

umask 077
mkdir -p "$(dirname "$output")"
printf 'secrets:\n  PRIVATE_MODEL_KEY: %s\n' "$key" > "$output"
chmod 0600 "$output"
printf 'wrote %s\n' "$output"

