package authz

import (
	"context"
	"strings"
	"testing"

	"github.com/Dynamisch-LLC/agentgate/internal/metrics"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc/codes"
)

func TestDiscoveryMethodOf(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		method string
		ok     bool
	}{
		{"tools/list", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, "tools/list", true},
		{"server/discover", `{"jsonrpc":"2.0","id":"d","method":"server/discover","params":{"_meta":{}}}`, "server/discover", true},
		{"whitespace ok", "  \n{ \"jsonrpc\" : \"2.0\", \"method\" : \"tools/list\" }\n", "tools/list", true},
		{"tools/call is not discovery", `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_tickets"}}`, "", false},
		{"resources/read is not discovery", `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"file:///x"}}`, "", false},
		{"initialize is not discovery", `{"jsonrpc":"2.0","id":1,"method":"initialize"}`, "", false},
		{"nested method key is ignored", `{"method":"tools/call","params":{"method":"tools/list"}}`, "", false},
		// Parser-differential attempts: a body that one JSON parser reads as tools/list and
		// another as tools/call must never be treated as discovery.
		{"duplicate method key, list last", `{"method":"tools/call","params":{"name":"delete_ticket"},"method":"tools/list"}`, "", false},
		{"duplicate method key, list first", `{"method":"tools/list","method":"tools/call"}`, "", false},
		{"differently cased key", `{"Method":"tools/list"}`, "", false},
		{"escaped key spelling twice", `{"method":"tools/list","method":"tools/call"}`, "", false},
		{"trailing second object", `{"method":"tools/list"}{"method":"tools/call","params":{"name":"delete_ticket"}}`, "", false},
		{"batch array", `[{"method":"tools/list"}]`, "", false},
		{"method not a string", `{"method":5}`, "", false},
		{"no method", `{"jsonrpc":"2.0","id":1}`, "", false},
		{"empty body", ``, "", false},
		{"garbage", `not json`, "", false},
		{"truncated object", `{"method":"tools/list"`, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			method, ok := discoveryMethodOf([]byte(c.body))
			if ok != c.ok || method != c.method {
				t.Fatalf("discoveryMethodOf(%q) = (%q, %v), want (%q, %v)", c.body, method, ok, c.method, c.ok)
			}
		})
	}
}

func TestCheck_DiscoveryAllowedWithoutDecisionService(t *testing.T) {
	// No decision service and no registry: discovery must not need either.
	s := NewServer(NewAdapter(AdapterConfig{DefaultWorkspaceID: "default", AllowStaticWorkspace: true}), nil)
	resp, err := s.Check(context.Background(), checkWithBody(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetStatus().GetCode() != int32(codes.OK) {
		t.Fatalf("status = %v, want OK", resp.GetStatus())
	}
	ok := resp.GetOkResponse()
	if ok == nil {
		t.Fatal("expected an OK response for tools/list")
	}
	if getOptionHeader(ok.GetHeaders(), "x-agentgate-decision") != "allow" {
		t.Fatal("x-agentgate-decision must be allow")
	}
	if getOptionHeader(ok.GetHeaders(), "x-agentgate-reason") != "discovery_allow" {
		t.Fatalf("x-agentgate-reason = %q, want discovery_allow", getOptionHeader(ok.GetHeaders(), "x-agentgate-reason"))
	}
	// The agent's inbound token must never reach the backend, discovery included.
	removed := false
	for _, h := range ok.GetHeadersToRemove() {
		if strings.EqualFold(h, "authorization") {
			removed = true
		}
	}
	if !removed {
		t.Fatal("authorization header must be removed before the request reaches the backend")
	}
	// No minted credential on a discovery call.
	for _, h := range ok.GetHeaders() {
		if strings.EqualFold(h.GetHeader().GetKey(), "authorization") {
			t.Fatal("discovery must not carry a credential")
		}
	}
}

func TestCheck_DuplicateMethodKeyNeverBypassesAuthorization(t *testing.T) {
	s := NewServer(NewAdapter(AdapterConfig{DefaultWorkspaceID: "default", AllowStaticWorkspace: true}), nil)
	resp, _ := s.Check(context.Background(), checkWithBody(
		`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"delete_ticket","arguments":{"ticket_id":"TCK-1001"}},"method":"tools/list"}`))
	if resp.GetOkResponse() != nil {
		t.Fatal("a body with two method keys must never be allowed through as discovery")
	}
	if resp.GetDeniedResponse() == nil {
		t.Fatal("expected a denial")
	}
}

func TestCheck_UnlistedMethodsStillDenied(t *testing.T) {
	s := NewServer(NewAdapter(AdapterConfig{DefaultWorkspaceID: "default", AllowStaticWorkspace: true}), nil)
	for _, method := range []string{"initialize", "resources/read", "resources/list", "prompts/get", "prompts/list", "subscriptions/listen", "completion/complete", "logging/setLevel", "ping"} {
		resp, _ := s.Check(context.Background(), checkWithBody(`{"jsonrpc":"2.0","id":4,"method":"`+method+`","params":{}}`))
		if resp.GetDeniedResponse() == nil {
			t.Fatalf("%s must be denied", method)
		}
	}
}

func TestCheck_DiscoveryCountedInMetrics(t *testing.T) {
	m := metrics.NewMetrics("test-version")
	s := NewServerWithOptions(NewAdapter(AdapterConfig{DefaultWorkspaceID: "default", AllowStaticWorkspace: true}), nil, ServerOptions{
		DecisionsTotal:          m.DecisionsTotal,
		DecisionDurationSeconds: m.DecisionDurationSeconds,
	})
	_, _ = s.Check(context.Background(), checkWithBody(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	var buf strings.Builder
	if err := m.Registry.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `agentgate_decisions_total{decision="ALLOW",reason="discovery_allow"} 1`) {
		t.Fatalf("discovery allow not counted; metrics:\n%s", buf.String())
	}
}

var _ = authv3.CheckRequest{}
