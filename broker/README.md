# Attested secret broker deployment

The shipping demo uses the canonical
[`tinfoilsh/keyserver`](https://github.com/tinfoilsh/keyserver) broker with its
file backend. The repository pins an audited broker commit rather than copying
its trust-critical verification code.

Run this after `scripts/prepare-model.sh`:

```bash
./scripts/prepare-broker.sh
```

The script checks out the pinned broker commit, runs its race-enabled tests,
builds it, and writes a mode-`0400` `secrets.json` plus an editable mode-`0600`
`policy.yaml`. Review the policy and replace its tag and domain before
deployment. Install it as mode `0400`. All generated output stays under ignored
`.private/`.

## Install on the customer broker host

Create an unprivileged `attested-secrets` account and install:

- `.private/broker/attested-secret-broker` at
  `/opt/tinfoil/private-model-demo/attested-secret-broker`;
- `.private/broker/secrets.json` at
  `/etc/tinfoil-private-model-demo/secrets.json`;
- `.private/broker/policy.yaml` at
  `/etc/tinfoil-private-model-demo/policy.yaml`;
- `broker/broker.env.example` as
  `/etc/tinfoil-private-model-demo/broker.env`, with certificate paths
  adjusted for the broker hostname; and
- `broker/attested-secret-broker.service` in `/etc/systemd/system/`.

Keep the secret file and TLS private key readable only by the service account.
The broker terminates TLS itself because forwarded client-certificate headers
are not an authentication boundary.

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now attested-secret-broker.service
curl --fail https://broker.example.com/health
```

Set `vault-url` in measured `tinfoil-config.yml` to that HTTPS origin. Despite
the legacy field name, it may point to any compatible attested secret broker.
The broker consumes one fresh, single-use challenge, verifies the v3 document
offline, binds the verified TLS key to the `/fetch` connection, and returns only
the secret names authorized by the pinned release policy.
