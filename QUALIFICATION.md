# Qualification

## Pins

- Model source revision: `476854d00ede130660aba430d15f9347ad2e7d0e`
- Model SHA-256: `8030f04528538d47bda434f6f0bdf3952c40a58123e4d5e755332f23731a8684`
- Modelwrap: `v0.2.1`
- Modelwrap SHA-256: `06a5a42e2126a5ed612c4b0617d191ed9c65bb5b3a183863e92ec959cf097f57`
- Inference image: `ghcr.io/ggml-org/llama.cpp@sha256:d740abe8d85d092a9bdffcc19c0f85a3b7e7e9a0b9588d12ae224c15c436e17e`
- Model identity: `tinfoilsh/demo-private-smollm2@2ca51f69b2fbf3f46eea72a3195388c5e93f84c0d36ccd439d91cf529ee8e7a3`
- Active EMWP SHA-256: `1f755efa099d64add1ea360bc83fd956e8b9b6a7a6523f9b7160dae9da10f38b`
- CVM source: `tinfoilsh/cvmimage@7687b9764e8ea4da905ff10f2150bc6b7876c995`
- CVM rootfs SHA-256: `1469ae7966561b0b2b81a4ca8e122aa148d8ff1bec0089fd0fc71b6c654d77d2`
- CVM kernel SHA-256: `b1e042ac790ffbd1f8031d13e870f7b907d2bf685e755d43b2975c78752b42eb`
- CVM initrd SHA-256: `fc1d0b7a6703e1b6c7dd34113aea3c3f5f5ef51decb3363736103c46efc9a24d`
- CVM roothash-file SHA-256: `8fea51650c7e9cbb416ae7f43701ed71e200c40f5f2fc35c71d2f1457e8b9723`
- Provider source: `tinfoilsh/example-secret-keys@517e90b`

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
| Provider with released CVM and v3 provenance | Pass | Blocked on approved CVM release |
| Private workload repository collateral | Pass | Blocked on private-collateral support |

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
- A temporary development provider on the Box3 CVM bridge released exactly the
  one missing model key over TLS 1.3. The enclave authenticated its measured
  private CA, completed the certificate-bound challenge protocol, mounted the
  model, and reached ready without the key in external config.
- The temporary provider, private CA keys, external configs, listener, and
  firewall rule were removed after the test. Box3 was left with no deployments.

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

## Remaining release gates

1. Review and merge the measured private-CA support if local or air-gapped
   customer vaults are a release requirement.
2. Choose public workload provenance or implement authenticated private GitHub
   collateral for ATC and `freshness-witness`.
3. Cut the approved CVM release through the normal supply-chain workflow.
4. Configure the provider for `tinfoilsh/demo-private-model-delivery`, remove
   the local external-config key, and prove successful v3-gated key release.
