# Qualification

## Pins

- Model source revision: `476854d00ede130660aba430d15f9347ad2e7d0e`
- Model SHA-256: `8030f04528538d47bda434f6f0bdf3952c40a58123e4d5e755332f23731a8684`
- Modelwrap: `v0.2.1`
- Modelwrap SHA-256: `06a5a42e2126a5ed612c4b0617d191ed9c65bb5b3a183863e92ec959cf097f57`
- Inference image: `ghcr.io/ggml-org/llama.cpp@sha256:d740abe8d85d092a9bdffcc19c0f85a3b7e7e9a0b9588d12ae224c15c436e17e`
- Model identity: `tinfoilsh/demo-private-smollm2@2ca51f69b2fbf3f46eea72a3195388c5e93f84c0d36ccd439d91cf529ee8e7a3`
- EMWP SHA-256: `2e942987774d20df64b4ef0288da70efff8893608ed7b52d468e604c438599c9`

## Matrix

| Test | Expected | Result |
| --- | --- | --- |
| Pinned plaintext model inference | Valid chat completion | Pending |
| Modelwrap built-in EMWP verification | Pass | Pending |
| Wrong model key | Fail closed | Pending |
| Local CVM boot with encrypted model | Pass | Pending |
| In-CVM inference | Valid chat completion | Pending |
| Model key visible to workload container | No | Pending |
| Missing model key | Fail closed | Pending |
| Tampered ciphertext | Fail closed | Pending |
| Wrong attested model identity | Fail closed | Pending |
| Provider without client certificate | Reject | Pending |
| Provider with unpublished local CVM | Reject at provenance | Pending |
| Provider with released CVM and v3 provenance | Pass | Blocked on approved CVM release |

