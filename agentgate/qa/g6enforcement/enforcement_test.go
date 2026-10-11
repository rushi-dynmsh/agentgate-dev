package g6enforcement_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Dynamisch-LLC/agentgate/internal/argdecl"
	"github.com/Dynamisch-LLC/agentgate/internal/audit"
	"github.com/Dynamisch-LLC/agentgate/internal/authz"
	"github.com/Dynamisch-LLC/agentgate/internal/decision"
	"github.com/Dynamisch-LLC/agentgate/internal/fixturepolicy"
	"github.com/Dynamisch-LLC/agentgate/internal/identity"
	"github.com/Dynamisch-LLC/agentgate/internal/toolregistry"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultGatewayURL = "http://localhost:3000"
	defaultBackendURL = "http://localhost:9101"
)

func getGatewayURL() string {
	if u := os.Getenv("AGENTGATE_GATEWAY_URL"); u != "" {
		return u
	}
	return defaultGatewayURL
}

func getBackendURL() string {
	if u := os.Getenv("AGENTGATE_BACKEND_COUNT_URL"); u != "" {
		return u
	}
	return defaultBackendURL
}

type backendCountResponse struct {
	Count       int `json:"count"`
	Invocations []struct {
		Timestamp string `json:"timestamp"`
		ToolName  string `json:"tool_name"`
	} `json:"invocations"`
}

