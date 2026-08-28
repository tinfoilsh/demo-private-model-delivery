# Qualification

This file records the original `v0.0.2` qualification. The current blueprint
uses `tinfoilsh/keyserver` as its canonical Key Broker Service (KBS); historical
provider names and commit pins below are retained as test evidence.

## Pins

- Model source revision: `476854d00ede130660aba430d15f9347ad2e7d0e`
- Model SHA-256: `8030f04528538d47bda434f6f0bdf3952c40a58123e4d5e755332f23731a8684`
- Modelwrap: `v0.2.1`
- Modelwrap SHA-256: `06a5a42e2126a5ed612c4b0617d191ed9c65bb5b3a183863e92ec959cf097f57`
- Inference image: `ghcr.io/ggml-org/llama.cpp@sha256:d740abe8d85d092a9bdffcc19c0f85a3b7e7e9a0b9588d12ae224c15c436e17e`
- Model identity: `tinfoilsh/demo-private-smollm2@2ca51f69b2fbf3f46eea72a3195388c5e93f84c0d36ccd439d91cf529ee8e7a3`
- Active EMWP SHA-256: `4bb94a1f7a4b34c43d6a29bd9361787477e081224e4604d1800d97e1393fd9cb`
- CVM source: `tinfoilsh/cvmimage@5c4ae156f3a4945dd1e0231a90416d3d8c14a464`
- CVM rootfs SHA-256: `4fa1860446c551eacc84da6bc62f182c0e066e4a7da0ed7b821d7b2675ec4644`
- CVM kernel SHA-256: `b1e042ac790ffbd1f8031d13e870f7b907d2bf685e755d43b2975c78752b42eb`
- CVM initrd SHA-256: `fc1d0b7a6703e1b6c7dd34113aea3c3f5f5ef51decb3363736103c46efc9a24d`
- CVM roothash-file SHA-256: `1d6af1ba93af0c4b01853a4384e388cf66941b4bdd43d4af16b4767884455cab`
- Provider source: `tinfoilsh/example-secret-keys@a92de50`

## Matrix

| Test | Expected | Result |
| --- | --- | --- |
| Pinned plaintext model inference | Valid chat completion | Pass |
| Modelwrap built-in EMWP verification | Pass | Pass |
| Kernel dm-crypt + dm-verity + EROFS round trip | Pass | Pass |
| Wrong model key | Fail closed | Pass |
| Missing model key | Fail closed | Pass |
| Tampered ciphertext | Fail closed | Pass |
| Wrong attested model identity | Fail closed | Pass |
| Local CVM boot with encrypted model | Pass | Pass |
| In-CVM inference | Valid chat completion | Pass |
| Inference container restart | Model remains available | Pass |
| Per-container model grant | Model file visible only to inference | Pass |
| Model key exposure to workload | Model key absent from environment | Pass |
| Model key declared for workload container | No | Pass; runtime validation rejects this configuration |
| Provider health over TLS 1.3 | Pass | Pass |
| Provider under TLS 1.2 | Reject | Pass |
| Provider without client certificate | Reject | Pass |
| Forged forwarded-certificate header | Reject | Pass |
| Challenge used by another client certificate | Reject and consume | Pass |
| Challenge replay | Reject | Pass |
| Unauthorized repository | Reject | Pass |
| Malformed v3 document | Reach verifier and reject | Pass |
| Provider response/log error paths | Never expose model key | Pass |
| Strict provider with unpublished local CVM | Reject at provenance | Pass |
| Development provider with unpublished local CVM | Complete protocol except verification | Pass |
| Measured private CA for vault HTTPS | Connect and fetch | Pass |
| Plaintext HTTP vault URL | Reject | Pass |
| Provider with released CVM and v3 provenance | Pass | Pass |
| Fresh `tinfoil-go feat/v3` verification | Pass | Pass at `c1fac8b` |
| Full enclave redeploy and inference | Pass | Pass |
| Private workload repository collateral | Not required | Public config repo; private model remains encrypted |

