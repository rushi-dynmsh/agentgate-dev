package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dynamisch-LLC/agentgate/internal/audit"
)

func TestChainVerifier_ValidChain(t *testing.T) {
	store := audit.NewMemoryStore()
	defer store.Close()
	ctx := context.Background()
	ws := "ws-chain-valid"

	prevHash := audit.GenesisHash
	for i := 1; i <= 3; i++ {
		rec := audit.DecisionRecord{
			WorkspaceID:       ws,
			ExecutionID:       "ex-chain",
			Timestamp:         time.Now().UTC(),
			Decision:          "ALLOW",
			Reason:            "policy_allow",
			PrincipalAgentID:  "agent",
			ToolBackendID:     "b",
			ToolName:          "t",
			ToolRisk:          "read",
			PolicyVersion:     "v1",
			PolicyHash:        "v1",
			RedactedArguments: map[string]string{"k": "v"},
			PrevHash:          prevHash,
		}
		canonical, err := audit.ComputeCanonicalPayload(rec)
		if err != nil {
			t.Fatal(err)
		}
		rec.CanonicalPayload = string(canonical)
		rec.RowHash = audit.ComputeRowHash(prevHash, canonical)
		prevHash = rec.RowHash

		_, err = store.AppendDecision(ctx, rec)
		if err != nil {
			t.Fatal(err)
		}
	}

	verifier := audit.NewChainVerifier()
	res, err := verifier.VerifyWorkspace(ctx, store, ws)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid {
		t.Fatalf("expected valid chain, got invalid: %s", res.Details)
	}
	if res.TotalRecords != 3 {
		t.Fatalf("expected 3 records, got %d", res.TotalRecords)
	}
}

