// Package credential defines the boundary by which AgentGate obtains a
// downstream credential for the real MCP backend, resolving O-001
// (docs/DECISIONS/OPEN_DECISIONS.md).
//
// # Status: interface only, no concrete Issuer yet
//
// This is G7 Task B's "design half"
// (docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md §3). The *implementation*
// half — a concrete Issuer for a real backend — is explicitly blocked on
// AI/Gateway's Task B choosing which real/realistic backend to target and
// documenting what credential model it needs
// (deploy/demo/CREDENTIAL_REQUIREMENTS.md, not yet written). Building a
// concrete mechanism against a guessed backend risks having to redo it once
// the real requirements are known — so this package defines only the shape
// AgentGate's own side of that exchange needs, not how any particular
// backend's credential protocol works.
//
// # Trust boundary
//
// An Issuer is invoked only after Cedar has already produced [decision.Allow]
// — never before, and never to influence the authorization decision itself.
// Its own failure is a distinct, later failure mode: the call was authorized,
// but AgentGate could not safely reach the backend on the caller's behalf,
// so it must still deny (docs/SECURITY/PRODUCTION-INVARIANTS.md fail-closed;
// [decision.ReasonCredentialIssuanceFailure]).
//
// # Non-negotiables for any future Issuer implementation
//
//   - The credential handed to the backend must never be the client's raw
//     inbound bearer token (no passthrough) — see
//     docs/SECURITY/PRODUCTION-INVARIANTS.md.
//   - Issuance failure must deny the call, not send it uncredentialed.
//   - The caller/on-behalf-of identity that produced this credential must
//     remain auditable — see [Credential.Ref] and
//     internal/audit.DecisionRecord.DownstreamCredentialRef.
//   - Replay/confusion risk should be addressed appropriately for the
//     chosen mechanism — e.g. scoping the credential to the specific
//     decision.Request's ExecutionID where the backend's model supports it.
package credential

import (
	"context"

	"github.com/Dynamisch-LLC/agentgate/internal/decision"
)

// Credential is a downstream credential to attach to an already-ALLOWed
// call before it reaches the real MCP backend.
type Credential struct {
	// HeaderName is the HTTP header agentgateway should set on the upstream
	// request — e.g. "Authorization" for a bearer/OAuth token, or a
	// backend-specific API-key header. Not fixed by this package: it
	// depends on the concrete backend AI/Gateway's Task B chooses.
	HeaderName string

	// HeaderValue is the credential material for that header. Never the
	// client's raw inbound bearer token — see the package doc's trust
	// boundary. Never logged or persisted; only [Credential.Ref] is
	// audit-safe.
	HeaderValue string

	// Ref is a non-secret reference or audience identifier for this
	// credential — safe to persist in the audit trail
	// (internal/audit.DecisionRecord.DownstreamCredentialRef). What it
	// contains is mechanism-specific (a token ID, an audience URI, a scope
	// string); the one hard rule is that reconstructing HeaderValue from
	// Ref alone must not be possible.
	Ref string
}

// Issuer mints a downstream [Credential] for a call Cedar has already
// allowed. Concrete implementations are backend-specific (API-key
// injection, OAuth token exchange, a short-lived scoped PAT, ...); this
// package intentionally does not provide one — see the package doc.
type Issuer interface {
	// Issue is called only for a decision.Request that has already produced
	// res.Decision == decision.Allow. A non-nil error means the caller MUST
	// deny the request rather than forward it uncredentialed — see
	// [decision.ReasonCredentialIssuanceFailure].
	Issue(ctx context.Context, req decision.Request, res decision.Result) (Credential, error)
}
