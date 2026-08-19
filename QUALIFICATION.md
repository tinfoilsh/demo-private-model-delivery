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
| Local CVM boot with encrypted model | Pass | Blocked by box3 guest boot |
| In-CVM inference | Valid chat completion | Blocked by box3 guest boot |
| Model key visible to workload container | No | Static/tests pass; runtime blocked |
| Provider health over TLS 1.3 | Pass | Pass |
| Provider under TLS 1.2 | Reject | Pass |
| Provider without client certificate | Reject | Pass |
| Forged forwarded-certificate header | Reject | Pass |
| Challenge used by another client certificate | Reject and consume | Pass |
| Challenge replay | Reject | Pass |
| Unauthorized repository | Reject | Pass |
| Malformed v3 document | Reach verifier and reject | Pass |
| Provider response/log error paths | Never expose model key | Pass |
| Provider with unpublished local CVM | Reject at provenance | Blocked before provider by box3 guest boot |
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

## Box3 blocker

Box3 currently fails before the guest exposes networking or a console:

- full SEV-SNP `tinctl dev-launch` starts QEMU but never reaches the shim;
- explicit non-CC/dummy-attestation mode behaves the same;
- a manual SeaBIOS launch reaches a running Linux vCPU through QMP but still
  never initializes the serial console, guest network, or shim; and
- recent host logs show the same behavior with the released CVM `v0.11.0`, so
  this is not specific to the encrypted model or local CVM tip.

All test VMs and the manual QEMU process were removed. `tinctl ls` returned no
deployments afterward. Reboot or repair box3's current guest boot path before
rerunning the two runtime-blocked rows.

## Remaining release gates

1. Repair box3 guest boot and rerun encrypted in-CVM inference plus workload
   environment inspection.
2. Choose public workload provenance or implement authenticated private GitHub
   collateral for ATC and `freshness-witness`.
3. Cut the approved CVM release through the normal supply-chain workflow.
4. Configure the provider for `tinfoilsh/demo-private-model-delivery`, remove
   the local external-config key, and prove successful v3-gated key release.