func resetBackendCount(t *testing.T) {
	t.Helper()
	resp, err := http.Post(getBackendURL()+"/_g6/reset", "application/json", nil)
	if err != nil {
		t.Fatalf("failed to reset backend count at %s: %v", getBackendURL(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to reset backend count: status %d", resp.StatusCode)
	}
}

func getBackendCount(t *testing.T) int {
	t.Helper()
	resp, err := http.Get(getBackendURL() + "/_g6/count")
	if err != nil {
		t.Fatalf("failed to get backend count from %s: %v", getBackendURL(), err)
	}
	defer resp.Body.Close()

	var countResp backendCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		t.Fatalf("failed to decode backend count: %v", err)
	}
	return countResp.Count
}

// isLiveGatewayAvailable reports whether the deploy/g6 Docker topology's probe
// backend is reachable. It gates the *_LiveE2E test functions below, which
// each call skipIfLiveUnavailable to skip loudly (not silently) when it isn't.
func isLiveGatewayAvailable() bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(getBackendURL() + "/healthz")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// skipIfLiveUnavailable skips the calling test with a clear, actionable message
// when the live deploy/g6 topology isn't reachable, instead of silently
// substituting a different (in-process) code path — see G7 Task A closeout.
func skipIfLiveUnavailable(t *testing.T) {
	t.Helper()
	if !isLiveGatewayAvailable() {
		t.Skipf("SKIPPED: no live gateway/backend reachable at %s — start the deploy/g6 "+
			"docker-compose topology (or set AGENTGATE_BACKEND_COUNT_URL) to run this live E2E "+
			"suite; see deploy/g6/README.md", getBackendURL())
	}
}

func sendMCPRequest(url string, headers map[string]string, body []byte) (int, string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(respBytes), nil
}

// jwtMetadata builds the MetadataContext an in-process test must set to
// establish identity. internal/authz.Adapter.extractClaims reads identity
// ONLY from Attributes.MetadataContext.FilterMetadata["envoy.filters.http.jwt_authn"]
// (agentgateway's post-verification JWT claims) — never from HTTP headers
// (docs/SECURITY/PRODUCTION-INVARIANTS.md; adapter.go's own comment: "Unverified
// client headers ... MUST NEVER be trusted as authenticated identity"). A test
// that sets only Headers for identity fields (x-agent-id, x-roles, etc.) is
// silently testing "missing identity" in the unit harness, not whatever
// specific violation it names — found while fixing G7 Task A; corrected for
// every scenario below that needs to reach a check *past* identity mapping.
func jwtMetadata(claims map[string]any) *corev3.Metadata {
	jwtClaims, _ := structpb.NewStruct(claims)
	return &corev3.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			"envoy.filters.http.jwt_authn": jwtClaims,
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// In-Process Checkpoint Harness for 100% Deterministic Scenario Verification
// ─────────────────────────────────────────────────────────────────────────────

type inProcessHarness struct {
	server       *authz.Server
	toolReg      *toolregistry.Registry
	evaluator    audit.DecisionEvaluator
	backendCalls int
}

func setupInProcessHarness(t *testing.T) *inProcessHarness {
	t.Helper()
	eng, err := decision.NewEngine([]byte(fixturepolicy.CedarSource))
	if err != nil {
		t.Fatalf("setup cedar engine: %v", err)
	}
	return setupInProcessHarnessWithEvaluator(t, &staticEvaluator{eng: eng})
}

// setupInProcessHarnessWithEvaluator builds the same adapter/registry topology as
// setupInProcessHarness but lets the caller substitute the decision evaluator —
// used by TestScenario07_AgentGateUnavailable_Unit to simulate a genuine
// decision-service outage rather than asserting on an unused local variable.
func setupInProcessHarnessWithEvaluator(t *testing.T, evaluator audit.DecisionEvaluator) *inProcessHarness {
	t.Helper()

	mapper, err := identity.NewMapper(identity.MapperConfig{
		AgentIDClaim:    "sub",
		RolesClaim:      "roles",
		OnBehalfOfClaim: "obo",
	})
	if err != nil {
		t.Fatalf("setup mapper: %v", err)
	}

	readSchema := []byte(`{"type":"object","properties":{"verbose":{"type":"boolean"}}}`)
	readFP, _ := toolregistry.FingerprintSchema(readSchema)

	toolReg, err := toolregistry.NewRegistry([]toolregistry.RegistryEntry{
		{
			ToolID:                toolregistry.ToolID{BackendID: "mcp-probe", ToolName: "read_status"},
			Risk:                  toolregistry.RiskRead,
			RegisteredFingerprint: readFP,
		},
		{
			ToolID:                toolregistry.ToolID{BackendID: "mcp-probe", ToolName: "admin_action"},
			Risk:                  toolregistry.RiskDestructive,
			RegisteredFingerprint: readFP,
		},
		{
			ToolID:                toolregistry.ToolID{BackendID: "mcp-probe", ToolName: "drifted_tool"},
			Risk:                  toolregistry.RiskRead,
			RegisteredFingerprint: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	})
	if err != nil {
		t.Fatalf("setup registry: %v", err)
	}

	readDecls, _ := argdecl.NewDeclarationSet([]argdecl.Declaration{
		{Name: "verbose", Type: argdecl.ArgTypeBool, Required: false},
	})

	adapter := authz.NewAdapter(authz.AdapterConfig{
		DefaultWorkspaceID:   "default",
		AllowStaticWorkspace: true,
		DefaultBackendID:     "mcp-probe",
		IdentityMapper:       mapper,
		ToolRegistry:         toolReg,
		ArgDeclarations: map[string]*argdecl.DeclarationSet{
			"read_status": readDecls,
		},
	})

	memStore := audit.NewMemoryStore()
	auditSvc := audit.NewAuditedDecisionService(evaluator, memStore, audit.NewRedactor("", nil))
	srv := authz.NewServer(adapter, auditSvc)

	return &inProcessHarness{
		server:    srv,
		toolReg:   toolReg,
		evaluator: evaluator,
	}
}

// staticEvaluator evaluates against a fixed Cedar engine and records the last
// decision.Request it received, so tests can assert on exactly what reached
// policy evaluation (e.g. that an undeclared argument never leaked into it).
type staticEvaluator struct {
	eng *decision.Engine

	mu      sync.Mutex
	lastReq decision.Request
}

func (s *staticEvaluator) EvaluateWithActivePolicy(_ context.Context, _ string, req decision.Request) (decision.Result, error) {
	s.mu.Lock()
	s.lastReq = req
	s.mu.Unlock()
	return s.eng.Evaluate(req), nil
}

func (s *staticEvaluator) LastRequest() decision.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastReq
}

// failingEvaluator simulates AgentGate's core decision path being genuinely
// unavailable (Cedar unreachable, database down, process crashed) — used to
// verify Scenario 7's real requirement: a decision-evaluation failure must
// deny, never implicitly allow.
type failingEvaluator struct{}

func (failingEvaluator) EvaluateWithActivePolicy(_ context.Context, _ string, _ decision.Request) (decision.Result, error) {
	return decision.Result{}, errors.New("simulated AgentGate outage: decision engine unreachable")
}

func (h *inProcessHarness) executeCall(req *authv3.CheckRequest) (bool, int32) {
	resp, err := h.server.Check(context.Background(), req)
	if err != nil || resp == nil || resp.Status == nil {
		return false, int32(codes.Internal)
	}
	if resp.Status.Code == int32(codes.OK) {
		h.backendCalls++
		return true, resp.Status.Code
	}
	return false, resp.Status.Code
}

// ─────────────────────────────────────────────────────────────────────────────
// 12 Mandatory Gate G6 Scenarios
//
// Each scenario has two independent test functions:
//   - TestScenarioNN_<Name>       — always runs in-process; fast, deterministic,
//                                    proves the adapter/decision/audit logic in
//                                    isolation from any network topology.
//   - TestScenarioNN_<Name>_LiveE2E — proves the same behavior through the real
//                                    agentgateway -> AgentGate -> probe-mcp
//                                    topology. Skips loudly (t.Skip, with a
//                                    clear message) rather than silently
//                                    falling back when that topology isn't up.
//
// Before G7 Task A, these were one function each: the live assertions were
// wrapped in "if isLiveGatewayAvailable()" and unconditionally followed by the
// in-process assertions, so every run reported 12/12 passing with zero live
// coverage in CI or a bare `go test ./...`, with no indication that the live
// path had been skipped rather than proven. See
// docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md §2 for the full finding.
// ─────────────────────────────────────────────────────────────────────────────

// Scenario 1: Authenticated + allowed known tool -> Backend count = 1

func TestScenario01_AuthenticatedAllowedKnownTool(t *testing.T) {
	h := setupInProcessHarness(t)
	jwtClaims, _ := structpb.NewStruct(map[string]any{
		"sub":          "agent-reader",
		"roles":        "reader",
		"workspace_id": "default",
	})
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"read_status","arguments":{"verbose":true}}}`,
				},
			},
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.filters.http.jwt_authn": jwtClaims,
				},
			},
		},
	}
	allowed, code := h.executeCall(req)
	if !allowed || code != int32(codes.OK) || h.backendCalls != 1 {
		t.Fatalf("expected ALLOW with code OK, got allowed=%v code=%d backendCount=%d", allowed, code, h.backendCalls)
	}
}

func TestScenario01_AuthenticatedAllowedKnownTool_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"read_status","arguments":{"verbose":true}}}`
	status, _, err := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), []byte(body))
	if err != nil || status != http.StatusOK {
		t.Fatalf("expected HTTP 200 on allowed tool, got status %d err %v", status, err)
	}
	if count := getBackendCount(t); count != 1 {
		t.Fatalf("expected backend count 1, got %d", count)
	}
}

