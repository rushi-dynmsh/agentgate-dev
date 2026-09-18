# AgentGate Backend API Reference

**Status:** Current implementation reference
**Backend:** `agentgate/` Go module
**Last reviewed:** 2026-09-18

This document describes the backend surfaces that currently exist in code. It separates public runtime APIs from internal Go package contracts and clearly marks planned endpoints that are not implemented yet.

## 1. Backend Responsibilities

AgentGate owns:

- Identity mapping from trusted gateway claims.
- Tool governance and risk classification.
- Cedar policy evaluation.
- Candidate, active, and historical policy lifecycle.
- Durable authorization audit.
- Fail-closed enforcement.
- The operator governance REST API.
- Envoy `ext_authz` gRPC integration.

AgentGate does **not** own:

- MCP transport or tool discovery. `agentgateway` owns those.
- The actual business operation. The downstream MCP server owns that.
- Per-record ownership. The tool/backend remains the system of record.
- Raw inbound bearer-token passthrough to downstream services.

## 2. Listening Surfaces

The production executable is:

```text
agentgate/cmd/agentgate
```

Default listeners:

| Surface | Default | Purpose |
|---|---:|---|
| HTTP | `:8090` | Health, readiness, and authenticated governance REST API |
| Envoy ext_authz gRPC | `:9001` | Per-request authorization checks from `agentgateway` |

The architecture documentation contains older `:9000` references. The current Go configuration and executable default to `:9001` for the authorization gRPC listener.

## 3. HTTP Authentication

Governance endpoints require the configured admin token.

Configuration:

```text
AGENTGATE_ADMIN_TOKEN
```

Local development default:

```text
agentgate-admin-secret-dev
```

Accepted request forms:

```http
Authorization: Bearer <admin-token>
```

or:

```http
X-AgentGate-Admin-Key: <admin-token>
```

The comparison is constant-time. Missing or invalid credentials return:

```http
401 Unauthorized
```

Response shape:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "valid admin credentials required"
  }
}
```

This is a development/static admin-token mechanism, not OIDC, user login, or per-user authorization.

## 4. Common HTTP Conventions

Base URL:

```text
http://localhost:8090
```

Frontend development URL:

```text
http://localhost:5173
```

Vite proxies `/api` requests to the backend HTTP listener.

JSON requests should include:

```http
Content-Type: application/json
```

Common error shape:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "human-readable explanation"
  }
}
```

Common status codes:

| Status | Meaning |
|---:|---|
| `200` | Successful read, validation, preview, dry-run, activation, or rollback |
| `201` | Candidate policy created |
| `400` | Malformed request, missing path data, or invalid policy |
| `401` | Missing or invalid admin credentials |
| `404` | Workspace policy/version not found |
| `501` | Optional governance integration is not configured |
| `500` | Backend/storage/internal failure |

## 5. Health and Readiness APIs

These endpoints do not use admin authentication.

### `GET /healthz`

Liveness check. It confirms the process can handle an HTTP request.

Successful response:

```http
200 OK
```

```text
ok
```

PowerShell:

```powershell
Invoke-WebRequest http://localhost:8090/healthz
```

### `GET /readyz`

Readiness check. It confirms the listener is bound and, when configured, dependent checks such as PostgreSQL pass.

Successful response:

```http
200 OK
```

```text
ready
```

Possible unavailable responses:

```http
503 Service Unavailable
```

```text
not ready
```

or:

```text
dependency not ready: <reason>
```

PowerShell:

```powershell
Invoke-WebRequest http://localhost:8090/readyz
```

## 6. Policy Governance APIs

All policy endpoints require admin authentication.

The workspace path is required on every route:

```text
/api/v1/workspaces/{workspace_id}/policies
```

The local default workspace is:

```text
default
```

### Policy record

The backend returns this policy record shape:

```json
{
  "id": 1,
  "workspace_id": "default",
  "version": "6cea...ef277",
  "content": "permit(...)",
  "state": "active",
  "description": "Agent tool access policy",
  "created_at": "2026-09-18T12:00:00Z",
  "activated_at": "2026-09-18T12:05:00Z"
}
```

