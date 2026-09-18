package govapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Dynamisch-LLC/agentgate/internal/audit"
)

// G7 Task C: read-only audit-events and tools endpoints. See
// docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md §5 and
// frontend/app/PROPOSED_G8_API_CONTRACTS.md for the reviewed contract these
// tests hold the handlers to.

func TestReadEndpoints_RejectMissingOrWrongToken(t *testing.T) {
	ts, _, _, _ := setupTestServerWithReadDeps("secret-token")
	defer ts.Close()

	paths := []string{
		"/api/v1/workspaces/ws-1/audit-events",
		"/api/v1/workspaces/ws-1/tools",
	}
	for _, path := range paths {
		// Missing auth.
		req, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: request failed: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401 for missing token, got %d", path, resp.StatusCode)
		}

		// Wrong auth.
		req2, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		req2.Header.Set("Authorization", "Bearer wrong-token")
		resp2, err := http.DefaultClient.Do(req2)
		if err != nil {
			t.Fatalf("%s: request failed: %v", path, err)
		}
		resp2.Body.Close()
		if resp2.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401 for wrong token, got %d", path, resp2.StatusCode)
		}
	}
}

func TestReadEndpoints_NotImplementedWithoutDependencies(t *testing.T) {
	// setupTestServer (not setupTestServerWithReadDeps) wires nil for both
	// the audit store and tool registry, per NewHandler's documented
	// nil-is-safe contract.
	ts, _ := setupTestServer("secret-token")
	defer ts.Close()

	for _, path := range []string{
		"/api/v1/workspaces/ws-1/audit-events",
		"/api/v1/workspaces/ws-1/tools",
	} {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: request failed: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Fatalf("%s: expected 501 without dependencies configured, got %d", path, resp.StatusCode)
		}
	}
}

func TestAuditEvents_EmptyWorkspaceReturnsEmptyList(t *testing.T) {
	adminToken := "secret-token"
	ts, _, _, _ := setupTestServerWithReadDeps(adminToken)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/workspaces/never-seen/audit-events", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body AuditEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Events) != 0 {
		t.Fatalf("expected empty events for a workspace with no records, got %d", len(body.Events))
	}
	if body.NextBeforeSequence != nil {
		t.Fatalf("expected no next_before_sequence when there are no events, got %v", *body.NextBeforeSequence)
	}
}

func TestAuditEvents_ListAndPaginate(t *testing.T) {
	adminToken := "secret-token"
	ts, _, auditStore, _ := setupTestServerWithReadDeps(adminToken)
	defer ts.Close()

	ws := "ws-audit"
	for i := 0; i < 5; i++ {
		_, err := auditStore.AppendDecision(context.Background(), audit.DecisionRecord{
			WorkspaceID:      ws,
			Decision:         "ALLOW",
			Reason:           "policy_match",
			PrincipalAgentID: "agent-1",
			ToolBackendID:    "default",
			ToolName:         "read_status",
			ToolRisk:         "read",
			CanonicalPayload: "sensitive-internal-artifact",
			PrevHash:         "sensitive-internal-artifact",
			RowHash:          "sensitive-internal-artifact",
		})
		if err != nil {
			t.Fatalf("seed record %d: %v", i, err)
		}
	}

	get := func(path string) AuditEventsResponse {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, resp.StatusCode)
		}
		var body AuditEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return body
	}

	// First page: limit=2, newest first (sequence 5, 4).
	page1 := get("/api/v1/workspaces/" + ws + "/audit-events?limit=2")
	if len(page1.Events) != 2 {
		t.Fatalf("expected 2 events on first page, got %d", len(page1.Events))
	}
	if page1.Events[0].SequenceNumber != 5 || page1.Events[1].SequenceNumber != 4 {
		t.Fatalf("expected newest-first [5,4], got [%d,%d]", page1.Events[0].SequenceNumber, page1.Events[1].SequenceNumber)
	}
	if page1.NextBeforeSequence == nil || *page1.NextBeforeSequence != 4 {
		t.Fatalf("expected next_before_sequence=4, got %v", page1.NextBeforeSequence)
	}

	// Second page: before_sequence=4 must exclude sequence 4 itself.
	page2 := get("/api/v1/workspaces/" + ws + "/audit-events?limit=2&before_sequence=4")
	if len(page2.Events) != 2 {
		t.Fatalf("expected 2 events on second page, got %d", len(page2.Events))
	}
	if page2.Events[0].SequenceNumber != 3 || page2.Events[1].SequenceNumber != 2 {
		t.Fatalf("expected [3,2], got [%d,%d]", page2.Events[0].SequenceNumber, page2.Events[1].SequenceNumber)
	}

	// Verify redaction-adjacent fields never appear on the wire at all —
	// not just absent from the Go struct, but absent from the raw JSON,
	// since a stray field added later to AuditEventView would compile fine
	// but silently leak these if this test only checked the typed struct.
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/workspaces/"+ws+"/audit-events?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	events, _ := raw["events"].([]any)
	if len(events) != 1 {
		t.Fatalf("expected 1 raw event, got %d", len(events))
	}
	event, _ := events[0].(map[string]any)
	for _, forbidden := range []string{"canonical_payload", "prev_hash", "row_hash"} {
		if _, present := event[forbidden]; present {
			t.Fatalf("security violation: chain-verification field %q leaked into the audit-events API response", forbidden)
		}
	}
}

