import { useEffect, useState } from "react";
import { X } from "lucide-react";
import {
  deriveGovernanceState,
  toolIDString,
  type ToolGovernanceView,
} from "@contract/models/tool-inventory";
import { ComingSoonPage } from "../../components/ui/ComingSoonPage";
import { StatusDot, type StatusTone } from "../../components/ui/StatusDot";

const FIXTURE_TOOLS: ToolGovernanceView[] = [
  { toolId: { backendId: "customer-mcp", toolName: "get_customer" }, known: true, risk: "read", registeredFingerprint: "a19cc0d4...", driftStatus: "none" },
  { toolId: { backendId: "customer-mcp", toolName: "update_customer" }, known: true, risk: "write", registeredFingerprint: "5f2b8e91...", driftStatus: "none" },
  { toolId: { backendId: "customer-mcp", toolName: "delete_customer" }, known: true, risk: "destructive", registeredFingerprint: "e70a5c22...", driftStatus: "none" },
  { toolId: { backendId: "customer-mcp", toolName: "export_customer_report" }, known: true, risk: "write", registeredFingerprint: "9d31af04...", driftStatus: "detected" },
  { toolId: { backendId: "billing-mcp", toolName: "issue_refund" }, known: false, driftStatus: "unknown_tool" },
];

const STATE_TONE: Record<string, StatusTone> = {
  authorized: "allow",
  unknown: "deny",
  drifted: "deny",
  missing_risk: "candidate",
  stale: "candidate",
  error: "deny",
};

const STATE_LABEL: Record<string, string> = {
  authorized: "Authorized",
  unknown: "Unknown tool",
  drifted: "Drift detected",
  missing_risk: "Missing risk",
  stale: "Stale",
  error: "Error",
};

export function ToolsPage() {
  const [selectedTool, setSelectedTool] = useState<ToolGovernanceView | null>(null);

  useEffect(() => {
    if (!selectedTool) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setSelectedTool(null);
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [selectedTool]);

  return (
    <ComingSoonPage
      title="Tools & resources"
      subtitle="Manage tool classifications and governance."
      note="Preview data. The real toolregistry contract (G2) exists and is tested, but no GovernanceClient method exposes it to a live backend yet — this is fixture data, not a live feed."
    >
      <table className="ag-table">
        <thead>
          <tr>
            <th>Tool</th>
            <th>Backend</th>
            <th>Risk</th>
            <th>Governance state</th>
            <th>Schema fingerprint</th>
          </tr>
        </thead>
        <tbody>
          {FIXTURE_TOOLS.map((tool) => {
            const s = deriveGovernanceState(tool);
            return (
              <tr
                key={toolIDString(tool.toolId)}
                className="ag-clickable-row"
                onClick={() => setSelectedTool(tool)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setSelectedTool(tool);
                  }
                }}
                tabIndex={0}
                aria-label={`View details for ${tool.toolId.toolName}`}
              >
                <td className="ag-mono" style={{ color: "var(--ag-text-primary)" }}>{tool.toolId.toolName}</td>
                <td className="ag-mono">{tool.toolId.backendId}</td>
                <td>{tool.risk ?? "—"}</td>
                <td>
                  <StatusDot tone={STATE_TONE[s] ?? "neutral"}>{STATE_LABEL[s] ?? s}</StatusDot>
                </td>
                <td className="ag-mono">{tool.registeredFingerprint?.slice(0, 12) ?? "—"}</td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {selectedTool && (() => {
        const state = deriveGovernanceState(selectedTool);
        return (
          <div
            className="ag-modal-backdrop"
            role="presentation"
            onClick={() => setSelectedTool(null)}
          >
            <section
              className="ag-modal"
              role="dialog"
              aria-modal="true"
              aria-labelledby="tool-details-title"
              onClick={(event) => event.stopPropagation()}
            >
              <div className="ag-modal-header">
                <div>
                  <div className="ag-eyebrow">Tool governance</div>
                  <h2 id="tool-details-title" className="ag-modal-title">
                    {selectedTool.toolId.toolName}
                  </h2>
                </div>
                <button
                  type="button"
                  className="ag-icon-button"
                  onClick={() => setSelectedTool(null)}
                  aria-label="Close tool details"
                  title="Close"
                >
                  <X size={17} />
                </button>
              </div>

              <div className="ag-modal-meta">
                <div>
                  <span>Backend</span>
                  <strong className="ag-mono">{selectedTool.toolId.backendId}</strong>
                </div>
                <div>
                  <span>Risk</span>
                  <strong>{selectedTool.risk ?? "Not classified"}</strong>
                </div>
                <div>
                  <span>Governance state</span>
                  <StatusDot tone={STATE_TONE[state] ?? "neutral"}>{STATE_LABEL[state] ?? state}</StatusDot>
                </div>
                <div>
                  <span>Schema status</span>
                  <strong>{selectedTool.driftStatus === "none" ? "Matches registered schema" : selectedTool.driftStatus}</strong>
                </div>
              </div>

              <div className="ag-modal-section">
                <h3>Tool identity</h3>
                <p className="ag-mono">{toolIDString(selectedTool.toolId)}</p>
              </div>

              <div className="ag-modal-section">
                <h3>Registered schema fingerprint</h3>
                <pre className="ag-policy-source">{selectedTool.registeredFingerprint ?? "No fingerprint registered."}</pre>
              </div>
            </section>
          </div>
        );
      })()}
    </ComingSoonPage>
  );
}