Valid states:

```text
candidate
active
historical
```

### `GET /api/v1/workspaces/{workspace_id}/policies`

Lists every policy version in the workspace.

Request:

```http
GET /api/v1/workspaces/default/policies
Authorization: Bearer <admin-token>
```

Response:

```json
{
  "workspace_id": "default",
  "policies": [
    {
      "workspace_id": "default",
      "version": "v1.0.0",
      "content": "permit(...)",
      "state": "historical",
      "description": "default baseline policy",
      "created_at": "2026-09-18T12:00:00Z"
    }
  ]
}
```

### `GET /api/v1/workspaces/{workspace_id}/policies/{version}`

Returns one policy version.

Request:

```http
GET /api/v1/workspaces/default/policies/v1.0.0
Authorization: Bearer <admin-token>
```

Returns the `PolicyRecord` object.

Not found:

```http
404 Not Found
```

### `POST /api/v1/workspaces/{workspace_id}/policies/validate`

Validates Cedar source without storing or activating it.

Request:

```json
{
  "content": "permit(principal, action, resource);"
}
```

Valid response:

```json
{
  "valid": true,
  "version": "content-hash"
}
```

Invalid Cedar response:

```json
{
  "valid": false,
  "errors": ["<Cedar validation error>"]
}
```

Important: invalid Cedar uses HTTP `200` with `valid: false`; it is a validation result, not a transport failure.

### `POST /api/v1/workspaces/{workspace_id}/policies`

Creates a validated candidate policy.

Request:

```json
{
  "content": "permit(principal, action, resource);",
  "description": "Allow reader access to read tools"
}
```

Successful response:

```http
201 Created
```

The response is a `PolicyRecord` with:

```text
state = candidate
```

Invalid policy:

```http
400 Bad Request
```

```json
{
  "error": {
    "code": "INVALID_POLICY",
    "message": "..."
  }
}
```

### `POST /api/v1/workspaces/{workspace_id}/policies/{version}/preview`

Evaluates sample inputs against one stored policy version.

Request:

```json
{
  "sample_requests": [
    {
      "principal_id": "agent-reader",
      "principal_roles": ["reader"],
      "resource_id": "default/read_status",
      "resource_risk": "read"
    }
  ]
}
```

Response:

```json
{
  "version": "candidate-version",
  "results": [
    {
      "allowed": true,
      "matched": true,
      "had_error": false
    }
  ]
}
```

This endpoint previews a single version. It is distinct from the active-versus-candidate dry-run endpoint.

### `POST /api/v1/workspaces/{workspace_id}/policies/{version}/dryrun`

Compares a candidate policy with the current active policy without changing active state.

Request:

```json
{
  "sample_requests": [
    {
      "execution_id": "dryrun-1",
      "principal_id": "agent-reader",
      "principal_roles": ["reader"],
      "on_behalf_of": "user-1",
      "backend_id": "default",
      "tool_name": "read_status",
      "risk": "read"
    }
  ]
}
```

Response:

```json
{
  "candidate_version": "candidate-version",
  "results": [
    {
      "active_decision": "DENY",
      "active_reason": "no_matching_policy",
      "active_policy_version": "v1.0.0",
      "candidate_decision": "ALLOW",
      "candidate_reason": "policy_allow",
      "candidate_policy_version": "candidate-version",
      "changed": true
    }
  ]
}
```

Dry-run must not mutate the active policy or active engine.

If governance integration is unavailable:

```http
501 Not Implemented
```

### `POST /api/v1/workspaces/{workspace_id}/policies/{version}/activate`

Atomically makes a validated candidate the active policy.

Request:

```http
POST /api/v1/workspaces/default/policies/candidate-version/activate
Authorization: Bearer <admin-token>
```

No request body is required.

Response:

```json
{
  "workspace_id": "default",
  "active_version": "candidate-version",
  "activated_at": "2026-09-18T12:05:00Z"
}
```

The backend changes the previous active version to historical and swaps the active in-memory policy engine atomically.

