# Demo Backend Credential Requirements — for Backend's G7 Task B

**Backend, this is what you're waiting on** (`docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md` §3,
§4). AI/Gateway's Task B chose `support-desk-mcp`
(`deploy/demo/support-desk-mcp/`) — a realistic support-ticketing backend with three tools of
genuine, mixed risk: `list_tickets` (read), `update_ticket_status` (write), `delete_ticket`
(destructive). This describes what credential model a **real** version of this kind of backend
would need `internal/credential.Issuer` (already designed —
`agentgate/internal/credential/credential.go`) to produce, so you can build the concrete
implementation against a real, specific target rather than a guess.

**In the demo itself, right now:** a stub shared-secret header
(`X-Demo-Backend-Key: demo-backend-shared-secret-dev`, checked nowhere yet — `support-desk-mcp`
doesn't even validate it today, since this demo topology proves the *authorization* path, not a
credential-exchange implementation that doesn't exist). Do not wait for the real mechanism to use
the demo — see DoD item 5 in `docs/PHASES/G7_WORKSTREAMS/02_AI_GATEWAY_G7.md` §3.

## What a real support-ticketing backend's credential model looks like

Real systems in this category (Zendesk, Freshdesk, Jira Service Desk, and similar) are HTTP APIs
authenticated with **OAuth 2.0**, not a shared static key. Realistically:

- **Token type:** OAuth 2.0 access token (bearer), issued via client-credentials or
  on-behalf-of/token-exchange grant — never an API key shared across all callers, and never the
  caller's own inbound token forwarded unchanged (the non-negotiable this whole mechanism exists
  to satisfy).
- **Audience:** scoped to the specific backend instance/tenant (e.g. `https://api.example-support-desk.com`),
  not a generic "any backend" token.
- **Lifetime:** short-lived (minutes, not hours) — the credential should not outlive the single
  MCP call it authorizes. `internal/credential.Credential` doesn't carry an explicit expiry field
  today; if the real mechanism needs one, that's a small, additive extension, not a redesign.
- **Scope, mapped to this demo's actual risk split:**
  - `tickets:read` — sufficient for `list_tickets`.
  - `tickets:write` — required for `update_ticket_status`, in addition to `tickets:read`.
  - `tickets:delete` — required for `delete_ticket`, the narrowest, most sensitive scope; a real
    deployment should treat granting this scope to a service identity as itself an
    auditable/governed action, not a one-time setup step.
- **Caller/on-behalf-of identity:** the token should encode (or be traceable back to, via
  `internal/credential.Credential.Ref`) which AgentGate-authenticated principal it was minted
  for — most real OAuth providers support this via a `sub`/`act` (actor) claim in a JWT access
  token, or via an opaque token plus a server-side mapping. Either way, `Credential.Ref` should be
  enough to answer "which caller did this backend call happen on behalf of" from the audit trail
  alone, without needing the token itself.
- **Replay/confusion scoping:** ideally the token (or its server-side mapping) is scoped to the
  specific `decision.Request.ExecutionID` that authorized it, so a captured/replayed credential
  can't be reused for a different call. Not every OAuth provider supports per-call token scoping;
  where it isn't available, short lifetime is the fallback mitigation — document which one a real
  implementation actually gets.

## What this demo does NOT need from Backend

- No real OAuth client registration, no real IdP — this is a fixture backend, not a real SaaS
  tenant.
- No per-scope enforcement in `support-desk-mcp` itself — it's a demo backend proving AgentGate's
  authorization path, not a hardened API with its own defense-in-depth.

## Handoff

This document is the deliverable; `docs/PHASES/PROGRESS_AND_ROADMAP.md` §2 tracks it as
unblocking Backend's Task B implementation. Once Backend's real `credential.Issuer` exists, swap
it in here (see README.md's "Follow-up: real credentials" section) — not a new ticket, per
`02_AI_GATEWAY_G7.md` §3 DoD item 5.