func TestTools_ListReflectsRegistryConfiguration(t *testing.T) {
	adminToken := "secret-token"
	ts, _, _, _ := setupTestServerWithReadDeps(adminToken)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/workspaces/any-workspace/tools", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body ToolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Source != "static_configuration" {
		t.Fatalf("expected source=static_configuration, got %q", body.Source)
	}
	// setupTestServerWithReadDeps registers exactly read_status and admin_action.
	if len(body.Tools) != 2 {
		t.Fatalf("expected 2 configured tools, got %d: %+v", len(body.Tools), body.Tools)
	}
	byName := make(map[string]ToolView, len(body.Tools))
	for _, tool := range body.Tools {
		byName[tool.ToolID.ToolName] = tool
	}
	readStatus, ok := byName["read_status"]
	if !ok || !readStatus.Known || readStatus.Risk != "read" || readStatus.RegisteredFingerprint == "" {
		t.Fatalf("expected known read_status tool with risk=read and a fingerprint, got %+v", readStatus)
	}
	adminAction, ok := byName["admin_action"]
	if !ok || !adminAction.Known || adminAction.Risk != "destructive" {
		t.Fatalf("expected known admin_action tool with risk=destructive, got %+v", adminAction)
	}
}

func TestAuditEvents_LimitIsCappedAndValidated(t *testing.T) {
	adminToken := "secret-token"
	ts, _, auditStore, _ := setupTestServerWithReadDeps(adminToken)
	defer ts.Close()

	ws := "ws-limit"
	for i := 0; i < 3; i++ {
		if _, err := auditStore.AppendDecision(context.Background(), audit.DecisionRecord{
			WorkspaceID: ws,
			Decision:    "ALLOW",
		}); err != nil {
			t.Fatalf("seed record %d: %v", i, err)
		}
	}

	// A limit above the cap is silently capped, not rejected.
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/workspaces/"+ws+"/audit-events?limit=99999", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for an over-cap limit, got %d", resp.StatusCode)
	}
	var body AuditEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Events) != 3 {
		t.Fatalf("expected all 3 seeded records (well under the cap), got %d", len(body.Events))
	}

	// A non-numeric limit is rejected outright, not silently defaulted.
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/workspaces/"+ws+"/audit-events?limit=not-a-number", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for a non-numeric limit, got %d", resp2.StatusCode)
	}
}
