#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work_dir=${WORK_DIR:-"$repo_root/.private"}
key_file=${KEY_FILE:-"$work_dir/private-model-key.b64"}
kbs_source="$work_dir/kbs-src"
kbs_dir="$work_dir/kbs"
kbs_binary="$kbs_dir/key-broker-service"
secrets_file="$kbs_dir/secrets.json"
policy_file="$kbs_dir/policy.yaml"

kbs_url=https://github.com/tinfoilsh/keyserver.git
kbs_commit=adedf59f6dbf2b8de6ac967cd8ab6158160a939a

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
mkdir -p "$kbs_source" "$kbs_dir"
if [[ ! -d "$kbs_source/.git" ]]; then
  git -C "$kbs_source" init -q
  git -C "$kbs_source" remote add origin "$kbs_url"
fi
if [[ -n $(git -C "$kbs_source" status --short) ]]; then
  printf 'refusing to overwrite modified KBS source: %s\n' "$kbs_source" >&2
  exit 1
fi
git -C "$kbs_source" fetch --quiet --depth=1 origin "$kbs_commit"
git -C "$kbs_source" checkout --quiet --detach "$kbs_commit"
[[ $(git -C "$kbs_source" rev-parse HEAD) == "$kbs_commit" ]]

(cd "$kbs_source" && go test -race ./...)
(cd "$kbs_source" && go build -trimpath -o "$kbs_binary" .)

key=$(tr -d '\n' < "$key_file")
[[ $(printf '%s' "$key" | base64 -d | wc -c) -eq 64 ]]
printf '{"models/private-smollm2":{"key":"%s"}}\n' "$key" > "$secrets_file"
cp "$repo_root/kbs/policy.example.yaml" "$policy_file"
chmod 0500 "$kbs_binary"
chmod 0400 "$secrets_file"
chmod 0600 "$policy_file"

printf 'kbs_commit=%s\n' "$kbs_commit"
printf 'kbs_binary=%s\n' "$kbs_binary"
printf 'secrets_file=%s\n' "$secrets_file"
printf 'policy_file=%s\n' "$policy_file"
