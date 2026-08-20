#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work_dir=${WORK_DIR:-"$repo_root/.private"}
key_file=${KEY_FILE:-"$work_dir/private-model-key.b64"}
provider_source="$work_dir/provider-src"
provider_dir="$work_dir/vault"
provider_binary="$provider_dir/attested-secrets-server"
secrets_file="$provider_dir/secrets.json"

provider_url=https://github.com/tinfoilsh/example-secret-keys.git
provider_commit=517e90b4c12492a08d7af090821106673e672da1

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
mkdir -p "$provider_source" "$provider_dir"
if [[ ! -d "$provider_source/.git" ]]; then
  git -C "$provider_source" init -q
  git -C "$provider_source" remote add origin "$provider_url"
fi
if [[ -n $(git -C "$provider_source" status --short) ]]; then
  printf 'refusing to overwrite modified provider source: %s\n' "$provider_source" >&2
  exit 1
fi
git -C "$provider_source" fetch --quiet --depth=1 origin "$provider_commit"
git -C "$provider_source" checkout --quiet --detach "$provider_commit"
[[ $(git -C "$provider_source" rev-parse HEAD) == "$provider_commit" ]]

(cd "$provider_source" && go test -race ./...)
(cd "$provider_source" && go build -trimpath -o "$provider_binary" .)

key=$(tr -d '\n' < "$key_file")
[[ $(printf '%s' "$key" | base64 -d | wc -c) -eq 64 ]]
printf '{"PRIVATE_MODEL_KEY":"%s"}\n' "$key" > "$secrets_file"
chmod 0500 "$provider_binary"
chmod 0400 "$secrets_file"

printf 'provider_commit=%s\n' "$provider_commit"
printf 'provider_binary=%s\n' "$provider_binary"
printf 'secrets_file=%s\n' "$secrets_file"
