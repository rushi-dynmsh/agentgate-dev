# Go Backend — G7 Governance Read-Surface Contract (Handoff)

**Ticket:** `docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md` §5 (Task C)
**Status:** Implemented, matches the reviewed proposal — see "Provenance" below before treating
this as frozen.
**Reviewed source:** `frontend/app/PROPOSED_G8_API_CONTRACTS.md` — Frontend drafted both
endpoints' shapes before Backend implemented them (an inverted, and better, review order than the
default: no risk of "built ahead of an agreed contract," since the agreement came first).

This is the canonical reference for Frontend (and anyone else) to integrate against the two new
governance read-only endpoints without reading Go source.

---

## Provenance — what "reviewed" means here

Frontend wrote `frontend/app/PROPOSED_G8_API_CONTRACTS.md` first. Backend implemented to that
spec exactly — same field names, same JSON shapes, same pagination model, same exclusions. There
is no divergence to reconcile. **What has not yet happened:** Frontend (the actual author of
`frontend/app`, not this agent) confirming they reviewed *this implementation* against their own
proposal and are satisfied it matches before either team calls the contract "frozen." Until that
explicit confirmation happens, treat this as implemented-to-spec, not yet formally frozen.

---

## Endpoint 1 — Audit query

`GET /api/v1/workspaces/{workspace_id}/audit-events`

Auth: same `AdminAuthMiddleware` as every existing `govapi` route (`X-AgentGate-Admin-Key` header
or `Authorization: Bearer <admin-token>`). No new auth mechanism.

| Query param | Required | Behavior |
|---|---|---|
| `limit` | No, default 50 | Capped server-side at 100 regardless of what's requested. A non-numeric value is rejected with 400, not silently defaulted. |
| `before_sequence` | No | Returns records with `sequence_number` strictly less than this value. Omit for the first/newest page. |

Response: `govapi.AuditEventsResponse` (`agentgate/internal/govapi/types.go`) — `workspace_id`,
`events` (array of `govapi.AuditEventView`, newest first), and `next_before_sequence` (the oldest
`sequence_number` in this page; present whenever at least one event was returned, omitted
otherwise). Page backward by resending `next_before_sequence` as the next request's
`before_sequence`; an empty `events` array is the signal there are no more pages.

`AuditEventView` is a **read-only projection** of `audit.StoredRecord` — every operator-facing
field (`id`, `workspace_id`, `sequence_number`, `execution_id`, `timestamp`, `event_type`,
`decision`, `reason`, `principal_agent_id`, `principal_roles`, `principal_on_behalf_of`,
`tool_backend_id`, `tool_name`, `tool_risk`, `policy_version`, `policy_hash`,
`redacted_arguments`) is included as-is. **`canonical_payload`, `prev_hash`, and `row_hash` are
deliberately excluded** — they are internal chain-verification artifacts, not operator data; a
dedicated test (`TestAuditEvents_ListAndPaginate` in `internal/govapi/read_endpoints_test.go`)
asserts they never appear in the raw JSON response, not merely that the Go struct omits them.

**Implementation note, not a contract change:** `audit.Store.ListRecords` had no cursor parameter.
Rather than change that method's signature (which would touch G5's already-verified callers —
`internal/audit/verifier.go`, `qa/g5audit`), a new method,
`ListRecordsBefore(ctx, workspaceID, beforeSequence, limit)`, was added to the `Store` interface
and implemented in both `MemoryStore` and `PostgresStore`, mirroring `ListRecords` exactly except
for the added `WHERE sequence_number < $N` clause.

## Endpoint 2 — Tool inventory

`GET /api/v1/workspaces/{workspace_id}/tools`

Auth: same as above. No query parameters.

Response: `govapi.ToolsResponse` — `workspace_id`, `source` (always the literal string
`"static_configuration"`), and `tools` (array of `govapi.ToolView`: `tool_id` (`{backend_id,
tool_name}`), `known`, `risk`, `registered_fingerprint`).

**`source: "static_configuration"` is load-bearing, not decorative.** `internal/toolregistry.Registry`
is a static, startup-loaded list (`cmd/agentgate/main.go` builds it from a literal
`[]RegistryEntry{...}`) — this endpoint enumerates *configured* tools only. It has no visibility
into tools seen on live traffic that aren't in that list; an unclassified live tool still fails
closed at enforcement time exactly as before, it simply never appears here. Do not build UI
copy/empty-states that imply "no unclassified tools exist" — only "no unclassified tools are
*configured*."

**Implementation note:** `Registry` previously only supported `Lookup` by exact `ToolID` — there
was no enumeration method. Added `Registry.List() []GovernanceRecord`
(`internal/toolregistry/toolregistry.go`), returning every configured entry, sorted deterministically
by `(BackendID, ToolName)`. It never performs drift checking (no live fingerprint exists for a bulk
listing) — `DriftStatus` is intentionally omitted from `ToolView` rather than surfaced as a
misleadingly-always-`"none"` field.

## Explicit non-goals (both endpoints)

- No write path of any kind. Tool classification changes are not part of this contract.
- No new redaction rules — `redacted_arguments` is already redacted at write time
  (`internal/audit/redact.go`); this is a pure read wrapper over that.
- No workspace-scoped tool registry — `internal/toolregistry.Registry` has no workspace concept in
  v1 (matches `decision.Request.WorkspaceID`'s own documented single-tenant status); `workspace_id`
  in the tools response is an echo field, not a filter.

## Error shape

Both endpoints reuse the existing `govapi.ErrorResponse{Error: {Code, Message}}` shape exactly —
`BAD_REQUEST` (malformed `limit`/`before_sequence`, or missing `workspace_id`), `NOT_IMPLEMENTED`
(dependency not configured — should not occur in production wiring, only in test/partial setups),
`INTERNAL_ERROR` (store failure), `UNAUTHORIZED` (from `AdminAuthMiddleware`, unchanged).

## Tests

`agentgate/internal/govapi/read_endpoints_test.go` — auth rejection (both endpoints), the
nil-dependency 501 contract, empty-workspace empty-list behavior, pagination correctness (two
pages, verifying `before_sequence` actually excludes the boundary record), the
chain-verification-field-exclusion check on raw JSON, tool listing against a known fixture
registry, and limit capping/validation.
