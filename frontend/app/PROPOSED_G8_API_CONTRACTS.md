# Proposed G8 Read API Contracts

**UI location:** `frontend/app/`
**Status:** Draft for Backend review

This repository's working React application is `frontend/app`; there is no separate `admin-ui/` directory. These contracts are proposed for the next checkpoint and are not wired into the UI yet.

## Audit Events

### Request

```http
GET /api/v1/workspaces/{workspace_id}/audit-events?limit=50&before_sequence=120
Authorization: Bearer <admin-token>
```

- `limit` is optional, defaults to 50, and is capped by the backend at 100.
- `before_sequence` is optional and returns records with sequence numbers before the supplied value.
- Results should be returned newest first for operator scanning.
- The next request can use the oldest returned `sequence_number` as its next `before_sequence` value.

### Response

```json
{
  "workspace_id": "default",
  "events": [
    {
      "id": 42,
      "workspace_id": "default",
      "sequence_number": 119,
      "execution_id": "exec-123",
      "timestamp": "2026-09-18T12:00:00Z",
      "event_type": "decision",
      "decision": "DENY",
      "reason": "policy_deny",
      "principal_agent_id": "agent-reader",
      "principal_roles": ["reader"],
      "principal_on_behalf_of": "user-1",
      "tool_backend_id": "default",
      "tool_name": "write_status",
      "tool_risk": "write",
      "policy_version": "v1.0.0",
      "policy_hash": "...",
      "redacted_arguments": {}
    }
  ],
  "next_before_sequence": 119
}
```

The UI needs the operator-facing fields from `audit.StoredRecord`: identity, workspace, execution, timestamp, event type, decision, reason, tool, policy provenance, and already-redacted arguments.

`canonical_payload`, `prev_hash`, and `row_hash` are chain-verification artifacts. They should be excluded from the normal response and remain available only to a dedicated verification/admin workflow. The UI must never reconstruct or expose raw arguments; it should display only `redacted_arguments`.

## Tool Inventory

### Request

```http
GET /api/v1/workspaces/{workspace_id}/tools
Authorization: Bearer <admin-token>
```

### Response

```json
{
  "workspace_id": "default",
  "source": "static_configuration",
  "tools": [
    {
      "tool_id": {
        "backend_id": "default",
        "tool_name": "read_status"
      },
      "known": true,
      "risk": "read",
      "registered_fingerprint": "..."
    }
  ]
}
```

`source: static_configuration` is intentional. This endpoint lists tools configured in AgentGate's startup registry. It does not discover new or unclassified tools from live traffic. Unknown live tools still fail closed and are not added to this list automatically.

This is read-only. Tool classification changes are not part of this contract; there is no UI write path proposed here.

## Frontend Review Notes

- Both endpoints must use the existing admin bearer-token middleware.
- Errors should reuse the existing `{ "error": { "code": "...", "message": "..." } }` shape.
- The frontend will not wire Audit or Tools to these endpoints until Backend implements and freezes the response shape.
