package authz

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/Dynamisch-LLC/agentgate/internal/decision"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

// reasonDiscoveryAllow labels the metrics counter for allow-listed discovery
// calls. These are not authorization decisions about a governed action, so they
// are counted in metrics and not written to the tamper-evident audit chain.
const reasonDiscoveryAllow decision.ReasonCode = "discovery_allow"

// discoveryMethods is the complete list of MCP methods AgentGate lets through
// without a tool decision. They only describe the server (versions, capabilities,
// tool names and schemas); none reads backend data or changes state. Every other
// method, including resources/*, prompts/*, subscriptions/* and initialize, is
// still denied. Adding a method here is a deliberate security decision.
var discoveryMethods = map[string]struct{}{
	"server/discover": {},
	"tools/list":      {},
}

// bodyOfCheck returns the request body carried in the ext_authz check.
func bodyOfCheck(req *authv3.CheckRequest) []byte {
	if req == nil || req.Attributes == nil || req.Attributes.Request == nil || req.Attributes.Request.Http == nil {
		return nil
	}
	h := req.Attributes.Request.Http
	if len(h.RawBody) > 0 {
		return h.RawBody
	}
	return []byte(h.Body)
}

// discoveryMethodOf reports the method of a body that is a single JSON object
// whose one "method" member names an allow-listed discovery method.
//
// It is deliberately stricter than encoding/json. A body that one parser reads
// as tools/list and another as tools/call (duplicate "method" keys, differently
// cased keys, a second object after the first) would otherwise let a tool call
// skip authorization if the backend parses it differently from AgentGate. Any
// such ambiguity returns false, which sends the request down the normal path
// where everything except tools/call is denied.
func discoveryMethodOf(body []byte) (string, bool) {
	dec := json.NewDecoder(bytes.NewReader(body))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return "", false
	}
	method, seen := "", false
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return "", false
		}
		key, ok := keyTok.(string)
		if !ok {
			return "", false
		}
		if strings.EqualFold(key, "method") {
			if seen || key != "method" {
				return "", false
			}
			seen = true
			if err := dec.Decode(&method); err != nil {
				return "", false
			}
			continue
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return "", false
		}
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return "", false
	}
	if _, err := dec.Token(); err != io.EOF {
		return "", false
	}
	if !seen {
		return "", false
	}
	if _, ok := discoveryMethods[method]; !ok {
		return "", false
	}
	return method, true
}

// buildDiscoveryAllowResponse allows an allow-listed discovery call. The agent's
// inbound token is stripped and no credential is minted: the backend must never
// see the agent's token, and discovery needs no credential.
func (s *Server) buildDiscoveryAllowResponse() *authv3.CheckResponse {
	return &authv3.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.OK),
			Message: "discovery method allowed by agentgate",
		},
		HttpResponse: &authv3.CheckResponse_OkResponse{
			OkResponse: &authv3.OkHttpResponse{
				Headers: []*corev3.HeaderValueOption{
					{Header: &corev3.HeaderValue{Key: "x-agentgate-decision", Value: "allow"}},
					{Header: &corev3.HeaderValue{Key: "x-agentgate-reason", Value: string(reasonDiscoveryAllow)}},
					{Header: &corev3.HeaderValue{Key: "x-agentgate-execution-id", Value: generateExecutionID()}},
				},
				HeadersToRemove: []string{"authorization"},
			},
		},
	}
}
