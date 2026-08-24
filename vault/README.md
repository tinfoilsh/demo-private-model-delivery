# Attested vault deployment

The shipping demo uses the small, standalone provider in
[`tinfoilsh/example-secret-keys`](https://github.com/tinfoilsh/example-secret-keys).
This repository pins the audited provider commit instead of copying its
trust-critical verification code.

Run this after `scripts/prepare-model.sh`:

```bash
./scripts/prepare-vault.sh
```

The script checks out provider commit
`a92de50148d22e778be96a3f413eaabbefe745a5`, runs its race-enabled tests,
builds it, and writes a mode-`0400` `secrets.json` containing only
`PRIVATE_MODEL_KEY`. All output stays under ignored `.private/`.

## Install on the customer vault host

Create an unprivileged `attested-secrets` account and install:

- `.private/vault/attested-secrets-server` at
  `/opt/tinfoil/private-model-demo/attested-secrets-server`;
- `.private/vault/secrets.json` at
  `/etc/tinfoil-private-model-demo/secrets.json`;
- `vault/provider.env.example` as
  `/etc/tinfoil-private-model-demo/provider.env`, with certificate paths
  adjusted for the vault hostname; and
- `vault/attested-secret-provider.service` in `/etc/systemd/system/`.

Keep the secret file and TLS private key readable only by the service account.
The service terminates TLS itself because forwarded client-certificate headers
are not an authentication boundary.

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now attested-secret-provider.service
curl --fail https://vault.example.com/health
```

Set `vault-url` in the measured `tinfoil-config.yml` to that HTTPS origin. The
provider authorizes the workload repository, consumes a certificate-bound
challenge, verifies the fresh v3 document offline, and returns only the secret
names declared by that workload.

For an isolated protocol test before publishable enclave provenance exists,
see `dev-vault/README.md`. Never deploy the development server as a real vault.