## Executed tests

- `go test ./...` passed for the exact `cvmimage` tip.
- `go test -race ./...` passed for the attested provider.
- Modelwrap's privileged end-to-end suite passed pack, userspace encryption,
  kernel dm-crypt decryption, dm-verity verification, EROFS mounting, and
  superblock-tamper coverage.
- The active EMWP passed correct-key verification and independently rejected a
  wrong key, a missing key, a flipped ciphertext byte, and a changed model
  identity/salt.
- The digest-pinned `llama.cpp` image loaded the pinned GGUF and returned a
  valid OpenAI-compatible chat completion on box3.
- `dev-vault.tinfoil.sh` runs the provider directly on port 443 with
  `CAP_NET_BIND_SERVICE`; the obsolete Caddy and legacy vault services are
  disabled but preserved for rollback.
- The local shipping image booted the encrypted EMWP, served a valid chat
  completion, and served another completion after the inference container was
  restarted.
- The named model path was present in the inference container and the model
  file was absent from the ungranted debug-toolbox container. The model key was
  also absent from the inference container environment before restart.
- A source-restricted temporary development provider on
  `dev-vault.tinfoil.sh` released exactly the one missing model key over TLS
  1.3. The enclave completed the certificate-bound challenge protocol, mounted
  the model, and reached ready without the key in external config.
- The upstream `cvmimage` build for the exact source commit completed
  successfully after the local Nix shipping-image build passed.
- The public `v0.0.2` demo release authenticated `cvmimage v0.11.1`, and its
  freshness witness was issued from `freshness-witness` main.
- The strict provider at `dev-vault.tinfoil.sh` used
  `example-secret-keys@a92de50`, verified the fresh v3 envelope offline, and
  released exactly one secret. Boot reported `handed off 0 workload secret(s);
  fetched 1 from vault`.
- The released workload returned valid OpenAI-compatible inference with model
  path `/tinfoil/models/private-smollm2/SmolLM2-135M-Instruct.Q4_K_M.gguf`.
- `tinfoil-go feat/v3@c1fac8b` independently verified the promoted enclave as
  `tinfoilsh/demo-private-model-delivery@v0.0.2` with digest
  `bf2082e351b021e7a460c5260720a40dc4feb79165426f42e300856cbc036c16`.
- A full redeploy produced an independently attested update candidate, fetched
  the key again, completed inference, and promoted successfully.
- Released `tinfoil-go v0.15.4` reached its legacy online VCEK lookup but could
  not complete because both `kds-proxy.tinfoil.sh` and AMD KDS timed out from
  Box3. The v3 verifier and provider were unaffected because all collateral was
  carried in the attestation envelope.
- Box3 was left with no demo deployments after qualification. The dedicated
  development Vault remains configured for this demo repository.

## Box3 resolution

The earlier conclusion that Box3 could not boot guests was incorrect. The
actual sequence was:

- Box3 ran `tinfoild 0.6.13`; canonical `infra` upgraded it to `0.6.15`.
- The current debug image then showed Linux and PID 1 booting normally in about
  two seconds.
- The current `infra-harness` CPU config used mutable `busybox:latest`; hardened
  CVM validation correctly rejected it before shim readiness.
- Pinning the existing BusyBox digest made the smoke workload reach ready and
  serve traffic.
- The local shipping image then booted this private-model configuration and
  completed inference both with direct control injection and with vault release.

`tinfoild#159` remains the narrow fix for unnecessary certificate-token fetches
in self-signed dev launches. Qualification used a harmless placeholder token;
shipping control-plane deployments already carry their real token.

## Follow-ups

1. Review and deploy `controlplane#582` so the production heartbeat keeps this
   demo's freshness witness renewed automatically.
2. Merge measured private-CA support only if local or air-gapped customer
   Vaults are a release requirement.
3. Restore legacy AMD KDS reachability before using released v2 verification
   as an independent compatibility check from Box3.