// Scenario 2: Authenticated + denied known tool -> Backend count = 0

func TestScenario02_AuthenticatedDeniedKnownTool(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"admin_action"}}`,
				},
			},
			// Valid, non-ambiguous identity — this scenario proves an
			// *authenticated* reader is denied by policy for a destructive
			// tool, not merely denied for lacking identity.
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-reader",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: expected DENY, got allowed=%v code=%d backendCount=%d", allowed, code, h.backendCalls)
	}
}

func TestScenario02_AuthenticatedDeniedKnownTool_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"admin_action"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on denied tool, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: denied tool reached backend count=%d", count)
	}
}

// Scenario 3: Unknown tool -> Backend count = 0

func TestScenario03_UnknownTool(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"completely_unknown_op"}}`,
				},
			},
			// Valid identity — this scenario proves the *tool registry
			// lookup* denies, not that identity was missing.
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-reader",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: unknown tool was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario03_UnknownTool_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"completely_unknown_op"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on unknown tool, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: unknown tool reached backend count=%d", count)
	}
}

// Scenario 4: Missing identity / unauthenticated -> Backend count = 0

func TestScenario04_MissingIdentity(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"read_status"}}`,
				},
			},
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: missing identity call was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario04_MissingIdentity_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"read_status"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), map[string]string{}, []byte(body))
	// With no JWT at all, agentgateway's jwtAuth (mode: strict) rejects the
	// request itself — 401 — before it ever reaches AgentGate's ext_authz
	// callout, so AgentGate's own 403 is never produced for this specific
	// case (defense in depth: two independent layers both deny, at
	// different points). Confirmed live (G7 Task A) — not guessed.
	if status != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401 (gateway-level JWT rejection) on missing identity, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: unauthenticated call reached backend count=%d", count)
	}
}

// Scenario 5: Ambiguous identity (on_behalf_of == agent_id) -> Backend count = 0

func TestScenario05_AmbiguousIdentity(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"read_status"}}`,
				},
			},
			// The ambiguity is agent_id == on_behalf_of in the *authenticated*
			// JWT claims (identity.Mapper's actual check) — client headers are
			// never consulted for identity, so setting only headers here would
			// deny for missing identity, not for the ambiguity this proves.
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-same",
				"obo":          "agent-same",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: ambiguous identity was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario05_AmbiguousIdentity_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"read_status"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-same",
		"obo":   "agent-same",
		"roles": "reader",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on ambiguous identity, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: ambiguous identity call reached backend count=%d", count)
	}
}

