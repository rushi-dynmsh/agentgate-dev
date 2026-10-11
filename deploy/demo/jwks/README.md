# Demo dev-fixture JWT signing key

Same non-secret dev fixture as `deploy/g6/jwks/` — copied here (not
referenced across directories) so `deploy/demo/` stays self-contained and
portable on its own, consistent with it being an additive topology. See
`deploy/g6/jwks/README.md` for the full explanation of what this key is,
where the matching private half lives (hardcoded Go constants in
`agentgate/qa/g6enforcement/jwt_fixture_test.go` and
`deploy/g6/probe-client/jwt.go` — both reused unchanged for this topology,
since it's the same signing key), and how to regenerate it.
