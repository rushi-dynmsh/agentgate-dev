import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { X } from "lucide-react";
import { useGovernance } from "../../state/GovernanceProvider";
import { toLifecycleSummaryView } from "@contract/view/policy-lifecycle-view";
import { StatusDot, type StatusTone } from "../../components/ui/StatusDot";
import type { PolicyRecord } from "@contract/models/governance";

const STATE_TONE: Record<PolicyRecord["state"], StatusTone> = {
  active: "allow",
  candidate: "candidate",
  historical: "neutral",
};

function formatVersion(version: string | undefined): string {
  if (!version) return "none";
  if (version.length <= 24) return version;
  return `${version.slice(0, 12)}...${version.slice(-8)}`;
}

export function DashboardPage() {
  const { state, ready } = useGovernance();
  const summary = toLifecycleSummaryView(state.workspaceId, state.policies);
  const activePolicy = state.policies.find((p) => p.state === "active");
  const [showActivePolicy, setShowActivePolicy] = useState(false);

  useEffect(() => {
    if (!showActivePolicy) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setShowActivePolicy(false);
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [showActivePolicy]);

  return (
    <div className="ag-page">
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">Welcome back</h1>
          <p className="ag-page-subtitle">
            Monitor and govern what your AI agents are allowed to do.
          </p>
        </div>
      </div>

      {!ready ? (
        <div className="ag-empty">Loading workspace state&hellip;</div>
      ) : (
        <>
          <div className="ag-stat-grid">
            <div className="ag-stat-card">
              <div className="ag-stat-label">Total policies</div>
              <div className="ag-stat-value">{summary.totalPolicies}</div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label" style={{ color: "var(--ag-allow)" }}>Active version</div>
              <div
                className="ag-stat-value ag-stat-value-hash ag-mono"
                style={{ fontSize: 18, color: "var(--ag-allow)" }}
                title={summary.activeVersion ?? "No active policy"}
              >
                {formatVersion(summary.activeVersion)}
              </div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label" style={{ color: "var(--ag-candidate)" }}>Candidates</div>
              <div className="ag-stat-value" style={{ color: "var(--ag-candidate)" }}>{summary.candidateCount}</div>
            </div>
            <div className="ag-stat-card">
              <div className="ag-stat-label">Historical</div>
              <div className="ag-stat-value">{summary.historicalCount}</div>
            </div>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1.4fr 1fr", gap: 24 }}>
            <div>
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12 }}>
                <h2 style={{ fontSize: 14, fontWeight: 500, margin: 0 }}>Recent policies</h2>
                <Link to="/policies" className="ag-btn ag-btn-ghost ag-btn-sm">
                  View all
                </Link>
              </div>
              {state.policies.length === 0 ? (
                <div className="ag-empty">No policies yet.</div>
              ) : (
                <table className="ag-table">
                  <thead>
                    <tr>
                      <th>Version</th>
                      <th>Status</th>
                      <th>Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.policies.slice(0, 5).map((p) => (
                      <tr key={p.version}>
                        <td className="ag-mono">{p.version}</td>
                        <td>
                          <StatusDot tone={STATE_TONE[p.state]}>{p.state}</StatusDot>
                        </td>
                        <td>{p.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            <div>
              <h2 style={{ fontSize: 14, fontWeight: 500, margin: "0 0 12px" }}>Active policy</h2>
              <button
                type="button"
                className="ag-card ag-active-policy-card"
                onClick={() => activePolicy && setShowActivePolicy(true)}
                disabled={!activePolicy}
                aria-label={activePolicy ? "View active policy details" : "No active policy"}
              >
                <div className="ag-card-body">
                  {activePolicy ? (
                    <>
                      <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 6 }}>
                        {activePolicy.description}
                      </div>
                      <div className="ag-mono" style={{ marginBottom: 12 }}>
                        {activePolicy.version}
                      </div>
                      <StatusDot tone="allow">Active</StatusDot>
                    </>
                  ) : (
                    <div style={{ color: "var(--ag-text-secondary)", fontSize: 13.5 }}>
                      No policy is currently active. Authorization decisions will fail closed
                      until one is activated.
                    </div>
                  )}
                </div>
              </button>
            </div>
          </div>

          {showActivePolicy && activePolicy && (
            <div
              className="ag-modal-backdrop"
              role="presentation"
              onClick={() => setShowActivePolicy(false)}
            >
              <section
                className="ag-modal"
                role="dialog"
                aria-modal="true"
                aria-labelledby="active-policy-dialog-title"
                onClick={(event) => event.stopPropagation()}
              >
                <div className="ag-modal-header">
                  <div>
                    <div className="ag-eyebrow">Active policy</div>
                    <h2 id="active-policy-dialog-title" className="ag-modal-title">
                      Policy details
                    </h2>
                  </div>
                  <button
                    type="button"
                    className="ag-icon-button"
                    onClick={() => setShowActivePolicy(false)}
                    aria-label="Close policy details"
                    title="Close"
                  >
                    <X size={17} />
                  </button>
                </div>

                <div className="ag-modal-meta">
                  <div>
                    <span>Version</span>
                    <strong className="ag-mono">{activePolicy.version}</strong>
                  </div>
                  <div>
                    <span>Workspace</span>
                    <strong className="ag-mono">{activePolicy.workspace_id}</strong>
                  </div>
                  <div>
                    <span>Status</span>
                    <StatusDot tone="allow">Active</StatusDot>
                  </div>
                  <div>
                    <span>Activated</span>
                    <strong>{activePolicy.activated_at ? new Date(activePolicy.activated_at).toLocaleString() : "Not recorded"}</strong>
                  </div>
                </div>

                <div className="ag-modal-section">
                  <h3>Description</h3>
                  <p>{activePolicy.description || "No description provided."}</p>
                </div>

                <div className="ag-modal-section">
                  <h3>Cedar policy source</h3>
                  <pre className="ag-policy-source">{activePolicy.content}</pre>
                </div>
              </section>
            </div>
          )}
        </>
      )}
    </div>
  );
}