// Scenario 6: Malformed authorization request (non-tool call) -> Backend count = 0

func TestScenario06_MalformedAuthzRequest_NonToolMethod(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{"x-agent-id": "agent-reader", "x-roles": "reader"},
					Body:    `{"jsonrpc":"2.0","id":6,"method":"initialize","params":{}}`,
				},
			},
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: non-tool method was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario06_MalformedAuthzRequest_NonToolMethod_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":6,"method":"initialize","params":{"protocolVersion":"2026-07-28"}}`
	// Valid identity, so this proves AgentGate's own method check denies it
	// (adapter.go parses/rejects the method before identity is even
	// examined) — not that agentgateway's gateway-level JWT check denied it
	// first for an unrelated reason (see Scenario04's 401 case, which is
	// exactly that different failure mode).
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected non-tool call to be denied by authz, got %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: unauthorized non-tool call reached backend count=%d", count)
	}
}

// Scenario 7: AgentGate unavailable -> Fail closed, Backend count = 0
//
// Before G7 Task A this test asserted `var nilServer *authz.Server; if
// nilServer != nil { t.Fatal(...) }` — a tautology on a variable that is
// never assigned anything but nil, followed by an unrelated raw-socket
// connection-refused check against an address nothing in this codebase
// serves. It could not fail and proved nothing about AgentGate's actual
// fail-closed behavior. This version wires a DecisionEvaluator that returns
// a genuine error (simulating the Cedar/database layer being unreachable)
// through the real production path — authz.Server -> audit.AuditedDecisionService
// -> the failing evaluator — and asserts the call is denied, not allowed.
//
// Note for AI/Gateway team (G7 Task A, environment half): this is the
// Go-level unit proof that a decision-evaluation failure denies. The live-
// topology equivalent — physically stopping the agentgate process and
// confirming agentgateway's ext_authz callout fails closed — belongs in
// deploy/g6's live matrix, not this Go unit test; it needs real process
// orchestration this file has no access to.
func TestScenario07_AgentGateUnavailable(t *testing.T) {
	h := setupInProcessHarnessWithEvaluator(t, failingEvaluator{})

	jwtClaims, _ := structpb.NewStruct(map[string]any{
		"sub":          "agent-reader",
		"roles":        "reader",
		"workspace_id": "default",
	})
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					// A well-formed, otherwise-allowed request: if the evaluator's
					// outage were not fail-closed, this would return ALLOW.
					Body: `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"read_status","arguments":{"verbose":true}}}`,
				},
			},
			MetadataContext: &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.filters.http.jwt_authn": jwtClaims,
				},
			},
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: decision-service outage was not denied, got allowed=%v code=%d backendCount=%d", allowed, code, h.backendCalls)
	}
}

