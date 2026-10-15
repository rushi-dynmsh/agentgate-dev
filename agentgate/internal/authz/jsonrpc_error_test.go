package authz

import (
	"encoding/json"
	"testing"

	"github.com/Dynamisch-LLC/agentgate/internal/decision"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

func checkWithBody(body string) *authv3.CheckRequest {
	return &authv3.CheckRequest{Attributes: &authv3.AttributeContext{Request: &authv3.AttributeContext_Request{
		Http: &authv3.AttributeContext_HttpRequest{Method: "POST", Path: "/", RawBody: []byte(body)},
	}}}
}

func TestRPCIDFromCheck_NumericID(t *testing.T) {
	got := rpcIDFromCheck(checkWithBody(`{"jsonrpc":"2.0","id":7,"method":"tools/call"}`))
	if string(got) != "7" {
		t.Fatalf("id = %s, want 7", got)
	}
}

func TestRPCIDFromCheck_StringID(t *testing.T) {
	got := rpcIDFromCheck(checkWithBody(`{"jsonrpc":"2.0","id":"abc","method":"tools/call"}`))
	if string(got) != `"abc"` {
		t.Fatalf("id = %s, want \"abc\"", got)
	}
}

func TestRPCIDFromCheck_MissingOrInvalidIsOmitted(t *testing.T) {
	for _, body := range []string{``, `not json`, `{"jsonrpc":"2.0","method":"x"}`, `[{"id":1}]`, `{"id":{"nested":1}}`, `{"id":null}`, `{"id":true}`} {
		if got := rpcIDFromCheck(checkWithBody(body)); len(got) != 0 {
			t.Fatalf("body %q: id = %s, want none", body, got)
		}
	}
	if got := rpcIDFromCheck(nil); len(got) != 0 {
		t.Fatalf("nil request: id = %s, want none", got)
	}
}

func TestDeniedBody_IsJSONRPCError(t *testing.T) {
	body := deniedJSONRPCBody(json.RawMessage("7"), decision.ReasonPolicyDeny, "an explicit policy denied this request", "exec-1")
	var got struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Reason      string `json:"reason"`
				ExecutionID string `json:"execution_id"`
			} `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("body is not JSON: %v\n%s", err, body)
	}
	if got.JSONRPC != "2.0" || string(got.ID) != "7" {
		t.Fatalf("envelope wrong: %+v", got)
	}
	if got.Error.Code != DeniedJSONRPCErrorCode {
		t.Fatalf("code = %d, want %d", got.Error.Code, DeniedJSONRPCErrorCode)
	}
	if got.Error.Message != "AgentGate Authorization Denied: policy_deny (an explicit policy denied this request)" {
		t.Fatalf("message = %q", got.Error.Message)
	}
	if got.Error.Data.Reason != "policy_deny" || got.Error.Data.ExecutionID != "exec-1" {
		t.Fatalf("data = %+v", got.Error.Data)
	}
}

func TestDeniedBody_OmitsIDWhenUnreadable(t *testing.T) {
	body := deniedJSONRPCBody(nil, decision.ReasonMalformedRequest, "bad body", "exec-2")
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("body is not JSON: %v\n%s", err, body)
	}
	if _, present := raw["id"]; present {
		t.Fatalf("id must be omitted when unreadable, got body %s", body)
	}
}

func TestDeniedCodeIsOutsideReservedRanges(t *testing.T) {
	if DeniedJSONRPCErrorCode >= -32768 && DeniedJSONRPCErrorCode <= -32000 {
		t.Fatalf("code %d is inside the JSON-RPC reserved range; MCP 2026-07-28 says to allocate outside it", DeniedJSONRPCErrorCode)
	}
}

func TestDeniedBody_UnknownMethodIsJSONRPC(t *testing.T) {
	s := NewServer(NewAdapter(AdapterConfig{DefaultWorkspaceID: "default", AllowStaticWorkspace: true}), nil)
	resp, _ := s.Check(t.Context(), checkWithBody(`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"file:///x"}}`))
	denied := resp.GetDeniedResponse()
	if denied == nil {
		t.Fatal("expected a denied response for an unlisted method")
	}
	var env struct {
		ID    json.RawMessage    `json:"id"`
		Error struct{ Code int } `json:"error"`
	}
	if err := json.Unmarshal([]byte(denied.GetBody()), &env); err != nil || string(env.ID) != "3" || env.Error.Code != DeniedJSONRPCErrorCode {
		t.Fatalf("denied body not a JSON-RPC error for id 3: %q (err %v)", denied.GetBody(), err)
	}
	if getOptionHeader(denied.GetHeaders(), "content-type") != "application/json" {
		t.Fatal("denied response must set content-type: application/json")
	}
	if getOptionHeader(denied.GetHeaders(), "x-agentgate-decision") != "deny" {
		t.Fatal("x-agentgate-decision header must be preserved")
	}
}
