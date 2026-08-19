# Private model delivery demo

This repository is a minimal end-to-end blueprint for delivering an encrypted
model to a Tinfoil CVM. The repository contains no model weights, encryption
keys, vault secrets, or CVM artifacts.

The demo uses:

- `modelwrap v0.2.1` to produce an encrypted, dm-verity-protected EMWP;
- a pinned SmolLM2 135M GGUF as a fast CPU inference workload;
- a digest-pinned `llama.cpp` server image;
- `PRIVATE_MODEL_KEY` as the boot-only model key secret; and
- `dev-vault.tinfoil.sh` as the attested secret provider.

## Prepare the encrypted model

```bash
./scripts/prepare-model.sh
```

The script downloads and verifies the pinned public source model, generates a
local 64-byte model key, and creates a verified EMWP. Its `.private/` output is
ignored by Git. For a real private model, replace the download with a local
model directory while keeping the same Modelwrap flow.

To create a local external config for pre-release testing:

```bash
./scripts/render-external-config.sh
```

The generated file is mode `0600`, ignored by Git, and must never be committed.
It bypasses the provider only for local qualification before the CVM image has
publishable provenance.

## Launch with a local CVM build

Build `shipping-image` from the intended `cvmimage` commit in an isolated
worktree, install the EMWP at the host model path shown by
`scripts/prepare-model.sh`, and launch without publishing any CVM artifacts:

```bash
tinctl dev-launch /path/to/cvmimage/result \
  --config ./tinfoil-config.yml \
  --external-config ./.private/external-config.yml \
  --skip-manifest \
  --watch
```

The inference API is OpenAI-compatible at `/v1/chat/completions`. Remove
`--external-config` after the CVM release exists and the attested provider is
configured with the same `PRIVATE_MODEL_KEY`.

See `QUALIFICATION.md` for the tested matrix and the final release boundary.