// Scenario 8: Policy evaluation failure (broken role without amount) -> Backend count = 0

func TestScenario08_PolicyEvaluationFailure(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"read_status"}}`,
				},
			},
			// Valid identity with an unrecognized role — this scenario proves
			// Cedar's own evaluation failure denies, not a missing-identity
			// short-circuit before Cedar is ever reached.
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-broken",
				"roles":        "broken",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: policy evaluation failure was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario08_PolicyEvaluationFailure_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	// Broken role deliberately triggers Cedar evaluation error when context.amount is absent
	body := `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"read_status"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-broken",
		"roles": "broken",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on policy evaluation failure, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: policy evaluation failure reached backend count=%d", count)
	}
}

// Scenario 9: Tool fingerprint / schema mismatch -> Backend count = 0

func TestScenario09_ToolFingerprintMismatch(t *testing.T) {
	h := setupInProcessHarness(t)

	// drifted_tool was registered with dummy zero hash, live fingerprint check causes drift.
	// x-tool-fingerprint is read directly from headers by the adapter (not an
	// identity field), so it stays as a header; identity itself must come
	// from JWT metadata for this scenario to reach the fingerprint check at all.
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"x-tool-fingerprint": "1111111111111111111111111111111111111111111111111111111111111111",
					},
					Body: `{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"drifted_tool"}}`,
				},
			},
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-reader",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: drifted tool was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario09_ToolFingerprintMismatch_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"read_status"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), mergeHeaders(jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), map[string]string{
		"x-tool-fingerprint": "mismatched_drift_fingerprint",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on tool fingerprint mismatch, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: drifted tool reached backend count=%d", count)
	}
}

// Scenario 10: Malicious client-supplied classification -> Backend count = 0