The current API does not accept an activation reason. The frontend reason field is currently UI-only.

### `POST /api/v1/workspaces/{workspace_id}/policies/rollback`

Makes a historical policy active again.

Request:

```json
{
  "target_version": "v1.0.0"
}
```

Response:

```json
{
  "workspace_id": "default",
  "active_version": "v1.0.0",
  "rolled_back_from": "candidate-version",
  "activated_at": "2026-09-18T12:10:00Z"
}
```

Rollback is a policy activation operation and is audited through the backend mutation/audit integration.

## 7. Runtime Authorization API: Envoy ext_authz gRPC

The runtime authorization entrypoint is the standard Envoy v3 service:

```text
envoy.service.auth.v3.Authorization/Check
```

AgentGate registers this service on the configured gRPC listener, currently `:9001`.

This is not a browser REST endpoint. `agentgateway` calls it for each governed MCP request.

### Incoming request

The gateway sends an Envoy `CheckRequest` containing:

- HTTP request metadata.
- MCP JSON-RPC request body in the HTTP body fields.
- Verified JWT claims in gateway filter metadata.
- Trusted route context extensions when configured.
- Optional execution/request ID headers.
- Backend and schema-fingerprint context used by the adapter.

The MCP body must be JSON-RPC `tools/call`:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "read_status",
    "arguments": {
      "verbose": true
    }
  }
}
```

Only `tools/call` is authorized by the current adapter. Missing, malformed, or different methods are denied.

### Trusted identity mapping

The adapter reads claims only from gateway-verified JWT metadata. It maps configured claims into:

```text
agent ID
roles
on-behalf-of identity
workspace
```

Client-supplied identity headers are not trusted as an alternative identity source.

Workspace resolution order:

1. Verified JWT `workspace_id` claim.
2. Verified JWT `workspace` claim.
3. Trusted gateway route context extension `workspace_id`.
4. Explicit static fallback for the current single-workspace development deployment.
5. Deny if no trusted workspace can be resolved and fallback is disabled.

### Runtime decision input

The adapter converts the gateway request into the internal decision contract:

```text
execution_id
workspace_id
identity.agent_id
identity.roles
identity.on_behalf_of
tool.backend_id
tool.name
classification.known
classification.risk
resolved declared arguments
```

### Tool governance checks

Before Cedar evaluation, AgentGate checks:

- Tool registry exists.
- Backend ID and tool name are registered.
- Tool risk is classified.
- Live schema fingerprint matches the registered fingerprint.
- Declared arguments have valid types.
- Required arguments are present.
- Explicit nulls and undeclared arguments fail closed.

### Authorization result: allow

An allowed check returns Envoy gRPC `OK` with response headers such as:

```text
x-agentgate-decision: allow
x-agentgate-reason: policy_allow
x-agentgate-execution-id: <id>
x-agentgate-policy-version: <version>
```

Only after durable audit succeeds can an `ALLOW` result be returned.

### Authorization result: deny

A denied check returns Envoy `PermissionDenied` and HTTP `403` response details with headers such as:

```text
x-agentgate-decision: deny
x-agentgate-reason: policy_deny
x-agentgate-execution-id: <id>
```

The downstream MCP backend must not receive a denied request.

### Fail-closed runtime actions

The runtime returns DENY for:

- Missing or invalid identity.
- Missing trusted workspace.
- Unknown tool.
- Missing risk classification.
- Schema drift.
- Malformed JSON-RPC.
- Unsupported MCP method.
- Invalid or undeclared arguments.
- No active policy.
- Cedar evaluation error.
- Audit persistence failure.
- Decision service failure.
- AgentGate outage as observed by the gateway.

Adaptation failures also attempt to create a durable DENY audit record using a safe fallback request.

## 8. Internal Decision Contract

The gRPC adapter calls the internal decision service rather than Cedar directly.

Conceptual request:

```json
{
  "execution_id": "exec-123",
  "workspace_id": "default",
  "identity": {
    "agent_id": "agent-reader",
    "on_behalf_of": "user-1",
    "roles": ["reader"]
  },
  "tool": {
    "backend_id": "default",
    "name": "read_status"
  },
  "classification": {
    "known": true,
    "risk": "read"
  },
  "arguments": {}
}
```

Conceptual result:

```json
{
  "decision": "ALLOW",
  "reason": "policy_allow",
  "message": "",
  "policy_version": "candidate-version",
  "execution_id": "exec-123"
}
```

The decision engine is internal Go code, not a public HTTP API.

## 9. Policy Actions and Rules

The current Cedar model uses one governed action:

```text
AgentGate::Action::"InvokeTool"
```

Common risk values:

```text
read
write
destructive
```

Typical role behavior:

```text
reader + read         -> ALLOW when policy grants it
reader + write        -> DENY unless explicitly granted
admin + read/write    -> ALLOW when policy grants it
any role + destructive -> DENY by default
payer + write + amount <= limit -> ALLOW when declared by policy
```

Cedar remains deny-by-default. No matching policy is not an error that can become ALLOW.

## 10. Storage and Durability

Policy storage:

- PostgreSQL is the persistent production store.
- Active policy engines are held in AgentGate memory.
- Development can run with an in-memory store.
- Activation replaces the active valid engine atomically.
- Failed loads leave the last known-good engine active.

Audit storage:

- PostgreSQL append-only audit records.
- Redacted arguments.
- SHA-256 row chaining.
- Database immutability trigger.
- Separate application and migration privileges.
- An ALLOW is converted to DENY if durable audit persistence fails.

## 11. APIs Not Implemented Yet

These are planned backend read surfaces, not current routes:

```http
GET /api/v1/workspaces/{workspace_id}/audit-events
GET /api/v1/workspaces/{workspace_id}/tools
```

They are needed to make the frontend Audit and Tools pages live.

The planned tools endpoint should list statically configured governance records. It should not imply dynamic discovery of every unknown live tool.

The planned audit endpoint should return redacted operator-facing records and should exclude internal chain-verification fields by default, such as `canonical_payload`, `prev_hash`, and `row_hash`.

Other pending backend work:

- Downstream scoped credentials/token exchange.
- On-behalf-of propagation to downstream MCP servers.
- No raw inbound-token passthrough.
- Configurable claims mapping instead of hardcoded claim names.
- Production OIDC/admin authentication.
- MCP revision compatibility matrix.
- Rate limiting and session/aggregate quotas.
- Response-side governance.
- Backup/recovery and production deployment hardening.

## 12. Quick API Test Commands

Set the development token:

```powershell
$token = "agentgate-admin-secret-dev"
```

Health:

```powershell
Invoke-WebRequest http://localhost:8090/healthz
Invoke-WebRequest http://localhost:8090/readyz
```

List policies:

```powershell
Invoke-RestMethod `
  -Uri http://localhost:8090/api/v1/workspaces/default/policies `
  -Headers @{ Authorization = "Bearer $token" }
```

Validate Cedar:

```powershell
$body = @{ content = 'permit(principal, action, resource);' } | ConvertTo-Json
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8090/api/v1/workspaces/default/policies/validate `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body $body
```

Create candidate:

```powershell
$body = @{
  content = 'permit(principal, action, resource);'
  description = 'Test candidate policy'
} | ConvertTo-Json
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8090/api/v1/workspaces/default/policies `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body $body
```

Activate a known version:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8090/api/v1/workspaces/default/policies/<version>/activate `
  -Headers @{ Authorization = "Bearer $token" }
```

Rollback:

```powershell
$body = @{ target_version = "<version>" } | ConvertTo-Json
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8090/api/v1/workspaces/default/policies/rollback `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body $body
```

## 13. Frontend Relationship

The React frontend calls only the authenticated governance REST API:

```text
frontend/app
  -> HttpGovernanceClient
  -> Vite /api proxy
  -> AgentGate :8090
```

The frontend does not call the ext_authz gRPC service and does not execute MCP tools.

The runtime path is separate:

```text
MCP client
  -> agentgateway
  -> AgentGate :9001 ext_authz Check
  -> downstream MCP server only after ALLOW
```
