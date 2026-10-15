package authz

import (
	"encoding/json"
	"fmt"

	"github.com/Dynamisch-LLC/agentgate/internal/decision"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

// DeniedJSONRPCErrorCode is the JSON-RPC error code AgentGate returns for an
// authorization denial. MCP 2026-07-28 reserves -32020..-32099 for itself and
// marks -32000..-32019 legacy, and says application-defined codes belong outside
// -32768..-32000, so this sits outside the JSON-RPC reserved range (mnemonic: HTTP 403).
const DeniedJSONRPCErrorCode = -40301

// rpcIDFromCheck returns the JSON-RPC id (a string or a number) of the request
// carried in the ext_authz check. It returns nil when there is no readable id, so
// the error response omits "id" as MCP 2026-07-28 allows for unreadable requests.
func rpcIDFromCheck(req *authv3.CheckRequest) json.RawMessage {
	body := bodyOfCheck(req)
	var env struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(body, &env); err != nil || len(env.ID) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(env.ID, &v); err != nil {
		return nil
	}
	switch v.(type) {
	case string, float64:
		return env.ID
	}
	return nil
}

// deniedJSONRPCBody renders an authorization denial as a JSON-RPC 2.0 error.
// A nil or empty id omits the "id" member.
func deniedJSONRPCBody(id json.RawMessage, reason decision.ReasonCode, msg, execID string) string {
	message := fmt.Sprintf("AgentGate Authorization Denied: %s", reason)
	if msg != "" {
		message = fmt.Sprintf("AgentGate Authorization Denied: %s (%s)", reason, msg)
	}
	out, _ := json.Marshal(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id,omitempty"`
		Error   any             `json:"error"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]any{
			"code":    DeniedJSONRPCErrorCode,
			"message": message,
			"data":    map[string]string{"reason": string(reason), "execution_id": execID},
		},
	})
	return string(out)
}
