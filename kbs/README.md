# Key Broker Service deployment

The shipping demo uses the canonical
[`tinfoilsh/keyserver`](https://github.com/tinfoilsh/keyserver) KBS with its
file backend. The repository pins an audited KBS commit rather than copying
its trust-critical verification code.

Run this after `scripts/prepare-model.sh`:

```bash
./scripts/prepare-kbs.sh
```

The script checks out the pinned KBS commit, runs its race-enabled tests,
builds it, and writes a mode-`0400` `secrets.json` plus an editable mode-`0600`
`policy.yaml`. Review the policy and replace its tag and domain before
deployment. Install it as mode `0400`. All generated output stays under ignored
`.private/`.

## Install on the customer KBS host

Create an unprivileged `attested-secrets` account and install:

- `.private/kbs/key-broker-service` at
  `/opt/tinfoil/private-model-demo/key-broker-service`;
- `.private/kbs/secrets.json` at
  `/etc/tinfoil-private-model-demo/secrets.json`;
- `.private/kbs/policy.yaml` at
  `/etc/tinfoil-private-model-demo/policy.yaml`;
- `kbs/kbs.env.example` as
  `/etc/tinfoil-private-model-demo/kbs.env`, with certificate paths adjusted
  for the KBS hostname; and
- `kbs/key-broker-service.service` in `/etc/systemd/system/`.

Keep the secret file and TLS private key readable only by the service account.
The KBS terminates TLS itself because forwarded client-certificate headers
are not an authentication boundary.

```bash
sudo install -d -o root -g root -m 0755 /opt/tinfoil/private-model-demo
sudo install -d -o attested-secrets -g attested-secrets -m 0700 \
  /etc/tinfoil-private-model-demo
sudo install -o root -g root -m 0555 \
  .private/kbs/key-broker-service \
  /opt/tinfoil/private-model-demo/key-broker-service
sudo install -o attested-secrets -g attested-secrets -m 0400 \
  .private/kbs/{secrets.json,policy.yaml} \
  /etc/tinfoil-private-model-demo/
sudo install -o root -g root -m 0644 \
  kbs/kbs.env.example \
  /etc/tinfoil-private-model-demo/kbs.env
sudo install -o root -g root -m 0644 \
  kbs/key-broker-service.service \
  /etc/systemd/system/key-broker-service.service
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now key-broker-service.service
curl --fail https://kbs.example.com/health
```

Set `vault-url` in measured `tinfoil-config.yml` to that HTTPS origin. Despite
the legacy field name, it may point to any compatible KBS. The KBS consumes one
fresh, single-use challenge, verifies the v3 document offline, binds the
verified TLS key to the `/fetch` connection, and returns only the secret names
authorized by the pinned release policy.
