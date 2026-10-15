package v1security_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const mcpRevision = "2026-07-28"
const deniedErrorCode = -40301 // must equal authz.DeniedJSONRPCErrorCode

// requireMCPCompat skips like the rest of the live suite when no gateway is configured.
func requireMCPCompat(t *testing.T) *TestConfig {
	t.Helper()
	return requireLiveEnvironment(t)
}

type rpcResp struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// mcpPost sends one spec-shaped 2026-07-28 request. headerMethod lets a test lie in
// the Mcp-Method header while the body says something else.
func mcpPost(t *testing.T, cfg *TestConfig, token string, id int, method, headerMethod, name string, params map[string]any) (int, rpcResp, []byte) {
	t.Helper()
	if params == nil {
		params = map[string]any{}
	}
	params["_meta"] = map[string]any{
		"io.modelcontextprotocol/protocolVersion":    mcpRevision,
		"io.modelcontextprotocol/clientInfo":         map[string]any{"name": "agentgate-compat-test", "version": "1.0.0"},
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
	}
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	req, _ := http.NewRequest(http.MethodPost, cfg.GatewayURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", mcpRevision)
	req.Header.Set("Mcp-Method", headerMethod)
	if name != "" {
		req.Header.Set("Mcp-Name", name)
	}
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", method, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out rpcResp
	if err := json.Unmarshal(raw, &out); err != nil {
		// 2026-07-28 lets a server answer with a request-scoped SSE stream; the
		// final "data:" frame carries the JSON-RPC response.
		for _, line := range bytes.Split(raw, []byte("\n")) {
			if data, ok := bytes.CutPrefix(bytes.TrimSpace(line), []byte("data:")); ok {
				var frame rpcResp
				if json.Unmarshal(bytes.TrimSpace(data), &frame) == nil {
					out = frame
				}
			}
		}
	}
	return resp.StatusCode, out, raw
}

func TestV1_18_RealClientListAndCall(t *testing.T) {
	cfg := requireMCPCompat(t)
	token := cfg.FetchToken(t, "agent-reader", cfg.ReaderSecret)

	status, list, raw := mcpPost(t, cfg, token, 1, "tools/list", "tools/list", "", nil)
	if status != http.StatusOK || list.Error != nil || len(list.Result) == 0 {
		t.Fatalf("tools/list: status %d body %s", status, raw)
	}
	if !bytes.Contains(list.Result, []byte(`list_tickets`)) {
		t.Fatalf("tools/list result lacks list_tickets: %s", list.Result)
	}

	status, call, raw := mcpPost(t, cfg, token, 2, "tools/call", "tools/call", "list_tickets",
		map[string]any{"name": "list_tickets", "arguments": map[string]any{}})
	if status != http.StatusOK || call.Error != nil || string(call.ID) != "2" {
		t.Fatalf("tools/call list_tickets: status %d body %s", status, raw)
	}

	status, denied, raw := mcpPost(t, cfg, token, 3, "tools/call", "tools/call", "delete_ticket",
		map[string]any{"name": "delete_ticket", "arguments": map[string]any{"ticket_id": "TCK-1001"}})
	if status != http.StatusForbidden || denied.Error == nil || denied.Error.Code != deniedErrorCode || string(denied.ID) != "3" {
		t.Fatalf("reader delete_ticket must be a 403 JSON-RPC error with id 3: status %d body %s", status, raw)
	}
}

func TestV1_19_HeaderSpoofDoesNotBypassAuthorization(t *testing.T) {
	cfg := requireMCPCompat(t)
	token := cfg.FetchToken(t, "agent-reader", cfg.ReaderSecret)
	cfg.ResetBackend(t)
	before := cfg.GetBackendState(t)

	// Header claims a harmless list; body is a destructive call.
	status, _, raw := mcpPost(t, cfg, token, 9, "tools/call", "tools/list", "",
		map[string]any{"name": "delete_ticket", "arguments": map[string]any{"ticket_id": "TCK-1001"}})
	if status == http.StatusOK {
		t.Fatalf("spoofed Mcp-Method header bypassed authorization: %s", raw)
	}
	after := cfg.GetBackendState(t)
	if after.Count != before.Count {
		t.Fatalf("backend was invoked despite denial: before %d after %d", before.Count, after.Count)
	}
}

func TestV1_20_BatchBodyDenied(t *testing.T) {
	cfg := requireMCPCompat(t)
	token := cfg.FetchToken(t, "agent-reader", cfg.ReaderSecret)
	cfg.ResetBackend(t)
	before := cfg.GetBackendState(t)

	body := []byte(`[{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_ticket","arguments":{"ticket_id":"TCK-1001"}}}]`)
	req, _ := http.NewRequest(http.MethodPost, cfg.GatewayURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("MCP-Protocol-Version", mcpRevision)
	req.Header.Set("Mcp-Method", "tools/call")
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal("a JSON-RPC batch body must be denied, not passed through")
	}
	if after := cfg.GetBackendState(t); after.Count != before.Count {
		t.Fatalf("batch reached the backend: before %d after %d", before.Count, after.Count)
	}
}

func TestV1_21_DataReadingMethodsDenied(t *testing.T) {
	cfg := requireMCPCompat(t)
	token := cfg.FetchToken(t, "agent-reader", cfg.ReaderSecret)
	cases := []struct {
		method string
		name   string
		params map[string]any
	}{
		{"resources/read", "file:///etc/passwd", map[string]any{"uri": "file:///etc/passwd"}},
		{"prompts/get", "any", map[string]any{"name": "any"}},
	}
	for i, c := range cases {
		status, rpc, raw := mcpPost(t, cfg, token, 30+i, c.method, c.method, c.name, c.params)
		if status != http.StatusForbidden || rpc.Error == nil || rpc.Error.Code != deniedErrorCode {
			t.Fatalf("%s must be denied by the gateway/AgentGate (403 + code %d), not forwarded: status %d body %s", c.method, deniedErrorCode, status, raw)
		}
	}
}

func TestV1_22_SubscriptionsListenDenied(t *testing.T) {
	cfg := requireMCPCompat(t)
	token := cfg.FetchToken(t, "agent-reader", cfg.ReaderSecret)
	status, rpc, raw := mcpPost(t, cfg, token, 40, "subscriptions/listen", "subscriptions/listen", "", map[string]any{})
	if status != http.StatusForbidden || rpc.Error == nil || rpc.Error.Code != deniedErrorCode {
		t.Fatalf("subscriptions/listen must be denied in v1: status %d body %s", status, raw)
	}
}
