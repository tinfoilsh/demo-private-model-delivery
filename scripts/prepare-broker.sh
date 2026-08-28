#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work_dir=${WORK_DIR:-"$repo_root/.private"}
key_file=${KEY_FILE:-"$work_dir/private-model-key.b64"}
broker_source="$work_dir/broker-src"
broker_dir="$work_dir/broker"
broker_binary="$broker_dir/attested-secret-broker"
secrets_file="$broker_dir/secrets.json"
policy_file="$broker_dir/policy.yaml"

broker_url=https://github.com/tinfoilsh/keyserver.git
broker_commit=adedf59f6dbf2b8de6ac967cd8ab6158160a939a

for command in git go; do
  command -v "$command" >/dev/null || {
    printf 'missing required command: %s\n' "$command" >&2
    exit 1
  }
done
[[ -f "$key_file" ]] || {
  printf 'missing model key; run scripts/prepare-model.sh first\n' >&2
  exit 1
}

umask 077
mkdir -p "$broker_source" "$broker_dir"
if [[ ! -d "$broker_source/.git" ]]; then
  git -C "$broker_source" init -q
  git -C "$broker_source" remote add origin "$broker_url"
fi
if [[ -n $(git -C "$broker_source" status --short) ]]; then
  printf 'refusing to overwrite modified broker source: %s\n' "$broker_source" >&2
  exit 1
fi
git -C "$broker_source" fetch --quiet --depth=1 origin "$broker_commit"
git -C "$broker_source" checkout --quiet --detach "$broker_commit"
[[ $(git -C "$broker_source" rev-parse HEAD) == "$broker_commit" ]]

(cd "$broker_source" && go test -race ./...)
(cd "$broker_source" && go build -trimpath -o "$broker_binary" .)

key=$(tr -d '\n' < "$key_file")
[[ $(printf '%s' "$key" | base64 -d | wc -c) -eq 64 ]]
printf '{"models/private-smollm2":{"key":"%s"}}\n' "$key" > "$secrets_file"
cp "$repo_root/broker/policy.example.yaml" "$policy_file"
chmod 0500 "$broker_binary"
chmod 0400 "$secrets_file"
chmod 0600 "$policy_file"

printf 'broker_commit=%s\n' "$broker_commit"
printf 'broker_binary=%s\n' "$broker_binary"
printf 'secrets_file=%s\n' "$secrets_file"
printf 'policy_file=%s\n' "$policy_file"