func TestScenario10_MaliciousClientClassification(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					// x-agentgate-risk is an attacker-controlled header the
					// adapter never reads for classification (verified: no
					// reference to it in internal/authz/adapter.go) — kept here
					// to document that it has no effect, not because it does
					// anything. Real classification comes only from
					// internal/toolregistry's registered risk for admin_action
					// (destructive). Identity must be authenticated (JWT
					// metadata, not headers) for this to prove policy denial
					// rather than a missing-identity short-circuit.
					Headers: map[string]string{
						"x-agentgate-risk": "read",
					},
					Body: `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"admin_action"}}`,
				},
			},
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-reader",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: spoofed classification was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario10_MaliciousClientClassification_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	// Attacker attempts to override risk level via client headers
	body := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"admin_action"}}`
	status, _, _ := sendMCPRequest(getGatewayURL(), mergeHeaders(jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), map[string]string{
		"x-agentgate-risk": "read",
		"x-tool-risk":      "read",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on spoofed classification, got status %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: spoofed classification reached backend count=%d", count)
	}
}

// Scenario 11: Malformed JSON / body disagreement -> Backend count = 0

func TestScenario11_MalformedJSON(t *testing.T) {
	h := setupInProcessHarness(t)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Headers: map[string]string{"x-agent-id": "agent-reader", "x-roles": "reader"},
					Body:    `{"jsonrpc":"2.0", broken syntax...`,
				},
			},
		},
	}
	allowed, code := h.executeCall(req)
	if allowed || code != int32(codes.PermissionDenied) || h.backendCalls != 0 {
		t.Fatalf("security violation: malformed JSON was not denied, backendCount=%d", h.backendCalls)
	}
}

func TestScenario11_MalformedJSON_LiveE2E(t *testing.T) {
	skipIfLiveUnavailable(t)
	resetBackendCount(t)
	body := `{"jsonrpc":"2.0", broken syntax...`
	// Valid identity + a tightened assertion (403, not merely "not 200") —
	// this scenario's original loose check would have silently accepted a
	// gateway-level 401 (no JWT) as "proof" of AgentGate's own malformed-body
	// handling, exactly the class of issue G7 Task A exists to close. With a
	// real JWT, this now genuinely reaches and exercises AgentGate's own
	// JSON parsing failure path.
	status, _, _ := sendMCPRequest(getGatewayURL(), jwtBearerHeader(map[string]any{
		"sub":   "agent-reader",
		"roles": "reader",
	}), []byte(body))
	if status != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 on broken JSON, got %d", status)
	}
	if count := getBackendCount(t); count != 0 {
		t.Fatalf("security violation: malformed JSON reached backend count=%d", count)
	}
}

// Scenario 12: Oversized body with an undeclared argument -> the undeclared
// argument must never reach policy evaluation, regardless of the resulting
// ALLOW/DENY outcome.
//
// Before G7 Task A this test's only real assertion was on the DENY path's
// error code; on the ALLOW path it just called t.Log and always passed. That
// framing assumed ALLOW was the wrong outcome — it isn't. There is no
// request-body size limit anywhere in the Go adapter (verified: no such
// check exists in internal/authz), and argdecl's whitelist already causes
// undeclared arguments to be silently dropped before they can reach
// decision.Request.Arguments (internal/argdecl's own package doc: "Undeclared
// arguments in raw are silently ignored (never reach policy)"). So for
// read_status (which declares only "verbose"), a huge undeclared "payload"
// argument is expected to be ALLOWED — the actual security property is that
// "payload" itself must never appear in what Cedar evaluated. This version
// asserts that directly via the evaluator's captured last request, instead of
// inferring it indirectly from an ALLOW/DENY code that can't distinguish
// "stripped safely" from "leaked but didn't matter."
//
// A request body size limit is a distinct, legitimate hardening item (DoS
// protection, not authorization correctness) — already tracked as backend
// work in docs/PHASES/PROGRESS_AND_ROADMAP.md §3 (G10, "Request size + time
// limits; rate limiting"). Not added here: out of this ticket's scope.
func TestScenario12_OversizedBody(t *testing.T) {
	eng, err := decision.NewEngine([]byte(fixturepolicy.CedarSource))
	if err != nil {
		t.Fatalf("setup cedar engine: %v", err)
	}
	evaluator := &staticEvaluator{eng: eng}
	h := setupInProcessHarnessWithEvaluator(t, evaluator)

	hugePayload := strings.Repeat("A", 1048576+100)
	req := &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Body: fmt.Sprintf(`{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"read_status","arguments":{"verbose":true,"payload":"%s"}}}`, hugePayload),
				},
			},
			MetadataContext: jwtMetadata(map[string]any{
				"sub":          "agent-reader",
				"roles":        "reader",
				"workspace_id": "default",
			}),
		},
	}
	allowed, code := h.executeCall(req)

	if !allowed {
		// Also acceptable, as long as it's a genuine policy denial, not an
		// adapter crash — a malformed/Internal code here would itself be a bug.
		if code != int32(codes.PermissionDenied) {
			t.Fatalf("expected either ALLOW (undeclared arg safely stripped) or PermissionDenied, got code %d", code)
		}
		return
	}

	// The actual security property: the undeclared, oversized "payload"
	// argument must never have reached the policy engine.
	lastArgs := evaluator.LastRequest().Arguments
	if _, leaked := lastArgs["payload"]; leaked {
		t.Fatalf("security violation: undeclared oversized argument \"payload\" reached decision.Request.Arguments (leaked into policy evaluation)")
	}
	// The declared argument should still be present and correct.
	if v, ok := lastArgs["verbose"]; !ok || v.String() != "true" {
		t.Fatalf("expected declared argument \"verbose\"=true to reach policy evaluation, got %+v", lastArgs["verbose"])
	}
}
