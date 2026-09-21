# G6 dev-fixture JWT signing key

**Not a secret.** This is a throwaway RSA key pair generated solely to let
`deploy/g6`'s live topology demonstrate real JWT-verified identity end to
end, the same way `deploy/g4`/other checkpoints use clearly-fake dev
credentials (see the security note at the top of
`deploy/g6/docker-compose.yml`). Never use it, or the pattern it
represents, for anything beyond this local/CI integration topology.

## What's here

- `dev-fixture-jwks.json` — the **public** half, in standard JWKS format.
  `deploy/g6/agentgateway.yaml`'s `jwtAuth.jwks.file` points at this file;
  agentgateway uses it to verify token signatures.

## Where the matching private half lives

The private key that signs test tokens is **not** a separate file here —
it's hardcoded as a Go string constant, duplicated in the two places that
need to mint tokens for this topology:

- `agentgate/qa/g6enforcement/jwt_fixture_test.go` (the Go live-E2E test suite)
- `deploy/g6/probe-client/jwt.go` (the manual/script-driven probe client)

Duplicated deliberately, not by oversight: the two callers are in separate
Go modules (`agentgate/go.mod` vs `deploy/g6/probe-client/go.mod`) with no
existing shared-code path between them, and this is a small (~30-line),
test-only fixture — not worth introducing a third shared module for. If a
third caller ever needs to mint tokens for this topology, that's the
trigger to actually factor it out.

## Regenerating this key

If it's ever compromised (e.g. accidentally reused somewhere it shouldn't
be) or just needs rotating: generate a new RSA-2048 key pair, update the
`n`/`e` values here from the new public key, and update the hardcoded PEM
constant in both Go files above with the new private key. There's no
migration concern — nothing this key protects is durable data (audit
records don't depend on JWT signatures remaining verifiable later).
