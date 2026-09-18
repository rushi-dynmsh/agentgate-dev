import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useGovernance } from "../../state/GovernanceProvider";
import type { DryRunSample, DryRunCompareResponse } from "@contract/models/governance";
import { StatusDot } from "../../components/ui/StatusDot";

const SAMPLE_REQUESTS: DryRunSample[] = [
  { principal_id: "agent-reader", principal_roles: ["reader"], backend_id: "default", tool_name: "read_status", risk: "read" },
  { principal_id: "agent-reader", principal_roles: ["reader"], backend_id: "default", tool_name: "write_status", risk: "write" },
  { principal_id: "agent-admin", principal_roles: ["admin"], backend_id: "default", tool_name: "write_status", risk: "write" },
  { principal_id: "agent-admin", principal_roles: ["admin"], backend_id: "default", tool_name: "read_status", risk: "read" },
  { principal_id: "agent-reader", principal_roles: ["reader"], backend_id: "default", tool_name: "delete_status", risk: "destructive" },
];

export function DryRunPage() {
  const { version } = useParams<{ version: string }>();
  const { dryRunCompare } = useGovernance();
  const navigate = useNavigate();

  const [result, setResult] = useState<DryRunCompareResponse | null>(null);
  const [running, setRunning] = useState(false);

  const handleRunDryRun = async () => {
    if (!version) return;
    setRunning(true);
    try {
      setResult(await dryRunCompare(version, SAMPLE_REQUESTS));
    } finally {
      setRunning(false);
    }
  };

  const changedCount = result?.results.filter((r) => r.changed).length ?? 0;
  const newAllow = result?.results.filter((r) => r.changed && r.candidate_decision === "ALLOW").length ?? 0;
  const newDeny = result?.results.filter((r) => r.changed && r.candidate_decision === "DENY").length ?? 0;

  return (
    <div className="ag-page">
      <div className="ag-eyebrow">
        <Link to="/policies">Policies</Link> &nbsp;/&nbsp; <span className="ag-mono">{version}</span> &nbsp;/&nbsp; Dry run
      </div>
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">Dry run results</h1>
          <p className="ag-page-subtitle">
            Comparison against the current active policy for candidate <span className="ag-mono">{version}</span>.
          </p>
        </div>
      </div>

      {!result ? (
        <div style={{ padding: "56px 0", textAlign: "center" }}>
          <p style={{ color: "var(--ag-text-secondary)", marginBottom: 18, fontSize: 14 }}>
            Run a comparison across {SAMPLE_REQUESTS.length} representative sample requests to see
            exactly which decisions would change if this candidate were activated.
          </p>
          <button className="ag-btn ag-btn-primary" onClick={handleRunDryRun} disabled={running}>
            {running ? "Running…" : "Run dry run"}
          </button>
        </div>
      ) : (
        <>
          <div className="ag-stat-grid">
            <div className="ag-stat-card">
              <div className="ag-stat-label">Total compared</div>
              <div className="ag-stat-value">{result.results.length}</div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label" style={{ color: "var(--ag-candidate)" }}>Changed decisions</div>
              <div className="ag-stat-value" style={{ color: "var(--ag-candidate)" }}>{changedCount}</div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label" style={{ color: "var(--ag-allow)" }}>New allow</div>
              <div className="ag-stat-value" style={{ color: "var(--ag-allow)" }}>{newAllow}</div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label" style={{ color: "var(--ag-deny)" }}>New deny</div>
              <div className="ag-stat-value" style={{ color: "var(--ag-deny)" }}>{newDeny}</div>
            </div>
          </div>

          <table className="ag-table" style={{ marginBottom: 24 }}>
            <thead>
              <tr>
                <th>Principal</th>
                <th>Tool</th>
                <th>Active</th>
                <th>Candidate</th>
                <th>Change</th>
              </tr>
            </thead>
            <tbody>
              {result.results.map((r, i) => {
                const sample = SAMPLE_REQUESTS[i];
                return (
                  <tr key={i}>
                    <td className="ag-mono">{sample?.principal_id}</td>
                    <td className="ag-mono">{sample?.tool_name}</td>
                    <td>
                      <StatusDot tone={r.active_decision === "ALLOW" ? "allow" : "deny"}>
                        {r.active_decision}
                      </StatusDot>
                    </td>
                    <td>
                      <StatusDot tone={r.candidate_decision === "ALLOW" ? "allow" : "deny"}>
                        {r.candidate_decision}
                      </StatusDot>
                    </td>
                    <td style={{ color: r.changed ? "var(--ag-candidate)" : "var(--ag-text-muted)", fontWeight: r.changed ? 500 : 400 }}>
                      {r.changed ? "Changed" : "No change"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>

          <div style={{ display: "flex", justifyContent: "flex-end", gap: 10 }}>
            <Link to="/policies" className="ag-btn ag-btn-ghost">Back</Link>
            <button
              className="ag-btn ag-btn-primary"
              onClick={() => navigate(`/policies/${encodeURIComponent(version ?? "")}/activate`)}
            >
              Proceed to activation
            </button>
          </div>
        </>
      )}
    </div>
  );
}
