#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work_dir=${WORK_DIR:-"$repo_root/.private"}
model_dir="$work_dir/plain"
tools_dir="$work_dir/tools"
output_dir="$work_dir/artifacts"
cache_dir="$work_dir/cache"
key_file="$work_dir/private-model-key.b64"

model_revision=476854d00ede130660aba430d15f9347ad2e7d0e
model_file=SmolLM2-135M-Instruct.Q4_K_M.gguf
model_sha256=8030f04528538d47bda434f6f0bdf3952c40a58123e4d5e755332f23731a8684
model_identity=tinfoilsh/demo-private-smollm2
modelwrap_version=v0.2.1
modelwrap_sha256=06a5a42e2126a5ed612c4b0617d191ed9c65bb5b3a183863e92ec959cf097f57

umask 077
mkdir -p "$model_dir" "$tools_dir" "$output_dir" "$cache_dir"

model_path="$model_dir/$model_file"
if [[ ! -f "$model_path" ]]; then
  curl -fL --retry 5 --output "$model_path" \
    "https://huggingface.co/QuantFactory/SmolLM2-135M-Instruct-GGUF/resolve/$model_revision/$model_file?download=true"
fi
printf '%s  %s\n' "$model_sha256" "$model_path" | sha256sum -c -

modelwrap="$tools_dir/modelwrap"
if [[ ! -f "$modelwrap" ]]; then
  curl -fL --retry 5 --output "$modelwrap" \
    "https://github.com/tinfoilsh/modelwrap/releases/download/$modelwrap_version/modelwrap"
  chmod 0755 "$modelwrap"
fi
printf '%s  %s\n' "$modelwrap_sha256" "$modelwrap" | sha256sum -c -

if [[ ! -f "$key_file" ]]; then
  openssl rand 64 | base64 -w0 > "$key_file"
  printf '\n' >> "$key_file"
fi
[[ $(base64 -d "$key_file" | wc -c) -eq 64 ]]

"$modelwrap" \
  --model-dir "$model_dir" \
  --encrypt \
  --key-file "$key_file" \
  --verify \
  --output "$output_dir" \
  --cache "$cache_dir" \
  "$model_identity"

if find "$output_dir" -type f ! -user "$(id -u)" -print -quit | grep -q .; then
  sudo chown -R "$(id -u):$(id -g)" "$output_dir"
fi
find "$output_dir" -type d -exec chmod 0700 {} +
find "$output_dir" -type f -exec chmod 0600 {} +

info_file=$(find "$output_dir/$model_identity" -name '*.emwp.info' -type f -print -quit)
artifact=${info_file%.info}
revision=$(basename "$artifact" .emwp)
ref=$(cat "$info_file")

printf 'model=%s@%s\n' "$model_identity" "$revision"
printf 'emwp=%s\n' "$ref"
printf 'artifact=%s\n' "$artifact"
printf 'artifact_sha256=%s\n' "$(sha256sum "$artifact" | cut -d' ' -f1)"
printf 'key_file=%s\n' "$key_file"
printf 'host_path=/mnt/large/tinfoil/models/%s/%s.emwp\n' "$model_identity" "$revision"

