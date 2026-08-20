# Development-only vault

This server exercises the same HTTPS client-certificate, `/challenge`, `/fetch`,
single-use nonce, repository, and exact-secret-selection flow as the shipping
provider. It deliberately does **not** verify the submitted attestation
document, allowing a locally built CVM with unpublished provenance to test the
rest of private model delivery.

It refuses to start unless `-unsafe-skip-attestation` is present and binds only
to loopback unless a second `-unsafe-allow-non-loopback` acknowledgement is
provided. Non-loopback mode also requires `-allow-source-cidr`; protect the
port with a host firewall for that same CVM test network, use a disposable
model key, and remove the service after the test.

The transport remains HTTPS. Generate a development CA and server certificate,
then place that CA certificate in the measured CVM configuration via
`vault-ca`. Never replace HTTPS with plaintext HTTP: the host hypervisor and
network are outside the CVM trust boundary and could otherwise read or replace
the model key.

```bash
go test ./...
go build -trimpath -o dev-vault .

./dev-vault \
  -unsafe-skip-attestation \
  -repo tinfoilsh/confidential-debug \
  -secrets ../.private/vault/secrets.json \
  -tls-cert ./server.crt \
  -tls-key ./server.key
```

Use the strict provider in `vault/` for every release qualification.