// TestChainVerifier_DownstreamCredentialRefIsChainProtected is a regression
// test for a specific footgun: ComputeCanonicalPayload and VerifyWorkspace's
// own record reconstruction must copy DownstreamCredentialRef identically,
// or every record with it populated would spuriously fail verification (the
// recomputed canonical payload would omit a field the stored one includes).
// G7, added alongside internal/credential's package boundary — see
// docs/PHASES/G7_WORKSTREAMS/01_BACKEND_G7.md §3.
func TestChainVerifier_DownstreamCredentialRefIsChainProtected(t *testing.T) {
	store := audit.NewMemoryStore()
	defer store.Close()
	ctx := context.Background()
	ws := "ws-credential-ref"

	rec := audit.DecisionRecord{
		WorkspaceID:             ws,
		ExecutionID:             "ex-cred",
		Timestamp:               time.Now().UTC(),
		Decision:                "ALLOW",
		Reason:                  "policy_allow",
		PrincipalAgentID:        "agent",
		ToolBackendID:           "b",
		ToolName:                "t",
		ToolRisk:                "read",
		PolicyVersion:           "v1",
		PolicyHash:              "v1",
		RedactedArguments:       map[string]string{"k": "v"},
		DownstreamCredentialRef: "cred-ref-abc123",
		PrevHash:                audit.GenesisHash,
	}
	canonical, err := audit.ComputeCanonicalPayload(rec)
	if err != nil {
		t.Fatal(err)
	}
	rec.CanonicalPayload = string(canonical)
	rec.RowHash = audit.ComputeRowHash(audit.GenesisHash, canonical)

	if _, err := store.AppendDecision(ctx, rec); err != nil {
		t.Fatal(err)
	}

	verifier := audit.NewChainVerifier()
	res, err := verifier.VerifyWorkspace(ctx, store, ws)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid {
		t.Fatalf("expected a record with DownstreamCredentialRef set to verify cleanly, got invalid: %s", res.Details)
	}

	// Tampering with the credential ref specifically (not any other field)
	// must still be caught, exactly like tampering with any other field.
	records, err := store.ListRecords(ctx, ws, 1)
	if err != nil {
		t.Fatal(err)
	}
	records[0].DownstreamCredentialRef = "tampered-ref"
	tamperedCanonical, err := audit.ComputeCanonicalPayload(audit.DecisionRecord{
		WorkspaceID:             records[0].WorkspaceID,
		ExecutionID:             records[0].ExecutionID,
		Timestamp:               records[0].Timestamp,
		EventType:               records[0].EventType,
		Decision:                records[0].Decision,
		Reason:                  records[0].Reason,
		PrincipalAgentID:        records[0].PrincipalAgentID,
		PrincipalRoles:          records[0].PrincipalRoles,
		PrincipalOnBehalfOf:     records[0].PrincipalOnBehalfOf,
		ToolBackendID:           records[0].ToolBackendID,
		ToolName:                records[0].ToolName,
		ToolRisk:                records[0].ToolRisk,
		PolicyVersion:           records[0].PolicyVersion,
		PolicyHash:              records[0].PolicyHash,
		RedactedArguments:       records[0].RedactedArguments,
		DownstreamCredentialRef: records[0].DownstreamCredentialRef,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(tamperedCanonical) == rec.CanonicalPayload {
		t.Fatal("expected tampering DownstreamCredentialRef to change the canonical payload — it appears not to be chain-protected")
	}
}

func TestChainVerifier_DetectsTamperedRow(t *testing.T) {
	store := audit.NewMemoryStore()
	defer store.Close()
	ctx := context.Background()
	ws := "ws-tamper"

	rec := audit.DecisionRecord{
		WorkspaceID:      ws,
		ExecutionID:      "ex-tamper",
		Timestamp:        time.Now().UTC(),
		Decision:         "ALLOW",
		Reason:           "policy_allow",
		PrincipalAgentID: "agent",
		ToolBackendID:    "b",
		ToolName:         "t",
		ToolRisk:         "read",
		PrevHash:         audit.GenesisHash,
	}
	canonical, _ := audit.ComputeCanonicalPayload(rec)
	rec.CanonicalPayload = string(canonical)
	rec.RowHash = audit.ComputeRowHash(audit.GenesisHash, canonical)
	stored, _ := store.AppendDecision(ctx, rec)

	// Tamper simulation on store
	tampered := *stored
	tampered.Decision = "DENY" // modified payload without hash update
	store.SimulateTamper(ws, stored.SequenceNumber, tampered)

	verifier := audit.NewChainVerifier()
	res, err := verifier.VerifyWorkspace(ctx, store, ws)
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid {
		t.Fatal("verifier should have caught tampered row!")
	}
	if res.BrokenSequence != 1 {
		t.Fatalf("expected broken at seq 1, got %d", res.BrokenSequence)
	}
}

func TestChainVerifier_DetectsBrokenPrevHashChain(t *testing.T) {
	store := audit.NewMemoryStore()
	defer store.Close()
	ctx := context.Background()
	ws := "ws-chain-break"

	// Record 1
	rec1 := audit.DecisionRecord{
		WorkspaceID:      ws,
		ExecutionID:      "ex-1",
		Timestamp:        time.Now().UTC(),
		Decision:         "ALLOW",
		Reason:           "policy_allow",
		PrincipalAgentID: "agent",
		ToolBackendID:    "b",
		ToolName:         "t",
		ToolRisk:         "read",
		PrevHash:         audit.GenesisHash,
	}
	c1, _ := audit.ComputeCanonicalPayload(rec1)
	rec1.CanonicalPayload = string(c1)
	rec1.RowHash = audit.ComputeRowHash(audit.GenesisHash, c1)
	_, _ = store.AppendDecision(ctx, rec1)

	// Record 2 with broken prev_hash
	rec2 := audit.DecisionRecord{
		WorkspaceID:      ws,
		ExecutionID:      "ex-2",
		Timestamp:        time.Now().UTC(),
		Decision:         "ALLOW",
		Reason:           "policy_allow",
		PrincipalAgentID: "agent",
		ToolBackendID:    "b",
		ToolName:         "t",
		ToolRisk:         "read",
		PrevHash:         "invalid-broken-prev-hash",
	}
	c2, _ := audit.ComputeCanonicalPayload(rec2)
	rec2.CanonicalPayload = string(c2)
	rec2.RowHash = audit.ComputeRowHash(rec2.PrevHash, c2)
	_, _ = store.AppendDecision(ctx, rec2)

	verifier := audit.NewChainVerifier()
	res, err := verifier.VerifyWorkspace(ctx, store, ws)
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid {
		t.Fatal("verifier should have caught broken prev_hash linkage")
	}
	if res.BrokenSequence != 2 {
		t.Fatalf("expected broken at seq 2, got %d", res.BrokenSequence)
	}
}
